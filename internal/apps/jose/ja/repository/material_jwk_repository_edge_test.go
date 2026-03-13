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

func TestMaterialJWKRepository_GetActiveMaterialWhenNoneActive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	materialRepo := NewMaterialJWKRepository(testDB)
	elasticRepo := NewElasticJWKRepository(testDB)

	// Create parent elastic JWK.
	elasticID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	parentJWK := &cryptoutilAppsJoseJaModel.ElasticJWK{
		ID:           *elasticID,
		TenantID:     *tenantID,
		KID:          "parent-no-active-" + elasticID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:      cryptoutilAppsJoseJaModel.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		CreatedAt:    time.Now().UTC(),
	}
	require.NoError(t, elasticRepo.Create(ctx, parentJWK))

	defer func() {
		_ = elasticRepo.Delete(ctx, parentJWK.ID)
	}()

	// Create material but mark it as inactive.
	materialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	material := &cryptoutilAppsJoseJaModel.MaterialJWK{
		ID:            *materialID,
		ElasticJWKID:  parentJWK.ID,
		MaterialKID:   "inactive-material-" + materialID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
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

func TestMaterialJWKRepository_GetByIDEdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := repo.GetByID(ctx, tt.id)
			require.Error(t, err, "GetByID should fail for non-existent ID")
		})
	}
}

func TestMaterialJWKRepository_GetByMaterialKIDWithSpecialChars(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	materialRepo := NewMaterialJWKRepository(testDB)
	elasticRepo := NewElasticJWKRepository(testDB)

	// Create parent elastic JWK.
	elasticID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	parentJWK := &cryptoutilAppsJoseJaModel.ElasticJWK{
		ID:           *elasticID,
		TenantID:     *tenantID,
		KID:          "parent-" + elasticID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:      cryptoutilAppsJoseJaModel.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
			material := &cryptoutilAppsJoseJaModel.MaterialJWK{
				ID:            *materialID,
				ElasticJWKID:  parentJWK.ID,
				MaterialKID:   tt.materialKID + "-" + materialID.String()[:cryptoutilSharedMagic.IMMinPasswordLength], // Make unique.
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
