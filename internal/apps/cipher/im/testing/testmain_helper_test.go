// Copyright (c) 2025 Justin Cranford
//

package testing_test

import (
"context"
"net"
"strconv"
"testing"
"time"

"github.com/stretchr/testify/require"

cryptoutilAppsCipherImServerConfig "cryptoutil/internal/apps/cipher/im/server/config"
cryptoutilAppsCipherImTesting "cryptoutil/internal/apps/cipher/im/testing"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestSetupTestServer_SuccessfulSetup tests that SetupTestServer creates a fully
// configured server with all required resources populated.
// NOT parallel: uses file::memory:?cache=shared which is process-wide.
func TestSetupTestServer_SuccessfulSetup(t *testing.T) {
ctx := context.Background()

resources, err := cryptoutilAppsCipherImTesting.SetupTestServer(ctx, false)
require.NoError(t, err)
require.NotNil(t, resources)

defer resources.Shutdown(context.Background())

tests := []struct {
name   string
verify func()
}{
{name: "DB is initialized", verify: func() { require.NotNil(t, resources.DB) }},
{name: "SQLDB is initialized", verify: func() { require.NotNil(t, resources.SQLDB) }},
{name: "CipherIMServer is initialized", verify: func() { require.NotNil(t, resources.CipherIMServer) }},
{name: "BaseURL is not empty", verify: func() {
require.NotEmpty(t, resources.BaseURL)
require.Contains(t, resources.BaseURL, "https://")
}},
{name: "AdminURL is not empty", verify: func() {
require.NotEmpty(t, resources.AdminURL)
require.Contains(t, resources.AdminURL, "https://")
}},
{name: "JWKGenService is initialized", verify: func() { require.NotNil(t, resources.JWKGenService) }},
{name: "TelemetryService is initialized", verify: func() { require.NotNil(t, resources.TelemetryService) }},
{name: "TLSCfg is initialized", verify: func() { require.NotNil(t, resources.TLSCfg) }},
{name: "HTTPClient is initialized", verify: func() { require.NotNil(t, resources.HTTPClient) }},
{name: "Shutdown function is initialized", verify: func() { require.NotNil(t, resources.Shutdown) }},
{name: "PublicPort is assigned", verify: func() { require.Greater(t, resources.CipherIMServer.PublicPort(), 0) }},
{name: "AdminPort is assigned", verify: func() { require.Greater(t, resources.CipherIMServer.AdminPort(), 0) }},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
t.Parallel()
tt.verify()
})
}
}

// TestSetupTestServer_CancelledContext tests that SetupTestServer returns an error
// when given a cancelled context (covers NewFromConfig error path).
// NOT parallel: uses file::memory:?cache=shared which is process-wide.
func TestSetupTestServer_CancelledContext(t *testing.T) {
ctx, cancel := context.WithCancel(context.Background())
cancel() // Cancel immediately.

resources, err := cryptoutilAppsCipherImTesting.SetupTestServer(ctx, false)
require.Error(t, err)
require.Nil(t, resources)
}

// TestStartCipherIMService_SuccessfulStart tests that StartCipherIMService creates
// and starts a fully operational server from configuration.
// NOT parallel: uses file::memory:?cache=shared which is process-wide.
func TestStartCipherIMService_SuccessfulStart(t *testing.T) {
cfg := cryptoutilAppsCipherImServerConfig.DefaultTestConfig()

server := cryptoutilAppsCipherImTesting.StartCipherIMService(cfg)
require.NotNil(t, server)

defer func() {
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

_ = server.Shutdown(shutdownCtx)

// Close DB to destroy the in-memory database for test isolation.
if sqlDB, err := server.DB().DB(); err == nil {
_ = sqlDB.Close()
}
}()

tests := []struct {
name   string
verify func()
}{
{name: "PublicPort is assigned", verify: func() { require.Greater(t, server.PublicPort(), 0) }},
{name: "AdminPort is assigned", verify: func() { require.Greater(t, server.AdminPort(), 0) }},
{name: "DB is initialized", verify: func() { require.NotNil(t, server.DB()) }},
{name: "App is initialized", verify: func() { require.NotNil(t, server.App()) }},
{name: "JWKGen is initialized", verify: func() { require.NotNil(t, server.JWKGen()) }},
{name: "Telemetry is initialized", verify: func() { require.NotNil(t, server.Telemetry()) }},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
t.Parallel()
tt.verify()
})
}
}

// TestStartCipherIMService_NilConfig tests that StartCipherIMService panics
// when given a nil configuration (covers NewFromConfig error/panic path).
func TestStartCipherIMService_NilConfig(t *testing.T) {
t.Parallel()

require.Panics(t, func() {
cryptoutilAppsCipherImTesting.StartCipherIMService(nil)
}, "StartCipherIMService should panic with nil config")
}

// TestStartCipherIMService_PortConflict tests that StartCipherIMService panics
// when the configured port is already in use (covers errChan server start error path).
// NOT parallel: uses file::memory:?cache=shared which is process-wide.
func TestStartCipherIMService_PortConflict(t *testing.T) {
// Occupy a port to create a conflict.
lc := net.ListenConfig{}

listener, err := lc.Listen(context.Background(), "tcp", cryptoutilSharedMagic.IPv4Loopback+":0")
require.NoError(t, err)

defer func() { _ = listener.Close() }()

_, portStr, err := net.SplitHostPort(listener.Addr().String())
require.NoError(t, err)

port, err := strconv.ParseUint(portStr, 10, 16)
require.NoError(t, err)

// Create config that binds to the occupied port — server Start() will fail.
cfg := cryptoutilAppsCipherImServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, uint16(port), true)

require.Panics(t, func() {
cryptoutilAppsCipherImTesting.StartCipherIMService(cfg)
}, "StartCipherIMService should panic when port is occupied")
}
