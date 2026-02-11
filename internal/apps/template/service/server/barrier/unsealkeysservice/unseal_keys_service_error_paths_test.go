// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestDeriveJWKsFromMChooseNCombinations_EmptyCombinations tests error when combinations are empty.
func TestDeriveJWKsFromMChooseNCombinations_EmptyCombinations(t *testing.T) {
	t.Parallel()

	// This should trigger "no combinations" error (0 choose anything = 0 combinations)
	emptySecrets := [][]byte{}
	chooseN := 0

	jwks, err := deriveJWKsFromMChooseNCombinations(emptySecrets, chooseN)
	require.Error(t, err)
	require.Nil(t, jwks)
	// Empty combinations result triggers the "no combinations" error path
	require.Contains(t, err.Error(), "no combinations")
}

// TestDeriveJWKsFromMChooseNCombinations_InvalidChooseN tests invalid chooseN values.
func TestDeriveJWKsFromMChooseNCombinations_InvalidChooseN(t *testing.T) {
	t.Parallel()

	secrets := [][]byte{
		[]byte("first secret with sufficient length for testing"),
		[]byte("second secret with sufficient length for testing"),
	}

	tests := []struct {
		name     string
		chooseN  int
		errMatch string
	}{
		{"negative-chooseN", -1, "failed to compute"},
		{"chooseN-exceeds-m", 5, "failed to compute"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwks, err := deriveJWKsFromMChooseNCombinations(secrets, tc.chooseN)
			require.Error(t, err)
			require.Nil(t, jwks)
			require.Contains(t, err.Error(), tc.errMatch)
		})
	}
}

// TestEncryptKey_DecryptKey_ErrorPaths tests error paths in encrypt/decrypt key functions.
func TestEncryptKey_DecryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	// Test decryptKey with invalid encrypted data
	// Create valid unseal JWKs first
	secrets := [][]byte{
		[]byte("shared secret one with sufficient length"),
		[]byte("shared secret two with sufficient length"),
	}

	jwks, err := deriveJWKsFromMChooseNCombinations(secrets, 2)
	require.NoError(t, err)
	require.NotEmpty(t, jwks)

	// Try to decrypt invalid data
	invalidEncryptedData := []byte("not valid encrypted key data")
	_, err = decryptKey(jwks, invalidEncryptedData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt root JWK")
}

// TestEncryptData_DecryptData_ErrorPaths tests error paths in encrypt/decrypt data functions.
func TestEncryptData_DecryptData_ErrorPaths(t *testing.T) {
	t.Parallel()

	// Create valid unseal JWKs
	secrets := [][]byte{
		[]byte("shared secret alpha with sufficient length"),
		[]byte("shared secret beta with sufficient length"),
	}

	jwks, err := deriveJWKsFromMChooseNCombinations(secrets, 2)
	require.NoError(t, err)
	require.NotEmpty(t, jwks)

	// Try to decrypt invalid data
	invalidEncryptedData := []byte("not valid encrypted data bytes")
	_, err = decryptData(jwks, invalidEncryptedData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt data")
}

// TestDeriveJWKsFromMChooseNCombinations_HappyPath tests successful JWK derivation.
func TestDeriveJWKsFromMChooseNCombinations_HappyPath(t *testing.T) {
	t.Parallel()

	secrets := [][]byte{
		[]byte("first shared secret with sufficient length for test"),
		[]byte("second shared secret with sufficient length for test"),
		[]byte("third shared secret with sufficient length for test"),
	}

	// Test 2-of-3 combinations (should generate 3 JWKs: AB, AC, BC)
	jwks, err := deriveJWKsFromMChooseNCombinations(secrets, 2)
	require.NoError(t, err)
	require.Len(t, jwks, 3) // C(3,2) = 3 combinations

	// Verify all JWKs are valid
	for i, jwk := range jwks {
		require.NotNil(t, jwk, "JWK %d should not be nil", i)
	}
}

// TestDeriveJWKsFromMChooseNCombinations_Deterministic tests deterministic JWK derivation.
func TestDeriveJWKsFromMChooseNCombinations_Deterministic(t *testing.T) {
	t.Parallel()

	secrets := [][]byte{
		[]byte("deterministic secret one with sufficient length"),
		[]byte("deterministic secret two with sufficient length"),
	}

	// Derive JWKs twice with same inputs
	jwks1, err := deriveJWKsFromMChooseNCombinations(secrets, 2)
	require.NoError(t, err)
	require.Len(t, jwks1, 1) // C(2,2) = 1 combination

	jwks2, err := deriveJWKsFromMChooseNCombinations(secrets, 2)
	require.NoError(t, err)
	require.Len(t, jwks2, 1)

	// Verify KIDs match (deterministic derivation)
	kid1, _ := jwks1[0].KeyID()
	kid2, _ := jwks2[0].KeyID()
	require.Equal(t, kid1, kid2, "Derived KIDs should match for same inputs")
}

// TestNewUnsealKeysServiceSharedSecrets_MaxSecrets tests max shared secrets boundary.
func TestNewUnsealKeysServiceSharedSecrets_MaxSecrets(t *testing.T) {
	t.Parallel()

	// Create exactly MaxUnsealSharedSecrets (256) secrets
	secrets := make([][]byte, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	for i := 0; i < cryptoutilSharedMagic.MaxUnsealSharedSecrets; i++ {
		secrets[i] = []byte("shared secret with sufficient minimum length here")
	}

	// This should fail because count >= MaxUnsealSharedSecrets
	service, err := NewUnsealKeysServiceSharedSecrets(secrets, 10)
	require.Error(t, err)
	require.Nil(t, service)
	require.Contains(t, err.Error(), "can't be greater than")
}
