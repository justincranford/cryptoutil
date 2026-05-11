package businesslogic

import (
	"context"
	"testing"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilAppsFrameworkServiceServerMiddleware "cryptoutil/internal/apps-framework/service/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm-kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type testStack struct {
	service *BusinessLogicService
	ctx     context.Context
	core    *testCoreFixture
}

func setupTestStack(tb testing.TB) *testStack {
	tb.Helper()

	ctx := context.Background()

	// Use shared fixtures from TestMain.
	if testCore == nil || testBarrierService == nil {
		tb.Fatalf("test fixtures not initialized - TestMain may have failed")
	}

	// Cleanup function to reset test data
	// Note: This truncates tables between tests to ensure isolation
	// while reusing the shared DB connection.
	tb.Cleanup(func() {
		// Truncate tables for clean slate in next test
		if err := testCore.DB.Exec("DELETE FROM elastic_keys").Error; err != nil && err.Error() != "table elastic_keys does not exist" {
			tb.Logf("cleanup error deleting elastic_keys: %v", err)
		}

		if err := testCore.DB.Exec("DELETE FROM material_keys").Error; err != nil && err.Error() != "table material_keys does not exist" {
			tb.Logf("cleanup error deleting material_keys: %v", err)
		}
	})

	// Create per-test repository and service for isolation
	ormRepo, err := cryptoutilOrmRepository.NewOrmRepository(
		ctx, testCore.Basic.TelemetryService, testCore.DB,
		testCore.Basic.JWKGenService, false,
	)
	testify.NoError(tb, err)
	tb.Cleanup(func() { ormRepo.Shutdown() })

	service, err := NewBusinessLogicService(
		ctx, testCore.Basic.TelemetryService, testCore.Basic.JWKGenService,
		ormRepo, testBarrierService,
	)
	testify.NoError(tb, err)

	tenantID := googleUuid.New()
	rc := &cryptoutilAppsFrameworkServiceServerMiddleware.RealmContext{TenantID: tenantID}
	testCtx := context.WithValue(ctx, cryptoutilAppsFrameworkServiceServerMiddleware.RealmContextKey{}, rc)

	return &testStack{service: service, ctx: testCtx, core: testCore}
}

func seedElasticKey(t *testing.T, stack *testStack, name string, alg cryptoutilOpenapiModel.ElasticKeyAlgorithm, status cryptoutilKmsServer.ElasticKeyStatus) googleUuid.UUID {
	t.Helper()

	tenantID := cryptoutilAppsFrameworkServiceServerMiddleware.GetRealmContext(stack.ctx).TenantID
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

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestGetElasticKeyByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		stack := setupTestStack(t)
		ekID := seedElasticKey(t, stack, "get-by-id", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
		ek, err := stack.service.GetElasticKeyByElasticKeyID(stack.ctx, &ekID)
		testify.NoError(t, err)
		testify.NotNil(t, ek)
		testify.Equal(t, "get-by-id", *ek.Name)
	})

	t.Run("not found", func(t *testing.T) {
		stack := setupTestStack(t)
		missingID := googleUuid.New()
		_, err := stack.service.GetElasticKeyByElasticKeyID(stack.ctx, &missingID)
		testify.Error(t, err)
	})
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestGetElasticKeys(t *testing.T) {
	stack := setupTestStack(t)
	seedElasticKey(t, stack, "list-a", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedElasticKey(t, stack, "list-b", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	all, err := stack.service.GetElasticKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.Len(t, all, 2)
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestGetMaterialKeysForElasticKey(t *testing.T) {
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "mat-for-ek", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID)
	seedMaterialKey(t, stack, ekID)
	mks, err := stack.service.GetMaterialKeysForElasticKey(stack.ctx, &ekID, nil)
	testify.NoError(t, err)
	testify.Len(t, mks, 2)
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestGetMaterialKeys(t *testing.T) {
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "mat-all", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID)
	mks, err := stack.service.GetMaterialKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.GreaterOrEqual(t, len(mks), 1)
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestGetMaterialKeyByIDs(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		stack := setupTestStack(t)
		ekID := seedElasticKey(t, stack, "mat-by-ids", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
		mkID := seedMaterialKey(t, stack, ekID)
		mk, err := stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(stack.ctx, &ekID, &mkID)
		testify.NoError(t, err)
		testify.NotNil(t, mk)
	})

	t.Run("not found", func(t *testing.T) {
		stack := setupTestStack(t)
		ekID := seedElasticKey(t, stack, "mat-nf", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
		missingMK := googleUuid.New()
		_, err := stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(stack.ctx, &ekID, &missingMK)
		testify.Error(t, err)
	})
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestUpdateElasticKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stack := setupTestStack(t)
		ekID := seedElasticKey(t, stack, "update-me", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
		newDesc := "updated-desc"
		updated, err := stack.service.UpdateElasticKey(stack.ctx, &ekID, &cryptoutilKmsServer.ElasticKeyUpdate{
			Name:        "updated-name",
			Description: &newDesc,
		})
		testify.NoError(t, err)
		testify.Equal(t, "updated-name", *updated.Name)
	})

	t.Run("not found", func(t *testing.T) {
		stack := setupTestStack(t)
		missingID := googleUuid.New()
		newDesc := "desc"
		_, err := stack.service.UpdateElasticKey(stack.ctx, &missingID, &cryptoutilKmsServer.ElasticKeyUpdate{
			Name:        "x",
			Description: &newDesc,
		})
		testify.Error(t, err)
	})
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestDeleteElasticKey(t *testing.T) {
	tests := []struct {
		name    string
		status  cryptoutilKmsServer.ElasticKeyStatus
		wantErr string
	}{
		{name: "active", status: cryptoutilKmsServer.Active},
		{name: "status disabled", status: cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled)},
		{name: "import failed", status: cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.ImportFailed)},
		{name: "pending import", status: cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)},
		{name: "generate failed", status: cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.GenerateFailed)},
		{name: "invalid status creating", status: cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Creating), wantErr: "cannot delete ElasticKey in status"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stack := setupTestStack(t)
			ekID := seedElasticKey(t, stack, "del-"+tc.name, cryptoutilOpenapiModel.A256GCMDir, tc.status)

			err := stack.service.DeleteElasticKey(stack.ctx, &ekID)
			if tc.wantErr != "" {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.wantErr)

				return
			}

			testify.NoError(t, err)
		})
	}

	t.Run("not found", func(t *testing.T) {
		stack := setupTestStack(t)
		missingID := googleUuid.New()
		err := stack.service.DeleteElasticKey(stack.ctx, &missingID)
		testify.Error(t, err)
	})
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestRevokeMaterialKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stack := setupTestStack(t)
		ekID := seedElasticKey(t, stack, "revoke-mk", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
		mkID := seedMaterialKey(t, stack, ekID)
		err := stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
		testify.NoError(t, err)
	})

	t.Run("already revoked", func(t *testing.T) {
		stack := setupTestStack(t)
		ekID := seedElasticKey(t, stack, "revoke-dup", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
		mkID := seedMaterialKey(t, stack, ekID)
		err := stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
		testify.NoError(t, err)
		err = stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
		testify.Error(t, err)
		testify.Contains(t, err.Error(), "already revoked")
	})

	t.Run("not found", func(t *testing.T) {
		stack := setupTestStack(t)
		ekID := googleUuid.New()
		mkID := googleUuid.New()
		err := stack.service.RevokeMaterialKey(stack.ctx, &ekID, &mkID)
		testify.Error(t, err)
	})
}

// Sequential: tests mutate shared testCore database with WAL write locks; parallel execution causes SQLite lock contention.
func TestDeleteMaterialKey_NotImplemented(t *testing.T) {
	stack := setupTestStack(t)
	ekID := googleUuid.New()
	mkID := googleUuid.New()
	err := stack.service.DeleteMaterialKey(stack.ctx, &ekID, &mkID)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "not implemented")
}
