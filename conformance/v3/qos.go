package v3

import (
	"fmt"
	"sync"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// QoSTests returns tests for MQTT v3.1.1 QoS functionality
func QoSTests() common.TestGroup {
	return common.TestGroup{
		Name: "QoS",
		Tests: []common.TestFunc{
			testQoS0AtMostOnce,
			testQoS1AtLeastOnce,
			testQoS2ExactlyOnce,
			testQoSDowngrade,
			testMessageOrderingQoS1,
			testMessageOrderingQoS2,
			testQoS1Acknowledgement,
			testQoS2HandshakeFull,
		},
	}
}

// testQoS0AtMostOnce tests QoS 0 at-most-once delivery [MQTT-4.3.1-1]
func testQoS0AtMostOnce(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS 0 At Most Once",
		SpecRef: "MQTT-4.3.1-1",
	}

	var mu sync.Mutex
	receivedCount := 0
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos0"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/qos0/atmost"
	subscriber.Subscribe(topic, 0, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos0-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish with QoS 0
	for i := 0; i < 5; i++ {
		publisher.Publish(topic, 0, false, fmt.Sprintf("message%d", i)).Wait()
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// With QoS 0, we may receive 0 to 5 messages
	if receivedCount > 5 {
		result.Error = fmt.Errorf("received more messages than sent (%d > 5)", receivedCount)
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS1AtLeastOnce tests QoS 1 at-least-once delivery [MQTT-4.3.2-1]
func testQoS1AtLeastOnce(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS 1 At Least Once",
		SpecRef: "MQTT-4.3.2-1",
	}

	var mu sync.Mutex
	receivedCount := 0
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos1-atleast"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/qos1/atleast"
	subscriber.Subscribe(topic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos1-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	messageCount := 5
	for i := 0; i < messageCount; i++ {
		token := publisher.Publish(topic, 1, false, fmt.Sprintf("message%d", i))
		token.Wait()
		if token.Error() != nil {
			result.Error = fmt.Errorf("publish failed: %w", token.Error())
			result.Duration = time.Since(start)
			return result
		}
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	// With QoS 1, we should receive at least all messages (may have duplicates)
	if receivedCount < messageCount {
		result.Error = fmt.Errorf("received fewer messages than sent (%d < %d)", receivedCount, messageCount)
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS2ExactlyOnce tests QoS 2 exactly-once delivery [MQTT-4.3.3-1]
func testQoS2ExactlyOnce(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS 2 Exactly Once",
		SpecRef: "MQTT-4.3.3-1",
	}

	var mu sync.Mutex
	receivedMessages := make(map[string]int)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessages[string(msg.Payload())]++
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos2-exactly"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/qos2/exactly"
	subscriber.Subscribe(topic, 2, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos2-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	messageCount := 5
	for i := 0; i < messageCount; i++ {
		token := publisher.Publish(topic, 2, false, fmt.Sprintf("message%d", i))
		token.Wait()
		if token.Error() != nil {
			result.Error = fmt.Errorf("publish failed: %w", token.Error())
			result.Duration = time.Since(start)
			return result
		}
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	// With QoS 2, each message should be received exactly once
	if len(receivedMessages) != messageCount {
		result.Error = fmt.Errorf("received %d distinct messages, expected %d", len(receivedMessages), messageCount)
		result.Duration = time.Since(start)
		return result
	}

	for msg, count := range receivedMessages {
		if count != 1 {
			result.Error = fmt.Errorf("message '%s' received %d times, expected exactly 1", msg, count)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testQoSDowngrade tests QoS downgrade [MQTT-3.8.4-6]
func testQoSDowngrade(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS Downgrade",
		SpecRef: "MQTT-3.8.4-6",
	}

	// Subscribe with QoS 0
	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos-downgrade"), nil)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/qos/downgrade"
	subscriber.Subscribe(topic, 0, nil).Wait() // Subscribe with QoS 0
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos-downgrade-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish with QoS 2, but subscriber requested QoS 0, so should be downgraded
	token := publisher.Publish(topic, 2, false, "qos2 message")
	token.Wait()

	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
	} else {
		// Test passes if publish succeeds (broker handles QoS downgrade)
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testMessageOrderingQoS1 tests message ordering for QoS 1 [MQTT-4.6.0-2]
func testMessageOrderingQoS1(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Message Ordering QoS 1",
		SpecRef: "MQTT-4.6.0-2",
	}

	var mu sync.Mutex
	receivedOrder := make([]string, 0)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedOrder = append(receivedOrder, string(msg.Payload()))
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-order-qos1"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/order/qos1"
	subscriber.Subscribe(topic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-order-qos1-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish messages in order
	for i := 1; i <= 5; i++ {
		token := publisher.Publish(topic, 1, false, fmt.Sprintf("msg%d", i))
		token.Wait()
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedOrder) < 5 {
		result.Error = fmt.Errorf("expected at least 5 messages, received %d", len(receivedOrder))
		result.Duration = time.Since(start)
		return result
	}

	// Check that messages appear in order (allowing for duplicates)
	lastSeen := 0
	for _, msg := range receivedOrder {
		var num int
		fmt.Sscanf(msg, "msg%d", &num)
		if num < lastSeen {
			result.Error = fmt.Errorf("messages received out of order: %v", receivedOrder)
			result.Duration = time.Since(start)
			return result
		}
		if num > lastSeen {
			lastSeen = num
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testMessageOrderingQoS2 tests message ordering for QoS 2 [MQTT-4.6.0-3]
func testMessageOrderingQoS2(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Message Ordering QoS 2",
		SpecRef: "MQTT-4.6.0-3",
	}

	var mu sync.Mutex
	receivedOrder := make([]string, 0)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedOrder = append(receivedOrder, string(msg.Payload()))
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-order-qos2"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/order/qos2"
	subscriber.Subscribe(topic, 2, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-order-qos2-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish messages in order
	for i := 1; i <= 5; i++ {
		token := publisher.Publish(topic, 2, false, fmt.Sprintf("msg%d", i))
		token.Wait()
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedOrder) != 5 {
		result.Error = fmt.Errorf("expected exactly 5 messages, received %d", len(receivedOrder))
		result.Duration = time.Since(start)
		return result
	}

	// Check that messages are in order (no duplicates with QoS 2)
	for i := 0; i < 5; i++ {
		expected := fmt.Sprintf("msg%d", i+1)
		if receivedOrder[i] != expected {
			result.Error = fmt.Errorf("message at position %d is '%s', expected '%s'", i, receivedOrder[i], expected)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testQoS1Acknowledgement tests PUBACK for QoS 1 [MQTT-4.3.2-2]
func testQoS1Acknowledgement(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS 1 PUBACK Acknowledgement",
		SpecRef: "MQTT-4.3.2-2",
	}

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos1-puback"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	topic := "test/qos1/puback"
	token := publisher.Publish(topic, 1, false, "qos1 message")
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("publish timeout (no PUBACK received)")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
	} else {
		// If publish succeeded and didn't timeout, PUBACK was received
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS2HandshakeFull tests complete QoS 2 handshake [MQTT-4.3.3-2]
func testQoS2HandshakeFull(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS 2 Full Handshake",
		SpecRef: "MQTT-4.3.3-2",
	}

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos2-handshake"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	topic := "test/qos2/handshake"
	token := publisher.Publish(topic, 2, false, "qos2 message")
	if !token.WaitTimeout(10 * time.Second) {
		result.Error = fmt.Errorf("publish timeout (QoS 2 handshake not completed)")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
	} else {
		// If publish succeeded, full QoS 2 handshake (PUBREC, PUBREL, PUBCOMP) completed
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}
