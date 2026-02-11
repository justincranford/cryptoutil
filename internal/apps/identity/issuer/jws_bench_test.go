// Copyright (c) 2025 Justin Cranford

package issuer

import (
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
		"RS256",
		1*time.Hour,
		1*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create JWS issuer: %v", err)
	}

	claims := map[string]any{
		"sub":   "user123",
		"aud":   "test-audience",
		"scope": "read write",
		"exp":   1234567890,
		"iat":   1234567800,
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
		"RS256",
		1*time.Hour,
		1*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create JWS issuer: %v", err)
	}

	claims := map[string]any{
		"sub":   "user123",
		"aud":   "client-app",
		"nonce": "random-nonce",
		"exp":   1234567890,
		"iat":   1234567800,
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
		"RS256",
		1*time.Hour,
		1*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create JWS issuer: %v", err)
	}

	claims := map[string]any{
		"sub": "user123",
		"aud": "test-audience",
		"exp": 9999999999,
		"iat": 1234567800,
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
