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

func TestJWEService_Encrypt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		algorithm string
	}{
		{
			name:      "RSA-2048 encryption",
			algorithm: cryptoutilMagic.JoseKeyTypeRSA2048,
		},
		{
			name:      "RSA-3072 encryption",
			algorithm: cryptoutilMagic.JoseKeyTypeRSA3072,
		},
		{
			name:      "Oct-128 direct encryption",
			algorithm: cryptoutilMagic.JoseKeyTypeOct128,
		},
		{
			name:      "Oct-256 direct encryption",
			algorithm: cryptoutilMagic.JoseKeyTypeOct256,
		},
		{
			name:      "ECP-256 encryption",
			algorithm: cryptoutilMagic.JoseKeyTypeECP256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			// Create encryption key.
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, tt.algorithm, joseJADomain.KeyUseEnc, 10)
			require.NoError(t, err)

			// Encrypt plaintext.
			plaintext := []byte("Hello, World!")
			jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, jweCompact)
		})
	}
}

func TestJWEService_Encrypt_InvalidKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key (not encryption).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

func TestJWEService_Encrypt_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to encrypt with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jweSvc.Encrypt(ctx, wrongTenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWEService_Encrypt_NonExistentKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to encrypt with non-existent key - should fail.
	_, err := jweSvc.Encrypt(ctx, tenantID, googleUuid.New(), []byte("test"))
	require.Error(t, err)
}

func TestJWEService_Decrypt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		algorithm string
	}{
		{
			name:      "RSA-2048 roundtrip",
			algorithm: cryptoutilMagic.JoseKeyTypeRSA2048,
		},
		{
			name:      "RSA-3072 roundtrip",
			algorithm: cryptoutilMagic.JoseKeyTypeRSA3072,
		},
		{
			name:      "Oct-128 roundtrip",
			algorithm: cryptoutilMagic.JoseKeyTypeOct128,
		},
		{
			name:      "Oct-256 roundtrip",
			algorithm: cryptoutilMagic.JoseKeyTypeOct256,
		},
		{
			name:      "ECP-256 roundtrip",
			algorithm: cryptoutilMagic.JoseKeyTypeECP256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
			tenantID := googleUuid.New()

			// Create encryption key.
			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, tt.algorithm, joseJADomain.KeyUseEnc, 10)
			require.NoError(t, err)

			// Encrypt plaintext.
			plaintext := []byte("Hello, World!")
			jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, plaintext)
			require.NoError(t, err)

			// Decrypt.
			decrypted, err := jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)
		})
	}
}

func TestJWEService_Decrypt_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Encrypt plaintext.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	// Try to decrypt with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jweSvc.Decrypt(ctx, wrongTenantID, elasticJWK.ID, jweCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWEService_Decrypt_InvalidJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to decrypt invalid JWE - should fail.
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, "not.a.valid.jwe.compact")
	require.Error(t, err)
}

func TestJWEService_EncryptWithKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Encrypt with specific KID.
	plaintext := []byte("Hello, KID!")
	jweCompact, err := jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	// Decrypt and verify.
	decrypted, err := jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestJWEService_EncryptWithKID_InvalidKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseAlgRS256, joseJADomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

func TestJWEService_EncryptWithKID_InvalidKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to encrypt with invalid KID - should fail.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, "invalid-kid", []byte("test"))
	require.Error(t, err)
}

func TestJWEService_EncryptWithKID_NonExistentKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to encrypt with non-existent elastic JWK - should fail.
	_, err := jweSvc.EncryptWithKID(ctx, tenantID, googleUuid.New(), "some-kid", []byte("test"))
	require.Error(t, err)
}

func TestJWEService_EncryptWithKID_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to encrypt with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = jweSvc.EncryptWithKID(ctx, wrongTenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestJWEService_EncryptWithKID_MaterialBelongsToOtherKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two encryption keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)
	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotEqual(t, elasticJWK1.ID, elasticJWK2.ID)

	// Try to encrypt using elasticJWK1 but with material from elasticJWK2 - should fail.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong")
}

func TestJWEService_Decrypt_NoMatchingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two separate encryption keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)
	elasticJWK2, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Encrypt with elasticJWK1.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK1.ID, []byte("secret"))
	require.NoError(t, err)

	// Try to decrypt with elasticJWK2 (different key) - should fail.
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK2.ID, jweCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWEService_Decrypt_NonExistentKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilMagic.JoseKeyTypeRSA2048, joseJADomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Encrypt plaintext.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	// Try to decrypt with non-existent key - should fail.
	_, err = jweSvc.Decrypt(ctx, tenantID, googleUuid.New(), jweCompact)
	require.Error(t, err)
}
