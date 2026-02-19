// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestJWTService_CreateJWT(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		algorithm string
	}{
		{
			name:      "RS256 JWT",
			algorithm: cryptoutilSharedMagic.JoseAlgRS256,
		},
		{
			name:      "ES256 JWT",
			algorithm: cryptoutilSharedMagic.JoseAlgES256,
		},
		{
			name:      "EdDSA JWT",
			algorithm: cryptoutilSharedMagic.JoseAlgEdDSA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			// Create signing key.
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, tt.algorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
			require.NoError(t, err)

			// Create JWT.
			exp := time.Now().UTC().Add(time.Hour)
			claims := &JWTClaims{
				Issuer:    "test-issuer",
				Subject:   "test-subject",
				Audience:  []string{"test-audience"},
				ExpiresAt: &exp,
				JTI:       googleUuid.New().String(),
			}
			token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
			require.NoError(t, err)
			require.NotEmpty(t, token)
		})
	}
}

func TestJWTService_CreateJWT_InvalidKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key (not signing).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to create JWT with encryption key - should fail.
	claims := &JWTClaims{Issuer: "test"}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for signing")
}

func TestJWTService_CreateJWT_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to create JWT with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	claims := &JWTClaims{Issuer: "test"}
	_, err = jwtSvc.CreateJWT(ctx, wrongTenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWTService_ValidateJWT(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		algorithm string
	}{
		{
			name:      "RS256 JWT roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgRS256,
		},
		{
			name:      "ES256 JWT roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgES256,
		},
		{
			name:      "EdDSA JWT roundtrip",
			algorithm: cryptoutilSharedMagic.JoseAlgEdDSA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			// Create signing key.
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, tt.algorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
			require.NoError(t, err)

			// Create JWT.
			exp := time.Now().UTC().Add(time.Hour)
			originalClaims := &JWTClaims{
				Issuer:    "test-issuer",
				Subject:   "test-subject",
				Audience:  []string{"test-audience"},
				ExpiresAt: &exp,
				JTI:       googleUuid.New().String(),
			}
			token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, originalClaims)
			require.NoError(t, err)

			// Validate JWT.
			validatedClaims, err := jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
			require.NoError(t, err)
			require.Equal(t, originalClaims.Issuer, validatedClaims.Issuer)
			require.Equal(t, originalClaims.Subject, validatedClaims.Subject)
			require.Equal(t, originalClaims.JTI, validatedClaims.JTI)
		})
	}
}

func TestJWTService_ValidateJWT_Expired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create expired JWT.
	exp := time.Now().UTC().Add(-time.Hour) // Already expired.
	claims := &JWTClaims{
		Issuer:    "test-issuer",
		ExpiresAt: &exp,
	}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	// Validate expired JWT - should fail.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestJWTService_ValidateJWT_NotYetValid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create JWT with future not-before.
	nbf := time.Now().UTC().Add(time.Hour) // Not yet valid.
	exp := time.Now().UTC().Add(2 * time.Hour)
	claims := &JWTClaims{
		Issuer:    "test-issuer",
		NotBefore: &nbf,
		ExpiresAt: &exp,
	}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	// Validate JWT that's not yet valid - should fail.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet valid")
}

func TestJWTService_ValidateJWT_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create JWT.
	exp := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{
		Issuer:    "test",
		ExpiresAt: &exp,
	}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	// Try to validate with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jwtSvc.ValidateJWT(ctx, wrongTenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWTService_ValidateJWT_InvalidToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to validate invalid token - should fail.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, "not.a.valid.jwt")
	require.Error(t, err)
}
