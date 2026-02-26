// Copyright (c) 2025 Justin Cranford

package hash

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHashLowEntropyNonDeterministic tests the wrapper function for PBKDF2-based password hashing.
func TestHashLowEntropyNonDeterministic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret string
	}{
		{
			name:   "simple_password",
			secret: "password123",
		},
		{
			name:   "short_password",
			secret: "abc",
		},
		{
			name:   "long_password",
			secret: strings.Repeat("a", cryptoutilSharedMagic.MaxUnsealSharedSecrets),
		},
		{
			name:   "unicode_password",
			secret: "пароль密码🔐",
		},
		{
			name:   "special_chars",
			secret: "p@$$w0rd!#%&",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Hash the secret
			hash, err := HashLowEntropyNonDeterministic(tt.secret)
			require.NoError(t, err, "HashLowEntropyNonDeterministic should not error")
			require.NotEmpty(t, hash, "hash should not be empty")

			// Verify format: {1}$pbkdf2-sha256$iter$salt$dk
			require.True(t, strings.HasPrefix(hash, "{1}$pbkdf2-sha256$"), "hash should have correct prefix")

			parts := strings.Split(hash, "$")
			require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, len(parts), "hash should have 5 parts")

			// Verify non-determinism: same secret produces different hashes
			hash2, err := HashLowEntropyNonDeterministic(tt.secret)
			require.NoError(t, err, "second hash should not error")
			require.NotEqual(t, hash, hash2, "two hashes of same secret should differ (random salt)")
		})
	}
}

// TestHashSecretPBKDF2 tests the direct PBKDF2 hash function.
func TestHashSecretPBKDF2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret string
	}{
		{
			name:   "standard_password",
			secret: "SecurePassword123!",
		},
		{
			name:   "two_char_string",
			secret: "ab",
		},
		{
			name:   "single_char",
			secret: "x",
		},
		{
			name:   "max_length",
			secret: strings.Repeat("long", cryptoutilSharedMagic.MinSerialNumberBits), // 256 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashSecretPBKDF2(tt.secret)
			require.NoError(t, err, "HashSecretPBKDF2 should not error")
			require.NotEmpty(t, hash, "hash should not be empty")

			// Verify format
			require.True(t, strings.HasPrefix(hash, "{1}$pbkdf2-sha256$"), "hash should use version 1 format")

			// Verify hash components
			parts := strings.Split(hash, "$")
			require.Len(t, parts, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, "hash should have 5 components")
			require.Equal(t, "{1}", parts[0], "version should be {1}")
			require.Equal(t, cryptoutilSharedMagic.PBKDF2Prefix, parts[1], "algorithm should be pbkdf2-sha256")
			require.Equal(t, "600000", parts[2], "iterations should be 600000")

			// Salt and derived key should be base64-encoded
			require.NotEmpty(t, parts[3], "salt should not be empty")
			require.NotEmpty(t, parts[4], "derived key should not be empty")
		})
	}
}

// TestPBKDF2ParameterSetVariants tests SHA-384 and SHA-512 parameter set functions.
func TestPBKDF2ParameterSetVariants(t *testing.T) {
	t.Parallel()

	t.Run("PBKDF2SHA384ParameterSetV1", func(t *testing.T) {
		t.Parallel()

		params := PBKDF2SHA384ParameterSetV1()
		require.NotNil(t, params)
		require.Equal(t, "1", params.Version)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2SHA384HashName, params.HashName)
		require.Equal(t, cryptoutilSharedMagic.IMPBKDF2Iterations, params.Iterations)
		require.Equal(t, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, params.SaltLength)
		require.Equal(t, cryptoutilSharedMagic.HMACSHA384KeySize, params.KeyLength) // SHA-384 = 48 bytes
		require.NotNil(t, params.HashFunc)
	})

	t.Run("PBKDF2SHA384ParameterSetV2", func(t *testing.T) {
		t.Parallel()

		params := PBKDF2SHA384ParameterSetV2()
		require.NotNil(t, params)
		require.Equal(t, "2", params.Version)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2SHA384HashName, params.HashName)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2V2Iterations, params.Iterations)
		require.Equal(t, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, params.SaltLength)
		require.Equal(t, cryptoutilSharedMagic.HMACSHA384KeySize, params.KeyLength)
		require.NotNil(t, params.HashFunc)
	})

	t.Run("PBKDF2SHA384ParameterSetV3", func(t *testing.T) {
		t.Parallel()

		params := PBKDF2SHA384ParameterSetV3()
		require.NotNil(t, params)
		require.Equal(t, "3", params.Version)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2SHA384HashName, params.HashName)
		require.Equal(t, cryptoutilSharedMagic.JoseJADefaultListLimit, params.Iterations)
		require.Equal(t, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, params.SaltLength)
		require.Equal(t, cryptoutilSharedMagic.HMACSHA384KeySize, params.KeyLength)
		require.NotNil(t, params.HashFunc)
	})

	t.Run("PBKDF2SHA512ParameterSetV1", func(t *testing.T) {
		t.Parallel()

		params := PBKDF2SHA512ParameterSetV1()
		require.NotNil(t, params)
		require.Equal(t, "1", params.Version)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2SHA512HashName, params.HashName)
		require.Equal(t, cryptoutilSharedMagic.IMPBKDF2Iterations, params.Iterations)
		require.Equal(t, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, params.SaltLength)
		require.Equal(t, cryptoutilSharedMagic.MinSerialNumberBits, params.KeyLength) // SHA-512 = 64 bytes
		require.NotNil(t, params.HashFunc)
	})

	t.Run("PBKDF2SHA512ParameterSetV2", func(t *testing.T) {
		t.Parallel()

		params := PBKDF2SHA512ParameterSetV2()
		require.NotNil(t, params)
		require.Equal(t, "2", params.Version)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2SHA512HashName, params.HashName)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2V2Iterations, params.Iterations)
		require.Equal(t, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, params.SaltLength)
		require.Equal(t, cryptoutilSharedMagic.MinSerialNumberBits, params.KeyLength)
		require.NotNil(t, params.HashFunc)
	})

	t.Run("PBKDF2SHA512ParameterSetV3", func(t *testing.T) {
		t.Parallel()

		params := PBKDF2SHA512ParameterSetV3()
		require.NotNil(t, params)
		require.Equal(t, "3", params.Version)
		require.Equal(t, cryptoutilSharedMagic.PBKDF2SHA512HashName, params.HashName)
		require.Equal(t, cryptoutilSharedMagic.JoseJADefaultListLimit, params.Iterations)
		require.Equal(t, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, params.SaltLength)
		require.Equal(t, cryptoutilSharedMagic.MinSerialNumberBits, params.KeyLength)
		require.NotNil(t, params.HashFunc)
	})
}

// TestGetDefaultParameterSet tests the registry default parameter set lookup.
func TestGetDefaultParameterSet(t *testing.T) {
	t.Parallel()

	registry := GetGlobalRegistry()

	// Test getting default parameter set
	params := registry.GetDefaultParameterSet()
	require.NotNil(t, params)
	require.Equal(t, "1", params.Version)
	require.Equal(t, cryptoutilSharedMagic.IMPBKDF2Iterations, params.Iterations)

	// Test getting default version string
	defaultVersion := registry.GetDefaultVersion()
	require.Equal(t, "1", defaultVersion)

	// Test getting specific version "1"
	params, err := registry.GetParameterSet("1")
	require.NoError(t, err)
	require.NotNil(t, params)
	require.Equal(t, "1", params.Version)

	// Test getting invalid version (should error)
	_, err = registry.GetParameterSet("99")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}
