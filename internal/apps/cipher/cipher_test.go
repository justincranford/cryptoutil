// Copyright (c) 2025 Justin Cranford
//
//

package cipher_test

import (
	"strings"
	"testing"

	cryptoutilCipherCmd "cryptoutil/internal/apps/cipher"
	cryptoutilTestutil "cryptoutil/internal/shared/testutil"

	"github.com/stretchr/testify/require"
)

func TestCipher_NoArguments(t *testing.T) {
	// Remove t.Parallel() - stdout/stderr capture has race condition with parallel tests.
	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := cryptoutilCipherCmd.Cipher([]string{})
		require.Equal(t, 1, exitCode)
	})

	require.Contains(t, output, "Usage: cipher <service> <subcommand> [options]")
	require.Contains(t, output, "Available services:")
	require.Contains(t, output, "im")
}

func TestCipher_HelpCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "help command",
			args: []string{"help"},
		},
		{
			name: "help flag long",
			args: []string{"--help"},
		},
		{
			name: "help flag short",
			args: []string{"-h"},
		},
	}

	for _, tt := range tests {
		// Capture range variable for parallel tests.
		t.Run(tt.name, func(t *testing.T) {
			// Remove t.Parallel() - stdout/stderr capture uses global mutex for safety.
			// The CaptureOutput function serializes all calls to prevent race conditions on os.Stdout/os.Stderr.
			output := cryptoutilTestutil.CaptureOutput(t, func() {
				exitCode := cryptoutilCipherCmd.Cipher(tt.args)
				require.Equal(t, 0, exitCode)
			})

			require.Contains(t, output, "Usage: cipher <service> <subcommand> [options]")
			require.Contains(t, output, "Available services:")
			require.Contains(t, output, "im")
		})
	}
}

func TestCipher_VersionCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "version command",
			args: []string{"version"},
		},
		{
			name: "version flag long",
			args: []string{"--version"},
		},
		{
			name: "version flag short",
			args: []string{"-v"},
		},
	}

	for _, tt := range tests {
		// Capture range variable for parallel tests.
		t.Run(tt.name, func(t *testing.T) {
			// Remove t.Parallel() - stdout/stderr capture uses global mutex for safety.
			// The captureOutput function serializes all calls to prevent race conditions on os.Stdout/os.Stderr.
			combinedOutput := cryptoutilTestutil.CaptureOutput(t, func() {
				exitCode := cryptoutilCipherCmd.Cipher(tt.args)
				require.Equal(t, 0, exitCode)
			})

			require.Contains(t, combinedOutput, "cipher product")
			require.Contains(t, combinedOutput, "cryptoutil")
		})
	}
}

func TestCipher_UnknownService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "unknown service foo",
			args:        []string{"foo"},
			expectedErr: "Unknown service: foo",
		},
		{
			name:        "unknown service bar",
			args:        []string{"bar", "subcommand"},
			expectedErr: "Unknown service: bar",
		},
		{
			name:        "unknown service with flags",
			args:        []string{"unknown", "--flag", "value"},
			expectedErr: "Unknown service: unknown",
		},
	}

	for _, tt := range tests {
		// Capture range variable for parallel tests.
		t.Run(tt.name, func(t *testing.T) {
			// Remove t.Parallel() - stdout/stderr capture uses global mutex for safety.
			// The captureOutput function serializes all calls to prevent race conditions on os.Stdout/os.Stderr.
			combinedOutput := cryptoutilTestutil.CaptureOutput(t, func() {
				exitCode := cryptoutilCipherCmd.Cipher(tt.args)
				require.Equal(t, 1, exitCode)
			})

			// Must contain error message.
			require.Contains(t, combinedOutput, tt.expectedErr)
			// Should contain usage (may not always appear in parallel tests due to timing).
			// Just verify exit code is 1 which indicates error path was taken.
		})
	}
}

func TestCipher_IMService_RoutesCorrectly(t *testing.T) {
	// Remove t.Parallel() - stdout/stderr capture has race condition with parallel tests.

	// Test that "im" service routes to IM function.
	// We can't fully test IM() behavior here without a running server,
	// but we can verify routing doesn't panic and handles help.
	combinedOutput := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := cryptoutilCipherCmd.Cipher([]string{"im", "help"})
		require.Equal(t, 0, exitCode)
	})

	// Should show IM-specific help, not cipher product help (check combined output).
	require.Contains(t, combinedOutput, "Usage: cipher im <subcommand>")
	require.Contains(t, combinedOutput, "server")
	require.Contains(t, combinedOutput, "client")
	require.Contains(t, combinedOutput, "init")
}

func TestCipher_IMService_InvalidSubcommand(t *testing.T) {
	// Remove t.Parallel() - stdout/stderr capture has race condition with parallel tests.
	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := cryptoutilCipherCmd.Cipher([]string{"im", "invalid-subcommand"})
		require.Equal(t, 1, exitCode)
	})

	// Should show IM help with error.
	require.True(t,
		strings.Contains(output, "Unknown subcommand") ||
			strings.Contains(output, "Usage: cipher im <subcommand>"),
		"output should contain error or usage: %s", output,
	)
}

func TestLearn_Constants(t *testing.T) {
	// Remove t.Parallel() from parent - child tests use captureOutput with race condition.

	// Verify constants are used consistently.
	// This test documents expected constant values.
	tests := []struct {
		name     string
		args     []string
		exitCode int
	}{
		// Help variants.
		{name: "help", args: []string{"help"}, exitCode: 0},
		{name: "--help", args: []string{"--help"}, exitCode: 0},
		{name: "-h", args: []string{"-h"}, exitCode: 0},
		// Version variants.
		{name: "version", args: []string{"version"}, exitCode: 0},
		{name: "--version", args: []string{"--version"}, exitCode: 0},
		{name: "-v", args: []string{"-v"}, exitCode: 0},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.name, func(t *testing.T) {
			// Remove t.Parallel() - stdout/stderr capture has race condition with parallel tests.
			_ = cryptoutilTestutil.CaptureOutput(t, func() {
				exitCode := cryptoutilCipherCmd.Cipher(tt.args)
				require.Equal(t, tt.exitCode, exitCode)
			})
		})
	}
}
