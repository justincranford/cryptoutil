// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

func TestMaterialJWKRepository_RotateMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	newMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:           *id,
		ElasticJWKID: *elasticJWKID,
		MaterialKID:  "new-material",
		Active:       true,
	}

	err = repo.RotateMaterial(ctx, *elasticJWKID, newMaterial)
	require.Error(t, err)
	// Transaction or any step could fail.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "sql: database is closed"),
		"Expected database error, got: %v", err)
}

// TestMaterialJWKRepository_RotateMaterialCreateError tests the "failed to create new material" error path
// inside RotateMaterial by using a duplicate MaterialKID to cause a constraint violation.
func TestMaterialJWKRepository_RotateMaterialCreateError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)
	elasticRepo := NewElasticJWKRepository(testDB)

	// Create unique test data.
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	materialKID := googleUuid.NewString() // Use UUID for uniqueness.

	// First create an ElasticJWK to satisfy foreign key constraint.
	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *elasticJWKID,
		TenantID:     *tenantID,
		KID:          googleUuid.NewString(),
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
	}
	err := elasticRepo.Create(ctx, elasticJWK)
	require.NoError(t, err)

	// Create first material with a specific MaterialKID.
	firstMaterialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	firstMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:            *firstMaterialID,
		ElasticJWKID:  *elasticJWKID,
		MaterialKID:   materialKID, // This KID will be duplicated.
		PrivateJWKJWE: "encrypted-private-1",
		PublicJWKJWE:  "encrypted-public-1",
		Active:        false,
	}
	err = repo.Create(ctx, firstMaterial)
	require.NoError(t, err)

	// Now try to rotate with a NEW material that uses the SAME MaterialKID (duplicate).
	secondMaterialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	duplicateMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:            *secondMaterialID,
		ElasticJWKID:  *elasticJWKID,
		MaterialKID:   materialKID, // DUPLICATE - should cause UNIQUE constraint violation.
		PrivateJWKJWE: "encrypted-private-2",
		PublicJWKJWE:  "encrypted-public-2",
		Active:        true,
	}

	// This should fail on the "Create" inside the transaction due to duplicate MaterialKID.
	err = repo.RotateMaterial(ctx, *elasticJWKID, duplicateMaterial)
	require.Error(t, err)
	// Should hit the "failed to create new material" error path.
	require.True(t,
		strings.Contains(err.Error(), "failed to create new material") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed"),
		"Expected create material error, got: %v", err)
}

func TestMaterialJWKRepository_RetireMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	err = repo.RetireMaterial(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to retire material JWK"))
}

func TestMaterialJWKRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	err = repo.Delete(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete material JWK"))
}

func TestMaterialJWKRepository_CountMaterialsDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err = repo.CountMaterials(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to count material JWKs"))
}

// ====================
// AuditConfig Repository Database Error Tests
// ====================

func TestAuditConfigRepository_GetDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err = repo.Get(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit config"))
}

func TestAuditConfigRepository_GetAllForTenantDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err = repo.GetAllForTenant(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit configs for tenant"))
}

func TestAuditConfigRepository_UpsertDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     *tenantID,
		Operation:    cryptoutilAppsJoseJaDomain.OperationSign,
		Enabled:      true,
		SamplingRate: cryptoutilSharedMagic.Tolerance50Percent,
	}

	err = repo.Upsert(ctx, config)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to upsert audit config"))
}

func TestAuditConfigRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	err = repo.Delete(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete audit config"))
}

func TestAuditConfigRepository_ShouldAuditDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err = repo.ShouldAudit(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to") ||
		strings.Contains(err.Error(), "sql: database is closed"))
}

// ====================
// AuditLog Repository Database Error Tests
// ====================
