// Copyright (c) 2025 Justin Cranford
//
//

package sysinfo

var defaultSysInfoProvider = &DefaultSysInfoProvider{}

// DefaultSysInfoProvider is the default implementation of SysInfoProvider.
type DefaultSysInfoProvider struct{}

func (sp *DefaultSysInfoProvider) RuntimeGoArch() string {
	return RuntimeGoArch()
}

func (sp *DefaultSysInfoProvider) RuntimeGoOS() string {
	return RuntimeGoOS()
}

func (sp *DefaultSysInfoProvider) RuntimeNumCPU() int {
	return RuntimeNumCPU()
}

func (sp *DefaultSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return CPUInfo()
}

func (sp *DefaultSysInfoProvider) RAMSize() (uint64, error) {
	return RAMSize()
}

func (sp *DefaultSysInfoProvider) OSHostname() (string, error) {
	return OSHostname()
}

func (sp *DefaultSysInfoProvider) HostID() (string, error) {
	return HostID()
}

func (sp *DefaultSysInfoProvider) UserInfo() (string, string, string, error) {
	return UserInfo()
}
