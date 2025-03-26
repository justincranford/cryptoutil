package service

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilPointer "cryptoutil/internal/pointer"
	cryptoutilRepositoryOrm "cryptoutil/internal/repository/orm"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type OpenapiOrmMapper struct{}

func NewMapper() *OpenapiOrmMapper {
	return &OpenapiOrmMapper{}
}

func (m *OpenapiOrmMapper) toGormKeyPoolInsert(openapiKeyPoolCreate *cryptoutilOpenapiModel.KeyPoolCreate) *cryptoutilRepositoryOrm.KeyPool {
	return &cryptoutilRepositoryOrm.KeyPool{
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

func (m *OpenapiOrmMapper) toOpenapiKeyPools(gormKeyPools *[]cryptoutilRepositoryOrm.KeyPool) *[]cryptoutilOpenapiModel.KeyPool {
	openapiKeyPools := make([]cryptoutilOpenapiModel.KeyPool, len(*gormKeyPools))
	for i, gormKeyPool := range *gormKeyPools {
		openapiKeyPools[i] = *m.toOpenapiKeyPool(&gormKeyPool)
	}
	return &openapiKeyPools
}

func (*OpenapiOrmMapper) toOpenapiKeyPool(gormKeyPool *cryptoutilRepositoryOrm.KeyPool) *cryptoutilOpenapiModel.KeyPool {
	return &cryptoutilOpenapiModel.KeyPool{
		Id:                  (*cryptoutilOpenapiModel.KeyPoolId)(cryptoutilPointer.StringPtr(gormKeyPool.KeyPoolID.String())),
		Name:                &gormKeyPool.KeyPoolName,
		Description:         &gormKeyPool.KeyPoolDescription,
		Algorithm:           (*cryptoutilOpenapiModel.KeyPoolAlgorithm)(&gormKeyPool.KeyPoolAlgorithm),
		Provider:            (*cryptoutilOpenapiModel.KeyPoolProvider)(&gormKeyPool.KeyPoolProvider),
		IsVersioningAllowed: &gormKeyPool.KeyPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKeyPool.KeyPoolIsImportAllowed,
		IsExportAllowed:     &gormKeyPool.KeyPoolIsExportAllowed,
		Status:              (*cryptoutilOpenapiModel.KeyPoolStatus)(&gormKeyPool.KeyPoolStatus),
	}
}

func (m *OpenapiOrmMapper) toOpenapiKeys(gormKeys *[]cryptoutilRepositoryOrm.Key) *[]cryptoutilOpenapiModel.Key {
	openapiKeys := make([]cryptoutilOpenapiModel.Key, len(*gormKeys))
	for i, gormKey := range *gormKeys {
		openapiKeys[i] = *m.toOpenapiKey(&gormKey)
	}
	return &openapiKeys
}

func (*OpenapiOrmMapper) toOpenapiKey(gormKey *cryptoutilRepositoryOrm.Key) *cryptoutilOpenapiModel.Key {
	return &cryptoutilOpenapiModel.Key{
		KeyId:        &gormKey.KeyID,
		KeyPoolId:    (*cryptoutilOpenapiModel.KeyPoolId)(cryptoutilPointer.StringPtr(gormKey.KeyPoolID.String())),
		GenerateDate: (*cryptoutilOpenapiModel.KeyGenerateDate)(gormKey.KeyGenerateDate),
	}
}

func (*OpenapiOrmMapper) toKeyPoolProviderEnum(openapiKeyPoolProvider *cryptoutilOpenapiModel.KeyPoolProvider) *cryptoutilRepositoryOrm.KeyPoolProviderEnum {
	gormKeyPoolProvider := cryptoutilRepositoryOrm.KeyPoolProviderEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolProvider
}

func (*OpenapiOrmMapper) toKKeyPoolAlgorithmEnum(openapiKeyPoolProvider *cryptoutilOpenapiModel.KeyPoolAlgorithm) *cryptoutilRepositoryOrm.KeyPoolAlgorithmEnum {
	gormKeyPoolAlgorithm := cryptoutilRepositoryOrm.KeyPoolAlgorithmEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolAlgorithm
}

func (*OpenapiOrmMapper) toKeyPoolInitialStatus(openapiKeyPoolIsImportAllowed *cryptoutilOpenapiModel.KeyPoolIsImportAllowed) *cryptoutilRepositoryOrm.KeyPoolStatusEnum {
	var gormKeyPoolStatus cryptoutilRepositoryOrm.KeyPoolStatusEnum
	if *openapiKeyPoolIsImportAllowed {
		gormKeyPoolStatus = cryptoutilRepositoryOrm.KeyPoolStatusEnum("pending_import")
	} else {
		gormKeyPoolStatus = cryptoutilRepositoryOrm.KeyPoolStatusEnum("pending_generate")
	}
	return &gormKeyPoolStatus
}

// PostKeyPool

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKeyPoolSuccess(gormKeyPool *cryptoutilRepositoryOrm.KeyPool) cryptoutilOpenapiServer.PostKeypoolResponseObject {
	openapiPostKeypoolResponseObject := cryptoutilOpenapiServer.PostKeypool200JSONResponse(*m.toOpenapiKeyPool(gormKeyPool))
	return &openapiPostKeypoolResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyPoolError(err error) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilOpenapiModel.HTTP500InternalServerError{Error: "failed to insert Key Pool"}}, fmt.Errorf("failed to insert Key Pool: %w", err)
}

// GetKeyPool

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKeyPoolSuccess(gormKeyPools *[]cryptoutilRepositoryOrm.KeyPool) cryptoutilOpenapiServer.GetKeypoolResponseObject {
	openapiGetKeypoolResponseObject := cryptoutilOpenapiServer.GetKeypool200JSONResponse(*m.toOpenapiKeyPools(gormKeyPools))
	return &openapiGetKeypoolResponseObject
}

func (m *OpenapiOrmMapper) toOpenapiResponseSelectKeyPoolError(err error) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypool404JSONResponse{HTTP404NotFound: cryptoutilOpenapiModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilOpenapiModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

// PostKeyPoolKeyPoolIDKey

func (m *OpenapiOrmMapper) toOpenapiResponseInsertKeySuccess(gormKey *cryptoutilRepositoryOrm.Key) cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject {
	openapiPostKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey200JSONResponse(*m.toOpenapiKey(gormKey))
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyInvalidKeyPoolID(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilOpenapiModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeySelectKeyPoolError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilOpenapiModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyInvalidKeyPoolStatus() (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilOpenapiModel.HTTP400BadRequest{Error: "Key Pool invalid initial state"}}, fmt.Errorf("Key Pool invalid initial state")
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyGenerateKeyMaterialError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return &cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilOpenapiModel.HTTP500InternalServerError{Error: err.Error()}}, nil
}

func (*OpenapiOrmMapper) toOpenapiResponseInsertKeyError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilOpenapiModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

// GetKeyPoolKeyPoolIDKey

func (m *OpenapiOrmMapper) toOpenapiResponseGetKeySuccess(gormKeys *[]cryptoutilRepositoryOrm.Key) cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject {
	openapiGetKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey200JSONResponse(*m.toOpenapiKeys(gormKeys))
	return &openapiGetKeypoolKeyPoolIDKeyResponseObject
}

func (*OpenapiOrmMapper) toOpenapiResponseGetKeyInvalidKeyPoolIDError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilOpenapiModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKeyNoKeyPoolIDFoundError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey404JSONResponse{HTTP404NotFound: cryptoutilOpenapiModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilOpenapiModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

func (m *OpenapiOrmMapper) toOpenapiResponseGetKeyFindError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilOpenapiModel.HTTP500InternalServerError{Error: "failed to get Keys"}}, fmt.Errorf("failed to get Keys: %w", err)
}
