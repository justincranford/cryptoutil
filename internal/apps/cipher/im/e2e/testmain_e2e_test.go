// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	templateE2E "cryptoutil/internal/apps/template/testing/e2e"
	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient  *http.Client
	composeManager    *templateE2E.ComposeManager
	composeFile       = "deployments/compose/cipher-im/compose.yml"

	// Three cipher-im instances with different backends.
	sqliteInstance    = "cipher-im-sqlite"
	postgres1Instance = "cipher-im-postgres-1"
	postgres2Instance = "cipher-im-postgres-2"

	// Service URLs (mapped from container ports to host ports).
	sqlitePublicURL     = "https://127.0.0.1:8080"
	sqliteAdminURL      = "https://127.0.0.1:9090"
	postgres1PublicURL  = "https://127.0.0.1:8081"
	postgres1AdminURL   = "https://127.0.0.1:9091"
	postgres2PublicURL  = "https://127.0.0.1:8082"
	postgres2AdminURL   = "https://127.0.0.1:9092"
	otelCollectorURL    = "http://127.0.0.1:4317"
	grafanaURL          = "http://127.0.0.1:3000"
)

// TestMain orchestrates docker compose lifecycle for E2E tests.
// This validates production-ready deployment with PostgreSQL, telemetry, and multiple instances.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize compose manager with reusable helper.
	composeManager = templateE2E.NewComposeManager(composeFile)
	sharedHTTPClient = cryptoutilTLS.NewClientForTest()

	// Step 1: Start docker compose stack.
	if err := composeManager.Start(ctx); err != nil {
		fmt.Printf("Failed to start docker compose: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for all services to be healthy.
	fmt.Println("Waiting for services to be healthy...")
	healthTimeout := 60 * time.Second

	if err := composeManager.WaitForHealth(sqliteAdminURL, healthTimeout); err != nil {
		fmt.Printf("SQLite instance health check failed: %v\n", err)
		_ = composeManager.Stop(ctx)
		os.Exit(1)
	}
	if err := composeManager.WaitForHealth(postgres1AdminURL, healthTimeout); err != nil {
		fmt.Printf("PostgreSQL-1 instance health check failed: %v\n", err)
		_ = composeManager.Stop(ctx)
		os.Exit(1)
	}
	if err := composeManager.WaitForHealth(postgres2AdminURL, healthTimeout); err != nil {
		fmt.Printf("PostgreSQL-2 instance health check failed: %v\n", err)
		_ = composeManager.Stop(ctx)
		os.Exit(1)
	}

	fmt.Println("All services healthy. Running tests...")

	// Step 3: Run tests.
	exitCode := m.Run()

	// Step 4: Cleanup docker compose stack.
	_ = composeManager.Stop(ctx)

	os.Exit(exitCode)
}
