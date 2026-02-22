// Copyright (c) 2025 Justin Cranford
//
//

package identity

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentity_NoArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Identity([]string{}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Usage: identity <service> <subcommand> [options]")
	require.Contains(t, output, "Available services:")
	require.Contains(t, output, "authz")
	require.Contains(t, output, "idp")
	require.Contains(t, output, "rp")
	require.Contains(t, output, "rs")
	require.Contains(t, output, "spa")
}

func TestIdentity_HelpCommand(t *testing.T) {
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

			exitCode := Identity(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			output := stdout.String() + stderr.String()
			require.Contains(t, output, "Usage: identity <service> <subcommand> [options]")
			require.Contains(t, output, "Available services:")
			require.Contains(t, output, "authz")
		})
	}
}

func TestIdentity_VersionCommand(t *testing.T) {
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

			exitCode := Identity(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			require.Contains(t, combinedOutput, "identity product")
		})
	}
}

func TestIdentity_UnknownService(t *testing.T) {
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

			exitCode := Identity(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 1, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			require.Contains(t, combinedOutput, tt.expectedErr)
		})
	}
}

func TestIdentity_ServiceRouting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		serviceName     string
		expectedUsage   string
	}{
		{name: "authz service", serviceName: "authz", expectedUsage: "Usage: identity authz <subcommand>"},
		{name: "idp service", serviceName: "idp", expectedUsage: "Usage: identity idp <subcommand>"},
		{name: "rp service", serviceName: "rp", expectedUsage: "Usage: identity rp <subcommand>"},
		{name: "rs service", serviceName: "rs", expectedUsage: "Usage: identity rs <subcommand>"},
		{name: "spa service", serviceName: "spa", expectedUsage: "Usage: identity spa <subcommand>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Identity([]string{tt.serviceName, "help"}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			require.Contains(t, combinedOutput, tt.expectedUsage)
			require.Contains(t, combinedOutput, "server")
			require.Contains(t, combinedOutput, "client")
			require.Contains(t, combinedOutput, "init")
		})
	}
}

func TestIdentity_EntryPoint(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Identity([]string{"--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Usage: identity")
}
