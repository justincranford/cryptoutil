// Copyright (c) 2025 Justin Cranford
//
//

//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"testing"

	cryptoutilAppsTemplateTestingE2eInfra "cryptoutil/internal/apps/template/service/testing/e2e_infra"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient *http.Client
	composeManager   *cryptoutilAppsTemplateTestingE2eInfra.ComposeManager

	// Three sm-im instances with different backends (actual container names).
	sqliteContainer    = cryptoutilSharedMagic.IME2ESQLiteContainer      // "sm-im-app-sqlite-1"
	postgres1Container = cryptoutilSharedMagic.IME2EPostgreSQL1Container // "sm-im-app-postgres-1"
	postgres2Container = cryptoutilSharedMagic.IME2EPostgreSQL2Container // "sm-im-app-postgres-2"

	// Service URLs (mapped from container ports to host ports).
	sqlitePublicURL    = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IME2ESQLitePublicPort)      // "https://127.0.0.1:8700"
	postgres1PublicURL = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IME2EPostgreSQL1PublicPort) // "https://127.0.0.1:8701"
	postgres2PublicURL = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IME2EPostgreSQL2PublicPort) // "https://127.0.0.1:8702"
	grafanaURL         = fmt.Sprintf("http://127.0.0.1:%d", cryptoutilSharedMagic.IME2EGrafanaPort)            // "http://127.0.0.1:3000"

	healthChecks = map[string]string{
		sqliteContainer:    sqlitePublicURL + cryptoutilSharedMagic.IME2EHealthEndpoint,
		postgres1Container: postgres1PublicURL + cryptoutilSharedMagic.IME2EHealthEndpoint,
		postgres2Container: postgres2PublicURL + cryptoutilSharedMagic.IME2EHealthEndpoint,
	}
)

// TestMain orchestrates docker compose lifecycle for E2E tests.
// This validates production-ready deployment with PostgreSQL, telemetry, and multiple instances.
//
// ENVIRONMENTAL NOTE: These E2E tests require Docker Desktop to be running on Windows.
// Without Docker Desktop, the tests will fail with errors like:
// - "unable to get image... open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified"
// - "Failed to start docker compose: exit status 1"
// This is an environmental requirement, not a code issue. The integration tests (in ../integration/)
// provide sufficient coverage using SQLite in-memory and do not require Docker.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize compose manager with reusable helper.
	composeManager = cryptoutilAppsTemplateTestingE2eInfra.NewComposeManager(cryptoutilSharedMagic.IME2EComposeFile, "dev", "postgres")
	sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

	// Step 1: Start docker compose stack.
	if err := composeManager.Start(ctx); err != nil {
		fmt.Printf("Failed to start docker compose: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for all services to be healthy using public /health endpoint.
	fmt.Println("Waiting for all sm-im instances to be healthy...")

	if err := composeManager.WaitForMultipleServices(healthChecks, cryptoutilSharedMagic.IME2EHealthTimeout); err != nil {
		fmt.Printf("Service health checks failed: %v\n", err)

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
