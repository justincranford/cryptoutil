// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestHandleHealth_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Initialize database and repositories.
	cfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg)
	require.NoError(t, err, "Failed to create repository factory")

	// Initialize token service.
	appCfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout,
			RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
			IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout,
		},
	}
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, appCfg.Tokens)

	// Create AuthZ service.
	authzSvc := cryptoutilIdentityAuthz.NewService(appCfg, repoFactory, tokenSvc)

	// Create Fiber app and register routes.
	app := fiber.New()
	authzSvc.RegisterRoutes(app)

	// Create test request.
	req := httptest.NewRequest("GET", "/health", nil)

	// Execute request.
	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request failed")

	defer func() { //nolint:errcheck // Test cleanup - error intentionally ignored
		_ = resp.Body.Close()
	}()

	// Validate response.
	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected 200 OK status")
}

func TestHandleHealth_DatabaseUnavailable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Initialize database with valid DSN.
	cfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg)
	require.NoError(t, err, "Failed to create repository factory")

	// Close the database connection to simulate unavailability.
	db := repoFactory.DB()
	sqlDB, err := db.DB()
	require.NoError(t, err, "Failed to get SQL DB")
	err = sqlDB.Close()
	require.NoError(t, err, "Failed to close database")

	// Initialize token service.
	appCfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout,
			RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
			IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout,
		},
	}
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, appCfg.Tokens)

	// Create AuthZ service.
	authzSvc := cryptoutilIdentityAuthz.NewService(appCfg, repoFactory, tokenSvc)

	// Create Fiber app and register routes.
	app := fiber.New()
	authzSvc.RegisterRoutes(app)

	// Create test request.
	req := httptest.NewRequest("GET", "/health", nil)

	// Execute request.
	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request failed")

	defer func() { //nolint:errcheck // Test cleanup - error intentionally ignored
		_ = resp.Body.Close()
	}()

	// Validate response - should return 503 when database is unavailable.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode, "Expected 503 Service Unavailable status")
}
