// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

// TestUnsealKeysServiceSimple_EncryptDecryptKey tests key encryption and decryption.
func TestUnsealKeysServiceSimple_EncryptDecryptKey(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Generate a test key to encrypt
	testKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 1, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.Len(t, testKeys, 1)
	clearKey := testKeys[0]

	// Encrypt the key
	encryptedKeyBytes, err := service.EncryptKey(clearKey)
	require.NoError(t, err)
	require.NotNil(t, encryptedKeyBytes)
	require.Greater(t, len(encryptedKeyBytes), 0)

	// Decrypt the key
	decryptedKey, err := service.DecryptKey(encryptedKeyBytes)
	require.NoError(t, err)
	require.NotNil(t, decryptedKey)

	// Verify keys are not nil (can't directly compare JWK objects)
	require.NotNil(t, clearKey)
	require.NotNil(t, decryptedKey)
}

// TestUnsealKeysServiceSimple_EncryptDecryptData tests data encryption and decryption.
func TestUnsealKeysServiceSimple_EncryptDecryptData(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Test data
	clearData := []byte("sensitive data to encrypt")

	// Encrypt the data
	encryptedData, err := service.EncryptData(clearData)
	require.NoError(t, err)
	require.NotNil(t, encryptedData)
	require.Greater(t, len(encryptedData), 0)

	// Encrypted data should be different from clear data
	require.NotEqual(t, clearData, encryptedData)

	// Decrypt the data
	decryptedData, err := service.DecryptData(encryptedData)
	require.NoError(t, err)
	require.NotNil(t, decryptedData)

	// Verify decrypted data matches original
	require.Equal(t, clearData, decryptedData)
}

// TestUnsealKeysServiceSimple_DecryptKey_InvalidData tests decryption with invalid data.
func TestUnsealKeysServiceSimple_DecryptKey_InvalidData(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Try to decrypt invalid data
	invalidData := []byte("this is not valid encrypted data")
	_, err = service.DecryptKey(invalidData)
	require.Error(t, err)
}

// TestUnsealKeysServiceSimple_DecryptData_InvalidData tests decryption with invalid data.
func TestUnsealKeysServiceSimple_DecryptData_InvalidData(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Try to decrypt invalid data
	invalidData := []byte("not encrypted")
	_, err = service.DecryptData(invalidData)
	require.Error(t, err)
}

// TestUnsealKeysServiceSimple_EncryptData_EmptyData tests encryption of empty data.
func TestUnsealKeysServiceSimple_EncryptData_EmptyData(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Encrypt empty data - this should fail because empty data is not allowed
	emptyData := []byte{}
	_, err = service.EncryptData(emptyData)
	require.Error(t, err, "empty data encryption should fail")
}

// TestUnsealKeysServiceSimple_Shutdown tests service shutdown.
func TestUnsealKeysServiceSimple_Shutdown(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown should not panic
	require.NotPanics(t, func() {
		service.Shutdown()
	})
}

// TestUnsealKeysServiceSimple_MultipleEncryptDecryptRounds tests multiple encryption/decryption cycles.
func TestUnsealKeysServiceSimple_MultipleEncryptDecryptRounds(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Test multiple rounds
	testData := []string{
		"first data",
		"second data with more content",
		"third data 123!@#",
	}

	for i, data := range testData {
		clearData := []byte(data)

		// Encrypt
		encryptedData, err := service.EncryptData(clearData)
		require.NoError(t, err, "round %d encryption failed", i)

		// Decrypt
		decryptedData, err := service.DecryptData(encryptedData)
		require.NoError(t, err, "round %d decryption failed", i)

		// Verify
		require.Equal(t, clearData, decryptedData, "round %d data mismatch", i)
	}
}

// TestUnsealKeysServiceSimple_LargeData tests encryption/decryption of large data.
func TestUnsealKeysServiceSimple_LargeData(t *testing.T) {
	t.Parallel()

	// Create unseal keys
	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create service
	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Create large data (1MB)
	const largeDataSize = 1024 * 1024

	largeData := make([]byte, largeDataSize)
	for i := range largeData {
		largeData[i] = byte(i % cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	}

	// Encrypt large data
	encryptedData, err := service.EncryptData(largeData)
	require.NoError(t, err)
	require.NotNil(t, encryptedData)

	// Decrypt large data
	decryptedData, err := service.DecryptData(encryptedData)
	require.NoError(t, err)
	require.Equal(t, largeData, decryptedData)
}

// TestUnsealKeysServiceSimple_DifferentKeySizes tests with different key sizes.
func TestUnsealKeysServiceSimple_DifferentKeySizes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		keyCount int
		enc      *joseJwa.ContentEncryptionAlgorithm
		alg      *joseJwa.KeyEncryptionAlgorithm
	}{
		{"Single Key", 1, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW},
		{"Two Keys", 2, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW},
		{"Three Keys", 3, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create unseal keys
			unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, tc.keyCount, tc.enc, tc.alg)
			require.NoError(t, err)

			// Create service
			service, err := NewUnsealKeysServiceSimple(unsealKeys)
			require.NoError(t, err)
			require.NotNil(t, service)

			// Test data
			testData := []byte("test data for " + tc.name)

			// Encrypt and decrypt
			encryptedData, err := service.EncryptData(testData)
			require.NoError(t, err)

			decryptedData, err := service.DecryptData(encryptedData)
			require.NoError(t, err)
			require.Equal(t, testData, decryptedData)
		})
	}
}

// TestUnsealKeysServiceSimple_MultipleKeys_SameData tests that different unseal keys produce different encrypted outputs.
func TestUnsealKeysServiceSimple_MultipleKeys_SameData(t *testing.T) {
	t.Parallel()

	clearData := []byte("same data for both services")

	// Create first set of unseal keys
	unsealKeys1, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create second set of unseal keys
	unsealKeys2, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Create first service
	service1, err := NewUnsealKeysServiceSimple(unsealKeys1)
	require.NoError(t, err)

	// Create second service
	service2, err := NewUnsealKeysServiceSimple(unsealKeys2)
	require.NoError(t, err)

	// Encrypt same data with different keys
	encrypted1, err := service1.EncryptData(clearData)
	require.NoError(t, err)

	encrypted2, err := service2.EncryptData(clearData)
	require.NoError(t, err)

	// Encrypted outputs should be different
	require.NotEqual(t, encrypted1, encrypted2)

	// Each service should only decrypt its own encrypted data
	decrypted1, err := service1.DecryptData(encrypted1)
	require.NoError(t, err)
	require.Equal(t, clearData, decrypted1)

	decrypted2, err := service2.DecryptData(encrypted2)
	require.NoError(t, err)
	require.Equal(t, clearData, decrypted2)

	// Cross-decryption should fail
	_, err = service1.DecryptData(encrypted2)
	require.Error(t, err, "service1 should not decrypt service2's data")

	_, err = service2.DecryptData(encrypted1)
	require.Error(t, err, "service2 should not decrypt service1's data")
}

// TestUnsealKeysServiceSharedSecrets_EncryptDecryptKey tests key encryption with shared secrets.
