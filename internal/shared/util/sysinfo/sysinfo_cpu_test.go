// Copyright (c) 2025-2026 Justin Cranford.
package sysinfo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultSysInfoProvider_CPUInfo(t *testing.T) {
	t.Parallel()

	vendorID, family, physicalID, modelName, err := defaultSysInfoProvider.CPUInfo()
	if err != nil {
		require.ErrorContains(t, err, "failed to get CPU info")

		return
	}

	require.NotEmpty(t, vendorID)
	require.NotEmpty(t, family)
	require.NotEmpty(t, physicalID)
	require.NotEmpty(t, modelName)
}
