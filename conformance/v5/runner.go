package v5

import (
	"fmt"

	"github.com/bromq-dev/testmqtt/conformance/common"
)

// AllTestGroups returns all available test groups
func AllTestGroups() []TestGroup {
	return []TestGroup{
		// Phase 1: Core Packet Format
		RemainingLengthTests(),
		PacketValidationTests(),
		UTF8ValidationTests(),

		// Phase 2: Control Packets
		ConnectionTests(),
		PublishSubscribeTests(),
		SubscribeExtendedTests(),
		UnsubscribeTests(),
		PingTests(),
		DisconnectTests(),

		// Phase 2.5: QoS Handshake Details
		QoSHandshakeTests(),

		// QoS and Flow Control
		QoSTests(),
		FlowControlTests(),

		// Advanced Features
		TopicTests(),
		TopicAliasTests(),
		MessageExpiryTests(),
		SubscriptionIdentifierTests(),
		SharedSubscriptionTests(),
		SessionTests(),
		WillTests(),
		PropertiesTests(),
		CONNACKPropertiesTests(),

		// Error Handling
		ErrorHandlingTests(),

		// Negative Tests
		NegativeTests(),
		AdditionalNegativeTests(),
	}
}

// RunTests executes MQTT v5 conformance tests
func RunTests(broker string, filter string, verbose bool) error {
	groups := AllTestGroups()

	fmt.Printf("\n%s\n", common.TitleStyle.Render("MQTT v5.0 Conformance Tests"))
	fmt.Printf("%s\n", common.SubtitleStyle.Render(fmt.Sprintf("Broker: %s", broker)))
	if verbose {
		fmt.Printf("%s\n", common.SubtitleStyle.Render("Verbose mode: ON"))
	}
	fmt.Println()

	totalTests := 0
	passedTests := 0
	failedTests := 0
	var failedResults []TestResult

	for _, group := range groups {
		if !common.ShouldRunGroup(group.Name, filter) {
			continue
		}

		fmt.Printf("\n%s\n", common.GroupStyle.Render(group.Name))

		for _, testFunc := range group.Tests {
			result := testFunc(broker)
			totalTests++

			status := common.PassStyle.Render("✓ PASS")
			if !result.Passed {
				status = common.FailStyle.Render("✗ FAIL")
				failedTests++
				failedResults = append(failedResults, result)
			} else {
				passedTests++
			}

			specRef := ""
			if result.SpecRef != "" {
				specRef = fmt.Sprintf(" [%s]", result.SpecRef)
			}

			fmt.Printf("  %s %s%s (%v)\n", status, result.Name, specRef, result.Duration)
		}
	}

	// Detailed failure report first (if verbose and failures exist)
	if verbose && failedTests > 0 {
		fmt.Printf("\n%s\n", common.FailStyle.Render("═══ Detailed Failure Report ═══"))
		for i, result := range failedResults {
			fmt.Printf("\n%s\n", common.FailStyle.Render(fmt.Sprintf("Failure #%d: %s", i+1, result.Name)))
			fmt.Printf("  Spec Reference: %s\n", result.SpecRef)
			fmt.Printf("  Duration: %v\n", result.Duration)
			fmt.Printf("  Error: %v\n", result.Error)
		}
	}

	// Summary
	fmt.Printf("\n%s\n", common.SummaryStyle.Render("Summary"))
	fmt.Printf("  Total:  %d\n", totalTests)
	fmt.Printf("  Passed: %s\n", common.PassStyle.Render(fmt.Sprintf("%d", passedTests)))
	if failedTests > 0 {
		fmt.Printf("  Failed: %s\n", common.FailStyle.Render(fmt.Sprintf("%d", failedTests)))
	}

	if failedTests > 0 {
		return fmt.Errorf("%d test(s) failed", failedTests)
	}

	return nil
}
