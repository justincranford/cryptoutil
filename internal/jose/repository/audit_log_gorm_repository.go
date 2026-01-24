// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"cryptoutil/internal/jose/domain"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// Verify interface compliance at compile time.
var _ AuditLogRepository = (*AuditLogGormRepository)(nil)

// AuditLogGormRepository implements AuditLogRepository using GORM.
type AuditLogGormRepository struct {
	db *gorm.DB
}

// NewAuditLogGormRepository creates a new AuditLogGormRepository.
func NewAuditLogGormRepository(db *gorm.DB) *AuditLogGormRepository {
	return &AuditLogGormRepository{db: db}
}

// Create creates a new audit log entry.
func (r *AuditLogGormRepository) Create(ctx context.Context, entry *domain.AuditLogEntry) error {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	// Generate ID if not set.
	if entry.ID == googleUuid.Nil {
		entry.ID = googleUuid.New()
	}

	err := db.WithContext(ctx).Create(entry).Error
	if err != nil {
		return fmt.Errorf("failed to create audit log entry: %w", err)
	}

	return nil
}

// CreateWithSampling creates an audit log entry only if sampling check passes.
// Returns true if entry was created, false if skipped due to sampling.
func (r *AuditLogGormRepository) CreateWithSampling(ctx context.Context, entry *domain.AuditLogEntry, samplingRate float64) (bool, error) {
	// Check sampling rate (0.0 means never log, 1.0 means always log).
	if samplingRate <= 0.0 {
		return false, nil
	}

	if samplingRate < 1.0 {
		//nolint:gosec // math/rand is acceptable for sampling (not security-critical).
		if rand.Float64() > samplingRate {
			return false, nil
		}
	}

	err := r.Create(ctx, entry)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetByID retrieves an audit log entry by ID.
func (r *AuditLogGormRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*domain.AuditLogEntry, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	var entry domain.AuditLogEntry

	err := db.WithContext(ctx).Where("id = ?", id.String()).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("audit log entry not found: %s", id)
		}

		return nil, fmt.Errorf("failed to get audit log entry: %w", err)
	}

	return &entry, nil
}

// ListByTenantRealm retrieves audit log entries for a tenant/realm with pagination.
func (r *AuditLogGormRepository) ListByTenantRealm(ctx context.Context, tenantID, realmID googleUuid.UUID, offset, limit int) ([]domain.AuditLogEntry, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	var entries []domain.AuditLogEntry

	err := db.WithContext(ctx).
		Where("tenant_id = ? AND realm_id = ?", tenantID.String(), realmID.String()).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list audit log entries: %w", err)
	}

	return entries, nil
}

// ListByOperation retrieves audit log entries for a specific operation with pagination.
func (r *AuditLogGormRepository) ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]domain.AuditLogEntry, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	var entries []domain.AuditLogEntry

	err := db.WithContext(ctx).
		Where("tenant_id = ? AND operation = ?", tenantID.String(), operation).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list audit log entries by operation: %w", err)
	}

	return entries, nil
}

// ListByResource retrieves audit log entries for a specific resource with pagination.
func (r *AuditLogGormRepository) ListByResource(ctx context.Context, resourceType, resourceID string, offset, limit int) ([]domain.AuditLogEntry, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	var entries []domain.AuditLogEntry

	err := db.WithContext(ctx).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list audit log entries by resource: %w", err)
	}

	return entries, nil
}

// ListByTimeRange retrieves audit log entries within a time range with pagination.
func (r *AuditLogGormRepository) ListByTimeRange(ctx context.Context, tenantID googleUuid.UUID, start, end time.Time, offset, limit int) ([]domain.AuditLogEntry, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	startMillis := start.UnixMilli()
	endMillis := end.UnixMilli()

	var entries []domain.AuditLogEntry

	err := db.WithContext(ctx).
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID.String(), startMillis, endMillis).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list audit log entries by time range: %w", err)
	}

	return entries, nil
}

// Count returns the total number of audit log entries for a tenant.
func (r *AuditLogGormRepository) Count(ctx context.Context, tenantID googleUuid.UUID) (int64, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	var count int64

	err := db.WithContext(ctx).Model(&domain.AuditLogEntry{}).Where("tenant_id = ?", tenantID.String()).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count audit log entries: %w", err)
	}

	return count, nil
}

// DeleteOlderThan deletes audit log entries older than the specified time.
// Returns the number of entries deleted.
func (r *AuditLogGormRepository) DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, before time.Time) (int64, error) {
	db := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db)

	beforeMillis := before.UnixMilli()

	result := db.WithContext(ctx).
		Where("tenant_id = ? AND created_at < ?", tenantID.String(), beforeMillis).
		Delete(&domain.AuditLogEntry{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old audit log entries: %w", result.Error)
	}

	return result.RowsAffected, nil
}
