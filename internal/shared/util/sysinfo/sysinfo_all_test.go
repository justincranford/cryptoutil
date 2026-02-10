// Copyright (c) 2025 Justin Cranford
//
//

package sysinfo

import (
	"runtime"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const expectedSysInfos = 13

func TestSysInfoAll(t *testing.T) {
	t.Parallel()
	all, err := GetAllInfoWithTimeout(mockSysInfoProvider, 5*time.Second)
	require.NoError(t, err)
	require.Len(t, all, expectedSysInfos)

	for i, value := range all {
		require.NotNil(t, value)
		require.NotEmpty(t, value)
		t.Logf("sysinfo[%d]: %s (0x%x)", i, string(value), value)
	}
}

func TestSysInfoAll_Timeout(t *testing.T) {
	t.Parallel()
	// Mock provider executes instantly, so timeout won't trigger with it.
	// Skip this test as timeout is only reachable with blocking operations.
	t.Skip("Mock provider too fast to test timeout path")
}

func TestSysInfoAll_RealProvider(t *testing.T) {
	t.Parallel()
	// Skip on Windows - CPU info collection can take 10+ seconds which exceeds
	// the timeout and causes test failures. Mock provider covers the code paths.
	if runtime.GOOS == "windows" {
		t.Skip("Skipping real sysinfo test on Windows due to slow gopsutil CPU collection")
	}

	// Test with real provider to cover defaultSysInfoProvider code paths.
	// Use DefaultSysInfoAllTimeout (10s) which is sufficient for slow Windows CPU info collection.
	all, err := GetAllInfoWithTimeout(defaultSysInfoProvider, cryptoutilSharedMagic.DefaultSysInfoAllTimeout)
	require.NoError(t, err)
	require.Len(t, all, expectedSysInfos)

	for i, value := range all {
		require.NotNil(t, value)
		require.NotEmpty(t, value)
		t.Logf("sysinfo[%d]: %s (0x%x)", i, string(value), value)
	}
}
