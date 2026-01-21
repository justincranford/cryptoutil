// Copyright (c) 2025 Justin Cranford
//
//

package rootkeysservice

import (
	"context"
	"errors"
	"fmt"
	"log"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"
)

// RootKeysService manages root encryption keys in the barrier hierarchy.
type RootKeysService struct {
	telemetryService  *cryptoutilTelemetry.TelemetryService
	jwkGenService     *cryptoutilJose.JWKGenService
	ormRepository     *cryptoutilOrmRepository.OrmRepository
	unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
}

// NewRootKeysService creates a new RootKeysService with the specified dependencies.
func NewRootKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) (*RootKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if unsealKeysService == nil {
		return nil, fmt.Errorf("unsealKeysService must be non-nil")
	}

	err := initializeFirstRootJWK(jwkGenService, ormRepository, unsealKeysService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize first root JWK: %w", err)
	}

	return &RootKeysService{telemetryService: telemetryService, jwkGenService: jwkGenService, ormRepository: ormRepository, unsealKeysService: unsealKeysService}, nil
}

func initializeFirstRootJWK(jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) error {
	var encryptedRootKeyLatest *cryptoutilOrmRepository.BarrierRootKey

	var err error

	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedRootKeyLatest, err = sqlTransaction.GetRootKeyLatest() // encrypted root JWK from DB
		// NOTE: "record not found" is EXPECTED on first run - don't treat as fatal error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get root key latest: %w", err)
		}

		return nil
	})

	// DEBUG: Log error handling decision
	isRecordNotFoundErr := errors.Is(err, gorm.ErrRecordNotFound)
	log.Printf("DEBUG initializeFirstRootJWK: err=%v, isRecordNotFound=%v, encryptedRootKeyLatest=%v", err, isRecordNotFoundErr, encryptedRootKeyLatest)

	if err != nil && !isRecordNotFoundErr {
		return fmt.Errorf("failed to get encrypted root JWK latest from DB: %w", err)
	}

	if encryptedRootKeyLatest == nil {
		log.Printf("DEBUG initializeFirstRootJWK: Creating first root JWK")

		rootKeyKidUUID, clearRootKey, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
		if err != nil {
			log.Printf("DEBUG initializeFirstRootJWK: GenerateJWEJWK failed: %v", err)

			return fmt.Errorf("failed to generate first root JWK latest: %w", err)
		}

		log.Printf("DEBUG initializeFirstRootJWK: Generated JWK with kid=%v", rootKeyKidUUID)

		encryptedRootKeyBytes, err := unsealKeysService.EncryptKey(clearRootKey)
		if err != nil {
			log.Printf("DEBUG initializeFirstRootJWK: EncryptKey failed: %v", err)

			return fmt.Errorf("failed to encrypt first root JWK: %w", err)
		}

		log.Printf("DEBUG initializeFirstRootJWK: Encrypted root JWK, len=%d", len(encryptedRootKeyBytes))

		firstEncryptedRootKey := &cryptoutilOrmRepository.BarrierRootKey{UUID: *rootKeyKidUUID, Encrypted: string(encryptedRootKeyBytes), KEKUUID: googleUuid.Nil}

		err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			return sqlTransaction.AddRootKey(firstEncryptedRootKey)
		})
		if err != nil {
			log.Printf("DEBUG initializeFirstRootJWK: AddRootKey failed: %v", err)

			return fmt.Errorf("failed to encrypt and store first root JWK: %w", err)
		}

		log.Printf("DEBUG initializeFirstRootJWK: Successfully created first root JWK")
	}

	return nil
}

// EncryptKey encrypts an intermediate key with the latest root key.
func (i *RootKeysService) EncryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearIntermediateKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedRootKeyLatest, err := sqlTransaction.GetRootKeyLatest() // encrypted root JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted root JWK latest from DB: %w", err)
	}

	rootKeyLatestKidUUID := encryptedRootKeyLatest.GetUUID()

	decryptedRootKeyLatest, err := i.unsealKeysService.DecryptKey([]byte(encryptedRootKeyLatest.GetEncrypted()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt root JWK latest: %w", err)
	}

	_, encryptedIntermediateKeyBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{decryptedRootKeyLatest}, clearIntermediateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt intermediate JWK with root JWK: %w", err)
	}

	return encryptedIntermediateKeyBytes, &rootKeyLatestKidUUID, nil
}

// DecryptKey decrypts an intermediate key encrypted with a root key.
func (i *RootKeysService) DecryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedIntermediateKeyBytes []byte) (joseJwk.Key, error) {
	encryptedIntermediateKey, err := joseJwe.Parse(encryptedIntermediateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted intermediate key message: %w", err)
	}

	var rootKeyKidUUIDString string

	err = encryptedIntermediateKey.ProtectedHeaders().Get(joseJwk.KeyIDKey, &rootKeyKidUUIDString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted intermediate key message kid UUID: %w", err)
	}

	rootKeyKidUUID, err := googleUuid.Parse(rootKeyKidUUIDString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	encryptedRootKey, err := sqlTransaction.GetRootKey(&rootKeyKidUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get root key: %w", err)
	}

	decryptedRootKey, err := i.unsealKeysService.DecryptKey([]byte(encryptedRootKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root key: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedIntermediateKey, err := cryptoutilJose.DecryptKey([]joseJwk.Key{decryptedRootKey}, encryptedIntermediateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key: %w", err)
	}

	return decryptedIntermediateKey, nil
}

// Shutdown releases all resources held by the RootKeysService.
func (i *RootKeysService) Shutdown() {
	i.unsealKeysService = nil
	i.ormRepository = nil
	i.jwkGenService = nil
	i.telemetryService = nil
}
