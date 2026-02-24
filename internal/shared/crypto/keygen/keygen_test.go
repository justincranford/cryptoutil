// Copyright (c) 2025 Justin Cranford

package keygen

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	rsa "crypto/rsa"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"github.com/cloudflare/circl/sign/ed448"
	"github.com/stretchr/testify/require"
)

// TestGenerateRSAKeyPair tests RSA key pair generation.
func TestGenerateRSAKeyPair(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		rsaBits int
		prob    float32
	}{
		{"RSA 2048", 2048, cryptoutilSharedMagic.TestProbAlways},
		{"RSA 3072", 3072, cryptoutilSharedMagic.TestProbTenth},
		{"RSA 4096", 4096, cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			keyPair, err := GenerateRSAKeyPair(tc.rsaBits)
			require.NoError(t, err)
			require.NotNil(t, keyPair)
			require.NotNil(t, keyPair.Private)
			require.NotNil(t, keyPair.Public)

			// Verify key type
			privateKey, ok := keyPair.Private.(*rsa.PrivateKey)
			require.True(t, ok, "private key should be *rsa.PrivateKey")
			require.Equal(t, tc.rsaBits, privateKey.N.BitLen())

			publicKey, ok := keyPair.Public.(*rsa.PublicKey)
			require.True(t, ok, "public key should be *rsa.PublicKey")
			require.Equal(t, tc.rsaBits, publicKey.N.BitLen())
		})
	}
}

// TestGenerateRSAKeyPairFunction tests RSA key pair generation function factory.
func TestGenerateRSAKeyPairFunction(t *testing.T) {
	t.Parallel()

	const rsaBits = 2048

	genFunc := GenerateRSAKeyPairFunction(rsaBits)
	require.NotNil(t, genFunc)

	keyPair, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, keyPair)
	require.NotNil(t, keyPair.Private)
	require.NotNil(t, keyPair.Public)
}

// TestGenerateECDSAKeyPair tests ECDSA key pair generation.
func TestGenerateECDSAKeyPair(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		curve elliptic.Curve
		prob  float32
	}{
		{"P-256", elliptic.P256(), cryptoutilSharedMagic.TestProbAlways},
		{"P-384", elliptic.P384(), cryptoutilSharedMagic.TestProbTenth},
		{"P-521", elliptic.P521(), cryptoutilSharedMagic.TestProbTenth},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			keyPair, err := GenerateECDSAKeyPair(tc.curve)
			require.NoError(t, err)
			require.NotNil(t, keyPair)
			require.NotNil(t, keyPair.Private)
			require.NotNil(t, keyPair.Public)

			// Verify key type
			privateKey, ok := keyPair.Private.(*ecdsa.PrivateKey)
			require.True(t, ok, "private key should be *ecdsa.PrivateKey")
			require.Equal(t, tc.curve, privateKey.Curve)

			publicKey, ok := keyPair.Public.(*ecdsa.PublicKey)
			require.True(t, ok, "public key should be *ecdsa.PublicKey")
			require.Equal(t, tc.curve, publicKey.Curve)
		})
	}
}

// TestGenerateECDSAKeyPairFunction tests ECDSA key pair generation function factory.
func TestGenerateECDSAKeyPairFunction(t *testing.T) {
	t.Parallel()

	curve := elliptic.P256()
	genFunc := GenerateECDSAKeyPairFunction(curve)
	require.NotNil(t, genFunc)

	keyPair, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, keyPair)
	require.NotNil(t, keyPair.Private)
	require.NotNil(t, keyPair.Public)
}

// TestGenerateECDHKeyPair tests ECDH key pair generation.
func TestGenerateECDHKeyPair(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		curve ecdh.Curve
		prob  float32
	}{
		{"P-256", ecdh.P256(), cryptoutilSharedMagic.TestProbAlways},
		{"P-384", ecdh.P384(), cryptoutilSharedMagic.TestProbTenth},
		{"P-521", ecdh.P521(), cryptoutilSharedMagic.TestProbTenth},
		{"X25519", ecdh.X25519(), cryptoutilSharedMagic.TestProbAlways},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			keyPair, err := GenerateECDHKeyPair(tc.curve)
			require.NoError(t, err)
			require.NotNil(t, keyPair)
			require.NotNil(t, keyPair.Private)
			require.NotNil(t, keyPair.Public)

			// Verify key type
			privateKey, ok := keyPair.Private.(*ecdh.PrivateKey)
			require.True(t, ok, "private key should be *ecdh.PrivateKey")
			require.NotNil(t, privateKey)

			publicKey, ok := keyPair.Public.(*ecdh.PublicKey)
			require.True(t, ok, "public key should be *ecdh.PublicKey")
			require.NotNil(t, publicKey)
		})
	}
}

// TestGenerateECDHKeyPairFunction tests ECDH key pair generation function factory.
func TestGenerateECDHKeyPairFunction(t *testing.T) {
	t.Parallel()

	curve := ecdh.P256()
	genFunc := GenerateECDHKeyPairFunction(curve)
	require.NotNil(t, genFunc)

	keyPair, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, keyPair)
	require.NotNil(t, keyPair.Private)
	require.NotNil(t, keyPair.Public)
}

// TestGenerateEDDSAKeyPair tests EdDSA key pair generation.
func TestGenerateEDDSAKeyPair(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		curve  string
		prob   float32
		verify func(*testing.T, *KeyPair)
	}{
		{
			name:  "Ed25519",
			curve: EdCurveEd25519,
			prob:  cryptoutilSharedMagic.TestProbAlways,
			verify: func(t *testing.T, keyPair *KeyPair) {
				privateKey, ok := keyPair.Private.(ed25519.PrivateKey)
				require.True(t, ok, "private key should be ed25519.PrivateKey")
				require.Len(t, privateKey, ed25519.PrivateKeySize)

				publicKey, ok := keyPair.Public.(ed25519.PublicKey)
				require.True(t, ok, "public key should be ed25519.PublicKey")
				require.Len(t, publicKey, ed25519.PublicKeySize)
			},
		},
		{
			name:  "Ed448",
			curve: EdCurveEd448,
			prob:  cryptoutilSharedMagic.TestProbTenth,
			verify: func(t *testing.T, keyPair *KeyPair) {
				privateKey, ok := keyPair.Private.(ed448.PrivateKey)
				require.True(t, ok, "private key should be ed448.PrivateKey")
				require.Len(t, privateKey, ed448.PrivateKeySize)

				publicKey, ok := keyPair.Public.(ed448.PublicKey)
				require.True(t, ok, "public key should be ed448.PublicKey")
				require.Len(t, publicKey, ed448.PublicKeySize)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			keyPair, err := GenerateEDDSAKeyPair(tc.curve)
			require.NoError(t, err)
			require.NotNil(t, keyPair)
			require.NotNil(t, keyPair.Private)
			require.NotNil(t, keyPair.Public)

			tc.verify(t, keyPair)
		})
	}
}

// TestGenerateEDDSAKeyPair_InvalidCurve tests EdDSA with invalid curve.
func TestGenerateEDDSAKeyPair_InvalidCurve(t *testing.T) {
	t.Parallel()

	_, err := GenerateEDDSAKeyPair("invalid-curve")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported Ed curve")
}

// TestGenerateEDDSAKeyPairFunction tests EdDSA key pair generation function factory.
func TestGenerateEDDSAKeyPairFunction(t *testing.T) {
	t.Parallel()

	genFunc := GenerateEDDSAKeyPairFunction(EdCurveEd25519)
	require.NotNil(t, genFunc)

	keyPair, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, keyPair)
	require.NotNil(t, keyPair.Private)
	require.NotNil(t, keyPair.Public)
}

// TestGenerateAESKey tests AES key generation.
func TestGenerateAESKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		aesBits      int
		expectedSize int
		prob         float32
	}{
		{"AES-128", aesKeySize128, aesKeySize128 / bitsToBytes, cryptoutilSharedMagic.TestProbAlways},
		{"AES-192", aesKeySize192, aesKeySize192 / bitsToBytes, cryptoutilSharedMagic.TestProbQuarter},
		{"AES-256", aesKeySize256, aesKeySize256 / bitsToBytes, cryptoutilSharedMagic.TestProbQuarter},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			key, err := GenerateAESKey(tc.aesBits)
			require.NoError(t, err)
			require.NotNil(t, key)
			require.Len(t, key, tc.expectedSize)

			// Verify all bytes are not zero (highly unlikely for random key)
			allZero := true

			for _, b := range key {
				if b != 0 {
					allZero = false

					break
				}
			}

			require.False(t, allZero, "key should not be all zeros")
		})
	}
}

// TestGenerateAESKey_InvalidSize tests AES with invalid key size.
func TestGenerateAESKey_InvalidSize(t *testing.T) {
	t.Parallel()

	_, err := GenerateAESKey(100)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid AES key size")
}

// TestGenerateAESKeyFunction tests AES key generation function factory.
func TestGenerateAESKeyFunction(t *testing.T) {
	t.Parallel()

	genFunc := GenerateAESKeyFunction(aesKeySize256)
	require.NotNil(t, genFunc)

	key, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Len(t, key, aesKeySize256/bitsToBytes)
}

// TestGenerateAESHSKey tests AES-HMAC-SHA2 key generation.
func TestGenerateAESHSKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		aesHsBits    int
		expectedSize int
		prob         float32
	}{
		{"AES-HS-256", aesHsKeySize256, aesHsKeySize256 / bitsToBytes, cryptoutilSharedMagic.TestProbAlways},
		{"AES-HS-384", aesHsKeySize384, aesHsKeySize384 / bitsToBytes, cryptoutilSharedMagic.TestProbQuarter},
		{"AES-HS-512", aesHsKeySize512, aesHsKeySize512 / bitsToBytes, cryptoutilSharedMagic.TestProbQuarter},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			key, err := GenerateAESHSKey(tc.aesHsBits)
			require.NoError(t, err)
			require.NotNil(t, key)
			require.Len(t, key, tc.expectedSize)

			// Verify all bytes are not zero
			allZero := true

			for _, b := range key {
				if b != 0 {
					allZero = false

					break
				}
			}

			require.False(t, allZero, "key should not be all zeros")
		})
	}
}

// TestGenerateAESHSKey_InvalidSize tests AES-HS with invalid key size.
func TestGenerateAESHSKey_InvalidSize(t *testing.T) {
	t.Parallel()

	_, err := GenerateAESHSKey(100)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid AES HAMC-SHA2 key size")
}

// TestGenerateAESHSKeyFunction tests AES-HS key generation function factory.
func TestGenerateAESHSKeyFunction(t *testing.T) {
	t.Parallel()

	genFunc := GenerateAESHSKeyFunction(aesHsKeySize256)
	require.NotNil(t, genFunc)

	key, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Len(t, key, aesHsKeySize256/bitsToBytes)
}

// TestGenerateHMACKey tests HMAC key generation.
func TestGenerateHMACKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		hmacBits     int
		expectedSize int
		prob         float32
	}{
		{"HMAC 256", 256, 256 / bitsToBytes, cryptoutilSharedMagic.TestProbAlways},
		{"HMAC 512", 512, 512 / bitsToBytes, cryptoutilSharedMagic.TestProbQuarter},
		{"HMAC 1024", 1024, 1024 / bitsToBytes, cryptoutilSharedMagic.TestProbQuarter},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cryptoutilSharedUtilRandom.SkipByProbability(t, tc.prob)

			key, err := GenerateHMACKey(tc.hmacBits)
			require.NoError(t, err)
			require.NotNil(t, key)
			require.Len(t, key, tc.expectedSize)

			// Verify all bytes are not zero
			allZero := true

			for _, b := range key {
				if b != 0 {
					allZero = false

					break
				}
			}

			require.False(t, allZero, "key should not be all zeros")
		})
	}
}

// TestGenerateHMACKey_BelowMinimum tests HMAC with key size below minimum.
func TestGenerateHMACKey_BelowMinimum(t *testing.T) {
	t.Parallel()

	_, err := GenerateHMACKey(minHMACKeySize - 8) // Below minimum
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid HMAC key size")
}

// TestGenerateHMACKeyFunction tests HMAC key generation function factory.
func TestGenerateHMACKeyFunction(t *testing.T) {
	t.Parallel()

	genFunc := GenerateHMACKeyFunction(512)
	require.NotNil(t, genFunc)

	key, err := genFunc()
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Len(t, key, 512/bitsToBytes)
}

// TestKeyPair_isKey tests KeyPair implements Key interface.
func TestKeyPair_isKey(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// Verify it implements Key interface by calling isKey
	require.NotPanics(t, func() {
		keyPair.isKey()
	})

	// Verify it can be assigned to Key interface
	var _ Key = keyPair
}

// TestSecretKey_isKey tests SecretKey implements Key interface.
func TestSecretKey_isKey(t *testing.T) {
	t.Parallel()

	secretKey, err := GenerateAESKey(aesKeySize256)
	require.NoError(t, err)

	// Verify it implements Key interface by calling isKey
	require.NotPanics(t, func() {
		secretKey.isKey()
	})

	// Verify it can be assigned to Key interface
	var _ Key = secretKey
}
