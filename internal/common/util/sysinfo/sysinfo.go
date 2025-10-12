package sysinfo

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const EmptyString = ""

func RuntimeGoArch() string {
	return runtime.GOARCH
}

func RuntimeGoOS() string {
	return runtime.GOOS
}

func RuntimeNumCPU() int {
	return runtime.NumCPU()
}

// CPUInfo Returns VendorID, Family, Model, PhysicalID, ModelName.
func CPUInfo() (string, string, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get CPU info: %w", err)
	}
	for _, cpu := range cpuInfo {
		return cpu.VendorID, cpu.Family, cpu.PhysicalID, cpu.ModelName, nil
	}
	return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("no CPU info")
}

func RAMSize() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vmStats, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get RAM info: %w", err)
	}
	return vmStats.Total, nil
}

func OSHostname() (string, error) {
	return os.Hostname()
}

func HostID() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hostID, err := host.HostIDWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get host ID: %w", err)
	}
	return hostID, nil
}

// UserInfo Returns UserID, GroupID, Username.
func UserInfo() (string, string, string, error) {
	userInfo, err := user.Current()
	if err != nil {
		return EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get user info: %w", err)
	}
	return userInfo.Uid, userInfo.Gid, userInfo.Username, nil
}
