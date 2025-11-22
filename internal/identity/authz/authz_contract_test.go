// Copyright (c) 2025 Justin Cranford
//

package authz_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilApiIdentityAuthz "cryptoutil/api/identity/authz"
	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TestAuthZContractHealth verifies the /health endpoint complies with the OpenAPI spec.
func TestAuthZContractHealth(t *testing.T) {
	t.Parallel()

	// Load OpenAPI spec for validation.
	spec, err := cryptoutilApiIdentityAuthz.GetSwagger()
	require.NoError(t, err, "Failed to load OpenAPI spec")

	// Create AuthZ service with minimal config.
	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenDuration: 3600,
		},
	}

	dbConfig := config.Database
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	defer func() {
		_ = repoFactory.Close()
	}()

	// Create mock issuers for token service.
	mockJWSIssuer := &mockJWSIssuer{}
	mockJWEIssuer := &mockJWEIssuer{}
	mockUUIDIssuer := &mockUUIDIssuer{}

	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(mockJWSIssuer, mockJWEIssuer, mockUUIDIssuer, config.Tokens)
	authzSvc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	// Create Fiber app with routes.
	app := fiber.New()
	authzSvc.RegisterMiddleware(app)
	authzSvc.RegisterRoutes(app)

	// Test GET /health endpoint.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
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

	// Parse response body.
	var healthResp cryptoutilApiIdentityAuthz.HealthResponse
	err = json.Unmarshal(bodyBytes, &healthResp)
	require.NoError(t, err, "Failed to parse response body")

	// Verify business logic.
	require.Equal(t, cryptoutilApiIdentityAuthz.Healthy, healthResp.Status, "Health status should be healthy")
}

// Mock issuers for testing.
type mockJWSIssuer struct{}

func (m *mockJWSIssuer) Issue(ctx context.Context, claims map[string]any) (string, error) {
	return "mock.jws.token", nil
}

type mockJWEIssuer struct{}

func (m *mockJWEIssuer) Issue(ctx context.Context, payload []byte) (string, error) {
	return "mock.jwe.token", nil
}

type mockUUIDIssuer struct{}

func (m *mockUUIDIssuer) Issue(ctx context.Context) (string, error) {
	return "123e4567-e89b-12d3-a456-426614174000", nil
}
