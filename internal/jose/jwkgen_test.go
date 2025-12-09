// Copyright (c) 2025 Justin Cranford
//
//

package jose

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"testing"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

// TestGenerateRSAJWK tests RSA JWK generation.
func TestGenerateRSAJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rsaBits int
	}{
		{"RSA2048", cryptoutilMagic.RSAKeySize2048},
		{"RSA3072", cryptoutilMagic.RSAKeySize3072},
		{"RSA4096", cryptoutilMagic.RSAKeySize4096},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
	}{
		{"P256", elliptic.P256()},
		{"P384", elliptic.P384()},
		{"P521", elliptic.P521()},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
	}{
		{"P256", ecdh.P256()},
		{"P384", ecdh.P384()},
		{"P521", ecdh.P521()},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
	}{
		{"AES128", cryptoutilMagic.AESKeySize128},
		{"AES192", cryptoutilMagic.AESKeySize192},
		{"AES256", cryptoutilMagic.AESKeySize256},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
		name     string
		aesHsBits int
	}{
		{"AES128HS256", cryptoutilMagic.HMACKeySize256},
		{"AES192HS384", cryptoutilMagic.HMACKeySize384},
		{"AES256HS512", cryptoutilMagic.HMACKeySize512},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
	}{
		{"HMAC256", cryptoutilMagic.HMACKeySize256},
		{"HMAC384", cryptoutilMagic.HMACKeySize384},
		{"HMAC512", cryptoutilMagic.HMACKeySize512},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

	// Test successful build.
	rsaKey, err := cryptoutilKeyGen.GenerateRSAKeyPair(cryptoutilMagic.RSAKeySize2048)
	require.NoError(t, err)

	jwk, err := BuildJWK(KtyRSA, rsaKey.Private, nil)
	require.NoError(t, err)
	require.NotNil(t, jwk)
	require.Equal(t, KtyRSA, jwk.KeyType())

	// Test error propagation from keygen.
	keyGenErr := errors.New("key generation failed")
	jwk, err = BuildJWK(KtyRSA, nil, keyGenErr)
	require.Error(t, err)
	require.Nil(t, jwk)
	require.Contains(t, err.Error(), "failed to generate RSA")
}
