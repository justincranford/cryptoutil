// Copyright (c) 2025 Justin Cranford
//

// Package builder provides fluent API for constructing service applications.
package builder

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"github.com/getkin/kin-openapi/openapi3"
	fiber "github.com/gofiber/fiber/v2"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

// OpenAPIConfig configures OpenAPI handler registration.
type OpenAPIConfig struct {
	// SwaggerSpec is the parsed OpenAPI specification from oapi-codegen's GetSwagger().
	SwaggerSpec *openapi3.T

	// BrowserAPIBasePath is the base URL for browser-facing APIs (e.g., "/browser/api/v1").
	BrowserAPIBasePath string

	// ServiceAPIBasePath is the base URL for service-to-service APIs (e.g., "/service/api/v1").
	ServiceAPIBasePath string

	// EnableRequestValidation enables OpenAPI request validation middleware.
	// When true, incoming requests are validated against the OpenAPI spec.
	EnableRequestValidation bool

	// ValidatorOptions configures the request validator middleware.
	// If nil, default options are used.
	ValidatorOptions *fibermiddleware.Options
}

// NewDefaultOpenAPIConfig creates a default OpenAPI configuration.
// BrowserAPIBasePath defaults to "/browser/api/v1".
// ServiceAPIBasePath defaults to "/service/api/v1".
// EnableRequestValidation defaults to true.
func NewDefaultOpenAPIConfig(swaggerSpec *openapi3.T) *OpenAPIConfig {
	return &OpenAPIConfig{
		SwaggerSpec:             swaggerSpec,
		BrowserAPIBasePath:      cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
		ServiceAPIBasePath:      cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
		EnableRequestValidation: true,
		ValidatorOptions:        &fibermiddleware.Options{},
	}
}

// CreateRequestValidatorMiddleware creates an OpenAPI request validator middleware.
// The middleware validates incoming requests against the OpenAPI specification.
// Returns nil if the swagger spec is nil or validation is disabled.
func (c *OpenAPIConfig) CreateRequestValidatorMiddleware() fiber.Handler {
	if c.SwaggerSpec == nil || !c.EnableRequestValidation {
		return nil
	}

	opts := c.ValidatorOptions
	if opts == nil {
		opts = &fibermiddleware.Options{}
	}

	return fibermiddleware.OapiRequestValidatorWithOptions(c.SwaggerSpec, opts)
}

// BrowserMiddlewares returns the middleware stack for browser API paths.
// Includes request validation if enabled.
func (c *OpenAPIConfig) BrowserMiddlewares() []fiber.Handler {
	var middlewares []fiber.Handler

	if validator := c.CreateRequestValidatorMiddleware(); validator != nil {
		middlewares = append(middlewares, validator)
	}

	return middlewares
}

// ServiceMiddlewares returns the middleware stack for service API paths.
// Includes request validation if enabled.
func (c *OpenAPIConfig) ServiceMiddlewares() []fiber.Handler {
	var middlewares []fiber.Handler

	if validator := c.CreateRequestValidatorMiddleware(); validator != nil {
		middlewares = append(middlewares, validator)
	}

	return middlewares
}

// OpenAPIRegistrar provides a standardized way to register OpenAPI handlers.
// Domain services implement this interface to register their generated handlers.
type OpenAPIRegistrar interface {
	// RegisterOpenAPIHandlers registers the service's OpenAPI handlers on the Fiber app.
	// The config provides base paths and middleware configuration.
	RegisterOpenAPIHandlers(app *fiber.App, config *OpenAPIConfig) error
}
