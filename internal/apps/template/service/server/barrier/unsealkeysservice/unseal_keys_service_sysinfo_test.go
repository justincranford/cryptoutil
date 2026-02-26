// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedUtilSysinfo "cryptoutil/internal/shared/util/sysinfo"

	"github.com/stretchr/testify/require"
)

func TestNewUnsealKeysServiceFromSysInfo_HappyPath(t *testing.T) {
	t.Parallel()

	mockProvider := &cryptoutilSharedUtilSysinfo.MockSysInfoProvider{}

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(mockProvider)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)

	// Verify the service can encrypt and decrypt keys.
	testKey, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 1, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	encrypted, err := unsealKeysService.EncryptKey(testKey[0])
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := unsealKeysService.DecryptKey(encrypted)
	require.NoError(t, err)
	require.NotNil(t, decrypted)
}

func TestNewUnsealKeysServiceFromSysInfo_EncryptDecryptData(t *testing.T) {
	t.Parallel()

	mockProvider := &cryptoutilSharedUtilSysinfo.MockSysInfoProvider{}

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(mockProvider)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)

	testData := []byte("test data to encrypt")

	encrypted, err := unsealKeysService.EncryptData(testData)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := unsealKeysService.DecryptData(encrypted)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted)
}

func TestNewUnsealKeysServiceFromSysInfo_Shutdown(t *testing.T) {
	t.Parallel()

	mockProvider := &cryptoutilSharedUtilSysinfo.MockSysInfoProvider{}

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(mockProvider)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)

	// Shutdown should not panic.
	unsealKeysService.Shutdown()
}

// ErrorMockSysInfoProvider always returns an error for GetAllInfo.
type ErrorMockSysInfoProvider struct{}

func (e *ErrorMockSysInfoProvider) RuntimeGoArch() string {
	return cryptoutilSharedMagic.MockRuntimeGoArch
}

func (e *ErrorMockSysInfoProvider) RuntimeGoOS() string {
	return cryptoutilSharedMagic.MockRuntimeGoOS
}

func (e *ErrorMockSysInfoProvider) RuntimeNumCPU() int {
	return 4
}

func (e *ErrorMockSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return "", "", "", "", nil
}

func (e *ErrorMockSysInfoProvider) RAMSize() (uint64, error) {
	return 0, nil
}

func (e *ErrorMockSysInfoProvider) OSHostname() (string, error) {
	return "", nil
}

func (e *ErrorMockSysInfoProvider) HostID() (string, error) {
	return "", nil
}

func (e *ErrorMockSysInfoProvider) UserInfo() (string, string, string, error) {
	return "", "", "", nil
}

func TestNewUnsealKeysServiceFromSysInfo_WithEmptyValues(t *testing.T) {
	t.Parallel()

	// Test with a provider that returns minimal/empty values.
	// The service should still work as long as it can create JWKs.
	errorProvider := &ErrorMockSysInfoProvider{}

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(errorProvider)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

// EmptySysInfoProvider returns empty data to test empty handling.
type EmptySysInfoProvider struct{}

func (e *EmptySysInfoProvider) RuntimeGoArch() string {
	return ""
}

func (e *EmptySysInfoProvider) RuntimeGoOS() string {
	return ""
}

func (e *EmptySysInfoProvider) RuntimeNumCPU() int {
	return 0
}

func (e *EmptySysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return "", "", "", "", nil
}

func (e *EmptySysInfoProvider) RAMSize() (uint64, error) {
	return 0, nil
}

func (e *EmptySysInfoProvider) OSHostname() (string, error) {
	return "", nil
}

func (e *EmptySysInfoProvider) HostID() (string, error) {
	return "", nil
}

func (e *EmptySysInfoProvider) UserInfo() (string, string, string, error) {
	return "", "", "", nil
}

func TestNewUnsealKeysServiceFromSysInfo_MinimalData(t *testing.T) {
	t.Parallel()

	// Test with truly minimal data - this tests edge case of empty sysinfo.
	emptyProvider := &EmptySysInfoProvider{}

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(emptyProvider)

	// Should still succeed since the runtime values will be used.
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}
