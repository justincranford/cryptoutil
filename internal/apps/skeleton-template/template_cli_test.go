// Copyright (c) 2025-2026 Justin Cranford.
package template

import (
	"bytes"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestTemplate_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Template([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestTemplate_Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Template([]string{cryptoutilSharedMagic.CLIVersionCommand}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestTemplate_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Template([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)
}
