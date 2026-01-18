// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"

	"cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
)

// AuditConfigRepository manages per-tenant audit configuration settings.
type AuditConfigRepository interface {
	// Get retrieves audit config for a tenant and operation.
	Get(ctx context.Context, tenantID googleUuid.UUID, operation string) (*domain.AuditConfig, error)

	// GetAll retrieves all audit configs for a tenant.
	GetAll(ctx context.Context, tenantID googleUuid.UUID) ([]domain.AuditConfig, error)

	// Upsert creates or updates audit config for a tenant and operation.
	Upsert(ctx context.Context, config *domain.AuditConfig) error

	// Delete removes audit config for a tenant and operation.
	Delete(ctx context.Context, tenantID googleUuid.UUID, operation string) error

	// IsEnabled checks if audit is enabled for a tenant and operation, and returns the sampling rate.
	IsEnabled(ctx context.Context, tenantID googleUuid.UUID, operation string) (enabled bool, samplingRate float64, err error)
}
