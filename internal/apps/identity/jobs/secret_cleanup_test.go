// Copyright (c) 2025-2026 Justin Cranford.
//
//

package jobs

import (
	"context"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	cryptoutilTestHelpDB "cryptoutil/internal/apps-framework/service/test_help_db"
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
					SecretHash: testHash1,
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC(),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)), // Future expiration.
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
					SecretHash: testHash1,
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
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
					SecretHash: testHash1,
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(-1 * time.Hour)), // Expired.
				},
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   googleUuid.Must(googleUuid.NewV7()),
					Version:    2,
					SecretHash: "hash2",
					Status:     cryptoutilIdentityDomain.SecretStatusActive,
					CreatedAt:  time.Now().UTC(),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)), // Active.
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
					SecretHash: testHash1,
					Status:     cryptoutilIdentityDomain.SecretStatusExpired, // Already expired.
					CreatedAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
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
					SecretHash: testHash1,
					Status:     cryptoutilIdentityDomain.SecretStatusRevoked, // Revoked.
					CreatedAt:  time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
					ExpiresAt:  ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
					RevokedAt:  ptrTime(time.Now().UTC().Add(-cryptoutilSharedMagic.HashPrefixLength * time.Hour)),
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

	db := cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
	db.Logger = logger.Default.LogMode(logger.Silent)

	// Auto-migrate schema (clients, client_secret_versions, key_rotation_events).
	require.NoError(t, db.AutoMigrate(
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	))

	return db
}

// ptrTime returns a pointer to the given time.Time value.
func ptrTime(t time.Time) *time.Time {
	return &t
}
