// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdh"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

func TestCreateJWEJWKFromKey_RSA_AllAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enc     joseJwa.ContentEncryptionAlgorithm
		alg     joseJwa.KeyEncryptionAlgorithm
		keySize int
	}{
		{"RSA_OAEP_256_A256GCM", joseJwa.A256GCM(), joseJwa.RSA_OAEP_256(), 2048},
		{"RSA_OAEP_384_A256GCM", joseJwa.A256GCM(), joseJwa.RSA_OAEP_384(), 3072},
		{"RSA_OAEP_512_A256GCM", joseJwa.A256GCM(), joseJwa.RSA_OAEP_512(), 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(tt.keySize)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &tt.enc, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			require.Equal(t, KtyRSA, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWEJWKFromKey_ECDH_AllCurves tests ECDH with all curves.
func TestCreateJWEJWKFromKey_ECDH_AllCurves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		enc   joseJwa.ContentEncryptionAlgorithm
		alg   joseJwa.KeyEncryptionAlgorithm
		curve ecdh.Curve
	}{
		{"ECDH_ES_P256", joseJwa.A256GCM(), joseJwa.ECDH_ES(), ecdh.P256()},
		{"ECDH_ES_P384", joseJwa.A256GCM(), joseJwa.ECDH_ES(), ecdh.P384()},
		{"ECDH_ES_P521", joseJwa.A256GCM(), joseJwa.ECDH_ES(), ecdh.P521()},
		{"ECDH_ES_A256KW_P256", joseJwa.A256GCM(), joseJwa.ECDH_ES_A256KW(), ecdh.P256()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDHKeyPair(tt.curve)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &tt.enc, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			require.Equal(t, KtyEC, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWEJWKFromKey_AES_AllSizes tests AES secret key with all sizes.
func TestCreateJWEJWKFromKey_AES_AllSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enc     joseJwa.ContentEncryptionAlgorithm
		alg     joseJwa.KeyEncryptionAlgorithm
		keySize int
	}{
		{"A128KW_A128GCM", joseJwa.A128GCM(), joseJwa.A128KW(), 128},
		{"A192KW_A192GCM", joseJwa.A192GCM(), joseJwa.A192KW(), 192},
		{"A256KW_A256GCM", joseJwa.A256GCM(), joseJwa.A256KW(), 256},
		{"dir_A256GCM", joseJwa.A256GCM(), joseJwa.DIRECT(), 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(tt.keySize)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &tt.enc, &tt.alg, secretKey)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.Nil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.Empty(t, publicBytes)

			require.Equal(t, KtyOCT, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWEJWKFromKey_ErrorCases tests comprehensive error handling.
func TestCreateJWEJWKFromKey_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("NilKid", func(t *testing.T) {
		t.Parallel()

		enc := joseJwa.A256GCM()
		alg := joseJwa.A256KW()
		key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(nil, &enc, &alg, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK kid must be valid")
	})

	t.Run("NilEnc", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		alg := joseJwa.A256KW()
		key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, nil, &alg, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK enc must be non-nil")
	})

	t.Run("NilAlg", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		enc := joseJwa.A256GCM()
		key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &enc, nil, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK alg must be non-nil")
	})

	t.Run("NilKey", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		enc := joseJwa.A256GCM()
		alg := joseJwa.A256KW()

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, nil)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK key must be non-nil")
	})
}
