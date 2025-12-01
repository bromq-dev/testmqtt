package sim

import "time"

// Config holds the configuration for the MQTT traffic simulator
type Config struct {
	Source         string
	SourceUsername string
	SourcePassword string
	Topic          string
	Broker         string
	Username       string
	Password       string
	Verbose        bool
	QoS            int           // -1 to preserve source QoS, 0-2 to override
	NoRetain       bool          // Strip retain flag from republished messages
	QueueSize      int           // Max concurrent publishes
	Timeout        time.Duration // Publish timeout
	UnixTimestamp  bool          // Use unix timestamp instead of datetime
}
