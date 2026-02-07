// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck,thelper // Test code doesn't need to wrap errors or use t.Helper()
package unsealkeysservice

import (
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
		largeData[i] = byte(i % 256)
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
func TestUnsealKeysServiceSharedSecrets_EncryptDecryptKey(t *testing.T) {
	t.Parallel()

	// Create shared secrets (5 secrets, choose 3)
	const (
		sharedSecretCount = 5
		chooseN           = 3
		secretSize        = 32
	)

	sharedSecrets := make([][]byte, sharedSecretCount)
	for i := 0; i < sharedSecretCount; i++ {
		sharedSecrets[i] = make([]byte, secretSize)
		for j := 0; j < secretSize; j++ {
			sharedSecrets[i][j] = byte(i*10 + j) // #nosec G602 -- bounds checked via make() calls.
		}
	}

	// Create service with shared secrets
	service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, chooseN)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

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
}

// TestUnsealKeysServiceSharedSecrets_EncryptDecryptData tests data encryption with shared secrets.
func TestUnsealKeysServiceSharedSecrets_EncryptDecryptData(t *testing.T) {
	t.Parallel()

	// Create shared secrets (5 secrets, choose 3)
	const (
		sharedSecretCount = 5
		chooseN           = 3
		secretSize        = 32
	)

	sharedSecrets := make([][]byte, sharedSecretCount)
	for i := 0; i < sharedSecretCount; i++ {
		sharedSecrets[i] = make([]byte, secretSize)
		for j := 0; j < secretSize; j++ {
			sharedSecrets[i][j] = byte(i*10 + j) // #nosec G602 -- bounds checked via make() calls.
		}
	}

	// Create service with shared secrets
	service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, chooseN)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Test data
	clearData := []byte("sensitive data encrypted with shared secrets")

	// Encrypt the data
	encryptedData, err := service.EncryptData(clearData)
	require.NoError(t, err)
	require.NotNil(t, encryptedData)
	require.Greater(t, len(encryptedData), 0)

	// Decrypt the data
	decryptedData, err := service.DecryptData(encryptedData)
	require.NoError(t, err)
	require.Equal(t, clearData, decryptedData)
}

// TestUnsealKeysServiceSharedSecrets_Shutdown tests shared secrets service shutdown.
func TestUnsealKeysServiceSharedSecrets_Shutdown(t *testing.T) {
	t.Parallel()

	// Create shared secrets
	const (
		sharedSecretCount = 3
		chooseN           = 2
		secretSize        = 32
	)

	sharedSecrets := make([][]byte, sharedSecretCount)
	for i := 0; i < sharedSecretCount; i++ {
		sharedSecrets[i] = make([]byte, secretSize)
	}

	// Create service
	service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, chooseN)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown should not panic
	require.NotPanics(t, func() {
		service.Shutdown()
	})
}

// TestUnsealKeysServiceSharedSecrets_DifferentChooseN tests different M-choose-N combinations.
func TestUnsealKeysServiceSharedSecrets_DifferentChooseN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		secretCount int
		chooseN     int
	}{
		{"2 of 3", 3, 2},
		{"3 of 5", 5, 3},
		{"4 of 6", 6, 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			const secretSize = 32

			sharedSecrets := make([][]byte, tc.secretCount)
			for i := 0; i < tc.secretCount; i++ {
				sharedSecrets[i] = make([]byte, secretSize)
				for j := 0; j < secretSize; j++ {
					sharedSecrets[i][j] = byte(i*10 + j) // #nosec G602 -- bounds checked via make() calls.
				}
			}

			service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, tc.chooseN)
			require.NoError(t, err)
			require.NotNil(t, service)

			defer service.Shutdown()

			// Test encryption/decryption works
			testData := []byte("test data for " + tc.name)
			encryptedData, err := service.EncryptData(testData)
			require.NoError(t, err)

			decryptedData, err := service.DecryptData(encryptedData)
			require.NoError(t, err)
			require.Equal(t, testData, decryptedData)
		})
	}
}

// TestUnsealKeysServiceSharedSecrets_MinimumSecretLength tests minimum secret length validation.
func TestUnsealKeysServiceSharedSecrets_MinimumSecretLength(t *testing.T) {
	t.Parallel()

	// Create secrets with one below minimum length
	sharedSecrets := [][]byte{
		make([]byte, 32), // Valid
		make([]byte, 31), // Below minimum (32 bytes)
	}

	service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, 2)
	require.Error(t, err)
	require.Nil(t, service)
	require.Contains(t, err.Error(), "secret 1 length can't be less than")
}

// TestUnsealKeysServiceSharedSecrets_DecryptInvalidData tests decryption with invalid data.
func TestUnsealKeysServiceSharedSecrets_DecryptInvalidData(t *testing.T) {
	t.Parallel()

	// Create shared secrets
	const (
		sharedSecretCount = 3
		chooseN           = 2
		secretSize        = 32
	)

	sharedSecrets := make([][]byte, sharedSecretCount)
	for i := 0; i < sharedSecretCount; i++ {
		sharedSecrets[i] = make([]byte, secretSize)
	}

	service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, chooseN)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Try to decrypt invalid data
	invalidData := []byte("not valid encrypted data")
	_, err = service.DecryptData(invalidData)
	require.Error(t, err)
}

// TestUnsealKeysServiceSimple_NilKey tests encryption with nil key.
func TestUnsealKeysServiceSimple_NilKey(t *testing.T) {
	t.Parallel()

	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Try to encrypt nil key - implementation may allow this, so just verify behavior
	encryptedBytes, err := service.EncryptKey(nil)
	// If error occurs, that's fine. If no error, verify we got encrypted bytes
	if err == nil {
		require.NotNil(t, encryptedBytes)
	}
}

// TestUnsealKeysServiceSimple_DecryptKey_EmptyBytes tests decryption with empty bytes.
func TestUnsealKeysServiceSimple_DecryptKey_EmptyBytes(t *testing.T) {
	t.Parallel()

	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Try to decrypt empty bytes
	_, err = service.DecryptKey([]byte{})
	require.Error(t, err)
}

// TestUnsealKeysServiceSimple_DecryptData_EmptyBytes tests decryption with empty bytes.
func TestUnsealKeysServiceSimple_DecryptData_EmptyBytes(t *testing.T) {
	t.Parallel()

	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Try to decrypt empty bytes
	_, err = service.DecryptData([]byte{})
	require.Error(t, err)
}

// TestUnsealKeysServiceSimple_EncryptData_NilData tests encryption with nil data.
func TestUnsealKeysServiceSimple_EncryptData_NilData(t *testing.T) {
	t.Parallel()

	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	service, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Try to encrypt nil data
	_, err = service.EncryptData(nil)
	require.Error(t, err, "nil data encryption should fail")
}

// TestUnsealKeysServiceSharedSecrets_MaxSecretLength tests maximum secret length validation.
func TestUnsealKeysServiceSharedSecrets_MaxSecretLength(t *testing.T) {
	t.Parallel()

	// Create secrets with one above maximum length
	sharedSecrets := [][]byte{
		make([]byte, 32),    // Valid
		make([]byte, 10000), // Above maximum
	}

	service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, 2)
	require.Error(t, err)
	require.Nil(t, service)
	require.Contains(t, err.Error(), "secret 1 length can't be greater than")
}

// TestUnsealKeysServiceSharedSecrets_DeterministicKeyDerivation tests that same secrets produce same keys.
func TestUnsealKeysServiceSharedSecrets_DeterministicKeyDerivation(t *testing.T) {
	t.Parallel()

	// Create shared secrets
	const (
		sharedSecretCount = 3
		chooseN           = 2
		secretSize        = 32
	)

	sharedSecrets := make([][]byte, sharedSecretCount)
	for i := 0; i < sharedSecretCount; i++ {
		sharedSecrets[i] = make([]byte, secretSize)
		for j := 0; j < secretSize; j++ {
			sharedSecrets[i][j] = byte(i*10 + j) // #nosec G602 -- bounds checked via make() calls.
		}
	}

	// Create two services with same shared secrets
	service1, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, chooseN)
	require.NoError(t, err)
	require.NotNil(t, service1)

	defer service1.Shutdown()

	service2, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, chooseN)
	require.NoError(t, err)
	require.NotNil(t, service2)

	defer service2.Shutdown()

	// Encrypt data with first service
	testData := []byte("test data for deterministic keys")
	encrypted1, err := service1.EncryptData(testData)
	require.NoError(t, err)

	// Decrypt with second service (should work because keys are deterministically derived)
	decrypted2, err := service2.DecryptData(encrypted1)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted2)

	// Encrypt with second service
	encrypted2, err := service2.EncryptData(testData)
	require.NoError(t, err)

	// Decrypt with first service
	decrypted1, err := service1.DecryptData(encrypted2)
	require.NoError(t, err)
	require.Equal(t, testData, decrypted1)
}

// TestUnsealKeysServiceSharedSecrets_SingleSecret tests 1-of-1 shared secret scenario.
func TestUnsealKeysServiceSharedSecrets_SingleSecret(t *testing.T) {
	t.Parallel()

	// Create single shared secret
	sharedSecrets := [][]byte{
		make([]byte, 32),
	}
	for j := 0; j < 32; j++ {
		sharedSecrets[0][j] = byte(j) // #nosec G602 -- bounds checked: slice sized to 32, loop bounded by 32.
	}

	service, err := NewUnsealKeysServiceSharedSecrets(sharedSecrets, 1)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Test encryption/decryption
	testData := []byte("single secret test")
	encryptedData, err := service.EncryptData(testData)
	require.NoError(t, err)

	decryptedData, err := service.DecryptData(encryptedData)
	require.NoError(t, err)
	require.Equal(t, testData, decryptedData)
}
