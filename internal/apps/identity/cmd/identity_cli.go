// Copyright (c) 2025 Justin Cranford
//
//

// Package cmd provides the command-line interface for the identity service.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/apps/identity/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ExecuteIdentity is the entry point for the identity CLI, routing to authz, idp, rs, or spa-rp services.
func ExecuteIdentity(parameters []string) {
	if len(parameters) < 1 {
		fmt.Println("Usage: cryptoutil identity <service> [options]")
		fmt.Println("Services:")
		fmt.Println("  authz    - OAuth 2.1 Authorization Server")
		fmt.Println("  idp      - OIDC Identity Provider")
		fmt.Println("  rs       - Resource Server")
		fmt.Println("  spa-rp   - SPA Relying Party")
		os.Exit(1)
	}

	service := parameters[0]
	serviceParams := parameters[1:]

	switch service {
	case "authz":
		identityAuthz(serviceParams)
	case "idp":
		identityIdp(serviceParams)
	case "rs":
		identityRs(serviceParams)
	case "spa-rp":
		identitySpaRp(serviceParams)
	default:
		fmt.Printf("Unknown identity service: %s\n", service)
		fmt.Println("Available services: authz, idp, rs, spa-rp")
		os.Exit(1)
	}
}

const (
	configFlagShort = "-c"
	dsnFlag         = "-u"
	dsnFlagLong     = "--database-url"
)

// parseConfigFlag extracts config file path from parameters.
// Supports both "--config /path" and "--config=/path" formats.
func parseConfigFlag(parameters []string, defaultConfig string) string {
	for i, param := range parameters {
		// Support --config /path format.
		if param == cryptoutilSharedMagic.IdentityCLIFlagConfig || param == configFlagShort {
			if i+1 < len(parameters) {
				return parameters[i+1]
			}
		}

		// Support --config=/path format.
		if len(param) > len(cryptoutilSharedMagic.IdentityCLIFlagConfig) && param[:len(cryptoutilSharedMagic.IdentityCLIFlagConfig)+1] == cryptoutilSharedMagic.IdentityCLIFlagConfig+"=" {
			return param[len(cryptoutilSharedMagic.IdentityCLIFlagConfig)+1:]
		}

		// Support -c=/path format.
		if len(param) > len(configFlagShort) && param[:len(configFlagShort)+1] == configFlagShort+"=" {
			return param[len(configFlagShort)+1:]
		}
	}

	return defaultConfig
}

// parseDSNFlag extracts database URL from parameters.
// Supports both "-u value" and "-u=value" formats.
// If the value starts with "file://", it reads the DSN from that file path.
func parseDSNFlag(parameters []string) string {
	for i, param := range parameters {
		// Support -u /path or --database-url /path format.
		if param == dsnFlag || param == dsnFlagLong {
			if i+1 < len(parameters) {
				return resolveDSNValue(parameters[i+1])
			}
		}

		// Support -u=/path format.
		if len(param) > len(dsnFlag) && param[:len(dsnFlag)+1] == dsnFlag+"=" {
			return resolveDSNValue(param[len(dsnFlag)+1:])
		}

		// Support --database-url=/path format.
		if len(param) > len(dsnFlagLong) && param[:len(dsnFlagLong)+1] == dsnFlagLong+"=" {
			return resolveDSNValue(param[len(dsnFlagLong)+1:])
		}
	}

	return ""
}

// resolveDSNValue resolves a DSN value, reading from file if it's a file:// URL.
func resolveDSNValue(value string) string {
	if strings.HasPrefix(value, cryptoutilSharedMagic.FileURIScheme) {
		filePath := strings.TrimPrefix(value, cryptoutilSharedMagic.FileURIScheme)

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read DSN from file %s: %v\n", filePath, err)

			return ""
		}

		// Trim whitespace (including newlines) from the DSN.
		return strings.TrimSpace(string(data))
	}

	return value
}

func identityAuthz(parameters []string) {
	// Default config file.
	configFile := parseConfigFlag(parameters, "/app/run/authz-docker.yml")

	// Debug logging.
	fmt.Fprintf(os.Stderr, "identityAuthz: Loading config from: %s\n", configFile)

	wd, _ := os.Getwd()
	fmt.Fprintf(os.Stderr, "identityAuthz: Working directory: %s\n", wd)

	// Load configuration from YAML file.
	config, err := cryptoutilIdentityConfig.LoadFromFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", configFile, err)
		os.Exit(1)
	}

	// Override DSN from command line if provided (-u flag for Docker secrets).
	if dsn := parseDSNFlag(parameters); dsn != "" {
		fmt.Fprintf(os.Stderr, "identityAuthz: Using DSN from command line flag\n")

		config.Database.DSN = dsn
	}

	// Validate configuration.
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create context.
	ctx := context.Background()

	// Initialize repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize repository factory: %v\n", err)
		os.Exit(1)
	}

	// Run database migrations if auto_migrate enabled.
	if config.Database.AutoMigrate {
		fmt.Fprintf(os.Stderr, "Running database migrations...\n")

		if err := repoFactory.AutoMigrate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run database migrations: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Database migrations completed successfully\n")
	}

	// Bootstrap demo client for testing.
	if err := cryptoutilIdentityBootstrap.BootstrapClients(ctx, config, repoFactory); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bootstrap clients: %v\n", err)
		os.Exit(1)
	}

	// Create key generator and key rotation manager
	keyGenerator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	keyRotationPolicy := cryptoutilIdentityIssuer.DevelopmentKeyRotationPolicy()

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(keyRotationPolicy, keyGenerator, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create key rotation manager: %v\n", err)
		os.Exit(1)
	}

	// Generate initial signing key
	if err := keyRotationMgr.RotateSigningKey(ctx, config.Tokens.SigningAlgorithm); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate initial signing key: %v\n", err)
		os.Exit(1)
	}

	// Generate initial encryption key
	if err := keyRotationMgr.RotateEncryptionKey(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate initial encryption key: %v\n", err)
		os.Exit(1)
	}

	// Create JWS issuer with key rotation manager
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		config.Tokens.Issuer,
		keyRotationMgr,
		config.Tokens.SigningAlgorithm,
		config.Tokens.AccessTokenLifetime,
		config.Tokens.IDTokenLifetime,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create JWS issuer: %v\n", err)
		os.Exit(1)
	}

	// Create JWE issuer with key rotation manager
	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create JWE issuer: %v\n", err)
		os.Exit(1)
	}

	// Create UUID issuer
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create token service
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, config.Tokens)

	// Create AuthZ server
	authzServer := cryptoutilIdentityServer.NewAuthZServer(config, repoFactory, tokenSvc)

	// Create context with cancellation
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		fmt.Printf("Starting OAuth 2.1 Authorization Server on %s:%d\n", config.AuthZ.BindAddress, config.AuthZ.Port)

		if err := authzServer.Start(serverCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilSharedMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully
	if err := authzServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}

func identityIdp(parameters []string) {
	// Parse config file from parameters.
	configFile := parseConfigFlag(parameters, "configs/identity/idp.yml")

	// Load configuration from YAML file.
	config, err := cryptoutilIdentityConfig.LoadFromFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", configFile, err)
		os.Exit(1)
	}

	// Override DSN from command line if provided (-u flag for Docker secrets).
	if dsn := parseDSNFlag(parameters); dsn != "" {
		fmt.Fprintf(os.Stderr, "identityIdp: Using DSN from command line flag\n")

		config.Database.DSN = dsn
	}

	// Validate configuration.
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create context.
	ctx := context.Background()

	// Initialize repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize repository factory: %v\n", err)
		os.Exit(1)
	}

	// Bootstrap demo user for testing.
	if err := cryptoutilIdentityBootstrap.BootstrapUsers(ctx, repoFactory); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bootstrap users: %v\n", err)
		os.Exit(1)
	}

	// TODO: Create JWS, JWE, UUID issuers properly.
	// For now, use placeholders.
	jwsIssuer := &cryptoutilIdentityIssuer.JWSIssuer{}
	jweIssuer := &cryptoutilIdentityIssuer.JWEIssuer{}
	uuidIssuer := &cryptoutilIdentityIssuer.UUIDIssuer{}

	// Create token service
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, config.Tokens)

	// Create IdP server
	idpServer := cryptoutilIdentityServer.NewIDPServer(config, repoFactory, tokenSvc)

	// Create context with cancellation
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		fmt.Printf("Starting OIDC Identity Provider Server on %s:%d\n", config.IDP.BindAddress, config.IDP.Port)

		if err := idpServer.Start(serverCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilSharedMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully
	if err := idpServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}

func identityRs(parameters []string) {
	// Parse config file from parameters
	configFile := parseConfigFlag(parameters, "configs/identity/rs.yml")

	// Load configuration from YAML file
	config, err := cryptoutilIdentityConfig.LoadFromFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", configFile, err)
		os.Exit(1)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create context
	ctx := context.Background()

	// Create token service (stub for now - would be initialized from issuer module)
	var tokenSvc *cryptoutilIdentityIssuer.TokenService

	// Create RS server
	rsServer, err := cryptoutilIdentityServer.NewRSServer(ctx, config, nil, tokenSvc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create RS server: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		fmt.Printf("Starting Resource Server on %s:%d\n", config.RS.BindAddress, config.RS.Port)

		if err := rsServer.Start(serverCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilSharedMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully
	if err := rsServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}

func identitySpaRp(_ []string) {
	fmt.Println("Starting SPA Relying Party...")
	// TODO: Implement SPA Relying Party
	fmt.Println("SPA Relying Party not yet implemented")
	os.Exit(1)
}
