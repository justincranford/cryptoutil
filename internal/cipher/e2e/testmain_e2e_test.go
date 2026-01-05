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

	cryptoutilCipherServer "cryptoutil/internal/cipher/server"
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
	testCipherIMServer     *cryptoutilCipherServer.CipherIMServer
	baseURL                string
	adminURL               string
)

// TestMain initializes shared resources once for all E2E tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize shared telemetry service.
	telemetrySettings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)

	var err error

	sharedTelemetryService, err = cryptoutilTelemetry.NewTelemetryService(ctx, telemetrySettings)
	if err != nil {
		panic("failed to initialize shared telemetry service: " + err.Error())
	}
	defer sharedTelemetryService.Shutdown() // LIFO: cleanup last service created first.

	// Initialize shared JWK generation service.
	sharedJWKGenService, err = cryptoutilJose.NewJWKGenService(ctx, sharedTelemetryService, false)
	if err != nil {
		panic("failed to initialize shared JWK generation service: " + err.Error())
	}
	defer sharedJWKGenService.Shutdown() // LIFO: cleanup after telemetry.

	// Initialize shared TLS configuration (for server).
	sharedTLSConfig, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		panic("failed to generate shared TLS configuration: " + err.Error())
	}

	// Initialize shared HTTP client (for test requests).
	sharedHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.CipherDefaultTimeout,
	}

	db, err := initTestDB()
	if err != nil {
		panic("failed to initialize test database: " + err.Error())
	}

	// Extract sql.DB for proper cleanup.
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get sql.DB from gorm.DB: " + err.Error())
	}

	defer func() {
		_ = sqlDB.Close() // LIFO: close database after services using it.
	}()

	testCipherIMServer, baseURL, adminURL, err = createTestCipherIMServer(db)
	if err != nil {
		panic("failed to create test cipher-im server: " + err.Error())
	}

	defer func() {
		_ = testCipherIMServer.Shutdown(context.Background())
	}() // LIFO: shutdown server.

	// Run all E2E tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}
