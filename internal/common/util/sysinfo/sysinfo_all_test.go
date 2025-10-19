package sysinfo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const expectedSysInfos = 13

func TestSysInfoAll(t *testing.T) {
	all, err := GetAllInfoWithTimeout(mockSysInfoProvider, 5*time.Second)
	require.NoError(t, err)
	require.Len(t, all, expectedSysInfos)

	for i, value := range all {
		require.NotNil(t, value)
		require.NotEmpty(t, value)
		t.Logf("sysinfo[%d]: %s (0x%x)", i, string(value), value)
	}
}
