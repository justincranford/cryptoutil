// Copyright (c) 2025 Justin Cranford

package learn_test

import (
	"testing"

	cryptoutilLearnCmd "cryptoutil/internal/cmd/learn"

	"github.com/stretchr/testify/require"
)

// TestIM_VersionCommand tests the IM version command.
// This test achieves 100% coverage for the printIMVersion function (currently 0.0%).
func TestIM_VersionCommand(t *testing.T) {
	t.Parallel()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"version"})
		require.Equal(t, 0, exitCode, "version command should exit with code 0")
	})

	// Verify output contains expected version information.
	require.Contains(t, stdout, "learn-im service", "output should contain service name")
	require.Contains(t, stdout, "cryptoutil learn product", "output should contain product name")
	require.Empty(t, stderr, "stderr should be empty for successful version command")
}

// TestIM_VersionFlag tests the IM --version flag.
func TestIM_VersionFlag(t *testing.T) {
	t.Parallel()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"--version"})
		require.Equal(t, 0, exitCode, "--version flag should exit with code 0")
	})

	require.Contains(t, stdout, "learn-im service", "output should contain service name")
	require.Empty(t, stderr, "stderr should be empty for successful version flag")
}

// TestIM_VersionShortFlag tests the IM -v flag.
func TestIM_VersionShortFlag(t *testing.T) {
	t.Parallel()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"-v"})
		require.Equal(t, 0, exitCode, "-v flag should exit with code 0")
	})

	require.Contains(t, stdout, "learn-im service", "output should contain service name")
	require.Empty(t, stderr, "stderr should be empty for successful version short flag")
}
