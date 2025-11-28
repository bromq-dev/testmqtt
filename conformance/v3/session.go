package v3

import (
	"fmt"
	"sync"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SessionTests returns tests for MQTT v3.1.1 Session State functionality
func SessionTests() common.TestGroup {
	return common.TestGroup{
		Name: "Session State",
		Tests: []common.TestFunc{
			testSessionStatePersistence,
			testSubscriptionPersistence,
			testQoS1MessagePersistence,
			testQoS2MessagePersistence,
			testCleanSessionClearsState,
			testRetainedNotPartOfSession,
		},
	}
}

// testSessionStatePersistence tests session state persists across connections [MQTT-3.1.2-4]
func testSessionStatePersistence(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Session State Persistence",
		SpecRef: "MQTT-3.1.2-4",
	}

	clientID := common.GenerateClientID("test-session-persist")

	// First connection with Clean Session = false
	client1, err := CreateAndConnectClientWithSession(broker, clientID, false, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Subscribe to a topic
	topic := "test/session/persist"
	client1.Subscribe(topic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	// Disconnect
	client1.Disconnect(250)
	time.Sleep(200 * time.Millisecond)

	// Reconnect with same client ID and Clean Session = false
	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	client2, err := CreateAndConnectClientWithSession(broker, clientID, false, messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client2.Disconnect(250)

	time.Sleep(100 * time.Millisecond)

	// Publish to the topic (subscription should still exist)
	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-session-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 1, false, "test message").Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !receivedMessage {
		result.Error = fmt.Errorf("message not received (session state not persisted)")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscriptionPersistence tests subscriptions persist [MQTT-3.1.2-4]
func testSubscriptionPersistence(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Subscription Persistence",
		SpecRef: "MQTT-3.1.2-4",
	}

	clientID := common.GenerateClientID("test-sub-persist")
	topic := "test/session/subscription"

	// Connect and subscribe with Clean Session = false
	client1, err := CreateAndConnectClientWithSession(broker, clientID, false, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	client1.Subscribe(topic, 1, nil).Wait()
	client1.Disconnect(250)
	time.Sleep(200 * time.Millisecond)

	// Publish while client is offline
	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-sub-persist-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 1, false, "offline message").Wait()
	time.Sleep(200 * time.Millisecond)

	// Reconnect and should receive the queued message
	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	client2, err := CreateAndConnectClientWithSession(broker, clientID, false, messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client2.Disconnect(250)

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if !receivedMessage {
		result.Error = fmt.Errorf("queued message not received after reconnect")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS1MessagePersistence tests QoS 1 messages persist [MQTT-3.1.2-5]
func testQoS1MessagePersistence(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS 1 Message Persistence",
		SpecRef: "MQTT-3.1.2-5",
	}

	clientID := common.GenerateClientID("test-qos1-persist")
	topic := "test/session/qos1"

	// Connect and subscribe with Clean Session = false
	client1, err := CreateAndConnectClientWithSession(broker, clientID, false, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	client1.Subscribe(topic, 1, nil).Wait()
	client1.Disconnect(250)
	time.Sleep(200 * time.Millisecond)

	// Publish QoS 1 while offline
	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-qos1-persist-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 1, false, "qos1 offline").Wait()
	time.Sleep(200 * time.Millisecond)

	// Reconnect
	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	client2, err := CreateAndConnectClientWithSession(broker, clientID, false, messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client2.Disconnect(250)

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if !receivedMessage {
		result.Error = fmt.Errorf("QoS 1 message not delivered after reconnect")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testQoS2MessagePersistence tests QoS 2 messages persist [MQTT-3.1.2-5]
func testQoS2MessagePersistence(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "QoS 2 Message Persistence",
		SpecRef: "MQTT-3.1.2-5",
	}

	clientID := common.GenerateClientID("test-qos2-persist")
	topic := "test/session/qos2"

	// Connect and subscribe with Clean Session = false
	client1, err := CreateAndConnectClientWithSession(broker, clientID, false, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	client1.Subscribe(topic, 2, nil).Wait()
	client1.Disconnect(250)
	time.Sleep(200 * time.Millisecond)

	// Publish QoS 2 while offline
	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-qos2-persist-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 2, false, "qos2 offline").Wait()
	time.Sleep(200 * time.Millisecond)

	// Reconnect
	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	client2, err := CreateAndConnectClientWithSession(broker, clientID, false, messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client2.Disconnect(250)

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if !receivedMessage {
		result.Error = fmt.Errorf("QoS 2 message not delivered after reconnect")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testCleanSessionClearsState tests Clean Session = true clears state [MQTT-3.1.2-6]
func testCleanSessionClearsState(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Clean Session Clears State",
		SpecRef: "MQTT-3.1.2-6",
	}

	clientID := common.GenerateClientID("test-clean-clears")
	topic := "test/session/clean"

	// Connect with Clean Session = false and subscribe
	client1, err := CreateAndConnectClientWithSession(broker, clientID, false, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	client1.Subscribe(topic, 1, nil).Wait()
	client1.Disconnect(250)
	time.Sleep(200 * time.Millisecond)

	// Reconnect with Clean Session = true (should clear state)
	client2, err := CreateAndConnectClientWithSession(broker, clientID, true, nil)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	client2.Disconnect(250)
	time.Sleep(200 * time.Millisecond)

	// Publish message
	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-clean-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 1, false, "should not receive").Wait()
	time.Sleep(200 * time.Millisecond)

	// Reconnect again - should NOT receive message (subscription was cleared)
	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	client3, err := CreateAndConnectClientWithSession(broker, clientID, false, messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("third connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client3.Disconnect(250)

	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if receivedMessage {
		result.Error = fmt.Errorf("received message after Clean Session cleared state")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testRetainedNotPartOfSession tests retained messages are not part of session state [MQTT-3.1.2.7]
func testRetainedNotPartOfSession(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Retained Messages Not Part of Session",
		SpecRef: "MQTT-3.1.2.7",
	}

	topic := "test/session/retained"

	// Publish retained message
	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-retained-session-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	publisher.Publish(topic, 1, true, "retained message").Wait()
	publisher.Disconnect(250)
	time.Sleep(200 * time.Millisecond)

	// Connect with Clean Session = true
	clientID := common.GenerateClientID("test-retained-session")
	var mu sync.Mutex
	var receivedRetained bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedRetained = true
		mu.Unlock()
	}

	client, err := CreateAndConnectClientWithSession(broker, clientID, true, messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("client connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Subscribe - should receive retained message even with Clean Session
	client.Subscribe(topic, 1, nil).Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !receivedRetained {
		result.Error = fmt.Errorf("retained message not received (retained should persist independent of session)")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}
