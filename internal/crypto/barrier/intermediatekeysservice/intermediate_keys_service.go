package intermediatekeysservice

import (
	"context"
	"errors"
	"fmt"

	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type IntermediateKeysService struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	aes256KeyGenPool *cryptoutilKeygen.KeyGenPool
	rootKeysService  *cryptoutilRootKeysService.RootKeysService
}

func NewIntermediateKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) (*IntermediateKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if rootKeysService == nil {
		return nil, fmt.Errorf("rootKeysService must be non-nil")
	} else if aes256KeyGenPool == nil {
		return nil, fmt.Errorf("aes256KeyGenPool must be non-nil")
	}
	err := initializeFirstIntermediateJwk(ormRepository, rootKeysService, aes256KeyGenPool)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize first intermediate JWK: %w", err)
	}
	return &IntermediateKeysService{telemetryService: telemetryService, ormRepository: ormRepository, rootKeysService: rootKeysService, aes256KeyGenPool: aes256KeyGenPool}, nil
}

func initializeFirstIntermediateJwk(ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) error {
	var encryptedIntermediateKeyLatest *cryptoutilOrmRepository.BarrierIntermediateKey
	var err error
	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedIntermediateKeyLatest, err = sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK from DB
		return err
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get encrypted intermediate JWK latest from DB: %w", err)
	}
	if encryptedIntermediateKeyLatest == nil {
		intermediateKeyKidUuid, clearIntermediateKey, _, err := cryptoutilJose.GenerateAesJWKFromPool(&cryptoutilJose.AlgKekDIRECT, aes256KeyGenPool)
		if err != nil {
			return fmt.Errorf("failed to generate first intermediate JWK: %w", err)
		}
		var encryptedIntermediateKeyBytes []byte
		var rootKeyKidUuid *googleUuid.UUID
		err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			encryptedIntermediateKeyBytes, rootKeyKidUuid, err = rootKeysService.EncryptKey(sqlTransaction, clearIntermediateKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt first intermediate JWK: %w", err)
			}
			firstEncryptedIntermediateKey := &cryptoutilOrmRepository.BarrierIntermediateKey{UUID: *intermediateKeyKidUuid, Encrypted: string(encryptedIntermediateKeyBytes), KEKUUID: *rootKeyKidUuid}
			err = sqlTransaction.AddIntermediateKey(firstEncryptedIntermediateKey)
			if err != nil {
				return fmt.Errorf("failed to store first intermediate JWK: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to encrypt and store first intermediate first JWK: %w", err)
		}
	}
	return nil
}

func (i *IntermediateKeysService) EncryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearContentKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedIntermediateKeyLatest, err := sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted intermediate JWK latest from DB: %w", err)
	}
	intermediateKeyLatestKidUuid := encryptedIntermediateKeyLatest.GetUUID()
	decryptedIntermediateKeyLatest, err := i.rootKeysService.DecryptKey(sqlTransaction, []byte(encryptedIntermediateKeyLatest.GetEncrypted()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt intermediate JWK latest: %w", err)
	}
	_, encryptedContentKeyBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{decryptedIntermediateKeyLatest}, clearContentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content JWK with intermediate JWK: %w", err)
	}
	return encryptedContentKeyBytes, &intermediateKeyLatestKidUuid, nil
}

func (i *IntermediateKeysService) DecryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentKeyBytes []byte) (joseJwk.Key, error) {
	encryptedContentKey, err := joseJwe.Parse(encryptedContentKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted content key message: %w", err)
	}
	var intermediateKeyKidUuidString string
	err = encryptedContentKey.ProtectedHeaders().Get(joseJwk.KeyIDKey, &intermediateKeyKidUuidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted content key message kid UUID: %w", err)
	}
	intermediateKeyKidUuid, err := googleUuid.Parse(intermediateKeyKidUuidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	encryptedIntermediateKey, err := sqlTransaction.GetIntermediateKey(intermediateKeyKidUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get intermediate key: %w", err)
	}
	decryptedIntermediateKey, err := i.rootKeysService.DecryptKey(sqlTransaction, []byte(encryptedIntermediateKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedContentKey, err := cryptoutilJose.DecryptKey([]joseJwk.Key{decryptedIntermediateKey}, []byte(encryptedContentKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content key: %w", err)
	}

	return decryptedContentKey, nil
}

func (i *IntermediateKeysService) Shutdown() {
	i.telemetryService = nil
	i.ormRepository = nil
	i.rootKeysService = nil
}
