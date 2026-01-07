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

type MockSysInfoProvider struct{}

func (mock *MockSysInfoProvider) RuntimeGoArch() string {
	return mockGoArch
}

func (mock *MockSysInfoProvider) RuntimeGoOS() string {
	return mockGoOS
}

func (mock *MockSysInfoProvider) RuntimeNumCPU() int {
	return mockNumCPU
}

func (mock *MockSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return mockCPUVendor, mockCPUFamily, mockCPUModel, mockCPUName, nil
}

func (mock *MockSysInfoProvider) RAMSize() (uint64, error) {
	return mockRAMSizeMB, nil
}

func (mock *MockSysInfoProvider) OSHostname() (string, error) {
	return mockHostname, nil
}

func (mock *MockSysInfoProvider) HostID() (string, error) {
	return mockHostID, nil
}

func (mock *MockSysInfoProvider) UserInfo() (string, string, string, error) {
	return mockUserID, mockGroupID, mockUsername, nil
}
