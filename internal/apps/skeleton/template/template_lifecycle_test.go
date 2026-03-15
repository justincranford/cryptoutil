// Copyright (c) 2025 Justin Cranford
//

package template

import (
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

// TestTemplate_ServerLifecycle verifies the full server start → signal → graceful shutdown path.
// Sequential: uses pflag.CommandLine global state via Parse() and process-level signals.
func TestTemplate_ServerLifecycle(t *testing.T) {
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("syscall.SIGINT is not supported on Windows.")
	}

	var stdout, stderr cryptoutilSharedTestutil.SafeBuffer

	exitCodeCh := make(chan int, 1)

	go func() {
		exitCodeCh <- templateServerStart(
			[]string{"--profile=test", "--bind-public-port=0", "--bind-private-port=0"},
			&stdout, &stderr,
		)
	}()

	// Wait for server to be fully started and listening.
	require.Eventually(t, func() bool {
		return strings.Contains(stdout.String(), "Starting skeleton-template service")
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
	require.Contains(t, combined, "skeleton-template service stopped")
}
