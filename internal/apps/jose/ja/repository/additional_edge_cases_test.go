// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// TestElasticJWKRepository_GetByIDWithInvalidID tests GetByID with various invalid IDs.
func TestElasticJWKRepository_GetByIDWithInvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	tests := []struct {
		name string
		id   googleUuid.UUID
	}{
		{
			name: "all zeros UUID",
			id:   googleUuid.MustParse("00000000-0000-0000-0000-000000000000"),
		},
		{
			name: "all ones UUID",
			id:   googleUuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		},
		{
			name: "random non-existent UUID",
			id:   googleUuid.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := repo.GetByID(ctx, tt.id)
			require.Error(t, err, "GetByID should fail for non-existent ID")
		})
	}
}

// TestElasticJWKRepository_GetWithSpecialCharactersInKID tests Get with special characters in KID.
func TestElasticJWKRepository_GetWithSpecialCharactersInKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	tests := []struct {
		name string
		kid  string
	}{
		{
			name: "KID with dashes",
			kid:  "test-key-id-with-dashes",
		},
		{
			name: "KID with underscores",
			kid:  "test_key_id_with_underscores",
		},
		{
			name: "KID with periods",
			kid:  "test.key.id.with.periods",
		},
		{
			name: "KID with mixed special chars",
			kid:  "test-key_id.mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
			tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

			jwk := &cryptoutilAppsJoseJaDomain.ElasticJWK{
				ID:           *id,
				TenantID:     *tenantID,
				KID:          tt.kid + "-" + id.String()[:8], // Make unique per test run.
				KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
				Algorithm:    "RS256",
				Use:          "sig",
				MaxMaterials: 10,
				CreatedAt:    time.Now().UTC(),
			}

			// Create.
			err := repo.Create(ctx, jwk)
			require.NoError(t, err)

			// Cleanup.
			t.Cleanup(func() {
				_ = repo.Delete(ctx, jwk.ID)
			})

			// Retrieve using Get.
			retrieved, err := repo.Get(ctx, *tenantID, jwk.KID)
			require.NoError(t, err)
			require.Equal(t, jwk.KID, retrieved.KID)
		})
	}
}

// TestElasticJWKRepository_ListWithEmptyDatabase tests List when no JWKs exist.
func TestElasticJWKRepository_ListWithEmptyDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Use a unique tenant ID that has no JWKs.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwks, total, err := repo.List(ctx, *tenantID, 0, 100)
	require.NoError(t, err)
	require.Empty(t, jwks)
	require.Equal(t, int64(0), total)
}

// TestElasticJWKRepository_UpdateNonExistentJWK tests Update on non-existent JWK.
func TestElasticJWKRepository_UpdateNonExistentJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	nonExistentJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "non-existent-" + id.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now().UTC(),
	}

	// Attempting to update a non-existent JWK should succeed (GORM creates it).
	err := repo.Update(ctx, nonExistentJWK)
	require.NoError(t, err)

	// Cleanup.
	t.Cleanup(func() {
		_ = repo.Delete(ctx, nonExistentJWK.ID)
	})

	// Verify it was created.
	retrieved, err := repo.GetByID(ctx, nonExistentJWK.ID)
	require.NoError(t, err)
	require.Equal(t, nonExistentJWK.KID, retrieved.KID)
}

// TestElasticJWKRepository_DeleteAlreadyDeleted tests Delete on already deleted JWK.
func TestElasticJWKRepository_DeleteAlreadyDeleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwk := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-double-delete-" + id.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now().UTC(),
	}

	// Create.
	require.NoError(t, repo.Create(ctx, jwk))

	// Delete once.
	err := repo.Delete(ctx, jwk.ID)
	require.NoError(t, err)

	// Delete again (idempotent - should not error).
	err = repo.Delete(ctx, jwk.ID)
	require.NoError(t, err)
}

// TestMaterialJWKRepository_GetByIDEdgeCases tests GetByID with edge case UUIDs.
func TestMaterialJWKRepository_GetByIDEdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

	tests := []struct {
		name string
		id   googleUuid.UUID
	}{
		{
			name: "nil UUID pattern",
			id:   googleUuid.MustParse("00000000-0000-0000-0000-000000000000"),
		},
		{
			name: "max UUID pattern",
			id:   googleUuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := repo.GetByID(ctx, tt.id)
			require.Error(t, err, "GetByID should fail for non-existent ID")
		})
	}
}

// TestMaterialJWKRepository_GetByMaterialKIDWithSpecialChars tests GetByMaterialKID with special characters.
func TestMaterialJWKRepository_GetByMaterialKIDWithSpecialChars(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	materialRepo := NewMaterialJWKRepository(testDB)
	elasticRepo := NewElasticJWKRepository(testDB)

	// Create parent elastic JWK.
	elasticID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	parentJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *elasticID,
		TenantID:     *tenantID,
		KID:          "parent-" + elasticID.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now().UTC(),
	}
	require.NoError(t, elasticRepo.Create(ctx, parentJWK))

	t.Cleanup(func() {
		_ = elasticRepo.Delete(ctx, parentJWK.ID)
	})

	tests := []struct {
		name        string
		materialKID string
	}{
		{
			name:        "material KID with dashes",
			materialKID: "material-key-id-dashes",
		},
		{
			name:        "material KID with underscores",
			materialKID: "material_key_id_underscores",
		},
		{
			name:        "material KID with periods",
			materialKID: "material.key.id.periods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			materialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

			// Use sample encrypted JWK data.
			material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
				ID:            *materialID,
				ElasticJWKID:  parentJWK.ID,
				MaterialKID:   tt.materialKID + "-" + materialID.String()[:8], // Make unique.
				PrivateJWKJWE: "sample-jwe-private-" + materialID.String(),
				PublicJWKJWE:  "sample-jwe-public-" + materialID.String(),
				Active:        true,
				CreatedAt:     time.Now().UTC(),
			}

			// Create.
			err := materialRepo.Create(ctx, material)
			require.NoError(t, err)

			// Cleanup.
			t.Cleanup(func() {
				_ = materialRepo.Delete(ctx, material.ID)
			})

			// Retrieve.
			retrieved, err := materialRepo.GetByMaterialKID(ctx, material.MaterialKID)
			require.NoError(t, err)
			require.Equal(t, material.MaterialKID, retrieved.MaterialKID)
		})
	}
}

// TestMaterialJWKRepository_GetActiveMaterialWhenNoneActive tests GetActiveMaterial when no materials are active.
func TestMaterialJWKRepository_GetActiveMaterialWhenNoneActive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	materialRepo := NewMaterialJWKRepository(testDB)
	elasticRepo := NewElasticJWKRepository(testDB)

	// Create parent elastic JWK.
	elasticID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	parentJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *elasticID,
		TenantID:     *tenantID,
		KID:          "parent-no-active-" + elasticID.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now().UTC(),
	}
	require.NoError(t, elasticRepo.Create(ctx, parentJWK))

	defer func() {
		_ = elasticRepo.Delete(ctx, parentJWK.ID)
	}()

	// Create material but mark it as inactive.
	materialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:            *materialID,
		ElasticJWKID:  parentJWK.ID,
		MaterialKID:   "inactive-material-" + materialID.String()[:8],
		PrivateJWKJWE: "sample-jwe-private-" + materialID.String(),
		PublicJWKJWE:  "sample-jwe-public-" + materialID.String(),
		Active:        false, // Not active.
		CreatedAt:     time.Now().UTC(),
	}
	require.NoError(t, materialRepo.Create(ctx, material))

	defer func() {
		_ = materialRepo.Delete(ctx, material.ID)
	}()

	// GetActiveMaterial should return error when no active materials exist.
	_, err := materialRepo.GetActiveMaterial(ctx, parentJWK.ID)
	require.Error(t, err)
}

// TestAuditConfigRepository_UpsertMultipleTimes tests Upsert idempotency.
func TestAuditConfigRepository_UpsertMultipleTimes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditConfigRepository(testDB)

	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     *tenantID,
		Operation:    "test-upsert-multi",
		Enabled:      true,
		SamplingRate: 0.5,
	}

	// Upsert first time.
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Cleanup.
	defer func() {
		_ = repo.Delete(ctx, *tenantID, config.Operation)
	}()

	// Upsert again with different sampling rate.
	config.SamplingRate = 0.8

	err = repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Verify the update.
	retrieved, err := repo.Get(ctx, *tenantID, config.Operation)
	require.NoError(t, err)
	require.InDelta(t, 0.8, retrieved.SamplingRate, 0.01)

	// Upsert third time with disabled.
	config.Enabled = false

	err = repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Verify.
	retrieved, err = repo.Get(ctx, *tenantID, config.Operation)
	require.NoError(t, err)
	require.False(t, retrieved.Enabled)
}

// TestAuditLogRepository_CreateMultipleEntries tests creating many audit logs.
func TestAuditLogRepository_CreateMultipleEntries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewAuditLogRepository(testDB)

	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	const numEntries = 20

	// Create many audit log entries.
	for i := 0; i < numEntries; i++ {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
			ID:        *id,
			TenantID:  *tenantID,
			Operation: "bulk-test",
			Success:   true,
			RequestID: id.String(),
			CreatedAt: time.Now().UTC().Add(time.Duration(-i) * time.Minute), // Different timestamps.
		}

		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// Cleanup.
	t.Cleanup(func() {
		_, _ = repo.DeleteOlderThan(ctx, *tenantID, -1) // Delete all.
	})

	// List them.
	entries, total, err := repo.List(ctx, *tenantID, 0, 100)
	require.NoError(t, err)
	require.Equal(t, int64(numEntries), total)
	require.Len(t, entries, numEntries)
}
