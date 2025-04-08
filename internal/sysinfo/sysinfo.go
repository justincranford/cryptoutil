package sysinfo

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	"os/user"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

var (
	onceRuntimeNumCPU   sync.Once
	cachedRuntimeNumCPU uint16

	onceGoArch   sync.Once
	cachedGoArch *string

	onceGoOS   sync.Once
	cachedGoOS *string

	onceCPUInfo      sync.Once
	cachedCPUInfo    *string
	cachedCPUInfoErr error

	onceRAMSize      sync.Once
	cachedRAMSize    uint64
	cachedRAMSizeErr error

	onceDiskSize      sync.Once
	cachedDiskSize    *string
	cachedDiskSizeErr error

	onceOSHostname      sync.Once
	cachedOSHostname    *string
	cachedOSHostnameErr error

	onceHostInfo                   sync.Once
	cachedHostname                 *string
	cachedHostOs                   *string
	cachedHostPlatform             *string
	cachedHostPlatformFamily       *string
	cachedHostPlatformVersion      *string
	cachedHostKernelVersion        *string
	cachedHostKernelArch           *string
	cachedHostVirtualizationSystem *string
	cachedHostVirtualizationRole   *string
	cachedHostHostID               *string
	cachedHostInfoErr              error

	onceUserStat      sync.Once
	cachedUsername    *string
	cachedUserName    *string
	cachedUserUid     *string
	cachedUserGid     *string
	cachedUserHomeDir *string
	cachedUserErr     error
)

func RuntimeNumCPU(cached bool) uint16 {
	if cached {
		onceRuntimeNumCPU.Do(func() {
			cachedRuntimeNumCPU = runtimeNumCPUFunc()
		})
		return cachedRuntimeNumCPU
	}
	uncachedRuntimeNumCPU := runtimeNumCPUFunc()
	onceRuntimeNumCPU.Do(func() {
		cachedRuntimeNumCPU = uncachedRuntimeNumCPU
	})
	return uncachedRuntimeNumCPU
}

func GoArch(cached bool) *string {
	if cached {
		onceGoArch.Do(func() {
			cachedGoArch = runtimeGoArchFunc()
		})
		return cachedGoArch
	}
	uncachedGoArch := runtimeGoArchFunc()
	onceGoArch.Do(func() {
		cachedGoArch = uncachedGoArch
	})
	return uncachedGoArch
}

func GoOS(cached bool) *string {
	if cached {
		onceGoOS.Do(func() {
			cachedGoOS = runtimeGoOSFunc()
		})
		return cachedGoOS
	}
	uncachedGoOS := runtimeGoOSFunc()
	onceGoOS.Do(func() {
		cachedGoOS = uncachedGoOS
	})
	return uncachedGoOS
}

func CPUInfo(cached bool) (*string, error) {
	if cached {
		onceCPUInfo.Do(func() {
			cachedCPUInfo, cachedCPUInfoErr = cpuInfoFunc()
		})
		return cachedCPUInfo, cachedCPUInfoErr
	}
	uncachedCPUInfo, err := cpuInfoFunc()
	onceCPUInfo.Do(func() {
		cachedCPUInfo, cachedCPUInfoErr = uncachedCPUInfo, err
	})
	return uncachedCPUInfo, err
}

func RAMSize(cached bool) (uint64, error) {
	if cached {
		onceRAMSize.Do(func() {
			cachedRAMSize, cachedRAMSizeErr = virtualMemoryFunc()
		})
		return cachedRAMSize, cachedRAMSizeErr
	}
	uncachedRAMSize, err := virtualMemoryFunc()
	onceRAMSize.Do(func() {
		cachedRAMSize, cachedRAMSizeErr = uncachedRAMSize, err
	})
	return uncachedRAMSize, err
}

func DiskSize(cached bool) (*string, error) {
	if cached {
		onceDiskSize.Do(func() {
			cachedDiskSize, cachedDiskSizeErr = diskUsageFunc()
		})
		return cachedDiskSize, cachedDiskSizeErr
	}
	uncachedDiskSize, err := diskUsageFunc()
	onceDiskSize.Do(func() {
		cachedDiskSize, cachedDiskSizeErr = uncachedDiskSize, err
	})
	return uncachedDiskSize, err
}

func OSHostname(cached bool) (*string, error) {
	if cached {
		onceOSHostname.Do(func() {
			cachedOSHostname, cachedOSHostnameErr = osHostnameFunc()
		})
		return cachedOSHostname, cachedOSHostnameErr
	}
	uncachedOSHostname, err := osHostnameFunc()
	onceOSHostname.Do(func() {
		cachedOSHostname, cachedOSHostnameErr = uncachedOSHostname, err
	})
	return uncachedOSHostname, err
}

func HostInfo(cached bool) (*string, *string, *string, *string, *string, *string, *string, *string, *string, *string, error) {
	if cached {
		onceHostInfo.Do(func() {
			cachedHostname, cachedHostOs, cachedHostPlatform, cachedHostPlatformFamily, cachedHostPlatformVersion, cachedHostKernelVersion, cachedHostKernelArch, cachedHostVirtualizationSystem, cachedHostVirtualizationRole, cachedHostHostID, cachedHostInfoErr = hostInfoFunc()
		})
		return cachedHostname, cachedHostOs, cachedHostPlatform, cachedHostPlatformFamily, cachedHostPlatformVersion, cachedHostKernelVersion, cachedHostKernelArch, cachedHostVirtualizationSystem, cachedHostVirtualizationRole, cachedHostHostID, cachedHostInfoErr
	}
	uncachedHostname, uncachedHostOs, uncachedHostPlatform, uncachedHostPlatformFamily, uncachedHostPlatformVersion, uncachedHostKernelVersion, uncachedHostKernelArch, uncachedHostVirtualizationSystem, uncachedHostVirtualizationRole, uncachedHostHostID, err := hostInfoFunc()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	onceHostInfo.Do(func() {
		cachedHostname, cachedHostOs, cachedHostPlatform, cachedHostPlatformFamily, cachedHostPlatformVersion, cachedHostKernelVersion, cachedHostKernelArch, cachedHostVirtualizationSystem, cachedHostVirtualizationRole, cachedHostHostID, cachedHostInfoErr = uncachedHostname, uncachedHostOs, uncachedHostPlatform, uncachedHostPlatformFamily, uncachedHostPlatformVersion, uncachedHostKernelVersion, uncachedHostKernelArch, uncachedHostVirtualizationSystem, uncachedHostVirtualizationRole, uncachedHostHostID, err
	})
	return uncachedHostname, uncachedHostOs, uncachedHostPlatform, uncachedHostPlatformFamily, uncachedHostPlatformVersion, uncachedHostKernelVersion, uncachedHostKernelArch, uncachedHostVirtualizationSystem, uncachedHostVirtualizationRole, uncachedHostHostID, err

}

func UserStat(cached bool) (*string, *string, *string, *string, *string, error) {
	if cached {
		onceUserStat.Do(func() {
			cachedUsername, cachedUserName, cachedUserUid, cachedUserGid, cachedUserHomeDir, cachedUserErr = userCurrentFunc()
		})
		return cachedUsername, cachedUserName, cachedUserUid, cachedUserGid, cachedUserHomeDir, cachedUserErr
	}
	uncachedUsername, uncachedUserName, uncachedUserUid, uncachedUserGid, uncachedUserHomeDir, err := userCurrentFunc()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	onceUserStat.Do(func() {
		cachedUsername, cachedUserName, cachedUserUid, cachedUserGid, cachedUserHomeDir, cachedUserErr = uncachedUsername, uncachedUserName, uncachedUserUid, uncachedUserGid, uncachedUserHomeDir, err
	})
	return uncachedUsername, uncachedUserName, uncachedUserUid, uncachedUserGid, uncachedUserHomeDir, err
}

var runtimeNumCPUFunc = func() uint16 {
	return uint16(runtime.NumCPU())
}

var runtimeGoArchFunc = func() *string {
	arch := runtime.GOARCH
	return &arch
}

var runtimeGoOSFunc = func() *string {
	os := runtime.GOOS
	return &os
}

var cpuInfoFunc = func() (*string, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}
	var details string
	for _, ci := range cpuInfo {
		details += fmt.Sprintf("CPU=%d, VendorID=%s, Family=%s, Model=%s, Stepping=%d, PhysicalID=%s, CoreID=%s, Cores=%d, ModelName=%s, Mhz=%f, CacheSize=%d, Flags=%v, Microcode=%s\n", ci.CPU, ci.VendorID, ci.Family, ci.Model, ci.Stepping, ci.PhysicalID, ci.CoreID, ci.Cores, ci.ModelName, ci.Mhz, ci.CacheSize, ci.Flags, ci.Microcode)
	}
	return &details, nil
}

var virtualMemoryFunc = func() (uint64, error) {
	vmStats, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("failed to get RAM info: %w", err)
	}
	return vmStats.Total, nil
}

var diskUsageFunc = func() (*string, error) {
	diskStats, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get Disk info: %w", err)
	}
	details := strconv.FormatUint(diskStats.Total, 10)
	return &details, nil
}

var osHostnameFunc = func() (*string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &hostname, nil
}

var hostInfoFunc = func() (*string, *string, *string, *string, *string, *string, *string, *string, *string, *string, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to get Host info: %w", err)
	}
	return &hostInfo.Hostname, &hostInfo.OS, &hostInfo.Platform, &hostInfo.PlatformFamily, &hostInfo.PlatformVersion, &hostInfo.KernelVersion, &hostInfo.KernelArch, &hostInfo.VirtualizationSystem, &hostInfo.VirtualizationRole, &hostInfo.HostID, nil
}

var userCurrentFunc = func() (*string, *string, *string, *string, *string, error) {
	userInfo, err := user.Current()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get Host info: %w", err)
	}
	return &userInfo.Username, &userInfo.Name, &userInfo.Uid, &userInfo.Gid, &userInfo.HomeDir, nil
}
