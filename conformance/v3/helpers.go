package v3

import (
	"fmt"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// CheckConnection performs a preflight check to verify broker connectivity and authentication
func CheckConnection(cfg common.Config) error {
	// First check TCP reachability
	if err := common.CheckBrokerReachable(cfg.Broker); err != nil {
		return fmt.Errorf("broker not reachable: %w", err)
	}

	// Now try an actual MQTT connection with auth
	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("preflight"), nil)
	if err != nil {
		if cfg.Username != "" {
			return fmt.Errorf("MQTT connection failed (check credentials): %w", err)
		}
		return fmt.Errorf("MQTT connection failed: %w", err)
	}

	client.Disconnect(250)
	return nil
}

// CreateAndConnectClient creates and connects a MQTT v3.1.1 client with optional message handler
func CreateAndConnectClient(cfg common.Config, clientID string, onMessage mqtt.MessageHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	if onMessage != nil {
		opts.SetDefaultPublishHandler(onMessage)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		return nil, fmt.Errorf("connection timeout")
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	return client, nil
}

// CreateAndConnectClientWithSession creates and connects a MQTT v3.1.1 client with Clean Session control
func CreateAndConnectClientWithSession(cfg common.Config, clientID string, cleanSession bool, onMessage mqtt.MessageHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(cleanSession)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	if onMessage != nil {
		opts.SetDefaultPublishHandler(onMessage)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		return nil, fmt.Errorf("connection timeout")
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	return client, nil
}

// CreateAndConnectClientWithWill creates a client with a will message
func CreateAndConnectClientWithWill(cfg common.Config, clientID string, willTopic string, willPayload []byte, willQos byte, willRetained bool, onMessage mqtt.MessageHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetWill(willTopic, string(willPayload), willQos, willRetained)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	if onMessage != nil {
		opts.SetDefaultPublishHandler(onMessage)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		return nil, fmt.Errorf("connection timeout")
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	return client, nil
}

// CreateClientWithKeepAlive creates a client with specified keep-alive interval
func CreateClientWithKeepAlive(cfg common.Config, clientID string, keepAlive time.Duration, onMessage mqtt.MessageHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetKeepAlive(keepAlive)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	if onMessage != nil {
		opts.SetDefaultPublishHandler(onMessage)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		return nil, fmt.Errorf("connection timeout")
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	return client, nil
}
