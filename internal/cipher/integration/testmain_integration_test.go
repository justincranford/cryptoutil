// Copyright (c) 2025 Justin Cranford
//
//

//go:build !windows

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var sharedServer *server.CipherIMServer

// TestMain initializes cipher-im server with automatic PostgreSQL testcontainer provisioning.
// Service-template handles container lifecycle, database connection, and cleanup automatically.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Configure automatic PostgreSQL testcontainer provisioning.
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-integration-test")
	settings.DatabaseURL = ""         // Empty = use testcontainer.
	settings.DatabaseContainer = "required" // Require PostgreSQL testcontainer.

	cfg := &config.AppConfig{
		ServerSettings: *settings,
		JWTSecret:      uuid.Must(uuid.NewUUID()).String(),
	}

	// Create server with automatic infrastructure (PostgreSQL testcontainer, telemetry, etc.).
	var err error

	sharedServer, err = server.NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	// Start server in background (Start() blocks until shutdown).
	errChan := make(chan error, 1)

	go func() {
		if startErr := sharedServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for both servers to bind to ports.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	var publicPort int
	var adminPort int

	for i := 0; i < maxWaitAttempts; i++ {
		publicPort = sharedServer.PublicPort()

		adminPortValue, _ := sharedServer.AdminPort()
		adminPort = adminPortValue

		if publicPort > 0 && adminPort > 0 {
			break
		}

		select {
		case err := <-errChan:
			panic(fmt.Sprintf("server start error: %v", err))
		case <-time.After(waitInterval):
		}
	}

	if publicPort == 0 {
		panic("public server did not bind to port")
	}

	if adminPort == 0 {
		panic("admin server did not bind to port")
	}

	// Run all tests.
	exitCode := m.Run()

	// Automatic cleanup (database container, connections, services).
	_ = sharedServer.Shutdown(context.Background())

	os.Exit(exitCode)
}
