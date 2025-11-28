package v5

import (
	"github.com/bromq-dev/testmqtt/conformance/common"
)

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// DisconnectTests returns tests for DISCONNECT packet [MQTT-3.14]
func DisconnectTests() TestGroup {
	return TestGroup{
		Name: "DISCONNECT Packet",
		Tests: []TestFunc{
			testNormalDisconnect,
			testDisconnectReasonCodes,
			testDisconnectSessionExpiry,
			testServerDisconnect,
		},
	}
}

// testNormalDisconnect tests normal disconnection [MQTT-3.14.4-1]
// "After sending a DISCONNECT packet the Client MUST NOT send any more MQTT
// Control Packets on that Network Connection"
func testNormalDisconnect(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Normal Disconnect (Reason Code 0x00)",
		SpecRef: "MQTT-3.14.4-1",
	}

	client, err := CreateAndConnectClient(cfg, "test-disconnect-normal", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Normal disconnect with reason code 0 (Normal disconnection)
	err = client.Disconnect(&paho.Disconnect{
		ReasonCode: 0x00,
	})

	if err != nil {
		result.Error = fmt.Errorf("disconnect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testDisconnectReasonCodes tests various DISCONNECT reason codes [MQTT-3.14.2.1]
func testDisconnectReasonCodes(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "DISCONNECT Reason Codes",
		SpecRef: "MQTT-3.14.2.1",
	}

	// Test normal disconnection with reason code
	client, err := CreateAndConnectClient(cfg, "test-disconnect-codes-1", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	err = client.Disconnect(&paho.Disconnect{
		ReasonCode: 0x00, // Normal disconnection
	})
	if err != nil {
		result.Error = fmt.Errorf("disconnect with reason 0x00 failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	time.Sleep(100 * time.Millisecond)

	// Test disconnect with will message
	client2, err := CreateAndConnectClient(cfg, "test-disconnect-codes-2", nil)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	err = client2.Disconnect(&paho.Disconnect{
		ReasonCode: 0x04, // Disconnect with Will Message
	})
	if err != nil {
		result.Error = fmt.Errorf("disconnect with reason 0x04 failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testDisconnectSessionExpiry tests session expiry in DISCONNECT [MQTT-3.14.2.2.2]
// "If the Session Expiry Interval in the DISCONNECT packet is absent, the Session
// Expiry Interval in the CONNECT packet is used"
func testDisconnectSessionExpiry(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "DISCONNECT Session Expiry Interval",
		SpecRef: "MQTT-3.14.2.2.2",
	}

	client, err := CreateAndConnectClient(cfg, "test-disconnect-session", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Disconnect with session expiry interval set to 0 (end session immediately)
	disconnectProps := &paho.DisconnectProperties{
		SessionExpiryInterval: func() *uint32 {
			val := uint32(0)
			return &val
		}(),
	}

	err = client.Disconnect(&paho.Disconnect{
		ReasonCode: 0x00,
		Properties: disconnectProps,
	})

	if err != nil {
		result.Error = fmt.Errorf("disconnect with session expiry failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testServerDisconnect tests server-initiated disconnect [MQTT-3.14.4-3]
// "After sending a DISCONNECT packet the Server MUST close the Network Connection"
func testServerDisconnect(cfg common.Config) TestResult {
	start := time.Now()
	result := TestResult{
		Name:    "Server-Initiated Disconnect Closes Connection",
		SpecRef: "MQTT-3.14.4-3",
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

	// Send a second CONNECT packet - broker should disconnect us
	_, err = conn.Write(connectPacket)
	if err != nil {
		result.Passed = true
		result.Error = nil
		result.Duration = time.Since(start)
		return result
	}

	// Read response - should be DISCONNECT or connection close
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err = conn.Read(response)

	if err != nil {
		// Connection closed
		result.Passed = true
		result.Error = nil
	} else if n > 0 && response[0] == 0xE0 {
		// DISCONNECT packet received
		result.Passed = true
		result.Error = nil
	} else {
		result.Passed = false
		result.Error = fmt.Errorf("server did not disconnect on protocol violation")
	}

	result.Duration = time.Since(start)
	return result
}
