// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"context"
	"errors"
	"fmt"
	"log"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// IntermediateKeysService manages intermediate encryption keys in the key hierarchy.
type IntermediateKeysService struct {
	telemetryService *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService    *cryptoutilSharedCryptoJose.JWKGenService
	repository       Repository
	rootKeysService  *RootKeysService
}

// intermediateGenerateJWEJWKFn is an injectable var for testing the error path when generating a new intermediate JWK.
var intermediateGenerateJWEJWKFn = func(svc *cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	return svc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
}

// NewIntermediateKeysService creates a new IntermediateKeysService with the specified dependencies.
func NewIntermediateKeysService(telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, repository Repository, rootKeysService *RootKeysService) (*IntermediateKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if repository == nil {
		return nil, fmt.Errorf("repository must be non-nil")
	} else if rootKeysService == nil {
		return nil, fmt.Errorf("rootKeysService must be non-nil")
	}

	err := initializeFirstIntermediateJWK(jwkGenService, repository, rootKeysService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize first intermediate JWK: %w", err)
	}

	return &IntermediateKeysService{telemetryService: telemetryService, jwkGenService: jwkGenService, repository: repository, rootKeysService: rootKeysService}, nil
}

func initializeFirstIntermediateJWK(jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, repository Repository, rootKeysService *RootKeysService) error {
	var encryptedIntermediateKeyLatest *IntermediateKey

	var err error

	err = repository.WithTransaction(context.Background(), func(sqlTransaction Transaction) error {
		encryptedIntermediateKeyLatest, err = sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK from DB
		// NOTE: "no intermediate key found" is EXPECTED on first run - don't treat as fatal error
		if err != nil && !errors.Is(err, ErrNoIntermediateKeyFound) {
			return fmt.Errorf("failed to get intermediate key latest: %w", err)
		}

		return nil
	})

	// DEBUG: Log error handling decision.
	isNoIntermediateKeyErr := errors.Is(err, ErrNoIntermediateKeyFound)
	log.Printf("DEBUG initializeFirstIntermediateJWK: err=%v, isNoIntermediateKeyFound=%v, encryptedIntermediateKeyLatest=%v", err, isNoIntermediateKeyErr, encryptedIntermediateKeyLatest)

	if err != nil && !isNoIntermediateKeyErr {
		return fmt.Errorf("failed to get encrypted intermediate JWK latest from DB: %w", err)
	}

	if encryptedIntermediateKeyLatest == nil {
		log.Printf("DEBUG initializeFirstIntermediateJWK: Creating first intermediate JWK")

		intermediateKeyKidUUID, clearIntermediateKey, _, _, _, err := intermediateGenerateJWEJWKFn(jwkGenService)
		if err != nil {
			log.Printf("DEBUG initializeFirstIntermediateJWK: GenerateJWEJWK failed: %v", err)

			return fmt.Errorf("failed to generate first intermediate JWK: %w", err)
		}

		log.Printf("DEBUG initializeFirstIntermediateJWK: Generated JWK with kid=%v", intermediateKeyKidUUID)

		var encryptedIntermediateKeyBytes []byte

		var rootKeyKidUUID *googleUuid.UUID

		err = repository.WithTransaction(context.Background(), func(sqlTransaction Transaction) error {
			encryptedIntermediateKeyBytes, rootKeyKidUUID, err = rootKeysService.EncryptKey(sqlTransaction, clearIntermediateKey)
			if err != nil {
				log.Printf("DEBUG initializeFirstIntermediateJWK: EncryptKey failed: %v", err)

				return fmt.Errorf("failed to encrypt first intermediate JWK: %w", err)
			}

			log.Printf("DEBUG initializeFirstIntermediateJWK: Encrypted intermediate JWK, len=%d, rootKeyKid=%v", len(encryptedIntermediateKeyBytes), rootKeyKidUUID)

			firstEncryptedIntermediateKey := &IntermediateKey{UUID: *intermediateKeyKidUUID, Encrypted: string(encryptedIntermediateKeyBytes), KEKUUID: *rootKeyKidUUID}

			err = sqlTransaction.AddIntermediateKey(firstEncryptedIntermediateKey)
			if err != nil {
				log.Printf("DEBUG initializeFirstIntermediateJWK: AddIntermediateKey failed: %v", err)

				return fmt.Errorf("failed to store first intermediate JWK: %w", err)
			}

			log.Printf("DEBUG initializeFirstIntermediateJWK: Successfully stored first intermediate JWK")

			return nil
		})
		if err != nil {
			log.Printf("DEBUG initializeFirstIntermediateJWK: Transaction failed: %v", err)

			return fmt.Errorf("failed to encrypt and store first intermediate first JWK: %w", err)
		}

		log.Printf("DEBUG initializeFirstIntermediateJWK: Successfully created first intermediate JWK")
	}

	return nil
}

// EncryptKey encrypts a content key using the latest intermediate key.
func (i *IntermediateKeysService) EncryptKey(sqlTransaction Transaction, clearContentKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedIntermediateKeyLatest, err := sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted intermediate JWK latest from DB: %w", err)
	}

	intermediateKeyLatestKidUUID := encryptedIntermediateKeyLatest.UUID

	decryptedIntermediateKeyLatest, err := i.rootKeysService.DecryptKey(sqlTransaction, []byte(encryptedIntermediateKeyLatest.Encrypted))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt intermediate JWK latest: %w", err)
	}

	_, encryptedContentKeyBytes, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{decryptedIntermediateKeyLatest}, clearContentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content JWK with intermediate JWK: %w", err)
	}

	return encryptedContentKeyBytes, &intermediateKeyLatestKidUUID, nil
}

// DecryptKey decrypts a content key using the identified intermediate key.
func (i *IntermediateKeysService) DecryptKey(sqlTransaction Transaction, encryptedContentKeyBytes []byte) (joseJwk.Key, error) {
	encryptedContentKey, err := joseJwe.Parse(encryptedContentKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted content key message: %w", err)
	}

	var intermediateKeyKidUUIDString string

	err = encryptedContentKey.ProtectedHeaders().Get(joseJwk.KeyIDKey, &intermediateKeyKidUUIDString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted content key message kid UUID: %w", err)
	}

	intermediateKeyKidUUID, err := googleUuid.Parse(intermediateKeyKidUUIDString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	encryptedIntermediateKey, err := sqlTransaction.GetIntermediateKey(&intermediateKeyKidUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get intermediate key: %w", err)
	}

	decryptedIntermediateKey, err := i.rootKeysService.DecryptKey(sqlTransaction, []byte(encryptedIntermediateKey.Encrypted))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedContentKey, err := cryptoutilSharedCryptoJose.DecryptKey([]joseJwk.Key{decryptedIntermediateKey}, encryptedContentKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content key: %w", err)
	}

	return decryptedContentKey, nil
}

// Shutdown gracefully shuts down the IntermediateKeysService.
func (i *IntermediateKeysService) Shutdown() {
	i.telemetryService = nil
	i.repository = nil
	i.jwkGenService = nil
	i.rootKeysService = nil
}
