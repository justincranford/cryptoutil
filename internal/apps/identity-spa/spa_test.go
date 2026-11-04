// Copyright (c) 2025-2026 Justin Cranford.
package spa

import (
	"bytes"
	"testing"

	cryptoutilTestCli "cryptoutil/internal/apps-framework/service/test_help_cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestSPA(t *testing.T) {
	t.Parallel()

	cryptoutilTestCli.RunCLITests(t, Spa)
}

func TestStartupzHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Spa([]string{"startupz", cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)

	require.Zero(t, exitCode)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "Usage: "+cryptoutilSharedMagic.IdentityProductName+" "+cryptoutilSharedMagic.SPAServiceName+" startupz [options]")
	require.Contains(t, stderr.String(), "GET /admin/api/v1/startupz")
	require.Contains(t, stderr.String(), cryptoutilSharedMagic.CLICACertFlag)
	require.Contains(t, stderr.String(), cryptoutilSharedMagic.CLICertFlag)
	require.Contains(t, stderr.String(), cryptoutilSharedMagic.CLIKeyFlag)
}
