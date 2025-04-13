package businesslogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	cryptoutilBarrierService "cryptoutil/internal/crypto/barrier/barrierservice"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
)

type BusinessLogicService struct {
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	serviceOrmMapper *serviceOrmMapper
	aes256KeyGenPool *cryptoutilKeygen.KeyGenPool
	aes192KeyGenPool *cryptoutilKeygen.KeyGenPool
	aes128KeyGenPool *cryptoutilKeygen.KeyGenPool
	uuidV7KeyGenPool *cryptoutilKeygen.KeyGenPool
	barrierService   *cryptoutilBarrierService.BarrierService
}

func NewBusinessLogicService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, barrierService *cryptoutilBarrierService.BarrierService) (*BusinessLogicService, error) {
	aes256KeyGenPoolConfig, err1 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-256", 2, 2, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	aes192KeyGenPoolConfig, err2 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-192", 2, 2, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(192))
	aes128KeyGenPoolConfig, err3 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-128", 2, 2, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(128))
	uuidV7KeyGenPoolConfig, err4 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service UUIDv7", 2, 2, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateUUIDv7Function())
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2, err3, err4))
	}

	aes256KeyGenPool, err1 := cryptoutilKeygen.NewGenKeyPool(aes256KeyGenPoolConfig)
	aes192KeyGenPool, err2 := cryptoutilKeygen.NewGenKeyPool(aes192KeyGenPoolConfig)
	aes128KeyGenPool, err3 := cryptoutilKeygen.NewGenKeyPool(aes128KeyGenPoolConfig)
	uuidV7KeyGenPool, err4 := cryptoutilKeygen.NewGenKeyPool(uuidV7KeyGenPoolConfig)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4))
	}

	return &BusinessLogicService{ormRepository: ormRepository, serviceOrmMapper: NewMapper(), aes256KeyGenPool: aes256KeyGenPool, aes192KeyGenPool: aes192KeyGenPool, aes128KeyGenPool: aes128KeyGenPool, uuidV7KeyGenPool: uuidV7KeyGenPool, barrierService: barrierService}, nil
}

func (s *BusinessLogicService) AddKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilBusinessLogicModel.KeyPoolCreate) (*cryptoutilBusinessLogicModel.KeyPool, error) {
	keyPoolID := s.uuidV7KeyGenPool.Get().Private.(googleUuid.UUID)
	repositoryKeyPoolToInsert := s.serviceOrmMapper.toOrmAddKeyPool(keyPoolID, openapiKeyPoolCreate)

	var insertedKeyPool *cryptoutilOrmRepository.KeyPool
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		err := sqlTransaction.AddKeyPool(repositoryKeyPoolToInsert)
		if err != nil {
			return fmt.Errorf("failed to add KeyPool: %w", err)
		}

		err = TransitionState(cryptoutilBusinessLogicModel.Creating, cryptoutilBusinessLogicModel.KeyPoolStatus(repositoryKeyPoolToInsert.KeyPoolStatus))
		if repositoryKeyPoolToInsert.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate {
			return fmt.Errorf("invalid KeyPoolStatus transition detected: %w", err)
		}

		if repositoryKeyPoolToInsert.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate {
			return nil // import first key manually later
		}

		// generate first key automatically now
		repositoryKey, err := s.generateKeyInsert(keyPoolID, string(repositoryKeyPoolToInsert.KeyPoolAlgorithm))
		if err != nil {
			return fmt.Errorf("failed to generate key material: %w", err)
		}

		err = sqlTransaction.AddKeyPoolKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateKeyPoolStatus(keyPoolID, cryptoutilOrmRepository.Active)
		if err != nil {
			return fmt.Errorf("failed to update KeyPoolStatus to active: %w", err)
		}

		insertedKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get updated KeyPool from DB: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add key pool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKeyPool(insertedKeyPool), nil
}

var repositoryKeyPool *cryptoutilOrmRepository.KeyPool

func (s *BusinessLogicService) GetKeyPoolByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID) (*cryptoutilBusinessLogicModel.KeyPool, error) {
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get KeyPool: %w", err)
	}
	return s.serviceOrmMapper.toServiceKeyPool(repositoryKeyPool), nil
}

func (s *BusinessLogicService) GetKeyPools(ctx context.Context, keyPoolQueryParams *cryptoutilBusinessLogicModel.KeyPoolsQueryParams) ([]cryptoutilBusinessLogicModel.KeyPool, error) {
	ormKeyPoolsQueryParams, err := s.serviceOrmMapper.toOrmGetKeyPoolsQueryParams(keyPoolQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Key Pools parameters: %w", err)
	}
	var repositoryKeyPools []cryptoutilOrmRepository.KeyPool
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPools, err = sqlTransaction.GetKeyPools(ormKeyPoolsQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list KeyPools: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list KeyPools: %w", err)
	}
	return s.serviceOrmMapper.toServiceKeyPools(repositoryKeyPools), nil
}

func (s *BusinessLogicService) GenerateKeyInPoolKey(ctx context.Context, keyPoolID googleUuid.UUID, _ *cryptoutilBusinessLogicModel.KeyGenerate) (*cryptoutilBusinessLogicModel.Key, error) {
	var repositoryKey *cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err := sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
		}

		if repositoryKeyPool.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate && repositoryKeyPool.KeyPoolStatus != cryptoutilOrmRepository.Active {
			return fmt.Errorf("invalid KeyPoolStatus detected for generate Key: %w", err)
		}

		repositoryKey, err = s.generateKeyInsert(repositoryKeyPool.KeyPoolID, string(repositoryKeyPool.KeyPoolAlgorithm))
		if err != nil {
			return fmt.Errorf("failed to generate key material: %w", err)
		}

		err = sqlTransaction.AddKeyPoolKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to insert Key: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	openapiPostKeypoolKeyPoolIDKeyResponseObject := *s.serviceOrmMapper.toServiceKey(repositoryKey)
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject, nil
}

func (s *BusinessLogicService) GetKeysByKeyPool(ctx context.Context, keyPoolID googleUuid.UUID, keyPoolKeysQueryParams *cryptoutilBusinessLogicModel.KeyPoolKeysQueryParams) ([]cryptoutilBusinessLogicModel.Key, error) {
	ormKeyPoolKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeyPoolKeysQueryParams(keyPoolKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Key Pool Keys parameters: %w", err)
	}
	var repositoryKeys []cryptoutilOrmRepository.Key
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeys, err = sqlTransaction.GetKeyPoolKeys(keyPoolID, ormKeyPoolKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKeys(repositoryKeys), nil
}

func (s *BusinessLogicService) GetKeys(ctx context.Context, keysQueryParams *cryptoutilBusinessLogicModel.KeysQueryParams) ([]cryptoutilBusinessLogicModel.Key, error) {
	ormKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeysQueryParams(keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Keys parameters: %w", err)
	}
	var repositoryKeys []cryptoutilOrmRepository.Key
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeys, err = sqlTransaction.GetKeys(ormKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKeys(repositoryKeys), nil
}

func (s *BusinessLogicService) GetKeyByKeyPoolAndKeyID(ctx context.Context, keyPoolID googleUuid.UUID, keyID googleUuid.UUID) (*cryptoutilBusinessLogicModel.Key, error) {
	var repositoryKey *cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKey, err = sqlTransaction.GetKeyPoolKey(keyPoolID, keyID)
		if err != nil {
			return fmt.Errorf("failed to get Key by KeyPoolID and KeyID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKey(repositoryKey), nil
}

func (s *BusinessLogicService) generateKeyInsert(keyPoolID googleUuid.UUID, keyPoolAlgorithm string) (*cryptoutilOrmRepository.Key, error) {
	keyID := s.uuidV7KeyGenPool.Get().Private.(googleUuid.UUID)

	// TODO Generate JWK instead of []byte
	clearKeyMaterial, err := s.GenerateKeyMaterial(keyPoolAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key material: %w", err)
	}
	repositoryKeyGenerateDate := time.Now().UTC()

	encryptedKeyMaterial, err := s.barrierService.EncryptContent(clearKeyMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt Key material: %w", err)
	}

	return &cryptoutilOrmRepository.Key{
		KeyPoolID:       keyPoolID,
		KeyID:           keyID,
		KeyMaterial:     encryptedKeyMaterial,
		KeyGenerateDate: &repositoryKeyGenerateDate,
	}, nil
}

func (s *BusinessLogicService) GenerateKeyMaterial(algorithm string) ([]byte, error) {
	switch string(algorithm) {
	case "AES-256", "AES256":
		return s.aes256KeyGenPool.Get().Private.([]byte), nil
	case "AES-192", "AES192":
		return s.aes192KeyGenPool.Get().Private.([]byte), nil
	case "AES-128", "AES128":
		return s.aes128KeyGenPool.Get().Private.([]byte), nil
	default:
		return nil, fmt.Errorf("unsuppported algorithm")
	}
}
