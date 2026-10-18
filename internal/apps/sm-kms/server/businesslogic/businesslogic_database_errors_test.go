// Copyright (c) 2025-2026 Justin Cranford.
//

package businesslogic

import (
	"context"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilAppsFrameworkServiceServerMiddleware "cryptoutil/internal/apps-framework/service/server/middleware"
	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm-kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func TestListBusinessLogic_DatabaseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		invoke func(*BusinessLogicService, context.Context, *googleUuid.UUID) error
	}{
		{
			name: "elastic keys query",
			invoke: func(service *BusinessLogicService, ctx context.Context, _ *googleUuid.UUID) error {
				_, err := service.GetElasticKeys(ctx, nil)

				return err
			},
		},
		{
			name: "material keys query",
			invoke: func(service *BusinessLogicService, ctx context.Context, _ *googleUuid.UUID) error {
				_, err := service.GetMaterialKeys(ctx, nil)

				return err
			},
		},
		{
			name: "material keys for elastic key query",
			invoke: func(service *BusinessLogicService, ctx context.Context, elasticKeyID *googleUuid.UUID) error {
				_, err := service.GetMaterialKeysForElasticKey(ctx, elasticKeyID, nil)

				return err
			},
		},
		{
			name: "add elastic key database write",
			invoke: func(service *BusinessLogicService, ctx context.Context, _ *googleUuid.UUID) error {
				description := "database error test"
				_, err := service.AddElasticKey(ctx, &cryptoutilKmsServer.ElasticKeyCreate{
					Algorithm:   string(cryptoutilOpenapiModel.A256GCMDir),
					Description: &description,
					Name:        "database-error",
				})

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
			ctx := context.Background()
			ormRepository, err := cryptoutilOrmRepository.NewOrmRepository(
				ctx,
				testCore.Basic.TelemetryService,
				db,
				testCore.Basic.JWKGenService,
				false,
			)
			testify.NoError(t, err)

			defer ormRepository.Shutdown()

			service, err := NewBusinessLogicService(ctx, testCore.Basic.TelemetryService, testCore.Basic.JWKGenService, ormRepository, testBarrierService)
			testify.NoError(t, err)

			tenantID, err := googleUuid.NewV7()
			testify.NoError(t, err)

			realmCtx := &cryptoutilAppsFrameworkServiceServerMiddleware.RealmContext{TenantID: tenantID}
			requestCtx := context.WithValue(ctx, cryptoutilAppsFrameworkServiceServerMiddleware.RealmContextKey{}, realmCtx)
			elasticKeyID, err := googleUuid.NewV7()
			testify.NoError(t, err)

			err = tc.invoke(service, requestCtx, &elasticKeyID)
			testify.Error(t, err)
			testify.NotContains(t, err.Error(), "%!w(<nil>)")
		})
	}
}
