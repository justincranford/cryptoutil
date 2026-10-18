// Copyright (c) 2025-2026 Justin Cranford.
//
//

package sysinfo

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const expectedSysInfos = 13

func TestSysInfoAll(t *testing.T) {
	t.Parallel()

	all, err := GetAllInfoWithTimeout(mockSysInfoProvider, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	require.NoError(t, err)
	require.Len(t, all, expectedSysInfos)

	for i, value := range all {
		require.NotNil(t, value)
		require.NotEmpty(t, value)
		t.Logf("sysinfo[%d]: %s (0x%x)", i, string(value), value)
	}
}

func TestSysInfoAll_DefaultProvider(t *testing.T) {
	t.Parallel()

	all, err := GetAllInfoWithTimeout(newPortableDefaultSysInfoProvider(), cryptoutilSharedMagic.DefaultSysInfoAllTimeout)
	require.NoError(t, err)
	require.Len(t, all, expectedSysInfos)

	for i, value := range all {
		require.NotNil(t, value)
		require.NotEmpty(t, value)
		t.Logf("sysinfo[%d]: %s (0x%x)", i, string(value), value)
	}
}
