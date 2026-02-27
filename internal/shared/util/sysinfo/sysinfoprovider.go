// Copyright (c) 2025 Justin Cranford
//
//

package sysinfo

// SysInfoProvider defines the interface for retrieving system information.
type SysInfoProvider interface {
	RuntimeGoArch() string
	RuntimeGoOS() string
	RuntimeNumCPU() int
	CPUInfo() (string, string, string, string, error)
	RAMSize() (uint64, error)
	OSHostname() (string, error)
	HostID() (string, error)
	UserInfo() (string, string, string, error)
}
