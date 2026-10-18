// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"context"
	"fmt"
	"os"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceServerUnsealKeys "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilAppsFrameworkServiceTestHelpDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"gorm.io/gorm"
)

var (
	testContext        = context.Background()
	testDB             *gorm.DB
	testJWKGenService  *cryptoutilSharedCryptoJose.JWKGenService
	testBarrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service
)

func TestMain(m *testing.M) {
	exitCode := func() (code int) {
		db, dbCleanup, err := cryptoutilAppsFrameworkServiceTestHelpDb.NewInMemorySQLiteDBForTestMain()
		if err != nil {
			panic(fmt.Sprintf("failed to create test database: %v", err))
		}
		defer func() {
			if cleanupErr := dbCleanup(); cleanupErr != nil {
				fmt.Fprintf(os.Stderr, "failed to close test database: %v\n", cleanupErr)

				code = 1
			}
		}()

		sqlDB, err := db.DB()
		if err != nil {
			panic(fmt.Sprintf("failed to access test database: %v", err))
		}

		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)

		testDB = db

		settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

		telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(testContext, settings.ToTelemetrySettings())
		if err != nil {
			panic(fmt.Sprintf("failed to create telemetry service: %v", err))
		}
		defer telemetryService.Shutdown()

		testJWKGenService, err = cryptoutilSharedCryptoJose.NewJWKGenService(testContext, telemetryService, false)
		if err != nil {
			panic(fmt.Sprintf("failed to create JWK generator: %v", err))
		}
		defer testJWKGenService.Shutdown()

		_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(
			&cryptoutilSharedCryptoJose.EncA256GCM,
			&cryptoutilSharedCryptoJose.AlgA256KW,
		)
		if err != nil {
			panic(fmt.Sprintf("failed to generate unseal JWK: %v", err))
		}

		unsealKeysService, err := cryptoutilAppsFrameworkServiceServerUnsealKeys.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
		if err != nil {
			panic(fmt.Sprintf("failed to create unseal keys service: %v", err))
		}
		defer unsealKeysService.Shutdown()

		if err := db.AutoMigrate(
			&cryptoutilAppsFrameworkServiceServerBarrier.RootKey{},
			&cryptoutilAppsFrameworkServiceServerBarrier.IntermediateKey{},
			&cryptoutilAppsFrameworkServiceServerBarrier.ContentKey{},
			&cryptoutilAppsSmKmsModel.ElasticJWK{},
			&cryptoutilAppsSmKmsModel.MaterialJWK{},
			&cryptoutilAppsSmKmsModel.AuditConfig{},
			&cryptoutilAppsSmKmsModel.AuditLogEntry{},
		); err != nil {
			panic(fmt.Sprintf("failed to migrate test tables: %v", err))
		}

		barrierRepo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
		if err != nil {
			panic(fmt.Sprintf("failed to create barrier repository: %v", err))
		}
		defer barrierRepo.Shutdown()

		testBarrierService, err = cryptoutilAppsFrameworkServiceServerBarrier.NewService(
			testContext,
			telemetryService,
			testJWKGenService,
			barrierRepo,
			unsealKeysService,
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create barrier service: %v", err))
		}
		defer testBarrierService.Shutdown()

		code = m.Run()

		return code
	}()
	os.Exit(exitCode)
}

func testRepositories() (
	cryptoutilAppsSmKmsRepository.ElasticJWKRepository,
	cryptoutilAppsSmKmsRepository.MaterialJWKRepository,
) {
	return cryptoutilAppsSmKmsRepository.NewElasticJWKRepository(testDB),
		cryptoutilAppsSmKmsRepository.NewMaterialJWKRepository(testDB)
}
