// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"testing"

	cryptoutilAppsTemplateTestingE2e "cryptoutil/internal/apps/template/testing/e2e"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient *http.Client
	composeManager   *cryptoutilAppsTemplateTestingE2e.ComposeManager

	// Five identity service instances (actual container names).
	authzContainer = cryptoutilSharedMagic.IdentityE2EAuthzContainer // "identity-authz-e2e"
	idpContainer   = cryptoutilSharedMagic.IdentityE2EIDPContainer   // "identity-idp-e2e"
	rsContainer    = cryptoutilSharedMagic.IdentityE2ERSContainer    // "identity-rs-e2e"
	rpContainer    = cryptoutilSharedMagic.IdentityE2ERPContainer    // "identity-rp-e2e"
	spaContainer   = cryptoutilSharedMagic.IdentityE2ESPAContainer   // "identity-spa-e2e"

	// Service URLs (mapped from container ports to host ports).
	authzPublicURL = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IdentityE2EAuthzPublicPort) // "https://127.0.0.1:8100"
	idpPublicURL   = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IdentityE2EIDPPublicPort)   // "https://127.0.0.1:8101"
	rsPublicURL    = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IdentityE2ERSPublicPort)    // "https://127.0.0.1:8110"
	rpPublicURL    = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IdentityE2ERPPublicPort)    // "https://127.0.0.1:8120"
	spaPublicURL   = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.IdentityE2ESPAPublicPort)   // "https://127.0.0.1:8130"

	healthChecks = map[string]string{
		authzContainer: authzPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint,
		idpContainer:   idpPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint,
		rsContainer:    rsPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint,
		rpContainer:    rpPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint,
		spaContainer:   spaPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint,
	}
)

// TestMain orchestrates docker compose lifecycle for E2E tests.
// This validates production-ready deployment with all identity services.
//
// ENVIRONMENTAL NOTE: These E2E tests require Docker Desktop to be running on Windows.
// Without Docker Desktop, the tests will fail with errors like:
// - "unable to get image... open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified"
// - "Failed to start docker compose: exit status 1"
// This is an environmental requirement, not a code issue.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize compose manager with reusable helper.
	composeManager = cryptoutilAppsTemplateTestingE2e.NewComposeManager(cryptoutilSharedMagic.IdentityE2EComposeFile)
	sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

	// Step 1: Start docker compose stack.
	if err := composeManager.Start(ctx); err != nil {
		fmt.Printf("Failed to start docker compose: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for all services to be healthy using public /health endpoint.
	fmt.Println("Waiting for all identity service instances to be healthy...")

	if err := composeManager.WaitForMultipleServices(healthChecks, cryptoutilSharedMagic.IdentityE2EHealthTimeout); err != nil {
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
