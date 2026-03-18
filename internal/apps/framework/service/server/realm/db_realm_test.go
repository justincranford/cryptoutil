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
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbName)
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
			policy:  &PasswordPolicyConfig{Algorithm: cryptoutilSharedMagic.PBKDF2DefaultAlgorithm, Iterations: cryptoutilSharedMagic.PBKDF2Iterations, SaltBytes: cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, HashBytes: cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes},
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
