// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package ja

import (
"bytes"
"strings"
"sync"
"syscall"
"testing"
"time"

"github.com/stretchr/testify/require"

cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestJA_SubcommandHelpFlags(t *testing.T) {
t.Parallel()

tests := []struct {
subcommand string
helpTexts  []string
}{
{subcommand: "client", helpTexts: []string{"jose ja client", "Run client operations"}},
{subcommand: "init", helpTexts: []string{"jose ja init", "Initialize database schema"}},
{subcommand: "health", helpTexts: []string{"jose ja health", "Check service health"}},
{subcommand: "livez", helpTexts: []string{"jose ja livez", "Check service liveness"}},
{subcommand: "readyz", helpTexts: []string{"jose ja readyz", "Check service readiness"}},
{subcommand: "shutdown", helpTexts: []string{"jose ja shutdown", "Trigger graceful shutdown"}},
}

for _, tc := range tests {
t.Run(tc.subcommand, func(t *testing.T) {
t.Parallel()

for _, flag := range []string{"--help", "-h", "help"} {
var stdout, stderr bytes.Buffer

exitCode := Ja([]string{tc.subcommand, flag}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode, "%s %s should succeed", tc.subcommand, flag)

combined := stdout.String() + stderr.String()
for _, expected := range tc.helpTexts {
require.Contains(t, combined, expected, "%s output should contain: %s", flag, expected)
}
}
})
}
}

func TestJA_MainHelp(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ja([]string{"--help"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "jose ja")
}

func TestJA_Version(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ja([]string{"version"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)
}

func TestJA_SubcommandNotImplemented(t *testing.T) {
t.Parallel()

tests := []struct {
subcommand string
errorText  string
}{
{subcommand: "client", errorText: "not yet implemented"},
{subcommand: "init", errorText: "not yet implemented"},
}

for _, tc := range tests {
t.Run(tc.subcommand, func(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ja([]string{tc.subcommand}, nil, &stdout, &stderr)
require.Equal(t, 1, exitCode, "%s should exit with 1", tc.subcommand)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, tc.errorText)
})
}
}

func TestJA_ServerHelp(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ja([]string{"server", "--help"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "jose ja server")
}

func TestJA_UnknownSubcommand(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ja([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Unknown subcommand")
}

// TestJA_ServerParseError verifies the Parse error path in jaServerStart.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestJA_ServerParseError(t *testing.T) {
var stdout, stderr bytes.Buffer

//nolint:goconst // Test-specific invalid flag, not a magic string.
exitCode := jaServerStart([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Failed to parse configuration")
}

// TestJA_ServerCreateError verifies the NewFromConfig error path.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestJA_ServerCreateError(t *testing.T) {
var stdout, stderr bytes.Buffer

// Server creation fails because PostgreSQL is not running on the default port.
exitCode := jaServerStart([]string{}, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Failed to create server")
}

// TestJA_ServerLifecycle tests the full server start → signal → shutdown lifecycle.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestJA_ServerLifecycle(t *testing.T) {
var mu sync.Mutex

stdout := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}
stderr := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}

exitCodeChan := make(chan int, 1)

go func() {
exitCodeChan <- jaServerStart([]string{
"--dev",
"--database-url", cryptoutilSharedMagic.SQLiteInMemoryDSN,
}, stdout, stderr)
}()

// Wait for server to start (goroutine prints "Starting jose-ja service" to stdout).
const maxWait = 30 * time.Second

startTime := time.Now()

for time.Since(startTime) < maxWait {
mu.Lock()

output := stdout.buf.String()

mu.Unlock()

if strings.Contains(output, "Starting jose-ja service") {
break
}

time.Sleep(100 * time.Millisecond) //nolint:mnd // Polling interval for server startup.
}

// Give server a moment to fully bind listeners and register signal handler.
time.Sleep(500 * time.Millisecond) //nolint:mnd // Wait for server to finish binding.

// Send SIGTERM to trigger graceful shutdown via the signal handler.
err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
require.NoError(t, err)

// Wait for jaServerStart to return.
select {
case exitCode := <-exitCodeChan:
require.Equal(t, 0, exitCode, "Server should exit cleanly after SIGTERM")
case <-time.After(maxWait):
t.Fatal("Server did not shut down within timeout")
}
}

// TestJA_ServerStartError tests the errChan error path when the server fails
// to start (e.g., address already in use from the previous lifecycle test).
// Sequential: uses viper global state via ParseWithFlagSet.
func TestJA_ServerStartError(t *testing.T) {
	// Port 8800 is still bound by the server from TestJA_ServerLifecycle
	// (production code does not call srv.Shutdown after SIGTERM signal).
	// This means Start() will fail with "address already in use", triggering
	// the errChan error path.
	var mu sync.Mutex

	stdout := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}
	stderr := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}

	exitCodeChan := make(chan int, 1)

	go func() {
		exitCodeChan <- jaServerStart([]string{
			"--dev",
			"--database-url", cryptoutilSharedMagic.SQLiteInMemoryDSN,
		}, stdout, stderr)
	}()

	// Wait for jaServerStart to return (should fail due to address in use).
	select {
	case exitCode := <-exitCodeChan:
		require.Equal(t, 1, exitCode, "Server should fail when address is already in use")
	case <-time.After(30 * time.Second): //nolint:mnd // Generous timeout for server start failure.
		t.Fatal("Server did not return within timeout")
	}

	mu.Lock()
	defer mu.Unlock()

	output := stdout.buf.String() + stderr.buf.String()
	require.Contains(t, output, "Server error")
}

// syncWriter is a thread-safe io.Writer that wraps a bytes.Buffer.
type syncWriter struct {
	buf *bytes.Buffer
	mu  *sync.Mutex
}

func (w *syncWriter) Write(p []byte) (int, error) {
w.mu.Lock()
defer w.mu.Unlock()

return w.buf.Write(p)
}
