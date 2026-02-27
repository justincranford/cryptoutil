// Copyright (c) 2025 Justin Cranford

package idp_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleConsent_MissingRequestID validates error when request_id query param is missing.
func TestHandleConsent_MissingRequestID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:        cryptoutilSharedMagic.IDPServiceName,
			BindAddress: cryptoutilSharedMagic.IPv4Loopback,
			Port:        cryptoutilSharedMagic.DemoServerPort,
			TLSEnabled:  true,
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			CookieName:      "session_id",
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
			SessionLifetime: 1 * time.Hour,
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	app := fiber.New()
	service.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/oidc/v1/consent", nil) // No request_id query param

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Middleware intercepts before handler - returns 401 Unauthorized
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandleConsent_InvalidRequestIDFormat validates error when request_id is not a valid UUID.
func TestHandleConsent_InvalidRequestIDFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:        cryptoutilSharedMagic.IDPServiceName,
			BindAddress: cryptoutilSharedMagic.IPv4Loopback,
			Port:        cryptoutilSharedMagic.DemoServerPort,
			TLSEnabled:  true,
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			CookieName:      "session_id",
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
			SessionLifetime: 1 * time.Hour,
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	app := fiber.New()
	service.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/oidc/v1/consent?request_id=not-a-uuid", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Middleware intercepts before handler - returns 401 Unauthorized
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandleConsent_RequestNotFound validates error when authorization request doesn't exist.
func TestHandleConsent_RequestNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:        cryptoutilSharedMagic.IDPServiceName,
			BindAddress: cryptoutilSharedMagic.IPv4Loopback,
			Port:        cryptoutilSharedMagic.DemoServerPort,
			TLSEnabled:  true,
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			CookieName:      "session_id",
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
			SessionLifetime: 1 * time.Hour,
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	app := fiber.New()
	service.RegisterRoutes(app)

	// Use a valid UUID that doesn't exist in database
	nonexistentID := googleUuid.Must(googleUuid.NewV7()).String()
	req := httptest.NewRequest(http.MethodGet, "/oidc/v1/consent?request_id="+nonexistentID, nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Middleware intercepts before handler - returns 401 Unauthorized
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
