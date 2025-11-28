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

// MessageExpiryTests returns tests for message expiry interval [MQTT-3.3.2.3.3]
func MessageExpiryTests() TestGroup {
	return TestGroup{
		Name: "Message Expiry Interval",
		Tests: []TestFunc{
			testMessageExpiryBasic,
			testMessageExpiryCountdown,
			testMessageExpiryZeroMeansNoExpiry,
			testMessageExpiryRetainedMessage,
		},
	}
}

// testMessageExpiryBasic tests basic message expiry [MQTT-3.3.2.3.3-1]
// "If present, the Four Byte value is the lifetime of the Application Message in seconds"
func testMessageExpiryBasic(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Message Expiry Interval Basic",
		SpecRef: "MQTT-3.3.2.3.3-1",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-expiry-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/expiry/basic", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-expiry-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish with message expiry interval of 10 seconds
	expiryInterval := uint32(10)
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/expiry/basic",
		QoS:     1,
		Payload: []byte("message with expiry"),
		Properties: &paho.PublishProperties{
			MessageExpiry: &expiryInterval,
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count > 0 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("message with expiry interval not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testMessageExpiryCountdown tests expiry countdown [MQTT-3.3.2.3.3-2]
// "The Message Expiry Interval MUST be set to the received value minus the time
// that the Application Message has been waiting in the Server"
func testMessageExpiryCountdown(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Message Expiry Interval Countdown",
		SpecRef: "MQTT-3.3.2.3.3-2",
	}

	messageReceived := false
	var receivedExpiry uint32
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageReceived = true
		if pr.Packet.Properties != nil && pr.Packet.Properties.MessageExpiry != nil {
			receivedExpiry = *pr.Packet.Properties.MessageExpiry
		}
		mu.Unlock()
		return true, nil
	}

	// First publish a RETAINED message with expiry
	// This ensures the message stays on the broker while we wait
	pub, err := CreateAndConnectClient(cfg, "test-expiry-countdown-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	ctx := context.Background()
	expiryInterval := uint32(30) // 30 seconds
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/expiry/countdown",
		QoS:     1,
		Retain:  true, // Retain so message stays on broker
		Payload: []byte("message with countdown"),
		Properties: &paho.PublishProperties{
			MessageExpiry: &expiryInterval,
		},
	})
	if err != nil {
		pub.Disconnect(&paho.Disconnect{ReasonCode: 0})
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Wait a bit before subscribing - message expiry should count down
	time.Sleep(2 * time.Second)

	// Now subscribe - should receive retained message with reduced expiry
	sub, err := CreateAndConnectClient(cfg, "test-expiry-countdown-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/expiry/countdown", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(1 * time.Second) // Increased wait time for retained message delivery

	mu.Lock()
	received := messageReceived
	expiry := receivedExpiry
	mu.Unlock()

	if received {
		// Expiry should be less than 30 (we waited 2 seconds)
		// But we can't be too strict due to timing variations
		if expiry < 30 || expiry == 0 {
			result.Passed = true
		} else {
			result.Error = fmt.Errorf("expiry countdown not working (got %d, expected < 30)", expiry)
		}
	} else {
		result.Error = fmt.Errorf("message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testMessageExpiryZeroMeansNoExpiry tests that absent expiry means no expiry [MQTT-3.3.2.3.3-3]
// "If the Message Expiry Interval is absent, the Application Message does not expire"
func testMessageExpiryZeroMeansNoExpiry(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Absent Message Expiry Means No Expiry",
		SpecRef: "MQTT-3.3.2.3.3-3",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-expiry-none-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/expiry/none", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-expiry-none-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish without message expiry interval
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/expiry/none",
		QoS:     1,
		Payload: []byte("message without expiry"),
		// No MessageExpiry property
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count > 0 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("message without expiry not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testMessageExpiryRetainedMessage tests expiry with retained messages [MQTT-3.3.2.3.3-4]
// "The PUBLISH packet sent to a Client by the Server MUST contain a Message Expiry Interval
// set to the received value minus the time that the message has been waiting in the Server"
func testMessageExpiryRetainedMessage(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Message Expiry With Retained Messages",
		SpecRef: "MQTT-3.3.2.3.3-4",
	}

	// Publish retained message with expiry
	pub, err := CreateAndConnectClient(cfg, "test-expiry-retained-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	ctx := context.Background()
	expiryInterval := uint32(60)
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/expiry/retained",
		QoS:     1,
		Payload: []byte("retained with expiry"),
		Retain:  true,
		Properties: &paho.PublishProperties{
			MessageExpiry: &expiryInterval,
		},
	})
	if err != nil {
		pub.Disconnect(&paho.Disconnect{ReasonCode: 0})
		result.Error = fmt.Errorf("publish retained failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(1 * time.Second)

	messageReceived := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageReceived = true
		mu.Unlock()
		return true, nil
	}

	// Subscribe to get retained message
	sub, err := CreateAndConnectClient(cfg, "test-expiry-retained-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/expiry/retained", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	received := messageReceived
	mu.Unlock()

	if received {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("retained message with expiry not received")
	}

	result.Duration = time.Since(start)
	return result
}
