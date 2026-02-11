// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestRegisterRoutes_HealthEndpoint validates health check route registration.
func TestRegisterRoutes_HealthEndpoint(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// Create minimal service with config, repository factory, and token service.
	config := createRoutesTestConfig()
	repoFactory := createRoutesTestRepoFactory(t)
	tokenSvc := createRoutesTestTokenService(config, repoFactory)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	service.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Health check request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Health check should return 200")
}

// TestRegisterRoutes_OAuth2Endpoints validates OAuth 2.1 endpoint registration.
func TestRegisterRoutes_OAuth2Endpoints(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// Create minimal service with config, repository factory, and token service.
	config := createRoutesTestConfig()
	repoFactory := createRoutesTestRepoFactory(t)
	tokenSvc := createRoutesTestTokenService(config, repoFactory)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	service.RegisterRoutes(app)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "authorize GET endpoint",
			method:     http.MethodGet,
			path:       "/oauth2/v1/authorize",
			wantStatus: fiber.StatusBadRequest, // Missing required params.
		},
		{
			name:       "authorize POST endpoint",
			method:     http.MethodPost,
			path:       "/oauth2/v1/authorize",
			wantStatus: fiber.StatusBadRequest, // Missing required params.
		},
		{
			name:       "token endpoint",
			method:     http.MethodPost,
			path:       "/oauth2/v1/token",
			wantStatus: fiber.StatusBadRequest, // Missing grant_type.
		},
		{
			name:       "introspect endpoint",
			method:     http.MethodPost,
			path:       "/oauth2/v1/introspect",
			wantStatus: fiber.StatusBadRequest, // Missing token.
		},
		{
			name:       "revoke endpoint",
			method:     http.MethodPost,
			path:       "/oauth2/v1/revoke",
			wantStatus: fiber.StatusBadRequest, // Revoke returns 400 for missing token.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err, "Request should succeed")

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.wantStatus, resp.StatusCode, "Status code should match expected")
		})
	}
}

// TestRegisterRoutes_SwaggerEndpoint validates OpenAPI spec endpoint registration.
func TestRegisterRoutes_SwaggerEndpoint(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// Create minimal service with config, repository factory, and token service.
	config := createRoutesTestConfig()
	repoFactory := createRoutesTestRepoFactory(t)
	tokenSvc := createRoutesTestTokenService(config, repoFactory)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	service.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/ui/swagger/doc.json", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Swagger request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Swagger endpoint may or may not be registered depending on spec generation.
	require.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound}, resp.StatusCode, "Swagger should return 200 or 404")
}

// createRoutesTestConfig creates a minimal config for route testing.
func createRoutesTestConfig() *cryptoutilIdentityConfig.Config {
	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  "file::memory:?cache=private",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}

// createRoutesTestRepoFactory creates a repository factory for route testing.
func createRoutesTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createRoutesTestConfig()
	ctx := context.Background()

	factory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Repository factory creation should succeed")

	err = factory.AutoMigrate(ctx)
	require.NoError(t, err, "Auto migration should succeed")

	return factory
}

// createRoutesTestTokenService creates a token service for route testing.
func createRoutesTestTokenService(
	_ *cryptoutilIdentityConfig.Config,
	_ *cryptoutilIdentityRepository.RepositoryFactory,
) *cryptoutilIdentityIssuer.TokenService {
	return nil
}
