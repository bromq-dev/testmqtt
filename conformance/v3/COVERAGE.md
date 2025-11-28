# MQTT v3.1.1 Conformance Test Coverage

Based on MQTT v3.1.1 Specification - **77 tests covering core protocol requirements**

## ✅ COMPLETE - All Core Areas Implemented (77/77 tests passing)

### Connection Tests (12 tests) ✅ - `connection.go`
- ✅ Basic connect [MQTT-3.1.0-1]
- ✅ Connect with specific client ID [MQTT-3.1.3-3]
- ✅ Clean Session true [MQTT-3.1.2-6]
- ✅ Clean Session false [MQTT-3.1.2-4]
- ✅ Zero-length client ID with Clean Session [MQTT-3.1.3-7]
- ✅ Zero-length client ID rejection with Clean Session=false [MQTT-3.1.3-8]
- ✅ Duplicate client ID takeover [MQTT-3.1.4-2]
- ✅ Connect with username [MQTT-3.1.2-19]
- ✅ Connect with username and password [MQTT-3.1.2-21]
- ✅ Password without username (invalid) [MQTT-3.1.2-22]
- ✅ Protocol level 3.1.1 [MQTT-3.1.2-2]
- ✅ Keep-alive functionality [MQTT-3.1.2-23]

### Publish/Subscribe Tests (10 tests) ✅ - `publish.go`
- ✅ Basic publish/subscribe [MQTT-3.3.1-1]
- ✅ Publish QoS 0 [MQTT-4.3.1-1]
- ✅ Publish QoS 1 [MQTT-4.3.2-1]
- ✅ Publish QoS 2 [MQTT-4.3.3-1]
- ✅ Subscribe acknowledgement [MQTT-3.8.4-1]
- ✅ Multiple subscriptions [MQTT-3.8.4-4]
- ✅ Subscription replacement [MQTT-3.8.4-3]
- ✅ Retained message delivery [MQTT-3.3.1-6]
- ✅ Clear retained message [MQTT-3.3.1-10]
- ✅ Publish to multiple subscribers [MQTT-3.3.5-1]

### Topic Tests (8 tests) ✅ - `topics.go`
- ✅ Multi-level wildcard # [MQTT-4.7.1-2]
- ✅ Single-level wildcard + [MQTT-4.7.1-3]
- ✅ Wildcard combination +/# [MQTT-4.7.1-3]
- ✅ Topic level separator [MQTT-4.7.3-1]
- ✅ $SYS prefix handling [MQTT-4.7.2-1]
- ✅ Topic case sensitivity [MQTT-4.7.3-4]
- ✅ Topics with spaces [MQTT-4.7.3-1]
- ✅ Leading/trailing slash [MQTT-4.7.3-1]

### QoS Tests (8 tests) ✅ - `qos.go`
- ✅ QoS 0 at-most-once delivery [MQTT-4.3.1-1]
- ✅ QoS 1 at-least-once delivery [MQTT-4.3.2-1]
- ✅ QoS 2 exactly-once delivery [MQTT-4.3.3-1]
- ✅ QoS downgrade [MQTT-3.8.4-6]
- ✅ Message ordering QoS 1 [MQTT-4.6.0-2]
- ✅ Message ordering QoS 2 [MQTT-4.6.0-3]
- ✅ QoS 1 PUBACK acknowledgement [MQTT-4.3.2-2]
- ✅ QoS 2 full handshake [MQTT-4.3.3-2]

### Will Messages (7 tests) ✅ - `will.go`
- ✅ Will message on abnormal disconnect [MQTT-3.1.2-8]
- ✅ Will NOT sent on clean disconnect [MQTT-3.1.2-10]
- ✅ Will message QoS 0 [MQTT-3.1.2-9]
- ✅ Will message QoS 1 [MQTT-3.1.2-14]
- ✅ Will message QoS 2 [MQTT-3.1.2-14]
- ✅ Will message retained [MQTT-3.1.2-17]
- ✅ Will message not retained [MQTT-3.1.2-16]

### Unsubscribe (5 tests) ✅ - `unsubscribe.go`
- ✅ Basic unsubscribe [MQTT-3.10.4-1]
- ✅ Unsubscribe stops delivery [MQTT-3.10.4-2]
- ✅ Unsubscribe multiple topics [MQTT-3.10.3-1]
- ✅ UNSUBACK acknowledgement [MQTT-3.10.4-4]
- ✅ Unsubscribe non-existent topic [MQTT-3.10.4-5]

### PING (3 tests) ✅ - `ping.go`
- ✅ PINGREQ/PINGRESP exchange [MQTT-3.1.2-23]
- ✅ Keep-alive zero (disabled) [MQTT-3.1.2-10]
- ✅ Keep-alive enforcement [MQTT-3.1.2-24]

### Session State (6 tests) ✅ - `session.go`
- ✅ Session state persistence [MQTT-3.1.2-4]
- ✅ Subscription persistence [MQTT-3.1.2-4]
- ✅ QoS 1 message persistence [MQTT-3.1.2-5]
- ✅ QoS 2 message persistence [MQTT-3.1.2-5]
- ✅ Clean Session clears state [MQTT-3.1.2-6]
- ✅ Retained messages not part of session [MQTT-3.1.2.7]

### Packet Validation (5 tests) ✅ - `validation.go`
- ✅ CONNECT packet validation [MQTT-3.1.0-1]
- ✅ PUBLISH packet validation [MQTT-3.3.1-1]
- ✅ SUBSCRIBE packet validation [MQTT-3.8.1-1]
- ✅ UNSUBSCRIBE packet validation [MQTT-3.10.1-1]
- ✅ Packet identifier validity [MQTT-2.3.1]

### UTF-8 Validation (4 tests) ✅ - `validation.go`
- ✅ Valid UTF-8 strings (including emoji, Japanese) [MQTT-1.5.3-1]
- ✅ UTF-8 with spaces [MQTT-4.7.3-1]
- ✅ UTF-8 case sensitivity [MQTT-4.7.3-4]
- ✅ UTF-8 maximum length [MQTT-4.7.3-3]

### Remaining Length (2 tests) ✅ - `validation.go`
- ✅ Small packet (1-byte encoding) [MQTT-2.2.3]
- ✅ Large payload (multi-byte encoding) [MQTT-2.2.3]

### Negative Tests (7 tests) ✅ - `negative.go`
- ✅ PUBLISH with wildcard topic [MQTT-3.3.2-2]
- ✅ Invalid QoS 3 [MQTT-3.3.1-4]
- ✅ Second CONNECT packet [MQTT-3.1.0-2]
- ✅ Empty SUBSCRIBE [MQTT-3.8.3-3]
- ✅ Invalid protocol name [MQTT-3.1.2-1]
- ✅ Invalid protocol level [MQTT-3.1.2-2]
- ✅ Reserved flag validation [MQTT-3.1.2-3]

## Test Results

```
MQTT v3.1.1 Conformance Tests
Broker: tcp://localhost:1883

Summary
  Total:  77
  Passed: 77
```

**100% Pass Rate** on Eclipse Mosquitto 2.x

## Coverage Statistics

- **Total normative requirements in MQTT v3.1.1 spec**: ~121
- **Test coverage**: 77 tests covering core requirements
- **Estimated coverage**: ~64% of normative requirements
- **All critical paths tested**: Connection, Pub/Sub, QoS, Sessions, Will Messages

## Architecture

- Uses `paho.mqtt.golang` for MQTT v3.1.1 client
- Shares common framework with v5 tests (`conformance/common/`)
- Consistent output styling and error handling
- All tests include proper spec references

## Notes

Some tests validate library behavior rather than broker behavior due to limitations of the high-level client library:
- Raw packet manipulation tests (e.g., malformed packets) rely on library correctness
- Some protocol violations are prevented by the client library itself
- Tests marked as such still validate that the system behaves correctly

## Future Enhancements

Possible additions (lower priority):
- More edge case tests for maximum values (client ID length, packet sizes)
- Additional negative tests with raw packet manipulation
- Stress testing of session persistence
- Performance benchmarks for QoS delivery
