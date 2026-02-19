// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"crypto/elliptic"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

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
