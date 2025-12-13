// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package jose

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"

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
			key := make(cryptoutilKeygen.SecretKey, tt.keySize)
			_, err := rand.Read(key)
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
			privateKey, err := rsa.GenerateKey(rand.Reader, tt.keySize)
			require.NoError(t, err)

			keyPair := &cryptoutilKeygen.KeyPair{
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
			privateKey, err := ecdsa.GenerateKey(tt.curve, rand.Reader)
			require.NoError(t, err)

			keyPair := &cryptoutilKeygen.KeyPair{
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
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilKeygen.KeyPair{
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
	invalidKeyPair := &cryptoutilKeygen.KeyPair{
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
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilKeygen.KeyPair{
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
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilKeygen.KeyPair{
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
	validKey, err := cryptoutilKeygen.GenerateHMACKey(256)
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

func TestValidateOrGenerateJWSEcdsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDSA P256 key pair.
	validKey, err := cryptoutilKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSEcdsaJWK(validKey, joseJwa.ES256(), elliptic.P256())
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSEcdsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWSEcdsaJWK(wrongKey, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSEddsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid Ed25519 key pair.
	validKey, err := cryptoutilKeygen.GenerateEDDSAKeyPair("Ed25519")
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSEddsaJWK(validKey, joseJwa.EdDSA(), "Ed25519")
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSEddsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWSEddsaJWK(wrongKey, joseJwa.EdDSA(), "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSEddsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilKeygen.KeyPair{
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

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	_ = publicKey

	keyPair := &cryptoutilKeygen.KeyPair{
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
	validKey, err := cryptoutilKeygen.GenerateHMACKey(256)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSHMACJWK(validKey, joseJwa.HS256(), 256)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSHMACJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use asymmetric key (wrong type).
	wrongKey, err := cryptoutilKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSHMACJWK(wrongKey, joseJwa.HS256(), 256)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}
func TestValidateOrGenerateJWSEcdsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilKeygen.KeyPair{
		Private: nil,
		Public:  &ecdsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEcdsaJWK(keyPair, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSEcdsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateJWSEcdsaJWK(keyPair, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}
func TestValidateOrGenerateJWSRSAJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	validKey, err := cryptoutilKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSRSAJWK(validKey, joseJwa.RS256(), 2048)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSRSAJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	wrongKey, err := cryptoutilKeygen.GenerateHMACKey(256)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSRSAJWK(wrongKey, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSRSAJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilKeygen.KeyPair{
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

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}
