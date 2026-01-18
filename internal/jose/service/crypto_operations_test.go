// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// testNonExistentKID is used for tests that need a KID that doesn't exist.
const testNonExistentKID = "nonexistent-kid-12345"

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

// TestGetDecryptedMaterialJWK_Success tests successful retrieval of decrypted material JWK.
func TestGetDecryptedMaterialJWK_Success(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK with an asymmetric key.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Get the decrypted material JWK.
	privateJWK, publicJWK, err := testElasticJWKSvc.GetDecryptedMaterialJWK(testCtx, createResp.MaterialJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, privateJWK)
	require.NotNil(t, publicJWK)
}

// TestGetDecryptedMaterialJWK_NotFound tests that GetDecryptedMaterialJWK fails for non-existent material.
func TestGetDecryptedMaterialJWK_NotFound(t *testing.T) {
	t.Parallel()

	// Try to get a non-existent material JWK.
	nonExistentID := googleUuid.New()
	_, _, err := testElasticJWKSvc.GetDecryptedMaterialJWK(testCtx, nonExistentID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material JWK")
}

// TestGetDecryptedPublicJWKs_TenantMismatch tests that GetDecryptedPublicJWKs fails for wrong tenant.
func TestGetDecryptedPublicJWKs_TenantMismatch(t *testing.T) {
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

	// Try to get public JWKs with wrong tenant.
	_, err = testElasticJWKSvc.GetDecryptedPublicJWKs(testCtx, wrongTenantID, realmID, createResp.ElasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found in specified tenant/realm")
}

// TestGetDecryptedPublicJWKs_RealmMismatch tests that GetDecryptedPublicJWKs fails for wrong realm.
func TestGetDecryptedPublicJWKs_RealmMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongRealmID := googleUuid.New()

	// Create an Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Try to get public JWKs with wrong realm.
	_, err = testElasticJWKSvc.GetDecryptedPublicJWKs(testCtx, tenantID, wrongRealmID, createResp.ElasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found in specified tenant/realm")
}

// TestSign_RealmMismatch tests that signing fails when realm doesn't match.
func TestSign_RealmMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongRealmID := googleUuid.New()

	// Create an Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Try to sign with wrong realm.
	_, err = testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     tenantID,
		RealmID:      wrongRealmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Payload:      []byte("test"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found in specified tenant/realm")
}

// TestEncrypt_RealmMismatch tests that encryption fails when realm doesn't match.
func TestEncrypt_RealmMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongRealmID := googleUuid.New()

	// Create an Elastic JWK for encryption.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ECDH-ES+A256KW",
		USE:      "enc",
	})
	require.NoError(t, err)

	// Try to encrypt with wrong realm.
	_, err = testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      wrongRealmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("test plaintext"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found in specified tenant/realm")
}

// TestVerify_RealmMismatch tests that verification fails when material_kid belongs to different tenant.
func TestVerify_RealmMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongTenantID := googleUuid.New()

	// Create and sign with correct tenant.
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
		Payload:      []byte("test payload"),
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

// TestDecrypt_RealmMismatch tests that decryption fails when material_kid belongs to different tenant.
func TestDecrypt_RealmMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	wrongTenantID := googleUuid.New()

	// Create and encrypt with correct tenant.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ECDH-ES+A256KW",
		USE:      "enc",
	})
	require.NoError(t, err)

	encryptResp, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("test plaintext"),
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

// TestListElasticJWKs_Success tests successful listing of Elastic JWKs.
func TestListElasticJWKs_Success(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create a few Elastic JWKs.
	for i := 0; i < 3; i++ {
		_, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
			TenantID: tenantID,
			RealmID:  realmID,
			KTY:      "EC",
			ALG:      "ES256",
			USE:      "sig",
		})
		require.NoError(t, err)
	}

	// List the Elastic JWKs.
	elasticJWKs, err := testElasticJWKSvc.ListElasticJWKs(testCtx, tenantID, realmID, 0, 100)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(elasticJWKs), 3)
}

// TestGetElasticJWK_NotFound tests that GetElasticJWK fails for non-existent elastic JWK.
func TestGetElasticJWK_NotFound(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Try to get a non-existent elastic JWK.
	_, err := testElasticJWKSvc.GetElasticJWK(testCtx, tenantID, realmID, testNonExistentKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestGetActiveMaterialJWK_NotFound tests that GetActiveMaterialJWK fails for non-existent elastic JWK.
func TestGetActiveMaterialJWK_NotFound(t *testing.T) {
	t.Parallel()

	nonExistentID := googleUuid.New()

	// Try to get active material for non-existent elastic JWK.
	_, _, _, err := testElasticJWKSvc.GetActiveMaterialJWK(testCtx, nonExistentID)
	require.Error(t, err)
}

// TestGetMaterialJWKByKID_NotFound tests that GetMaterialJWKByKID fails for non-existent material KID.
func TestGetMaterialJWKByKID_NotFound(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK to have a valid elastic ID.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Try to get material by non-existent KID.
	_, _, _, err = testElasticJWKSvc.GetMaterialJWKByKID(testCtx, createResp.ElasticJWK.ID, testNonExistentKID)
	require.Error(t, err)
}

// TestVerify_InvalidJWS tests that verification fails with invalid JWS.
func TestVerify_InvalidJWS(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	// Try to verify invalid JWS.
	_, err := testElasticJWKSvc.Verify(testCtx, &VerifyRequest{
		TenantID:        tenantID,
		JWSMessageBytes: []byte("invalid-jws-data"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWS message")
}

// TestDecrypt_InvalidJWE tests that decryption fails with invalid JWE.
func TestDecrypt_InvalidJWE(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	// Try to decrypt invalid JWE.
	_, err := testElasticJWKSvc.Decrypt(testCtx, &DecryptRequest{
		TenantID:        tenantID,
		JWEMessageBytes: []byte("invalid-jwe-data"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWE message")
}

// TestSign_NotFound tests that signing fails when elastic JWK doesn't exist.
func TestSign_NotFound(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	nonExistentID := googleUuid.New()

	// Try to sign with non-existent elastic JWK.
	_, err := testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: nonExistentID,
		Payload:      []byte("test"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get elastic JWK")
}

// TestEncrypt_NotFound tests that encryption fails when elastic JWK doesn't exist.
func TestEncrypt_NotFound(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	nonExistentID := googleUuid.New()

	// Try to encrypt with non-existent elastic JWK.
	_, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: nonExistentID,
		Plaintext:    []byte("test"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get elastic JWK")
}

// TestVerify_MaterialNotFound tests that verification fails when material JWK lookup fails.
func TestVerify_MaterialNotFound(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create and sign with correct tenant.
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
		Payload:      []byte("test payload"),
	})
	require.NoError(t, err)

	// Create a JWS with non-existent kid (manually crafted would be complex, so we just test with wrong tenant).
	differentTenantID := googleUuid.New()
	_, err = testElasticJWKSvc.Verify(testCtx, &VerifyRequest{
		TenantID:        differentTenantID,
		JWSMessageBytes: signResp.JWSMessageBytes,
	})
	require.Error(t, err)
}

// TestDecrypt_MaterialNotFound tests that decryption fails when material JWK lookup fails.
func TestDecrypt_MaterialNotFound(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create and encrypt with correct tenant.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ECDH-ES+A256KW",
		USE:      "enc",
	})
	require.NoError(t, err)

	encryptResp, err := testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    []byte("test plaintext"),
	})
	require.NoError(t, err)

	// Try to decrypt with different tenant (material will not be found for this tenant).
	differentTenantID := googleUuid.New()
	_, err = testElasticJWKSvc.Decrypt(testCtx, &DecryptRequest{
		TenantID:        differentTenantID,
		JWEMessageBytes: encryptResp.JWEMessageBytes,
	})
	require.Error(t, err)
}

// TestListElasticJWKs_EmptyResult tests listing returns empty for non-existent tenant/realm.
func TestListElasticJWKs_EmptyResult(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// List Elastic JWKs for tenant/realm with no JWKs.
	elasticJWKs, err := testElasticJWKSvc.ListElasticJWKs(testCtx, tenantID, realmID, 0, 100)
	require.NoError(t, err)
	require.Empty(t, elasticJWKs)
}

// TestCanRotate_Success tests that CanRotate returns true for a valid elastic JWK.
func TestCanRotate_Success(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Check if we can rotate.
	canRotate, count, err := testElasticJWKSvc.CanRotate(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.True(t, canRotate)
	require.Equal(t, int64(1), count)
}

// TestCanRotate_NotFound tests that CanRotate returns 0 count for non-existent elastic JWK.
func TestCanRotate_NotFound(t *testing.T) {
	t.Parallel()

	nonExistentID := googleUuid.New()

	// Check if we can rotate non-existent JWK - returns true with 0 count (no materials).
	canRotate, count, err := testElasticJWKSvc.CanRotate(testCtx, nonExistentID)
	require.NoError(t, err)
	require.True(t, canRotate) // Can rotate because count is 0 < MaxMaterialsPerElasticJWK
	require.Equal(t, int64(0), count)
}

// TestGetMaterialCount_Success tests that GetMaterialCount returns correct count.
func TestGetMaterialCount_Success(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK.
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Get material count.
	count, err := testElasticJWKSvc.GetMaterialCount(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Rotate and check count again.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)

	count, err = testElasticJWKSvc.GetMaterialCount(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

// TestGetMaterialCount_NotFound tests that GetMaterialCount returns 0 for non-existent elastic JWK.
func TestGetMaterialCount_NotFound(t *testing.T) {
	t.Parallel()

	nonExistentID := googleUuid.New()

	// Get material count for non-existent JWK - returns 0 (no materials).
	count, err := testElasticJWKSvc.GetMaterialCount(testCtx, nonExistentID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

// TestGetDecryptedPublicJWKs_NotFound tests that GetDecryptedPublicJWKs fails for non-existent elastic JWK.
func TestGetDecryptedPublicJWKs_NotFound(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	nonExistentID := googleUuid.New()

	// Try to get public JWKs for non-existent elastic JWK.
	_, err := testElasticJWKSvc.GetDecryptedPublicJWKs(testCtx, tenantID, realmID, nonExistentID)
	require.Error(t, err)
}

// TestCreateElasticJWK_InvalidSignatureAlgorithm tests that creating a signing JWK with invalid algorithm fails.
func TestCreateElasticJWK_InvalidSignatureAlgorithm(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Try to create with invalid signing algorithm.
	_, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "INVALID-SIG-ALG",
		USE:      "sig",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported signature algorithm")
}

// TestCreateElasticJWK_InvalidEncryptionAlgorithm tests that creating an encryption JWK with invalid algorithm fails.
func TestCreateElasticJWK_InvalidEncryptionAlgorithm(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Try to create with invalid encryption algorithm.
	_, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "INVALID-ENC-ALG",
		USE:      "enc",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported encryption algorithm")
}

// TestSign_UseTypeMismatch tests that signing fails when the elastic JWK is not a signing key.
func TestSign_UseTypeMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an encryption JWK (use=enc).
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ECDH-ES",
		USE:      "enc",
	})
	require.NoError(t, err)

	// Try to sign with an encryption key - should fail.
	payload := []byte("test payload to sign")
	_, err = testElasticJWKSvc.Sign(testCtx, &SignRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Payload:      payload,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a signing key")
}

// TestEncrypt_UseTypeMismatch tests that encryption fails when the elastic JWK is not an encryption key.
func TestEncrypt_UseTypeMismatch(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create a signing JWK (use=sig).
	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	})
	require.NoError(t, err)

	// Try to encrypt with a signing key - should fail.
	plaintext := []byte("test plaintext to encrypt")
	_, err = testElasticJWKSvc.Encrypt(testCtx, &EncryptRequest{
		TenantID:     tenantID,
		RealmID:      realmID,
		ElasticJWKID: createResp.ElasticJWK.ID,
		Plaintext:    plaintext,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not an encryption key")
}
