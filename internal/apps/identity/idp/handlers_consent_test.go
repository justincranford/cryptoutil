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

// TestHandleConsent_GET validates GET /consent endpoint displays consent page.
func TestHandleConsent_GET(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
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
			expectedStatus:   http.StatusUnauthorized,
			expectedContains: "User not authenticated",
		},
		{
			name:             "Missing request ID",
			requestID:        "",
			expectedStatus:   http.StatusUnauthorized,
			expectedContains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reqURL := "/oidc/v1/consent"
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

// TestHandleConsentSubmit_POST validates POST /consent endpoint processes consent decisions.
func TestHandleConsentSubmit_POST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
		&cryptoutilIdentityDomain.AuthorizationRequest{},
		&cryptoutilIdentityDomain.ConsentDecision{},
	)
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
		requestID      string
		approvedScopes string
		expectedStatus int
	}{
		{
			name:           "Missing request_id",
			requestID:      "",
			approvedScopes: "openid profile",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid request_id format",
			requestID:      "invalid-uuid",
			approvedScopes: "openid profile",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Nonexistent request_id",
			requestID:      googleUuid.Must(googleUuid.NewV7()).String(),
			approvedScopes: "openid profile",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			formData := url.Values{}
			formData.Set("request_id", tc.requestID)
			formData.Set("approved_scopes", tc.approvedScopes)

			req := httptest.NewRequest(
				http.MethodPost,
				"/oidc/v1/consent",
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
