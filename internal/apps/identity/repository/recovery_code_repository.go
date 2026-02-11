// Copyright (c) 2025 Justin Cranford

package repository

import (
	"context"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// RecoveryCodeRepository defines operations for recovery code persistence.
type RecoveryCodeRepository interface {
	// Create stores a new recovery code.
	Create(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error

	// CreateBatch stores multiple recovery codes in a transaction.
	CreateBatch(ctx context.Context, codes []*cryptoutilIdentityDomain.RecoveryCode) error

	// GetByUserID retrieves all recovery codes for a user.
	GetByUserID(ctx context.Context, userID googleUuid.UUID) ([]*cryptoutilIdentityDomain.RecoveryCode, error)

	// GetByID retrieves a recovery code by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.RecoveryCode, error)

	// Update modifies an existing recovery code (typically to mark as used).
	Update(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error

	// DeleteByUserID removes all recovery codes for a user (regeneration scenario).
	DeleteByUserID(ctx context.Context, userID googleUuid.UUID) error

	// DeleteExpired removes all expired recovery codes.
	DeleteExpired(ctx context.Context) (int64, error)

	// CountUnused returns count of unused, unexpired codes for a user.
	CountUnused(ctx context.Context, userID googleUuid.UUID) (int64, error)
}
