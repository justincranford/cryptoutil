// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"time"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
)

// AuditLogRepository manages audit log entries for cryptographic operations.
type AuditLogRepository interface {
	// Create creates a new audit log entry (respects sampling).
	Create(ctx context.Context, entry *cryptoutilJoseDomain.AuditLogEntry) error

	// CreateWithSampling creates an audit log entry only if sampling check passes.
	// Returns true if entry was created, false if skipped due to sampling.
	CreateWithSampling(ctx context.Context, entry *cryptoutilJoseDomain.AuditLogEntry, samplingRate float64) (bool, error)

	// GetByID retrieves an audit log entry by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilJoseDomain.AuditLogEntry, error)

	// ListByTenantRealm retrieves audit log entries for a tenant/realm with pagination.
	ListByTenantRealm(ctx context.Context, tenantID, realmID googleUuid.UUID, offset, limit int) ([]cryptoutilJoseDomain.AuditLogEntry, error)

	// ListByOperation retrieves audit log entries for a specific operation with pagination.
	ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]cryptoutilJoseDomain.AuditLogEntry, error)

	// ListByResource retrieves audit log entries for a specific resource with pagination.
	ListByResource(ctx context.Context, resourceType, resourceID string, offset, limit int) ([]cryptoutilJoseDomain.AuditLogEntry, error)

	// ListByTimeRange retrieves audit log entries within a time range with pagination.
	ListByTimeRange(ctx context.Context, tenantID googleUuid.UUID, start, end time.Time, offset, limit int) ([]cryptoutilJoseDomain.AuditLogEntry, error)

	// Count returns the total number of audit log entries for a tenant.
	Count(ctx context.Context, tenantID googleUuid.UUID) (int64, error)

	// DeleteOlderThan deletes audit log entries older than the specified time.
	// Returns the number of entries deleted.
	DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, before time.Time) (int64, error)
}
