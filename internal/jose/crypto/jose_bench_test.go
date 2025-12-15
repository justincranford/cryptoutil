// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

const benchmarkPayload = `{"sub":"test-user","aud":"test-audience","exp":1735689600,"data":"Lorem ipsum dolor sit amet"}`

// BenchmarkJWSSign_ES256 measures JWT signing with ECDSA P-256.
func BenchmarkJWSSign_ES256(b *testing.B) {
	alg := joseJwa.ES256()
	_, jwk, _, _, _, err := GenerateJWSJWKForAlg(&alg)
	require.NoError(b, err)

	jwks := []joseJwk.Key{jwk}
	clearBytes := []byte(benchmarkPayload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := SignBytes(jwks, clearBytes)
		if err != nil {
			b.Fatalf("failed to sign bytes: %v", err)
		}
	}
}

// BenchmarkJWSVerify_ES256 measures JWT verification with ECDSA P-256.
func BenchmarkJWSVerify_ES256(b *testing.B) {
	alg := joseJwa.ES256()
	_, signJWK, verifyJWK, _, _, err := GenerateJWSJWKForAlg(&alg)
	require.NoError(b, err)

	signJWKs := []joseJwk.Key{signJWK}
	verifyJWKs := []joseJwk.Key{verifyJWK}
	clearBytes := []byte(benchmarkPayload)

	_, jwsMessageBytes, err := SignBytes(signJWKs, clearBytes)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := VerifyBytes(verifyJWKs, jwsMessageBytes)
		if err != nil {
			b.Fatalf("failed to verify bytes: %v", err)
		}
	}
}

// BenchmarkJWSSign_RS256 measures JWT signing with RSA 2048.
func BenchmarkJWSSign_RS256(b *testing.B) {
	alg := joseJwa.RS256()
	_, jwk, _, _, _, err := GenerateJWSJWKForAlg(&alg)
	require.NoError(b, err)

	jwks := []joseJwk.Key{jwk}
	clearBytes := []byte(benchmarkPayload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := SignBytes(jwks, clearBytes)
		if err != nil {
			b.Fatalf("failed to sign bytes: %v", err)
		}
	}
}

// BenchmarkJWSVerify_RS256 measures JWT verification with RSA 2048.
func BenchmarkJWSVerify_RS256(b *testing.B) {
	alg := joseJwa.RS256()
	_, signJWK, verifyJWK, _, _, err := GenerateJWSJWKForAlg(&alg)
	require.NoError(b, err)

	signJWKs := []joseJwk.Key{signJWK}
	verifyJWKs := []joseJwk.Key{verifyJWK}
	clearBytes := []byte(benchmarkPayload)

	_, jwsMessageBytes, err := SignBytes(signJWKs, clearBytes)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := VerifyBytes(verifyJWKs, jwsMessageBytes)
		if err != nil {
			b.Fatalf("failed to verify bytes: %v", err)
		}
	}
}

// BenchmarkJWEEncrypt_A256GCM measures JWE encryption with AES-256-GCM.
func BenchmarkJWEEncrypt_A256GCM(b *testing.B) {
	alg := joseJwa.A256GCMKW()
	enc := joseJwa.A256GCM()
	_, jwk, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(b, err)

	jwks := []joseJwk.Key{jwk}
	clearBytes := []byte(benchmarkPayload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := EncryptBytes(jwks, clearBytes)
		if err != nil {
			b.Fatalf("failed to encrypt bytes: %v", err)
		}
	}
}

// BenchmarkJWEDecrypt_A256GCM measures JWE decryption with AES-256-GCM.
func BenchmarkJWEDecrypt_A256GCM(b *testing.B) {
	alg := joseJwa.A256GCMKW()
	enc := joseJwa.A256GCM()
	_, jwk, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(b, err)

	jwks := []joseJwk.Key{jwk}
	clearBytes := []byte(benchmarkPayload)

	_, jweMessageBytes, err := EncryptBytes(jwks, clearBytes)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DecryptBytes(jwks, jweMessageBytes)
		if err != nil {
			b.Fatalf("failed to decrypt bytes: %v", err)
		}
	}
}

// BenchmarkJWEEncrypt_RSA_OAEP measures JWE encryption with RSA-OAEP.
func BenchmarkJWEEncrypt_RSA_OAEP(b *testing.B) {
	alg := joseJwa.RSA_OAEP()
	enc := joseJwa.A256GCM()
	_, privateJWK, publicJWK, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(b, err)

	_ = privateJWK // Not used in encrypt-only benchmark
	jwks := []joseJwk.Key{publicJWK}
	clearBytes := []byte(benchmarkPayload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := EncryptBytes(jwks, clearBytes)
		if err != nil {
			b.Fatalf("failed to encrypt bytes: %v", err)
		}
	}
}

// BenchmarkJWEDecrypt_RSA_OAEP measures JWE decryption with RSA-OAEP.
func BenchmarkJWEDecrypt_RSA_OAEP(b *testing.B) {
	alg := joseJwa.RSA_OAEP()
	enc := joseJwa.A256GCM()
	_, privateJWK, publicJWK, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(b, err)

	encryptJWKs := []joseJwk.Key{publicJWK}
	decryptJWKs := []joseJwk.Key{privateJWK}
	clearBytes := []byte(benchmarkPayload)

	_, jweMessageBytes, err := EncryptBytes(encryptJWKs, clearBytes)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DecryptBytes(decryptJWKs, jweMessageBytes)
		if err != nil {
			b.Fatalf("failed to decrypt bytes: %v", err)
		}
	}
}

// BenchmarkJWEEncrypt_ECDH_ES measures JWE encryption with ECDH-ES+A256KW.
func BenchmarkJWEEncrypt_ECDH_ES(b *testing.B) {
	alg := joseJwa.ECDH_ES_A256KW()
	enc := joseJwa.A256GCM()
	_, privateJWK, publicJWK, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(b, err)

	_ = privateJWK // Not used in encrypt-only benchmark
	jwks := []joseJwk.Key{publicJWK}
	clearBytes := []byte(benchmarkPayload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := EncryptBytes(jwks, clearBytes)
		if err != nil {
			b.Fatalf("failed to encrypt bytes: %v", err)
		}
	}
}

// BenchmarkJWEDecrypt_ECDH_ES measures JWE decryption with ECDH-ES+A256KW.
func BenchmarkJWEDecrypt_ECDH_ES(b *testing.B) {
	alg := joseJwa.ECDH_ES_A256KW()
	enc := joseJwa.A256GCM()
	_, privateJWK, publicJWK, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(b, err)

	encryptJWKs := []joseJwk.Key{publicJWK}
	decryptJWKs := []joseJwk.Key{privateJWK}
	clearBytes := []byte(benchmarkPayload)

	_, jweMessageBytes, err := EncryptBytes(encryptJWKs, clearBytes)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DecryptBytes(decryptJWKs, jweMessageBytes)
		if err != nil {
			b.Fatalf("failed to decrypt bytes: %v", err)
		}
	}
}

// BenchmarkJWSRoundTrip_ES256 measures full sign+verify cycle with ECDSA P-256.
func BenchmarkJWSRoundTrip_ES256(b *testing.B) {
	alg := joseJwa.ES256()
	_, signJWK, verifyJWK, _, _, err := GenerateJWSJWKForAlg(&alg)
	require.NoError(b, err)

	signJWKs := []joseJwk.Key{signJWK}
	verifyJWKs := []joseJwk.Key{verifyJWK}
	clearBytes := []byte(benchmarkPayload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, jwsMessageBytes, err := SignBytes(signJWKs, clearBytes)
		if err != nil {
			b.Fatalf("failed to sign bytes: %v", err)
		}

		_, err = VerifyBytes(verifyJWKs, jwsMessageBytes)
		if err != nil {
			b.Fatalf("failed to verify bytes: %v", err)
		}
	}
}

// BenchmarkJWERoundTrip_A256GCM measures full encrypt+decrypt cycle with AES-256-GCM.
func BenchmarkJWERoundTrip_A256GCM(b *testing.B) {
	alg := joseJwa.A256GCMKW()
	enc := joseJwa.A256GCM()
	_, jwk, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(b, err)

	jwks := []joseJwk.Key{jwk}
	clearBytes := []byte(benchmarkPayload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, jweMessageBytes, err := EncryptBytes(jwks, clearBytes)
		if err != nil {
			b.Fatalf("failed to encrypt bytes: %v", err)
		}

		_, err = DecryptBytes(jwks, jweMessageBytes)
		if err != nil {
			b.Fatalf("failed to decrypt bytes: %v", err)
		}
	}
}
