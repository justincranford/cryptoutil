// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

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

// TestIM_ServerHelp verifies that server --help prints usage and returns 0.
func TestIM_ServerHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"server", "--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "cipher im server")
}

// TestIM_ServerParseError verifies the parse error path in imServiceServerStart.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestIM_ServerParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	//nolint:goconst // Test-specific invalid flag, not a magic string.
	exitCode := imServiceServerStart([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to parse configuration")
}

// TestIM_ServerCreateError verifies the NewFromConfig error path.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestIM_ServerCreateError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Server creation fails because PostgreSQL is not running on the default port.
	exitCode := imServiceServerStart([]string{}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to create server")
}

// TestIM_ServerLifecycle tests the full server start -> signal -> shutdown lifecycle.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestIM_ServerLifecycle(t *testing.T) {
	var mu sync.Mutex

	stdout := &imSyncWriter{buf: &bytes.Buffer{}, mu: &mu}
	stderr := &imSyncWriter{buf: &bytes.Buffer{}, mu: &mu}

	exitCodeChan := make(chan int, 1)

	go func() {
		exitCodeChan <- imServiceServerStart([]string{
			"--dev",
			"--database-url", cryptoutilSharedMagic.SQLiteInMemoryDSN,
			"--bind-private-port=0",
		}, stdout, stderr)
	}()

	// Wait for server to start (goroutine prints "Starting cipher-im service" to stdout).
	const maxWait = 30 * time.Second

	startTime := time.Now().UTC()

	for time.Since(startTime) < maxWait {
		mu.Lock()

		output := stdout.buf.String()

		mu.Unlock()

		if strings.Contains(output, "Starting cipher-im service") {
			break
		}

		time.Sleep(100 * time.Millisecond) //nolint:mnd // Polling interval for server startup.
	}

	// Give server a moment to fully bind listeners and register signal handler.
	time.Sleep(500 * time.Millisecond) //nolint:mnd // Wait for server to finish binding.

	// Send SIGTERM to trigger graceful shutdown via the signal handler.
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	require.NoError(t, err)

	// Wait for imServiceServerStart to return.
	select {
	case exitCode := <-exitCodeChan:
		require.Equal(t, 0, exitCode, "Server should exit cleanly after SIGTERM")
	case <-time.After(maxWait):
		t.Fatal("Server did not shut down within timeout")
	}
}

// TestIM_ServerStartError tests the errChan error path when the server fails
// to start (e.g., address already in use from the previous lifecycle test).
// Sequential: uses viper global state via ParseWithFlagSet.
func TestIM_ServerStartError(t *testing.T) {
	// Port 8700 is still bound by the server from TestIM_ServerLifecycle
	// (production code does not call srv.Shutdown after SIGTERM signal).
	// This means Start() will fail with "address already in use", triggering
	// the errChan error path.
	var mu sync.Mutex

	stdout := &imSyncWriter{buf: &bytes.Buffer{}, mu: &mu}
	stderr := &imSyncWriter{buf: &bytes.Buffer{}, mu: &mu}

	exitCodeChan := make(chan int, 1)

	go func() {
		exitCodeChan <- imServiceServerStart([]string{
			"--dev",
			"--database-url", cryptoutilSharedMagic.SQLiteInMemoryDSN,
			"--bind-private-port=0",
		}, stdout, stderr)
	}()

	// Wait for imServiceServerStart to return (should fail due to address in use).
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

// imSyncWriter is a thread-safe io.Writer that wraps a bytes.Buffer.
type imSyncWriter struct {
	buf *bytes.Buffer
	mu  *sync.Mutex
}

func (w *imSyncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buf.Write(p)
}
