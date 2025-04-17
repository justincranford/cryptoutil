package intermediatekeysservice

import (
	"context"
	"errors"
	"fmt"

	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type IntermediateKeysService struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	rootKeysService  *cryptoutilRootKeysService.RootKeysService
}

func NewIntermediateKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService) (*IntermediateKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	}

	var encryptedIntermediateKeyLatest *cryptoutilOrmRepository.BarrierIntermediateKey
	err := ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		encryptedIntermediateKeyLatest, err = sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK from DB
		return err
	})
	// TODO handle no row gracefully
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get encrypted intermediate JWK latest from DB")
	}
	if encryptedIntermediateKeyLatest == nil {
		firstClearIntermediateKey, _, firstClearIntermediateKeyLatestKidUuid, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
		if err != nil {
			return nil, fmt.Errorf("failed to generate intermediate JWK latest: %w", err)
		}
		clearRootJwkKidUuid, err := cryptoutilJose.ExtractKidUuid(rootKeysService.GetLatest())
		if err != nil {
			return nil, fmt.Errorf("failed to extract intermediate JWK latest kid UUID: %w", err)
		}
		_, firstEncryptedIntermediateKeyLatestJweMessageBytes, err := cryptoutilJose.EncryptKey(rootKeysService.GetAll(), firstClearIntermediateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt root JWK: %w", err)
		}

		// put new, encrypted intermediate JWK latest in DB
		err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			return sqlTransaction.AddIntermediateKey(&cryptoutilOrmRepository.BarrierIntermediateKey{UUID: firstClearIntermediateKeyLatestKidUuid, Encrypted: string(firstEncryptedIntermediateKeyLatestJweMessageBytes), KEKUUID: *clearRootJwkKidUuid})
		})
		if err != nil {
			return nil, fmt.Errorf("failed to store intermediate JWK latest kid: %w", err)
		}
	}

	return &IntermediateKeysService{telemetryService: telemetryService, ormRepository: ormRepository, rootKeysService: rootKeysService}, nil
}

func (i *IntermediateKeysService) GetLatest(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) (joseJwk.Key, error) {
	encryptedIntermediateKeyLatest, err := sqlTransaction.GetIntermediateKeyLatest() // encrypted intermediate JWK latest from DB
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted intermediate JWK latest from DB")
	}
	decryptedIntermediateKeyLatest, err := cryptoutilJose.DecryptKey(i.rootKeysService.GetAll(), []byte(encryptedIntermediateKeyLatest.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate JWK latest: %w", err)
	}
	return decryptedIntermediateKeyLatest, nil
}

func (i *IntermediateKeysService) Get(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, uuid googleUuid.UUID) (joseJwk.Key, error) {
	encryptedIntermediateKey, err := sqlTransaction.GetIntermediateKey(uuid) // encrypted intermediate JWK from DB
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted intermediate JWK from DB")
	}
	decryptedIntermediateKey, err := cryptoutilJose.DecryptKey(i.rootKeysService.GetAll(), []byte(encryptedIntermediateKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate JWK for uuid %s: %w", uuid.String(), err)
	}
	return decryptedIntermediateKey, nil
}

func (i *IntermediateKeysService) GetAll(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) ([]joseJwk.Key, error) {
	encryptedIntermediateKeys, err := sqlTransaction.GetIntermediateKeys() // encrypted intermediate JWK from DB
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted intermediate JWK from DB")
	}
	decryptedIntermediateKeys := make([]joseJwk.Key, 0, len(encryptedIntermediateKeys))
	for index, encryptedIntermediateKey := range encryptedIntermediateKeys {
		decryptedIntermediateKey, err := cryptoutilJose.DecryptKey(i.rootKeysService.GetAll(), []byte(encryptedIntermediateKey.GetEncrypted()))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt intermediate JWK %d: %w", index, err)
		}
		decryptedIntermediateKeys = append(decryptedIntermediateKeys, decryptedIntermediateKey)
	}
	return decryptedIntermediateKeys, nil
}

func (i *IntermediateKeysService) Shutdown() {
	i.telemetryService = nil
	i.ormRepository = nil
	i.rootKeysService = nil
}
