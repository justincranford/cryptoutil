// Copyright (c) 2025 Justin Cranford

//go:build !fuzz

package keygen

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	rsa "crypto/rsa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestRSAKeyGenerationProperties verifies RSA key generation properties.
func TestRSAKeyGenerationProperties(t *testing.T) {
	t.Parallel()

	// Reduce iterations for expensive RSA key generation (10 instead of default 100)
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = cryptoutilSharedMagic.JoseJADefaultMaxMaterials
	properties := gopter.NewProperties(params)

	// Property 1: RSA key generation produces valid keys
	properties.Property("RSA keys are valid for supported bit sizes", prop.ForAll(
		func(bits uint) bool {
			// Only test supported RSA bit sizes
			supportedBits := []int{cryptoutilSharedMagic.DefaultMetricsBatchSize, cryptoutilSharedMagic.RSA3072KeySize, cryptoutilSharedMagic.RSA4096KeySize}
			//nolint:gosec // G115: Intentional modulo for property test array selection
			rsaBits := supportedBits[int(bits)%len(supportedBits)]

			keyPair, err := GenerateRSAKeyPair(rsaBits)
			if err != nil {
				return false // Should not error for valid bit sizes
			}

			// Verify key pair structure
			if keyPair == nil || keyPair.Private == nil || keyPair.Public == nil {
				return false
			}

			// Verify private key type
			rsaPriv, ok := keyPair.Private.(*rsa.PrivateKey)
			if !ok {
				return false
			}

			// Verify public key type and size
			rsaPub, ok := keyPair.Public.(*rsa.PublicKey)
			if !ok {
				return false
			}

			// Verify key size matches requested size
			return rsaPriv.N.BitLen() == rsaBits && rsaPub.N.BitLen() == rsaBits
		},
		gen.UInt(),
	))

	// Property 2: RSA key generation produces unique keys
	properties.Property("RSA keys are unique across generations", prop.ForAll(
		func() bool {
			keyPair1, err1 := GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
			keyPair2, err2 := GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)

			if err1 != nil || err2 != nil {
				return false
			}

			rsaPriv1, ok1 := keyPair1.Private.(*rsa.PrivateKey)

			rsaPriv2, ok2 := keyPair2.Private.(*rsa.PrivateKey)
			if !ok1 || !ok2 {
				return false
			}

			// Different keys should have different moduli
			return rsaPriv1.N.Cmp(rsaPriv2.N) != 0
		},
	))

	properties.TestingRun(t)
}

// TestECDSAKeyGenerationProperties verifies ECDSA key generation properties.
func TestECDSAKeyGenerationProperties(t *testing.T) {
	t.Parallel()

	// Reduce iterations for faster test execution (25 instead of default 100)
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = cryptoutilSharedMagic.TLSMaxValidityCACertYears
	properties := gopter.NewProperties(params)

	// Property 1: ECDSA key generation produces valid keys for all supported curves
	properties.Property("ECDSA keys are valid for all supported curves", prop.ForAll(
		func(curveIdx uint) bool {
			// Test all supported NIST curves
			curves := []elliptic.Curve{
				elliptic.P256(),
				elliptic.P384(),
				elliptic.P521(),
			}
			//nolint:gosec // G115: Intentional modulo for property test array selection
			curve := curves[int(curveIdx)%len(curves)]

			keyPair, err := GenerateECDSAKeyPair(curve)
			if err != nil {
				return false
			}

			if keyPair == nil || keyPair.Private == nil || keyPair.Public == nil {
				return false
			}

			// Verify private key type
			ecdsaPriv, ok := keyPair.Private.(*ecdsa.PrivateKey)
			if !ok {
				return false
			}

			// Verify public key type
			ecdsaPub, ok := keyPair.Public.(*ecdsa.PublicKey)
			if !ok {
				return false
			}

			// Verify curve matches
			return ecdsaPriv.Curve == curve && ecdsaPub.Curve == curve
		},
		gen.UInt(),
	))

	// Property 2: ECDSA key generation produces unique keys
	properties.Property("ECDSA keys are unique across generations", prop.ForAll(
		func() bool {
			keyPair1, err1 := GenerateECDSAKeyPair(elliptic.P256())
			keyPair2, err2 := GenerateECDSAKeyPair(elliptic.P256())

			if err1 != nil || err2 != nil {
				return false
			}

			ecdsaPriv1, ok1 := keyPair1.Private.(*ecdsa.PrivateKey)

			ecdsaPriv2, ok2 := keyPair2.Private.(*ecdsa.PrivateKey)
			if !ok1 || !ok2 {
				return false
			}

			// Different keys should have different private values
			return ecdsaPriv1.D.Cmp(ecdsaPriv2.D) != 0
		},
	))

	properties.TestingRun(t)
}

// TestECDHKeyGenerationProperties verifies ECDH key generation properties.
func TestECDHKeyGenerationProperties(t *testing.T) {
	t.Parallel()

	// Reduce iterations for faster test execution (25 instead of default 100)
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = cryptoutilSharedMagic.TLSMaxValidityCACertYears
	properties := gopter.NewProperties(params)

	// Property 1: ECDH key generation produces valid keys
	properties.Property("ECDH keys are valid for all supported curves", prop.ForAll(
		func(curveIdx uint) bool {
			curves := []ecdh.Curve{
				ecdh.P256(),
				ecdh.P384(),
				ecdh.P521(),
			}
			//nolint:gosec // G115: Intentional modulo for property test array selection
			curve := curves[int(curveIdx)%len(curves)]

			keyPair, err := GenerateECDHKeyPair(curve)
			if err != nil {
				return false
			}

			return keyPair != nil && keyPair.Private != nil && keyPair.Public != nil
		},
		gen.UInt(),
	))

	properties.TestingRun(t)
}

// TestEdDSAKeyGenerationProperties verifies EdDSA key generation properties.
func TestEdDSAKeyGenerationProperties(t *testing.T) {
	t.Parallel()

	// Reduce iterations for faster test execution (25 instead of default 100)
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = cryptoutilSharedMagic.TLSMaxValidityCACertYears
	properties := gopter.NewProperties(params)

	// Property 1: EdDSA key generation produces valid Ed25519 keys
	properties.Property("EdDSA keys are valid for Ed25519", prop.ForAll(
		func() bool {
			keyPair, err := GenerateEDDSAKeyPair(cryptoutilSharedMagic.EdCurveEd25519)
			if err != nil {
				return false
			}

			if keyPair == nil || keyPair.Private == nil || keyPair.Public == nil {
				return false
			}

			// Verify key types
			edPriv, ok1 := keyPair.Private.(ed25519.PrivateKey)
			edPub, ok2 := keyPair.Public.(ed25519.PublicKey)

			if !ok1 || !ok2 {
				return false
			}

			// Verify key sizes (Ed25519 has fixed sizes)
			return len(edPriv) == ed25519.PrivateKeySize && len(edPub) == ed25519.PublicKeySize
		},
	))

	// Property 2: EdDSA key generation produces unique keys
	properties.Property("EdDSA keys are unique across generations", prop.ForAll(
		func() bool {
			keyPair1, err1 := GenerateEDDSAKeyPair(cryptoutilSharedMagic.EdCurveEd25519)
			keyPair2, err2 := GenerateEDDSAKeyPair(cryptoutilSharedMagic.EdCurveEd25519)

			if err1 != nil || err2 != nil {
				return false
			}

			edPriv1, ok1 := keyPair1.Private.(ed25519.PrivateKey)

			edPriv2, ok2 := keyPair2.Private.(ed25519.PrivateKey)
			if !ok1 || !ok2 {
				return false
			}

			// Different keys should have different private key bytes
			for i := 0; i < ed25519.PrivateKeySize; i++ {
				if edPriv1[i] != edPriv2[i] {
					return true
				}
			}

			return false // Should be different
		},
	))

	properties.TestingRun(t)
}

// TestAESKeyGenerationProperties verifies AES key generation properties.
func TestAESKeyGenerationProperties(t *testing.T) {
	t.Parallel()

	// Fast symmetric key generation - use default 100 iterations
	properties := gopter.NewProperties(nil)

	// Property 1: AES key generation produces correct key sizes
	properties.Property("AES keys have correct size for supported bit lengths", prop.ForAll(
		func(bits uint) bool {
			supportedBits := []int{cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits, cryptoutilSharedMagic.SymmetricKeySize192, cryptoutilSharedMagic.MaxUnsealSharedSecrets}
			//nolint:gosec // G115: Intentional modulo for property test array selection
			aesBits := supportedBits[int(bits)%len(supportedBits)]

			key, err := GenerateAESKey(aesBits)
			if err != nil {
				return false
			}

			expectedBytes := aesBits / cryptoutilSharedMagic.IMMinPasswordLength

			return len(key) == expectedBytes
		},
		gen.UInt(),
	))

	// Property 2: AES key generation produces unique keys
	properties.Property("AES keys are unique across generations", prop.ForAll(
		func() bool {
			key1, err1 := GenerateAESKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
			key2, err2 := GenerateAESKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)

			if err1 != nil || err2 != nil {
				return false
			}

			// Different keys should have at least one different byte
			for i := 0; i < len(key1); i++ {
				if key1[i] != key2[i] {
					return true
				}
			}

			return false
		},
	))

	properties.TestingRun(t)
}

// TestHMACKeyGenerationProperties verifies HMAC key generation properties.
func TestHMACKeyGenerationProperties(t *testing.T) {
	t.Parallel()

	// Fast symmetric key generation - use default 100 iterations
	properties := gopter.NewProperties(nil)

	// Property 1: HMAC key generation produces correct key sizes
	properties.Property("HMAC keys have correct size for supported bit lengths", prop.ForAll(
		func(bits uint) bool {
			supportedBits := []int{cryptoutilSharedMagic.MaxUnsealSharedSecrets, cryptoutilSharedMagic.SymmetricKeySize384, cryptoutilSharedMagic.DefaultTracesBatchSize}
			//nolint:gosec // G115: Intentional modulo for property test array selection
			hmacBits := supportedBits[int(bits)%len(supportedBits)]

			key, err := GenerateHMACKey(hmacBits)
			if err != nil {
				return false
			}

			expectedBytes := hmacBits / cryptoutilSharedMagic.IMMinPasswordLength

			return len(key) == expectedBytes
		},
		gen.UInt(),
	))

	// Property 2: HMAC key generation produces unique keys
	properties.Property("HMAC keys are unique across generations", prop.ForAll(
		func() bool {
			key1, err1 := GenerateHMACKey(cryptoutilSharedMagic.DefaultTracesBatchSize)
			key2, err2 := GenerateHMACKey(cryptoutilSharedMagic.DefaultTracesBatchSize)

			if err1 != nil || err2 != nil {
				return false
			}

			// Different keys should have at least one different byte
			for i := 0; i < len(key1); i++ {
				if key1[i] != key2[i] {
					return true
				}
			}

			return false
		},
	))

	properties.TestingRun(t)
}
