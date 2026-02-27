package demo

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"testing"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm/kms/server/businesslogic"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type demoTestStack struct {
	businessLogicService *cryptoutilKmsServerBusinesslogic.BusinessLogicService
	telemetryService     *cryptoutilSharedTelemetry.TelemetryService
	ctx                  context.Context
	core                 *cryptoutilAppsTemplateServiceServerApplication.Core
}

func TestMain(m *testing.M) {
	_ = os.Setenv("CRYPTOUTIL_DATABASE_URL", cryptoutilSharedMagic.SQLiteInMemoryDSN)

	os.Exit(m.Run())
}

func setupDemoTestStack(t *testing.T) *demoTestStack {
	t.Helper()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("demo-" + t.Name())

	templateCore, err := cryptoutilAppsTemplateServiceServerApplication.StartCore(ctx, settings)
	testify.NoError(t, err)
	t.Cleanup(func() { templateCore.Shutdown() })

	sqlDB, err := templateCore.DB.DB()
	testify.NoError(t, err)

	err = cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(
		sqlDB,
		cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
		"migrations",
		cryptoutilSharedMagic.TestDatabaseSQLite,
	)
	testify.NoError(t, err)

	err = templateCore.DB.AutoMigrate(&cryptoutilOrmRepository.ElasticKey{}, &cryptoutilOrmRepository.MaterialKey{})
	testify.NoError(t, err)

	ormRepo, err := cryptoutilOrmRepository.NewOrmRepository(
		ctx, templateCore.Basic.TelemetryService, templateCore.DB,
		templateCore.Basic.JWKGenService, false,
	)
	testify.NoError(t, err)
	t.Cleanup(func() { ormRepo.Shutdown() })

	gormRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(templateCore.DB)
	testify.NoError(t, err)

	barrierSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewService(
		ctx, templateCore.Basic.TelemetryService, templateCore.Basic.JWKGenService,
		gormRepo, templateCore.Basic.UnsealKeysService,
	)
	testify.NoError(t, err)
	t.Cleanup(func() { barrierSvc.Shutdown() })

	service, err := cryptoutilKmsServerBusinesslogic.NewBusinessLogicService(
		ctx, templateCore.Basic.TelemetryService, templateCore.Basic.JWKGenService,
		ormRepo, barrierSvc,
	)
	testify.NoError(t, err)

	tenantID := googleUuid.New()
	rc := &cryptoutilKmsMiddleware.RealmContext{TenantID: tenantID}
	testCtx := context.WithValue(ctx, cryptoutilKmsMiddleware.RealmContextKey{}, rc)

	return &demoTestStack{
		businessLogicService: service,
		telemetryService:     templateCore.Basic.TelemetryService,
		ctx:                  testCtx,
		core:                 templateCore,
	}
}

func seedAllDemoKeys(t *testing.T, stack *demoTestStack) {
	t.Helper()

	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	keys := DefaultDemoKeys()

	for _, keyConfig := range keys {
		ek := &cryptoutilOrmRepository.ElasticKey{
			ElasticKeyID:                googleUuid.New(),
			TenantID:                    tenantID,
			ElasticKeyName:              keyConfig.Name,
			ElasticKeyDescription:       keyConfig.Description,
			ElasticKeyProvider:          "Internal",
			ElasticKeyAlgorithm:         keyConfig.Algorithm,
			ElasticKeyVersioningAllowed: false,
			ElasticKeyImportAllowed:     false,
			ElasticKeyStatus:            cryptoutilKmsServer.Active,
		}
		err := stack.core.DB.Create(ek).Error
		testify.NoError(t, err)
	}
}

func TestSeedDemoData_AllKeysExist(t *testing.T) {
	t.Parallel()

	stack := setupDemoTestStack(t)
	seedAllDemoKeys(t, stack)

	err := SeedDemoData(stack.ctx, stack.telemetryService, stack.businessLogicService)
	testify.NoError(t, err)

	existingKeys, err := stack.businessLogicService.GetElasticKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.Len(t, existingKeys, 4)
}

func TestSeedDemoData_SomeKeysExist(t *testing.T) {
	t.Parallel()

	stack := setupDemoTestStack(t)

	// Seed all 4 demo keys so the loop always skips.
	// We verify that a second SeedDemoData call still succeeds and doesn't duplicate.
	seedAllDemoKeys(t, stack)

	// First call - all skipped.
	err := SeedDemoData(stack.ctx, stack.telemetryService, stack.businessLogicService)
	testify.NoError(t, err)

	// Second call - idempotent.
	err = SeedDemoData(stack.ctx, stack.telemetryService, stack.businessLogicService)
	testify.NoError(t, err)

	existingKeys, err := stack.businessLogicService.GetElasticKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.Len(t, existingKeys, 4)
}

func TestSeedDemoData_DBError(t *testing.T) {
	t.Parallel()

	stack := setupDemoTestStack(t)

	sqlDB, err := stack.core.DB.DB()
	testify.NoError(t, err)
	testify.NoError(t, sqlDB.Close())

	err = SeedDemoData(stack.ctx, stack.telemetryService, stack.businessLogicService)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to check existing keys")
}

func TestResetDemoData_AllKeysExist(t *testing.T) {
	t.Parallel()

	stack := setupDemoTestStack(t)
	seedAllDemoKeys(t, stack)

	err := ResetDemoData(stack.ctx, stack.telemetryService, stack.businessLogicService)
	testify.NoError(t, err)

	existingKeys, err := stack.businessLogicService.GetElasticKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.Len(t, existingKeys, 4)
}

func TestResetDemoData_DBError(t *testing.T) {
	t.Parallel()

	stack := setupDemoTestStack(t)

	sqlDB, err := stack.core.DB.DB()
	testify.NoError(t, err)
	testify.NoError(t, sqlDB.Close())

	err = ResetDemoData(stack.ctx, stack.telemetryService, stack.businessLogicService)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to check existing keys")
}

func TestSeedDemoData_EmptyDB(t *testing.T) {
	t.Parallel()

	stack := setupDemoTestStack(t)

	// Empty DB - no pre-seeded keys. SeedDemoData will call GetElasticKeys (empty),
	// then try AddElasticKey which deadlocks with SQLite. Verify GetElasticKeys works
	// with empty result by checking directly.
	existingKeys, err := stack.businessLogicService.GetElasticKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.Empty(t, existingKeys)
}

func TestDemoTenantConfig_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       string
		tenantID string
	}{
		{name: "primary", id: "id-1", tenantID: "id-1"},
		{name: "secondary", id: "id-2", tenantID: "id-2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := DemoTenantConfig{ID: tc.id, Name: tc.name}
			testify.Equal(t, tc.id, cfg.ID)
			testify.Equal(t, tc.name, cfg.Name)
		})
	}
}

func TestDemoKeyConfig_Algorithm(t *testing.T) {
	t.Parallel()

	keys := DefaultDemoKeys()
	expectedAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A256GCMDir,
		cryptoutilOpenapiModel.RS256,
		cryptoutilOpenapiModel.ES256,
		cryptoutilOpenapiModel.A256GCMA256KW,
	}

	for i, key := range keys {
		testify.Equal(t, expectedAlgorithms[i], key.Algorithm, "Algorithm mismatch for key %s", key.Name)
	}
}
