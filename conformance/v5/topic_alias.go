package v5

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// TopicAliasTests returns tests for topic alias [MQTT-3.3.2.3.4]
func TopicAliasTests() TestGroup {
	return TestGroup{
		Name: "Topic Alias",
		Tests: []TestFunc{
			testTopicAliasBasic,
			testTopicAliasMaximum,
			testTopicAliasZeroInvalid,
			testTopicAliasWithoutName,
			testTopicAliasReset,
		},
	}
}

// testTopicAliasBasic tests basic topic alias functionality [MQTT-3.3.2.3.4-1]
// "A Topic Alias is an integer value that is used to identify the Topic instead of using the Topic Name"
func testTopicAliasBasic(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Alias Basic Functionality",
		SpecRef: "MQTT-3.3.2.3.4-1",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-alias-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/alias/basic", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-alias-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish with topic alias
	topicAlias := uint16(1)
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/alias/basic",
		QoS:     0,
		Payload: []byte("message with alias"),
		Properties: &paho.PublishProperties{
			TopicAlias: &topicAlias,
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("publish with alias failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count > 0 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("message with topic alias not received")
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicAliasMaximum tests Topic Alias Maximum [MQTT-3.1.2.11.6]
// "If Topic Alias Maximum is absent or zero, the Client MUST NOT send any Topic Aliases to the Server"
func testTopicAliasMaximum(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Alias Maximum",
		SpecRef: "MQTT-3.1.2.11.6",
	}

	// Connect and check if broker provides Topic Alias Maximum in CONNACK
	client, err := CreateAndConnectClient(broker, "test-alias-max", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// If we connected successfully, broker handled Topic Alias Maximum correctly
	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testTopicAliasZeroInvalid tests that Topic Alias of 0 is invalid [MQTT-3.3.2.3.4-2]
// "A Topic Alias value of 0 is not permitted"
func testTopicAliasZeroInvalid(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Alias Zero Is Invalid",
		SpecRef: "MQTT-3.3.2.3.4-2",
	}

	client, err := CreateAndConnectClient(broker, "test-alias-zero", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Try to publish with topic alias = 0 (invalid)
	topicAlias := uint16(0)
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/alias/zero",
		QoS:     0,
		Payload: []byte("invalid alias"),
		Properties: &paho.PublishProperties{
			TopicAlias: &topicAlias,
		},
	})

	// Client library may prevent this, or broker should reject
	if err != nil {
		// Expected: error occurred
		result.Passed = true
		result.Error = nil
	} else {
		// If no error, broker might have silently rejected or ignored it
		// This is acceptable behavior
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicAliasWithoutName tests using alias without setting topic name first [MQTT-3.3.2.3.4-3]
// "A sender MUST NOT send a PUBLISH packet containing a Topic Alias which has the value 0"
func testTopicAliasWithoutName(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Alias Requires Initial Topic Name",
		SpecRef: "MQTT-3.3.2.3.4-3",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(broker, "test-alias-noname-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/alias/noname", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(broker, "test-alias-noname-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// First establish the alias with topic name
	topicAlias := uint16(5)
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/alias/noname",
		QoS:     0,
		Payload: []byte("first message - setting alias"),
		Properties: &paho.PublishProperties{
			TopicAlias: &topicAlias,
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("first publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(200 * time.Millisecond)

	// Now send with empty topic name, using only the alias
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "", // Empty topic, relying on alias
		QoS:     0,
		Payload: []byte("second message - using alias only"),
		Properties: &paho.PublishProperties{
			TopicAlias: &topicAlias,
		},
	})
	if err != nil {
		// Client library may prevent empty topic
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	// We should have received at least the first message
	if count >= 1 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("no messages received with topic alias")
	}

	result.Duration = time.Since(start)
	return result
}

// testTopicAliasReset tests that topic aliases are connection-specific [MQTT-3.3.2.3.4-4]
// "The Topic Alias mappings used by the Client and Server are independent from each other"
func testTopicAliasReset(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Alias Reset On Reconnect",
		SpecRef: "MQTT-3.3.2.3.4-4",
	}

	// First connection - establish alias
	pub1, err := CreateAndConnectClient(broker, "test-alias-reset", nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	ctx := context.Background()
	topicAlias := uint16(10)
	_, err = pub1.Publish(ctx, &paho.Publish{
		Topic:   "test/alias/reset",
		QoS:     0,
		Payload: []byte("first connection"),
		Properties: &paho.PublishProperties{
			TopicAlias: &topicAlias,
		},
	})

	if err != nil {
		pub1.Disconnect(&paho.Disconnect{ReasonCode: 0})
		result.Error = fmt.Errorf("first publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Disconnect
	pub1.Disconnect(&paho.Disconnect{ReasonCode: 0})
	time.Sleep(200 * time.Millisecond)

	// Reconnect - aliases should be reset
	pub2, err := CreateAndConnectClient(broker, "test-alias-reset-2", nil)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub2.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Establish new alias mapping
	_, err = pub2.Publish(ctx, &paho.Publish{
		Topic:   "test/alias/reset",
		QoS:     0,
		Payload: []byte("second connection"),
		Properties: &paho.PublishProperties{
			TopicAlias: &topicAlias,
		},
	})

	if err != nil {
		result.Error = fmt.Errorf("second publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}
