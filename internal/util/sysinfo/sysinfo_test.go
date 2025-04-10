package sysinfo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testSysInfoProviders = []SysInfoProvider{mockSysInfoProvider, defaultSysInfoProvider}

func TestRuntimeGoArch(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		arch := testSysInfoProvider.RuntimeGoArch()
		fmt.Printf("Time: %.3f msec >>> RuntimeGoArch: %s\n", float32(time.Since(start).Microseconds())/1000, arch)
		require.NotEmpty(t, arch)
	}
}

func TestRuntimeGoOS(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		os := testSysInfoProvider.RuntimeGoOS()
		fmt.Printf("Time: %.3f msec >>> RuntimeGoOS: %s\n", float32(time.Since(start).Microseconds())/1000, os)
		require.NotEmpty(t, os)
	}
}

func TestRuntimeNumCPU(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		numCPU := testSysInfoProvider.RuntimeNumCPU()
		fmt.Printf("Time: %.3f msec >>> RuntimeNumCPU: %d\n", float32(time.Since(start).Microseconds())/1000, numCPU)
		require.NotZero(t, numCPU)
	}
}

func TestCPUInfo(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		vendorID, family, physicalID, modelName, err := testSysInfoProvider.CPUInfo()
		fmt.Printf("Time: %.3f msec >>> CPUInfo: VendorID=%s, Family=%s, PhysicalID=%s, ModelName=%s\n", float32(time.Since(start).Microseconds())/1000, vendorID, family, physicalID, modelName)
		require.NoError(t, err)
		require.NotEmpty(t, vendorID, "vendorID was empty")
		require.NotEmpty(t, family, "family was empty")
		require.NotEmpty(t, physicalID, "physicalID was empty")
		require.NotEmpty(t, modelName, "modelName was empty")
	}
}

func TestRAMSize(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		ramSize, err := testSysInfoProvider.RAMSize()
		fmt.Printf("Time: %.3f msec >>> RAMSize: %d\n", float32(time.Since(start).Microseconds())/1000, ramSize)
		require.NoError(t, err)
		require.NotZero(t, ramSize)
	}
}

func TestOSHostname(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		hostname, err := testSysInfoProvider.OSHostname()
		fmt.Printf("Time: %.3f msec >>> OSHostname: %s\n", float32(time.Since(start).Microseconds())/1000, hostname)
		require.NoError(t, err)
		require.NotEmpty(t, hostname)
	}
}

func TestHostID(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		hostID, err := testSysInfoProvider.HostID()
		fmt.Printf("Time: %.3f msec >>> HostID: %s\n", float32(time.Since(start).Microseconds())/1000, hostID)
		require.NoError(t, err)
		require.NotEmpty(t, hostID)
	}
}

func TestUserInfo(t *testing.T) {
	for _, testSysInfoProvider := range testSysInfoProviders {
		start := time.Now()
		userID, groupID, username, err := testSysInfoProvider.UserInfo()
		fmt.Printf("Time: %.3f msec >>> UserInfo: UserID=%s, GroupID=%s, Username=%s\n", float32(time.Since(start).Microseconds())/1000, userID, groupID, username)
		require.NoError(t, err)
		require.NotEmpty(t, userID)
		require.NotEmpty(t, groupID)
		require.NotEmpty(t, username)
	}
}
