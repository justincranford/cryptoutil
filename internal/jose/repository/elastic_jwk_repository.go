// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
)

// ElasticJWKRepository manages Elastic JWKs with multi-tenancy support.
type ElasticJWKRepository interface {
	// Create creates a new Elastic JWK.
	Create(ctx context.Context, elasticJWK *cryptoutilJoseDomain.ElasticJWK) error

	// Get retrieves an Elastic JWK by tenant ID, realm ID, and KID.
	Get(ctx context.Context, tenantID, realmID googleUuid.UUID, kid string) (*cryptoutilJoseDomain.ElasticJWK, error)

	// GetByID retrieves an Elastic JWK by its ID.
	GetByID(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilJoseDomain.ElasticJWK, error)

	// List retrieves all Elastic JWKs for a tenant/realm with pagination.
	List(ctx context.Context, tenantID, realmID googleUuid.UUID, offset, limit int) ([]cryptoutilJoseDomain.ElasticJWK, error)

	// IncrementMaterialCount increments the material count for an Elastic JWK.
	IncrementMaterialCount(ctx context.Context, elasticJWKID googleUuid.UUID) error
}
