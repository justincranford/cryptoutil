// Copyright (c) 2025-2026 Justin Cranford.
//
//

package sysinfo

var defaultSysInfoProvider = &DefaultSysInfoProvider{}

// DefaultSysInfoProvider is the default implementation of SysInfoProvider.
type DefaultSysInfoProvider struct {
	cpuInfoFn    func() (string, string, string, string, error)
	ramSizeFn    func() (uint64, error)
	osHostnameFn func() (string, error)
	hostIDFn     func() (string, error)
	userInfoFn   func() (string, string, string, error)
}

// RuntimeGoArch returns the GOARCH runtime architecture.
func (sp *DefaultSysInfoProvider) RuntimeGoArch() string {
	return RuntimeGoArch()
}

// RuntimeGoOS returns the GOOS runtime operating system.
func (sp *DefaultSysInfoProvider) RuntimeGoOS() string {
	return RuntimeGoOS()
}

// RuntimeNumCPU returns the number of CPUs available.
func (sp *DefaultSysInfoProvider) RuntimeNumCPU() int {
	return RuntimeNumCPU()
}

// CPUInfo returns CPU information including VendorID, Family, PhysicalID, and ModelName.
func (sp *DefaultSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	if sp.cpuInfoFn != nil {
		return sp.cpuInfoFn()
	}

	return CPUInfo()
}

// RAMSize returns the total RAM size in bytes.
func (sp *DefaultSysInfoProvider) RAMSize() (uint64, error) {
	if sp.ramSizeFn != nil {
		return sp.ramSizeFn()
	}

	return RAMSize()
}

// OSHostname returns the OS hostname.
func (sp *DefaultSysInfoProvider) OSHostname() (string, error) {
	if sp.osHostnameFn != nil {
		return sp.osHostnameFn()
	}

	return OSHostname()
}

// HostID returns the unique host identifier.
func (sp *DefaultSysInfoProvider) HostID() (string, error) {
	if sp.hostIDFn != nil {
		return sp.hostIDFn()
	}

	return HostID()
}

// UserInfo returns user information including UserID, GroupID, and Username.
func (sp *DefaultSysInfoProvider) UserInfo() (string, string, string, error) {
	if sp.userInfoFn != nil {
		return sp.userInfoFn()
	}

	return UserInfo()
}
