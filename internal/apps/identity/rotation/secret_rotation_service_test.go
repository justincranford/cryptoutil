// Copyright (c) 2025 Justin Cranford
//
//

package rotation

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
	_ "modernc.org/sqlite" // Import modernc.org/sqlite driver.

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique database file for each test to prevent table conflicts.
	dbName := fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", t.Name())

	// Open SQLite with modernc.org/sqlite (CGO-free).
	sqlDB, err := sql.Open("sqlite", dbName)
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
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
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
	gracePeriod := 24 * time.Hour

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
	gracePeriod := 24 * time.Hour

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
	require.NoError(t, err)
	require.Nil(t, version, "Should return nil when no active version exists")

	// Rotate to create version 1.
	result, rotateErr := service.RotateClientSecret(
		ctx,
		clientID,
		24*time.Hour,
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
	gracePeriod := 24 * time.Hour

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
		24*time.Hour,
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

func TestRevokeSecretVersion_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())

	// Attempt to revoke non-existent version.
	err := service.RevokeSecretVersion(
		ctx,
		clientID,
		999,
		"admin-user",
		"test",
	)
	require.Error(t, err, "Revoking non-existent version should fail")
	require.Contains(t, err.Error(), "not found")
}

// TestGetActiveSecretVersions tests retrieving all active secret versions.
func TestGetActiveSecretVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupVersions    int
		revokeVersions   []int
		expectedActive   int
		expectedVersions []int
	}{
		{
			name:             "no_secrets",
			setupVersions:    0,
			revokeVersions:   nil,
			expectedActive:   0,
			expectedVersions: []int{},
		},
		{
			name:             "single_active_secret",
			setupVersions:    1,
			revokeVersions:   nil,
			expectedActive:   1,
			expectedVersions: []int{1},
		},
		{
			name:             "multiple_active_secrets",
			setupVersions:    3,
			revokeVersions:   nil,
			expectedActive:   3,
			expectedVersions: []int{3, 2, 1}, // DESC order.
		},
		{
			name:             "mixed_active_and_revoked",
			setupVersions:    4,
			revokeVersions:   []int{2},
			expectedActive:   3,
			expectedVersions: []int{4, 3, 1}, // Excludes version 2.
		},
		{
			name:             "all_revoked",
			setupVersions:    2,
			revokeVersions:   []int{1, 2},
			expectedActive:   0,
			expectedVersions: []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			service := NewSecretRotationService(db)
			ctx := context.Background()

			clientID := googleUuid.Must(googleUuid.NewV7())
			gracePeriod := 24 * time.Hour

			// Create secret versions.
			for i := 0; i < tc.setupVersions; i++ {
				_, err := service.RotateClientSecret(
					ctx,
					clientID,
					gracePeriod,
					"test-user",
					fmt.Sprintf("rotation-%d", i+1),
				)
				require.NoError(t, err)
			}

			// Revoke specified versions.
			for _, version := range tc.revokeVersions {
				err := service.RevokeSecretVersion(
					ctx,
					clientID,
					version,
					"admin-user",
					"test-revocation",
				)
				require.NoError(t, err)
			}

			// Get active secret versions.
			versions, err := service.GetActiveSecretVersions(ctx, clientID)
			require.NoError(t, err)

			require.Len(t, versions, tc.expectedActive, "Should return correct number of active versions")

			// Verify versions are in DESC order.
			for i, version := range versions {
				require.Equal(t, tc.expectedVersions[i], version.Version, "Version mismatch at index %d", i)
				require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, version.Status, "All returned versions should be active")
			}
		})
	}
}

// TestGenerateRandomSecret tests the internal secret generation function.
func TestGenerateRandomSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		length     int
		wantErr    bool
		minEncoded int // Minimum base64-encoded length.
	}{
		{
			name:       "standard_length_32",
			length:     32,
			wantErr:    false,
			minEncoded: 40, // base64(32 bytes) = ~43 chars.
		},
		{
			name:       "standard_length_64",
			length:     64,
			wantErr:    false,
			minEncoded: 80, // base64(64 bytes) = ~86 chars.
		},
		{
			name:       "minimum_length_16",
			length:     16,
			wantErr:    false,
			minEncoded: 20, // base64(16 bytes) = ~22 chars.
		},
		{
			name:       "zero_length",
			length:     0,
			wantErr:    false,
			minEncoded: 0, // base64(0 bytes) = empty string (expected behavior).
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			secret, err := generateRandomSecret(tc.length)

			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, secret)
			} else {
				require.NoError(t, err)

				// Special case: zero-length generates empty string.
				if tc.length == 0 {
					require.Empty(t, secret, "Zero-length secret should be empty string")
				} else {
					require.NotEmpty(t, secret, "Should generate non-empty secret")
					require.GreaterOrEqual(t, len(secret), tc.minEncoded, "Base64-encoded secret should meet minimum length")

					// Verify secret is URL-safe base64.
					require.NotContains(t, secret, "+", "Should use URL-safe encoding")
					require.NotContains(t, secret, "/", "Should use URL-safe encoding")
				}
			}
		})
	}

	// Test uniqueness - generate multiple secrets and verify they're different.
	t.Run("uniqueness", func(t *testing.T) {
		t.Parallel()

		const (
			numSecrets   = 10
			secretLength = 32
		)

		secrets := make(map[string]bool)

		for i := 0; i < numSecrets; i++ {
			secret, err := generateRandomSecret(secretLength)
			require.NoError(t, err)

			// Check for duplicates.
			require.False(t, secrets[secret], "Generated secret should be unique (duplicate found)")

			secrets[secret] = true
		}

		require.Len(t, secrets, numSecrets, "Should generate exactly %d unique secrets", numSecrets)
	})
}

func TestSecretRotationService_GetActiveSecretVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	service := NewSecretRotationService(db)

	// Create test client.
	clientID, err := googleUuid.NewV7()
	require.NoError(t, err)

	// Create multiple secret versions with different statuses.
	activeVersion1 := &cryptoutilIdentityDomain.ClientSecretVersion{
		ID:        googleUuid.New(),
		ClientID:  clientID,
		Version:   1,
		Status:    cryptoutilIdentityDomain.SecretStatusActive,
		CreatedAt: time.Now().UTC().Add(-48 * time.Hour),
	}

	activeVersion2 := &cryptoutilIdentityDomain.ClientSecretVersion{
		ID:        googleUuid.New(),
		ClientID:  clientID,
		Version:   2,
		Status:    cryptoutilIdentityDomain.SecretStatusActive,
		CreatedAt: time.Now().UTC().Add(-24 * time.Hour),
	}

	revokedAt := time.Now().UTC().Add(-6 * time.Hour)

	revokedVersion := &cryptoutilIdentityDomain.ClientSecretVersion{
		ID:        googleUuid.New(),
		ClientID:  clientID,
		Version:   3,
		Status:    cryptoutilIdentityDomain.SecretStatusRevoked,
		CreatedAt: time.Now().UTC().Add(-12 * time.Hour),
		RevokedAt: &revokedAt,
	}

	// Insert versions.
	require.NoError(t, db.WithContext(ctx).Create(activeVersion1).Error)
	require.NoError(t, db.WithContext(ctx).Create(activeVersion2).Error)
	require.NoError(t, db.WithContext(ctx).Create(revokedVersion).Error)

	// Get active versions - should only return active ones, ordered by version DESC.
	activeVersions, err := service.GetActiveSecretVersions(ctx, clientID)
	require.NoError(t, err)
	require.Len(t, activeVersions, 2, "Should return only active versions")

	// Verify order (DESC by version).
	require.Equal(t, 2, activeVersions[0].Version, "First version should be 2")
	require.Equal(t, 1, activeVersions[1].Version, "Second version should be 1")
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, activeVersions[0].Status)
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, activeVersions[1].Status)
}
