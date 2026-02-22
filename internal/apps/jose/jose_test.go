// Copyright (c) 2025 Justin Cranford
//
//

package jose

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJose_NoArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Jose([]string{}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Usage: jose <service> <subcommand> [options]")
	require.Contains(t, output, "Available services:")
	require.Contains(t, output, "ja")
}

func TestJose_HelpCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "help command", args: []string{"help"}},
		{name: "help flag long", args: []string{"--help"}},
		{name: "help flag short", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Jose(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			output := stdout.String() + stderr.String()
			require.Contains(t, output, "Usage: jose <service> <subcommand> [options]")
			require.Contains(t, output, "Available services:")
			require.Contains(t, output, "ja")
		})
	}
}

func TestJose_VersionCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "version command", args: []string{"version"}},
		{name: "version flag long", args: []string{"--version"}},
		{name: "version flag short", args: []string{"-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Jose(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			require.Contains(t, combinedOutput, "jose product")
		})
	}
}

func TestJose_UnknownService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{name: "unknown service", args: []string{"nonexistent"}, expectedErr: "Unknown service: nonexistent"},
		{name: "empty service name", args: []string{""}, expectedErr: "Unknown service: "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Jose(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 1, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			require.Contains(t, combinedOutput, tt.expectedErr)
		})
	}
}

func TestJose_JAService_RoutesCorrectly(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Jose([]string{"ja", "help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combinedOutput := stdout.String() + stderr.String()
	require.Contains(t, combinedOutput, "Usage: jose ja <subcommand>")
	require.Contains(t, combinedOutput, "server")
	require.Contains(t, combinedOutput, "client")
	require.Contains(t, combinedOutput, "init")
}

func TestJose_JAService_InvalidSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Jose([]string{"ja", "invalid-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	output := stdout.String() + stderr.String()
	require.True(t,
		strings.Contains(output, "Unknown subcommand") ||
			strings.Contains(output, "Usage: jose ja <subcommand>"),
		"output should contain error or usage: %s", output,
	)
}

func TestJose_EntryPoint(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Jose([]string{"--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Usage: jose")
}
