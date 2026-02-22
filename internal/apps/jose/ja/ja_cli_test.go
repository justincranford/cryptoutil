// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package ja

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

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
