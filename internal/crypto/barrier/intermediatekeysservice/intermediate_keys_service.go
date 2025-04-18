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

	var encryptedIntermediateKeyLatest *cryptoutilOrmRepository.BarrierIntermediateKey
	err := ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		encryptedIntermediateKeyLatest, err = sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK from DB
		return err
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get encrypted intermediate JWK latest from DB")
	}
	if encryptedIntermediateKeyLatest == nil {
		firstClearIntermediateKey, _, firstClearIntermediateKeyLatestKidUuid, err := cryptoutilJose.GenerateAesJWKFromPool(cryptoutilJose.AlgDIRECT, aes256KeyGenPool)
		if err != nil {
			return nil, fmt.Errorf("failed to generate intermediate JWK latest: %w", err)
		}
		clearRootJwkKidUuid, err := cryptoutilJose.ExtractKidUuid(rootKeysService.GetLatest())
		if err != nil {
			return nil, fmt.Errorf("failed to extract intermediate JWK latest kid UUID: %w", err)
		}
		_, firstEncryptedIntermediateKeyLatestBytes, err := cryptoutilJose.EncryptKey(rootKeysService.GetAll(), firstClearIntermediateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt root JWK: %w", err)
		}

		// put new, encrypted intermediate JWK latest in DB
		err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			return sqlTransaction.AddIntermediateKey(&cryptoutilOrmRepository.BarrierIntermediateKey{UUID: firstClearIntermediateKeyLatestKidUuid, Encrypted: string(firstEncryptedIntermediateKeyLatestBytes), KEKUUID: *clearRootJwkKidUuid})
		})
		if err != nil {
			return nil, fmt.Errorf("failed to store intermediate JWK latest kid: %w", err)
		}
	}

	return &IntermediateKeysService{telemetryService: telemetryService, ormRepository: ormRepository, rootKeysService: rootKeysService, aes256KeyGenPool: aes256KeyGenPool}, nil
}

func (i *IntermediateKeysService) EncryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearContentKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedIntermediateKeyLatest, err := sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted intermediate JWK latest from DB")
	}
	encryptedIntermediateKeyLatestKidUuid := encryptedIntermediateKeyLatest.GetUUID()
	decryptedIntermediateKeyLatest, err := cryptoutilJose.DecryptKey(i.rootKeysService.GetAll(), []byte(encryptedIntermediateKeyLatest.GetEncrypted()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt intermediate JWK latest: %w", err)
	}
	_, encryptedContentKeyBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{decryptedIntermediateKeyLatest}, clearContentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt content JWK with intermediate JWK")
	}
	return encryptedContentKeyBytes, &encryptedIntermediateKeyLatestKidUuid, nil
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
		return nil, fmt.Errorf("failed to get intermediate key")
	}
	decryptedIntermediateKey, err := cryptoutilJose.DecryptKey(i.rootKeysService.GetAll(), []byte(encryptedIntermediateKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key")
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedContentKey, err := cryptoutilJose.DecryptKey([]joseJwk.Key{decryptedIntermediateKey}, []byte(encryptedContentKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content key")
	}

	return decryptedContentKey, nil
}

func (i *IntermediateKeysService) Shutdown() {
	i.telemetryService = nil
	i.ormRepository = nil
	i.rootKeysService = nil
}
