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

// ConsentDecisionRepository implements repository.ConsentDecisionRepository using GORM.
type ConsentDecisionRepository struct {
	db *gorm.DB
}

// NewConsentDecisionRepository creates a new GORM-based consent decision repository.
func NewConsentDecisionRepository(db *gorm.DB) *ConsentDecisionRepository {
	return &ConsentDecisionRepository{db: db}
}

// Create creates a new consent decision.
func (r *ConsentDecisionRepository) Create(ctx context.Context, consent *cryptoutilIdentityDomain.ConsentDecision) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(consent).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to create consent decision: %w", err),
		)
	}

	return nil
}

// GetByID retrieves a consent decision by ID.
func (r *ConsentDecisionRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.ConsentDecision, error) {
	var consent cryptoutilIdentityDomain.ConsentDecision

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", id).
		First(&consent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrConsentNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get consent decision by ID: %w", err),
		)
	}

	return &consent, nil
}

// GetByUserClientScope retrieves a consent decision by user, client, and scope.
func (r *ConsentDecisionRepository) GetByUserClientScope(ctx context.Context, userID googleUuid.UUID, clientID, scope string) (*cryptoutilIdentityDomain.ConsentDecision, error) {
	var consent cryptoutilIdentityDomain.ConsentDecision

	now := time.Now().UTC()

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("user_id = ? AND client_id = ? AND scope = ? AND revoked_at IS NULL AND expires_at > ?", userID, clientID, scope, now).
		First(&consent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrConsentNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get consent decision by user/client/scope: %w", err),
		)
	}

	return &consent, nil
}

// Update updates an existing consent decision.
func (r *ConsentDecisionRepository) Update(ctx context.Context, consent *cryptoutilIdentityDomain.ConsentDecision) error {
	result := getDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilIdentityDomain.ConsentDecision{}).
		Where("id = ?", consent.ID).
		Updates(consent)

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to update consent decision: %w", result.Error),
		)
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrConsentNotFound
	}

	return nil
}

// Delete deletes a consent decision by ID.
func (r *ConsentDecisionRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", id).
		Delete(&cryptoutilIdentityDomain.ConsentDecision{})

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to delete consent decision: %w", result.Error),
		)
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrConsentNotFound
	}

	return nil
}

// RevokeByID revokes a consent decision by ID.
func (r *ConsentDecisionRepository) RevokeByID(ctx context.Context, id googleUuid.UUID) error {
	now := time.Now().UTC()
	result := getDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilIdentityDomain.ConsentDecision{}).
		Where("id = ?", id).
		Update("revoked_at", now)

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to revoke consent decision: %w", result.Error),
		)
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrConsentNotFound
	}

	return nil
}

// DeleteExpired deletes all expired consent decisions.
func (r *ConsentDecisionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.ConsentDecision{})

	if result.Error != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to delete expired consent decisions: %w", result.Error),
		)
	}

	return result.RowsAffected, nil
}
