// Copyright (c) 2025 Justin Cranford
//
//

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
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleLogout_POST validates POST /logout endpoint terminates sessions.
func TestHandleLogout_POST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Session{},
	)
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
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)

	app := fiber.New()
	service.RegisterRoutes(app)

	// Create test user.
	testUser := &cryptoutilIdentityDomain.User{
		Sub:   googleUuid.Must(googleUuid.NewV7()).String(),
		Email: "test@example.com",
	}
	userRepo := repoFactory.UserRepository()
	require.NoError(t, userRepo.Create(ctx, testUser))

	sessionRepo := repoFactory.SessionRepository()

	tests := []struct {
		name           string
		sessionCookie  string
		expectedStatus int
		createSession  bool
	}{
		{
			name:           "Valid session cookie",
			sessionCookie:  "", // Will be set by createSession
			expectedStatus: http.StatusOK,
			createSession:  true,
		},
		{
			name:           "Missing session cookie",
			sessionCookie:  "",
			expectedStatus: http.StatusUnauthorized,
			createSession:  false,
		},
		{
			name:           "Invalid session cookie",
			sessionCookie:  googleUuid.Must(googleUuid.NewV7()).String(),
			expectedStatus: http.StatusUnauthorized,
			createSession:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sessionCookie := tc.sessionCookie

			// Create fresh session for each test that needs one.
			if tc.createSession {
				testSession := &cryptoutilIdentityDomain.Session{
					UserID:                testUser.ID,
					SessionID:             googleUuid.Must(googleUuid.NewV7()).String(),
					IPAddress:             cryptoutilSharedMagic.IPv4Loopback,
					UserAgent:             "test-agent",
					IssuedAt:              time.Now().UTC(),
					ExpiresAt:             time.Now().UTC().Add(1 * time.Hour),
					LastSeenAt:            time.Now().UTC(),
					Active:                boolPtr(true),
					AuthenticationMethods: []string{cryptoutilSharedMagic.AuthMethodUsernamePassword},
					AuthenticationTime:    time.Now().UTC(),
				}
				require.NoError(t, sessionRepo.Create(ctx, testSession))
				sessionCookie = testSession.SessionID
			}

			req := httptest.NewRequest(http.MethodPost, "/oidc/v1/logout", nil)

			if sessionCookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: sessionCookie,
				})
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

// TestHandleEndSession_MissingParams validates error when neither id_token_hint nor client_id provided.
func TestHandleEndSession_MissingParams(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/oidc/v1/endsession", nil) // No params

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandleEndSession_WithClientID validates success with client_id parameter.
func TestHandleEndSession_WithClientID(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/oidc/v1/endsession?client_id=test-client", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Should return 200 (session cleared, no redirect URI)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
