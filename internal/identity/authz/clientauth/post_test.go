// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

func TestPostAuthenticator_MethodName(t *testing.T) {
	t.Parallel()

	repo := &mockClientRepo{}
	auth := NewPostAuthenticator(repo)
	require.Equal(t, cryptoutilIdentityMagic.ClientAuthMethodSecretPost, auth.Method())
}

func TestPostAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	testClientID := "test-client-id"
	testClientSecret := "test-client-secret"
	testClientIDUUID := googleUuid.New()

	// Hash the client secret for storage using PBKDF2 (format: pbkdf2$iter$salt$hash).
	hashedSecret, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2(testClientSecret)
	require.NoError(t, err)

	repo := &mockClientRepo{
		clients: map[string]*cryptoutilIdentityDomain.Client{
			testClientID: {
				ID:                      testClientIDUUID,
				ClientID:                testClientID,
				ClientSecret:            hashedSecret,
				ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				AllowedGrantTypes:       []string{"authorization_code"},
				AllowedResponseTypes:    []string{"code"},
				AllowedScopes:           []string{"openid"},
				RedirectURIs:            []string{"https://example.com/callback"},
				RequirePKCE:             boolPtr(true),
				AccessTokenLifetime:     3600,
				RefreshTokenLifetime:    86400,
				IDTokenLifetime:         3600,
			},
		},
	}

	auth := NewPostAuthenticator(repo)

	tests := []struct {
		name        string
		clientID    string
		credential  string
		wantErr     bool
		wantErrType error
	}{
		{
			name:        "valid post auth",
			clientID:    testClientID,
			credential:  testClientSecret,
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name:        "invalid client secret",
			clientID:    testClientID,
			credential:  "wrong-secret",
			wantErr:     true,
			wantErrType: cryptoutilIdentityAppErr.ErrInvalidClientSecret,
		},
		{
			name:        "client not found",
			clientID:    "nonexistent",
			credential:  "some-secret",
			wantErr:     true,
			wantErrType: cryptoutilIdentityAppErr.ErrClientNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			client, err := auth.Authenticate(ctx, tc.clientID, tc.credential)

			if tc.wantErr {
				require.Error(t, err)

				if tc.wantErrType != nil {
					require.ErrorIs(t, err, tc.wantErrType)
				}

				require.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				require.Equal(t, testClientID, client.ClientID)
			}
		})
	}
}

func TestPostAuthenticator_ValidateAuthMethod(t *testing.T) {
	t.Parallel()

	repo := &mockClientRepo{}
	auth := NewPostAuthenticator(repo)

	tests := []struct {
		name     string
		client   *cryptoutilIdentityDomain.Client
		expected bool
	}{
		{
			name: "valid auth method",
			client: &cryptoutilIdentityDomain.Client{
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			},
			expected: true,
		},
		{
			name: "invalid auth method - BASIC",
			client: &cryptoutilIdentityDomain.Client{
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
			},
			expected: false,
		},
		{
			name: "invalid auth method - private_key_jwt",
			client: &cryptoutilIdentityDomain.Client{
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodPrivateKeyJWT,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := auth.validateAuthMethod(tc.client)
			require.Equal(t, tc.expected, result)
		})
	}
}
