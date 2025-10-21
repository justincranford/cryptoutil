package sysinfo

import (
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

const (
	// Mock values for testing.
	mockNumCPU    = cryptoutilMagic.CountMockCPU
	mockRAMSizeMB = cryptoutilMagic.CountMockRAMMB
	mockCPUFamily = cryptoutilMagic.MockCPUFamily
	mockCPUModel  = cryptoutilMagic.MockCPUModel
)

var mockSysInfoProvider = &MockSysInfoProvider{}

type MockSysInfoProvider struct{}

func (mock *MockSysInfoProvider) RuntimeGoArch() string {
	return "amd64"
}

func (mock *MockSysInfoProvider) RuntimeGoOS() string {
	return "linux"
}

func (mock *MockSysInfoProvider) RuntimeNumCPU() int {
	return mockNumCPU
}

func (mock *MockSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return "GenuineIntel", mockCPUFamily, mockCPUModel, "Intel(R) Core(TM) i7-8550U", nil
}

func (mock *MockSysInfoProvider) RAMSize() (uint64, error) {
	return mockRAMSizeMB, nil
}

func (mock *MockSysInfoProvider) OSHostname() (string, error) {
	return "mock-hostname", nil
}

func (mock *MockSysInfoProvider) HostID() (string, error) {
	return "mock-host-id", nil
}

func (mock *MockSysInfoProvider) UserInfo() (string, string, string, error) {
	return "mock-user-id-1000", "mock-group-id-1000", "mock-username", nil
}
