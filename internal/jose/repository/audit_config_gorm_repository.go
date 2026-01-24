// Copyright (c) 2025 Justin Cranford
//
//

// Package repository provides data access layer implementations for JOSE domain models.
package repository

import (
	"context"
	"errors"
	"fmt"

	"cryptoutil/internal/jose/domain"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// Verify interface compliance at compile time.
var _ AuditConfigRepository = (*AuditConfigGormRepository)(nil)

// AuditConfigGormRepository implements AuditConfigRepository using GORM.
type AuditConfigGormRepository struct {
	db *gorm.DB
}

// NewAuditConfigGormRepository creates a new AuditConfigGormRepository.
func NewAuditConfigGormRepository(db *gorm.DB) *AuditConfigGormRepository {
	return &AuditConfigGormRepository{db: db}
}

// Get retrieves audit config for a tenant and operation.
func (r *AuditConfigGormRepository) Get(ctx context.Context, tenantID googleUuid.UUID, operation string) (*domain.AuditConfig, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	var config domain.AuditConfig

	err := db.WithContext(ctx).Where("tenant_id = ? AND operation = ?", tenantID.String(), operation).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("audit config not found for tenant %s operation %s", tenantID, operation)
		}

		return nil, fmt.Errorf("failed to get audit config: %w", err)
	}

	return &config, nil
}

// GetAll retrieves all audit configs for a tenant.
func (r *AuditConfigGormRepository) GetAll(ctx context.Context, tenantID googleUuid.UUID) ([]domain.AuditConfig, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	var configs []domain.AuditConfig

	err := db.WithContext(ctx).Where("tenant_id = ?", tenantID.String()).Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get audit configs: %w", err)
	}

	return configs, nil
}

// Upsert creates or updates audit config for a tenant and operation.
func (r *AuditConfigGormRepository) Upsert(ctx context.Context, config *domain.AuditConfig) error {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	// Use Save which does upsert based on primary key (tenant_id + operation).
	err := db.WithContext(ctx).Save(config).Error
	if err != nil {
		return fmt.Errorf("failed to upsert audit config: %w", err)
	}

	return nil
}

// Delete removes audit config for a tenant and operation.
func (r *AuditConfigGormRepository) Delete(ctx context.Context, tenantID googleUuid.UUID, operation string) error {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	result := db.WithContext(ctx).Where("tenant_id = ? AND operation = ?", tenantID.String(), operation).Delete(&domain.AuditConfig{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete audit config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("audit config not found for tenant %s operation %s", tenantID, operation)
	}

	return nil
}

// IsEnabled checks if audit is enabled for a tenant and operation, and returns the sampling rate.
// If no config exists, returns disabled=false with zero sampling rate (conservative default).
func (r *AuditConfigGormRepository) IsEnabled(ctx context.Context, tenantID googleUuid.UUID, operation string) (bool, float64, error) {
	config, err := r.Get(ctx, tenantID, operation)
	if err != nil {
		// If no config exists, audit is disabled (conservative default).
		return false, 0.0, nil //nolint:nilerr // Not found is not an error for IsEnabled.
	}

	return config.Enabled, config.SamplingRate, nil
}
