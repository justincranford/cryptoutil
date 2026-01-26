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

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// ClientProfileRepositoryGORM implements the ClientProfileRepository interface using GORM.
type ClientProfileRepositoryGORM struct {
	db *gorm.DB
}

// NewClientProfileRepository creates a new ClientProfileRepositoryGORM.
func NewClientProfileRepository(db *gorm.DB) *ClientProfileRepositoryGORM {
	return &ClientProfileRepositoryGORM{db: db}
}

// Create creates a new client profile.
func (r *ClientProfileRepositoryGORM) Create(ctx context.Context, profile *cryptoutilIdentityDomain.ClientProfile) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(profile).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create client profile: %w", err))
	}

	return nil
}

// GetByID retrieves a client profile by ID.
func (r *ClientProfileRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.ClientProfile, error) {
	var profile cryptoutilIdentityDomain.ClientProfile
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrClientProfileNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get client profile by ID: %w", err))
	}

	return &profile, nil
}

// GetByName retrieves a client profile by name.
func (r *ClientProfileRepositoryGORM) GetByName(ctx context.Context, name string) (*cryptoutilIdentityDomain.ClientProfile, error) {
	var profile cryptoutilIdentityDomain.ClientProfile
	if err := getDB(ctx, r.db).WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrClientProfileNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get client profile by name: %w", err))
	}

	return &profile, nil
}

// Update updates an existing client profile.
func (r *ClientProfileRepositoryGORM) Update(ctx context.Context, profile *cryptoutilIdentityDomain.ClientProfile) error {
	profile.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(profile).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update client profile: %w", err))
	}

	return nil
}

// Delete deletes a client profile by ID (soft delete).
func (r *ClientProfileRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.ClientProfile{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete client profile: %w", err))
	}

	return nil
}

// List lists client profiles with pagination.
func (r *ClientProfileRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.ClientProfile, error) {
	var profiles []*cryptoutilIdentityDomain.ClientProfile
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&profiles).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list client profiles: %w", err))
	}

	return profiles, nil
}

// Count returns the total number of client profiles.
func (r *ClientProfileRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.ClientProfile{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count client profiles: %w", err))
	}

	return count, nil
}
