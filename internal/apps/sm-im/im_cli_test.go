// Copyright (c) 2025-2026 Justin Cranford.
package im

import (
	"bytes"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestIM_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestIM_Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{cryptoutilSharedMagic.CLIVersionCommand}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestIM_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)
}
