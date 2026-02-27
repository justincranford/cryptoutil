// Copyright (c) 2025 Justin Cranford
//

package barrier

import (
	"context"
	"testing"

	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestRotateContentKey_NoIntermediateKeyFound(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.getIntermediateKeyLatestReturnsNil = true

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "no intermediate key found")
}

// TestRotateContentKey_AddContentKeyFailure tests error when adding new content key fails.
func TestRotateContentKey_AddContentKeyFailure(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	// Generate a valid encrypted key chain: unseal -> root -> intermediate
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	_, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Encrypt root key with unseal key
	encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
	require.NoError(t, err)

	// Encrypt intermediate key with root key
	_, encryptedIntermediateKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWK)
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: string(encryptedIntermediateKey),
		KEKUUID:   *rootKeyUUID,
	}
	mockRepo.tx.rootKey = &RootKey{
		UUID:      *rootKeyUUID,
		Encrypted: string(encryptedRootKey),
	}
	mockRepo.tx.addContentKeyErr = errMockServiceFailure

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to store new content key")
}

// TestRotateContentKey_GetRootKeyForDecryptionFailure tests error when getting root key for decryption fails.
func TestRotateContentKey_GetRootKeyForDecryptionFailure(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	// Generate a valid encrypted intermediate key chain: root -> intermediate
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	_, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Encrypt intermediate key with root key (needs valid JWE so kid can be extracted)
	_, encryptedIntermediateKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWK)
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: string(encryptedIntermediateKey),
		KEKUUID:   *rootKeyUUID,
	}
	mockRepo.tx.getRootKeyErr = errMockServiceFailure

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to get root key")
}

// TestRotateContentKey_InvalidJWEFormat tests error when intermediate key JWE is invalid.
func TestRotateContentKey_InvalidJWEFormat(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: "not-a-valid-jwe-string",
		KEKUUID:   googleUuid.New(),
	}

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to parse encrypted intermediate key")
}

// TestRotateContentKey_DecryptRootKeyFailure tests error when decrypting root key fails.
func TestRotateContentKey_DecryptRootKeyFailure(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	// Create a different unseal key to encrypt the root key (so decryption will fail)
	_, differentUnsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Generate root key and encrypt with the different unseal key
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Encrypt root key with different unseal key (not the one in unsealService)
	_, encryptedRootKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{differentUnsealJWK}, clearRootJWK)
	require.NoError(t, err)

	// Generate intermediate key and encrypt with root key
	_, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	_, encryptedIntermediateKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWK)
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: string(encryptedIntermediateKey),
		KEKUUID:   *rootKeyUUID,
	}
	mockRepo.tx.rootKey = &RootKey{
		UUID:      *rootKeyUUID,
		Encrypted: string(encryptedRootKey), // Encrypted with wrong unseal key
	}

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to decrypt root key")
}

// TestRotateContentKey_DecryptIntermediateKeyFailure tests error when decrypting intermediate key fails.
func TestRotateContentKey_DecryptIntermediateKeyFailure(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	// Generate root key and encrypt with unseal service
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
	require.NoError(t, err)

	// Generate a different root key to encrypt intermediate (so decryption will fail)
	_, differentRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Generate intermediate key and encrypt with DIFFERENT root key
	_, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Note: We need to set the kid in the JWE to match our rootKeyUUID so the lookup succeeds,
	// but encrypt with differentRootJWK so decryption fails.
	// We can do this by setting the kid on the encryption key.
	err = differentRootJWK.Set(joseJwk.KeyIDKey, rootKeyUUID.String())
	require.NoError(t, err)

	_, encryptedIntermediateKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{differentRootJWK}, clearIntermediateJWK)
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: string(encryptedIntermediateKey), // Encrypted with wrong root key
		KEKUUID:   *rootKeyUUID,
	}
	mockRepo.tx.rootKey = &RootKey{
		UUID:      *rootKeyUUID,
		Encrypted: string(encryptedRootKey),
	}

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to decrypt intermediate key")
}

// TestNewRotationService_NilParameters tests constructor validation.
func TestNewRotationService_NilParameters(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)
	mockRepo := newMockServiceRepository()

	tests := []struct {
		name               string
		jwkGenService      *cryptoutilSharedCryptoJose.JWKGenService
		repository         Repository
		unsealKeysService  cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrContain string
	}{
		{
			name:               "nil jwkGenService",
			jwkGenService:      nil,
			repository:         mockRepo,
			unsealKeysService:  unsealService,
			expectedErrContain: "jwkGenService must be non-nil",
		},
		{
			name:               "nil repository",
			jwkGenService:      jwkGenService,
			repository:         nil,
			unsealKeysService:  unsealService,
			expectedErrContain: "repository must be non-nil",
		},
		{
			name:               "nil unsealKeysService",
			jwkGenService:      jwkGenService,
			repository:         mockRepo,
			unsealKeysService:  nil,
			expectedErrContain: "unsealKeysService must be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := NewRotationService(tt.jwkGenService, tt.repository, tt.unsealKeysService)
			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tt.expectedErrContain)
		})
	}
}

// TestRotationService_WithTransactionError tests error when WithTransaction itself fails.
func TestRotationService_WithTransactionError(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.withTxErr = errMockServiceFailure
	mockRepo.shouldCallTxFn = false

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("RotateRootKey_TransactionError", func(t *testing.T) {
		result, err := rotationService.RotateRootKey(ctx, "test")
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "root key rotation transaction failed")
	})

	t.Run("RotateIntermediateKey_TransactionError", func(t *testing.T) {
		result, err := rotationService.RotateIntermediateKey(ctx, "test")
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "intermediate key rotation transaction failed")
	})

	t.Run("RotateContentKey_TransactionError", func(t *testing.T) {
		result, err := rotationService.RotateContentKey(ctx, "test")
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "content key rotation transaction failed")
	})
}
