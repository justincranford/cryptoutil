// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cryptoutilIdentityBootstrap "cryptoutil/internal/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/identity/server"
)

func Execute() {
	executable := os.Args[0] // Example executable: ./cryptoutil
	if len(os.Args) < 2 {
		printUsage(executable)
		os.Exit(1)
	}

	command := os.Args[1]     // Example command: server
	parameters := os.Args[2:] // Example parameters: --config-file, --port, --host, etc.

	switch command {
	case "server":
		server(parameters)
	case "identity":
		identity(parameters)
	// case "kv":
	// 	kv(parameters)
	case "help":
		printUsage(executable)
	default:
		printUsage(executable)
		fmt.Printf("Unknown command: %s %s\n", executable, command)
		os.Exit(1)
	}
}

func identity(parameters []string) {
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
	configFlag      = "--config"
	configFlagShort = "-c"
)

func identityAuthz(parameters []string) {
	// Default config file
	configFile := "/app/run/authz-docker.yml"

	// Parse command-line flags for config override
	for i, param := range parameters {
		if (param == configFlag || param == configFlagShort) && i+1 < len(parameters) {
			configFile = parameters[i+1]

			break
		}
	}

	// Debug logging
	fmt.Fprintf(os.Stderr, "identityAuthz: Loading config from: %s\n", configFile)

	wd, _ := os.Getwd()
	fmt.Fprintf(os.Stderr, "identityAuthz: Working directory: %s\n", wd)

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

	// Initialize repository factory
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize repository factory: %v\n", err)
		os.Exit(1)
	}

	// Run database migrations if auto_migrate enabled
	if config.Database.AutoMigrate {
		fmt.Fprintf(os.Stderr, "Running database migrations...\n")

		if err := repoFactory.AutoMigrate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run database migrations: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Database migrations completed successfully\n")
	}

	// Bootstrap demo client for testing
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilIdentityMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully
	if err := authzServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}

func identityIdp(parameters []string) {
	// Default config file
	configFile := "configs/identity/idp.yml"

	// Parse command-line flags for config override
	for i, param := range parameters {
		if (param == configFlag || param == configFlagShort) && i+1 < len(parameters) {
			configFile = parameters[i+1]

			break
		}
	}

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

	// Initialize repository factory
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize repository factory: %v\n", err)
		os.Exit(1)
	}

	// TODO: Create JWS, JWE, UUID issuers properly
	// For now, use placeholders
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilIdentityMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully
	if err := idpServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}

func identityRs(parameters []string) {
	// Default config file
	configFile := "configs/identity/rs.yml"

	// Parse command-line flags for config override
	for i, param := range parameters {
		if (param == "--config" || param == "-c") && i+1 < len(parameters) {
			configFile = parameters[i+1]

			break
		}
	}

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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilIdentityMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully
	if err := rsServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}

func identitySpaRp(parameters []string) {
	fmt.Println("Starting SPA Relying Party...")
	// TODO: Implement SPA Relying Party
	fmt.Println("SPA Relying Party not yet implemented")
	os.Exit(1)
}

func printUsage(executable string) {
	fmt.Printf("Usage: %s <command> [options]\n", executable)
	fmt.Println("  server")
	fmt.Println("  identity")
}
