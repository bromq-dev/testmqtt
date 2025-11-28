package conformance

import (
	"github.com/bromq-dev/testmqtt/conformance/common"
	v3 "github.com/bromq-dev/testmqtt/conformance/v3"
)

// RunV3Tests executes MQTT v3.1.1 conformance tests
func RunV3Tests(broker, username, password, tests string, verbose bool) error {
	cfg := common.Config{
		Broker:   broker,
		Username: username,
		Password: password,
	}
	return v3.RunTests(cfg, tests, verbose)
}
