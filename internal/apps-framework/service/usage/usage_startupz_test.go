// Copyright (c) 2025-2026 Justin Cranford.
package usage_test

import (
	"testing"

	cryptoutilAppsFrameworkServiceUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestBuildUsageStartupz(t *testing.T) {
	t.Parallel()

	usageText := cryptoutilAppsFrameworkServiceUsage.BuildUsageStartupz(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
	)

	require.Contains(t, usageText, "Usage: sm kms startupz [options]")
	require.Contains(t, usageText, cryptoutilSharedMagic.DefaultAdminStartupzEndpoint)
	require.Contains(t, usageText, cryptoutilSharedMagic.CLICACertFlag)
	require.Contains(t, usageText, cryptoutilSharedMagic.CLICertFlag)
	require.Contains(t, usageText, cryptoutilSharedMagic.CLIKeyFlag)
}
