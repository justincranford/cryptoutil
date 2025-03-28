package service

import (
	cryptoutilService "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilRepositoryOrm "cryptoutil/internal/repository/orm"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type serviceOrmMapper struct{}

func NewMapper() *serviceOrmMapper {
	return &serviceOrmMapper{}
}

func (m *serviceOrmMapper) toGormKeyPoolInsert(openapiKeyPoolCreate *cryptoutilService.KeyPoolCreate) *cryptoutilRepositoryOrm.KeyPool {
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

func (m *serviceOrmMapper) toServiceKeyPools(gormKeyPools []cryptoutilRepositoryOrm.KeyPool) []cryptoutilService.KeyPool {
	openapiKeyPools := make([]cryptoutilService.KeyPool, len(gormKeyPools))
	for i, gormKeyPool := range gormKeyPools {
		openapiKeyPools[i] = *m.toServiceKeyPool(&gormKeyPool)
	}
	return openapiKeyPools
}

func (*serviceOrmMapper) toServiceKeyPool(gormKeyPool *cryptoutilRepositoryOrm.KeyPool) *cryptoutilService.KeyPool {
	return &cryptoutilService.KeyPool{
		Id:                  (*cryptoutilService.KeyPoolId)(&gormKeyPool.KeyPoolID),
		Name:                &gormKeyPool.KeyPoolName,
		Description:         &gormKeyPool.KeyPoolDescription,
		Algorithm:           (*cryptoutilService.KeyPoolAlgorithm)(&gormKeyPool.KeyPoolAlgorithm),
		Provider:            (*cryptoutilService.KeyPoolProvider)(&gormKeyPool.KeyPoolProvider),
		IsVersioningAllowed: &gormKeyPool.KeyPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKeyPool.KeyPoolIsImportAllowed,
		IsExportAllowed:     &gormKeyPool.KeyPoolIsExportAllowed,
		Status:              (*cryptoutilService.KeyPoolStatus)(&gormKeyPool.KeyPoolStatus),
	}
}

func (m *serviceOrmMapper) toServiceKeys(gormKeys []cryptoutilRepositoryOrm.Key) []cryptoutilService.Key {
	openapiKeys := make([]cryptoutilService.Key, len(gormKeys))
	for i, gormKey := range gormKeys {
		openapiKeys[i] = *m.toServiceKey(&gormKey)
	}
	return openapiKeys
}

func (*serviceOrmMapper) toServiceKey(gormKey *cryptoutilRepositoryOrm.Key) *cryptoutilService.Key {
	return &cryptoutilService.Key{
		KeyId:        &gormKey.KeyID,
		KeyPoolId:    (*cryptoutilService.KeyPoolId)(&gormKey.KeyPoolID),
		GenerateDate: (*cryptoutilService.KeyGenerateDate)(gormKey.KeyGenerateDate),
	}
}

func (*serviceOrmMapper) toKeyPoolProviderEnum(openapiKeyPoolProvider *cryptoutilService.KeyPoolProvider) *cryptoutilRepositoryOrm.KeyPoolProviderEnum {
	gormKeyPoolProvider := cryptoutilRepositoryOrm.KeyPoolProviderEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolProvider
}

func (*serviceOrmMapper) toKKeyPoolAlgorithmEnum(openapiKeyPoolProvider *cryptoutilService.KeyPoolAlgorithm) *cryptoutilRepositoryOrm.KeyPoolAlgorithmEnum {
	gormKeyPoolAlgorithm := cryptoutilRepositoryOrm.KeyPoolAlgorithmEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolAlgorithm
}

func (*serviceOrmMapper) toKeyPoolInitialStatus(openapiKeyPoolIsImportAllowed *cryptoutilService.KeyPoolIsImportAllowed) *cryptoutilRepositoryOrm.KeyPoolStatusEnum {
	var gormKeyPoolStatus cryptoutilRepositoryOrm.KeyPoolStatusEnum
	if *openapiKeyPoolIsImportAllowed {
		gormKeyPoolStatus = cryptoutilRepositoryOrm.KeyPoolStatusEnum("pending_import")
	} else {
		gormKeyPoolStatus = cryptoutilRepositoryOrm.KeyPoolStatusEnum("pending_generate")
	}
	return &gormKeyPoolStatus
}

// PostKeyPool

func (m *serviceOrmMapper) toOpenapiInsertKeyPoolResponseSuccess(gormKeyPool *cryptoutilRepositoryOrm.KeyPool) cryptoutilOpenapiServer.PostKeypoolResponseObject {
	openapiPostKeypoolResponseObject := cryptoutilOpenapiServer.PostKeypool200JSONResponse(*m.toServiceKeyPool(gormKeyPool))
	return &openapiPostKeypoolResponseObject
}

func (*serviceOrmMapper) toOpenapiInsertKeyPoolResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilService.HTTP500InternalServerError{Error: "failed to insert Key Pool"}}, fmt.Errorf("failed to insert Key Pool: %w", err)
}

// GetKeyPool

func (m *serviceOrmMapper) toOpenapiSelectKeyPoolResponseSuccess(gormKeyPools []cryptoutilRepositoryOrm.KeyPool) cryptoutilOpenapiServer.GetKeypoolResponseObject {
	openapiGetKeypoolResponseObject := cryptoutilOpenapiServer.GetKeypool200JSONResponse(m.toServiceKeyPools(gormKeyPools))
	return &openapiGetKeypoolResponseObject
}

func (m *serviceOrmMapper) toOpenapiSelectKeyPoolResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypool400JSONResponse{HTTP400BadRequest: cryptoutilService.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypool404JSONResponse{HTTP404NotFound: cryptoutilService.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilService.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

// PostKeyPoolKeyPoolIDKey

func (m *serviceOrmMapper) toOpenapiInsertKeySuccessResponseError(gormKey *cryptoutilRepositoryOrm.Key) cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject {
	openapiPostKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey200JSONResponse(*m.toServiceKey(gormKey))
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject
}

func (*serviceOrmMapper) toOpenapiInsertKeyInvalidKeyPoolIDResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilService.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (*serviceOrmMapper) toOpenapiInsertKeySelectKeyPoolResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilService.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

func (*serviceOrmMapper) toOpenapiInsertKeyInvalidKeyPoolStatusResponseError() (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilService.HTTP400BadRequest{Error: "Key Pool invalid initial state"}}, fmt.Errorf("Key Pool invalid initial state")
}

func (*serviceOrmMapper) toOpenapiInsertKeyGenerateKeyMaterialResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return &cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilService.HTTP500InternalServerError{Error: err.Error()}}, nil
}

func (*serviceOrmMapper) toOpenapiInsertKeyResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilService.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

// GetKeyPoolKeyPoolIDKey

func (m *serviceOrmMapper) toOpenapiGetKeyResponseSuccess(gormKeys []cryptoutilRepositoryOrm.Key) cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject {
	openapiGetKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey200JSONResponse(m.toServiceKeys(gormKeys))
	return &openapiGetKeypoolKeyPoolIDKeyResponseObject
}

func (*serviceOrmMapper) toOpenapiGetKeyInvalidKeyPoolIDResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilService.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (m *serviceOrmMapper) toOpenapiGetKeyNoKeyPoolIDFoundResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey404JSONResponse{HTTP404NotFound: cryptoutilService.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilService.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

func (m *serviceOrmMapper) toOpenapiGetKeyFindResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilService.HTTP500InternalServerError{Error: "failed to get Keys"}}, fmt.Errorf("failed to get Keys: %w", err)
}
