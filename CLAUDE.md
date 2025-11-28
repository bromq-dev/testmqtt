# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

testmqtt is a comprehensive MQTT broker testing tool built in Go. It provides three main testing capabilities:
1. **Conformance Testing**: Validates MQTT broker compliance with MQTT 3.1.1 and MQTT 5.0 specifications
2. **Performance Benchmarking**: One-off performance measurements
3. **Stress Testing**: Load testing with configurable publishers, subscribers, and sustained duration

## Build and Run

Build the project:
```bash
go build -o bin/testmqtt .
```

The binary is built to the `bin/` directory (which is untracked in git).

Install globally:
```bash
go install github.com/bromq-dev/testmqtt@latest
```

Run the tool:
```bash
# Conformance tests
./bin/testmqtt conformance -version 5 -broker tcp://localhost:1883 -tests all

# Conformance tests with authentication
./bin/testmqtt conformance -version 5 -broker tcp://localhost:1883 -u myuser -p mypass

# Performance stress test
./bin/testmqtt performance stress -broker tcp://localhost:1883 -duration 60s -publishers 100 -subscribers 10 -topics 10 -qos 1

# Performance benchmark (one-off)
./bin/testmqtt performance bench -broker tcp://localhost:1883 -messages 10000 -payload-size 256 -qos 0

# Performance rounds (increasing load)
./bin/testmqtt performance round -broker tcp://localhost:1883 -rounds 10 -increment 100
```

Local test broker:
```bash
docker compose up -d     # Start Eclipse Mosquitto on port 1883
docker compose down      # Stop the broker
```

## Architecture

### Directory Structure

- `main.go`: Application entry point
- `internal/cmd/`: CLI command implementations (cobra)
- `internal/conformance/`: Conformance test runners (delegates to conformance/ packages)
- `conformance/`: MQTT conformance test suites
  - `v5/`: MQTT 5.0 conformance tests - organized by category:
    - **Phase 1: Core Packet Format**
      - `remaining_length.go`: Remaining length encoding tests (1-4 bytes)
      - `packet_validation.go`: Fixed header, flags, reserved packet types
      - `utf8_validation.go`: UTF-8 string validation tests
    - **Phase 2: Control Packets**
      - `connection.go`: CONNECT/CONNACK packet tests
      - `publish.go`: PUBLISH/SUBSCRIBE packet tests
      - `unsubscribe.go`: UNSUBSCRIBE/UNSUBACK tests
      - `ping.go`: PINGREQ/PINGRESP and keep-alive tests
      - `disconnect.go`: DISCONNECT packet and reason codes
    - **Advanced Features**
      - `qos.go`: QoS 0/1/2 delivery and flow control
      - `topics.go`: Topic wildcards and validation
      - `session.go`: Session management and persistence
      - `will.go`: Will message tests
      - `properties.go`: MQTT v5 properties tests
    - **Error Handling**
      - `negative.go`: Protocol violations and malformed packets
    - **Framework**
      - `runner.go`: Test execution framework with styled output
      - `helpers.go`: v5-specific test helpers
      - `types.go`: Shared test types
      - `TODO.md`: Comprehensive test coverage plan (~150-200 tests needed)
  - `v3/`: MQTT 3.1.1 conformance tests (TODO)
  - `common/`: Shared helpers between v3 and v5 (DRY utilities)
- `performance/`: Performance testing modules
  - `bench/`: One-off benchmark tests (TODO)
  - `stress/`: Load/stress testing (TODO)
  - `round/`: Multi-round incremental load tests (TODO)
- `spec/`: MQTT specification documents (v3.1.1 and v5.0)

### Key Technologies

- **CLI Framework**: cobra (commands) + viper (configuration)
- **Terminal Output**: bubbletea + gum + lipgloss for enhanced terminal output (progress bars, status displays, formatted results)
- **MQTT Clients**:
  - `github.com/eclipse/paho.mqtt.golang`: MQTT 3.1.1
  - `github.com/eclipse/paho.golang`: MQTT 5.0

### Implementation Notes

Most modules are currently stubs/placeholders (marked with TODO files). When implementing:
- Conformance tests should reference specifications in `spec/` directory
- Use appropriate Paho client based on MQTT version being tested
- Use bubbletea/gum/lipgloss for fancy terminal output (progress, status updates) during test execution
- CLI commands follow cobra conventions with flag-based configuration
