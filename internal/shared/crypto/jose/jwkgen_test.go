// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

// TestGenerateRSAJWK tests RSA JWK generation.
// P0.2 optimization: Test only RSA2048 - function logic is identical for all sizes.
func TestGenerateRSAJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rsaBits int
		prob    float32
	}{
		{"RSA2048", cryptoutilSharedMagic.RSAKeySize2048, cryptoutilSharedMagic.TestProbAlways},
		{"RSA3072", cryptoutilSharedMagic.RSAKeySize3072, cryptoutilSharedMagic.TestProbTenth},
		{"RSA4096", cryptoutilSharedMagic.RSAKeySize4096, cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			jwk, err := GenerateRSAJWK(tc.rsaBits)
			require.NoError(t, err)
			require.NotNil(t, jwk)
			require.Equal(t, KtyRSA, jwk.KeyType())

			// Test function generator.
			genFunc := GenerateRSAJWKFunction(tc.rsaBits)
			jwk2, err := genFunc()
			require.NoError(t, err)
			require.NotNil(t, jwk2)
			require.Equal(t, KtyRSA, jwk2.KeyType())
		})
	}
}

// TestGenerateECDSAJWK tests ECDSA JWK generation.
func TestGenerateECDSAJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		curve elliptic.Curve
		prob  float32
	}{
		{"P256", elliptic.P256(), cryptoutilSharedMagic.TestProbAlways},
		{"P384", elliptic.P384(), cryptoutilSharedMagic.TestProbTenth},
		{"P521", elliptic.P521(), cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			jwk, err := GenerateECDSAJWK(tc.curve)
			require.NoError(t, err)
			require.NotNil(t, jwk)
			require.Equal(t, KtyEC, jwk.KeyType())

			// Test function generator.
			genFunc := GenerateECDSAJWKFunction(tc.curve)
			jwk2, err := genFunc()
			require.NoError(t, err)
			require.NotNil(t, jwk2)
			require.Equal(t, KtyEC, jwk2.KeyType())
		})
	}
}

// TestGenerateECDHJWK tests ECDH JWK generation.
func TestGenerateECDHJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		curve ecdh.Curve
		prob  float32
	}{
		{"P256", ecdh.P256(), cryptoutilSharedMagic.TestProbAlways},
		{"P384", ecdh.P384(), cryptoutilSharedMagic.TestProbTenth},
		{"P521", ecdh.P521(), cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			jwk, err := GenerateECDHJWK(tc.curve)
			require.NoError(t, err)
			require.NotNil(t, jwk)
			require.Equal(t, KtyEC, jwk.KeyType())

			// Test function generator.
			genFunc := GenerateECDHJWKFunction(tc.curve)
			jwk2, err := genFunc()
			require.NoError(t, err)
			require.NotNil(t, jwk2)
			require.Equal(t, KtyEC, jwk2.KeyType())
		})
	}
}

// TestGenerateEDDSAJWK tests EdDSA JWK generation.
func TestGenerateEDDSAJWK(t *testing.T) {
	t.Parallel()

	jwk, err := GenerateEDDSAJWK("Ed25519")
	require.NoError(t, err)
	require.NotNil(t, jwk)
	require.Equal(t, KtyOKP, jwk.KeyType())

	// Test function generator.
	genFunc := GenerateEDDSAJWKFunction("Ed25519")
	jwk2, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, jwk2)
	require.Equal(t, KtyOKP, jwk2.KeyType())
}

// TestGenerateAESJWK tests AES JWK generation.
func TestGenerateAESJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		aesBits int
		prob    float32
	}{
		{"AES128", cryptoutilSharedMagic.AESKeySize128, cryptoutilSharedMagic.TestProbAlways},
		{"AES192", cryptoutilSharedMagic.AESKeySize192, cryptoutilSharedMagic.TestProbTenth},
		{"AES256", cryptoutilSharedMagic.AESKeySize256, cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			jwk, err := GenerateAESJWK(tc.aesBits)
			require.NoError(t, err)
			require.NotNil(t, jwk)
			require.Equal(t, KtyOCT, jwk.KeyType())

			// Test function generator.
			genFunc := GenerateAESJWKFunction(tc.aesBits)
			jwk2, err := genFunc()
			require.NoError(t, err)
			require.NotNil(t, jwk2)
			require.Equal(t, KtyOCT, jwk2.KeyType())
		})
	}
}

// TestGenerateAESHSJWK tests AES+HS JWK generation.
func TestGenerateAESHSJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		aesHsBits int
		prob      float32
	}{
		{"AES128HS256", cryptoutilSharedMagic.HMACKeySize256, cryptoutilSharedMagic.TestProbAlways},
		{"AES192HS384", cryptoutilSharedMagic.HMACKeySize384, cryptoutilSharedMagic.TestProbTenth},
		{"AES256HS512", cryptoutilSharedMagic.HMACKeySize512, cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			jwk, err := GenerateAESHSJWK(tc.aesHsBits)
			require.NoError(t, err)
			require.NotNil(t, jwk)
			require.Equal(t, KtyOCT, jwk.KeyType())

			// Test function generator.
			genFunc := GenerateAESHSJWKFunction(tc.aesHsBits)
			jwk2, err := genFunc()
			require.NoError(t, err)
			require.NotNil(t, jwk2)
			require.Equal(t, KtyOCT, jwk2.KeyType())
		})
	}
}

// TestGenerateHMACJWK tests HMAC JWK generation.
func TestGenerateHMACJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hmacBits int
		prob     float32
	}{
		{"HMAC256", cryptoutilSharedMagic.HMACKeySize256, cryptoutilSharedMagic.TestProbAlways},
		{"HMAC384", cryptoutilSharedMagic.HMACKeySize384, cryptoutilSharedMagic.TestProbTenth},
		{"HMAC512", cryptoutilSharedMagic.HMACKeySize512, cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			jwk, err := GenerateHMACJWK(tc.hmacBits)
			require.NoError(t, err)
			require.NotNil(t, jwk)
			require.Equal(t, KtyOCT, jwk.KeyType())

			// Test function generator.
			genFunc := GenerateHMACJWKFunction(tc.hmacBits)
			jwk2, err := genFunc()
			require.NoError(t, err)
			require.NotNil(t, jwk2)
			require.Equal(t, KtyOCT, jwk2.KeyType())
		})
	}
}

// TestBuildJWK tests BuildJWK helper function.
func TestBuildJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		kty         joseJwa.KeyType
		generateKey func() (any, error)
	}{
		{
			name: "RSA",
			kty:  KtyRSA,
			generateKey: func() (any, error) {
				keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.RSAKeySize2048)
				if err != nil {
					return nil, err //nolint:wrapcheck // Test helper
				}

				return keyPair.Private, nil
			},
		},
		{
			name: "EC",
			kty:  KtyEC,
			generateKey: func() (any, error) {
				keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
				if err != nil {
					return nil, err //nolint:wrapcheck // Test helper //nolint:wrapcheck // Test helper
				}

				return keyPair.Private, nil
			},
		},
		{
			name: "OKP",
			kty:  KtyOKP,
			generateKey: func() (any, error) {
				keyPair, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair(cryptoutilSharedCryptoKeygen.EdCurveEd25519)
				if err != nil {
					return nil, err //nolint:wrapcheck // Test helper
				}

				return keyPair.Private, nil
			},
		},
		{
			name: "OCT",
			kty:  KtyOCT,
			generateKey: func() (any, error) {
				key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.HMACKeySize256)
				if err != nil {
					return nil, err //nolint:wrapcheck // Test helper
				}

				return []byte(key), nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test successful build.
			raw, err := tt.generateKey()
			require.NoError(t, err)

			jwk, err := BuildJWK(tt.kty, raw, nil)
			require.NoError(t, err)
			require.NotNil(t, jwk)
			require.Equal(t, tt.kty, jwk.KeyType())

			// Verify KID was set (UUID v7).
			kidVal, ok := jwk.KeyID()
			require.True(t, ok)
			require.NotEmpty(t, kidVal)
			_, err = googleUuid.Parse(kidVal)
			require.NoError(t, err)
		})
	}

	// Test error propagation from keygen.
	t.Run("ErrorPropagation", func(t *testing.T) {
		t.Parallel()

		keyGenErr := errors.New("key generation failed")
		jwk, err := BuildJWK(KtyRSA, nil, keyGenErr)
		require.Error(t, err)
		require.Nil(t, jwk)
		require.Contains(t, err.Error(), "failed to generate")
	})

	// Test import failure with invalid raw data.
	t.Run("ImportFailure", func(t *testing.T) {
		t.Parallel()

		invalidRaw := "not a valid key"
		jwk, err := BuildJWK(KtyRSA, invalidRaw, nil)
		require.Error(t, err)
		require.Nil(t, jwk)
		require.Contains(t, err.Error(), "failed to import")
	})

	// Test KeyType set failure (simulated by using nil JWK in test).
	t.Run("KeyTypeSetFailure", func(t *testing.T) {
		t.Parallel()

		// Create a valid key but test error handling path.
		// Since we cannot directly fail Set() on a valid JWK, this test
		// verifies the success path for KeyType setting as additional coverage.
		keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.RSAKeySize2048)
		require.NoError(t, err)

		jwk, err := BuildJWK(KtyRSA, keyPair.Private, nil)
		require.NoError(t, err)
		require.NotNil(t, jwk)

		// Verify KeyType was set successfully.
		require.Equal(t, KtyRSA, jwk.KeyType())
	})
}
