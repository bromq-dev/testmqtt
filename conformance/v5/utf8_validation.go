package v5

import (
	"github.com/bromq-dev/testmqtt/conformance/common"
)

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// UTF8ValidationTests returns tests for UTF-8 string validation [MQTT-1.5.4]
func UTF8ValidationTests() TestGroup {
	return TestGroup{
		Name: "UTF-8 String Validation",
		Tests: []TestFunc{
			testUTF8WellFormed,
			testUTF8NoNull,
			testUTF8NoSurrogates,
			testUTF8ValidClientID,
			testUTF8ValidTopicName,
			testUTF8InvalidSequence,
		},
	}
}

// testUTF8WellFormed tests that UTF-8 strings must be well-formed [MQTT-1.5.4-1]
// "The character data in a UTF-8 Encoded String MUST be well-formed UTF-8"
func testUTF8WellFormed(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UTF-8 Strings Must Be Well-Formed",
		SpecRef: "MQTT-1.5.4-1",
	}

	// Test with valid UTF-8 characters including multi-byte sequences
	client, err := CreateAndConnectClient(cfg, "test-utf8-valid-\u4E2D\u6587", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect with valid UTF-8 failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Publish with valid UTF-8 in topic (Chinese characters)
	_, err = client.Publish(ctx, &paho.Publish{
		Topic:   "test/\u4E2D\u6587/topic",
		QoS:     0,
		Payload: []byte("valid UTF-8"),
	})

	if err != nil {
		result.Error = fmt.Errorf("publish with valid UTF-8 topic failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUTF8NoNull tests that null character is not allowed [MQTT-1.5.4-2]
// "A UTF-8 Encoded String MUST NOT include an encoding of the null character U+0000"
func testUTF8NoNull(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UTF-8 Strings Must Not Contain Null (U+0000)",
		SpecRef: "MQTT-1.5.4-2",
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

	// Send CONNECT with null character in client ID
	connectWithNull := []byte{
		0x10,       // CONNECT
		0x11,       // Remaining length 17
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
		0x00,       // Properties length
		0x00, 0x05, // Client ID length (5)
		't', 'e', 0x00, 's', 't', // Client ID with null
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(connectWithNull)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Read response - should be error CONNACK or disconnect
	response := make([]byte, 256)
	n, err := conn.Read(response)

	if err != nil || n == 0 {
		// Connection closed - broker rejected
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == 0x20 {
		// CONNACK - check reason code
		if n >= 4 && response[3] != 0x00 {
			result.Passed = true
			result.Error = nil
		} else {
			result.Passed = false
			result.Error = fmt.Errorf("broker accepted null character in client ID")
		}
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("unexpected broker response to null character")
	}

	result.Duration = time.Since(start)
	return result
}

// testUTF8NoSurrogates tests that UTF-16 surrogates are not allowed [MQTT-1.5.4-3]
// "A UTF-8 Encoded String MUST NOT include encodings of code points between U+D800 and U+DFFF"
func testUTF8NoSurrogates(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UTF-8 Must Not Contain Surrogates (U+D800 to U+DFFF)",
		SpecRef: "MQTT-1.5.4-3",
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

	// Send CONNECT with UTF-16 surrogate in client ID
	// UTF-16 surrogate U+D800 encoded as 0xED 0xA0 0x80 (invalid UTF-8)
	connectWithSurrogate := []byte{
		0x10,       // CONNECT
		0x11,       // Remaining length
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
		0x00,       // Properties length
		0x00, 0x05, // Client ID length
		't', 0xED, 0xA0, 0x80, 't', // Client ID with surrogate
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(connectWithSurrogate)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Read response
	response := make([]byte, 256)
	n, err := conn.Read(response)

	if err != nil || n == 0 {
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == 0x20 {
		if n >= 4 && response[3] != 0x00 {
			result.Passed = true
			result.Error = nil
		} else {
			result.Passed = false
			result.Error = fmt.Errorf("broker accepted UTF-16 surrogate in client ID")
		}
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("unexpected broker response to surrogate")
	}

	result.Duration = time.Since(start)
	return result
}

// testUTF8ValidClientID tests that client IDs must be valid UTF-8 [MQTT-3.1.3-4]
func testUTF8ValidClientID(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Client ID Must Be Valid UTF-8",
		SpecRef: "MQTT-3.1.3-4",
	}

	// Test various valid UTF-8 client IDs
	testIDs := []string{
		"simple-ascii-123",
		"utf8-\u00E9\u00E0\u00FC",  // Latin with accents
		"emoji-\U0001F600",         // Emoji
		"\u4E2D\u6587",             // Chinese
		"\u0420\u0443\u0441\u0441", // Russian
	}

	for _, clientID := range testIDs {
		client, err := CreateAndConnectClient(cfg, clientID, nil)
		if err != nil {
			result.Error = fmt.Errorf("failed to connect with valid UTF-8 client ID '%s': %w", clientID, err)
			result.Duration = time.Since(start)
			return result
		}
		client.Disconnect(&paho.Disconnect{ReasonCode: 0})
		time.Sleep(100 * time.Millisecond)
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUTF8ValidTopicName tests that topic names must be valid UTF-8 [MQTT-4.7.3-3]
func testUTF8ValidTopicName(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Topic Names Must Be Valid UTF-8",
		SpecRef: "MQTT-4.7.3-3",
	}

	client, err := CreateAndConnectClient(cfg, "test-utf8-topics", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(&paho.Disconnect{ReasonCode: 0})

	ctx := context.Background()

	// Test various valid UTF-8 topic names
	testTopics := []string{
		"test/simple/topic",
		"test/\u00E9\u00E0\u00FC/topic",       // Latin with accents
		"test/\U0001F600/topic",               // Emoji
		"test/\u4E2D\u6587/topic",             // Chinese
		"test/\u0420\u0443\u0441\u0441/topic", // Russian
	}

	for _, topic := range testTopics {
		_, err = client.Publish(ctx, &paho.Publish{
			Topic:   topic,
			QoS:     0,
			Payload: []byte("test"),
		})
		if err != nil {
			result.Error = fmt.Errorf("failed to publish to valid UTF-8 topic '%s': %w", topic, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUTF8InvalidSequence tests that invalid UTF-8 sequences are rejected
func testUTF8InvalidSequence(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Reject Invalid UTF-8 Sequences",
		SpecRef: "MQTT-1.5.4-1",
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

	// Send CONNECT with invalid UTF-8 sequence (continuation byte without start)
	connectInvalid := []byte{
		0x10,       // CONNECT
		0x11,       // Remaining length
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
		0x00,       // Properties length
		0x00, 0x05, // Client ID length
		't', 0x80, 0x81, 's', 't', // Invalid UTF-8: orphan continuation bytes
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(connectInvalid)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Read response
	response := make([]byte, 256)
	n, err := conn.Read(response)

	if err != nil || n == 0 {
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == 0x20 {
		if n >= 4 && response[3] != 0x00 {
			result.Passed = true
			result.Error = nil
		} else {
			result.Passed = false
			result.Error = fmt.Errorf("broker accepted invalid UTF-8 sequence")
		}
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("unexpected broker response to invalid UTF-8")
	}

	result.Duration = time.Since(start)
	return result
}
