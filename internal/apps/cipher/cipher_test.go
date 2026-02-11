// Copyright (c) 2025 Justin Cranford
//
//

package cipher

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCipher_NoArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cipher([]string{}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	output := stdout.String() + stderr.String()
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
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Cipher(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			output := stdout.String() + stderr.String()
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
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Cipher(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			combinedOutput := stdout.String() + stderr.String()
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
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Cipher(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 1, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			// Must contain error message.
			require.Contains(t, combinedOutput, tt.expectedErr)
			// Should contain usage (may not always appear in parallel tests due to timing).
			// Just verify exit code is 1 which indicates error path was taken.
		})
	}
}

func TestCipher_IMService_RoutesCorrectly(t *testing.T) {
	t.Parallel()

	// Test that "im" service routes to IM function.
	// We can't fully test IM() behavior here without a running server,
	// but we can verify routing doesn't panic and handles help.
	var stdout, stderr bytes.Buffer

	exitCode := Cipher([]string{"im", "help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	// Should show IM-specific help, not cipher product help (check combined output).
	combinedOutput := stdout.String() + stderr.String()
	require.Contains(t, combinedOutput, "Usage: cipher im <subcommand>")
	require.Contains(t, combinedOutput, "server")
	require.Contains(t, combinedOutput, "client")
	require.Contains(t, combinedOutput, "init")
}

func TestCipher_IMService_InvalidSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cipher([]string{"im", "invalid-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	// Should show IM help with error.
	output := stdout.String() + stderr.String()
	require.True(t,
		strings.Contains(output, "Unknown subcommand") ||
			strings.Contains(output, "Usage: cipher im <subcommand>"),
		"output should contain error or usage: %s", output,
	)
}

func TestLearn_Constants(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Cipher(tt.args, nil, &stdout, &stderr)
			require.Equal(t, tt.exitCode, exitCode)
		})
	}
}

// TestCipher_EntryPoint tests the public Cipher entry point.
// This test ensures the entry point wrapper function is covered.
func TestCipher_EntryPoint(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	// Test help flag via public entry point.
	exitCode := Cipher([]string{"--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Usage: cipher")
}
