// Copyright (c) 2025 Justin Cranford
//

package testing

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilE2E "cryptoutil/internal/template/testing/e2e"
)

// TestServerResources holds shared resources created by SetupTestServer.
type TestServerResources struct {
	// Database resources
	DB    *gorm.DB
	SQLDB *sql.DB

	// Server resources
	CipherIMServer *server.CipherIMServer
	BaseURL        string
	AdminURL       string

	// Shared services
	JWKGenService    *cryptoutilJose.JWKGenService
	TelemetryService *cryptoutilTelemetry.TelemetryService
	TLSCfg           *cryptoutilTLSGenerator.TLSGeneratedSettings

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
func SetupTestServer(ctx context.Context, useInMemoryDB bool) (*TestServerResources, error) {
	resources := &TestServerResources{}

	// Setup database.
	var dsn string
	if useInMemoryDB {
		dsn = "file::memory:?cache=shared"
	} else {
		dbID, err := googleUuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("failed to generate DB ID: %w", err)
		}

		dsn = "file:" + dbID.String() + "?mode=memory&cache=shared"
	}

	// CRITICAL: Store sql.DB reference in returned resources.
	// In-memory SQLite databases are destroyed when all connections close.
	// Storing reference prevents GC from closing connection during parallel test execution.
	var err error

	resources.SQLDB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Configure SQLite for concurrent operations.
	if _, err := resources.SQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	if _, err := resources.SQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	resources.SQLDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	resources.SQLDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	resources.SQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	resources.DB, err = gorm.Open(sqlite.Dialector{Conn: resources.SQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to create GORM DB: %w", err)
	}

	// Run migrations.
	if err := repository.ApplyMigrations(resources.SQLDB, repository.DatabaseTypeSQLite); err != nil {
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize telemetry.
	resources.TelemetryService, err = cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	if err != nil {
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to create telemetry: %w", err)
	}

	// Initialize JWK Generation Service.
	resources.JWKGenService, err = cryptoutilJose.NewJWKGenService(ctx, resources.TelemetryService, false)
	if err != nil {
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to create JWK service: %w", err)
	}

	// Generate TLS config for HTTP client.
	resources.TLSCfg, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Generate JWT secret.
	jwtSecretID, err := googleUuid.NewV7()
	if err != nil {
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// Create AppConfig with test settings.
	cfg := &config.AppConfig{
		ServerSettings: *cryptoutilE2E.NewTestServerSettingsWithService("cipher-im-test"),
		JWTSecret:      jwtSecretID.String(),
	}

	// Create full server.
	resources.CipherIMServer, err = server.New(ctx, cfg, resources.DB, repository.DatabaseTypeSQLite)
	if err != nil {
		resources.JWKGenService.Shutdown()
		resources.TelemetryService.Shutdown()
		_ = resources.SQLDB.Close()

		return nil, fmt.Errorf("failed to create CipherIMServer: %w", err)
	}

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

		adminPortValue, _ := resources.CipherIMServer.AdminPort()
		adminPort = adminPortValue

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

	resources.BaseURL = fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, publicPort)
	resources.AdminURL = fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, adminPort)

	// Create HTTP client with test TLS config.
	resources.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.CipherDefaultTimeout,
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

// StartCipherIMServer creates and starts a cipher-im server from config.
// This is a simpler helper for integration tests that provide their own AppConfig.
//
// The server is started in the background and this function waits for both public
// and admin servers to bind to their ports before returning.
//
// Example usage:
//
//	appConfig := &config.AppConfig{...}
//	server := StartCipherIMServer(appConfig)
//	defer server.Shutdown(context.Background())
func StartCipherIMServer(appConfig *config.AppConfig) *server.CipherIMServer {
	ctx := context.Background()

	cipherImServer, err := server.NewFromConfig(ctx, appConfig)
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

		adminPortValue, _ := cipherImServer.AdminPort()
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

	return cipherImServer
}
