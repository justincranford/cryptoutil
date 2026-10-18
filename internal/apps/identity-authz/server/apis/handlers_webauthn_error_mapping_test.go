// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	json "encoding/json"
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

const webAuthnErrorMappingCredentialID = "webauthn-error-mapping-credential"

func TestFinishWebAuthnHandlers_MapSessionErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ceremony   cryptoutilIdentityMfa.WebAuthnCeremonyType
		expired    bool
		wantStatus int
	}{
		{name: "registration session not found", ceremony: cryptoutilIdentityMfa.WebAuthnCeremonyRegistration, wantStatus: fiber.StatusNotFound},
		{name: "registration session expired", ceremony: cryptoutilIdentityMfa.WebAuthnCeremonyRegistration, expired: true, wantStatus: fiber.StatusGone},
		{name: "authentication session not found", ceremony: cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication, wantStatus: fiber.StatusNotFound},
		{name: "authentication session expired", ceremony: cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication, expired: true, wantStatus: fiber.StatusGone},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userID := googleUuid.Must(googleUuid.NewV7())

			sessionID := googleUuid.Must(googleUuid.NewV7())
			if tc.expired {
				sessionID = createExpiredWebAuthnSession(t, userID, tc.ceremony)
			}

			config := *totpTestConfig
			tokens := *totpTestConfig.Tokens
			tokens.Issuer = webAuthnTestIssuer
			config.Tokens = &tokens
			service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			requestBody := webAuthnFinishRequestBody(t, tc.ceremony, userID, sessionID)

			requestPath := webAuthnRegistrationFinishPath
			if tc.ceremony == cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication {
				requestPath = webAuthnAuthenticationFinishPath
			}

			request := httptest.NewRequest(http.MethodPost, requestPath, strings.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, tc.wantStatus, response.StatusCode)
		})
	}
}

func TestFinishWebAuthnHandlers_RejectVerificationFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ceremony cryptoutilIdentityMfa.WebAuthnCeremonyType
		issuer   string
		path     string
	}{
		{
			name:     testRegistrationCase,
			ceremony: cryptoutilIdentityMfa.WebAuthnCeremonyRegistration,
			issuer:   webAuthnExampleOrgIssuer,
			path:     webAuthnRegistrationFinishPath,
		},
		{
			name:     testAuthenticationCase,
			ceremony: cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication,
			issuer:   webAuthnTestIssuer,
			path:     webAuthnAuthenticationFinishPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userID := googleUuid.Must(googleUuid.NewV7())
			credential := cryptoutilIdentityMfa.WebAuthnCredential{
				ID:              googleUuid.Must(googleUuid.NewV7()),
				UserID:          userID,
				CredentialID:    []byte(webAuthnErrorMappingCredentialID),
				PublicKey:       []byte("public-key"),
				AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
				Transports:      testEmptyJSONArray,
				DisplayName:     webAuthnErrorMappingCredentialDisplayName,
			}

			webAuthnConfig := cryptoutilIdentityMfa.WebAuthnConfig{
				RPDisplayName: webAuthnRPDisplayName,
				RPID:          webAuthnRPID,
				RPOrigins:     []string{tc.issuer},
			}
			if tc.issuer == webAuthnTestIssuer {
				webAuthnConfig.RPID = cryptoutilSharedMagic.HostnameLocalhost

				require.NoError(t, totpTestRepoFactory.DB().Create(&credential).Error)
			}

			webAuthnService, err := cryptoutilIdentityMfa.NewWebAuthnService(totpTestRepoFactory.DB(), webAuthnConfig)
			require.NoError(t, err)

			user := &cryptoutilIdentityMfa.WebAuthnUser{
				ID:          userID,
				Name:        webAuthnFinishUsername,
				DisplayName: webAuthnFinishDisplayName,
			}

			var sessionID googleUuid.UUID

			switch tc.ceremony {
			case cryptoutilIdentityMfa.WebAuthnCeremonyRegistration:
				_, sessionID, err = webAuthnService.BeginRegistration(t.Context(), user)
			case cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication:
				user.Credentials = []cryptoutilIdentityMfa.WebAuthnCredential{credential}
				_, sessionID, err = webAuthnService.BeginAuthentication(t.Context(), user)
			default:
				require.FailNowf(t, "unexpected WebAuthn ceremony type", "got %q", tc.ceremony)
			}

			require.NoError(t, err)

			config := *totpTestConfig
			tokens := *totpTestConfig.Tokens
			tokens.Issuer = tc.issuer
			config.Tokens = &tokens
			service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			requestBody := webAuthnRegistrationFinishRequestBody(t, userID.String(), sessionID.String())
			if tc.ceremony == cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication {
				requestBody = webAuthnAuthenticationFinishRequestBody(t, userID, sessionID)
			}

			request := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
		})
	}
}

func webAuthnFinishRequestBody(t *testing.T, ceremony cryptoutilIdentityMfa.WebAuthnCeremonyType, userID, sessionID googleUuid.UUID) string {
	t.Helper()

	if ceremony == cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication {
		return webAuthnAuthenticationFinishRequestBody(t, userID, sessionID)
	}

	return webAuthnRegistrationFinishRequestBody(t, userID.String(), sessionID.String())
}

func webAuthnRegistrationFinishRequestBody(t *testing.T, userID, sessionID string) string {
	t.Helper()

	attestationObject, err := hex.DecodeString(webAuthnRegistrationAttestationFixture)
	require.NoError(t, err)
	clientDataJSON, err := hex.DecodeString(webAuthnRegistrationClientDataFixture)
	require.NoError(t, err)
	credentialID, err := hex.DecodeString(webAuthnRegistrationCredentialIDFixture)
	require.NoError(t, err)

	encodedCredentialID := base64.RawURLEncoding.EncodeToString(credentialID)
	responseBody, err := json.Marshal(map[string]any{
		webAuthnCredentialIDJSONField:    encodedCredentialID,
		webAuthnRawCredentialIDJSONField: encodedCredentialID,
		webAuthnCredentialTypeJSONField:  webAuthnRegistrationCredentialType,
		webAuthnCredentialResponseJSONField: map[string]string{
			testWebAuthnAttestationObjectField: base64.RawURLEncoding.EncodeToString(attestationObject),
			webAuthnClientDataJSONField:        base64.RawURLEncoding.EncodeToString(clientDataJSON),
		},
	})
	require.NoError(t, err)

	request, err := json.Marshal(cryptoutilIdentityAuthz.FinishWebAuthnRegistrationRequest{
		UserID:    userID,
		Username:  webAuthnFinishUsername,
		SessionID: sessionID,
		Response:  base64.StdEncoding.EncodeToString(responseBody),
	})
	require.NoError(t, err)

	return string(request)
}

func webAuthnAuthenticationFinishRequestBody(t *testing.T, userID, sessionID googleUuid.UUID) string {
	t.Helper()

	credentialID := []byte(webAuthnErrorMappingCredentialID)
	authenticatorData := make([]byte, 37)
	clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"test-challenge","origin":"http://localhost:8080"}`)
	responseBody, err := json.Marshal(map[string]any{
		webAuthnCredentialIDJSONField:    base64.RawURLEncoding.EncodeToString(credentialID),
		webAuthnRawCredentialIDJSONField: base64.RawURLEncoding.EncodeToString(credentialID),
		webAuthnCredentialTypeJSONField:  webAuthnAuthenticationCredentialType,
		webAuthnCredentialResponseJSONField: map[string]string{
			testWebAuthnAuthenticatorDataField: base64.RawURLEncoding.EncodeToString(authenticatorData),
			webAuthnClientDataJSONField:        base64.RawURLEncoding.EncodeToString(clientDataJSON),
			testWebAuthnSignatureField:         base64.RawURLEncoding.EncodeToString([]byte{1}),
			testWebAuthnUserHandleField:        base64.RawURLEncoding.EncodeToString(userID[:]),
		},
	})
	require.NoError(t, err)

	request, err := json.Marshal(cryptoutilIdentityAuthz.FinishWebAuthnAuthenticationRequest{
		UserID:    userID.String(),
		Username:  webAuthnFinishUsername,
		SessionID: sessionID.String(),
		Response:  base64.StdEncoding.EncodeToString(responseBody),
	})
	require.NoError(t, err)

	return string(request)
}

func createExpiredWebAuthnSession(t *testing.T, userID googleUuid.UUID, ceremony cryptoutilIdentityMfa.WebAuthnCeremonyType) googleUuid.UUID {
	t.Helper()

	webAuthnService, err := cryptoutilIdentityMfa.NewWebAuthnService(totpTestRepoFactory.DB(), cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: webAuthnRPDisplayName,
		RPID:          cryptoutilSharedMagic.HostnameLocalhost,
		RPOrigins:     []string{webAuthnTestIssuer},
	})
	require.NoError(t, err)

	user := &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          userID,
		Name:        webAuthnFinishUsername,
		DisplayName: webAuthnFinishDisplayName,
	}

	var sessionID googleUuid.UUID

	switch ceremony {
	case cryptoutilIdentityMfa.WebAuthnCeremonyRegistration:
		_, sessionID, err = webAuthnService.BeginRegistration(context.Background(), user)
	case cryptoutilIdentityMfa.WebAuthnCeremonyAuthentication:
		credential := cryptoutilIdentityMfa.WebAuthnCredential{
			ID:              googleUuid.Must(googleUuid.NewV7()),
			UserID:          userID,
			CredentialID:    []byte(webAuthnErrorMappingCredentialID),
			PublicKey:       []byte("public-key"),
			AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
			Transports:      testEmptyJSONArray,
			DisplayName:     webAuthnErrorMappingCredentialDisplayName,
		}
		require.NoError(t, totpTestRepoFactory.DB().Create(&credential).Error)
		user.Credentials = []cryptoutilIdentityMfa.WebAuthnCredential{credential}
		_, sessionID, err = webAuthnService.BeginAuthentication(context.Background(), user)
	default:
		require.FailNowf(t, "unexpected WebAuthn ceremony type", "got %q", ceremony)
	}

	require.NoError(t, err)
	require.NoError(t, totpTestRepoFactory.DB().Model(&cryptoutilIdentityMfa.WebAuthnSession{}).
		Where("id = ?", sessionID).
		Update("expires_at", time.Now().UTC().Add(-time.Minute)).Error)

	return sessionID
}
