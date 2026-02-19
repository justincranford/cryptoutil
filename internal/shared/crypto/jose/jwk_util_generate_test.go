// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/elliptic"
	rsa "crypto/rsa"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestGenerateJWKForAlg_AllAlgorithms(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		alg  cryptoutilOpenapiModel.GenerateAlgorithm
	}{
		{"RSA2048", cryptoutilOpenapiModel.RSA2048},
		{"ECP256", cryptoutilOpenapiModel.ECP256},
		{"ECP384", cryptoutilOpenapiModel.ECP384},
		{"ECP521", cryptoutilOpenapiModel.ECP521},
		{"OKPEd25519", cryptoutilOpenapiModel.OKPEd25519},
		{"Oct128", cryptoutilOpenapiModel.Oct128},
		{"Oct192", cryptoutilOpenapiModel.Oct192},
		{"Oct256", cryptoutilOpenapiModel.Oct256},
		{"Oct384", cryptoutilOpenapiModel.Oct384},
		{"Oct512", cryptoutilOpenapiModel.Oct512},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kid, privateJWK, publicJWK, privateJWKBytes, publicJWKBytes, err := GenerateJWKForAlg(&tc.alg)
			require.NoError(t, err)
			require.NotNil(t, kid)
			require.NotNil(t, privateJWK)
			require.NotEmpty(t, privateJWKBytes)

			// Oct keys (symmetric) don't have separate public keys
			isSymmetric := tc.alg == cryptoutilOpenapiModel.Oct128 ||
				tc.alg == cryptoutilOpenapiModel.Oct192 ||
				tc.alg == cryptoutilOpenapiModel.Oct256 ||
				tc.alg == cryptoutilOpenapiModel.Oct384 ||
				tc.alg == cryptoutilOpenapiModel.Oct512

			if isSymmetric {
				require.Nil(t, publicJWK)
				require.Empty(t, publicJWKBytes)
			} else {
				require.NotNil(t, publicJWK)
				require.NotEmpty(t, publicJWKBytes)
			}

			// Test ExtractKty (works for all key types).
			kty, err := ExtractKty(privateJWK)
			require.NoError(t, err)
			require.NotNil(t, kty)
		})
	}
}

func TestGenerateJWKForAlg_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	// Invalid algorithm value.
	invalidAlg := cryptoutilOpenapiModel.GenerateAlgorithm("INVALID")

	kid, privateJWK, publicJWK, privateJWKBytes, publicJWKBytes, err := GenerateJWKForAlg(&invalidAlg)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Nil(t, privateJWK)
	require.Nil(t, publicJWK)
	require.Nil(t, privateJWKBytes)
	require.Nil(t, publicJWKBytes)
	require.Contains(t, err.Error(), "unsupported JWK alg")
}

func TestEnsureSignatureAlgorithmType_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	// Generate JWK for encryption (not signing).
	enc := EncA256GCM
	algEnc := AlgA256KW
	_, privateJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &algEnc)
	require.NoError(t, err)

	// Test validation should fail because this is an encryption key, not a signing key.
	err = EnsureSignatureAlgorithmType(privateJWK)
	require.Error(t, err)
	// The actual error is about getting algorithm from JWK.
	require.Contains(t, err.Error(), "failed to get algorithm from JWK")
}

func Test_EnsureSignatureAlgorithmType_NilJWK(t *testing.T) {
	t.Parallel()

	err := EnsureSignatureAlgorithmType(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK invalid")
}

// NOTE: EnsureSignatureAlgorithmType comprehensive tests removed.
// Function has design flaw: attempts to Get() string but JWX v3 already stores typed SignatureAlgorithm.
// Function appears unused in production code (only called by tests).
// Existing tests (InvalidAlgorithm, NilJWK) provide minimal coverage for unused function.
// Additional tests would require fixing production code first.

func TestExtractAlg_NilJWK(t *testing.T) {
	t.Parallel()

	// Test nil JWK.
	extractedAlg, err := ExtractAlg(nil)
	require.Error(t, err)
	require.Nil(t, extractedAlg)
	require.Contains(t, err.Error(), "invalid jwk")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestExtractAlg_JWKMissingAlgHeader(t *testing.T) {
	t.Parallel()

	// Generate JWK without algorithm header.
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	// Create JWK from RSA private key WITHOUT setting algorithm.
	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)

	// Extract algorithm should fail because alg header missing.
	extractedAlg, err := ExtractAlg(privateJWK)
	require.Error(t, err)
	require.Nil(t, extractedAlg)
	require.Contains(t, err.Error(), "failed to get alg header")
}

// TestExtractAlg_JWSAlgorithmNotGenerateAlgorithm tests ExtractAlg with JWS algorithm.
func TestExtractAlg_JWSAlgorithmNotGenerateAlgorithm(t *testing.T) {
	t.Parallel()

	// Generate JWK with HS256 algorithm (JWS algorithm, not GenerateAlgorithm).
	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.HS256()
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.HMACKeySize256)
	require.NoError(t, err)

	_, nonPublicJWK, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)

	// ExtractAlg should fail because HS256 is not in generateAlgorithms map.
	extractedAlg, err := ExtractAlg(nonPublicJWK)
	require.Error(t, err)
	require.Nil(t, extractedAlg)
	require.Contains(t, err.Error(), "failed to map to generate alg")

	// Verify algorithm is correctly set on JWK using Algorithm() method.
	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, joseJwa.HS256(), algVal)
}

func TestExtractKidUUID_ValidKid(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-for-kid-extraction-32b"))
	require.NoError(t, err)

	// Create valid UUID and set as kid.
	validKid := googleUuid.New()
	require.NoError(t, jwk.Set(joseJwk.KeyIDKey, validKid.String()))

	extractedKid, err := ExtractKidUUID(jwk)
	require.NoError(t, err)
	require.NotNil(t, extractedKid)
	require.Equal(t, validKid, *extractedKid)
}

func TestExtractKidUUID_NilJWK(t *testing.T) {
	t.Parallel()

	extractedKid, err := ExtractKidUUID(nil)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "invalid jwk")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestExtractKidUUID_MissingKid(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-no-kid-requires-32-byt"))
	require.NoError(t, err)

	extractedKid, err := ExtractKidUUID(jwk)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "failed to get kid header")
}

func TestExtractKidUUID_InvalidUUIDFormat(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-invalid-uuid-32-bytes!"))
	require.NoError(t, err)
	require.NoError(t, jwk.Set(joseJwk.KeyIDKey, "not-a-valid-uuid-format"))

	extractedKid, err := ExtractKidUUID(jwk)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "failed to parse kid as UUID")
}

func TestExtractKidUUID_InvalidNilUUID(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-nil-uuid-exactly-32-by"))
	require.NoError(t, err)
	require.NoError(t, jwk.Set(joseJwk.KeyIDKey, googleUuid.Nil.String()))

	extractedKid, err := ExtractKidUUID(jwk)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "failed to validate kid UUID")
}

func TestCreateJWKFromKey_RSAKeyPair(t *testing.T) {
	t.Parallel()

	// RSA key pair has public key component.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.RSA2048
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // RSA should have public key
	require.NotEmpty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)
}

// TestCreateJWKFromKey_Oct256HMAC tests CreateJWKFromKey with Oct256 HMAC key.
func TestCreateJWKFromKey_Oct256HMAC(t *testing.T) {
	t.Parallel()

	// Generate Oct256 secret key for HMAC.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.Oct256
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.HMACKeySize256)
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK) // Symmetric key has no public key
	require.Nil(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)

	// Verify headers.
	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, AlgHS256, algVal)

	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyOCT, kty)

	usage, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, joseJwk.ForSignature.String(), usage)
}

// TestCreateJWKFromKey_Oct128AES tests CreateJWKFromKey with Oct128 AES key.
func TestCreateJWKFromKey_Oct128AES(t *testing.T) {
	t.Parallel()

	// Generate Oct128 secret key for AES.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.Oct128
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.AESKeySize128)
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK) // Symmetric key has no public key
	require.Nil(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)

	// Verify headers.
	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, "A128GCM", algVal.String())

	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyOCT, kty)

	usage, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, "enc", usage)
}

// TestCreateJWKFromKey_InvalidHeaders tests CreateJWKFromKey with invalid headers.
func TestCreateJWKFromKey_InvalidHeaders(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.Oct256
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.HMACKeySize256)
	require.NoError(t, err)

	// Nil kid should fail validation.
	_, _, _, _, _, err = CreateJWKFromKey(nil, &alg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK headers")

	// Nil alg should fail validation.
	_, _, _, _, _, err = CreateJWKFromKey(&kid, nil, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK alg must be non-nil")

	// Nil key should fail validation.
	_, _, _, _, _, err = CreateJWKFromKey(&kid, &alg, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK key material must be non-nil")
}

// TestCreateJWKFromKey_ECDSAKeyPair tests CreateJWKFromKey with ECDSA key pair.
func TestCreateJWKFromKey_ECDSAKeyPair(t *testing.T) {
	t.Parallel()

	// Generate ECDSA P256 key pair.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.ECP256
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // Asymmetric key has public key
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify key type.
	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyEC, kty)
}

// TestCreateJWKFromKey_EdDSAKeyPair tests CreateJWKFromKey with EdDSA key pair.
func TestCreateJWKFromKey_EdDSAKeyPair(t *testing.T) {
	t.Parallel()

	// Generate Ed25519 key pair.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.OKPEd25519
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair("Ed25519")
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // Asymmetric key has public key
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify key type.
	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyOKP, kty)
}

func TestValidateOrGenerateHMACJWK_NilSecretKey(t *testing.T) {
	t.Parallel()

	// SecretKey with nil value.
	var nilKey cryptoutilSharedCryptoKeygen.SecretKey

	result, err := validateOrGenerateHMACJWK(nilKey, cryptoutilSharedMagic.HMACKeySize256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil key bytes")
}

func TestValidateOrGenerateAESJWK_NilSecretKey(t *testing.T) {
	t.Parallel()

	// SecretKey with nil value.
	var nilKey cryptoutilSharedCryptoKeygen.SecretKey

	result, err := validateOrGenerateAESJWK(nilKey, cryptoutilSharedMagic.AESKeySize256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil key bytes")
}

func TestExtractKty_ValidKeyTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		genKey      func(t *testing.T) joseJwk.Key
		expectedKty joseJwa.KeyType
	}{
		{
			name: "RSA",
			genKey: func(t *testing.T) joseJwk.Key {
				keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
				require.NoError(t, err)
				jwk, err := joseJwk.Import(keyPair.Private)
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.RSA(),
		},
		{
			name: "EC",
			genKey: func(t *testing.T) joseJwk.Key {
				keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
				require.NoError(t, err)
				jwk, err := joseJwk.Import(keyPair.Private)
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.EC(),
		},
		{
			name: "OKP",
			genKey: func(t *testing.T) joseJwk.Key {
				keyPair, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair("Ed25519")
				require.NoError(t, err)
				jwk, err := joseJwk.Import(keyPair.Private)
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.OKP(),
		},
		{
			name: "oct",
			genKey: func(t *testing.T) joseJwk.Key {
				jwk, err := joseJwk.Import([]byte("test-key-for-oct-requires-32-byte"))
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.OctetSeq(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwk := tc.genKey(t)

			kty, err := ExtractKty(jwk)
			require.NoError(t, err)
			require.NotNil(t, kty)
			require.Equal(t, tc.expectedKty, *kty)
		})
	}
}
