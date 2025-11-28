package v3

import (
	"fmt"
	"sync"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// TopicTests returns tests for MQTT v3.1.1 topic functionality
func TopicTests() common.TestGroup {
	return common.TestGroup{
		Name: "Topics",
		Tests: []common.TestFunc{
			testTopicWildcardMultiLevel,
			testTopicWildcardSingleLevel,
			testTopicWildcardCombination,
			testTopicLevelSeparator,
			testTopicSystemPrefix,
			testTopicCaseSensitivity,
			testTopicWithSpaces,
			testTopicLeadingTrailingSlash,
		},
	}
}

// testTopicWildcardMultiLevel tests multi-level wildcard # [MQTT-4.7.1-2]
func testTopicWildcardMultiLevel(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic Multi-Level Wildcard #",
		SpecRef: "MQTT-4.7.1-2",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-multi-wildcard"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// Subscribe to sport/tennis/# should match all below
	token := subscriber.Subscribe("sport/tennis/#", 0, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-multi-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Publish to various topics that should match
	publisher.Publish("sport/tennis/player1", 0, false, "msg1").Wait()
	publisher.Publish("sport/tennis/player1/ranking", 0, false, "msg2").Wait()
	publisher.Publish("sport/tennis/player1/score/wimbledon", 0, false, "msg3").Wait()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) != 3 {
		result.Error = fmt.Errorf("expected 3 messages, received %d", len(receivedTopics))
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicWildcardSingleLevel tests single-level wildcard + [MQTT-4.7.1-3]
func testTopicWildcardSingleLevel(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic Single-Level Wildcard +",
		SpecRef: "MQTT-4.7.1-3",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-single-wildcard"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// Subscribe to sport/tennis/+ should match only one level
	token := subscriber.Subscribe("sport/tennis/+", 0, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-single-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Should match
	publisher.Publish("sport/tennis/player1", 0, false, "msg1").Wait()
	publisher.Publish("sport/tennis/player2", 0, false, "msg2").Wait()

	// Should NOT match (too many levels)
	publisher.Publish("sport/tennis/player1/ranking", 0, false, "msg3").Wait()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) != 2 {
		result.Error = fmt.Errorf("expected 2 messages, received %d", len(receivedTopics))
	} else if receivedTopics["sport/tennis/player1/ranking"] {
		result.Error = fmt.Errorf("received message that should not have matched")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicWildcardCombination tests combination of + and # wildcards
func testTopicWildcardCombination(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic Wildcard Combination +/#",
		SpecRef: "MQTT-4.7.1-3",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-combo-wildcard"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// Subscribe to +/tennis/# should match any first level, then tennis, then anything
	token := subscriber.Subscribe("+/tennis/#", 0, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-combo-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish("sport/tennis/player1", 0, false, "msg1").Wait()
	publisher.Publish("event/tennis/tournament", 0, false, "msg2").Wait()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) != 2 {
		result.Error = fmt.Errorf("expected 2 messages, received %d", len(receivedTopics))
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicLevelSeparator tests topic level separator / handling [MQTT-4.7.3-1]
func testTopicLevelSeparator(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic Level Separator",
		SpecRef: "MQTT-4.7.3-1",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-separator"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// Subscribe to multiple distinct topics
	subscriber.Subscribe("finance", 0, nil).Wait()
	subscriber.Subscribe("/finance", 0, nil).Wait()

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-sep-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// These are different topics
	publisher.Publish("finance", 0, false, "msg1").Wait()
	publisher.Publish("/finance", 0, false, "msg2").Wait()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) != 2 {
		result.Error = fmt.Errorf("expected 2 distinct messages, received %d", len(receivedTopics))
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicSystemPrefix tests $SYS topics not matched by wildcards [MQTT-4.7.2-1]
func testTopicSystemPrefix(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic $SYS Prefix",
		SpecRef: "MQTT-4.7.2-1",
	}

	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		if msg.Topic() == "$SYS/test/topic" {
			receivedMessage = true
		}
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-sys"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// Subscribe to # should NOT receive $SYS topics
	subscriber.Subscribe("#", 0, nil).Wait()

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-sys-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	// Attempt to publish to $SYS topic (broker may reject this)
	publisher.Publish("$SYS/test/topic", 0, false, "sys message").Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// Should NOT receive $SYS topic through # wildcard
	if receivedMessage {
		result.Error = fmt.Errorf("received $SYS topic through # wildcard")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicCaseSensitivity tests that topics are case sensitive [MQTT-4.7.3-1]
func testTopicCaseSensitivity(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic Case Sensitivity",
		SpecRef: "MQTT-4.7.3-4",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-case"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// Subscribe to lowercase only
	subscriber.Subscribe("accounts", 0, nil).Wait()

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-case-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish("accounts", 0, false, "msg1").Wait()
	publisher.Publish("Accounts", 0, false, "msg2").Wait() // Should NOT match
	publisher.Publish("ACCOUNTS", 0, false, "msg3").Wait() // Should NOT match

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) != 1 {
		result.Error = fmt.Errorf("expected 1 message, received %d (topics are case sensitive)", len(receivedTopics))
	} else if !receivedTopics["accounts"] {
		result.Error = fmt.Errorf("did not receive message on expected topic")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicWithSpaces tests that topics can include spaces [MQTT-4.7.3-1]
func testTopicWithSpaces(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic With Spaces",
		SpecRef: "MQTT-4.7.3-1",
	}

	var mu sync.Mutex
	var receivedMessage bool
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedMessage = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-spaces"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	topic := "accounts payable"
	subscriber.Subscribe(topic, 0, nil).Wait()

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-spaces-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish(topic, 0, false, "message").Wait()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !receivedMessage {
		result.Error = fmt.Errorf("message not received on topic with spaces")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicLeadingTrailingSlash tests leading/trailing slash creates distinct topics [MQTT-4.7.3-1]
func testTopicLeadingTrailingSlash(broker string) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Topic Leading/Trailing Slash",
		SpecRef: "MQTT-4.7.3-1",
	}

	var mu sync.Mutex
	receivedTopics := make(map[string]bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		receivedTopics[msg.Topic()] = true
		mu.Unlock()
	}

	subscriber, err := CreateAndConnectClient(broker, common.GenerateClientID("test-slash"), messageHandler)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subscriber.Disconnect(250)

	// These are all different topics
	subscriber.Subscribe("topic", 0, nil).Wait()
	subscriber.Subscribe("/topic", 0, nil).Wait()
	subscriber.Subscribe("topic/", 0, nil).Wait()
	subscriber.Subscribe("/topic/", 0, nil).Wait()

	time.Sleep(100 * time.Millisecond)

	publisher, err := CreateAndConnectClient(broker, common.GenerateClientID("test-slash-pub"), nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer publisher.Disconnect(250)

	publisher.Publish("topic", 0, false, "msg1").Wait()
	publisher.Publish("/topic", 0, false, "msg2").Wait()
	publisher.Publish("topic/", 0, false, "msg3").Wait()
	publisher.Publish("/topic/", 0, false, "msg4").Wait()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedTopics) != 4 {
		result.Error = fmt.Errorf("expected 4 distinct topics, received %d", len(receivedTopics))
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}
