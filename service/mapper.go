package service

import (
	"cryptoutil/internal/openapi/model"
	"cryptoutil/internal/openapi/server"
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

func (m *OpenapiOrmMapper) toGormKEKPoolInsert(openapiKEKPoolCreate *model.KEKPoolCreate) *orm.KEKPool {
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

func (m *OpenapiOrmMapper) toOpenapiKEKPools(gormKEKPools *[]orm.KEKPool) *[]model.KEKPool {
	openapiKEKPools := make([]model.KEKPool, len(*gormKEKPools))
	for i, gormKekPool := range *gormKEKPools {
		openapiKEKPools[i] = *m.toOpenapiKEKPool(&gormKekPool)
	}
	return &openapiKEKPools
}

func (*OpenapiOrmMapper) toOpenapiKEKPool(gormKekPool *orm.KEKPool) *model.KEKPool {
	return &model.KEKPool{
		Id:                  (*model.KEKPoolId)(util.StringPtr(gormKekPool.KEKPoolID.String())),
		Name:                &gormKekPool.KEKPoolName,
		Description:         &gormKekPool.KEKPoolDescription,
		Algorithm:           (*model.KEKPoolAlgorithm)(&gormKekPool.KEKPoolAlgorithm),
		Provider:            (*model.KEKPoolProvider)(&gormKekPool.KEKPoolProvider),
		IsVersioningAllowed: &gormKekPool.KEKPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKekPool.KEKPoolIsImportAllowed,
		IsExportAllowed:     &gormKekPool.KEKPoolIsExportAllowed,
		Status:              (*model.KEKPoolStatus)(&gormKekPool.KEKPoolStatus),
	}
}

func (m *OpenapiOrmMapper) toOpenapiKEKs(gormKEKs *[]orm.KEK) *[]model.KEK {
	openapiKEKs := make([]model.KEK, len(*gormKEKs))
	for i, gormKEK := range *gormKEKs {
		openapiKEKs[i] = *m.toOpenapiKEK(&gormKEK)
	}
	return &openapiKEKs
}

func (*OpenapiOrmMapper) toOpenapiKEK(gormKEK *orm.KEK) *model.KEK {
	return &model.KEK{
		KekId:        &gormKEK.KEKID,
		KekPoolId:    (*model.KEKPoolId)(util.StringPtr(gormKEK.KEKPoolID.String())),
		GenerateDate: (*model.KEKGenerateDate)(gormKEK.KEKGenerateDate),
	}
}

func (*OpenapiOrmMapper) toKEKPoolProviderEnum(openapiKEKPoolProvider *model.KEKPoolProvider) *orm.KEKPoolProviderEnum {
	gormKEKPoolProvider := orm.KEKPoolProviderEnum(*openapiKEKPoolProvider)
	return &gormKEKPoolProvider
}

func (*OpenapiOrmMapper) toKKEKPoolAlgorithmEnum(openapiKEKPoolProvider *model.KEKPoolAlgorithm) *orm.KEKPoolAlgorithmEnum {
	gormKEKPoolAlgorithm := orm.KEKPoolAlgorithmEnum(*openapiKEKPoolProvider)
	return &gormKEKPoolAlgorithm
}

func (*OpenapiOrmMapper) toKEKPoolInitialStatus(openapiKEKPoolIsImportAllowed *model.KEKPoolIsImportAllowed) *orm.KEKPoolStatusEnum {
	var gormKEKPoolStatus orm.KEKPoolStatusEnum
	if *openapiKEKPoolIsImportAllowed {
		gormKEKPoolStatus = orm.KEKPoolStatusEnum("pending_import")
	} else {
		gormKEKPoolStatus = orm.KEKPoolStatusEnum("pending_generate")
	}
	return &gormKEKPoolStatus
}

// PostKEKPool

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKEKPoolSuccess(gormKEKPool *orm.KEKPool) server.PostKekpoolResponseObject {
	openapiPostKekpoolResponseObject := server.PostKekpool200JSONResponse(*m.toOpenapiKEKPool(gormKEKPool))
	return &openapiPostKekpoolResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKPoolError(err error) (server.PostKekpoolResponseObject, error) {
	return server.PostKekpool500JSONResponse{HTTP500InternalServerError: model.HTTP500InternalServerError{Error: "failed to insert KEK Pool"}}, fmt.Errorf("failed to insert KEK Pool: %w", err)
}

// GetKEKPool

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKEKPoolSuccess(gormKEKPools *[]orm.KEKPool) server.GetKekpoolResponseObject {
	openapiGetKekpoolResponseObject := server.GetKekpool200JSONResponse(*m.toOpenapiKEKPools(gormKEKPools))
	return &openapiGetKekpoolResponseObject
}

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKEKPoolError(err error) (server.GetKekpoolResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return server.GetKekpool404JSONResponse{HTTP404NotFound: model.HTTP404NotFound{Error: "KEK Pool not found"}}, fmt.Errorf("KEK Pool not found: %w", err)
	}
	return server.GetKekpool500JSONResponse{HTTP500InternalServerError: model.HTTP500InternalServerError{Error: "failed to get KEK Pool"}}, fmt.Errorf("failed to get KEK Pool: %w", err)
}

// PostKEKPoolKEKPoolIDKEK

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKEKSuccess(gormKEK *orm.KEK) server.PostKekpoolKekPoolIDKekResponseObject {
	openapiPostKekpoolKekPoolIDKekResponseObject := server.PostKekpoolKekPoolIDKek200JSONResponse(*m.toOpenapiKEK(gormKEK))
	return &openapiPostKekpoolKekPoolIDKekResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKInvalidKEKPoolID(err error) (server.PostKekpoolKekPoolIDKekResponseObject, error) {
	return server.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400BadRequest: model.HTTP400BadRequest{Error: "KEK Pool ID"}}, fmt.Errorf("KEK Pool ID: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKSelectKEKPoolError(err error) (server.PostKekpoolKekPoolIDKekResponseObject, error) {
	return server.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: model.HTTP500InternalServerError{Error: "failed to insert KEK"}}, fmt.Errorf("failed to insert KEK: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKInvalidKEKPoolStatus() (server.PostKekpoolKekPoolIDKekResponseObject, error) {
	return server.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400BadRequest: model.HTTP400BadRequest{Error: "KEK Pool invalid initial state"}}, fmt.Errorf("KEK Pool invalid initial state")
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKGenerateKeyMaterialError(err error) (server.PostKekpoolKekPoolIDKekResponseObject, error) {
	return &server.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: model.HTTP500InternalServerError{Error: err.Error()}}, nil
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKError(err error) (server.PostKekpoolKekPoolIDKekResponseObject, error) {
	return server.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: model.HTTP500InternalServerError{Error: "failed to insert KEK"}}, fmt.Errorf("failed to insert KEK: %w", err)
}

// GetKEKPoolKEKPoolIDKEK

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKSuccess(gormKEKs *[]orm.KEK) server.GetKekpoolKekPoolIDKekResponseObject {
	openapiGetKekpoolKekPoolIDKekResponseObject := server.GetKekpoolKekPoolIDKek200JSONResponse(*m.toOpenapiKEKs(gormKEKs))
	return &openapiGetKekpoolKekPoolIDKekResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseGetKEKInvalidKEKPoolIDError(err error) (server.GetKekpoolKekPoolIDKekResponseObject, error) {
	return server.GetKekpoolKekPoolIDKek400JSONResponse{HTTP400BadRequest: model.HTTP400BadRequest{Error: "KEK Pool ID"}}, fmt.Errorf("KEK Pool ID: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKNoKEKPoolIDFoundError(err error) (server.GetKekpoolKekPoolIDKekResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return server.GetKekpoolKekPoolIDKek404JSONResponse{HTTP404NotFound: model.HTTP404NotFound{Error: "KEK Pool not found"}}, fmt.Errorf("KEK Pool not found: %w", err)
	}
	return server.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: model.HTTP500InternalServerError{Error: "failed to get KEK Pool"}}, fmt.Errorf("failed to get KEK Pool: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKFindError(err error) (server.GetKekpoolKekPoolIDKekResponseObject, error) {
	return server.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: model.HTTP500InternalServerError{Error: "failed to get KEKs"}}, fmt.Errorf("failed to get KEKs: %w", err)
}
