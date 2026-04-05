// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// testNonExistentOperation is a test constant to satisfy goconst linter.
const testNonExistentOperation = "non_existent_operation"

// TestAuditConfigRepository_GetNonExistent tests getting non-existent audit config.
func TestAuditConfigRepository_GetNonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	nonExistentTenant := googleUuid.New()
	operation := testNonExistentOperation

	_, err := repo.Get(ctx, nonExistentTenant, operation)

	// Should error (record not found).
	require.Error(t, err)
}

// TestAuditConfigRepository_DeleteNonExistent tests deleting non-existent config.
func TestAuditConfigRepository_DeleteNonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	nonExistentTenant := googleUuid.New()
	operation := testNonExistentOperation

	// GORM Delete with WHERE clause that matches nothing.
	// No error, zero rows affected.
	err := repo.Delete(ctx, nonExistentTenant, operation)

	require.NoError(t, err)
}

// TestAuditConfigRepository_ShouldAuditDefaultBehavior tests default audit behavior.
func TestAuditConfigRepository_ShouldAuditDefaultBehavior(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	tenantID := googleUuid.New()
	operation := "unconfigured_operation"

	// When no config exists, should NOT error - implementation uses sampling (1% by default).
	// We don't verify the return value because it's probabilistic.
	shouldAudit, err := repo.ShouldAudit(ctx, tenantID, operation)
	require.NoError(t, err)
	// shouldAudit is probabilistic based on JoseJAAuditFallbackSamplingRate (1%)
	// Just verify type is correct - don't assert value.
	_ = shouldAudit
}

// TestAuditLogRepository_CreateWithInvalidElasticJWKID tests creation with invalid FK.
func TestAuditLogRepository_CreateWithInvalidElasticJWKID(t *testing.T) {
	t.Parallel()

	// This test would require FK constraints in schema.
	t.Skip("TODO P2.4: Add FK constraint tests when schema includes foreign keys")
}

// TestAuditLogRepository_ListPaginationBoundary tests pagination edge cases.
func TestAuditLogRepository_ListPaginationBoundary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	tenantID := googleUuid.New()

	tests := []struct {
		name      string
		offset    int
		limit     int
		wantError bool
	}{
		{
			name:      "zero limit",
			offset:    0,
			limit:     0,
			wantError: false,
		},
		{
			name:      "negative offset",
			offset:    -1,
			limit:     cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			wantError: false,
		},
		{
			name:      "very large limit",
			offset:    0,
			limit:     1000000,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := repo.List(ctx, tenantID, tt.offset, tt.limit)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestAuditLogRepository_ListByElasticJWKNonExistent tests listing for non-existent JWK.
func TestAuditLogRepository_ListByElasticJWKNonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	nonExistentJWKID := googleUuid.New()

	entries, total, err := repo.ListByElasticJWK(ctx, nonExistentJWKID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, entries)
}

// TestAuditLogRepository_ListByOperationEdgeCases tests operation filter edge cases.
func TestAuditLogRepository_ListByOperationEdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	tenantID := googleUuid.New()

	tests := []struct {
		name      string
		operation string
	}{
		{
			name:      "empty operation",
			operation: "",
		},
		{
			name:      "very long operation name",
			operation: string(make([]byte, cryptoutilSharedMagic.JoseJADefaultListLimit)),
		},
		{
			name:      "special characters",
			operation: "operation'with\"quotes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entries, total, err := repo.ListByOperation(ctx, tenantID, tt.operation, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

			require.NoError(t, err)
			require.Equal(t, int64(0), total)
			require.Empty(t, entries)
		})
	}
}

// TestAuditLogRepository_GetByRequestIDNonExistent tests getting non-existent request ID.
func TestAuditLogRepository_GetByRequestIDNonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	nonExistentRequestID := "non-existent-request-id"

	_, err := repo.GetByRequestID(ctx, nonExistentRequestID)

	require.Error(t, err)
}

// TestAuditLogRepository_DeleteOlderThanEdgeCases tests deletion time boundary.
func TestAuditLogRepository_DeleteOlderThanEdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	tenantID := googleUuid.New()

	// Test with zero duration (delete everything).
	deleted, err := repo.DeleteOlderThan(ctx, tenantID, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, deleted, int64(0))

	// Test with negative duration (invalid but handled gracefully).
	deleted, err = repo.DeleteOlderThan(ctx, tenantID, -1)
	require.NoError(t, err)
	require.GreaterOrEqual(t, deleted, int64(0))
}

// TestAuditConfigRepository_ContextCancellation tests context cancellation behavior.
func TestAuditConfigRepository_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewAuditConfigRepository(testDB)
	tenantID := googleUuid.New()

	_, err := repo.GetAllForTenant(ctx, tenantID)

	// Driver-specific behavior.
	if err == nil {
		t.Skip("Database driver doesn't propagate context cancellation")
	}

	require.Error(t, err)
}

// TestAuditLogRepository_ContextCancellation tests context cancellation behavior.
func TestAuditLogRepository_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewAuditLogRepository(testDB)
	tenantID := googleUuid.New()

	_, _, err := repo.List(ctx, tenantID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

	// Driver-specific behavior.
	if err == nil {
		t.Skip("Database driver doesn't propagate context cancellation")
	}

	require.Error(t, err)
}

func TestAuditConfigRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	err := repo.Delete(ctx, googleUuid.New(), cryptoutilAppsJoseJaModel.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete audit config"))
}

func TestAuditConfigRepository_GetAllForTenantDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err := repo.GetAllForTenant(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit configs for tenant"))
}

func TestAuditConfigRepository_GetDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err := repo.Get(ctx, googleUuid.New(), cryptoutilAppsJoseJaModel.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit config"))
}

func TestAuditConfigRepository_ShouldAuditDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err := repo.ShouldAudit(ctx, googleUuid.New(), cryptoutilAppsJoseJaModel.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to") ||
		strings.Contains(err.Error(), "sql: database is closed"))
}

func TestAuditConfigRepository_UpsertDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	config := &cryptoutilAppsJoseJaModel.AuditConfig{
		TenantID:     *tenantID,
		Operation:    cryptoutilAppsJoseJaModel.OperationSign,
		Enabled:      true,
		SamplingRate: cryptoutilSharedMagic.Tolerance50Percent,
	}

	err := repo.Upsert(ctx, config)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to upsert audit config"))
}

func TestAuditLogRepository_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	entry := &cryptoutilAppsJoseJaModel.AuditLogEntry{
		ID:        *id,
		TenantID:  *tenantID,
		Operation: cryptoutilAppsJoseJaModel.OperationSign,
		Success:   true,
		RequestID: googleUuid.NewString(),
	}

	err := repo.Create(ctx, entry)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create audit log entry"))
}

func TestAuditLogRepository_DeleteOlderThanDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, err := repo.DeleteOlderThan(ctx, googleUuid.New(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete old audit log entries"))
}

func TestAuditLogRepository_GetByRequestIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, err := repo.GetByRequestID(ctx, googleUuid.NewString())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit log entry by request ID"))
}

func TestAuditLogRepository_ListByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, _, err := repo.ListByElasticJWK(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count audit log entries") ||
			strings.Contains(err.Error(), "failed to list audit log entries"),
		"Expected count or list error, got: %v", err)
}

func TestAuditLogRepository_ListByOperationDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, _, err := repo.ListByOperation(ctx, googleUuid.New(), cryptoutilAppsJoseJaModel.OperationSign, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count audit log entries") ||
			strings.Contains(err.Error(), "failed to list audit log entries"),
		"Expected count or list error, got: %v", err)
}

func TestAuditLogRepository_ListDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, _, err := repo.List(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count audit log entries") ||
			strings.Contains(err.Error(), "failed to list audit log entries"),
		"Expected count or list error, got: %v", err)
}
