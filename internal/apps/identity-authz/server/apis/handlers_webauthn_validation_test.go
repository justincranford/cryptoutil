// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"encoding/base64"
	json "encoding/json"
	"io"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	webAuthnTestIssuer                     = "http://localhost:8080"
	webAuthnRegistrationBeginPath          = "/oidc/v1/mfa/webauthn/register/begin"
	webAuthnRegistrationFinishPath         = "/oidc/v1/mfa/webauthn/register/finish"
	webAuthnAuthenticationBeginPath        = "/oidc/v1/mfa/webauthn/authenticate/begin"
	webAuthnAuthenticationFinishPath       = "/oidc/v1/mfa/webauthn/authenticate/finish"
	webAuthnCredentialsPath                = "/oidc/v1/mfa/webauthn/credentials"
	webAuthnEmptyObject                    = `{}`
	webAuthnMalformedJSON                  = "{"
	webAuthnInvalidUserRegistrationRequest = `{"user_id":"invalid","username":"passkey-user"}`
)

func TestWebAuthnHandlers_ValidationAndDatabasePaths(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	credentialID := googleUuid.Must(googleUuid.NewV7()).String()
	sessionID := googleUuid.Must(googleUuid.NewV7()).String()
	encodedInvalidCredentialResponse := base64.StdEncoding.EncodeToString([]byte(webAuthnEmptyObject))
	config := *totpTestConfig
	tokens := *totpTestConfig.Tokens
	tokens.Issuer = webAuthnTestIssuer
	config.Tokens = &tokens
	service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "registration malformed body", method: http.MethodPost, path: webAuthnRegistrationBeginPath, body: webAuthnMalformedJSON, wantStatus: fiber.StatusBadRequest},
		{name: "registration missing user", method: http.MethodPost, path: webAuthnRegistrationBeginPath, body: webAuthnEmptyObject, wantStatus: fiber.StatusBadRequest},
		{name: "registration missing username", method: http.MethodPost, path: webAuthnRegistrationBeginPath, body: `{"user_id":"` + userID + `"}`, wantStatus: fiber.StatusBadRequest},
		{name: "registration invalid user ID", method: http.MethodPost, path: webAuthnRegistrationBeginPath, body: webAuthnInvalidUserRegistrationRequest, wantStatus: fiber.StatusBadRequest},
		{name: "registration begins", method: http.MethodPost, path: webAuthnRegistrationBeginPath, body: `{"user_id":"` + userID + `","username":"passkey-user"}`, wantStatus: fiber.StatusOK},
		{name: "registration finish missing fields", method: http.MethodPost, path: webAuthnRegistrationFinishPath, body: webAuthnEmptyObject, wantStatus: fiber.StatusBadRequest},
		{name: "registration finish invalid user ID", method: http.MethodPost, path: webAuthnRegistrationFinishPath, body: `{"user_id":"invalid","username":"passkey-user","session_id":"` + sessionID + `","response":"e30="}`, wantStatus: fiber.StatusBadRequest},
		{name: "registration finish invalid session ID", method: http.MethodPost, path: webAuthnRegistrationFinishPath, body: `{"user_id":"` + userID + `","username":"passkey-user","session_id":"invalid","response":"e30="}`, wantStatus: fiber.StatusBadRequest},
		{name: "registration finish invalid base64", method: http.MethodPost, path: webAuthnRegistrationFinishPath, body: `{"user_id":"` + userID + `","username":"passkey-user","session_id":"` + sessionID + `","response":"!"}`, wantStatus: fiber.StatusBadRequest},
		{name: "registration finish invalid credential response", method: http.MethodPost, path: webAuthnRegistrationFinishPath, body: `{"user_id":"` + userID + `","username":"passkey-user","session_id":"` + sessionID + `","response":"` + encodedInvalidCredentialResponse + `"}`, wantStatus: fiber.StatusBadRequest},
		{name: "authentication malformed body", method: http.MethodPost, path: webAuthnAuthenticationBeginPath, body: webAuthnMalformedJSON, wantStatus: fiber.StatusBadRequest},
		{name: "authentication missing user", method: http.MethodPost, path: webAuthnAuthenticationBeginPath, body: webAuthnEmptyObject, wantStatus: fiber.StatusBadRequest},
		{name: "authentication missing username", method: http.MethodPost, path: webAuthnAuthenticationBeginPath, body: `{"user_id":"` + userID + `"}`, wantStatus: fiber.StatusBadRequest},
		{name: "authentication invalid user ID", method: http.MethodPost, path: webAuthnAuthenticationBeginPath, body: webAuthnInvalidUserRegistrationRequest, wantStatus: fiber.StatusBadRequest},
		{name: "authentication without credentials", method: http.MethodPost, path: webAuthnAuthenticationBeginPath, body: `{"user_id":"` + userID + `","username":"passkey-user"}`, wantStatus: fiber.StatusNotFound},
		{name: "authentication finish missing fields", method: http.MethodPost, path: webAuthnAuthenticationFinishPath, body: webAuthnEmptyObject, wantStatus: fiber.StatusBadRequest},
		{name: "authentication finish invalid user ID", method: http.MethodPost, path: webAuthnAuthenticationFinishPath, body: `{"user_id":"invalid","username":"passkey-user","session_id":"` + sessionID + `","response":"e30="}`, wantStatus: fiber.StatusBadRequest},
		{name: "authentication finish invalid session ID", method: http.MethodPost, path: webAuthnAuthenticationFinishPath, body: `{"user_id":"` + userID + `","username":"passkey-user","session_id":"invalid","response":"e30="}`, wantStatus: fiber.StatusBadRequest},
		{name: "authentication finish invalid base64", method: http.MethodPost, path: webAuthnAuthenticationFinishPath, body: `{"user_id":"` + userID + `","username":"passkey-user","session_id":"` + sessionID + `","response":"!"}`, wantStatus: fiber.StatusBadRequest},
		{name: "authentication finish invalid assertion", method: http.MethodPost, path: webAuthnAuthenticationFinishPath, body: `{"user_id":"` + userID + `","username":"passkey-user","session_id":"` + sessionID + `","response":"` + encodedInvalidCredentialResponse + `"}`, wantStatus: fiber.StatusBadRequest},
		{name: "credential list invalid user ID", method: http.MethodGet, path: webAuthnCredentialsPath + "/invalid", wantStatus: fiber.StatusBadRequest},
		{name: "credential list empty", method: http.MethodGet, path: webAuthnCredentialsPath + "/" + userID, wantStatus: fiber.StatusOK},
		{name: "credential deletion malformed body", method: http.MethodDelete, path: webAuthnCredentialsPath, body: webAuthnMalformedJSON, wantStatus: fiber.StatusBadRequest},
		{name: "credential deletion missing fields", method: http.MethodDelete, path: webAuthnCredentialsPath, body: webAuthnEmptyObject, wantStatus: fiber.StatusBadRequest},
		{name: "credential deletion invalid user ID", method: http.MethodDelete, path: webAuthnCredentialsPath, body: `{"user_id":"invalid","credential_id":"` + credentialID + `"}`, wantStatus: fiber.StatusBadRequest},
		{name: "credential deletion invalid credential ID", method: http.MethodDelete, path: webAuthnCredentialsPath, body: `{"user_id":"` + userID + `","credential_id":"invalid"}`, wantStatus: fiber.StatusBadRequest},
		{name: "credential deletion missing credential", method: http.MethodDelete, path: webAuthnCredentialsPath, body: `{"user_id":"` + userID + `","credential_id":"` + credentialID + `"}`, wantStatus: fiber.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			service.RegisterRoutes(app)

			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			responseBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, response.StatusCode, string(responseBody))

			if tc.wantStatus == fiber.StatusOK {
				var result map[string]any
				require.NoError(t, json.Unmarshal(responseBody, &result))
				require.NotEmpty(t, result)
			}
		})
	}
}

func TestWebAuthnHandlers_ListCredentials(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7())
	lastUsedAt := time.Now().UTC()
	credentials := []cryptoutilIdentityMfa.WebAuthnCredential{
		{
			ID:              googleUuid.Must(googleUuid.NewV7()),
			UserID:          userID,
			CredentialID:    []byte("unused-credential-id"),
			PublicKey:       []byte("public-key"),
			AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
			DisplayName:     "unused credential",
		},
		{
			ID:              googleUuid.Must(googleUuid.NewV7()),
			UserID:          userID,
			CredentialID:    []byte("used-credential-id"),
			PublicKey:       []byte("public-key"),
			AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
			DisplayName:     "used credential",
			LastUsedAt:      &lastUsedAt,
		},
	}
	require.NoError(t, totpTestRepoFactory.DB().Create(&credentials).Error)

	config := *totpTestConfig
	tokens := *totpTestConfig.Tokens
	tokens.Issuer = webAuthnTestIssuer
	config.Tokens = &tokens
	service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)

	app := fiber.New()
	service.RegisterRoutes(app)

	request := httptest.NewRequest(http.MethodGet, webAuthnCredentialsPath+"/"+userID.String(), nil)
	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	var result cryptoutilIdentityAuthz.ListWebAuthnCredentialsResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Len(t, result.Credentials, len(credentials))

	credentialsByID := make(map[string]cryptoutilIdentityAuthz.WebAuthnCredentialResponse, len(result.Credentials))
	for _, credential := range result.Credentials {
		credentialsByID[credential.ID] = credential
	}

	require.Empty(t, credentialsByID[credentials[0].ID.String()].LastUsedAt)
	require.Equal(t, lastUsedAt.Format(time.RFC3339), credentialsByID[credentials[1].ID.String()].LastUsedAt)
	require.Equal(t, credentials[1].DisplayName, credentialsByID[credentials[1].ID.String()].DisplayName)
}

func TestWebAuthnHandlers_DeleteCredential(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7())
	credential := cryptoutilIdentityMfa.WebAuthnCredential{
		ID:              googleUuid.Must(googleUuid.NewV7()),
		UserID:          userID,
		CredentialID:    []byte("credential-to-delete"),
		PublicKey:       []byte("public-key"),
		AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
		DisplayName:     "credential to delete",
	}
	require.NoError(t, totpTestRepoFactory.DB().Create(&credential).Error)

	config := *totpTestConfig
	tokens := *totpTestConfig.Tokens
	tokens.Issuer = webAuthnTestIssuer
	config.Tokens = &tokens
	service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)

	app := fiber.New()
	service.RegisterRoutes(app)

	body := `{"user_id":"` + userID.String() + `","credential_id":"` + credential.ID.String() + `"}`
	request := httptest.NewRequest(http.MethodDelete, webAuthnCredentialsPath, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, response.StatusCode)

	var result map[string]bool
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.True(t, result["success"])

	var count int64
	require.NoError(t, totpTestRepoFactory.DB().Model(&cryptoutilIdentityMfa.WebAuthnCredential{}).Where("id = ?", credential.ID).Count(&count).Error)
	require.Zero(t, count)
}

func TestWebAuthnHandlers_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	userUUID := googleUuid.Must(googleUuid.NewV7())
	userID := userUUID.String()
	sessionUUID := googleUuid.Must(googleUuid.NewV7())
	sessionID := sessionUUID.String()
	config := *totpTestConfig
	tokens := *totpTestConfig.Tokens
	tokens.Issuer = ":"
	config.Tokens = &tokens
	service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   webAuthnRegistrationBeginCase,
			method: http.MethodPost,
			path:   webAuthnRegistrationBeginPath,
			body:   `{"user_id":"` + userID + `","username":"passkey-user"}`,
		},
		{
			name:   webAuthnAuthenticationBeginCase,
			method: http.MethodPost,
			path:   webAuthnAuthenticationBeginPath,
			body:   `{"user_id":"` + userID + `","username":"passkey-user"}`,
		},
		{name: webAuthnCredentialListCase, method: http.MethodGet, path: webAuthnCredentialsPath + "/" + userID},
		{
			name:   "credential deletion",
			method: http.MethodDelete,
			path:   webAuthnCredentialsPath,
			body:   `{"user_id":"` + userID + `","credential_id":"` + googleUuid.Must(googleUuid.NewV7()).String() + `"}`,
		},
		{
			name:   testRegistrationFinishCase,
			method: http.MethodPost,
			path:   webAuthnRegistrationFinishPath,
			body:   webAuthnRegistrationFinishRequestBody(t, userID, sessionID),
		},
		{
			name:   testAuthenticationFinishCase,
			method: http.MethodPost,
			path:   webAuthnAuthenticationFinishPath,
			body:   webAuthnAuthenticationFinishRequestBody(t, userUUID, sessionUUID),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			service.RegisterRoutes(app)

			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
		})
	}
}

func TestWebAuthnHandlers_DatabaseLoadFailure(t *testing.T) {
	t.Parallel()

	config, repoFactory, err := newTOTPIntegrationTestFixture()
	require.NoError(t, err)
	require.NoError(t, repoFactory.Close())

	tokens := *config.Tokens
	tokens.Issuer = webAuthnTestIssuer
	config.Tokens = &tokens

	userUUID := googleUuid.Must(googleUuid.NewV7())
	userID := userUUID.String()
	sessionUUID := googleUuid.Must(googleUuid.NewV7())
	sessionID := sessionUUID.String()
	credentialID := googleUuid.Must(googleUuid.NewV7()).String()
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	tests := []struct {
		name        string
		method      string
		routePath   string
		requestPath string
		body        string
		handler     func(*fiber.Ctx) error
	}{
		{
			name:        "registration begin",
			method:      http.MethodPost,
			routePath:   webAuthnRegistrationBeginPath,
			requestPath: webAuthnRegistrationBeginPath,
			body:        `{"user_id":"` + userID + `","username":"passkey-user"}`,
			handler:     service.BeginWebAuthnRegistration,
		},
		{
			name:        "authentication begin",
			method:      http.MethodPost,
			routePath:   webAuthnAuthenticationBeginPath,
			requestPath: webAuthnAuthenticationBeginPath,
			body:        `{"user_id":"` + userID + `","username":"passkey-user"}`,
			handler:     service.BeginWebAuthnAuthentication,
		},
		{
			name:        "credential list",
			method:      http.MethodGet,
			routePath:   webAuthnCredentialsPath + "/:user_id",
			requestPath: webAuthnCredentialsPath + "/" + userID,
			handler:     service.ListWebAuthnCredentials,
		},
		{
			name:        "credential delete",
			method:      http.MethodDelete,
			routePath:   webAuthnCredentialsPath,
			requestPath: webAuthnCredentialsPath,
			body:        `{"user_id":"` + userID + `","credential_id":"` + credentialID + `"}`,
			handler:     service.DeleteWebAuthnCredential,
		},
		{
			name:        testRegistrationFinishCase,
			method:      http.MethodPost,
			routePath:   webAuthnRegistrationFinishPath,
			requestPath: webAuthnRegistrationFinishPath,
			body:        webAuthnRegistrationFinishRequestBody(t, userID, sessionID),
			handler:     service.FinishWebAuthnRegistration,
		},
		{
			name:        testAuthenticationFinishCase,
			method:      http.MethodPost,
			routePath:   webAuthnAuthenticationFinishPath,
			requestPath: webAuthnAuthenticationFinishPath,
			body:        webAuthnAuthenticationFinishRequestBody(t, userUUID, sessionUUID),
			handler:     service.FinishWebAuthnAuthentication,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Add(tc.method, tc.routePath, tc.handler)

			request := httptest.NewRequest(tc.method, tc.requestPath, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")

			response, requestErr := app.Test(request, -1)

			require.NoError(t, requestErr)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
		})
	}
}

func TestWebAuthnHandlers_BeginSessionPersistenceFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		path               string
		wantError          string
		requiresCredential bool
	}{
		{
			name:      testRegistrationCase,
			path:      webAuthnRegistrationBeginPath,
			wantError: "Failed to begin registration",
		},
		{
			name:               testAuthenticationCase,
			path:               webAuthnAuthenticationBeginPath,
			wantError:          "Failed to begin authentication",
			requiresCredential: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, repoFactory, err := newTOTPIntegrationTestFixture()

			require.NoError(t, err)
			defer func() {
				require.NoError(t, repoFactory.Close())
			}()

			tokens := *config.Tokens
			tokens.Issuer = webAuthnTestIssuer
			config.Tokens = &tokens

			userID := googleUuid.Must(googleUuid.NewV7())
			if tc.requiresCredential {
				credential := cryptoutilIdentityMfa.WebAuthnCredential{
					ID:              googleUuid.Must(googleUuid.NewV7()),
					UserID:          userID,
					CredentialID:    []byte("session-persistence-credential"),
					PublicKey:       []byte("public-key"),
					AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
					Transports:      testEmptyJSONArray,
					DisplayName:     "session persistence",
				}
				require.NoError(t, repoFactory.DB().Create(&credential).Error)
			}

			require.NoError(t, repoFactory.DB().Migrator().DropTable(&cryptoutilIdentityMfa.WebAuthnSession{}))

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			body := `{"user_id":"` + userID.String() + `","username":"passkey-user"}`
			request := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			responseBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusInternalServerError, response.StatusCode, string(responseBody))

			var result map[string]string
			require.NoError(t, json.Unmarshal(responseBody, &result))
			require.Equal(t, tc.wantError, result[cryptoutilSharedMagic.StringError])
		})
	}
}
