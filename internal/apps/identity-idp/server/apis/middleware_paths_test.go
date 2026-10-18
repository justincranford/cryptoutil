// Copyright (c) 2025-2026 Justin Cranford.

package apis

import (
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
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newAPITestTokenService() *cryptoutilIdentityIssuer.TokenService {
	return cryptoutilIdentityIssuer.NewTokenService(
		nil,
		nil,
		cryptoutilIdentityIssuer.NewUUIDIssuer(),
		&cryptoutilIdentityConfig.TokenConfig{AccessTokenFormat: cryptoutilSharedMagic.TokenFormatUUID},
	)
}

func createMiddlewareTestSession(
	t *testing.T,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	ctx context.Context,
	active bool,
	expiresAt time.Time,
) *cryptoutilIdentityDomain.Session {
	t.Helper()

	session := &cryptoutilIdentityDomain.Session{
		SessionID:             googleUuid.Must(googleUuid.NewV7()).String(),
		UserID:                googleUuid.Must(googleUuid.NewV7()),
		IssuedAt:              time.Now().UTC(),
		ExpiresAt:             expiresAt,
		Active:                &active,
		AuthenticationTime:    time.Now().UTC(),
		AuthenticationMethods: []string{"password"},
	}
	require.NoError(t, repoFactory.SessionRepository().Create(ctx, session))

	return session
}

func runMiddlewareRequest(t *testing.T, app *fiber.App, authorization, sessionID string) int {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	if authorization != "" {
		request.Header.Set(fiber.HeaderAuthorization, authorization)
	}

	if sessionID != "" {
		request.AddCookie(&http.Cookie{
			Name:     testSessionCookieName,
			Value:    sessionID,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}

	response, err := app.Test(request, -1)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }()

	return response.StatusCode
}

func TestAuthMiddleware_RequestPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string
		statusCode int
	}{
		{name: "missing cookie", statusCode: fiber.StatusUnauthorized},
		{
			name: "unknown session",
			setup: func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string {
				return googleUuid.Must(googleUuid.NewV7()).String()
			},
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name: "inactive session",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createMiddlewareTestSession(t, repoFactory, ctx, false, time.Now().UTC().Add(time.Hour)).SessionID
			},
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name: "expired session",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createMiddlewareTestSession(t, repoFactory, ctx, true, time.Now().UTC().Add(-time.Hour)).SessionID
			},
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name: "active session",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createMiddlewareTestSession(t, repoFactory, ctx, true, time.Now().UTC().Add(time.Hour)).SessionID
			},
			statusCode: fiber.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, repoFactory, ctx := newAPITestService(t)

			sessionID := ""
			if tc.setup != nil {
				sessionID = tc.setup(t, repoFactory, ctx)
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/", service.AuthMiddleware(), func(c *fiber.Ctx) error {
				require.NotNil(t, c.Locals("session"))

				return c.SendStatus(fiber.StatusNoContent)
			})

			statusCode := runMiddlewareRequest(t, app, "", sessionID)
			require.Equal(t, tc.statusCode, statusCode)
		})
	}
}

func TestTokenAuthMiddleware_RequestPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		authorization string
		statusCode    int
	}{
		{name: "missing header", statusCode: fiber.StatusUnauthorized},
		{name: "invalid header scheme", authorization: testBasicAuthorizationHeader, statusCode: fiber.StatusUnauthorized},
		{name: "missing bearer token", authorization: cryptoutilSharedMagic.AuthorizationBearer, statusCode: fiber.StatusUnauthorized},
		{name: testInvalidUUIDTokenCase, authorization: testInvalidBearerAuthorizationHeader, statusCode: fiber.StatusUnauthorized},
		{name: "valid UUID token", authorization: cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix + googleUuid.Must(googleUuid.NewV7()).String(), statusCode: fiber.StatusNoContent},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, _, _ := newAPITestService(t)
			service.tokenSvc = newAPITestTokenService()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/", service.TokenAuthMiddleware(), func(c *fiber.Ctx) error {
				require.NotNil(t, c.Locals("claims"))

				return c.SendStatus(fiber.StatusNoContent)
			})

			statusCode := runMiddlewareRequest(t, app, tc.authorization, "")
			require.Equal(t, tc.statusCode, statusCode)
		})
	}
}

func TestHybridAuthMiddleware_RequestPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		authorization string
		setup         func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string
		wantAuth      string
		statusCode    int
	}{
		{name: "no credentials", statusCode: fiber.StatusUnauthorized},
		{name: "unsupported authorization falls through", authorization: testBasicAuthorizationHeader, statusCode: fiber.StatusUnauthorized},
		{name: "invalid bearer is rejected before cookie", authorization: testInvalidBearerAuthorizationHeader, statusCode: fiber.StatusUnauthorized},
		{
			name:          "valid bearer",
			authorization: cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix + googleUuid.Must(googleUuid.NewV7()).String(),
			wantAuth:      cryptoutilSharedMagic.ClientAuthMethodBearerToken,
			statusCode:    fiber.StatusNoContent,
		},
		{
			name: "unknown session cookie",
			setup: func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string {
				return googleUuid.Must(googleUuid.NewV7()).String()
			},
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name: "inactive session cookie",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createMiddlewareTestSession(t, repoFactory, ctx, false, time.Now().UTC().Add(time.Hour)).SessionID
			},
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name: "expired session cookie",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createMiddlewareTestSession(t, repoFactory, ctx, true, time.Now().UTC().Add(-time.Hour)).SessionID
			},
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name: "active session cookie",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createMiddlewareTestSession(t, repoFactory, ctx, true, time.Now().UTC().Add(time.Hour)).SessionID
			},
			wantAuth:   "session_cookie",
			statusCode: fiber.StatusNoContent,
		},
		{
			name:          "bearer authentication takes precedence",
			authorization: cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix + googleUuid.Must(googleUuid.NewV7()).String(),
			setup: func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string {
				return googleUuid.Must(googleUuid.NewV7()).String()
			},
			wantAuth:   cryptoutilSharedMagic.ClientAuthMethodBearerToken,
			statusCode: fiber.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, repoFactory, ctx := newAPITestService(t)
			service.tokenSvc = newAPITestTokenService()

			sessionID := ""
			if tc.setup != nil {
				sessionID = tc.setup(t, repoFactory, ctx)
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/", service.HybridAuthMiddleware(), func(c *fiber.Ctx) error {
				require.NotNil(t, c.Locals("claims"))
				require.Equal(t, tc.wantAuth, c.Locals("auth_method"))

				return c.SendStatus(fiber.StatusNoContent)
			})

			statusCode := runMiddlewareRequest(t, app, tc.authorization, sessionID)
			require.Equal(t, tc.statusCode, statusCode)
		})
	}
}
