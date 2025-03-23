package service

import (
	"cryptoutil/api/openapi"
	"cryptoutil/orm"
	"cryptoutil/util"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type OpenapiOrmMapper struct{}

func NewMapper() *OpenapiOrmMapper {
	return &OpenapiOrmMapper{}
}

func (m *OpenapiOrmMapper) toGormKEKPoolInsert(openapiKEKPoolCreate *openapi.KEKPoolCreate) *orm.KEKPool {
	return &orm.KEKPool{
		KEKPoolName:                openapiKEKPoolCreate.Name,
		KEKPoolDescription:         openapiKEKPoolCreate.Description,
		KEKPoolProvider:            *m.toKEKPoolProviderEnum(openapiKEKPoolCreate.Provider),
		KEKPoolAlgorithm:           *m.toKKEKPoolAlgorithmEnum(openapiKEKPoolCreate.Algorithm),
		KEKPoolIsVersioningAllowed: *openapiKEKPoolCreate.IsVersioningAllowed,
		KEKPoolIsImportAllowed:     *openapiKEKPoolCreate.IsImportAllowed,
		KEKPoolIsExportAllowed:     *openapiKEKPoolCreate.IsExportAllowed,
		KEKPoolStatus:              *m.toKEKPoolInitialStatus(openapiKEKPoolCreate.IsImportAllowed),
	}
}

func (m *OpenapiOrmMapper) toOpenapiKEKPools(gormKEKPools *[]orm.KEKPool) *[]openapi.KEKPool {
	openapiKEKPools := make([]openapi.KEKPool, len(*gormKEKPools))
	for i, gormKekPool := range *gormKEKPools {
		openapiKEKPools[i] = *m.toOpenapiKEKPool(&gormKekPool)
	}
	return &openapiKEKPools
}

func (*OpenapiOrmMapper) toOpenapiKEKPool(gormKekPool *orm.KEKPool) *openapi.KEKPool {
	return &openapi.KEKPool{
		Id:                  (*openapi.KEKPoolId)(util.StringPtr(gormKekPool.KEKPoolID.String())),
		Name:                &gormKekPool.KEKPoolName,
		Description:         &gormKekPool.KEKPoolDescription,
		Algorithm:           (*openapi.KEKPoolAlgorithm)(&gormKekPool.KEKPoolAlgorithm),
		Provider:            (*openapi.KEKPoolProvider)(&gormKekPool.KEKPoolProvider),
		IsVersioningAllowed: &gormKekPool.KEKPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKekPool.KEKPoolIsImportAllowed,
		IsExportAllowed:     &gormKekPool.KEKPoolIsExportAllowed,
		Status:              (*openapi.KEKPoolStatus)(&gormKekPool.KEKPoolStatus),
	}
}

func (m *OpenapiOrmMapper) toOpenapiKEKs(gormKEKs *[]orm.KEK) *[]openapi.KEK {
	openapiKEKs := make([]openapi.KEK, len(*gormKEKs))
	for i, gormKEK := range *gormKEKs {
		openapiKEKs[i] = *m.toOpenapiKEK(&gormKEK)
	}
	return &openapiKEKs
}

func (*OpenapiOrmMapper) toOpenapiKEK(gormKEK *orm.KEK) *openapi.KEK {
	return &openapi.KEK{
		KekId:        &gormKEK.KEKID,
		KekPoolId:    (*openapi.KEKPoolId)(util.StringPtr(gormKEK.KEKPoolID.String())),
		GenerateDate: (*openapi.KEKGenerateDate)(gormKEK.KEKGenerateDate),
	}
}

func (*OpenapiOrmMapper) toKEKPoolProviderEnum(openapiKEKPoolProvider *openapi.KEKPoolProvider) *orm.KEKPoolProviderEnum {
	gormKEKPoolProvider := orm.KEKPoolProviderEnum(*openapiKEKPoolProvider)
	return &gormKEKPoolProvider
}

func (*OpenapiOrmMapper) toKKEKPoolAlgorithmEnum(openapiKEKPoolProvider *openapi.KEKPoolAlgorithm) *orm.KEKPoolAlgorithmEnum {
	gormKEKPoolAlgorithm := orm.KEKPoolAlgorithmEnum(*openapiKEKPoolProvider)
	return &gormKEKPoolAlgorithm
}

func (*OpenapiOrmMapper) toKEKPoolInitialStatus(openapiKEKPoolIsImportAllowed *openapi.KEKPoolIsImportAllowed) *orm.KEKPoolStatusEnum {
	var gormKEKPoolStatus orm.KEKPoolStatusEnum
	if *openapiKEKPoolIsImportAllowed {
		gormKEKPoolStatus = orm.KEKPoolStatusEnum("pending_import")
	} else {
		gormKEKPoolStatus = orm.KEKPoolStatusEnum("pending_generate")
	}
	return &gormKEKPoolStatus
}

// PostKEKPool

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKEKPoolSuccess(gormKEKPool *orm.KEKPool) openapi.PostKekpoolResponseObject {
	openapiPostKekpoolResponseObject := openapi.PostKekpool200JSONResponse(*m.toOpenapiKEKPool(gormKEKPool))
	return &openapiPostKekpoolResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKPoolError(err error) (openapi.PostKekpoolResponseObject, error) {
	return openapi.PostKekpool500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to insert KEK Pool")}}, fmt.Errorf("failed to insert KEK Pool: %w", err)
}

// GetKEKPool

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKEKPoolSuccess(gormKEKPools *[]orm.KEKPool) openapi.GetKekpoolResponseObject {
	openapiGetKekpoolResponseObject := openapi.GetKekpool200JSONResponse(*m.toOpenapiKEKPools(gormKEKPools))
	return &openapiGetKekpoolResponseObject
}

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKEKPoolError(err error) (openapi.GetKekpoolResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return openapi.GetKekpool404JSONResponse{HTTP404JSONResponse: openapi.HTTP404JSONResponse{Error: util.StringPtr("KEK Pool not found")}}, fmt.Errorf("KEK Pool not found: %w", err)
	}
	return openapi.GetKekpool500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to get KEK Pool")}}, fmt.Errorf("failed to get KEK Pool: %w", err)
}

// PostKEKPoolKEKPoolIDKEK

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKEKSuccess(gormKEK *orm.KEK) openapi.PostKekpoolKekPoolIDKekResponseObject {
	openapiPostKekpoolKekPoolIDKekResponseObject := openapi.PostKekpoolKekPoolIDKek200JSONResponse(*m.toOpenapiKEK(gormKEK))
	return &openapiPostKekpoolKekPoolIDKekResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKInvalidKEKPoolID(err error) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	return openapi.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: util.StringPtr("KEK Pool ID")}}, fmt.Errorf("KEK Pool ID: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKSelectKEKPoolError(err error) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	return openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to insert KEK")}}, fmt.Errorf("failed to insert KEK: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKInvalidKEKPoolStatus() (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	return openapi.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: util.StringPtr("KEK Pool invalid initial state")}}, fmt.Errorf("KEK Pool invalid initial state")
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKGenerateKeyMaterialError(err error) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	return &openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr(err.Error())}}, nil
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKError(err error) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	return openapi.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to insert KEK")}}, fmt.Errorf("failed to insert KEK: %w", err)
}

// GetKEKPoolKEKPoolIDKEK

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKSuccess(gormKEKs *[]orm.KEK) openapi.GetKekpoolKekPoolIDKekResponseObject {
	openapiGetKekpoolKekPoolIDKekResponseObject := openapi.GetKekpoolKekPoolIDKek200JSONResponse(*m.toOpenapiKEKs(gormKEKs))
	return &openapiGetKekpoolKekPoolIDKekResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseGetKEKInvalidKEKPoolIDError(err error) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	return openapi.GetKekpoolKekPoolIDKek400JSONResponse{HTTP400JSONResponse: openapi.HTTP400JSONResponse{Error: util.StringPtr("KEK Pool ID")}}, fmt.Errorf("KEK Pool ID: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKNoKEKPoolIDFoundError(err error) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return openapi.GetKekpoolKekPoolIDKek404JSONResponse{HTTP404JSONResponse: openapi.HTTP404JSONResponse{Error: util.StringPtr("KEK Pool not found")}}, fmt.Errorf("KEK Pool not found: %w", err)
	}
	return openapi.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to get KEK Pool")}}, fmt.Errorf("failed to get KEK Pool: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKFindError(err error) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	return openapi.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500JSONResponse: openapi.HTTP500JSONResponse{Error: util.StringPtr("failed to get KEKs")}}, fmt.Errorf("failed to get KEKs: %w", err)
}
