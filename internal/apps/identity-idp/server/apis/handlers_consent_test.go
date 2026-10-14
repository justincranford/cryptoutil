// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis_test

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity-idp/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleConsent_GET validates GET /consent endpoint displays consent page.
func TestHandleConsent_GET(t *testing.T) {
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
			Port:        cryptoutilSharedMagic.TestServerPort,
			TLSEnabled:  true,
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			CookieName:      testSessionID,
			CookieHTTPOnly:  true,
			CookieSameSite:  testLax,
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
			name:             testValidRequestID,
			requestID:        googleUuid.Must(googleUuid.NewV7()).String(),
			expectedStatus:   http.StatusUnauthorized,
			expectedContains: "User not authenticated",
		},
		{
			name:             testMissingRequestID,
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
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
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
			Name:        cryptoutilSharedMagic.IDPServiceName,
			BindAddress: cryptoutilSharedMagic.IPv4Loopback,
			Port:        cryptoutilSharedMagic.TestServerPort,
			TLSEnabled:  true,
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			CookieName:      testSessionID,
			CookieHTTPOnly:  true,
			CookieSameSite:  testLax,
			SessionLifetime: 1 * time.Hour,
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)

	app := fiber.New()
	service.RegisterRoutes(app)

	// Create test user.
	testUser := &cryptoutilIdentityDomain.User{
		Sub:   googleUuid.Must(googleUuid.NewV7()).String(),
		Email: testTestExampleCom,
	}
	userRepo := repoFactory.UserRepository()
	require.NoError(t, userRepo.Create(ctx, testUser))

	// Create test client.
	testClient := &cryptoutilIdentityDomain.Client{
		ClientID:     testTestClient,
		RedirectURIs: []string{testHTTPSClientExampleComCallback},
	}
	clientRepo := repoFactory.ClientRepository()
	require.NoError(t, clientRepo.Create(ctx, testClient))

	// Create test authorization request.
	authzReq := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     testClient.ClientID,
		RedirectURI:  testClient.RedirectURIs[0],
		ResponseType: cryptoutilSharedMagic.ResponseTypeCode,
		Scope:        testOpenidProfile,
		State:        testTestState,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
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
			name:           testMissingRequestIDValueCFA47702,
			requestID:      "",
			approvedScopes: testOpenidProfile,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           testInvalidRequestIDFormat,
			requestID:      testInvalidUUID,
			approvedScopes: testOpenidProfile,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           testNonexistentRequestID,
			requestID:      googleUuid.Must(googleUuid.NewV7()).String(),
			approvedScopes: testOpenidProfile,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			formData := url.Values{}
			formData.Set(testRequestID, tc.requestID)
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
