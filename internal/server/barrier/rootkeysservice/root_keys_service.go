package rootkeysservice

import (
	"context"
	"errors"
	"fmt"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"
)

type RootKeysService struct {
	telemetryService  *cryptoutilTelemetry.TelemetryService
	jwkGenService     *cryptoutilJose.JwkGenService
	ormRepository     *cryptoutilOrmRepository.OrmRepository
	unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
}

func NewRootKeysService(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JwkGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) (*RootKeysService, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if unsealKeysService == nil {
		return nil, fmt.Errorf("unsealKeysService must be non-nil")
	}
	err := initializeFirstRootJwk(jwkGenService, ormRepository, unsealKeysService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize first root JWK: %w", err)
	}
	return &RootKeysService{telemetryService: telemetryService, jwkGenService: jwkGenService, ormRepository: ormRepository, unsealKeysService: unsealKeysService}, nil
}

func initializeFirstRootJwk(jwkGenService *cryptoutilJose.JwkGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) error {
	var encryptedRootKeyLatest *cryptoutilOrmRepository.BarrierRootKey
	var err error
	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedRootKeyLatest, err = sqlTransaction.GetRootKeyLatest() // encrypted root JWK from DB
		return err
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get encrypted root JWK latest from DB: %w", err)
	}
	if encryptedRootKeyLatest == nil {
		rootKeyKidUuid, clearRootKey, _, _, _, err := jwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
		if err != nil {
			return fmt.Errorf("failed to generate first root JWK latest: %w", err)
		}
		encryptedRootKeyBytes, err := unsealKeysService.EncryptKey(clearRootKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt first root JWK: %w", err)
		}
		firstEncryptedRootKey := &cryptoutilOrmRepository.BarrierRootKey{UUID: *rootKeyKidUuid, Encrypted: string(encryptedRootKeyBytes), KEKUUID: googleUuid.Nil}
		err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			return sqlTransaction.AddRootKey(firstEncryptedRootKey)
		})
		if err != nil {
			return fmt.Errorf("failed to encrypt and store first root JWK: %w", err)
		}
	}
	return nil
}

func (i *RootKeysService) EncryptKey(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearIntermediateKey joseJwk.Key) ([]byte, *googleUuid.UUID, error) {
	encryptedRootKeyLatest, err := sqlTransaction.GetRootKeyLatest() // encrypted root JWK latest from DB
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get encrypted root JWK latest from DB: %w", err)
	}
	rootKeyLatestKidUuid := encryptedRootKeyLatest.GetUUID()
	decryptedRootKeyLatest, err := i.unsealKeysService.DecryptKey([]byte(encryptedRootKeyLatest.GetEncrypted()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt root JWK latest: %w", err)
	}

	_, encryptedIntermediateKeyBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{decryptedRootKeyLatest}, clearIntermediateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt intermediate JWK with root JWK: %w", err)
	}
	return encryptedIntermediateKeyBytes, &rootKeyLatestKidUuid, nil
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
	encryptedRootKey, err := sqlTransaction.GetRootKey(&rootKeyKidUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get root key: %w", err)
	}
	decryptedRootKey, err := i.unsealKeysService.DecryptKey([]byte(encryptedRootKey.GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root key: %w", err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	decryptedIntermediateKey, err := cryptoutilJose.DecryptKey([]joseJwk.Key{decryptedRootKey}, []byte(encryptedIntermediateKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt intermediate key: %w", err)
	}

	return decryptedIntermediateKey, nil
}

func (u *RootKeysService) Shutdown() {
	u.unsealKeysService = nil
	u.ormRepository = nil
	u.jwkGenService = nil
	u.telemetryService = nil
}
