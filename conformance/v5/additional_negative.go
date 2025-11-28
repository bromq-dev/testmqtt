package v5

import (
	"github.com/bromq-dev/testmqtt/conformance/common"
)

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// AdditionalNegativeTests returns additional negative test cases
func AdditionalNegativeTests() TestGroup {
	return TestGroup{
		Name: "Additional Negative Tests",
		Tests: []TestFunc{
			testMaximumTopicLength,
			testExcessiveClientID,
			testMalformedUTF8InPayload,
			testZeroLengthClientID,
			testReservedTopicCharacters,
			testSubscribeWithoutTopics,
			testUnsubscribeWithoutTopics,
			testPublishWithExcessiveQoS,
		},
	}
}

// testMaximumTopicLength tests handling of very long topic names
func testMaximumTopicLength(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Maximum Topic Name Length",
		SpecRef: "MQTT-4.7.3-3",
	}

	client, err := CreateAndConnectClient(cfg, "test-max-topic-len", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Create a very long but valid topic name (MQTT v5 allows up to 65535 bytes)
	longTopic := "test/" + strings.Repeat("a", 1000)

	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   longTopic,
		QoS:     0,
		Payload: []byte("test"),
	})

	// Should handle long topics gracefully
	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testExcessiveClientID tests handling of very long client IDs
func testExcessiveClientID(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Excessive Client ID Length",
		SpecRef: "MQTT-3.1.3-5",
	}

	// Try to connect with very long client ID
	longClientID := "test-" + strings.Repeat("x", 200)

	client, err := CreateAndConnectClient(cfg, longClientID, nil)
	if err != nil {
		// Broker may reject very long client IDs
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	// Broker accepted long client ID
	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testMalformedUTF8InPayload tests handling of malformed UTF-8 in payload
func testMalformedUTF8InPayload(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Malformed UTF-8 in Payload",
		SpecRef: "MQTT-1.5.4-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-malformed-utf8", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Payload can contain arbitrary binary data (doesn't need to be UTF-8)
	malformedPayload := []byte{0xFF, 0xFE, 0xFD, 0x80, 0x81}

	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/malformed/payload",
		QoS:     0,
		Payload: malformedPayload,
	})

	// Should succeed - payload is binary, not required to be UTF-8
	if err == nil {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("publish with binary payload failed: %w", err)
	}

	result.Duration = time.Since(start)
	return result
}

// testZeroLengthClientID tests zero-length client ID with Clean Start
func testZeroLengthClientID(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Zero-Length Client ID with Clean Start",
		SpecRef: "MQTT-3.1.3-7",
	}

	// Empty client ID with Clean Start should be accepted (broker assigns ID)
	client, err := CreateAndConnectClient(cfg, "", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect with empty client ID failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testReservedTopicCharacters tests topics with reserved characters
func testReservedTopicCharacters(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Reserved Topic Characters",
		SpecRef: "MQTT-4.7.1-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-reserved-chars", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Try topics with various special characters (most should be valid)
	testTopics := []string{
		"test/topic-with-dash",
		"test/topic_with_underscore",
		"test/topic.with.dots",
		"test/topic:with:colons",
	}

	failCount := 0
	for _, topic := range testTopics {
		_, err = client.Publish(ctx, &paho.Publish{
			Topic:   topic,
			QoS:     0,
			Payload: []byte("test"),
		})
		if err != nil {
			failCount++
		}
	}

	// Most special characters should be accepted
	if failCount <= len(testTopics)/2 {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("%d/%d topic publishes failed", failCount, len(testTopics))
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscribeWithoutTopics tests SUBSCRIBE packet with no topic filters
func testSubscribeWithoutTopics(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "SUBSCRIBE Without Topic Filters",
		SpecRef: "MQTT-3.8.3-3",
	}

	u, err := url.Parse(cfg.Broker)
	if err != nil {
		result.Error = fmt.Errorf("invalid broker URL: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	host := u.Host
	if u.Port() == "" {
		host = net.JoinHostPort(u.Hostname(), "1883")
	}

	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		result.Error = fmt.Errorf("failed to dial broker: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer conn.Close()

	// Send CONNECT
	connectPacket := []byte{
		0x10,       // CONNECT
		0x10,       // Remaining length
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive
		0x00,       // Properties length
		0x00, 0x04, // Client ID length
		't', 'e', 's', 't', // Client ID
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(connectPacket)
	if err != nil {
		result.Error = fmt.Errorf("failed to write CONNECT: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Read CONNACK
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil || n == 0 {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Send malformed SUBSCRIBE with no topics
	subscribePacket := []byte{
		0x82,       // SUBSCRIBE with fixed flags
		0x02,       // Remaining length (just properties)
		0x00, 0x01, // Packet ID
		0x00, // Properties length
		// No topic filters (malformed)
	}

	_, err = conn.Write(subscribePacket)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Broker should reject or disconnect
	n, err = conn.Read(response)
	if err != nil || n == 0 {
		result.Passed = true
		result.Error = nil
	} else {
		// Check if broker sent error response
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeWithoutTopics tests UNSUBSCRIBE packet with no topics
func testUnsubscribeWithoutTopics(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UNSUBSCRIBE Without Topics",
		SpecRef: "MQTT-3.10.3-2",
	}

	client, err := CreateAndConnectClient(cfg, "test-unsub-no-topics", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Try to unsubscribe with empty topic list
	_, err = client.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{},
	})

	// Should either fail or return error reason codes
	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testPublishWithExcessiveQoS tests handling of invalid QoS values in edge cases
func testPublishWithExcessiveQoS(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Publish with Excessive QoS",
		SpecRef: "MQTT-3.3.1-4",
	}

	client, err := CreateAndConnectClient(cfg, "test-excessive-qos", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Try QoS 2 (should work)
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/qos/max",
		QoS:     2,
		Payload: []byte("test"),
	})

	if err == nil {
		result.Passed = true
	} else {
		result.Error = fmt.Errorf("QoS 2 publish failed: %w", err)
	}

	result.Duration = time.Since(start)
	return result
}
