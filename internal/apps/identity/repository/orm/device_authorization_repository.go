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

// DeviceAuthorizationRepository implements repository.DeviceAuthorizationRepository using GORM.
type DeviceAuthorizationRepository struct {
	db *gorm.DB
}

// NewDeviceAuthorizationRepository creates a new GORM-based device authorization repository.
func NewDeviceAuthorizationRepository(db *gorm.DB) *DeviceAuthorizationRepository {
	return &DeviceAuthorizationRepository{db: db}
}

// Create creates a new device authorization request.
func (r *DeviceAuthorizationRepository) Create(ctx context.Context, auth *cryptoutilIdentityDomain.DeviceAuthorization) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(auth).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to create device authorization: %w", err),
		)
	}

	return nil
}

// GetByDeviceCode retrieves a device authorization by device code.
func (r *DeviceAuthorizationRepository) GetByDeviceCode(ctx context.Context, deviceCode string) (*cryptoutilIdentityDomain.DeviceAuthorization, error) {
	var auth cryptoutilIdentityDomain.DeviceAuthorization

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("device_code = ?", deviceCode).
		First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get device authorization by device code: %w", err),
		)
	}

	return &auth, nil
}

// GetByUserCode retrieves a device authorization by user code.
func (r *DeviceAuthorizationRepository) GetByUserCode(ctx context.Context, userCode string) (*cryptoutilIdentityDomain.DeviceAuthorization, error) {
	var auth cryptoutilIdentityDomain.DeviceAuthorization

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("user_code = ?", userCode).
		First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get device authorization by user code: %w", err),
		)
	}

	return &auth, nil
}

// GetByID retrieves a device authorization by primary key UUID.
func (r *DeviceAuthorizationRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.DeviceAuthorization, error) {
	var auth cryptoutilIdentityDomain.DeviceAuthorization

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", id).
		First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get device authorization by ID: %w", err),
		)
	}

	return &auth, nil
}

// Update updates an existing device authorization.
func (r *DeviceAuthorizationRepository) Update(ctx context.Context, auth *cryptoutilIdentityDomain.DeviceAuthorization) error {
	result := getDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilIdentityDomain.DeviceAuthorization{}).
		Where("id = ?", auth.ID).
		Updates(auth)

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to update device authorization: %w", result.Error),
		)
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound
	}

	return nil
}

// DeleteExpired deletes all expired device authorizations.
func (r *DeviceAuthorizationRepository) DeleteExpired(ctx context.Context) error {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.DeviceAuthorization{})

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to delete expired device authorizations: %w", result.Error),
		)
	}

	return nil
}
