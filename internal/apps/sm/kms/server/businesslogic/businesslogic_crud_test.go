package businesslogic

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"io/fs"
	"os"
	"testing"
	"time"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
	cryptoutilKmsServerRepository "cryptoutil/internal/apps/sm/kms/server/repository"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

const currentDir = "."

type testMergedMigrations struct {
	templateFS   fs.FS
	templatePath string
	domainFS     fs.FS
	domainPath   string
}

func (m *testMergedMigrations) Open(name string) (fs.File, error) {
	if m.domainFS != nil {
		p := m.domainPath
		if name != currentDir && name != "" {
			p = m.domainPath + "/" + name
		}

		if f, err := m.domainFS.Open(p); err == nil {
			return f, nil
		}
	}

	p := m.templatePath
	if name != currentDir && name != "" {
		p = m.templatePath + "/" + name
	}

	return m.templateFS.Open(p)
}

func (m *testMergedMigrations) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry

	tp := m.templatePath
	if name != currentDir && name != "" {
		tp = m.templatePath + "/" + name
	}

	if te, err := fs.ReadDir(m.templateFS, tp); err == nil {
		entries = append(entries, te...)
	}

	if m.domainFS != nil {
		dp := m.domainPath
		if name != currentDir && name != "" {
			dp = m.domainPath + "/" + name
		}

		if de, err := fs.ReadDir(m.domainFS, dp); err == nil {
			entries = append(entries, de...)
		}
	}

	return entries, nil
}

func (m *testMergedMigrations) ReadFile(name string) ([]byte, error) {
	if m.domainFS != nil {
		dp := m.domainPath
		if name != currentDir && name != "" {
			dp = m.domainPath + "/" + name
		}

		if data, err := fs.ReadFile(m.domainFS, dp); err == nil {
			return data, nil
		}
	}

	tp := m.templatePath
	if name != currentDir && name != "" {
		tp = m.templatePath + "/" + name
	}

	return fs.ReadFile(m.templateFS, tp)
}

type testStack struct {
	service *BusinessLogicService
	ctx     context.Context
	core    *cryptoutilAppsTemplateServiceServerApplication.Core
}

func setupTestStack(t *testing.T) *testStack {
	t.Helper()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("biz-crud-" + t.Name())

	templateCore, err := cryptoutilAppsTemplateServiceServerApplication.StartCore(ctx, settings)
	testify.NoError(t, err)
	t.Cleanup(func() { templateCore.Shutdown() })

	sqlDB, err := templateCore.DB.DB()
	testify.NoError(t, err)

	mergedFS := &testMergedMigrations{
		templateFS:   cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
		templatePath: "migrations",
		domainFS:     cryptoutilKmsServerRepository.MigrationsFS,
		domainPath:   "migrations",
	}

	err = cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(sqlDB, mergedFS, "", cryptoutilSharedMagic.TestDatabaseSQLite)
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

	service, err := NewBusinessLogicService(
		ctx, templateCore.Basic.TelemetryService, templateCore.Basic.JWKGenService,
		ormRepo, barrierSvc,
	)
	testify.NoError(t, err)

	tenantID := googleUuid.New()
	rc := &cryptoutilKmsMiddleware.RealmContext{TenantID: tenantID}
	testCtx := context.WithValue(ctx, cryptoutilKmsMiddleware.RealmContextKey{}, rc)

	return &testStack{service: service, ctx: testCtx, core: templateCore}
}

func seedElasticKey(t *testing.T, stack *testStack, name string, alg cryptoutilOpenapiModel.ElasticKeyAlgorithm, status cryptoutilKmsServer.ElasticKeyStatus) googleUuid.UUID {
	t.Helper()

	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	ekID := googleUuid.New()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       "test-desc",
		ElasticKeyProvider:          "Internal",
		ElasticKeyAlgorithm:         alg,
		ElasticKeyVersioningAllowed: false,
		ElasticKeyImportAllowed:     false,
		ElasticKeyStatus:            status,
	}
	err := stack.core.DB.Create(ek).Error
	testify.NoError(t, err)

	return ekID
}

func seedMaterialKey(t *testing.T, stack *testStack, ekID googleUuid.UUID) googleUuid.UUID {
	t.Helper()

	mkID := googleUuid.New()
	now := time.Now().UTC().UnixMilli()
	mk := &cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  ekID,
		MaterialKeyID:                 mkID,
		MaterialKeyClearPublic:        nil,
		MaterialKeyEncryptedNonPublic: []byte("encrypted-placeholder"),
		MaterialKeyGenerateDate:       &now,
	}
	err := stack.core.DB.Create(mk).Error
	testify.NoError(t, err)

	return mkID
}

func TestGetElasticKeyByID(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "get-by-id", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	ek, err := stack.service.GetElasticKeyByElasticKeyID(stack.ctx, &ekID)
	testify.NoError(t, err)
	testify.NotNil(t, ek)
	testify.Equal(t, "get-by-id", *ek.Name)
}

func TestGetElasticKeyByID_NotFound(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	missingID := googleUuid.New()
	_, err := stack.service.GetElasticKeyByElasticKeyID(stack.ctx, &missingID)
	testify.Error(t, err)
}

func TestGetElasticKeys(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	seedElasticKey(t, stack, "list-a", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedElasticKey(t, stack, "list-b", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	all, err := stack.service.GetElasticKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.Len(t, all, 2)
}

func TestGetMaterialKeysForElasticKey(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "mat-for-ek", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID)
	seedMaterialKey(t, stack, ekID)
	mks, err := stack.service.GetMaterialKeysForElasticKey(stack.ctx, &ekID, nil)
	testify.NoError(t, err)
	testify.Len(t, mks, 2)
}

func TestGetMaterialKeys(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "mat-all", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID)
	mks, err := stack.service.GetMaterialKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.GreaterOrEqual(t, len(mks), 1)
}

func TestGetMaterialKeyByIDs(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "mat-by-ids", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	mkID := seedMaterialKey(t, stack, ekID)
	mk, err := stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(stack.ctx, &ekID, &mkID)
	testify.NoError(t, err)
	testify.NotNil(t, mk)
}

func TestGetMaterialKeyByIDs_NotFound(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "mat-nf", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	missingMK := googleUuid.New()
	_, err := stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(stack.ctx, &ekID, &missingMK)
	testify.Error(t, err)
}

func TestUpdateElasticKey(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "update-me", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	newDesc := "updated-desc"
	updated, err := stack.service.UpdateElasticKey(stack.ctx, &ekID, &cryptoutilKmsServer.ElasticKeyUpdate{
		Name:        "updated-name",
		Description: &newDesc,
	})
	testify.NoError(t, err)
	testify.Equal(t, "updated-name", *updated.Name)
}

func TestUpdateElasticKey_NotFound(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	missingID := googleUuid.New()
	newDesc := "desc"
	_, err := stack.service.UpdateElasticKey(stack.ctx, &missingID, &cryptoutilKmsServer.ElasticKeyUpdate{
		Name:        "x",
		Description: &newDesc,
	})
	testify.Error(t, err)
}

func TestDeleteElasticKey_Active(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "del-active", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	err := stack.service.DeleteElasticKey(stack.ctx, &ekID)
	testify.NoError(t, err)
}

func TestDeleteElasticKey_Disabled(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "del-disabled", cryptoutilOpenapiModel.A256GCMDir,
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled))
	err := stack.service.DeleteElasticKey(stack.ctx, &ekID)
	testify.NoError(t, err)
}

func TestDeleteElasticKey_ImportFailed(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "del-impfail", cryptoutilOpenapiModel.A256GCMDir,
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.ImportFailed))
	err := stack.service.DeleteElasticKey(stack.ctx, &ekID)
	testify.NoError(t, err)
}

func TestDeleteElasticKey_PendingImport(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "del-pendimp", cryptoutilOpenapiModel.A256GCMDir,
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport))
	err := stack.service.DeleteElasticKey(stack.ctx, &ekID)
	testify.NoError(t, err)
}

func TestDeleteElasticKey_GenerateFailed(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "del-genfail", cryptoutilOpenapiModel.A256GCMDir,
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.GenerateFailed))
	err := stack.service.DeleteElasticKey(stack.ctx, &ekID)
	testify.NoError(t, err)
}

func TestDeleteElasticKey_InvalidStatus(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "del-creating", cryptoutilOpenapiModel.A256GCMDir,
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Creating))
	err := stack.service.DeleteElasticKey(stack.ctx, &ekID)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "cannot delete ElasticKey in status")
}

func TestDeleteElasticKey_NotFound(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	missingID := googleUuid.New()
	err := stack.service.DeleteElasticKey(stack.ctx, &missingID)
	testify.Error(t, err)
}

func TestRevokeMaterialKey(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "revoke-mk", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	mkID := seedMaterialKey(t, stack, ekID)
	err := stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
	testify.NoError(t, err)
}

func TestRevokeMaterialKey_AlreadyRevoked(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "revoke-dup", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	mkID := seedMaterialKey(t, stack, ekID)
	err := stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
	testify.NoError(t, err)
	err = stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "already revoked")
}

func TestRevokeMaterialKey_NotFound(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := googleUuid.New()
	mkID := googleUuid.New()
	err := stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
	testify.Error(t, err)
}

func TestDeleteMaterialKey_NotImplemented(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := googleUuid.New()
	mkID := googleUuid.New()
	err := stack.service.DeleteMaterialKey(stack.ctx, &ekID, &mkID)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "not implemented")
}

func TestMain(m *testing.M) {
	_ = os.Setenv("CRYPTOUTIL_DATABASE_URL", cryptoutilSharedMagic.SQLiteInMemoryDSN) //nolint:errcheck // TestMain cannot use t.Setenv

	os.Exit(m.Run())
}
