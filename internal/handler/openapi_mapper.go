package handler

import (
	cryptoutilBusinessModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type openapiMapper struct{}

func NewMapper() *openapiMapper {
	return &openapiMapper{}
}

func (m *openapiMapper) toOpenapiGetKeypoolResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypool400JSONResponse{HTTP400BadRequest: cryptoutilBusinessModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypool500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "unexpected error"}}, fmt.Errorf("unexpected error: %w", err)
}

func (m *openapiMapper) toOpenapiGetKeypoolKeyPoolIDKeyResponseError(err error) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if errors.Is(err, fiber.ErrBadRequest) {
		return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: cryptoutilBusinessModel.HTTP400BadRequest{Error: "Bad input"}}, fmt.Errorf("Bad input: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: cryptoutilBusinessModel.HTTP500InternalServerError{Error: "unexpected error"}}, fmt.Errorf("unexpected error: %w", err)
}
