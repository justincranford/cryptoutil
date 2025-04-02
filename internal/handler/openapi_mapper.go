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

func (m *openapiMapper) toOpenapiGetKeypoolResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypool400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "unexpected error"}}, fmt.Errorf("unexpected error: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeypoolKeyPoolIDKeyResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "unexpected error"}}, fmt.Errorf("unexpected error: %w", err)
}

// PostKeyPool

func (m *openapiMapper) toOpenapiInsertKeyPoolResponseSuccess(gormKeyPool *cryptoutilServiceModel.KeyPool) cryptoutilOpenapiServer.PostKeypoolResponseObject {
	openapiPostKeypoolResponseObject := cryptoutilOpenapiServer.PostKeypool200JSONResponse(*gormKeyPool)
	return &openapiPostKeypoolResponseObject
}

func (*openapiMapper) toOpenapiInsertKeyPoolResponseError(err error) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	return cryptoutilOpenapiServer.PostKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to insert Key Pool"}}, fmt.Errorf("failed to insert Key Pool: %w", err)
}

// GetKeyPool

func (m *openapiMapper) toServiceModelGetKeyPoolQueryParams(openapiGetKeyPoolQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolParams) *cryptoutilServiceModel.KeyPoolQueryParams {
	filters := cryptoutilServiceModel.KeyPoolQueryParams{
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

func (m *openapiMapper) toOpenapiSelectKeyPoolResponseSuccess(gormKeyPools []cryptoutilServiceModel.KeyPool) cryptoutilOpenapiServer.GetKeypoolResponseObject {
	openapiGetKeypoolResponseObject := cryptoutilOpenapiServer.GetKeypool200JSONResponse(gormKeyPools)
	return &openapiGetKeypoolResponseObject
}

func (m *openapiMapper) toOpenapiSelectKeyPoolResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypool400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypool404JSONResponse{HTTP404NotFound: cryptoutilServiceModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
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

func (m *openapiMapper) toServiceModelGetKeyQueryParams(openapiGetKeyQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyParams) *cryptoutilServiceModel.KeyQueryParams {
	filters := cryptoutilServiceModel.KeyQueryParams{
		Pool: openapiGetKeyQueryParamsObject.Pool,
		Id:   openapiGetKeyQueryParamsObject.Id,
		Sort: openapiGetKeyQueryParamsObject.Sort,
		Page: openapiGetKeyQueryParamsObject.Page,
		Size: openapiGetKeyQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiMapper) toOpenapiGetKeyResponseSuccess(keys []cryptoutilServiceModel.Key) cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject {
	openapiGetKeypoolKeyPoolIDKeyResponseObject := cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey200JSONResponse(keys)
	return &openapiGetKeypoolKeyPoolIDKeyResponseObject
}

func (*openapiMapper) toOpenapiGetKeyInvalidKeyPoolIDResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{Error: "Key Pool ID"}}, fmt.Errorf("Key Pool ID: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeyNoKeyPoolIDFoundResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey404JSONResponse{HTTP404NotFound: cryptoutilServiceModel.HTTP404NotFound{Error: "Key Pool not found"}}, fmt.Errorf("Key Pool not found: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to get Key Pool"}}, fmt.Errorf("failed to get Key Pool: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeyFindResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{Error: "failed to get Keys"}}, fmt.Errorf("failed to get Keys: %w", err)
}
