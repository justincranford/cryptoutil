// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for SM-KMS businesslogic tests.

package businesslogic

import (
	"context"
	"os"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilAppsFrameworkServiceTestHelpDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm-kms/server/repository/orm"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

type testBasicFixture struct {
	TelemetryService *cryptoutilSharedTelemetry.TelemetryService
	JWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
}

type testCoreFixture struct {
	DB    *gorm.DB
	Basic *testBasicFixture
}

var (
	testCore           *testCoreFixture
	testBarrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	testDB, dbCleanup, err := cryptoutilAppsFrameworkServiceTestHelpDb.NewInMemorySQLiteDBForTestMain()
	if err != nil {
		panic("TestMain: failed to create test database: " + err.Error())
	}
	defer dbCleanup()

	telemetrySettings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetrySettings.ToTelemetrySettings())
	if err != nil {
		panic("TestMain: failed to create telemetry service: " + err.Error())
	}
	defer telemetryService.Shutdown()

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK generator: " + err.Error())
	}
	defer jwkGenService.Shutdown()

	_, testUnsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate unseal JWK: " + err.Error())
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal keys service: " + err.Error())
	}
	defer unsealKeysService.Shutdown()

	if err := testDB.AutoMigrate(
		&cryptoutilAppsFrameworkServiceServerBarrier.RootKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.ContentKey{},
	); err != nil {
		panic("TestMain: failed to migrate barrier tables: " + err.Error())
	}

	barrierRepo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(testDB)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown()

	testBarrierService, err = cryptoutilAppsFrameworkServiceServerBarrier.NewService(ctx, telemetryService, jwkGenService, barrierRepo, unsealKeysService)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}
	defer testBarrierService.Shutdown()

	testCore = &testCoreFixture{
		DB: testDB,
		Basic: &testBasicFixture{
			TelemetryService: telemetryService,
			JWKGenService:    jwkGenService,
		},
	}
	if err := testCore.DB.AutoMigrate(&cryptoutilOrmRepository.ElasticKey{}, &cryptoutilOrmRepository.MaterialKey{}); err != nil {
		panic("TestMain: failed to migrate KMS tables: " + err.Error())
	}

	exitCode := m.Run()

	os.Exit(exitCode)
}
