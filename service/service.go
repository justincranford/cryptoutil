package service

import (
	"context"
	"fmt"
	"time"

	"cryptoutil/internal/openapi/model"
	"cryptoutil/internal/openapi/server"
	ormService "cryptoutil/internal/orm"

	"github.com/google/uuid"
)

type KEKPoolService struct {
	ormService       *ormService.Service
	openapiOrmMapper *OpenapiOrmMapper
}

func NewService(dbService *ormService.Service) *KEKPoolService {
	return &KEKPoolService{ormService: dbService, openapiOrmMapper: NewMapper()}
}

func (s *KEKPoolService) PostKEKPool(ctx context.Context, openapiKEKPoolCreate *model.KEKPoolCreate) (server.PostKekpoolResponseObject, error) {
	gormKEKPoolInsert := s.openapiOrmMapper.toGormKEKPoolInsert(openapiKEKPoolCreate)
	gormCreateKEKResult := s.ormService.GormDB.Create(gormKEKPoolInsert)
	if gormCreateKEKResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKPoolError(gormCreateKEKResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseInsertKEKPoolSuccess(gormKEKPoolInsert), nil
}

func (s *KEKPoolService) GetKEKPool(ctx context.Context) (server.GetKekpoolResponseObject, error) {
	var gormKEKPools []ormService.KEKPool
	gormFindKEKPoolsResult := s.ormService.GormDB.Find(&gormKEKPools)
	if gormFindKEKPoolsResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseSelectKEKPoolError(gormFindKEKPoolsResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseSelectKEKPoolSuccess(&gormKEKPools), nil
}

func (s *KEKPoolService) PostKEKPoolKEKPoolIDKEK(ctx context.Context, kekPoolID *string, _ *model.KEKGenerate) (server.PostKekpoolKekPoolIDKekResponseObject, error) {
	_, err := uuid.Parse(*kekPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKSelectKEKPoolError(err)
	}

	var gormKEKPool ormService.KEKPool
	gormSelectKEKPoolResult := s.ormService.GormDB.First(&gormKEKPool, "kek_pool_id=?", *kekPoolID)
	if gormSelectKEKPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKSelectKEKPoolError(gormSelectKEKPoolResult.Error)
	}

	if gormKEKPool.KEKPoolStatus != ormService.PendingGenerate {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKInvalidKEKPoolStatus()
	}

	var gormKEKPoolMaxID int // COALESCE clause returns 0 if no KEKs found for KEK Pool
	s.ormService.GormDB.Model(&ormService.KEK{}).Where("kek_pool_id=?", *kekPoolID).Select("COALESCE(MAX(kek_id), 0)").Scan(&gormKEKPoolMaxID)

	gormKEK, err := s.generateKEKInsert(&gormKEKPool, gormKEKPoolMaxID+1)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKGenerateKeyMaterialError(err)
	}

	gormCreateKEKResult := s.ormService.GormDB.Create(gormKEK)
	if gormCreateKEKResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKError(gormSelectKEKPoolResult.Error)
	}

	return s.openapiOrmMapper.toOpenapiResponseInsertKEKSuccess(gormKEK), nil
}

func (s *KEKPoolService) GetKEKPoolKEKPoolIDKEK(ctx context.Context, kekPoolID *string) (server.GetKekpoolKekPoolIDKekResponseObject, error) {
	_, err := uuid.Parse(*kekPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKEKInvalidKEKPoolIDError(err)
	}

	var gormKekPool ormService.KEKPool
	gormKEKPool := s.ormService.GormDB.First(&gormKekPool, "kek_pool_id=?", *kekPoolID)
	if gormKEKPool.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKEKNoKEKPoolIDFoundError(err)
	}

	var gormKeks []ormService.KEK
	query := s.ormService.GormDB.Where("kek_pool_id=?", *kekPoolID)
	gormKEKPool = query.Find(&gormKeks)
	if gormKEKPool.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKEKFindError(err)
	}

	return s.openapiOrmMapper.toOpenapiResponseGetKEKSuccess(&gormKeks), nil
}

func (s *KEKPoolService) generateKEKInsert(gormKEKPool *ormService.KEKPool, kekPoolNextID int) (*ormService.KEK, error) {
	gormKEKKeyMaterial, err := generateKeyMaterial(string(gormKEKPool.KEKPoolAlgorithm))
	if err != nil {
		return nil, fmt.Errorf("failed to generate KEK material: %w", err)
	}
	gormKEKGenerateDate := time.Now().UTC()

	return &ormService.KEK{
		KEKPoolID:       gormKEKPool.KEKPoolID,
		KEKID:           kekPoolNextID,
		KEKMaterial:     gormKEKKeyMaterial,
		KEKGenerateDate: &gormKEKGenerateDate,
	}, nil
}
