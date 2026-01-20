// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	joseJADomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

func TestJWKSService_GetJWKS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create multiple keys for the tenant.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	_, _, err = elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgES256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	_, _, err = elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Get JWKS.
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.NotNil(t, jwks)
	require.Len(t, jwks.Keys, 3)
}

func TestJWKSService_GetJWKS_EmptyTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Get JWKS for tenant with no keys.
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.NotNil(t, jwks)
	require.Empty(t, jwks.Keys)
}

func TestJWKSService_GetJWKSForElasticKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Get JWKS for specific elastic key.
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, jwks)
	require.Len(t, jwks.Keys, 1)
}

func TestJWKSService_GetJWKSForElasticKey_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to get JWKS with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jwksSvc.GetJWKSForElasticKey(ctx, wrongTenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWKSService_GetJWKSForElasticKey_NonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to get JWKS for non-existent key - should fail.
	_, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, googleUuid.New())
	require.Error(t, err)
}

func TestJWKSService_GetPublicJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Get public JWK by KID.
	publicJWK, err := jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.NoError(t, err)
	require.NotNil(t, publicJWK)
	require.Equal(t, material.MaterialKID, publicJWK.KeyID)
	require.Equal(t, elasticJWK.Use, publicJWK.Use)
	require.Equal(t, elasticJWK.Algorithm, publicJWK.Algorithm)
}

func TestJWKSService_GetPublicJWK_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create key.
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to get public JWK with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jwksSvc.GetPublicJWK(ctx, wrongTenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWKSService_GetPublicJWK_InvalidKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to get public JWK with invalid KID - should fail.
	_, err := jwksSvc.GetPublicJWK(ctx, tenantID, "invalid-kid-that-does-not-exist")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWKSService_GetJWKS_MultipleTenants(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)

	// Create keys for two different tenants.
	tenantID1 := googleUuid.New()
	tenantID2 := googleUuid.New()

	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID1, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	_, _, err = elasticSvc.CreateElasticJWK(ctx, tenantID1, cryptoutilMagic.JoseAlgES256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	_, _, err = elasticSvc.CreateElasticJWK(ctx, tenantID2, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Get JWKS for tenant1 - should have 2 keys.
	jwks1, err := jwksSvc.GetJWKS(ctx, tenantID1)
	require.NoError(t, err)
	require.Len(t, jwks1.Keys, 2)

	// Get JWKS for tenant2 - should have 1 key.
	jwks2, err := jwksSvc.GetJWKS(ctx, tenantID2)
	require.NoError(t, err)
	require.Len(t, jwks2.Keys, 1)
}

func TestJWKSService_GetJWKS_VerifyPublicKeyOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create key.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Get JWKS.
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Len(t, jwks.Keys, 1)

	// Verify that the key is public only (no private key material).
	publicJWK := jwks.Keys[0]
	require.True(t, publicJWK.IsPublic(), "JWKS should only contain public keys")
}
