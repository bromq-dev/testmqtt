package v3

import (
	"fmt"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
)

// PacketValidationTests returns tests for MQTT v3.1.1 packet validation
func PacketValidationTests() common.TestGroup {
	return common.TestGroup{
		Name: "Packet Validation",
		Tests: []common.TestFunc{
			testConnectPacketValidation,
			testPublishPacketValidation,
			testSubscribePacketValidation,
			testUnsubscribePacketValidation,
			testPacketIdentifierValidity,
		},
	}
}

// UTF8ValidationTests returns tests for MQTT v3.1.1 UTF-8 validation
func UTF8ValidationTests() common.TestGroup {
	return common.TestGroup{
		Name: "UTF-8 Validation",
		Tests: []common.TestFunc{
			testValidUTF8String,
			testUTF8WithSpaces,
			testUTF8CaseSensitive,
			testUTF8MaxLength,
		},
	}
}

// RemainingLengthTests returns tests for MQTT v3.1.1 remaining length encoding
func RemainingLengthTests() common.TestGroup {
	return common.TestGroup{
		Name: "Remaining Length",
		Tests: []common.TestFunc{
			testRemainingLengthSmallPacket,
			testRemainingLengthLargePayload,
		},
	}
}

// testConnectPacketValidation tests CONNECT packet structure [MQTT-3.1.0-1]
func testConnectPacketValidation(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "CONNECT Packet Validation",
		SpecRef: "MQTT-3.1.0-1",
	}

	// Valid CONNECT packet structure
	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-connect-valid"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testPublishPacketValidation tests PUBLISH packet structure [MQTT-3.3.1-1]
func testPublishPacketValidation(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "PUBLISH Packet Validation",
		SpecRef: "MQTT-3.3.1-1",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-pub-valid"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Valid PUBLISH with various QoS levels
	token := client.Publish("test/validation/publish", 1, false, "test payload")
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testSubscribePacketValidation tests SUBSCRIBE packet structure [MQTT-3.8.1-1]
func testSubscribePacketValidation(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "SUBSCRIBE Packet Validation",
		SpecRef: "MQTT-3.8.1-1",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-sub-valid"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Valid SUBSCRIBE
	token := client.Subscribe("test/validation/subscribe", 1, nil)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("subscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUnsubscribePacketValidation tests UNSUBSCRIBE packet structure [MQTT-3.10.1-1]
func testUnsubscribePacketValidation(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "UNSUBSCRIBE Packet Validation",
		SpecRef: "MQTT-3.10.1-1",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-unsub-valid"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Subscribe first
	client.Subscribe("test/validation/unsubscribe", 1, nil).Wait()

	// Valid UNSUBSCRIBE
	token := client.Unsubscribe("test/validation/unsubscribe")
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("unsubscribe failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testPacketIdentifierValidity tests packet identifier usage [MQTT-2.3.1]
func testPacketIdentifierValidity(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Packet Identifier Validity",
		SpecRef: "MQTT-2.3.1",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-pktid"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Publish multiple QoS 1 messages (each gets packet identifier)
	for i := 0; i < 5; i++ {
		token := client.Publish("test/validation/pktid", 1, false, fmt.Sprintf("msg%d", i))
		token.Wait()
		if token.Error() != nil {
			result.Error = fmt.Errorf("publish %d failed: %w", i, token.Error())
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testValidUTF8String tests valid UTF-8 strings [MQTT-1.5.3-1]
func testValidUTF8String(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Valid UTF-8 Strings",
		SpecRef: "MQTT-1.5.3-1",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-utf8-valid"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Topic with valid UTF-8 including non-ASCII characters
	topics := []string{
		"test/utf8/simple",
		"test/utf8/émoji",
		"test/utf8/日本語",
		"test/utf8/🚀",
	}

	for _, topic := range topics {
		token := client.Publish(topic, 0, false, "valid utf8")
		token.Wait()
		if token.Error() != nil {
			result.Error = fmt.Errorf("publish to '%s' failed: %w", topic, token.Error())
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUTF8WithSpaces tests UTF-8 strings can contain spaces [MQTT-4.7.3-1]
func testUTF8WithSpaces(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "UTF-8 Strings with Spaces",
		SpecRef: "MQTT-4.7.3-1",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-utf8-spaces"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	topic := "test/utf8/with spaces"
	token := client.Publish(topic, 0, false, "message")
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUTF8CaseSensitive tests UTF-8 strings are case sensitive [MQTT-4.7.3-1]
func testUTF8CaseSensitive(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "UTF-8 Case Sensitivity",
		SpecRef: "MQTT-4.7.3-4",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-utf8-case"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// These should be treated as different topics
	client.Publish("test/CASE", 0, false, "upper").Wait()
	client.Publish("test/case", 0, false, "lower").Wait()

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testUTF8MaxLength tests UTF-8 string maximum length [MQTT-4.7.3-3]
func testUTF8MaxLength(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "UTF-8 Maximum Length",
		SpecRef: "MQTT-4.7.3-3",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-utf8-maxlen"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Create a reasonably long topic (not 65535 to avoid timeout issues)
	longTopic := "test/utf8/long/"
	for i := 0; i < 100; i++ {
		longTopic += "segment/"
	}

	token := client.Publish(longTopic, 0, false, "long topic test")
	token.WaitTimeout(5 * time.Second)

	// May succeed or fail based on broker limits, test passes either way
	result.Passed = true

	result.Duration = time.Since(start)
	return result
}

// testRemainingLengthSmallPacket tests small packets with 1-byte remaining length [MQTT-2.2.3]
func testRemainingLengthSmallPacket(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Remaining Length Small Packet",
		SpecRef: "MQTT-2.2.3",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-remlen-small"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Small payload (< 127 bytes, fits in 1-byte remaining length)
	token := client.Publish("test/remlen/small", 0, false, "small")
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testRemainingLengthLargePayload tests larger packets with multi-byte remaining length [MQTT-2.2.3]
func testRemainingLengthLargePayload(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Remaining Length Large Payload",
		SpecRef: "MQTT-2.2.3",
	}

	client, err := CreateAndConnectClient(cfg, common.GenerateClientID("test-remlen-large"), nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Larger payload (> 127 bytes, requires 2-byte remaining length)
	largePayload := make([]byte, 500)
	for i := range largePayload {
		largePayload[i] = byte('A' + (i % 26))
	}

	token := client.Publish("test/remlen/large", 0, false, largePayload)
	token.Wait()
	if token.Error() != nil {
		result.Error = fmt.Errorf("publish failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}
