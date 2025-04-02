package handler

import (
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type openapiMapper struct{}

func NewMapper() *openapiMapper {
	return &openapiMapper{}
}

// GetKeyPool

func (m *openapiMapper) toServiceModelGetKeyPoolQueryParams(openapiGetKeyPoolQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolsParams) *cryptoutilServiceModel.KeyPoolsQueryParams {
	filters := cryptoutilServiceModel.KeyPoolsQueryParams{
		Id:                openapiGetKeyPoolQueryParamsObject.Id,
		Name:              openapiGetKeyPoolQueryParamsObject.Name,
		Provider:          openapiGetKeyPoolQueryParamsObject.Provider,
		Algorithm:         openapiGetKeyPoolQueryParamsObject.Algorithm,
		VersioningAllowed: openapiGetKeyPoolQueryParamsObject.VersioningAllowed,
		ImportAllowed:     openapiGetKeyPoolQueryParamsObject.ImportAllowed,
		ExportAllowed:     openapiGetKeyPoolQueryParamsObject.ExportAllowed,
		Status:            openapiGetKeyPoolQueryParamsObject.Status,
		Sort:              openapiGetKeyPoolQueryParamsObject.Sort,
		Page:              openapiGetKeyPoolQueryParamsObject.Page,
		Size:              openapiGetKeyPoolQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiMapper) toOpenapiSelectKeyPoolResponseSuccess(gormKeyPools []cryptoutilServiceModel.KeyPool) cryptoutilOpenapiServer.GetKeypoolsResponseObject {
	openapiGetKeypoolResponseObject := cryptoutilOpenapiServer.GetKeypools200JSONResponse(gormKeyPools)
	return &openapiGetKeypoolResponseObject
}

func (m *openapiMapper) toOpenapiSelectKeyPoolResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolsResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypools400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypools404JSONResponse{HTTP404NotFound: cryptoutilServiceModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypools500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeypoolResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolsResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypools400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypools500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "unexpected error"}}, fmt.Errorf("unexpected error: %w", err)
}

// PostKeyPool

func (m *openapiMapper) toOpenapiInsertKeyPoolResponseSuccess(gormKeyPool *cryptoutilServiceModel.KeyPool) cryptoutilOpenapiServer.PostKeypoolResponseObject {
	openapiPostKeypoolResponseObject := cryptoutilOpenapiServer.PostKeypool200JSONResponse(*gormKeyPool)
	return &openapiPostKeypoolResponseObject
}

func (*openapiMapper) toOpenapiInsertKeyPoolResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to insert Key Pool"}}, fmt.Errorf("failed to insert Key Pool: %w", err)
}

// PostKeyPoolKeyPoolIDKey

func (m *openapiMapper) toOpenapiInsertKeySuccessResponseError(key *cryptoutilServiceModel.Key) cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject {
	openapiPostKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey200JSONResponse(*key)
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject
}

func (*openapiMapper) toOpenapiInsertKeyInvalidKeyPoolIDResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (*openapiMapper) toOpenapiInsertKeySelectKeyPoolResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

func (*openapiMapper) toOpenapiInsertKeyInvalidKeyPoolStatusResponseError() (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Key Pool invalid initial state"}}, fmt.Errorf("Key Pool invalid initial state")
}

func (*openapiMapper) toOpenapiInsertKeyGenerateKeyMaterialResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return &cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: err.Error()}}, nil
}

func (*openapiMapper) toOpenapiInsertKeyResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to insert Key"}}, fmt.Errorf("failed to insert Key: %w", err)
}

// GetKeyPoolKeyPoolIDKey

func (m *openapiMapper) toServiceModelGetKeyPoolKeysQueryParams(openapiGetKeyQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysParams) *cryptoutilServiceModel.KeyPoolKeysQueryParams {
	filters := cryptoutilServiceModel.KeyPoolKeysQueryParams{
		Id:   openapiGetKeyQueryParamsObject.Id,
		Sort: openapiGetKeyQueryParamsObject.Sort,
		Page: openapiGetKeyQueryParamsObject.Page,
		Size: openapiGetKeyQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiMapper) toServiceModelGetKeysQueryParams(openapiGetKeyQueryParamsObject *cryptoutilOpenapiServer.GetKeysParams) *cryptoutilServiceModel.KeysQueryParams {
	filters := cryptoutilServiceModel.KeysQueryParams{
		Pool: openapiGetKeyQueryParamsObject.Pool,
		Id:   openapiGetKeyQueryParamsObject.Id,
		Sort: openapiGetKeyQueryParamsObject.Sort,
		Page: openapiGetKeyQueryParamsObject.Page,
		Size: openapiGetKeyQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiMapper) toOpenapiGetKeyResponseSuccess(keys []cryptoutilServiceModel.Key) cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject {
	openapiGetKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys200JSONResponse(keys)
	return &openapiGetKeypoolKeyPoolIDKeyResponseObject
}

func (*openapiMapper) toOpenapiGetKeyInvalidKeyPoolIDResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeypoolKeyPoolIDKeyResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "unexpected error"}}, fmt.Errorf("unexpected error: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeyNoKeyPoolIDFoundResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys404JSONResponse{HTTP404NotFound: cryptoutilServiceModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeyFindResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to get Keys"}}, fmt.Errorf("failed to get Keys: %w", err)
}
