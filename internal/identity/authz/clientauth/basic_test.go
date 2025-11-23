// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"encoding/base64"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// MockClientRepo is a mock implementation of ClientRepository for testing.
type mockClientRepo struct {
	clients map[string]*cryptoutilIdentityDomain.Client
}

func (m *mockClientRepo) GetByClientID(_ context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error) {
	client, ok := m.clients[clientID]
	if !ok {
		return nil, cryptoutilIdentityAppErr.ErrClientNotFound
	}

	return client, nil
}

func (m *mockClientRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) GetAll(_ context.Context) ([]*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) Create(_ context.Context, _ *cryptoutilIdentityDomain.Client) error {
	return nil
}

func (m *mockClientRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.Client) error {
	return nil
}

func (m *mockClientRepo) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

func (m *mockClientRepo) List(_ context.Context, _ int, _ int) ([]*cryptoutilIdentityDomain.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func TestBasicAuthenticator_MethodName(t *testing.T) {
	t.Parallel()

	repo := &mockClientRepo{}
	auth := NewBasicAuthenticator(repo)
	require.Equal(t, cryptoutilIdentityMagic.ClientAuthMethodSecretBasic, auth.Method())
}

func TestBasicAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	testClientID := "test-client-id"
	testClientSecret := "test-client-secret"
	testClientIDUUID := googleUuid.New()

	repo := &mockClientRepo{
		clients: map[string]*cryptoutilIdentityDomain.Client{
			testClientID: {
				ID:                      testClientIDUUID,
				ClientID:                testClientID,
				ClientSecret:            testClientSecret,
				ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
				AllowedGrantTypes:       []string{"authorization_code"},
				AllowedResponseTypes:    []string{"code"},
				AllowedScopes:           []string{"openid"},
				RedirectURIs:            []string{"https://example.com/callback"},
				RequirePKCE:             true,
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
			credential:  base64.StdEncoding.EncodeToString([]byte(testClientID + ":" + testClientSecret)),
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name:        "invalid base64 encoding",
			clientID:    testClientID,
			credential:  "not-base64!!!",
			wantErr:     true,
			wantErrType: nil, // fmt.Errorf wrapper
		},
		{
			name:        "missing colon separator",
			clientID:    testClientID,
			credential:  base64.StdEncoding.EncodeToString([]byte("no-colon")),
			wantErr:     true,
			wantErrType: cryptoutilIdentityAppErr.ErrInvalidClientAuth,
		},
		{
			name:        "client ID mismatch",
			clientID:    testClientID,
			credential:  base64.StdEncoding.EncodeToString([]byte("wrong-id:" + testClientSecret)),
			wantErr:     true,
			wantErrType: cryptoutilIdentityAppErr.ErrInvalidClientAuth,
		},
		{
			name:        "invalid client secret",
			clientID:    testClientID,
			credential:  base64.StdEncoding.EncodeToString([]byte(testClientID + ":wrong-secret")),
			wantErr:     true,
			wantErrType: cryptoutilIdentityAppErr.ErrInvalidClientSecret,
		},
		{
			name:        "client not found",
			clientID:    "nonexistent",
			credential:  base64.StdEncoding.EncodeToString([]byte("nonexistent:secret")),
			wantErr:     true,
			wantErrType: nil, // fmt.Errorf wrapper
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
