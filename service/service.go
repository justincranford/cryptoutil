package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cryptoutil/api/openapi"
	"cryptoutil/orm"
	"cryptoutil/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KEKPoolService struct {
	ormService       *orm.Service
	openapiOrmMapper *OpenapiOrmMapper
}

func NewService(dbService *orm.Service) *KEKPoolService {
	return &KEKPoolService{ormService: dbService, openapiOrmMapper: NewMapper()}
}

func (s *KEKPoolService) PostKEKPool(ctx context.Context, openapiRequest openapi.PostKekpoolRequestObject) (openapi.PostKekpoolResponseObject, error) {
	gormKEKPool := orm.KEKPool{
		KEKPoolName:                openapiRequest.Body.Name,
		KEKPoolDescription:         openapiRequest.Body.Description,
		KEKPoolProvider:            s.openapiOrmMapper.toKEKPoolProviderEnum(*openapiRequest.Body.Provider),
		KEKPoolAlgorithm:           s.openapiOrmMapper.toKKEKPoolAlgorithmEnum(*openapiRequest.Body.Algorithm),
		KEKPoolIsVersioningAllowed: *openapiRequest.Body.IsVersioningAllowed,
		KEKPoolIsImportAllowed:     *openapiRequest.Body.IsImportAllowed,
		KEKPoolIsExportAllowed:     *openapiRequest.Body.IsExportAllowed,
		KEKPoolStatus:              s.openapiOrmMapper.toKEKPoolStatusImportOrGenerate(*openapiRequest.Body.IsImportAllowed),
	}

	gormCreateKEKResult := s.ormService.GormDB.Create(&gormKEKPool)
	if gormCreateKEKResult.Error != nil {
		return &openapi.PostKekpool500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to insert KEK Pool")}}, fmt.Errorf("failed to insert KEK Pool: %w", gormCreateKEKResult.Error)
	}

	openapiResponse := openapi.PostKekpool200JSONResponse(s.openapiOrmMapper.toOpenapiKEKPool(gormKEKPool))
	return &openapiResponse, nil
}

func (s *KEKPoolService) GetKEKPool(ctx context.Context, openapiRequest openapi.GetKekpoolRequestObject) (openapi.GetKekpoolResponseObject, error) {
	var gormKEKPools []orm.KEKPool
	gormFindKEKPoolsResult := s.ormService.GormDB.Find(&gormKEKPools)
	if gormFindKEKPoolsResult.Error != nil {
		if errors.Is(gormFindKEKPoolsResult.Error, gorm.ErrRecordNotFound) {
			return &openapi.GetKekpool404JSONResponse{HTTP404JSONResponse: openapi.HTTP404JSONResponse{Error: util.StringPtr("KEK Pool not found")}}, fmt.Errorf("KEK Pool not found: %w", gormFindKEKPoolsResult.Error)
		}
		return &openapi.GetKekpool500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to get KEK Pool")}}, fmt.Errorf("failed to get KEK Pool: %w", gormFindKEKPoolsResult.Error)
	}

	openapiResponse := openapi.GetKekpool200JSONResponse(s.openapiOrmMapper.toOpenapiKEKPools(gormKEKPools))
	return &openapiResponse, nil
}

func (s *KEKPoolService) PostKEKPoolKEKPoolIDKEK(ctx context.Context, openapiRequest openapi.PostKekpoolKekPoolIDKekRequestObject) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	gormKEKPoolID, err := uuid.Parse(openapiRequest.KekPoolID)
	if err != nil {
		return &openapi.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: util.StringPtr("KEK Pool ID")}}, fmt.Errorf("failed to get KEK Pool: %w", err)
	}

	var gormKEKPool orm.KEKPool
	gormSelectKEKPoolResult := s.ormService.GormDB.First(&gormKEKPool, "kek_pool_id=?", openapiRequest.KekPoolID)
	if gormSelectKEKPoolResult.Error != nil {
		if errors.Is(gormSelectKEKPoolResult.Error, gorm.ErrRecordNotFound) {
			return &openapi.PostKekpoolKekPoolIDKek404JSONResponse{HTTP404JSONResponse: openapi.HTTP404JSONResponse{Error: util.StringPtr("KEK Pool not found")}}, fmt.Errorf("KEK Pool not found: %w", err)
		}
		return &openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("KEK Pool find error")}}, fmt.Errorf("KEK Pool find error: %w", err)
	}

	if gormKEKPool.KEKPoolStatus != orm.PendingGenerate && gormKEKPool.KEKPoolStatus != orm.PendingImport {
		return &openapi.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: util.StringPtr("KEK Pool invalid initial state")}}, fmt.Errorf("KEK Pool invalid initial state")
	}

	gormKEKKeyMaterial, err := generateKEKMaterial(string(gormKEKPool.KEKPoolAlgorithm))
	if err != nil {
		return &openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr(err.Error())}}, nil
	}
	gormKEKGenerateDate := time.Now().UTC()

	var gormMaxKEKIDForKEKPool int // COALESCE clause returns 0 if no KEKs found for KEK Pool
	s.ormService.GormDB.Model(&orm.KEK{}).Where("kek_pool_id=?", openapiRequest.KekPoolID).Select("COALESCE(MAX(kek_id), 0)").Scan(&gormMaxKEKIDForKEKPool)
	gormNextKEKIDForKEKPool := gormMaxKEKIDForKEKPool + 1

	gormKEK := orm.KEK{
		KEKPoolID:       gormKEKPoolID,
		KEKID:           gormNextKEKIDForKEKPool,
		KEKMaterial:     gormKEKKeyMaterial,
		KEKGenerateDate: &gormKEKGenerateDate,
	}

	gormCreateKEKResult := s.ormService.GormDB.Create(&gormKEK)
	if gormCreateKEKResult.Error != nil {
		return &openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("KEK Pool find error")}}, fmt.Errorf("KEK Pool find error: %w", gormCreateKEKResult.Error)
	}

	openapiResponse := openapi.PostKekpoolKekPoolIDKek200JSONResponse(s.openapiOrmMapper.toOpenapiKEK(gormKEK)) // KeyMaterial is not returned
	return &openapiResponse, nil
}

func (s *KEKPoolService) GetKEKPoolKEKPoolIDKEK(ctx context.Context, openapiRequest openapi.GetKekpoolKekPoolIDKekRequestObject) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	_, err := uuid.Parse(openapiRequest.KekPoolID)
	if err != nil {
		return &openapi.GetKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: util.StringPtr("KEK Pool ID is not valid UUID")}}, fmt.Errorf("KEK Pool ID is not valid UUID: %w", err)
	}

	var gormKekPool orm.KEKPool
	gormKEKPool := s.ormService.GormDB.First(&gormKekPool, "kek_pool_id=?", openapiRequest.KekPoolID)
	if gormKEKPool.Error != nil {
		if errors.Is(gormKEKPool.Error, gorm.ErrRecordNotFound) {
			return &openapi.GetKekpoolKekPoolIDKek404JSONResponse{HTTP404JSONResponse: openapi.HTTP404JSONResponse{Error: util.StringPtr("KEK Pool id not found")}}, fmt.Errorf("KEK Pool id not found: %w", gormKEKPool.Error)
		}
		return &openapi.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("KEK Pool id lookup error")}}, fmt.Errorf("KEK Pool id lookup error: %w", gormKEKPool.Error)
	}

	var gormKeks []orm.KEK
	query := s.ormService.GormDB.Where("kek_pool_id=?", openapiRequest.KekPoolID)
	gormKEKPool = query.Find(&gormKeks)
	if gormKEKPool.Error != nil {
		return &openapi.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("KEKs lookup error")}}, fmt.Errorf("KEKs lookup error: %w", gormKEKPool.Error)
	}

	openapiResponse := openapi.GetKekpoolKekPoolIDKek200JSONResponse(s.openapiOrmMapper.toOpenapiKEKs(gormKeks)) // KeyMaterial is not returned
	return &openapiResponse, nil
}
