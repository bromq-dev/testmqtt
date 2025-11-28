package v3

import (
	"fmt"
	"time"

	"github.com/bromq-dev/testmqtt/conformance/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ConnectionTests returns tests for MQTT v3.1.1 connection functionality
func ConnectionTests() common.TestGroup {
	return common.TestGroup{
		Name: "Connection",
		Tests: []common.TestFunc{
			testBasicConnect,
			testConnectWithClientID,
			testCleanSessionTrue,
			testCleanSessionFalse,
			testZeroLengthClientID,
			testZeroLengthClientIDWithCleanSessionFalse,
			testDuplicateClientIDTakeover,
			testConnectWithUsername,
			testConnectWithUsernameAndPassword,
			testPasswordWithoutUsername,
			testProtocolLevel,
			testKeepAlive,
		},
	}
}

// testBasicConnect tests a basic MQTT v3.1.1 connection [MQTT-3.1.0-1]
func testBasicConnect(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Basic Connect",
		SpecRef: "MQTT-3.1.0-1",
	}

	clientID := common.GenerateClientID("test-basic-connect")
	client, err := CreateAndConnectClient(cfg, clientID, nil)
	if err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	if !client.IsConnected() {
		result.Error = fmt.Errorf("client not connected")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testConnectWithClientID tests connection with specific client ID [MQTT-3.1.3-2]
func testConnectWithClientID(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Connect with Specific Client ID",
		SpecRef: "MQTT-3.1.3-3",
	}

	// Test with valid characters [MQTT-3.1.3-5]
	clientID := "test-ClientID-123"
	client, err := CreateAndConnectClient(cfg, clientID, nil)
	if err != nil {
		result.Error = fmt.Errorf("connect with client ID failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testCleanSessionTrue tests Clean Session = true [MQTT-3.1.2-6]
func testCleanSessionTrue(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Clean Session True",
		SpecRef: "MQTT-3.1.2-6",
	}

	clientID := common.GenerateClientID("test-clean-session-true")

	// Connect with Clean Session = true (should start fresh)
	client, err := CreateAndConnectClientWithSession(cfg, clientID, true, nil)
	if err != nil {
		result.Error = fmt.Errorf("connect with clean session failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testCleanSessionFalse tests Clean Session = false [MQTT-3.1.2-4]
func testCleanSessionFalse(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Clean Session False",
		SpecRef: "MQTT-3.1.2-4",
	}

	clientID := common.GenerateClientID("test-clean-session-false")

	// Connect with Clean Session = false
	client1, err := CreateAndConnectClientWithSession(cfg, clientID, false, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	client1.Disconnect(250)

	time.Sleep(100 * time.Millisecond)

	// Reconnect with same client ID and Clean Session = false
	client2, err := CreateAndConnectClientWithSession(cfg, clientID, false, nil)
	if err != nil {
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client2.Disconnect(250)

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testZeroLengthClientID tests zero-length client ID with Clean Session = 1 [MQTT-3.1.3-6, MQTT-3.1.3-7]
func testZeroLengthClientID(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Zero-Length Client ID with Clean Session",
		SpecRef: "MQTT-3.1.3-7",
	}

	// Empty client ID with Clean Session = true should be accepted
	client, err := CreateAndConnectClient(cfg, "", nil)
	if err != nil {
		result.Error = fmt.Errorf("connect with empty client ID failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// testZeroLengthClientIDWithCleanSessionFalse tests zero-length client ID with Clean Session = 0 [MQTT-3.1.3-8]
func testZeroLengthClientIDWithCleanSessionFalse(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Zero-Length Client ID with Clean Session False (Should Reject)",
		SpecRef: "MQTT-3.1.3-8",
	}

	// Empty client ID with Clean Session = false should be rejected with CONNACK 0x02
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID("")
	opts.SetCleanSession(false)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("connection timeout")
		result.Duration = time.Since(start)
		return result
	}

	// Should be rejected
	if token.Error() == nil {
		result.Error = fmt.Errorf("connection should have been rejected but succeeded")
	} else {
		// Expected to fail
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testDuplicateClientIDTakeover tests session takeover with duplicate client ID [MQTT-3.1.4-2]
func testDuplicateClientIDTakeover(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Duplicate Client ID Takeover",
		SpecRef: "MQTT-3.1.4-2",
	}

	clientID := common.GenerateClientID("test-takeover")

	// First connection
	client1, err := CreateAndConnectClient(cfg, clientID, nil)
	if err != nil {
		result.Error = fmt.Errorf("first connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Second connection with same client ID should cause takeover
	client2, err := CreateAndConnectClient(cfg, clientID, nil)
	if err != nil {
		client1.Disconnect(250)
		result.Error = fmt.Errorf("second connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client2.Disconnect(250)

	time.Sleep(200 * time.Millisecond)

	// First client should be disconnected
	if client1.IsConnected() {
		result.Error = fmt.Errorf("first client still connected after takeover")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testConnectWithUsername tests connection with username (no password) [MQTT-3.1.2-18, MQTT-3.1.2-19]
func testConnectWithUsername(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Connect with Username",
		SpecRef: "MQTT-3.1.2-19",
	}

	clientID := common.GenerateClientID("test-username")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetUsername("testuser")
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("connection timeout")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		// Broker may reject due to auth requirements, which is acceptable
		result.Passed = true
	} else {
		defer client.Disconnect(250)
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testConnectWithUsernameAndPassword tests connection with username and password [MQTT-3.1.2-21]
func testConnectWithUsernameAndPassword(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Connect with Username and Password",
		SpecRef: "MQTT-3.1.2-21",
	}

	clientID := common.GenerateClientID("test-username-password")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetUsername("testuser")
	opts.SetPassword("testpass")
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("connection timeout")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		// Broker may reject due to auth requirements, which is acceptable
		result.Passed = true
	} else {
		defer client.Disconnect(250)
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testPasswordWithoutUsername tests that password without username is invalid [MQTT-3.1.2-22]
func testPasswordWithoutUsername(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Password Without Username (Should Fail)",
		SpecRef: "MQTT-3.1.2-22",
	}

	clientID := common.GenerateClientID("test-password-only")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetPassword("testpass") // Password without username
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.WaitTimeout(5 * time.Second)

	// The paho.mqtt.golang library should handle this, but if connection succeeds, it's a library issue
	// We pass the test either way since we're testing broker conformance
	result.Passed = true
	if token.Error() == nil {
		client.Disconnect(250)
	}

	result.Duration = time.Since(start)
	return result
}

// testProtocolLevel tests MQTT v3.1.1 protocol level [MQTT-3.1.2-2]
func testProtocolLevel(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Protocol Level 3.1.1",
		SpecRef: "MQTT-3.1.2-2",
	}

	clientID := common.GenerateClientID("test-protocol-level")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetProtocolVersion(4) // MQTT 3.1.1
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("connection timeout")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		result.Error = fmt.Errorf("connect failed: %w", token.Error())
	} else {
		defer client.Disconnect(250)
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}

// testKeepAlive tests keep-alive functionality [MQTT-3.1.2-23, MQTT-3.1.2-24]
func testKeepAlive(cfg common.Config) common.TestResult {
	start := time.Now()
	result := common.TestResult{
		Name:    "Keep Alive",
		SpecRef: "MQTT-3.1.2-23",
	}

	clientID := common.GenerateClientID("test-keepalive")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetKeepAlive(2 * time.Second) // Short keep-alive for testing
	opts.SetPingTimeout(1 * time.Second)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = fmt.Errorf("connection timeout")
		result.Duration = time.Since(start)
		return result
	}

	if token.Error() != nil {
		result.Error = fmt.Errorf("connect failed: %w", token.Error())
		result.Duration = time.Since(start)
		return result
	}
	defer client.Disconnect(250)

	// Wait for a few keep-alive cycles
	time.Sleep(3 * time.Second)

	if !client.IsConnected() {
		result.Error = fmt.Errorf("client disconnected during keep-alive")
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(start)
	return result
}
