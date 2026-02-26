// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // Use modernc CGO-free SQLite.

)

func TestDBRealmRepository_UpdatePassword(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	// Create test user.
	user := &DBRealmUser{
		ID:       "pwd-user-1",
		RealmID:  "realm-1",
		Username: "pwduser",
		Enabled:  true,
	}
	err = repo.CreateUser(ctx, user, "oldpassword")
	require.NoError(t, err)

	// Update password.
	err = repo.UpdatePassword(ctx, "pwd-user-1", "newpassword")
	require.NoError(t, err)

	// Verify old password fails.
	result, err := repo.Authenticate(ctx, "realm-1", "pwduser", "oldpassword")
	require.NoError(t, err)
	require.False(t, result.Authenticated)

	// Verify new password works.
	result, err = repo.Authenticate(ctx, "realm-1", "pwduser", "newpassword")
	require.NoError(t, err)
	require.True(t, result.Authenticated)
}

func TestDBRealmRepository_EnableDisableUser(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	// Create test user.
	user := &DBRealmUser{
		ID:       "toggle-user-1",
		RealmID:  "realm-1",
		Username: "toggleuser",
		Enabled:  true,
	}
	err = repo.CreateUser(ctx, user, "password")
	require.NoError(t, err)

	// Disable user.
	err = repo.DisableUser(ctx, "toggle-user-1")
	require.NoError(t, err)

	found, err := repo.GetUser(ctx, "toggle-user-1")
	require.NoError(t, err)
	require.False(t, found.Enabled)

	// Enable user.
	err = repo.EnableUser(ctx, "toggle-user-1")
	require.NoError(t, err)

	found, err = repo.GetUser(ctx, "toggle-user-1")
	require.NoError(t, err)
	require.True(t, found.Enabled)
}

func TestDBRealmRepository_DeleteUser(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	// Create test user.
	user := &DBRealmUser{
		ID:       "delete-user-1",
		RealmID:  "realm-1",
		Username: "deleteuser",
		Enabled:  true,
	}
	err = repo.CreateUser(ctx, user, "password")
	require.NoError(t, err)

	// Delete user.
	err = repo.DeleteUser(ctx, "delete-user-1")
	require.NoError(t, err)

	// Verify user is deleted.
	found, err := repo.GetUser(ctx, "delete-user-1")
	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, found)
}

func TestDBRealmRepository_ListUsers(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	// Create test users.
	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		user := &DBRealmUser{
			ID:       "list-user-" + string(rune('0'+i)),
			RealmID:  "realm-1",
			Username: "listuser" + string(rune('0'+i)),
			Enabled:  true,
		}
		err = repo.CreateUser(ctx, user, "password")
		require.NoError(t, err)
	}

	// List all users.
	users, err := repo.ListUsers(ctx, "realm-1", 0, 0)
	require.NoError(t, err)
	require.Len(t, users, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

	// List with limit.
	users, err = repo.ListUsers(ctx, "realm-1", 2, 0)
	require.NoError(t, err)
	require.Len(t, users, 2)

	// List with offset.
	users, err = repo.ListUsers(ctx, "realm-1", 2, 2)
	require.NoError(t, err)
	require.Len(t, users, 2)
}

func TestDBRealmRepository_CountUsers(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	// Initially zero.
	count, err := repo.CountUsers(ctx, "count-realm")
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Create users.
	for i := 0; i < 3; i++ {
		user := &DBRealmUser{
			ID:       "count-user-" + string(rune('0'+i)),
			RealmID:  "count-realm",
			Username: "countuser" + string(rune('0'+i)),
			Enabled:  true,
		}
		err = repo.CreateUser(ctx, user, "password")
		require.NoError(t, err)
	}

	// Count again.
	count, err = repo.CountUsers(ctx, "count-realm")
	require.NoError(t, err)
	require.Equal(t, int64(3), count)
}

func TestDBRealmRepository_UpdateUser(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	realmID := fmt.Sprintf("realm-%d", time.Now().UTC().UnixNano())
	userID := fmt.Sprintf("user-%d", time.Now().UTC().UnixNano())

	user := &DBRealmUser{
		ID:       userID,
		RealmID:  realmID,
		Username: "updateuser",
		Email:    "old@example.com",
		Roles:    "[]",
		Enabled:  true,
	}

	err = repo.CreateUser(ctx, user, "password")
	require.NoError(t, err)

	user.Email = "new@example.com"
	user.Roles = `["admin","user"]`

	err = repo.UpdateUser(ctx, user)
	require.NoError(t, err)

	updated, err := repo.GetUser(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "new@example.com", updated.Email)
	require.Equal(t, `["admin","user"]`, updated.Roles)
}
