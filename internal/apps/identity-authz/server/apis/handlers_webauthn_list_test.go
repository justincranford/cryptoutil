// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestListWebAuthnCredentials_WithAndWithoutLastUsedAt(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7())
	notUsedCredentialID := googleUuid.Must(googleUuid.NewV7())
	notUsedCredential := &cryptoutilIdentityMfa.WebAuthnCredential{
		ID:              googleUuid.Must(googleUuid.NewV7()),
		UserID:          userID,
		CredentialID:    notUsedCredentialID[:],
		PublicKey:       []byte{1},
		AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
		DisplayName:     "Not used",
	}
	lastUsed := time.Now().UTC().Add(-time.Hour)
	usedCredentialID := googleUuid.Must(googleUuid.NewV7())
	usedCredential := &cryptoutilIdentityMfa.WebAuthnCredential{
		ID:              googleUuid.Must(googleUuid.NewV7()),
		UserID:          userID,
		CredentialID:    usedCredentialID[:],
		PublicKey:       []byte{2},
		AttestationType: cryptoutilSharedMagic.WebAuthnAttestationTypeNone,
		DisplayName:     "Previously used",
		LastUsedAt:      &lastUsed,
	}

	require.NoError(t, totpTestRepoFactory.DB().Create(notUsedCredential).Error)
	require.NoError(t, totpTestRepoFactory.DB().Create(usedCredential).Error)

	config := *totpTestConfig
	tokens := *totpTestConfig.Tokens
	tokens.Issuer = webAuthnTestIssuer
	config.Tokens = &tokens

	service := cryptoutilIdentityAuthz.NewService(&config, totpTestRepoFactory, nil)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	service.RegisterRoutes(app)

	request := httptest.NewRequest(http.MethodGet, webAuthnCredentialsPath+"/"+userID.String(), nil)
	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()

	require.Equal(t, fiber.StatusOK, response.StatusCode)

	var result cryptoutilIdentityAuthz.ListWebAuthnCredentialsResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.Len(t, result.Credentials, 2)

	credentialsByID := make(map[string]cryptoutilIdentityAuthz.WebAuthnCredentialResponse, len(result.Credentials))
	for _, credential := range result.Credentials {
		credentialsByID[credential.ID] = credential
	}

	notUsedResponse, ok := credentialsByID[notUsedCredential.ID.String()]
	require.True(t, ok)
	require.Equal(t, notUsedCredential.DisplayName, notUsedResponse.DisplayName)
	require.NotEmpty(t, notUsedResponse.CreatedAt)
	require.Empty(t, notUsedResponse.LastUsedAt)

	usedResponse, ok := credentialsByID[usedCredential.ID.String()]
	require.True(t, ok)
	require.Equal(t, usedCredential.DisplayName, usedResponse.DisplayName)
	require.Equal(t, lastUsed.Format(time.RFC3339), usedResponse.LastUsedAt)
}
