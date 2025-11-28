package conformance

import (
	v3 "github.com/bromq-dev/testmqtt/conformance/v3"
)

// RunV3Tests executes MQTT v3.1.1 conformance tests
func RunV3Tests(broker string, tests string, verbose bool) error {
	return v3.RunTests(broker, tests, verbose)
}
