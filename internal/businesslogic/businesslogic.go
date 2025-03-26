package service

import (
	"context"
	"fmt"
	"time"

	cryptoutilCryptoKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilRepositoryOrm "cryptoutil/internal/repository/orm"

	"github.com/google/uuid"
)

type KeyPoolService struct {
	ormService       *cryptoutilRepositoryOrm.Service
	openapiOrmMapper *OpenapiOrmMapper
}

func NewService(dbService *cryptoutilRepositoryOrm.Service) *KeyPoolService {
	return &KeyPoolService{ormService: dbService, openapiOrmMapper: NewMapper()}
}

func (s *KeyPoolService) PostKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilOpenapiModel.KeyPoolCreate) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	gormKeyPoolInsert := s.openapiOrmMapper.toGormKeyPoolInsert(openapiKeyPoolCreate)
	gormCreateKeyResult := s.ormService.GormDB.Create(gormKeyPoolInsert)
	if gormCreateKeyResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeyPoolError(gormCreateKeyResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseInsertKeyPoolSuccess(gormKeyPoolInsert), nil
}

func (s *KeyPoolService) GetKeyPool(ctx context.Context) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	var gormKeyPools []cryptoutilRepositoryOrm.KeyPool
	gormFindKeyPoolsResult := s.ormService.GormDB.Find(&gormKeyPools)
	if gormFindKeyPoolsResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseSelectKeyPoolError(gormFindKeyPoolsResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseSelectKeyPoolSuccess(&gormKeyPools), nil
}

func (s *KeyPoolService) PostKeyPoolKeyPoolIDKey(ctx context.Context, keyPoolID *string, _ *cryptoutilOpenapiModel.KeyGenerate) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	_, err := uuid.Parse(*keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeySelectKeyPoolError(err)
	}

	var gormKeyPool cryptoutilRepositoryOrm.KeyPool
	gormSelectKeyPoolResult := s.ormService.GormDB.First(&gormKeyPool, "key_pool_id=?", *keyPoolID)
	if gormSelectKeyPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeySelectKeyPoolError(gormSelectKeyPoolResult.Error)
	}

	if gormKeyPool.KeyPoolStatus != cryptoutilRepositoryOrm.PendingGenerate {
		return s.openapiOrmMapper.toOpenapiResponseInsertKeyInvalidKeyPoolStatus()
	}

	var gormKeyPoolMaxID int // COALESCE clause returns 0 if no Keys found for Key Pool
	s.ormService.GormDB.Model(&cryptoutilRepositoryOrm.Key{}).Where("key_pool_id=?", *keyPoolID).Select("COALESCE(MAX(key_id), 0)").Scan(&gormKeyPoolMaxID)

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

func (s *KeyPoolService) GetKeyPoolKeyPoolIDKey(ctx context.Context, keyPoolID *string) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	_, err := uuid.Parse(*keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKeyInvalidKeyPoolIDError(err)
	}

	var gormKeyPool cryptoutilRepositoryOrm.KeyPool
	gormKeyPoolResult := s.ormService.GormDB.First(&gormKeyPool, "key_pool_id=?", *keyPoolID)
	if gormKeyPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKeyNoKeyPoolIDFoundError(err)
	}

	var gormKeys []cryptoutilRepositoryOrm.Key
	query := s.ormService.GormDB.Where("key_pool_id=?", *keyPoolID)
	gormKeysResult := query.Find(&gormKeys)
	if gormKeysResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKeyFindError(err)
	}

	return s.openapiOrmMapper.toOpenapiResponseGetKeySuccess(&gormKeys), nil
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
