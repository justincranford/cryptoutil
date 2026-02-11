// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides Identity demo implementation.
package demo

import (
	"bytes"
	"context"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/url"
	"strings"
	"time"

	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/apps/identity/server"
)

// Identity demo step counts.
const (
	identityStepCount      = 5
	identityHealthTimeout  = 30 * time.Second
	identityHealthInterval = 500 * time.Millisecond

	// Configuration constants.
	identityServerReadTimeout  = 30 * time.Second
	identityServerWriteTimeout = 30 * time.Second
	identityServerIdleTimeout  = 120 * time.Second
	identityDemoPort           = 18080 // Fixed port for demo to allow health checks
	identityAdminPort          = 19090 // Fixed admin port for demo
	identityDBMaxOpenConns     = 5
	identityDBMaxIdleConns     = 2
	identityDBConnMaxLifetime  = 60 * time.Minute
	identityDBConnMaxIdleTime  = 10 * time.Minute
	identityAccessTokenTTL     = 3600 * time.Second
	identityIDTokenTTL         = 3600 * time.Second
	identityRefreshTokenTTL    = 86400 * time.Second
	identityServerStartDelay   = 100 * time.Millisecond
	identityShutdownTimeout    = 5 * time.Second
	identityHTTPClientTimeout  = 5 * time.Second
	identityHTTPLongTimeout    = 10 * time.Second
	identityJWTPartCount       = 3
	identityTokenTruncateLen   = 50

	// Step tracking for result calculation.
	identityStepsAfterConfig = 1
	identityStepsAfterServer = 2
	identityStepsAfterHealth = 3

	// Demo client credentials (shared with integration demo).
	identityDemoClientID     = "demo-client"
	identityDemoClientSecret = "demo-secret"
)

// identityDemoServer holds the running Identity server instance.
type identityDemoServer struct {
	config      *cryptoutilIdentityConfig.Config
	server      *cryptoutilIdentityServer.AuthZServer
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	cancelFunc  context.CancelFunc
	baseURL     string
}

// runIdentityDemo executes the Identity demo.
func runIdentityDemo(ctx context.Context, config *Config) int {
	progress := NewProgressDisplay(config)
	errors := NewErrorAggregator("identity")
	startTime := time.Now().UTC()

	progress.Info("Starting Identity Demo")
	progress.Info("=======================")
	progress.SetTotalSteps(identityStepCount)

	// Step 1: Parse configuration.
	progress.StartStep("Parsing configuration")

	identityConfig, err := parseIdentityConfig()
	if err != nil {
		progress.FailStep("Parsing configuration", err)
		errors.Add("config", "failed to parse configuration", err)

		result := errors.ToResult(0, identityStepCount-1)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Parsed configuration")

	// Step 2: Start server.
	progress.StartStep("Starting Identity AuthZ server")

	demoServer, err := startIdentityServer(ctx, identityConfig)
	if err != nil {
		progress.FailStep("Starting Identity AuthZ server", err)
		errors.Add("server", "failed to start Identity server", err)

		result := errors.ToResult(1, identityStepCount-2)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	defer func() {
		progress.Debug("Shutting down Identity server")
		stopIdentityServer(demoServer)
	}()

	progress.CompleteStep("Started Identity AuthZ server on " + demoServer.baseURL)

	// Step 3: Wait for health checks.
	progress.StartStep("Waiting for health checks")

	if err := waitForIdentityHealth(ctx, demoServer, config.HealthTimeout); err != nil {
		progress.FailStep("Health checks", err)
		errors.Add("health", "health checks failed", err)

		result := errors.ToResult(identityStepsAfterServer, identityStepCount-identityStepsAfterHealth)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Health checks passed")

	// Step 4: Verify OpenID configuration.
	progress.StartStep("Verifying OpenID configuration")

	if err := verifyOpenIDConfiguration(ctx, demoServer, progress); err != nil {
		progress.FailStep("OpenID configuration", err)
		errors.Add("openid", "OpenID configuration verification failed", err)

		if !config.ContinueOnError {
			result := errors.ToResult(identityStepsAfterHealth, identityStepCount-identityStepsAfterHealth-1)
			result.DurationMS = time.Since(startTime).Milliseconds()
			progress.PrintSummary(result)

			return result.ExitCode()
		}
	} else {
		progress.CompleteStep("OpenID configuration verified")
	}

	// Step 5: Demonstrate OAuth 2.1 client_credentials flow.
	progress.StartStep("Demonstrating OAuth 2.1 client_credentials flow")

	if err := demonstrateClientCredentialsFlow(ctx, demoServer, progress); err != nil {
		progress.FailStep("OAuth 2.1 flow", err)
		errors.Add("oauth", "OAuth 2.1 demonstration failed", err)
	} else {
		progress.CompleteStep("OAuth 2.1 client_credentials flow demonstrated")
	}

	// Calculate final result.
	passed := identityStepCount - errors.Count()

	result := errors.ToResult(passed, 0)
	result.DurationMS = time.Since(startTime).Milliseconds()

	progress.PrintSummary(result)

	return result.ExitCode()
}

// parseIdentityConfig creates configuration for Identity demo.
func parseIdentityConfig() (*cryptoutilIdentityConfig.Config, error) {
	return &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:             "identity-demo",
			BindAddress:      "127.0.0.1",
			Port:             identityDemoPort,
			TLSEnabled:       false,
			ReadTimeout:      identityServerReadTimeout,
			WriteTimeout:     identityServerWriteTimeout,
			IdleTimeout:      identityServerIdleTimeout,
			AdminEnabled:     true,
			AdminBindAddress: "127.0.0.1",
			AdminPort:        identityAdminPort,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type:            "sqlite",
			DSN:             ":memory:",
			MaxOpenConns:    identityDBMaxOpenConns,
			MaxIdleConns:    identityDBMaxIdleConns,
			ConnMaxLifetime: identityDBConnMaxLifetime,
			ConnMaxIdleTime: identityDBConnMaxIdleTime,
			AutoMigrate:     true,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  identityAccessTokenTTL,
			AccessTokenFormat:    "jws",
			IDTokenLifetime:      identityIDTokenTTL,
			IDTokenFormat:        "jws",
			RefreshTokenLifetime: identityRefreshTokenTTL,
			RefreshTokenFormat:   "uuid",
			Issuer:               "https://identity-demo.local",
			SigningAlgorithm:     "RS256",
		},
	}, nil
}

// startIdentityServer starts the Identity AuthZ server.
func startIdentityServer(ctx context.Context, config *cryptoutilIdentityConfig.Config) (*identityDemoServer, error) {
	// Initialize repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository factory: %w", err)
	}

	// Run database migrations.
	if err := repoFactory.AutoMigrate(ctx); err != nil {
		_ = repoFactory.Close()

		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Bootstrap demo client.
	if err := cryptoutilIdentityBootstrap.BootstrapClients(ctx, config, repoFactory); err != nil {
		_ = repoFactory.Close()

		return nil, fmt.Errorf("failed to bootstrap clients: %w", err)
	}

	// Create production key generator.
	keyGenerator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()

	// Create key rotation manager with default policy.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		keyGenerator,
		nil, // No callback needed for demo
	)
	if err != nil {
		_ = repoFactory.Close()

		return nil, fmt.Errorf("failed to create key rotation manager: %w", err)
	}

	// Rotate initial signing key.
	if err := keyRotationMgr.RotateSigningKey(ctx, config.Tokens.SigningAlgorithm); err != nil {
		_ = repoFactory.Close()

		return nil, fmt.Errorf("failed to rotate initial signing key: %w", err)
	}

	// Rotate initial encryption key.
	if err := keyRotationMgr.RotateEncryptionKey(ctx); err != nil {
		_ = repoFactory.Close()

		return nil, fmt.Errorf("failed to rotate initial encryption key: %w", err)
	}

	// Create JWS issuer.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		config.Tokens.Issuer,
		keyRotationMgr,
		config.Tokens.SigningAlgorithm,
		config.Tokens.AccessTokenLifetime,
		config.Tokens.IDTokenLifetime,
	)
	if err != nil {
		_ = repoFactory.Close()

		return nil, fmt.Errorf("failed to create JWS issuer: %w", err)
	}

	// Create JWE issuer.
	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	if err != nil {
		_ = repoFactory.Close()

		return nil, fmt.Errorf("failed to create JWE issuer: %w", err)
	}

	// Create UUID issuer.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, config.Tokens)

	// Create AuthZ server.
	authzServer := cryptoutilIdentityServer.NewAuthZServer(config, repoFactory, tokenSvc)

	// Create context with cancellation.
	serverCtx, cancel := context.WithCancel(ctx)

	// Start server in goroutine.
	go func() {
		_ = authzServer.Start(serverCtx)
	}()

	// Give server time to start and determine actual port.
	time.Sleep(identityServerStartDelay)

	// Build base URL with actual port.
	baseURL := fmt.Sprintf("http://%s:%d", config.AuthZ.BindAddress, config.AuthZ.Port)

	return &identityDemoServer{
		config:      config,
		server:      authzServer,
		repoFactory: repoFactory,
		cancelFunc:  cancel,
		baseURL:     baseURL,
	}, nil
}

// stopIdentityServer stops the Identity server.
func stopIdentityServer(demoServer *identityDemoServer) {
	if demoServer == nil {
		return
	}

	// Cancel context to trigger shutdown.
	if demoServer.cancelFunc != nil {
		demoServer.cancelFunc()
	}

	// Stop server gracefully.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), identityShutdownTimeout)
	defer shutdownCancel()

	if demoServer.server != nil {
		_ = demoServer.server.Stop(shutdownCtx)
	}

	// Close repository.
	if demoServer.repoFactory != nil {
		_ = demoServer.repoFactory.Close()
	}
}

// waitForIdentityHealth waits for Identity server health checks to pass.
func waitForIdentityHealth(ctx context.Context, demoServer *identityDemoServer, timeout time.Duration) error {
	deadline := time.Now().UTC().Add(timeout)
	healthURL := demoServer.baseURL + "/health"

	client := &http.Client{Timeout: identityHTTPClientTimeout}

	for time.Now().UTC().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("health check interrupted: %w", ctx.Err())
		case <-time.After(identityHealthInterval):
			// Continue polling.
		}
	}

	return fmt.Errorf("health checks did not pass within %v", timeout)
}

// verifyOpenIDConfiguration verifies the OpenID configuration endpoint.
func verifyOpenIDConfiguration(ctx context.Context, demoServer *identityDemoServer, progress *ProgressDisplay) error {
	configURL := demoServer.baseURL + "/.well-known/openid-configuration"

	client := &http.Client{Timeout: identityHTTPLongTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, configURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch OpenID configuration: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(body, &config); err != nil {
		return fmt.Errorf("failed to parse OpenID configuration: %w", err)
	}

	// Verify required fields.
	requiredFields := []string{"issuer", "token_endpoint", "jwks_uri"}
	for _, field := range requiredFields {
		if _, ok := config[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	progress.Debug(fmt.Sprintf("OpenID Configuration: issuer=%s", config["issuer"]))
	progress.Debug(fmt.Sprintf("  token_endpoint: %s", config["token_endpoint"]))
	progress.Debug(fmt.Sprintf("  jwks_uri: %s", config["jwks_uri"]))

	return nil
}

// demonstrateClientCredentialsFlow demonstrates the OAuth 2.1 client_credentials flow.
func demonstrateClientCredentialsFlow(ctx context.Context, demoServer *identityDemoServer, progress *ProgressDisplay) error {
	tokenURL := demoServer.baseURL + "/oauth2/v1/token"

	// Build token request using demo client credentials.
	clientID := identityDemoClientID
	clientSecret := identityDemoClientSecret

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "read write")

	client := &http.Client{Timeout: identityHTTPLongTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic auth for client authentication.
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	progress.Debug(fmt.Sprintf("Requesting token from: %s", tokenURL))
	progress.Debug(fmt.Sprintf("  client_id: %s", clientID))
	progress.Debug("  grant_type: client_credentials")
	progress.Debug("  scope: read write")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp map[string]any
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Verify token response.
	accessToken, ok := tokenResp["access_token"].(string)
	if !ok || accessToken == "" {
		return fmt.Errorf("missing access_token in response")
	}

	tokenType, _ := tokenResp["token_type"].(string)
	expiresIn, _ := tokenResp["expires_in"].(float64)

	progress.Debug("Token response received successfully:")
	progress.Debug(fmt.Sprintf("  token_type: %s", tokenType))
	progress.Debug(fmt.Sprintf("  expires_in: %.0f seconds", expiresIn))
	progress.Debug(fmt.Sprintf("  access_token: %s... (truncated)", accessToken[:min(identityTokenTruncateLen, len(accessToken))]))

	// Decode and display JWT claims (for demo purposes).
	parts := strings.Split(accessToken, ".")
	if len(parts) == identityJWTPartCount {
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var claims map[string]any
			if json.Unmarshal(payload, &claims) == nil {
				var prettyJSON bytes.Buffer
				if json.Indent(&prettyJSON, payload, "    ", "  ") == nil {
					progress.Debug("  JWT payload:")

					for _, line := range strings.Split(prettyJSON.String(), "\n") {
						progress.Debug("    " + line)
					}
				}
			}
		}
	}

	return nil
}
