// Copyright (c) 2025 Justin Cranford
//
//

package sysinfo

import (
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	// Mock values for testing.
	mockNumCPU    = cryptoutilMagic.MockCPUCount
	mockRAMSizeMB = cryptoutilMagic.MockRAMMB
	mockCPUFamily = cryptoutilMagic.MockCPUFamily
	mockCPUModel  = cryptoutilMagic.MockCPUModel
	mockGoArch    = cryptoutilMagic.MockRuntimeGoArch
	mockGoOS      = cryptoutilMagic.MockRuntimeGoOS
	mockCPUVendor = cryptoutilMagic.MockCPUVendorID
	mockCPUName   = cryptoutilMagic.MockCPUModelName
	mockHostname  = cryptoutilMagic.MockHostname
	mockHostID    = cryptoutilMagic.MockHostID
	mockUserID    = cryptoutilMagic.MockUserID
	mockGroupID   = cryptoutilMagic.MockGroupID
	mockUsername  = cryptoutilMagic.MockUsername
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
