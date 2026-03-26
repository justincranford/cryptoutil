// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic_test

import (
	"crypto/ecdh"
	"crypto/elliptic"
	crand "crypto/rand"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
)

// P2.3.4: KMS Performance Baseline Benchmarks.
// These benchmarks establish performance baselines for KMS cryptographic operations.
// Note: Full E2E benchmarks require complex stack setup - these focus on core crypto.

const aes256Bits = 256

// BenchmarkAESKeyGeneration measures AES key generation performance.
func BenchmarkAESKeyGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(aes256Bits)
		if err != nil {
			b.Fatalf("AES key generation failed: %v", err)
		}
	}
}

// BenchmarkECDSAKeyGeneration measures ECDSA key generation performance.
func BenchmarkECDSAKeyGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
		if err != nil {
			b.Fatalf("ECDSA key generation failed: %v", err)
		}
	}
}

// BenchmarkECDHKeyGeneration measures ECDH key generation performance.
func BenchmarkECDHKeyGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cryptoutilSharedCryptoKeygen.GenerateECDHKeyPair(ecdh.P256())
		if err != nil {
			b.Fatalf("ECDH key generation failed: %v", err)
		}
	}
}

// BenchmarkRSAKeyGeneration measures RSA key generation performance.
func BenchmarkRSAKeyGeneration(b *testing.B) {
	const rsaBits = 2048

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(rsaBits)
		if err != nil {
			b.Fatalf("RSA key generation failed: %v", err)
		}
	}
}

// BenchmarkEdDSAKeyGeneration measures EdDSA key generation performance.
func BenchmarkEdDSAKeyGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair(cryptoutilSharedCryptoKeygen.EdCurveEd25519)
		if err != nil {
			b.Fatalf("EdDSA key generation failed: %v", err)
		}
	}
}

// BenchmarkJWKSign_ES256 measures JWT signing performance with ES256.
func BenchmarkJWKSign_ES256(b *testing.B) {
	// Generate ECDSA key for signing.
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	if err != nil {
		b.Fatalf("failed to generate ECDSA key: %v", err)
	}

	jwk, err := joseJwk.Import(keyPair.Private)
	if err != nil {
		b.Fatalf("failed to import key: %v", err)
	}

	if err := jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES256()); err != nil {
		b.Fatalf("failed to set algorithm: %v", err)
	}

	if err := jwk.Set(joseJwk.KeyIDKey, "bench-es256-key"); err != nil {
		b.Fatalf("failed to set kid: %v", err)
	}

	// Test payload.
	const payloadSize = 1024

	payload := make([]byte, payloadSize)

	if _, err := crand.Read(payload); err != nil {
		b.Fatalf("failed to generate payload: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := joseJws.Sign(payload, joseJws.WithKey(joseJwa.ES256(), jwk))
		if err != nil {
			b.Fatalf("JWS signing failed: %v", err)
		}
	}
}

// BenchmarkJWKVerify_ES256 measures JWT verification performance with ES256.
func BenchmarkJWKVerify_ES256(b *testing.B) {
	// Generate ECDSA key for signing.
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	if err != nil {
		b.Fatalf("failed to generate ECDSA key: %v", err)
	}

	jwk, err := joseJwk.Import(keyPair.Private)
	if err != nil {
		b.Fatalf("failed to import key: %v", err)
	}

	if err := jwk.Set(joseJwk.AlgorithmKey, joseJwa.ES256()); err != nil {
		b.Fatalf("failed to set algorithm: %v", err)
	}

	if err := jwk.Set(joseJwk.KeyIDKey, "bench-es256-key"); err != nil {
		b.Fatalf("failed to set kid: %v", err)
	}

	// Sign a payload first.
	const payloadSize = 1024

	payload := make([]byte, payloadSize)

	if _, err := crand.Read(payload); err != nil {
		b.Fatalf("failed to generate payload: %v", err)
	}

	signature, err := joseJws.Sign(payload, joseJws.WithKey(joseJwa.ES256(), jwk))
	if err != nil {
		b.Fatalf("failed to sign payload: %v", err)
	}

	// Get public key for verification.
	pubKey, err := jwk.PublicKey()
	if err != nil {
		b.Fatalf("failed to get public key: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := joseJws.Verify(signature, joseJws.WithKey(joseJwa.ES256(), pubKey))
		if err != nil {
			b.Fatalf("JWS verification failed: %v", err)
		}
	}
}

// BenchmarkHMACKeyGeneration measures HMAC key generation performance.
func BenchmarkHMACKeyGeneration(b *testing.B) {
	const hmac256Bits = 256

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(hmac256Bits)
		if err != nil {
			b.Fatalf("HMAC key generation failed: %v", err)
		}
	}
}

// BenchmarkKeyGenerationParallel measures parallel key generation performance.
func BenchmarkKeyGenerationParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(aes256Bits)
			if err != nil {
				b.Fatalf("AES key generation failed: %v", err)
			}
		}
	})
}

// BenchmarkPayloadSizes measures key generation for different algorithms.
func BenchmarkPayloadSizes(b *testing.B) {
	b.Run("AES128", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits)
			if err != nil {
				b.Fatalf("AES128 generation failed: %v", err)
			}
		}
	})

	b.Run("AES192", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.SymmetricKeySize192)
			if err != nil {
				b.Fatalf("AES192 generation failed: %v", err)
			}
		}
	})

	b.Run("AES256", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
			if err != nil {
				b.Fatalf("AES256 generation failed: %v", err)
			}
		}
	})

	b.Run("ECDSA_P256", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
			if err != nil {
				b.Fatalf("ECDSA_P256 generation failed: %v", err)
			}
		}
	})

	b.Run("ECDSA_P384", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
			if err != nil {
				b.Fatalf("ECDSA_P384 generation failed: %v", err)
			}
		}
	})
}

// BenchmarkJWKCreation measures JWK creation performance from raw keys.
func BenchmarkJWKCreation(b *testing.B) {
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jwk, err := joseJwk.Import(keyPair.Private)
		if err != nil {
			b.Fatalf("JWK import failed: %v", err)
		}

		kid := fmt.Sprintf("key-%d", i)
		if err := jwk.Set(joseJwk.KeyIDKey, kid); err != nil {
			b.Fatalf("failed to set kid: %v", err)
		}
	}
}
