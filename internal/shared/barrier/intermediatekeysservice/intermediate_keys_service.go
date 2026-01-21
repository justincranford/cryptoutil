// Copyright (c) 2025 Justin Cranford
//
//

// Package intermediatekeysservice provides intermediate-level key management for the barrier encryption hierarchy.
package intermediatekeysservice

import (
	"context"
	"errors"
	"fmt"
	"log"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilRootKeysService "cryptoutil/internal/shared/barrier/rootkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// IntermediateKeysService manages intermediate encryption keys in the barrier hierarchy.
type IntermediateKeysService struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	jwkGenService    *cryptoutilJose.JWKGenService
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	rootKeysService  *cryptoutilRootKeysService.RootKeysService
}

// NewIntermediateKeysService creates a new IntermediateKeysService with the specified dependencies.
func NewIntermediateKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService) (*IntermediateKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if rootKeysService == nil {
		return nil, fmt.Errorf("rootKeysService must be non-nil")
	}

	err := initializeFirstIntermediateJWK(jwkGenService, ormRepository, rootKeysService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize first intermediate JWK: %w", err)
	}

	return &IntermediateKeysService{telemetryService: telemetryService, jwkGenService: jwkGenService, ormRepository: ormRepository, rootKeysService: rootKeysService}, nil
}

func initializeFirstIntermediateJWK(jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService) error {
	var encryptedIntermediateKeyLatest *cryptoutilOrmRepository.BarrierIntermediateKey

	var err error

	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedIntermediateKeyLatest, err = sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK from DB
		// NOTE: "record not found" is EXPECTED on first run - don't treat as fatal error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get intermediate key latest: %w", err)
		}

		return nil
	})

	// DEBUG: Log error handling decision.
	isRecordNotFoundErr := errors.Is(err, gorm.ErrRecordNotFound)
	log.Printf("DEBUG initializeFirstIntermediateJWK: err=%v, isRecordNotFound=%v, encryptedIntermediateKeyLatest=%v", err, isRecordNotFoundErr, encryptedIntermediateKeyLatest)

	if err != nil && !isRecordNotFoundErr {
		return fmt.Errorf("failed to get encrypted intermediate JWK latest from DB: %w", err)
	}

	if encryptedIntermediateKeyLatest == nil {
		log.Printf("DEBUG initializeFirstIntermediateJWK: Creating first intermediate JWK")

		intermediateKeyKidUUID, clearIntermediateKey, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
		if err != nil {
			log.Printf("DEBUG initializeFirstIntermediateJWK: GenerateJWEJWK failed: %v", err)

			return fmt.Errorf("failed to generate first intermediate JWK: %w", err)
		}

		log.Printf("DEBUG initializeFirstIntermediateJWK: Generated JWK with kid=%v", intermediateKeyKidUUID)

		var encryptedIntermediateKeyBytes []byte

		var rootKeyKidUUID *googleUuid.UUID

		err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			encryptedIntermediateKeyBytes, rootKeyKidUUID, err = rootKeysService.EncryptKey(sqlTransaction, clearIntermediateKey)
			if err != nil {
				log.Printf("DEBUG initializeFirstIntermediateJWK: EncryptKey failed: %v", err)

				return fmt.Errorf("failed to encrypt first intermediate JWK: %w", err)
			}

			log.Printf("DEBUG initializeFirstIntermediateJWK: Encrypted intermediate JWK, len=%d, rootKeyKid=%v", len(encryptedIntermediateKeyBytes), rootKeyKidUUID)

			firstEncryptedIntermediateKey := &cryptoutilOrmRepository.BarrierIntermediateKey{UUID: *intermediateKeyKidUUID, Encrypted: string(encryptedIntermediateKeyBytes), KEKUUID: *rootKeyKidUUID}

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

// EncryptKey encrypts a content key with the latest intermediate key.
func (i *IntermediateKeysService) EncryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearContentKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedIntermediateKeyLatest, err := sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted intermediate JWK latest from DB: %w", err)
	}

	intermediateKeyLatestKidUUID := encryptedIntermediateKeyLatest.GetUUID()

	decryptedIntermediateKeyLatest, err := i.rootKeysService.DecryptKey(sqlTransaction, []byte(encryptedIntermediateKeyLatest.GetEncrypted()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt intermediate JWK latest: %w", err)
	}

	_, encryptedContentKeyBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{decryptedIntermediateKeyLatest}, clearContentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content JWK with intermediate JWK: %w", err)
	}

	return encryptedContentKeyBytes, &intermediateKeyLatestKidUUID, nil
}

// DecryptKey decrypts a content key encrypted with an intermediate key.
func (i *IntermediateKeysService) DecryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentKeyBytes []byte) (joseJwk.Key, error) {
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

	decryptedIntermediateKey, err := i.rootKeysService.DecryptKey(sqlTransaction, []byte(encryptedIntermediateKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedContentKey, err := cryptoutilJose.DecryptKey([]joseJwk.Key{decryptedIntermediateKey}, encryptedContentKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content key: %w", err)
	}

	return decryptedContentKey, nil
}

// Shutdown releases all resources held by the IntermediateKeysService.
func (i *IntermediateKeysService) Shutdown() {
	i.telemetryService = nil
	i.ormRepository = nil
	i.jwkGenService = nil
	i.rootKeysService = nil
}
