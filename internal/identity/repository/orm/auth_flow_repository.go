// Copyright (c) 2025 Justin Cranford
//
//

// Package orm provides GORM-based repository implementations for identity domain entities.
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

// AuthFlowRepositoryGORM implements the AuthFlowRepository interface using GORM.
type AuthFlowRepositoryGORM struct {
	db *gorm.DB
}

// NewAuthFlowRepository creates a new AuthFlowRepositoryGORM.
func NewAuthFlowRepository(db *gorm.DB) *AuthFlowRepositoryGORM {
	return &AuthFlowRepositoryGORM{db: db}
}

// Create creates a new authorization flow.
func (r *AuthFlowRepositoryGORM) Create(ctx context.Context, flow *cryptoutilIdentityDomain.AuthFlow) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(flow).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create auth flow: %w", err))
	}

	return nil
}

// GetByID retrieves an authorization flow by ID.
func (r *AuthFlowRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.AuthFlow, error) {
	var flow cryptoutilIdentityDomain.AuthFlow
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&flow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrAuthFlowNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get auth flow by ID: %w", err))
	}

	return &flow, nil
}

// GetByName retrieves an authorization flow by name.
func (r *AuthFlowRepositoryGORM) GetByName(ctx context.Context, name string) (*cryptoutilIdentityDomain.AuthFlow, error) {
	var flow cryptoutilIdentityDomain.AuthFlow
	if err := getDB(ctx, r.db).WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&flow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrAuthFlowNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get auth flow by name: %w", err))
	}

	return &flow, nil
}

// Update updates an existing authorization flow.
func (r *AuthFlowRepositoryGORM) Update(ctx context.Context, flow *cryptoutilIdentityDomain.AuthFlow) error {
	flow.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(flow).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update auth flow: %w", err))
	}

	return nil
}

// Delete deletes an authorization flow by ID (soft delete).
func (r *AuthFlowRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.AuthFlow{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete auth flow: %w", err))
	}

	return nil
}

// List lists authorization flows with pagination.
func (r *AuthFlowRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.AuthFlow, error) {
	var flows []*cryptoutilIdentityDomain.AuthFlow
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&flows).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list auth flows: %w", err))
	}

	return flows, nil
}

// Count returns the total number of authorization flows.
func (r *AuthFlowRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.AuthFlow{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count auth flows: %w", err))
	}

	return count, nil
}
