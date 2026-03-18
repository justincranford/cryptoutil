// Copyright (c) 2025 Justin Cranford
//

// Package builder provides fluent API for constructing service applications.
package builder

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"

	fiber "github.com/gofiber/fiber/v2"
)

// ErrStrictServerNoBasePath indicates that no base path was configured.
var ErrStrictServerNoBasePath = errors.New("strict server requires at least one base path (browser or service)")

// StrictServerConfig configures OpenAPI strict server registration.
// This abstraction supports oapi-codegen's strict server pattern where:
// 1. Domain service implements generated StrictServerInterface.
// 2. StrictServerInterface is wrapped with NewStrictHandler().
// 3. Handler is registered with RegisterHandlersWithOptions().
type StrictServerConfig struct {
	// BrowserAPIBasePath is the base URL for browser-facing APIs (e.g., "/browser/api/v1").
	BrowserAPIBasePath string

	// ServiceAPIBasePath is the base URL for service-to-service APIs (e.g., "/service/api/v1").
	ServiceAPIBasePath string

	// BrowserMiddlewares are middlewares applied to browser API routes.
	BrowserMiddlewares []fiber.Handler

	// ServiceMiddlewares are middlewares applied to service API routes.
	ServiceMiddlewares []fiber.Handler

	// RegisterBrowserHandlers is a function that registers handlers on the browser API router.
	// This should call the generated RegisterHandlersWithOptions() function.
	// Parameters: router, baseURL, middlewares.
	RegisterBrowserHandlers func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error

	// RegisterServiceHandlers is a function that registers handlers on the service API router.
	// This should call the generated RegisterHandlersWithOptions() function.
	// Parameters: router, baseURL, middlewares.
	RegisterServiceHandlers func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error
}

// NewDefaultStrictServerConfig creates a default StrictServerConfig with standard paths.
// BrowserAPIBasePath defaults to "/browser/api/v1".
// ServiceAPIBasePath defaults to "/service/api/v1".
func NewDefaultStrictServerConfig() *StrictServerConfig {
	return &StrictServerConfig{
		BrowserAPIBasePath: cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
		ServiceAPIBasePath: cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
		BrowserMiddlewares: nil,
		ServiceMiddlewares: nil,
	}
}

// WithBrowserBasePath sets the browser API base path.
func (c *StrictServerConfig) WithBrowserBasePath(basePath string) *StrictServerConfig {
	c.BrowserAPIBasePath = basePath

	return c
}

// WithServiceBasePath sets the service API base path.
func (c *StrictServerConfig) WithServiceBasePath(basePath string) *StrictServerConfig {
	c.ServiceAPIBasePath = basePath

	return c
}

// WithBrowserMiddlewares sets the browser API middlewares.
func (c *StrictServerConfig) WithBrowserMiddlewares(middlewares ...fiber.Handler) *StrictServerConfig {
	c.BrowserMiddlewares = middlewares

	return c
}

// WithServiceMiddlewares sets the service API middlewares.
func (c *StrictServerConfig) WithServiceMiddlewares(middlewares ...fiber.Handler) *StrictServerConfig {
	c.ServiceMiddlewares = middlewares

	return c
}

// WithBrowserHandlerRegistration sets the browser handler registration function.
func (c *StrictServerConfig) WithBrowserHandlerRegistration(fn func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error) *StrictServerConfig {
	c.RegisterBrowserHandlers = fn

	return c
}

// WithServiceHandlerRegistration sets the service handler registration function.
func (c *StrictServerConfig) WithServiceHandlerRegistration(fn func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error) *StrictServerConfig {
	c.RegisterServiceHandlers = fn

	return c
}

// Validate checks that the configuration is valid.
func (c *StrictServerConfig) Validate() error {
	if c.BrowserAPIBasePath == "" && c.ServiceAPIBasePath == "" {
		return ErrStrictServerNoBasePath
	}

	return nil
}

// RegisterHandlers registers the strict server handlers on the provided Fiber app.
// It registers both browser and service handlers if their registration functions are set.
func (c *StrictServerConfig) RegisterHandlers(app *fiber.App) error {
	if c.RegisterBrowserHandlers != nil && c.BrowserAPIBasePath != "" {
		if err := c.RegisterBrowserHandlers(app, c.BrowserAPIBasePath, c.BrowserMiddlewares); err != nil {
			return err
		}
	}

	if c.RegisterServiceHandlers != nil && c.ServiceAPIBasePath != "" {
		if err := c.RegisterServiceHandlers(app, c.ServiceAPIBasePath, c.ServiceMiddlewares); err != nil {
			return err
		}
	}

	return nil
}

// StrictServerRegistrar provides a standardized way for domain services to register
// their OpenAPI strict server handlers. Domain services implement this interface
// and provide it to the ServerBuilder.
type StrictServerRegistrar interface {
	// RegisterStrictServerHandlers registers the service's strict server handlers.
	// The implementation should:
	// 1. Create the StrictServerInterface implementation with business logic.
	// 2. Wrap it with NewStrictHandler().
	// 3. Call RegisterHandlersWithOptions() with the provided router.
	RegisterStrictServerHandlers(app *fiber.App, config *StrictServerConfig) error
}
