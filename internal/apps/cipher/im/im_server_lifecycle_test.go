// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

)

// TestIM_ServerHelp verifies that server --help prints usage and returns 0.
func TestIM_ServerHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"server", "--help"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "cipher im server")
}

// TestIM_ServerParseError verifies the parse error path in imServiceServerStart.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestIM_ServerParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	//nolint:goconst // Test-specific invalid flag, not a magic string.
	exitCode := imServiceServerStart([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to parse configuration")
}

// TestIM_ServerCreateError verifies the NewFromConfig error path.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestIM_ServerCreateError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Server creation fails because PostgreSQL is not running on the default port.
	exitCode := imServiceServerStart([]string{}, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	combined := stdout.String() + stderr.String()
	require.Contains(t, combined, "Failed to create server")
}

