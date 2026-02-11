// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

// TestHandleLogin_GET validates the GET /login endpoint displays the login page.
func TestHandleLogin_GET(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run database migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	// Initialize IDP service.
	config := &cryptoutilIdentityConfig.Config{
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "idp",
			BindAddress: "127.0.0.1",
			Port:        8080,
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

	tests := []struct {
		name             string
		requestID        string
		expectedStatus   int
		expectedContains string
	}{
		{
			name:             "Valid request ID",
			requestID:        googleUuid.Must(googleUuid.NewV7()).String(),
			expectedStatus:   http.StatusOK,
			expectedContains: "request_id",
		},
		{
			name:             "Missing request ID",
			requestID:        "",
			expectedStatus:   http.StatusBadRequest,
			expectedContains: "request_id is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reqURL := "/oidc/v1/login"
			if tc.requestID != "" {
				reqURL += "?request_id=" + tc.requestID
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

// TestHandleLoginSubmit_POST validates POST /login endpoint authenticates users.
func TestHandleLoginSubmit_POST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run database migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	// Initialize IDP service.
	config := &cryptoutilIdentityConfig.Config{
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "idp",
			BindAddress: "127.0.0.1",
			Port:        8080,
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

	// Create test client.
	testClient := &cryptoutilIdentityDomain.Client{
		ClientID:     "test-client",
		RedirectURIs: []string{"https://client.example.com/callback"},
	}
	clientRepo := repoFactory.ClientRepository()
	require.NoError(t, clientRepo.Create(ctx, testClient))

	// Create test authorization request.
	authzReq := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     testClient.ClientID,
		RedirectURI:  testClient.RedirectURIs[0],
		ResponseType: "code",
		Scope:        "openid profile",
		State:        "test-state",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	}
	authzReqRepo := repoFactory.AuthorizationRequestRepository()
	require.NoError(t, authzReqRepo.Create(ctx, authzReq))

	tests := []struct {
		name           string
		username       string
		password       string
		requestID      string
		expectedStatus int
	}{
		{
			name:           "Missing username",
			username:       "",
			password:       "password123",
			requestID:      authzReq.ID.String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing password",
			username:       "testuser",
			password:       "",
			requestID:      authzReq.ID.String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing request_id",
			username:       "testuser",
			password:       "password123",
			requestID:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid request_id format",
			username:       "testuser",
			password:       "password123",
			requestID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Nonexistent request_id",
			username:       "testuser",
			password:       "password123",
			requestID:      googleUuid.Must(googleUuid.NewV7()).String(),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			formData := url.Values{}
			formData.Set("username", tc.username)
			formData.Set("password", tc.password)
			formData.Set("request_id", tc.requestID)

			req := httptest.NewRequest(
				http.MethodPost,
				"/oidc/v1/login",
				strings.NewReader(formData.Encode()),
			)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
