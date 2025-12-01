package cmd

import (
	"fmt"
	"time"

	"github.com/bromq-dev/testmqtt/internal/sim"
	"github.com/spf13/cobra"
)

var (
	simVersion        string
	simSource         string
	simSourceUsername string
	simSourcePassword string
	simTopic          string
	simBroker         string
	simUsername       string
	simPassword       string
	simVerbose        bool
	simQoS            int
	simNoRetain       bool
	simQueueSize      int
	simTimeout        time.Duration
	simUnixTimestamp  bool
)

var simCmd = &cobra.Command{
	Use:   "sim",
	Short: "Simulate MQTT traffic by bridging from a source broker",
	Long: `Simulate MQTT traffic by subscribing to a source broker and republishing
all messages to a target broker. This is useful for:
- Replicating real-world traffic patterns for testing
- Load testing with production-like message flows
- Debugging broker behavior with live data

All message properties are preserved (topic, payload, QoS, retain flag,
and MQTT v5 properties).`,
	Example: `  # Bridge all traffic from test.mosquitto.org to local broker
  testmqtt sim --source tcp://test.mosquitto.org:1883 --broker tcp://localhost:1883

  # Bridge specific topic with authentication
  testmqtt sim --source tcp://source:1883 --topic "sensors/#" \
    --broker tcp://localhost:1883 --username admin --password secret

  # Use MQTT v3.1.1 instead of v5
  testmqtt sim --version 3 --source tcp://source:1883 --broker tcp://localhost:1883`,
	RunE:         runSim,
	SilenceUsage: true,
}

func init() {
	simCmd.Flags().StringVarP(&simVersion, "version", "v", "5", "MQTT version (3 or 5)")
	simCmd.Flags().StringVar(&simSource, "source", "tcp://test.mosquitto.org:1883", "Source broker URL")
	simCmd.Flags().StringVar(&simSourceUsername, "source-username", "", "Source broker username")
	simCmd.Flags().StringVar(&simSourcePassword, "source-password", "", "Source broker password")
	simCmd.Flags().StringVarP(&simTopic, "topic", "t", "#", "Topic filter to subscribe on source")
	simCmd.Flags().StringVarP(&simBroker, "broker", "b", "tcp://localhost:1883", "Target broker URL")
	simCmd.Flags().StringVarP(&simUsername, "username", "u", "", "Target broker username")
	simCmd.Flags().StringVarP(&simPassword, "password", "p", "", "Target broker password")
	simCmd.Flags().BoolVar(&simVerbose, "verbose", false, "Log each message being bridged")
	simCmd.Flags().IntVarP(&simQoS, "qos", "q", -1, "Override QoS for republishing (0, 1, 2). -1 preserves source QoS")
	simCmd.Flags().BoolVar(&simNoRetain, "no-retain", false, "Strip retain flag from republished messages")
	simCmd.Flags().IntVar(&simQueueSize, "queue-size", 1000, "Max concurrent publishes in flight")
	simCmd.Flags().DurationVar(&simTimeout, "timeout", 100*time.Millisecond, "Publish timeout (drops if exceeded)")
	simCmd.Flags().BoolVar(&simUnixTimestamp, "unix-ts", false, "Use unix timestamp instead of datetime")
}

func runSim(cmd *cobra.Command, args []string) error {
	cfg := sim.Config{
		Source:         simSource,
		SourceUsername: simSourceUsername,
		SourcePassword: simSourcePassword,
		Topic:          simTopic,
		Broker:         simBroker,
		Username:       simUsername,
		Password:       simPassword,
		Verbose:        simVerbose,
		QoS:            simQoS,
		NoRetain:       simNoRetain,
		QueueSize:      simQueueSize,
		Timeout:        simTimeout,
		UnixTimestamp:  simUnixTimestamp,
	}

	switch simVersion {
	case "5":
		return sim.RunV5(cfg)
	case "3":
		return sim.RunV3(cfg)
	default:
		return fmt.Errorf("unsupported MQTT version: %s (supported: 3, 5)", simVersion)
	}
}
