package v3

import (
	"fmt"
	"sync"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// UnsubscribeTests returns tests for MQTT v3.1.1 Unsubscribe functionality
func UnsubscribeTests() common.TestGroup {
	return common.TestGroup{
		Name: "Unsubscribe",
		Tests: []common.TestFunc{
			testBasicUnsubscribe,
			testUnsubscribeStopsDelivery,
			testUnsubscribeMultipleTopics,
			testUnsubscribeAcknowledgement,
			testUnsubscribeNonExistentTopic,
		},
	}
}

// testBasicUnsubscribe tests basic unsubscribe functionality [MQTT-3.10.4-1]
func testBasicUnsubscribe(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Basic Unsubscribe",
		SpecRef: "MQTT-3.10.4-1",
	}

	var mu sync.Mutex
	receivedCount := 0
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/unsubscribe/basic"
	subscriber.Subscribe(topic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsub-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish before unsubscribe
	publisher.Publish(topic, 1, false, "msg1").Wait()
	time.Sleep(500 * time.Millisecond)

	// Unsubscribe
	token := subscriber.Unsubscribe(topic)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	countBeforeUnsub := receivedCount
	mu.Unlock()

	// Publish after unsubscribe
	publisher.Publish(topic, 1, false, "msg2").Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedCount != countBeforeUnsub {
		result.Error = fmt.Errorf("received message after unsubscribe")
	} else if countBeforeUnsub == 0 {
		result.Error = fmt.Errorf("did not receive message before unsubscribe")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeStopsDelivery tests unsubscribe stops new message delivery [MQTT-3.10.4-2]
func testUnsubscribeStopsDelivery(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Unsubscribe Stops Delivery",
		SpecRef: "MQTT-3.10.4-2",
	}

	var mu sync.Mutex
	var receivedAfterUnsub bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedAfterUnsub = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsub-stop"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "test/unsubscribe/stop"
	subscriber.Subscribe(topic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	// Unsubscribe
	subscriber.Unsubscribe(topic).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsub-stop-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish after unsubscribe - should not be received
	publisher.Publish(topic, 1, false, "after unsub").Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedAfterUnsub {
		result.Error = fmt.Errorf("received message after unsubscribe")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeMultipleTopics tests unsubscribe from multiple topics [MQTT-3.10.3-1]
func testUnsubscribeMultipleTopics(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Unsubscribe Multiple Topics",
		SpecRef: "MQTT-3.10.3-1",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsub-multi"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic1 := "test/unsubscribe/multi/1"
	topic2 := "test/unsubscribe/multi/2"

	// Subscribe to both
	topics := map[string]byte{
		topic1: 1,
		topic2: 1,
	}
	subscriber.SubscribeMultiple(topics, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsub-multi-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish to both topics
	publisher.Publish(topic1, 1, false, "msg1").Wait()
	publisher.Publish(topic2, 1, false, "msg2").Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	if len(receivedTopics) != 2 {
		result.Error = fmt.Errorf("did not receive messages on both topics before unsubscribe")
		mu.Unlock()
		result.Duration = time.Since(start)
		return result
	}
	receivedTopics = make(map[string]bool) // Clear
	mu.Unlock()

	// Unsubscribe from both
	subscriber.Unsubscribe(topic1, topic2).Wait()
	time.Sleep(100 * time.Millisecond)

	// Publish again - should not be received
	publisher.Publish(topic1, 1, false, "msg3").Wait()
	publisher.Publish(topic2, 1, false, "msg4").Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) > 0 {
		result.Error = fmt.Errorf("received messages after unsubscribe from multiple topics")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeAcknowledgement tests UNSUBACK is sent [MQTT-3.10.4-4]
func testUnsubscribeAcknowledgement(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Unsubscribe Acknowledgement",
		SpecRef: "MQTT-3.10.4-4",
	}

	client, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsuback"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	topic := "test/unsubscribe/ack"
	client.Subscribe(topic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	token := client.Unsubscribe(topic)
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("unsubscribe timeout (no UNSUBACK)")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", token.Error())
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeNonExistentTopic tests unsubscribe from non-existent topic still gets UNSUBACK [MQTT-3.10.4-5]
func testUnsubscribeNonExistentTopic(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Unsubscribe Non-Existent Topic",
		SpecRef: "MQTT-3.10.4-5",
	}

	client, err := CreateAndConnectClient(broker, common.GenerateClientID("test-unsub-nonexist"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Unsubscribe from topic we never subscribed to
	topic := "test/unsubscribe/nonexistent"
	token := client.Unsubscribe(topic)
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("unsubscribe timeout (no UNSUBACK)")
		result.Duration = time.Since(start)
		return result
	}

	// Should still get UNSUBACK even if not subscribed
	if token.Error() != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", token.Error())
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}
