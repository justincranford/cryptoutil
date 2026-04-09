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

// corruptPrivateJWK returns a corruption function that sets the material's private_jwk_jwe to the given value.
func corruptPrivateJWK(t *testing.T, ctx context.Context, materialID googleUuid.UUID, value string) {
	t.Helper()

	err := testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", materialID).Update("private_jwk_jwe", value).Error
	require.NoError(t, err)
}

// corruptPublicJWK sets the material's public_jwk_jwe to the given value.
func corruptPublicJWK(t *testing.T, ctx context.Context, materialID googleUuid.UUID, value string) {
	t.Helper()

	err := testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", materialID).Update("public_jwk_jwe", value).Error
	require.NoError(t, err)
}

// barrierEncryptedInvalidJSON returns a base64-encoded barrier-encrypted blob of invalid JSON.
func barrierEncryptedInvalidJSON(t *testing.T, ctx context.Context) string {
	t.Helper()

	encrypted, err := testBarrierService.EncryptContentWithContext(ctx, []byte("not-valid-json"))
	require.NoError(t, err)

	return base64.StdEncoding.EncodeToString(encrypted)
}

func TestJWEService_DatabaseError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		invoke func(svc JWEService, ctx context.Context) error
	}{
		{name: "decrypt closed DB", invoke: func(svc JWEService, ctx context.Context) error {
			_, err := svc.Decrypt(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJlbmMiOiJBMjU2R0NNIn0.test.test.test.test")

			return err
		}},
		{name: "encrypt closed DB", invoke: func(svc JWEService, ctx context.Context) error {
			_, err := svc.Encrypt(ctx, googleUuid.New(), googleUuid.New(), []byte("test plaintext"))

			return err
		}},
		{name: "encrypt with KID closed DB", invoke: func(svc JWEService, ctx context.Context) error {
			_, err := svc.EncryptWithKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid", []byte("test plaintext"))

			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)
			svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

			err := tc.invoke(svc, context.Background())
			require.Error(t, err)
			require.True(t,
				strings.Contains(err.Error(), "failed to") ||
					strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "parse"),
				"Expected database, not-found, or parse error, got: %v", err)
		})
	}
}

func TestJWEService_Decrypt_CorruptedMaterial(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		corruptValue func(t *testing.T, ctx context.Context) string
	}{
		{name: "barrier decrypt failure", corruptValue: func(_ *testing.T, _ context.Context) string {
			return base64.StdEncoding.EncodeToString([]byte("not-barrier"))
		}},
		{name: "invalid base64 equals signs", corruptValue: func(_ *testing.T, _ context.Context) string {
			return "===invalid==="
		}},
		{name: "invalid base64 exclamation", corruptValue: func(_ *testing.T, _ context.Context) string {
			return "not-valid-base64!!!"
		}},
		{name: "invalid base64 mixed", corruptValue: func(_ *testing.T, _ context.Context) string {
			return "invalid-base64!!!"
		}},
		{name: "corrupted private JWK parse", corruptValue: func(t *testing.T, ctx context.Context) string {
			t.Helper()

			return barrierEncryptedInvalidJSON(t, ctx)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			encrypted, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
			require.NoError(t, err)

			corruptPrivateJWK(t, ctx, material.ID, tc.corruptValue(t, ctx))

			_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, encrypted)
			require.Error(t, err)
			require.Contains(t, err.Error(), "no matching key found")
		})
	}
}

func TestJWEService_EncryptDecrypt_SymmetricKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	plaintext := []byte("secret message for symmetric key")
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	decrypted, err := jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestJWEService_EncryptWithKID_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, ctx context.Context) (tenantID, elasticJWKID googleUuid.UUID, materialKID string)
		wantErr string
	}{
		{name: "algorithm key mismatch", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
			require.NoError(t, err)

			return tenantID, elasticJWK.ID, material.MaterialKID
		}, wantErr: "failed to create encrypter"},
		{name: "corrupted base64", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			corruptPublicJWK(t, ctx, material.ID, "===invalid===")

			return tenantID, elasticJWK.ID, material.MaterialKID
		}, wantErr: "failed to decode public JWK JWE"},
		{name: "corrupted JWE", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			corruptPublicJWK(t, ctx, material.ID, base64.StdEncoding.EncodeToString([]byte("not-a-jwe")))

			return tenantID, elasticJWK.ID, material.MaterialKID
		}, wantErr: "failed to decrypt public JWK"},
		{name: "corrupted public JWK parse", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			corruptPublicJWK(t, ctx, material.ID, barrierEncryptedInvalidJSON(t, ctx))

			return tenantID, elasticJWK.ID, material.MaterialKID
		}, wantErr: "failed to parse public JWK"},
		{name: "material not found", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID, "nonexistent-kid"
		}, wantErr: "failed to get material"},
		{name: "material wrong elastic JWK", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			_, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, elasticJWK1.ID, material2.MaterialKID
		}, wantErr: "material key does not belong to elastic JWK"},
		{name: "unsupported algorithm", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID, material.MaterialKID
		}, wantErr: "unsupported algorithm for JWE"},
		{name: "wrong key use", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID, string) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID, material.MaterialKID
		}, wantErr: "not configured for encryption"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			tenantID, elasticJWKID, materialKID := tc.setup(t, ctx)
			jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)

			_, err := jweSvc.EncryptWithKID(ctx, tenantID, elasticJWKID, materialKID, []byte("test"))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestJWEService_Encrypt_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, ctx context.Context) (tenantID, elasticJWKID googleUuid.UUID)
		wantErr string
	}{
		{name: "algorithm key mismatch", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
			require.NoError(t, err)

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to create encrypter"},
		{name: "barrier decrypt failure", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			corruptPublicJWK(t, ctx, material.ID, base64.StdEncoding.EncodeToString([]byte("not-barrier")))

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to decrypt public JWK"},
		{name: "corrupted base64", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			corruptPublicJWK(t, ctx, material.ID, "not-valid-base64!!!")

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to decode public JWK JWE"},
		{name: "corrupted public JWK parse", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			corruptPublicJWK(t, ctx, material.ID, barrierEncryptedInvalidJSON(t, ctx))

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to parse public JWK"},
		{name: "no active material", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)
			err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID
		}, wantErr: "failed to get active material"},
		{name: "unsupported algorithm", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID
		}, wantErr: "unsupported algorithm for JWE"},
		{name: "wrong key use", setup: func(t *testing.T, ctx context.Context) (googleUuid.UUID, googleUuid.UUID) {
			t.Helper()

			tenantID := googleUuid.New()
			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
			require.NoError(t, err)

			return tenantID, elasticJWK.ID
		}, wantErr: "not configured for encryption"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			tenantID, elasticJWKID := tc.setup(t, ctx)
			jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)

			_, err := jweSvc.Encrypt(ctx, tenantID, elasticJWKID, []byte("test"))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
