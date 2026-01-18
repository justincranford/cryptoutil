// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Tests in this file use the shared testElasticJWKSvc from elastic_jwk_service_test.go TestMain.

// TestSign_Success tests successful signing with an Elastic JWK.
func TestSign_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kty  string
		alg  string
	}{
		{name: "RSA_RS256", kty: "RSA", alg: "RS256"},
		{name: "EC_ES256", kty: "EC", alg: "ES256"},
		{name: "OKP_EdDSA", kty: "OKP", alg: "EdDSA"},
		{name: "oct_HS256", kty: "oct", alg: "HS256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenantID := googleUuid.New()
			realmID := googleUuid.New()

			// Create an Elastic JWK for signing.
			createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
				TenantID: tenantID,
				RealmID:  realmID,
				KTY:      tt.kty,
				ALG:      tt.alg,
				USE:      "sig",
			})
			require.NoError(t, err)
			require.NotNil(t, createResp)

			// Sign some data.
			payload := []byte("test payload for signing")
			signResp, err := testElasticJWKSvc.Sign(testCtx, &SignRequest{
				TenantID:     tenantID,
				RealmID:      realmID,
				ElasticJWKID: createResp.ElasticJWK.ID,
				Payload:      payload,
			})
			require.NoError(t, err)
			require.NotNil(t, signResp)
			require.NotNil(t, signResp.JWSMessage)
			require.NotEmpty(t, signResp.JWSMessageBytes)
			require.NotEmpty(t, signResp.MaterialKID)

			// The material KID should match the active material.
			require.Equal(t, createResp.MaterialJWK.MaterialKID, signResp.MaterialKID)
		})
	}
}

// TestSign_TenantMismatch tests that signing fails when tenant doesn't match.
func TestSign_TenantMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongTenantID := googleUuid.New()

	// Create an Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Try to sign with wrong tenant.
	_, err = testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     wrongTenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Payload:      []byte("test"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found in specified tenant/realm")
}

// TestSign_WrongUse tests that signing fails when the key is for encryption.
func TestSign_WrongUse(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK for encryption.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RSA-OAEP-256",
		USE:      "enc",
	})
	require.NoError(t, err)

	// Try to sign with an encryption key.
	_, err = testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Payload:      []byte("test"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a signing key")
}

// TestEncrypt_Success tests successful encryption with an Elastic JWK.
func TestEncrypt_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kty  string
		alg  string
	}{
		{name: "RSA_RSA-OAEP-256", kty: "RSA", alg: "RSA-OAEP-256"},
		{name: "EC_ECDH-ES+A256KW", kty: "EC", alg: "ECDH-ES+A256KW"},
		{name: "oct_A256GCM_dir", kty: "oct", alg: "A256GCM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenantID := googleUuid.New()
			realmID := googleUuid.New()

			// Create an Elastic JWK for encryption.
			createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
				TenantID: tenantID,
				RealmID:  realmID,
				KTY:      tt.kty,
				ALG:      tt.alg,
				USE:      "enc",
			})
			require.NoError(t, err)
			require.NotNil(t, createResp)

			// Encrypt some data.
			plaintext := []byte("test plaintext for encryption")
			encryptResp, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
				TenantID:     tenantID,
				RealmID:      realmID,
				ElasticJWKID: createResp.ElasticJWK.ID,
				Plaintext:    plaintext,
			})
			require.NoError(t, err)
			require.NotNil(t, encryptResp)
			require.NotNil(t, encryptResp.JWEMessage)
			require.NotEmpty(t, encryptResp.JWEMessageBytes)
			require.NotEmpty(t, encryptResp.MaterialKID)

			// The material KID should match the active material.
			require.Equal(t, createResp.MaterialJWK.MaterialKID, encryptResp.MaterialKID)
		})
	}
}

// TestEncrypt_TenantMismatch tests that encryption fails when tenant doesn't match.
func TestEncrypt_TenantMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongTenantID := googleUuid.New()

	// Create an Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RSA-OAEP-256",
		USE:      "enc",
	})
	require.NoError(t, err)

	// Try to encrypt with wrong tenant.
	_, err = testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     wrongTenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("test"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found in specified tenant/realm")
}

// TestEncrypt_WrongUse tests that encryption fails when the key is for signing.
func TestEncrypt_WrongUse(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK for signing.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Try to encrypt with a signing key.
	_, err = testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("test"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not an encryption key")
}

// TestSignAndVerify_RoundTrip tests sign then verify works correctly.
func TestSignAndVerify_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kty  string
		alg  string
	}{
		{name: "RSA_RS256", kty: "RSA", alg: "RS256"},
		{name: "EC_ES256", kty: "EC", alg: "ES256"},
		{name: "OKP_EdDSA", kty: "OKP", alg: "EdDSA"},
		{name: "oct_HS256", kty: "oct", alg: "HS256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenantID := googleUuid.New()
			realmID := googleUuid.New()

			// Create an Elastic JWK for signing.
			createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
				TenantID: tenantID,
				RealmID:  realmID,
				KTY:      tt.kty,
				ALG:      tt.alg,
				USE:      "sig",
			})
			require.NoError(t, err)

			// Sign some data.
			payload := []byte("test payload for round trip")
			signResp, err := testElasticJWKSvc.Sign(testCtx, &SignRequest{
				TenantID:     tenantID,
				RealmID:      realmID,
				ElasticJWKID: createResp.ElasticJWK.ID,
				Payload:      payload,
			})
			require.NoError(t, err)

			// Verify the signature.
			verifyResp, err := testElasticJWKSvc.Verify(testCtx, &VerifyRequest{
				TenantID:        tenantID,
				JWSMessageBytes: signResp.JWSMessageBytes,
			})
			require.NoError(t, err)
			require.NotNil(t, verifyResp)
			require.Equal(t, payload, verifyResp.Payload)
			require.Equal(t, signResp.MaterialKID, verifyResp.MaterialKID)
		})
	}
}

// TestEncryptAndDecrypt_RoundTrip tests encrypt then decrypt works correctly.
func TestEncryptAndDecrypt_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kty  string
		alg  string
	}{
		{name: "RSA_RSA-OAEP-256", kty: "RSA", alg: "RSA-OAEP-256"},
		{name: "EC_ECDH-ES+A256KW", kty: "EC", alg: "ECDH-ES+A256KW"},
		{name: "oct_A256GCM_dir", kty: "oct", alg: "A256GCM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tenantID := googleUuid.New()
			realmID := googleUuid.New()

			// Create an Elastic JWK for encryption.
			createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
				TenantID: tenantID,
				RealmID:  realmID,
				KTY:      tt.kty,
				ALG:      tt.alg,
				USE:      "enc",
			})
			require.NoError(t, err)

			// Encrypt some data.
			plaintext := []byte("test plaintext for round trip")
			encryptResp, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
				TenantID:     tenantID,
				RealmID:      realmID,
				ElasticJWKID: createResp.ElasticJWK.ID,
				Plaintext:    plaintext,
			})
			require.NoError(t, err)

			// Decrypt the data.
			decryptResp, err := testElasticJWKSvc.Decrypt(testCtx, &DecryptRequest{
				TenantID:        tenantID,
				JWEMessageBytes: encryptResp.JWEMessageBytes,
			})
			require.NoError(t, err)
			require.NotNil(t, decryptResp)
			require.Equal(t, plaintext, decryptResp.Plaintext)
			require.Equal(t, encryptResp.MaterialKID, decryptResp.MaterialKID)
		})
	}
}

// TestVerify_TenantMismatch tests that verification fails when tenant doesn't match.
func TestVerify_TenantMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongTenantID := googleUuid.New()

	// Create an Elastic JWK and sign data.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	signResp, err := testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Payload:      []byte("test"),
	})
	require.NoError(t, err)

	// Try to verify with wrong tenant.
	_, err = testElasticJWKSvc.Verify(testCtx, &VerifyRequest{
		TenantID:        wrongTenantID,
		JWSMessageBytes: signResp.JWSMessageBytes,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found for tenant")
}

// TestDecrypt_TenantMismatch tests that decryption fails when tenant doesn't match.
func TestDecrypt_TenantMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongTenantID := googleUuid.New()

	// Create an Elastic JWK and encrypt data.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RSA-OAEP-256",
		USE:      "enc",
	})
	require.NoError(t, err)

	encryptResp, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("test"),
	})
	require.NoError(t, err)

	// Try to decrypt with wrong tenant.
	_, err = testElasticJWKSvc.Decrypt(testCtx, &DecryptRequest{
		TenantID:        wrongTenantID,
		JWEMessageBytes: encryptResp.JWEMessageBytes,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found for tenant")
}

// TestVerify_HistoricalMaterial tests that verification works with rotated (historical) materials.
func TestVerify_HistoricalMaterial(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK and sign data.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Sign with the first material.
	signResp1, err := testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Payload:      []byte("first payload"),
	})
	require.NoError(t, err)

	// Rotate the material.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)

	// Sign with the new material.
	signResp2, err := testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Payload:      []byte("second payload"),
	})
	require.NoError(t, err)

	// Material KIDs should be different.
	require.NotEqual(t, signResp1.MaterialKID, signResp2.MaterialKID)

	// Verify the first signature (historical material).
	verifyResp1, err := testElasticJWKSvc.Verify(testCtx, &VerifyRequest{
		TenantID:        tenantID,
		JWSMessageBytes: signResp1.JWSMessageBytes,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("first payload"), verifyResp1.Payload)
	require.Equal(t, signResp1.MaterialKID, verifyResp1.MaterialKID)

	// Verify the second signature (active material).
	verifyResp2, err := testElasticJWKSvc.Verify(testCtx, &VerifyRequest{
		TenantID:        tenantID,
		JWSMessageBytes: signResp2.JWSMessageBytes,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("second payload"), verifyResp2.Payload)
	require.Equal(t, signResp2.MaterialKID, verifyResp2.MaterialKID)
}

// TestDecrypt_HistoricalMaterial tests that decryption works with rotated (historical) materials.
func TestDecrypt_HistoricalMaterial(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK and encrypt data.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RSA-OAEP-256",
		USE:      "enc",
	})
	require.NoError(t, err)

	// Encrypt with the first material.
	encryptResp1, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("first plaintext"),
	})
	require.NoError(t, err)

	// Rotate the material.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)

	// Encrypt with the new material.
	encryptResp2, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("second plaintext"),
	})
	require.NoError(t, err)

	// Material KIDs should be different.
	require.NotEqual(t, encryptResp1.MaterialKID, encryptResp2.MaterialKID)

	// Decrypt the first ciphertext (historical material).
	decryptResp1, err := testElasticJWKSvc.Decrypt(testCtx, &DecryptRequest{
		TenantID:        tenantID,
		JWEMessageBytes: encryptResp1.JWEMessageBytes,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("first plaintext"), decryptResp1.Plaintext)
	require.Equal(t, encryptResp1.MaterialKID, decryptResp1.MaterialKID)

	// Decrypt the second ciphertext (active material).
	decryptResp2, err := testElasticJWKSvc.Decrypt(testCtx, &DecryptRequest{
		TenantID:        tenantID,
		JWEMessageBytes: encryptResp2.JWEMessageBytes,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("second plaintext"), decryptResp2.Plaintext)
	require.Equal(t, encryptResp2.MaterialKID, decryptResp2.MaterialKID)
}

// TestGetDecryptedPublicJWKs_Success tests successful retrieval of public JWKs.
func TestGetDecryptedPublicJWKs_Success(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an asymmetric Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Rotate a few times to have multiple materials.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)

	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)

	// Get the public JWKs.
	publicJWKs, err := testElasticJWKSvc.GetDecryptedPublicJWKs(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.Len(t, publicJWKs, 3) // Original + 2 rotations
}

// TestGetDecryptedPublicJWKs_SymmetricKeyFails tests that symmetric keys cannot be exposed via JWKS.
func TestGetDecryptedPublicJWKs_SymmetricKeyFails(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create a symmetric Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "oct",
		ALG:      "HS256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Try to get public JWKs (should fail for symmetric keys).
	_, err = testElasticJWKSvc.GetDecryptedPublicJWKs(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "symmetric keys cannot be exposed via JWKS")
}
