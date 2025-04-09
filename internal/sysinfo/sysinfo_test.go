package sysinfo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimeNumCPU(t *testing.T) {
	uncached := RuntimeNumCPU(false)

	// Validate uncached values
	require.Greater(t, uncached, uint16(0), "Expected CPU count to be greater than 0")

	cached := RuntimeNumCPU(true)

	// Validate cached values
	require.Greater(t, cached, uint16(0), "Expected CPU count to be greater than 0")

	// Validate cached == uncached
	require.Equal(t, uncached, cached, "Mismatch between uncached and cached CPU count")
}

func TestGoArch(t *testing.T) {
	uncached := GoArch(false)

	// Validate uncached values
	require.NotNil(t, uncached, "Expected non-nil architecture")
	require.NotEmpty(t, *uncached, "Expected non-empty architecture")

	cached := GoArch(true)

	// Validate cached values
	require.NotNil(t, cached, "Expected non-nil cached architecture")
	require.NotEmpty(t, *cached, "Expected non-empty architecture")

	// Validate cached == uncached
	require.Equal(t, *uncached, *cached, "Mismatch between uncached and cached architecture")
	require.Equal(t, uncached, cached, "Mismatch between uncached and cached architecture")
}

func TestGoOS(t *testing.T) {
	uncached := GoOS(false)

	// Validate uncached values
	require.NotNil(t, uncached, "Expected non-nil OS")
	require.NotEmpty(t, *uncached, "Expected non-empty OS")

	cached := GoOS(true)

	// Validate cached values
	require.NotNil(t, cached, "Expected non-nil cached OS")

	// Validate cached == uncached
	require.Equal(t, *uncached, *cached, "Mismatch between uncached and cached OS")
	require.Equal(t, uncached, cached, "Mismatch between uncached and cached OS")
}

func TestCPUInfo(t *testing.T) {
	uncached, uncachedErr := CPUInfo(false)

	// Validate uncached values
	require.NoError(t, uncachedErr, "Unexpected error fetching uncached CPU info")
	require.NotNil(t, uncached, "Expected non-nil CPU info")
	require.NotEmpty(t, *uncached, "Expected non-empty CPU info")

	cached, cachedErr := CPUInfo(true)

	// Validate cached values
	require.NoError(t, cachedErr, "Unexpected error fetching cached CPU info")
	require.NotNil(t, cached, "Expected non-nil cached CPU info")
	require.NotEmpty(t, *cached, "Expected non-empty CPU info")

	// Validate cached == uncached
	require.Equal(t, uncachedErr, cachedErr, "Mismatch between uncached and cached errors")
	require.Equal(t, *uncached, *cached, "Mismatch between uncached and cached CPU info")
	require.Equal(t, uncached, cached, "Mismatch between uncached and cached CPU info")
}

func TestRAMSize(t *testing.T) {
	uncached, uncachedErr := RAMSize(false)

	// Validate uncached values
	require.NoError(t, uncachedErr, "Unexpected error fetching uncached RAM size")
	require.Greater(t, uncached, uint64(0), "Expected RAM size to be greater than 0")

	cached, cachedErr := RAMSize(true)

	// Validate cached values
	require.NoError(t, cachedErr, "Unexpected error fetching cached RAM size")
	require.Greater(t, cached, uint64(0), "Expected RAM size to be greater than 0")

	// Validate cached == uncached
	require.Equal(t, uncachedErr, cachedErr, "Mismatch between uncached and cached errors")
	require.Equal(t, uncached, cached, "Mismatch between uncached and cached RAM size")
}

func TestDiskSize(t *testing.T) {
	uncached, uncachedErr := DiskSize(false)

	// Validate uncached values
	require.NoError(t, uncachedErr, "Unexpected error fetching uncached disk size")
	require.NotNil(t, uncached, "Expected non-nil disk size")
	require.NotEmpty(t, *uncached, "Expected non-empty disk size")

	cached, cachedErr := DiskSize(true)

	// Validate cached values
	require.NoError(t, cachedErr, "Unexpected error fetching cached disk size")
	require.NotNil(t, cached, "Expected non-nil cached disk size")
	require.NotEmpty(t, *cached, "Expected non-empty disk size")

	// Validate cached == uncached
	require.Equal(t, uncachedErr, cachedErr, "Mismatch between uncached and cached errors")
	require.Equal(t, *uncached, *cached, "Mismatch between uncached and cached disk size")
	require.Equal(t, uncached, cached, "Mismatch between uncached and cached disk size")
}

func TestOSHostname(t *testing.T) {
	uncached, uncachedErr := OSHostname(false)

	// Validate uncached values
	require.NoError(t, uncachedErr, "Unexpected error fetching uncached hostname")
	require.NotNil(t, uncached, "Expected non-nil hostname")
	require.NotEmpty(t, *uncached, "Expected non-empty hostname")

	cached, cachedErr := OSHostname(true)

	// Validate cached values
	require.NoError(t, cachedErr, "Unexpected error fetching cached hostname")
	require.Equal(t, uncachedErr, cachedErr, "Mismatch between uncached and cached errors")
	require.NotNil(t, cached, "Expected non-nil cached hostname")

	// Validate cached == uncached
	require.Equal(t, *uncached, *cached, "Mismatch between uncached and cached hostname")
	require.Equal(t, uncached, cached, "Mismatch between uncached and cached hostname")
}

func TestHostInfo(t *testing.T) {
	uncachedHostname, uncachedOS, uncachedPlatform, uncachedPlatformFamily, uncachedPlatformVersion, uncachedKernelVersion, uncachedKernelArch, uncachedVirtualizationSystem, uncachedVirtualizationRole, uncachedHostID, uncachedErr := HostInfo(false)

	// Validate uncached values
	require.NoError(t, uncachedErr, "Unexpected error fetching uncached host info")
	require.NotNil(t, uncachedHostname, "Expected non-nil hostname")
	require.NotEmpty(t, *uncachedHostname, "Expected non-empty hostname")
	require.NotNil(t, uncachedOS, "Expected non-nil OS")
	require.NotEmpty(t, *uncachedOS, "Expected non-empty OS")
	require.NotNil(t, uncachedPlatform, "Expected non-nil platform")
	require.NotEmpty(t, *uncachedPlatform, "Expected non-empty platform")
	require.NotNil(t, uncachedPlatformFamily, "Expected non-nil platform family")
	require.NotEmpty(t, *uncachedPlatformFamily, "Expected non-empty platform family")
	require.NotNil(t, uncachedPlatformVersion, "Expected non-nil platform version")
	require.NotEmpty(t, *uncachedPlatformVersion, "Expected non-empty platform version")
	require.NotNil(t, uncachedKernelVersion, "Expected non-nil kernel version")
	require.NotEmpty(t, *uncachedKernelVersion, "Expected non-empty kernel version")
	require.NotNil(t, uncachedKernelArch, "Expected non-nil kernel architecture")
	require.NotEmpty(t, *uncachedKernelArch, "Expected non-empty kernel architecture")
	require.NotNil(t, uncachedVirtualizationSystem, "Expected non-nil virtualization system")
	// require.NotEmpty(t, *uncachedVirtualizationSystem, "Expected non-empty virtualization system")
	require.NotNil(t, uncachedVirtualizationRole, "Expected non-nil virtualization role")
	// require.NotEmpty(t, *uncachedVirtualizationRole, "Expected non-empty virtualization role")
	require.NotNil(t, uncachedHostID, "Expected non-nil host ID")
	require.NotEmpty(t, *uncachedHostID, "Expected non-empty host ID")

	cachedHostname, cachedOS, cachedPlatform, cachedPlatformFamily, cachedPlatformVersion, cachedKernelVersion, cachedKernelArch, cachedVirtualizationSystem, cachedVirtualizationRole, cachedHostID, cachedErr := HostInfo(true)

	// Validate cached values
	require.NoError(t, cachedErr, "Unexpected error fetching cached host info")
	require.NotNil(t, cachedHostname, "Expected non-nil hostname")
	require.NotEmpty(t, *cachedHostname, "Expected non-empty hostname")
	require.NotNil(t, cachedOS, "Expected non-nil OS")
	require.NotEmpty(t, *cachedOS, "Expected non-empty OS")
	require.NotNil(t, cachedPlatform, "Expected non-nil platform")
	require.NotEmpty(t, *cachedPlatform, "Expected non-empty platform")
	require.NotNil(t, cachedPlatformFamily, "Expected non-nil platform family")
	require.NotEmpty(t, *cachedPlatformFamily, "Expected non-empty platform family")
	require.NotNil(t, cachedPlatformVersion, "Expected non-nil platform version")
	require.NotEmpty(t, *cachedPlatformVersion, "Expected non-empty platform version")
	require.NotNil(t, cachedKernelVersion, "Expected non-nil kernel version")
	require.NotEmpty(t, *cachedKernelVersion, "Expected non-empty kernel version")
	require.NotNil(t, cachedKernelArch, "Expected non-nil kernel architecture")
	require.NotEmpty(t, *cachedKernelArch, "Expected non-empty kernel architecture")
	require.NotNil(t, cachedVirtualizationSystem, "Expected non-nil virtualization system")
	// require.NotEmpty(t, *cachedVirtualizationSystem, "Expected non-empty virtualization system")
	require.NotNil(t, cachedVirtualizationRole, "Expected non-nil virtualization role")
	// require.NotEmpty(t, *cachedVirtualizationRole, "Expected non-empty virtualization role")
	require.NotNil(t, cachedHostID, "Expected non-nil host ID")
	require.NotEmpty(t, *cachedHostID, "Expected non-empty host ID")

	// Validate cached == uncached
	require.Equal(t, uncachedErr, cachedErr, "Mismatch between uncached and cached errors")
	require.Equal(t, *uncachedHostname, *cachedHostname, "Mismatch between uncached and cached hostname")
	require.Equal(t, *uncachedOS, *cachedOS, "Mismatch between uncached and cached OS")
	require.Equal(t, *uncachedPlatform, *cachedPlatform, "Mismatch between uncached and cached platform")
	require.Equal(t, *uncachedPlatformFamily, *cachedPlatformFamily, "Mismatch between uncached and cached platform family")
	require.Equal(t, *uncachedPlatformVersion, *cachedPlatformVersion, "Mismatch between uncached and cached platform version")
	require.Equal(t, *uncachedKernelVersion, *cachedKernelVersion, "Mismatch between uncached and cached kernel version")
	require.Equal(t, *uncachedKernelArch, *cachedKernelArch, "Mismatch between uncached and cached kernel architecture")
	require.Equal(t, *uncachedVirtualizationSystem, *cachedVirtualizationSystem, "Mismatch between uncached and cached virtualization system")
	require.Equal(t, *uncachedVirtualizationRole, *cachedVirtualizationRole, "Mismatch between uncached and cached virtualization role")
	require.Equal(t, *uncachedHostID, *cachedHostID, "Mismatch between uncached and cached host ID")
}

func TestUserStat(t *testing.T) {
	uncachedUsername, uncachedName, uncachedUid, uncachedGid, uncachedHomeDir, uncachedErr := UserStat(false)

	// Validate uncached values
	require.NoError(t, uncachedErr, "Unexpected error fetching uncached user stats")
	require.NotNil(t, uncachedUsername, "Expected non-nil username")
	require.NotEmpty(t, *uncachedUsername, "Expected non-empty username")
	require.NotNil(t, uncachedName, "Expected non-nil name")
	require.NotEmpty(t, *uncachedName, "Expected non-empty name")
	require.NotNil(t, uncachedUid, "Expected non-nil UID")
	require.NotEmpty(t, *uncachedUid, "Expected non-empty UID")
	require.NotNil(t, uncachedGid, "Expected non-nil GID")
	require.NotEmpty(t, *uncachedGid, "Expected non-empty GID")
	require.NotNil(t, uncachedHomeDir, "Expected non-nil home directory")
	require.NotEmpty(t, *uncachedHomeDir, "Expected non-empty home directory")

	cachedUsername, cachedName, cachedUid, cachedGid, cachedHomeDir, cachedErr := UserStat(true)

	// Validate cached values
	require.NoError(t, cachedErr, "Unexpected error fetching cached user stats")
	require.NotNil(t, cachedUsername, "Expected non-nil username")
	require.NotEmpty(t, *cachedUsername, "Expected non-empty username")
	require.NotNil(t, cachedName, "Expected non-nil name")
	require.NotEmpty(t, *cachedName, "Expected non-empty name")
	require.NotNil(t, cachedUid, "Expected non-nil UID")
	require.NotEmpty(t, *cachedUid, "Expected non-empty UID")
	require.NotNil(t, cachedGid, "Expected non-nil GID")
	require.NotEmpty(t, *cachedGid, "Expected non-empty GID")
	require.NotNil(t, cachedHomeDir, "Expected non-nil home directory")
	require.NotEmpty(t, *cachedHomeDir, "Expected non-empty home directory")

	// Validate cached == uncached
	require.Equal(t, uncachedErr, cachedErr, "Mismatch between uncached and cached errors")
	require.Equal(t, *uncachedUsername, *cachedUsername, "Mismatch between uncached and cached username")
	require.Equal(t, *uncachedName, *cachedName, "Mismatch between uncached and cached name")
	require.Equal(t, *uncachedUid, *cachedUid, "Mismatch between uncached and cached UID")
	require.Equal(t, *uncachedGid, *cachedGid, "Mismatch between uncached and cached GID")
	require.Equal(t, *uncachedHomeDir, *cachedHomeDir, "Mismatch between uncached and cached home directory")
}
