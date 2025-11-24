package authz_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"

	identityAuthz "cryptoutil/internal/identity/authz"
	identityDomain "cryptoutil/internal/identity/domain"
	identityTestutils "cryptoutil/internal/identity/test/testutils"
)

// BenchmarkTokenIssuance measures OAuth 2.1 token issuance performance.
func BenchmarkTokenIssuance(b *testing.B) {
	ctx := context.Background()

	// Setup test database and repository.
	repoFactory := identityTestutils.SetupTestDatabase(b, ctx)
	b.Cleanup(func() {
		_ = repoFactory.Close()
	})

	// Create test client with authorization code grant.
	client := &identityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                googleUuid.NewString(),
		ClientSecret:            googleUuid.NewString(),
		TokenEndpointAuthMethod: identityDomain.ClientAuthMethodSecretPost,
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		RedirectURIs:            []string{"https://client.example.com/callback"},
		Scopes:                  []string{"openid", "profile", "email"},
		Active:                  true,
	}
	err := repoFactory.ClientRepository().Create(ctx, client)
	if err != nil {
		b.Fatalf("failed to create test client: %v", err)
	}

	// Create test user.
	user := &identityDomain.User{
		ID:       googleUuid.Must(googleUuid.NewV7()),
		Username: "bench_user",
		Email:    "bench@example.com",
		Active:   true,
	}
	err = repoFactory.UserRepository().Create(ctx, user)
	if err != nil {
		b.Fatalf("failed to create test user: %v", err)
	}

	// Create authorization request.
	authzReq := &identityDomain.AuthorizationRequest{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     client.ClientID,
		UserID:       user.ID,
		ResponseType: "code",
		RedirectURI:  client.RedirectURIs[0],
		Scopes:       client.Scopes,
		State:        googleUuid.NewString(),
		Active:       true,
	}
	err = repoFactory.AuthorizationRequestRepository().Create(ctx, authzReq)
	if err != nil {
		b.Fatalf("failed to create authz request: %v", err)
	}

	// Create token service.
	tokenService := &identityAuthz.TokenService{
		RepoFactory: repoFactory,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark token issuance.
		_, err := tokenService.IssueAccessToken(ctx, client, user, authzReq.Scopes)
		if err != nil {
			b.Fatalf("token issuance failed: %v", err)
		}
	}
}

// BenchmarkJWTValidation measures JWT signature validation performance.
func BenchmarkJWTValidation(b *testing.B) {
	ctx := context.Background()

	// Generate RSA key pair for signing.
	privateKey, err := joseJwk.FromRaw(generateRSAPrivateKey(b))
	if err != nil {
		b.Fatalf("failed to create private JWK: %v", err)
	}

	if err := privateKey.Set(joseJwk.AlgorithmKey, joseJwa.RS256()); err != nil {
		b.Fatalf("failed to set algorithm: %v", err)
	}

	publicKey, err := privateKey.PublicKey()
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

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateKey))
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

// BenchmarkJWKSRetrieval measures JWKS endpoint performance (placeholder).
func BenchmarkJWKSRetrieval(b *testing.B) {
	ctx := context.Background()

	// Setup test database.
	repoFactory := identityTestutils.SetupTestDatabase(b, ctx)
	b.Cleanup(func() {
		_ = repoFactory.Close()
	})

	// Create test signing keys.
	for i := 0; i < 3; i++ {
		privateKey := generateRSAPrivateKey(b)
		publicJWK, err := joseJwk.FromRaw(privateKey.Public())
		if err != nil {
			b.Fatalf("failed to create public JWK: %v", err)
		}
		publicBytes, err := joseJwk.EncodePEM(publicJWK)
		if err != nil {
			b.Fatalf("failed to encode public key: %v", err)
		}

		key := &identityDomain.Key{
			ID:        googleUuid.Must(googleUuid.NewV7()),
			Usage:     "signing",
			Algorithm: "RS256",
			PublicKey: string(publicBytes),
			Active:    true,
		}
		if err := repoFactory.KeyRepository().Create(ctx, key); err != nil {
			b.Fatalf("failed to create key: %v", err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark JWKS retrieval.
		keys, err := repoFactory.KeyRepository().FindByUsage(ctx, "signing", true)
		if err != nil {
			b.Fatalf("JWKS retrieval failed: %v", err)
		}
		if len(keys) != 3 {
			b.Fatalf("expected 3 keys, got %d", len(keys))
		}
	}
}

// generateRSAPrivateKey generates an RSA private key for benchmarking.
func generateRSAPrivateKey(b *testing.B) interface{} {
	b.Helper()

	key, err := joseJwk.GenerateKey(joseJwa.RSA, joseJwk.WithKeySize(2048))
	if err != nil {
		b.Fatalf("failed to generate RSA key: %v", err)
	}

	var rawKey interface{}
	if err := key.Raw(&rawKey); err != nil {
		b.Fatalf("failed to get raw key: %v", err)
	}

	return rawKey
}
