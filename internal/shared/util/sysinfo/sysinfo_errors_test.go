// Copyright (c) 2025 Justin Cranford
//

package sysinfo

import (
	"context"
	"errors"
	"os/user"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/require"
)

var errInjected = errors.New("injected failure")

// TestCPUInfo_Success tests the happy path when gopsutil returns valid CPU info.
func TestCPUInfo_Success(t *testing.T) {
	t.Parallel()

	vendorID, family, physicalID, modelName, err := cpuInfoWithFn(context.Background(), func(_ context.Context) ([]cpu.InfoStat, error) {
		return []cpu.InfoStat{{
			VendorID:   cryptoutilSharedMagic.MockCPUVendorID,
			Family:     cryptoutilSharedMagic.MockCPUFamily,
			PhysicalID: cryptoutilSharedMagic.MockCPUModel,
			ModelName:  cryptoutilSharedMagic.MockCPUModelName,
		}}, nil
	})
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.MockCPUVendorID, vendorID)
	require.Equal(t, cryptoutilSharedMagic.MockCPUFamily, family)
	require.Equal(t, cryptoutilSharedMagic.MockCPUModel, physicalID)
	require.Equal(t, cryptoutilSharedMagic.MockCPUModelName, modelName)
}

// TestCPUInfo_Error tests error path when gopsutil fails.
func TestCPUInfo_Error(t *testing.T) {
	t.Parallel()

	_, _, _, _, err := cpuInfoWithFn(context.Background(), func(_ context.Context) ([]cpu.InfoStat, error) {
		return nil, errInjected
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get CPU info")
}

// TestCPUInfo_NoCPU tests the "no CPU info" error path when gopsutil returns empty slice.
func TestCPUInfo_NoCPU(t *testing.T) {
	t.Parallel()

	_, _, _, _, err := cpuInfoWithFn(context.Background(), func(_ context.Context) ([]cpu.InfoStat, error) {
		return []cpu.InfoStat{}, nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no CPU info")
}

// TestRAMSize_Error tests error path when gopsutil fails.
func TestRAMSize_Error(t *testing.T) {
	t.Parallel()

	_, err := ramSizeWithFn(context.Background(), func(_ context.Context) (*mem.VirtualMemoryStat, error) {
		return nil, errInjected
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get RAM info")
}

// TestOSHostname_Error tests error path when os.Hostname fails.
func TestOSHostname_Error(t *testing.T) {
	t.Parallel()

	_, err := osHostnameWithFn(func() (string, error) {
		return "", errInjected
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get OS hostname")
}

// TestHostID_Error tests error path when gopsutil fails.
func TestHostID_Error(t *testing.T) {
	t.Parallel()

	_, err := hostIDWithFn(context.Background(), func(_ context.Context) (string, error) {
		return "", errInjected
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get host ID")
}

// TestUserInfo_Error tests error path when user.Current fails.
func TestUserInfo_Error(t *testing.T) {
	t.Parallel()

	_, _, _, err := userInfoWithFn(func() (*user.User, error) {
		return nil, errInjected
	})
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
