// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"fmt"

	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// ContentKeysService encrypts and decrypts content data using content keys.
type ContentKeysService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	jwkGenService           *cryptoutilJose.JWKGenService
	repository              Repository
	intermediateKeysService *IntermediateKeysService
}

// NewContentKeysService creates a new ContentKeysService with the specified dependencies.
func NewContentKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, repository Repository, intermediateKeysService *IntermediateKeysService) (*ContentKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if repository == nil {
		return nil, fmt.Errorf("repository must be non-nil")
	} else if intermediateKeysService == nil {
		return nil, fmt.Errorf("intermediateKeysService must be non-nil")
	}

	return &ContentKeysService{telemetryService: telemetryService, jwkGenService: jwkGenService, repository: repository, intermediateKeysService: intermediateKeysService}, nil
}

// EncryptContent encrypts content data and returns the encrypted bytes and encryption key ID.
func (s *ContentKeysService) EncryptContent(sqlTransaction Transaction, clearContentBytes []byte) ([]byte, *googleUuid.UUID, error) {
	contentKeyKidUUID, clearContentKey, _, _, _, err := s.jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate content JWK: %w", err)
	}

	_, encryptedContentJWEMessageBytes, err := cryptoutilJose.EncryptBytesWithContext([]joseJwk.Key{clearContentKey}, clearContentBytes, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content with JWK: %w", err)
	}

	encryptedContentKeyJWEMessageBytes, intermediateKeyKidUUID, err := s.intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content JWK with intermediate JWK: %w", err)
	}

	err = sqlTransaction.AddContentKey(&ContentKey{UUID: *contentKeyKidUUID, Encrypted: string(encryptedContentKeyJWEMessageBytes), KEKUUID: *intermediateKeyKidUUID})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add content key to DB: %w", err)
	}

	return encryptedContentJWEMessageBytes, contentKeyKidUUID, nil
}

// DecryptContent decrypts content data using the content key identified in the JWE message.
func (s *ContentKeysService) DecryptContent(sqlTransaction Transaction, encryptedContentJWEMessageBytes []byte) ([]byte, error) {
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

	decryptedContentKey, err := s.intermediateKeysService.DecryptKey(sqlTransaction, []byte(encryptedContentKey.Encrypted))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content key: %w", err)
	}

	decryptedBytes, err := cryptoutilJose.DecryptBytesWithContext([]joseJwk.Key{decryptedContentKey}, encryptedContentJWEMessageBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content with content key: %w", err)
	}

	return decryptedBytes, nil
}

// Shutdown gracefully shuts down the ContentKeysService.
func (s *ContentKeysService) Shutdown() {
	s.telemetryService = nil
	s.repository = nil
	s.intermediateKeysService = nil
}
