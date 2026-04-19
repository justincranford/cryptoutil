// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e_test

import (
	"net"
	"testing"
	"time"
)

// TestE2E_AdminPortIsolation verifies that the admin port (127.0.0.1:9090) is NOT exposed
// to the host. This confirms the network isolation policy: admin endpoints are only accessible
// from inside the container network.
//
// Note: Full admin mTLS tests (happy/sad path with client certs) require connecting from inside
// the container network. Those are validated by the Docker Compose healthcheck which calls
// `/app/sm-kms livez` against 127.0.0.1:9090 — if that healthcheck passes, admin TLS works.
func TestE2E_AdminPortIsolation(t *testing.T) {
	t.Parallel()

	// Admin ports are 127.0.0.1:9090 INSIDE containers — not exposed to host.
	// We verify the host cannot reach those ports (connection refused or timeout).
	// If a port is accidentally exposed, this test catches it.
	adminAddrs := []struct {
		name string
		addr string
	}{
		// These would be the admin ports if incorrectly exposed.
		// The compose.yml does NOT publish 9090, so all should be unreachable from host.
		{"sqlite-1-admin", "127.0.0.1:9090"},
		{"sqlite-2-admin", "127.0.0.1:9091"},
		{"postgres-1-admin", "127.0.0.1:9092"},
		{"postgres-2-admin", "127.0.0.1:9093"},
	}

	for _, tt := range adminAddrs {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn, err := net.DialTimeout("tcp", tt.addr, 2*time.Second)
			if err == nil {
				_ = conn.Close()
				// If we reach here, an admin port leaked to host — fail the test.
				t.Errorf("admin port %s is unexpectedly reachable from host — admin ports must NOT be exposed", tt.addr)
			}
			// Connection refused or timeout = admin port correctly not exposed.
			// This is the expected (happy) path.
		})
	}
}
