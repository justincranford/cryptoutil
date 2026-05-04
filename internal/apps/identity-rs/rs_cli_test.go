// Copyright (c) 2025-2026 Justin Cranford.
package rs

import (
	"bytes"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestRS_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Rs([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestRS_Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Rs([]string{cryptoutilSharedMagic.CLIVersionCommand}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestRS_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Rs([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)
}
