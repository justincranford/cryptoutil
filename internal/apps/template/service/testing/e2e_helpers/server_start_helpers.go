// Copyright (c) 2025 Justin Cranford

// Package e2e_helpers provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from cipher-im implementation to support 9-service migration.
package e2e_helpers

import (
"fmt"
"time"

cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DualPortServer defines the interface for servers that expose both public and admin port binding status.
// All cryptoutil service servers (CipherIMServer, JoseJAServer, KMSServer) implement this interface.
type DualPortServer interface {
PublicPort() int
AdminPort() int
}

// WaitForDualServerPorts waits for both public and admin server ports to bind.
// It polls at regular intervals and checks the error channel for startup failures.
//
// Returns nil when both ports are bound, or an error if the server fails to start
// or ports do not bind within the configured timeout.
func WaitForDualServerPorts(server DualPortServer, errChan <-chan error) error {
params := DefaultServerWaitParams()

for i := 0; i < params.MaxWaitAttempts; i++ {
publicPort := server.PublicPort()
adminPort := server.AdminPort()

if publicPort > 0 && adminPort > 0 {
return nil
}

select {
case err := <-errChan:
return fmt.Errorf("server start error: %w", err)
case <-time.After(params.WaitInterval):
}
}

switch {
case server.PublicPort() == 0:
return fmt.Errorf("public server did not bind to port after %d attempts", params.MaxWaitAttempts)
default:
return fmt.Errorf("admin server did not bind to port after %d attempts", params.MaxWaitAttempts)
}
}

// StartDualPortServerAsync starts a server function in a background goroutine.
// Returns an error channel for monitoring startup failures.
//
// The startFn should call server.Start() with appropriate parameters. Using a closure
// allows callers to adapt different server signatures (Start(ctx) vs Start()).
func StartDualPortServerAsync(startFn func() error) <-chan error {
errChan := make(chan error, 1)

go func() {
if err := startFn(); err != nil {
errChan <- err
}
}()

return errChan
}

// MustStartAndWaitForDualPorts starts a server in background and waits for both public
// and admin ports to bind. Panics on failure — intended for TestMain usage where
// test infrastructure setup failures should abort the entire test suite.
//
// Example usage in TestMain:
//
//server, _ := NewFromConfig(ctx, cfg)
//e2e_helpers.MustStartAndWaitForDualPorts(server, func() error {
//    return server.Start(ctx)
//})
//defer server.Shutdown(ctx)
func MustStartAndWaitForDualPorts(server DualPortServer, startFn func() error) {
errChan := StartDualPortServerAsync(startFn)
if err := WaitForDualServerPorts(server, errChan); err != nil {
panic(fmt.Sprintf("failed to start server: %v", err))
}
}

// DualPortBaseURLs returns the public and admin base URLs for a DualPortServer.
// Uses the standard format: https://127.0.0.1:<port>.
func DualPortBaseURLs(server DualPortServer) (publicURL, adminURL string) {
publicURL = fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, server.PublicPort())
adminURL = fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, server.AdminPort())

return publicURL, adminURL
}
