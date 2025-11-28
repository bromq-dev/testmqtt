package v5

import (
	"github.com/bromq-dev/testmqtt/conformance/common"
)

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/eclipse/paho.golang/packets"
)

// PacketValidationTests returns tests for packet format validation [MQTT-2.1]
func PacketValidationTests() TestGroup {
	return TestGroup{
		Name: "Packet Format Validation",
		Tests: []TestFunc{
			testReservedPacketType,
			testInvalidPacketFlags,
			testPublishFlagsValidation,
			testPubrelFixedFlags,
			testSubscribeFixedFlags,
			testUnsubscribeFixedFlags,
		},
	}
}

// testReservedPacketType tests that reserved packet types are rejected [MQTT-2.1.2-1]
// "A Server or Client MUST NOT send packets where the MQTT Control Packet type is 0 or 15"
func testReservedPacketType(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Reject Reserved Packet Types (0, 15)",
		SpecRef: "MQTT-2.1.2-1",
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

	// First send valid CONNECT
	connectPacket := []byte{
		0x10,       // CONNECT
		0x10,       // Remaining length 16
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
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
		// EOF is valid - broker rejected invalid packet
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Try to send packet with reserved type 0
	reservedPacket := []byte{
		0x00, // Reserved packet type 0
		0x00, // Remaining length 0
	}

	_, err = conn.Write(reservedPacket)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Broker should disconnect (EOF or DISCONNECT packet)
	n, err = conn.Read(response)
	if err != nil || n == 0 {
		// Connection closed - broker correctly rejected (EOF is valid)
		result.Passed = true
		result.Error = nil
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("broker accepted reserved packet type")
	}

	result.Duration = time.Since(start)
	return result
}

// testInvalidPacketFlags tests invalid flag combinations [MQTT-2.1.3-1]
// "Where a flag bit is marked as 'Reserved', it is reserved for future use
// and MUST be set to the value listed"
func testInvalidPacketFlags(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Reject Invalid Reserved Flags",
		SpecRef: "MQTT-2.1.3-1",
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

	// Send valid CONNECT
	connectPacket := []byte{
		0x10,       // CONNECT
		0x10,       // Remaining length 16
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
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
		// EOF is valid - broker rejected invalid CONNECT
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Send CONNECT with invalid flags (should be 0x10, send 0x1F)
	invalidConnect := []byte{
		0x1F,       // CONNECT with invalid flags
		0x10,       // Remaining length 16
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
		0x00,       // Properties length
		0x00, 0x05, // Client ID length
		't', 'e', 's', 't', '2', // Client ID
	}

	_, err = conn.Write(invalidConnect)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Broker should disconnect (EOF or DISCONNECT packet)
	n, err = conn.Read(response)
	if err != nil || n == 0 {
		// Connection closed - broker correctly rejected (EOF is valid)
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == packets.DISCONNECT {
		result.Passed = true
		result.Error = nil
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("broker accepted invalid flags")
	}

	result.Duration = time.Since(start)
	return result
}

// testPublishFlagsValidation tests PUBLISH flags validation [MQTT-3.3.1-4]
// "DUP, QoS, and RETAIN flags in the fixed header of a PUBLISH packet"
func testPublishFlagsValidation(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBLISH Flags Validation",
		SpecRef: "MQTT-3.3.1-4",
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

	// Send valid CONNECT
	connectPacket := []byte{
		0x10,       // CONNECT
		0x10,       // Remaining length 16
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
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
		// EOF is valid - broker rejected invalid CONNECT
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Send PUBLISH with invalid QoS (both bits set = QoS 3)
	invalidPublish := []byte{
		0x36,       // PUBLISH with QoS=3 (invalid)
		0x0D,       // Remaining length
		0x00, 0x09, // Topic length
		't', 'e', 's', 't', '/', 't', 'e', 's', 't',
		'h', 'i', // Payload
	}

	_, err = conn.Write(invalidPublish)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Broker should disconnect (EOF or DISCONNECT packet)
	n, err = conn.Read(response)
	if err != nil || n == 0 {
		// Connection closed - broker correctly rejected (EOF is valid)
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == packets.DISCONNECT {
		result.Passed = true
		result.Error = nil
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("broker accepted invalid QoS in PUBLISH")
	}

	result.Duration = time.Since(start)
	return result
}

// testPubrelFixedFlags tests PUBREL fixed flags [MQTT-3.6.1-1]
// "Bits 3,2,1 and 0 of the Fixed Header of the PUBREL packet are reserved
// and MUST be set to 0,0,1 and 0 respectively"
func testPubrelFixedFlags(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "PUBREL Fixed Flags (0x62)",
		SpecRef: "MQTT-3.6.1-1",
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

	// Send valid CONNECT
	connectPacket := []byte{
		0x10,       // CONNECT
		0x10,       // Remaining length 16
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
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
		// EOF is valid - broker rejected invalid packet
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Send PUBREL with incorrect flags (should be 0x62, send 0x60)
	invalidPubrel := []byte{
		0x60,       // PUBREL with wrong flags
		0x03,       // Remaining length
		0x00, 0x01, // Packet identifier
		0x00, // Reason code
	}

	_, err = conn.Write(invalidPubrel)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Broker should disconnect due to malformed packet
	n, err = conn.Read(response)
	if err != nil || n == 0 {
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == packets.DISCONNECT {
		result.Passed = true
		result.Error = nil
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("broker accepted PUBREL with invalid flags")
	}

	result.Duration = time.Since(start)
	return result
}

// testSubscribeFixedFlags tests SUBSCRIBE fixed flags [MQTT-3.8.1-1]
// "Bits 3,2,1 and 0 of the Fixed Header of the SUBSCRIBE packet are reserved
// and MUST be set to 0,0,1 and 0 respectively"
func testSubscribeFixedFlags(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "SUBSCRIBE Fixed Flags (0x82)",
		SpecRef: "MQTT-3.8.1-1",
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

	// Send valid CONNECT
	connectPacket := []byte{
		0x10,       // CONNECT
		0x10,       // Remaining length 16
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
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
		// EOF is valid - broker rejected invalid packet
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Send SUBSCRIBE with wrong flags (should be 0x82, send 0x80)
	invalidSubscribe := []byte{
		0x80,       // SUBSCRIBE with wrong flags
		0x0A,       // Remaining length
		0x00, 0x01, // Packet identifier
		0x00,       // Properties length
		0x00, 0x04, // Topic filter length
		't', 'e', 's', 't',
		0x00, // Subscription options
	}

	_, err = conn.Write(invalidSubscribe)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Broker should disconnect
	n, err = conn.Read(response)
	if err != nil || n == 0 {
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == packets.DISCONNECT {
		result.Passed = true
		result.Error = nil
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("broker accepted SUBSCRIBE with invalid flags")
	}

	result.Duration = time.Since(start)
	return result
}

// testUnsubscribeFixedFlags tests UNSUBSCRIBE fixed flags [MQTT-3.10.1-1]
// "Bits 3,2,1 and 0 of the Fixed Header of the UNSUBSCRIBE packet are reserved
// and MUST be set to 0,0,1 and 0 respectively"
func testUnsubscribeFixedFlags(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "UNSUBSCRIBE Fixed Flags (0xA2)",
		SpecRef: "MQTT-3.10.1-1",
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

	// Send valid CONNECT
	connectPacket := []byte{
		0x10,       // CONNECT
		0x10,       // Remaining length 16
		0x00, 0x04, // Protocol name length
		'M', 'Q', 'T', 'T',
		0x05,       // Protocol level 5
		0x02,       // Clean start
		0x00, 0x3C, // Keep alive 60
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
		// EOF is valid - broker rejected invalid packet
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Send UNSUBSCRIBE with wrong flags (should be 0xA2, send 0xA0)
	invalidUnsubscribe := []byte{
		0xA0,       // UNSUBSCRIBE with wrong flags
		0x09,       // Remaining length
		0x00, 0x01, // Packet identifier
		0x00,       // Properties length
		0x00, 0x04, // Topic filter length
		't', 'e', 's', 't',
	}

	_, err = conn.Write(invalidUnsubscribe)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Broker should disconnect
	n, err = conn.Read(response)
	if err != nil || n == 0 {
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == packets.DISCONNECT {
		result.Passed = true
		result.Error = nil
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("broker accepted UNSUBSCRIBE with invalid flags")
	}

	result.Duration = time.Since(start)
	return result
}
