// Copyright (c) 2025 Justin Cranford
//
//

// Package contentkeysservice provides content-level key management for the barrier encryption hierarchy.
package contentkeysservice

import (
	"fmt"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilIntermediateKeysService "cryptoutil/internal/shared/barrier/intermediatekeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// ContentKeysService manages content encryption keys in the barrier hierarchy.
type ContentKeysService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	jwkGenService           *cryptoutilJose.JWKGenService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
}

// NewContentKeysService creates a new ContentKeysService with the specified dependencies.
func NewContentKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService) (*ContentKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if intermediateKeysService == nil {
		return nil, fmt.Errorf("intermediateKeysService must be non-nil")
	}

	return &ContentKeysService{telemetryService: telemetryService, jwkGenService: jwkGenService, ormRepository: ormRepository, intermediateKeysService: intermediateKeysService}, nil
}

// EncryptContent encrypts content bytes with a newly generated content key.
func (s *ContentKeysService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearContentBytes []byte) ([]byte, *googleUuid.UUID, error) {
	contentKeyKidUUID, clearContentKey, _, _, _, err := s.jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate content JWK: %w", err)
	}

	_, encryptedContentJWEMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{clearContentKey}, clearContentBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content with JWK: %w", err)
	}

	encryptedContentKeyJWEMessageBytes, intermediateKeyKidUUID, err := s.intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content JWK with intermediate JWK: %w", err)
	}

	err = sqlTransaction.AddContentKey(&cryptoutilOrmRepository.BarrierContentKey{UUID: *contentKeyKidUUID, Encrypted: string(encryptedContentKeyJWEMessageBytes), KEKUUID: *intermediateKeyKidUUID})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add content key to DB: %w", err)
	}

	return encryptedContentJWEMessageBytes, contentKeyKidUUID, nil
}

// DecryptContent decrypts encrypted JWE message bytes using the content key hierarchy.
func (s *ContentKeysService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentJWEMessageBytes []byte) ([]byte, error) {
	encryptedContentJWEMessage, err := joseJwe.Parse(encryptedContentJWEMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message: %w", err)
	}

	var encryptedContentKeyKidString string

	err = encryptedContentJWEMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &encryptedContentKeyKidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message kid: %w", err)
	}

	encryptedContentKeyKidUUID, err := googleUuid.Parse(encryptedContentKeyKidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}

	encryptedContentKey, err := sqlTransaction.GetContentKey(&encryptedContentKeyKidUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted content key: %w", err)
	}

	decryptedContentKey, err := s.intermediateKeysService.DecryptKey(sqlTransaction, []byte(encryptedContentKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content key: %w", err)
	}

	decryptedBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedContentKey}, encryptedContentJWEMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content with content key: %w", err)
	}

	return decryptedBytes, nil
}

// Shutdown releases all resources held by the ContentKeysService.
func (s *ContentKeysService) Shutdown() {
	s.telemetryService = nil
	s.ormRepository = nil
	s.intermediateKeysService = nil
}
