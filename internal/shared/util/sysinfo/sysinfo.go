// Copyright (c) 2025 Justin Cranford
//
//

// Package sysinfo provides system information utilities for runtime environment detection.
package sysinfo

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const (
	// EmptyString is a constant for an empty string.
	EmptyString = cryptoutilSharedMagic.EmptyString

	// Timeouts for system info queries to prevent hanging.
	cpuInfoTimeout = cryptoutilSharedMagic.DefaultSysInfoCPUTimeout
	memoryTimeout  = cryptoutilSharedMagic.DefaultSysInfoMemoryTimeout
	hostIDTimeout  = cryptoutilSharedMagic.DefaultSysInfoHostTimeout
)

// RuntimeGoArch returns the GOARCH runtime architecture.
func RuntimeGoArch() string {
	return runtime.GOARCH
}

// RuntimeGoOS returns the GOOS runtime operating system.
func RuntimeGoOS() string {
	return runtime.GOOS
}

// RuntimeNumCPU returns the number of CPUs available.
func RuntimeNumCPU() int {
	return runtime.NumCPU()
}

// sysinfoFns holds injectable function vars for testing OS-level calls.
var (
	sysinfoGetCPUInfoFn           = func(ctx context.Context) ([]cpu.InfoStat, error) { return cpu.InfoWithContext(ctx) }
	sysinfoGetVirtualMemoryFn     = func(ctx context.Context) (*mem.VirtualMemoryStat, error) { return mem.VirtualMemoryWithContext(ctx) }
	sysinfoGetOSHostnameFn        = os.Hostname
	sysinfoGetHostIDFn            = func(ctx context.Context) (string, error) { return host.HostIDWithContext(ctx) }
	sysinfoGetUserCurrentFn       = user.Current
)

// CPUInfo Returns VendorID, Family, Model, PhysicalID, ModelName.
func CPUInfo() (string, string, string, string, error) {
        ctx, cancel := context.WithTimeout(context.Background(), cpuInfoTimeout)
        defer cancel()

        cpuInfo, err := sysinfoGetCPUInfoFn(ctx)
        if err != nil {
                return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get CPU info: %w", err)
        }

        for _, cpu := range cpuInfo {
                return cpu.VendorID, cpu.Family, cpu.PhysicalID, cpu.ModelName, nil
        }

        return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("no CPU info")
}

// RAMSize returns the total RAM size in bytes.
func RAMSize() (uint64, error) {
        ctx, cancel := context.WithTimeout(context.Background(), memoryTimeout)
        defer cancel()

        vmStats, err := sysinfoGetVirtualMemoryFn(ctx)
        if err != nil {
                return 0, fmt.Errorf("failed to get RAM info: %w", err)
        }

        return vmStats.Total, nil
}

// OSHostname returns the OS hostname.
func OSHostname() (string, error) {
        hostname, err := sysinfoGetOSHostnameFn()
        if err != nil {
                return "", fmt.Errorf("failed to get OS hostname: %w", err)
        }

        return hostname, nil
}

// HostID returns the unique host identifier.
func HostID() (string, error) {
        ctx, cancel := context.WithTimeout(context.Background(), hostIDTimeout)
        defer cancel()

        hostID, err := sysinfoGetHostIDFn(ctx)
        if err != nil {
                return "", fmt.Errorf("failed to get host ID: %w", err)
        }

        return hostID, nil
}

// UserInfo Returns UserID, GroupID, Username.
func UserInfo() (string, string, string, error) {
        userInfo, err := sysinfoGetUserCurrentFn()
	if err != nil {
		return EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo.Uid, userInfo.Gid, userInfo.Username, nil
}
