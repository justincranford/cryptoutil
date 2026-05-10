// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"context"
	"os"
	"testing"

	"gorm.io/gorm"

	cryptoutilTestHelpDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

var (
	testCtx              context.Context
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testGormDB           *gorm.DB
	testOrmRepository    *OrmRepository
)

func TestMain(m *testing.M) {
	testCtx = context.Background()

	var err error

	testTelemetryService, err = cryptoutilSharedTelemetry.NewTelemetryService(testCtx, &cryptoutilSharedTelemetry.TelemetrySettings{
		LogLevel:        cryptoutilSharedMagic.DefaultLogLevelInfo,
		OTLPEnabled:     false,
		OTLPService:     "framework-orm-test",
		OTLPInstance:    "test",
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

	var dbCleanup func()

	testGormDB, dbCleanup, err = cryptoutilTestHelpDb.NewInMemorySQLiteDBForTestMain()
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create in-memory SQLite DB")

	defer dbCleanup()

	testOrmRepository, err = NewOrmRepository(testCtx, testTelemetryService, testGormDB, testJWKGenService, false)
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create OrmRepository")

	os.Exit(m.Run())
}
