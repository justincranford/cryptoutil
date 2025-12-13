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
		{"HS256", joseJwa.HS256, 32},
		{"HS384", joseJwa.HS384, 48},
		{"HS512", joseJwa.HS512, 64},
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
			require.Equal(t, kid.String(), nonPublicJWK.KeyID())
			require.Equal(t, tt.alg, nonPublicJWK.Algorithm())
			require.Equal(t, joseJwk.OctetSeq, nonPublicJWK.KeyType())
			require.Equal(t, joseJwk.ForSignature.String(), nonPublicJWK.KeyUsage())
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
		{"RS256", joseJwa.RS256, 2048},
		{"RS384", joseJwa.RS384, 2048},
		{"RS512", joseJwa.RS512, 2048},
		{"PS256", joseJwa.PS256, 2048},
		{"PS384", joseJwa.PS384, 2048},
		{"PS512", joseJwa.PS512, 2048},
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
			require.Equal(t, kid.String(), nonPublicJWK.KeyID())
			require.Equal(t, tt.alg, nonPublicJWK.Algorithm())
			require.Equal(t, joseJwk.RSA, nonPublicJWK.KeyType())
			require.Equal(t, joseJwk.ForSignature.String(), nonPublicJWK.KeyUsage())
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
		{"ES256", joseJwa.ES256, elliptic.P256()},
		{"ES384", joseJwa.ES384, elliptic.P384()},
		{"ES512", joseJwa.ES512, elliptic.P521()},
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
			require.Equal(t, kid.String(), nonPublicJWK.KeyID())
			require.Equal(t, tt.alg, nonPublicJWK.Algorithm())
			require.Equal(t, joseJwk.EC, nonPublicJWK.KeyType())
			require.Equal(t, joseJwk.ForSignature.String(), nonPublicJWK.KeyUsage())
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

	alg := joseJwa.EdDSA

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify headers
	require.Equal(t, kid.String(), nonPublicJWK.KeyID())
	require.Equal(t, alg, nonPublicJWK.Algorithm())
	require.Equal(t, joseJwk.OKP, nonPublicJWK.KeyType())
	require.Equal(t, joseJwk.ForSignature.String(), nonPublicJWK.KeyUsage())
}

// TestCreateJWSJWKFromKey_UnsupportedKeyType tests error for unsupported key types.
func TestCreateJWSJWKFromKey_UnsupportedKeyType(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.RS256
	invalidKey := "not-a-valid-key"

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, invalidKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported key type")
}

// TestCreateJWSJWKFromKey_NilKid tests error for nil KID.
func TestCreateJWSJWKFromKey_NilKid(t *testing.T) {
	t.Parallel()

	alg := joseJwa.RS256
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
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilKeygen.KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}

	_, _, _, _, _, err = CreateJWSJWKFromKey(&kid, nil, keyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

// TestCreateJWSJWKFromKey_NilKey tests error for nil key.
func TestCreateJWSJWKFromKey_NilKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.RS256

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}
