// Copyright (c) 2025 Justin Cranford

package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// RecoveryCodeRepository implements RecoveryCodeRepository using GORM.
type RecoveryCodeRepository struct {
	db *gorm.DB
}

// NewRecoveryCodeRepository creates a new GORM-based recovery code repository.
func NewRecoveryCodeRepository(db *gorm.DB) *RecoveryCodeRepository {
	return &RecoveryCodeRepository{db: db}
}

// Create stores a new recovery code.
func (r *RecoveryCodeRepository) Create(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(code).Error; err != nil {
		return fmt.Errorf("failed to create recovery code: %w", err)
	}

	return nil
}

// CreateBatch stores multiple recovery codes in a transaction.
func (r *RecoveryCodeRepository) CreateBatch(ctx context.Context, codes []*cryptoutilIdentityDomain.RecoveryCode) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(codes).Error; err != nil {
		return fmt.Errorf("failed to create recovery codes batch: %w", err)
	}

	return nil
}

// GetByUserID retrieves all recovery codes for a user.
func (r *RecoveryCodeRepository) GetByUserID(ctx context.Context, userID googleUuid.UUID) ([]*cryptoutilIdentityDomain.RecoveryCode, error) {
	var codes []*cryptoutilIdentityDomain.RecoveryCode

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("user_id = ?", userID.String()).
		Find(&codes).Error; err != nil {
		return nil, fmt.Errorf("failed to get recovery codes by user ID: %w", err)
	}

	return codes, nil
}

// GetByID retrieves a recovery code by ID.
func (r *RecoveryCodeRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.RecoveryCode, error) {
	var code cryptoutilIdentityDomain.RecoveryCode

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", id.String()).
		First(&code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound
		}

		return nil, fmt.Errorf("failed to get recovery code by ID: %w", err)
	}

	return &code, nil
}

// Update modifies an existing recovery code.
func (r *RecoveryCodeRepository) Update(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Save(code).Error; err != nil {
		return fmt.Errorf("failed to update recovery code: %w", err)
	}

	return nil
}

// DeleteByUserID removes all recovery codes for a user.
func (r *RecoveryCodeRepository) DeleteByUserID(ctx context.Context, userID googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("user_id = ?", userID.String()).
		Delete(&cryptoutilIdentityDomain.RecoveryCode{}).Error; err != nil {
		return fmt.Errorf("failed to delete recovery codes by user ID: %w", err)
	}

	return nil
}

// DeleteExpired removes all expired recovery codes.
func (r *RecoveryCodeRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.RecoveryCode{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired recovery codes: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// CountUnused returns count of unused, unexpired codes for a user.
func (r *RecoveryCodeRepository) CountUnused(ctx context.Context, userID googleUuid.UUID) (int64, error) {
	var count int64

	if err := getDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilIdentityDomain.RecoveryCode{}).
		Where("user_id = ? AND used = ? AND expires_at > ?", userID.String(), false, time.Now().UTC()).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unused recovery codes: %w", err)
	}

	return count, nil
}
