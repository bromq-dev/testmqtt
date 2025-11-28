package v3

import (
	"fmt"
	"sync"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// PublishSubscribeTests returns tests for MQTT v3.1.1 publish/subscribe functionality
func PublishSubscribeTests() common.TestGroup {
	return common.TestGroup{
		Name: "Publish/Subscribe",
		Tests: []common.TestFunc{
			testBasicPublishSubscribe,
			testPublishQoS0,
			testPublishQoS1,
			testPublishQoS2,
			testSubscribeAcknowledgement,
			testMultipleSubscriptions,
			testSubscriptionReplacement,
			testRetainedMessage,
			testRetainedMessageClear,
			testPublishToMultipleSubscribers,
		},
	}
}

// testBasicPublishSubscribe tests basic publish and subscribe [MQTT-3.3.1-1]
func testBasicPublishSubscribe(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Basic Publish/Subscribe",
		SpecRef: "MQTT-3.3.1-1",
	}

	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/basic/pubsub"
	token := subscriber.Subscribe(topic, 0, nil)
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("subscribe timeout")
		result.Duration = time.Since(start)
		return result
	}
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	token = publisher.Publish(topic, 0, false, "test message")
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("publish timeout")
		result.Duration = time.Since(start)
		return result
	}
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !receivedMessage {
		result.Error = fmt.Errorf("message not received")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testPublishQoS0 tests QoS 0 publish [MQTT-4.3.1-1]
func testPublishQoS0(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Publish QoS 0",
		SpecRef: "MQTT-4.3.1-1",
	}

	var mu sync.Mutex
	var receivedCount int
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos0-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/qos0"
	token := subscriber.Subscribe(topic, 0, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos0-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	token = publisher.Publish(topic, 0, false, "QoS 0 message")
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedCount == 0 {
		result.Error = fmt.Errorf("message not received")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testPublishQoS1 tests QoS 1 publish [MQTT-4.3.2-1]
func testPublishQoS1(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Publish QoS 1",
		SpecRef: "MQTT-4.3.2-1",
	}

	var mu sync.Mutex
	var receivedCount int
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos1-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/qos1"
	token := subscriber.Subscribe(topic, 1, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos1-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	token = publisher.Publish(topic, 1, false, "QoS 1 message")
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedCount == 0 {
		result.Error = fmt.Errorf("message not received")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testPublishQoS2 tests QoS 2 publish [MQTT-4.3.3-1]
func testPublishQoS2(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Publish QoS 2",
		SpecRef: "MQTT-4.3.3-1",
	}

	var mu sync.Mutex
	var receivedCount int
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos2-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/qos2"
	token := subscriber.Subscribe(topic, 2, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-qos2-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	token = publisher.Publish(topic, 2, false, "QoS 2 message")
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedCount == 0 {
		result.Error = fmt.Errorf("message not received")
	} else if receivedCount > 1 {
		result.Error = fmt.Errorf("message received %d times (expected exactly once)", receivedCount)
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscribeAcknowledgement tests SUBSCRIBE acknowledgement [MQTT-3.8.4-1]
func testSubscribeAcknowledgement(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Subscribe Acknowledgement",
		SpecRef: "MQTT-3.8.4-1",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-suback"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	topic := "test/suback"
	token := client.Subscribe(topic, 1, nil)
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("subscribe timeout")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testMultipleSubscriptions tests multiple subscriptions [MQTT-3.8.4-4]
func testMultipleSubscriptions(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Multiple Subscriptions",
		SpecRef: "MQTT-3.8.4-4",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-multi-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// Subscribe to multiple topics
	topics := map[string]byte{
		"test/multi/topic1": 0,
		"test/multi/topic2": 1,
	}

	token := subscriber.SubscribeMultiple(topics, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-multi-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish("test/multi/topic1", 0, false, "message1").Wait()
	publisher.Publish("test/multi/topic2", 1, false, "message2").Wait()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) != 2 {
		result.Error = fmt.Errorf("expected messages on 2 topics, got %d", len(receivedTopics))
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscriptionReplacement tests subscription replacement [MQTT-3.8.4-3]
func testSubscriptionReplacement(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Subscription Replacement",
		SpecRef: "MQTT-3.8.4-3",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-sub-replace"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	topic := "test/replace"

	// First subscription with QoS 0
	token := client.Subscribe(topic, 0, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("first subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	// Replace with QoS 1 subscription
	token = client.Subscribe(topic, 1, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("second subscribe failed: %w", token.Error())
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testRetainedMessage tests retained message delivery [MQTT-3.3.1-5, MQTT-3.3.1-6]
func testRetainedMessage(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Retained Message",
		SpecRef: "MQTT-3.3.1-6",
	}

	topic := "test/retained"

	// Publish retained message
	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-retained-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	token := publisher.Publish(topic, 1, true, "retained message")
	token.Wait()
	publisher.Disconnect(250)

	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(200 * time.Millisecond)

	// Subscribe and expect to receive retained message
	var mu sync.Mutex
	var receivedRetained bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedRetained = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-retained-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	token = subscriber.Subscribe(topic, 1, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !receivedRetained {
		result.Error = fmt.Errorf("retained message not received")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testRetainedMessageClear tests clearing retained message [MQTT-3.3.1-10, MQTT-3.3.1-11]
func testRetainedMessageClear(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Clear Retained Message",
		SpecRef: "MQTT-3.3.1-10",
	}

	topic := "test/retained/clear"

	// Publish retained message
	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-clear-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 1, true, "retained").Wait()
	time.Sleep(100 * time.Millisecond)

	// Clear retained message with zero-byte payload
	publisher.Publish(topic, 1, true, "").Wait()
	time.Sleep(100 * time.Millisecond)

	// Subscribe and expect NOT to receive retained message
	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-clear-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(topic, 1, nil).Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedMessage {
		result.Error = fmt.Errorf("received message when retained should have been cleared")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testPublishToMultipleSubscribers tests publish to multiple subscribers [MQTT-3.3.5-1]
func testPublishToMultipleSubscribers(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Publish to Multiple Subscribers",
		SpecRef: "MQTT-3.3.5-1",
	}

	topic := "test/multi/subscribers"
	var wg sync.WaitGroup
	var mu sync.Mutex
	receivedCount := 0

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
	}

	// Create multiple subscribers
	sub1, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-multi-sub1"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber1 connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub1.Disconnect(250)

	sub2, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-multi-sub2"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber2 connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub2.Disconnect(250)

	wg.Add(2)
	sub1.Subscribe(topic, 1, nil).Wait()
	sub2.Subscribe(topic, 1, nil).Wait()

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-multi-pub3"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 1, false, "broadcast").Wait()

	// Wait for messages with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		if receivedCount == 2 {
			result.Passed = true
		} else {
			result.Error = fmt.Errorf("expected 2 subscribers to receive message, got %d", receivedCount)
		}
	case <-time.After(2 * time.Second):
		result.Error = fmt.Errorf("timeout waiting for messages")
	}

	result.Duration = time.Since(start)
	return result
}
