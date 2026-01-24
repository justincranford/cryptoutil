// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
)

// MaterialJWKRepository manages Material JWKs (versioned key material) for Elastic JWKs.
type MaterialJWKRepository interface {
	// Create creates a new Material JWK.
	Create(ctx context.Context, materialJWK *cryptoutilJoseDomain.MaterialJWK) error

	// GetByID retrieves a Material JWK by its UUID.
	GetByID(ctx context.Context, materialJWKID googleUuid.UUID) (*cryptoutilJoseDomain.MaterialJWK, error)

	// GetByMaterialKID retrieves a Material JWK by its material KID within an Elastic JWK.
	// Used for lookups that know the elastic JWK context.
	GetByMaterialKID(ctx context.Context, elasticJWKID googleUuid.UUID, materialKID string) (*cryptoutilJoseDomain.MaterialJWK, error)

	// GetByMaterialKIDGlobal retrieves a Material JWK by its material KID globally.
	// Used for decrypt/verify operations that only have the material_kid from JWS/JWE headers.
	GetByMaterialKIDGlobal(ctx context.Context, materialKID string) (*cryptoutilJoseDomain.MaterialJWK, error)

	// ListByElasticJWK retrieves all Material JWKs for an Elastic JWK with pagination.
	ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]cryptoutilJoseDomain.MaterialJWK, error)

	// GetActiveMaterial retrieves the currently active Material JWK for an Elastic JWK.
	// Used for encrypt/sign operations that use the current key version.
	GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilJoseDomain.MaterialJWK, error)

	// RotateMaterial performs key rotation:
	// 1. Sets old active material's retired_at to current time
	// 2. Inserts new material with active = TRUE
	// This operation is performed in a transaction.
	RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *cryptoutilJoseDomain.MaterialJWK) error

	// CountMaterials returns the count of Material JWKs for an Elastic JWK.
	// Used to enforce the max_materials limit (default 1000).
	CountMaterials(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error)
}
