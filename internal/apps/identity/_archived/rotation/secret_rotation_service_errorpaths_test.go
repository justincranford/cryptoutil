// Copyright (c) 2025 Justin Cranford
//
//

package rotation

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // Import modernc.org/sqlite driver.

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// setupTestDBWithTables creates a test DB and returns it along with the underlying sql.DB for manipulation.
func setupTestDBWithTables(t *testing.T) (*gorm.DB, *sql.DB) {
	t.Helper()

	dbName := fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", t.Name())

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbName)
	require.NoError(t, err, "Failed to open SQL database")

	ctx := context.Background()

	_, pragmaErr := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, pragmaErr, "Failed to enable WAL mode")

	_, timeoutErr := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, timeoutErr, "Failed to set busy timeout")

	db, gormErr := gorm.Open(sqlite.Dialector{
		Conn: sqlDB,
	}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, gormErr, "Failed to create GORM database")

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(0)

	migrateErr := db.AutoMigrate(
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	)
	require.NoError(t, migrateErr, "Failed to run migrations")

	return db, sqlDB
}

func TestRotateClientSecret_GenerateSecretError(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	service.generateSecretFn = func(_ int) (string, error) {
		return "", fmt.Errorf("simulated generate failure")
	}

	ctx := context.Background()
	clientID := googleUuid.Must(googleUuid.NewV7())

	_, err := service.RotateClientSecret(ctx, clientID, time.Hour, "test", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate new secret")
}

func TestRotateClientSecret_HashSecretError(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	service.hashSecretFn = func(_ string) (string, error) {
		return "", fmt.Errorf("simulated hash failure")
	}

	ctx := context.Background()
	clientID := googleUuid.Must(googleUuid.NewV7())

	_, err := service.RotateClientSecret(ctx, clientID, time.Hour, "test", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash new secret")
}

func TestRotateClientSecret_DBQueryError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupTestDBWithTables(t)
	service := NewSecretRotationService(db)

	ctx := context.Background()

	// Drop the versions table to cause a DB error during the transaction.
	_, dropErr := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS client_secret_versions")
	require.NoError(t, dropErr)

	clientID := googleUuid.Must(googleUuid.NewV7())

	_, err := service.RotateClientSecret(ctx, clientID, time.Hour, "test", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rotation transaction failed")
}

func TestRotateClientSecret_EventCreateError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupTestDBWithTables(t)
	service := NewSecretRotationService(db)

	ctx := context.Background()

	// Drop the events table to cause an error only at event creation.
	_, dropErr := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS key_rotation_events")
	require.NoError(t, dropErr)

	clientID := googleUuid.Must(googleUuid.NewV7())

	_, err := service.RotateClientSecret(ctx, clientID, time.Hour, "test", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rotation transaction failed")
}

func TestGetActiveSecretVersion_DBError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupTestDBWithTables(t)
	service := NewSecretRotationService(db)

	ctx := context.Background()

	_, dropErr := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS client_secret_versions")
	require.NoError(t, dropErr)

	clientID := googleUuid.Must(googleUuid.NewV7())

	_, err := service.GetActiveSecretVersion(ctx, clientID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to query active version")
}

func TestValidateSecretDuringGracePeriod_DBError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupTestDBWithTables(t)
	service := NewSecretRotationService(db)

	ctx := context.Background()

	_, dropErr := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS client_secret_versions")
	require.NoError(t, dropErr)

	clientID := googleUuid.Must(googleUuid.NewV7())

	_, _, err := service.ValidateSecretDuringGracePeriod(ctx, clientID, "secret")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to query active versions")
}

func TestGetActiveSecretVersions_DBError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupTestDBWithTables(t)
	service := NewSecretRotationService(db)

	ctx := context.Background()

	_, dropErr := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS client_secret_versions")
	require.NoError(t, dropErr)

	clientID := googleUuid.Must(googleUuid.NewV7())

	_, err := service.GetActiveSecretVersions(ctx, clientID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to query active secrets")
}

func TestRevokeSecretVersion_DBQueryError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupTestDBWithTables(t)
	service := NewSecretRotationService(db)

	ctx := context.Background()

	_, dropErr := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS client_secret_versions")
	require.NoError(t, dropErr)

	clientID := googleUuid.Must(googleUuid.NewV7())

	err := service.RevokeSecretVersion(ctx, clientID, 1, "admin", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to revoke secret version")
}

func TestRevokeSecretVersion_EventCreateError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupTestDBWithTables(t)
	service := NewSecretRotationService(db)

	ctx := context.Background()
	clientID := googleUuid.Must(googleUuid.NewV7())

	// Create a version first (events table still exists).
	_, rotateErr := service.RotateClientSecret(ctx, clientID, time.Hour, "test", "test")
	require.NoError(t, rotateErr)

	// Now drop the events table.
	_, dropErr := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS key_rotation_events")
	require.NoError(t, dropErr)

	// Revoke should fail at event creation.
	err := service.RevokeSecretVersion(ctx, clientID, 1, "admin", "incident")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to revoke secret version")
}
