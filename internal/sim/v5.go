package sim

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/eclipse/paho.golang/paho"
)

// RunV5 runs the MQTT v5 traffic simulator
func RunV5(cfg Config) error {
	// Styles for output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Println(headerStyle.Render("MQTT v5 Traffic Simulator"))
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

	// Cancellable context for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Message counters and shutdown flag
	var receivedCount uint64
	var deliveredCount uint64
	var errorCount uint64
	var shuttingDown atomic.Bool

	// Semaphore to limit concurrent publishes
	sem := make(chan struct{}, cfg.QueueSize)

	// Target client with mutex for reconnection
	var targetMu sync.RWMutex
	var targetClient *paho.Client
	var targetConn interface{ Close() error }

	// Source connection with mutex for reconnection
	var sourceMu sync.Mutex
	var sourceConn interface{ Close() error }

	// Connect to target broker
	connectTarget := func() error {
		targetMu.Lock()
		defer targetMu.Unlock()

		if targetConn != nil {
			targetConn.Close()
		}

		conn, err := common.DialBroker(cfg.Broker)
		if err != nil {
			return fmt.Errorf("failed to dial target broker: %w", err)
		}

		client := paho.NewClient(paho.ClientConfig{
			ClientID: common.GenerateClientID("sim-target"),
			Conn:     conn,
		})

		cp := &paho.Connect{
			KeepAlive:  60,
			ClientID:   common.GenerateClientID("sim-target"),
			CleanStart: true,
		}
		if cfg.Username != "" {
			cp.UsernameFlag = true
			cp.Username = cfg.Username
		}
		if cfg.Password != "" {
			cp.PasswordFlag = true
			cp.Password = []byte(cfg.Password)
		}

		connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
		defer connectCancel()

		_, err = client.Connect(connectCtx, cp)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to connect to target broker: %w", err)
		}

		targetClient = client
		targetConn = conn
		return nil
	}

	// Message handler - republish to target
	onPublish := func(pr paho.PublishReceived) (bool, error) {
		atomic.AddUint64(&receivedCount, 1)

		if shuttingDown.Load() {
			return true, nil
		}

		// Try to acquire semaphore, drop if full
		select {
		case sem <- struct{}{}:
		default:
			return true, nil
		}

		// Determine QoS and retain
		qos := pr.Packet.QoS
		if cfg.QoS >= 0 {
			qos = byte(cfg.QoS)
		}
		retain := pr.Packet.Retain
		if cfg.NoRetain {
			retain = false
		}

		pub := &paho.Publish{
			Topic:   pr.Packet.Topic,
			QoS:     qos,
			Retain:  retain,
			Payload: pr.Packet.Payload,
		}

		if pr.Packet.Properties != nil {
			pub.Properties = &paho.PublishProperties{
				PayloadFormat:   pr.Packet.Properties.PayloadFormat,
				MessageExpiry:   pr.Packet.Properties.MessageExpiry,
				ContentType:     pr.Packet.Properties.ContentType,
				ResponseTopic:   pr.Packet.Properties.ResponseTopic,
				CorrelationData: pr.Packet.Properties.CorrelationData,
				User:            pr.Packet.Properties.User,
			}
		}

		if cfg.Verbose {
			fmt.Printf("%s [%s] QoS:%d Retain:%v Payload:%d bytes\n",
				infoStyle.Render("→"),
				pr.Packet.Topic,
				pr.Packet.QoS,
				pr.Packet.Retain,
				len(pr.Packet.Payload))
		}

		atomic.AddUint64(&deliveredCount, 1)

		go func() {
			defer func() { <-sem }()

			if shuttingDown.Load() {
				return
			}

			pubCtx, pubCancel := context.WithTimeout(ctx, cfg.Timeout)
			defer pubCancel()

			targetMu.RLock()
			client := targetClient
			targetMu.RUnlock()

			if client != nil {
				_, err := client.Publish(pubCtx, pub)
				if err != nil {
					atomic.AddUint64(&errorCount, 1)
				}
			}
		}()

		return true, nil
	}

	// Connect to source broker
	connectSource := func() error {
		sourceMu.Lock()
		defer sourceMu.Unlock()

		if sourceConn != nil {
			sourceConn.Close()
		}

		conn, err := common.DialBroker(cfg.Source)
		if err != nil {
			return fmt.Errorf("failed to dial source broker: %w", err)
		}

		client := paho.NewClient(paho.ClientConfig{
			ClientID:          common.GenerateClientID("sim-source"),
			Conn:              conn,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){onPublish},
		})

		cp := &paho.Connect{
			KeepAlive:  60,
			ClientID:   common.GenerateClientID("sim-source"),
			CleanStart: true,
		}
		if cfg.SourceUsername != "" {
			cp.UsernameFlag = true
			cp.Username = cfg.SourceUsername
		}
		if cfg.SourcePassword != "" {
			cp.PasswordFlag = true
			cp.Password = []byte(cfg.SourcePassword)
		}

		connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
		defer connectCancel()

		_, err = client.Connect(connectCtx, cp)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to connect to source broker: %w", err)
		}

		// Subscribe
		_, err = client.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: cfg.Topic, QoS: 2},
			},
		})
		if err != nil {
			client.Disconnect(&paho.Disconnect{ReasonCode: 0})
			conn.Close()
			return fmt.Errorf("failed to subscribe: %w", err)
		}

		sourceConn = conn
		return nil
	}

	// Initial connections
	if err := connectTarget(); err != nil {
		return err
	}
	fmt.Println(successStyle.Render("  ✓ Connected to target broker"))

	if err := connectSource(); err != nil {
		targetMu.Lock()
		if targetConn != nil {
			targetConn.Close()
		}
		targetMu.Unlock()
		return err
	}
	fmt.Println(successStyle.Render("  ✓ Connected to source broker"))
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

	var lastReceived, lastDelivered, lastErrors uint64
	var sourceStallCount int

	for {
		select {
		case <-sigChan:
			fmt.Println()
			fmt.Println(headerStyle.Render("Shutting down..."))

			shuttingDown.Store(true)
			cancel()

			sourceMu.Lock()
			if sourceConn != nil {
				sourceConn.Close()
			}
			sourceMu.Unlock()

			targetMu.Lock()
			if targetConn != nil {
				targetConn.Close()
			}
			targetMu.Unlock()

			finalReceived := atomic.LoadUint64(&receivedCount)
			finalDelivered := atomic.LoadUint64(&deliveredCount)
			fmt.Printf("\n%s Total: %d received, %d delivered\n", successStyle.Render("✓"), finalReceived, finalDelivered)
			return nil

		case <-ticker.C:
			received := atomic.LoadUint64(&receivedCount)
			delivered := atomic.LoadUint64(&deliveredCount)
			errors := atomic.LoadUint64(&errorCount)

			deltaReceived := received - lastReceived
			deltaDelivered := delivered - lastDelivered
			deltaErrors := errors - lastErrors

			lastReceived = received
			lastDelivered = delivered
			lastErrors = errors

			// Detect source stall (no messages received)
			if deltaReceived == 0 && received > 0 {
				sourceStallCount++
				if sourceStallCount >= 3 {
					fmt.Printf("%s Source stall detected, reconnecting...\n", warnStyle.Render("!"))
					if err := connectSource(); err != nil {
						fmt.Printf("%s Source reconnect failed: %v\n", warnStyle.Render("!"), err)
					} else {
						fmt.Printf("%s Reconnected to source broker\n", successStyle.Render("✓"))
						sourceStallCount = 0
					}
				}
			} else {
				sourceStallCount = 0
			}

			// Detect target issues (high error rate)
			if deltaErrors > 100 || (deltaDelivered > 0 && float64(deltaErrors)/float64(deltaDelivered) > 0.5) {
				fmt.Printf("%s High error rate (%d errors), reconnecting to target...\n", warnStyle.Render("!"), deltaErrors)
				if err := connectTarget(); err != nil {
					fmt.Printf("%s Target reconnect failed: %v\n", warnStyle.Render("!"), err)
				} else {
					fmt.Printf("%s Reconnected to target broker\n", successStyle.Render("✓"))
					// Reset error count after reconnect
					atomic.StoreUint64(&errorCount, 0)
					lastErrors = 0
				}
			}

			// Calculate rates
			recvRate := float64(deltaReceived) / 5.0
			sentRate := float64(deltaDelivered) / 5.0

			var tickPct float64
			if deltaReceived > 0 {
				tickPct = float64(deltaDelivered) / float64(deltaReceived) * 100
			}

			var totalPct float64
			if received > 0 {
				totalPct = float64(delivered) / float64(received) * 100
			}

			var timestamp string
			if cfg.UnixTimestamp {
				timestamp = fmt.Sprintf("%d", time.Now().Unix())
			} else {
				timestamp = time.Now().Format("2006-01-02 15:04:05")
			}

			// Show errors if any
			errStr := ""
			if deltaErrors > 0 {
				errStr = fmt.Sprintf("  err: %d", deltaErrors)
			}
			fmt.Printf("%s %d/%d (%.1f%%)  |  total: %d/%d (%.1f%%)  rate: %.1f/%.1f msg/s%s\n",
				infoStyle.Render(timestamp), deltaDelivered, deltaReceived, tickPct, delivered, received, totalPct, sentRate, recvRate, errStr)
		}
	}
}
