package v5

import (
	"context"
	"fmt"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	"github.com/eclipse/paho.golang/paho"
)

// CreateAndConnectClient creates and connects a MQTT v5 client with optional message handler
func CreateAndConnectClient(cfg common.Config, clientID string, onPublish func(paho.PublishReceived) (bool, error)) (*paho.Client, error) {
	conn, err := common.DialBroker(cfg.Broker)
	if err != nil {
		return nil, err
	}

	config := paho.ClientConfig{
		ClientID: clientID,
		Conn:     conn,
	}

	if onPublish != nil {
		config.OnPublishReceived = []func(paho.PublishReceived) (bool, error){onPublish}
	}

	client := paho.NewClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   clientID,
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

	_, err = client.Connect(ctx, cp)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// CreateAndConnectClientWithSession creates and connects a MQTT v5 client with session control
func CreateAndConnectClientWithSession(cfg common.Config, clientID string, cleanStart bool, onPublish func(paho.PublishReceived) (bool, error)) (*paho.Client, error) {
	conn, err := common.DialBroker(cfg.Broker)
	if err != nil {
		return nil, err
	}

	config := paho.ClientConfig{
		ClientID: clientID,
		Conn:     conn,
	}

	if onPublish != nil {
		config.OnPublishReceived = []func(paho.PublishReceived) (bool, error){onPublish}
	}

	client := paho.NewClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set session expiry interval to 300 seconds if not using clean start
	sessionExpiry := uint32(300)
	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   clientID,
		CleanStart: cleanStart,
	}

	if !cleanStart {
		cp.Properties = &paho.ConnectProperties{
			SessionExpiryInterval: &sessionExpiry,
		}
	}

	if cfg.Username != "" {
		cp.UsernameFlag = true
		cp.Username = cfg.Username
	}
	if cfg.Password != "" {
		cp.PasswordFlag = true
		cp.Password = []byte(cfg.Password)
	}

	_, err = client.Connect(ctx, cp)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}
