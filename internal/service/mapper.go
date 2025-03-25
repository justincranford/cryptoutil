package service

import (
	cryptoutilModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/openapi/server"
	cryptoutilOrmService "cryptoutil/internal/orm"
	cryptoutilUtil "cryptoutil/internal/util"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type OpenapiOrmMapper struct{}

func NewMapper() *OpenapiOrmMapper {
	return &OpenapiOrmMapper{}
}

func (m *OpenapiOrmMapper) toGormKEKPoolInsert(openapiKEKPoolCreate *cryptoutilModel.KEKPoolCreate) *cryptoutilOrmService.KEKPool {
	return &cryptoutilOrmService.KEKPool{
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

func (m *OpenapiOrmMapper) toOpenapiKEKPools(gormKEKPools *[]cryptoutilOrmService.KEKPool) *[]cryptoutilModel.KEKPool {
	openapiKEKPools := make([]cryptoutilModel.KEKPool, len(*gormKEKPools))
	for i, gormKekPool := range *gormKEKPools {
		openapiKEKPools[i] = *m.toOpenapiKEKPool(&gormKekPool)
	}
	return &openapiKEKPools
}

func (*OpenapiOrmMapper) toOpenapiKEKPool(gormKekPool *cryptoutilOrmService.KEKPool) *cryptoutilModel.KEKPool {
	return &cryptoutilModel.KEKPool{
		Id:                  (*cryptoutilModel.KEKPoolId)(cryptoutilUtil.StringPtr(gormKekPool.KEKPoolID.String())),
		Name:                &gormKekPool.KEKPoolName,
		Description:         &gormKekPool.KEKPoolDescription,
		Algorithm:           (*cryptoutilModel.KEKPoolAlgorithm)(&gormKekPool.KEKPoolAlgorithm),
		Provider:            (*cryptoutilModel.KEKPoolProvider)(&gormKekPool.KEKPoolProvider),
		IsVersioningAllowed: &gormKekPool.KEKPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKekPool.KEKPoolIsImportAllowed,
		IsExportAllowed:     &gormKekPool.KEKPoolIsExportAllowed,
		Status:              (*cryptoutilModel.KEKPoolStatus)(&gormKekPool.KEKPoolStatus),
	}
}

func (m *OpenapiOrmMapper) toOpenapiKEKs(gormKEKs *[]cryptoutilOrmService.KEK) *[]cryptoutilModel.KEK {
	openapiKEKs := make([]cryptoutilModel.KEK, len(*gormKEKs))
	for i, gormKEK := range *gormKEKs {
		openapiKEKs[i] = *m.toOpenapiKEK(&gormKEK)
	}
	return &openapiKEKs
}

func (*OpenapiOrmMapper) toOpenapiKEK(gormKEK *cryptoutilOrmService.KEK) *cryptoutilModel.KEK {
	return &cryptoutilModel.KEK{
		KekId:        &gormKEK.KEKID,
		KekPoolId:    (*cryptoutilModel.KEKPoolId)(cryptoutilUtil.StringPtr(gormKEK.KEKPoolID.String())),
		GenerateDate: (*cryptoutilModel.KEKGenerateDate)(gormKEK.KEKGenerateDate),
	}
}

func (*OpenapiOrmMapper) toKEKPoolProviderEnum(openapiKEKPoolProvider *cryptoutilModel.KEKPoolProvider) *cryptoutilOrmService.KEKPoolProviderEnum {
	gormKEKPoolProvider := cryptoutilOrmService.KEKPoolProviderEnum(*openapiKEKPoolProvider)
	return &gormKEKPoolProvider
}

func (*OpenapiOrmMapper) toKKEKPoolAlgorithmEnum(openapiKEKPoolProvider *cryptoutilModel.KEKPoolAlgorithm) *cryptoutilOrmService.KEKPoolAlgorithmEnum {
	gormKEKPoolAlgorithm := cryptoutilOrmService.KEKPoolAlgorithmEnum(*openapiKEKPoolProvider)
	return &gormKEKPoolAlgorithm
}

func (*OpenapiOrmMapper) toKEKPoolInitialStatus(openapiKEKPoolIsImportAllowed *cryptoutilModel.KEKPoolIsImportAllowed) *cryptoutilOrmService.KEKPoolStatusEnum {
	var gormKEKPoolStatus cryptoutilOrmService.KEKPoolStatusEnum
	if *openapiKEKPoolIsImportAllowed {
		gormKEKPoolStatus = cryptoutilOrmService.KEKPoolStatusEnum("pending_import")
	} else {
		gormKEKPoolStatus = cryptoutilOrmService.KEKPoolStatusEnum("pending_generate")
	}
	return &gormKEKPoolStatus
}

// PostKEKPool

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKEKPoolSuccess(gormKEKPool *cryptoutilOrmService.KEKPool) cryptoutilServer.PostKekpoolResponseObject {
	openapiPostKekpoolResponseObject := cryptoutilServer.PostKekpool200JSONResponse(*m.toOpenapiKEKPool(gormKEKPool))
	return &openapiPostKekpoolResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKPoolError(err error) (cryptoutilServer.PostKekpoolResponseObject, error) {
	return cryptoutilServer.PostKekpool500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to insert KEK Pool"}}, fmt.Errorf("failed to insert KEK Pool: %w", err)
}

// GetKEKPool

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKEKPoolSuccess(gormKEKPools *[]cryptoutilOrmService.KEKPool) cryptoutilServer.GetKekpoolResponseObject {
	openapiGetKekpoolResponseObject := cryptoutilServer.GetKekpool200JSONResponse(*m.toOpenapiKEKPools(gormKEKPools))
	return &openapiGetKekpoolResponseObject
}

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKEKPoolError(err error) (cryptoutilServer.GetKekpoolResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilServer.GetKekpool404JSONResponse{HTTP404NotFound: cryptoutilModel.HTTP404NotFound{Error: "KEK Pool not found"}}, fmt.Errorf("KEK Pool not found: %w", err)
	}
	return cryptoutilServer.GetKekpool500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to get KEK Pool"}}, fmt.Errorf("failed to get KEK Pool: %w", err)
}

// PostKEKPoolKEKPoolIDKEK

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKEKSuccess(gormKEK *cryptoutilOrmService.KEK) cryptoutilServer.PostKekpoolKekPoolIDKekResponseObject {
	openapiPostKekpoolKekPoolIDKekResponseObject := cryptoutilServer.PostKekpoolKekPoolIDKek200JSONResponse(*m.toOpenapiKEK(gormKEK))
	return &openapiPostKekpoolKekPoolIDKekResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKInvalidKEKPoolID(err error) (cryptoutilServer.PostKekpoolKekPoolIDKekResponseObject, error) {
	return cryptoutilServer.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400BadRequest: cryptoutilModel.HTTP400BadRequest{Error: "KEK Pool ID"}}, fmt.Errorf("KEK Pool ID: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKSelectKEKPoolError(err error) (cryptoutilServer.PostKekpoolKekPoolIDKekResponseObject, error) {
	return cryptoutilServer.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to insert KEK"}}, fmt.Errorf("failed to insert KEK: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKInvalidKEKPoolStatus() (cryptoutilServer.PostKekpoolKekPoolIDKekResponseObject, error) {
	return cryptoutilServer.PostKekpoolKekPoolIDKek400JSONResponse{HTTP400BadRequest: cryptoutilModel.HTTP400BadRequest{Error: "KEK Pool invalid initial state"}}, fmt.Errorf("KEK Pool invalid initial state")
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKGenerateKeyMaterialError(err error) (cryptoutilServer.PostKekpoolKekPoolIDKekResponseObject, error) {
	return &cryptoutilServer.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: err.Error()}}, nil
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKEKError(err error) (cryptoutilServer.PostKekpoolKekPoolIDKekResponseObject, error) {
	return cryptoutilServer.PostKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to insert KEK"}}, fmt.Errorf("failed to insert KEK: %w", err)
}

// GetKEKPoolKEKPoolIDKEK

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKSuccess(gormKEKs *[]cryptoutilOrmService.KEK) cryptoutilServer.GetKekpoolKekPoolIDKekResponseObject {
	openapiGetKekpoolKekPoolIDKekResponseObject := cryptoutilServer.GetKekpoolKekPoolIDKek200JSONResponse(*m.toOpenapiKEKs(gormKEKs))
	return &openapiGetKekpoolKekPoolIDKekResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseGetKEKInvalidKEKPoolIDError(err error) (cryptoutilServer.GetKekpoolKekPoolIDKekResponseObject, error) {
	return cryptoutilServer.GetKekpoolKekPoolIDKek400JSONResponse{HTTP400BadRequest: cryptoutilModel.HTTP400BadRequest{Error: "KEK Pool ID"}}, fmt.Errorf("KEK Pool ID: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKNoKEKPoolIDFoundError(err error) (cryptoutilServer.GetKekpoolKekPoolIDKekResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilServer.GetKekpoolKekPoolIDKek404JSONResponse{HTTP404NotFound: cryptoutilModel.HTTP404NotFound{Error: "KEK Pool not found"}}, fmt.Errorf("KEK Pool not found: %w", err)
	}
	return cryptoutilServer.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to get KEK Pool"}}, fmt.Errorf("failed to get KEK Pool: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKEKFindError(err error) (cryptoutilServer.GetKekpoolKekPoolIDKekResponseObject, error) {
	return cryptoutilServer.GetKekpoolKekPoolIDKek500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to get KEKs"}}, fmt.Errorf("failed to get KEKs: %w", err)
}
