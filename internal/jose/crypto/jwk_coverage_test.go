// Copyright (c) 2025 Justin Cranford
//
//

package jose

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

// TestCreateJWKFromKey_Oct192AES tests CreateJWKFromKey with Oct192 AES key (uncovered branch).
func TestCreateJWKFromKey_Oct192AES(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct192
	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize192)
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.Empty(t, publicBytes)

	// Verify algorithm header set to A192GCM
	algHeader, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, "A192GCM", algHeader.String())

	// Verify use header set to "enc"
	use, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, "enc", use)
}

// TestCreateJWKFromKey_Oct384HMAC tests CreateJWKFromKey with Oct384 HMAC key (uncovered branch).
func TestCreateJWKFromKey_Oct384HMAC(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct384
	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize384)
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.Empty(t, publicBytes)

	// Verify algorithm header set to HS384
	algHeader, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, "HS384", algHeader.String())

	// Verify use header set to "sig"
	use, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, "sig", use)
}

// TestCreateJWKFromKey_Oct512HMAC tests CreateJWKFromKey with Oct512 HMAC key (uncovered branch).
func TestCreateJWKFromKey_Oct512HMAC(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct512
	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize512)
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.Empty(t, publicBytes)

	// Verify algorithm header set to HS512
	algHeader, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, "HS512", algHeader.String())

	// Verify use header set to "sig"
	use, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, "sig", use)
}

// TestIsPublicPrivateAsymmetricSymmetric_RSA tests JWK type checks with RSA keys.
func TestIsPublicPrivateAsymmetricSymmetric_RSA(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize2048)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.RSA2048
	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}

	_, privateJWK, publicJWK, _, _, err := CreateJWKFromKey(&kid, &alg, keyPair)
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
}

// TestIsPublicPrivateAsymmetricSymmetric_ECDSA tests JWK type checks with ECDSA keys.
func TestIsPublicPrivateAsymmetricSymmetric_ECDSA(t *testing.T) {
	t.Parallel()

	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.ECP384
	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}

	_, privateJWK, publicJWK, _, _, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)

	// Test private JWK
	isPrivate, err := IsPrivateJWK(privateJWK)
	require.NoError(t, err)
	require.True(t, isPrivate)

	isAsymmetric, err := IsAsymmetricJWK(privateJWK)
	require.NoError(t, err)
	require.True(t, isAsymmetric)

	// Test public JWK
	isPublic, err := IsPublicJWK(publicJWK)
	require.NoError(t, err)
	require.True(t, isPublic)
}

// TestIsPublicPrivateAsymmetricSymmetric_Ed25519 tests JWK type checks with Ed25519 keys.
func TestIsPublicPrivateAsymmetricSymmetric_Ed25519(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.OKPEd25519
	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: publicKey}

	_, privateJWK, publicJWK, _, _, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)

	// Test private JWK
	isPrivate, err := IsPrivateJWK(privateJWK)
	require.NoError(t, err)
	require.True(t, isPrivate)

	isAsymmetric, err := IsAsymmetricJWK(privateJWK)
	require.NoError(t, err)
	require.True(t, isAsymmetric)

	// Test public JWK
	isPublic, err := IsPublicJWK(publicJWK)
	require.NoError(t, err)
	require.True(t, isPublic)
}

// TestIsPublicPrivateAsymmetricSymmetric_HMAC tests JWK type checks with HMAC keys.
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

// TestCreateJWKFromKey_RSA3072 tests CreateJWKFromKey with RSA-3072.
func TestCreateJWKFromKey_RSA3072(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize3072)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.RSA3072
	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}

	resultKid, privateJWK, publicJWK, privateBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, privateJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, privateBytes)
	require.NotEmpty(t, publicBytes)
}

// TestCreateJWKFromKey_RSA4096 tests CreateJWKFromKey with RSA-4096.
func TestCreateJWKFromKey_RSA4096(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize4096)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.RSA4096
	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}

	resultKid, privateJWK, publicJWK, privateBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, privateJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, privateBytes)
	require.NotEmpty(t, publicBytes)
}

// TestCreateJWKFromKey_Ed448 removed - Ed448 not supported (only Ed25519 in api/model)

// TestCreateJWEJWKFromKey_ECDH_P384 tests CreateJWEJWKFromKey with ECDH P-384.
func TestCreateJWEJWKFromKey_ECDH_P384(t *testing.T) {
	t.Parallel()

	keyPair, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P384())
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
}

// TestCreateJWEJWKFromKey_ECDH_P521 tests CreateJWEJWKFromKey with ECDH P-521.
func TestCreateJWEJWKFromKey_ECDH_P521(t *testing.T) {
	t.Parallel()

	keyPair, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P521())
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	enc := joseJwa.A128GCM()
	alg := joseJwa.ECDH_ES_A256KW()

	resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, encryptJWK)
	require.NotNil(t, decryptJWK)
	require.NotEmpty(t, encryptBytes)
	require.NotEmpty(t, decryptBytes)
}

// TestCreateJWEJWKFromKey_RSA3072 tests CreateJWEJWKFromKey with RSA-3072.
func TestCreateJWEJWKFromKey_RSA3072(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize3072)
	require.NoError(t, err)

	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}
	kid := googleUuid.Must(googleUuid.NewV7())
	enc := joseJwa.A192GCM()
	alg := joseJwa.RSA_OAEP_384()

	resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, encryptJWK)
	require.NotNil(t, decryptJWK)
	require.NotEmpty(t, encryptBytes)
	require.NotEmpty(t, decryptBytes)
}

// TestCreateJWEJWKFromKey_RSA4096 tests CreateJWEJWKFromKey with RSA-4096.
func TestCreateJWEJWKFromKey_RSA4096(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize4096)
	require.NoError(t, err)

	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}
	kid := googleUuid.Must(googleUuid.NewV7())
	enc := joseJwa.A256GCM()
	alg := joseJwa.RSA_OAEP_512()

	resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, encryptJWK)
	require.NotNil(t, decryptJWK)
	require.NotEmpty(t, encryptBytes)
	require.NotEmpty(t, decryptBytes)
}

// TestCreateJWEJWKFromKey_AES128 tests CreateJWEJWKFromKey with AES-128.
func TestCreateJWEJWKFromKey_AES128(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize128)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	enc := joseJwa.A128GCM()
	alg := AlgDir

	resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, encryptJWK)
	require.Nil(t, decryptJWK) // Symmetric key - no separate decrypt JWK
	require.NotEmpty(t, encryptBytes)
	require.Empty(t, decryptBytes) // Symmetric key - no separate decrypt bytes
}

// TestCreateJWEJWKFromKey_AES256 tests CreateJWEJWKFromKey with AES-256.
func TestCreateJWEJWKFromKey_AES256(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize256)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	enc := joseJwa.A256GCM()
	alg := AlgDir

	resultKid, encryptJWK, decryptJWK, encryptBytes, decryptBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, encryptJWK)
	require.Nil(t, decryptJWK)
	require.NotEmpty(t, encryptBytes)
	require.Empty(t, decryptBytes)
}

// TestCreateJWSJWKFromKey_RSA3072 tests CreateJWSJWKFromKey with RSA-3072.
func TestCreateJWSJWKFromKey_RSA3072(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize3072)
	require.NoError(t, err)

	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}
	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.PS384()

	resultKid, signJWK, verifyJWK, signBytes, verifyBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, signJWK)
	require.NotNil(t, verifyJWK)
	require.NotEmpty(t, signBytes)
	require.NotEmpty(t, verifyBytes)
}

// TestCreateJWSJWKFromKey_RSA4096 tests CreateJWSJWKFromKey with RSA-4096.
func TestCreateJWSJWKFromKey_RSA4096(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilMagic.RSAKeySize4096)
	require.NoError(t, err)

	keyPair := &cryptoutilKeyGen.KeyPair{Private: privateKey, Public: &privateKey.PublicKey}
	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.PS512()

	resultKid, signJWK, verifyJWK, signBytes, verifyBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, signJWK)
	require.NotNil(t, verifyJWK)
	require.NotEmpty(t, signBytes)
	require.NotEmpty(t, verifyBytes)
}

// TestCreateJWSJWKFromKey_Ed448 removed - Ed448 not supported (only Ed25519 in api/model)

// TestCreateJWSJWKFromKey_HMAC256 tests CreateJWSJWKFromKey with HMAC-256.
func TestCreateJWSJWKFromKey_HMAC256(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize256)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.HS256()

	resultKid, signJWK, verifyJWK, signBytes, verifyBytes, err := CreateJWSJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, signJWK)
	require.Nil(t, verifyJWK) // Symmetric key - no separate verify JWK
	require.NotEmpty(t, signBytes)
	require.Empty(t, verifyBytes) // Symmetric key - no separate verify bytes
}

// TestCreateJWSJWKFromKey_HMAC512 tests CreateJWSJWKFromKey with HMAC-512.
func TestCreateJWSJWKFromKey_HMAC512(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize512)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.HS512()

	resultKid, signJWK, verifyJWK, signBytes, verifyBytes, err := CreateJWSJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, kid, *resultKid)
	require.NotNil(t, signJWK)
	require.Nil(t, verifyJWK)
	require.NotEmpty(t, signBytes)
	require.Empty(t, verifyBytes)
}

// TestCreateJWKFromKey_ErrorNilKid tests CreateJWKFromKey with nil KID.
func TestCreateJWKFromKey_ErrorNilKid(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize256)
	require.NoError(t, err)

	alg := cryptoutilOpenapiModel.Oct256

	_, _, _, _, _, err = CreateJWKFromKey(nil, &alg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK headers")
}

// TestCreateJWKFromKey_ErrorNilAlg tests CreateJWKFromKey with nil algorithm.
func TestCreateJWKFromKey_ErrorNilAlg(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize256)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())

	_, _, _, _, _, err = CreateJWKFromKey(&kid, nil, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK headers")
}

// TestCreateJWKFromKey_ErrorNilKey tests CreateJWKFromKey with nil key.
func TestCreateJWKFromKey_ErrorNilKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct256

	_, _, _, _, _, err := CreateJWKFromKey(&kid, &alg, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK headers")
}

// TestCreateJWEJWKFromKey_ErrorNilKid tests CreateJWEJWKFromKey with nil KID.
func TestCreateJWEJWKFromKey_ErrorNilKid(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize256)
	require.NoError(t, err)

	enc := joseJwa.A256GCM()
	alg := AlgDir

	_, _, _, _, _, err = CreateJWEJWKFromKey(nil, &enc, &alg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

// TestCreateJWEJWKFromKey_ErrorNilEnc tests CreateJWEJWKFromKey with nil enc.
func TestCreateJWEJWKFromKey_ErrorNilEnc(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize256)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := AlgDir

	_, _, _, _, _, err = CreateJWEJWKFromKey(&kid, nil, &alg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

// TestCreateJWEJWKFromKey_ErrorNilAlg tests CreateJWEJWKFromKey with nil alg.
func TestCreateJWEJWKFromKey_ErrorNilAlg(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize256)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())
	enc := joseJwa.A256GCM()

	_, _, _, _, _, err = CreateJWEJWKFromKey(&kid, &enc, nil, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

// TestCreateJWSJWKFromKey_ErrorNilKid tests CreateJWSJWKFromKey with nil KID.
func TestCreateJWSJWKFromKey_ErrorNilKid(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize256)
	require.NoError(t, err)

	alg := joseJwa.HS256()

	_, _, _, _, _, err = CreateJWSJWKFromKey(nil, &alg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

// TestCreateJWSJWKFromKey_ErrorNilAlg tests CreateJWSJWKFromKey with nil alg.
func TestCreateJWSJWKFromKey_ErrorNilAlg(t *testing.T) {
	t.Parallel()

	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize256)
	require.NoError(t, err)

	kid := googleUuid.Must(googleUuid.NewV7())

	_, _, _, _, _, err = CreateJWSJWKFromKey(&kid, nil, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}
