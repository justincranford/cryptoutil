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

// AuthProfileRepositoryGORM implements the AuthProfileRepository interface using GORM.
type AuthProfileRepositoryGORM struct {
	db *gorm.DB
}

// NewAuthProfileRepository creates a new AuthProfileRepositoryGORM.
func NewAuthProfileRepository(db *gorm.DB) *AuthProfileRepositoryGORM {
	return &AuthProfileRepositoryGORM{db: db}
}

// Create creates a new authentication profile.
func (r *AuthProfileRepositoryGORM) Create(ctx context.Context, profile *cryptoutilIdentityDomain.AuthProfile) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(profile).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create auth profile: %w", err))
	}

	return nil
}

// GetByID retrieves an authentication profile by ID.
func (r *AuthProfileRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.AuthProfile, error) {
	var profile cryptoutilIdentityDomain.AuthProfile
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrAuthProfileNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get auth profile by ID: %w", err))
	}

	return &profile, nil
}

// GetByName retrieves an authentication profile by name.
func (r *AuthProfileRepositoryGORM) GetByName(ctx context.Context, name string) (*cryptoutilIdentityDomain.AuthProfile, error) {
	var profile cryptoutilIdentityDomain.AuthProfile
	if err := getDB(ctx, r.db).WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrAuthProfileNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get auth profile by name: %w", err))
	}

	return &profile, nil
}

// Update updates an existing authentication profile.
func (r *AuthProfileRepositoryGORM) Update(ctx context.Context, profile *cryptoutilIdentityDomain.AuthProfile) error {
	profile.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(profile).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update auth profile: %w", err))
	}

	return nil
}

// Delete deletes an authentication profile by ID (soft delete).
func (r *AuthProfileRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.AuthProfile{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete auth profile: %w", err))
	}

	return nil
}

// List lists authentication profiles with pagination.
func (r *AuthProfileRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.AuthProfile, error) {
	var profiles []*cryptoutilIdentityDomain.AuthProfile
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&profiles).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list auth profiles: %w", err))
	}

	return profiles, nil
}

// Count returns the total number of authentication profiles.
func (r *AuthProfileRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.AuthProfile{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count auth profiles: %w", err))
	}

	return count, nil
}
