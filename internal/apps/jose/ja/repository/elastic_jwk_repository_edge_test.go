// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose/ja/model"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestElasticJWKRepository_DeleteAlreadyDeleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwk := &cryptoutilAppsJoseJaModel.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-double-delete-" + id.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:      cryptoutilAppsJoseJaModel.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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

func TestElasticJWKRepository_GetByIDWithInvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	tests := []struct {
		name string
		id   googleUuid.UUID
	}{
		{
			name: "nil UUID",
			id:   googleUuid.UUID{},
		},
		{
			name: "max UUID",
			id:   googleUuid.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
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

			jwk := &cryptoutilAppsJoseJaModel.ElasticJWK{
				ID:           *id,
				TenantID:     *tenantID,
				KID:          tt.kid + "-" + id.String()[:cryptoutilSharedMagic.IMMinPasswordLength], // Make unique per test run.
				KeyType:      cryptoutilAppsJoseJaModel.KeyTypeRSA,
				Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
				Use:          cryptoutilSharedMagic.JoseKeyUseSig,
				MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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

func TestElasticJWKRepository_ListWithEmptyDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Use a unique tenant ID that has no JWKs.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwks, total, err := repo.List(ctx, *tenantID, 0, cryptoutilSharedMagic.JoseJAMaxMaterials)
	require.NoError(t, err)
	require.Empty(t, jwks)
	require.Equal(t, int64(0), total)
}

func TestElasticJWKRepository_UpdateNonExistentJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	nonExistentJWK := &cryptoutilAppsJoseJaModel.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "non-existent-" + id.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:      cryptoutilAppsJoseJaModel.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
