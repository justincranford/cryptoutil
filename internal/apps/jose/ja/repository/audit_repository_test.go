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
		SamplingRate: cryptoutilSharedMagic.Tolerance50Percent,
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
			SamplingRate: cryptoutilSharedMagic.Tolerance10Percent,
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
		SamplingRate: cryptoutilSharedMagic.TestProbQuarter,
	}
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	defer func() {
		_ = repo.Delete(ctx, *tenantID, operation)
	}()

	// Verify created.
	retrieved, err := repo.Get(ctx, *tenantID, operation)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.TestProbQuarter, retrieved.SamplingRate)

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
		SamplingRate: cryptoutilSharedMagic.Tolerance50Percent,
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
		SamplingRate: cryptoutilSharedMagic.TestProbAlways,
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
	config.SamplingRate = cryptoutilSharedMagic.BaselineContributionZero
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
	entries, total, err := repo.List(ctx, *tenantID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
	entries, total, err = repo.List(ctx, nonExistentTenant, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, entries)
}

