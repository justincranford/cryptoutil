// Copyright (c) 2025 Justin Cranford
//
//

// Package repository provides data access layer implementations for JOSE-JA service.
package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditConfigRepository defines the interface for audit configuration persistence.
type AuditConfigRepository interface {
	// Get retrieves audit config for a tenant and operation.
	Get(ctx context.Context, tenantID googleUuid.UUID, operation string) (*cryptoutilAppsJoseJaDomain.AuditConfig, error)

	// GetAllForTenant retrieves all audit configs for a tenant.
	GetAllForTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsJoseJaDomain.AuditConfig, error)

	// Upsert creates or updates audit config for a tenant and operation.
	Upsert(ctx context.Context, config *cryptoutilAppsJoseJaDomain.AuditConfig) error

	// Delete removes audit config for a tenant and operation.
	Delete(ctx context.Context, tenantID googleUuid.UUID, operation string) error

	// ShouldAudit checks if an operation should be audited based on sampling.
	ShouldAudit(ctx context.Context, tenantID googleUuid.UUID, operation string) (bool, error)
}

// gormAuditConfigRepository implements AuditConfigRepository using GORM.
type gormAuditConfigRepository struct {
	db *gorm.DB
}

// NewAuditConfigRepository creates a new GORM-based audit config repository.
func NewAuditConfigRepository(db *gorm.DB) AuditConfigRepository {
	return &gormAuditConfigRepository{db: db}
}

// Get retrieves audit config for a tenant and operation.
func (r *gormAuditConfigRepository) Get(ctx context.Context, tenantID googleUuid.UUID, operation string) (*cryptoutilAppsJoseJaDomain.AuditConfig, error) {
	var config cryptoutilAppsJoseJaDomain.AuditConfig
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND operation = ?", tenantID.String(), operation).
		First(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit config: %w", err)
	}

	return &config, nil
}

// GetAllForTenant retrieves all audit configs for a tenant.
func (r *gormAuditConfigRepository) GetAllForTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsJoseJaDomain.AuditConfig, error) {
	var configs []*cryptoutilAppsJoseJaDomain.AuditConfig
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit configs for tenant: %w", err)
	}

	return configs, nil
}

// Upsert creates or updates audit config for a tenant and operation.
func (r *gormAuditConfigRepository) Upsert(ctx context.Context, config *cryptoutilAppsJoseJaDomain.AuditConfig) error {
	// Use Save which does upsert for GORM.
	if err := r.db.WithContext(ctx).Save(config).Error; err != nil {
		return fmt.Errorf("failed to upsert audit config: %w", err)
	}

	return nil
}

// Delete removes audit config for a tenant and operation.
func (r *gormAuditConfigRepository) Delete(ctx context.Context, tenantID googleUuid.UUID, operation string) error {
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND operation = ?", tenantID.String(), operation).
		Delete(&cryptoutilAppsJoseJaDomain.AuditConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete audit config: %w", err)
	}

	return nil
}

// ShouldAudit checks if an operation should be audited based on sampling.
func (r *gormAuditConfigRepository) ShouldAudit(ctx context.Context, tenantID googleUuid.UUID, operation string) (bool, error) {
	config, err := r.Get(ctx, tenantID, operation)
	if err != nil {
		// If record not found, default to auditing with fallback sampling rate.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//nolint:gosec // Not cryptographic - only for sampling decision.
			return rand.Float64() < cryptoutilSharedMagic.JoseJAAuditFallbackSamplingRate, nil
		}
		// For other errors, return the error.
		return false, fmt.Errorf("failed to check audit config: %w", err)
	}

	if !config.Enabled {
		return false, nil
	}

	//nolint:gosec // Not cryptographic - only for sampling decision.
	return rand.Float64() < config.SamplingRate, nil
}

// AuditLogRepository defines the interface for audit log persistence.
type AuditLogRepository interface {
	// Create stores a new audit log entry.
	Create(ctx context.Context, entry *cryptoutilAppsJoseJaDomain.AuditLogEntry) error

	// List retrieves audit log entries for a tenant with pagination.
	List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error)

	// ListByElasticJWK retrieves audit log entries for an Elastic JWK.
	ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error)

	// ListByOperation retrieves audit log entries by operation type.
	ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error)

	// GetByRequestID retrieves an audit log entry by request ID.
	GetByRequestID(ctx context.Context, requestID string) (*cryptoutilAppsJoseJaDomain.AuditLogEntry, error)

	// DeleteOlderThan removes audit log entries older than the specified time.
	DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error)
}

// gormAuditLogRepository implements AuditLogRepository using GORM.
type gormAuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new GORM-based audit log repository.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &gormAuditLogRepository{db: db}
}

// Create stores a new audit log entry.
func (r *gormAuditLogRepository) Create(ctx context.Context, entry *cryptoutilAppsJoseJaDomain.AuditLogEntry) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create audit log entry: %w", err)
	}

	return nil
}

// List retrieves audit log entries for a tenant with pagination.
func (r *gormAuditLogRepository) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error) {
	var (
		entries []*cryptoutilAppsJoseJaDomain.AuditLogEntry
		total   int64
	)

	// Count total.

	if err := r.db.WithContext(ctx).
		Model(&cryptoutilAppsJoseJaDomain.AuditLogEntry{}).
		Where("tenant_id = ?", tenantID.String()).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit log entries: %w", err)
	}

	// Fetch page.
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list audit log entries: %w", err)
	}

	return entries, total, nil
}

// ListByElasticJWK retrieves audit log entries for an Elastic JWK.
func (r *gormAuditLogRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error) {
	var (
		entries []*cryptoutilAppsJoseJaDomain.AuditLogEntry
		total   int64
	)

	// Count total.

	if err := r.db.WithContext(ctx).
		Model(&cryptoutilAppsJoseJaDomain.AuditLogEntry{}).
		Where("elastic_jwk_id = ?", elasticJWKID.String()).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit log entries: %w", err)
	}

	// Fetch page.
	if err := r.db.WithContext(ctx).
		Where("elastic_jwk_id = ?", elasticJWKID.String()).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list audit log entries: %w", err)
	}

	return entries, total, nil
}

// ListByOperation retrieves audit log entries by operation type.
func (r *gormAuditLogRepository) ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error) {
	var (
		entries []*cryptoutilAppsJoseJaDomain.AuditLogEntry
		total   int64
	)

	// Count total.

	if err := r.db.WithContext(ctx).
		Model(&cryptoutilAppsJoseJaDomain.AuditLogEntry{}).
		Where("tenant_id = ? AND operation = ?", tenantID.String(), operation).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit log entries: %w", err)
	}

	// Fetch page.
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND operation = ?", tenantID.String(), operation).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list audit log entries: %w", err)
	}

	return entries, total, nil
}

// GetByRequestID retrieves an audit log entry by request ID.
func (r *gormAuditLogRepository) GetByRequestID(ctx context.Context, requestID string) (*cryptoutilAppsJoseJaDomain.AuditLogEntry, error) {
	var entry cryptoutilAppsJoseJaDomain.AuditLogEntry
	if err := r.db.WithContext(ctx).
		Where("request_id = ?", requestID).
		First(&entry).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit log entry by request ID: %w", err)
	}

	return &entry, nil
}

// DeleteOlderThan removes audit log entries older than the specified time.
func (r *gormAuditLogRepository) DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at < datetime('now', ?)", tenantID.String(), fmt.Sprintf("-%d days", days)).
		Delete(&cryptoutilAppsJoseJaDomain.AuditLogEntry{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old audit log entries: %w", result.Error)
	}

	return result.RowsAffected, nil
}
