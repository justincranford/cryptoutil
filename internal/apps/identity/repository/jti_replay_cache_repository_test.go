// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;"); err != nil {
		require.FailNowf(t, "failed to enable WAL mode", "%v", err)
	}

	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;"); err != nil {
		require.FailNowf(t, "failed to set busy timeout", "%v", err)
	}

	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	gormDB, err := db.DB()
	require.NoError(t, err)

	gormDB.SetMaxOpenConns(5)
	gormDB.SetMaxIdleConns(5)
	gormDB.SetConnMaxLifetime(0)
	gormDB.SetConnMaxIdleTime(0)

	// Auto-migrate JTI replay cache table.
	err = db.AutoMigrate(&cryptoutilIdentityDomain.JTIReplayCache{})
	require.NoError(t, err)

	return db
}

// TestJTIReplayCacheRepository_Store tests storing JTI entries.
func TestJTIReplayCacheRepository_Store(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := NewJTIReplayCacheRepository(db)

	ctx := context.Background()
	jti := "test-jti-" + googleUuid.NewString()
	clientID := googleUuid.New()
	expiresAt := time.Now().UTC().Add(10 * time.Minute)

	// First store should succeed
	err := repo.Store(ctx, jti, clientID, expiresAt)
	require.NoError(t, err)

	// Second store with same JTI should fail (replay detected)
	err = repo.Store(ctx, jti, clientID, expiresAt)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JTI replay detected")
}

// TestJTIReplayCacheRepository_Exists tests checking JTI existence.
func TestJTIReplayCacheRepository_Exists(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := NewJTIReplayCacheRepository(db)

	ctx := context.Background()
	jti := "test-jti-" + googleUuid.NewString()
	clientID := googleUuid.New()
	expiresAt := time.Now().UTC().Add(10 * time.Minute)

	// JTI should not exist initially
	exists, err := repo.Exists(ctx, jti)
	require.NoError(t, err)
	require.False(t, exists)

	// Store JTI
	err = repo.Store(ctx, jti, clientID, expiresAt)
	require.NoError(t, err)

	// JTI should now exist
	exists, err = repo.Exists(ctx, jti)
	require.NoError(t, err)
	require.True(t, exists)
}

// TestJTIReplayCacheRepository_DeleteExpired tests deleting expired JTI entries.
func TestJTIReplayCacheRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := NewJTIReplayCacheRepository(db)

	ctx := context.Background()
	clientID := googleUuid.New()

	// Store expired JTI (already expired)
	expiredJTI := "expired-jti-" + googleUuid.NewString()
	err := repo.Store(ctx, expiredJTI, clientID, time.Now().UTC().Add(-time.Minute))
	require.NoError(t, err)

	// Store valid JTI (expires in future)
	validJTI := "valid-jti-" + googleUuid.NewString()
	err = repo.Store(ctx, validJTI, clientID, time.Now().UTC().Add(10*time.Minute))
	require.NoError(t, err)

	// Delete expired entries
	deleted, err := repo.DeleteExpired(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, deleted, int64(1), "At least expired JTI should be deleted")

	// Expired JTI should be gone
	exists, err := repo.Exists(ctx, expiredJTI)
	require.NoError(t, err)
	require.False(t, exists, "Expired JTI should be deleted")

	// Valid JTI should still exist
	exists, err = repo.Exists(ctx, validJTI)
	require.NoError(t, err)
	require.True(t, exists, "Valid JTI should still exist")
}
