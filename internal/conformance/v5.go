package conformance

import (
	"github.com/bromq-dev/testmqtt/conformance/common"
	v5 "github.com/bromq-dev/testmqtt/conformance/v5"
)

// RunV5Tests executes MQTT v5 conformance tests
func RunV5Tests(broker, username, password, tests string, verbose bool) error {
	cfg := common.Config{
		Broker:   broker,
		Username: username,
		Password: password,
	}
	return v5.RunTests(cfg, tests, verbose)
}
