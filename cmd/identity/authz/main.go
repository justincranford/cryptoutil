package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/identity/server"
)

func main() {
	// Parse command-line flags.
	configFile := flag.String("config", "configs/identity/authz.yml", "Path to configuration file")
	flag.Parse()

	// TODO: Load configuration from YAML file.
	// For now, create minimal configuration.
	config := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "authz-server",
			BindAddress: "127.0.0.1",
			Port:        cryptoutilIdentityMagic.DefaultAuthZPort,
			TLSEnabled:  false,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: "jws",
			Issuer:            "https://authz.example.com",
		},
	}

	_ = configFile

	// Create context.
	ctx := context.Background()

	// Initialize repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize repository factory: %v\n", err)
		os.Exit(1)
	}

	// TODO: Create JWS, JWE, UUID issuers properly.
	// For now, use placeholders.
	jwsIssuer := &cryptoutilIdentityIssuer.JWSIssuer{}
	jweIssuer := &cryptoutilIdentityIssuer.JWEIssuer{}
	uuidIssuer := &cryptoutilIdentityIssuer.UUIDIssuer{}

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, config.Tokens)

	// Create AuthZ server.
	authzServer := cryptoutilIdentityServer.NewAuthZServer(config, repoFactory, tokenSvc)

	// Create context with cancellation.
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine.
	go func() {
		fmt.Printf("Starting OAuth 2.1 Authorization Server on %s:%d\n", config.AuthZ.BindAddress, config.AuthZ.Port)

		if err := authzServer.Start(serverCtx); err != nil {
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilIdentityMagic.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Stop server gracefully.
	if err := authzServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}
