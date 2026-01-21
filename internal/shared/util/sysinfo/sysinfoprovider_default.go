// Copyright (c) 2025 Justin Cranford
//
//

package sysinfo

var defaultSysInfoProvider = &DefaultSysInfoProvider{}

// DefaultSysInfoProvider is the default implementation of SysInfoProvider.
type DefaultSysInfoProvider struct{}

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
	return CPUInfo()
}

// RAMSize returns the total RAM size in bytes.
func (sp *DefaultSysInfoProvider) RAMSize() (uint64, error) {
	return RAMSize()
}

// OSHostname returns the OS hostname.
func (sp *DefaultSysInfoProvider) OSHostname() (string, error) {
	return OSHostname()
}

// HostID returns the unique host identifier.
func (sp *DefaultSysInfoProvider) HostID() (string, error) {
	return HostID()
}

// UserInfo returns user information including UserID, GroupID, and Username.
func (sp *DefaultSysInfoProvider) UserInfo() (string, string, string, error) {
	return UserInfo()
}
