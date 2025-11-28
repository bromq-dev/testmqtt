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

// QoSTests returns all QoS-related conformance tests
func QoSTests() TestGroup {
	return TestGroup{
		Name: "QoS",
		Tests: []TestFunc{
			testQoS0,
			testQoS1,
			testQoS2,
			testQoS1Duplicate,
			testQoS2ExactlyOnce,
			testPacketIdentifier,
		},
	}
}

// testQoS0 tests QoS 0 message delivery [MQTT-4.3.1-1]
// "The receiver does not respond to the message and does not make any attempt
// at re-delivery"
func testQoS0(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "QoS 0 Delivery",
		SpecRef: "MQTT-4.3.1-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.QoS == 0 {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-qos0", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/qos0", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-qos0", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/qos0",
		QoS:     0,
		Payload: []byte("qos 0 message"),
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
		result.Error = fmt.Errorf("QoS 0 message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS1 tests QoS 1 message delivery [MQTT-4.3.2-1]
// "The receiver sends a PUBACK packet in response to a PUBLISH packet"
func testQoS1(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "QoS 1 Delivery",
		SpecRef: "MQTT-4.3.2-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.QoS == 1 {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-qos1", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/qos1", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-qos1", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/qos1",
		QoS:     1,
		Payload: []byte("qos 1 message"),
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
		result.Error = fmt.Errorf("QoS 1 message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS2 tests QoS 2 message delivery [MQTT-4.3.3-1]
// "This is the highest QoS level, for use when neither loss nor duplication
// of messages are acceptable"
func testQoS2(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "QoS 2 Delivery",
		SpecRef: "MQTT-4.3.3-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.QoS == 2 {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-qos2", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/qos2", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-qos2", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/qos2",
		QoS:     2,
		Payload: []byte("qos 2 message"),
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
		result.Error = fmt.Errorf("QoS 2 message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS1Duplicate tests QoS 1 duplicate handling
// Tests that messages can be delivered with QoS 1
func testQoS1Duplicate(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "QoS 1 At Least Once",
		SpecRef: "MQTT-4.3.2-2",
	}

	receivedCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-qos1-dup", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/qos1/dup", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-qos1-dup", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/qos1/dup",
		QoS:     1,
		Payload: []byte("qos 1 at-least-once"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := receivedCount
	mu.Unlock()

	// Should receive at least one message
	result.Passed = (count >= 1)
	if !result.Passed {
		result.Error = fmt.Errorf("expected at least 1 message, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS2ExactlyOnce tests QoS 2 exactly-once delivery
// Verifies that QoS 2 messages are delivered exactly once
func testQoS2ExactlyOnce(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "QoS 2 Exactly-Once",
		SpecRef: "MQTT-4.3.3-2",
	}

	receivedCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-sub-qos2-once", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/qos2/once", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pub-qos2-once", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/qos2/once",
		QoS:     2,
		Payload: []byte("qos 2 exactly-once"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	count := receivedCount
	mu.Unlock()

	// Should receive exactly one message
	result.Passed = (count == 1)
	if !result.Passed {
		result.Error = fmt.Errorf("expected exactly 1 message, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testPacketIdentifier tests packet identifier requirements [MQTT-2.2.1-3]
// "Each time a Client sends a new SUBSCRIBE, UNSUBSCRIBE, or PUBLISH (where QoS > 0)
// MQTT Control Packet it MUST assign it a non-zero Packet Identifier that is
// currently unused"
func testPacketIdentifier(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Packet Identifier Assignment",
		SpecRef: "MQTT-2.2.1-3",
	}

	// The paho client library handles packet identifiers automatically
	// We test that multiple QoS > 0 publishes work correctly
	pub, err := CreateAndConnectClient(cfg, "test-pub-pktid", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Publish multiple QoS 1 messages
	for i := 0; i < 5; i++ {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   fmt.Sprintf("test/pktid/%d", i),
			QoS:     1,
			Payload: []byte(fmt.Sprintf("message %d", i)),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish %d failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}
