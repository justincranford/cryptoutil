// Copyright (c) 2025 Justin Cranford
//
//

package jobs

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

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestCleanupExpiredSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		secrets  []*cryptoutilIdentityDomain.ClientSecretVersion
		wantRows int64
	}{
		{
			name: "no expired secrets",
			secrets: []*cryptoutilIdentityDomain.ClientSecretVersion{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   googleUuid.Must(googleUuid.NewV7()),
					Version:    1,
					SecretHash: "hash1",
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC(),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(24 * time.Hour)), // Future expiration.
				},
			},
			wantRows: 0,
		},
		{
			name: "one expired secret",
			secrets: []*cryptoutilIdentityDomain.ClientSecretVersion{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   googleUuid.Must(googleUuid.NewV7()),
					Version:    1,
					SecretHash: "hash1",
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC().Add(-48 * time.Hour),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(-1 * time.Hour)), // Past expiration.
				},
			},
			wantRows: 1,
		},
		{
			name: "mixed active and expired secrets",
			secrets: []*cryptoutilIdentityDomain.ClientSecretVersion{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   googleUuid.Must(googleUuid.NewV7()),
					Version:    1,
					SecretHash: "hash1",
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC().Add(-48 * time.Hour),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(-1 * time.Hour)), // Expired.
				},
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   googleUuid.Must(googleUuid.NewV7()),
					Version:    2,
					SecretHash: "hash2",
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC(),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(24 * time.Hour)), // Active.
				},
			},
			wantRows: 1,
		},
		{
			name: "already expired secrets not updated",
			secrets: []*cryptoutilIdentityDomain.ClientSecretVersion{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   googleUuid.Must(googleUuid.NewV7()),
					Version:    1,
					SecretHash: "hash1",
					Status:     cryptoutilIdentityDomain.SecretStatusExpired, // Already expired.
					CreatedAt:  time.Now().UTC().Add(-48 * time.Hour),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
				},
			},
			wantRows: 0,
		},
		{
			name: "revoked secrets not updated",
			secrets: []*cryptoutilIdentityDomain.ClientSecretVersion{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   googleUuid.Must(googleUuid.NewV7()),
					Version:    1,
					SecretHash: "hash1",
					Status:     cryptoutilIdentityDomain.SecretStatusRevoked, // Revoked.
					CreatedAt:  time.Now().UTC().Add(-48 * time.Hour),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
					RevokedAt:  ptrTime(time.Now().UTC().Add(-12 * time.Hour)),
				},
			},
			wantRows: 0,
		},
	}

	for _, tc := range tests {
		// Capture range variable.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			ctx := context.Background()

			// Seed test data.
			for _, secret := range tc.secrets {
				err := db.Create(secret).Error
				require.NoError(t, err)
			}

			// Run cleanup.
			rowsAffected, err := CleanupExpiredSecrets(ctx, db)
			require.NoError(t, err)
			require.Equal(t, tc.wantRows, rowsAffected)

			// Verify status updated correctly.
			if tc.wantRows > 0 {
				var updatedSecrets []cryptoutilIdentityDomain.ClientSecretVersion

				err = db.Where("status = ?", cryptoutilIdentityDomain.SecretStatusExpired).Find(&updatedSecrets).Error
				require.NoError(t, err)
				require.Equal(t, int(tc.wantRows), len(updatedSecrets))
			}
		})
	}
}

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	dbID := googleUuid.Must(googleUuid.NewV7())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

	// Open database connection using modernc.org/sqlite (CGO-free).
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	// Apply SQLite PRAGMA settings.
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		require.FailNowf(t, "failed to enable WAL mode", "%v", err)
	}

	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		require.FailNowf(t, "failed to set busy timeout", "%v", err)
	}

	// Create GORM database with explicit connection.
	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate schema (clients, client_secret_versions, key_rotation_events).
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close() //nolint:errcheck // Test cleanup - error not critical.
	})

	return db
}

// ptrTime returns a pointer to the given time.Time value.
func ptrTime(t time.Time) *time.Time {
	return &t
}
