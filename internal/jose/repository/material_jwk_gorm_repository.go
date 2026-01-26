// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// materialJWKGormRepository is a GORM-based implementation of MaterialJWKRepository.
type materialJWKGormRepository struct {
	db *gorm.DB
}

// NewMaterialJWKRepository creates a new MaterialJWKRepository.
func NewMaterialJWKRepository(db *gorm.DB) MaterialJWKRepository {
	return &materialJWKGormRepository{db: db}
}

// Create creates a new Material JWK.
func (r *materialJWKGormRepository) Create(ctx context.Context, materialJWK *cryptoutilJoseDomain.MaterialJWK) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Create(materialJWK).Error; err != nil {
		return fmt.Errorf("failed to create material JWK: %w", err)
	}

	return nil
}

// GetByID retrieves a Material JWK by its UUID.
func (r *materialJWKGormRepository) GetByID(ctx context.Context, materialJWKID googleUuid.UUID) (*cryptoutilJoseDomain.MaterialJWK, error) {
	var materialJWK cryptoutilJoseDomain.MaterialJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("id = ?", materialJWKID).
		First(&materialJWK).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("material JWK not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get material JWK: %w", err)
	}

	return &materialJWK, nil
}

// GetByMaterialKID retrieves a Material JWK by its material KID within an Elastic JWK.
func (r *materialJWKGormRepository) GetByMaterialKID(ctx context.Context, elasticJWKID googleUuid.UUID, materialKID string) (*cryptoutilJoseDomain.MaterialJWK, error) {
	var materialJWK cryptoutilJoseDomain.MaterialJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("elastic_jwk_id = ? AND material_kid = ?", elasticJWKID, materialKID).
		First(&materialJWK).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("material JWK not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get material JWK: %w", err)
	}

	return &materialJWK, nil
}

// GetByMaterialKIDGlobal retrieves a Material JWK by its material KID globally.
// The material_kid is globally unique (UUID), so no elastic_jwk_id is needed.
func (r *materialJWKGormRepository) GetByMaterialKIDGlobal(ctx context.Context, materialKID string) (*cryptoutilJoseDomain.MaterialJWK, error) {
	var materialJWK cryptoutilJoseDomain.MaterialJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("material_kid = ?", materialKID).
		First(&materialJWK).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("material JWK not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get material JWK: %w", err)
	}

	return &materialJWK, nil
}

// ListByElasticJWK retrieves all Material JWKs for an Elastic JWK with pagination.
func (r *materialJWKGormRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]cryptoutilJoseDomain.MaterialJWK, error) {
	var materialJWKs []cryptoutilJoseDomain.MaterialJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("elastic_jwk_id = ?", elasticJWKID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&materialJWKs).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to list material JWKs: %w", err)
	}

	return materialJWKs, nil
}

// GetActiveMaterial retrieves the currently active Material JWK for an Elastic JWK.
func (r *materialJWKGormRepository) GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilJoseDomain.MaterialJWK, error) {
	var materialJWK cryptoutilJoseDomain.MaterialJWK

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("elastic_jwk_id = ? AND active = ?", elasticJWKID, true).
		First(&materialJWK).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no active material JWK found: %w", err)
		}

		return nil, fmt.Errorf("failed to get active material JWK: %w", err)
	}

	return &materialJWK, nil
}

// RotateMaterial performs key rotation atomically.
func (r *materialJWKGormRepository) RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *cryptoutilJoseDomain.MaterialJWK) error {
	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Retire the currently active material.
		now := time.Now().UTC().UnixMilli()

		err := tx.Model(&cryptoutilJoseDomain.MaterialJWK{}).
			Where("elastic_jwk_id = ? AND active = ?", elasticJWKID, true).
			Updates(map[string]any{
				"active":     false,
				"retired_at": now,
			}).Error
		if err != nil {
			return fmt.Errorf("failed to retire active material: %w", err)
		}

		// Ensure new material is set as active.
		newMaterial.Active = true
		newMaterial.ElasticJWKID = elasticJWKID

		// Create the new active material.
		if err := tx.Create(newMaterial).Error; err != nil {
			return fmt.Errorf("failed to create new material: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to rotate material: %w", err)
	}

	return nil
}

// CountMaterials returns the count of Material JWKs for an Elastic JWK.
func (r *materialJWKGormRepository) CountMaterials(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error) {
	var count int64

	err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilJoseDomain.MaterialJWK{}).
		Where("elastic_jwk_id = ?", elasticJWKID).
		Count(&count).
		Error
	if err != nil {
		return 0, fmt.Errorf("failed to count material JWKs: %w", err)
	}

	return count, nil
}
