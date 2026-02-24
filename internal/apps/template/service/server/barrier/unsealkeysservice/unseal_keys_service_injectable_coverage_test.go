// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
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

// TestNewUnsealKeysServiceFromSysInfo_EmptySysinfos tests the empty sysinfos error path.
// Uses injectable getAllInfoWithTimeoutFn to return empty slice, covering numSysinfos == 0 branch.
//
// NOTE: Must NOT use t.Parallel() - modifies package-level getAllInfoWithTimeoutFn.
func TestNewUnsealKeysServiceFromSysInfo_EmptySysinfos(t *testing.T) {
	originalFn := getAllInfoWithTimeoutFn
	getAllInfoWithTimeoutFn = func(_ cryptoutilSharedUtilSysinfo.SysInfoProvider, _ time.Duration) ([][]byte, error) {
		return [][]byte{}, nil
	}

	defer func() { getAllInfoWithTimeoutFn = originalFn }()

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(&cryptoutilSharedUtilSysinfo.DefaultSysInfoProvider{})

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "empty sysinfos not supported")
}

// TestNewUnsealKeysServiceFromSysInfo_SingleSysinfo tests the numSysinfos == 1 branch.
// Uses injectable getAllInfoWithTimeoutFn to return a single element, covering chooseN = 1 path.
//
// NOTE: Must NOT use t.Parallel() - modifies package-level getAllInfoWithTimeoutFn.
func TestNewUnsealKeysServiceFromSysInfo_SingleSysinfo(t *testing.T) {
	originalFn := getAllInfoWithTimeoutFn
	getAllInfoWithTimeoutFn = func(_ cryptoutilSharedUtilSysinfo.SysInfoProvider, _ time.Duration) ([][]byte, error) {
		return [][]byte{[]byte("single-sysinfo-entry-for-testing")}, nil
	}

	defer func() { getAllInfoWithTimeoutFn = originalFn }()

	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(&cryptoutilSharedUtilSysinfo.DefaultSysInfoProvider{})

	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

// TestDeriveJWKs_SecondHKDFError tests the second HKDF error path in deriveJWKsFromMChooseNCombinations.
// Uses call-count injection: first HKDF call (for KID) succeeds, second (for secret) fails.
//
// NOTE: Must NOT use t.Parallel() - modifies package-level hkdfWithSHA256Fn.
func TestDeriveJWKs_SecondHKDFError(t *testing.T) {
	callCount := 0
	originalFn := hkdfWithSHA256Fn

	hkdfWithSHA256Fn = func(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
		callCount++
		if callCount == 2 {
			return nil, fmt.Errorf("simulated second HKDF error")
		}

		return cryptoutilSharedCryptoDigests.HKDFwithSHA256(secret, salt, info, outputBytesLength)
	}

	defer func() { hkdfWithSHA256Fn = originalFn }()

	sharedSecret := make([]byte, cryptoutilSharedMagic.MinSharedSecretLength)
	sharedSecretsM := [][]byte{sharedSecret}

	unsealKeysService, err := NewUnsealKeysServiceSharedSecrets(sharedSecretsM, 1)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to derive unseal JWK secret bytes")
}

// TestNewUnsealKeysServiceFromSettings_DevMode_GenerateBytesError tests the GenerateBytes error path.
// Uses injectable generateBytesFn to simulate random byte generation failure in DevMode.
//
// NOTE: Must NOT use t.Parallel() - modifies package-level generateBytesFn.
func TestNewUnsealKeysServiceFromSettings_DevMode_GenerateBytesError(t *testing.T) {
	originalFn := generateBytesFn
	generateBytesFn = func(_ int) ([]byte, error) {
		return nil, fmt.Errorf("simulated random generation error")
	}

	defer func() { generateBytesFn = originalFn }()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-dev-generate-error")
	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings.ToTelemetrySettings())
	settings.DevMode = true

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)

	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to generate random bytes for dev mode")
}
