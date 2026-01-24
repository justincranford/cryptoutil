// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

func TestElasticJWKService_CreateElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		algorithm    string
		use          string
		maxMaterials int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid RS256 signing key",
			algorithm:    "RS256",
			use:          "sig",
			maxMaterials: 10,
			wantErr:      false,
		},
		{
			name:         "valid ES256 signing key",
			algorithm:    "ES256",
			use:          "sig",
			maxMaterials: 5,
			wantErr:      false,
		},
		{
			name:         "valid EdDSA signing key",
			algorithm:    "EdDSA",
			use:          "sig",
			maxMaterials: 3,
			wantErr:      false,
		},
		{
			name:         "valid A256GCM encryption key",
			algorithm:    "A256GCM",
			use:          "enc",
			maxMaterials: 20,
			wantErr:      false,
		},
		{
			name:         "default max materials when zero",
			algorithm:    "RS256",
			use:          "sig",
			maxMaterials: 0,
			wantErr:      false,
		},
		{
			name:         "invalid algorithm",
			algorithm:    "INVALID",
			use:          "sig",
			maxMaterials: 10,
			wantErr:      true,
			errContains:  "invalid algorithm",
		},
		{
			name:         "invalid key use",
			algorithm:    "RS256",
			use:          "invalid",
			maxMaterials: 10,
			wantErr:      true,
			errContains:  "invalid key use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenantID, err := cryptoutilSharedUtilRandom.GenerateUUIDv7()
			require.NoError(t, err)

			svc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)

			elasticJWK, material, err := svc.CreateElasticJWK(ctx, *tenantID, tt.algorithm, tt.use, tt.maxMaterials)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, elasticJWK)
				require.Nil(t, material)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, elasticJWK)
			require.NotNil(t, material)

			// Verify elastic JWK fields.
			require.Equal(t, *tenantID, elasticJWK.TenantID)
			require.Equal(t, tt.algorithm, elasticJWK.Algorithm)
			require.Equal(t, tt.use, elasticJWK.Use)

			if tt.maxMaterials <= 0 {
				require.Equal(t, 10, elasticJWK.MaxMaterials) // Default value.
			} else {
				require.Equal(t, tt.maxMaterials, elasticJWK.MaxMaterials)
			}

			require.Equal(t, 1, elasticJWK.CurrentMaterialCount)
			require.NotEmpty(t, elasticJWK.KID)

			// Verify material fields.
			require.Equal(t, elasticJWK.ID, material.ElasticJWKID)
			require.True(t, material.Active)
			require.NotEmpty(t, material.MaterialKID)
			require.NotEmpty(t, material.PrivateJWKJWE)
			require.NotEmpty(t, material.PublicJWKJWE)
			require.Equal(t, 1, material.BarrierVersion)

			// Cleanup.
			require.NoError(t, svc.DeleteElasticJWK(ctx, *tenantID, elasticJWK.ID))
		})
	}
}

func TestElasticJWKService_GetElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)

	// Create a test elastic JWK.
	tenantID, err := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	require.NoError(t, err)

	elasticJWK, _, err := svc.CreateElasticJWK(ctx, *tenantID, "RS256", "sig", 10)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = svc.DeleteElasticJWK(ctx, *tenantID, elasticJWK.ID)
	})

	tests := []struct {
		name        string
		tenantID    googleUuid.UUID
		id          googleUuid.UUID
		wantErr     bool
		errContains string
	}{
		{
			name:     "existing JWK",
			tenantID: *tenantID,
			id:       elasticJWK.ID,
			wantErr:  false,
		},
		{
			name:        "wrong tenant",
			tenantID:    googleUuid.New(),
			id:          elasticJWK.ID,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "non-existent JWK",
			tenantID:    *tenantID,
			id:          googleUuid.New(),
			wantErr:     true,
			errContains: "failed to get",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			retrieved, err := svc.GetElasticJWK(ctx, tt.tenantID, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, retrieved)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, retrieved)
			require.Equal(t, elasticJWK.ID, retrieved.ID)
			require.Equal(t, elasticJWK.TenantID, retrieved.TenantID)
			require.Equal(t, elasticJWK.KID, retrieved.KID)
		})
	}
}

func TestElasticJWKService_ListElasticJWKs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)

	// Create a unique tenant for this test.
	tenantID, err := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	require.NoError(t, err)

	// Create multiple elastic JWKs.
	var createdIDs []googleUuid.UUID

	for i := 0; i < 3; i++ {
		elasticJWK, _, err := svc.CreateElasticJWK(ctx, *tenantID, "RS256", "sig", 10)
		require.NoError(t, err)

		createdIDs = append(createdIDs, elasticJWK.ID)
	}

	t.Cleanup(func() {
		for _, id := range createdIDs {
			_ = svc.DeleteElasticJWK(ctx, *tenantID, id)
		}
	})

	tests := []struct {
		name        string
		tenantID    googleUuid.UUID
		offset      int
		limit       int
		expectCount int
		wantErr     bool
	}{
		{
			name:        "list all",
			tenantID:    *tenantID,
			offset:      0,
			limit:       100,
			expectCount: 3,
			wantErr:     false,
		},
		{
			name:        "list with pagination",
			tenantID:    *tenantID,
			offset:      0,
			limit:       2,
			expectCount: 2,
			wantErr:     false,
		},
		{
			name:        "list with offset",
			tenantID:    *tenantID,
			offset:      1,
			limit:       100,
			expectCount: 2,
			wantErr:     false,
		},
		{
			name:        "list empty tenant",
			tenantID:    googleUuid.New(),
			offset:      0,
			limit:       100,
			expectCount: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticJWKs, total, err := svc.ListElasticJWKs(ctx, tt.tenantID, tt.offset, tt.limit)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Len(t, elasticJWKs, tt.expectCount)

			if tt.tenantID == *tenantID {
				require.GreaterOrEqual(t, total, int64(3))
			} else {
				require.Equal(t, int64(0), total)
			}
		})
	}
}

func TestElasticJWKService_DeleteElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)

	tests := []struct {
		name        string
		setup       func() (tenantID, id googleUuid.UUID)
		wantErr     bool
		errContains string
	}{
		{
			name: "delete existing JWK",
			setup: func() (tenantID, id googleUuid.UUID) {
				tid, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
				elasticJWK, _, _ := svc.CreateElasticJWK(ctx, *tid, "RS256", "sig", 10)

				return *tid, elasticJWK.ID
			},
			wantErr: false,
		},
		{
			name: "delete non-existent JWK",
			setup: func() (tenantID, id googleUuid.UUID) {
				tid, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

				return *tid, googleUuid.New()
			},
			wantErr:     true,
			errContains: "failed to get",
		},
		{
			name: "delete with wrong tenant",
			setup: func() (tenantID, id googleUuid.UUID) {
				tid, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
				elasticJWK, _, _ := svc.CreateElasticJWK(ctx, *tid, "RS256", "sig", 10)
				wrongTenant := googleUuid.New()

				return wrongTenant, elasticJWK.ID
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenantID, id := tt.setup()
			err := svc.DeleteElasticJWK(ctx, tenantID, id)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)

				return
			}

			require.NoError(t, err)

			// Verify deletion.
			_, err = svc.GetElasticJWK(ctx, tenantID, id)
			require.Error(t, err)
		})
	}
}

func TestMapAlgorithmToKeyType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		expected  string
	}{
		{"RS256", cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"RS384", cryptoutilSharedMagic.JoseAlgRS384, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"RS512", cryptoutilSharedMagic.JoseAlgRS512, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"PS256", cryptoutilSharedMagic.JoseAlgPS256, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"PS384", cryptoutilSharedMagic.JoseAlgPS384, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"PS512", cryptoutilSharedMagic.JoseAlgPS512, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"ES256", cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{"ES384", cryptoutilSharedMagic.JoseAlgES384, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{"ES512", cryptoutilSharedMagic.JoseAlgES512, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{"EdDSA", cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaDomain.KeyTypeOKP},
		{"A128GCM", cryptoutilSharedMagic.JoseEncA128GCM, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"A192GCM", cryptoutilSharedMagic.JoseEncA192GCM, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"A256GCM", cryptoutilSharedMagic.JoseEncA256GCM, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"A128CBC-HS256", cryptoutilSharedMagic.JoseEncA128CBCHS256, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"A256CBC-HS512", cryptoutilSharedMagic.JoseEncA256CBCHS512, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"INVALID", "INVALID", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapAlgorithmToKeyType(tt.algorithm)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestMapToGenerateAlgorithm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		expectNil bool
	}{
		{"RS256", cryptoutilSharedMagic.JoseAlgRS256, false},
		{"PS256", cryptoutilSharedMagic.JoseAlgPS256, false},
		{"ES256", cryptoutilSharedMagic.JoseAlgES256, false},
		{"ES384", cryptoutilSharedMagic.JoseAlgES384, false},
		{"ES512", cryptoutilSharedMagic.JoseAlgES512, false},
		{"EdDSA", cryptoutilSharedMagic.JoseAlgEdDSA, false},
		{"A128GCM", cryptoutilSharedMagic.JoseEncA128GCM, false},
		{"A192GCM", cryptoutilSharedMagic.JoseEncA192GCM, false},
		{"A256GCM", cryptoutilSharedMagic.JoseEncA256GCM, false},
		{"RSA/2048", cryptoutilSharedMagic.JoseKeyTypeRSA2048, false},
		{"RSA/3072", cryptoutilSharedMagic.JoseKeyTypeRSA3072, false},
		{"RSA/4096", cryptoutilSharedMagic.JoseKeyTypeRSA4096, false},
		{"EC/P256", cryptoutilSharedMagic.JoseKeyTypeECP256, false},
		{"EC/P384", cryptoutilSharedMagic.JoseKeyTypeECP384, false},
		{"EC/P521", cryptoutilSharedMagic.JoseKeyTypeECP521, false},
		{"OKP/Ed25519", cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, false},
		{"oct/128", cryptoutilSharedMagic.JoseKeyTypeOct128, false},
		{"oct/192", cryptoutilSharedMagic.JoseKeyTypeOct192, false},
		{"oct/256", cryptoutilSharedMagic.JoseKeyTypeOct256, false},
		{"INVALID", "INVALID", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapToGenerateAlgorithm(tt.algorithm)

			if tt.expectNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
		})
	}
}
