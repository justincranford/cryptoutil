// Copyright (c) 2025 Justin Cranford
//
//

package keygen

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// IMPORTANT: All Fuzz* test function names MUST be unique and MUST NOT be substrings of any other fuzz test names.
// This ensures cross-platform compatibility with the `-fuzz` parameter (no quotes or regex needed).

// FuzzGenerateRSAKeyPair tests RSA key pair generation with various bit sizes.
func FuzzGenerateRSAKeyPair(f *testing.F) {
	// Add seed corpus with valid RSA key sizes
	f.Add(cryptoutilSharedMagic.RSAKeySize2048)
	f.Add(cryptoutilSharedMagic.RSAKeySize3072)
	f.Add(cryptoutilSharedMagic.RSAKeySize4096)

	f.Fuzz(func(t *testing.T, rsaBits int) {
		// Only test valid RSA key sizes to avoid expected errors
		if rsaBits != cryptoutilSharedMagic.RSAKeySize2048 && rsaBits != cryptoutilSharedMagic.RSAKeySize3072 && rsaBits != cryptoutilSharedMagic.RSAKeySize4096 {
			t.Skip("Skipping invalid RSA key size for fuzzing")
		}

		keyPair, err := GenerateRSAKeyPair(rsaBits)
		require.NoError(t, err, "GenerateRSAKeyPair should not fail with valid input")

		require.NotNil(t, keyPair, "GenerateRSAKeyPair should return a valid key pair")
		require.NotNil(t, keyPair.Private, "GenerateRSAKeyPair should return a valid private key")
		require.NotNil(t, keyPair.Public, "GenerateRSAKeyPair should return a valid public key")
	})
}

// FuzzGenerateECDSAKeyPair tests ECDSA key pair generation with various curves.
func FuzzGenerateECDSAKeyPair(f *testing.F) {
	// Add seed corpus with curve identifiers (we'll map them to actual curves)
	f.Add("P256")
	f.Add("P384")
	f.Add("P521")

	f.Fuzz(func(t *testing.T, curveName string) {
		var curve elliptic.Curve

		// Map string to actual curve
		switch curveName {
		case ECCurveP256:
			curve = elliptic.P256()
		case ECCurveP384:
			curve = elliptic.P384()
		case ECCurveP521:
			curve = elliptic.P521()
		default:
			t.Skip("Skipping unknown curve for fuzzing")
		}

		keyPair, err := GenerateECDSAKeyPair(curve)
		require.NoError(t, err, "GenerateECDSAKeyPair should not fail with valid input")

		require.NotNil(t, keyPair, "GenerateECDSAKeyPair should return a valid key pair")
		require.NotNil(t, keyPair.Private, "GenerateECDSAKeyPair should return a valid private key")
		require.NotNil(t, keyPair.Public, "GenerateECDSAKeyPair should return a valid public key")
	})
}

// FuzzGenerateECDHKeyPair tests ECDH key pair generation with various curves.
func FuzzGenerateECDHKeyPair(f *testing.F) {
	// Add seed corpus with curve identifiers
	f.Add("P256")
	f.Add("P384")
	f.Add("P521")

	f.Fuzz(func(t *testing.T, curveName string) {
		var curve ecdh.Curve

		// Map string to actual curve
		switch curveName {
		case ECCurveP256:
			curve = ecdh.P256()
		case ECCurveP384:
			curve = ecdh.P384()
		case ECCurveP521:
			curve = ecdh.P521()
		default:
			t.Skip("Skipping unknown curve for fuzzing")
		}

		keyPair, err := GenerateECDHKeyPair(curve)
		require.NoError(t, err, "GenerateECDHKeyPair should not fail with valid input")

		require.NotNil(t, keyPair, "GenerateECDHKeyPair should return a valid key pair")
		require.NotNil(t, keyPair.Private, "GenerateECDHKeyPair should return a valid private key")
		require.NotNil(t, keyPair.Public, "GenerateECDHKeyPair should return a valid public key")
	})
}

// FuzzGenerateEDDSAKeyPair tests EdDSA key pair generation with various curves.
func FuzzGenerateEDDSAKeyPair(f *testing.F) {
	// Add seed corpus with valid EdDSA curve names
	f.Add("Ed25519")
	f.Add("Ed448")

	f.Fuzz(func(t *testing.T, edCurve string) {
		// Only test valid EdDSA curves
		if edCurve != EdCurveEd25519 && edCurve != EdCurveEd448 {
			t.Skip("Skipping invalid EdDSA curve for fuzzing")
		}

		keyPair, err := GenerateEDDSAKeyPair(edCurve)
		require.NoError(t, err, "GenerateEDDSAKeyPair should not fail with valid input")

		require.NotNil(t, keyPair, "GenerateEDDSAKeyPair should return a valid key pair")
		require.NotNil(t, keyPair.Private, "GenerateEDDSAKeyPair should return a valid private key")
		require.NotNil(t, keyPair.Public, "GenerateEDDSAKeyPair should return a valid public key")
	})
}

// FuzzGenerateAESKey tests AES key generation with various key sizes.
func FuzzGenerateAESKey(f *testing.F) {
	// Add seed corpus with valid AES key sizes
	f.Add(cryptoutilSharedMagic.AESKeySize128)
	f.Add(cryptoutilSharedMagic.AESKeySize192)
	f.Add(cryptoutilSharedMagic.AESKeySize256)

	f.Fuzz(func(t *testing.T, aesBits int) {
		// Only test valid AES key sizes
		if aesBits != cryptoutilSharedMagic.AESKeySize128 && aesBits != cryptoutilSharedMagic.AESKeySize192 && aesBits != cryptoutilSharedMagic.AESKeySize256 {
			t.Skip("Skipping invalid AES key size for fuzzing")
		}

		key, err := GenerateAESKey(aesBits)
		require.NoError(t, err, "GenerateAESKey should not fail with valid input")

		expectedLength := aesBits / cryptoutilSharedMagic.BitsToBytes
		require.Len(t, key, expectedLength, "GenerateAESKey should return key of correct length")
	})
}

// FuzzGenerateAESHSKey tests AES HMAC-SHA2 key generation with various key sizes.
func FuzzGenerateAESHSKey(f *testing.F) {
	// Add seed corpus with valid AES HS key sizes
	f.Add(cryptoutilSharedMagic.AESHSKeySize256)
	f.Add(cryptoutilSharedMagic.AESHSKeySize384)
	f.Add(cryptoutilSharedMagic.AESHSKeySize512)

	f.Fuzz(func(t *testing.T, aesHsBits int) {
		// Only test valid AES HS key sizes
		if aesHsBits != cryptoutilSharedMagic.AESHSKeySize256 && aesHsBits != cryptoutilSharedMagic.AESHSKeySize384 && aesHsBits != cryptoutilSharedMagic.AESHSKeySize512 {
			t.Skip("Skipping invalid AES HS key size for fuzzing")
		}

		key, err := GenerateAESHSKey(aesHsBits)
		require.NoError(t, err, "GenerateAESHSKey should not fail with valid input")

		expectedLength := aesHsBits / cryptoutilSharedMagic.BitsToBytes
		require.Len(t, key, expectedLength, "GenerateAESHSKey should return key of correct length")
	})
}

// FuzzGenerateHMACKey tests HMAC key generation with various key sizes.
func FuzzGenerateHMACKey(f *testing.F) {
	// Add seed corpus with valid HMAC key sizes
	f.Add(cryptoutilSharedMagic.MinHMACKeySize)
	f.Add(cryptoutilSharedMagic.HMACKeySize384)
	f.Add(cryptoutilSharedMagic.HMACKeySize512)

	f.Fuzz(func(t *testing.T, hmacBits int) {
		// Only test valid HMAC key sizes (minimum 256 bits)
		if hmacBits < cryptoutilSharedMagic.MinHMACKeySize {
			t.Skip("Skipping invalid HMAC key size for fuzzing")
		}

		key, err := GenerateHMACKey(hmacBits)
		require.NoError(t, err, "GenerateHMACKey should not fail with valid input")

		expectedLength := hmacBits / cryptoutilSharedMagic.BitsToBytes
		require.Len(t, key, expectedLength, "GenerateHMACKey should return key of correct length")
	})
}
