package sysinfo

import (
	"fmt"
	"os"
	"runtime"

	"os/user"

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

// CPUInfo Returns VendorID, Family, Model, PhysicalID, ModelName
func CPUInfo() (string, string, string, string, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get CPU info: %w", err)
	}
	for _, cpu := range cpuInfo {
		return cpu.VendorID, cpu.Family, cpu.PhysicalID, cpu.ModelName, nil
	}
	return EmptyString, EmptyString, EmptyString, EmptyString, fmt.Errorf("no CPU info")
}

func RAMSize() (uint64, error) {
	vmStats, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("failed to get RAM info: %w", err)
	}
	return vmStats.Total, nil
}

func OSHostname() (string, error) {
	return os.Hostname()
}

func HostID() (string, error) {
	hostID, err := host.HostID()
	if err != nil {
		return "", fmt.Errorf("failed to get host ID: %w", err)
	}
	return hostID, nil
}

// UserInfo Returns UserID, GroupID, Username
func UserInfo() (string, string, string, error) {
	userInfo, err := user.Current()
	if err != nil {
		return EmptyString, EmptyString, EmptyString, fmt.Errorf("failed to get user info: %w", err)
	}
	return userInfo.Uid, userInfo.Gid, userInfo.Username, nil
}
