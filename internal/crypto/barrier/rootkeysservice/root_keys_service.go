package rootkeysservice

import (
	"context"
	"errors"
	"fmt"

	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"
)

type RootKeysService struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	unsealRepository cryptoutilUnsealRepository.UnsealRepository
	aes256KeyGenPool *cryptoutilKeygen.KeyGenPool
}

func (u *RootKeysService) Shutdown() {
	u.aes256KeyGenPool = nil
	u.unsealRepository = nil
	u.ormRepository = nil
	u.telemetryService = nil
}

func NewRootKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealRepository cryptoutilUnsealRepository.UnsealRepository, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) (*RootKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if unsealRepository == nil {
		return nil, fmt.Errorf("unsealRepository must be non-nil")
	} else if aes256KeyGenPool == nil {
		return nil, fmt.Errorf("aes256KeyGenPool must be non-nil")
	}

	unsealJwks := unsealRepository.UnsealJwks() // unseal keys from unseal repository
	if len(unsealJwks) == 0 {
		return nil, fmt.Errorf("no unseal JWKs")
	}

	var encryptedRootKeyLatest *cryptoutilOrmRepository.BarrierRootKey
	err := ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		encryptedRootKeyLatest, err = sqlTransaction.GetRootKeyLatest() // encrypted intermediate JWK from DB
		return err
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get encrypted root JWK latest from DB")
	}
	if encryptedRootKeyLatest == nil {
		firstClearRootKey, _, firstClearRootKeyLatestKidUuid, err := cryptoutilJose.GenerateAesJWKFromPool(cryptoutilJose.AlgA256GCMKW, aes256KeyGenPool)
		if err != nil {
			return nil, fmt.Errorf("failed to generate root JWK latest: %w", err)
		}
		_, firstEncryptedRootKeyLatestBytes, err := cryptoutilJose.EncryptKey(unsealRepository.UnsealJwks(), firstClearRootKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt root JWK: %w", err)
		}

		// put new, encrypted root JWK latest in DB
		err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			return sqlTransaction.AddRootKey(&cryptoutilOrmRepository.BarrierRootKey{UUID: firstClearRootKeyLatestKidUuid, Encrypted: string(firstEncryptedRootKeyLatestBytes), KEKUUID: googleUuid.Nil})
		})
		if err != nil {
			return nil, fmt.Errorf("failed to store encrypted root JWK: %w", err)
		}
	}

	return &RootKeysService{telemetryService: telemetryService, ormRepository: ormRepository, unsealRepository: unsealRepository, aes256KeyGenPool: aes256KeyGenPool}, nil
}

func (i *RootKeysService) EncryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearIntermediateKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedRootKeyLatest, err := sqlTransaction.GetRootKeyLatest() // encrypted root JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted root JWK latest from DB")
	}
	encryptedRootKeyLatestKidUuid := encryptedRootKeyLatest.GetUUID()
	decryptedRootKeyLatest, err := cryptoutilJose.DecryptKey(i.unsealRepository.UnsealJwks(), []byte(encryptedRootKeyLatest.GetEncrypted()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt root JWK latest: %w", err)
	}
	_, encryptedIntermediateKeyBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{decryptedRootKeyLatest}, clearIntermediateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt intermediate JWK with root JWK")
	}
	return encryptedIntermediateKeyBytes, &encryptedRootKeyLatestKidUuid, nil
}

func (i *RootKeysService) DecryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedIntermediateKeyBytes []byte) (joseJwk.Key, error) {
	encryptedIntermediateKey, err := joseJwe.Parse(encryptedIntermediateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted intermediate key message: %w", err)
	}
	var rootKeyKidUuidString string
	err = encryptedIntermediateKey.ProtectedHeaders().Get(joseJwk.KeyIDKey, &rootKeyKidUuidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted intermediate key message kid UUID: %w", err)
	}
	rootKeyKidUuid, err := googleUuid.Parse(rootKeyKidUuidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	encryptedRootKey, err := sqlTransaction.GetRootKey(rootKeyKidUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get root key")
	}
	decryptedRootKey, err := cryptoutilJose.DecryptKey(i.unsealRepository.UnsealJwks(), []byte(encryptedRootKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root key")
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedIntermediateKey, err := cryptoutilJose.DecryptKey([]joseJwk.Key{decryptedRootKey}, []byte(encryptedIntermediateKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key")
	}

	return decryptedIntermediateKey, nil
}
