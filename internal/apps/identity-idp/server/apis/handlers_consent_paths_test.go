// Copyright (c) 2025-2026 Justin Cranford.

package apis

import (
	"context"
	"io"
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
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	consentTestClientID                   = "consent-test-client"
	consentTestClientURL                  = "https://client.example/callback"
	consentTestExpiredRequestCaseName     = "expired request"
	consentTestMalformedRequestIDCaseName = "malformed request ID"
	consentTestMissingRequestIDCaseName   = "missing request ID"
	consentTestRequestNotFoundCaseName    = "request not found"
	consentTestScope                      = "openid profile"
	consentTestState                      = "consent-test-state"
	consentTestUnauthenticatedCaseName    = "user not authenticated"
	consentTestUsername                   = "consent-test-user"
)

func newAPITestService(t *testing.T) (*Service, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) {
	t.Helper()

	ctx := context.Background()
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	})
	require.NoError(t, err)
	require.NoError(t, repoFactory.DB().AutoMigrate(
		&cryptoutilIdentityDomain.AuthorizationRequest{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.ConsentDecision{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
		&cryptoutilIdentityDomain.Session{},
		&cryptoutilIdentityDomain.User{},
	))

	service := NewService(&cryptoutilIdentityConfig.Config{
		Sessions: &cryptoutilIdentityConfig.SessionConfig{CookieName: testSessionCookieName},
	}, repoFactory, nil)

	return service, repoFactory, ctx
}

func createConsentTestUser(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) *cryptoutilIdentityDomain.User {
	t.Helper()

	user := &cryptoutilIdentityDomain.User{
		Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
		PreferredUsername: consentTestUsername,
	}
	require.NoError(t, repoFactory.UserRepository().Create(ctx, user))

	return user
}

func createConsentTestClient(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) *cryptoutilIdentityDomain.Client {
	t.Helper()

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     consentTestClientID,
		Name:         "Consent Test Client",
		RedirectURIs: []string{consentTestClientURL},
	}
	require.NoError(t, repoFactory.ClientRepository().Create(ctx, client))

	return client
}

func createConsentTestRequest(
	t *testing.T,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	ctx context.Context,
	userID *googleUuid.UUID,
	expiresAt time.Time,
) *cryptoutilIdentityDomain.AuthorizationRequest {
	t.Helper()

	request := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.Must(googleUuid.NewV7()),
		ClientID:            consentTestClientID,
		RedirectURI:         consentTestClientURL,
		ResponseType:        cryptoutilSharedMagic.ResponseTypeCode,
		Scope:               consentTestScope,
		State:               consentTestState,
		CodeChallenge:       "consent-test-code-challenge",
		CodeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           expiresAt,
	}
	if userID != nil {
		request.UserID = cryptoutilIdentityDomain.NewNullableUUID(userID)
	}

	require.NoError(t, repoFactory.AuthorizationRequestRepository().Create(ctx, request))

	return request
}

func consentResponse(t *testing.T, app *fiber.App, method, target, body string) *http.Response {
	t.Helper()

	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if method == http.MethodPost {
		request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)
	}

	response, err := app.Test(request, -1)
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })

	return response
}

func responseBody(t *testing.T, response *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return string(body)
}

func TestHandleConsent_RequestPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string
		status         int
		bodyContains   string
		locationPrefix string
	}{
		{name: consentTestMissingRequestIDCaseName, status: fiber.StatusBadRequest, bodyContains: valueRequestIDIsRequired},
		{name: consentTestMalformedRequestIDCaseName, setup: func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string {
			return testInvalidUUIDValue
		}, status: fiber.StatusBadRequest, bodyContains: valueInvalidRequestIDFormat},
		{name: consentTestRequestNotFoundCaseName, setup: func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string {
			return googleUuid.Must(googleUuid.NewV7()).String()
		}, status: fiber.StatusBadRequest, bodyContains: valueAuthorizationRequestNotFoundOrExpired},
		{
			name: consentTestExpiredRequestCaseName,
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createConsentTestRequest(t, repoFactory, ctx, nil, time.Now().UTC().Add(-time.Hour)).ID.String()
			},
			status: fiber.StatusBadRequest, bodyContains: valueAuthorizationRequestHasExpired,
		},
		{
			name: consentTestUnauthenticatedCaseName,
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createConsentTestRequest(t, repoFactory, ctx, nil, time.Now().UTC().Add(time.Hour)).ID.String()
			},
			status: fiber.StatusBadRequest, bodyContains: valueUserNotAuthenticated,
		},
		{
			name: "client not found",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				user := createConsentTestUser(t, repoFactory, ctx)

				return createConsentTestRequest(t, repoFactory, ctx, &user.ID, time.Now().UTC().Add(time.Hour)).ID.String()
			},
			status: fiber.StatusBadRequest, bodyContains: valueClientNotFound,
		},
		{
			name: "user not found",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				createConsentTestClient(t, repoFactory, ctx)

				missingUserID := googleUuid.Must(googleUuid.NewV7())

				return createConsentTestRequest(t, repoFactory, ctx, &missingUserID, time.Now().UTC().Add(time.Hour)).ID.String()
			},
			status: fiber.StatusBadRequest, bodyContains: valueUserNotFound,
		},
		{
			name: "renders consent page",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				user := createConsentTestUser(t, repoFactory, ctx)
				createConsentTestClient(t, repoFactory, ctx)

				return createConsentTestRequest(t, repoFactory, ctx, &user.ID, time.Now().UTC().Add(time.Hour)).ID.String()
			},
			status: fiber.StatusOK, bodyContains: consentTestUsername,
		},
		{
			name: "existing consent redirects with authorization code",
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				user := createConsentTestUser(t, repoFactory, ctx)
				createConsentTestClient(t, repoFactory, ctx)
				request := createConsentTestRequest(t, repoFactory, ctx, &user.ID, time.Now().UTC().Add(time.Hour))
				require.NoError(t, repoFactory.ConsentDecisionRepository().Create(ctx, &cryptoutilIdentityDomain.ConsentDecision{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    user.ID,
					ClientID:  consentTestClientID,
					Scope:     consentTestScope,
					GrantedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}))

				return request.ID.String()
			},
			status: fiber.StatusFound, locationPrefix: consentTestClientURL + "?code=",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, repoFactory, ctx := newAPITestService(t)

			requestID := ""
			if tc.setup != nil {
				requestID = tc.setup(t, repoFactory, ctx)
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/consent", service.handleConsent)

			target := "/consent"
			if requestID != "" {
				target += "?request_id=" + url.QueryEscape(requestID)
			}

			response := consentResponse(t, app, http.MethodGet, target, "")

			defer func() { _ = response.Body.Close() }()

			require.Equal(t, tc.status, response.StatusCode)

			if tc.bodyContains != "" {
				require.Contains(t, responseBody(t, response), tc.bodyContains)
			}

			if tc.locationPrefix != "" {
				require.True(t, strings.HasPrefix(response.Header.Get(fiber.HeaderLocation), tc.locationPrefix))
			}
		})
	}
}

func TestHandleConsentSubmit_RequestPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestID      string
		decision       string
		setup          func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory, context.Context) string
		status         int
		bodyContains   string
		locationPrefix string
	}{
		{name: consentTestMissingRequestIDCaseName, decision: testConsentDecisionApprove, status: fiber.StatusBadRequest, bodyContains: valueRequestIDIsRequired},
		{name: "missing decision", requestID: googleUuid.Must(googleUuid.NewV7()).String(), status: fiber.StatusBadRequest, bodyContains: "decision is required"},
		{name: "consent denied", requestID: "unused", decision: "deny", status: fiber.StatusForbidden, bodyContains: "User denied consent"},
		{name: consentTestMalformedRequestIDCaseName, requestID: testInvalidUUIDValue, decision: testConsentDecisionApprove, status: fiber.StatusBadRequest, bodyContains: valueInvalidRequestIDFormat},
		{name: consentTestRequestNotFoundCaseName, requestID: googleUuid.Must(googleUuid.NewV7()).String(), decision: testConsentDecisionApprove, status: fiber.StatusBadRequest, bodyContains: valueAuthorizationRequestNotFoundOrExpired},
		{
			name:     consentTestExpiredRequestCaseName,
			decision: testConsentDecisionApprove,
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createConsentTestRequest(t, repoFactory, ctx, nil, time.Now().UTC().Add(-time.Hour)).ID.String()
			},
			status: fiber.StatusBadRequest, bodyContains: valueAuthorizationRequestHasExpired,
		},
		{
			name:     consentTestUnauthenticatedCaseName,
			decision: testConsentDecisionApprove,
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				return createConsentTestRequest(t, repoFactory, ctx, nil, time.Now().UTC().Add(time.Hour)).ID.String()
			},
			status: fiber.StatusBadRequest, bodyContains: valueUserNotAuthenticated,
		},
		{
			name:     "approved consent redirects with authorization code",
			decision: testConsentDecisionApprove,
			setup: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, ctx context.Context) string {
				user := createConsentTestUser(t, repoFactory, ctx)
				request := createConsentTestRequest(t, repoFactory, ctx, &user.ID, time.Now().UTC().Add(time.Hour))

				return request.ID.String()
			},
			status: fiber.StatusFound, locationPrefix: consentTestClientURL + "?code=",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, repoFactory, ctx := newAPITestService(t)

			requestID := tc.requestID
			if tc.setup != nil {
				requestID = tc.setup(t, repoFactory, ctx)
			}

			form := url.Values{}
			if requestID != "" {
				form.Set("request_id", requestID)
			}

			if tc.decision != "" {
				form.Set("decision", tc.decision)
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post("/consent", service.handleConsentSubmit)
			response := consentResponse(t, app, http.MethodPost, "/consent", form.Encode())

			defer func() { _ = response.Body.Close() }()

			require.Equal(t, tc.status, response.StatusCode)

			if tc.bodyContains != "" {
				require.Contains(t, responseBody(t, response), tc.bodyContains)
			}

			if tc.locationPrefix != "" {
				require.True(t, strings.HasPrefix(response.Header.Get(fiber.HeaderLocation), tc.locationPrefix))
			}
		})
	}
}
