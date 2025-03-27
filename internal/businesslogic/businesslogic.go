package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	cryptoutilCryptoKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilBusinessModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilRepositoryOrm "cryptoutil/internal/repository/orm"
)

type KeyPoolService struct {
	ormService       *cryptoutilRepositoryOrm.RepositoryOrm
	openapiOrmMapper *openapiOrmMapper
}

func NewService(dbService *cryptoutilRepositoryOrm.RepositoryOrm) *KeyPoolService {
	return &KeyPoolService{ormService: dbService, openapiOrmMapper: NewMapper()}
}

func (s *KeyPoolService) AddKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilBusinessModel.KeyPoolCreate) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	gormKeyPoolInsert := s.openapiOrmMapper.toGormKeyPoolInsert(openapiKeyPoolCreate)
	err := s.ormService.AddKeyPool(gormKeyPoolInsert)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyPoolResponseError(err)
	}
	return s.openapiOrmMapper.toOpenapiInsertKeyPoolResponseSuccess(gormKeyPoolInsert), nil
}

func (s *KeyPoolService) ListKeyPools(ctx context.Context) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	gormKeyPools, err := s.ormService.FindKeyPools()
	if err != nil {
		return s.openapiOrmMapper.toOpenapiSelectKeyPoolResponseError(err)
	}
	return s.openapiOrmMapper.toOpenapiSelectKeyPoolResponseSuccess(&gormKeyPools), nil
}

func (s *KeyPoolService) GenerateKeyInPoolKey(ctx context.Context, keyPoolID uuid.UUID, _ *cryptoutilBusinessModel.KeyGenerate) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	gormKeyPool, err := s.ormService.GetKeyPoolByID(keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeySelectKeyPoolResponseError(err)
	}

	if gormKeyPool.KeyPoolStatus != cryptoutilRepositoryOrm.PendingGenerate {
		return s.openapiOrmMapper.toOpenapiInsertKeyInvalidKeyPoolStatusResponseError()
	}

	gormKeyPoolMaxID, err := s.ormService.ListMaxKeyIDByKeyPoolID(keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyGenerateKeyMaterialResponseError(err)
	}

	gormKey, err := s.generateKeyInsert(gormKeyPool, gormKeyPoolMaxID+1)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyGenerateKeyMaterialResponseError(err)
	}

	err = s.ormService.AddKey(gormKey)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyResponseError(err)
	}

	return s.openapiOrmMapper.toOpenapiInsertKeySuccessResponseError(gormKey), nil
}

func (s *KeyPoolService) ListKeysByKeyPool(ctx context.Context, keyPoolID uuid.UUID) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	gormKeys, err := s.ormService.ListKeysByKeyPoolID(keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiGetKeyFindResponseError(err)
	}

	return s.openapiOrmMapper.toOpenapiGetKeyResponseSuccess(&gormKeys), nil
}

func (s *KeyPoolService) generateKeyInsert(gormKeyPool *cryptoutilRepositoryOrm.KeyPool, keyPoolNextID int) (*cryptoutilRepositoryOrm.Key, error) {
	gormKeyKeyMaterial, err := cryptoutilCryptoKeygen.GenerateKeyMaterial(string(gormKeyPool.KeyPoolAlgorithm))
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key material: %w", err)
	}
	gormKeyGenerateDate := time.Now().UTC()

	return &cryptoutilRepositoryOrm.Key{
		KeyPoolID:       gormKeyPool.KeyPoolID,
		KeyID:           keyPoolNextID,
		KeyMaterial:     gormKeyKeyMaterial,
		KeyGenerateDate: &gormKeyGenerateDate,
	}, nil
}
