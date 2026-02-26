// Copyright (c) 2025 Justin Cranford
//

// Package testing provides test utilities and helpers for sm-im server testing.
package testing

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	http "net/http"

	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestServerResources holds shared resources created by SetupTestServer.
type TestServerResources struct {
	// Database resources
	DB    *gorm.DB
	SQLDB *sql.DB

	// Server resources
	SmIMServer *cryptoutilAppsSmImServer.SmIMServer
	BaseURL    string
	AdminURL   string

	// Shared services
	JWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	TelemetryService *cryptoutilSharedTelemetry.TelemetryService
	TLSCfg           *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings

	// HTTP client for tests
	HTTPClient *http.Client

	// Shutdown function to clean up all resources
	Shutdown func(ctx context.Context)
}

// SetupTestServer creates a fully configured sm-im server with all dependencies for testing.
// It returns TestServerResources containing the server, database, shared services, and a shutdown function.
//
// The caller MUST call resources.Shutdown(ctx) when done to clean up all resources.
//
// Example usage:
//
//	resources, err := SetupTestServer(ctx, false)
//	if err != nil {
//	    panic(err)
//	}
//	defer resources.Shutdown(context.Background())
func SetupTestServer(ctx context.Context, _ bool) (*TestServerResources, error) {
	resources := &TestServerResources{}

	// Setup database DSN - always use normalized in-memory format for tests.
	// Note: WAL mode is handled by application_core.go which special-cases in-memory databases.
	dsn := cryptoutilSharedMagic.SQLiteInMemoryDSN

	// Generate TLS config for HTTP client.
	var err error

	resources.TLSCfg, err = cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilSharedMagic.HostnameLocalhost},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Create SmIMServerSettings with test settings.
	cfg := &cryptoutilAppsSmImServerConfig.SmIMServerSettings{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceTestingE2eHelpers.NewTestServerSettingsWithService("sm-im-test"),
	}
	cfg.DatabaseURL = dsn // Set database URL for NewFromConfig

	// Create full server using NewFromConfig.
	resources.SmIMServer, err = cryptoutilAppsSmImServer.NewFromConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SmIMServer: %w", err)
	}

	// Set resources from server.
	resources.DB = resources.SmIMServer.DB()
	resources.SQLDB, _ = resources.DB.DB()
	resources.JWKGenService = resources.SmIMServer.JWKGen()
	resources.TelemetryService = resources.SmIMServer.Telemetry()

	// Start server in background and wait for both ports to bind.
	errChan := cryptoutilAppsTemplateServiceTestingE2eHelpers.StartDualPortServerAsync(func() error {
		return resources.SmIMServer.Start(ctx)
	})

	if err = cryptoutilAppsTemplateServiceTestingE2eHelpers.WaitForDualServerPorts(resources.SmIMServer, errChan); err != nil {
		_ = resources.SmIMServer.Shutdown(ctx)
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to wait for sm-im server ports: %w", err)
	}

	resources.BaseURL, resources.AdminURL = cryptoutilAppsTemplateServiceTestingE2eHelpers.DualPortBaseURLs(resources.SmIMServer)

	// Create HTTP client with test TLS config.
	resources.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilSharedMagic.IMDefaultTimeout,
	}

	// Setup shutdown function.
	resources.Shutdown = func(ctx context.Context) {
		_ = resources.SmIMServer.Shutdown(ctx)
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()
	}

	return resources, nil
}

// StartSmIMService creates and starts a sm-im server from config.
// This is a simpler helper for integration tests that provide their own SmIMServerSettings.
//
// The server is started in the background and this function waits for both public
// and admin servers to bind to their ports before returning.
//
// Example usage:
//
//	SmIMServerSettings := &config.SmIMServerSettings{...}
//	server := StartSmIMService(SmIMServerSettings)
//	defer server.Shutdown(context.Background())
func StartSmIMService(SmIMServerSettings *cryptoutilAppsSmImServerConfig.SmIMServerSettings) *cryptoutilAppsSmImServer.SmIMServer {
	ctx := context.Background()

	smIMServer, err := cryptoutilAppsSmImServer.NewFromConfig(ctx, SmIMServerSettings)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	// Use generic template helper for goroutine start + dual port polling + panic-on-failure.
	cryptoutilAppsTemplateServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(smIMServer, func() error {
		return smIMServer.Start(ctx)
	})

	return smIMServer
}
