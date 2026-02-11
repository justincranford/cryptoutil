// Copyright (c) 2025 Justin Cranford
//

package builder

import (
	"errors"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultStrictServerConfig(t *testing.T) {
	t.Parallel()

	config := NewDefaultStrictServerConfig()

	require.NotNil(t, config)
	require.Equal(t, "/browser/api/v1", config.BrowserAPIBasePath)
	require.Equal(t, "/service/api/v1", config.ServiceAPIBasePath)
	require.Nil(t, config.BrowserMiddlewares)
	require.Nil(t, config.ServiceMiddlewares)
	require.Nil(t, config.RegisterBrowserHandlers)
	require.Nil(t, config.RegisterServiceHandlers)
}

func TestStrictServerConfig_Validate_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *StrictServerConfig
	}{
		{
			name:   "default config",
			config: NewDefaultStrictServerConfig(),
		},
		{
			name: "browser only",
			config: &StrictServerConfig{
				BrowserAPIBasePath: "/browser/api/v1",
				ServiceAPIBasePath: "",
			},
		},
		{
			name: "service only",
			config: &StrictServerConfig{
				BrowserAPIBasePath: "",
				ServiceAPIBasePath: "/service/api/v1",
			},
		},
		{
			name: "both paths",
			config: &StrictServerConfig{
				BrowserAPIBasePath: "/browser/api/v2",
				ServiceAPIBasePath: "/service/api/v2",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			require.NoError(t, err)
		})
	}
}

func TestStrictServerConfig_Validate_NoBasePath(t *testing.T) {
	t.Parallel()

	config := &StrictServerConfig{
		BrowserAPIBasePath: "",
		ServiceAPIBasePath: "",
	}

	err := config.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrStrictServerNoBasePath)
}

func TestStrictServerConfig_FluentSetters(t *testing.T) {
	t.Parallel()

	middleware1 := func(c *fiber.Ctx) error { return c.Next() }
	middleware2 := func(c *fiber.Ctx) error { return c.Next() }

	config := NewDefaultStrictServerConfig().
		WithBrowserBasePath("/custom/browser").
		WithServiceBasePath("/custom/service").
		WithBrowserMiddlewares(middleware1, middleware2).
		WithServiceMiddlewares(middleware1)

	require.Equal(t, "/custom/browser", config.BrowserAPIBasePath)
	require.Equal(t, "/custom/service", config.ServiceAPIBasePath)
	require.Len(t, config.BrowserMiddlewares, 2)
	require.Len(t, config.ServiceMiddlewares, 1)
}

func TestStrictServerConfig_HandlerRegistration(t *testing.T) {
	t.Parallel()

	var (
		browserCalled, serviceCalled           bool
		capturedBrowserURL, capturedServiceURL string
	)

	browserHandler := func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error {
		browserCalled = true
		capturedBrowserURL = baseURL

		return nil
	}

	serviceHandler := func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error {
		serviceCalled = true
		capturedServiceURL = baseURL

		return nil
	}

	config := NewDefaultStrictServerConfig().
		WithBrowserHandlerRegistration(browserHandler).
		WithServiceHandlerRegistration(serviceHandler)

	app := fiber.New()
	err := config.RegisterHandlers(app)
	require.NoError(t, err)

	require.True(t, browserCalled)
	require.True(t, serviceCalled)
	require.Equal(t, "/browser/api/v1", capturedBrowserURL)
	require.Equal(t, "/service/api/v1", capturedServiceURL)
}

func TestStrictServerConfig_RegisterHandlers_BrowserOnly(t *testing.T) {
	t.Parallel()

	var browserCalled bool

	config := &StrictServerConfig{
		BrowserAPIBasePath: "/browser/api/v1",
		ServiceAPIBasePath: "", // No service path.
		RegisterBrowserHandlers: func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error {
			browserCalled = true

			return nil
		},
	}

	app := fiber.New()
	err := config.RegisterHandlers(app)
	require.NoError(t, err)
	require.True(t, browserCalled)
}

func TestStrictServerConfig_RegisterHandlers_ServiceOnly(t *testing.T) {
	t.Parallel()

	var serviceCalled bool

	config := &StrictServerConfig{
		BrowserAPIBasePath: "", // No browser path.
		ServiceAPIBasePath: "/service/api/v1",
		RegisterServiceHandlers: func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error {
			serviceCalled = true

			return nil
		},
	}

	app := fiber.New()
	err := config.RegisterHandlers(app)
	require.NoError(t, err)
	require.True(t, serviceCalled)
}

func TestStrictServerConfig_RegisterHandlers_BrowserError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("browser registration failed")

	config := NewDefaultStrictServerConfig().
		WithBrowserHandlerRegistration(func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error {
			return expectedErr
		})

	app := fiber.New()
	err := config.RegisterHandlers(app)
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestStrictServerConfig_RegisterHandlers_ServiceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("service registration failed")

	config := NewDefaultStrictServerConfig().
		WithServiceHandlerRegistration(func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error {
			return expectedErr
		})

	app := fiber.New()
	err := config.RegisterHandlers(app)
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestStrictServerConfig_NoHandlersRegistered(t *testing.T) {
	t.Parallel()

	config := NewDefaultStrictServerConfig()
	// No handlers set.

	app := fiber.New()
	err := config.RegisterHandlers(app)
	require.NoError(t, err) // Should succeed, just do nothing.
}

func TestStrictServerConfig_MiddlewaresPassedToHandlers(t *testing.T) {
	t.Parallel()

	middleware1 := func(c *fiber.Ctx) error { return c.Next() }
	middleware2 := func(c *fiber.Ctx) error { return c.Next() }

	var capturedMiddlewares []fiber.Handler

	config := NewDefaultStrictServerConfig().
		WithBrowserMiddlewares(middleware1, middleware2).
		WithBrowserHandlerRegistration(func(router fiber.Router, baseURL string, middlewares []fiber.Handler) error {
			capturedMiddlewares = middlewares

			return nil
		})

	app := fiber.New()
	err := config.RegisterHandlers(app)
	require.NoError(t, err)
	require.Len(t, capturedMiddlewares, 2)
}

func TestWithStrictServer_NilConfig(t *testing.T) {
	t.Parallel()

	// WithStrictServer with nil should return builder without error.
	// (validated by checking that Build doesn't fail due to strict server config)
	builder := &ServerBuilder{}
	result := builder.WithStrictServer(nil)
	require.Same(t, builder, result)
	require.Nil(t, builder.strictServerConfig)
}

func TestWithStrictServer_ValidConfig(t *testing.T) {
	t.Parallel()

	config := NewDefaultStrictServerConfig()
	builder := &ServerBuilder{}
	result := builder.WithStrictServer(config)

	require.Same(t, builder, result)
	require.NotNil(t, builder.strictServerConfig)
	require.Equal(t, config, builder.strictServerConfig)
}

func TestWithStrictServer_InvalidConfig(t *testing.T) {
	t.Parallel()

	// Config with no base paths should fail validation.
	config := &StrictServerConfig{
		BrowserAPIBasePath: "",
		ServiceAPIBasePath: "",
	}
	builder := &ServerBuilder{}
	result := builder.WithStrictServer(config)

	require.Same(t, builder, result)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "invalid strict server config")
}

func TestWithStrictServer_ErrorAccumulation(t *testing.T) {
	t.Parallel()

	// If builder already has error, WithStrictServer should not modify it.
	existingErr := errors.New("existing error")
	builder := &ServerBuilder{err: existingErr}
	config := NewDefaultStrictServerConfig()

	result := builder.WithStrictServer(config)
	require.Same(t, builder, result)
	require.ErrorIs(t, builder.err, existingErr)
	require.Nil(t, builder.strictServerConfig) // Should not be set.
}
