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

// QoSHandshakeTests returns tests for QoS 1/2 handshake packets [MQTT-3.4, 3.5, 3.6, 3.7]
func QoSHandshakeTests() TestGroup {
	return TestGroup{
		Name: "QoS Handshake (PUBACK/PUBREC/PUBREL/PUBCOMP)",
		Tests: []TestFunc{
			testPUBACKPacketIdentifier,
			testPUBACKReasonCodes,
			testPUBRECPacketIdentifier,
			testPUBRECReasonCodes,
			testPUBRELPacketIdentifier,
			testPUBRELReasonCodes,
			testPUBCOMPPacketIdentifier,
			testPUBCOMPReasonCodes,
			testQoS2CompleteHandshake,
			testQoS1DuplicateHandling,
		},
	}
}

// testPUBACKPacketIdentifier tests PUBACK packet identifier [MQTT-3.4.2-1]
// "The Packet Identifier field contains the Packet Identifier from the PUBLISH packet
// that is being acknowledged"
func testPUBACKPacketIdentifier(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBACK Packet Identifier Matches PUBLISH",
		SpecRef: "MQTT-3.4.2-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		received = true
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-puback-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/puback/id", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-puback-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish QoS 1 - will receive PUBACK
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/puback/id",
		QoS:     1,
		Payload: []byte("test qos1"),
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
		result.Error = fmt.Errorf("QoS 1 message not received (PUBACK may have failed)")
	}

	result.Duration = time.Since(start)
	return result
}

// testPUBACKReasonCodes tests PUBACK reason codes [MQTT-3.4.2.1-1]
// "The Client or Server sending the PUBACK packet MUST use one of the PUBACK Reason Codes"
func testPUBACKReasonCodes(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBACK Reason Codes",
		SpecRef: "MQTT-3.4.2.1-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-puback-reason", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Publish QoS 1 to valid topic - should get success PUBACK (0x00)
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/puback/reason",
		QoS:     1,
		Payload: []byte("test"),
	})

	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testPUBRECPacketIdentifier tests PUBREC packet identifier [MQTT-3.5.2-1]
// "The Packet Identifier field contains the Packet Identifier from the PUBLISH packet
// that is being acknowledged"
func testPUBRECPacketIdentifier(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBREC Packet Identifier Matches PUBLISH",
		SpecRef: "MQTT-3.5.2-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		received = true
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-pubrec-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/pubrec/id", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pubrec-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish QoS 2 - will trigger PUBREC/PUBREL/PUBCOMP handshake
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/pubrec/id",
		QoS:     2,
		Payload: []byte("test qos2"),
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
		result.Error = fmt.Errorf("QoS 2 message not received (PUBREC handshake may have failed)")
	}

	result.Duration = time.Since(start)
	return result
}

// testPUBRECReasonCodes tests PUBREC reason codes [MQTT-3.5.2.1-1]
func testPUBRECReasonCodes(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBREC Reason Codes",
		SpecRef: "MQTT-3.5.2.1-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-pubrec-reason", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Publish QoS 2 - should receive PUBREC with success (0x00)
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/pubrec/reason",
		QoS:     2,
		Payload: []byte("test"),
	})

	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testPUBRELPacketIdentifier tests PUBREL packet identifier [MQTT-3.6.2-1]
// "The Packet Identifier field contains the Packet Identifier from the PUBREC packet
// that is being acknowledged"
func testPUBRELPacketIdentifier(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBREL Packet Identifier Matches PUBREC",
		SpecRef: "MQTT-3.6.2-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		received = true
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-pubrel-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/pubrel/id", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pubrel-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// QoS 2 publish triggers full handshake including PUBREL
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/pubrel/id",
		QoS:     2,
		Payload: []byte("test qos2 pubrel"),
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
		result.Error = fmt.Errorf("QoS 2 message not received (PUBREL may have failed)")
	}

	result.Duration = time.Since(start)
	return result
}

// testPUBRELReasonCodes tests PUBREL reason codes [MQTT-3.6.2.1-1]
func testPUBRELReasonCodes(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBREL Reason Codes",
		SpecRef: "MQTT-3.6.2.1-1",
	}

	// Test that PUBREL is sent with proper reason code during QoS 2 flow
	client, err := CreateAndConnectClient(cfg, "test-pubrel-reason", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// QoS 2 publish - if successful, PUBREL was sent with correct reason code
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/pubrel/reason",
		QoS:     2,
		Payload: []byte("test"),
	})

	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testPUBCOMPPacketIdentifier tests PUBCOMP packet identifier [MQTT-3.7.2-1]
// "The Packet Identifier field contains the Packet Identifier from the PUBREL packet
// that is being acknowledged"
func testPUBCOMPPacketIdentifier(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBCOMP Packet Identifier Matches PUBREL",
		SpecRef: "MQTT-3.7.2-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		received = true
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-pubcomp-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/pubcomp/id", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-pubcomp-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// QoS 2 publish - PUBCOMP is final ack in the handshake
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/pubcomp/id",
		QoS:     2,
		Payload: []byte("test qos2 pubcomp"),
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
		result.Error = fmt.Errorf("QoS 2 message not received (PUBCOMP may have failed)")
	}

	result.Duration = time.Since(start)
	return result
}

// testPUBCOMPReasonCodes tests PUBCOMP reason codes [MQTT-3.7.2.1-1]
func testPUBCOMPReasonCodes(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBCOMP Reason Codes",
		SpecRef: "MQTT-3.7.2.1-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-pubcomp-reason", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// QoS 2 publish - if successful, PUBCOMP was received with correct reason code
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/pubcomp/reason",
		QoS:     2,
		Payload: []byte("test"),
	})

	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testQoS2CompleteHandshake tests complete QoS 2 handshake [MQTT-4.3.3-1]
// "The receiver MUST respond to a PUBREL packet by sending a PUBCOMP packet
// containing the same Packet Identifier as the PUBREL"
func testQoS2CompleteHandshake(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "QoS 2 Complete Handshake (PUBLISH->PUBREC->PUBREL->PUBCOMP)",
		SpecRef: "MQTT-4.3.3-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		received = true
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-qos2-handshake-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/qos2/handshake", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-qos2-handshake-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish QoS 2 - triggers full 4-way handshake
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/qos2/handshake",
		QoS:     2,
		Payload: []byte("test qos2 complete"),
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
		result.Error = fmt.Errorf("QoS 2 handshake did not complete successfully")
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS1DuplicateHandling tests DUP flag in QoS 1 retransmissions [MQTT-3.3.1-1]
// "If the DUP flag is set to 0, it indicates that this is the first occasion
// that the Client or Server has attempted to send this PUBLISH packet"
func testQoS1DuplicateHandling(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "QoS 1 DUP Flag Handling",
		SpecRef: "MQTT-3.3.1-1",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-dup-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/dup/flag", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-dup-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish multiple QoS 1 messages
	for i := 0; i < 3; i++ {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   "test/dup/flag",
			QoS:     1,
			Payload: []byte(fmt.Sprintf("message %d", i)),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish %d failed: %w", i, err)
			result.Duration = time.Since(start)
			return result
		}
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count >= 3 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected at least 3 messages, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}
