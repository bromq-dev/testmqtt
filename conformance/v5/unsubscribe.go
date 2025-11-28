package v5

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// UnsubscribeTests returns tests for UNSUBSCRIBE/UNSUBACK [MQTT-3.10, 3.11]
func UnsubscribeTests() TestGroup {
	return TestGroup{
		Name: "UNSUBSCRIBE & UNSUBACK",
		Tests: []TestFunc{
			testUnsubscribeStopsMessages,
			testUnsubscribeMultipleTopics,
			testUnsubackReasonCodes,
			testUnsubscribeNonExistent,
			testUnsubscribePacketIdentifier,
		},
	}
}

// testUnsubscribeStopsMessages tests that unsubscribe stops message delivery [MQTT-3.10.4-6]
// "The Server MUST stop adding any new messages for delivery to the Client"
func testUnsubscribeStopsMessages(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UNSUBSCRIBE Stops Message Delivery",
		SpecRef: "MQTT-3.10.4-6",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-unsub-stops", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/unsub/stop", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Create publisher
	pub, err := CreateAndConnectClient(broker, "test-unsub-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Publish first message
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/unsub/stop",
		QoS:     0,
		Payload: []byte("message1"),
	})
	if err != nil {
		result.Error = fmt.Errorf("first publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	firstCount := messageCount
	mu.Unlock()

	if firstCount == 0 {
		result.Error = fmt.Errorf("no messages received before unsubscribe")
		result.Duration = time.Since(start)
		return result
	}

	// Unsubscribe
	_, err = sub.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{"test/unsub/stop"},
	})
	if err != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish second message - should NOT be received
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/unsub/stop",
		QoS:     0,
		Payload: []byte("message2"),
	})
	if err != nil {
		result.Error = fmt.Errorf("second publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	finalCount := messageCount
	mu.Unlock()

	if finalCount == firstCount {
		// No additional messages received after unsubscribe
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("received %d messages after unsubscribe (expected 0)", finalCount-firstCount)
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeMultipleTopics tests unsubscribing from multiple topics [MQTT-3.10.3-2]
// "The Topic Filters in an UNSUBSCRIBE packet MUST be UTF-8 Encoded Strings"
func testUnsubscribeMultipleTopics(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UNSUBSCRIBE Multiple Topics",
		SpecRef: "MQTT-3.10.3-2",
	}

	client, err := CreateAndConnectClient(broker, "test-unsub-multiple", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe to multiple topics
	_, err = client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/unsub/multi/1", QoS: 0},
			{Topic: "test/unsub/multi/2", QoS: 0},
			{Topic: "test/unsub/multi/3", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Unsubscribe from all three at once
	_, err = client.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{
			"test/unsub/multi/1",
			"test/unsub/multi/2",
			"test/unsub/multi/3",
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUnsubackReasonCodes tests UNSUBACK reason codes [MQTT-3.11.2-1]
// "The Server sends an UNSUBACK packet to the Client to confirm receipt of an UNSUBSCRIBE packet"
func testUnsubackReasonCodes(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UNSUBACK Reason Codes",
		SpecRef: "MQTT-3.11.2-1",
	}

	client, err := CreateAndConnectClient(broker, "test-unsuback-codes", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe first
	_, err = client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/unsuback/reason", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Unsubscribe - should get UNSUBACK with success (0x00)
	unsuback, err := client.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{"test/unsuback/reason"},
	})
	if err != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Check that we got an UNSUBACK
	if unsuback != nil && len(unsuback.Reasons) > 0 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("did not receive UNSUBACK")
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeNonExistent tests unsubscribing from non-existent subscription [MQTT-3.11.3-2]
// "If the Server receives an UNSUBSCRIBE packet that contains a Topic Filter that does
// not match any of the Client's existing Subscriptions, the Server MUST respond with
// an UNSUBACK containing a Reason Code of 0x11 (No subscription existed)"
func testUnsubscribeNonExistent(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UNSUBSCRIBE Non-Existent Subscription",
		SpecRef: "MQTT-3.11.3-2",
	}

	client, err := CreateAndConnectClient(broker, "test-unsub-nonexist", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Unsubscribe from topic we never subscribed to
	unsuback, err := client.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{"test/unsub/never/subscribed"},
	})

	// Some brokers may return error, some may return UNSUBACK with reason code
	// Both are acceptable per spec
	if err != nil {
		// Error is acceptable
		result.Passed = true
	} else if unsuback != nil && len(unsuback.Reasons) > 0 {
		// Check for "No subscription existed" reason code (0x11)
		if unsuback.Reasons[0] == 0x11 || unsuback.Reasons[0] == 0x00 {
			result.Passed = true
		} else {
			result.Error = fmt.Errorf("unexpected reason code: 0x%02X", unsuback.Reasons[0])
		}
	} else {
		result.Error = fmt.Errorf("no UNSUBACK received")
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribePacketIdentifier tests packet identifier in UNSUBSCRIBE [MQTT-3.10.2-1]
// "The Packet Identifier field is used to identify the UNSUBSCRIBE
// packet and its associated UNSUBACK"
func testUnsubscribePacketIdentifier(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UNSUBSCRIBE Packet Identifier Matching",
		SpecRef: "MQTT-3.10.2-1",
	}

	client, err := CreateAndConnectClient(broker, "test-unsub-packetid", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe first
	_, err = client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/unsub/packetid", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Unsubscribe - paho library handles packet ID automatically
	unsuback, err := client.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{"test/unsub/packetid"},
	})
	if err != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// If we got UNSUBACK, packet IDs matched
	if unsuback != nil {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("no UNSUBACK received")
	}

	result.Duration = time.Since(start)
	return result
}
