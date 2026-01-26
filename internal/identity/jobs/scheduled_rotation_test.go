// Copyright (c) 2025 Justin Cranford
//
//

package jobs

import (
	"context"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRotation "cryptoutil/internal/identity/rotation"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

func TestScheduledRotation_NoExpiringSecrets(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret (version 1 auto-generated).
	client := createTestClient(t, db)

	// Get initial secret version.
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(db)
	versions, err := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1, "Client should have 1 active secret after creation")

	// Set expiration far in future (30 days).
	futureExpiration := time.Now().UTC().Add(30 * 24 * time.Hour)
	err = db.Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
		Where("client_id = ? AND version = ?", client.ID, versions[0].Version).
		Update("expires_at", futureExpiration).Error
	require.NoError(t, err)

	// Run scheduled rotation (threshold 7 days).
	config := &ScheduledRotationConfig{
		ExpirationThreshold: 7 * 24 * time.Hour,
		CheckInterval:       time.Hour,
	}

	rotatedCount, err2 := ScheduledRotation(ctx, db, config)
	require.NoError(t, err2)
	require.Equal(t, 0, rotatedCount, "Should not rotate secrets expiring >7 days")

	// Verify no new secret version created.
	versionsAfter, err3 := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err3)
	require.Len(t, versionsAfter, 1)
	require.Equal(t, versions[0].Version, versionsAfter[0].Version)
}

func TestScheduledRotation_OneExpiringSecret(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret (version 1 auto-generated).
	client := createTestClient(t, db)

	// Get initial secret version.
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(db)
	versions, err := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1, "Client should have 1 active secret after creation")

	// Set expiration within threshold (3 days).
	soonExpiration := time.Now().UTC().Add(3 * 24 * time.Hour)
	err = db.Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
		Where("client_id = ? AND version = ?", client.ID, versions[0].Version).
		Update("expires_at", soonExpiration).Error
	require.NoError(t, err)

	// Verify only 1 client exists before rotation.
	var clientCount int64

	db.Model(&cryptoutilIdentityDomain.Client{}).Count(&clientCount)
	require.Equal(t, int64(1), clientCount, "Should have exactly 1 client before rotation")

	// Run scheduled rotation (threshold 7 days).
	config := &ScheduledRotationConfig{
		ExpirationThreshold: 7 * 24 * time.Hour,
		CheckInterval:       time.Hour,
	}

	rotatedCount, err2 := ScheduledRotation(ctx, db, config)
	require.NoError(t, err2)
	require.Equal(t, 1, rotatedCount, "Should rotate exactly 1 client")

	// Verify new secret version created.
	versionsAfter, err3 := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err3)
	require.Len(t, versionsAfter, 2, "Should have 2 active versions (old + new)")
	require.Equal(t, 2, versionsAfter[0].Version, "Should have version 2 first (DESC order)")
	require.Equal(t, 1, versionsAfter[1].Version, "Should have version 1 second (DESC order)")
}

func TestScheduledRotation_MultipleExpiringSecrets(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create 3 clients with secrets (version 1 auto-generated).
	client1 := createTestClient(t, db)
	client2 := createTestClient(t, db)
	client3 := createTestClient(t, db)
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(db)

	soonExpiration := time.Now().UTC().Add(3 * 24 * time.Hour)

	// Set expiration for all 3 clients.
	for _, clientID := range []googleUuid.UUID{client1.ID, client2.ID, client3.ID} {
		versions, err := rotationService.GetActiveSecretVersions(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, versions, 1)

		err = db.Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
			Where("client_id = ? AND version = ?", clientID, versions[0].Version).
			Update("expires_at", soonExpiration).Error
		require.NoError(t, err)
	}

	// Run scheduled rotation (threshold 7 days).
	config := &ScheduledRotationConfig{
		ExpirationThreshold: 7 * 24 * time.Hour,
		CheckInterval:       time.Hour,
	}

	rotatedCount, err := ScheduledRotation(ctx, db, config)
	require.NoError(t, err)
	require.Equal(t, 3, rotatedCount, "Should rotate all 3 clients")

	// Verify each client has 2 active versions.
	for _, clientID := range []googleUuid.UUID{client1.ID, client2.ID, client3.ID} {
		versionsAfter, err2 := rotationService.GetActiveSecretVersions(ctx, clientID)
		require.NoError(t, err2)
		require.Len(t, versionsAfter, 2, "Each client should have 2 active versions")
	}
}

func TestScheduledRotation_SecretsOutsideThreshold(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create 2 clients with secrets (version 1 auto-generated).
	client1 := createTestClient(t, db)
	client2 := createTestClient(t, db)
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(db)

	// Client 1: expiring soon (3 days).
	versions1, err := rotationService.GetActiveSecretVersions(ctx, client1.ID)
	require.NoError(t, err)
	require.Len(t, versions1, 1, "Client1 should have 1 active secret after creation")

	soonExpiration := time.Now().UTC().Add(3 * 24 * time.Hour)
	err = db.Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
		Where("client_id = ? AND version = ?", client1.ID, versions1[0].Version).
		Update("expires_at", soonExpiration).Error
	require.NoError(t, err)

	// Client 2: expiring later (10 days).
	versions2, err2 := rotationService.GetActiveSecretVersions(ctx, client2.ID)
	require.NoError(t, err2)
	require.Len(t, versions2, 1, "Client2 should have 1 active secret after creation")

	laterExpiration := time.Now().UTC().Add(10 * 24 * time.Hour)
	err = db.Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
		Where("client_id = ? AND version = ?", client2.ID, versions2[0].Version).
		Update("expires_at", laterExpiration).Error
	require.NoError(t, err)

	// Run scheduled rotation (threshold 7 days).
	config := &ScheduledRotationConfig{
		ExpirationThreshold: 7 * 24 * time.Hour,
		CheckInterval:       time.Hour,
	}

	rotatedCount, err3 := ScheduledRotation(ctx, db, config)
	require.NoError(t, err3)
	require.Equal(t, 1, rotatedCount, "Should rotate only client1 (3 days), not client2 (10 days)")

	// Verify client1 has 2 active versions.
	versionsAfter1, err4 := rotationService.GetActiveSecretVersions(ctx, client1.ID)
	require.NoError(t, err4)
	require.Len(t, versionsAfter1, 2, "Client1 should have 2 active versions")

	// Verify client2 still has 1 active version.
	versionsAfter2, err5 := rotationService.GetActiveSecretVersions(ctx, client2.ID)
	require.NoError(t, err5)
	require.Len(t, versionsAfter2, 1, "Client2 should still have 1 active version")
}

func TestScheduledRotation_DefaultConfig(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret (version 1 auto-generated).
	client := createTestClient(t, db)
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(db)

	versions, err := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1, "Client should have 1 active secret after creation")

	// Set expiration within default threshold (3 days < 7 days default).
	soonExpiration := time.Now().UTC().Add(3 * 24 * time.Hour)
	err = db.Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
		Where("client_id = ? AND version = ?", client.ID, versions[0].Version).
		Update("expires_at", soonExpiration).Error
	require.NoError(t, err)

	// Run with nil config (uses default).
	rotatedCount, err2 := ScheduledRotation(ctx, db, nil)
	require.NoError(t, err2)
	require.Equal(t, 1, rotatedCount, "Should rotate using default config (7 days threshold)")

	// Verify new secret version created.
	versionsAfter, err3 := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err3)
	require.Len(t, versionsAfter, 2)
}

func TestScheduledRotation_AlreadyRotatedSecrets(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret (version 1 auto-generated).
	client := createTestClient(t, db)
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(db)

	versions, err := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1, "Client should have 1 active secret after creation")

	// Set version 1 to expire soon.
	soonExpiration := time.Now().UTC().Add(3 * 24 * time.Hour)
	err = db.Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
		Where("client_id = ? AND version = ?", client.ID, versions[0].Version).
		Update("expires_at", soonExpiration).Error
	require.NoError(t, err)

	// Manually rotate (creates version 2, sets version 1 expiration to grace period end).
	_, err = rotationService.RotateClientSecret(
		ctx, client.ID, 24*time.Hour, "test-admin", "manual rotation")
	require.NoError(t, err)

	// Verify version 2 exists and version 1 has updated expiration.
	versionsAfter, err2 := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err2)
	require.Len(t, versionsAfter, 2)

	// After manual rotation, version 2 has NO expiration (newest).
	// Version 1 has expiration set to grace period end (24 hours from rotation).
	// ScheduledRotation will find version 2 as the latest active secret with no expiration.
	// Since version 2 has no expiration, it won't be rotated (not expiring within threshold).
	config := &ScheduledRotationConfig{
		ExpirationThreshold: 7 * 24 * time.Hour,
		CheckInterval:       time.Hour,
	}

	rotatedCount, err3 := ScheduledRotation(ctx, db, config)
	require.NoError(t, err3)
	require.Equal(t, 0, rotatedCount, "Should skip rotation (version 2 has no expiration)")

	// Verify still 2 active versions (no duplicate rotation).
	versionsFinal, err4 := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err4)
	require.Len(t, versionsFinal, 2)
}

// createTestClient creates a client for testing.
func createTestClient(t *testing.T, db *gorm.DB) *cryptoutilIdentityDomain.Client {
	t.Helper()

	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
		ClientSecret:            "will-be-replaced-by-create",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}

	// Use the same transaction pattern as ClientRepositoryGORM.Create().
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create client record.
		if err := tx.Create(client).Error; err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Generate and hash initial secret (version 1).
		initialSecret := "test-secret-" + googleUuid.Must(googleUuid.NewV7()).String()

		secretHash, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(initialSecret)
		if err != nil {
			return fmt.Errorf("failed to hash initial secret: %w", err)
		} // Create ClientSecretVersion.

		version := &cryptoutilIdentityDomain.ClientSecretVersion{
			ID:         googleUuid.Must(googleUuid.NewV7()),
			ClientID:   client.ID,
			Version:    1,
			SecretHash: secretHash,
			Status:     cryptoutilIdentityDomain.SecretStatusActive,
			CreatedAt:  time.Now().UTC(),
			ExpiresAt:  nil,
		}
		if err := tx.Create(version).Error; err != nil {
			return fmt.Errorf("failed to create initial secret version: %w", err)
		}

		// Create KeyRotationEvent audit log.
		oldVersion := 0
		newVersion := 1
		event := &cryptoutilIdentityDomain.KeyRotationEvent{
			ID:            googleUuid.Must(googleUuid.NewV7()),
			EventType:     "secret_created",
			KeyType:       "client_secret",
			KeyID:         client.ID.String(),
			Timestamp:     time.Now().UTC(),
			Initiator:     "system",
			OldKeyVersion: &oldVersion,
			NewKeyVersion: &newVersion,
			Reason:        "Initial client creation (test)",
			Success:       boolPtr(true),
		}

		return tx.Create(event).Error
	})
	require.NoError(t, err)

	// Version 1 secret created via transaction.
	return client
}
