// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	"context"
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/stretchr/testify/require"
)

// TestNewUnsealKeysServiceFromSettings_VerboseMode tests verbose logging path.
func TestNewUnsealKeysServiceFromSettings_VerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-verbose-mode")
	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	telemetryService.VerboseMode = true

	// Use DevMode=true to avoid sysinfo collection timeout on Windows (CPU info takes 4+ seconds).
	// This still tests the verbose logging path in NewUnsealKeysServiceFromSettings.
	settings.DevMode = true

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

// TestNewUnsealKeysServiceFromSettings_SimpleMode_NegativeN tests negative N value in simple mode.
func TestNewUnsealKeysServiceFromSettings_SimpleMode_NegativeN(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-negative-n")
	settings.DevMode = false
	settings.UnsealMode = "-5" // Negative N

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "N must be > 0")
}

// TestNewUnsealKeysServiceFromSettings_SimpleMode_FileCountMismatch tests N mode with wrong file count.
func TestNewUnsealKeysServiceFromSettings_SimpleMode_FileCountMismatch(t *testing.T) {
	t.Parallel()

	// Create 1 JWK file but expect 3
	jwks, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 1, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	jwkFile := filepath.Join(tmpDir, "jwk1.json")

	jwkBytes, err := json.Marshal(jwks[0])
	require.NoError(t, err)

	err = os.WriteFile(jwkFile, jwkBytes, 0o600)
	require.NoError(t, err)

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-simple-file-mismatch")
	settings.DevMode = false
	settings.UnsealMode = "3" // Expecting 3 files
	settings.UnsealFiles = []string{jwkFile}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "expected 3 shared secret files, got 1")
}

// TestUnsealKeysServiceFromSysInfo_EncryptDecryptKey tests sysinfo service key operations.
func TestUnsealKeysServiceFromSysInfo_EncryptDecryptKey(t *testing.T) {
	t.Parallel()

	// Use existing helper to create service from real sysinfo
	service := RequireNewFromSysInfoForTest()
	require.NotNil(t, service)

	// Generate a test key
	testKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 1, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	clearKey := testKeys[0]

	// Encrypt the key
	encryptedKey, err := service.EncryptKey(clearKey)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	// Decrypt the key
	decryptedKey, err := service.DecryptKey(encryptedKey)
	require.NoError(t, err)
	require.NotNil(t, decryptedKey)
}
