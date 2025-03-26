package service

import (
	cryptoutilBusinessModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilPointer "cryptoutil/internal/pointer"
	cryptoutilRepositoryOrm "cryptoutil/internal/repository/orm"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type openapiOrmMapper struct{}

func NewMapper() *openapiOrmMapper {
	return &openapiOrmMapper{}
}

func (m *openapiOrmMapper) toGormKeyPoolInsert(openapiKeyPoolCreate *cryptoutilBusinessModel.KeyPoolCreate) *cryptoutilRepositoryOrm.KeyPool {
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

func (m *openapiOrmMapper) toOpenapiKeyPools(gormKeyPools *[]cryptoutilRepositoryOrm.KeyPool) *[]cryptoutilBusinessModel.KeyPool {
	openapiKeyPools := make([]cryptoutilBusinessModel.KeyPool, len(*gormKeyPools))
	for i, gormKeyPool := range *gormKeyPools {
		openapiKeyPools[i] = *m.toOpenapiKeyPool(&gormKeyPool)
	}
	return &openapiKeyPools
}

func (*openapiOrmMapper) toOpenapiKeyPool(gormKeyPool *cryptoutilRepositoryOrm.KeyPool) *cryptoutilBusinessModel.KeyPool {
	return &cryptoutilBusinessModel.KeyPool{
		Id:                  (*cryptoutilBusinessModel.KeyPoolId)(cryptoutilPointer.StringPtr(gormKeyPool.KeyPoolID.String())),
		Name:                &gormKeyPool.KeyPoolName,
		Description:         &gormKeyPool.KeyPoolDescription,
		Algorithm:           (*cryptoutilBusinessModel.KeyPoolAlgorithm)(&gormKeyPool.KeyPoolAlgorithm),
		Provider:            (*cryptoutilBusinessModel.KeyPoolProvider)(&gormKeyPool.KeyPoolProvider),
		IsVersioningAllowed: &gormKeyPool.KeyPoolIsVersioningAllowed,
		IsImportAllowed:     &gormKeyPool.KeyPoolIsImportAllowed,
		IsExportAllowed:     &gormKeyPool.KeyPoolIsExportAllowed,
		Status:              (*cryptoutilBusinessModel.KeyPoolStatus)(&gormKeyPool.KeyPoolStatus),
	}
}

func (m *openapiOrmMapper) toOpenapiKeys(gormKeys *[]cryptoutilRepositoryOrm.Key) *[]cryptoutilBusinessModel.Key {
	openapiKeys := make([]cryptoutilBusinessModel.Key, len(*gormKeys))
	for i, gormKey := range *gormKeys {
		openapiKeys[i] = *m.toOpenapiKey(&gormKey)
	}
	return &openapiKeys
}

func (*openapiOrmMapper) toOpenapiKey(gormKey *cryptoutilRepositoryOrm.Key) *cryptoutilBusinessModel.Key {
	return &cryptoutilBusinessModel.Key{
		KeyId:        &gormKey.KeyID,
		KeyPoolId:    (*cryptoutilBusinessModel.KeyPoolId)(cryptoutilPointer.StringPtr(gormKey.KeyPoolID.String())),
		GenerateDate: (*cryptoutilBusinessModel.KeyGenerateDate)(gormKey.KeyGenerateDate),
	}
}

func (*openapiOrmMapper) toKeyPoolProviderEnum(openapiKeyPoolProvider *cryptoutilBusinessModel.KeyPoolProvider) *cryptoutilRepositoryOrm.KeyPoolProviderEnum {
	gormKeyPoolProvider := cryptoutilRepositoryOrm.KeyPoolProviderEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolProvider
}

func (*openapiOrmMapper) toKKeyPoolAlgorithmEnum(openapiKeyPoolProvider *cryptoutilBusinessModel.KeyPoolAlgorithm) *cryptoutilRepositoryOrm.KeyPoolAlgorithmEnum {
	gormKeyPoolAlgorithm := cryptoutilRepositoryOrm.KeyPoolAlgorithmEnum(*openapiKeyPoolProvider)
	return &gormKeyPoolAlgorithm
}

func (*openapiOrmMapper) toKeyPoolInitialStatus(openapiKeyPoolIsImportAllowed *cryptoutilBusinessModel.KeyPoolIsImportAllowed) *cryptoutilRepositoryOrm.KeyPoolStatusEnum {
	var gormKeyPoolStatus cryptoutilRepositoryOrm.KeyPoolStatusEnum
	if *openapiKeyPoolIsImportAllowed {
		gormKeyPoolStatus = cryptoutilRepositoryOrm.KeyPoolStatusEnum("pending_import")
	} else {
		gormKeyPoolStatus = cryptoutilRepositoryOrm.KeyPoolStatusEnum("pending_generate")
	}
	return &gormKeyPoolStatus
}

// PostKeyPool

func (m *openapiOrmMapper) toOpenapiInsertKeyPoolResponseSuccess(gormKeyPool *cryptoutilRepositoryOrm.KeyPool) cryptoutilOpenapiServer.PostKeypoolResponseObject {
	openapiPostKeypoolResponseObject := cryptoutilOpenapiServer.PostKeypool200JSONResponse(*m.toOpenapiKeyPool(gormKeyPool))
	return &openapiPostKeypoolResponseObject
}

func (*openapiOrmMapper) toOpenapiInsertKeyPoolResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "failed to insert Key Pool"}}, fmt.Errorf("failed to insert Key Pool: %w", err)
}

// GetKeyPool

func (m *openapiOrmMapper) toOpenapiSelectKeyPoolResponseSuccess(gormKeyPools *[]cryptoutilRepositoryOrm.KeyPool) cryptoutilOpenapiServer.GetKeypoolResponseObject {
	openapiGetKeypoolResponseObject := cryptoutilOpenapiServer.GetKeypool200JSONResponse(*m.toOpenapiKeyPools(gormKeyPools))
	return &openapiGetKeypoolResponseObject
}

func (m *openapiOrmMapper) toOpenapiSelectKeyPoolResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypool400JSONResponse{HTTP400BadRequest: cryptoutilBusinessModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypool404JSONResponse{HTTP404NotFound: cryptoutilBusinessModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

// PostKeyPoolKeyPoolIDKey

func (m *openapiOrmMapper) toOpenapiInsertKeySuccessResponseError(gormKey *cryptoutilRepositoryOrm.Key) cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject {
	openapiPostKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey200JSONResponse(*m.toOpenapiKey(gormKey))
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject
}

func (*openapiOrmMapper) toOpenapiInsertKeyInvalidKeyPoolIDResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilBusinessModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (*openapiOrmMapper) toOpenapiInsertKeySelectKeyPoolResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

func (*openapiOrmMapper) toOpenapiInsertKeyInvalidKeyPoolStatusResponseError() (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilBusinessModel.HTTP400BadRequest{Error: "Key Pool invalid initial state"}}, fmt.Errorf("Key Pool invalid initial state")
}

func (*openapiOrmMapper) toOpenapiInsertKeyGenerateKeyMaterialResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return &cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: err.Error()}}, nil
}

func (*openapiOrmMapper) toOpenapiInsertKeyResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

// GetKeyPoolKeyPoolIDKey

func (m *openapiOrmMapper) toOpenapiGetKeyResponseSuccess(gormKeys *[]cryptoutilRepositoryOrm.Key) cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject {
	openapiGetKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey200JSONResponse(*m.toOpenapiKeys(gormKeys))
	return &openapiGetKeypoolKeyPoolIDKeyResponseObject
}

func (*openapiOrmMapper) toOpenapiGetKeyInvalidKeyPoolIDResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilBusinessModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (m *openapiOrmMapper) toOpenapiGetKeyNoKeyPoolIDFoundResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey404JSONResponse{HTTP404NotFound: cryptoutilBusinessModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

func (m *openapiOrmMapper) toOpenapiGetKeyFindResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "failed to get Keys"}}, fmt.Errorf("failed to get Keys: %w", err)
}
