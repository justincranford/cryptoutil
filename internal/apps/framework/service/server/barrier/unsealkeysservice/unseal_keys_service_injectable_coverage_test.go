// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilSysinfo "cryptoutil/internal/shared/util/sysinfo"

	"github.com/stretchr/testify/require"
)

const (
	testGoArch = cryptoutilSharedMagic.MockRuntimeGoArch
	testGoOS   = cryptoutilSharedMagic.MockRuntimeGoOS
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

func (f *failingSysInfoProvider) OSHostname() (string, error) {
	return cryptoutilSharedMagic.DefaultOTLPHostnameDefault, nil
}

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
	settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("test-verbose-invalid")
	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings.ToTelemetrySettings())
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
// deriveJWKsFromMChooseNCombinationsInternal and the corresponding error in newUnsealKeysServiceSharedSecretsInternal.
func TestNewUnsealKeysServiceSharedSecrets_HKDFError(t *testing.T) {
	t.Parallel()

	stubHKDF := func(_, _, _ []byte, _ int) ([]byte, error) {
		return nil, fmt.Errorf("simulated HKDF error")
	}

	sharedSecret := make([]byte, cryptoutilSharedMagic.MinSharedSecretLength)
	sharedSecretsM := [][]byte{sharedSecret}

	unsealKeysService, err := newUnsealKeysServiceSharedSecretsInternal(sharedSecretsM, 1, stubHKDF)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to create unseal JWK combinations")
}

// TestNewUnsealKeysServiceFromSysInfo_HKDFError tests the HKDF error path in
// deriveJWKsFromMChooseNCombinationsInternal when called from the sysinfo path.
func TestNewUnsealKeysServiceFromSysInfo_HKDFError(t *testing.T) {
	t.Parallel()

	stubHKDF := func(_, _, _ []byte, _ int) ([]byte, error) {
		return nil, fmt.Errorf("simulated HKDF error")
	}

	unsealKeysService, err := newUnsealKeysServiceFromSysInfoInternal(&cryptoutilSharedUtilSysinfo.DefaultSysInfoProvider{}, cryptoutilSharedUtilSysinfo.GetAllInfoWithTimeout, stubHKDF)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to create unseal JWKs")
}

// TestNewUnsealKeysServiceFromSysInfo_EmptySysinfos tests the empty sysinfos error path.
// Uses fn-param injection to return empty slice, covering numSysinfos == 0 branch.
func TestNewUnsealKeysServiceFromSysInfo_EmptySysinfos(t *testing.T) {
	t.Parallel()

	stubGetAllInfo := func(_ cryptoutilSharedUtilSysinfo.SysInfoProvider, _ time.Duration) ([][]byte, error) {
		return [][]byte{}, nil
	}

	unsealKeysService, err := newUnsealKeysServiceFromSysInfoInternal(&cryptoutilSharedUtilSysinfo.DefaultSysInfoProvider{}, stubGetAllInfo, cryptoutilSharedCryptoDigests.HKDFwithSHA256)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "empty sysinfos not supported")
}

// TestNewUnsealKeysServiceFromSysInfo_SingleSysinfo tests the numSysinfos == 1 branch.
// Uses fn-param injection to return a single element, covering chooseN = 1 path.
func TestNewUnsealKeysServiceFromSysInfo_SingleSysinfo(t *testing.T) {
	t.Parallel()

	stubGetAllInfo := func(_ cryptoutilSharedUtilSysinfo.SysInfoProvider, _ time.Duration) ([][]byte, error) {
		return [][]byte{[]byte("single-sysinfo-entry-for-testing")}, nil
	}

	unsealKeysService, err := newUnsealKeysServiceFromSysInfoInternal(&cryptoutilSharedUtilSysinfo.DefaultSysInfoProvider{}, stubGetAllInfo, cryptoutilSharedCryptoDigests.HKDFwithSHA256)

	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

// TestDeriveJWKs_SecondHKDFError tests the second HKDF error path in deriveJWKsFromMChooseNCombinationsInternal.
// Uses call-count injection: first HKDF call (for KID) succeeds, second (for secret) fails.
func TestDeriveJWKs_SecondHKDFError(t *testing.T) {
	t.Parallel()

	callCount := 0
	stubHKDF := func(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
		callCount++
		if callCount == 2 {
			return nil, fmt.Errorf("simulated second HKDF error")
		}

		return cryptoutilSharedCryptoDigests.HKDFwithSHA256(secret, salt, info, outputBytesLength)
	}

	sharedSecret := make([]byte, cryptoutilSharedMagic.MinSharedSecretLength)
	sharedSecretsM := [][]byte{sharedSecret}

	unsealKeysService, err := newUnsealKeysServiceSharedSecretsInternal(sharedSecretsM, 1, stubHKDF)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to derive unseal JWK secret bytes")
}

// TestNewUnsealKeysServiceFromSettings_DevMode_GenerateBytesError tests the GenerateBytes error path.
// Uses fn-param injection to simulate random byte generation failure in DevMode.
func TestNewUnsealKeysServiceFromSettings_DevMode_GenerateBytesError(t *testing.T) {
	t.Parallel()

	stubGenerateBytes := func(_ int) ([]byte, error) {
		return nil, fmt.Errorf("simulated random generation error")
	}

	ctx := context.Background()
	settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("test-dev-generate-error")
	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings.ToTelemetrySettings())
	settings.DevMode = true

	unsealKeysService, err := newUnsealKeysServiceFromSettingsInternal(ctx, telemetryService, settings, stubGenerateBytes)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to generate random bytes for dev mode")
}
