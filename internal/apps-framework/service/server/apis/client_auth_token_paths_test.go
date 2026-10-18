// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"context"
	"encoding/base64"
	json "encoding/json"
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps-framework/service/server/service"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type clientAuthClientRepositoryStub struct {
	cryptoutilAppsFrameworkServiceServerRepository.ClientRepository
	client *cryptoutilAppsFrameworkServiceServerRepository.Client
	err    error
}

func (s *clientAuthClientRepositoryStub) GetByClientID(
	context.Context,
	string,
) (*cryptoutilAppsFrameworkServiceServerRepository.Client, error) {
	return s.client, s.err
}

type clientAuthTenantRepositoryStub struct {
	cryptoutilAppsFrameworkServiceServerRepository.TenantRepository
	tenant *cryptoutilAppsFrameworkServiceServerRepository.Tenant
	err    error
}

func (s *clientAuthTenantRepositoryStub) GetByID(
	context.Context,
	googleUuid.UUID,
) (*cryptoutilAppsFrameworkServiceServerRepository.Tenant, error) {
	return s.tenant, s.err
}

type clientAuthRealmServiceStub struct {
	cryptoutilAppsFrameworkServiceServerService.RealmService
	realm *cryptoutilAppsFrameworkServiceServerRepository.TenantRealm
	err   error
}

func (s *clientAuthRealmServiceStub) GetFirstActiveRealm(
	context.Context,
	googleUuid.UUID,
) (*cryptoutilAppsFrameworkServiceServerRepository.TenantRealm, error) {
	return s.realm, s.err
}

func TestHandleIssueServiceToken_ErrorPaths(t *testing.T) {
	t.Parallel()

	clientLookupErr := errors.New("injected client repository error")
	secretVerificationErr := errors.New("injected secret verification error")
	tenantLookupErr := errors.New("injected tenant repository error")
	realmLookupErr := errors.New("injected realm lookup error")
	sessionIssueErr := errors.New("injected session issue error")
	tenantNotFoundSummary := "tenant not found"
	tenantNotFoundErr := cryptoutilSharedApperr.NewHTTP404NotFound(
		&tenantNotFoundSummary,
		errors.New("tenant lookup found no record"),
	)
	realmID, err := googleUuid.NewV7()
	require.NoError(t, err)

	tests := []struct {
		name             string
		clientLookupErr  error
		secretVerifyErr  error
		tenantLookupErr  error
		realmService     bool
		realm            *cryptoutilAppsFrameworkServiceServerRepository.TenantRealm
		realmLookupErr   error
		sessionIssueErr  error
		wantStatus       int
		wantErrorMessage string
		wantRealmID      googleUuid.UUID
	}{
		{
			name:             "client repository error",
			clientLookupErr:  clientLookupErr,
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorMessage: clientAuthErrorSessionFailed,
		},
		{
			name:             "secret verification error",
			secretVerifyErr:  secretVerificationErr,
			wantStatus:       fiber.StatusUnauthorized,
			wantErrorMessage: clientAuthErrorInvalidCredentials,
		},
		{
			name:             "tenant repository error",
			tenantLookupErr:  tenantLookupErr,
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorMessage: clientAuthErrorSessionFailed,
		},
		{
			name:             "tenant not found is unauthorized",
			tenantLookupErr:  tenantNotFoundErr,
			wantStatus:       fiber.StatusUnauthorized,
			wantErrorMessage: clientAuthErrorInvalidCredentials,
		},
		{
			name:         "active realm included",
			realmService: true,
			realm:        &cryptoutilAppsFrameworkServiceServerRepository.TenantRealm{RealmID: realmID},
			wantStatus:   fiber.StatusOK,
			wantRealmID:  realmID,
		},
		{
			name:           "realm lookup failure falls back to nil realm",
			realmService:   true,
			realmLookupErr: realmLookupErr,
			wantStatus:     fiber.StatusOK,
			wantRealmID:    googleUuid.Nil,
		},
		{
			name:             "session issue error",
			sessionIssueErr:  sessionIssueErr,
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorMessage: clientAuthErrorSessionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, tenant := newActiveClientAuthEntities(t)
			sessionManager := newMockSessionManagerSuccess()

			var issuedRealmID googleUuid.UUID

			sessionManager.issueServiceFunc = func(
				_ context.Context,
				_ string,
				_ googleUuid.UUID,
				realmID googleUuid.UUID,
			) (string, error) {
				issuedRealmID = realmID

				if tc.sessionIssueErr != nil {
					return "", tc.sessionIssueErr
				}

				return testServiceClientToken, nil
			}

			handlers := &clientAuthenticationHandlers{
				clientRepository: &clientAuthClientRepositoryStub{
					client: client,
					err:    tc.clientLookupErr,
				},
				tenantRepository: &clientAuthTenantRepositoryStub{
					tenant: tenant,
					err:    tc.tenantLookupErr,
				},
				sessionManager: sessionManager,
				verifySecretFn: func(string, string) (bool, error) {
					if tc.secretVerifyErr != nil {
						return false, tc.secretVerifyErr
					}

					return true, nil
				},
			}
			if tc.realmService {
				handlers.realmService = &clientAuthRealmServiceStub{
					realm: tc.realm,
					err:   tc.realmLookupErr,
				}
			}

			response := performClientTokenRequest(t, handlers)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			if tc.wantErrorMessage != "" {
				var body map[string]string
				require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
				require.Equal(t, tc.wantErrorMessage, body[cryptoutilSharedMagic.StringError])

				return
			}

			require.Equal(t, tc.wantRealmID, issuedRealmID)
		})
	}
}

func TestParseBasicClientCredentials_ErrorCases(t *testing.T) {
	t.Parallel()

	encodedValidCredentials := base64.StdEncoding.EncodeToString([]byte("client-id:client-secret"))
	tests := []struct {
		name       string
		header     string
		wantClient string
		wantSecret string
		wantOK     bool
	}{
		{name: "missing scheme separator"},
		{name: "unsupported scheme", header: cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix + encodedValidCredentials},
		{name: "empty credentials", header: "Basic   "},
		{name: "invalid base64", header: "Basic !"},
		{name: "missing colon", header: "Basic " + base64.StdEncoding.EncodeToString([]byte("client-secret"))},
		{name: "empty client ID", header: "Basic " + base64.StdEncoding.EncodeToString([]byte(":client-secret"))},
		{name: "empty client secret", header: "Basic " + base64.StdEncoding.EncodeToString([]byte("client-id:"))},
		{
			name:       "valid case-insensitive scheme",
			header:     "bAsIc " + encodedValidCredentials,
			wantClient: "client-id",
			wantSecret: "client-secret",
			wantOK:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clientID, clientSecret, ok := parseBasicClientCredentials(tc.header)

			require.Equal(t, tc.wantOK, ok)
			require.Equal(t, tc.wantClient, clientID)
			require.Equal(t, tc.wantSecret, clientSecret)
		})
	}
}

func TestClientAuthJSON_EncodingError(t *testing.T) {
	t.Parallel()

	var responseErr error

	app := newClientAuthTestApp()
	app.Get("/", func(c *fiber.Ctx) error {
		responseErr = clientAuthJSON(c, fiber.StatusOK, make(chan int))

		return responseErr
	})

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
	require.ErrorContains(t, responseErr, "write client authentication response")
}

func TestIsNotFoundAppError_NonNotFoundErrors(t *testing.T) {
	t.Parallel()

	badRequestSummary := "invalid request"
	tests := []struct {
		name string
		err  error
	}{
		{name: "nil error"},
		{name: "plain error", err: errors.New("plain repository failure")},
		{
			name: "HTTP bad request",
			err: cryptoutilSharedApperr.NewHTTP400BadRequest(
				&badRequestSummary,
				errors.New("invalid request"),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.False(t, isNotFoundAppError(tc.err))
		})
	}
}

func TestGenerateServiceClientCredentials_ErrorPaths(t *testing.T) {
	t.Parallel()

	clientIDErr := errors.New("injected UUID generation error")
	secretErr := errors.New("injected random source error")
	generatedUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	tests := []struct {
		name          string
		generateUUID  func() (googleUuid.UUID, error)
		readRandom    func([]byte) (int, error)
		wantErrorText string
	}{
		{
			name: "UUID generation error",
			generateUUID: func() (googleUuid.UUID, error) {
				return googleUuid.Nil, clientIDErr
			},
			readRandom: func([]byte) (int, error) {
				return 0, nil
			},
			wantErrorText: "generate client ID",
		},
		{
			name: "random source error",
			generateUUID: func() (googleUuid.UUID, error) {
				return generatedUUID, nil
			},
			readRandom: func([]byte) (int, error) {
				return 0, secretErr
			},
			wantErrorText: "generate client secret",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clientID, clientSecret, err := generateServiceClientCredentialsWith(
				tc.generateUUID,
				tc.readRandom,
			)

			require.ErrorContains(t, err, tc.wantErrorText)
			require.Empty(t, clientID)
			require.Empty(t, clientSecret)
		})
	}
}

func newActiveClientAuthEntities(
	t *testing.T,
) (*cryptoutilAppsFrameworkServiceServerRepository.Client, *cryptoutilAppsFrameworkServiceServerRepository.Tenant) {
	t.Helper()

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)
	clientID, err := googleUuid.NewV7()
	require.NoError(t, err)

	tenant := &cryptoutilAppsFrameworkServiceServerRepository.Tenant{
		ID:     tenantID,
		Name:   "token-test-" + tenantID.String(),
		Active: 1,
	}
	client := &cryptoutilAppsFrameworkServiceServerRepository.Client{
		ID:               clientID,
		TenantID:         tenantID,
		ClientID:         clientID.String(),
		ClientSecretHash: "test-secret-hash",
		Active:           1,
	}

	return client, tenant
}

func performClientTokenRequest(t *testing.T, handlers *clientAuthenticationHandlers) *http.Response {
	t.Helper()

	app := newClientAuthTestApp()
	app.Post(cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath, handlers.handleIssueServiceToken)

	request := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
		nil,
	)
	request.SetBasicAuth(testServiceClientToken, testServiceClientSecret)

	response, err := app.Test(request, -1)
	require.NoError(t, err)

	return response
}
