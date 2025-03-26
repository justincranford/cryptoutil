package service

import (
	"context"
	"fmt"
	"time"

	cryptoutilCryptoKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilBusinessModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilRepositoryOrm "cryptoutil/internal/repository/orm"

	"github.com/google/uuid"
)

type KeyPoolService struct {
	ormService       *cryptoutilRepositoryOrm.Service
	openapiOrmMapper *openapiOrmMapper
}

func NewService(dbService *cryptoutilRepositoryOrm.Service) *KeyPoolService {
	return &KeyPoolService{ormService: dbService, openapiOrmMapper: NewMapper()}
}

func (s *KeyPoolService) PostKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilBusinessModel.KeyPoolCreate) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	gormKeyPoolInsert := s.openapiOrmMapper.toGormKeyPoolInsert(openapiKeyPoolCreate)
	gormCreateKeyResult := s.ormService.GormDB.Create(gormKeyPoolInsert)
	if gormCreateKeyResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyPoolResponseError(gormCreateKeyResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiInsertKeyPoolResponseSuccess(gormKeyPoolInsert), nil
}

func (s *KeyPoolService) GetKeyPool(ctx context.Context) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	var gormKeyPools []cryptoutilRepositoryOrm.KeyPool
	gormFindKeyPoolsResult := s.ormService.GormDB.Find(&gormKeyPools)
	if gormFindKeyPoolsResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiSelectKeyPoolResponseError(gormFindKeyPoolsResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiSelectKeyPoolResponseSuccess(&gormKeyPools), nil
}

func (s *KeyPoolService) PostKeyPoolKeyPoolIDKey(ctx context.Context, keyPoolID *string, _ *cryptoutilBusinessModel.KeyGenerate) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	_, err := uuid.Parse(*keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyInvalidKeyPoolIDResponseError(err)
	}

	var gormKeyPool cryptoutilRepositoryOrm.KeyPool
	gormSelectKeyPoolResult := s.ormService.GormDB.First(&gormKeyPool, "key_pool_id=?", *keyPoolID)
	if gormSelectKeyPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeySelectKeyPoolResponseError(gormSelectKeyPoolResult.Error)
	}

	if gormKeyPool.KeyPoolStatus != cryptoutilRepositoryOrm.PendingGenerate {
		return s.openapiOrmMapper.toOpenapiInsertKeyInvalidKeyPoolStatusResponseError()
	}

	var gormKeyPoolMaxID int // COALESCE clause returns 0 if no Keys found for Key Pool
	s.ormService.GormDB.Model(&cryptoutilRepositoryOrm.Key{}).Where("key_pool_id=?", *keyPoolID).Select("COALESCE(MAX(key_id), 0)").Scan(&gormKeyPoolMaxID)

	gormKey, err := s.generateKeyInsert(&gormKeyPool, gormKeyPoolMaxID+1)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyGenerateKeyMaterialResponseError(err)
	}

	gormCreateKeyResult := s.ormService.GormDB.Create(gormKey)
	if gormCreateKeyResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiInsertKeyResponseError(gormSelectKeyPoolResult.Error)
	}

	return s.openapiOrmMapper.toOpenapiInsertKeySuccessResponseError(gormKey), nil
}

func (s *KeyPoolService) GetKeyPoolKeyPoolIDKey(ctx context.Context, keyPoolID *string) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	_, err := uuid.Parse(*keyPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiGetKeyInvalidKeyPoolIDResponseError(err)
	}

	var gormKeyPool cryptoutilRepositoryOrm.KeyPool
	gormKeyPoolResult := s.ormService.GormDB.First(&gormKeyPool, "key_pool_id=?", *keyPoolID)
	if gormKeyPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiGetKeyNoKeyPoolIDFoundResponseError(err)
	}

	var gormKeys []cryptoutilRepositoryOrm.Key
	query := s.ormService.GormDB.Where("key_pool_id=?", *keyPoolID)
	gormKeysResult := query.Find(&gormKeys)
	if gormKeysResult.Error != nil {
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
