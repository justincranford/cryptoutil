// Copyright (c) 2025 Justin Cranford
//
//

// Package cmd provides command-line entry points for KMS server operations.
package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilKMSServer "cryptoutil/internal/apps/sm/kms/server"
	cryptoutilServerApplication "cryptoutil/internal/apps/sm/kms/server/application"
)

// Server handles the KMS server command and subcommands.
func Server(parameters []string) {
	// Reuse same Settings for start, ready, live, stop sub-commands, since they need to share private API coordinates.
	settings, err := cryptoutilAppsTemplateServiceConfig.Parse(parameters, true)
	if err != nil {
		log.Fatal("Error parsing config:", err)
	}

	switch settings.SubCommand {
	case "start":
		startServer(settings)
	case "stop":
		err := cryptoutilServerApplication.SendServerListenerShutdownRequest(settings)
		if err != nil {
			log.Fatalf("failed to stop server application: %v", err)
		}
	case "live":
		_, err := cryptoutilServerApplication.SendServerListenerLivenessCheck(settings)
		if err != nil {
			log.Fatalf("failed to check server liveness: %v", err)
		}
	case "ready":
		_, err := cryptoutilServerApplication.SendServerListenerReadinessCheck(settings)
		if err != nil {
			log.Fatalf("failed to check server readiness: %v", err)
		}
	case "init":
		err := cryptoutilServerApplication.ServerInit(settings)
		if err != nil {
			log.Fatalf("failed to init server application: %v", err)
		}
	default:
		log.Fatalf("unknown subcommand: %v", settings.SubCommand)
	}
}

// startServer starts the KMS server using the new ServerBuilder-based implementation.
func startServer(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create KMS server using ServerBuilder infrastructure.
	kmsServer, err := cryptoutilKMSServer.NewKMSServer(ctx, settings)
	if err != nil {
		log.Fatalf("failed to create KMS server: %v", err)
	}

	// Setup signal handling for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if startErr := kmsServer.Start(); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for shutdown signal or server error.
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		cancel()
		kmsServer.Shutdown()
	case startErr := <-errChan:
		log.Fatalf("server error: %v", startErr)
	}
}
