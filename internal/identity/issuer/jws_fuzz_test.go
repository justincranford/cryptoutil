// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	"testing"
	"time"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// FuzzJWSTokenParsing tests JWS token parsing with various inputs.
func FuzzJWSTokenParsing(f *testing.F) {
	// Seed corpus with valid and invalid tokens.
	f.Add("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.signature")
	f.Add("invalid.token")
	f.Add("")
	f.Add("a.b.c.d.e")
	f.Add("...")
	f.Add("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTl9.signature")
	f.Add("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjF9.signature")

	// Create legacy JWS issuer for testing.
	issuer, err := NewJWSIssuerLegacy(
		"https://test.example.com",
		[]byte("test-signing-key"),
		"RS256",
		cryptoutilIdentityMagic.DefaultAccessTokenLifetime,
		cryptoutilIdentityMagic.DefaultIDTokenLifetime,
	)
	if err != nil {
		f.Fatalf("failed to create issuer: %v", err)
	}

	ctx := context.Background()

	f.Fuzz(func(_ *testing.T, token string) {
		// Validate token - should not panic.
		_, err := issuer.ValidateToken(ctx, token)

		// We don't care about the error, just that it doesn't panic.
		_ = err
	})
}

// FuzzJWSClaimsMarshaling tests JWT claims marshaling with various claim values.
func FuzzJWSClaimsMarshaling(f *testing.F) {
	// Seed corpus with various claim structures.
	f.Add("test-sub", "test-aud", "openid profile")
	f.Add("", "", "")
	f.Add("very-long-subject-"+string(make([]byte, 1000)), "aud", "scope")
	f.Add("sub", "aud-with-special-chars-!@#$%^&*()", "scope1 scope2 scope3")

	// Create legacy JWS issuer.
	issuer, err := NewJWSIssuerLegacy(
		"https://test.example.com",
		[]byte("test-signing-key"),
		"RS256",
		cryptoutilIdentityMagic.DefaultAccessTokenLifetime,
		cryptoutilIdentityMagic.DefaultIDTokenLifetime,
	)
	if err != nil {
		f.Fatalf("failed to create issuer: %v", err)
	}

	ctx := context.Background()

	f.Fuzz(func(_ *testing.T, sub, aud, scope string) {
		claims := map[string]any{
			cryptoutilIdentityMagic.ClaimSub:   sub,
			cryptoutilIdentityMagic.ClaimAud:   aud,
			cryptoutilIdentityMagic.ParamScope: scope,
		}

		// Issue token - should not panic.
		_, err := issuer.IssueAccessToken(ctx, claims)

		// We don't care about the error, just that it doesn't panic.
		_ = err
	})
}

// FuzzJWSIDTokenGeneration tests ID token generation with various OIDC claims.
// cspell:ignore JWSID
func FuzzJWSIDTokenGeneration(f *testing.F) {
	// Seed corpus.
	f.Add("user123", "client456", "John Doe", "john@example.com")
	f.Add("", "", "", "")
	f.Add("sub", "aud", "", "")
	f.Add("very-long-"+string(make([]byte, 500)), "aud", "name", "email@test.com")

	// Create legacy JWS issuer.
	issuer, err := NewJWSIssuerLegacy(
		"https://test.example.com",
		[]byte("test-signing-key"),
		"RS256",
		cryptoutilIdentityMagic.DefaultAccessTokenLifetime,
		cryptoutilIdentityMagic.DefaultIDTokenLifetime,
	)
	if err != nil {
		f.Fatalf("failed to create issuer: %v", err)
	}

	ctx := context.Background()

	f.Fuzz(func(_ *testing.T, sub, aud, name, email string) {
		claims := map[string]any{
			cryptoutilIdentityMagic.ClaimSub: sub,
			cryptoutilIdentityMagic.ClaimAud: aud,
		}

		if name != "" {
			claims["name"] = name
		}

		if email != "" {
			claims["email"] = email
		}

		// Issue ID token - should not panic.
		_, err := issuer.IssueIDToken(ctx, claims)

		// We don't care about the error, just that it doesn't panic.
		_ = err
	})
}

// FuzzJWSExpirationValidation tests token expiration validation.
func FuzzJWSExpirationValidation(f *testing.F) {
	// Seed with various expiration times.
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(time.Now().UTC().Add(1 * time.Hour).Unix())
	f.Add(time.Now().UTC().Add(-1 * time.Hour).Unix())
	f.Add(int64(9999999999))

	// Create legacy JWS issuer.
	issuer, err := NewJWSIssuerLegacy(
		"https://test.example.com",
		[]byte("test-signing-key"),
		"RS256",
		cryptoutilIdentityMagic.DefaultAccessTokenLifetime,
		cryptoutilIdentityMagic.DefaultIDTokenLifetime,
	)
	if err != nil {
		f.Fatalf("failed to create issuer: %v", err)
	}

	ctx := context.Background()

	f.Fuzz(func(_ *testing.T, exp int64) {
		claims := map[string]any{
			cryptoutilIdentityMagic.ClaimSub: "test-user",
			cryptoutilIdentityMagic.ClaimAud: "test-client",
			cryptoutilIdentityMagic.ClaimExp: float64(exp),
		}

		// Issue token.
		token, err := issuer.buildJWS(claims)
		if err != nil {
			return
		}

		// Validate token - should not panic.
		_, err = issuer.ValidateToken(ctx, token)

		// We don't care about the error, just that it doesn't panic.
		_ = err
	})
}
