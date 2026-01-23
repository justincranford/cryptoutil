// Copyright (c) 2025 Justin Cranford

package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	templateE2E "cryptoutil/internal/apps/template/testing/e2e"
	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient *http.Client
	composeManager   *templateE2E.ComposeManager

	// Five identity service instances (actual container names).
	authzContainer = cryptoutilMagic.IdentityE2EAuthzContainer // "identity-authz-e2e"
	idpContainer   = cryptoutilMagic.IdentityE2EIDPContainer   // "identity-idp-e2e"
	rsContainer    = cryptoutilMagic.IdentityE2ERSContainer    // "identity-rs-e2e"
	rpContainer    = cryptoutilMagic.IdentityE2ERPContainer    // "identity-rp-e2e"
	spaContainer   = cryptoutilMagic.IdentityE2ESPAContainer   // "identity-spa-e2e"

	// Service URLs (mapped from container ports to host ports).
	authzPublicURL = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilMagic.IdentityE2EAuthzPublicPort) // "https://127.0.0.1:18000"
	idpPublicURL   = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilMagic.IdentityE2EIDPPublicPort)   // "https://127.0.0.1:18100"
	rsPublicURL    = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilMagic.IdentityE2ERSPublicPort)    // "https://127.0.0.1:18200"
	rpPublicURL    = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilMagic.IdentityE2ERPPublicPort)    // "https://127.0.0.1:18300"
	spaPublicURL   = fmt.Sprintf("https://127.0.0.1:%d", cryptoutilMagic.IdentityE2ESPAPublicPort)   // "https://127.0.0.1:18400"

	healthChecks = map[string]string{
		authzContainer: authzPublicURL + cryptoutilMagic.IdentityE2EHealthEndpoint,
		idpContainer:   idpPublicURL + cryptoutilMagic.IdentityE2EHealthEndpoint,
		rsContainer:    rsPublicURL + cryptoutilMagic.IdentityE2EHealthEndpoint,
		rpContainer:    rpPublicURL + cryptoutilMagic.IdentityE2EHealthEndpoint,
		spaContainer:   spaPublicURL + cryptoutilMagic.IdentityE2EHealthEndpoint,
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
	composeManager = templateE2E.NewComposeManager(cryptoutilMagic.IdentityE2EComposeFile)
	sharedHTTPClient = cryptoutilTLS.NewClientForTest()

	// Step 1: Start docker compose stack.
	if err := composeManager.Start(ctx); err != nil {
		fmt.Printf("Failed to start docker compose: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for all services to be healthy using public /health endpoint.
	fmt.Println("Waiting for all identity service instances to be healthy...")

	if err := composeManager.WaitForMultipleServices(healthChecks, cryptoutilMagic.IdentityE2EHealthTimeout); err != nil {
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
