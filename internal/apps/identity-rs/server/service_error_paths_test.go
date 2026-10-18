// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testScopeRoutePath             = "/scope"
	testScopeReadValue             = "read:resource"
	testScopeReadAndWriteValue     = "read:resource write:resource"
	testScopeWriteValue            = "write:resource"
	testServiceCORSPath            = "/cors"
	testServiceCORSOriginHeader    = "Access-Control-Allow-Origin"
	testServiceCORSMethodsHeader   = "Access-Control-Allow-Methods"
	testServiceCORSHeadersHeader   = "Access-Control-Allow-Headers"
	testServiceAuthorizationHeader = "Authorization"
)

type failingTokenService struct {
	validateErr error
}

func (s *failingTokenService) ValidateAccessToken(context.Context, string) (map[string]any, error) {
	return nil, s.validateErr
}

func (*failingTokenService) IsTokenActive(map[string]any) bool {
	return true
}

func (*failingTokenService) IntrospectToken(context.Context, string) (*cryptoutilIdentityIssuer.TokenMetadata, error) {
	return nil, nil
}

func TestService_RequireScopesValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		claims         map[string]any
		requiredScopes []string
		wantStatus     int
	}{
		{name: "claims missing", requiredScopes: []string{testScopeReadValue}, wantStatus: http.StatusUnauthorized},
		{
			name:           "scope claim missing",
			claims:         map[string]any{cryptoutilSharedMagic.ClaimClientID: "client"},
			requiredScopes: []string{testScopeReadValue},
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "scope claim not a string",
			claims:         map[string]any{cryptoutilSharedMagic.ClaimScope: 42},
			requiredScopes: []string{testScopeReadValue},
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "later required scope missing",
			claims:         map[string]any{cryptoutilSharedMagic.ClaimScope: testScopeReadValue},
			requiredScopes: []string{testScopeReadValue, testScopeWriteValue},
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "all required scopes granted",
			claims:         map[string]any{cryptoutilSharedMagic.ClaimScope: testScopeReadAndWriteValue},
			requiredScopes: []string{testScopeReadValue, testScopeWriteValue},
			wantStatus:     http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := NewService(
				&cryptoutilIdentityConfig.Config{},
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				nil,
			)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get(
				testScopeRoutePath,
				func(c *fiber.Ctx) error {
					if tc.claims != nil {
						c.Locals("token_claims", tc.claims)
					}

					return c.Next()
				},
				service.RequireScopes(tc.requiredScopes...),
				func(c *fiber.Ctx) error {
					return c.SendStatus(http.StatusOK)
				},
			)

			response, err := app.Test(httptest.NewRequest(http.MethodGet, testScopeRoutePath, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)
		})
	}
}

func TestService_ProtectedHandlersRejectMissingClaims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		method  string
		path    string
		handler func(*Service) fiber.Handler
	}{
		{
			name:   "get protected resource",
			method: http.MethodGet,
			path:   "/protected",
			handler: func(s *Service) fiber.Handler {
				return s.handleProtectedResource
			},
		},
		{
			name:   "create resource",
			method: http.MethodPost,
			path:   "/create",
			handler: func(s *Service) fiber.Handler {
				return s.handleCreateResource
			},
		},
		{
			name:   "delete resource",
			method: http.MethodDelete,
			path:   "/delete",
			handler: func(s *Service) fiber.Handler {
				return s.handleDeleteResource
			},
		},
		{
			name:   "admin users",
			method: http.MethodGet,
			path:   "/users",
			handler: func(s *Service) fiber.Handler {
				return s.handleAdminUsers
			},
		},
		{
			name:   "admin metrics",
			method: http.MethodGet,
			path:   "/metrics",
			handler: func(s *Service) fiber.Handler {
				return s.handleAdminMetrics
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := NewService(
				&cryptoutilIdentityConfig.Config{},
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				nil,
			)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Add(tc.method, tc.path, tc.handler(service))

			response, err := app.Test(httptest.NewRequest(tc.method, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
	}
}

func TestService_RegisterMiddlewareHandlesCORSPreflight(t *testing.T) {
	t.Parallel()

	service := NewService(
		&cryptoutilIdentityConfig.Config{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	service.RegisterMiddleware(app)
	app.Options(testServiceCORSPath, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(http.MethodOptions, testServiceCORSPath, nil), -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "*", response.Header.Get(testServiceCORSOriginHeader))
	require.Contains(t, response.Header.Get(testServiceCORSMethodsHeader), http.MethodDelete)
	require.Contains(t, response.Header.Get(testServiceCORSHeadersHeader), testServiceAuthorizationHeader)
}

func TestTokenValidator_ValidationServiceFailure(t *testing.T) {
	t.Parallel()

	validator := NewTokenValidator(
		&failingTokenService{validateErr: errors.New("injected token validation failure")},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/validate", validator.ValidateToken(), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/validate", nil)
	request.Header.Set("Authorization", cryptoutilSharedMagic.AuthorizationBearerPrefix+"valid-token")

	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
}
