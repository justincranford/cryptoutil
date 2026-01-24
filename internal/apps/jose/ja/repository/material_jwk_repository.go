// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// MaterialJWKRepository defines the interface for Material JWK persistence.
type MaterialJWKRepository interface {
	// Create stores a new Material JWK.
	Create(ctx context.Context, materialJWK *cryptoutilAppsJoseJaDomain.MaterialJWK) error

	// GetByMaterialKID retrieves a Material JWK by its material KID.
	GetByMaterialKID(ctx context.Context, materialKID string) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error)

	// GetByID retrieves a Material JWK by its UUID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error)

	// GetActiveMaterial retrieves the active Material JWK for an Elastic JWK.
	GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error)

	// ListByElasticJWK retrieves all Material JWKs for an Elastic JWK.
	ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.MaterialJWK, int64, error)

	// RotateMaterial atomically retires the current active material and activates a new one.
	RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *cryptoutilAppsJoseJaDomain.MaterialJWK) error

	// RetireMaterial marks a material as retired.
	RetireMaterial(ctx context.Context, id googleUuid.UUID) error

	// Delete removes a Material JWK by ID.
	Delete(ctx context.Context, id googleUuid.UUID) error

	// CountMaterials counts the number of materials for an Elastic JWK.
	CountMaterials(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error)
}

// gormMaterialJWKRepository implements MaterialJWKRepository using GORM.
type gormMaterialJWKRepository struct {
	db *gorm.DB
}

// NewMaterialJWKRepository creates a new GORM-based Material JWK repository.
func NewMaterialJWKRepository(db *gorm.DB) MaterialJWKRepository {
	return &gormMaterialJWKRepository{db: db}
}

// Create stores a new Material JWK.
func (r *gormMaterialJWKRepository) Create(ctx context.Context, materialJWK *cryptoutilAppsJoseJaDomain.MaterialJWK) error {
	if err := r.db.WithContext(ctx).Create(materialJWK).Error; err != nil {
		return fmt.Errorf("failed to create material JWK: %w", err)
	}

	return nil
}

// GetByMaterialKID retrieves a Material JWK by its material KID.
func (r *gormMaterialJWKRepository) GetByMaterialKID(ctx context.Context, materialKID string) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	var materialJWK cryptoutilAppsJoseJaDomain.MaterialJWK
	if err := r.db.WithContext(ctx).
		Where("material_kid = ?", materialKID).
		First(&materialJWK).Error; err != nil {
		return nil, fmt.Errorf("failed to get material JWK by KID: %w", err)
	}

	return &materialJWK, nil
}

// GetByID retrieves a Material JWK by its UUID.
func (r *gormMaterialJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	var materialJWK cryptoutilAppsJoseJaDomain.MaterialJWK
	if err := r.db.WithContext(ctx).
		Where("id = ?", id.String()).
		First(&materialJWK).Error; err != nil {
		return nil, fmt.Errorf("failed to get material JWK by ID: %w", err)
	}

	return &materialJWK, nil
}

// GetActiveMaterial retrieves the active Material JWK for an Elastic JWK.
func (r *gormMaterialJWKRepository) GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	var materialJWK cryptoutilAppsJoseJaDomain.MaterialJWK
	if err := r.db.WithContext(ctx).
		Where("elastic_jwk_id = ? AND active = ?", elasticJWKID.String(), true).
		First(&materialJWK).Error; err != nil {
		return nil, fmt.Errorf("failed to get active material JWK: %w", err)
	}

	return &materialJWK, nil
}

// ListByElasticJWK retrieves all Material JWKs for an Elastic JWK.
func (r *gormMaterialJWKRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.MaterialJWK, int64, error) {
	var (
		materialJWKs []*cryptoutilAppsJoseJaDomain.MaterialJWK
		total        int64
	)

	// Count total.

	if err := r.db.WithContext(ctx).
		Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("elastic_jwk_id = ?", elasticJWKID.String()).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count material JWKs: %w", err)
	}

	// Fetch page.
	if err := r.db.WithContext(ctx).
		Where("elastic_jwk_id = ?", elasticJWKID.String()).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&materialJWKs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list material JWKs: %w", err)
	}

	return materialJWKs, total, nil
}

// RotateMaterial atomically retires the current active material and activates a new one.
func (r *gormMaterialJWKRepository) RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *cryptoutilAppsJoseJaDomain.MaterialJWK) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Retire current active material.
		now := time.Now()
		if err := tx.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
			Where("elastic_jwk_id = ? AND active = ?", elasticJWKID.String(), true).
			Updates(map[string]any{
				"active":     false,
				"retired_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to retire current material: %w", err)
		}

		// Create new active material.
		newMaterial.Active = true
		if err := tx.Create(newMaterial).Error; err != nil {
			return fmt.Errorf("failed to create new material: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to rotate material JWK: %w", err)
	}

	return nil
}

// RetireMaterial marks a material as retired.
func (r *gormMaterialJWKRepository) RetireMaterial(ctx context.Context, id googleUuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", id.String()).
		Updates(map[string]any{
			"active":     false,
			"retired_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to retire material JWK: %w", err)
	}

	return nil
}

// Delete removes a Material JWK by ID.
func (r *gormMaterialJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("id = ?", id.String()).
		Delete(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Error; err != nil {
		return fmt.Errorf("failed to delete material JWK: %w", err)
	}

	return nil
}

// CountMaterials counts the number of materials for an Elastic JWK.
func (r *gormMaterialJWKRepository) CountMaterials(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("elastic_jwk_id = ?", elasticJWKID.String()).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count material JWKs: %w", err)
	}

	return count, nil
}
