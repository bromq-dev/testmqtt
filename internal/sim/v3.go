package sim

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	"github.com/charmbracelet/lipgloss"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// RunV3 runs the MQTT v3.1.1 traffic simulator
func RunV3(cfg Config) error {
	// Styles for output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Println(headerStyle.Render("MQTT v3.1.1 Traffic Simulator"))
	fmt.Println()

	// Check source broker connectivity
	fmt.Printf("Connecting to source: %s\n", cfg.Source)
	if err := common.CheckBrokerReachable(cfg.Source); err != nil {
		return fmt.Errorf("source broker not reachable: %w", err)
	}

	// Check target broker connectivity
	fmt.Printf("Connecting to target: %s\n", cfg.Broker)
	if err := common.CheckBrokerReachable(cfg.Broker); err != nil {
		return fmt.Errorf("target broker not reachable: %w", err)
	}

	// Message counters and shutdown flag
	var receivedCount uint64
	var deliveredCount uint64
	var shuttingDown atomic.Bool

	// Connect to target broker first (publisher)
	targetOpts := mqtt.NewClientOptions()
	targetOpts.AddBroker(cfg.Broker)
	targetOpts.SetClientID(common.GenerateClientID("sim-target"))
	targetOpts.SetCleanSession(true)
	targetOpts.SetConnectTimeout(5 * time.Second)
	targetOpts.SetAutoReconnect(true)
	targetOpts.SetKeepAlive(60 * time.Second)

	if cfg.Username != "" {
		targetOpts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		targetOpts.SetPassword(cfg.Password)
	}

	targetClient := mqtt.NewClient(targetOpts)
	token := targetClient.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("target broker connection timeout")
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to connect to target broker: %w", token.Error())
	}
	fmt.Println(successStyle.Render("  ✓ Connected to target broker"))

	// Message handler - republish to target with full passthrough
	onMessage := func(client mqtt.Client, msg mqtt.Message) {
		atomic.AddUint64(&receivedCount, 1)

		// Skip if shutting down
		if shuttingDown.Load() {
			return
		}

		if cfg.Verbose {
			fmt.Printf("%s [%s] QoS:%d Retain:%v Payload:%d bytes\n",
				infoStyle.Render("→"),
				msg.Topic(),
				msg.Qos(),
				msg.Retained(),
				len(msg.Payload()))
		}

		// Determine QoS and retain
		qos := msg.Qos()
		if cfg.QoS >= 0 {
			qos = byte(cfg.QoS)
		}
		retain := msg.Retained()
		if cfg.NoRetain {
			retain = false
		}

		// Fire and forget - count as sent when dispatched
		atomic.AddUint64(&deliveredCount, 1)
		go func(topic string, qos byte, retained bool, payload []byte) {
			if shuttingDown.Load() {
				return
			}
			targetClient.Publish(topic, qos, retained, payload)
		}(msg.Topic(), qos, retain, msg.Payload())
	}

	// Connect to source broker (subscriber)
	sourceOpts := mqtt.NewClientOptions()
	sourceOpts.AddBroker(cfg.Source)
	sourceOpts.SetClientID(common.GenerateClientID("sim-source"))
	sourceOpts.SetCleanSession(true)
	sourceOpts.SetConnectTimeout(5 * time.Second)
	sourceOpts.SetAutoReconnect(true)
	sourceOpts.SetKeepAlive(60 * time.Second)
	sourceOpts.SetDefaultPublishHandler(onMessage)

	if cfg.SourceUsername != "" {
		sourceOpts.SetUsername(cfg.SourceUsername)
	}
	if cfg.SourcePassword != "" {
		sourceOpts.SetPassword(cfg.SourcePassword)
	}

	sourceClient := mqtt.NewClient(sourceOpts)
	token = sourceClient.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		targetClient.Disconnect(250)
		return fmt.Errorf("source broker connection timeout")
	}
	if token.Error() != nil {
		targetClient.Disconnect(250)
		return fmt.Errorf("failed to connect to source broker: %w", token.Error())
	}
	fmt.Println(successStyle.Render("  ✓ Connected to source broker"))

	// Subscribe to source topic at QoS 2 to preserve message QoS
	token = sourceClient.Subscribe(cfg.Topic, 2, nil)
	if !token.WaitTimeout(5 * time.Second) {
		sourceClient.Disconnect(250)
		targetClient.Disconnect(250)
		return fmt.Errorf("subscribe timeout")
	}
	if token.Error() != nil {
		sourceClient.Disconnect(250)
		targetClient.Disconnect(250)
		return fmt.Errorf("failed to subscribe to source topic: %w", token.Error())
	}
	fmt.Printf(successStyle.Render("  ✓ Subscribed to: %s\n"), cfg.Topic)

	fmt.Println()
	fmt.Println(headerStyle.Render("Bridging traffic... (Ctrl+C to stop)"))
	fmt.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Status ticker
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastReceived, lastDelivered uint64

	for {
		select {
		case <-sigChan:
			fmt.Println()
			fmt.Println(headerStyle.Render("Shutting down..."))

			// Signal shutdown immediately
			shuttingDown.Store(true)

			// Disconnect clients (with short timeout)
			sourceClient.Disconnect(0)
			targetClient.Disconnect(0)

			finalReceived := atomic.LoadUint64(&receivedCount)
			finalDelivered := atomic.LoadUint64(&deliveredCount)
			fmt.Printf("\n%s Total: %d received, %d delivered\n", successStyle.Render("✓"), finalReceived, finalDelivered)
			return nil

		case <-ticker.C:
			received := atomic.LoadUint64(&receivedCount)
			delivered := atomic.LoadUint64(&deliveredCount)

			deltaReceived := received - lastReceived
			deltaDelivered := delivered - lastDelivered

			lastReceived = received
			lastDelivered = delivered

			// Calculate rates (per second, ticker is 5s)
			recvRate := float64(deltaReceived) / 5.0
			sentRate := float64(deltaDelivered) / 5.0

			// Calculate tick delivery percentage
			var tickPct float64
			if deltaReceived > 0 {
				tickPct = float64(deltaDelivered) / float64(deltaReceived) * 100
			}

			// Calculate total delivery percentage
			var totalPct float64
			if received > 0 {
				totalPct = float64(delivered) / float64(received) * 100
			}

			fmt.Printf("%s %d/%d (%.1f%%)  |  total: %d/%d (%.1f%%)  rate: %.1f/%.1f msg/s\n",
				infoStyle.Render("•"), deltaDelivered, deltaReceived, tickPct, delivered, received, totalPct, sentRate, recvRate)
		}
	}
}
