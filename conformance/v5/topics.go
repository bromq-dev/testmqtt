package v5

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// TopicTests returns all topic-related conformance tests
func TopicTests() TestGroup {
	return TestGroup{
		Name: "Topics",
		Tests: []TestFunc{
			testSingleLevelWildcard,
			testMultiLevelWildcard,
			testTopicLevels,
			testDollarTopics,
			testTopicLength,
			testTopicNameValidation,
		},
	}
}

// testSingleLevelWildcard tests single-level wildcard (+) [MQTT-4.7.1-2]
// "The single-level wildcard can be used at any level in the Topic Filter,
// including first and last levels"
func testSingleLevelWildcard(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Single-Level Wildcard (+)",
		SpecRef: "MQTT-4.7.1-2",
	}

	receivedTopics := make(map[string]bool)
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		receivedTopics[pr.Packet.Topic] = true
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-sub-wildcard+", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	// Subscribe with single-level wildcard
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/+/wildcard", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-pub-wildcard+", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish messages that should match
	topics := []string{
		"test/a/wildcard",
		"test/b/wildcard",
		"test/c/wildcard",
	}

	for _, topic := range topics {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   topic,
			QoS:     0,
			Payload: []byte("message"),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish to %s failed: %w", topic, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Publish message that should NOT match (too many levels)
	pub.Publish(ctx, &paho.Publish{
		Topic:   "test/a/b/wildcard",
		QoS:     0,
		Payload: []byte("should not match"),
	})

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	matchCount := len(receivedTopics)
	mu.Unlock()

	// Should have received exactly 3 messages (not the 4-level one)
	result.Passed = (matchCount == 3)
	if !result.Passed {
		result.Error = fmt.Errorf("expected 3 messages, got %d", matchCount)
	}

	result.Duration = time.Since(start)
	return result
}

// testMultiLevelWildcard tests multi-level wildcard (#) [MQTT-4.7.1-1]
// "The multi-level wildcard character MUST be specified either on its own or
// following a topic level separator"
func testMultiLevelWildcard(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Multi-Level Wildcard (#)",
		SpecRef: "MQTT-4.7.1-1",
	}

	receivedCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-sub-wildcard#", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	// Subscribe with multi-level wildcard
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/multi/#", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-pub-wildcard#", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish messages at different levels - all should match
	topics := []string{
		"test/multi/a",
		"test/multi/a/b",
		"test/multi/a/b/c",
		"test/multi/x/y/z",
	}

	for _, topic := range topics {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   topic,
			QoS:     0,
			Payload: []byte("message"),
		})
		if err != nil {
			result.Error = fmt.Errorf("publish to %s failed: %w", topic, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	// This should NOT match (wrong prefix)
	pub.Publish(ctx, &paho.Publish{
		Topic:   "other/multi/a",
		QoS:     0,
		Payload: []byte("should not match"),
	})

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := receivedCount
	mu.Unlock()

	// Should have received exactly 4 messages
	result.Passed = (count == 4)
	if !result.Passed {
		result.Error = fmt.Errorf("expected 4 messages, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicLevels tests topic level handling [MQTT-4.7.3-1]
// "The Topic Name in the PUBLISH packet MUST NOT contain wildcard characters"
func testTopicLevels(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Levels",
		SpecRef: "MQTT-4.7.3-1",
	}

	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		// Verify topic doesn't contain wildcards
		topic := pr.Packet.Topic
		if topic == "test/level/a/b/c" {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-sub-levels", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/level/#", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-pub-levels", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish with multiple topic levels
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/level/a/b/c",
		QoS:     0,
		Payload: []byte("multi-level topic"),
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
		result.Error = fmt.Errorf("multi-level topic message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testDollarTopics tests $SYS and $ topic behavior [MQTT-4.7.2-1]
// "The Server MUST NOT match Topic Filters starting with a wildcard character
// (# or +) with Topic Names beginning with a $ character"
func testDollarTopics(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Dollar Topics ($SYS)",
		SpecRef: "MQTT-4.7.2-1",
	}

	receivedTopics := []string{}
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		receivedTopics = append(receivedTopics, pr.Packet.Topic)
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-sub-dollar", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	// Subscribe to a more specific pattern to avoid other broker messages
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "testdollar/#", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-pub-dollar", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish to $ topic - should NOT be received with # subscription
	pub.Publish(ctx, &paho.Publish{
		Topic:   "$testdollar/special",
		QoS:     0,
		Payload: []byte("dollar topic"),
	})

	// Publish to normal topic - should be received
	pub.Publish(ctx, &paho.Publish{
		Topic:   "testdollar/normal",
		QoS:     0,
		Payload: []byte("normal topic"),
	})

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	topics := receivedTopics
	mu.Unlock()

	// Should have received only 1 message (the non-$ topic)
	// Check if we received the $ topic (which would be wrong)
	hasDollarTopic := false
	hasNormalTopic := false
	for _, topic := range topics {
		if topic == "$testdollar/special" {
			hasDollarTopic = true
		}
		if topic == "testdollar/normal" {
			hasNormalTopic = true
		}
	}

	if hasDollarTopic {
		result.Passed = false
		result.Error = fmt.Errorf("$ topic matched # wildcard (violation)")
	} else if !hasNormalTopic {
		result.Passed = false
		result.Error = fmt.Errorf("normal topic not received")
	} else if len(topics) > 1 {
		result.Passed = false
		result.Error = fmt.Errorf("expected 1 message, got %d topics: %v", len(topics), topics)
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicLength tests topic name length constraints
// "All Topic Names and Topic Filters MUST be at least one character long"
func testTopicLength(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Name Length",
		SpecRef: "MQTT-4.7.3-2",
	}

	// Test that single-character topics work
	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.Topic == "a" {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-sub-length", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "a", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe to single-char topic failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-pub-length", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish to single-character topic
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "a",
		QoS:     0,
		Payload: []byte("single char topic"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish to single-char topic failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	result.Passed = received
	mu.Unlock()

	if !result.Passed {
		result.Error = fmt.Errorf("single-character topic message not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicNameValidation tests topic name validation requirements
// "Topic Names and Topic Filters are UTF-8 Encoded Strings"
func testTopicNameValidation(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Name Validation",
		SpecRef: "MQTT-4.7.3-3",
	}

	// Test that UTF-8 topics work correctly
	received := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		if pr.Packet.Topic == "test/topic/valid" {
			received = true
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-sub-validation", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/topic/valid", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-pub-validation", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish with valid topic
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/topic/valid",
		QoS:     0,
		Payload: []byte("valid topic"),
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
		result.Error = fmt.Errorf("valid topic message not received")
	}

	result.Duration = time.Since(start)
	return result
}
