package v5

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// SharedSubscriptionTests returns tests for shared subscriptions [MQTT-4.8.2]
func SharedSubscriptionTests() TestGroup {
	return TestGroup{
		Name: "Shared Subscriptions",
		Tests: []TestFunc{
			testSharedSubscriptionBasic,
			testSharedSubscriptionLoadBalancing,
			testSharedSubscriptionQoS,
			testSharedSubscriptionAndNormalSubscription,
			testSharedSubscriptionMultipleGroups,
		},
	}
}

// testSharedSubscriptionBasic tests basic shared subscription [MQTT-4.8.2-1]
// "Shared Subscriptions are defined using the $share prefix"
func testSharedSubscriptionBasic(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Shared Subscription Basic",
		SpecRef: "MQTT-4.8.2-1",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	// Create two subscribers in the same share group
	sub1, err := CreateAndConnectClient(broker, "test-share-basic-1", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber 1 connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub1.Disconnect(&paho.Disconnect{ReasonCode: 0})

	sub2, err := CreateAndConnectClient(broker, "test-share-basic-2", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber 2 connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub2.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Both subscribe to the same shared subscription
	shareName := "$share/group1/test/share/basic"
	_, err = sub1.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: shareName, QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscriber 1 subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	_, err = sub2.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: shareName, QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscriber 2 subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish a message
	pub, err := CreateAndConnectClient(broker, "test-share-basic-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/share/basic",
		QoS:     0,
		Payload: []byte("shared message"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	// Message should be received by only ONE subscriber (not both)
	if count == 1 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected 1 message delivery, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testSharedSubscriptionLoadBalancing tests load balancing [MQTT-4.8.2-2]
// "The Server MUST distribute the messages to the subscribers in the group"
func testSharedSubscriptionLoadBalancing(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Shared Subscription Load Balancing",
		SpecRef: "MQTT-4.8.2-2",
	}

	count1 := 0
	count2 := 0
	var mu sync.Mutex

	onPublish1 := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		count1++
		mu.Unlock()
		return true, nil
	}

	onPublish2 := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		count2++
		mu.Unlock()
		return true, nil
	}

	// Create two subscribers in the same share group
	sub1, err := CreateAndConnectClient(broker, "test-share-lb-1", onPublish1)
	if err != nil {
		result.Error = fmt.Errorf("subscriber 1 connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub1.Disconnect(&paho.Disconnect{ReasonCode: 0})

	sub2, err := CreateAndConnectClient(broker, "test-share-lb-2", onPublish2)
	if err != nil {
		result.Error = fmt.Errorf("subscriber 2 connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub2.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Both subscribe to the same shared subscription
	shareName := "$share/group2/test/share/loadbalance"
	_, err = sub1.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: shareName, QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscriber 1 subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	_, err = sub2.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: shareName, QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscriber 2 subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish multiple messages
	pub, err := CreateAndConnectClient(broker, "test-share-lb-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	messageCount := 10
	for i := 0; i < messageCount; i++ {
		_, err = pub.Publish(ctx, &paho.Publish{
			Topic:   "test/share/loadbalance",
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
	c1 := count1
	c2 := count2
	mu.Unlock()

	total := c1 + c2

	// All messages should be received, distributed between both subscribers
	if total == messageCount && c1 > 0 && c2 > 0 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("load balancing failed: sub1=%d, sub2=%d, total=%d (expected %d)", c1, c2, total, messageCount)
	}

	result.Duration = time.Since(start)
	return result
}

// testSharedSubscriptionQoS tests QoS with shared subscriptions [MQTT-4.8.2]
func testSharedSubscriptionQoS(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Shared Subscription with QoS",
		SpecRef: "MQTT-4.8.2",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	// Create subscriber with QoS 1
	sub, err := CreateAndConnectClient(broker, "test-share-qos-1", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	shareName := "$share/group3/test/share/qos"
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: shareName, QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish with QoS 1
	pub, err := CreateAndConnectClient(broker, "test-share-qos-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/share/qos",
		QoS:     1,
		Payload: []byte("qos message"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count == 1 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected 1 message, got %d", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testSharedSubscriptionAndNormalSubscription tests mixing shared and normal subscriptions
func testSharedSubscriptionAndNormalSubscription(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Shared and Normal Subscriptions Coexist",
		SpecRef: "MQTT-4.8.2",
	}

	sharedCount := 0
	normalCount := 0
	var mu sync.Mutex

	onPublishShared := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		sharedCount++
		mu.Unlock()
		return true, nil
	}

	onPublishNormal := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		normalCount++
		mu.Unlock()
		return true, nil
	}

	// Create shared subscriber
	subShared, err := CreateAndConnectClient(broker, "test-share-mixed-shared", onPublishShared)
	if err != nil {
		result.Error = fmt.Errorf("shared subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subShared.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Create normal subscriber
	subNormal, err := CreateAndConnectClient(broker, "test-share-mixed-normal", onPublishNormal)
	if err != nil {
		result.Error = fmt.Errorf("normal subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subNormal.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe with shared subscription
	shareName := "$share/group4/test/share/mixed"
	_, err = subShared.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: shareName, QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("shared subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Subscribe with normal subscription to the same topic
	_, err = subNormal.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/share/mixed", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("normal subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish message
	pub, err := CreateAndConnectClient(broker, "test-share-mixed-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/share/mixed",
		QoS:     0,
		Payload: []byte("mixed message"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	shared := sharedCount
	normal := normalCount
	mu.Unlock()

	// Both should receive the message (shared and normal subscriptions are independent)
	if shared == 1 && normal == 1 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected shared=1, normal=1, got shared=%d, normal=%d", shared, normal)
	}

	result.Duration = time.Since(start)
	return result
}

// testSharedSubscriptionMultipleGroups tests multiple share groups on same topic
func testSharedSubscriptionMultipleGroups(broker string) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Multiple Share Groups on Same Topic",
		SpecRef: "MQTT-4.8.2",
	}

	countGroup1 := 0
	countGroup2 := 0
	var mu sync.Mutex

	onPublishGroup1 := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		countGroup1++
		mu.Unlock()
		return true, nil
	}

	onPublishGroup2 := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		countGroup2++
		mu.Unlock()
		return true, nil
	}

	// Create subscribers in different share groups
	subGroup1, err := CreateAndConnectClient(broker, "test-share-groups-1", onPublishGroup1)
	if err != nil {
		result.Error = fmt.Errorf("group1 subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subGroup1.Disconnect(&paho.Disconnect{ReasonCode: 0})

	subGroup2, err := CreateAndConnectClient(broker, "test-share-groups-2", onPublishGroup2)
	if err != nil {
		result.Error = fmt.Errorf("group2 subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer subGroup2.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe to different share groups but same topic
	_, err = subGroup1.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "$share/groupA/test/share/groups", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("group1 subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	_, err = subGroup2.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "$share/groupB/test/share/groups", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("group2 subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish message
	pub, err := CreateAndConnectClient(broker, "test-share-groups-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/share/groups",
		QoS:     0,
		Payload: []byte("multi-group message"),
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	g1 := countGroup1
	g2 := countGroup2
	mu.Unlock()

	// Both groups should receive the message (different groups are independent)
	if g1 == 1 && g2 == 1 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected group1=1, group2=1, got group1=%d, group2=%d", g1, g2)
	}

	result.Duration = time.Since(start)
	return result
}
