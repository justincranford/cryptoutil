// Copyright (c) 2025 Justin Cranford
//

package rs_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilApiIdentityRs "cryptoutil/api/identity/rs"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityRs "cryptoutil/internal/identity/rs"
)

// TestRSContractPublicHealth verifies the /api/v1/public/health endpoint complies with the OpenAPI spec.
func TestRSContractPublicHealth(t *testing.T) {
	t.Parallel()

	// Load OpenAPI spec for validation.
	spec, err := cryptoutilApiIdentityRs.GetSwagger()
	require.NoError(t, err, "Failed to load OpenAPI spec")

	// Create RS service with minimal config.
	config := &cryptoutilIdentityConfig.Config{}

	// Mock token service for testing.
	mockTokenSvc := &mockTokenService{}
	rsSvc := cryptoutilIdentityRs.NewService(config, nil, mockTokenSvc)

	// Create Fiber app with routes.
	app := fiber.New()
	rsSvc.RegisterMiddleware(app)
	rsSvc.RegisterRoutes(app)

	// Test GET /api/v1/public/health endpoint.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Failed to execute request")

	defer func() {
		_ = resp.Body.Close()
	}()

	// Verify response against OpenAPI spec.
	router, err := gorillamux.NewRouter(spec)
	require.NoError(t, err, "Failed to create OpenAPI router")

	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err, "Failed to find route in OpenAPI spec")

	// Read response body for validation.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	// Create validation input.
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	// Create response validation input.
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 resp.StatusCode,
		Header:                 resp.Header,
		Body:                   io.NopCloser(bytes.NewReader(bodyBytes)),
	}

	// Validate response against spec.
	err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
	require.NoError(t, err, "Response does not match OpenAPI spec")

	// Verify HTTP status code.
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")
}

// mockTokenService is a mock implementation of TokenService for testing.
type mockTokenService struct{}

func (m *mockTokenService) VerifyAccessToken(ctx context.Context, token string) (map[string]any, error) {
	return map[string]any{
		"sub":   "user123",
		"scope": "read:resource",
	}, nil
}
