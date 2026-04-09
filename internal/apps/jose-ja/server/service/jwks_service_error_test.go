package service

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

func TestJWKSService_DatabaseError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		invoke  func(t *testing.T, ctx context.Context) error
		wantErr string
	}{
		{name: "GetJWKS closed DB", invoke: func(t *testing.T, ctx context.Context) error {
			t.Helper()

			elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)
			svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

			_, err := svc.GetJWKS(ctx, googleUuid.New())

			return err
		}},
		{name: "GetJWKSForElasticKey closed DB", invoke: func(t *testing.T, ctx context.Context) error {
			t.Helper()

			elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)
			svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

			_, err := svc.GetJWKSForElasticKey(ctx, googleUuid.New(), googleUuid.New())

			return err
		}},
		{name: "GetJWKSForElasticKey list materials closed DB", invoke: func(t *testing.T, ctx context.Context) error {
			t.Helper()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			brokenMaterialRepo := closedDBMaterialRepo(t)
			jwksSvc := NewJWKSService(testElasticRepo, brokenMaterialRepo, testBarrierService)

			_, err = jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)

			return err
		}, wantErr: "failed to list materials"},
		{name: "GetPublicJWK closed DB", invoke: func(t *testing.T, ctx context.Context) error {
			t.Helper()

			elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)
			svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

			_, err := svc.GetPublicJWK(ctx, googleUuid.New(), "test-kid")

			return err
		}},
		{name: "GetPublicJWK elastic JWK deleted", invoke: func(t *testing.T, ctx context.Context) error {
			t.Helper()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			err = testDB.Delete(&cryptoutilAppsJoseJaModel.ElasticJWK{}, "id = ?", elasticJWK.ID).Error
			require.NoError(t, err)

			_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)

			return err
		}, wantErr: "failed to get elastic JWK"},
		{name: "GetPublicJWK wrong KID", invoke: func(t *testing.T, ctx context.Context) error {
			t.Helper()

			jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)

			_, err := jwksSvc.GetPublicJWK(ctx, googleUuid.New(), "nonexistent-kid")

			return err
		}, wantErr: "failed to get material"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.invoke(t, context.Background())
			require.Error(t, err)

			if tc.wantErr != "" {
				require.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.True(t,
					strings.Contains(err.Error(), "failed to") ||
						strings.Contains(err.Error(), "not found"),
					"Expected database or not-found error, got: %v", err)
			}
		})
	}
}

func TestJWKSService_GetJWKSForElasticKey_GracefulDegradation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, elasticID, materialID googleUuid.UUID)
	}{
		{name: "barrier decrypt skip", setup: func(t *testing.T, ctx context.Context, _, _, materialID googleUuid.UUID) {
			t.Helper()
			corruptPublicJWK(t, ctx, materialID, base64.StdEncoding.EncodeToString([]byte("not-barrier")))
		}},
		{name: "base64 decode skip", setup: func(t *testing.T, ctx context.Context, _, _, materialID googleUuid.UUID) {
			t.Helper()
			corruptPublicJWK(t, ctx, materialID, corruptBase64Padding)
		}},
		{name: "corrupted public JWK parse", setup: func(t *testing.T, ctx context.Context, _, _, materialID googleUuid.UUID) {
			t.Helper()
			corruptPublicJWK(t, ctx, materialID, barrierEncryptedInvalidJSON(t, ctx))
		}},
		{name: "inactive materials", setup: func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, elasticID, materialID googleUuid.UUID) {
			t.Helper()

			rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			err := rotationSvc.RetireMaterial(ctx, tenantID, elasticID, materialID)
			require.NoError(t, err)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			tc.setup(t, ctx, tenantID, elasticJWK.ID, material.ID)

			jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
			require.NoError(t, err)
			require.Empty(t, jwks.Keys)
		})
	}
}

func TestJWKSService_GetJWKS_GracefulDegradation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, elasticID, materialID googleUuid.UUID) googleUuid.UUID
	}{
		{name: "barrier decrypt failure", setup: func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, _, materialID googleUuid.UUID) googleUuid.UUID {
			t.Helper()
			corruptPublicJWK(t, ctx, materialID, base64.StdEncoding.EncodeToString([]byte("not-barrier")))

			return tenantID
		}},
		{name: "base64 decode failure", setup: func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, _, materialID googleUuid.UUID) googleUuid.UUID {
			t.Helper()
			corruptPublicJWK(t, ctx, materialID, corruptBase64Padding)

			return tenantID
		}},
		{name: "corrupted base64 mixed", setup: func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, _, materialID googleUuid.UUID) googleUuid.UUID {
			t.Helper()
			corruptPublicJWK(t, ctx, materialID, corruptBase64Exclamation)

			return tenantID
		}},
		{name: "corrupted public JWK parse", setup: func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, _, materialID googleUuid.UUID) googleUuid.UUID {
			t.Helper()
			corruptPublicJWK(t, ctx, materialID, barrierEncryptedInvalidJSON(t, ctx))

			return tenantID
		}},
		{name: "no active material", setup: func(t *testing.T, ctx context.Context, tenantID googleUuid.UUID, elasticID, materialID googleUuid.UUID) googleUuid.UUID {
			t.Helper()

			rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			err := rotationSvc.RetireMaterial(ctx, tenantID, elasticID, materialID)
			require.NoError(t, err)

			return tenantID
		}},
		{name: "empty for wrong tenant", setup: func(_ *testing.T, _ context.Context, _ googleUuid.UUID, _, _ googleUuid.UUID) googleUuid.UUID {
			return googleUuid.New()
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			queryTenantID := tc.setup(t, ctx, tenantID, elasticJWK.ID, material.ID)

			jwks, err := jwksSvc.GetJWKS(ctx, queryTenantID)
			require.NoError(t, err)
			require.Empty(t, jwks.Keys)
		})
	}
}

func TestJWKSService_GetPublicJWK_CorruptedMaterial(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		corruptValue func(t *testing.T, ctx context.Context) string
		wantErr      string
	}{
		{name: "base64 decode failure equals", corruptValue: func(_ *testing.T, _ context.Context) string {
			return corruptBase64Padding
		}, wantErr: "failed to decode public JWK JWE"},
		{name: "base64 decode failure exclamation", corruptValue: func(_ *testing.T, _ context.Context) string {
			return corruptBase64Exclamation
		}, wantErr: "failed to decode public JWK JWE"},
		{name: "barrier decrypt failure", corruptValue: func(_ *testing.T, _ context.Context) string {
			return base64.StdEncoding.EncodeToString([]byte("not-barrier"))
		}, wantErr: "failed to decrypt public JWK"},
		{name: "corrupted public JWK parse", corruptValue: func(t *testing.T, ctx context.Context) string {
			t.Helper()

			return barrierEncryptedInvalidJSON(t, ctx)
		}, wantErr: "failed to parse public JWK"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			corruptPublicJWK(t, ctx, material.ID, tc.corruptValue(t, ctx))

			_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestVerify_ListMaterialsDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	jws, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)

	brokenMaterialRepo := closedDBMaterialRepo(t)
	brokenJWSSvc := NewJWSService(testElasticRepo, brokenMaterialRepo, testBarrierService)

	_, err = brokenJWSSvc.Verify(ctx, tenantID, elasticJWK.ID, jws)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list materials")
}
