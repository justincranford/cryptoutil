// Copyright (c) 2025 Justin Cranford

package issuer

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"
)

// BenchmarkJWSIssueAccessToken benchmarks JWS access token generation.
func BenchmarkJWSIssueAccessToken(b *testing.B) {
	ctx := context.Background()

	keyRotationMgr, err := NewKeyRotationManager(
		DefaultKeyRotationPolicy(),
		NewProductionKeyGenerator(),
		nil,
	)
	if err != nil {
		b.Fatalf("failed to create key rotation manager: %v", err)
	}

	jwsIssuer, err := NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create JWS issuer: %v", err)
	}

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:   "user123",
		cryptoutilSharedMagic.ClaimAud:   "test-audience",
		cryptoutilSharedMagic.ClaimScope: "read write",
		cryptoutilSharedMagic.ClaimExp:   1234567890,
		cryptoutilSharedMagic.ClaimIat:   1234567800,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := jwsIssuer.IssueAccessToken(ctx, claims)
		if err != nil {
			b.Fatalf("failed to issue access token: %v", err)
		}
	}
}

// BenchmarkJWSIssueIDToken benchmarks JWS ID token generation.
func BenchmarkJWSIssueIDToken(b *testing.B) {
	ctx := context.Background()

	keyRotationMgr, err := NewKeyRotationManager(
		DefaultKeyRotationPolicy(),
		NewProductionKeyGenerator(),
		nil,
	)
	if err != nil {
		b.Fatalf("failed to create key rotation manager: %v", err)
	}

	jwsIssuer, err := NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create JWS issuer: %v", err)
	}

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:   "user123",
		cryptoutilSharedMagic.ClaimAud:   "client-app",
		cryptoutilSharedMagic.ClaimNonce: "random-nonce",
		cryptoutilSharedMagic.ClaimExp:   1234567890,
		cryptoutilSharedMagic.ClaimIat:   1234567800,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := jwsIssuer.IssueIDToken(ctx, claims)
		if err != nil {
			b.Fatalf("failed to issue ID token: %v", err)
		}
	}
}

// BenchmarkJWSValidateToken benchmarks JWS token validation.
func BenchmarkJWSValidateToken(b *testing.B) {
	ctx := context.Background()

	keyRotationMgr, err := NewKeyRotationManager(
		DefaultKeyRotationPolicy(),
		NewProductionKeyGenerator(),
		nil,
	)
	if err != nil {
		b.Fatalf("failed to create key rotation manager: %v", err)
	}

	jwsIssuer, err := NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create JWS issuer: %v", err)
	}

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub: "user123",
		cryptoutilSharedMagic.ClaimAud: "test-audience",
		cryptoutilSharedMagic.ClaimExp: 9999999999,
		cryptoutilSharedMagic.ClaimIat: 1234567800,
	}

	token, err := jwsIssuer.IssueAccessToken(ctx, claims)
	if err != nil {
		b.Fatalf("failed to issue access token: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := jwsIssuer.ValidateToken(ctx, token)
		if err != nil {
			b.Fatalf("failed to validate token: %v", err)
		}
	}
}
