// Copyright (c) 2025 Justin Cranford
//

package template

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplate_SubcommandHelpFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		subcommand string
		helpTexts  []string
	}{
		{subcommand: "client", helpTexts: []string{"skeleton template client", "Run client operations"}},
		{subcommand: "init", helpTexts: []string{"skeleton template init", "Initialize database schema"}},
		{subcommand: "health", helpTexts: []string{"skeleton template health", "Check service health"}},
		{subcommand: "livez", helpTexts: []string{"skeleton template livez", "Check service liveness"}},
		{subcommand: "readyz", helpTexts: []string{"skeleton template readyz", "Check service readiness"}},
		{subcommand: "shutdown", helpTexts: []string{"skeleton template shutdown", "Trigger graceful shutdown"}},
	}

	for _, tc := range tests {
		t.Run(tc.subcommand, func(t *testing.T) {
			t.Parallel()

			for _, flag := range []string{"--help", "-h", "help"} {
				var stdout, stderr bytes.Buffer

				exitCode := Template([]string{tc.subcommand, flag}, nil, &stdout, &stderr)
				require.Equal(t, 0, exitCode, "%s %s should succeed", tc.subcommand, flag)

				combined := stdout.String() + stderr.String()
				for _, expected := range tc.helpTexts {
					require.Contains(t, combined, expected, "%s output should contain: %s", flag, expected)
				}
			}
		})
	}
}

func TestTemplate_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Template([]string{"--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "skeleton template")
}

func TestTemplate_Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Template([]string{"version"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestTemplate_SubcommandNotImplemented(t *testing.T) {
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

			exitCode := Template([]string{tc.subcommand}, nil, &stdout, &stderr)
			require.Equal(t, 1, exitCode, "%s should exit with 1", tc.subcommand)

			combined := stdout.String() + stderr.String()
			require.Contains(t, combined, tc.errorText)
		})
	}
}

func TestTemplate_ServerHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Template([]string{"server", "--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "skeleton template server")
}

func TestTemplate_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Template([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Unknown subcommand")
}

// TestTemplate_ServerParseError verifies the Parse error path in templateServerStart.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestTemplate_ServerParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	//nolint:goconst // Test-specific invalid flag, not a magic string.
	exitCode := templateServerStart([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to parse configuration")
}

// TestTemplate_ServerCreateError verifies the NewFromConfig error path.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestTemplate_ServerCreateError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Server creation fails because PostgreSQL is not running on the default port.
	exitCode := templateServerStart([]string{}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to create server")
}
