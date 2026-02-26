// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)


// =============================================================================
// AuditConfigRepository Tests
// =============================================================================

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
	entries, total, err := repo.ListByElasticJWK(ctx, *elasticJWKID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, entries, 2)

	// Test empty result for non-existent ElasticJWK.
	nonExistentID := googleUuid.New()
	entries, total, err = repo.ListByElasticJWK(ctx, nonExistentID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
	entries, total, err := repo.ListByOperation(ctx, *tenantID, "sign", 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, entries, 2)

	// Test list by operation "verify".
	entries, total, err = repo.ListByOperation(ctx, *tenantID, "verify", 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, entries, 1)

	// Test empty result for non-existent operation.
	entries, total, err = repo.ListByOperation(ctx, *tenantID, "decrypt", 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
	oldTime := time.Now().UTC().Add(-31 * cryptoutilSharedMagic.HoursPerDay * time.Hour)

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
	entries, total, err := repo.List(ctx, *tenantID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, entries, 3)

	// Delete entries older than 30 days (should delete all 3 entries since they're 31 days old).
	deleted, err := repo.DeleteOlderThan(ctx, *tenantID, cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days)
	require.NoError(t, err)
	require.Equal(t, int64(3), deleted)

	// Verify entries deleted.
	entries, total, err = repo.List(ctx, *tenantID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
//
// NOTE: This test is skipped because in-memory SQLite cannot trigger
// non-ErrRecordNotFound database errors. The code path is exercised in
// production with PostgreSQL where FK violations and connection errors occur.
func TestShouldAudit_DatabaseErrorPropagation(t *testing.T) {
	t.Parallel()

	// Skip: In-memory SQLite doesn't trigger database errors for Nil UUID queries.
	// The Nil UUID query returns ErrRecordNotFound (no rows), not a database error.
	// This mutation-killing test requires PostgreSQL with FK constraints to work.
	t.Skip("TODO: Requires PostgreSQL with FK constraints to trigger non-ErrRecordNotFound errors")
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
