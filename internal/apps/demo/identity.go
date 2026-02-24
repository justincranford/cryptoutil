// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides Identity demo implementation.
package demo

import (
	"context"
	"fmt"
	http "net/http"
	"time"

	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/apps/identity/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
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
	identityDemoClientID     = cryptoutilSharedMagic.DemoClientID
	identityDemoClientSecret = cryptoutilSharedMagic.DemoClientSecret
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

	// Build base URL and poll for server readiness instead of sleeping.
	baseURL := fmt.Sprintf("http://%s:%d", config.AuthZ.BindAddress, config.AuthZ.Port)
	healthURL := baseURL + "/health"
	client := &http.Client{Timeout: identityHTTPClientTimeout}

	if err := cryptoutilSharedUtilPoll.Until(ctx, identityHealthTimeout, identityHealthInterval, func(pollCtx context.Context) (bool, error) {
		return isHTTPHealthy(pollCtx, client, healthURL), nil
	}); err != nil {
		cancel()

		return nil, fmt.Errorf("identity server failed to become ready: %w", err)
	}

	return &identityDemoServer{
		config:      config,
		server:      authzServer,
		repoFactory: repoFactory,
		cancelFunc:  cancel,
		baseURL:     baseURL,
	}, nil
}

// stopIdentityServer stops the Identity server.
