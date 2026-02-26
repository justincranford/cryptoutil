// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
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

// TestRegisterRoutes_RouteRegistration validates route registration.
func TestRegisterRoutes_RouteRegistration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Initialize IDP service.
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
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:           "https://localhost:8080",
			SigningAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)

	app := fiber.New()
	service.RegisterRoutes(app)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET /health",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET /.well-known/openid-configuration",
			method:         http.MethodGet,
			path:           cryptoutilSharedMagic.PathDiscovery,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET /oidc/v1/login missing request_id",
			method:         http.MethodGet,
			path:           "/oidc/v1/login",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET /oidc/v1/consent missing session",
			method:         http.MethodGet,
			path:           "/oidc/v1/consent?request_id=" + googleUuid.Must(googleUuid.NewV7()).String(),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "POST /oidc/v1/logout missing session",
			method:         http.MethodPost,
			path:           "/oidc/v1/logout",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "GET /oidc/v1/userinfo missing auth",
			method:         http.MethodGet,
			path:           "/oidc/v1/userinfo",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
