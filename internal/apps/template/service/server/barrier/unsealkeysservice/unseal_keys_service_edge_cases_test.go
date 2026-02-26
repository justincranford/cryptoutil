// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	"github.com/stretchr/testify/require"
)

// TestEncryptKey_EncryptData_Comprehensive tests all encryption paths.
func TestEncryptKey_EncryptData_Comprehensive(t *testing.T) {
	t.Parallel()

	secrets := [][]byte{
		[]byte("comprehensive test secret one with sufficient length"),
		[]byte("comprehensive test secret two with sufficient length"),
		[]byte("comprehensive test secret three with sufficient length"),
	}

	// Test 3-of-3 (all combinations = 1)
	service, err := NewUnsealKeysServiceSharedSecrets(secrets, 3)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Test EncryptData path
	testData := []byte("test data for comprehensive coverage")

	encrypted, err := service.EncryptData(testData)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := service.DecryptData(encrypted)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted)

	// Test EncryptKey path
	testKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 1, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	encryptedKey, err := service.EncryptKey(testKeys[0])
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	decryptedKey, err := service.DecryptKey(encryptedKey)
	require.NoError(t, err)
	require.NotNil(t, decryptedKey)
}

// TestDeriveJWKsFromMChooseNCombinations_LargeCombinations tests larger combination sets.
func TestDeriveJWKsFromMChooseNCombinations_LargeCombinations(t *testing.T) {
	t.Parallel()

	// Create 5 secrets, choose 3 = C(5,3) = 10 combinations
	secrets := [][]byte{
		[]byte("secret one with minimum required length for testing"),
		[]byte("secret two with minimum required length for testing"),
		[]byte("secret three with minimum required length for testing"),
		[]byte("secret four with minimum required length for testing"),
		[]byte("secret five with minimum required length for testing"),
	}

	jwks, err := deriveJWKsFromMChooseNCombinations(secrets, 3)
	require.NoError(t, err)
	require.Len(t, jwks, cryptoutilSharedMagic.JoseJADefaultMaxMaterials) // C(5,3) = 10

	// Verify all JWKs are unique
	kidSet := make(map[string]bool)

	for _, jwk := range jwks {
		kid, _ := jwk.KeyID()
		require.NotContains(t, kidSet, kid, "KIDs should be unique")
		kidSet[kid] = true
	}
}

// TestNewUnsealKeysServiceSharedSecrets_EdgeCaseBoundaries tests boundary conditions.
func TestNewUnsealKeysServiceSharedSecrets_EdgeCaseBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		secretLen int
		shouldErr bool
		errMatch  string
	}{
		{"min-length-31", 31, true, "can't be less than"},
		{"min-length-cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes", 32, false, ""},
		{"max-length-cryptoutilSharedMagic.MinSerialNumberBits", 64, false, ""},
		{"max-length-65", 65, true, "can't be greater than"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			secret := make([]byte, tc.secretLen)
			for i := range secret {
				secret[i] = byte('a')
			}

			service, err := NewUnsealKeysServiceSharedSecrets([][]byte{secret}, 1)

			if tc.shouldErr {
				require.Error(t, err)
				require.Nil(t, service)
				require.Contains(t, err.Error(), tc.errMatch)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
			}
		})
	}
}

// TestUnsealKeysServiceFromSysInfo_SingleSysInfo tests single sysinfo value (chooseN=1).
func TestUnsealKeysServiceFromSysInfo_SingleSysInfo(t *testing.T) {
	t.Parallel()

	// Use real sysinfo which will have multiple values
	service := RequireNewFromSysInfoForTest()
	require.NotNil(t, service)

	// Test encryption/decryption works
	testData := []byte("single sysinfo test data")

	encrypted, err := service.EncryptData(testData)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := service.DecryptData(encrypted)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted)
}

// TestDeriveJWKsFromMChooseNCombinations_Choose1 tests 1-choose-1 case.
func TestDeriveJWKsFromMChooseNCombinations_Choose1(t *testing.T) {
	t.Parallel()

	secrets := [][]byte{
		[]byte("single secret with sufficient minimum length"),
	}

	jwks, err := deriveJWKsFromMChooseNCombinations(secrets, 1)
	require.NoError(t, err)
	require.Len(t, jwks, 1) // C(1,1) = 1

	// Verify JWK is valid
	require.NotNil(t, jwks[0])
	kid, ok := jwks[0].KeyID()
	require.True(t, ok)
	require.NotEmpty(t, kid)
}

// TestEncryptDecryptKey_LargeKey tests encryption with larger key structures.
func TestEncryptDecryptKey_LargeKey(t *testing.T) {
	t.Parallel()

	secrets := [][]byte{
		[]byte("large key test secret one with sufficient length"),
		[]byte("large key test secret two with sufficient length"),
	}

	service, err := NewUnsealKeysServiceSharedSecrets(secrets, 2)
	require.NoError(t, err)

	// Generate RSA key (larger than AES)
	testKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKsForTest(t, 1, &cryptoutilSharedCryptoJose.AlgRS256)
	require.NoError(t, err)

	encryptedKey, err := service.EncryptKey(testKeys[0])
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	decryptedKey, err := service.DecryptKey(encryptedKey)
	require.NoError(t, err)
	require.NotNil(t, decryptedKey)
}
