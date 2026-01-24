// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"errors"
	"fmt"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// elasticJWKGormRepository is a GORM-based implementation of ElasticJWKRepository.
type elasticJWKGormRepository struct {
	db *gorm.DB
}

// NewElasticJWKRepository creates a new ElasticJWKRepository.
func NewElasticJWKRepository(db *gorm.DB) ElasticJWKRepository {
	return &elasticJWKGormRepository{db: db}
}

// Create creates a new Elastic JWK with tenant isolation.
func (r *elasticJWKGormRepository) Create(ctx context.Context, elasticJWK *cryptoutilJoseDomain.ElasticJWK) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Create(elasticJWK).Error; err != nil {
		return fmt.Errorf("failed to create elastic JWK: %w", err)
	}

	return nil
}

// Get retrieves an Elastic JWK by tenant ID, realm ID, and KID with tenant enforcement.
func (r *elasticJWKGormRepository) Get(ctx context.Context, tenantID, realmID googleUuid.UUID, kid string) (*cryptoutilJoseDomain.ElasticJWK, error) {
	var elasticJWK cryptoutilJoseDomain.ElasticJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ? AND realm_id = ? AND kid = ?", tenantID, realmID, kid).
		First(&elasticJWK).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("elastic JWK not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	return &elasticJWK, nil
}

// GetByID retrieves an Elastic JWK by its ID (no tenant filtering).
func (r *elasticJWKGormRepository) GetByID(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilJoseDomain.ElasticJWK, error) {
	var elasticJWK cryptoutilJoseDomain.ElasticJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", elasticJWKID).
		First(&elasticJWK).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("elastic JWK not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	return &elasticJWK, nil
}

// List retrieves all Elastic JWKs for a tenant/realm with pagination.
func (r *elasticJWKGormRepository) List(ctx context.Context, tenantID, realmID googleUuid.UUID, offset, limit int) ([]cryptoutilJoseDomain.ElasticJWK, error) {
	var elasticJWKs []cryptoutilJoseDomain.ElasticJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ? AND realm_id = ?", tenantID, realmID).
		Offset(offset).
		Limit(limit).
		Find(&elasticJWKs).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to list elastic JWKs: %w", err)
	}

	return elasticJWKs, nil
}

// IncrementMaterialCount increments the material count for an Elastic JWK.
func (r *elasticJWKGormRepository) IncrementMaterialCount(ctx context.Context, elasticJWKID googleUuid.UUID) error {
	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilJoseDomain.ElasticJWK{}).
		Where("id = ?", elasticJWKID).
		Update("current_material_count", gorm.Expr("current_material_count + 1")).
		Error
	if err != nil {
		return fmt.Errorf("failed to increment material count: %w", err)
	}

	return nil
}
