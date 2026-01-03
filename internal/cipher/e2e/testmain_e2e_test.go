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

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"

	"cryptoutil/internal/cipher/server"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilBarrier "cryptoutil/internal/template/server/barrier"
)

// Shared test resources (initialized once per package).
var (
	sharedTelemetryService *cryptoutilTelemetry.TelemetryService
	sharedJWKGenService    *cryptoutilJose.JWKGenService
	sharedTLSConfig        *cryptoutilTLSGenerator.TLSGeneratedSettings
	sharedHTTPClient       *http.Client
	testPublicServer       *server.PublicServer
	baseURL                string
	testBarrierService     *cryptoutilBarrier.BarrierService
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
	defer sqlDB.Close() // LIFO: close database after services using it.

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := sharedJWKGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	if err != nil {
		panic("failed to generate unseal JWK: " + err.Error())
	}

	// Initialize unseal keys service.
	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	if err != nil {
		panic("failed to create unseal service: " + err.Error())
	}
	defer unsealService.Shutdown() // LIFO: cleanup unseal service.

	// Initialize barrier repository.
	barrierRepo, err := cryptoutilBarrier.NewGormBarrierRepository(db)
	if err != nil {
		panic("failed to initialize barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown() // LIFO: cleanup barrier repository.

	// Initialize barrier service for E2E tests.
	testBarrierService, err = cryptoutilBarrier.NewBarrierService(ctx, sharedTelemetryService, sharedJWKGenService, barrierRepo, unsealService)
	if err != nil {
		panic("failed to initialize test barrier service: " + err.Error())
	}
	defer testBarrierService.Shutdown() // LIFO: cleanup barrier service.

	testPublicServer, baseURL, err = createTestPublicServer(db)
	if err != nil {
		panic("failed to create test public server: " + err.Error())
	}

	defer func() {
		_ = testPublicServer.Shutdown(context.Background())
	}() // LIFO: shutdown server.

	// Run all E2E tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}
