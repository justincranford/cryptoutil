// Copyright (c) 2025 Justin Cranford
//

package builder

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	fiber "github.com/gofiber/fiber/v2"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
	"github.com/stretchr/testify/require"
)

// TestNewDefaultOpenAPIConfig verifies default configuration values.
func TestNewDefaultOpenAPIConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		swaggerSpec *openapi3.T
		wantNil     bool
	}{
		{
			name:        "nil swagger spec",
			swaggerSpec: nil,
			wantNil:     false,
		},
		{
			name:        "valid swagger spec",
			swaggerSpec: &openapi3.T{},
			wantNil:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := NewDefaultOpenAPIConfig(tc.swaggerSpec)

			require.NotNil(t, config)
			require.Equal(t, tc.swaggerSpec, config.SwaggerSpec)
			require.Equal(t, "/browser/api/v1", config.BrowserAPIBasePath)
			require.Equal(t, "/service/api/v1", config.ServiceAPIBasePath)
			require.True(t, config.EnableRequestValidation)
			require.NotNil(t, config.ValidatorOptions)
		})
	}
}

// TestOpenAPIConfig_CreateRequestValidatorMiddleware tests middleware creation.
func TestOpenAPIConfig_CreateRequestValidatorMiddleware(t *testing.T) {
	t.Parallel()

	// Create a minimal valid OpenAPI spec for testing.
	minimalSpec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	tests := []struct {
		name       string
		config     *OpenAPIConfig
		wantNil    bool
		wantReason string
	}{
		{
			name: "nil swagger spec returns nil",
			config: &OpenAPIConfig{
				SwaggerSpec:             nil,
				EnableRequestValidation: true,
			},
			wantNil:    true,
			wantReason: "nil swagger spec should return nil middleware",
		},
		{
			name: "validation disabled returns nil",
			config: &OpenAPIConfig{
				SwaggerSpec:             minimalSpec,
				EnableRequestValidation: false,
			},
			wantNil:    true,
			wantReason: "disabled validation should return nil middleware",
		},
		{
			name: "valid config with nil options",
			config: &OpenAPIConfig{
				SwaggerSpec:             minimalSpec,
				EnableRequestValidation: true,
				ValidatorOptions:        nil,
			},
			wantNil:    false,
			wantReason: "valid config should return middleware",
		},
		{
			name: "valid config with options",
			config: &OpenAPIConfig{
				SwaggerSpec:             minimalSpec,
				EnableRequestValidation: true,
				ValidatorOptions:        &fibermiddleware.Options{},
			},
			wantNil:    false,
			wantReason: "valid config with options should return middleware",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			middleware := tc.config.CreateRequestValidatorMiddleware()

			if tc.wantNil {
				require.Nil(t, middleware, tc.wantReason)
			} else {
				require.NotNil(t, middleware, tc.wantReason)
			}
		})
	}
}

// TestOpenAPIConfig_Middlewares tests BrowserMiddlewares and ServiceMiddlewares.
func TestOpenAPIConfig_Middlewares(t *testing.T) {
	t.Parallel()

	minimalSpec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	tests := []struct {
		name      string
		config    *OpenAPIConfig
		wantCount int
	}{
		{
			name: "validation disabled - empty middlewares",
			config: &OpenAPIConfig{
				SwaggerSpec:             minimalSpec,
				EnableRequestValidation: false,
			},
			wantCount: 0,
		},
		{
			name: "validation enabled - one middleware",
			config: &OpenAPIConfig{
				SwaggerSpec:             minimalSpec,
				EnableRequestValidation: true,
			},
			wantCount: 1,
		},
		{
			name: "nil spec - empty middlewares",
			config: &OpenAPIConfig{
				SwaggerSpec:             nil,
				EnableRequestValidation: true,
			},
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			browserMiddlewares := tc.config.BrowserMiddlewares()
			serviceMiddlewares := tc.config.ServiceMiddlewares()

			require.Len(t, browserMiddlewares, tc.wantCount)
			require.Len(t, serviceMiddlewares, tc.wantCount)
		})
	}
}

// TestOpenAPIConfig_MiddlewareExecution tests that middleware actually executes.
func TestOpenAPIConfig_MiddlewareExecution(t *testing.T) {
	t.Parallel()

	// Create a minimal valid OpenAPI spec with a test path.
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{
			Extensions: map[string]any{},
		},
	}

	// Add a path to the spec.
	spec.Paths.Set("/test", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "getTest",
			Responses: &openapi3.Responses{
				Extensions: map[string]any{},
			},
		},
	})

	// Initialize the responses with a 200 response.
	spec.Paths.Find("/test").Get.Responses.Set("200", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: ptr("OK"),
		},
	})

	config := &OpenAPIConfig{
		SwaggerSpec:             spec,
		BrowserAPIBasePath:      "/browser/api/v1",
		ServiceAPIBasePath:      "/service/api/v1",
		EnableRequestValidation: true,
		ValidatorOptions:        &fibermiddleware.Options{},
	}

	// Get browser middlewares.
	middlewares := config.BrowserMiddlewares()
	require.Len(t, middlewares, 1)

	// Create a test Fiber app with the middleware.
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Use the middleware on a test route.
	app.Use(middlewares[0])
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Make a request - the middleware should process it.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// The request should succeed (path is defined in spec).
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// ptr returns a pointer to a string.
func ptr(s string) *string {
	return &s
}

// TestOpenAPIRegistrar_Interface verifies interface can be implemented.
func TestOpenAPIRegistrar_Interface(t *testing.T) {
	t.Parallel()

	// Verify that a concrete type can implement the interface.
	var _ OpenAPIRegistrar = &mockOpenAPIRegistrar{}
}

// mockOpenAPIRegistrar is a test implementation of OpenAPIRegistrar.
type mockOpenAPIRegistrar struct {
	registerCalled bool
}

func (m *mockOpenAPIRegistrar) RegisterOpenAPIHandlers(app *fiber.App, config *OpenAPIConfig) error {
	m.registerCalled = true

	return nil
}
