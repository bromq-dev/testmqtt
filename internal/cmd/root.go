package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "testmqtt",
	Short: "A comprehensive MQTT broker testing tool",
	Long: `testmqtt is a comprehensive MQTT broker testing tool that provides:
- Conformance testing for MQTT 3.1.1 and MQTT 5.0
- Performance benchmarking
- Stress testing
- Traffic simulation (bridge messages between brokers)`,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(conformanceCmd)
	rootCmd.AddCommand(performanceCmd)
	rootCmd.AddCommand(simCmd)
}
