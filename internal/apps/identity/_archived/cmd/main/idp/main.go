// Copyright (c) 2025 Justin Cranford
//
//

// Package main is the entry point for the Identity Provider server.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/apps/identity/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func main() {
	// Parse command-line flags.
	configFile := flag.String("config", "configs/identity/idp.yml", "Path to configuration file")

	flag.Parse()

	// Load configuration from YAML file.
	config, err := cryptoutilIdentityConfig.LoadFromFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", *configFile, err)
		os.Exit(1)
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

	// Run database migrations.
	if err := repoFactory.AutoMigrate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run database migrations: %v\n", err)
		os.Exit(1)
	}

	// Bootstrap demo client for testing.
	if err := cryptoutilIdentityBootstrap.BootstrapClients(ctx, config, repoFactory); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bootstrap clients: %v\n", err)
		os.Exit(1)
	}

	// Create production key generator.
	keyGenerator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()

	// Create key rotation manager with default policy.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		keyGenerator,
		func(keyID string) {
			fmt.Printf("Key rotated: %s\n", keyID)
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create key rotation manager: %v\n", err)
		os.Exit(1)
	}

	// Rotate initial signing key.
	if err := keyRotationMgr.RotateSigningKey(ctx, config.Tokens.SigningAlgorithm); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rotate initial signing key: %v\n", err)
		os.Exit(1)
	}

	// Rotate initial encryption key.
	if err := keyRotationMgr.RotateEncryptionKey(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rotate initial encryption key: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Failed to create JWS issuer: %v\n", err)
		os.Exit(1)
	}

	// Create JWE issuer.
	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create JWE issuer: %v\n", err)
		os.Exit(1)
	}

	// Create UUID issuer.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, config.Tokens)

	// Create IdP server.
	idpServer := cryptoutilIdentityServer.NewIDPServer(config, repoFactory, tokenSvc)

	// Create context with cancellation.
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine.
	go func() {
		fmt.Printf("Starting OIDC Identity Provider Server on %s:%d\n", config.IDP.BindAddress, config.IDP.Port)

		if err := idpServer.Start(serverCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down server...")

	// Create shutdown context with timeout.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilSharedMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully.
	if err := idpServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}
