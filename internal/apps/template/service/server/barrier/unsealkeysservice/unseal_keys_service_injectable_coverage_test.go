// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"context"
	"fmt"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilSysinfo "cryptoutil/internal/shared/util/sysinfo"

	"github.com/stretchr/testify/require"
)

const (
	testGoArch = "amd64"
	testGoOS   = "linux"
)

// failingSysInfoProvider is a SysInfoProvider that returns an error from CPUInfo.
type failingSysInfoProvider struct{}

func (f *failingSysInfoProvider) RuntimeGoArch() string { return testGoArch }

func (f *failingSysInfoProvider) RuntimeGoOS() string { return testGoOS }

func (f *failingSysInfoProvider) RuntimeNumCPU() int { return 1 }

func (f *failingSysInfoProvider) CPUInfo() (string, string, string, string, error) {
	return "", "", "", "", fmt.Errorf("simulated CPU info error")
}

func (f *failingSysInfoProvider) RAMSize() (uint64, error) { return 0, nil }

func (f *failingSysInfoProvider) OSHostname() (string, error) { return "localhost", nil }

func (f *failingSysInfoProvider) HostID() (string, error) { return "host-id", nil }

func (f *failingSysInfoProvider) UserInfo() (string, string, string, error) {
	return "uid", "gid", "username", nil
}

// Verify failingSysInfoProvider implements SysInfoProvider.
var _ cryptoutilSharedUtilSysinfo.SysInfoProvider = (*failingSysInfoProvider)(nil)

// TestNewUnsealKeysServiceFromSettings_VerboseMode_InvalidMode tests VerboseMode logging
// with DevMode=false and an invalid UnsealMode that causes early error return.
// This covers the VerboseMode slogger path in from_settings.go.
func TestNewUnsealKeysServiceFromSettings_VerboseMode_InvalidMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-verbose-invalid")
	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	telemetryService.VerboseMode = true

	settings.DevMode = false
	settings.UnsealMode = "invalid-unseal-mode"

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "invalid unseal mode")
}

// TestNewUnsealKeysServiceFromSysInfo_FailingProvider tests the sysinfo error path.
// Uses a provider that returns an error, covering the GetAllInfoWithTimeout error branch.
func TestNewUnsealKeysServiceFromSysInfo_FailingProvider(t *testing.T) {
	t.Parallel()

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(&failingSysInfoProvider{})

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to get sysinfo")
}

// TestNewUnsealKeysServiceSharedSecrets_HKDFError tests the HKDF error path in
// deriveJWKsFromMChooseNCombinations and the corresponding error in NewUnsealKeysServiceSharedSecrets.
//
// NOTE: Must NOT use t.Parallel() - modifies package-level hkdfWithSHA256Fn.
func TestNewUnsealKeysServiceSharedSecrets_HKDFError(t *testing.T) {
	originalFn := hkdfWithSHA256Fn
	hkdfWithSHA256Fn = func(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
		return nil, fmt.Errorf("simulated HKDF error")
	}

	defer func() { hkdfWithSHA256Fn = originalFn }()

	sharedSecret := make([]byte, cryptoutilSharedMagic.MinSharedSecretLength)
	sharedSecretsM := [][]byte{sharedSecret}

	unsealKeysService, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 1)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to create unseal JWK combinations")
}

// TestNewUnsealKeysServiceFromSysInfo_HKDFError tests the HKDF error path in
// deriveJWKsFromMChooseNCombinations when called from the sysinfo path.
//
// NOTE: Must NOT use t.Parallel() - modifies package-level hkdfWithSHA256Fn.
func TestNewUnsealKeysServiceFromSysInfo_HKDFError(t *testing.T) {
	originalFn := hkdfWithSHA256Fn
	hkdfWithSHA256Fn = func(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
		return nil, fmt.Errorf("simulated HKDF error")
	}

	defer func() { hkdfWithSHA256Fn = originalFn }()

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(&cryptoutilSharedUtilSysinfo.DefaultSysInfoProvider{})

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to create unseal JWKs")
}
