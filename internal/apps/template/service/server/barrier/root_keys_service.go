// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"context"
	"errors"
	"fmt"
	"log"

	// Repository interface used instead of OrmRepository.
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// RootKeysService manages root encryption keys at the top of the key hierarchy.
type RootKeysService struct {
	telemetryService  *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService     *cryptoutilSharedCryptoJose.JWKGenService
	repository        Repository
	unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
}

// NewRootKeysService creates a new RootKeysService with the specified dependencies.
func NewRootKeysService(telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, repository Repository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) (*RootKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if repository == nil {
		return nil, fmt.Errorf("repository must be non-nil")
	} else if unsealKeysService == nil {
		return nil, fmt.Errorf("unsealKeysService must be non-nil")
	}

	err := initializeFirstRootJWK(jwkGenService, repository, unsealKeysService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize first root JWK: %w", err)
	}

	return &RootKeysService{telemetryService: telemetryService, jwkGenService: jwkGenService, repository: repository, unsealKeysService: unsealKeysService}, nil
}

func initializeFirstRootJWK(jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, repository Repository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) error {
	var encryptedRootKeyLatest *RootKey

	var err error

	err = repository.WithTransaction(context.Background(), func(tx Transaction) error {
		encryptedRootKeyLatest, err = tx.GetRootKeyLatest() // encrypted root JWK from DB
		// NOTE: "no root key found" is EXPECTED on first run - don't treat as fatal error
		if err != nil && !errors.Is(err, ErrNoRootKeyFound) {
			return fmt.Errorf("failed to get root key latest: %w", err)
		}

		return nil
	})

	// DEBUG: Log error handling decision
	isNoRootKeyErr := errors.Is(err, ErrNoRootKeyFound)
	log.Printf("DEBUG initializeFirstRootJWK: err=%v, isNoRootKeyFound=%v, encryptedRootKeyLatest=%v", err, isNoRootKeyErr, encryptedRootKeyLatest)

	if err != nil && !isNoRootKeyErr {
		return fmt.Errorf("failed to get encrypted root JWK latest from DB: %w", err)
	}

	if encryptedRootKeyLatest == nil {
		log.Printf("DEBUG initializeFirstRootJWK: Creating first root JWK")

		rootKeyKidUUID, clearRootKey, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
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

		firstEncryptedRootKey := &RootKey{UUID: *rootKeyKidUUID, Encrypted: string(encryptedRootKeyBytes), KEKUUID: googleUuid.Nil}

		err = repository.WithTransaction(context.Background(), func(tx Transaction) error {
			return tx.AddRootKey(firstEncryptedRootKey)
		})
		if err != nil {
			log.Printf("DEBUG initializeFirstRootJWK: AddRootKey failed: %v", err)

			return fmt.Errorf("failed to encrypt and store first root JWK: %w", err)
		}

		log.Printf("DEBUG initializeFirstRootJWK: Successfully created first root JWK")
	}

	return nil
}

// EncryptKey encrypts an intermediate key using the latest root key.
func (i *RootKeysService) EncryptKey(tx Transaction, clearIntermediateKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedRootKeyLatest, err := tx.GetRootKeyLatest() // encrypted root JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted root JWK latest from DB: %w", err)
	}

	rootKeyLatestKidUUID := encryptedRootKeyLatest.UUID

	decryptedRootKeyLatest, err := i.unsealKeysService.DecryptKey([]byte(encryptedRootKeyLatest.Encrypted))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt root JWK latest: %w", err)
	}

	_, encryptedIntermediateKeyBytes, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{decryptedRootKeyLatest}, clearIntermediateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt intermediate JWK with root JWK: %w", err)
	}

	return encryptedIntermediateKeyBytes, &rootKeyLatestKidUUID, nil
}

// DecryptKey decrypts an intermediate key using the identified root key.
func (i *RootKeysService) DecryptKey(sqlTransaction Transaction, encryptedIntermediateKeyBytes []byte) (joseJwk.Key, error) {
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

	decryptedRootKey, err := i.unsealKeysService.DecryptKey([]byte(encryptedRootKey.Encrypted))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root key: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedIntermediateKey, err := cryptoutilSharedCryptoJose.DecryptKey([]joseJwk.Key{decryptedRootKey}, encryptedIntermediateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key: %w", err)
	}

	return decryptedIntermediateKey, nil
}

// Shutdown gracefully shuts down the RootKeysService.
func (i *RootKeysService) Shutdown() {
	i.unsealKeysService = nil
	i.repository = nil
	i.jwkGenService = nil
	i.telemetryService = nil
}
