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

// PushedAuthorizationRequestRepository implements repository.PushedAuthorizationRequestRepository using GORM.
type PushedAuthorizationRequestRepository struct {
	db *gorm.DB
}

// NewPushedAuthorizationRequestRepository creates a new GORM-based PAR repository.
func NewPushedAuthorizationRequestRepository(db *gorm.DB) *PushedAuthorizationRequestRepository {
	return &PushedAuthorizationRequestRepository{db: db}
}

// Create creates a new pushed authorization request.
func (r *PushedAuthorizationRequestRepository) Create(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(req).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to create pushed authorization request: %w", err),
		)
	}

	return nil
}

// GetByRequestURI retrieves a PAR by its request_uri value.
func (r *PushedAuthorizationRequestRepository) GetByRequestURI(ctx context.Context, requestURI string) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error) {
	var req cryptoutilIdentityDomain.PushedAuthorizationRequest

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("request_uri = ?", requestURI).
		First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrPushedAuthorizationRequestNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get pushed authorization request by request_uri: %w", err),
		)
	}

	return &req, nil
}

// GetByID retrieves a PAR by its primary key ID.
func (r *PushedAuthorizationRequestRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error) {
	var req cryptoutilIdentityDomain.PushedAuthorizationRequest

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", id).
		First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrPushedAuthorizationRequestNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get pushed authorization request by ID: %w", err),
		)
	}

	return &req, nil
}

// Update modifies an existing PAR (typically to mark as used).
func (r *PushedAuthorizationRequestRepository) Update(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Save(req).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to update pushed authorization request: %w", err),
		)
	}

	return nil
}

// DeleteExpired removes all expired PAR entries from the database.
func (r *PushedAuthorizationRequestRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.PushedAuthorizationRequest{})

	if result.Error != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to delete expired pushed authorization requests: %w", result.Error),
		)
	}

	return result.RowsAffected, nil
}
