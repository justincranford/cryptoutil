// Copyright (c) 2025 Justin Cranford
//

// Package testing provides test utilities and helpers for cipher-im server testing.
package testing

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	http "net/http"
	"time"

	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsCipherImServer "cryptoutil/internal/apps/cipher/im/server"
	cryptoutilAppsCipherImServerConfig "cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilAppsTemplateServiceTestingE2e "cryptoutil/internal/apps/template/service/testing/e2e"
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
	CipherIMServer *cryptoutilAppsCipherImServer.CipherIMServer
	BaseURL        string
	AdminURL       string

	// Shared services
	JWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	TelemetryService *cryptoutilSharedTelemetry.TelemetryService
	TLSCfg           *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings

	// HTTP client for tests
	HTTPClient *http.Client

	// Shutdown function to clean up all resources
	Shutdown func(ctx context.Context)
}

// SetupTestServer creates a fully configured cipher-im server with all dependencies for testing.
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
	dsn := "file::memory:?cache=shared"

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

	// Create CipherImServerSettings with test settings.
	cfg := &cryptoutilAppsCipherImServerConfig.CipherImServerSettings{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceTestingE2e.NewTestServerSettingsWithService("cipher-im-test"),
	}
	cfg.DatabaseURL = dsn // Set database URL for NewFromConfig

	// Create full server using NewFromConfig.
	resources.CipherIMServer, err = cryptoutilAppsCipherImServer.NewFromConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CipherIMServer: %w", err)
	}

	// Set resources from server.
	resources.DB = resources.CipherIMServer.DB()
	resources.SQLDB, _ = resources.DB.DB()
	resources.JWKGenService = resources.CipherIMServer.JWKGen()
	resources.TelemetryService = resources.CipherIMServer.Telemetry()

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := resources.CipherIMServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for both servers to bind to ports.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	var (
		publicPort int
		adminPort  int
	)

	for i := 0; i < maxWaitAttempts; i++ {
		publicPort = resources.CipherIMServer.PublicPort()

		adminPort = resources.CipherIMServer.AdminPort()

		if publicPort > 0 && adminPort > 0 {
			break
		}

		select {
		case err := <-errChan:
			resources.JWKGenService.Shutdown()
			resources.TelemetryService.Shutdown()
			_ = resources.SQLDB.Close()

			return nil, fmt.Errorf("server start error: %w", err)
		case <-time.After(waitInterval):
		}
	}

	if publicPort == 0 {
		_ = resources.CipherIMServer.Shutdown(ctx)
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("public server did not bind to port")
	}

	if adminPort == 0 {
		_ = resources.CipherIMServer.Shutdown(ctx)
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("admin server did not bind to port")
	}

	resources.BaseURL = fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, publicPort)
	resources.AdminURL = fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, adminPort)

	// Create HTTP client with test TLS config.
	resources.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilSharedMagic.CipherDefaultTimeout,
	}

	// Setup shutdown function.
	resources.Shutdown = func(ctx context.Context) {
		_ = resources.CipherIMServer.Shutdown(ctx)
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()
	}

	return resources, nil
}

// StartCipherIMService creates and starts a cipher-im server from config.
// This is a simpler helper for integration tests that provide their own CipherImServerSettings.
//
// The server is started in the background and this function waits for both public
// and admin servers to bind to their ports before returning.
//
// Example usage:
//
//	CipherImServerSettings := &config.CipherImServerSettings{...}
//	server := StartCipherIMService(CipherImServerSettings)
//	defer server.Shutdown(context.Background())
func StartCipherIMService(CipherImServerSettings *cryptoutilAppsCipherImServerConfig.CipherImServerSettings) *cryptoutilAppsCipherImServer.CipherIMServer {
	ctx := context.Background()

	cipherImServer, err := cryptoutilAppsCipherImServer.NewFromConfig(ctx, CipherImServerSettings)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	// Start server in background (Start() blocks until shutdown).
	errChan := make(chan error, 1)

	go func() {
		if startErr := cipherImServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for both servers to bind to ports.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	var (
		publicPort int
		adminPort  int
	)

	for i := 0; i < maxWaitAttempts; i++ {
		publicPort = cipherImServer.PublicPort()

		adminPort = cipherImServer.AdminPort()

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

	return cipherImServer
}
