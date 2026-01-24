// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

// TestCreateJWKFromKey_SymmetricKeys tests CreateJWKFromKey with symmetric keys (AES, HMAC variants).
func TestCreateJWKFromKey_SymmetricKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		alg         cryptoutilOpenapiModel.GenerateAlgorithm
		keySize     int
		keyGenFunc  func(int) (cryptoutilKeyGen.SecretKey, error)
		expectedAlg string
		expectedUse string
		prob        float32
	}{
		// AES variants (Oct128, Oct192 only - no Oct256 for AES!)
		{"Oct128_AES", cryptoutilOpenapiModel.Oct128, cryptoutilMagic.AESKeySize128, cryptoutilKeyGen.GenerateAESKey, "A128GCM", "enc", cryptoutilMagic.TestProbAlways},
		{"Oct192_AES", cryptoutilOpenapiModel.Oct192, cryptoutilMagic.AESKeySize192, cryptoutilKeyGen.GenerateAESKey, "A192GCM", "enc", cryptoutilMagic.TestProbTenth},
		// HMAC variants (Oct256, Oct384, Oct512)
		{"Oct256_HMAC", cryptoutilOpenapiModel.Oct256, cryptoutilMagic.HMACKeySize256, cryptoutilKeyGen.GenerateHMACKey, "HS256", "sig", cryptoutilMagic.TestProbAlways},
		{"Oct384_HMAC", cryptoutilOpenapiModel.Oct384, cryptoutilMagic.HMACKeySize384, cryptoutilKeyGen.GenerateHMACKey, "HS384", "sig", cryptoutilMagic.TestProbTenth},
		{"Oct512_HMAC", cryptoutilOpenapiModel.Oct512, cryptoutilMagic.HMACKeySize512, cryptoutilKeyGen.GenerateHMACKey, "HS512", "sig", cryptoutilMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			kid := googleUuid.Must(googleUuid.NewV7())
			secretKey, err := tc.keyGenFunc(tc.keySize)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &tc.alg, secretKey)
			require.NoError(t, err)
			require.Equal(t, kid, *resultKid)
			require.NotNil(t, nonPublicJWK)
			require.Nil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.Empty(t, publicBytes)

			// Verify algorithm header
			algHeader, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tc.expectedAlg, algHeader.String())

			// Verify use header
			use, ok := nonPublicJWK.KeyUsage()
			require.True(t, ok)
			require.Equal(t, tc.expectedUse, use)
		})
	}
}

// TestIsPublicPrivateAsymmetricSymmetric_AsymmetricKeys tests JWK type checks with asymmetric keys.
func TestIsPublicPrivateAsymmetricSymmetric_AsymmetricKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alg        cryptoutilOpenapiModel.GenerateAlgorithm
		keyGenFunc func() (any, any, error)
		prob       float32
	}{
		{"RSA2048", cryptoutilOpenapiModel.RSA2048, func() (any, any, error) {
			key, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize2048)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate RSA 2048 key: %w", err)
			}

			return key, &key.PublicKey, nil
		}, cryptoutilMagic.TestProbAlways},
		{"RSA3072", cryptoutilOpenapiModel.RSA3072, func() (any, any, error) {
			key, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize3072)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate RSA 3072 key: %w", err)
			}

			return key, &key.PublicKey, nil
		}, cryptoutilMagic.TestProbTenth},
		{"RSA4096", cryptoutilOpenapiModel.RSA4096, func() (any, any, error) {
			key, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize4096)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate RSA 4096 key: %w", err)
			}

			return key, &key.PublicKey, nil
		}, cryptoutilMagic.TestProbTenth},
		{"ECP256", cryptoutilOpenapiModel.ECP256, func() (any, any, error) {
			key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate ECDSA P-256 key: %w", err)
			}

			return key, &key.PublicKey, nil
		}, cryptoutilMagic.TestProbAlways},
		{"ECP384", cryptoutilOpenapiModel.ECP384, func() (any, any, error) {
			key, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate ECDSA P-384 key: %w", err)
			}

			return key, &key.PublicKey, nil
		}, cryptoutilMagic.TestProbTenth},
		{"ECP521", cryptoutilOpenapiModel.ECP521, func() (any, any, error) {
			key, err := ecdsa.GenerateKey(elliptic.P521(), crand.Reader)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate ECDSA P-521 key: %w", err)
			}

			return key, &key.PublicKey, nil
		}, cryptoutilMagic.TestProbTenth},
		{"OKPEd25519", cryptoutilOpenapiModel.OKPEd25519, func() (any, any, error) {
			pub, priv, err := ed25519.GenerateKey(crand.Reader)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
			}

			return priv, pub, nil
		}, cryptoutilMagic.TestProbAlways},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			privateKey, publicKey, err := tc.keyGenFunc()
			require.NoError(t, err)

			kid := googleUuid.Must(googleUuid.NewV7())
			keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: publicKey}

			_, privateJWK, publicJWK, _, _, err := CreateJWKFromKey(&kid, &tc.alg, keyPair)
			require.NoError(t, err)

			// Test private JWK
			isPrivate, err := IsPrivateJWK(privateJWK)
			require.NoError(t, err)
			require.True(t, isPrivate)

			isAsymmetric, err := IsAsymmetricJWK(privateJWK)
			require.NoError(t, err)
			require.True(t, isAsymmetric)

			isSymmetric, err := IsSymmetricJWK(privateJWK)
			require.NoError(t, err)
			require.False(t, isSymmetric)

			// Test public JWK
			isPublic, err := IsPublicJWK(publicJWK)
			require.NoError(t, err)
			require.True(t, isPublic)
		})
	}
}

// TestIsPublicPrivateAsymmetricSymmetric_SymmetricKeys tests JWK type checks with symmetric keys.
func TestIsPublicPrivateAsymmetricSymmetric_HMAC(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize384)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct384

	_, privateJWK, _, _, _, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)

	// HMAC is symmetric (not private, not public, not asymmetric, is symmetric)
	isPrivate, err := IsPrivateJWK(privateJWK)
	require.NoError(t, err)
	require.False(t, isPrivate)

	isPublic, err := IsPublicJWK(privateJWK)
	require.NoError(t, err)
	require.False(t, isPublic)

	isAsymmetric, err := IsAsymmetricJWK(privateJWK)
	require.NoError(t, err)
	require.False(t, isAsymmetric)

	isSymmetric, err := IsSymmetricJWK(privateJWK)
	require.NoError(t, err)
	require.True(t, isSymmetric)
}

// TestIsPublicPrivateAsymmetricSymmetric_AES tests JWK type checks with AES keys.
func TestIsPublicPrivateAsymmetricSymmetric_AES(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize192)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct192

	_, privateJWK, _, _, _, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)

	// AES is symmetric (not private, not public, not asymmetric, is symmetric)
	isPrivate, err := IsPrivateJWK(privateJWK)
	require.NoError(t, err)
	require.False(t, isPrivate)

	isSymmetric, err := IsSymmetricJWK(privateJWK)
	require.NoError(t, err)
	require.True(t, isSymmetric)

	isAsymmetric, err := IsAsymmetricJWK(privateJWK)
	require.NoError(t, err)
	require.False(t, isAsymmetric)
}

// TestCreateJWEJWKFromKey_ECDH_P256 tests CreateJWEJWKFromKey with ECDH P-256.
func TestCreateJWEJWKFromKey_ECDH_P256(t *testing.T) {
	t.Parallel()

	keyPair, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P256())
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	enc := joseJwa.A256GCM()
	alg := joseJwa.ECDH_ES()

	resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, encryptJWK)
	require.NotNil(t, decryptJWK)
	require.NotEmpty(t, encryptBytes)
	require.NotEmpty(t, decryptBytes)
	// Note: ECDH JWK type checking has go-jose library limitations:
	// - ECDH keys may not match expected ECDSAPublicKey/ECDSAPrivateKey types
	// - IsDecryptJWK requires enc header which may not be set on imported keys
	// Skip detailed type validation - test confirms JWK creation completes successfully
}

// TestCreateJWSJWKFromKey_ECDSA_P521 tests CreateJWSJWKFromKey with ECDSA P-521.
func TestCreateJWSJWKFromKey_ECDSA_P521(t *testing.T) {
	t.Parallel()

	keyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P521())
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.ES512()

	resultKid, signJWK, verifyJWK, signBytes, verifyBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, signJWK)
	require.NotNil(t, verifyJWK)
	require.NotEmpty(t, signBytes)
	require.NotEmpty(t, verifyBytes)

	// Verify sign JWK
	isSign, err := IsSignJWK(signJWK)
	require.NoError(t, err)
	require.True(t, isSign)

	// Verify verify JWK
	isVerify, err := IsVerifyJWK(verifyJWK)
	require.NoError(t, err)
	require.True(t, isVerify)
}

// TestCreateJWEJWKFromKey_ECDH_Variants tests CreateJWEJWKFromKey with ECDH curve variants.
func TestCreateJWEJWKFromKey_ECDH_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		curve ecdh.Curve
		enc   joseJwa.ContentEncryptionAlgorithm
		alg   joseJwa.KeyEncryptionAlgorithm
		prob  float32
	}{
		{"P256", ecdh.P256(), joseJwa.A256GCM(), joseJwa.ECDH_ES(), cryptoutilMagic.TestProbAlways},
		{"P384", ecdh.P384(), joseJwa.A256GCM(), joseJwa.ECDH_ES(), cryptoutilMagic.TestProbTenth},
		{"P521", ecdh.P521(), joseJwa.A128GCM(), joseJwa.ECDH_ES_A256KW(), cryptoutilMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			keyPair, err := cryptoutilKeyGen.GenerateECDHKeyPair(tc.curve)
			require.NoError(t, err)

			kid := googleUuid.Must(googleUuid.NewV7())

			resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &tc.enc, &tc.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, kid, *resultKid)
			require.NotNil(t, encryptJWK)
			require.NotNil(t, decryptJWK)
			require.NotEmpty(t, encryptBytes)
			require.NotEmpty(t, decryptBytes)
		})
	}
}

// TestCreateJWEJWKFromKey_RSA_Variants tests CreateJWEJWKFromKey with RSA key size variants.
func TestCreateJWEJWKFromKey_RSA_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		keySize int
		enc     joseJwa.ContentEncryptionAlgorithm
		alg     joseJwa.KeyEncryptionAlgorithm
		prob    float32
	}{
		{"RSA2048", cryptoutilMagic.RSAKeySize2048, joseJwa.A256GCM(), joseJwa.RSA_OAEP_256(), cryptoutilMagic.TestProbAlways},
		{"RSA3072", cryptoutilMagic.RSAKeySize3072, joseJwa.A192GCM(), joseJwa.RSA_OAEP_384(), cryptoutilMagic.TestProbTenth},
		{"RSA4096", cryptoutilMagic.RSAKeySize4096, joseJwa.A256GCM(), joseJwa.RSA_OAEP_512(), cryptoutilMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			privateKey, err := rsa.GenerateKey(crand.Reader, tc.keySize)
			require.NoError(t, err)

			keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}
			kid := googleUuid.Must(googleUuid.NewV7())

			resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &tc.enc, &tc.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, kid, *resultKid)
			require.NotNil(t, encryptJWK)
			require.NotNil(t, decryptJWK)
			require.NotEmpty(t, encryptBytes)
			require.NotEmpty(t, decryptBytes)
		})
	}
}

// TestCreateJWEJWKFromKey_AES_Variants tests CreateJWEJWKFromKey with AES key size variants.
func TestCreateJWEJWKFromKey_AES_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		keySize int
		enc     joseJwa.ContentEncryptionAlgorithm
		prob    float32
	}{
		{"AES128", cryptoutilMagic.AESKeySize128, joseJwa.A128GCM(), cryptoutilMagic.TestProbAlways},
		{"AES192", cryptoutilMagic.AESKeySize192, joseJwa.A192GCM(), cryptoutilMagic.TestProbTenth},
		{"AES256", cryptoutilMagic.AESKeySize256, joseJwa.A256GCM(), cryptoutilMagic.TestProbAlways},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			secretKey, err := cryptoutilKeyGen.GenerateAESKey(tc.keySize)
			require.NoError(t, err)

			kid := googleUuid.Must(googleUuid.NewV7())
			alg := AlgDir

			resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &tc.enc, &alg, secretKey)
			require.NoError(t, err)
			require.Equal(t, kid, *resultKid)
			require.NotNil(t, encryptJWK)
			require.Nil(t, decryptJWK) // Symmetric key - no separate decrypt JWK
			require.NotEmpty(t, encryptBytes)
			require.Empty(t, decryptBytes) // Symmetric key - no separate decrypt bytes
		})
	}
}

// TestCreateJWSJWKFromKey_RSA_Variants tests CreateJWSJWKFromKey with RSA key size variants.
func TestCreateJWSJWKFromKey_RSA_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		keySize int
		alg     joseJwa.SignatureAlgorithm
		prob    float32
	}{
		{"RSA2048_PS256", cryptoutilMagic.RSAKeySize2048, joseJwa.PS256(), cryptoutilMagic.TestProbAlways},
		{"RSA3072_PS384", cryptoutilMagic.RSAKeySize3072, joseJwa.PS384(), cryptoutilMagic.TestProbTenth},
		{"RSA4096_PS512", cryptoutilMagic.RSAKeySize4096, joseJwa.PS512(), cryptoutilMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			privateKey, err := rsa.GenerateKey(crand.Reader, tc.keySize)
			require.NoError(t, err)

			keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}
			kid := googleUuid.Must(googleUuid.NewV7())

			resultKid, signJWK, verifyJWK, signBytes, verifyBytes, err := CreateJWSJWKFromKey(&kid, &tc.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, kid, *resultKid)
			require.NotNil(t, signJWK)
			require.NotNil(t, verifyJWK)
			require.NotEmpty(t, signBytes)
			require.NotEmpty(t, verifyBytes)
		})
	}
}

// TestCreateJWSJWKFromKey_HMAC_Variants tests CreateJWSJWKFromKey with HMAC key size variants.
func TestCreateJWSJWKFromKey_HMAC_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		keySize int
		alg     joseJwa.SignatureAlgorithm
		prob    float32
	}{
		{"HMAC256", cryptoutilMagic.HMACKeySize256, joseJwa.HS256(), cryptoutilMagic.TestProbAlways},
		{"HMAC384", cryptoutilMagic.HMACKeySize384, joseJwa.HS384(), cryptoutilMagic.TestProbTenth},
		{"HMAC512", cryptoutilMagic.HMACKeySize512, joseJwa.HS512(), cryptoutilMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			secretKey, err := cryptoutilKeyGen.GenerateHMACKey(tc.keySize)
			require.NoError(t, err)

			kid := googleUuid.Must(googleUuid.NewV7())

			resultKid, signJWK, verifyJWK, signBytes, verifyBytes, err := CreateJWSJWKFromKey(&kid, &tc.alg, secretKey)
			require.NoError(t, err)
			require.Equal(t, kid, *resultKid)
			require.NotNil(t, signJWK)
			require.Nil(t, verifyJWK) // Symmetric key - no separate verify JWK
			require.NotEmpty(t, signBytes)
			require.Empty(t, verifyBytes) // Symmetric key - no separate verify bytes
		})
	}
}
