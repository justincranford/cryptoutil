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
)

func TestJWSService_Sign(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		algorithm string
	}{
		{
			name:      "RS256 signing",
			algorithm: cryptoutilSharedMagic.JoseAlgRS256,
		},
		{
			name:      "RS384 signing",
			algorithm: cryptoutilSharedMagic.JoseAlgRS384,
		},
		{
			name:      "PS256 signing",
			algorithm: cryptoutilSharedMagic.JoseAlgPS256,
		},
		{
			name:      "ES256 signing",
			algorithm: cryptoutilSharedMagic.JoseAlgES256,
		},
		{
			name:      "EdDSA signing",
			algorithm: cryptoutilSharedMagic.JoseAlgEdDSA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			// Create signing key.
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, tt.algorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
			require.NoError(t, err)

			// Sign payload.
			payload := []byte("Hello, World!")
			jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, payload)
			require.NoError(t, err)
			require.NotEmpty(t, jwsCompact)
		})
	}
}

func TestJWSService_Sign_InvalidKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key (not signing).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to sign with encryption key - should fail.
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for signing")
}

func TestJWSService_Sign_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to sign with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jwsSvc.Sign(ctx, wrongTenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWSService_Sign_NonExistentKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to sign with non-existent key - should fail.
	_, err := jwsSvc.Sign(ctx, tenantID, googleUuid.New(), []byte("test"))
	require.Error(t, err)
}

func TestJWSService_Verify(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		algorithm string
	}{
		{
			name:      "RS256 roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgRS256,
		},
		{
			name:      "RS384 roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgRS384,
		},
		{
			name:      "PS256 roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgPS256,
		},
		{
			name:      "ES256 roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgES256,
		},
		{
			name:      "EdDSA roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgEdDSA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			// Create signing key.
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, tt.algorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
			require.NoError(t, err)

			// Sign payload.
			payload := []byte("Hello, World!")
			jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, payload)
			require.NoError(t, err)

			// Verify.
			verified, err := jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, jwsCompact)
			require.NoError(t, err)
			require.Equal(t, payload, verified)
		})
	}
}

func TestJWSService_Verify_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Sign payload.
	jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	// Try to verify with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jwsSvc.Verify(ctx, wrongTenantID, elasticJWK.ID, jwsCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWSService_Verify_InvalidJWS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to verify invalid JWS - should fail.
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, "not.a.valid.jws")
	require.Error(t, err)
}

func TestJWSService_SignWithKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Sign with specific KID.
	payload := []byte("Hello, KID!")
	jwsCompact, err := jwsSvc.SignWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, payload)
	require.NoError(t, err)
	require.NotEmpty(t, jwsCompact)

	// Verify.
	verified, err := jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, jwsCompact)
	require.NoError(t, err)
	require.Equal(t, payload, verified)
}

func TestJWSService_SignWithKID_InvalidKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to sign with encryption key - should fail.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for signing")
}

func TestJWSService_SignWithKID_InvalidKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to sign with invalid KID - should fail.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK.ID, "invalid-kid", []byte("test"))
	require.Error(t, err)
}

func TestJWSService_SignWithKID_NonExistentKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to sign with non-existent elastic JWK - should fail.
	_, err := jwsSvc.SignWithKID(ctx, tenantID, googleUuid.New(), "some-kid", []byte("test"))
	require.Error(t, err)
}

func TestJWSService_SignWithKID_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to sign with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jwsSvc.SignWithKID(ctx, wrongTenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWSService_SignWithKID_MaterialBelongsToOtherKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two signing keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotEqual(t, elasticJWK1.ID, elasticJWK2.ID)

	// Try to sign using elasticJWK1 but with material from elasticJWK2 - should fail.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong")
}

func TestJWSService_Verify_NoMatchingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two separate signing keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	elasticJWK2, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Sign with elasticJWK1.
	jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK1.ID, []byte("secret"))
	require.NoError(t, err)

	// Try to verify with elasticJWK2 (different key) - should fail.
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK2.ID, jwsCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWSService_Verify_NonExistentKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Sign payload.
	jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	// Try to verify with non-existent key - should fail.
	_, err = jwsSvc.Verify(ctx, tenantID, googleUuid.New(), jwsCompact)
	require.Error(t, err)
}
