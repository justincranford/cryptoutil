// Copyright (c) 2025 Justin Cranford

package crypto

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// TestCreateJWSJWKFromKey_HMAC tests HMAC key import.
func TestCreateJWSJWKFromKey_HMAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     joseJwa.SignatureAlgorithm
		keySize int
	}{
		{"HS256", joseJwa.HS256(), 32},
		{"HS384", joseJwa.HS384(), 48},
		{"HS512", joseJwa.HS512(), 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			key := make(cryptoutilSharedCryptoKeygen.SecretKey, tt.keySize)
			_, err := crand.Read(key)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, key)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.Nil(t, publicJWK) // HMAC has no public key
			require.NotEmpty(t, nonPublicBytes)
			require.Empty(t, publicBytes)

			// Verify headers
			keyID, ok := nonPublicJWK.KeyID()
			require.True(t, ok)
			require.Equal(t, kid.String(), keyID)

			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyOCT, nonPublicJWK.KeyType())

			usage, ok := nonPublicJWK.KeyUsage()
			require.True(t, ok)
			require.Equal(t, joseJwk.ForSignature.String(), usage)
		})
	}
}

// TestCreateJWSJWKFromKey_RSA tests RSA key import.
func TestCreateJWSJWKFromKey_RSA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     joseJwa.SignatureAlgorithm
		keySize int
	}{
		{"RS256", joseJwa.RS256(), 2048},
		{"RS384", joseJwa.RS384(), 2048},
		{"RS512", joseJwa.RS512(), 2048},
		{"PS256", joseJwa.PS256(), 2048},
		{"PS384", joseJwa.PS384(), 2048},
		{"PS512", joseJwa.PS512(), 2048},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			privateKey, err := rsa.GenerateKey(crand.Reader, tt.keySize)
			require.NoError(t, err)

			keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
				Private: privateKey,
				Public:  &privateKey.PublicKey,
			}

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			// Verify headers
			keyID, ok := nonPublicJWK.KeyID()
			require.True(t, ok)
			require.Equal(t, kid.String(), keyID)

			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyRSA, nonPublicJWK.KeyType())

			usage, ok := nonPublicJWK.KeyUsage()
			require.True(t, ok)
			require.Equal(t, joseJwk.ForSignature.String(), usage)
		})
	}
}

// TestCreateJWSJWKFromKey_ECDSA tests ECDSA key import.
func TestCreateJWSJWKFromKey_ECDSA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		alg   joseJwa.SignatureAlgorithm
		curve elliptic.Curve
	}{
		{"ES256", joseJwa.ES256(), elliptic.P256()},
		{"ES384", joseJwa.ES384(), elliptic.P384()},
		{"ES512", joseJwa.ES512(), elliptic.P521()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			privateKey, err := ecdsa.GenerateKey(tt.curve, crand.Reader)
			require.NoError(t, err)

			keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
				Private: privateKey,
				Public:  &privateKey.PublicKey,
			}

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			// Verify headers
			keyID, ok := nonPublicJWK.KeyID()
			require.True(t, ok)
			require.Equal(t, kid.String(), keyID)

			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyEC, nonPublicJWK.KeyType())

			usage, ok := nonPublicJWK.KeyUsage()
			require.True(t, ok)
			require.Equal(t, joseJwk.ForSignature.String(), usage)
		})
	}
}

// TestCreateJWSJWKFromKey_EdDSA tests EdDSA key import.
func TestCreateJWSJWKFromKey_EdDSA(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	_, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  privateKey.Public(),
	}

	alg := joseJwa.EdDSA()

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify headers
	keyID, ok := nonPublicJWK.KeyID()
	require.True(t, ok)
	require.Equal(t, kid.String(), keyID)

	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, alg, algVal)

	require.Equal(t, KtyOKP, nonPublicJWK.KeyType())

	usage, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, joseJwk.ForSignature.String(), usage)
}

// TestCreateJWSJWKFromKey_UnsupportedKeyType tests error for unsupported key types.
func TestCreateJWSJWKFromKey_UnsupportedKeyType(t *testing.T) {
	t.Parallel()

	// Test with KeyPair containing unsupported key type (int instead of crypto key)
	kid := googleUuid.New()
	alg := joseJwa.RS256()

	// KeyPair with invalid Private field type will hit default case in switch
	invalidKeyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: 12345, // Invalid type - not *rsa.PrivateKey, *ecdsa.PrivateKey, or ed25519.PrivateKey
		Public:  nil,
	}

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, invalidKeyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestCreateJWSJWKFromKey_NilKid tests error for nil KID.
func TestCreateJWSJWKFromKey_NilKid(t *testing.T) {
	t.Parallel()

	alg := joseJwa.RS256()
	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}

	_, _, _, _, _, err = CreateJWSJWKFromKey(nil, &alg, keyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

// TestCreateJWSJWKFromKey_NilAlg tests error for nil algorithm.
func TestCreateJWSJWKFromKey_NilAlg(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()

	// Use simple KeyPair with valid key type to test nil alg validation
	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}

	// Should error on nil alg validation before key type checks
	_, _, _, _, _, err = CreateJWSJWKFromKey(&kid, nil, keyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

func TestCreateJWSJWKFromKey_SetKidError(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.HS256()
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(256)
	require.NoError(t, err)

	_, _, _, _, _, err = CreateJWSJWKFromKey(&kid, &alg, validKey)
	require.NoError(t, err)
}

// TestCreateJWSJWKFromKey_NilKey tests error for nil key.
func TestCreateJWSJWKFromKey_NilKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.RS256()

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

func TestCreateJWSJWKFromKey_PublicKeyExtraction(t *testing.T) {
	t.Parallel()

	// Generate RSA key pair.
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	kid := googleUuid.New()
	alg := joseJwa.RS256()

	// Create JWK from key pair.
	_, nonPublicJWK, publicJWK, _, clearPublicBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, clearPublicBytes)

	// Verify public key extracted successfully.
	var kidFromPublic string

	require.NoError(t, publicJWK.Get(joseJwk.KeyIDKey, &kidFromPublic))
	require.Equal(t, kid.String(), kidFromPublic)
}

func TestCreateJWSJWKFromKey_HMACNoPublicKey(t *testing.T) {
	t.Parallel()

	// HMAC has no public key, so publicJWK should be nil and clearPublicBytes empty.
	kid := googleUuid.New()
	alg := joseJwa.HS256()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, 32)
	_, _ = crand.Read(key)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK) // HMAC should have no public key
	require.Empty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)
}

// TestCreateJWSJWKFromKey_ImportSecretKeyError tests validation error for empty HMAC key.
func TestCreateJWSJWKFromKey_ImportSecretKeyError(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.HS256()
	// Empty SecretKey fails validation before import
	emptyKey := cryptoutilSharedCryptoKeygen.SecretKey("")

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, emptyKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

// TestCreateJWSJWKFromKey_ImportKeyPairError tests validation error for nil KeyPair.
func TestCreateJWSJWKFromKey_ImportKeyPairError(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.ES256()
	// KeyPair with nil Private fails validation before import
	invalidKeyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  nil,
	}

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, invalidKeyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

func TestValidateOrGenerateJWSEcdsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDSA P256 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSEcdsaJWK(validKey, joseJwa.ES256(), elliptic.P256())
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSEcdsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWSEcdsaJWK(wrongKey, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSEcdsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	// KeyPair with nil private key.
	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: (*ecdsa.PrivateKey)(nil),
		Public:  &ecdsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEcdsaJWK(keyPair, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil ECDSA private key")
}

func TestValidateOrGenerateJWSEcdsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDSA P256 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	// Create KeyPair with valid private and nil public.
	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: validKey.Private,
		Public:  (*ecdsa.PublicKey)(nil),
	}

	validated, err := validateOrGenerateJWSEcdsaJWK(keyPair, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil ECDSA public key")
}

func TestValidateOrGenerateJWSEddsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid Ed25519 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair("Ed25519")
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSEddsaJWK(validKey, joseJwa.EdDSA(), "Ed25519")
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSEddsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWSEddsaJWK(wrongKey, joseJwa.EdDSA(), "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSEddsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  ed25519.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSEddsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	_ = publicKey

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSHMACJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid HMAC 256 key.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(256)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSHMACJWK(validKey, joseJwa.HS256(), 256)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSHMACJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use asymmetric key (wrong type).
	wrongKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSHMACJWK(wrongKey, joseJwa.HS256(), 256)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSHMACJWK_NilSecretKey(t *testing.T) {
	t.Parallel()

	// SecretKey with nil value.
	var nilKey cryptoutilSharedCryptoKeygen.SecretKey

	result, err := validateOrGenerateJWSHMACJWK(nilKey, joseJwa.HS256(), 256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil key bytes")
}

func TestValidateOrGenerateJWSHMACJWK_WrongKeyLength(t *testing.T) {
	t.Parallel()

	// Generate 512-bit key but expect 256-bit.
	wrongLengthKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(512)
	require.NoError(t, err)

	result, err := validateOrGenerateJWSHMACJWK(wrongLengthKey, joseJwa.HS256(), 256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid key length")
}

func TestValidateOrGenerateJWSRSAJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	validKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSRSAJWK(validKey, joseJwa.RS256(), 2048)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSRSAJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	wrongKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(256)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSRSAJWK(wrongKey, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSRSAJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  &rsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSRSAJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestCreateJWSJWKFromKey_ECDSA_AllCurves tests ECDSA with all curves.
func TestCreateJWSJWKFromKey_ECDSA_AllCurves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		alg   joseJwa.SignatureAlgorithm
		curve elliptic.Curve
	}{
		{"ES256_P256", joseJwa.ES256(), elliptic.P256()},
		{"ES384_P384", joseJwa.ES384(), elliptic.P384()},
		{"ES512_P521", joseJwa.ES512(), elliptic.P521()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(tt.curve)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			// Verify algorithm
			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyEC, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWSJWKFromKey_EdDSA_Ed25519 tests EdDSA Ed25519 key creation.
func TestCreateJWSJWKFromKey_EdDSA_Ed25519(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.EdDSA()
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair("Ed25519")
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify algorithm
	algValue, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, alg, algValue)

	require.Equal(t, KtyOKP, nonPublicJWK.KeyType())
}

// TestCreateJWSJWKFromKey_ErrorCases tests error handling.
func TestCreateJWSJWKFromKey_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("NilKid", func(t *testing.T) {
		t.Parallel()

		alg := joseJwa.HS256()
		key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(nil, &alg, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWS JWK kid must be valid")
	})

	t.Run("NilAlg", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, nil, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWS JWK alg must be non-nil")
	})

	t.Run("NilKey", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		alg := joseJwa.HS256()

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, nil)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWS JWK key material must be non-nil")
	})
}

// TestCreateJWSJWKFromKey_HMAC_AllSizes tests all HMAC sizes.
func TestCreateJWSJWKFromKey_HMAC_AllSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     joseJwa.SignatureAlgorithm
		keySize int
	}{
		{"HS256_256bit", joseJwa.HS256(), 256},
		{"HS384_384bit", joseJwa.HS384(), 384},
		{"HS512_512bit", joseJwa.HS512(), 512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(tt.keySize)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, key)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.Nil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.Empty(t, publicBytes)

			// Verify algorithm
			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyOCT, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWSJWKFromKey_RSA_AllSizes tests all RSA key sizes and algorithms.
func TestCreateJWSJWKFromKey_RSA_AllSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     joseJwa.SignatureAlgorithm
		keySize int
	}{
		{"RS256_2048", joseJwa.RS256(), 2048},
		{"RS384_3072", joseJwa.RS384(), 3072},
		{"RS512_4096", joseJwa.RS512(), 4096},
		{"PS256_2048", joseJwa.PS256(), 2048},
		{"PS384_3072", joseJwa.PS384(), 3072},
		{"PS512_4096", joseJwa.PS512(), 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(tt.keySize)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			// Verify algorithm
			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyRSA, nonPublicJWK.KeyType())
		})
	}
}
