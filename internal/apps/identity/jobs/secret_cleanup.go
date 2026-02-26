// Copyright (c) 2025 Justin Cranford
//
//

package jobs

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// CleanupExpiredSecrets marks active secrets with past expiration as expired.
// This job should run periodically (e.g., hourly) to enforce grace period expiration.
func CleanupExpiredSecrets(ctx context.Context, db *gorm.DB) (int64, error) {
	now := time.Now().UTC()

	// Update active secrets with expiration in the past.
	result := db.WithContext(ctx).
		Model(&cryptoutilIdentityDomain.ClientSecretVersion{}).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", cryptoutilIdentityDomain.SecretStatusActive, now).
		Update(cryptoutilSharedMagic.StringStatus, cryptoutilIdentityDomain.SecretStatusExpired)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired secrets: %w", result.Error)
	}

	return result.RowsAffected, nil
}
