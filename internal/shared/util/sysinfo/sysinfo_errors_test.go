// Copyright (c) 2025-2026 Justin Cranford.
//

package sysinfo

import (
	"context"
	"errors"
	"os/user"
	"strconv"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/require"
)

const sysinfoInjectedFailureMessage = "injected failure"

var errInjected = errors.New(sysinfoInjectedFailureMessage)

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

func runSysInfoErrorTests(t *testing.T) {
	t.Helper()

	tests := []struct {
		name        string
		wantErr     string
		wantValues  []string
		collectInfo func() ([]string, error)
	}{
		{
			name:       "ram_size_error_handling",
			wantErr:    "failed to get RAM info",
			wantValues: []string{"0"},
			collectInfo: func() ([]string, error) {
				size, err := ramSizeWithFn(context.Background(), func(_ context.Context) (*mem.VirtualMemoryStat, error) {
					return nil, errInjected
				})

				return []string{strconv.FormatUint(size, 10)}, err
			},
		},
		{
			name:       "os_hostname_error_handling",
			wantErr:    "failed to get OS hostname",
			wantValues: []string{EmptyString},
			collectInfo: func() ([]string, error) {
				hostname, err := osHostnameWithFn(func() (string, error) {
					return EmptyString, errInjected
				})

				return []string{hostname}, err
			},
		},
		{
			name:       "host_id_error_handling",
			wantErr:    "failed to get host ID",
			wantValues: []string{EmptyString},
			collectInfo: func() ([]string, error) {
				hostID, err := hostIDWithFn(context.Background(), func(_ context.Context) (string, error) {
					return EmptyString, errInjected
				})

				return []string{hostID}, err
			},
		},
		{
			name:       "user_info_error_handling",
			wantErr:    "failed to get user info",
			wantValues: []string{EmptyString, EmptyString, EmptyString},
			collectInfo: func() ([]string, error) {
				userID, groupID, username, err := userInfoWithFn(func() (*user.User, error) {
					return nil, errInjected
				})

				return []string{userID, groupID, username}, err
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			values, err := tc.collectInfo()
			require.ErrorIs(t, err, errInjected)
			require.ErrorContains(t, err, tc.wantErr)
			require.Equal(t, tc.wantValues, values)
		})
	}
}

func TestGetAllInfoWithTimeout_ProviderErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		inject     func(*DefaultSysInfoProvider)
		wantErrMsg string
	}{
		{
			name: "cpu info error",
			inject: func(provider *DefaultSysInfoProvider) {
				provider.cpuInfoFn = func() (string, string, string, string, error) {
					return "", "", "", "", errInjected
				}
			},
			wantErrMsg: sysinfoInjectedFailureMessage,
		},
		{
			name: "ram size error",
			inject: func(provider *DefaultSysInfoProvider) {
				provider.ramSizeFn = func() (uint64, error) {
					return 0, errInjected
				}
			},
			wantErrMsg: sysinfoInjectedFailureMessage,
		},
		{
			name: "hostname error",
			inject: func(provider *DefaultSysInfoProvider) {
				provider.osHostnameFn = func() (string, error) {
					return "", errInjected
				}
			},
			wantErrMsg: sysinfoInjectedFailureMessage,
		},
		{
			name: "host ID error",
			inject: func(provider *DefaultSysInfoProvider) {
				provider.hostIDFn = func() (string, error) {
					return "", errInjected
				}
			},
			wantErrMsg: sysinfoInjectedFailureMessage,
		},
		{
			name: "user info error",
			inject: func(provider *DefaultSysInfoProvider) {
				provider.userInfoFn = func() (string, string, string, error) {
					return "", "", "", errInjected
				}
			},
			wantErrMsg: sysinfoInjectedFailureMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			provider := newPortableDefaultSysInfoProvider()
			tc.inject(provider)

			_, err := GetAllInfoWithTimeout(provider, cryptoutilSharedMagic.DefaultSysInfoAllTimeout)
			require.ErrorIs(t, err, errInjected)
			require.ErrorContains(t, err, tc.wantErrMsg)
		})
	}
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

// TestSysInfoAll_Timeout verifies that provider calls exceeding the deadline report timeouts.
func TestSysInfoAll_Timeout(t *testing.T) {
	t.Parallel()

	releaseCh := make(chan struct{})
	provider := &blockingProviderForTimeout{releaseCh: releaseCh}

	// Release blocker after test regardless of outcome.
	defer close(releaseCh)

	// 1ms timeout — all goroutines should hit the timeout path.
	_, err := GetAllInfoWithTimeout(provider, time.Millisecond)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to collect system information")
	require.ErrorIs(t, err, context.DeadlineExceeded)

	for _, timeoutMessage := range []string{
		"CPU info timeout",
		"RAM info timeout",
		"hostname info timeout",
		"host ID info timeout",
		"user info timeout",
	} {
		require.ErrorContains(t, err, timeoutMessage)
	}
}
