package clientauth

import (
	"context"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func mustNewUUID() googleUuid.UUID {
	id, _ := googleUuid.NewV7()
	return id
}

func TestPBKDF2Hasher_HashSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "valid password",
			plaintext: "mySecretPassword123",
			wantErr:   false,
		},
		{
			name:      "empty password",
			plaintext: "",
			wantErr:   false,
		},
		{
			name:      "unicode password",
			plaintext: "パスワード🔐",
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hasher := NewPBKDF2Hasher()
			hashed, err := hasher.HashSecret(tc.plaintext)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, hashed)
			require.True(t, strings.HasPrefix(hashed, "$pbkdf2-sha256$"), "Hash should have PBKDF2 format")
			require.NotEqual(t, tc.plaintext, hashed, "Hash should differ from plaintext")
		})
	}
}

func TestPBKDF2Hasher_CompareSecret(t *testing.T) {
	t.Parallel()

	hasher := NewPBKDF2Hasher()
	plaintext := "mySecretPassword123"
	hashed, err := hasher.HashSecret(plaintext)
	require.NoError(t, err)

	tests := []struct {
		name      string
		hashed    string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "matching password",
			hashed:    hashed,
			plaintext: plaintext,
			wantErr:   false,
		},
		{
			name:      "wrong password",
			hashed:    hashed,
			plaintext: "wrongPassword",
			wantErr:   true,
		},
		{
			name:      "invalid hash format",
			hashed:    "invalid-hash",
			plaintext: plaintext,
			wantErr:   true,
		},
		{
			name:      "empty plaintext",
			hashed:    hashed,
			plaintext: "",
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := hasher.CompareSecret(tc.hashed, tc.plaintext)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestMigrateClientSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name            string
		clients         []*cryptoutilIdentityDomain.Client
		expectedMigrate int
		wantErr         bool
	}{
		{
			name: "migrate plaintext secrets",
			clients: []*cryptoutilIdentityDomain.Client{
				{
					ID:           mustNewUUID(),
					ClientID:     "plaintext-client",
					ClientSecret: "plaintext-secret",
					Enabled:      true,
				},
			},
			expectedMigrate: 1,
			wantErr:         false,
		},
		{
			name: "skip public clients",
			clients: []*cryptoutilIdentityDomain.Client{
				{
					ID:           mustNewUUID(),
					ClientID:     "public-client",
					ClientSecret: "",
					Enabled:      true,
				},
			},
			expectedMigrate: 0,
			wantErr:         false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockClientRepository{clients: make([]*cryptoutilIdentityDomain.Client, len(tc.clients))}
			copy(mockRepo.clients, tc.clients)
			hasher := NewPBKDF2Hasher()

			migrated, err := MigrateClientSecrets(ctx, mockRepo, hasher)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedMigrate, migrated)
		})
	}
}

func TestSecretBasedAuthenticator_AuthenticateBasic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hasher := NewPBKDF2Hasher()
	hashedSecret, err := hasher.HashSecret("correct-secret")
	require.NoError(t, err)

	clientID, _ := googleUuid.NewV7()
	enabledClient := &cryptoutilIdentityDomain.Client{
		ID:           clientID,
		ClientID:     "enabled-client",
		ClientSecret: hashedSecret,
		Enabled:      true,
	}

	disabledClientID, _ := googleUuid.NewV7()
	disabledClient := &cryptoutilIdentityDomain.Client{
		ID:           disabledClientID,
		ClientID:     "disabled-client",
		ClientSecret: hashedSecret,
		Enabled:      false,
	}

	tests := []struct {
		name         string
		client       *cryptoutilIdentityDomain.Client
		clientID     string
		clientSecret string
		wantErr      bool
	}{
		{
			name:         "valid credentials",
			client:       enabledClient,
			clientID:     "enabled-client",
			clientSecret: "correct-secret",
			wantErr:      false,
		},
		{
			name:         "wrong secret",
			client:       enabledClient,
			clientID:     "enabled-client",
			clientSecret: "wrong-secret",
			wantErr:      true,
		},
		{
			name:         "disabled client",
			client:       disabledClient,
			clientID:     "disabled-client",
			clientSecret: "correct-secret",
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockClientRepository{clients: []*cryptoutilIdentityDomain.Client{tc.client}}
			authenticator := NewSecretBasedAuthenticator(mockRepo)

			client, err := authenticator.AuthenticateBasic(ctx, tc.clientID, tc.clientSecret)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, client)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)
			require.Equal(t, tc.clientID, client.ClientID)
		})
	}
}

// mockClientRepository implements cryptoutilIdentityRepository.ClientRepository for testing.
type mockClientRepository struct {
	clients []*cryptoutilIdentityDomain.Client
}

func (m *mockClientRepository) Create(ctx context.Context, client *cryptoutilIdentityDomain.Client) error {
	m.clients = append(m.clients, client)
	return nil
}

func (m *mockClientRepository) GetByClientID(ctx context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error) {
	for _, c := range m.clients {
		if c.ClientID == clientID {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockClientRepository) GetAll(ctx context.Context) ([]*cryptoutilIdentityDomain.Client, error) {
	return m.clients, nil
}

func (m *mockClientRepository) Update(ctx context.Context, client *cryptoutilIdentityDomain.Client) error {
	for i, c := range m.clients {
		if c.ID == client.ID {
			m.clients[i] = client
			return nil
		}
	}
	return nil
}

func (m *mockClientRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	for i, c := range m.clients {
		if c.ID == id {
			m.clients = append(m.clients[:i], m.clients[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockClientRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Client, error) {
	for _, c := range m.clients {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockClientRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.clients)), nil
}

func (m *mockClientRepository) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Client, error) {
	if offset >= len(m.clients) {
		return []*cryptoutilIdentityDomain.Client{}, nil
	}
	end := offset + limit
	if end > len(m.clients) {
		end = len(m.clients)
	}
	return m.clients[offset:end], nil
}
