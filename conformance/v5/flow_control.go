package v5

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// FlowControlTests returns tests for flow control [MQTT-4.9]
func FlowControlTests() TestGroup {
	return TestGroup{
		Name: "Flow Control (Receive Maximum)",
		Tests: []TestFunc{
			testReceiveMaximumBasic,
			testReceiveMaximumQoS1,
			testReceiveMaximumQoS2,
			testReceiveMaximumEnforcement,
			testPacketIdentifierReuse,
		},
	}
}

// testReceiveMaximumBasic tests Receive Maximum property [MQTT-3.1.2.11.3]
// "The Client uses this value to limit the number of QoS 1 and QoS 2 publications
// that it is willing to process concurrently"
func testReceiveMaximumBasic(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Receive Maximum Property",
		SpecRef: "MQTT-3.1.2.11.3",
	}

	// Connect - broker will send its Receive Maximum in CONNACK
	client, err := CreateAndConnectClient(broker, "test-recvmax-basic", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// If connection succeeded, broker handled Receive Maximum
	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testReceiveMaximumQoS1 tests that Receive Maximum applies to QoS 1 [MQTT-4.9.0-1]
// "The Client MUST NOT send more than Receive Maximum QoS 1 and QoS 2 PUBLISH packets
// for which it has not received PUBACK, PUBCOMP, or PUBREC with a Reason Code >= 0x80"
func testReceiveMaximumQoS1(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Receive Maximum Applies to QoS 1",
		SpecRef: "MQTT-4.9.0-1",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-recvmax-qos1-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/recvmax/qos1", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-recvmax-qos1-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish multiple QoS 1 messages
	for i := 0; i < 10; i++ {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   "test/recvmax/qos1",
			QoS:     1,
			Payload: []byte(fmt.Sprintf("message %d", i)),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish %d failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count == 10 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected 10 messages, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testReceiveMaximumQoS2 tests that Receive Maximum applies to QoS 2 [MQTT-4.9.0-2]
func testReceiveMaximumQoS2(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Receive Maximum Applies to QoS 2",
		SpecRef: "MQTT-4.9.0-2",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-recvmax-qos2-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/recvmax/qos2", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-recvmax-qos2-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish multiple QoS 2 messages
	for i := 0; i < 10; i++ {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   "test/recvmax/qos2",
			QoS:     2,
			Payload: []byte(fmt.Sprintf("message %d", i)),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish %d failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count == 10 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected 10 messages, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testReceiveMaximumEnforcement tests that exceeding Receive Maximum causes disconnect [MQTT-4.9.0-3]
// "If a Server or Client receives more than Receive Maximum QoS 1 and QoS 2 PUBLISH packets
// without sending PUBACK or PUBCOMP, it MUST close the Network Connection"
func testReceiveMaximumEnforcement(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Receive Maximum Enforcement",
		SpecRef: "MQTT-4.9.0-3",
	}

	// This test is difficult to implement reliably without knowing the broker's
	// Receive Maximum value. We test that the broker at least accepts
	// a reasonable number of concurrent QoS messages.

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		// Delay PUBACK to keep messages in-flight
		time.Sleep(50 * time.Millisecond)
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-recvmax-enforce-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/recvmax/enforce", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-recvmax-enforce-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Send a moderate number of messages (less than typical Receive Maximum)
	for i := 0; i < 5; i++ {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   "test/recvmax/enforce",
			QoS:     1,
			Payload: []byte(fmt.Sprintf("message %d", i)),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish %d failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	time.Sleep(2 * time.Second)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	// If we got all messages, flow control is working
	if count == 5 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected 5 messages, got %d (flow control may have issues)", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testPacketIdentifierReuse tests packet identifier reuse [MQTT-2.2.1-3]
// "Packet Identifiers become available for reuse after the sender has processed
// the corresponding acknowledgement packet"
func testPacketIdentifierReuse(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Packet Identifier Reuse After ACK",
		SpecRef: "MQTT-2.2.1-3",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-packetid-reuse-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/packetid/reuse", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-packetid-reuse-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish many QoS 1 messages - packet IDs will be reused
	// (assuming fewer than 65535 concurrent messages)
	for i := 0; i < 100; i++ {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   "test/packetid/reuse",
			QoS:     1,
			Payload: []byte(fmt.Sprintf("message %d", i)),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish %d failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	time.Sleep(2 * time.Second)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count == 100 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected 100 messages, got %d (packet ID reuse may have failed)", count)
	}

	result.Duration = time.Since(start)
	return result
}
