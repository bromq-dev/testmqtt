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

// SubscribeExtendedTests returns extended subscription tests [MQTT-3.8, 3.9]
func SubscribeExtendedTests() TestGroup {
	return TestGroup{
		Name: "SUBSCRIBE Extended Features",
		Tests: []TestFunc{
			testSubscribePacketIdentifier,
			testSubscribeMultipleFilters,
			testSubscriptionOptions,
			testSubscribeQoSDowngrade,
			testSUBACKReasonCodes,
			testRetainAsPublished,
			testNoLocal,
			testRetainHandling,
		},
	}
}

// testSubscribePacketIdentifier tests SUBSCRIBE packet identifier [MQTT-3.8.2-1]
// "The Packet Identifier field is used to identify the SUBSCRIBE packet"
func testSubscribePacketIdentifier(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "SUBSCRIBE Packet Identifier",
		SpecRef: "MQTT-3.8.2-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-sub-packetid", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe - paho handles packet ID automatically
	suback, err := client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/sub/id", QoS: 1},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// If we got SUBACK, packet ID was handled correctly
	if suback != nil && len(suback.Reasons) > 0 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("no SUBACK received")
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscribeMultipleFilters tests multiple subscriptions in one packet [MQTT-3.8.3-3]
// "The Payload of a SUBSCRIBE packet MUST contain at least one Topic Filter and Subscription Options pair"
func testSubscribeMultipleFilters(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "SUBSCRIBE Multiple Topic Filters",
		SpecRef: "MQTT-3.8.3-3",
	}

	client, err := CreateAndConnectClient(cfg, "test-sub-multi", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe to multiple topics at once
	suback, err := client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/multi/1", QoS: 0},
			{Topic: "test/multi/2", QoS: 1},
			{Topic: "test/multi/3", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// SUBACK should have 3 reason codes (one per subscription)
	if suback != nil && len(suback.Reasons) == 3 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("expected 3 reason codes in SUBACK, got %d", len(suback.Reasons))
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscriptionOptions tests subscription options [MQTT-3.8.3.1]
// "Subscription Options contains fields QoS, NL (No Local), RAP (Retain As Published), and Retain Handling"
func testSubscriptionOptions(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Subscription Options",
		SpecRef: "MQTT-3.8.3.1",
	}

	client, err := CreateAndConnectClient(cfg, "test-sub-options", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe with various options
	suback, err := client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic:             "test/options/1",
				QoS:               1,
				NoLocal:           false,
				RetainAsPublished: false,
				RetainHandling:    0, // Send retained messages
			},
			{
				Topic:             "test/options/2",
				QoS:               2,
				NoLocal:           true,
				RetainAsPublished: true,
				RetainHandling:    1, // Send retained only on new sub
			},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	if suback != nil && len(suback.Reasons) == 2 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("subscription with options failed")
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscribeQoSDowngrade tests QoS downgrade [MQTT-3.8.4-5]
// "The Server might grant a lower QoS than the Client requested"
func testSubscribeQoSDowngrade(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "SUBSCRIBE QoS Downgrade Allowed",
		SpecRef: "MQTT-3.8.4-5",
	}

	client, err := CreateAndConnectClient(cfg, "test-sub-downgrade", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe with QoS 2
	suback, err := client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/qos/downgrade", QoS: 2},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Check SUBACK - granted QoS can be 0, 1, or 2
	if suback != nil && len(suback.Reasons) > 0 {
		grantedQoS := suback.Reasons[0]
		if grantedQoS <= 2 {
			result.Passed = true
		} else {
			result.Error = fmt.Errorf("invalid granted QoS: 0x%02X", grantedQoS)
		}
	} else {
		result.Error = fmt.Errorf("no SUBACK received")
	}

	result.Duration = time.Since(start)
	return result
}

// testSUBACKReasonCodes tests SUBACK reason codes [MQTT-3.9.3-1]
// "The SUBACK packet contains a list of Reason Codes"
func testSUBACKReasonCodes(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "SUBACK Reason Codes",
		SpecRef: "MQTT-3.9.3-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-suback-reason", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe - should get success reason codes
	suback, err := client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: "test/suback/reason", QoS: 0},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Check for valid reason code (0x00, 0x01, 0x02 for granted QoS 0, 1, 2)
	if suback != nil && len(suback.Reasons) > 0 {
		reason := suback.Reasons[0]
		// 0x00, 0x01, 0x02 = Granted QoS 0, 1, 2
		if reason <= 0x02 {
			result.Passed = true
		} else {
			result.Error = fmt.Errorf("unexpected reason code: 0x%02X", reason)
		}
	} else {
		result.Error = fmt.Errorf("no SUBACK received")
	}

	result.Duration = time.Since(start)
	return result
}

// testRetainAsPublished tests Retain As Published option [MQTT-3.8.3.1-3]
// "If Retain As Published is set to 1, the Server MUST set the RETAIN flag equal
// to the RETAIN flag in the PUBLISH packet"
func testRetainAsPublished(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Retain As Published Option",
		SpecRef: "MQTT-3.8.3.1-3",
	}

	var mu sync.Mutex
	receivedRetain := false

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		receivedRetain = pr.Packet.Retain
		mu.Unlock()
		return true, nil
	}

	sub, err := CreateAndConnectClient(cfg, "test-rap-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe with RetainAsPublished = true
	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic:             "test/rap",
				QoS:               0,
				RetainAsPublished: true,
			},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	pub, err := CreateAndConnectClient(cfg, "test-rap-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(100 * time.Millisecond)

	// Publish with retain flag
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/rap",
		QoS:     0,
		Payload: []byte("retained message"),
		Retain:  true,
	})
	if err != nil {
		result.Error = fmt.Errorf("publish failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	retain := receivedRetain
	mu.Unlock()

	// With RetainAsPublished=true, we should receive the retain flag
	result.Passed = retain
	if !result.Passed {
		result.Error = fmt.Errorf("retain flag not preserved (RetainAsPublished not working)")
	}

	result.Duration = time.Since(start)
	return result
}

// testNoLocal tests No Local option [MQTT-3.8.3.1-2]
// "If No Local is set to 1, Application Messages MUST NOT be forwarded to
// a connection with a ClientID equal to the ClientID of the publishing connection"
func testNoLocal(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "No Local Subscription Option",
		SpecRef: "MQTT-3.8.3.1-2",
	}

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	client, err := CreateAndConnectClient(cfg, "test-nolocal", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Subscribe with NoLocal = true
	_, err = client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic:   "test/nolocal",
				QoS:     0,
				NoLocal: true,
			},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Publish to our own subscription with NoLocal=true
	// We should NOT receive this message
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/nolocal",
		QoS:     0,
		Payload: []byte("should not receive"),
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

	if count == 0 {
		// Correctly did not receive our own message
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("received %d messages with NoLocal=true (should be 0)", count)
	}

	result.Duration = time.Since(start)
	return result
}

// testRetainHandling tests Retain Handling option [MQTT-3.8.3.1-4]
// "Retain Handling indicates whether retained messages are sent when the subscription is established"
func testRetainHandling(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Retain Handling Option",
		SpecRef: "MQTT-3.8.3.1-4",
	}

	// First publish a retained message
	pub, err := CreateAndConnectClient(cfg, "test-retainhandle-pub", nil)
	if err != nil {
		result.Error = fmt.Errorf("publisher connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	ctx := context.Background()
	_, err = pub.Publish(ctx, &paho.Publish{
		Topic:   "test/retainhandle",
		QoS:     0,
		Payload: []byte("retained"),
		Retain:  true,
	})
	if err != nil {
		pub.Disconnect(&paho.Disconnect{ReasonCode: 0})
		result.Error = fmt.Errorf("publish retained failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	pub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	time.Sleep(200 * time.Millisecond)

	messageCount := 0
	var mu sync.Mutex

	onPublish := func(pr paho.PublishReceived) (bool, error) {
		mu.Lock()
		messageCount++
		mu.Unlock()
		return true, nil
	}

	// Subscribe with RetainHandling = 2 (do not send retained messages)
	sub, err := CreateAndConnectClient(cfg, "test-retainhandle-sub", onPublish)
	if err != nil {
		result.Error = fmt.Errorf("subscriber connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer sub.Disconnect(&paho.Disconnect{ReasonCode: 0})

	_, err = sub.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic:          "test/retainhandle",
				QoS:            0,
				RetainHandling: 2, // Do not send retained messages
			},
		},
	})
	if err != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	if count == 0 {
		// With RetainHandling=2, we should not receive retained message
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("received %d retained messages with RetainHandling=2 (should be 0)", count)
	}

	result.Duration = time.Since(start)
	return result
}
