// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"

	cryptoutilJoseJADomain "cryptoutil/internal/apps/jose/ja/domain"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// ElasticJWKRepository defines the interface for Elastic JWK persistence.
// CRITICAL: Methods filter by tenant_id ONLY - realms are authn-only, NOT data scope.
type ElasticJWKRepository interface {
	// Create stores a new Elastic JWK.
	Create(ctx context.Context, elasticJWK *cryptoutilJoseJADomain.ElasticJWK) error

	// Get retrieves an Elastic JWK by KID within a tenant.
	Get(ctx context.Context, tenantID googleUuid.UUID, kid string) (*cryptoutilJoseJADomain.ElasticJWK, error)

	// GetByID retrieves an Elastic JWK by its UUID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilJoseJADomain.ElasticJWK, error)

	// List retrieves all Elastic JWKs for a tenant with pagination.
	List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilJoseJADomain.ElasticJWK, int64, error)

	// Update updates an existing Elastic JWK.
	Update(ctx context.Context, elasticJWK *cryptoutilJoseJADomain.ElasticJWK) error

	// Delete removes an Elastic JWK by ID.
	Delete(ctx context.Context, id googleUuid.UUID) error

	// IncrementMaterialCount atomically increments the material count.
	IncrementMaterialCount(ctx context.Context, id googleUuid.UUID) error

	// DecrementMaterialCount atomically decrements the material count.
	DecrementMaterialCount(ctx context.Context, id googleUuid.UUID) error
}

// gormElasticJWKRepository implements ElasticJWKRepository using GORM.
type gormElasticJWKRepository struct {
	db *gorm.DB
}

// NewElasticJWKRepository creates a new GORM-based Elastic JWK repository.
func NewElasticJWKRepository(db *gorm.DB) ElasticJWKRepository {
	return &gormElasticJWKRepository{db: db}
}

// Create stores a new Elastic JWK.
func (r *gormElasticJWKRepository) Create(ctx context.Context, elasticJWK *cryptoutilJoseJADomain.ElasticJWK) error {
	if err := r.db.WithContext(ctx).Create(elasticJWK).Error; err != nil {
		return fmt.Errorf("failed to create elastic JWK: %w", err)
	}

	return nil
}

// Get retrieves an Elastic JWK by KID within a tenant.
// CRITICAL: Filters by tenant_id ONLY - realms are authn-only, NOT data scope.
func (r *gormElasticJWKRepository) Get(ctx context.Context, tenantID googleUuid.UUID, kid string) (*cryptoutilJoseJADomain.ElasticJWK, error) {
	var elasticJWK cryptoutilJoseJADomain.ElasticJWK
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND kid = ?", tenantID.String(), kid).
		First(&elasticJWK).Error; err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	return &elasticJWK, nil
}

// GetByID retrieves an Elastic JWK by its UUID.
func (r *gormElasticJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilJoseJADomain.ElasticJWK, error) {
	var elasticJWK cryptoutilJoseJADomain.ElasticJWK
	if err := r.db.WithContext(ctx).
		Where("id = ?", id.String()).
		First(&elasticJWK).Error; err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK by ID: %w", err)
	}

	return &elasticJWK, nil
}

// List retrieves all Elastic JWKs for a tenant with pagination.
// CRITICAL: Filters by tenant_id ONLY - realms are authn-only, NOT data scope.
func (r *gormElasticJWKRepository) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilJoseJADomain.ElasticJWK, int64, error) {
	var (
		elasticJWKs []*cryptoutilJoseJADomain.ElasticJWK
		total       int64
	)

	// Count total.

	if err := r.db.WithContext(ctx).
		Model(&cryptoutilJoseJADomain.ElasticJWK{}).
		Where("tenant_id = ?", tenantID.String()).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count elastic JWKs: %w", err)
	}

	// Fetch page.
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&elasticJWKs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list elastic JWKs: %w", err)
	}

	return elasticJWKs, total, nil
}

// Update updates an existing Elastic JWK.
func (r *gormElasticJWKRepository) Update(ctx context.Context, elasticJWK *cryptoutilJoseJADomain.ElasticJWK) error {
	if err := r.db.WithContext(ctx).Save(elasticJWK).Error; err != nil {
		return fmt.Errorf("failed to update elastic JWK: %w", err)
	}

	return nil
}

// Delete removes an Elastic JWK by ID.
func (r *gormElasticJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("id = ?", id.String()).
		Delete(&cryptoutilJoseJADomain.ElasticJWK{}).Error; err != nil {
		return fmt.Errorf("failed to delete elastic JWK: %w", err)
	}

	return nil
}

// IncrementMaterialCount atomically increments the material count.
func (r *gormElasticJWKRepository) IncrementMaterialCount(ctx context.Context, id googleUuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&cryptoutilJoseJADomain.ElasticJWK{}).
		Where("id = ?", id.String()).
		UpdateColumn("current_material_count", gorm.Expr("current_material_count + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment material count: %w", err)
	}

	return nil
}

// DecrementMaterialCount atomically decrements the material count.
func (r *gormElasticJWKRepository) DecrementMaterialCount(ctx context.Context, id googleUuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&cryptoutilJoseJADomain.ElasticJWK{}).
		Where("id = ? AND current_material_count > 0", id.String()).
		UpdateColumn("current_material_count", gorm.Expr("current_material_count - 1")).Error; err != nil {
		return fmt.Errorf("failed to decrement material count: %w", err)
	}

	return nil
}
