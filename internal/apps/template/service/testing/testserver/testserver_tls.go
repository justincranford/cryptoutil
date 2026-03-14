// Copyright (c) 2025 Justin Cranford
//
//

package testserver

import (
"crypto/tls"
"crypto/x509"
"testing"

cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
)

// TestTLSBundle holds the root CA pool needed for test HTTP client TLS configuration.
// It enables test HTTP clients to validate server certificates without InsecureSkipVerify.
type TestTLSBundle struct {
rootCAPool *x509.CertPool
}

// NewTestTLSBundle creates a TestTLSBundle from a running ServiceServer's root CA pool.
// Must be called after testserver.StartAndWait to ensure the server has started and its
// TLS chain is available.
func NewTestTLSBundle(t testing.TB, srv cryptoutilAppsTemplateServiceServer.ServiceServer) *TestTLSBundle {
t.Helper()

if srv == nil {
t.Fatalf("server cannot be nil for NewTestTLSBundle")

return nil
}

pool := srv.TLSRootCAPool()
if pool == nil {
t.Fatalf("server.TLSRootCAPool() returned nil: server may not have started or uses a mock public server")

return nil
}

return &TestTLSBundle{rootCAPool: pool}
}

// TLSClientConfig returns a *tls.Config configured to trust the bundle's root CA certificate.
// Use this instead of InsecureSkipVerify: true to securely validate server certificates in tests.
func TLSClientConfig(t testing.TB, bundle *TestTLSBundle) *tls.Config {
t.Helper()

if bundle == nil {
t.Fatalf("TLS bundle cannot be nil for TLSClientConfig")

return nil
}

return &tls.Config{
MinVersion: tls.VersionTLS13,
RootCAs:    bundle.rootCAPool,
}
}
