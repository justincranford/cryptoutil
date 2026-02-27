// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// jtiReplayCacheRepository implements JTIReplayCacheRepository using GORM.
type jtiReplayCacheRepository struct {
	db *gorm.DB
}

// NewJTIReplayCacheRepository creates a new JTI replay cache repository.
func NewJTIReplayCacheRepository(db *gorm.DB) JTIReplayCacheRepository {
	return &jtiReplayCacheRepository{db: db}
}

// Store stores a JTI to prevent replay attacks.
// Returns error if JTI already exists (replay detected).
func (r *jtiReplayCacheRepository) Store(ctx context.Context, jti string, clientID googleUuid.UUID, expiresAt time.Time) error {
	entry := &cryptoutilIdentityDomain.JTIReplayCache{
		JTI:       jti,
		ClientID:  clientID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}

	// Use GORM Create which will fail on duplicate primary key.
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		// Check if error is duplicate key constraint.
		if isDuplicateKeyError(err) {
			return fmt.Errorf("JTI replay detected: jti %s already used", jti)
		}

		return fmt.Errorf("failed to store JTI: %w", err)
	}

	return nil
}

// Exists checks if a JTI exists in the cache.
func (r *jtiReplayCacheRepository) Exists(ctx context.Context, jti string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&cryptoutilIdentityDomain.JTIReplayCache{}).
		Where("jti = ?", jti).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check JTI existence: %w", err)
	}

	return count > 0, nil
}

// DeleteExpired removes expired JTI entries from the cache.
func (r *jtiReplayCacheRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.JTIReplayCache{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired JTI entries: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// isDuplicateKeyError checks if error is a duplicate key constraint violation.
// Works for both SQLite and PostgreSQL.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// SQLite: "UNIQUE constraint failed"
	// PostgreSQL: "duplicate key value violates unique constraint"
	return contains(errMsg, "UNIQUE constraint failed") || contains(errMsg, "duplicate key value")
}

// contains checks if string s contains substring substr (case-insensitive check not needed for error messages).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
