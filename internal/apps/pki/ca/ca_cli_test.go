// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package ca

import (
"bytes"
"sync"
"syscall"
"testing"
"time"

"github.com/stretchr/testify/require"

cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"strings"
)

func TestCA_SubcommandHelpFlags(t *testing.T) {
t.Parallel()

tests := []struct {
subcommand string
helpTexts  []string
}{
{subcommand: "client", helpTexts: []string{"pki ca client", "Run client operations"}},
{subcommand: "init", helpTexts: []string{"pki ca init", "Initialize database schema"}},
{subcommand: "health", helpTexts: []string{"pki ca health", "Check service health"}},
{subcommand: "livez", helpTexts: []string{"pki ca livez", "Check service liveness"}},
{subcommand: "readyz", helpTexts: []string{"pki ca readyz", "Check service readiness"}},
{subcommand: "shutdown", helpTexts: []string{"pki ca shutdown", "Trigger graceful shutdown"}},
}

for _, tc := range tests {
t.Run(tc.subcommand, func(t *testing.T) {
t.Parallel()

for _, flag := range []string{"--help", "-h", "help"} {
var stdout, stderr bytes.Buffer

exitCode := Ca([]string{tc.subcommand, flag}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode, "%s %s should succeed", tc.subcommand, flag)

combined := stdout.String() + stderr.String()
for _, expected := range tc.helpTexts {
require.Contains(t, combined, expected, "%s output should contain: %s", flag, expected)
}
}
})
}
}

func TestCA_MainHelp(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"--help"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "pki ca")
}

func TestCA_Version(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"version"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)
}

func TestCA_SubcommandNotImplemented(t *testing.T) {
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

exitCode := Ca([]string{tc.subcommand}, nil, &stdout, &stderr)
require.Equal(t, 1, exitCode, "%s should exit with 1", tc.subcommand)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, tc.errorText)
})
}
}

func TestCA_ServerHelp(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"server", "--help"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "pki ca server")
}

func TestCA_UnknownSubcommand(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Unknown subcommand")
}

// TestCA_ServerParseError verifies the Parse error path in caServerStart.
func TestCA_ServerParseError(t *testing.T) {
var stdout, stderr bytes.Buffer

//nolint:goconst // Test-specific invalid flag, not a magic string.
exitCode := caServerStart([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Failed to parse configuration")
}

// TestCA_ServerCreateError verifies the NewFromConfig error path.
func TestCA_ServerCreateError(t *testing.T) {
var stdout, stderr bytes.Buffer

exitCode := caServerStart([]string{}, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Failed to create server")
}

// TestCA_ServerLifecycle tests the full server start → signal → shutdown lifecycle.
func TestCA_ServerLifecycle(t *testing.T) {
var mu sync.Mutex

stdout := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}
stderr := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}

exitCodeChan := make(chan int, 1)

go func() {
exitCodeChan <- caServerStart([]string{
"--dev",
"--database-url", cryptoutilSharedMagic.SQLiteInMemoryDSN,
}, stdout, stderr)
}()

const maxWait = 30 * time.Second

startTime := time.Now().UTC()

for time.Since(startTime) < maxWait {
mu.Lock()

output := stdout.buf.String()

mu.Unlock()

if strings.Contains(output, "Starting pki-ca service") {
break
}

time.Sleep(100 * time.Millisecond) //nolint:mnd // Polling interval for server startup.
}

time.Sleep(500 * time.Millisecond) //nolint:mnd // Wait for server to finish binding.

err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
require.NoError(t, err)

select {
case exitCode := <-exitCodeChan:
require.Equal(t, 0, exitCode, "Server should exit cleanly after SIGTERM")
case <-time.After(maxWait):
t.Fatal("Server did not shut down within timeout")
}
}

// TestCA_ServerStartError tests the errChan error path when port is already in use.
func TestCA_ServerStartError(t *testing.T) {
var mu sync.Mutex

stdout := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}
stderr := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}

exitCodeChan := make(chan int, 1)

go func() {
exitCodeChan <- caServerStart([]string{
"--dev",
"--database-url", cryptoutilSharedMagic.SQLiteInMemoryDSN,
}, stdout, stderr)
}()

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

func TestCA_SubcommandLiveServer(t *testing.T) {
tests := []struct {
subcommand       string
url              string
expectedExitCode int
expectedOutputs  []string
}{
{
subcommand:       "health",
url:              publicBaseURL + "/service/api/v1",
expectedExitCode: 0,
expectedOutputs:  []string{"HTTP 200"},
},
{
subcommand:       "livez",
url:              adminBaseURL,
expectedExitCode: 0,
expectedOutputs:  []string{"HTTP 200"},
},
{
subcommand:       "readyz",
url:              adminBaseURL,
expectedExitCode: 0,
expectedOutputs:  []string{"HTTP 200"},
},
}

for _, tc := range tests {
t.Run(tc.subcommand, func(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{tc.subcommand, "--url", tc.url}, nil, &stdout, &stderr)
require.Equal(t, tc.expectedExitCode, exitCode, "%s should succeed", tc.subcommand)

output := stdout.String() + stderr.String()
for _, expected := range tc.expectedOutputs {
require.Contains(t, output, expected, "Output should contain: %s", expected)
}
})
}
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
