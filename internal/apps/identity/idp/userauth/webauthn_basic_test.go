// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// TestWebAuthnUser_WebAuthnID tests WebAuthnID.
func TestWebAuthnUser_WebAuthnID(t *testing.T) {
	t.Parallel()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	domainUser := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser",
		Name:              "Test User",
	}

	user := &WebAuthnUser{
		user:        domainUser,
		credentials: nil,
	}

	webAuthnID := user.WebAuthnID()
	require.NotNil(t, webAuthnID, "WebAuthnID should not be nil")
	require.Equal(t, []byte(userID.String()), webAuthnID, "WebAuthnID should return user ID bytes")
}

// TestWebAuthnUser_WebAuthnName tests WebAuthnName.
func TestWebAuthnUser_WebAuthnName(t *testing.T) {
	t.Parallel()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	domainUser := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser@example.com",
		Name:              "Test User",
	}

	user := &WebAuthnUser{
		user:        domainUser,
		credentials: nil,
	}

	name := user.WebAuthnName()
	require.Equal(t, "testuser@example.com", name, "WebAuthnName should return preferred username")
}

// TestWebAuthnUser_WebAuthnDisplayName tests WebAuthnDisplayName.
func TestWebAuthnUser_WebAuthnDisplayName(t *testing.T) {
	t.Parallel()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	domainUser := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser",
		Name:              "Test Display Name",
	}

	user := &WebAuthnUser{
		user:        domainUser,
		credentials: nil,
	}

	displayName := user.WebAuthnDisplayName()
	require.Equal(t, "Test Display Name", displayName, "WebAuthnDisplayName should return name")
}

// TestWebAuthnUser_WebAuthnDisplayNameFallback tests WebAuthnDisplayName fallback to PreferredUsername.
func TestWebAuthnUser_WebAuthnDisplayNameFallback(t *testing.T) {
	t.Parallel()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	domainUser := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser",
		Name:              "", // Empty name should fall back to PreferredUsername.
	}

	user := &WebAuthnUser{
		user:        domainUser,
		credentials: nil,
	}

	displayName := user.WebAuthnDisplayName()
	require.Equal(t, "testuser", displayName, "WebAuthnDisplayName should fall back to preferred username when name is empty")
}

// TestWebAuthnUser_WebAuthnIcon tests WebAuthnIcon.
func TestWebAuthnUser_WebAuthnIcon(t *testing.T) {
	t.Parallel()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	domainUser := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser",
	}

	user := &WebAuthnUser{
		user:        domainUser,
		credentials: nil,
	}

	icon := user.WebAuthnIcon()
	require.Equal(t, "", icon, "WebAuthnIcon should return empty string")
}

// TestWebAuthnUser_WebAuthnCredentials tests WebAuthnCredentials with nil credentials.
func TestWebAuthnUser_WebAuthnCredentials_Nil(t *testing.T) {
	t.Parallel()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	domainUser := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser",
	}

	user := &WebAuthnUser{
		user:        domainUser,
		credentials: nil,
	}

	creds := user.WebAuthnCredentials()
	require.Nil(t, creds, "WebAuthnCredentials should return nil for nil credentials")
}

// TestWebAuthnAuthenticator_NewAuthenticatorNilConfig tests NewWebAuthnAuthenticator with nil config.
func TestWebAuthnAuthenticator_NewAuthenticatorNilConfig(t *testing.T) {
	t.Parallel()

	auth, err := NewWebAuthnAuthenticator(nil, nil, nil)
	require.Error(t, err, "NewWebAuthnAuthenticator should fail with nil config")
	require.Contains(t, err.Error(), "config cannot be nil", "Error should indicate config is nil")
	require.Nil(t, auth, "Authenticator should be nil on error")
}

// TestWebAuthnAuthenticator_NewAuthenticatorNilCredentialStore tests NewWebAuthnAuthenticator with nil credential store.
func TestWebAuthnAuthenticator_NewAuthenticatorNilCredentialStore(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	auth, err := NewWebAuthnAuthenticator(config, nil, nil)
	require.Error(t, err, "NewWebAuthnAuthenticator should fail with nil credential store")
	require.Contains(t, err.Error(), "credential store cannot be nil", "Error should indicate credential store is nil")
	require.Nil(t, auth, "Authenticator should be nil on error")
}

// TestWebAuthnAuthenticator_NewAuthenticatorNilChallengeStore tests NewWebAuthnAuthenticator with nil challenge store.
func TestWebAuthnAuthenticator_NewAuthenticatorNilChallengeStore(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	// Create a mock credential store.
	credStore := &mockWebAuthnCredentialStore{}

	auth, err := NewWebAuthnAuthenticator(config, credStore, nil)
	require.Error(t, err, "NewWebAuthnAuthenticator should fail with nil challenge store")
	require.Contains(t, err.Error(), "challenge store cannot be nil", "Error should indicate challenge store is nil")
	require.Nil(t, auth, "Authenticator should be nil on error")
}

// TestWebAuthnAuthenticator_Method tests Method.
func TestWebAuthnAuthenticator_Method(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	credStore := &mockWebAuthnCredentialStore{}
	challengeStore := NewInMemoryChallengeStore()

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err, "NewWebAuthnAuthenticator should succeed")

	method := auth.Method()
	require.Equal(t, "passkey_webauthn", method, "Method should return 'passkey_webauthn'")
}

// mockWebAuthnCredentialStore implements CredentialStore for testing.
type mockWebAuthnCredentialStore struct {
	credentials map[string][]*Credential
}

func (m *mockWebAuthnCredentialStore) StoreCredential(_ context.Context, cred *Credential) error {
	if m.credentials == nil {
		m.credentials = make(map[string][]*Credential)
	}

	m.credentials[cred.UserID] = append(m.credentials[cred.UserID], cred)

	return nil
}

func (m *mockWebAuthnCredentialStore) GetCredential(_ context.Context, credID string) (*Credential, error) {
	for _, creds := range m.credentials {
		for _, cred := range creds {
			if cred.ID == credID {
				return cred, nil
			}
		}
	}

	return nil, nil
}

func (m *mockWebAuthnCredentialStore) GetUserCredentials(_ context.Context, userID string) ([]*Credential, error) {
	if m.credentials == nil {
		return nil, nil
	}

	return m.credentials[userID], nil
}

func (m *mockWebAuthnCredentialStore) DeleteCredential(_ context.Context, _ string) error {
	return nil
}

// TestWebAuthnAuthenticator_BeginRegistration_NilUser tests BeginRegistration with nil user.
func TestWebAuthnAuthenticator_BeginRegistration_NilUser(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	credStore := &mockWebAuthnCredentialStore{}
	challengeStore := NewInMemoryChallengeStore()

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err, "NewWebAuthnAuthenticator should succeed")

	creation, err := auth.BeginRegistration(context.Background(), nil)
	require.Error(t, err, "BeginRegistration should fail with nil user")
	require.Contains(t, err.Error(), "user cannot be nil", "Error should indicate user is nil")
	require.Nil(t, creation, "Creation should be nil on error")
}

// TestWebAuthnAuthenticator_BeginRegistration_Success tests successful BeginRegistration.
func TestWebAuthnAuthenticator_BeginRegistration_Success(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	credStore := &mockWebAuthnCredentialStore{}
	challengeStore := NewInMemoryChallengeStore()

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err, "NewWebAuthnAuthenticator should succeed")

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	user := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser",
		Name:              "Test User",
	}

	creation, err := auth.BeginRegistration(context.Background(), user)
	require.NoError(t, err, "BeginRegistration should succeed")
	require.NotNil(t, creation, "Creation should not be nil")
	require.NotNil(t, creation.Response, "Creation response should not be nil")
}

// TestWebAuthnAuthenticator_FinishRegistration_NilUser tests FinishRegistration with nil user.
func TestWebAuthnAuthenticator_FinishRegistration_NilUser(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	credStore := &mockWebAuthnCredentialStore{}
	challengeStore := NewInMemoryChallengeStore()

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err, "NewWebAuthnAuthenticator should succeed")

	cred, err := auth.FinishRegistration(context.Background(), nil, "challenge-id", nil)
	require.Error(t, err, "FinishRegistration should fail with nil user")
	require.Contains(t, err.Error(), "user cannot be nil", "Error should indicate user is nil")
	require.Nil(t, cred, "Credential should be nil on error")
}

// TestWebAuthnAuthenticator_FinishRegistration_NilResponse tests FinishRegistration with nil response.
func TestWebAuthnAuthenticator_FinishRegistration_NilResponse(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	credStore := &mockWebAuthnCredentialStore{}
	challengeStore := NewInMemoryChallengeStore()

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err, "NewWebAuthnAuthenticator should succeed")

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	user := &cryptoutilIdentityDomain.User{
		ID:                userID,
		PreferredUsername: "testuser",
	}

	cred, err := auth.FinishRegistration(context.Background(), user, "challenge-id", nil)
	require.Error(t, err, "FinishRegistration should fail with nil response")
	require.Contains(t, err.Error(), "credential creation response cannot be nil", "Error should indicate response is nil")
	require.Nil(t, cred, "Credential should be nil on error")
}

// TestWebAuthnAuthenticator_InitiateAuth_NoCredentials tests InitiateAuth with no registered credentials.
func TestWebAuthnAuthenticator_InitiateAuth_NoCredentials(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	credStore := &mockWebAuthnCredentialStore{}
	challengeStore := NewInMemoryChallengeStore()

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err, "NewWebAuthnAuthenticator should succeed")

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	challenge, err := auth.InitiateAuth(context.Background(), userID.String())
	require.Error(t, err, "InitiateAuth should fail with no credentials")
	require.Contains(t, err.Error(), "no WebAuthn credentials registered for user", "Error should indicate no credentials")
	require.Nil(t, challenge, "Challenge should be nil on error")
}

// TestWebAuthnAuthenticator_VerifyAuth_NilResponse tests VerifyAuth with nil response.
func TestWebAuthnAuthenticator_VerifyAuth_NilResponse(t *testing.T) {
	t.Parallel()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
	}

	credStore := &mockWebAuthnCredentialStore{}
	challengeStore := NewInMemoryChallengeStore()

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err, "NewWebAuthnAuthenticator should succeed")

	user, err := auth.VerifyAuth(context.Background(), "challenge-id", nil)
	require.Error(t, err, "VerifyAuth should fail with nil response")
	require.Contains(t, err.Error(), "credential assertion response cannot be nil", "Error should indicate response is nil")
	require.Nil(t, user, "User should be nil on error")
}
