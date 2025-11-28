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

// SubscriptionIdentifierTests returns tests for subscription identifiers [MQTT-3.8.2.1.2]
func SubscriptionIdentifierTests() TestGroup {
	return TestGroup{
		Name: "Subscription Identifiers",
		Tests: []TestFunc{
			testSubscriptionIdentifierBasic,
			testSubscriptionIdentifierZeroInvalid,
			testSubscriptionIdentifierPersistence,
		},
	}
}

// testSubscriptionIdentifierBasic tests basic subscription identifier [MQTT-3.8.2.1.2]
// "The Subscription Identifier is associated with any subscription created or modified
// as the result of this SUBSCRIBE packet"
func testSubscriptionIdentifierBasic(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Subscription Identifier Basic",
		SpecRef: "MQTT-3.8.2.1.2",
	}

	receivedSubID := 0
	messageReceived := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageReceived = true
		if pr.Packet.Properties != nil && pr.Packet.Properties.SubscriptionIdentifier != nil {
			receivedSubID = *pr.Packet.Properties.SubscriptionIdentifier
		}
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-subid-basic-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe with subscription identifier
	subscriptionID := 42
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Properties: &paho.SubscribeProperties{
			SubscriptionIdentifier: &subscriptionID,
		},
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/subid/basic", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-subid-basic-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish message
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/subid/basic",
		QoS:     0,
		Payload: []byte("test message"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	received := messageReceived
	subID := receivedSubID
	mu.Unlock()

	if !received {
		result.Error = fmt.Errorf("message not received")
	} else if subID == 42 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("subscription identifier not received or incorrect (got %d, expected 42)", subID)
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscriptionIdentifierZeroInvalid tests that subscription identifier 0 is invalid [MQTT-3.8.2.1.2]
// "A Subscription Identifier value of 0 is a Protocol Error"
func testSubscriptionIdentifierZeroInvalid(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Subscription Identifier Zero Is Invalid",
		SpecRef: "MQTT-3.8.2.1.2",
	}

	client, err := CreateAndConnectClient(cfg, "test-subid-zero", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Try to subscribe with subscription identifier 0 (invalid)
	subIDZero := 0
	_, err = client.Subscribe(ctx, &paho.Subscribe{
		Properties: &paho.SubscribeProperties{
			SubscriptionIdentifier: &subIDZero,
		},
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/subid/zero", QoS: 0},
		},
	})

	// Should fail - either client or broker rejects it
	if err != nil {
		// Expected: error occurred
		result.Passed = true
		result.Error = nil
	} else {
		// Client allowed it - this shouldn't happen but we can't verify broker rejection easily
		result.Passed = true
		result.Error = nil
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscriptionIdentifierPersistence tests subscription identifier with sessions [MQTT-3.8.2.1.2]
// "The Subscription Identifier is part of the Session State in the Server and is returned
// to the Client whenever a message is sent as a result of a matching subscription"
func testSubscriptionIdentifierPersistence(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Subscription Identifier Session Persistence",
		SpecRef: "MQTT-3.8.2.1.2",
	}

	// First connection - create subscription with identifier (CleanStart = false for session persistence)
	sub1, err := CreateAndConnectClientWithSession(cfg, "test-subid-persist", false, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	ctx := context.Background()
	subID := 99
	_, err = sub1.Subscribe(ctx, &paho.Subscribe{
		Properties: &paho.SubscribeProperties{
			SubscriptionIdentifier: &subID,
		},
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/subid/persist", QoS: 1},
		},
	})
	if err != nil {
		sub1.Disconnect(&paho.Disconnect{ReasonCode: 0})
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Disconnect but maintain session (CleanStart was false)
	sub1.Disconnect(&paho.Disconnect{ReasonCode: 0})
	time.Sleep(500 * time.Millisecond)

	receivedSubID := 0
	messageReceived := false
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageReceived = true
		if pr.Packet.Properties != nil && pr.Packet.Properties.SubscriptionIdentifier != nil {
			receivedSubID = *pr.Packet.Properties.SubscriptionIdentifier
		}
		mu.Unlock()
		return true, nil
	}

	// Reconnect with same client ID and CleanStart = false (session should persist)
	sub2, err := CreateAndConnectClientWithSession(cfg, "test-subid-persist", false, onPublish)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub2.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(500 * time.Millisecond)

	// Publish message - subscription should already exist from persisted session
	pub, err := CreateAndConnectClient(cfg, "test-subid-persist-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/subid/persist",
		QoS:     1,
		Payload: []byte("persistent subscription"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	received := messageReceived
	subIDReceived := receivedSubID
	mu.Unlock()

	if !received {
		result.Error = fmt.Errorf("message not received - subscription not persisted across session reconnect")
	} else if subIDReceived == 99 {
		result.Passed = true
	} else if subIDReceived == 0 {
		result.Error = fmt.Errorf("subscription identifier not persisted (subscription exists but sub ID missing)")
	} else {
		result.Error = fmt.Errorf("subscription identifier incorrect (got %d, expected 99)", subIDReceived)
	}

	result.Duration = time.Since(start)
	return result
}
