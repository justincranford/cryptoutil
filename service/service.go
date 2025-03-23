package service

import (
	"context"
	"fmt"
	"time"

	"cryptoutil/api/openapi"
	"cryptoutil/orm"

	"github.com/google/uuid"
)

type KEKPoolService struct {
	ormService       *orm.Service
	openapiOrmMapper *OpenapiOrmMapper
}

func NewService(dbService *orm.Service) *KEKPoolService {
	return &KEKPoolService{ormService: dbService, openapiOrmMapper: NewMapper()}
}

func (s *KEKPoolService) PostKEKPool(ctx context.Context, openapiKEKPoolCreate *openapi.KEKPoolCreate) (openapi.PostKekpoolResponseObject, error) {
	gormKEKPoolInsert := s.openapiOrmMapper.toGormKEKPoolInsert(openapiKEKPoolCreate)
	gormCreateKEKResult := s.ormService.GormDB.Create(gormKEKPoolInsert)
	if gormCreateKEKResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKPoolError(gormCreateKEKResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseInsertKEKPoolSuccess(gormKEKPoolInsert), nil
}

func (s *KEKPoolService) GetKEKPool(ctx context.Context) (openapi.GetKekpoolResponseObject, error) {
	var gormKEKPools []orm.KEKPool
	gormFindKEKPoolsResult := s.ormService.GormDB.Find(&gormKEKPools)
	if gormFindKEKPoolsResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseSelectKEKPoolError(gormFindKEKPoolsResult.Error)
	}
	return s.openapiOrmMapper.toOpenapiResponseSelectKEKPoolSuccess(&gormKEKPools), nil
}

func (s *KEKPoolService) PostKEKPoolKEKPoolIDKEK(ctx context.Context, kekPoolID *string) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	_, err := uuid.Parse(*kekPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKSelectKEKPoolError(err)
	}

	var gormKEKPool orm.KEKPool
	gormSelectKEKPoolResult := s.ormService.GormDB.First(&gormKEKPool, "kek_pool_id=?", *kekPoolID)
	if gormSelectKEKPoolResult.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKSelectKEKPoolError(gormSelectKEKPoolResult.Error)
	}

	if gormKEKPool.KEKPoolStatus != orm.PendingGenerate {
		return s.openapiOrmMapper.toOpenapiResponseInsertKEKInvalidKEKPoolStatus()
	}

	var gormKEKPoolMaxID int // COALESCE clause returns 0 if no KEKs found for KEK Pool
	s.ormService.GormDB.Model(&orm.KEK{}).Where("kek_pool_id=?", *kekPoolID).Select("COALESCE(MAX(kek_id), 0)").Scan(&gormKEKPoolMaxID)

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

func (s *KEKPoolService) GetKEKPoolKEKPoolIDKEK(ctx context.Context, kekPoolID *string) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	_, err := uuid.Parse(*kekPoolID)
	if err != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKEKInvalidKEKPoolIDError(err)
	}

	var gormKekPool orm.KEKPool
	gormKEKPool := s.ormService.GormDB.First(&gormKekPool, "kek_pool_id=?", *kekPoolID)
	if gormKEKPool.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKEKNoKEKPoolIDFoundError(err)
	}

	var gormKeks []orm.KEK
	query := s.ormService.GormDB.Where("kek_pool_id=?", *kekPoolID)
	gormKEKPool = query.Find(&gormKeks)
	if gormKEKPool.Error != nil {
		return s.openapiOrmMapper.toOpenapiResponseGetKEKFindError(err)
	}

	return s.openapiOrmMapper.toOpenapiResponseGetKEKSuccess(&gormKeks), nil
}

func (s *KEKPoolService) generateKEKInsert(gormKEKPool *orm.KEKPool, kekPoolNextID int) (*orm.KEK, error) {
	gormKEKKeyMaterial, err := generateKeyMaterial(string(gormKEKPool.KEKPoolAlgorithm))
	if err != nil {
		return nil, fmt.Errorf("failed to generate KEK material: %w", err)
	}
	gormKEKGenerateDate := time.Now().UTC()

	return &orm.KEK{
		KEKPoolID:       gormKEKPool.KEKPoolID,
		KEKID:           kekPoolNextID,
		KEKMaterial:     gormKEKKeyMaterial,
		KEKGenerateDate: &gormKEKGenerateDate,
	}, nil
}
