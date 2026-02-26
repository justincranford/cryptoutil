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
			algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			use:          cryptoutilSharedMagic.JoseKeyUseSig,
			maxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			wantErr:      false,
		},
		{
			name:         "valid ES256 signing key",
			algorithm:    cryptoutilSharedMagic.JoseAlgES256,
			use:          cryptoutilSharedMagic.JoseKeyUseSig,
			maxMaterials: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			wantErr:      false,
		},
		{
			name:         "valid EdDSA signing key",
			algorithm:    cryptoutilSharedMagic.JoseAlgEdDSA,
			use:          cryptoutilSharedMagic.JoseKeyUseSig,
			maxMaterials: 3,
			wantErr:      false,
		},
		{
			name:         "valid A256GCM encryption key",
			algorithm:    cryptoutilSharedMagic.JoseEncA256GCM,
			use:          cryptoutilSharedMagic.JoseKeyUseEnc,
			maxMaterials: cryptoutilSharedMagic.MaxErrorDisplay,
			wantErr:      false,
		},
		{
			name:         "valid A128CBC-HS256 encryption key",
			algorithm:    cryptoutilSharedMagic.JoseEncA128CBCHS256,
			use:          cryptoutilSharedMagic.JoseKeyUseEnc,
			maxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			wantErr:      false,
		},
		{
			name:         "valid A192CBC-HS384 encryption key",
			algorithm:    cryptoutilSharedMagic.JoseEncA192CBCHS384,
			use:          cryptoutilSharedMagic.JoseKeyUseEnc,
			maxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			wantErr:      false,
		},
		{
			name:         "valid A256CBC-HS512 encryption key",
			algorithm:    cryptoutilSharedMagic.JoseEncA256CBCHS512,
			use:          cryptoutilSharedMagic.JoseKeyUseEnc,
			maxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			wantErr:      false,
		},
		{
			name:         "default max materials when zero",
			algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			use:          cryptoutilSharedMagic.JoseKeyUseSig,
			maxMaterials: 0,
			wantErr:      false,
		},
		{
			name:         "invalid algorithm",
			algorithm:    "INVALID",
			use:          cryptoutilSharedMagic.JoseKeyUseSig,
			maxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			wantErr:      true,
			errContains:  "invalid algorithm",
		},
		{
			name:         "invalid key use",
			algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			use:          "invalid",
			maxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
				require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, elasticJWK.MaxMaterials) // Default value.
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

	elasticJWK, _, err := svc.CreateElasticJWK(ctx, *tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseKeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
		elasticJWK, _, err := svc.CreateElasticJWK(ctx, *tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseKeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
			limit:       cryptoutilSharedMagic.JoseJAMaxMaterials,
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
			limit:       cryptoutilSharedMagic.JoseJAMaxMaterials,
			expectCount: 2,
			wantErr:     false,
		},
		{
			name:        "list empty tenant",
			tenantID:    googleUuid.New(),
			offset:      0,
			limit:       cryptoutilSharedMagic.JoseJAMaxMaterials,
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
				elasticJWK, _, _ := svc.CreateElasticJWK(ctx, *tid, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseKeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

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
				elasticJWK, _, _ := svc.CreateElasticJWK(ctx, *tid, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseKeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
		{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseAlgRS384, cryptoutilSharedMagic.JoseAlgRS384, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseAlgRS512, cryptoutilSharedMagic.JoseAlgRS512, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseAlgPS256, cryptoutilSharedMagic.JoseAlgPS256, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseAlgPS384, cryptoutilSharedMagic.JoseAlgPS384, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseAlgPS512, cryptoutilSharedMagic.JoseAlgPS512, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseAlgES256, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{cryptoutilSharedMagic.JoseAlgES384, cryptoutilSharedMagic.JoseAlgES384, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{cryptoutilSharedMagic.JoseAlgES512, cryptoutilSharedMagic.JoseAlgES512, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaDomain.KeyTypeOKP},
		{cryptoutilSharedMagic.JoseEncA128GCM, cryptoutilSharedMagic.JoseEncA128GCM, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseEncA192GCM, cryptoutilSharedMagic.JoseEncA192GCM, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseEncA256GCM, cryptoutilSharedMagic.JoseEncA256GCM, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseEncA128CBCHS256, cryptoutilSharedMagic.JoseEncA128CBCHS256, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseEncA192CBCHS384, cryptoutilSharedMagic.JoseEncA192CBCHS384, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseEncA256CBCHS512, cryptoutilSharedMagic.JoseEncA256CBCHS512, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseKeyTypeOct192, cryptoutilSharedMagic.JoseKeyTypeOct192, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseKeyTypeOct384, cryptoutilSharedMagic.JoseKeyTypeOct384, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseKeyTypeOct512, cryptoutilSharedMagic.JoseKeyTypeOct512, cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseKeyTypeRSA3072, cryptoutilSharedMagic.JoseKeyTypeRSA3072, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseKeyTypeRSA4096, cryptoutilSharedMagic.JoseKeyTypeRSA4096, cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{cryptoutilSharedMagic.JoseKeyTypeECP384, cryptoutilSharedMagic.JoseKeyTypeECP384, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{cryptoutilSharedMagic.JoseKeyTypeECP521, cryptoutilSharedMagic.JoseKeyTypeECP521, cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, cryptoutilAppsJoseJaDomain.KeyTypeOKP},
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
		{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgRS256, false},
		{cryptoutilSharedMagic.JoseAlgRS384, cryptoutilSharedMagic.JoseAlgRS384, false},
		{cryptoutilSharedMagic.JoseAlgRS512, cryptoutilSharedMagic.JoseAlgRS512, false},
		{cryptoutilSharedMagic.JoseAlgPS256, cryptoutilSharedMagic.JoseAlgPS256, false},
		{cryptoutilSharedMagic.JoseAlgPS384, cryptoutilSharedMagic.JoseAlgPS384, false},
		{cryptoutilSharedMagic.JoseAlgPS512, cryptoutilSharedMagic.JoseAlgPS512, false},
		{cryptoutilSharedMagic.JoseAlgES256, cryptoutilSharedMagic.JoseAlgES256, false},
		{cryptoutilSharedMagic.JoseAlgES384, cryptoutilSharedMagic.JoseAlgES384, false},
		{cryptoutilSharedMagic.JoseAlgES512, cryptoutilSharedMagic.JoseAlgES512, false},
		{cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilSharedMagic.JoseAlgEdDSA, false},
		{cryptoutilSharedMagic.JoseEncA128GCM, cryptoutilSharedMagic.JoseEncA128GCM, false},
		{cryptoutilSharedMagic.JoseEncA192GCM, cryptoutilSharedMagic.JoseEncA192GCM, false},
		{cryptoutilSharedMagic.JoseEncA256GCM, cryptoutilSharedMagic.JoseEncA256GCM, false},
		{cryptoutilSharedMagic.JoseEncA128CBCHS256, cryptoutilSharedMagic.JoseEncA128CBCHS256, false},
		{cryptoutilSharedMagic.JoseEncA192CBCHS384, cryptoutilSharedMagic.JoseEncA192CBCHS384, false},
		{cryptoutilSharedMagic.JoseEncA256CBCHS512, cryptoutilSharedMagic.JoseEncA256CBCHS512, false},
		{cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilSharedMagic.JoseKeyTypeRSA2048, false},
		{cryptoutilSharedMagic.JoseKeyTypeRSA3072, cryptoutilSharedMagic.JoseKeyTypeRSA3072, false},
		{cryptoutilSharedMagic.JoseKeyTypeRSA4096, cryptoutilSharedMagic.JoseKeyTypeRSA4096, false},
		{cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilSharedMagic.JoseKeyTypeECP256, false},
		{cryptoutilSharedMagic.JoseKeyTypeECP384, cryptoutilSharedMagic.JoseKeyTypeECP384, false},
		{cryptoutilSharedMagic.JoseKeyTypeECP521, cryptoutilSharedMagic.JoseKeyTypeECP521, false},
		{cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, false},
		{cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilSharedMagic.JoseKeyTypeOct128, false},
		{cryptoutilSharedMagic.JoseKeyTypeOct192, cryptoutilSharedMagic.JoseKeyTypeOct192, false},
		{cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilSharedMagic.JoseKeyTypeOct256, false},
		{cryptoutilSharedMagic.JoseKeyTypeOct384, cryptoutilSharedMagic.JoseKeyTypeOct384, false},
		{cryptoutilSharedMagic.JoseKeyTypeOct512, cryptoutilSharedMagic.JoseKeyTypeOct512, false},
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
