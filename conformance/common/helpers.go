package common

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"time"
)

// GenerateClientID generates a unique client ID for testing
func GenerateClientID(prefix string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), n.Int64())
}

// GenerateTopicName generates a unique topic name for testing
func GenerateTopicName(prefix string) string {
	return fmt.Sprintf("%s/%d", prefix, time.Now().UnixNano())
}

// WaitTimeout is a helper that waits for a condition with timeout
func WaitTimeout(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// RandomPayload generates random payload of specified size using crypto/rand
func RandomPayload(size int) []byte {
	payload := make([]byte, size)
	rand.Read(payload)
	return payload
}

// DialBroker parses broker URL and establishes TCP connection
func DialBroker(broker string) (net.Conn, error) {
	u, err := url.Parse(broker)
	if err != nil {
		return nil, fmt.Errorf("invalid broker URL: %w", err)
	}

	host := u.Host
	if u.Port() == "" {
		host = net.JoinHostPort(u.Hostname(), "1883")
	}

	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial broker: %w", err)
	}

	return conn, nil
}

// CheckBrokerReachable verifies the broker is reachable at the TCP level
func CheckBrokerReachable(broker string) error {
	conn, err := DialBroker(broker)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
