// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	"github.com/stretchr/testify/require"
)

const testUnsealMode2Of3 = "2-of-3"

// TestNewUnsealKeysServiceFromSettings_MofN_FileMismatch tests M-of-N mode with wrong file count.
func TestNewUnsealKeysServiceFromSettings_MofN_FileMismatch(t *testing.T) {
	t.Parallel()

	// Create temp dir with 1 file but expect 3 in 2-of-3 mode
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret1.bin")

	err := os.WriteFile(secretFile, []byte("this is a long enough shared secret for testing purposes"), 0o600)
	require.NoError(t, err)

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-mofn-file-mismatch")
	settings.DevMode = false
	settings.UnsealMode = testUnsealMode2Of3 // Expecting 3 files
	settings.UnsealFiles = []string{secretFile}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "expected 3 shared secret files, got 1")
}

// TestNewUnsealKeysServiceFromSettings_MofN_HappyPath tests successful M-of-N mode.
func TestNewUnsealKeysServiceFromSettings_MofN_HappyPath(t *testing.T) {
	t.Parallel()

	// Create temp dir with 3 files for 2-of-3 mode
	tmpDir := t.TempDir()
	secret1 := filepath.Join(tmpDir, "secret1.bin")
	secret2 := filepath.Join(tmpDir, "secret2.bin")
	secret3 := filepath.Join(tmpDir, "secret3.bin")

	err := os.WriteFile(secret1, []byte("first shared secret with sufficient length for validation"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(secret2, []byte("second shared secret with sufficient length for validation"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(secret3, []byte("third shared secret with sufficient length for validation"), 0o600)
	require.NoError(t, err)

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-mofn-happy")
	settings.DevMode = false
	settings.UnsealMode = testUnsealMode2Of3
	settings.UnsealFiles = []string{secret1, secret2, secret3}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

// TestNewUnsealKeysServiceFromSettings_SimpleMode_HappyPath tests N mode with JWK files.
func TestNewUnsealKeysServiceFromSettings_SimpleMode_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate 2 JWKs for simple mode
	jwks, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 2, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.Len(t, jwks, 2)

	// Write JWKs to temp files
	tmpDir := t.TempDir()
	jwk1File := filepath.Join(tmpDir, "jwk1.json")
	jwk2File := filepath.Join(tmpDir, "jwk2.json")

	jwk1Bytes, err := json.Marshal(jwks[0])
	require.NoError(t, err)

	jwk2Bytes, err := json.Marshal(jwks[1])
	require.NoError(t, err)

	err = os.WriteFile(jwk1File, jwk1Bytes, 0o600)
	require.NoError(t, err)

	err = os.WriteFile(jwk2File, jwk2Bytes, 0o600)
	require.NoError(t, err)

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-simple-mode-happy")
	settings.DevMode = false
	settings.UnsealMode = "2" // Simple mode with 2 JWK files
	settings.UnsealFiles = []string{jwk1File, jwk2File}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

// TestNewUnsealKeysServiceFromSettings_SimpleMode_InvalidJWK tests N mode with invalid JWK content.
func TestNewUnsealKeysServiceFromSettings_SimpleMode_InvalidJWK(t *testing.T) {
	t.Parallel()

	// Create temp dir with invalid JWK content
	tmpDir := t.TempDir()
	badJWKFile := filepath.Join(tmpDir, "bad.jwk")

	err := os.WriteFile(badJWKFile, []byte("this is not a valid JWK"), 0o600)
	require.NoError(t, err)

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-simple-mode-invalid-jwk")
	settings.DevMode = false
	settings.UnsealMode = "1" // Simple mode with 1 JWK file
	settings.UnsealFiles = []string{badJWKFile}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to parse JWK from file contents")
}

// TestNewUnsealKeysServiceFromSettings_MofN_ReadFilesError tests M-of-N with file read error.
func TestNewUnsealKeysServiceFromSettings_MofN_ReadFilesError(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-mofn-read-error")
	settings.DevMode = false
	settings.UnsealMode = testUnsealMode2Of3
	settings.UnsealFiles = []string{"/nonexistent/path1.bin", "/nonexistent/path2.bin", "/nonexistent/path3.bin"}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to read shared secrets files")
}

// TestNewUnsealKeysServiceFromSettings_SimpleMode_ReadFilesError tests N mode with file read error.
func TestNewUnsealKeysServiceFromSettings_SimpleMode_ReadFilesError(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-simple-read-error")
	settings.DevMode = false
	settings.UnsealMode = "2"
	settings.UnsealFiles = []string{"/nonexistent/jwk1.json", "/nonexistent/jwk2.json"}

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.Contains(t, err.Error(), "failed to read shared secrets files")
}

// TestUnsealKeysServiceFromSettings_InterfaceMethods tests all interface methods via FromSettings.
func TestUnsealKeysServiceFromSettings_InterfaceMethods(t *testing.T) {
	t.Parallel()

	ctx, telemetryService := createTestContext(t)
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test-interface-methods")
	settings.DevMode = true

	unsealKeysService, err := NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)

	// Test EncryptData and DecryptData
	testData := []byte("sensitive data for testing")

	encrypted, err := unsealKeysService.EncryptData(testData)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
	require.NotEqual(t, testData, encrypted)

	decrypted, err := unsealKeysService.DecryptData(encrypted)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted)

	// Test EncryptKey and DecryptKey
	jwks, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 1, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.Len(t, jwks, 1)

	encryptedKey, err := unsealKeysService.EncryptKey(jwks[0])
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	decryptedKey, err := unsealKeysService.DecryptKey(encryptedKey)
	require.NoError(t, err)
	require.NotNil(t, decryptedKey)

	// Test Shutdown
	require.NotPanics(t, func() {
		unsealKeysService.Shutdown()
	})
}
