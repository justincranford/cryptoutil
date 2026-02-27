// Copyright (c) 2025 Justin Cranford
//
//

package keygen

import (
	"crypto/ecdh"
	"crypto/elliptic"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkGenerateRSA2048KeyPair benchmarks RSA-2048 key pair generation.
func BenchmarkGenerateRSA2048KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
		require.NoError(b, err, "GenerateRSAKeyPair should not fail")
	}
}

// BenchmarkGenerateRSA3072KeyPair benchmarks RSA-3072 key pair generation.
func BenchmarkGenerateRSA3072KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRSAKeyPair(cryptoutilSharedMagic.RSA3072KeySize)
		require.NoError(b, err, "GenerateRSAKeyPair should not fail")
	}
}

// BenchmarkGenerateRSA4096KeyPair benchmarks RSA-4096 key pair generation.
func BenchmarkGenerateRSA4096KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRSAKeyPair(cryptoutilSharedMagic.RSA4096KeySize)
		require.NoError(b, err, "GenerateRSAKeyPair should not fail")
	}
}

// BenchmarkGenerateECDSAP256KeyPair benchmarks ECDSA P-256 key pair generation.
func BenchmarkGenerateECDSAP256KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateECDSAKeyPair(elliptic.P256())
		require.NoError(b, err, "GenerateECDSAKeyPair should not fail")
	}
}

// BenchmarkGenerateECDSAP384KeyPair benchmarks ECDSA P-384 key pair generation.
func BenchmarkGenerateECDSAP384KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateECDSAKeyPair(elliptic.P384())
		require.NoError(b, err, "GenerateECDSAKeyPair should not fail")
	}
}

// BenchmarkGenerateECDSAP521KeyPair benchmarks ECDSA P-521 key pair generation.
func BenchmarkGenerateECDSAP521KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateECDSAKeyPair(elliptic.P521())
		require.NoError(b, err, "GenerateECDSAKeyPair should not fail")
	}
}

// BenchmarkGenerateECDHKeyPair benchmarks ECDH key pair generation.
func BenchmarkGenerateECDHKeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateECDHKeyPair(ecdh.P256())
		require.NoError(b, err, "GenerateECDHKeyPair should not fail")
	}
}

// BenchmarkGenerateEd25519KeyPair benchmarks Ed25519 key pair generation.
func BenchmarkGenerateEd25519KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateEDDSAKeyPair(cryptoutilSharedMagic.EdCurveEd25519)
		require.NoError(b, err, "GenerateEDDSAKeyPair should not fail")
	}
}

// BenchmarkGenerateEd448KeyPair benchmarks Ed448 key pair generation.
func BenchmarkGenerateEd448KeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateEDDSAKeyPair(cryptoutilSharedMagic.EdCurveEd448)
		require.NoError(b, err, "GenerateEDDSAKeyPair should not fail")
	}
}

// BenchmarkGenerateAES128Key benchmarks AES-128 key generation.
func BenchmarkGenerateAES128Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateAESKey(cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits)
		require.NoError(b, err, "GenerateAESKey should not fail")
	}
}

// BenchmarkGenerateAES192Key benchmarks AES-192 key generation.
func BenchmarkGenerateAES192Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateAESKey(cryptoutilSharedMagic.SymmetricKeySize192)
		require.NoError(b, err, "GenerateAESKey should not fail")
	}
}

// BenchmarkGenerateAES256Key benchmarks AES-256 key generation.
func BenchmarkGenerateAES256Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateAESKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
		require.NoError(b, err, "GenerateAESKey should not fail")
	}
}

// BenchmarkGenerateHMAC256Key benchmarks HMAC-256 key generation.
func BenchmarkGenerateHMAC256Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
		require.NoError(b, err, "GenerateHMACKey should not fail")
	}
}

// BenchmarkGenerateHMAC384Key benchmarks HMAC-384 key generation.
func BenchmarkGenerateHMAC384Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateHMACKey(cryptoutilSharedMagic.SymmetricKeySize384)
		require.NoError(b, err, "GenerateHMACKey should not fail")
	}
}

// BenchmarkGenerateHMAC512Key benchmarks HMAC-512 key generation.
func BenchmarkGenerateHMAC512Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateHMACKey(cryptoutilSharedMagic.DefaultTracesBatchSize)
		require.NoError(b, err, "GenerateHMACKey should not fail")
	}
}
