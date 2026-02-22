// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package kms

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestKMS_ServerHelp verifies that server --help prints usage and returns 0.
func TestKMS_ServerHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Kms([]string{"server", "--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "sm kms server")
}

// TestKMS_MainHelp verifies that --help prints main usage and returns 0.
func TestKMS_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Kms([]string{"--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "sm kms")
}

// TestKMS_Version verifies the version subcommand.
func TestKMS_Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Kms([]string{"version"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "sm-kms")
}

// TestKMS_SubcommandHelpFlags verifies help flags for all subcommands.
func TestKMS_SubcommandHelpFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subcommand string
		contains   string
	}{
		{name: "client_help", subcommand: "client", contains: "sm kms client"},
		{name: "init_help", subcommand: "init", contains: "sm kms init"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Kms([]string{tc.subcommand, "--help"}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			combined := stdout.String() + stderr.String()
			require.Contains(t, combined, tc.contains)
		})
	}
}

// TestKMS_SubcommandNotImplemented verifies client and init subcommands return error.
func TestKMS_SubcommandNotImplemented(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subcommand string
		contains   string
	}{
		{name: "client", subcommand: "client", contains: "not yet implemented"},
		{name: "init", subcommand: "init", contains: "not yet implemented"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Kms([]string{tc.subcommand}, nil, &stdout, &stderr)
			require.Equal(t, 1, exitCode)

			combined := stdout.String() + stderr.String()
			require.Contains(t, combined, tc.contains)
		})
	}
}

// TestKMS_UnknownSubcommand verifies unknown subcommand returns error.
func TestKMS_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Kms([]string{"nonexistent"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Unknown subcommand")
}

// TestKMS_ServerParseError verifies the parse error path in kmsServerStart.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestKMS_ServerParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	//nolint:goconst // Test-specific invalid flag, not a magic string.
	exitCode := kmsServerStart([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to parse configuration")
}

// TestKMS_ServerCreateError verifies the NewKMSServer error path.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestKMS_ServerCreateError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Server creation fails because PostgreSQL is not running on the default port.
	exitCode := kmsServerStart([]string{}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to create server")
}

// NOTE: TestKMS_ServerLifecycle and TestKMS_ServerStartError are not included
// because NewKMSServer initialization hangs when using --dev mode with SQLite.
// The KMS server has dual initialization (ServerApplicationCore + ServerBuilder.Build)
// and the keygen pool workers block indefinitely during test execution.
// The srv.Start() goroutine, signal handler, and errChan paths remain uncovered.
