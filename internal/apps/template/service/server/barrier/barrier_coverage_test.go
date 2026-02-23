// Copyright (c) 2025 Justin Cranford
//

package barrier

import (
	"context"
	crand "crypto/rand"
	"testing"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// TestRootKeyInit_AddRootKeyFails verifies that initializeFirstRootJWK fails when
// AddRootKey returns an error after generate and encrypt succeed.
func TestRootKeyInit_AddRootKeyFails(t *testing.T) {
	t.Parallel()

	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.getRootKeyLatestReturnsNil = true
	mockRepo.tx.addRootKeyErr = errMockServiceFailure

	_, err := NewRootKeysService(telemetryService, jwkGenService, mockRepo, unsealService)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to initialize first root JWK")
}

// TestIntermediateKeyInit_RootEncryptKeyFails verifies that intermediate key init fails
// when the root key in the mock has invalid encrypted data, causing EncryptKey to fail
// (unseal decrypt error → root EncryptKey error → intermediate init tx error).
func TestIntermediateKeyInit_RootEncryptKeyFails(t *testing.T) {
	t.Parallel()

	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	mockRepo := newMockServiceRepository()
	// Root key exists but with invalid encrypted data (root init skips creation).
	mockRepo.tx.rootKey = &RootKey{UUID: googleUuid.New(), Encrypted: "invalid", KEKUUID: googleUuid.Nil}
	// No intermediate key → init tries to create one.
	mockRepo.tx.getIntermediateKeyLatestReturnsNil = true

	ctx := context.Background()

	_, err := NewService(ctx, telemetryService, jwkGenService, mockRepo, unsealService)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create intermediate keys service")
}

// TestIntermediateKeyInit_AddIntermediateKeyFails verifies that intermediate key init fails
// when AddIntermediateKey returns an error after generate and encrypt succeed.
func TestIntermediateKeyInit_AddIntermediateKeyFails(t *testing.T) {
	t.Parallel()

	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	// Create a properly encrypted root key for the mock.
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	encryptedRootKeyBytes, err := unsealService.EncryptKey(clearRootJWK)
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKeyBytes), KEKUUID: googleUuid.Nil}
	mockRepo.tx.getIntermediateKeyLatestReturnsNil = true
	mockRepo.tx.addIntermediateKeyErr = errMockServiceFailure

	ctx := context.Background()

	_, err = NewService(ctx, telemetryService, jwkGenService, mockRepo, unsealService)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create intermediate keys service")
}

// TestRootKeysService_DecryptKey_ErrorPaths tests error paths in RootKeysService.DecryptKey:
// GetRootKey failure and unseal DecryptKey failure.
func TestRootKeysService_DecryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	// Create a properly encrypted root key.
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	encryptedRootKeyBytes, err := unsealService.EncryptKey(clearRootJWK)
	require.NoError(t, err)

	// Construct RootKeysService with existing root key (init succeeds).
	mockRepo := newMockServiceRepository()
	mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKeyBytes), KEKUUID: googleUuid.Nil}

	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	t.Cleanup(func() { rootKeysService.Shutdown() })

	// Generate valid JWE bytes for DecryptKey input (encrypted with root key so kid matches).
	_, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	_, encryptedIntermediateBytes, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWK)
	require.NoError(t, err)

	t.Run("GetRootKeyFails", func(t *testing.T) {
		t.Parallel()

		failTx := &mockServiceTransaction{getRootKeyErr: errMockServiceFailure}

		_, err := rootKeysService.DecryptKey(failTx, encryptedIntermediateBytes)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get root key")
	})

	t.Run("UnsealDecryptFails", func(t *testing.T) {
		t.Parallel()

		badRootKeyTx := &mockServiceTransaction{
			rootKey: &RootKey{UUID: *rootKeyUUID, Encrypted: "invalid-encrypted-data"},
		}

		_, err := rootKeysService.DecryptKey(badRootKeyTx, encryptedIntermediateBytes)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt root key")
	})

	t.Run("NoKidInJWE", func(t *testing.T) {
		t.Parallel()

		// Create a raw AES key without kid, encrypt to produce JWE without kid header.
		rawKey := make([]byte, 32)
		_, err := crand.Read(rawKey)
		require.NoError(t, err)

		noKidKey, err := joseJwk.Import(rawKey)
		require.NoError(t, err)

		noKidJWEBytes, err := joseJwe.Encrypt([]byte("test-content"), joseJwe.WithKey(joseJwa.A256KW(), noKidKey), joseJwe.WithContentEncryption(joseJwa.A256GCM()))
		require.NoError(t, err)

		_, err = rootKeysService.DecryptKey(mockRepo.tx, noKidJWEBytes)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse encrypted intermediate key message kid UUID")
	})

	t.Run("JOSEDecryptKeyMismatch", func(t *testing.T) {
		t.Parallel()

		// Root key B is different from root key A (used to encrypt intermediate bytes).
		_, clearRootKeyB, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
		require.NoError(t, err)

		encryptedRootKeyBBytes, err := unsealService.EncryptKey(clearRootKeyB)
		require.NoError(t, err)

		// Mock returns root key B (different from A used for encryption).
		mismatchTx := &mockServiceTransaction{
			rootKey: &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKeyBBytes)},
		}

		_, err = rootKeysService.DecryptKey(mismatchTx, encryptedIntermediateBytes)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt intermediate key")
	})
}

// TestIntermediateKeysService_DecryptKey_KeyMismatch tests that DecryptKey fails when the
// intermediate key used for decryption differs from the one used for encryption.
// Covers the final JOSE DecryptKey error path in IntermediateKeysService.DecryptKey.
func TestIntermediateKeysService_DecryptKey_KeyMismatch(t *testing.T) {
	t.Parallel()

	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	// Create encrypted root key.
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	encryptedRootKeyBytes, err := unsealService.EncryptKey(clearRootJWK)
	require.NoError(t, err)

	// Create intermediate key A (used to encrypt content key).
	_, clearIntermediateJWKA, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	// Create intermediate key B (different, returned by mock during decryption).
	_, clearIntermediateJWKB, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	// Encrypt key B with root key (what mock will return).
	_, encryptedIntermediateKeyB, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWKB)
	require.NoError(t, err)

	// Encrypt a content key with key A (produces JWE with key A's kid).
	_, clearContentKey, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	_, encryptedContentKeyBytes, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearIntermediateJWKA}, clearContentKey)
	require.NoError(t, err)

	// Set up mock repo: root key exists, intermediate key returns B's encrypted data.
	mockRepo := newMockServiceRepository()
	mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKeyBytes), KEKUUID: googleUuid.Nil}
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: string(encryptedIntermediateKeyB),
		KEKUUID:   *rootKeyUUID,
	}

	// Construct services.
	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	t.Cleanup(func() { rootKeysService.Shutdown() })

	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, jwkGenService, mockRepo, rootKeysService)
	require.NoError(t, err)

	t.Cleanup(func() { intermediateKeysService.Shutdown() })

	// DecryptKey: parses JWE → gets kid A → GetIntermediateKey returns key B → decrypt B →
	// try to decrypt content with key B (encrypted with key A) → FAILS.
	_, err = intermediateKeysService.DecryptKey(mockRepo.tx, encryptedContentKeyBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt content key")
}

// TestIntermediateKeysService_DecryptKey_NoKidInJWE tests that DecryptKey fails when
// the JWE input has no kid header (ProtectedHeaders().Get fails).
func TestIntermediateKeysService_DecryptKey_NoKidInJWE(t *testing.T) {
	t.Parallel()

	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	// Create encrypted root key.
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	encryptedRootKeyBytes, err := unsealService.EncryptKey(clearRootJWK)
	require.NoError(t, err)

	// Set up mock repo.
	mockRepo := newMockServiceRepository()
	mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKeyBytes), KEKUUID: googleUuid.Nil}
	mockRepo.tx.intermediateKey = &IntermediateKey{UUID: googleUuid.New(), Encrypted: "dummy", KEKUUID: *rootKeyUUID}

	// Construct services.
	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	t.Cleanup(func() { rootKeysService.Shutdown() })

	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, jwkGenService, mockRepo, rootKeysService)
	require.NoError(t, err)

	t.Cleanup(func() { intermediateKeysService.Shutdown() })

	// Create a raw AES key without kid, then encrypt to produce JWE without kid header.
	rawKey := make([]byte, 32)
	_, err = crand.Read(rawKey)
	require.NoError(t, err)

	noKidKey, err := joseJwk.Import(rawKey)
	require.NoError(t, err)

	noKidJWEBytes, err := joseJwe.Encrypt([]byte("test-content"), joseJwe.WithKey(joseJwa.A256KW(), noKidKey), joseJwe.WithContentEncryption(joseJwa.A256GCM()))
	require.NoError(t, err)

	// DecryptKey: parse JWE → ProtectedHeaders().Get(kid) FAILS (no kid in header).
	_, err = intermediateKeysService.DecryptKey(mockRepo.tx, noKidJWEBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse encrypted content key message kid UUID")
}

// TestRotateContentKey_NoKidInIntermediateJWE tests that RotateContentKey fails
// when the intermediate key's encrypted JWE has no kid header.
func TestRotateContentKey_NoKidInIntermediateJWE(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	// Create a JWE without kid (raw key, no kid set).
	rawKey := make([]byte, 32)
	_, err := crand.Read(rawKey)
	require.NoError(t, err)

	noKidKey, err := joseJwk.Import(rawKey)
	require.NoError(t, err)

	noKidJWEBytes, err := joseJwe.Encrypt([]byte("test-key-material"), joseJwe.WithKey(joseJwa.A256KW(), noKidKey), joseJwe.WithContentEncryption(joseJwa.A256GCM()))
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: string(noKidJWEBytes), // JWE without kid header.
		KEKUUID:   googleUuid.New(),
	}

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()

	result, err := rotationService.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to get root key kid")
}
