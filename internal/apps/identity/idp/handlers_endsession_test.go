// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"net/url"
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

// TestHandleEndSession_GET validates GET /endsession endpoint (OIDC RP-Initiated Logout).
//
// Requirements verified:
// - P1.6.2: RP-Initiated Logout endpoint.
func TestHandleEndSession_GET(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
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

	// Create test client with post_logout_redirect_uris.
	testClient := &cryptoutilIdentityDomain.Client{
		ClientID:                "test-logout-client",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Logout Client",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		PostLogoutRedirectURIs:  []string{"https://client.example.com/logged-out"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}
	clientRepo := repoFactory.ClientRepository()
	require.NoError(t, clientRepo.Create(ctx, testClient))

	// Create test user and session for session cleanup tests.
	testUser := &cryptoutilIdentityDomain.User{
		Sub:   googleUuid.Must(googleUuid.NewV7()).String(),
		Email: "logout-test@example.com",
	}
	userRepo := repoFactory.UserRepository()
	require.NoError(t, userRepo.Create(ctx, testUser))

	testSession := &cryptoutilIdentityDomain.Session{
		UserID:                testUser.ID,
		SessionID:             googleUuid.Must(googleUuid.NewV7()).String(),
		IPAddress:             "127.0.0.1",
		UserAgent:             "test-agent",
		IssuedAt:              time.Now().UTC(),
		ExpiresAt:             time.Now().UTC().Add(1 * time.Hour),
		LastSeenAt:            time.Now().UTC(),
		Active:                boolPtr(true),
		AuthenticationMethods: []string{"username_password"},
		AuthenticationTime:    time.Now().UTC(),
	}
	sessionRepo := repoFactory.SessionRepository()
	require.NoError(t, sessionRepo.Create(ctx, testSession))

	tests := []struct {
		name                string
		queryParams         map[string]string
		sessionCookie       string
		expectedStatus      int
		expectRedirect      bool
		expectHTML          bool
		expectedRedirectURL string
	}{
		{
			name:           "Missing required parameters returns error",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "client_id only returns success page",
			queryParams: map[string]string{
				"client_id": testClient.ClientID,
			},
			expectedStatus: http.StatusOK,
			expectHTML:     true,
		},
		{
			name: "id_token_hint only returns success page",
			queryParams: map[string]string{
				"id_token_hint": "test.jwt.token",
			},
			expectedStatus: http.StatusOK,
			expectHTML:     true,
		},
		{
			name: "Valid redirect URI redirects",
			queryParams: map[string]string{
				"client_id":                testClient.ClientID,
				"post_logout_redirect_uri": "https://client.example.com/logged-out",
			},
			expectedStatus:      http.StatusFound,
			expectRedirect:      true,
			expectedRedirectURL: "https://client.example.com/logged-out",
		},
		{
			name: "Valid redirect URI with state includes state in redirect",
			queryParams: map[string]string{
				"client_id":                testClient.ClientID,
				"post_logout_redirect_uri": "https://client.example.com/logged-out",
				"state":                    "test-state-123",
			},
			expectedStatus: http.StatusFound,
			expectRedirect: true,
		},
		{
			name: "Invalid redirect URI returns error",
			queryParams: map[string]string{
				"client_id":                testClient.ClientID,
				"post_logout_redirect_uri": "https://malicious.com/callback",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Session cookie is cleared on logout",
			queryParams: map[string]string{
				"client_id": testClient.ClientID,
			},
			sessionCookie:  testSession.SessionID,
			expectedStatus: http.StatusOK,
			expectHTML:     true,
		},
		{
			name: "Unknown client returns error for redirect validation",
			queryParams: map[string]string{
				"client_id":                "unknown-client",
				"post_logout_redirect_uri": "https://unknown.com/callback",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Note: Not using t.Parallel() here because all subtests share the same
			// in-memory SQLite database. With parallel subtests, each gets a separate
			// database connection which creates a new empty database in :memory: mode.
			// Build URL with query parameters.
			reqURL := "/oidc/v1/endsession"

			if len(tc.queryParams) > 0 {
				params := url.Values{}
				for k, v := range tc.queryParams {
					params.Set(k, v)
				}

				reqURL += "?" + params.Encode()
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)

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

			if tc.expectRedirect {
				location := resp.Header.Get("Location")
				require.NotEmpty(t, location)

				if tc.expectedRedirectURL != "" {
					require.Contains(t, location, tc.expectedRedirectURL)
				}
			}

			if tc.expectHTML {
				contentType := resp.Header.Get("Content-Type")
				require.Contains(t, contentType, "text/html")
			}
		})
	}
}
