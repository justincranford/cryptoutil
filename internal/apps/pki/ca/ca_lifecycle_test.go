// Copyright (c) 2025 Justin Cranford
//

package ca

import (
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

// TestCA_ServerLifecycle verifies the full server start → signal → graceful shutdown path.
// Sequential: uses viper global state via ParseWithFlagSet and process-level signals.
func TestCA_ServerLifecycle(t *testing.T) {
	// Reset viper global state after test to prevent leaking --profile=test to subsequent tests.
	t.Cleanup(func() { viper.Reset() })

	var stdout, stderr cryptoutilSharedTestutil.SafeBuffer

	exitCodeCh := make(chan int, 1)

	go func() {
		exitCodeCh <- caServerStart(
			[]string{"--profile=test", "--bind-public-port=0", "--bind-private-port=0"},
			&stdout, &stderr,
		)
	}()

	// Wait for server to be fully started and listening.
	require.Eventually(t, func() bool {
		return strings.Contains(stdout.String(), "Starting pki-ca service")
	}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second, 200*time.Millisecond, "server should start within timeout")

	// Send SIGINT to trigger the signal handler and graceful shutdown.
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(syscall.SIGINT))

	// Wait for the function to return.
	select {
	case exitCode := <-exitCodeCh:
		require.Equal(t, 0, exitCode, "graceful shutdown should return exit code 0")
	case <-time.After(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second):
		t.Fatal("server did not shut down within timeout")
	}

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "pki-ca service stopped")
}
