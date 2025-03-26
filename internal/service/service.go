package service

import (
	"context"
	"fmt"
	"time"

	cryptoutilModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/openapi/server"
	ormService "cryptoutil/internal/orm"

	"github.com/google/uuid"
)

type KeyPoolService struct {
	ormService       *ormService.Service
	openapiOrmMapper *OpenapiOrmMapper
}

func NewService(dbService *ormService.Service) *KeyPoolService {
	return &KeyPoolService{ormService: dbService, openapiOrmMapper: NewMapper()}
}

func (s *KeyPoolService) PostKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilModel.KeyPoolCreate) (cryptoutilServer.PostKeypoolResponseObject, error) {
	gormKeyPoolInsert := s.openapiOrmMapper.toGormKeyPoolInsert(openapiKeyPoolCreate)
	gormCreateKeyResult := s.ormService.GormDB.Create(gormKeyPoolInsert)
	if gormCreateKeyResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeyPoolError(gormCreateKeyResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseInsertKeyPoolSuccess(gormKeyPoolInsert), nil
}

func (s *KeyPoolService) GetKeyPool(ctx context.Context) (cryptoutilServer.GetKeypoolResponseObject, error) {
	var gormKeyPools []ormService.KeyPool
	gormFindKeyPoolsResult := s.ormService.GormDB.Find(&gormKeyPools)
	if gormFindKeyPoolsResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseSelectKeyPoolError(gormFindKeyPoolsResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseSelectKeyPoolSuccess(&gormKeyPools), nil
}

func (s *KeyPoolService) PostKeyPoolKeyPoolIDKey(ctx context.Context, keyPoolID *string, _ *cryptoutilModel.KeyGenerate) (cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	_, err := uuid.Parse(*keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeySelectKeyPoolError(err)
	}

	var gormKeyPool ormService.KeyPool
	gormSelectKeyPoolResult := s.ormService.GormDB.First(&gormKeyPool, "key_pool_id=?", *keyPoolID)
	if gormSelectKeyPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeySelectKeyPoolError(gormSelectKeyPoolResult.Error)
	}

	if gormKeyPool.KeyPoolStatus != ormService.PendingGenerate {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeyInvalidKeyPoolStatus()
	}

	var gormKeyPoolMaxID int // COALESCE clause returns 0 if no Keys found for Key Pool
	s.ormService.GormDB.Model(&ormService.Key{}).Where("key_pool_id=?", *keyPoolID).Select("COALESCE(MAX(key_id), 0)").Scan(&gormKeyPoolMaxID)

	gormKey, err := s.generateKeyInsert(&gormKeyPool, gormKeyPoolMaxID+1)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeyGenerateKeyMaterialError(err)
	}

	gormCreateKeyResult := s.ormService.GormDB.Create(gormKey)
	if gormCreateKeyResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeyError(gormSelectKeyPoolResult.Error)
	}

	return s.openapiOrmMapper.toOpenapiResponseInsertKeySuccess(gormKey), nil
}

func (s *KeyPoolService) GetKeyPoolKeyPoolIDKey(ctx context.Context, keyPoolID *string) (cryptoutilServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	_, err := uuid.Parse(*keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKeyInvalidKeyPoolIDError(err)
	}

	var gormKeyPool ormService.KeyPool
	gormKeyPoolResult := s.ormService.GormDB.First(&gormKeyPool, "key_pool_id=?", *keyPoolID)
	if gormKeyPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKeyNoKeyPoolIDFoundError(err)
	}

	var gormKeys []ormService.Key
	query := s.ormService.GormDB.Where("key_pool_id=?", *keyPoolID)
	gormKeysResult := query.Find(&gormKeys)
	if gormKeysResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKeyFindError(err)
	}

	return s.openapiOrmMapper.toOpenapiResponseGetKeySuccess(&gormKeys), nil
}

func (s *KeyPoolService) generateKeyInsert(gormKeyPool *ormService.KeyPool, keyPoolNextID int) (*ormService.Key, error) {
	gormKeyKeyMaterial, err := generateKeyMaterial(string(gormKeyPool.KeyPoolAlgorithm))
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key material: %w", err)
	}
	gormKeyGenerateDate := time.Now().UTC()

	return &ormService.Key{
		KeyPoolID:       gormKeyPool.KeyPoolID,
		KeyID:           keyPoolNextID,
		KeyMaterial:     gormKeyKeyMaterial,
		KeyGenerateDate: &gormKeyGenerateDate,
	}, nil
}
