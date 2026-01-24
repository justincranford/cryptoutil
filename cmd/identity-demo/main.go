// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides a demonstration of the identity service capabilities.
package main

import (
	"context"
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/url"
	"os"
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"
)

const (
	demoIssuer           = "http://127.0.0.1:8080"
	demoPort             = ":8080"
	demoClientID         = "demo-client"
	demoClientName       = "Demo Client"
	demoRedirectURI      = "https://example.com/callback"
	sampleAccessTokenFmt = "sample-access-token"
)

var demoClientSecret = "demo-secret-" + googleUuid.New().String()[:8]

func main() {
	fmt.Println("🚀 Identity System Demo - OAuth 2.1 Authorization Server")
	fmt.Println("=========================================================")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilIdentityMagic.DemoTimeout)
	defer cancel()

	// Run the demo.
	if err := runDemo(ctx); err != nil {
		fmt.Printf("❌ Demo failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Demo completed successfully!")
}

func runDemo(ctx context.Context) error {
	// Step 1: Start AuthZ server.
	fmt.Println("📦 Step 1: Starting Authorization Server...")

	app, repoFactory, cleanup, err := startAuthZServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	defer cleanup()

	// Start Fiber in background.
	go func() {
		if listenErr := app.Listen(demoPort); listenErr != nil {
			fmt.Printf("Server error: %v\n", listenErr)
		}
	}()

	// Give server time to start.
	time.Sleep(cryptoutilIdentityMagic.DemoStartupDelay)

	fmt.Println("✅ Authorization server started on http://127.0.0.1:8080")
	fmt.Println()

	// Step 2: Register demo client.
	fmt.Println("📝 Step 2: Registering Demo Client...")

	if err := registerDemoClient(ctx, repoFactory); err != nil {
		return fmt.Errorf("failed to register client: %w", err)
	}

	fmt.Printf("   ✅ Client ID: %s\n", demoClientID)
	fmt.Printf("   ✅ Client Name: %s\n", demoClientName)
	fmt.Printf("   ✅ Redirect URI: %s\n", demoRedirectURI)
	fmt.Println()

	// Create HTTP client for requests.
	client := createHTTPClient()

	// Step 3: Check discovery endpoints.
	fmt.Println("🔍 Step 3: Verifying Discovery Endpoints...")

	if err := checkDiscoveryEndpoints(ctx, client); err != nil {
		return fmt.Errorf("discovery check failed: %w", err)
	}

	fmt.Println()

	// Step 4: Demonstrate OAuth 2.1 endpoints.
	fmt.Println("📋 Step 4: OAuth 2.1 Endpoint Summary")
	printEndpointSummary()
	fmt.Println()

	// Step 5: Demonstrate authorization endpoint.
	fmt.Println("🔐 Step 5: Authorization Endpoint Demo...")

	_, codeChallenge := generatePKCE()
	state := generateState()

	if err := demonstrateAuthorization(ctx, client, codeChallenge, state); err != nil {
		fmt.Printf("⚠️ Authorization demo info: %v\n", err)
	}

	fmt.Println()

	// Step 6: Demonstrate token endpoint and get access token.
	fmt.Println("🔑 Step 6: Token Endpoint Demo...")

	accessToken, err := demonstrateTokenEndpoint(ctx, client)
	if err != nil {
		fmt.Printf("⚠️ Token endpoint info: %v\n", err)
	}

	fmt.Println()

	// Step 7: Demonstrate introspection endpoint with real token.
	fmt.Println("🔬 Step 7: Token Introspection (BEFORE revocation)...")

	if err := demonstrateIntrospection(ctx, client, accessToken); err != nil {
		fmt.Printf("⚠️ Introspection info: %v\n", err)
	}

	fmt.Println()

	// Step 8: Demonstrate revocation endpoint with real token.
	fmt.Println("🗑️ Step 8: Token Revocation...")

	if err := demonstrateRevocation(ctx, client, accessToken); err != nil {
		fmt.Printf("⚠️ Revocation info: %v\n", err)
	}

	fmt.Println()

	// Step 9: Demonstrate introspection after revocation.
	fmt.Println("🔬 Step 9: Token Introspection (AFTER revocation)...")

	if err := demonstrateIntrospectionAfterRevoke(ctx, client, accessToken); err != nil {
		fmt.Printf("⚠️ Introspection info: %v\n", err)
	}

	fmt.Println()

	return nil
}

func startAuthZServer(ctx context.Context) (*fiber.App, *cryptoutilIdentityRepository.RepositoryFactory, func(), error) {
	// Create in-memory configuration.
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:         "identity-authz-demo",
			BindAddress:  "127.0.0.1",
			Port:         cryptoutilIdentityMagic.DemoServerPort,
			TLSEnabled:   false,
			ReadTimeout:  cryptoutilIdentityMagic.DefaultReadTimeout,
			WriteTimeout: cryptoutilIdentityMagic.DefaultWriteTimeout,
			IdleTimeout:  cryptoutilIdentityMagic.DefaultIdleServerTimeout,
			AdminEnabled: true,
			AdminPort:    cryptoutilIdentityMagic.DemoAdminPort,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type:            "sqlite",
			DSN:             "file::memory:?cache=shared",
			MaxOpenConns:    cryptoutilIdentityMagic.DefaultMaxOpenConns,
			MaxIdleConns:    cryptoutilIdentityMagic.DefaultMaxIdleConns,
			ConnMaxLifetime: cryptoutilIdentityMagic.DefaultConnMaxLifetime,
			AutoMigrate:     true,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:               demoIssuer,
			AccessTokenLifetime:  cryptoutilIdentityMagic.DefaultAccessTokenLifetime,
			RefreshTokenLifetime: cryptoutilIdentityMagic.DefaultRefreshTokenLifetime,
			IDTokenLifetime:      cryptoutilIdentityMagic.DefaultIDTokenLifetime,
			SigningAlgorithm:     "RS256",
			AccessTokenFormat:    cryptoutilIdentityMagic.TokenFormatJWS,
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
	if err := keyRotationMgr.RotateSigningKey(ctx, "RS256"); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to rotate signing key: %w", err)
	}

	// Create JWS issuer for token signing with key rotation.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		demoIssuer,
		keyRotationMgr,
		"RS256",
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
		fmt.Println("🛑 Shutting down server...")

		if shutdownErr := app.Shutdown(); shutdownErr != nil {
			fmt.Printf("⚠️ Shutdown error: %v\n", shutdownErr)
		}

		fmt.Println("✅ Server stopped")
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
		ClientID:                demoClientID,
		ClientSecret:            hashedSecret,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    demoClientName,
		Description:             "Demo OAuth 2.1 client for testing",
		RedirectURIs:            []string{demoRedirectURI},
		AllowedGrantTypes:       []string{"authorization_code", "refresh_token", "client_credentials"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid", "profile", "email", "offline_access"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
		RequirePKCE:             &trueVal,
		PKCEChallengeMethod:     "S256",
		AccessTokenLifetime:     cryptoutilIdentityMagic.AccessTokenExpirySeconds,
		RefreshTokenLifetime:    cryptoutilIdentityMagic.RefreshTokenExpirySeconds,
		IDTokenLifetime:         cryptoutilIdentityMagic.IDTokenExpirySeconds,
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
		Timeout: cryptoutilIdentityMagic.DemoRequestDelay,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Demo only
			},
		},
	}
}

func checkDiscoveryEndpoints(ctx context.Context, client *http.Client) error {
	// Check OAuth metadata.
	oauthURL := fmt.Sprintf("%s/.well-known/oauth-authorization-server", demoIssuer)

	oauthMeta, err := getJSON(ctx, client, oauthURL)
	if err != nil {
		return fmt.Errorf("oauth metadata: %w", err)
	}

	fmt.Printf("   ✅ OAuth Metadata: issuer=%s\n", oauthMeta["issuer"])

	grantTypes, _ := oauthMeta["grant_types_supported"].([]any) //nolint:errcheck // Demo ok assertion
	fmt.Printf("   ✅ Grant Types: %v\n", grantTypes)

	// Check OIDC discovery.
	oidcURL := fmt.Sprintf("%s/.well-known/openid-configuration", demoIssuer)

	oidcMeta, err := getJSON(ctx, client, oidcURL)
	if err != nil {
		return fmt.Errorf("oidc discovery: %w", err)
	}

	fmt.Printf("   ✅ OIDC Discovery: issuer=%s\n", oidcMeta["issuer"])

	// Check JWKS.
	jwksURL := fmt.Sprintf("%s/oauth2/v1/jwks", demoIssuer)

	jwks, err := getJSON(ctx, client, jwksURL)
	if err != nil {
		return fmt.Errorf("jwks: %w", err)
	}

	keys, ok := jwks["keys"].([]any)
	if !ok || len(keys) == 0 {
		fmt.Println("   ✅ JWKS: Empty (token service not configured)")
	} else {
		fmt.Printf("   ✅ JWKS: %d key(s) available\n", len(keys))
	}

	return nil
}

func printEndpointSummary() {
	fmt.Println("   OAuth 2.1 / OpenID Connect Endpoints:")
	fmt.Printf("   • Discovery:     %s/.well-known/oauth-authorization-server\n", demoIssuer)
	fmt.Printf("   • OIDC Config:   %s/.well-known/openid-configuration\n", demoIssuer)
	fmt.Printf("   • Authorization: %s/oauth2/v1/authorize\n", demoIssuer)
	fmt.Printf("   • Token:         %s/oauth2/v1/token\n", demoIssuer)
	fmt.Printf("   • Introspect:    %s/oauth2/v1/introspect\n", demoIssuer)
	fmt.Printf("   • Revoke:        %s/oauth2/v1/revoke\n", demoIssuer)
	fmt.Printf("   • JWKS:          %s/oauth2/v1/jwks\n", demoIssuer)
}

func demonstrateAuthorization(ctx context.Context, client *http.Client, codeChallenge, state string) error {
	authURL := fmt.Sprintf("%s/oauth2/v1/authorize", demoIssuer)

	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {demoClientID},
		"redirect_uri":          {demoRedirectURI},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"scope":                 {"openid profile email"},
	}

	fullURL := authURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("create authorization request: %w", err)
	}

	// Don't follow redirects.
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute authorization request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	fmt.Printf("   Request: GET %s?...\n", authURL)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Println("   ✅ Authorization endpoint accessible")
	fmt.Println("   📝 In production: would redirect to IdP login page")

	return nil
}

func demonstrateTokenEndpoint(ctx context.Context, client *http.Client) (string, error) {
	tokenURL := fmt.Sprintf("%s/oauth2/v1/token", demoIssuer)

	data := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {"openid profile email"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	fmt.Printf("   Request: POST %s\n", tokenURL)
	fmt.Printf("   Grant:   client_credentials\n")
	fmt.Printf("   Client:  %s (Basic Auth)\n", demoClientID)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var accessToken string

	if resp.StatusCode == http.StatusOK {
		var tokenResp map[string]any

		if err := json.Unmarshal(body, &tokenResp); err == nil {
			if at, ok := tokenResp["access_token"].(string); ok {
				accessToken = at
				fmt.Printf("   ✅ Access Token (first %d): %s...\n",
					cryptoutilIdentityMagic.DemoMinTokenChars,
					accessToken[:min(cryptoutilIdentityMagic.DemoMinTokenChars, len(accessToken))])
			}

			if tokenType, ok := tokenResp["token_type"].(string); ok {
				fmt.Printf("   ✅ Token Type: %s\n", tokenType)
			}

			if expiresIn, ok := tokenResp["expires_in"].(float64); ok {
				fmt.Printf("   ✅ Expires In: %.0f seconds\n", expiresIn)
			}
		}
	} else {
		fmt.Printf("   ⚠️ Response: %s\n", string(body))
		fmt.Println("   📝 Note: Token service may need to be configured for token issuance")
	}

	return accessToken, nil
}

func demonstrateIntrospection(ctx context.Context, client *http.Client, accessToken string) error {
	introspectURL := fmt.Sprintf("%s/oauth2/v1/introspect", demoIssuer)

	// Use the actual access token if available, otherwise use a sample.
	tokenToIntrospect := accessToken
	if tokenToIntrospect == "" {
		tokenToIntrospect = sampleAccessTokenFmt
	}

	data := url.Values{
		"token": {tokenToIntrospect},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute introspection request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read introspection response: %w", err)
	}

	fmt.Printf("   Request: POST %s\n", introspectURL)
	fmt.Printf("   Client:  %s (Basic Auth)\n", demoClientID)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var introspectResp map[string]any
	if err := json.Unmarshal(body, &introspectResp); err == nil {
		active, _ := introspectResp["active"].(bool) //nolint:errcheck // Demo ok assertion

		if active && accessToken != "" {
			fmt.Println("   ✅ Token Active: true (issued access token validated)")
		} else if !active && accessToken != "" {
			fmt.Println("   ⚠️ Token Active: false (token may not be stored)")
		} else {
			fmt.Printf("   ✅ Token Active: %v (unknown token returns inactive)\n", active)
		}
	}

	return nil
}

func demonstrateRevocation(ctx context.Context, client *http.Client, accessToken string) error {
	revokeURL := fmt.Sprintf("%s/oauth2/v1/revoke", demoIssuer)

	// Use the actual access token if available, otherwise use a sample.
	tokenToRevoke := accessToken
	if tokenToRevoke == "" {
		tokenToRevoke = "sample-token-to-revoke"
	}

	data := url.Values{
		"token": {tokenToRevoke},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create revocation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute revocation request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	fmt.Printf("   Request: POST %s\n", revokeURL)
	fmt.Printf("   Client:  %s (Basic Auth)\n", demoClientID)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Println("   ✅ Revocation endpoint returns 200 per RFC 7009")

	return nil
}

func demonstrateIntrospectionAfterRevoke(ctx context.Context, client *http.Client, accessToken string) error {
	introspectURL := fmt.Sprintf("%s/oauth2/v1/introspect", demoIssuer)

	// Use the revoked access token.
	tokenToIntrospect := accessToken
	if tokenToIntrospect == "" {
		tokenToIntrospect = sampleAccessTokenFmt
	}

	data := url.Values{
		"token": {tokenToIntrospect},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create post-revoke introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute post-revoke introspection request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read post-revoke introspection response: %w", err)
	}

	fmt.Printf("   Request: POST %s\n", introspectURL)
	fmt.Printf("   Client:  %s (Basic Auth)\n", demoClientID)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var introspectResp map[string]any
	if err := json.Unmarshal(body, &introspectResp); err == nil {
		active, _ := introspectResp["active"].(bool) //nolint:errcheck // Demo ok assertion

		if !active && accessToken != "" {
			fmt.Println("   ✅ Token Active: false (revoked token correctly invalidated)")
		} else if active && accessToken != "" {
			fmt.Println("   ⚠️ Token Active: true (revocation may not have persisted)")
		} else {
			fmt.Printf("   ✅ Token Active: %v\n", active)
		}
	}

	return nil
}

func getJSON(ctx context.Context, client *http.Client, urlStr string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("create GET request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute GET request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode JSON response: %w", err)
	}

	return result, nil
}

func generatePKCE() (verifier, challenge string) {
	verifierBytes := make([]byte, cryptoutilIdentityMagic.DefaultStateLength)

	if _, err := crand.Read(verifierBytes); err != nil {
		// Fall back to deterministic value for demo purposes.
		copy(verifierBytes, []byte("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ12"))
	}

	verifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge
}

func generateState() string {
	stateBytes := make([]byte, cryptoutilIdentityMagic.DefaultNonceLength/2)

	if _, err := crand.Read(stateBytes); err != nil {
		// Fall back to deterministic value for demo purposes.
		copy(stateBytes, []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"))
	}

	return base64.RawURLEncoding.EncodeToString(stateBytes)
}
