// Copyright (c) 2025 Justin Cranford
//

package sysinfo

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"os/user"
	"testing"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/require"
)

// TestCPUInfo_Error tests error path when gopsutil fails.
// Cannot be parallel since it modifies package-level vars.
func TestCPUInfo_Error(t *testing.T) {
	orig := sysinfoGetCPUInfoFn
	sysinfoGetCPUInfoFn = func(_ context.Context) ([]cpu.InfoStat, error) {
		return nil, errors.New("injected CPU info failure")
	}

	defer func() { sysinfoGetCPUInfoFn = orig }()

	_, _, _, _, err := CPUInfo()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get CPU info")
}

// TestCPUInfo_NoCPU tests the "no CPU info" error path when gopsutil returns empty slice.
// Cannot be parallel since it modifies package-level vars.
func TestCPUInfo_NoCPU(t *testing.T) {
	orig := sysinfoGetCPUInfoFn
	sysinfoGetCPUInfoFn = func(_ context.Context) ([]cpu.InfoStat, error) {
		return []cpu.InfoStat{}, nil // empty slice
	}

	defer func() { sysinfoGetCPUInfoFn = orig }()

	_, _, _, _, err := CPUInfo()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no CPU info")
}

// TestRAMSize_Error tests error path when gopsutil fails.
// Cannot be parallel since it modifies package-level vars.
func TestRAMSize_Error(t *testing.T) {
	orig := sysinfoGetVirtualMemoryFn
	sysinfoGetVirtualMemoryFn = func(_ context.Context) (*mem.VirtualMemoryStat, error) {
		return nil, errors.New("injected RAM info failure")
	}

	defer func() { sysinfoGetVirtualMemoryFn = orig }()

	_, err := RAMSize()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get RAM info")
}

// TestOSHostname_Error tests error path when os.Hostname fails.
// Cannot be parallel since it modifies package-level vars.
func TestOSHostname_Error(t *testing.T) {
	orig := sysinfoGetOSHostnameFn
	sysinfoGetOSHostnameFn = func() (string, error) {
		return "", errors.New("injected hostname failure")
	}

	defer func() { sysinfoGetOSHostnameFn = orig }()

	_, err := OSHostname()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get OS hostname")
}

// TestHostID_Error tests error path when gopsutil fails.
// Cannot be parallel since it modifies package-level vars.
func TestHostID_Error(t *testing.T) {
	orig := sysinfoGetHostIDFn
	sysinfoGetHostIDFn = func(_ context.Context) (string, error) {
		return "", errors.New("injected host ID failure")
	}

	defer func() { sysinfoGetHostIDFn = orig }()

	_, err := HostID()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get host ID")
}

// TestUserInfo_Error tests error path when user.Current fails.
// Cannot be parallel since it modifies package-level vars.
func TestUserInfo_Error(t *testing.T) {
	orig := sysinfoGetUserCurrentFn
	sysinfoGetUserCurrentFn = func() (*user.User, error) {
		return nil, errors.New("injected user info failure")
	}

	defer func() { sysinfoGetUserCurrentFn = orig }()

	_, _, _, err := UserInfo()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get user info")
}

// blockingProviderForTimeout is a SysInfoProvider that blocks until released.
// Used to trigger timeout paths in GetAllInfoWithTimeout.
type blockingProviderForTimeout struct {
	releaseCh chan struct{}
}

func (p *blockingProviderForTimeout) RuntimeGoArch() string {
	return cryptoutilSharedMagic.MockRuntimeGoArch
}
func (p *blockingProviderForTimeout) RuntimeGoOS() string {
	return cryptoutilSharedMagic.MockRuntimeGoOS
}
func (p *blockingProviderForTimeout) RuntimeNumCPU() int { return 1 }

func (p *blockingProviderForTimeout) CPUInfo() (string, string, string, string, error) {
	<-p.releaseCh

	return "", "", "", "", nil
}

func (p *blockingProviderForTimeout) RAMSize() (uint64, error) {
	<-p.releaseCh

	return 0, nil
}

func (p *blockingProviderForTimeout) OSHostname() (string, error) {
	<-p.releaseCh

	return "", nil
}

func (p *blockingProviderForTimeout) HostID() (string, error) {
	<-p.releaseCh

	return "", nil
}

func (p *blockingProviderForTimeout) UserInfo() (string, string, string, error) {
	<-p.releaseCh

	return "", "", "", nil
}

// TestGetAllInfoWithTimeout_Timeout tests that timeout paths are triggered.
func TestGetAllInfoWithTimeout_Timeout(t *testing.T) {
	t.Parallel()

	releaseCh := make(chan struct{})
	provider := &blockingProviderForTimeout{releaseCh: releaseCh}

	// Release blocker after test regardless of outcome.
	defer close(releaseCh)

	// 1ms timeout — all goroutines should hit the timeout path.
	_, err := GetAllInfoWithTimeout(provider, time.Millisecond)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to collect system information")
}
