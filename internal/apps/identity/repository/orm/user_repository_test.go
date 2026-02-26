// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               "user-create-sub",
		PreferredUsername: "createuser",
		Email:             "create@example.com",
		EmailVerified:     true,
		PasswordHash:      "hash123",
		Enabled:           true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.Sub, retrieved.Sub)
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr error
	}{
		{
			name:    "user not found",
			id:      googleUuid.Must(googleUuid.NewV7()),
			wantErr: cryptoutilIdentityAppErr.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user, err := repo.GetByID(ctx, tc.id)
			require.ErrorIs(t, err, tc.wantErr)
			require.Nil(t, user)
		})
	}
}

func TestUserRepository_GetBySub(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               "user-sub-test",
		PreferredUsername: "subuser",
		Email:             "sub@example.com",
		EmailVerified:     true,
		PasswordHash:      "hash123",
		Enabled:           true,
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	tests := []struct {
		name    string
		sub     string
		wantErr error
	}{
		{
			name:    "user found",
			sub:     "user-sub-test",
			wantErr: nil,
		},
		{
			name:    "user not found",
			sub:     "nonexistent",
			wantErr: cryptoutilIdentityAppErr.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user, err := repo.GetBySub(ctx, tc.sub)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, tc.sub, user.Sub)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               "user-username-test",
		PreferredUsername: "testusername",
		Email:             "username@example.com",
		EmailVerified:     true,
		PasswordHash:      "hash123",
		Enabled:           true,
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "user found",
			username: "testusername",
			wantErr:  nil,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			wantErr:  cryptoutilIdentityAppErr.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user, err := repo.GetByUsername(ctx, tc.username)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, tc.username, user.PreferredUsername)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               "user-email-test",
		PreferredUsername: "emailuser",
		Email:             "test@example.com",
		EmailVerified:     true,
		PasswordHash:      "hash123",
		Enabled:           true,
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "user found",
			email:   "test@example.com",
			wantErr: nil,
		},
		{
			name:    "user not found",
			email:   "nonexistent@example.com",
			wantErr: cryptoutilIdentityAppErr.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user, err := repo.GetByEmail(ctx, tc.email)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, tc.email, user.Email)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               "user-update-test",
		PreferredUsername: "updateuser",
		Email:             "update@example.com",
		EmailVerified:     false,
		PasswordHash:      "hash123",
		Enabled:           true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	user.EmailVerified = true
	err = repo.Update(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.True(t, retrieved.EmailVerified.Bool())
}

func TestUserRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               "user-delete-test",
		PreferredUsername: "deleteuser",
		Email:             "delete@example.com",
		EmailVerified:     true,
		PasswordHash:      "hash123",
		Enabled:           true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, user.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrUserNotFound)
	require.Nil(t, retrieved)
}

func TestUserRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	for i := range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		user := &cryptoutilIdentityDomain.User{
			ID:                googleUuid.Must(googleUuid.NewV7()),
			Sub:               "user-list-" + string(rune('0'+i)),
			PreferredUsername: "listuser" + string(rune('0'+i)),
			Email:             "list" + string(rune('0'+i)) + "@example.com",
			EmailVerified:     true,
			PasswordHash:      "hash123",
			Enabled:           true,
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	users, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	require.Len(t, users, 3)

	users, err = repo.List(ctx, 3, 3)
	require.NoError(t, err)
	require.Len(t, users, 2)
}

func TestUserRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewUserRepository(testDB.db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	for i := range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
		user := &cryptoutilIdentityDomain.User{
			ID:                googleUuid.Must(googleUuid.NewV7()),
			Sub:               "user-count-" + string(rune('0'+i)),
			PreferredUsername: "countuser" + string(rune('0'+i)),
			Email:             "count" + string(rune('0'+i)) + "@example.com",
			EmailVerified:     true,
			PasswordHash:      "hash123",
			Enabled:           true,
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), count)
}
