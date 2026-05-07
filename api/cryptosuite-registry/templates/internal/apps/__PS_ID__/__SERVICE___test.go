//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
package __SERVICE__

import (
	"bytes"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func Test__USAGE_PREFIX___MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := __ENTRY_FUNC__([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func Test__USAGE_PREFIX___Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := __ENTRY_FUNC__([]string{cryptoutilSharedMagic.CLIVersionCommand}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode)
}

func Test__USAGE_PREFIX___UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := __ENTRY_FUNC__([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)
}
