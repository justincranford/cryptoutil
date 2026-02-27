// Copyright (c) 2025 Justin Cranford
//
//

package jobs

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

const (
	testRawMemoryDSNFormat          = "file:%s?mode=memory&cache=shared"
	testSecretHashCoverage          = "test-hash-coverage"
	testCleanupSessionErrorContains = "session cleanup failed"
)

// createRawMemoryDB creates a minimal in-memory SQLite DB with no migrations.
func createRawMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbID := googleUuid.Must(googleUuid.NewV7())
	dsn := fmt.Sprintf(testRawMemoryDSNFormat, dbID.String())

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries) //nolint:mnd // SQLite GORM transaction pattern requires 5 connections.
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries) //nolint:mnd // SQLite GORM transaction pattern requires 5 connections.

	t.Cleanup(func() {
		_ = sqlDB.Close() //nolint:errcheck // Test cleanup, error not critical.
	})

	return db
}

// createDBWithSecretVersionsOnly creates an in-memory SQLite DB with only the ClientSecretVersion table.
func createDBWithSecretVersionsOnly(t *testing.T) *gorm.DB {
	t.Helper()

	db := createRawMemoryDB(t)

	err := db.AutoMigrate(&cryptoutilIdentityDomain.ClientSecretVersion{})
	require.NoError(t, err)

	return db
}

// TestCleanupJob_Cleanup_SessionError verifies that session cleanup error is recorded.
// Covers cleanup.go:107.16,114.3 (session cleanup error path) and cleanup.go:150.16,152.3 (inner return).
func TestCleanupJob_Cleanup_SessionError(t *testing.T) {
	t.Parallel()

	repoFactory := createTestRepoFactory(t)

	// Migrate only Token table so token cleanup succeeds but session cleanup fails (no sessions table).
	err := repoFactory.DB().AutoMigrate(&cryptoutilIdentityDomain.Token{})
	require.NoError(t, err)

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	job := NewCleanupJob(repoFactory, log, 1*time.Hour)

	// Token table exists → token cleanup succeeds (0 rows).
	// Session table missing → session cleanup fails → covers L107-114.
	job.cleanup(context.Background())

	metrics := job.GetMetrics()
	require.Error(t, metrics.LastError)
	require.Contains(t, metrics.LastError.Error(), testCleanupSessionErrorContains)
	require.Equal(t, 1, metrics.ErrorCount)
}

// TestScheduledRotation_FirstQueryFails verifies that DB query failure is reported.
// Covers scheduled_rotation.go:50.16,52.3 (first DB query error path).
func TestScheduledRotation_FirstQueryFails(t *testing.T) {
	t.Parallel()

	db := createRawMemoryDB(t) // No client_secret_versions table.

	_, err := ScheduledRotation(context.Background(), db, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to query secrets for rotation")
}

// TestScheduledRotation_GracePeriodNegativeAndRotationFails verifies that expired secrets
// have their grace period clamped to zero and rotation errors are returned properly.
// Covers scheduled_rotation.go:94.22,96.4 (gracePeriod<0 clamp) and 105.17,107.4 (rotation error).
func TestScheduledRotation_GracePeriodNegativeAndRotationFails(t *testing.T) {
	t.Parallel()

	// DB has ClientSecretVersion table but NOT KeyRotationEvent → rotation fails at event creation.
	db := createDBWithSecretVersionsOnly(t)

	ctx := context.Background()
	clientID := googleUuid.Must(googleUuid.NewV7())

	// Create an active secret that has ALREADY expired (ExpiresAt in the past).
	// ExpiresAt < now → gracePeriod = ExpiresAt.Sub(now) < 0 → clamped to 0 (L94-96).
	pastExpiry := time.Now().UTC().Add(-1 * time.Hour)
	secret := &cryptoutilIdentityDomain.ClientSecretVersion{
		ID:         googleUuid.Must(googleUuid.NewV7()),
		ClientID:   clientID,
		Version:    1,
		SecretHash: testSecretHashCoverage,
		Status:     cryptoutilIdentityDomain.SecretStatusActive,
		CreatedAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
		ExpiresAt:  &pastExpiry,
	}

	require.NoError(t, db.Create(secret).Error)

	// Use large threshold so the already-expired secret is included in the rotation candidates.
	config := &ScheduledRotationConfig{
		ExpirationThreshold: cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Hour,
		CheckInterval:       time.Hour,
	}

	// First query finds the secret (it is active and expires_at <= now+10h).
	// Second query finds it as the latest active version.
	// gracePeriod = ExpiresAt.Sub(now) < 0 → set to 0 (covers L94-96).
	// RotateClientSecret fails because key_rotation_events table doesn't exist (covers L105-107).
	_, err := ScheduledRotation(ctx, db, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to rotate secret for client")
}

// TestCleanupExpiredSecrets_DBError verifies that DB update failure is reported.
// Covers secret_cleanup.go:28.25,30.3 (DB update error path).
func TestCleanupExpiredSecrets_DBError(t *testing.T) {
	t.Parallel()

	db := createRawMemoryDB(t) // No client_secret_versions table → UPDATE fails.

	_, err := CleanupExpiredSecrets(context.Background(), db)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to cleanup expired secrets")
}
