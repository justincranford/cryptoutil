// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestTokenRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	token := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_create",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid", "profile"},
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	require.Equal(t, token.TokenValue, retrieved.TokenValue)
}

func TestTokenRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr error
	}{
		{
			name:    "token not found",
			id:      googleUuid.Must(googleUuid.NewV7()),
			wantErr: cryptoutilIdentityAppErr.ErrTokenNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token, err := repo.GetByID(ctx, tc.id)
			require.ErrorIs(t, err, tc.wantErr)
			require.Nil(t, token)
		})
	}
}

func TestTokenRepository_GetByTokenValue(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	testToken := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_getbyvalue",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid", "profile"},
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}

	err := repo.Create(ctx, testToken)
	require.NoError(t, err)

	tests := []struct {
		name       string
		tokenValue string
		wantErr    error
	}{
		{
			name:       "token found",
			tokenValue: "access_token_getbyvalue",
			wantErr:    nil,
		},
		{
			name:       "token not found",
			tokenValue: "nonexistent",
			wantErr:    cryptoutilIdentityAppErr.ErrTokenNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token, err := repo.GetByTokenValue(ctx, tc.tokenValue)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, token)
			} else {
				require.NoError(t, err)
				require.NotNil(t, token)
				require.Equal(t, tc.tokenValue, token.TokenValue)
			}
		})
	}
}

func TestTokenRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	token := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_update",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid", "profile"},
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	token.Scopes = []string{"openid", "profile", "email"}
	err = repo.Update(ctx, token)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	require.Len(t, retrieved.Scopes, 3)
}

func TestTokenRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	token := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_delete",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid", "profile"},
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	err = repo.Delete(ctx, token.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, token.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrTokenNotFound)
	require.Nil(t, retrieved)
}

func TestTokenRepository_RevokeByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	token := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_revoke",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid", "profile"},
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	err = repo.RevokeByID(ctx, token.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	require.True(t, retrieved.Revoked.Bool())
}

func TestTokenRepository_RevokeByTokenValue(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	token := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_revoke_value",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid", "profile"},
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	err = repo.RevokeByTokenValue(ctx, "access_token_revoke_value")
	require.NoError(t, err)

	retrieved, err := repo.GetByTokenValue(ctx, "access_token_revoke_value")
	require.NoError(t, err)
	require.True(t, retrieved.Revoked.Bool())
}

func TestTokenRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	for i := range 5 {
		token := &cryptoutilIdentityDomain.Token{
			ID:             googleUuid.Must(googleUuid.NewV7()),
			TokenValue:     "access_token_list_" + string(rune('0'+i)),
			TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
			TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
			ClientID:       googleUuid.Must(googleUuid.NewV7()),
			UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
			Scopes:         []string{"openid", "profile"},
			IssuedAt:       time.Now().UTC(),
			ExpiresAt:      time.Now().UTC().Add(time.Hour),
			Revoked:        false,
			RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
		}
		err := repo.Create(ctx, token)
		require.NoError(t, err)
	}

	tokens, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	require.Len(t, tokens, 3)

	tokens, err = repo.List(ctx, 3, 3)
	require.NoError(t, err)
	require.Len(t, tokens, 2)
}

func TestTokenRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	for i := range 5 {
		token := &cryptoutilIdentityDomain.Token{
			ID:             googleUuid.Must(googleUuid.NewV7()),
			TokenValue:     "access_token_count_" + string(rune('0'+i)),
			TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
			TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
			ClientID:       googleUuid.Must(googleUuid.NewV7()),
			UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
			Scopes:         []string{"openid", "profile"},
			IssuedAt:       time.Now().UTC(),
			ExpiresAt:      time.Now().UTC().Add(time.Hour),
			Revoked:        false,
			RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
		}
		err := repo.Create(ctx, token)
		require.NoError(t, err)
	}

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(5), count)
}

func TestTokenRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	// Create expired token.
	expiredToken := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_expired",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid"},
		IssuedAt:       time.Now().UTC().Add(-2 * time.Hour),
		ExpiresAt:      time.Now().UTC().Add(-1 * time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}
	err := repo.Create(ctx, expiredToken)
	require.NoError(t, err)

	// Create valid token.
	validToken := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_valid",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid"},
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(1 * time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}
	err = repo.Create(ctx, validToken)
	require.NoError(t, err)

	// Delete expired tokens.
	err = repo.DeleteExpired(ctx)
	require.NoError(t, err)

	// Expired token should be gone.
	_, err = repo.GetByID(ctx, expiredToken.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrTokenNotFound)

	// Valid token should still exist.
	_, err = repo.GetByID(ctx, validToken.ID)
	require.NoError(t, err)
}

func TestTokenRepository_DeleteExpiredBefore(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewTokenRepository(testDB.db)
	ctx := context.Background()

	// Create token expired 2 hours ago.
	token1 := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_expired_2h",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid"},
		IssuedAt:       time.Now().UTC().Add(-3 * time.Hour),
		ExpiresAt:      time.Now().UTC().Add(-2 * time.Hour),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}
	err := repo.Create(ctx, token1)
	require.NoError(t, err)

	// Create token expired 30 minutes ago.
	token2 := &cryptoutilIdentityDomain.Token{
		ID:             googleUuid.Must(googleUuid.NewV7()),
		TokenValue:     "access_token_expired_30m",
		TokenType:      cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:    cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:       googleUuid.Must(googleUuid.NewV7()),
		UserID:         cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		Scopes:         []string{"openid"},
		IssuedAt:       time.Now().UTC().Add(-1 * time.Hour),
		ExpiresAt:      time.Now().UTC().Add(-30 * time.Minute),
		Revoked:        false,
		RefreshTokenID: cryptoutilIdentityDomain.NullableUUID{Valid: false},
	}
	err = repo.Create(ctx, token2)
	require.NoError(t, err)

	// Delete tokens expired before 1 hour ago.
	cutoffTime := time.Now().UTC().Add(-1 * time.Hour)
	deletedCount, err := repo.DeleteExpiredBefore(ctx, cutoffTime)
	require.NoError(t, err)
	require.Equal(t, 1, deletedCount) // Only token1 deleted.

	// token1 should be gone.
	_, err = repo.GetByID(ctx, token1.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrTokenNotFound)

	// token2 should still exist.
	_, err = repo.GetByID(ctx, token2.ID)
	require.NoError(t, err)
}
