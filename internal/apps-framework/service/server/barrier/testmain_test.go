// Copyright (c) 2025-2026 Justin Cranford.

package barrier_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

func newBarrierTestTelemetryService(t *testing.T) (*cryptoutilSharedTelemetry.TelemetryService, error) {
	t.Helper()

	ctx := context.Background()
	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings()

	service, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create barrier test telemetry: %w", err)
	}

	t.Cleanup(func() { service.Shutdown() })

	return service, nil
}

func newBarrierTestJWKGenService(t *testing.T, telemetryService *cryptoutilSharedTelemetry.TelemetryService) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
	t.Helper()

	service, err := cryptoutilAppsFrameworkServiceServerBarrier.ExportedNewBarrierTestJWKGenService(context.Background(), telemetryService)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() { cryptoutilAppsFrameworkServiceServerBarrier.ExportedShutdownBarrierTestJWKGenService(service) })

	return service, nil
}

func newBarrierTestJWKGenServiceWithTelemetry(t *testing.T) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
	t.Helper()

	telemetryService, err := newBarrierTestTelemetryService(t)
	if err != nil {
		return nil, err
	}

	return newBarrierTestJWKGenService(t, telemetryService)
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup: Create shared heavyweight resources ONCE.
	dbID, err := googleUuid.NewV7()
	if err != nil {
		panic("TestMain: failed to create SQLite database ID: " + err.Error())
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	testSQLDB, err = sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	if err != nil {
		panic("TestMain: failed to open SQLite: " + err.Error())
	}

	// Configure SQLite for concurrent operations.
	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		panic("TestMain: failed to enable WAL: " + err.Error())
	}

	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		panic("TestMain: failed to set busy timeout: " + err.Error())
	}

	testSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	testDB, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("TestMain: failed to create GORM DB: " + err.Error())
	}

	// Create barrier tables.
	if err := createBarrierTables(testSQLDB); err != nil {
		panic("TestMain: failed to create tables: " + err.Error())
	}

	// Initialize telemetry.
	testTelemetryService, err = cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}
	defer testTelemetryService.Shutdown()

	// Initialize JWK gen service.
	testJWKGenService, err = cryptoutilAppsFrameworkServiceServerBarrier.ExportedNewBarrierTestJWKGenService(ctx, testTelemetryService)
	if err != nil {
		panic("TestMain: failed to create JWK gen service: " + err.Error())
	}
	defer cryptoutilAppsFrameworkServiceServerBarrier.ExportedShutdownBarrierTestJWKGenService(testJWKGenService)

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate unseal JWK: " + err.Error())
	}

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal service: " + err.Error())
	}
	defer unsealService.Shutdown()

	// Create barrier repository.
	barrierRepo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(testDB)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown()

	// Create barrier service.
	testService, err = cryptoutilAppsFrameworkServiceServerBarrier.NewService(
		ctx,
		testTelemetryService,
		testJWKGenService,
		barrierRepo,
		unsealService,
	)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}

	defer testService.Shutdown()
	defer func() {
		if closeErr := testSQLDB.Close(); closeErr != nil {
			panic("TestMain: failed to close test SQL DB: " + closeErr.Error())
		}
	}()

	// Run all tests and release shared resources before os.Exit skips defers.
	exitCode := m.Run()

	testService.Shutdown()
	barrierRepo.Shutdown()
	unsealService.Shutdown()
	cryptoutilAppsFrameworkServiceServerBarrier.ExportedShutdownBarrierTestJWKGenService(testJWKGenService)
	testTelemetryService.Shutdown()

	if closeErr := testSQLDB.Close(); closeErr != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to close test SQL DB: %v\n", closeErr)

		exitCode = 1
	}

	os.Exit(exitCode)
}
