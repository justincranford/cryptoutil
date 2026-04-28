// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package server_test

import (
	"bytes"
	"testing"

	cryptoutilAppsSmIm "cryptoutil/internal/apps/sm-im"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestIM_ServerHelp verifies that server --help prints usage and returns 0.
func TestIM_ServerHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsSmIm.Im([]string{"server", cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "sm im server")
}

// TestIM_ServerParseError verifies the parse error path in server CLI handling.
// Sequential: uses pflag.CommandLine global state via Parse().
func TestIM_ServerParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	//nolint:goconst // Test-specific invalid flag, not a magic string.
	exitCode := cryptoutilAppsSmIm.Im([]string{"server", "--this-flag-does-not-exist"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to parse configuration")
}

// TestIM_ServerCreateError verifies server creation error path.
// Sequential: uses pflag.CommandLine global state via Parse().
func TestIM_ServerCreateError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Server creation fails because PostgreSQL is not running on the default port.
	exitCode := cryptoutilAppsSmIm.Im([]string{"server"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to create server")
}
