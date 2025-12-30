// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// Shared test resources (initialized once per package).
var (
	sharedTelemetryService *cryptoutilTelemetry.TelemetryService
	sharedJWKGenService    *cryptoutilJose.JWKGenService
	sharedTLSConfig        *cryptoutilTLSGenerator.TLSGeneratedSettings
	sharedHTTPClient       *http.Client
)

// TestMain initializes shared resources once for all E2E tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize shared telemetry service.
	telemetrySettings := &cryptoutilConfig.ServerSettings{
		LogLevel:     "info",
		OTLPService:  "learn-im-e2e-shared",
		OTLPEnabled:  false, // E2E tests use in-process telemetry only.
		OTLPEndpoint: "grpc://" + cryptoutilMagic.HostnameLocalhost + ":" + "4317",
	}

	var err error
	sharedTelemetryService, err = cryptoutilTelemetry.NewTelemetryService(ctx, telemetrySettings)
	if err != nil {
		panic("failed to initialize shared telemetry service: " + err.Error())
	}

	// Initialize shared JWK generation service.
	sharedJWKGenService, err = cryptoutilJose.NewJWKGenService(ctx, sharedTelemetryService, false)
	if err != nil {
		sharedTelemetryService.Shutdown()
		panic("failed to initialize shared JWK generation service: " + err.Error())
	}

	// Initialize shared TLS configuration (for server).
	sharedTLSConfig, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		sharedTelemetryService.Shutdown()
		panic("failed to generate shared TLS configuration: " + err.Error())
	}

	// Initialize shared HTTP client (for test requests).
	sharedHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.LearnDefaultTimeout,
	}

	// Run all E2E tests.
	exitCode := m.Run()

	// Cleanup shared resources.
	sharedTelemetryService.Shutdown()

	os.Exit(exitCode)
}

// getSharedResources returns the shared test resources.
// Each test creates its own database and server, but reuses telemetry, JWK gen, TLS, and HTTP client.
func getSharedResources(t *testing.T) (*cryptoutilTelemetry.TelemetryService, *cryptoutilJose.JWKGenService, *cryptoutilTLSGenerator.TLSGeneratedSettings, *http.Client) {
	t.Helper()

	require.NotNil(t, sharedTelemetryService, "shared telemetry service must be initialized in TestMain")
	require.NotNil(t, sharedJWKGenService, "shared JWK generation service must be initialized in TestMain")
	require.NotNil(t, sharedTLSConfig, "shared TLS configuration must be initialized in TestMain")
	require.NotNil(t, sharedHTTPClient, "shared HTTP client must be initialized in TestMain")

	return sharedTelemetryService, sharedJWKGenService, sharedTLSConfig, sharedHTTPClient
}
