package cmd

import (
	"fmt"

	"github.com/bromq-dev/testmqtt/internal/conformance"
	"github.com/spf13/cobra"
)

var (
	cfVersion string
	cfBroker  string
	cfTests   string
	cfVerbose bool
)

var conformanceCmd = &cobra.Command{
	Use:          "conformance",
	Short:        "Run MQTT conformance tests",
	Long:         `Run conformance tests to validate MQTT broker compliance with MQTT specifications`,
	RunE:         runConformance,
	SilenceUsage: true,
}

func init() {
	conformanceCmd.Flags().StringVarP(&cfVersion, "version", "v", "5", "MQTT version (3 or 5)")
	conformanceCmd.Flags().StringVarP(&cfBroker, "broker", "b", "tcp://localhost:1883", "Broker URL")
	conformanceCmd.Flags().StringVarP(&cfTests, "tests", "t", "all", "Tests to run (all, or comma-separated list)")
	conformanceCmd.Flags().BoolVar(&cfVerbose, "verbose", false, "Enable verbose output with detailed failure information")
}

func runConformance(cmd *cobra.Command, args []string) error {
	switch cfVersion {
	case "5":
		return conformance.RunV5Tests(cfBroker, cfTests, cfVerbose)
	case "3":
		return conformance.RunV3Tests(cfBroker, cfTests, cfVerbose)
	default:
		return fmt.Errorf("unsupported MQTT version: %s (supported: 3, 5)", cfVersion)
	}
}
