// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"

	"cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
)

// MaterialJWKRepository manages Material JWKs (versioned key material) for Elastic JWKs.
type MaterialJWKRepository interface {
	// Create creates a new Material JWK.
	Create(ctx context.Context, materialJWK *domain.MaterialJWK) error

	// GetByMaterialKID retrieves a Material JWK by its material KID.
	// Used for decrypt/verify operations that specify a specific key version.
	GetByMaterialKID(ctx context.Context, elasticJWKID googleUuid.UUID, materialKID string) (*domain.MaterialJWK, error)

	// ListByElasticJWK retrieves all Material JWKs for an Elastic JWK with pagination.
	ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]domain.MaterialJWK, error)

	// GetActiveMaterial retrieves the currently active Material JWK for an Elastic JWK.
	// Used for encrypt/sign operations that use the current key version.
	GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*domain.MaterialJWK, error)

	// RotateMaterial performs key rotation:
	// 1. Sets old active material's retired_at to current time
	// 2. Inserts new material with active = TRUE
	// This operation is performed in a transaction.
	RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *domain.MaterialJWK) error

	// CountMaterials returns the count of Material JWKs for an Elastic JWK.
	// Used to enforce the max_materials limit (default 1000).
	CountMaterials(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error)
}
