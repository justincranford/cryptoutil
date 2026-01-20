// Copyright (c) 2025 Justin Cranford
//

// Package service provides business logic services for jose-ja.
package service

import (
	"context"
	"fmt"
	"time"

	joseJADomain "cryptoutil/internal/apps/jose/ja/domain"
	joseJARepository "cryptoutil/internal/apps/jose/ja/repository"

	googleUuid "github.com/google/uuid"
)

// AuditLogService provides business logic for audit logging operations.
type AuditLogService interface {
	// LogOperation logs an operation for audit purposes.
	LogOperation(ctx context.Context, tenantID googleUuid.UUID, elasticJWKID *googleUuid.UUID, operation, requestID string, success bool, errorMessage *string) error

	// ListAuditLogs lists audit logs for a tenant with pagination.
	ListAuditLogs(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error)

	// ListAuditLogsByElasticJWK lists audit logs for a specific elastic JWK.
	ListAuditLogsByElasticJWK(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error)

	// ListAuditLogsByOperation lists audit logs by operation type.
	ListAuditLogsByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error)

	// GetAuditConfig gets the audit configuration for a tenant.
	GetAuditConfig(ctx context.Context, tenantID googleUuid.UUID) (*joseJADomain.AuditConfig, error)

	// UpdateAuditConfig updates the audit configuration for a tenant.
	UpdateAuditConfig(ctx context.Context, tenantID googleUuid.UUID, config *joseJADomain.AuditConfig) error

	// CleanupOldLogs removes audit logs older than the specified number of days.
	CleanupOldLogs(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error)
}

// auditLogServiceImpl implements AuditLogService.
type auditLogServiceImpl struct {
	auditLogRepo    joseJARepository.AuditLogRepository
	auditConfigRepo joseJARepository.AuditConfigRepository
	elasticRepo     joseJARepository.ElasticJWKRepository
}

// NewAuditLogService creates a new AuditLogService.
func NewAuditLogService(
	auditLogRepo joseJARepository.AuditLogRepository,
	auditConfigRepo joseJARepository.AuditConfigRepository,
	elasticRepo joseJARepository.ElasticJWKRepository,
) AuditLogService {
	return &auditLogServiceImpl{
		auditLogRepo:    auditLogRepo,
		auditConfigRepo: auditConfigRepo,
		elasticRepo:     elasticRepo,
	}
}

// LogOperation logs an operation for audit purposes.
func (s *auditLogServiceImpl) LogOperation(ctx context.Context, tenantID googleUuid.UUID, elasticJWKID *googleUuid.UUID, operation, requestID string, success bool, errorMessage *string) error {
	// Check if audit is enabled for this operation.
	shouldAudit, err := s.auditConfigRepo.ShouldAudit(ctx, tenantID, operation)
	if err != nil {
		// On error, default to auditing.
		shouldAudit = true
	}

	if !shouldAudit {
		return nil // Audit disabled for this operation.
	}

	// Create audit log entry.
	entry := &joseJADomain.AuditLogEntry{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		ElasticJWKID: elasticJWKID,
		Operation:    operation,
		Success:      success,
		ErrorMessage: errorMessage,
		RequestID:    requestID,
		CreatedAt:    time.Now(),
	}

	// Store audit log.
	if err := s.auditLogRepo.Create(ctx, entry); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// ListAuditLogs lists audit logs for a tenant with pagination.
func (s *auditLogServiceImpl) ListAuditLogs(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error) {
	entries, total, err := s.auditLogRepo.List(ctx, tenantID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return entries, total, nil
}

// ListAuditLogsByElasticJWK lists audit logs for a specific elastic JWK.
func (s *auditLogServiceImpl) ListAuditLogsByElasticJWK(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error) {
	// Verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return nil, 0, fmt.Errorf("elastic JWK not found")
	}

	entries, total, err := s.auditLogRepo.ListByElasticJWK(ctx, elasticJWKID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return entries, total, nil
}

// ListAuditLogsByOperation lists audit logs by operation type.
func (s *auditLogServiceImpl) ListAuditLogsByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error) {
	entries, total, err := s.auditLogRepo.ListByOperation(ctx, tenantID, operation, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs by operation: %w", err)
	}

	return entries, total, nil
}

// GetAuditConfig gets the audit configuration for a tenant.
func (s *auditLogServiceImpl) GetAuditConfig(ctx context.Context, tenantID googleUuid.UUID) (*joseJADomain.AuditConfig, error) {
	configs, err := s.auditConfigRepo.GetAllForTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit configs: %w", err)
	}

	// Return first config or create default.
	if len(configs) > 0 {
		return configs[0], nil
	}

	// Return a default config.
	return &joseJADomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    joseJADomain.OperationGenerate, // Default operation.
		Enabled:      true,
		SamplingRate: 1.0, // 100% sampling by default.
	}, nil
}

// UpdateAuditConfig updates the audit configuration for a tenant.
func (s *auditLogServiceImpl) UpdateAuditConfig(ctx context.Context, tenantID googleUuid.UUID, config *joseJADomain.AuditConfig) error {
	// Ensure tenant ID is set.
	config.TenantID = tenantID

	if err := s.auditConfigRepo.Upsert(ctx, config); err != nil {
		return fmt.Errorf("failed to update audit config: %w", err)
	}

	return nil
}

// CleanupOldLogs removes audit logs older than the specified number of days.
func (s *auditLogServiceImpl) CleanupOldLogs(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error) {
	count, err := s.auditLogRepo.DeleteOlderThan(ctx, tenantID, days)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	return count, nil
}
