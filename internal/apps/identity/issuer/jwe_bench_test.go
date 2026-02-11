// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"context"
	"testing"
)

const benchmarkPlaintext = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoidGVzdC1hdWRpZW5jZSJ9.signature"

// BenchmarkJWEEncryptToken benchmarks JWE token encryption.
func BenchmarkJWEEncryptToken(b *testing.B) {
	ctx := context.Background()

	keyRotationMgr, err := NewKeyRotationManager(
		DefaultKeyRotationPolicy(),
		NewProductionKeyGenerator(),
		nil,
	)
	if err != nil {
		b.Fatalf("failed to create key rotation manager: %v", err)
	}

	jweIssuer, err := NewJWEIssuer(keyRotationMgr)
	if err != nil {
		b.Fatalf("failed to create JWE issuer: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := jweIssuer.EncryptToken(ctx, benchmarkPlaintext)
		if err != nil {
			b.Fatalf("failed to encrypt token: %v", err)
		}
	}
}

// BenchmarkJWEDecryptToken benchmarks JWE token decryption.
func BenchmarkJWEDecryptToken(b *testing.B) {
	ctx := context.Background()

	keyRotationMgr, err := NewKeyRotationManager(
		DefaultKeyRotationPolicy(),
		NewProductionKeyGenerator(),
		nil,
	)
	if err != nil {
		b.Fatalf("failed to create key rotation manager: %v", err)
	}

	jweIssuer, err := NewJWEIssuer(keyRotationMgr)
	if err != nil {
		b.Fatalf("failed to create JWE issuer: %v", err)
	}

	encrypted, err := jweIssuer.EncryptToken(ctx, benchmarkPlaintext)
	if err != nil {
		b.Fatalf("failed to encrypt token: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := jweIssuer.DecryptToken(ctx, encrypted)
		if err != nil {
			b.Fatalf("failed to decrypt token: %v", err)
		}
	}
}

// BenchmarkJWERoundTrip benchmarks full JWE encrypt/decrypt cycle.
func BenchmarkJWERoundTrip(b *testing.B) {
	ctx := context.Background()

	keyRotationMgr, err := NewKeyRotationManager(
		DefaultKeyRotationPolicy(),
		NewProductionKeyGenerator(),
		nil,
	)
	if err != nil {
		b.Fatalf("failed to create key rotation manager: %v", err)
	}

	jweIssuer, err := NewJWEIssuer(keyRotationMgr)
	if err != nil {
		b.Fatalf("failed to create JWE issuer: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		encrypted, encErr := jweIssuer.EncryptToken(ctx, benchmarkPlaintext)
		if encErr != nil {
			b.Fatalf("failed to encrypt token: %v", encErr)
		}

		_, decErr := jweIssuer.DecryptToken(ctx, encrypted)
		if decErr != nil {
			b.Fatalf("failed to decrypt token: %v", decErr)
		}
	}
}
