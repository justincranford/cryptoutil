// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerMiddleware "cryptoutil/internal/apps-framework/service/server/middleware"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps-framework/service/server/service"
	cryptoutilAppsFrameworkServiceTestHelpDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gorm.io/gorm"
)

const (
	testServiceClientSecret = "test-service-client-secret-with-at-least-128-bits"
	testServiceClientToken  = "service-client-session-token"
	testServiceSessionToken = "service-session-validation-token"
)

func TestClientAuthenticationRoutes_RegisterAndExchangeClientCredentials(t *testing.T) {
	t.Parallel()

	db := newClientAuthTestDB(t)
	clientRepository := cryptoutilAppsFrameworkServiceServerRepository.NewClientRepository(db)
	tenantRepository := cryptoutilAppsFrameworkServiceServerRepository.NewTenantRepository(db)
	registrationService := newClientAuthRegistrationService(db, clientRepository, tenantRepository)

	var (
		issuedClientID string
		issuedTenantID googleUuid.UUID
		issuedRealmID  googleUuid.UUID
	)

	sessionManager := newMockSessionManagerSuccess()
	sessionManager.serviceSessionExpiration = 15 * time.Minute
	sessionManager.issueServiceFunc = func(
		_ context.Context,
		clientID string,
		tenantID googleUuid.UUID,
		realmID googleUuid.UUID,
	) (string, error) {
		issuedClientID = clientID
		issuedTenantID = tenantID
		issuedRealmID = realmID

		return testServiceClientToken, nil
	}

	app := newClientAuthTestApp()
	app.Use(
		cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath+"/*",
		cryptoutilAppsFrameworkServiceServerMiddleware.ServiceSessionMiddleware(sessionManager),
	)
	RegisterClientAuthenticationRoutes(
		app,
		registrationService,
		clientRepository,
		tenantRepository,
		sessionManager,
		nil,
		cryptoutilSharedMagic.RateLimitDefaultRequestsPerMin,
	)
	app.Get(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath+"/protected", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	tenantUUID, err := googleUuid.NewV7()
	require.NoError(t, err)
	registerBody, err := json.Marshal(registerClientRequest{TenantName: "load-" + tenantUUID.String()})
	require.NoError(t, err)

	registerRequest := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath,
		bytes.NewReader(registerBody),
	)
	registerRequest.Header.Set("Content-Type", "application/json")

	registerResponse, err := app.Test(registerRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, registerResponse.Body.Close()) }()

	require.Equal(t, fiber.StatusCreated, registerResponse.StatusCode)

	var registered registerClientResponse
	require.NoError(t, json.NewDecoder(registerResponse.Body).Decode(&registered))
	require.NotEmpty(t, registered.ClientID)
	require.NotEmpty(t, registered.ClientSecret)
	require.NotEmpty(t, registered.TenantID)

	client, err := clientRepository.GetByClientID(t.Context(), registered.ClientID)
	require.NoError(t, err)
	validSecret, err := cryptoutilSharedCryptoHash.VerifySecretHKDFRandom(client.ClientSecretHash, registered.ClientSecret)
	require.NoError(t, err)
	require.True(t, validSecret)

	tokenRequest := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
		nil,
	)
	tokenRequest.SetBasicAuth(registered.ClientID, registered.ClientSecret)

	tokenRequestStarted := time.Now().UTC()

	tokenResponse, err := app.Test(tokenRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, tokenResponse.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, tokenResponse.StatusCode)

	var tokenResult struct {
		Token     string `json:"token"`
		TokenType string `json:"token_type"`
		ExpiresAt string `json:"expires_at"`
	}
	require.NoError(t, json.NewDecoder(tokenResponse.Body).Decode(&tokenResult))
	require.Equal(t, testServiceClientToken, tokenResult.Token)
	require.Equal(t, cryptoutilSharedMagic.AuthorizationBearer, tokenResult.TokenType)
	expiresAt, err := time.Parse(time.RFC3339, tokenResult.ExpiresAt)
	require.NoError(t, err)
	require.WithinDuration(t, tokenRequestStarted.Add(sessionManager.ServiceSessionExpiration()), expiresAt, time.Second)
	require.Equal(t, registered.ClientID, issuedClientID)
	require.Equal(t, googleUuid.MustParse(registered.TenantID), issuedTenantID)
	require.True(t, issuedRealmID == googleUuid.Nil)

	protectedRequest := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath+"/protected",
		nil,
	)
	protectedResponse, err := app.Test(protectedRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, protectedResponse.Body.Close()) }()

	require.Equal(t, fiber.StatusUnauthorized, protectedResponse.StatusCode)

	authenticatedRequest := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath+"/protected",
		nil,
	)
	authenticatedRequest.Header.Set(
		"Authorization",
		cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix+tokenResult.Token,
	)

	authenticatedResponse, err := app.Test(authenticatedRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, authenticatedResponse.Body.Close()) }()

	require.Equal(t, fiber.StatusNoContent, authenticatedResponse.StatusCode)
}

func TestClientAuthenticationRoutes_RejectInvalidCredentials(t *testing.T) {
	t.Parallel()

	db := newClientAuthTestDB(t)
	clientRepository := cryptoutilAppsFrameworkServiceServerRepository.NewClientRepository(db)
	tenantRepository := cryptoutilAppsFrameworkServiceServerRepository.NewTenantRepository(db)
	sessionManager := newMockSessionManagerSuccess()

	activeClient, activeTenant := seedServiceClientForAuthTest(t, clientRepository, tenantRepository, true, true)
	inactiveClient, _ := seedServiceClientForAuthTest(t, clientRepository, tenantRepository, false, true)
	inactiveTenantClient, _ := seedServiceClientForAuthTest(t, clientRepository, tenantRepository, true, false)
	unknownClientID, err := googleUuid.NewV7()
	require.NoError(t, err)

	handlers := newClientAuthenticationHandlersForTest(
		clientRepository,
		tenantRepository,
		sessionManager,
	)
	tests := []struct {
		name          string
		clientID      string
		clientSecret  string
		rawAuthHeader string
		wantHeader    bool
	}{
		{name: "missing authorization"},
		{name: "malformed authorization", rawAuthHeader: "Bearer token", wantHeader: true},
		{
			name:         "unknown client",
			clientID:     unknownClientID.String(),
			clientSecret: testServiceClientSecret,
		},
		{
			name:         "incorrect secret",
			clientID:     activeClient.ClientID,
			clientSecret: "incorrect-secret",
		},
		{
			name:         "inactive client",
			clientID:     inactiveClient.ClientID,
			clientSecret: testServiceClientSecret,
		},
		{
			name:         "inactive tenant",
			clientID:     inactiveTenantClient.ClientID,
			clientSecret: testServiceClientSecret,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post(
				cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
				handlers.handleIssueServiceToken,
			)

			req := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
				nil,
			)
			if tc.wantHeader {
				req.Header.Set("Authorization", tc.rawAuthHeader)
			} else if tc.clientID != "" {
				req.SetBasicAuth(tc.clientID, tc.clientSecret)
			}

			resp, err := app.Test(req, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

			var failure map[string]string
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&failure))
			require.Equal(
				t,
				map[string]string{cryptoutilSharedMagic.StringError: clientAuthErrorInvalidCredentials},
				failure,
			)
		})
	}

	require.True(t, activeClient.IsActive())
	require.True(t, activeTenant.IsActive())
}

func TestClientAuthenticationRoutes_RateLimiting(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	RegisterClientAuthenticationRoutes(app, nil, nil, nil, nil, nil, 1)

	var rateLimited bool

	for range cryptoutilSharedMagic.JoseJADefaultMaxMaterials {
		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
			nil,
		)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		if resp.StatusCode == fiber.StatusTooManyRequests {
			rateLimited = true

			break
		}
	}

	require.True(t, rateLimited, "client credential exchange should enforce its rate limit")
}

func newClientAuthTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var appErr *cryptoutilSharedApperr.Error
			if errors.As(err, &appErr) {
				return c.SendStatus(int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
			}

			return c.SendStatus(fiber.StatusInternalServerError)
		},
	})
}

func newClientAuthRegistrationService(
	db *gorm.DB,
	clientRepository cryptoutilAppsFrameworkServiceServerRepository.ClientRepository,
	tenantRepository cryptoutilAppsFrameworkServiceServerRepository.TenantRepository,
) cryptoutilAppsFrameworkServiceServerService.RegistrationService {
	roleRepository := cryptoutilAppsFrameworkServiceServerRepository.NewRoleRepository(db)

	return cryptoutilAppsFrameworkServiceServerService.NewRegistrationService(
		cryptoutilAppsFrameworkServiceServerService.NewTenantService(tenantRepository, roleRepository),
		cryptoutilAppsFrameworkServiceServerRepository.NewUserRepository(db),
		clientRepository,
		cryptoutilAppsFrameworkServiceServerRepository.NewUnverifiedUserRepository(db),
		cryptoutilAppsFrameworkServiceServerRepository.NewUnverifiedClientRepository(db),
		roleRepository,
		cryptoutilAppsFrameworkServiceServerRepository.NewUserRoleRepository(db),
		cryptoutilAppsFrameworkServiceServerRepository.NewClientRoleRepository(db),
	)
}

func newClientAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := cryptoutilAppsFrameworkServiceTestHelpDb.NewInMemorySQLiteDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	require.NoError(
		t,
		cryptoutilAppsFrameworkServiceServerRepository.ApplyMigrations(
			sqlDB,
			cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite,
			cryptoutilAppsFrameworkServiceServerRepository.MigrationsFS,
		),
	)

	return db
}

func seedServiceClientForAuthTest(
	t *testing.T,
	clientRepository cryptoutilAppsFrameworkServiceServerRepository.ClientRepository,
	tenantRepository cryptoutilAppsFrameworkServiceServerRepository.TenantRepository,
	clientActive bool,
	tenantActive bool,
) (*cryptoutilAppsFrameworkServiceServerRepository.Client, *cryptoutilAppsFrameworkServiceServerRepository.Tenant) {
	t.Helper()

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)

	tenant := &cryptoutilAppsFrameworkServiceServerRepository.Tenant{
		ID:     tenantID,
		Name:   "auth-test-" + tenantID.String(),
		Active: 1,
	}
	require.NoError(t, tenantRepository.Create(t.Context(), tenant))

	if !tenantActive {
		tenant.SetActive(false)
		require.NoError(t, tenantRepository.Update(t.Context(), tenant))
	}

	clientID, err := googleUuid.NewV7()
	require.NoError(t, err)
	clientSecretHash, err := cryptoutilSharedCryptoHash.HashHighEntropyNonDeterministic(testServiceClientSecret)
	require.NoError(t, err)

	client := &cryptoutilAppsFrameworkServiceServerRepository.Client{
		ID:               clientID,
		TenantID:         tenant.ID,
		ClientID:         clientID.String(),
		ClientSecretHash: clientSecretHash,
		Active:           1,
	}
	require.NoError(t, clientRepository.Create(t.Context(), client))

	if !clientActive {
		client.SetActive(false)
		require.NoError(t, clientRepository.Update(t.Context(), client))
	}

	return client, tenant
}

func newClientAuthenticationHandlersForTest(
	clientRepository cryptoutilAppsFrameworkServiceServerRepository.ClientRepository,
	tenantRepository cryptoutilAppsFrameworkServiceServerRepository.TenantRepository,
	sessionManager SessionManager,
) *clientAuthenticationHandlers {
	return &clientAuthenticationHandlers{
		clientRepository:      clientRepository,
		tenantRepository:      tenantRepository,
		sessionManager:        sessionManager,
		hashSecretFn:          cryptoutilSharedCryptoHash.HashHighEntropyNonDeterministic,
		verifySecretFn:        cryptoutilSharedCryptoHash.VerifySecretHKDFRandom,
		generateCredentialsFn: generateServiceClientCredentials,
	}
}
