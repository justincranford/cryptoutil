package service

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

const (
	corruptBase64Padding     = "===invalid==="
	corruptBase64Exclamation = "not-valid-base64!!!"
)

func TestJWTService_DatabaseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(JWTService, context.Context) error
	}{
		{name: "CreateJWT", call: func(svc JWTService, ctx context.Context) error {
			_, err := svc.CreateJWT(ctx, googleUuid.New(), googleUuid.New(), &JWTClaims{Issuer: "test"})

			return err
		}},
		{name: "CreateEncryptedJWT", call: func(svc JWTService, ctx context.Context) error {
			_, err := svc.CreateEncryptedJWT(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New(), &JWTClaims{Issuer: "test"})

			return err
		}},
		{name: "ValidateJWT", call: func(svc JWTService, ctx context.Context) error {
			_, err := svc.ValidateJWT(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJ0ZXN0In0.test")

			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)
			svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

			err := tc.call(svc, context.Background())
			require.Error(t, err)
			require.True(t,
				strings.Contains(err.Error(), "failed to") ||
					strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "parse"),
				"Expected database, not-found, or parse error, got: %v", err)
		})
	}
}

func TestJWTService_CreateJWT_PrivateKeyCorruption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		corruptFn func(t *testing.T, ctx context.Context) string
		wantErr   string
	}{
		{name: "barrier decrypt failure", corruptFn: func(_ *testing.T, _ context.Context) string {
			return base64.StdEncoding.EncodeToString([]byte("not-barrier"))
		}, wantErr: "failed to decrypt private JWK"},
		{name: "base64 decode failure padding", corruptFn: func(_ *testing.T, _ context.Context) string {
			return corruptBase64Padding
		}, wantErr: "failed to decode private JWK JWE"},
		{name: "corrupted private JWK parse", corruptFn: func(t *testing.T, ctx context.Context) string {
			t.Helper()

			encrypted, err := testBarrierService.EncryptContentWithContext(ctx, []byte("not-valid-json"))
			require.NoError(t, err)

			return base64.StdEncoding.EncodeToString(encrypted)
		}, wantErr: "failed to parse private JWK"},
		{name: "base64 decode failure exclamation", corruptFn: func(_ *testing.T, _ context.Context) string {
			return corruptBase64Exclamation
		}, wantErr: "failed to decode private JWK JWE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", tc.corruptFn(t, ctx)).Error
			require.NoError(t, err)

			expiry := time.Now().UTC().Add(time.Hour)
			_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, &JWTClaims{Subject: "test", ExpiresAt: &expiry})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestJWTService_ValidateJWT_PublicKeyCorruption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		corruptFn func(t *testing.T, ctx context.Context) string
		wantErr   string
	}{
		{name: "barrier decrypt failure", corruptFn: func(_ *testing.T, _ context.Context) string {
			return base64.StdEncoding.EncodeToString([]byte("not-barrier"))
		}, wantErr: "failed to decrypt public JWK"},
		{name: "base64 decode failure padding", corruptFn: func(_ *testing.T, _ context.Context) string {
			return corruptBase64Padding
		}, wantErr: "failed to decode public JWK JWE"},
		{name: "corrupted public JWK parse", corruptFn: func(t *testing.T, ctx context.Context) string {
			t.Helper()

			encrypted, err := testBarrierService.EncryptContentWithContext(ctx, []byte("not-valid-json"))
			require.NoError(t, err)

			return base64.StdEncoding.EncodeToString(encrypted)
		}, wantErr: "failed to parse public JWK"},
		{name: "base64 decode failure exclamation", corruptFn: func(_ *testing.T, _ context.Context) string {
			return corruptBase64Exclamation
		}, wantErr: "failed to decode public JWK JWE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			expiry := time.Now().UTC().Add(time.Hour)
			token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, &JWTClaims{Subject: "test", ExpiresAt: &expiry})
			require.NoError(t, err)

			err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", tc.corruptFn(t, ctx)).Error
			require.NoError(t, err)

			_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestJWTService_CreateEncryptedJWT_KeyValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, ctx context.Context, svc ElasticJWKService) (googleUuid.UUID, googleUuid.UUID, googleUuid.UUID)
		wantErr string
	}{
		{name: "encryption key wrong tenant", setup: func(t *testing.T, ctx context.Context, svc ElasticJWKService) (googleUuid.UUID, googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID, otherTenantID := googleUuid.New(), googleUuid.New()

			sigKey, _, err := svc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			encKey, _, err := svc.CreateElasticJWK(ctx, otherTenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, sigKey.ID, encKey.ID
		}, wantErr: "encryption key not found"},
		{name: "encryption key wrong use", setup: func(t *testing.T, ctx context.Context, svc ElasticJWKService) (googleUuid.UUID, googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()

			sigKey, _, err := svc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			otherSigKey, _, err := svc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, sigKey.ID, otherSigKey.ID
		}, wantErr: "not configured for encryption"},
		{name: "unsupported encryption algorithm", setup: func(t *testing.T, ctx context.Context, svc ElasticJWKService) (googleUuid.UUID, googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()

			sigKey, _, err := svc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			encKey, _, err := svc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, sigKey.ID, encKey.ID
		}, wantErr: "unsupported algorithm for JWE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID, sigKeyID, encKeyID := tc.setup(t, ctx, elasticSvc)

			expiry := time.Now().UTC().Add(time.Hour)
			_, err := jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigKeyID, encKeyID, &JWTClaims{Subject: "test-user", ExpiresAt: &expiry})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestJWTService_CreateJWT_KeyValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID)
		wantErr string
	}{
		{name: "algorithm key mismatch", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm).Error
			require.NoError(t, err)

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to create signer"},
		{name: "no active material", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to get active material"},
		{name: "unsupported algorithm", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to create signer"},
		{name: "wrong key use", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			tenantID := googleUuid.New()

			encKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, encKey.ID
		}, wantErr: "not configured for signing"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID, elasticJWKID := tc.setup(t, ctx)

			expiry := time.Now().UTC().Add(time.Hour)
			_, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWKID, &JWTClaims{Subject: "test", ExpiresAt: &expiry})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestJWTService_CreateEncryptedJWT_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	encrypted, err := testBarrierService.EncryptContentWithContext(ctx, []byte("not-valid-json"))
	require.NoError(t, err)

	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", encMaterial.ID).Update("public_jwk_jwe", base64.StdEncoding.EncodeToString(encrypted)).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, &JWTClaims{Subject: "test", ExpiresAt: &expiry})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

func TestJWTService_ValidateJWT_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, ctx context.Context, elasticSvc ElasticJWKService, jwtSvc JWTService) (googleUuid.UUID, googleUuid.UUID, string)
		wantErr string
	}{
		{name: "material KID not belong to elastic JWK", setup: func(t *testing.T, ctx context.Context, elasticSvc ElasticJWKService, jwtSvc JWTService) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()

			sigKey1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			sigKey2, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			expiry := time.Now().UTC().Add(time.Hour)
			token, err := jwtSvc.CreateJWT(ctx, tenantID, sigKey1.ID, &JWTClaims{Subject: "test-user", ExpiresAt: &expiry})
			require.NoError(t, err)

			return tenantID, sigKey2.ID, token
		}, wantErr: "does not belong to this elastic JWK"},
		{name: "invalid token no headers", setup: func(t *testing.T, ctx context.Context, elasticSvc ElasticJWKService, _ JWTService) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()

			sigKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, sigKey.ID, "invalid.token.structure"
		}, wantErr: "failed to parse JWT"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID, elasticJWKID, token := tc.setup(t, ctx, elasticSvc, jwtSvc)

			_, err := jwtSvc.ValidateJWT(ctx, tenantID, elasticJWKID, token)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
