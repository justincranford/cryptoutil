// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"context"
	"fmt"

	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// RotationService provides manual key rotation functionality.
// NOTE: This implementation uses elastic key rotation - new keys are created
// but old keys remain in the database for decrypting historical data.
// Re-encryption of dependent keys is NOT performed automatically.
type RotationService struct {
	jwkGenService     *cryptoutilSharedCryptoJose.JWKGenService
	repository        Repository
	unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
}

// NewRotationService creates a new rotation service.
func NewRotationService(
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	repository Repository,
	unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService,
) (*RotationService, error) {
	if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	}

	if repository == nil {
		return nil, fmt.Errorf("repository must be non-nil")
	}

	if unsealKeysService == nil {
		return nil, fmt.Errorf("unsealKeysService must be non-nil")
	}

	return &RotationService{
		jwkGenService:     jwkGenService,
		repository:        repository,
		unsealKeysService: unsealKeysService,
	}, nil
}

// RotateRootKeyResult contains the result of root key rotation.
type RotateRootKeyResult struct {
	OldKeyUUID googleUuid.UUID
	NewKeyUUID googleUuid.UUID
	Reason     string
}

// RotateRootKey creates a new root key encrypted with the unseal key.
// Old root keys remain in the database for decrypting historical intermediate keys.
// New intermediate keys will be encrypted with the new root key.
func (s *RotationService) RotateRootKey(ctx context.Context, reason string) (*RotateRootKeyResult, error) {
	var result *RotateRootKeyResult

	err := s.repository.WithTransaction(ctx, func(tx Transaction) error {
		// Get current root key (will be the old key after rotation)
		oldRootKey, err := tx.GetRootKeyLatest()
		if err != nil {
			return fmt.Errorf("failed to get current root key: %w", err)
		}

		if oldRootKey == nil {
			return fmt.Errorf("no root key found - cannot rotate")
		}

		// Generate new root JWK
		rootKeyKidUUID, clearRootKey, _, _, _, err := s.jwkGenService.GenerateJWEJWK(
			&cryptoutilSharedCryptoJose.EncA256GCM,
			&cryptoutilSharedCryptoJose.AlgDir,
		)
		if err != nil {
			return fmt.Errorf("failed to generate root JWK: %w", err)
		}

		// Encrypt with unseal key
		encryptedRootKeyBytes, err := s.unsealKeysService.EncryptKey(clearRootKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt root key: %w", err)
		}

		// Store new root key
		newRootKey := &RootKey{
			UUID:      *rootKeyKidUUID,
			Encrypted: string(encryptedRootKeyBytes),
			KEKUUID:   googleUuid.UUID{}, // Root keys have no parent
		}

		if err := tx.AddRootKey(newRootKey); err != nil {
			return fmt.Errorf("failed to store new root key: %w", err)
		}

		// Store rotation result
		result = &RotateRootKeyResult{
			OldKeyUUID: oldRootKey.UUID,
			NewKeyUUID: *rootKeyKidUUID,
			Reason:     reason,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("root key rotation transaction failed: %w", err)
	}

	return result, nil
}

// RotateIntermediateKeyResult contains the result of intermediate key rotation.
type RotateIntermediateKeyResult struct {
	OldKeyUUID googleUuid.UUID
	NewKeyUUID googleUuid.UUID
	Reason     string
}

// RotateIntermediateKey creates a new intermediate key encrypted with the current root key.
// Old intermediate keys remain in the database for decrypting historical content keys.
// New content keys will be encrypted with the new intermediate key.
func (s *RotationService) RotateIntermediateKey(ctx context.Context, reason string) (*RotateIntermediateKeyResult, error) {
	var result *RotateIntermediateKeyResult

	err := s.repository.WithTransaction(ctx, func(tx Transaction) error {
		// Get current intermediate key (will be the old key after rotation)
		oldIntermediateKey, err := tx.GetIntermediateKeyLatest()
		if err != nil {
			return fmt.Errorf("failed to get current intermediate key: %w", err)
		}

		if oldIntermediateKey == nil {
			return fmt.Errorf("no intermediate key found - cannot rotate")
		}

		// Get current root key for encryption
		currentRootKey, err := tx.GetRootKeyLatest()
		if err != nil {
			return fmt.Errorf("failed to get current root key: %w", err)
		}

		if currentRootKey == nil {
			return fmt.Errorf("no root key found")
		}

		// Decrypt root key
		clearRootKey, err := s.unsealKeysService.DecryptKey([]byte(currentRootKey.Encrypted))
		if err != nil {
			return fmt.Errorf("failed to decrypt root key: %w", err)
		}

		// Generate new intermediate JWK
		intermediateKeyKidUUID, clearIntermediateKey, _, _, _, err := s.jwkGenService.GenerateJWEJWK(
			&cryptoutilSharedCryptoJose.EncA256GCM,
			&cryptoutilSharedCryptoJose.AlgA256KW,
		)
		if err != nil {
			return fmt.Errorf("failed to generate intermediate JWK: %w", err)
		}

		// Encrypt new intermediate key with current root key
		_, encryptedIntermediateKeyBytes, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootKey}, clearIntermediateKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt intermediate key: %w", err)
		}

		// Store new intermediate key
		newIntermediateKey := &IntermediateKey{
			UUID:      *intermediateKeyKidUUID,
			Encrypted: string(encryptedIntermediateKeyBytes),
			KEKUUID:   currentRootKey.UUID,
		}

		if err := tx.AddIntermediateKey(newIntermediateKey); err != nil {
			return fmt.Errorf("failed to store new intermediate key: %w", err)
		}

		// Store rotation result
		result = &RotateIntermediateKeyResult{
			OldKeyUUID: oldIntermediateKey.UUID,
			NewKeyUUID: *intermediateKeyKidUUID,
			Reason:     reason,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("intermediate key rotation transaction failed: %w", err)
	}

	return result, nil
}

// RotateContentKeyResult contains the result of content key rotation.
type RotateContentKeyResult struct {
	NewKeyUUID googleUuid.UUID
	Reason     string
}

// RotateContentKey creates a new content key encrypted with the current intermediate key.
// Content keys don't have GetLatest so we can't track old key UUID.
// Old content keys remain in the database for decrypting historical data.
// New encryptions will use the new content key.
func (s *RotationService) RotateContentKey(ctx context.Context, reason string) (*RotateContentKeyResult, error) {
	var result *RotateContentKeyResult

	err := s.repository.WithTransaction(ctx, func(tx Transaction) error {
		// Get current intermediate key for encryption
		currentIntermediateKey, err := tx.GetIntermediateKeyLatest()
		if err != nil {
			return fmt.Errorf("failed to get current intermediate key: %w", err)
		}

		if currentIntermediateKey == nil {
			return fmt.Errorf("no intermediate key found")
		}

		// Parse and decrypt intermediate key to get clear key for encryption
		encryptedIntermediateKeyMsg, err := joseJwe.Parse([]byte(currentIntermediateKey.Encrypted))
		if err != nil {
			return fmt.Errorf("failed to parse encrypted intermediate key: %w", err)
		}

		// Get root key kid from intermediate key's JWE header
		var rootKeyKidString string

		err = encryptedIntermediateKeyMsg.ProtectedHeaders().Get(joseJwk.KeyIDKey, &rootKeyKidString)
		if err != nil {
			return fmt.Errorf("failed to get root key kid: %w", err)
		}

		rootKeyKidUUID, err := googleUuid.Parse(rootKeyKidString)
		if err != nil {
			return fmt.Errorf("failed to parse root key kid: %w", err)
		}

		// Get and decrypt root key
		encryptedRootKey, err := tx.GetRootKey(&rootKeyKidUUID)
		if err != nil {
			return fmt.Errorf("failed to get root key: %w", err)
		}

		clearRootKey, err := s.unsealKeysService.DecryptKey([]byte(encryptedRootKey.Encrypted))
		if err != nil {
			return fmt.Errorf("failed to decrypt root key: %w", err)
		}

		// Decrypt intermediate key
		clearIntermediateKey, err := cryptoutilSharedCryptoJose.DecryptKey([]joseJwk.Key{clearRootKey}, []byte(currentIntermediateKey.Encrypted))
		if err != nil {
			return fmt.Errorf("failed to decrypt intermediate key: %w", err)
		}

		// Generate new content JWK
		contentKeyKidUUID, clearContentKey, _, _, _, err := s.jwkGenService.GenerateJWEJWK(
			&cryptoutilSharedCryptoJose.EncA256GCM,
			&cryptoutilSharedCryptoJose.AlgA256KW,
		)
		if err != nil {
			return fmt.Errorf("failed to generate content JWK: %w", err)
		}

		// Encrypt content key with intermediate key
		_, encryptedContentKeyBytes, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearIntermediateKey}, clearContentKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt content key: %w", err)
		}

		// Store new content key
		newContentKey := &ContentKey{
			UUID:      *contentKeyKidUUID,
			Encrypted: string(encryptedContentKeyBytes),
			KEKUUID:   currentIntermediateKey.UUID,
		}

		if err := tx.AddContentKey(newContentKey); err != nil {
			return fmt.Errorf("failed to store new content key: %w", err)
		}

		// Store rotation result (no old key UUID since no GetLatest)
		result = &RotateContentKeyResult{
			NewKeyUUID: *contentKeyKidUUID,
			Reason:     reason,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("content key rotation transaction failed: %w", err)
	}

	return result, nil
}
