// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides integration demo implementation.
// This file demonstrates the full KMS and Identity server integration,
// including OAuth 2.1 token flow and authenticated operations.
package demo

import (
	"context"
	"fmt"
	"time"

	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/apps/identity/server"
	cryptoutilServerApplication "cryptoutil/internal/apps/sm/kms/server/application"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Integration demo step counts.
const (
	integrationStepCount = 7

	// Step tracking constants to avoid magic numbers.
	integrationStepsAfterIdentity = 1
	integrationStepsAfterKMS      = 2
	integrationStepsAfterHealth   = 3
	integrationStepsAfterToken    = 4
	integrationStepsAfterValidate = 5
	integrationStepsAfterKMSOp    = 6
	integrationStepsAfterAudit    = 7

	// Remaining steps calculations.
	integrationRemainingAfterKMS      = 5 // 7 - 2
	integrationRemainingAfterHealth   = 4 // 7 - 3
	integrationRemainingAfterToken    = 3 // 7 - 4
	integrationRemainingAfterValidate = 2 // 7 - 5
	integrationRemainingAfterKMSOp    = 1 // 7 - 6
	integrationRemainingAfterAudit    = 0 // 7 - 7

	// Integration timeouts.
	integrationHealthTimeout  = 30 * time.Second
	integrationHTTPTimeout    = 10 * time.Second
	integrationShutdownTime   = 5 * time.Second
	integrationServerStartup  = 100 * time.Millisecond
	integrationHealthInterval = 500 * time.Millisecond

	// Identity server configuration.
	integrationIdentityPort      = 18080
	integrationIdentityAdminPort = 19090

	// Database configuration.
	integrationDBMaxOpenConns   = 5
	integrationDBMaxIdleConns   = 2
	integrationDBConnMaxLife    = 60 * time.Minute
	integrationDBConnMaxIdle    = 10 * time.Minute
	integrationAccessTokenTTL   = 3600 * time.Second
	integrationIDTokenTTL       = 3600 * time.Second
	integrationRefreshTokenTTL  = 86400 * time.Second
	integrationServerReadTime   = 30 * time.Second
	integrationServerWriteTime  = 30 * time.Second
	integrationServerIdleTime   = 120 * time.Second
	integrationTokenTruncateLen = 50

	// Demo client credentials.
	integrationDemoClientID     = cryptoutilSharedMagic.DemoClientID
	integrationDemoClientSecret = cryptoutilSharedMagic.DemoClientSecret
)

// integrationServers holds running server instances for integration demo.
type integrationServers struct {
	identityConfig  *cryptoutilIdentityConfig.Config
	identityServer  *cryptoutilIdentityServer.AuthZServer
	identityRepo    *cryptoutilIdentityRepository.RepositoryFactory
	identityCancel  context.CancelFunc
	identityBaseURL string
	kmsServer       *cryptoutilServerApplication.ServerApplicationListener
	kmsSettings     *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	kmsBaseURL      string
}

// runIntegrationDemo executes the full integration demo (KMS + Identity).
func runIntegrationDemo(ctx context.Context, config *Config) int {
	progress := NewProgressDisplay(config)
	errors := NewErrorAggregator("integration")
	startTime := time.Now().UTC()

	progress.Info("Starting Integration Demo")
	progress.Info("=========================")
	progress.Info("This demo shows KMS and Identity server integration")
	progress.SetTotalSteps(integrationStepCount)

	var servers integrationServers

	defer func() {
		progress.Debug("Shutting down servers")
		stopIntegrationServers(&servers)
	}()

	// Step 1: Start Identity server.
	progress.StartStep("Starting Identity server")

	if err := startIntegrationIdentityServer(ctx, &servers); err != nil {
		progress.FailStep("Starting Identity server", err)
		errors.Add("identity_server", "failed to start Identity server", err)

		result := errors.ToResult(0, integrationStepCount-1)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Started Identity AuthZ server on " + servers.identityBaseURL)

	// Step 2: Start KMS server.
	progress.StartStep("Starting KMS server")

	if err := startIntegrationKMSServer(ctx, &servers); err != nil {
		progress.FailStep("Starting KMS server", err)
		errors.Add("kms_server", "failed to start KMS server", err)

		result := errors.ToResult(1, integrationStepCount-2)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Started KMS server on " + servers.kmsBaseURL)

	// Step 3: Wait for all services to be healthy.
	progress.StartStep("Waiting for all services")

	if err := waitForIntegrationHealth(ctx, &servers, config.HealthTimeout); err != nil {
		progress.FailStep("Service health checks", err)
		errors.Add("health", "health checks failed", err)

		result := errors.ToResult(integrationStepsAfterKMS, integrationRemainingAfterHealth)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("All service health checks passed")

	// Step 4: Get access token from Identity.
	progress.StartStep("Obtaining access token")

	accessToken, err := obtainIntegrationAccessToken(ctx, &servers, progress)
	if err != nil {
		progress.FailStep("Obtaining access token", err)
		errors.Add("token", "failed to obtain access token", err)

		result := errors.ToResult(integrationStepsAfterHealth, integrationRemainingAfterToken)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Obtained access token successfully")

	// Step 5: Validate token structure.
	progress.StartStep("Validating token structure")

	if err := validateIntegrationToken(accessToken, progress); err != nil {
		progress.FailStep("Token validation", err)
		errors.Add("validation", "token validation failed", err)

		result := errors.ToResult(integrationStepsAfterToken, integrationRemainingAfterValidate)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Token structure validated successfully")

	// Step 6: Perform authenticated KMS operation.
	progress.StartStep("Performing authenticated KMS operation")

	if err := performAuthenticatedKMSOperation(ctx, &servers, accessToken, progress); err != nil {
		progress.FailStep("Authenticated KMS operation", err)
		errors.Add("kms_operation", "authenticated KMS operation failed", err)

		result := errors.ToResult(integrationStepsAfterValidate, integrationRemainingAfterKMSOp)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Authenticated KMS operation completed")

	// Step 7: Verify integration audit trail.
	progress.StartStep("Verifying integration audit trail")

	if err := verifyIntegrationAuditTrail(progress); err != nil {
		progress.FailStep("Audit trail verification", err)
		errors.Add("audit", "audit trail verification failed", err)

		result := errors.ToResult(integrationStepsAfterKMSOp, integrationRemainingAfterAudit)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	progress.CompleteStep("Integration audit trail verified")

	// Calculate final result.
	passed := integrationStepCount - errors.Count()

	result := errors.ToResult(passed, 0)
	result.DurationMS = time.Since(startTime).Milliseconds()

	progress.PrintSummary(result)

	return result.ExitCode()
}

// startIntegrationIdentityServer starts the Identity server for integration demo.
func startIntegrationIdentityServer(ctx context.Context, servers *integrationServers) error {
	// Create Identity configuration.
	identityConfig := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:             "integration-identity",
			BindAddress:      "127.0.0.1",
			Port:             integrationIdentityPort,
			TLSEnabled:       false,
			ReadTimeout:      integrationServerReadTime,
			WriteTimeout:     integrationServerWriteTime,
			IdleTimeout:      integrationServerIdleTime,
			AdminEnabled:     true,
			AdminBindAddress: "127.0.0.1",
			AdminPort:        integrationIdentityAdminPort,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type:            "sqlite",
			DSN:             ":memory:",
			MaxOpenConns:    integrationDBMaxOpenConns,
			MaxIdleConns:    integrationDBMaxIdleConns,
			ConnMaxLifetime: integrationDBConnMaxLife,
			ConnMaxIdleTime: integrationDBConnMaxIdle,
			AutoMigrate:     true,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  integrationAccessTokenTTL,
			AccessTokenFormat:    "jws",
			IDTokenLifetime:      integrationIDTokenTTL,
			IDTokenFormat:        "jws",
			RefreshTokenLifetime: integrationRefreshTokenTTL,
			RefreshTokenFormat:   "uuid",
			Issuer:               "https://integration-demo.local",
			SigningAlgorithm:     "RS256",
		},
	}

	// Initialize repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, identityConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize repository factory: %w", err)
	}

	// Run database migrations.
	if err := repoFactory.AutoMigrate(ctx); err != nil {
		_ = repoFactory.Close()

		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Bootstrap demo client.
	if err := cryptoutilIdentityBootstrap.BootstrapClients(ctx, identityConfig, repoFactory); err != nil {
		_ = repoFactory.Close()

		return fmt.Errorf("failed to bootstrap clients: %w", err)
	}

	// Create production key generator.
	keyGenerator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()

	// Create key rotation manager with default policy.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		keyGenerator,
		nil,
	)
	if err != nil {
		_ = repoFactory.Close()

		return fmt.Errorf("failed to create key rotation manager: %w", err)
	}

	// Rotate initial signing key.
	if err := keyRotationMgr.RotateSigningKey(ctx, identityConfig.Tokens.SigningAlgorithm); err != nil {
		_ = repoFactory.Close()

		return fmt.Errorf("failed to rotate initial signing key: %w", err)
	}

	// Rotate initial encryption key.
	if err := keyRotationMgr.RotateEncryptionKey(ctx); err != nil {
		_ = repoFactory.Close()

		return fmt.Errorf("failed to rotate initial encryption key: %w", err)
	}

	// Create JWS issuer.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		identityConfig.Tokens.Issuer,
		keyRotationMgr,
		identityConfig.Tokens.SigningAlgorithm,
		identityConfig.Tokens.AccessTokenLifetime,
		identityConfig.Tokens.IDTokenLifetime,
	)
	if err != nil {
		_ = repoFactory.Close()

		return fmt.Errorf("failed to create JWS issuer: %w", err)
	}

	// Create JWE issuer.
	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	if err != nil {
		_ = repoFactory.Close()

		return fmt.Errorf("failed to create JWE issuer: %w", err)
	}

	// Create UUID issuer.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, identityConfig.Tokens)

	// Create AuthZ server.
	authzServer := cryptoutilIdentityServer.NewAuthZServer(identityConfig, repoFactory, tokenSvc)

	// Create context with cancellation.
	serverCtx, cancel := context.WithCancel(ctx)

	// Start server in goroutine.
	go func() {
		_ = authzServer.Start(serverCtx)
	}()

	// Give server time to start.
	time.Sleep(integrationServerStartup)

	// Build base URL.
	baseURL := fmt.Sprintf("http://%s:%d", identityConfig.AuthZ.BindAddress, identityConfig.AuthZ.Port)

	// Store server references.
	servers.identityConfig = identityConfig
	servers.identityServer = authzServer
	servers.identityRepo = repoFactory
	servers.identityCancel = cancel
	servers.identityBaseURL = baseURL

	return nil
}

// startIntegrationKMSServer starts the KMS server for integration demo.
func startIntegrationKMSServer(_ context.Context, servers *integrationServers) error {
	// Parse KMS configuration with dev and demo mode.
	args := []string{
		"start",
		"--dev",
		"--demo",
		"--log-level", "INFO",
		"--bind-public-port", "0",
		"--bind-private-port", "0",
	}

	settings, err := cryptoutilAppsTemplateServiceConfig.Parse(args, true)
	if err != nil {
		return fmt.Errorf("failed to parse KMS config: %w", err)
	}

	// Start KMS server.
	server, err := cryptoutilServerApplication.StartServerListenerApplication(settings)
	if err != nil {
		return fmt.Errorf("failed to start KMS server: %w", err)
	}

	// Start server in background.
	go server.StartFunction()

	// Give server time to start.
	time.Sleep(cryptoutilSharedMagic.DefaultServerStartupDelay)

	// Update settings with actual ports.
	settings.BindPublicPort = server.ActualPublicPort
	settings.BindPrivatePort = server.ActualPrivatePort

	// Build base URL with actual port.
	baseURL := fmt.Sprintf("https://%s:%d", settings.BindPublicAddress, server.ActualPublicPort)

	// Store server references.
	servers.kmsServer = server
	servers.kmsSettings = settings
	servers.kmsBaseURL = baseURL

	return nil
}

// stopIntegrationServers stops all running servers.
func stopIntegrationServers(servers *integrationServers) {
	if servers == nil {
		return
	}

	// Stop KMS server.
	if servers.kmsServer != nil {
		servers.kmsServer.ShutdownFunction()
	}

	// Stop Identity server.
	if servers.identityCancel != nil {
		servers.identityCancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), integrationShutdownTime)
	defer shutdownCancel()

	if servers.identityServer != nil {
		_ = servers.identityServer.Stop(shutdownCtx)
	}

	if servers.identityRepo != nil {
		_ = servers.identityRepo.Close()
	}
}

// waitForIntegrationHealth waits for both servers to be healthy.
