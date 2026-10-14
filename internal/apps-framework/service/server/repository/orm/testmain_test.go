// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"gorm.io/gorm"

	cryptoutilTestHelpDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

const testOTLPInstance = "test"

var (
	testCtx              context.Context
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testGormDB           *gorm.DB
	testOrmRepository    *OrmRepository
)

func TestMain(m *testing.M) {
	exitCode := func() (exitCode int) {
		testCtx = context.Background()

		var err error

		testTelemetryService, err = cryptoutilSharedTelemetry.NewTelemetryService(testCtx, &cryptoutilSharedTelemetry.TelemetrySettings{
			LogLevel:        cryptoutilSharedMagic.DefaultLogLevelInfo,
			OTLPEnabled:     false,
			OTLPService:     "framework-orm-test",
			OTLPInstance:    testOTLPInstance,
			OTLPVersion:     "0.0.0-test",
			OTLPEnvironment: "test",
			OTLPHostname:    cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
			OTLPEndpoint:    cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		})
		cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create TelemetryService")

		defer testTelemetryService.Shutdown()

		testJWKGenService, err = cryptoutilSharedCryptoJose.NewJWKGenService(testCtx, testTelemetryService, false)
		cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create JWKGenService")

		defer testJWKGenService.Shutdown()

		var dbCleanup func() error

		testGormDB, dbCleanup, err = cryptoutilTestHelpDb.NewInMemorySQLiteDBForTestMain()
		cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create in-memory SQLite DB")

		defer func() {
			if cleanupErr := dbCleanup(); cleanupErr != nil {
				fmt.Fprintf(os.Stderr, "TestMain: failed to close test database: %v\n", cleanupErr)

				exitCode = 1
			}
		}()

		testOrmRepository, err = NewOrmRepository(testCtx, testTelemetryService, testGormDB, testJWKGenService, false)
		cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create OrmRepository")

		exitCode = m.Run()

		return exitCode
	}()
	os.Exit(exitCode)
}
