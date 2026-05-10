// Copyright (c) 2025-2026 Justin Cranford.
//

package service

import (
	"context"
	"database/sql"
	"os"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilTestDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose-ja/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

var (
	// Test database.
	testDB *gorm.DB

	// Repositories.
	testElasticRepo     cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	testMaterialRepo    cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	testAuditLogRepo    cryptoutilAppsJoseJaRepository.AuditLogRepository
	testAuditConfigRepo cryptoutilAppsJoseJaRepository.AuditConfigRepository

	// Services (dependencies).
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testBarrierService   *cryptoutilAppsFrameworkServiceServerBarrier.Service
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var (
		dbCleanup func()
		err       error
	)

	testDB, dbCleanup, err = cryptoutilTestDb.NewInMemorySQLiteDBForTestMain()
	if err != nil {
		panic("TestMain: failed to create test DB: " + err.Error())
	}
	defer dbCleanup()

	// Run migrations using underlying sql.DB.
	testSQLDB, err := testDB.DB()
	if err != nil {
		panic("TestMain: failed to get sql.DB: " + err.Error())
	}

	if err := cryptoutilAppsJoseJaRepository.ApplyJoseJAMigrations(testSQLDB, cryptoutilAppsJoseJaRepository.DatabaseTypeSQLite); err != nil {
		panic("TestMain: failed to run migrations: " + err.Error())
	}

	// Create repositories.
	testElasticRepo = cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(testDB)
	testMaterialRepo = cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(testDB)
	testAuditLogRepo = cryptoutilAppsJoseJaRepository.NewAuditLogRepository(testDB)
	testAuditConfigRepo = cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(testDB)

	// Initialize telemetry.
	telemetrySettings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	testTelemetryService, err = cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetrySettings.ToTelemetrySettings())
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}
	defer testTelemetryService.Shutdown()

	// Initialize JWK Generation Service.
	testJWKGenService, err = cryptoutilSharedCryptoJose.NewJWKGenService(ctx, testTelemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK service: " + err.Error())
	}
	defer testJWKGenService.Shutdown()

	// Initialize Barrier Service.
	// Generate a simple test unseal key using JWE with A256GCM encryption and A256KW key wrapping.
	_, testUnsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate test unseal JWK: " + err.Error())
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal keys service: " + err.Error())
	}
	defer unsealKeysService.Shutdown()

	barrierRepo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(testDB)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown()

	testBarrierService, err = cryptoutilAppsFrameworkServiceServerBarrier.NewService(ctx, testTelemetryService, testJWKGenService, barrierRepo, unsealKeysService)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}
	defer testBarrierService.Shutdown()

	// Run all tests.
	exitCode := m.Run()

	// Cleanup happens via defer.
	os.Exit(exitCode)
}

func newClosedServiceDeps(t *testing.T) (cryptoutilAppsJoseJaRepository.ElasticJWKRepository, cryptoutilAppsJoseJaRepository.MaterialJWKRepository, cryptoutilAppsJoseJaRepository.AuditLogRepository, cryptoutilAppsJoseJaRepository.AuditConfigRepository) {
	t.Helper()

	closedDB := cryptoutilTestDb.NewClosedSQLiteDB(t, func(sqlDB *sql.DB) error {
		return cryptoutilAppsJoseJaRepository.ApplyJoseJAMigrations(sqlDB, cryptoutilAppsJoseJaRepository.DatabaseTypeSQLite)
	})

	elasticRepo := cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(closedDB)
	materialRepo := cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(closedDB)
	auditLogRepo := cryptoutilAppsJoseJaRepository.NewAuditLogRepository(closedDB)
	auditConfigRepo := cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(closedDB)

	return elasticRepo, materialRepo, auditLogRepo, auditConfigRepo
}

func closedDBMaterialRepo(t *testing.T) cryptoutilAppsJoseJaRepository.MaterialJWKRepository {
	t.Helper()

	closedDB := cryptoutilTestDb.NewClosedSQLiteDB(t, func(sqlDB *sql.DB) error {
		return cryptoutilAppsJoseJaRepository.ApplyJoseJAMigrations(sqlDB, cryptoutilAppsJoseJaRepository.DatabaseTypeSQLite)
	})

	return cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(closedDB)
}
