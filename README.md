# testmqtt

A comprehensive MQTT broker testing tool built in Go.

## Features

- **Conformance Testing**: Validate MQTT broker compliance with specifications
  - MQTT v3.1.1: 77 tests covering all core protocol features ✓
  - MQTT v5.0: 139 tests covering advanced features ✓
- **Performance Benchmarking**: One-off performance measurements
- **Stress Testing**: Load testing with configurable publishers, subscribers, and duration

## Quick Start

### Build

```bash
go build -o bin/testmqtt ./cmd/testmqtt
```

### Run Conformance Tests

```bash
# MQTT v3.1.1 conformance tests (77 tests)
./bin/testmqtt conformance --version 3 --broker tcp://localhost:1883

# MQTT v5.0 conformance tests (139 tests)
./bin/testmqtt conformance --version 5 --broker tcp://localhost:1883

# Run specific test groups
./bin/testmqtt conformance --version 3 --broker tcp://localhost:1883 --tests Connection,QoS

# Verbose output with detailed failure information
./bin/testmqtt conformance --version 3 --broker tcp://localhost:1883 --verbose
```

### Performance Testing

```bash
# Stress test
./bin/testmqtt performance stress --broker tcp://localhost:1883 --duration 60s --publishers 100 --subscribers 10 --topics 10 --qos 1

# One-off benchmark
./bin/testmqtt performance bench --broker tcp://localhost:1883 --messages 10000 --payload-size 256 --qos 0

# Multiple rounds with increasing load
./bin/testmqtt performance round --broker tcp://localhost:1883 --rounds 10 --increment 100
```

### Test with Local Broker

```bash
# Start Eclipse Mosquitto
docker compose up -d

# Run tests
./bin/testmqtt conformance --version 3 --broker tcp://localhost:1883

# Stop broker
docker compose down
```

## Conformance Test Coverage

### MQTT v3.1.1 (77 tests)
- Connection (12): Basic connect, clean session, client ID handling, authentication
- Publish/Subscribe (10): QoS 0/1/2, retained messages, multiple subscribers
- Topics (8): Wildcards (#, +), $SYS prefix, case sensitivity
- QoS (8): Delivery guarantees, message ordering, acknowledgements
- Will Messages (7): Abnormal disconnect, QoS levels, retained
- Unsubscribe (5): Stop delivery, acknowledgements
- PING (3): Keep-alive, heartbeat
- Session State (6): Persistence, clean session
- Packet Validation (5): CONNECT, PUBLISH, SUBSCRIBE structure
- UTF-8 Validation (4): Valid strings, encoding
- Remaining Length (2): Packet size encoding
- Negative Tests (7): Protocol violations

### MQTT v5.0 (139 tests)
- Core packet format validation
- All control packets (CONNECT, PUBLISH, SUBSCRIBE, etc.)
- QoS handshakes and flow control
- Advanced features (topic aliases, message expiry, subscription identifiers)
- Properties and user properties
- Enhanced authentication
- Error handling and negative tests

See `conformance/v3/COVERAGE.md` and `conformance/v5/TODO.md` for detailed coverage.

## Architecture

```
testmqtt/
├── cmd/testmqtt/          # Main application entry
├── internal/
│   ├── cmd/               # CLI commands (cobra)
│   └── conformance/       # Test runners
├── conformance/
│   ├── common/            # Shared test framework
│   ├── v3/                # MQTT v3.1.1 tests (77 tests)
│   └── v5/                # MQTT v5.0 tests (139 tests)
├── performance/           # Performance testing (TODO)
└── spec/                  # MQTT specifications
```

## Technologies Used

- Go 1.21+
- `github.com/eclipse/paho.mqtt.golang` - MQTT v3.1.1 client
- `github.com/eclipse/paho.golang` - MQTT v5.0 client
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
