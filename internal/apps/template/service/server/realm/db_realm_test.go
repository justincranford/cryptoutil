// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite" // Use modernc CGO-free SQLite.

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique database name per test to avoid conflicts.
	dbName := fmt.Sprintf("file:test_%d?mode=memory&cache=private", time.Now().UTC().UnixNano())

	// Use database/sql with modernc.org/sqlite driver.
	sqlDB, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)

	// Configure connection pool per instructions.
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)

	// Wrap with GORM using sqlite Dialector.
	db, err := gorm.Open(gormsqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return db
}

func TestNewDBRealmRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		db      *gorm.DB
		policy  *PasswordPolicyConfig
		wantErr bool
	}{
		{
			name:    "nil database",
			db:      nil,
			policy:  nil,
			wantErr: true,
		},
		{
			name:    "valid database with default policy",
			db:      setupTestDB(t),
			policy:  nil,
			wantErr: false,
		},
		{
			name:    "valid database with custom policy",
			db:      setupTestDB(t),
			policy:  &PasswordPolicyConfig{Algorithm: "SHA-256", Iterations: 100000, SaltBytes: 32, HashBytes: 32},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, err := NewDBRealmRepository(tc.db, tc.policy)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, repo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, repo)
			}
		})
	}
}

func TestDBRealmRepository_Migrate(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	// Verify table exists.
	var count int64

	err = db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='kms_realm_users'").Scan(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestDBRealmRepository_CreateUser(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		user     *DBRealmUser
		password string
		wantErr  bool
	}{
		{
			name:     "nil user",
			user:     nil,
			password: googleUuid.Must(googleUuid.NewV7()).String(),
			wantErr:  true,
		},
		{
			name: "empty password",
			user: &DBRealmUser{
				ID:       "user-1",
				RealmID:  "realm-1",
				Username: "testuser",
				Enabled:  true,
			},
			password: "",
			wantErr:  true,
		},
		{
			name: "valid user",
			user: &DBRealmUser{
				ID:       "user-2",
				RealmID:  "realm-1",
				Username: "validuser",
				Email:    "valid@example.com",
				Roles:    `["user"]`,
				Enabled:  true,
			},
			password: "validpassword",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.CreateUser(ctx, tc.user, tc.password)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, tc.user.PasswordHash)
				require.NotZero(t, tc.user.CreatedAt)
			}
		})
	}
}

func TestDBRealmRepository_GetUser(t *testing.T) {
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
		ID:       "get-user-1",
		RealmID:  "realm-1",
		Username: "getuser",
		Enabled:  true,
	}
	err = repo.CreateUser(ctx, user, "password")
	require.NoError(t, err)

	// Get existing user.
	found, err := repo.GetUser(ctx, "get-user-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "getuser", found.Username)

	// Get non-existent user.
	notFound, err := repo.GetUser(ctx, "nonexistent")
	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, notFound)
}

func TestDBRealmRepository_GetUserByUsername(t *testing.T) {
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
		ID:       "username-user-1",
		RealmID:  "realm-1",
		Username: "byusername",
		Enabled:  true,
	}
	err = repo.CreateUser(ctx, user, "password")
	require.NoError(t, err)

	// Get existing user.
	found, err := repo.GetUserByUsername(ctx, "realm-1", "byusername")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "username-user-1", found.ID)

	// Get non-existent user.
	notFound, err := repo.GetUserByUsername(ctx, "realm-1", "nonexistent")
	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, notFound)

	// Wrong realm.
	wrongRealm, err := repo.GetUserByUsername(ctx, "realm-2", "byusername")
	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, wrongRealm)
}

func TestDBRealmRepository_Authenticate(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	policy := DefaultPasswordPolicy()

	repo, err := NewDBRealmRepository(db, &policy)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Migrate(ctx)
	require.NoError(t, err)

	// Create test users.
	activeUser := &DBRealmUser{
		ID:       "auth-user-1",
		RealmID:  "realm-1",
		Username: "activeuser",
		Enabled:  true,
	}
	err = repo.CreateUser(ctx, activeUser, "correctpassword")
	require.NoError(t, err)

	disabledUser := &DBRealmUser{
		ID:       "auth-user-2",
		RealmID:  "realm-1",
		Username: "disableduser",
		Enabled:  false,
	}
	err = repo.CreateUser(ctx, disabledUser, "password")
	require.NoError(t, err)

	tests := []struct {
		name        string
		realmID     string
		username    string
		password    string
		wantAuth    bool
		wantErrCode AuthErrorCode
	}{
		{
			name:        "successful authentication",
			realmID:     "realm-1",
			username:    "activeuser",
			password:    "correctpassword",
			wantAuth:    true,
			wantErrCode: AuthErrorNone,
		},
		{
			name:        "wrong password",
			realmID:     "realm-1",
			username:    "activeuser",
			password:    "wrongpassword",
			wantAuth:    false,
			wantErrCode: AuthErrorPasswordMismatch,
		},
		{
			name:        "user not found",
			realmID:     "realm-1",
			username:    "nonexistent",
			password:    googleUuid.Must(googleUuid.NewV7()).String(),
			wantAuth:    false,
			wantErrCode: AuthErrorUserNotFound,
		},
		{
			name:        "disabled user",
			realmID:     "realm-1",
			username:    "disableduser",
			password:    googleUuid.Must(googleUuid.NewV7()).String(),
			wantAuth:    false,
			wantErrCode: AuthErrorUserDisabled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.Authenticate(ctx, tc.realmID, tc.username, tc.password)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tc.wantAuth, result.Authenticated)
			require.Equal(t, tc.wantErrCode, result.ErrorCode)
		})
	}
}

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
	for i := 0; i < 5; i++ {
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
	require.Len(t, users, 5)

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
