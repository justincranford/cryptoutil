// Copyright (c) 2025 Justin Cranford

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

	// Three skeleton-template instances with different backends (actual compose service names).
	sqliteContainer    = cryptoutilSharedMagic.SkeletonTemplateE2ESQLiteContainer      // "skeleton-template-app-sqlite-1"
	postgres1Container = cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL1Container // "skeleton-template-app-postgres-1"
	postgres2Container = cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL2Container // "skeleton-template-app-postgres-2"

	// Service URLs (mapped from container ports to host ports).
	sqlitePublicURL    = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.SkeletonTemplateE2ESQLitePublicPort)      // "https://127.0.0.1:18900"
	postgres1PublicURL = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL1PublicPort) // "https://127.0.0.1:18901"
	postgres2PublicURL = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL2PublicPort) // "https://127.0.0.1:18902"

	healthChecks = map[string]string{
		sqliteContainer:    sqlitePublicURL + cryptoutilSharedMagic.SkeletonTemplateE2EHealthEndpoint,
		postgres1Container: postgres1PublicURL + cryptoutilSharedMagic.SkeletonTemplateE2EHealthEndpoint,
		postgres2Container: postgres2PublicURL + cryptoutilSharedMagic.SkeletonTemplateE2EHealthEndpoint,
	}
)

// TestMain orchestrates docker compose lifecycle for skeleton-template E2E tests.
// This validates production-ready deployment with PostgreSQL, telemetry, and multiple instances.
//
// ENVIRONMENTAL NOTE: These E2E tests require Docker Desktop to be running.
// Without Docker Desktop, the tests will fail with compose errors.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize compose manager with reusable helper.
	composeManager = cryptoutilAppsTemplateTestingE2eInfra.NewComposeManager(cryptoutilSharedMagic.SkeletonTemplateE2EComposeFile, cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault, cryptoutilSharedMagic.DockerServicePostgres)
	sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

	// Step 1: Start docker compose stack.
	if err := composeManager.Start(ctx); err != nil {
		fmt.Printf("Failed to start docker compose: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for all services to be healthy using public /health endpoint.
	fmt.Println("Waiting for all skeleton-template instances to be healthy...")

	if err := composeManager.WaitForMultipleServices(healthChecks, cryptoutilSharedMagic.SkeletonTemplateE2EHealthTimeout); err != nil {
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
