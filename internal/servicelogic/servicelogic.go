package servicelogic

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

type KeyPoolService struct {
	ormRepository    *cryptoutilOrmRepository.RepositoryProvider
	serviceOrmMapper *serviceOrmMapper
	aes256Pool       *cryptoutilKeygen.KeyPool
	aes192Pool       *cryptoutilKeygen.KeyPool
	aes128Pool       *cryptoutilKeygen.KeyPool
	uuidV7Pool       *cryptoutilKeygen.KeyPool
}

func NewService(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider) *KeyPoolService {
	aes256Pool := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Service AES-256", 3, 1, cryptoutilKeygen.MaxKeys, cryptoutilKeygen.MaxTime, cryptoutilKeygen.GenerateAESKeyFunction(256))
	aes192Pool := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Service AES-192", 3, 1, cryptoutilKeygen.MaxKeys, cryptoutilKeygen.MaxTime, cryptoutilKeygen.GenerateAESKeyFunction(192))
	aes128Pool := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Service AES-128", 3, 1, cryptoutilKeygen.MaxKeys, cryptoutilKeygen.MaxTime, cryptoutilKeygen.GenerateAESKeyFunction(128))
	uuidV7Pool := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Service UUIDv7", 3, 1, cryptoutilKeygen.MaxKeys, cryptoutilKeygen.MaxTime, cryptoutilKeygen.GenerateUUIDv7Function())
	return &KeyPoolService{ormRepository: ormRepository, serviceOrmMapper: NewMapper(), aes256Pool: aes256Pool, aes192Pool: aes192Pool, aes128Pool: aes128Pool, uuidV7Pool: uuidV7Pool}
}

func (s *KeyPoolService) AddKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilServiceModel.KeyPoolCreate) (*cryptoutilServiceModel.KeyPool, error) {
	keyPoolID := s.uuidV7Pool.Get().Private.(googleUuid.UUID)
	repositoryKeyPoolToInsert := s.serviceOrmMapper.toOrmAddKeyPool(keyPoolID, openapiKeyPoolCreate)

	var insertedKeyPool *cryptoutilOrmRepository.KeyPool
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		err := sqlTransaction.AddKeyPool(repositoryKeyPoolToInsert)
		if err != nil {
			return fmt.Errorf("failed to add KeyPool: %w", err)
		}

		err = TransitionState(cryptoutilServiceModel.Creating, cryptoutilServiceModel.KeyPoolStatus(repositoryKeyPoolToInsert.KeyPoolStatus))
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

		err = sqlTransaction.AddKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateKeyPoolStatus(keyPoolID, cryptoutilOrmRepository.Active)
		if err != nil {
			return fmt.Errorf("failed to update KeyPoolStatus to active: %w", err)
		}

		insertedKeyPool, err = sqlTransaction.GetKeyPoolByKeyPoolID(keyPoolID)
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

func (s *KeyPoolService) GetKeyPoolByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID) (*cryptoutilServiceModel.KeyPool, error) {
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPoolByKeyPoolID(keyPoolID)
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

func (s *KeyPoolService) GetKeyPools(ctx context.Context, keyPoolQueryParams *cryptoutilServiceModel.KeyPoolsQueryParams) ([]cryptoutilServiceModel.KeyPool, error) {
	ormKeyPoolsQueryParams, err := s.serviceOrmMapper.toOrmGetKeyPoolsQueryParams(keyPoolQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Key Pools parameters: %w", err)
	}
	var repositoryKeyPools []cryptoutilOrmRepository.KeyPool
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
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

func (s *KeyPoolService) GenerateKeyInPoolKey(ctx context.Context, keyPoolID googleUuid.UUID, _ *cryptoutilServiceModel.KeyGenerate) (*cryptoutilServiceModel.Key, error) {
	var repositoryKey *cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		var err error
		repositoryKeyPool, err := sqlTransaction.GetKeyPoolByKeyPoolID(keyPoolID)
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

		err = sqlTransaction.AddKey(repositoryKey)
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

func (s *KeyPoolService) GetKeysByKeyPool(ctx context.Context, keyPoolID googleUuid.UUID, keyPoolKeysQueryParams *cryptoutilServiceModel.KeyPoolKeysQueryParams) ([]cryptoutilServiceModel.Key, error) {
	ormKeyPoolKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeyPoolKeysQueryParams(keyPoolKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Key Pool Keys parameters: %w", err)
	}
	var repositoryKeys []cryptoutilOrmRepository.Key
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		var err error
		repositoryKeys, err = sqlTransaction.FindKeysByKeyPoolID(keyPoolID, ormKeyPoolKeysQueryParams)
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

func (s *KeyPoolService) GetKeys(ctx context.Context, keysQueryParams *cryptoutilServiceModel.KeysQueryParams) ([]cryptoutilServiceModel.Key, error) {
	ormKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeysQueryParams(keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Keys parameters: %w", err)
	}
	var repositoryKeys []cryptoutilOrmRepository.Key
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
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

func (s *KeyPoolService) GetKeyByKeyPoolAndKeyID(ctx context.Context, keyPoolID googleUuid.UUID, keyID googleUuid.UUID) (*cryptoutilServiceModel.Key, error) {
	var repositoryKey *cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		var err error
		repositoryKey, err = sqlTransaction.GetKeyByKeyPoolIDAndKeyID(keyPoolID, keyID)
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

func (s *KeyPoolService) generateKeyInsert(keyPoolID googleUuid.UUID, keyPoolAlgorithm string) (*cryptoutilOrmRepository.Key, error) {
	keyID := s.uuidV7Pool.Get().Private.(googleUuid.UUID)

	repositoryKeyKeyMaterial, err := s.GenerateKeyMaterial(keyPoolAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key material: %w", err)
	}
	repositoryKeyGenerateDate := time.Now().UTC()

	return &cryptoutilOrmRepository.Key{
		KeyPoolID:       keyPoolID,
		KeyID:           keyID,
		KeyMaterial:     repositoryKeyKeyMaterial,
		KeyGenerateDate: &repositoryKeyGenerateDate,
	}, nil
}

func (s *KeyPoolService) GenerateKeyMaterial(algorithm string) ([]byte, error) {
	switch string(algorithm) {
	case "AES-256", "AES256":
		return s.aes256Pool.Get().Private.([]byte), nil
	case "AES-192", "AES192":
		return s.aes192Pool.Get().Private.([]byte), nil
	case "AES-128", "AES128":
		return s.aes128Pool.Get().Private.([]byte), nil
	default:
		return nil, fmt.Errorf("unsuppported algorithm")
	}
}
