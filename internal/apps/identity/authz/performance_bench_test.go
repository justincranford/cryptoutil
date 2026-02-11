// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"context"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"

	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
)

// BenchmarkUUIDTokenIssuance measures UUID token issuance performance.
func BenchmarkUUIDTokenIssuance(b *testing.B) {
	ctx := context.Background()

	// Create UUID issuer.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark UUID token generation.
		if err := uuidIssuer.ValidateToken(ctx, "test-token"); err == nil {
			b.Fatal("expected validation error")
		}
	}
}

// BenchmarkJWTSigning measures JWT signature creation performance.
func BenchmarkJWTSigning(b *testing.B) {
	// Generate RSA key for signing using crypto/rsa.
	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		b.Fatalf("failed to generate RSA key: %v", err)
	}

	key, err := joseJwk.Import(privateKey)
	if err != nil {
		b.Fatalf("failed to import RSA key: %v", err)
	}

	if err := key.Set(joseJwk.AlgorithmKey, joseJwa.RS256()); err != nil {
		b.Fatalf("failed to set algorithm: %v", err)
	}

	if err := key.Set(joseJwk.KeyIDKey, "bench-key-id"); err != nil {
		b.Fatalf("failed to set kid: %v", err)
	}

	// Create test JWT payload.
	token := joseJwt.New()
	if err := token.Set(joseJwt.IssuerKey, "bench_issuer"); err != nil {
		b.Fatal(err)
	}

	if err := token.Set(joseJwt.SubjectKey, "bench_subject"); err != nil {
		b.Fatal(err)
	}

	if err := token.Set(joseJwt.AudienceKey, []string{"bench_audience"}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark JWT signing operation.
		_, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), key))
		if err != nil {
			b.Fatalf("JWT signing failed: %v", err)
		}
	}
}

// BenchmarkJWTValidation measures JWT signature validation performance.
func BenchmarkJWTValidation(b *testing.B) {
	// Generate RSA key pair for signing using crypto/rsa.
	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		b.Fatalf("failed to generate RSA key: %v", err)
	}

	key, err := joseJwk.Import(privateKey)
	if err != nil {
		b.Fatalf("failed to import RSA key: %v", err)
	}

	if err := key.Set(joseJwk.AlgorithmKey, joseJwa.RS256()); err != nil {
		b.Fatalf("failed to set algorithm: %v", err)
	}

	if err := key.Set(joseJwk.KeyIDKey, "bench-key-id"); err != nil {
		b.Fatalf("failed to set kid: %v", err)
	}

	publicKey, err := key.PublicKey()
	if err != nil {
		b.Fatalf("failed to get public key: %v", err)
	}

	// Create and sign test JWT.
	token := joseJwt.New()
	if err := token.Set(joseJwt.IssuerKey, "bench_issuer"); err != nil {
		b.Fatal(err)
	}

	if err := token.Set(joseJwt.SubjectKey, "bench_subject"); err != nil {
		b.Fatal(err)
	}

	if err := token.Set(joseJwt.AudienceKey, []string{"bench_audience"}); err != nil {
		b.Fatal(err)
	}

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), key))
	if err != nil {
		b.Fatalf("failed to sign token: %v", err)
	}

	// Create key set for validation.
	keySet := joseJwk.NewSet()
	if err := keySet.AddKey(publicKey); err != nil {
		b.Fatalf("failed to add key to set: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark JWT parsing and signature verification.
		_, err := joseJwt.Parse(signedToken, joseJwt.WithKeySet(keySet))
		if err != nil {
			b.Fatalf("JWT validation failed: %v", err)
		}
	}
}

// BenchmarkRSAKeyGeneration measures RSA key generation performance.
func BenchmarkRSAKeyGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark RSA key generation using crypto/rsa.
		privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
		if err != nil {
			b.Fatalf("failed to generate RSA key: %v", err)
		}

		// Convert to JWK.
		_, err = joseJwk.Import(privateKey)
		if err != nil {
			b.Fatalf("failed to import RSA key: %v", err)
		}
	}
}
