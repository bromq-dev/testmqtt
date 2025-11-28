package v3

import (
	"fmt"
	"sync"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// WillTests returns tests for MQTT v3.1.1 Will Message functionality
func WillTests() common.TestGroup {
	return common.TestGroup{
		Name: "Will Messages",
		Tests: []common.TestFunc{
			testWillMessageOnAbnormalDisconnect,
			testWillMessageNotSentOnCleanDisconnect,
			testWillMessageQoS0,
			testWillMessageQoS1,
			testWillMessageQoS2,
			testWillMessageRetained,
			testWillMessageNotRetained,
		},
	}
}

// testWillMessageOnAbnormalDisconnect tests will message is sent on abnormal disconnect [MQTT-3.1.2-8]
func testWillMessageOnAbnormalDisconnect(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Will Message on Abnormal Disconnect",
		SpecRef: "MQTT-3.1.2-8",
	}

	// Subscribe to will topic
	var mu sync.Mutex
	var receivedWill bool
	willTopic := "test/will/abnormal"

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == willTopic {
			receivedWill = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-will-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(willTopic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	// Create client with will message
	client, err := CreateAndConnectClientWithWill(
		broker,
		common.GenerateClientID("test-will-client"),
		willTopic,
		[]byte("will message"),
		1,
		false,
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("client with will connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Force disconnect by getting the underlying connection (paho.mqtt.golang doesn't expose clean way)
	// We'll just disconnect without DISCONNECT packet by using very short timeout
	client.Disconnect(0) // 0ms timeout = abrupt close

	time.Sleep(1 * time.Second) // Wait for will to be published

	mu.Lock()
	defer mu.Unlock()
	if !receivedWill {
		result.Error = fmt.Errorf("will message not received after abnormal disconnect")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testWillMessageNotSentOnCleanDisconnect tests will message NOT sent on DISCONNECT [MQTT-3.1.2-10]
func testWillMessageNotSentOnCleanDisconnect(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Will Message Not Sent on Clean Disconnect",
		SpecRef: "MQTT-3.1.2-10",
	}

	// Subscribe to will topic
	var mu sync.Mutex
	var receivedWill bool
	willTopic := "test/will/clean"

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == willTopic {
			receivedWill = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-will-clean-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(willTopic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	// Create client with will message
	client, err := CreateAndConnectClientWithWill(
		broker,
		common.GenerateClientID("test-will-clean-client"),
		willTopic,
		[]byte("will message"),
		1,
		false,
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("client with will connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Clean disconnect with DISCONNECT packet
	client.Disconnect(250)

	time.Sleep(1 * time.Second) // Wait to ensure will is NOT published

	mu.Lock()
	defer mu.Unlock()
	if receivedWill {
		result.Error = fmt.Errorf("will message was sent on clean disconnect (should not be)")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testWillMessageQoS0 tests will message with QoS 0 [MQTT-3.1.2-9]
func testWillMessageQoS0(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Will Message QoS 0",
		SpecRef: "MQTT-3.1.2-9",
	}

	var mu sync.Mutex
	var receivedWill bool
	willTopic := "test/will/qos0"

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == willTopic {
			receivedWill = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-will-qos0-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(willTopic, 0, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	client, err := CreateAndConnectClientWithWill(
		broker,
		common.GenerateClientID("test-will-qos0-client"),
		willTopic,
		[]byte("will qos0"),
		0, // QoS 0
		false,
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("client with will connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)
	client.Disconnect(0) // Abnormal disconnect
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if !receivedWill {
		result.Error = fmt.Errorf("will message QoS 0 not received")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testWillMessageQoS1 tests will message with QoS 1 [MQTT-3.1.2-14]
func testWillMessageQoS1(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Will Message QoS 1",
		SpecRef: "MQTT-3.1.2-14",
	}

	var mu sync.Mutex
	var receivedWill bool
	willTopic := "test/will/qos1"

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == willTopic {
			receivedWill = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-will-qos1-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(willTopic, 1, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	client, err := CreateAndConnectClientWithWill(
		broker,
		common.GenerateClientID("test-will-qos1-client"),
		willTopic,
		[]byte("will qos1"),
		1, // QoS 1
		false,
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("client with will connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)
	client.Disconnect(0) // Abnormal disconnect
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if !receivedWill {
		result.Error = fmt.Errorf("will message QoS 1 not received")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testWillMessageQoS2 tests will message with QoS 2 [MQTT-3.1.2-14]
func testWillMessageQoS2(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Will Message QoS 2",
		SpecRef: "MQTT-3.1.2-14",
	}

	var mu sync.Mutex
	var receivedWill bool
	willTopic := "test/will/qos2"

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == willTopic {
			receivedWill = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-will-qos2-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(willTopic, 2, nil).Wait()
	time.Sleep(100 * time.Millisecond)

	client, err := CreateAndConnectClientWithWill(
		broker,
		common.GenerateClientID("test-will-qos2-client"),
		willTopic,
		[]byte("will qos2"),
		2, // QoS 2
		false,
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("client with will connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)
	client.Disconnect(0) // Abnormal disconnect
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if !receivedWill {
		result.Error = fmt.Errorf("will message QoS 2 not received")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testWillMessageRetained tests will message with retain flag [MQTT-3.1.2-17]
func testWillMessageRetained(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Will Message Retained",
		SpecRef: "MQTT-3.1.2-17",
	}

	willTopic := "test/will/retained"

	// Create client with retained will message
	client, err := CreateAndConnectClientWithWill(
		broker,
		common.GenerateClientID("test-will-retained-client"),
		willTopic,
		[]byte("retained will"),
		1,
		true, // Retained
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("client with will connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)
	client.Disconnect(0) // Abnormal disconnect to trigger will
	time.Sleep(1 * time.Second)

	// Now subscribe and should receive retained will
	var mu sync.Mutex
	var receivedRetained bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == willTopic {
			receivedRetained = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-will-retained-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(willTopic, 1, nil).Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !receivedRetained {
		result.Error = fmt.Errorf("retained will message not received by new subscriber")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testWillMessageNotRetained tests will message without retain flag [MQTT-3.1.2-16]
func testWillMessageNotRetained(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Will Message Not Retained",
		SpecRef: "MQTT-3.1.2-16",
	}

	willTopic := "test/will/notretained"

	// Create client with non-retained will message
	client, err := CreateAndConnectClientWithWill(
		broker,
		common.GenerateClientID("test-will-notretained-client"),
		willTopic,
		[]byte("non-retained will"),
		1,
		false, // Not retained
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("client with will connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)
	client.Disconnect(0) // Abnormal disconnect to trigger will
	time.Sleep(1 * time.Second)

	// Now subscribe and should NOT receive retained will
	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == willTopic {
			receivedMessage = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-will-notretained-sub"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	subscriber.Subscribe(willTopic, 1, nil).Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedMessage {
		result.Error = fmt.Errorf("non-retained will message was received by new subscriber (should not be)")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}
