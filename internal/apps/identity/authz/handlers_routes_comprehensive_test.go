// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestRegisterRoutes_AllEndpointsRegistered(t *testing.T) {
	t.Parallel()

	config := createRoutesComprehensiveTestConfig(t)
	repoFactory := createRoutesComprehensiveTestRepoFactory(t)
	tokenSvc := createRoutesComprehensiveTestTokenService(t)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	app := fiber.New()
	service.RegisterRoutes(app)

	// Test health endpoint.
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Health check should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Health check should return 200 OK")

	// Test OAuth 2.1 endpoints exist (will return errors without proper setup, but routes should be registered).
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/oauth2/v1/authorize"},
		{"POST", "/oauth2/v1/authorize"},
		{"POST", "/oauth2/v1/token"},
		{"POST", "/oauth2/v1/introspect"},
		{"POST", "/oauth2/v1/revoke"},
	}

	for _, endpoint := range endpoints {
		req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err, "Route %s %s should be registered", endpoint.method, endpoint.path)

		defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

		require.NotEqual(t, fiber.StatusNotFound, resp.StatusCode, "Route %s %s should not return 404", endpoint.method, endpoint.path)
	}
}

func createRoutesComprehensiveTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://identity.example.com",
		},
	}
}

func createRoutesComprehensiveTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         "file::memory:?cache=private&_id=" + googleUuid.NewString(),
		AutoMigrate: true,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run database migrations")

	return repoFactory
}

func createRoutesComprehensiveTestTokenService(t *testing.T) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	config := &cryptoutilIdentityConfig.TokenConfig{
		Issuer: "https://identity.example.com",
	}

	return cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config)
}
