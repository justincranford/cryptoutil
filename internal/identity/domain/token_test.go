package domain_test

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"cryptoutil/internal/identity/domain"
)

func TestToken_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		token          *domain.Token
		expectIDChange bool
	}{
		{
			name: "generates ID when empty",
			token: &domain.Token{
				TokenValue: "test_token",
			},
			expectIDChange: true,
		},
		{
			name: "preserves existing ID",
			token: &domain.Token{
				ID:         googleUuid.Must(googleUuid.NewV7()),
				TokenValue: "test_token",
			},
			expectIDChange: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			originalID := tc.token.ID

			err := tc.token.BeforeCreate(nil)
			require.NoError(t, err)

			if tc.expectIDChange {
				require.NotEqual(t, googleUuid.Nil, tc.token.ID, "ID should be generated")
			} else {
				require.Equal(t, originalID, tc.token.ID, "ID should be preserved")
			}
		})
	}
}

func TestToken_TableName(t *testing.T) {
	t.Parallel()

	token := domain.Token{}
	require.Equal(t, "tokens", token.TableName())
}

func TestToken_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		expired   bool
	}{
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			expired:   true,
		},
		{
			name:      "valid token",
			expiresAt: time.Now().Add(1 * time.Hour),
			expired:   false,
		},
		{
			name:      "token expiring now",
			expiresAt: time.Now().Add(1 * time.Second),
			expired:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token := &domain.Token{
				ExpiresAt: tc.expiresAt,
			}

			require.Equal(t, tc.expired, token.IsExpired())
		})
	}
}

func TestToken_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		revoked   bool
		valid     bool
	}{
		{
			name:      "valid active token",
			expiresAt: time.Now().Add(1 * time.Hour),
			revoked:   false,
			valid:     true,
		},
		{
			name:      "expired active token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			revoked:   false,
			valid:     false,
		},
		{
			name:      "valid revoked token",
			expiresAt: time.Now().Add(1 * time.Hour),
			revoked:   true,
			valid:     false,
		},
		{
			name:      "expired revoked token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			revoked:   true,
			valid:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token := &domain.Token{
				ExpiresAt: tc.expiresAt,
				Revoked:   tc.revoked,
			}

			require.Equal(t, tc.valid, token.IsValid())
		})
	}
}
