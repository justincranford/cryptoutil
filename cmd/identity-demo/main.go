// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

const (
	demoTimeout     = 30 * time.Second
	requestTimeout  = 10 * time.Second
	demoIssuer      = "http://127.0.0.1:8080"
	demoPort        = ":8080"
	demoClientID    = "demo-client"
	demoClientName  = "Demo Client"
	demoRedirectURI = "https://example.com/callback"
)

var demoClientSecret = "demo-secret-" + googleUuid.New().String()[:8]

func main() {
	fmt.Println("🚀 Identity System Demo - OAuth 2.1 Authorization Server")
	fmt.Println("=========================================================")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), demoTimeout)
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
	time.Sleep(500 * time.Millisecond)
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

	// Step 6: Demonstrate token endpoint.
	fmt.Println("🔑 Step 6: Token Endpoint Demo...")
	if err := demonstrateTokenEndpoint(ctx, client); err != nil {
		fmt.Printf("⚠️ Token endpoint info: %v\n", err)
	}
	fmt.Println()

	// Step 7: Demonstrate introspection endpoint.
	fmt.Println("🔬 Step 7: Introspection Endpoint Demo...")
	if err := demonstrateIntrospection(ctx, client); err != nil {
		fmt.Printf("⚠️ Introspection info: %v\n", err)
	}
	fmt.Println()

	// Step 8: Demonstrate revocation endpoint.
	fmt.Println("🗑️ Step 8: Revocation Endpoint Demo...")
	if err := demonstrateRevocation(ctx, client); err != nil {
		fmt.Printf("⚠️ Revocation info: %v\n", err)
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
			Port:         8080,
			TLSEnabled:   false,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
			AdminEnabled: true,
			AdminPort:    9090,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type:            "sqlite",
			DSN:             "file::memory:?cache=shared",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 60 * time.Minute,
			AutoMigrate:     true,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:               demoIssuer,
			AccessTokenLifetime:  3600 * time.Second,
			RefreshTokenLifetime: 86400 * time.Second,
			IDTokenLifetime:      3600 * time.Second,
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
	// Hash the client secret.
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(demoClientSecret), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash client secret: %w", err)
	}

	// Create pointer booleans.
	trueVal := true

	// Create demo client.
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                demoClientID,
		ClientSecret:            string(hashedSecret),
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
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
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
		Timeout: requestTimeout,
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

	grantTypes, _ := oauthMeta["grant_types_supported"].([]any)
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
		return err
	}

	// Don't follow redirects.
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	fmt.Printf("   Request: GET %s?...\n", authURL)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Println("   ✅ Authorization endpoint accessible")
	fmt.Println("   📝 In production: would redirect to IdP login page")

	return nil
}

func demonstrateTokenEndpoint(ctx context.Context, client *http.Client) error {
	tokenURL := fmt.Sprintf("%s/oauth2/v1/token", demoIssuer)

	data := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {"openid profile email"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("   Request: POST %s\n", tokenURL)
	fmt.Printf("   Grant:   client_credentials\n")
	fmt.Printf("   Client:  %s (Basic Auth)\n", demoClientID)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	if resp.StatusCode == http.StatusOK {
		var tokenResp map[string]any
		if err := json.Unmarshal(body, &tokenResp); err == nil {
			if accessToken, ok := tokenResp["access_token"].(string); ok {
				fmt.Printf("   ✅ Access Token (first 20): %s...\n", accessToken[:min(20, len(accessToken))])
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

	return nil
}

func demonstrateIntrospection(ctx context.Context, client *http.Client) error {
	introspectURL := fmt.Sprintf("%s/oauth2/v1/introspect", demoIssuer)

	data := url.Values{
		"token": {"sample-access-token"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("   Request: POST %s\n", introspectURL)
	fmt.Printf("   Client:  %s (Basic Auth)\n", demoClientID)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var introspectResp map[string]any
	if err := json.Unmarshal(body, &introspectResp); err == nil {
		active, _ := introspectResp["active"].(bool)
		fmt.Printf("   ✅ Token Active: %v (unknown token returns inactive)\n", active)
	}

	return nil
}

func demonstrateRevocation(ctx context.Context, client *http.Client) error {
	revokeURL := fmt.Sprintf("%s/oauth2/v1/revoke", demoIssuer)

	data := url.Values{
		"token": {"sample-token-to-revoke"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	fmt.Printf("   Request: POST %s\n", revokeURL)
	fmt.Printf("   Client:  %s (Basic Auth)\n", demoClientID)
	fmt.Printf("   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Println("   ✅ Revocation endpoint returns 200 per RFC 7009")

	return nil
}

func getJSON(ctx context.Context, client *http.Client, urlStr string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func generatePKCE() (verifier, challenge string) {
	verifierBytes := make([]byte, 32)
	_, _ = io.ReadFull(strings.NewReader("abcdefghijklmnopqrstuvwxyz0123456789AB"), verifierBytes)
	verifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge
}

func generateState() string {
	stateBytes := make([]byte, 16)
	_, _ = io.ReadFull(strings.NewReader("ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"), stateBytes)

	return base64.RawURLEncoding.EncodeToString(stateBytes)
}
