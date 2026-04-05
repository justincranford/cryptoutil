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

// cpuInfoWithFn returns CPU info using the provided context-aware fn (function-param injection).
func cpuInfoWithFn(ctx context.Context, fn func(context.Context) ([]cpu.InfoStat, error)) (string, string, string, string, error) {
	cpuInfo, err := fn(ctx)
	if err != nil {
		return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get CPU info: %w", err)
	}

	for _, c := range cpuInfo {
		return c.VendorID, c.Family, c.PhysicalID, c.ModelName, nil
	}

	return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("no CPU info")
}

// CPUInfo returns VendorID, Family, Model, PhysicalID, ModelName.
func CPUInfo() (string, string, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSysInfoCPUTimeout)
	defer cancel()

	return cpuInfoWithFn(ctx, func(c context.Context) ([]cpu.InfoStat, error) { return cpu.InfoWithContext(c) })
}

// ramSizeWithFn returns RAM size using the provided context-aware fn (function-param injection).
func ramSizeWithFn(ctx context.Context, fn func(context.Context) (*mem.VirtualMemoryStat, error)) (uint64, error) {
	vmStats, err := fn(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get RAM info: %w", err)
	}

	return vmStats.Total, nil
}

// RAMSize returns the total RAM size in bytes.
func RAMSize() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSysInfoMemoryTimeout)
	defer cancel()

	return ramSizeWithFn(ctx, func(c context.Context) (*mem.VirtualMemoryStat, error) { return mem.VirtualMemoryWithContext(c) })
}

// osHostnameWithFn returns the OS hostname using the provided fn (function-param injection).
func osHostnameWithFn(fn func() (string, error)) (string, error) {
	hostname, err := fn()
	if err != nil {
		return "", fmt.Errorf("failed to get OS hostname: %w", err)
	}

	return hostname, nil
}

// OSHostname returns the OS hostname.
func OSHostname() (string, error) {
	return osHostnameWithFn(os.Hostname)
}

// hostIDWithFn returns the host ID using the provided context-aware fn (function-param injection).
func hostIDWithFn(ctx context.Context, fn func(context.Context) (string, error)) (string, error) {
	hostID, err := fn(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get host ID: %w", err)
	}

	return hostID, nil
}

// HostID returns the unique host identifier.
func HostID() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSysInfoHostTimeout)
	defer cancel()

	return hostIDWithFn(ctx, func(c context.Context) (string, error) { return host.HostIDWithContext(c) })
}

// userInfoWithFn returns user info using the provided fn (function-param injection).
func userInfoWithFn(fn func() (*user.User, error)) (string, string, string, error) {
	userInfo, err := fn()
	if err != nil {
		return EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo.Uid, userInfo.Gid, userInfo.Username, nil
}

// UserInfo returns UserID, GroupID, Username.
func UserInfo() (string, string, string, error) {
	return userInfoWithFn(user.Current)
}
