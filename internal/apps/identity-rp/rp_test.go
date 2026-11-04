// Copyright (c) 2025-2026 Justin Cranford.
package rp

import (
	"bytes"
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestRP(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Rp)
}

func TestStartupzHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Rp([]string{"startupz", cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)

	require.Zero(t, exitCode)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "Usage: "+cryptoutilSharedMagic.IdentityProductName+" "+cryptoutilSharedMagic.RPServiceName+" startupz [options]")
	require.Contains(t, stderr.String(), "GET /admin/api/v1/startupz")
	require.Contains(t, stderr.String(), cryptoutilSharedMagic.CLICACertFlag)
	require.Contains(t, stderr.String(), cryptoutilSharedMagic.CLICertFlag)
	require.Contains(t, stderr.String(), cryptoutilSharedMagic.CLIKeyFlag)
}
