package common

import (
	"time"
)

// Config holds configuration for conformance tests
type Config struct {
	Broker   string
	Username string
	Password string
}

// TestResult represents the outcome of a conformance test
type TestResult struct {
	Name     string
	Passed   bool
	Error    error
	Duration time.Duration
	SpecRef  string // MQTT spec reference like "MQTT-3.1.0-1" (v5) or "MQTT-3.1-1" (v3.1.1)
}

// TestFunc is a function that runs a conformance test
type TestFunc func(cfg Config) TestResult

// TestGroup represents a collection of related tests
type TestGroup struct {
	Name  string
	Tests []TestFunc
}
