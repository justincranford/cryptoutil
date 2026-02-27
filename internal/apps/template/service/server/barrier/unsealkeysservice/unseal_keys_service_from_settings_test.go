// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/stretchr/testify/require"
)

const testUnsealModeSysinfo = "sysinfo"

// createTestContext creates context and telemetry service for testing.
func createTestContext(t *testing.T) (context.Context, *cryptoutilSharedTelemetry.TelemetryService) {
	t.Helper()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("unsealkeysservice-test")

	return ctx, cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings.ToTelemetrySettings())
}

func TestNewUnsealKeysServiceFromSettings_DevMode(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-dev-mode")
	settings.DevMode = true

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)

	// Verify shutdown works.
	unsealKeysService.Shutdown()
}

func TestNewUnsealKeysServiceFromSettings_SysInfoMode(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-sysinfo-mode")
	settings.DevMode = false
	settings.UnsealMode = testUnsealModeSysinfo

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

func TestNewUnsealKeysServiceFromSettings_InvalidMofNFormat(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-invalid-mofn")
	settings.DevMode = false
	settings.UnsealMode = "a-of-b-of-c" // Invalid format.

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "invalid unseal mode format")
}

func TestNewUnsealKeysServiceFromSettings_InvalidMValue(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-invalid-m")
	settings.DevMode = false
	settings.UnsealMode = "abc-of-3" // Invalid M value.

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "invalid M value")
}

func TestNewUnsealKeysServiceFromSettings_InvalidNValue(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-invalid-n")
	settings.DevMode = false
	settings.UnsealMode = "2-of-xyz" // Invalid N value.

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "invalid N value")
}

// TestNewUnsealKeysServiceFromSettings_MZeroBoundary tests that m=0 is rejected in M-of-N mode.
// Standalone test (not table-driven) ensures gremlins can reliably kill the m<=0 boundary mutant.
func TestNewUnsealKeysServiceFromSettings_MZeroBoundary(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-m-zero-boundary")
	settings.DevMode = false
	settings.UnsealMode = "0-of-3" // m=0, n=3: m must fail validation, not proceed to file reading.

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "invalid M-of-N values")
}

func TestNewUnsealKeysServiceFromSettings_InvalidMNValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		unsealMode string
		errMessage string
	}{
		{"m-zero", "0-of-3", "M must be > 0"},
		{"n-zero", "2-of-0", "M must be > 0"},
		{"m-greater-than-n", "5-of-3", "M must be > 0"},
		{"both-negative", "-1-of--2", "M must be > 0"}, // Parses as negative numbers, then fails M > 0 check.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, telemetryService := createTestContext(t)
			settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-invalid-mn-" + tc.name)
			settings.DevMode = false
			settings.UnsealMode = tc.unsealMode

			unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
			require.Error(t, err)
			require.Nil(t, unsealKeysService)
			require.Contains(t, err.Error(), tc.errMessage)
		})
	}
}

func TestNewUnsealKeysServiceFromSettings_InvalidSimpleMode(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-invalid-simple")
	settings.DevMode = false
	settings.UnsealMode = "invalid-mode" // Not a number, not M-of-N, not sysinfo.

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "invalid unseal mode")
}

func TestNewUnsealKeysServiceFromSettings_ZeroSimpleMode(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-zero-simple")
	settings.DevMode = false
	settings.UnsealMode = "0" // N=0 is invalid.

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "N must be > 0")
}

func TestNewUnsealKeysServiceFromSettings_MissingFiles(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-missing-files")
	settings.DevMode = false
	settings.UnsealMode = "1"
	settings.UnsealFiles = []string{"/nonexistent/file.key"}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to read shared secrets files")
}

func TestNewUnsealKeysServiceFromSettings_FileMismatch(t *testing.T) {
	t.Parallel()

	// Create a temp dir with one file but expect 2.
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key1.jwk")

	err := os.WriteFile(keyFile, []byte(`{"kty":"oct","k":"test"}`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-file-mismatch")
	settings.DevMode = false
	settings.UnsealMode = "2" // Expecting 2 files.
	settings.UnsealFiles = []string{keyFile}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "expected 2 shared secret files, got 1")
}

func TestUnsealKeysServiceFromSettings_EncryptDecrypt(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-encrypt-decrypt")
	settings.DevMode = true

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)

	// Cast to test the interface methods.
	fromSettings, ok := unsealKeysService.(*UnsealKeysServiceSharedSecrets)
	require.True(t, ok)

	// Test EncryptKey and DecryptKey via interface.
	testData := []byte("test data")

	encrypted, err := fromSettings.EncryptData(testData)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := fromSettings.DecryptData(encrypted)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted)
}
