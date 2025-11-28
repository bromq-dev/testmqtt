package conformance

import (
	v5 "github.com/bromq-dev/testmqtt/conformance/v5"
)

// RunV5Tests executes MQTT v5 conformance tests
func RunV5Tests(broker string, tests string, verbose bool) error {
	return v5.RunTests(broker, tests, verbose)
}
