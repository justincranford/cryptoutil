// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// MFAFactorRepositoryGORM implements the MFAFactorRepository interface using GORM.
type MFAFactorRepositoryGORM struct {
	db *gorm.DB
}

// NewMFAFactorRepository creates a new MFAFactorRepositoryGORM.
func NewMFAFactorRepository(db *gorm.DB) *MFAFactorRepositoryGORM {
	return &MFAFactorRepositoryGORM{db: db}
}

// Create creates a new MFA factor.
func (r *MFAFactorRepositoryGORM) Create(ctx context.Context, factor *cryptoutilIdentityDomain.MFAFactor) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(factor).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create MFA factor: %w", err))
	}

	return nil
}

// GetByID retrieves an MFA factor by ID.
func (r *MFAFactorRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.MFAFactor, error) {
	var factor cryptoutilIdentityDomain.MFAFactor
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&factor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrMFAFactorNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get MFA factor by ID: %w", err))
	}

	return &factor, nil
}

// GetByAuthProfileID retrieves MFA factors by authentication profile ID.
func (r *MFAFactorRepositoryGORM) GetByAuthProfileID(ctx context.Context, authProfileID googleUuid.UUID) ([]*cryptoutilIdentityDomain.MFAFactor, error) {
	var factors []*cryptoutilIdentityDomain.MFAFactor
	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("auth_profile_id = ? AND deleted_at IS NULL", authProfileID).
		Order("\"order\" ASC").
		Find(&factors).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get MFA factors by auth profile ID: %w", err))
	}

	return factors, nil
}

// Update updates an existing MFA factor.
func (r *MFAFactorRepositoryGORM) Update(ctx context.Context, factor *cryptoutilIdentityDomain.MFAFactor) error {
	factor.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(factor).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update MFA factor: %w", err))
	}

	return nil
}

// Delete deletes an MFA factor by ID (soft delete).
func (r *MFAFactorRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.MFAFactor{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete MFA factor: %w", err))
	}

	return nil
}

// List lists MFA factors with pagination.
func (r *MFAFactorRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.MFAFactor, error) {
	var factors []*cryptoutilIdentityDomain.MFAFactor
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&factors).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list MFA factors: %w", err))
	}

	return factors, nil
}

// Count returns the total number of MFA factors.
func (r *MFAFactorRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.MFAFactor{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count MFA factors: %w", err))
	}

	return count, nil
}
