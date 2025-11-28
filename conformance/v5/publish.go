package v5

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// PublishSubscribeTests returns all publish/subscribe conformance tests
func PublishSubscribeTests() TestGroup {
	return TestGroup{
		Name: "Publish/Subscribe",
		Tests: []TestFunc{
			testBasicPubSub,
			testMultipleSubscribers,
			testRetainedMessage,
			testEmptyPayload,
			testUnsubscribe,
		},
	}
}

// testBasicPubSub tests basic publish/subscribe [MQTT-3.3.1-1]
func testBasicPubSub(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Basic Publish/Subscribe",
		SpecRef: "MQTT-3.3.1-1",
	}

	// Set up message handler
	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		received = true
		mu.Unlock()
		return true, nil
	}

	// Create subscriber
	sub, err := CreateAndConnectClient(broker, "test-sub-basic", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Subscribe
	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/basic", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Create publisher
	pub, err := CreateAndConnectClient(broker, "test-pub-basic", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Give subscriber time to be ready
	time.Sleep(100 * time.Millisecond)

	// Publish
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/basic",
		QoS:     0,
		Payload: []byte("test message"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Wait for message
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	result.Passed = received
	mu.Unlock()

	if !result.Passed {
		result.Error = fmt.Errorf("message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testMultipleSubscribers tests multiple subscribers receiving messages
func testMultipleSubscribers(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Multiple Subscribers",
		SpecRef: "MQTT-3.3.1-1",
	}

	// Create multiple subscribers
	numSubs := 3
	received := make([]bool, numSubs)
	var mu sync.Mutex
	var clients []*paho.Client

	for i := 0; i < numSubs; i++ {
		idx := i
		onPublish := func(pr paho.PublishReceived) (bool, error) {
			mu.Lock()
			received[idx] = true
			mu.Unlock()
			return true, nil
		}

		sub, err := CreateAndConnectClient(broker, fmt.Sprintf("test-sub-multi-%d", i), onPublish)
		if err != nil {
			result.Error = fmt.Errorf("subscriber %d connect failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
		clients = append(clients, sub)

		ctx := context.Background()
		_, err = sub.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: "test/multi", QoS: 0},
			},
		})
		if err != nil {
			result.Error = fmt.Errorf("subscriber %d subscribe failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Defer cleanup
	for _, client := range clients {
		defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})
	}

	// Create publisher
	pub, err := CreateAndConnectClient(broker, "test-pub-multi", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Give subscribers time to be ready
	time.Sleep(100 * time.Millisecond)

	// Publish message
	ctx := context.Background()
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/multi",
		QoS:     0,
		Payload: []byte("broadcast message"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Wait for messages
	time.Sleep(500 * time.Millisecond)

	// Check all received
	mu.Lock()
	allReceived := true
	for i, r := range received {
		if !r {
			allReceived = false
			result.Error = fmt.Errorf("subscriber %d did not receive message", i)
			break
		}
	}
	mu.Unlock()

	result.Passed = allReceived
	result.Duration = time.Since(start)
	return result
}

// testRetainedMessage tests retained message functionality [MQTT-3.3.1-5]
// "When a new Non‑shared Subscription is made, the last retained message, if any,
// on each matching topic name is sent to the Client"
func testRetainedMessage(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Retained Message",
		SpecRef: "MQTT-3.3.1-5",
	}

	topic := fmt.Sprintf("test/retained/%d", time.Now().UnixNano())

	// Publish a retained message
	pub, err := CreateAndConnectClient(broker, "test-pub-retained", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	ctx := context.Background()
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   topic,
		QoS:     0,
		Retain:  true,
		Payload: []byte("retained message"),
	})
	if err != nil {
		pub.Disconnect(&paho.Disconnect{ReasonCode: 0})
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Wait a moment for the message to be retained
	time.Sleep(100 * time.Millisecond)

	// Subscribe with a new client - should receive the retained message
	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if string(pr.Packet.Payload) == "retained message" {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-sub-retained", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: topic, QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Wait for retained message
	time.Sleep(500 * time.Millisecond)

	// Clear the retained message
	pub2, _ := CreateAndConnectClient(broker, "test-pub-clear", nil)
	if pub2 != nil {
		pub2.Publish(ctx, &paho.Publish{Topic: topic, QoS: 0, Retain: true, Payload: []byte{}})
		pub2.Disconnect(&paho.Disconnect{ReasonCode: 0})
	}

	mu.Lock()
	result.Passed = received
	mu.Unlock()

	if !result.Passed {
		result.Error = fmt.Errorf("retained message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testEmptyPayload tests publishing with empty payload
// "A zero-byte Payload is valid"
func testEmptyPayload(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Empty Payload",
		SpecRef: "MQTT-3.3.3-1",
	}

	received := false
	receivedEmpty := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		received = true
		if len(pr.Packet.Payload) == 0 {
			receivedEmpty = true
		}
		mu.Unlock()
		return true, nil
	}

	// Create subscriber
	sub, err := CreateAndConnectClient(broker, "test-sub-empty", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/empty", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Create publisher
	pub, err := CreateAndConnectClient(broker, "test-pub-empty", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish with empty payload
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/empty",
		QoS:     0,
		Payload: []byte{},
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	result.Passed = received && receivedEmpty
	mu.Unlock()

	if !result.Passed {
		if !received {
			result.Error = fmt.Errorf("message not received")
		} else {
			result.Error = fmt.Errorf("payload was not empty")
		}
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribe tests unsubscribe functionality [MQTT-3.10.4-1]
// "The Server MUST stop adding any new messages which match the Topic Filters,
// for delivery to that Client"
func testUnsubscribe(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Unsubscribe",
		SpecRef: "MQTT-3.10.4-1",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	// Create subscriber
	sub, err := CreateAndConnectClient(broker, "test-sub-unsub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/unsub", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Create publisher
	pub, err := CreateAndConnectClient(broker, "test-pub-unsub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish first message
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/unsub",
		QoS:     0,
		Payload: []byte("message 1"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish 1 failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(200 * time.Millisecond)

	// Unsubscribe
	_, err = sub.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{"test/unsub"},
	})
	if err != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish second message - should NOT be received
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/unsub",
		QoS:     0,
		Payload: []byte("message 2"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish 2 failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	// Should have received exactly 1 message (before unsubscribe)
	result.Passed = (count == 1)
	if !result.Passed {
		result.Error = fmt.Errorf("expected 1 message, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}
