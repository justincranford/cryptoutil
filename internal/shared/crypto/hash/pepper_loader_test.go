// Copyright (c) 2025 Justin Cranford

package hash

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testPasswordValue = "testPassword123"

// TestLoadPepperFromSecret_HappyPath tests successful pepper loading.
func TestLoadPepperFromSecret_HappyPath(t *testing.T) {
	t.Parallel()

	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "test_pepper.secret")

	expectedPepper := "7t1qT7/OxY7lzqe8E5Q89AfNF2iNzu+QrvLJJe+V/WY="
	require.NoError(t, os.WriteFile(secretFile, []byte(expectedPepper+"\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// Test loading pepper
	pepper, err := LoadPepperFromSecret(secretFile)
	require.NoError(t, err)
	require.Equal(t, expectedPepper, pepper)
}

// TestLoadPepperFromSecret_FilePrefixSupport tests "file://" prefix handling.
func TestLoadPepperFromSecret_FilePrefixSupport(t *testing.T) {
	t.Parallel()

	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "test_pepper.secret")

	expectedPepper := "testPepperValue123"
	require.NoError(t, os.WriteFile(secretFile, []byte(expectedPepper), cryptoutilSharedMagic.CacheFilePermissions))

	// Test with "file://" prefix
	pepper, err := LoadPepperFromSecret(cryptoutilSharedMagic.FileURIScheme + secretFile)
	require.NoError(t, err)
	require.Equal(t, expectedPepper, pepper)
}

// TestLoadPepperFromSecret_WhitespaceTrimming tests whitespace trimming.
func TestLoadPepperFromSecret_WhitespaceTrimming(t *testing.T) {
	t.Parallel()

	// Create temporary secret file with whitespace
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "test_pepper.secret")

	expectedPepper := "pepperWithWhitespace"
	require.NoError(t, os.WriteFile(secretFile, []byte("\n  "+expectedPepper+"  \n\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// Test loading and trimming
	pepper, err := LoadPepperFromSecret(secretFile)
	require.NoError(t, err)
	require.Equal(t, expectedPepper, pepper)
}

// TestLoadPepperFromSecret_ErrorCases tests error handling.
func TestLoadPepperFromSecret_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) string // Returns secretPath
		expectedErrMsg string
	}{
		{
			name: "empty secret path",
			setupFunc: func(_ *testing.T) string {
				return ""
			},
			expectedErrMsg: "secret path is empty",
		},
		{
			name: "non-existent file",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.secret")
			},
			expectedErrMsg: "failed to read pepper secret",
		},
		{
			name: "empty file",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				emptyFile := filepath.Join(tempDir, "empty.secret")
				require.NoError(t, os.WriteFile(emptyFile, []byte("  \n  "), cryptoutilSharedMagic.CacheFilePermissions))

				return emptyFile
			},
			expectedErrMsg: "pepper secret file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			secretPath := tt.setupFunc(t)

			pepper, err := LoadPepperFromSecret(secretPath)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrMsg)
			require.Empty(t, pepper)
		})
	}
}

// TestConfigurePeppers_HappyPath tests successful pepper configuration in registry.
func TestConfigurePeppers_HappyPath(t *testing.T) {
	t.Parallel()

	// Create temporary secret files for v1, v2, v3
	tempDir := t.TempDir()

	pepper1 := "pepper_v1_" + googleUuid.NewString()
	pepper2 := "pepper_v2_" + googleUuid.NewString()
	pepper3 := "pepper_v3_" + googleUuid.NewString()

	secret1 := filepath.Join(tempDir, "pepper_v1.secret")
	secret2 := filepath.Join(tempDir, "pepper_v2.secret")
	secret3 := filepath.Join(tempDir, "pepper_v3.secret")

	require.NoError(t, os.WriteFile(secret1, []byte(pepper1), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(secret2, []byte(pepper2), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(secret3, []byte(pepper3), cryptoutilSharedMagic.CacheFilePermissions))

	// Create fresh registry
	registry := NewParameterSetRegistry()

	// Configure peppers
	peppers := []PepperConfig{
		{Version: "1", SecretPath: secret1},
		{Version: "2", SecretPath: secret2},
		{Version: "3", SecretPath: cryptoutilSharedMagic.FileURIScheme + secret3}, // Test file:// prefix
	}

	err := ConfigurePeppers(registry, peppers)
	require.NoError(t, err)

	// Verify peppers loaded into parameter sets
	params1, err := registry.GetParameterSet("1")
	require.NoError(t, err)
	require.Equal(t, pepper1, params1.Pepper)

	params2, err := registry.GetParameterSet("2")
	require.NoError(t, err)
	require.Equal(t, pepper2, params2.Pepper)

	params3, err := registry.GetParameterSet("3")
	require.NoError(t, err)
	require.Equal(t, pepper3, params3.Pepper)
}

// TestConfigurePeppers_ErrorCases tests error handling in pepper configuration.
func TestConfigurePeppers_ErrorCases(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validSecret := filepath.Join(tempDir, "valid.secret")
	require.NoError(t, os.WriteFile(validSecret, []byte("validPepper"), cryptoutilSharedMagic.CacheFilePermissions))

	tests := []struct {
		name           string
		registry       *ParameterSetRegistry
		peppers        []PepperConfig
		expectedErrMsg string
	}{
		{
			name:           "nil registry",
			registry:       nil,
			peppers:        []PepperConfig{{Version: "1", SecretPath: validSecret}},
			expectedErrMsg: "registry is nil",
		},
		{
			name:     "empty version",
			registry: NewParameterSetRegistry(),
			peppers: []PepperConfig{
				{Version: "", SecretPath: validSecret},
			},
			expectedErrMsg: "empty version",
		},
		{
			name:     "empty secret path",
			registry: NewParameterSetRegistry(),
			peppers: []PepperConfig{
				{Version: "1", SecretPath: ""},
			},
			expectedErrMsg: "empty secret path",
		},
		{
			name:     "non-existent secret file",
			registry: NewParameterSetRegistry(),
			peppers: []PepperConfig{
				{Version: "1", SecretPath: filepath.Join(tempDir, "nonexistent.secret")},
			},
			expectedErrMsg: "failed to load pepper",
		},
		{
			name:     "invalid version (not in registry)",
			registry: NewParameterSetRegistry(),
			peppers: []PepperConfig{
				{Version: "999", SecretPath: validSecret},
			},
			expectedErrMsg: "parameter set version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ConfigurePeppers(tt.registry, tt.peppers)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

// TestPepperedHashing_DifferentPeppersProduceDifferentHashes tests OWASP requirement.
// CRITICAL: Different peppers MUST produce different hashes for same password.
func TestPepperedHashing_DifferentPeppersProduceDifferentHashes(t *testing.T) {
	t.Parallel()

	// Create temporary secret files with different peppers
	tempDir := t.TempDir()

	pepper1 := "pepper_value_1"
	pepper2 := "pepper_value_2"

	secret1 := filepath.Join(tempDir, "pepper1.secret")
	secret2 := filepath.Join(tempDir, "pepper2.secret")

	require.NoError(t, os.WriteFile(secret1, []byte(pepper1), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(secret2, []byte(pepper2), cryptoutilSharedMagic.CacheFilePermissions))

	// Create two registries with different peppers
	registry1 := NewParameterSetRegistry()
	registry2 := NewParameterSetRegistry()

	require.NoError(t, ConfigurePeppers(registry1, []PepperConfig{
		{Version: "1", SecretPath: secret1},
	}))

	require.NoError(t, ConfigurePeppers(registry2, []PepperConfig{
		{Version: "1", SecretPath: secret2},
	}))

	// Hash same password with different peppers
	password := testPasswordValue

	params1 := registry1.GetDefaultParameterSet()
	params2 := registry2.GetDefaultParameterSet()

	hash1, err := HashSecretPBKDF2WithParams(password, params1)
	require.NoError(t, err)

	hash2, err := HashSecretPBKDF2WithParams(password, params2)
	require.NoError(t, err)

	// CRITICAL: Different peppers MUST produce different hashes
	require.NotEqual(t, hash1, hash2, "Different peppers MUST produce different hashes for same password")
}

// TestPepperedVerification_CorrectPepperRequired tests verification requires matching pepper.
func TestPepperedVerification_CorrectPepperRequired(t *testing.T) {
	t.Parallel()

	// Create temporary secret files
	tempDir := t.TempDir()

	correctPepper := "correct_pepper_value"
	wrongPepper := "wrong_pepper_value"

	correctSecret := filepath.Join(tempDir, "correct.secret")
	wrongSecret := filepath.Join(tempDir, "wrong.secret")

	require.NoError(t, os.WriteFile(correctSecret, []byte(correctPepper), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(wrongSecret, []byte(wrongPepper), cryptoutilSharedMagic.CacheFilePermissions))

	// Create registry with correct pepper
	registryCorrect := NewParameterSetRegistry()
	require.NoError(t, ConfigurePeppers(registryCorrect, []PepperConfig{
		{Version: "1", SecretPath: correctSecret},
	}))

	// Hash password with correct pepper
	password := testPasswordValue
	paramsCorrect := registryCorrect.GetDefaultParameterSet()

	hash, err := HashSecretPBKDF2WithParams(password, paramsCorrect)
	require.NoError(t, err)

	// Verify with CORRECT pepper (should succeed)
	valid, err := VerifySecretPBKDF2WithParams(hash, password, paramsCorrect)
	require.NoError(t, err)
	require.True(t, valid, "Verification with correct pepper MUST succeed")

	// Create registry with WRONG pepper
	registryWrong := NewParameterSetRegistry()
	require.NoError(t, ConfigurePeppers(registryWrong, []PepperConfig{
		{Version: "1", SecretPath: wrongSecret},
	}))

	paramsWrong := registryWrong.GetDefaultParameterSet()

	// Verify with WRONG pepper (should fail)
	valid, err = VerifySecretPBKDF2WithParams(hash, password, paramsWrong)
	require.NoError(t, err)
	require.False(t, valid, "Verification with wrong pepper MUST fail")
}
