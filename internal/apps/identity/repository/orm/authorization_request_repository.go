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

// AuthorizationRequestRepository implements repository.AuthorizationRequestRepository using GORM.
type AuthorizationRequestRepository struct {
	db *gorm.DB
}

// NewAuthorizationRequestRepository creates a new GORM-based authorization request repository.
func NewAuthorizationRequestRepository(db *gorm.DB) *AuthorizationRequestRepository {
	return &AuthorizationRequestRepository{db: db}
}

// Create creates a new authorization request.
func (r *AuthorizationRequestRepository) Create(ctx context.Context, request *cryptoutilIdentityDomain.AuthorizationRequest) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(request).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to create authorization request: %w", err),
		)
	}

	return nil
}

// GetByID retrieves an authorization request by ID.
func (r *AuthorizationRequestRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.AuthorizationRequest, error) {
	var request cryptoutilIdentityDomain.AuthorizationRequest

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", id).
		First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrAuthorizationRequestNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get authorization request by ID: %w", err),
		)
	}

	return &request, nil
}

// GetByCode retrieves an authorization request by authorization code.
func (r *AuthorizationRequestRepository) GetByCode(ctx context.Context, code string) (*cryptoutilIdentityDomain.AuthorizationRequest, error) {
	var request cryptoutilIdentityDomain.AuthorizationRequest

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("code = ?", code).
		First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrAuthorizationRequestNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get authorization request by code: %w", err),
		)
	}

	return &request, nil
}

// Update updates an existing authorization request.
func (r *AuthorizationRequestRepository) Update(ctx context.Context, request *cryptoutilIdentityDomain.AuthorizationRequest) error {
	result := getDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilIdentityDomain.AuthorizationRequest{}).
		Where("id = ?", request.ID).
		Updates(request)

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to update authorization request: %w", result.Error),
		)
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrAuthorizationRequestNotFound
	}

	return nil
}

// Delete deletes an authorization request by ID.
func (r *AuthorizationRequestRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", id).
		Delete(&cryptoutilIdentityDomain.AuthorizationRequest{})

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to delete authorization request: %w", result.Error),
		)
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrAuthorizationRequestNotFound
	}

	return nil
}

// DeleteExpired deletes all expired authorization requests.
func (r *AuthorizationRequestRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.AuthorizationRequest{})

	if result.Error != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to delete expired authorization requests: %w", result.Error),
		)
	}

	return result.RowsAffected, nil
}
