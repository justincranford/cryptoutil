// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides a demonstration of the identity service capabilities.
package demo

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	http "net/http"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var demoClientSecret = "demo-secret-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength]

var (
	outWriter io.Writer
	errWriter io.Writer
)

// Demo runs the identity service demonstration.
// args: Command-line arguments (including program name)
// stdin: Input stream (unused)
// stdout, stderr: Output streams for messages
// Returns: Exit code (0 for success, non-zero for errors).
func Demo(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cmdArgs := args[1:]

	return demo(cmdArgs, stdout, stderr)
}

// demo runs the identity service demonstration.
// args: Command-line arguments (not including program name)
// stdout, stderr: Output streams for messages
// Returns: Exit code (0 for success, non-zero for errors).
func demo(args []string, stdout, stderr io.Writer) int {
	_, _ = fmt.Fprintln(outWriter, "🚀 Identity System Demo - OAuth 2.1 Authorization Server")
	outWriter = stdout
	errWriter = stderr

	_, _ = fmt.Fprintln(outWriter, "================================================================")
	_, _ = fmt.Fprintln(stdout)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DemoTimeout)
	defer cancel()

	// Run the demo.
	if err := runDemo(ctx); err != nil {
		_, _ = fmt.Fprintf(errWriter, "❌ Demo failed: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprintln(outWriter)
	_, _ = fmt.Fprintln(outWriter, "✅ Demo completed successfully!")

	return 0
}

func runDemo(ctx context.Context) error {
	// Step 1: Start AuthZ server.
	_, _ = fmt.Fprintln(outWriter, "📦 Step 1: Starting Authorization Server...")

	app, repoFactory, cleanup, err := startAuthZServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	defer cleanup()

	// Start Fiber in background.
	go func() {
		if listenErr := app.Listen(cryptoutilSharedMagic.DemoPort); listenErr != nil {
			_, _ = fmt.Fprintf(outWriter, "Server error: %v\n", listenErr)
		}
	}()

	// Give server time to start.
	time.Sleep(cryptoutilSharedMagic.DemoStartupDelay)

	_, _ = fmt.Fprintln(outWriter, "✅ Authorization server started on http://127.0.0.1:8080")
	_, _ = fmt.Fprintln(outWriter)

	// Step 2: Register demo client.
	_, _ = fmt.Fprintln(outWriter, "📝 Step 2: Registering Demo Client...")

	if err := registerDemoClient(ctx, repoFactory); err != nil {
		return fmt.Errorf("failed to register client: %w", err)
	}

	_, _ = fmt.Fprintf(outWriter, "   ✅ Client ID: %s\n", cryptoutilSharedMagic.DemoClientID)
	_, _ = fmt.Fprintf(outWriter, "   ✅ Client Name: %s\n", cryptoutilSharedMagic.DemoClientName)
	_, _ = fmt.Fprintf(outWriter, "   ✅ Redirect URI: %s\n", cryptoutilSharedMagic.DemoRedirectURI)
	_, _ = fmt.Fprintln(outWriter)

	// Create HTTP client for requests.
	client := createHTTPClient()

	// Step 3: Check discovery endpoints.
	_, _ = fmt.Fprintln(outWriter, "🔍 Step 3: Verifying Discovery Endpoints...")

	if err := checkDiscoveryEndpoints(ctx, client); err != nil {
		return fmt.Errorf("discovery check failed: %w", err)
	}

	_, _ = fmt.Fprintln(outWriter)

	// Step 4: Demonstrate OAuth 2.1 endpoints.
	_, _ = fmt.Fprintln(outWriter, "📋 Step 4: OAuth 2.1 Endpoint Summary")

	printEndpointSummary()

	_, _ = fmt.Fprintln(outWriter)

	// Step 5: Demonstrate authorization endpoint.
	_, _ = fmt.Fprintln(outWriter, "🔐 Step 5: Authorization Endpoint Demo...")

	_, codeChallenge := generatePKCE()
	state := generateState()

	if err := demonstrateAuthorization(ctx, client, codeChallenge, state); err != nil {
		_, _ = fmt.Fprintf(outWriter, "⚠️ Authorization demo info: %v\n", err)
	}

	_, _ = fmt.Fprintln(outWriter)

	// Step 6: Demonstrate token endpoint and get access token.
	_, _ = fmt.Fprintln(outWriter, "🔑 Step 6: Token Endpoint Demo...")

	accessToken, err := demonstrateTokenEndpoint(ctx, client)
	if err != nil {
		_, _ = fmt.Fprintf(outWriter, "⚠️ Token endpoint info: %v\n", err)
	}

	_, _ = fmt.Fprintln(outWriter)

	// Step 7: Demonstrate introspection endpoint with real token.
	_, _ = fmt.Fprintln(outWriter, "🔬 Step 7: Token Introspection (BEFORE revocation)...")

	if err := demonstrateIntrospection(ctx, client, accessToken); err != nil {
		_, _ = fmt.Fprintf(outWriter, "⚠️ Introspection info: %v\n", err)
	}

	_, _ = fmt.Fprintln(outWriter)

	// Step 8: Demonstrate revocation endpoint with real token.
	_, _ = fmt.Fprintln(outWriter, "🗑️ Step 8: Token Revocation...")

	if err := demonstrateRevocation(ctx, client, accessToken); err != nil {
		_, _ = fmt.Fprintf(outWriter, "⚠️ Revocation info: %v\n", err)
	}

	_, _ = fmt.Fprintln(outWriter)

	// Step 9: Demonstrate introspection after revocation.
	_, _ = fmt.Fprintln(outWriter, "🔬 Step 9: Token Introspection (AFTER revocation)...")

	if err := demonstrateIntrospectionAfterRevoke(ctx, client, accessToken); err != nil {
		_, _ = fmt.Fprintf(outWriter, "⚠️ Introspection info: %v\n", err)
	}

	_, _ = fmt.Fprintln(outWriter)

	return nil
}

func startAuthZServer(ctx context.Context) (*fiber.App, *cryptoutilIdentityRepository.RepositoryFactory, func(), error) {
	// Create in-memory configuration.
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:         "identity-authz-demo",
			BindAddress:  cryptoutilSharedMagic.IPv4Loopback,
			Port:         cryptoutilSharedMagic.DemoServerPort,
			TLSEnabled:   false,
			ReadTimeout:  cryptoutilSharedMagic.DefaultReadTimeout,
			WriteTimeout: cryptoutilSharedMagic.DefaultWriteTimeout,
			IdleTimeout:  cryptoutilSharedMagic.DefaultIdleServerTimeout,
			AdminEnabled: true,
			AdminPort:    cryptoutilSharedMagic.DemoAdminPort,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type:            "sqlite",
			DSN:             cryptoutilSharedMagic.SQLiteInMemoryDSN,
			MaxOpenConns:    cryptoutilSharedMagic.DefaultMaxOpenConns,
			MaxIdleConns:    cryptoutilSharedMagic.DefaultMaxIdleConns,
			ConnMaxLifetime: cryptoutilSharedMagic.DefaultConnMaxLifetime,
			AutoMigrate:     true,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:               cryptoutilSharedMagic.DemoIssuer,
			AccessTokenLifetime:  cryptoutilSharedMagic.DefaultAccessTokenLifetime,
			RefreshTokenLifetime: cryptoutilSharedMagic.DefaultRefreshTokenLifetime,
			IDTokenLifetime:      cryptoutilSharedMagic.DefaultIDTokenLifetime,
			SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			AccessTokenFormat:    cryptoutilSharedMagic.TokenFormatJWS,
		},
	}

	// Create repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create repository factory: %w", err)
	}

	// Run migrations.
	if err := repoFactory.AutoMigrate(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create production key generator and key rotation manager.
	keyGenerator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DevelopmentKeyRotationPolicy(),
		keyGenerator,
		nil,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create key rotation manager: %w", err)
	}

	// Rotate to generate initial signing key.
	if err := keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to rotate signing key: %w", err)
	}

	// Create JWS issuer for token signing with key rotation.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		cryptoutilSharedMagic.DemoIssuer,
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		config.Tokens.AccessTokenLifetime,
		config.Tokens.IDTokenLifetime,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create JWS issuer: %w", err)
	}

	// Create UUID issuer for refresh tokens.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, nil, uuidIssuer, config.Tokens)

	// Create authz service with token service.
	authzService := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	// Create Fiber app and register routes.
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	authzService.RegisterRoutes(app)

	cleanup := func() {
		_, _ = fmt.Fprintln(outWriter, "🛑 Shutting down server...")

		if shutdownErr := app.Shutdown(); shutdownErr != nil {
			_, _ = fmt.Fprintf(outWriter, "⚠️ Shutdown error: %v\n", shutdownErr)
		}

		_, _ = fmt.Fprintln(outWriter, "✅ Server stopped")
	}

	return app, repoFactory, cleanup, nil
}

func registerDemoClient(ctx context.Context, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) error {
	// Hash the client secret with PBKDF2 (FIPS-compliant).
	hashedSecret, err := cryptoutilSharedCryptoPassword.HashPassword(demoClientSecret)
	if err != nil {
		return fmt.Errorf("failed to hash client secret: %w", err)
	}

	// Create pointer booleans.
	trueVal := true

	// Create demo client.
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                cryptoutilSharedMagic.DemoClientID,
		ClientSecret:            hashedSecret,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    cryptoutilSharedMagic.DemoClientName,
		Description:             "Demo OAuth 2.1 client for testing",
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode, cryptoutilSharedMagic.GrantTypeRefreshToken, cryptoutilSharedMagic.GrantTypeClientCredentials},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ScopeOfflineAccess},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
		RequirePKCE:             &trueVal,
		PKCEChallengeMethod:     cryptoutilSharedMagic.PKCEMethodS256,
		AccessTokenLifetime:     cryptoutilSharedMagic.AccessTokenExpirySeconds,
		RefreshTokenLifetime:    cryptoutilSharedMagic.RefreshTokenExpirySeconds,
		IDTokenLifetime:         cryptoutilSharedMagic.IDTokenExpirySeconds,
		Enabled:                 &trueVal,
	}

	// Save to repository.
	if err := repoFactory.ClientRepository().Create(ctx, client); err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: cryptoutilSharedMagic.DemoRequestDelay,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Demo only
			},
		},
	}
}

func checkDiscoveryEndpoints(ctx context.Context, client *http.Client) error {
	// Check OAuth metadata.
	oauthURL := fmt.Sprintf("%s/.well-known/oauth-authorization-server", cryptoutilSharedMagic.DemoIssuer)

	oauthMeta, err := getJSON(ctx, client, oauthURL)
	if err != nil {
		return fmt.Errorf("oauth metadata: %w", err)
	}

	_, _ = fmt.Fprintf(outWriter, "   ✅ OAuth Metadata: issuer=%s\n", oauthMeta["issuer"])

	grantTypes, _ := oauthMeta["grant_types_supported"].([]any) //nolint:errcheck // Demo ok assertion
	_, _ = fmt.Fprintf(outWriter, "   ✅ Grant Types: %v\n", grantTypes)

	// Check OIDC discovery.
	oidcURL := fmt.Sprintf("%s/.well-known/openid-configuration", cryptoutilSharedMagic.DemoIssuer)

	oidcMeta, err := getJSON(ctx, client, oidcURL)
	if err != nil {
		return fmt.Errorf("oidc discovery: %w", err)
	}

	_, _ = fmt.Fprintf(outWriter, "   ✅ OIDC Discovery: issuer=%s\n", oidcMeta["issuer"])

	// Check JWKS.
	jwksURL := fmt.Sprintf("%s/oauth2/v1/jwks", cryptoutilSharedMagic.DemoIssuer)

	jwks, err := getJSON(ctx, client, jwksURL)
	if err != nil {
		return fmt.Errorf("jwks: %w", err)
	}

	keys, ok := jwks["keys"].([]any)
	if !ok || len(keys) == 0 {
		_, _ = fmt.Fprintln(outWriter, "   ✅ JWKS: Empty (token service not configured)")
	} else {
		_, _ = fmt.Fprintf(outWriter, "   ✅ JWKS: %d key(s) available\n", len(keys))
	}

	return nil
}
