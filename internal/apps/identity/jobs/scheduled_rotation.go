// Copyright (c) 2025 Justin Cranford
//
//

package jobs

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRotation "cryptoutil/internal/apps/identity/rotation"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ScheduledRotationConfig holds configuration for scheduled rotation.
type ScheduledRotationConfig struct {
	ExpirationThreshold time.Duration // Rotate secrets expiring within this duration.
	CheckInterval       time.Duration // How often to check for secrets needing rotation.
}

// DefaultScheduledRotationConfig returns default configuration.
func DefaultScheduledRotationConfig() *ScheduledRotationConfig {
	return &ScheduledRotationConfig{
		ExpirationThreshold: cryptoutilSharedMagic.SecretRotationExpirationThreshold,
		CheckInterval:       cryptoutilSharedMagic.SecretRotationCheckInterval,
	}
}

// ScheduledRotation checks for secrets approaching expiration and rotates them automatically.
func ScheduledRotation(ctx context.Context, db *gorm.DB, config *ScheduledRotationConfig) (int, error) {
	if config == nil {
		config = DefaultScheduledRotationConfig()
	}

	now := time.Now().UTC()
	expirationCutoff := now.Add(config.ExpirationThreshold)

	// Find active secrets expiring within threshold.
	var secretsToRotate []cryptoutilIdentityDomain.ClientSecretVersion

	err := db.WithContext(ctx).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?",
			cryptoutilIdentityDomain.SecretStatusActive, expirationCutoff).
		Order("expires_at ASC").
		Find(&secretsToRotate).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query secrets for rotation: %w", err)
	}

	if len(secretsToRotate) == 0 {
		return 0, nil
	}

	// Group secrets by client ID (only rotate latest active secret per client).
	clientSecrets := make(map[string]cryptoutilIdentityDomain.ClientSecretVersion)
	rotatedCount := 0

	for _, secret := range secretsToRotate {
		existing, exists := clientSecrets[secret.ClientID.String()]
		if !exists || secret.Version > existing.Version {
			clientSecrets[secret.ClientID.String()] = secret
		}
	}

	// Filter out secrets that are NOT the latest active version for their client.
	// If a secret is expiring but is not the latest version, it's already in grace period from a prior rotation.
	for clientID, secret := range clientSecrets {
		var latestActiveVersion cryptoutilIdentityDomain.ClientSecretVersion

		err := db.WithContext(ctx).
			Where("client_id = ? AND status = ?", secret.ClientID, cryptoutilIdentityDomain.SecretStatusActive).
			Order("version DESC").
			First(&latestActiveVersion).Error
		if err != nil {
			return rotatedCount, fmt.Errorf("failed to query latest version for client %s: %w", clientID, err)
		}

		// Only keep this secret if it IS the latest active version.
		if secret.Version != latestActiveVersion.Version {
			delete(clientSecrets, clientID)
		}
	}

	// Rotate each client's secret.
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(db)

	for _, secret := range clientSecrets {
		// Calculate grace period: time until current secret expires.
		gracePeriod := secret.ExpiresAt.Sub(now)
		if gracePeriod < 0 {
			gracePeriod = 0
		}

		_, err := rotationService.RotateClientSecret(
			ctx,
			secret.ClientID,
			gracePeriod,
			cryptoutilSharedMagic.SystemInitiatorName,
			fmt.Sprintf("Automatic rotation (expiring in %s)", gracePeriod),
		)
		if err != nil {
			return rotatedCount, fmt.Errorf("failed to rotate secret for client %s: %w", secret.ClientID, err)
		}

		rotatedCount++
		// DEBUG: Log each rotation.
		// fmt.Printf("DEBUG: Rotated client %s (version %d), count now %d\n", secret.ClientID, secret.Version, rotatedCount)
	}

	return rotatedCount, nil
}
