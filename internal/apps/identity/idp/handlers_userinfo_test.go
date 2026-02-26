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
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleUserInfo_GET validates GET /userinfo endpoint returns user claims.
func TestHandleUserInfo_GET(t *testing.T) {
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
		Name:  "Test User",
	}
	userRepo := repoFactory.UserRepository()
	require.NoError(t, userRepo.Create(ctx, testUser))

	// Create test session.
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
	sessionRepo := repoFactory.SessionRepository()
	require.NoError(t, sessionRepo.Create(ctx, testSession))

	tests := []struct {
		name           string
		sessionCookie  string
		expectedStatus int
	}{
		{
			name:           "Valid session cookie",
			sessionCookie:  testSession.SessionID,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing session cookie",
			sessionCookie:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid session cookie",
			sessionCookie:  googleUuid.Must(googleUuid.NewV7()).String(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/oidc/v1/userinfo", nil)

			if tc.sessionCookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tc.sessionCookie,
				})
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
