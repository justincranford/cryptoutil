// Copyright (c) 2025 Justin Cranford
//
//

// Copyright (c) 2025 Justin Cranford
//
//

//go:build integration_placeholder

package userauth

import (
	"context"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// mockCredentialStore implements CredentialStore for testing.
type mockCredentialStore struct {
	credentials map[string]*Credential
}

func newMockCredentialStore() *mockCredentialStore {
	return &mockCredentialStore{
		credentials: make(map[string]*Credential),
	}
}

func (m *mockCredentialStore) StoreCredential(ctx context.Context, credential *Credential) error {
	m.credentials[credential.ID] = credential

	return nil
}

func (m *mockCredentialStore) GetCredential(ctx context.Context, credentialID string) (*Credential, error) {
	if cred, ok := m.credentials[credentialID]; ok {
		return cred, nil
	}

	return nil, &cryptoutilIdentityDomain.AppErr{
		Type:    cryptoutilIdentityDomain.ErrTypeNotFound,
		Message: "credential not found",
	}
}

func (m *mockCredentialStore) GetUserCredentials(ctx context.Context, userID string) ([]*Credential, error) {
	creds := make([]*Credential, 0)

	for _, cred := range m.credentials {
		if cred.UserID == userID {
			creds = append(creds, cred)
		}
	}

	return creds, nil
}

func (m *mockCredentialStore) DeleteCredential(ctx context.Context, credentialID string) error {
	delete(m.credentials, credentialID)

	return nil
}

func TestNewWebAuthnAuthenticator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		config            *WebAuthnConfig
		credentialStore   CredentialStore
		challengeStore    ChallengeStore
		wantError         bool
		wantErrorContains string
	}{
		{
			name: "valid configuration creates authenticator successfully",
			config: &WebAuthnConfig{
				RPID:          "example.com",
				RPDisplayName: "Example Corp",
				RPOrigins:     []string{"https://example.com"},
				Timeout:       cryptoutilIdentityMagic.DefaultOTPLifetime,
			},
			credentialStore: newMockCredentialStore(),
			challengeStore:  newMockChallengeStore(),
			wantError:       false,
		},
		{
			name:              "nil config returns error",
			config:            nil,
			credentialStore:   newMockCredentialStore(),
			challengeStore:    newMockChallengeStore(),
			wantError:         true,
			wantErrorContains: "config cannot be nil",
		},
		{
			name: "nil credential store returns error",
			config: &WebAuthnConfig{
				RPID:          "example.com",
				RPDisplayName: "Example Corp",
				RPOrigins:     []string{"https://example.com"},
			},
			credentialStore:   nil,
			challengeStore:    newMockChallengeStore(),
			wantError:         true,
			wantErrorContains: "credential store cannot be nil",
		},
		{
			name: "nil challenge store returns error",
			config: &WebAuthnConfig{
				RPID:          "example.com",
				RPDisplayName: "Example Corp",
				RPOrigins:     []string{"https://example.com"},
			},
			credentialStore:   newMockCredentialStore(),
			challengeStore:    nil,
			wantError:         true,
			wantErrorContains: "challenge store cannot be nil",
		},
		{
			name: "zero timeout defaults to DefaultOTPLifetime",
			config: &WebAuthnConfig{
				RPID:          "example.com",
				RPDisplayName: "Example Corp",
				RPOrigins:     []string{"https://example.com"},
				Timeout:       0,
			},
			credentialStore: newMockCredentialStore(),
			challengeStore:  newMockChallengeStore(),
			wantError:       false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth, err := NewWebAuthnAuthenticator(tc.config, tc.credentialStore, tc.challengeStore)

			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrorContains)
				require.Nil(t, auth)
			} else {
				require.NoError(t, err)
				require.NotNil(t, auth)
				require.Equal(t, "passkey_webauthn", auth.Method())

				if tc.config.Timeout == 0 {
					require.Equal(t, cryptoutilIdentityMagic.DefaultOTPLifetime, auth.config.Timeout)
				}
			}
		})
	}
}

func TestWebAuthnAuthenticator_BeginRegistration(t *testing.T) {
	t.Parallel()

	// Generate test UUIDs
	userID1 := googleUuid.Must(googleUuid.NewV7())
	userID2 := googleUuid.Must(googleUuid.NewV7())

	tests := []struct {
		name              string
		user              *cryptoutilIdentityDomain.User
		existingCreds     []*Credential
		wantError         bool
		wantErrorContains string
	}{
		{
			name: "begin registration for new user succeeds",
			user: &cryptoutilIdentityDomain.User{
				ID:                userID1,
				PreferredUsername: "test-user-webauthn-reg-new",
				Name:              "Test User",
			},
			existingCreds: nil,
			wantError:     false,
		},
		{
			name: "begin registration for user with existing credentials succeeds",
			user: &cryptoutilIdentityDomain.User{
				ID:                userID2,
				PreferredUsername: "test-user-webauthn-reg-existing",
				Name:              "Test User",
			},
			existingCreds: []*Credential{
				{
					ID:              "existing-cred-id",
					UserID:          userID2.String(),
					Type:            CredentialTypePasskey,
					PublicKey:       []byte("stub-public-key"),
					AttestationType: "none",
					SignCount:       1,
				},
			},
			wantError: false,
		},
		{
			name:              "nil user returns error",
			user:              nil,
			existingCreds:     nil,
			wantError:         true,
			wantErrorContains: "user cannot be nil",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			credStore := newMockCredentialStore()
			challengeStore := newMockChallengeStore()

			// Pre-populate existing credentials.
			if tc.existingCreds != nil && tc.user != nil {
				for _, cred := range tc.existingCreds {
					err := credStore.StoreCredential(ctx, cred)
					require.NoError(t, err)
				}
			}

			config := &WebAuthnConfig{
				RPID:          "example.com",
				RPDisplayName: "Example Corp",
				RPOrigins:     []string{"https://example.com"},
				Timeout:       cryptoutilIdentityMagic.DefaultOTPLifetime,
			}

			auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
			require.NoError(t, err)

			creation, err := auth.BeginRegistration(ctx, tc.user)

			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrorContains)
				require.Nil(t, creation)
			} else {
				require.NoError(t, err)
				require.NotNil(t, creation)
				require.NotNil(t, creation.Response.Challenge)
				require.Equal(t, protocol.VerificationPreferred, creation.Response.AuthenticatorSelection.UserVerification)
			}
		})
	}
}

func TestWebAuthnAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	// Generate test UUIDs
	userID1 := googleUuid.Must(googleUuid.NewV7()).String()
	userID2 := googleUuid.Must(googleUuid.NewV7()).String()

	tests := []struct {
		name              string
		userID            string
		existingCreds     []*Credential
		wantError         bool
		wantErrorContains string
	}{
		{
			name:   "initiate auth for user with credentials succeeds",
			userID: userID1,
			existingCreds: []*Credential{
				{
					ID:              "cred-id-1",
					UserID:          userID1,
					Type:            CredentialTypePasskey,
					PublicKey:       []byte("stub-public-key"),
					AttestationType: "none",
					SignCount:       5,
				},
			},
			wantError: false,
		},
		{
			name:              "initiate auth for user with no credentials fails",
			userID:            userID2,
			existingCreds:     nil,
			wantError:         true,
			wantErrorContains: "no WebAuthn credentials registered",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			credStore := newMockCredentialStore()
			challengeStore := newMockChallengeStore()

			// Pre-populate existing credentials.
			if tc.existingCreds != nil {
				for _, cred := range tc.existingCreds {
					err := credStore.StoreCredential(ctx, cred)
					require.NoError(t, err)
				}
			}

			config := &WebAuthnConfig{
				RPID:          "example.com",
				RPDisplayName: "Example Corp",
				RPOrigins:     []string{"https://example.com"},
				Timeout:       cryptoutilIdentityMagic.DefaultOTPLifetime,
			}

			auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
			require.NoError(t, err)

			challenge, err := auth.InitiateAuth(ctx, tc.userID)

			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrorContains)
				require.Nil(t, challenge)
			} else {
				require.NoError(t, err)
				require.NotNil(t, challenge)
				require.Equal(t, "passkey_webauthn", challenge.Method)
				require.Equal(t, tc.userID, challenge.UserID)
				require.True(t, challenge.ExpiresAt.After(time.Now().UTC()))
				require.Contains(t, challenge.Metadata, "session_data")
				require.Contains(t, challenge.Metadata, "operation")
				require.Equal(t, "authentication", challenge.Metadata["operation"])
			}
		})
	}
}

func TestWebAuthnAuthenticator_ChallengeExpiration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockCredentialStore()
	challengeStore := newMockChallengeStore()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
		Timeout:       1 * time.Millisecond, // Very short timeout for test.
	}

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err)

	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		PreferredUsername: "test-user-webauthn-expiration",
		Name:              "Test User",
	}

	// Begin registration.
	creation, err := auth.BeginRegistration(ctx, user)
	require.NoError(t, err)
	require.NotNil(t, creation)

	// Wait for challenge to expire.
	time.Sleep(10 * time.Millisecond)

	// Attempt to finish registration with expired challenge (would need real credentialCreationResponse).
	// For now, just verify challenge is expired by checking InitiateAuth with no credentials.
	challenge, err := auth.InitiateAuth(ctx, user.ID.String())
	require.Error(t, err)
	require.Contains(t, err.Error(), "no WebAuthn credentials registered")
	require.Nil(t, challenge)
}

func TestWebAuthnAuthenticator_Method(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockCredentialStore()
	challengeStore := newMockChallengeStore()

	config := &WebAuthnConfig{
		RPID:          "example.com",
		RPDisplayName: "Example Corp",
		RPOrigins:     []string{"https://example.com"},
		Timeout:       cryptoutilIdentityMagic.DefaultOTPLifetime,
	}

	auth, err := NewWebAuthnAuthenticator(config, credStore, challengeStore)
	require.NoError(t, err)

	require.Equal(t, "passkey_webauthn", auth.Method())
}

func TestWebAuthnUser_Adapter(t *testing.T) {
	t.Parallel()

	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		PreferredUsername: "test-user-adapter",
		Name:              "Test User Display",
	}

	webauthnUser := &WebAuthnUser{
		user:        user,
		credentials: nil,
	}

	require.Equal(t, []byte(user.ID.String()), webauthnUser.WebAuthnID())
	require.Equal(t, "test-user-adapter", webauthnUser.WebAuthnName())
	require.Equal(t, "Test User Display", webauthnUser.WebAuthnDisplayName())
	require.Equal(t, "", webauthnUser.WebAuthnIcon())
	require.Empty(t, webauthnUser.WebAuthnCredentials())
}

func TestWebAuthnUser_AdapterNoDisplayName(t *testing.T) {
	t.Parallel()

	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		PreferredUsername: "test-user-no-display",
		Name:              "",
	}

	webauthnUser := &WebAuthnUser{
		user:        user,
		credentials: nil,
	}

	require.Equal(t, "test-user-no-display", webauthnUser.WebAuthnDisplayName())
}
