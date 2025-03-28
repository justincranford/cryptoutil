// Package server provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	externalRef0 "cryptoutil/internal/openapi/model"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/oapi-codegen/runtime"
)

// GetKeypoolParams defines parameters for GetKeypool.
type GetKeypoolParams struct {
	Filter *externalRef0.QueryParamFilter `form:"filter,omitempty" json:"filter,omitempty"`
	Sort   *externalRef0.QueryParamSort   `form:"sort,omitempty" json:"sort,omitempty"`
	Page   *externalRef0.QueryParamPage   `form:"page,omitempty" json:"page,omitempty"`
}

// GetKeypoolKeyPoolIDKeyParams defines parameters for GetKeypoolKeyPoolIDKey.
type GetKeypoolKeyPoolIDKeyParams struct {
	Filter *externalRef0.QueryParamFilter `form:"filter,omitempty" json:"filter,omitempty"`
	Sort   *externalRef0.QueryParamSort   `form:"sort,omitempty" json:"sort,omitempty"`
	Page   *externalRef0.QueryParamPage   `form:"page,omitempty" json:"page,omitempty"`
}

// PostKeypoolJSONRequestBody defines body for PostKeypool for application/json ContentType.
type PostKeypoolJSONRequestBody = externalRef0.KeyPoolCreate

// PostKeypoolKeyPoolIDKeyJSONRequestBody defines body for PostKeypoolKeyPoolIDKey for application/json ContentType.
type PostKeypoolKeyPoolIDKeyJSONRequestBody = externalRef0.KeyGenerate

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// List all Key Pools. Supports optional filtering, sorting, and paging.
	// (GET /keypool)
	GetKeypool(c *fiber.Ctx, params GetKeypoolParams) error
	// Create a new Key Pool.
	// (POST /keypool)
	PostKeypool(c *fiber.Ctx) error
	// List all Keys in Key Pool. Supports optional filtering, sorting, and paging.
	// (GET /keypool/{keyPoolID}/key)
	GetKeypoolKeyPoolIDKey(c *fiber.Ctx, keyPoolID externalRef0.KeyPoolId, params GetKeypoolKeyPoolIDKeyParams) error
	// Generate a new Key in a Key Pool.
	// (POST /keypool/{keyPoolID}/key)
	PostKeypoolKeyPoolIDKey(c *fiber.Ctx, keyPoolID externalRef0.KeyPoolId) error
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

type MiddlewareFunc fiber.Handler

// GetKeypool operation middleware
func (siw *ServerInterfaceWrapper) GetKeypool(c *fiber.Ctx) error {

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetKeypoolParams

	var query url.Values
	query, err = url.ParseQuery(string(c.Request().URI().QueryString()))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for query string: %w", err).Error())
	}

	// ------------- Optional query parameter "filter" -------------

	err = runtime.BindQueryParameter("form", true, false, "filter", query, &params.Filter)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter filter: %w", err).Error())
	}

	// ------------- Optional query parameter "sort" -------------

	err = runtime.BindQueryParameter("form", true, false, "sort", query, &params.Sort)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter sort: %w", err).Error())
	}

	// ------------- Optional query parameter "page" -------------

	err = runtime.BindQueryParameter("form", true, false, "page", query, &params.Page)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter page: %w", err).Error())
	}

	return siw.Handler.GetKeypool(c, params)
}

// PostKeypool operation middleware
func (siw *ServerInterfaceWrapper) PostKeypool(c *fiber.Ctx) error {

	return siw.Handler.PostKeypool(c)
}

// GetKeypoolKeyPoolIDKey operation middleware
func (siw *ServerInterfaceWrapper) GetKeypoolKeyPoolIDKey(c *fiber.Ctx) error {

	var err error

	// ------------- Path parameter "keyPoolID" -------------
	var keyPoolID externalRef0.KeyPoolId

	err = runtime.BindStyledParameterWithOptions("simple", "keyPoolID", c.Params("keyPoolID"), &keyPoolID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter keyPoolID: %w", err).Error())
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetKeypoolKeyPoolIDKeyParams

	var query url.Values
	query, err = url.ParseQuery(string(c.Request().URI().QueryString()))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for query string: %w", err).Error())
	}

	// ------------- Optional query parameter "filter" -------------

	err = runtime.BindQueryParameter("form", true, false, "filter", query, &params.Filter)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter filter: %w", err).Error())
	}

	// ------------- Optional query parameter "sort" -------------

	err = runtime.BindQueryParameter("form", true, false, "sort", query, &params.Sort)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter sort: %w", err).Error())
	}

	// ------------- Optional query parameter "page" -------------

	err = runtime.BindQueryParameter("form", true, false, "page", query, &params.Page)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter page: %w", err).Error())
	}

	return siw.Handler.GetKeypoolKeyPoolIDKey(c, keyPoolID, params)
}

// PostKeypoolKeyPoolIDKey operation middleware
func (siw *ServerInterfaceWrapper) PostKeypoolKeyPoolIDKey(c *fiber.Ctx) error {

	var err error

	// ------------- Path parameter "keyPoolID" -------------
	var keyPoolID externalRef0.KeyPoolId

	err = runtime.BindStyledParameterWithOptions("simple", "keyPoolID", c.Params("keyPoolID"), &keyPoolID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("Invalid format for parameter keyPoolID: %w", err).Error())
	}

	return siw.Handler.PostKeypoolKeyPoolIDKey(c, keyPoolID)
}

// FiberServerOptions provides options for the Fiber server.
type FiberServerOptions struct {
	BaseURL     string
	Middlewares []MiddlewareFunc
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router fiber.Router, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, FiberServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router fiber.Router, si ServerInterface, options FiberServerOptions) {
	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	for _, m := range options.Middlewares {
		router.Use(fiber.Handler(m))
	}

	router.Get(options.BaseURL+"/keypool", wrapper.GetKeypool)

	router.Post(options.BaseURL+"/keypool", wrapper.PostKeypool)

	router.Get(options.BaseURL+"/keypool/:keyPoolID/key", wrapper.GetKeypoolKeyPoolIDKey)

	router.Post(options.BaseURL+"/keypool/:keyPoolID/key", wrapper.PostKeypoolKeyPoolIDKey)

}

type GetKeypoolRequestObject struct {
	Params GetKeypoolParams
}

type GetKeypoolResponseObject interface {
	VisitGetKeypoolResponse(ctx *fiber.Ctx) error
}

type GetKeypool200JSONResponse []externalRef0.KeyPool

func (response GetKeypool200JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(200)

	return ctx.JSON(&response)
}

type GetKeypool400JSONResponse struct{ externalRef0.HTTP400BadRequest }

func (response GetKeypool400JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(400)

	return ctx.JSON(&response)
}

type GetKeypool401JSONResponse struct {
	externalRef0.HTTP401Unauthorized
}

func (response GetKeypool401JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(401)

	return ctx.JSON(&response)
}

type GetKeypool403JSONResponse struct{ externalRef0.HTTP403Forbidden }

func (response GetKeypool403JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(403)

	return ctx.JSON(&response)
}

type GetKeypool404JSONResponse struct{ externalRef0.HTTP404NotFound }

func (response GetKeypool404JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(404)

	return ctx.JSON(&response)
}

type GetKeypool429JSONResponse struct {
	externalRef0.HTTP429TooManyRequests
}

func (response GetKeypool429JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(429)

	return ctx.JSON(&response)
}

type GetKeypool500JSONResponse struct {
	externalRef0.HTTP500InternalServerError
}

func (response GetKeypool500JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)

	return ctx.JSON(&response)
}

type GetKeypool502JSONResponse struct{ externalRef0.HTTP502BadGateway }

func (response GetKeypool502JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(502)

	return ctx.JSON(&response)
}

type GetKeypool503JSONResponse struct {
	externalRef0.HTTP503ServiceUnavailable
}

func (response GetKeypool503JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(503)

	return ctx.JSON(&response)
}

type GetKeypool504JSONResponse struct {
	externalRef0.HTTP504GatewayTimeout
}

func (response GetKeypool504JSONResponse) VisitGetKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(504)

	return ctx.JSON(&response)
}

type PostKeypoolRequestObject struct {
	Body *PostKeypoolJSONRequestBody
}

type PostKeypoolResponseObject interface {
	VisitPostKeypoolResponse(ctx *fiber.Ctx) error
}

type PostKeypool200JSONResponse externalRef0.KeyPool

func (response PostKeypool200JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(200)

	return ctx.JSON(&response)
}

type PostKeypool400JSONResponse struct{ externalRef0.HTTP400BadRequest }

func (response PostKeypool400JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(400)

	return ctx.JSON(&response)
}

type PostKeypool401JSONResponse struct {
	externalRef0.HTTP401Unauthorized
}

func (response PostKeypool401JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(401)

	return ctx.JSON(&response)
}

type PostKeypool403JSONResponse struct{ externalRef0.HTTP403Forbidden }

func (response PostKeypool403JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(403)

	return ctx.JSON(&response)
}

type PostKeypool404JSONResponse struct{ externalRef0.HTTP404NotFound }

func (response PostKeypool404JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(404)

	return ctx.JSON(&response)
}

type PostKeypool429JSONResponse struct {
	externalRef0.HTTP429TooManyRequests
}

func (response PostKeypool429JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(429)

	return ctx.JSON(&response)
}

type PostKeypool500JSONResponse struct {
	externalRef0.HTTP500InternalServerError
}

func (response PostKeypool500JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)

	return ctx.JSON(&response)
}

type PostKeypool502JSONResponse struct{ externalRef0.HTTP502BadGateway }

func (response PostKeypool502JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(502)

	return ctx.JSON(&response)
}

type PostKeypool503JSONResponse struct {
	externalRef0.HTTP503ServiceUnavailable
}

func (response PostKeypool503JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(503)

	return ctx.JSON(&response)
}

type PostKeypool504JSONResponse struct {
	externalRef0.HTTP504GatewayTimeout
}

func (response PostKeypool504JSONResponse) VisitPostKeypoolResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(504)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKeyRequestObject struct {
	KeyPoolID externalRef0.KeyPoolId `json:"keyPoolID"`
	Params    GetKeypoolKeyPoolIDKeyParams
}

type GetKeypoolKeyPoolIDKeyResponseObject interface {
	VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error
}

type GetKeypoolKeyPoolIDKey200JSONResponse []externalRef0.Key

func (response GetKeypoolKeyPoolIDKey200JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(200)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey400JSONResponse struct{ externalRef0.HTTP400BadRequest }

func (response GetKeypoolKeyPoolIDKey400JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(400)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey401JSONResponse struct {
	externalRef0.HTTP401Unauthorized
}

func (response GetKeypoolKeyPoolIDKey401JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(401)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey403JSONResponse struct{ externalRef0.HTTP403Forbidden }

func (response GetKeypoolKeyPoolIDKey403JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(403)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey404JSONResponse struct{ externalRef0.HTTP404NotFound }

func (response GetKeypoolKeyPoolIDKey404JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(404)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey429JSONResponse struct {
	externalRef0.HTTP429TooManyRequests
}

func (response GetKeypoolKeyPoolIDKey429JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(429)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey500JSONResponse struct {
	externalRef0.HTTP500InternalServerError
}

func (response GetKeypoolKeyPoolIDKey500JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey502JSONResponse struct{ externalRef0.HTTP502BadGateway }

func (response GetKeypoolKeyPoolIDKey502JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(502)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey503JSONResponse struct {
	externalRef0.HTTP503ServiceUnavailable
}

func (response GetKeypoolKeyPoolIDKey503JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(503)

	return ctx.JSON(&response)
}

type GetKeypoolKeyPoolIDKey504JSONResponse struct {
	externalRef0.HTTP504GatewayTimeout
}

func (response GetKeypoolKeyPoolIDKey504JSONResponse) VisitGetKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(504)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKeyRequestObject struct {
	KeyPoolID externalRef0.KeyPoolId `json:"keyPoolID"`
	Body      *PostKeypoolKeyPoolIDKeyJSONRequestBody
}

type PostKeypoolKeyPoolIDKeyResponseObject interface {
	VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error
}

type PostKeypoolKeyPoolIDKey200JSONResponse externalRef0.Key

func (response PostKeypoolKeyPoolIDKey200JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(200)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey400JSONResponse struct{ externalRef0.HTTP400BadRequest }

func (response PostKeypoolKeyPoolIDKey400JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(400)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey401JSONResponse struct {
	externalRef0.HTTP401Unauthorized
}

func (response PostKeypoolKeyPoolIDKey401JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(401)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey403JSONResponse struct{ externalRef0.HTTP403Forbidden }

func (response PostKeypoolKeyPoolIDKey403JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(403)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey404JSONResponse struct{ externalRef0.HTTP404NotFound }

func (response PostKeypoolKeyPoolIDKey404JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(404)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey429JSONResponse struct {
	externalRef0.HTTP429TooManyRequests
}

func (response PostKeypoolKeyPoolIDKey429JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(429)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey500JSONResponse struct {
	externalRef0.HTTP500InternalServerError
}

func (response PostKeypoolKeyPoolIDKey500JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey502JSONResponse struct{ externalRef0.HTTP502BadGateway }

func (response PostKeypoolKeyPoolIDKey502JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(502)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey503JSONResponse struct {
	externalRef0.HTTP503ServiceUnavailable
}

func (response PostKeypoolKeyPoolIDKey503JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(503)

	return ctx.JSON(&response)
}

type PostKeypoolKeyPoolIDKey504JSONResponse struct {
	externalRef0.HTTP504GatewayTimeout
}

func (response PostKeypoolKeyPoolIDKey504JSONResponse) VisitPostKeypoolKeyPoolIDKeyResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(504)

	return ctx.JSON(&response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// List all Key Pools. Supports optional filtering, sorting, and paging.
	// (GET /keypool)
	GetKeypool(ctx context.Context, request GetKeypoolRequestObject) (GetKeypoolResponseObject, error)
	// Create a new Key Pool.
	// (POST /keypool)
	PostKeypool(ctx context.Context, request PostKeypoolRequestObject) (PostKeypoolResponseObject, error)
	// List all Keys in Key Pool. Supports optional filtering, sorting, and paging.
	// (GET /keypool/{keyPoolID}/key)
	GetKeypoolKeyPoolIDKey(ctx context.Context, request GetKeypoolKeyPoolIDKeyRequestObject) (GetKeypoolKeyPoolIDKeyResponseObject, error)
	// Generate a new Key in a Key Pool.
	// (POST /keypool/{keyPoolID}/key)
	PostKeypoolKeyPoolIDKey(ctx context.Context, request PostKeypoolKeyPoolIDKeyRequestObject) (PostKeypoolKeyPoolIDKeyResponseObject, error)
}

type StrictHandlerFunc func(ctx *fiber.Ctx, args interface{}) (interface{}, error)

type StrictMiddlewareFunc func(f StrictHandlerFunc, operationID string) StrictHandlerFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// GetKeypool operation middleware
func (sh *strictHandler) GetKeypool(ctx *fiber.Ctx, params GetKeypoolParams) error {
	var request GetKeypoolRequestObject

	request.Params = params

	handler := func(ctx *fiber.Ctx, request interface{}) (interface{}, error) {
		return sh.ssi.GetKeypool(ctx.UserContext(), request.(GetKeypoolRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetKeypool")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	} else if validResponse, ok := response.(GetKeypoolResponseObject); ok {
		if err := validResponse.VisitGetKeypoolResponse(ctx); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// PostKeypool operation middleware
func (sh *strictHandler) PostKeypool(ctx *fiber.Ctx) error {
	var request PostKeypoolRequestObject

	var body PostKeypoolJSONRequestBody
	if err := ctx.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	request.Body = &body

	handler := func(ctx *fiber.Ctx, request interface{}) (interface{}, error) {
		return sh.ssi.PostKeypool(ctx.UserContext(), request.(PostKeypoolRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostKeypool")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	} else if validResponse, ok := response.(PostKeypoolResponseObject); ok {
		if err := validResponse.VisitPostKeypoolResponse(ctx); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetKeypoolKeyPoolIDKey operation middleware
func (sh *strictHandler) GetKeypoolKeyPoolIDKey(ctx *fiber.Ctx, keyPoolID externalRef0.KeyPoolId, params GetKeypoolKeyPoolIDKeyParams) error {
	var request GetKeypoolKeyPoolIDKeyRequestObject

	request.KeyPoolID = keyPoolID
	request.Params = params

	handler := func(ctx *fiber.Ctx, request interface{}) (interface{}, error) {
		return sh.ssi.GetKeypoolKeyPoolIDKey(ctx.UserContext(), request.(GetKeypoolKeyPoolIDKeyRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetKeypoolKeyPoolIDKey")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	} else if validResponse, ok := response.(GetKeypoolKeyPoolIDKeyResponseObject); ok {
		if err := validResponse.VisitGetKeypoolKeyPoolIDKeyResponse(ctx); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// PostKeypoolKeyPoolIDKey operation middleware
func (sh *strictHandler) PostKeypoolKeyPoolIDKey(ctx *fiber.Ctx, keyPoolID externalRef0.KeyPoolId) error {
	var request PostKeypoolKeyPoolIDKeyRequestObject

	request.KeyPoolID = keyPoolID

	var body PostKeypoolKeyPoolIDKeyJSONRequestBody
	if err := ctx.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	request.Body = &body

	handler := func(ctx *fiber.Ctx, request interface{}) (interface{}, error) {
		return sh.ssi.PostKeypoolKeyPoolIDKey(ctx.UserContext(), request.(PostKeypoolKeyPoolIDKeyRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostKeypoolKeyPoolIDKey")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	} else if validResponse, ok := response.(PostKeypoolKeyPoolIDKeyResponseObject); ok {
		if err := validResponse.VisitPostKeypoolKeyPoolIDKeyResponse(ctx); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"9d/6Z2T+OP9vAAAA//8Yxmzx0jsAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(path.Dir(pathToFile), "./openapi_spec_components.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
