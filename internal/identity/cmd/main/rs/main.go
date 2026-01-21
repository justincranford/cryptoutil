// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the resource server entry point.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityServer "cryptoutil/internal/identity/server"
)

func main() {
	// Parse command-line flags.
	configPath := flag.String("config", "configs/identity/rs.yml", "path to RS server configuration file")

	flag.Parse()

	// Load configuration from YAML file.
	cfg, err := cryptoutilIdentityConfig.LoadFromFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", *configPath, err)
		os.Exit(1)
	}

	// Validate configuration.
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create token service (stub for now - would be initialized from issuer module).
	var tokenSvc *cryptoutilIdentityIssuer.TokenService

	// Create RS server.
	ctx := context.Background()

	rsServer, err := cryptoutilIdentityServer.NewRSServer(ctx, cfg, logger, tokenSvc)
	if err != nil {
		log.Fatalf("failed to create RS server: %v", err)
	}

	// Start RS server in a goroutine.
	go func() {
		log.Printf("starting RS server on %s:%d", cfg.RS.BindAddress, cfg.RS.Port)

		if err := rsServer.Start(context.Background()); err != nil {
			log.Fatalf("RS server error: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	log.Println("shutting down RS server...")

	// Create a context with timeout for graceful shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cryptoutilIdentityMagic.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	if err := rsServer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "RS server shutdown error: %v\n", err)
		os.Exit(1)
	}

	log.Println("RS server stopped gracefully")
	// Use configPath to avoid unused variable error
	_ = configPath
}
