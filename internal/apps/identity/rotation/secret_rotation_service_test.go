// Copyright (c) 2025 Justin Cranford
//
//

package rotation

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
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

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique database file for each test to prevent table conflicts.
	dbName := fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", t.Name())

	// Open SQLite with modernc.org/sqlite (CGO-free).
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbName)
	require.NoError(t, err, "Failed to open SQL database")

	// Apply SQLite PRAGMA settings for concurrent operations.
	ctx := context.Background()

	_, pragmaErr := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, pragmaErr, "Failed to enable WAL mode")

	_, timeoutErr := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, timeoutErr, "Failed to set busy timeout")

	// Create GORM DB from existing sql.DB connection.
	db, gormErr := gorm.Open(sqlite.Dialector{
		Conn: sqlDB,
	}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, gormErr, "Failed to create GORM database")

	// Configure connection pool for GORM transaction pattern.
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(0)

	// Auto-migrate domain models.
	migrateErr := db.AutoMigrate(
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	)
	require.NoError(t, migrateErr, "Failed to run migrations")

	return db
}

func TestRotateClientSecret_FirstRotation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())
	gracePeriod := cryptoutilSharedMagic.HoursPerDay * time.Hour

	result, err := service.RotateClientSecret(
		ctx,
		clientID,
		gracePeriod,
		"test-user",
		"initial setup",
	)

	require.NoError(t, err, "First rotation should succeed")
	require.NotNil(t, result)
	require.Equal(t, 0, result.OldVersion, "First rotation should have no old version")
	require.Equal(t, 1, result.NewVersion, "First rotation should create version 1")
	require.NotEmpty(t, result.NewSecretPlaintext, "Should generate new secret")
	require.NotEqual(t, googleUuid.Nil, result.EventID, "Should create rotation event")
}

func TestRotateClientSecret_SubsequentRotation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())
	gracePeriod := cryptoutilSharedMagic.HoursPerDay * time.Hour

	// First rotation.
	result1, err1 := service.RotateClientSecret(
		ctx,
		clientID,
		gracePeriod,
		"test-user",
		"initial setup",
	)
	require.NoError(t, err1)
	require.Equal(t, 1, result1.NewVersion)

	// Second rotation.
	result2, err2 := service.RotateClientSecret(
		ctx,
		clientID,
		gracePeriod,
		"test-user",
		"scheduled rotation",
	)
	require.NoError(t, err2)
	require.NotNil(t, result2)
	require.Equal(t, 1, result2.OldVersion, "Should reference version 1 as old")
	require.Equal(t, 2, result2.NewVersion, "Should create version 2")
	require.NotEmpty(t, result2.NewSecretPlaintext)

	// Verify old version has expiration set.
	var oldVersion cryptoutilIdentityDomain.ClientSecretVersion

	err := db.Where("client_id = ? AND version = ?", clientID, 1).
		First(&oldVersion).Error

	require.NoError(t, err)
	require.NotNil(t, oldVersion.ExpiresAt, "Old version should have expiration")
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, oldVersion.Status, "Old version should still be active during grace period")
}

func TestGetActiveSecretVersion(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())

	// No active version initially.
	version, err := service.GetActiveSecretVersion(ctx, clientID)
	require.ErrorIs(t, err, ErrNoActiveSecretVersion)
	require.Nil(t, version, "Should return nil when no active version exists")

	// Rotate to create version 1.
	result, rotateErr := service.RotateClientSecret(
		ctx,
		clientID,
		cryptoutilSharedMagic.HoursPerDay*time.Hour,
		"test-user",
		"test",
	)
	require.NoError(t, rotateErr)

	// Query active version.
	version, err = service.GetActiveSecretVersion(ctx, clientID)
	require.NoError(t, err)
	require.NotNil(t, version)
	require.Equal(t, result.NewVersion, version.Version)
}

func TestValidateSecretDuringGracePeriod(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())
	gracePeriod := cryptoutilSharedMagic.HoursPerDay * time.Hour

	// First rotation.
	result1, err1 := service.RotateClientSecret(
		ctx,
		clientID,
		gracePeriod,
		"test-user",
		"initial",
	)
	require.NoError(t, err1)

	secret1 := result1.NewSecretPlaintext

	// Second rotation (version 1 now in grace period).
	result2, err2 := service.RotateClientSecret(
		ctx,
		clientID,
		gracePeriod,
		"test-user",
		"rotation",
	)
	require.NoError(t, err2)

	secret2 := result2.NewSecretPlaintext

	// Both secrets should validate during grace period.
	valid1, version1, validateErr1 := service.ValidateSecretDuringGracePeriod(ctx, clientID, secret1)
	require.NoError(t, validateErr1)
	require.True(t, valid1, "Old secret should be valid during grace period")
	require.Equal(t, result1.NewVersion, version1, "Should return version 1")

	valid2, version2, validateErr2 := service.ValidateSecretDuringGracePeriod(ctx, clientID, secret2)
	require.NoError(t, validateErr2)
	require.True(t, valid2, "New secret should be valid")
	require.Equal(t, result2.NewVersion, version2, "Should return version 2")

	// Invalid secret should not validate.
	validInvalid, _, validateErrInvalid := service.ValidateSecretDuringGracePeriod(
		ctx,
		clientID,
		"invalid-secret",
	)
	require.NoError(t, validateErrInvalid)
	require.False(t, validInvalid, "Invalid secret should not validate")
}

func TestRevokeSecretVersion(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())

	// Create version 1.
	result, rotateErr := service.RotateClientSecret(
		ctx,
		clientID,
		cryptoutilSharedMagic.HoursPerDay*time.Hour,
		"test-user",
		"initial",
	)
	require.NoError(t, rotateErr)

	// Revoke version 1.
	revokeErr := service.RevokeSecretVersion(
		ctx,
		clientID,
		result.NewVersion,
		"admin-user",
		"security incident",
	)
	require.NoError(t, revokeErr)

	// Verify version 1 is revoked.
	var version cryptoutilIdentityDomain.ClientSecretVersion

	err := db.Where("client_id = ? AND version = ?", clientID, result.NewVersion).
		First(&version).Error

	require.NoError(t, err)
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusRevoked, version.Status)
	require.NotNil(t, version.RevokedAt)
	require.Equal(t, "admin-user", version.RevokedBy)

	// Verify revocation event was created.
	var event cryptoutilIdentityDomain.KeyRotationEvent

	eventErr := db.Where("event_type = ? AND key_id = ?", cryptoutilIdentityDomain.EventTypeRevocation, clientID.String()).
		Order("timestamp DESC").
		First(&event).Error

	require.NoError(t, eventErr)
	require.Equal(t, cryptoutilIdentityDomain.EventTypeRevocation, event.EventType)
	require.Equal(t, "security incident", event.Reason)
}
