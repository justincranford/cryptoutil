// Copyright (c) 2025-2026 Justin Cranford.
package idp

import (
	"bytes"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestIDP_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Idp([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestIDP_Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Idp([]string{cryptoutilSharedMagic.CLIVersionCommand}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func TestIDP_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Idp([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)
}
