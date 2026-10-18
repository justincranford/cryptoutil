// Copyright (c) 2025-2026 Justin Cranford.
//
//

package sysinfo

import (
	"fmt"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

var testSysInfoProviders = []SysInfoProvider{mockSysInfoProvider, newPortableDefaultSysInfoProvider()}

func newPortableDefaultSysInfoProvider() *DefaultSysInfoProvider {
	return &DefaultSysInfoProvider{
		cpuInfoFn: mockSysInfoProvider.CPUInfo,
	}
}

func TestSysInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		testFn func(t *testing.T, provider SysInfoProvider)
	}{
		{
			name: "runtime_go_arch",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				arch := provider.RuntimeGoArch()
				fmt.Printf("Time: %.3f msec >>> RuntimeGoArch: %s\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, arch)
				require.NotEmpty(t, arch)
			},
		},
		{
			name: "runtime_go_os",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				os := provider.RuntimeGoOS()
				fmt.Printf("Time: %.3f msec >>> RuntimeGoOS: %s\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, os)
				require.NotEmpty(t, os)
			},
		},
		{
			name: "runtime_num_cpu",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				numCPU := provider.RuntimeNumCPU()
				fmt.Printf("Time: %.3f msec >>> RuntimeNumCPU: %d\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, numCPU)
				require.NotZero(t, numCPU)
			},
		},
		{
			name: "cpu_info",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				vendorID, family, physicalID, modelName, err := provider.CPUInfo()
				fmt.Printf("Time: %.3f msec >>> CPUInfo: VendorID=%s, Family=%s, PhysicalID=%s, ModelName=%s\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, vendorID, family, physicalID, modelName)
				require.NoError(t, err)
				require.NotEmpty(t, vendorID, "vendorID was empty")
				require.NotEmpty(t, family, "family was empty")
				require.NotEmpty(t, physicalID, "physicalID was empty")
				require.NotEmpty(t, modelName, "modelName was empty")
			},
		},

		{
			name: "ram_size",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				ramSize, err := provider.RAMSize()
				fmt.Printf("Time: %.3f msec >>> RAMSize: %d\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, ramSize)
				require.NoError(t, err)
				require.NotZero(t, ramSize)
			},
		},
		{
			name: "os_hostname",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				hostname, err := provider.OSHostname()
				fmt.Printf("Time: %.3f msec >>> OSHostname: %s\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, hostname)
				require.NoError(t, err)
				require.NotEmpty(t, hostname)
			},
		},
		{
			name: "host_id",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				hostID, err := provider.HostID()
				fmt.Printf("Time: %.3f msec >>> HostID: %s\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, hostID)
				require.NoError(t, err)
				require.NotEmpty(t, hostID)
			},
		},
		{
			name: "user_info",
			testFn: func(t *testing.T, provider SysInfoProvider) {
				t.Helper()

				start := time.Now().UTC()
				userID, groupID, username, err := provider.UserInfo()
				fmt.Printf("Time: %.3f msec >>> UserInfo: UserID=%s, GroupID=%s, Username=%s\n", float32(time.Since(start).Microseconds())/cryptoutilSharedMagic.JoseJADefaultListLimit, userID, groupID, username)
				require.NoError(t, err)
				require.NotEmpty(t, userID)
				require.NotEmpty(t, groupID)
				require.NotEmpty(t, username)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, provider := range testSysInfoProviders {
				tc.testFn(t, provider)
			}
		})
	}

	runSysInfoErrorTests(t)
}
