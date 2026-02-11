// Copyright (c) 2025 Justin Cranford
//
//

package sysinfo

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// Mock values for testing.
	mockNumCPU    = cryptoutilSharedMagic.MockCPUCount
	mockRAMSizeMB = cryptoutilSharedMagic.MockRAMMB
	mockCPUFamily = cryptoutilSharedMagic.MockCPUFamily
	mockCPUModel  = cryptoutilSharedMagic.MockCPUModel
	mockGoArch    = cryptoutilSharedMagic.MockRuntimeGoArch
	mockGoOS      = cryptoutilSharedMagic.MockRuntimeGoOS
	mockCPUVendor = cryptoutilSharedMagic.MockCPUVendorID
	mockCPUName   = cryptoutilSharedMagic.MockCPUModelName
	mockHostname  = cryptoutilSharedMagic.MockHostname
	mockHostID    = cryptoutilSharedMagic.MockHostID
	mockUserID    = cryptoutilSharedMagic.MockUserID
	mockGroupID   = cryptoutilSharedMagic.MockGroupID
	mockUsername  = cryptoutilSharedMagic.MockUsername
)

var mockSysInfoProvider = &MockSysInfoProvider{}

// MockSysInfoProvider is a mock implementation of the SysInfoProvider interface for testing.
type MockSysInfoProvider struct{}

// RuntimeGoArch returns a mock Go architecture string.
func (mock *MockSysInfoProvider) RuntimeGoArch() string {
	return mockGoArch
}

// RuntimeGoOS returns a mock Go operating system string.
func (mock *MockSysInfoProvider) RuntimeGoOS() string {
	return mockGoOS
}

// RuntimeNumCPU returns a mock number of CPUs.
func (mock *MockSysInfoProvider) RuntimeNumCPU() int {
	return mockNumCPU
}

// CPUInfo returns mock CPU information.
func (mock *MockSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return mockCPUVendor, mockCPUFamily, mockCPUModel, mockCPUName, nil
}

// RAMSize returns a mock RAM size in megabytes.
func (mock *MockSysInfoProvider) RAMSize() (uint64, error) {
	return mockRAMSizeMB, nil
}

// OSHostname returns a mock hostname.
func (mock *MockSysInfoProvider) OSHostname() (string, error) {
	return mockHostname, nil
}

// HostID returns a mock host identifier.
func (mock *MockSysInfoProvider) HostID() (string, error) {
	return mockHostID, nil
}

// UserInfo returns mock user information.
func (mock *MockSysInfoProvider) UserInfo() (string, string, string, error) {
	return mockUserID, mockGroupID, mockUsername, nil
}
