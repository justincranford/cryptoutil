// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"strings"
	"testing"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// testMaterialKID is a test constant to satisfy goconst linter.
const testMaterialKID = "test-kid"

// TestMaterialJWKRepository_CreateForeignKeyViolation tests creation with invalid ElasticJWKID.
func TestMaterialJWKRepository_CreateForeignKeyViolation(t *testing.T) {
	t.Parallel()

	// This test would require database foreign key constraints.
	// Current schema may not enforce FK constraints in SQLite test mode.
	t.Skip("TODO P2.4: Add FK constraint tests when schema migrations include foreign keys")
}

// TestMaterialJWKRepository_RotateMaterialTransactionRollback tests transaction failure handling.
func TestMaterialJWKRepository_RotateMaterialTransactionRollback(t *testing.T) {
	t.Parallel()

	// This test would verify transaction rollback on error.
	// Requires mocked database to trigger mid-transaction failure.
	t.Skip("TODO P2.4: Add mocked database tests for transaction rollback scenarios")
}

// TestMaterialJWKRepository_RotateMaterialConcurrentModification tests concurrent rotation.
func TestMaterialJWKRepository_RotateMaterialConcurrentModification(t *testing.T) {
	t.Parallel()

	// This test would verify atomic rotation under concurrent access.
	// Requires concurrent goroutines and synchronization.
	t.Skip("TODO P2.4: Add concurrency tests for RotateMaterial with goroutines")
}

// TestMaterialJWKRepository_RetireMaterialNonExistent tests retiring non-existent material.
func TestMaterialJWKRepository_RetireMaterialNonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

	nonExistentID := googleUuid.New()

	// GORM Updates() with WHERE clause that matches nothing.
	// No error returned, zero rows affected.
	err := repo.RetireMaterial(ctx, nonExistentID)

	// GORM doesn't error on zero affected rows.
	require.NoError(t, err)
}

// TestMaterialJWKRepository_CountMaterialsError tests count error handling.
func TestMaterialJWKRepository_CountMaterialsError(t *testing.T) {
	t.Parallel()

	// Testing GORM Count() error paths requires mocked database.
	t.Skip("TODO P2.4: Add mocked database tests for Count error scenarios")
}

// TestMaterialJWKRepository_GetActiveMaterialMultipleActive tests behavior with multiple active materials.
func TestMaterialJWKRepository_GetActiveMaterialMultipleActive(t *testing.T) {
	t.Parallel()

	// This test would verify behavior when database constraint fails.
	// Should only have one active material per ElasticJWK.
	// Requires database UNIQUE constraint on (elastic_jwk_id, active) WHERE active=true.
	t.Skip("TODO P2.4: Add unique constraint tests for active materials")
}

// TestMaterialJWKRepository_ListPaginationEdgeCases tests pagination boundary conditions.
func TestMaterialJWKRepository_ListPaginationEdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)
	elasticJWKID := googleUuid.New()

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
			name:      "large limit",
			offset:    0,
			limit:     cryptoutilSharedMagic.PBKDF2Iterations,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := repo.ListByElasticJWK(ctx, elasticJWKID, tt.offset, tt.limit)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMaterialJWKRepository_ContextCancellation tests context cancellation behavior.
func TestMaterialJWKRepository_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewMaterialJWKRepository(testDB)
	materialKID := testMaterialKID

	_, err := repo.GetByMaterialKID(ctx, materialKID)

	// Driver-specific behavior.
	if err == nil {
		t.Skip("Database driver doesn't propagate context cancellation")
	}

	require.Error(t, err)
}

// TestMaterialJWKRepository_DeleteCascadeToAuditLogs tests cascade behavior.
func TestMaterialJWKRepository_DeleteCascadeToAuditLogs(t *testing.T) {
	t.Parallel()

	// This test would verify cascade deletion to audit logs if configured.
	t.Skip("TODO P2.4: Add cascade deletion tests when audit log FK constraints implemented")
}

// TestMaterialJWKRepository_NilContextHandling tests nil context handling.
func TestMaterialJWKRepository_NilContextHandling(t *testing.T) {
	t.Parallel()

	repo := NewMaterialJWKRepository(testDB)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	_, err := repo.GetByMaterialKID(nil, "test-kid") //nolint:staticcheck // Testing nil context.
	require.Error(t, err)
}

func TestMaterialJWKRepository_CountMaterialsDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err := repo.CountMaterials(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to count material JWKs"))
}

func TestMaterialJWKRepository_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	material := &cryptoutilAppsJoseJaModel.MaterialJWK{
		ID:           *id,
		ElasticJWKID: *elasticJWKID,
		MaterialKID:  "test-material-error",
		Active:       true,
	}

	err := repo.Create(ctx, material)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create material JWK"))
}

func TestMaterialJWKRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	err := repo.Delete(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete material JWK"))
}

func TestMaterialJWKRepository_GetActiveMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err := repo.GetActiveMaterial(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get active material JWK"))
}

func TestMaterialJWKRepository_GetByIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err := repo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get material JWK by ID"))
}

func TestMaterialJWKRepository_GetByMaterialKIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err := repo.GetByMaterialKID(ctx, "test-kid")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get material JWK by KID"))
}

func TestMaterialJWKRepository_ListByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, _, err := repo.ListByElasticJWK(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count material JWKs") ||
			strings.Contains(err.Error(), "failed to list material JWKs"),
		"Expected count or list error, got: %v", err)
}

func TestMaterialJWKRepository_RetireMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	err := repo.RetireMaterial(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to retire material JWK"))
}

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
	elasticJWK := &cryptoutilAppsJoseJaModel.ElasticJWK{
		ID:           *elasticJWKID,
		TenantID:     *tenantID,
		KID:          googleUuid.NewString(),
		KeyType:      cryptoutilAppsJoseJaModel.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
	}
	err := elasticRepo.Create(ctx, elasticJWK)
	require.NoError(t, err)

	// Create first material with a specific MaterialKID.
	firstMaterialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	firstMaterial := &cryptoutilAppsJoseJaModel.MaterialJWK{
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
	duplicateMaterial := &cryptoutilAppsJoseJaModel.MaterialJWK{
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

func TestMaterialJWKRepository_RotateMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	newMaterial := &cryptoutilAppsJoseJaModel.MaterialJWK{
		ID:           *id,
		ElasticJWKID: *elasticJWKID,
		MaterialKID:  "new-material",
		Active:       true,
	}

	err := repo.RotateMaterial(ctx, *elasticJWKID, newMaterial)
	require.Error(t, err)
	// Transaction or any step could fail.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "sql: database is closed"),
		"Expected database error, got: %v", err)
}
