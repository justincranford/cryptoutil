// Copyright (c) 2025 Justin Cranford
//
//

package sysinfo

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var mockSysInfoProvider = &MockSysInfoProvider{}

// MockSysInfoProvider is a mock implementation of the SysInfoProvider interface for testing.
type MockSysInfoProvider struct{}

// RuntimeGoArch returns a mock Go architecture string.
func (mock *MockSysInfoProvider) RuntimeGoArch() string {
	return cryptoutilSharedMagic.MockRuntimeGoArch
}

// RuntimeGoOS returns a mock Go operating system string.
func (mock *MockSysInfoProvider) RuntimeGoOS() string {
	return cryptoutilSharedMagic.MockRuntimeGoOS
}

// RuntimeNumCPU returns a mock number of CPUs.
func (mock *MockSysInfoProvider) RuntimeNumCPU() int {
	return cryptoutilSharedMagic.MockCPUCount
}

// CPUInfo returns mock CPU information.
func (mock *MockSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return cryptoutilSharedMagic.MockCPUVendorID, cryptoutilSharedMagic.MockCPUFamily, cryptoutilSharedMagic.MockCPUModel, cryptoutilSharedMagic.MockCPUModelName, nil
}

// RAMSize returns a mock RAM size in megabytes.
func (mock *MockSysInfoProvider) RAMSize() (uint64, error) {
	return cryptoutilSharedMagic.MockRAMMB, nil
}

// OSHostname returns a mock hostname.
func (mock *MockSysInfoProvider) OSHostname() (string, error) {
	return cryptoutilSharedMagic.MockHostname, nil
}

// HostID returns a mock host identifier.
func (mock *MockSysInfoProvider) HostID() (string, error) {
	return cryptoutilSharedMagic.MockHostID, nil
}

// UserInfo returns mock user information.
func (mock *MockSysInfoProvider) UserInfo() (string, string, string, error) {
	return cryptoutilSharedMagic.MockUserID, cryptoutilSharedMagic.MockGroupID, cryptoutilSharedMagic.MockUsername, nil
}
