// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

// mockClientRepo moved to test_helpers_test.go (shared across all clientauth test files)

func TestBasicAuthenticator_MethodName(t *testing.T) {
	t.Parallel()

	repo := &mockClientRepo{}
	auth := NewBasicAuthenticator(repo)
	require.Equal(t, cryptoutilIdentityMagic.ClientAuthMethodSecretBasic, auth.Method())
}

const (
	testClientSecretStr = "test-client-secret"
	testClientIDStr     = "test-client-id"
)

func TestBasicAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	testClientID := testClientIDStr
	testClientSecret := testClientSecretStr
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
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
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

	auth := NewBasicAuthenticator(repo)

	tests := []struct {
		name        string
		clientID    string
		credential  string
		wantErr     bool
		wantErrType error
	}{
		{
			name:        "valid basic auth",
			clientID:    testClientID,
			credential:  testClientSecret, // Plaintext client secret
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name:        "invalid client secret",
			clientID:    testClientID,
			credential:  "wrong-secret", // Plaintext wrong secret
			wantErr:     true,
			wantErrType: cryptoutilIdentityAppErr.ErrInvalidClientSecret,
		},
		{
			name:        "client not found",
			clientID:    "nonexistent",
			credential:  testClientSecret, // Plaintext client secret
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

func TestBasicAuthenticator_ValidateAuthMethod(t *testing.T) {
	t.Parallel()

	repo := &mockClientRepo{}
	auth := NewBasicAuthenticator(repo)

	tests := []struct {
		name     string
		client   *cryptoutilIdentityDomain.Client
		expected bool
	}{
		{
			name: "valid auth method",
			client: &cryptoutilIdentityDomain.Client{
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
			},
			expected: true,
		},
		{
			name: "invalid auth method - POST",
			client: &cryptoutilIdentityDomain.Client{
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
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
