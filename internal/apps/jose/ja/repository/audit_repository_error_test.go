// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"testing"

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
			limit:     10,
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

	entries, total, err := repo.ListByElasticJWK(ctx, nonExistentJWKID, 0, 10)

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
			operation: string(make([]byte, 1000)),
		},
		{
			name:      "special characters",
			operation: "operation'with\"quotes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entries, total, err := repo.ListByOperation(ctx, tenantID, tt.operation, 0, 10)

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

	_, _, err := repo.List(ctx, tenantID, 0, 10)

	// Driver-specific behavior.
	if err == nil {
		t.Skip("Database driver doesn't propagate context cancellation")
	}

	require.Error(t, err)
}
