package service

import (
	cryptoutilModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/openapi/server"
	cryptoutilOrmService "cryptoutil/internal/orm"
	cryptoutilPointer "cryptoutil/internal/pointer"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type OpenapiOrmMapper struct{}

func NewMapper() *OpenapiOrmMapper {
	return &OpenapiOrmMapper{}
}

func (m *OpenapiOrmMapper) toGormKeyPoolInsert(openapiKeyPoolCreate *cryptoutilModel.KeyPoolCreate) *cryptoutilOrmService.KeyPool {
	return &cryptoutilOrmService.KeyPool{
		KeyPoolName:                openapiKeyPoolCreate.Name,
		KeyPoolDescription:         openapiKeyPoolCreate.Description,
		KeyPoolProvider:            *m.toKeyPoolProviderEnum(openapiKeyPoolCreate.Provider),
		KeyPoolAlgorithm:           *m.toKKeyPoolAlgorithmEnum(openapiKeyPoolCreate.Algorithm),
		KeyPoolIsVersioningAllowed: *openapiKeyPoolCreate.IsVersioningAllowed,
		KeyPoolIsImportAllowed:     *openapiKeyPoolCreate.IsImportAllowed,
		KeyPoolIsExportAllowed:     *openapiKeyPoolCreate.IsExportAllowed,
		KeyPoolStatus:              *m.toKeyPoolInitialStatus(openapiKeyPoolCreate.IsImportAllowed),
	}
}

func (m *OpenapiOrmMapper) toOpenapiKeyPools(gormKeyPools *[]cryptoutilOrmService.KeyPool) *[]cryptoutilModel.KeyPool {
	openapiKeyPools := make([]cryptoutilModel.KeyPool, len(*gormKeyPools))
	for i, gormKeyPool := range *gormKeyPools {
		openapiKeyPools[i] = *m.toOpenapiKeyPool(&gormKeyPool)
	}
	return &openapiKeyPools
}

func (*OpenapiOrmMapper) toOpenapiKeyPool(gormKeyPool *cryptoutilOrmService.KeyPool) *cryptoutilModel.KeyPool {
	return &cryptoutilModel.KeyPool{
		Id:                  (*cryptoutilModel.KeyPoolId)(cryptoutilPointer.StringPtr(gormKeyPool.KeyPoolID.String())),
		Name:                &gormKeyPool.KeyPoolName,
		Description:         &gormKeyPool.KeyPoolDescription,
		Algorithm:           (*cryptoutilModel.KeyPoolAlgorithm)(&gormKeyPool.KeyPoolAlgorithm),
		Provider:            (*cryptoutilModel.KeyPoolProvider)(&gormKeyPool.KeyPoolProvider),
		IsVersioningAllowed: &gormKeyPool.KeyPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKeyPool.KeyPoolIsImportAllowed,
		IsExportAllowed:     &gormKeyPool.KeyPoolIsExportAllowed,
		Status:              (*cryptoutilModel.KeyPoolStatus)(&gormKeyPool.KeyPoolStatus),
	}
}

func (m *OpenapiOrmMapper) toOpenapiKeys(gormKeys *[]cryptoutilOrmService.Key) *[]cryptoutilModel.Key {
	openapiKeys := make([]cryptoutilModel.Key, len(*gormKeys))
	for i, gormKey := range *gormKeys {
		openapiKeys[i] = *m.toOpenapiKey(&gormKey)
	}
	return &openapiKeys
}

func (*OpenapiOrmMapper) toOpenapiKey(gormKey *cryptoutilOrmService.Key) *cryptoutilModel.Key {
	return &cryptoutilModel.Key{
		KeyId:        &gormKey.KeyID,
		KeyPoolId:    (*cryptoutilModel.KeyPoolId)(cryptoutilPointer.StringPtr(gormKey.KeyPoolID.String())),
		GenerateDate: (*cryptoutilModel.KeyGenerateDate)(gormKey.KeyGenerateDate),
	}
}

func (*OpenapiOrmMapper) toKeyPoolProviderEnum(openapiKeyPoolProvider *cryptoutilModel.KeyPoolProvider) *cryptoutilOrmService.KeyPoolProviderEnum {
	gormKeyPoolProvider := cryptoutilOrmService.KeyPoolProviderEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolProvider
}

func (*OpenapiOrmMapper) toKKeyPoolAlgorithmEnum(openapiKeyPoolProvider *cryptoutilModel.KeyPoolAlgorithm) *cryptoutilOrmService.KeyPoolAlgorithmEnum {
	gormKeyPoolAlgorithm := cryptoutilOrmService.KeyPoolAlgorithmEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolAlgorithm
}

func (*OpenapiOrmMapper) toKeyPoolInitialStatus(openapiKeyPoolIsImportAllowed *cryptoutilModel.KeyPoolIsImportAllowed) *cryptoutilOrmService.KeyPoolStatusEnum {
	var gormKeyPoolStatus cryptoutilOrmService.KeyPoolStatusEnum
	if *openapiKeyPoolIsImportAllowed {
		gormKeyPoolStatus = cryptoutilOrmService.KeyPoolStatusEnum("pending_import")
	} else {
		gormKeyPoolStatus = cryptoutilOrmService.KeyPoolStatusEnum("pending_generate")
	}
	return &gormKeyPoolStatus
}

// PostKeyPool

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKeyPoolSuccess(gormKeyPool *cryptoutilOrmService.KeyPool) cryptoutilServer.PostKeypoolResponseObject {
	openapiPostKeypoolResponseObject := cryptoutilServer.PostKeypool200JSONResponse(*m.toOpenapiKeyPool(gormKeyPool))
	return &openapiPostKeypoolResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyPoolError(err error) (cryptoutilServer.PostKeypoolResponseObject, error) {
	return cryptoutilServer.PostKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to insert Key Pool"}}, fmt.Errorf("failed to insert Key Pool: %w", err)
}

// GetKeyPool

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKeyPoolSuccess(gormKeyPools *[]cryptoutilOrmService.KeyPool) cryptoutilServer.GetKeypoolResponseObject {
	openapiGetKeypoolResponseObject := cryptoutilServer.GetKeypool200JSONResponse(*m.toOpenapiKeyPools(gormKeyPools))
	return &openapiGetKeypoolResponseObject
}

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKeyPoolError(err error) (cryptoutilServer.GetKeypoolResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilServer.GetKeypool404JSONResponse{HTTP404NotFound: cryptoutilModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilServer.GetKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

// PostKeyPoolKeyPoolIDKey

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKeySuccess(gormKey *cryptoutilOrmService.Key) cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject {
	openapiPostKeypoolKeyPoolIDKeyResponseObject := cryptoutilServer.PostKeypoolKeyPoolIDKey200JSONResponse(*m.toOpenapiKey(gormKey))
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyInvalidKeyPoolID(err error) (cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeySelectKeyPoolError(err error) (cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyInvalidKeyPoolStatus() (cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilModel.HTTP400BadRequest{Error: "Key Pool invalid initial state"}}, fmt.Errorf("Key Pool invalid initial state")
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyGenerateKeyMaterialError(err error) (cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return &cryptoutilServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: err.Error()}}, nil
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyError(err error) (cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

// GetKeyPoolKeyPoolIDKey

func (m *OpenapiOrmMapper) toOpenapiResponseGetKeySuccess(gormKeys *[]cryptoutilOrmService.Key) cryptoutilServer.GetKeypoolKeyPoolIDKeyResponseObject {
	openapiGetKeypoolKeyPoolIDKeyResponseObject := cryptoutilServer.GetKeypoolKeyPoolIDKey200JSONResponse(*m.toOpenapiKeys(gormKeys))
	return &openapiGetKeypoolKeyPoolIDKeyResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseGetKeyInvalidKeyPoolIDError(err error) (cryptoutilServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilServer.GetKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKeyNoKeyPoolIDFoundError(err error) (cryptoutilServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilServer.GetKeypoolKeyPoolIDKey404JSONResponse{HTTP404NotFound: cryptoutilModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKeyFindError(err error) (cryptoutilServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilModel.HTTP500InternalServerError{Error: "failed to get Keys"}}, fmt.Errorf("failed to get Keys: %w", err)
}
