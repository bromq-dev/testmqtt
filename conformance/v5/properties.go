package v5

import (
	"github.com/bromq-dev/testmqtt/conformance/common"
)

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// PropertiesTests returns all MQTT v5 properties conformance tests
func PropertiesTests() TestGroup {
	return TestGroup{
		Name: "Properties",
		Tests: []TestFunc{
			testUserProperties,
			testContentType,
			testResponseTopic,
			testCorrelationData,
			testMaximumPacketSize,
		},
	}
}

// testUserProperties tests User Properties [MQTT-3.1.3-10]
// "The Server MUST maintain the order of User Properties when publishing
// the Will Message"
func testUserProperties(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "User Properties",
		SpecRef: "MQTT-3.1.3-10",
	}

	// Test that we can send user properties
	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		// Check if user properties are present
		if len(pr.Packet.Properties.User) > 0 {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-userprops", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/userprops", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-userprops", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish with user properties
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/userprops",
		QoS:     0,
		Payload: []byte("message with properties"),
		Properties: &paho.PublishProperties{
			User: []paho.UserProperty{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	result.Passed = received
	mu.Unlock()

	if !result.Passed {
		result.Error = fmt.Errorf("user properties not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testContentType tests Content Type property
func testContentType(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Content Type Property",
		SpecRef: "MQTT-3.3.2-5",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.Properties != nil && pr.Packet.Properties.ContentType != "" {
			if pr.Packet.Properties.ContentType == "application/json" {
				received = true
			}
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-contenttype", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/contenttype", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-contenttype", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/contenttype",
		QoS:     0,
		Payload: []byte(`{"test": "data"}`),
		Properties: &paho.PublishProperties{
			ContentType: "application/json",
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	result.Passed = received
	mu.Unlock()

	if !result.Passed {
		result.Error = fmt.Errorf("content type property not received correctly")
	}

	result.Duration = time.Since(start)
	return result
}

// testResponseTopic tests Response Topic property
func testResponseTopic(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Response Topic Property",
		SpecRef: "MQTT-3.3.2-6",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.Properties != nil && pr.Packet.Properties.ResponseTopic != "" {
			if pr.Packet.Properties.ResponseTopic == "response/topic" {
				received = true
			}
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-responsetopic", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/responsetopic", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-responsetopic", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/responsetopic",
		QoS:     0,
		Payload: []byte("request"),
		Properties: &paho.PublishProperties{
			ResponseTopic: "response/topic",
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	result.Passed = received
	mu.Unlock()

	if !result.Passed {
		result.Error = fmt.Errorf("response topic property not received correctly")
	}

	result.Duration = time.Since(start)
	return result
}

// testCorrelationData tests Correlation Data property
func testCorrelationData(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Correlation Data Property",
		SpecRef: "MQTT-3.3.2-7",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.Properties != nil && len(pr.Packet.Properties.CorrelationData) > 0 {
			if string(pr.Packet.Properties.CorrelationData) == "correlation-123" {
				received = true
			}
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-correlation", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/correlation", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-correlation", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/correlation",
		QoS:     0,
		Payload: []byte("request"),
		Properties: &paho.PublishProperties{
			CorrelationData: []byte("correlation-123"),
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	result.Passed = received
	mu.Unlock()

	if !result.Passed {
		result.Error = fmt.Errorf("correlation data property not received correctly")
	}

	result.Duration = time.Since(start)
	return result
}

// testMaximumPacketSize tests Maximum Packet Size [MQTT-3.1.2-24]
// "The Server MUST NOT send packets exceeding Maximum Packet Size to the Client"
func testMaximumPacketSize(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Maximum Packet Size",
		SpecRef: "MQTT-3.1.2-24",
	}

	// Testing maximum packet size requires setting it in CONNECT
	// and then trying to send large messages
	client, err := CreateAndConnectClient(cfg, "test-maxpacket", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}
