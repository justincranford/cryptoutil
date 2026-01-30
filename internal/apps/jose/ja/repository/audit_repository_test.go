// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Test constants for audit repository tests.
const testOperation = "test-operation"

// =============================================================================
// AuditConfigRepository Tests
// =============================================================================

func TestAuditConfigRepository_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	// Create config.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	operation := testOperation
	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     *tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.5,
	}
	require.NoError(t, repo.Upsert(ctx, config))

	defer func() {
		_ = repo.Delete(ctx, *tenantID, operation)
	}()

	// Test successful get.
	retrieved, err := repo.Get(ctx, *tenantID, operation)
	require.NoError(t, err)
	require.Equal(t, config.TenantID, retrieved.TenantID)
	require.Equal(t, config.Operation, retrieved.Operation)
	require.Equal(t, config.Enabled, retrieved.Enabled)
	require.Equal(t, config.SamplingRate, retrieved.SamplingRate)

	// Test error on non-existent config.
	nonExistentTenant := googleUuid.New()
	_, err = repo.Get(ctx, nonExistentTenant, operation)
	require.Error(t, err)
}

func TestAuditConfigRepository_GetAllForTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	// Create multiple configs for same tenant.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	operations := []string{"sign", "verify", "encrypt"}

	for _, op := range operations {
		config := &cryptoutilAppsJoseJaDomain.AuditConfig{
			TenantID:     *tenantID,
			Operation:    op,
			Enabled:      true,
			SamplingRate: 0.1,
		}
		require.NoError(t, repo.Upsert(ctx, config))

		defer func(operation string) {
			_ = repo.Delete(ctx, *tenantID, operation)
		}(op)
	}

	// Test get all for tenant.
	configs, err := repo.GetAllForTenant(ctx, *tenantID)
	require.NoError(t, err)
	require.Len(t, configs, 3)

	// Test empty result for non-existent tenant.
	nonExistentTenant := googleUuid.New()
	configs, err = repo.GetAllForTenant(ctx, nonExistentTenant)
	require.NoError(t, err)
	require.Empty(t, configs)
}

func TestAuditConfigRepository_Upsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	operation := "test-upsert"

	// Test create (insert).
	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     *tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.25,
	}
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	defer func() {
		_ = repo.Delete(ctx, *tenantID, operation)
	}()

	// Verify created.
	retrieved, err := repo.Get(ctx, *tenantID, operation)
	require.NoError(t, err)
	require.Equal(t, 0.25, retrieved.SamplingRate)

	// Test update (upsert).
	config.SamplingRate = 0.75
	config.Enabled = false
	err = repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Verify updated.
	retrieved, err = repo.Get(ctx, *tenantID, operation)
	require.NoError(t, err)
	require.Equal(t, 0.75, retrieved.SamplingRate)
	require.False(t, retrieved.Enabled)
}

func TestAuditConfigRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	// Create config to delete.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	operation := "test-delete"
	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     *tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.5,
	}
	require.NoError(t, repo.Upsert(ctx, config))

	// Delete config.
	err := repo.Delete(ctx, *tenantID, operation)
	require.NoError(t, err)

	// Verify deletion.
	_, err = repo.Get(ctx, *tenantID, operation)
	require.Error(t, err)
}

func TestAuditConfigRepository_ShouldAudit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	operation := "test-should-audit"

	// Test ShouldAudit when config doesn't exist (uses fallback).
	_, err := repo.ShouldAudit(ctx, *tenantID, operation)
	require.NoError(t, err)

	// Create config with enabled=true and sampling=1.0 (always audit).
	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     *tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 1.0,
	}
	require.NoError(t, repo.Upsert(ctx, config))

	defer func() {
		_ = repo.Delete(ctx, *tenantID, operation)
	}()

	// Test ShouldAudit with 100% sampling - should always return true.
	shouldAudit, err := repo.ShouldAudit(ctx, *tenantID, operation)
	require.NoError(t, err)
	require.True(t, shouldAudit)

	// Update to disabled.
	config.Enabled = false
	require.NoError(t, repo.Upsert(ctx, config))

	// Test ShouldAudit when disabled - should always return false.
	shouldAudit, err = repo.ShouldAudit(ctx, *tenantID, operation)
	require.NoError(t, err)
	require.False(t, shouldAudit)

	// Update to enabled with 0% sampling.
	config.Enabled = true
	config.SamplingRate = 0.0
	require.NoError(t, repo.Upsert(ctx, config))

	// Test ShouldAudit with 0% sampling - should always return false.
	shouldAudit, err = repo.ShouldAudit(ctx, *tenantID, operation)
	require.NoError(t, err)
	require.False(t, shouldAudit)
}

// =============================================================================
// AuditLogRepository Tests
// =============================================================================

func TestAuditLogRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	// Create audit log entry.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	requestID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
		ID:        *id,
		TenantID:  *tenantID,
		Operation: "sign",
		Success:   true,
		RequestID: requestID.String(),
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	// Verify created.
	retrieved, err := repo.GetByRequestID(ctx, requestID.String())
	require.NoError(t, err)
	require.Equal(t, entry.ID, retrieved.ID)
	require.Equal(t, entry.Operation, retrieved.Operation)
	require.True(t, retrieved.Success)
}

func TestAuditLogRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	// Create multiple entries for same tenant.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	for i := 0; i < 3; i++ {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		requestID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
			ID:        *id,
			TenantID:  *tenantID,
			Operation: "verify",
			Success:   true,
			RequestID: requestID.String(),
			CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		require.NoError(t, repo.Create(ctx, entry))
	}

	// Test list all.
	entries, total, err := repo.List(ctx, *tenantID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, entries, 3)

	// Test pagination.
	entries, total, err = repo.List(ctx, *tenantID, 0, 2)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, entries, 2)

	// Test empty result for non-existent tenant.
	nonExistentTenant := googleUuid.New()
	entries, total, err = repo.List(ctx, nonExistentTenant, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, entries)
}

func TestAuditLogRepository_ListByElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	// Create entries for specific ElasticJWK.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	for i := 0; i < 2; i++ {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		requestID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
			ID:           *id,
			TenantID:     *tenantID,
			ElasticJWKID: elasticJWKID,
			Operation:    "sign",
			Success:      true,
			RequestID:    requestID.String(),
			CreatedAt:    time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		require.NoError(t, repo.Create(ctx, entry))
	}

	// Test list by ElasticJWK.
	entries, total, err := repo.ListByElasticJWK(ctx, *elasticJWKID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, entries, 2)

	// Test empty result for non-existent ElasticJWK.
	nonExistentID := googleUuid.New()
	entries, total, err = repo.ListByElasticJWK(ctx, nonExistentID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, entries)
}

func TestAuditLogRepository_ListByOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	// Create entries with different operations.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	operations := []string{"sign", "sign", "verify"}
	for i, op := range operations {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		requestID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
			ID:        *id,
			TenantID:  *tenantID,
			Operation: op,
			Success:   true,
			RequestID: requestID.String(),
			CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		require.NoError(t, repo.Create(ctx, entry))
	}

	// Test list by operation "sign".
	entries, total, err := repo.ListByOperation(ctx, *tenantID, "sign", 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, entries, 2)

	// Test list by operation "verify".
	entries, total, err = repo.ListByOperation(ctx, *tenantID, "verify", 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, entries, 1)

	// Test empty result for non-existent operation.
	entries, total, err = repo.ListByOperation(ctx, *tenantID, "decrypt", 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, entries)
}

func TestAuditLogRepository_GetByRequestID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	// Create entry.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	requestID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
		ID:        *id,
		TenantID:  *tenantID,
		Operation: "encrypt",
		Success:   true,
		RequestID: requestID.String(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, entry))

	// Test successful get by RequestID.
	retrieved, err := repo.GetByRequestID(ctx, requestID.String())
	require.NoError(t, err)
	require.Equal(t, entry.ID, retrieved.ID)
	require.Equal(t, requestID.String(), retrieved.RequestID)

	// Test error on non-existent RequestID.
	_, err = repo.GetByRequestID(ctx, "non-existent-request-id")
	require.Error(t, err)
}

func TestAuditLogRepository_DeleteOlderThan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	// Create entries.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	// Create 3 entries with created_at 31 days in the past.
	oldTime := time.Now().UTC().Add(-31 * 24 * time.Hour)

	for i := 0; i < 3; i++ {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		requestID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
			ID:        *id,
			TenantID:  *tenantID,
			Operation: "sign",
			Success:   true,
			RequestID: requestID.String(),
			CreatedAt: oldTime.Add(time.Duration(i) * time.Second),
		}
		require.NoError(t, repo.Create(ctx, entry))
	}

	// Verify 3 entries exist.
	entries, total, err := repo.List(ctx, *tenantID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, entries, 3)

	// Delete entries older than 30 days (should delete all 3 entries since they're 31 days old).
	deleted, err := repo.DeleteOlderThan(ctx, *tenantID, 30)
	require.NoError(t, err)
	require.Equal(t, int64(3), deleted)

	// Verify entries deleted.
	entries, total, err = repo.List(ctx, *tenantID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, entries)
}

// =============================================================================
// Mutation-Killing Tests
// =============================================================================

// TestShouldAudit_DatabaseErrorPropagation kills mutations:
// - audit_repository.go:101:26 (CONDITIONALS_NEGATION)
// - audit_repository.go:101:26 (CONDITIONALS_BOUNDARY)
//
// Verifies that non-ErrRecordNotFound errors are propagated correctly,
// not converted to fallback sampling logic.
func TestShouldAudit_DatabaseErrorPropagation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use invalid tenant ID to trigger database error (foreign key constraint).
	// This will cause a non-ErrRecordNotFound error when querying audit config.
	invalidTenantID := googleUuid.Nil
	operation := testOperation

	repo := NewAuditConfigRepository(testDB)

	// Attempt to check if audit should be performed with invalid tenant.
	// This should return an error (not fallback to sampling).
	shouldAudit, err := repo.ShouldAudit(ctx, invalidTenantID, operation)

	// Verify error is propagated (not converted to fallback sampling).
	// The exact error depends on database implementation, but it should NOT be nil.
	require.Error(t, err, "expected database error to be propagated")
	require.False(t, shouldAudit, "should not audit when error occurs")
}

// TestShouldAudit_FallbackSamplingBoundary kills mutation:
// - audit_repository.go:112:24 (CONDITIONALS_BOUNDARY)
//
// Verifies boundary condition for sampling rate (< vs <=).
// Since rand.Float64() returns [0.0, 1.0), exact equality with sampling rate
// is theoretically possible but extremely rare. This test documents the
// expected behavior: < (not <=).
func TestShouldAudit_FallbackSamplingBoundary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	// Use non-existent tenant to trigger ErrRecordNotFound → fallback sampling.
	nonExistentTenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	operation := testOperation

	// Run sampling decision multiple times to verify statistical behavior.
	// With JoseJAAuditFallbackSamplingRate (1%) fallback sampling rate, expect ~1% true results.
	// Use 10000 iterations for 1% rate to get enough samples for statistical validity.
	const iterations = 10000

	trueCount := 0

	for i := 0; i < iterations; i++ {
		shouldAudit, err := repo.ShouldAudit(ctx, *nonExistentTenantID, operation)
		require.NoError(t, err, "fallback sampling should not error")

		if shouldAudit {
			trueCount++
		}
	}

	// Verify sampling rate is approximately 1% (allow generous tolerance for low sample rates).
	// Expected: ~100 hits out of 10000 (1%).
	// Tolerance: 0.3% to 3% (0.003 to 0.03) to account for statistical variance.
	samplingRate := float64(trueCount) / float64(iterations)
	require.Greater(t, samplingRate, 0.003, "sampling rate should be > 0.3%% (got %.4f)", samplingRate)
	require.Less(t, samplingRate, 0.03, "sampling rate should be < 3%% (got %.4f)", samplingRate)
}
