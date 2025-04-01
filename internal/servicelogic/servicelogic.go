package servicelogic

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"cryptoutil/internal/crypto/keygen"
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
}

func NewService(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider) *KeyPoolService {
	aes256Pool := keygen.NewKeyPool(ctx, telemetryService, "Service AES-256", 3, 1, keygen.MaxKeys, keygen.MaxTime, keygen.GenerateAESKeyFunction(256))
	aes192Pool := keygen.NewKeyPool(ctx, telemetryService, "Service AES-192", 3, 1, keygen.MaxKeys, keygen.MaxTime, keygen.GenerateAESKeyFunction(192))
	aes128Pool := keygen.NewKeyPool(ctx, telemetryService, "Service AES-128", 3, 1, keygen.MaxKeys, keygen.MaxTime, keygen.GenerateAESKeyFunction(128))
	return &KeyPoolService{ormRepository: ormRepository, serviceOrmMapper: NewMapper(), aes256Pool: aes256Pool, aes192Pool: aes192Pool, aes128Pool: aes128Pool}
}

func (s *KeyPoolService) AddKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilServiceModel.KeyPoolCreate) (*cryptoutilServiceModel.KeyPool, error) {
	gormKeyPoolToInsert := s.serviceOrmMapper.toOrmKeyPoolInsert(openapiKeyPoolCreate)

	var insertedKeyPool *cryptoutilOrmRepository.KeyPool
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		err := sqlTransaction.AddKeyPool(gormKeyPoolToInsert)
		if err != nil {
			return fmt.Errorf("failed to add KeyPool: %w", err)
		}

		err = TransitionState(cryptoutilServiceModel.Creating, cryptoutilServiceModel.KeyPoolStatus(gormKeyPoolToInsert.KeyPoolStatus))
		if gormKeyPoolToInsert.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate {
			return fmt.Errorf("invalid KeyPoolStatus transition detected: %w", err)
		}

		if gormKeyPoolToInsert.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate {
			return nil // done because no Key needs to be generated (i.e. it will be imported later)
		}

		gormKey, err := s.generateKeyInsert(gormKeyPoolToInsert.KeyPoolID, string(gormKeyPoolToInsert.KeyPoolAlgorithm), 1)
		if err != nil {
			return fmt.Errorf("failed to generate key material: %w", err)
		}

		err = sqlTransaction.AddKey(gormKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateKeyPoolStatus(gormKeyPoolToInsert.KeyPoolID, cryptoutilOrmRepository.Active)
		if err != nil {
			return fmt.Errorf("failed to update KeyPoolStatus to active: %w", err)
		}

		insertedKeyPool, err = sqlTransaction.GetKeyPoolByID(gormKeyPoolToInsert.KeyPoolID)
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

func (s *KeyPoolService) ListKeyPools(ctx context.Context) ([]cryptoutilServiceModel.KeyPool, error) {
	var gormKeyPools []cryptoutilOrmRepository.KeyPool
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		var err error
		gormKeyPools, err = sqlTransaction.FindKeyPools()
		if err != nil {
			return fmt.Errorf("failed to list KeyPools: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list KeyPools: %w", err)
	}
	return s.serviceOrmMapper.toServiceKeyPools(gormKeyPools), nil
}

func (s *KeyPoolService) GenerateKeyInPoolKey(ctx context.Context, keyPoolID uuid.UUID, _ *cryptoutilServiceModel.KeyGenerate) (*cryptoutilServiceModel.Key, error) {
	var gormKey *cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		var err error
		gormKeyPool, err := sqlTransaction.GetKeyPoolByID(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
		}

		if gormKeyPool.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate && gormKeyPool.KeyPoolStatus != cryptoutilOrmRepository.Active {
			return fmt.Errorf("invalid KeyPoolStatus detected for generate Key: %w", err)
		}

		gormKeyPoolMaxID, err := sqlTransaction.ListMaxKeyIDByKeyPoolID(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get max ID by KeyPoolID: %w", err)
		}

		gormKey, err = s.generateKeyInsert(gormKeyPool.KeyPoolID, string(gormKeyPool.KeyPoolAlgorithm), gormKeyPoolMaxID+1)
		if err != nil {
			return fmt.Errorf("failed to generate key material: %w", err)
		}

		err = sqlTransaction.AddKey(gormKey)
		if err != nil {
			return fmt.Errorf("failed to insert Key: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	openapiPostKeypoolKeyPoolIDKeyResponseObject := *s.serviceOrmMapper.toServiceKey(gormKey)
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject, nil
}

func (s *KeyPoolService) ListKeysByKeyPool(ctx context.Context, keyPoolID uuid.UUID) ([]cryptoutilServiceModel.Key, error) {
	var gormKeys []cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.RepositoryTransaction) error {
		var err error
		gormKeys, err = sqlTransaction.ListKeysByKeyPoolID(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKeys(gormKeys), nil
}

func (s *KeyPoolService) generateKeyInsert(keyPoolID uuid.UUID, keyPoolAlgorithm string, keyPoolNextID int) (*cryptoutilOrmRepository.Key, error) {
	gormKeyKeyMaterial, err := s.GenerateKeyMaterial(keyPoolAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key material: %w", err)
	}
	gormKeyGenerateDate := time.Now().UTC()

	return &cryptoutilOrmRepository.Key{
		KeyPoolID:       keyPoolID,
		KeyID:           keyPoolNextID,
		KeyMaterial:     gormKeyKeyMaterial,
		KeyGenerateDate: &gormKeyGenerateDate,
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
