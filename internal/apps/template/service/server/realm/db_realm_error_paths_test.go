// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"context"
"testing"

googleUuid "github.com/google/uuid"
"github.com/stretchr/testify/require"
)

// setupClosedDBRepo creates a DBRealmRepository with an already-closed DB.
// Used to test DB error paths without needing actual DB failures.
func setupClosedDBRepo(t *testing.T) (*DBRealmRepository, context.Context) {
t.Helper()

db := setupTestDB(t)
policy := DefaultPasswordPolicy()

repo, err := NewDBRealmRepository(db, &policy)
require.NoError(t, err)

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

return repo, context.Background()
}

// setupMigratedClosedDBRepo creates a DBRealmRepository, migrates the schema,
// then closes the DB to trigger DB error paths.
func setupMigratedClosedDBRepo(t *testing.T) (*DBRealmRepository, context.Context) {
t.Helper()

db := setupTestDB(t)
policy := DefaultPasswordPolicy()

repo, err := NewDBRealmRepository(db, &policy)
require.NoError(t, err)

ctx := context.Background()
err = repo.Migrate(ctx)
require.NoError(t, err)

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

return repo, ctx
}

// TestDBRealmRepository_Migrate_DBClosed verifies Migrate returns an error when
// the underlying DB is closed.
func TestDBRealmRepository_Migrate_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupClosedDBRepo(t)

err := repo.Migrate(ctx)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to migrate")
}

// TestDBRealmRepository_CreateUser_DBClosed verifies CreateUser returns a DB
// error when the underlying connection is closed.
func TestDBRealmRepository_CreateUser_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

user := &DBRealmUser{
ID:       googleUuid.Must(googleUuid.NewV7()).String(),
RealmID:  googleUuid.Must(googleUuid.NewV7()).String(),
Username: "testuser",
Enabled:  true,
}

err := repo.CreateUser(ctx, user, "testpassword")
require.Error(t, err)
require.Contains(t, err.Error(), "failed to create user")
}

// TestDBRealmRepository_UpdateUser_NilUser verifies UpdateUser returns an error
// for a nil user without DB interaction.
func TestDBRealmRepository_UpdateUser_NilUser(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
policy := DefaultPasswordPolicy()

repo, err := NewDBRealmRepository(db, &policy)
require.NoError(t, err)

ctx := context.Background()

err = repo.UpdateUser(ctx, nil)
require.Error(t, err)
require.Contains(t, err.Error(), "user cannot be nil")
}

// TestDBRealmRepository_UpdateUser_DBClosed verifies UpdateUser returns a DB
// error when the underlying connection is closed.
func TestDBRealmRepository_UpdateUser_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

user := &DBRealmUser{
ID:       googleUuid.Must(googleUuid.NewV7()).String(),
RealmID:  googleUuid.Must(googleUuid.NewV7()).String(),
Username: "user-for-update",
Enabled:  true,
}

err := repo.UpdateUser(ctx, user)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to update user")
}

// TestDBRealmRepository_GetUser_DBClosed verifies GetUser returns a non-record-
// not-found error when the underlying DB is closed.
func TestDBRealmRepository_GetUser_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

_, err := repo.GetUser(ctx, googleUuid.Must(googleUuid.NewV7()).String())
require.Error(t, err)
require.NotContains(t, err.Error(), "not found")
require.Contains(t, err.Error(), "failed to get user")
}

// TestDBRealmRepository_GetUserByUsername_DBClosed verifies GetUserByUsername
// returns a non-record-not-found error when the underlying DB is closed.
func TestDBRealmRepository_GetUserByUsername_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

_, err := repo.GetUserByUsername(ctx, googleUuid.Must(googleUuid.NewV7()).String(), "someuser")
require.Error(t, err)
require.Contains(t, err.Error(), "failed to get user by username")
}

// TestDBRealmRepository_UpdatePassword_DBClosed verifies UpdatePassword returns
// a DB error when the underlying connection is closed.
func TestDBRealmRepository_UpdatePassword_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

err := repo.UpdatePassword(ctx, googleUuid.Must(googleUuid.NewV7()).String(), "newpassword")
require.Error(t, err)
require.Contains(t, err.Error(), "failed to update password")
}

// TestDBRealmRepository_DeleteUser_DBClosed verifies DeleteUser returns a DB
// error when the underlying connection is closed.
func TestDBRealmRepository_DeleteUser_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

err := repo.DeleteUser(ctx, googleUuid.Must(googleUuid.NewV7()).String())
require.Error(t, err)
require.Contains(t, err.Error(), "failed to delete user")
}

// TestDBRealmRepository_ListUsers_DBClosed verifies ListUsers returns a DB
// error when the underlying connection is closed.
func TestDBRealmRepository_ListUsers_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

_, err := repo.ListUsers(ctx, googleUuid.Must(googleUuid.NewV7()).String(), cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 0)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to list users")
}

// TestDBRealmRepository_CountUsers_DBClosed verifies CountUsers returns a DB
// error when the underlying connection is closed.
func TestDBRealmRepository_CountUsers_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

_, err := repo.CountUsers(ctx, googleUuid.Must(googleUuid.NewV7()).String())
require.Error(t, err)
require.Contains(t, err.Error(), "failed to count users")
}

// TestDBRealmRepository_EnableUser_DBClosed verifies EnableUser returns a DB
// error when the underlying connection is closed.
func TestDBRealmRepository_EnableUser_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

err := repo.EnableUser(ctx, googleUuid.Must(googleUuid.NewV7()).String())
require.Error(t, err)
require.Contains(t, err.Error(), "failed to enable user")
}

// TestDBRealmRepository_DisableUser_DBClosed verifies DisableUser returns a DB
// error when the underlying connection is closed.
func TestDBRealmRepository_DisableUser_DBClosed(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

err := repo.DisableUser(ctx, googleUuid.Must(googleUuid.NewV7()).String())
require.Error(t, err)
require.Contains(t, err.Error(), "failed to disable user")
}

// TestDBRealmRepository_Authenticate_DBError verifies Authenticate returns a DB
// error (not ErrUserNotFound) when the underlying DB is closed.
func TestDBRealmRepository_Authenticate_DBError(t *testing.T) {
t.Parallel()

repo, ctx := setupMigratedClosedDBRepo(t)

result, err := repo.Authenticate(ctx, googleUuid.Must(googleUuid.NewV7()).String(), "user", "pass")
require.Error(t, err)
require.NotNil(t, result)
require.Equal(t, "database error", result.Error)
require.Contains(t, err.Error(), "failed to lookup user")
}
