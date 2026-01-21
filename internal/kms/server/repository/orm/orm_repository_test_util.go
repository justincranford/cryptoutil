// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
)

// RequireNewForTest creates a new ORM repository for testing and panics on error.
func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSQLRepository.SQLRepository, jwkGenService *cryptoutilJose.JWKGenService, settings *cryptoutilConfig.ServiceTemplateServerSettings) *OrmRepository {
	ormRepository, err := NewOrmRepository(ctx, telemetryService, sqlRepository, jwkGenService, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to create new ORM repository")

	return ormRepository
}

// CleanupDatabase removes all data from database tables to ensure test isolation.
// Should be called via t.Cleanup() at the start of each test that modifies database state.
func CleanupDatabase(t *testing.T, repo *OrmRepository) {
	t.Helper()
	t.Cleanup(func() {
		err := repo.WithTransaction(context.Background(), ReadWrite, func(tx *OrmTransaction) error {
			// Delete in reverse foreign key dependency order.
			if err := tx.state.gormTx.Exec("DELETE FROM material_keys").Error; err != nil {
				return err
			}

			if err := tx.state.gormTx.Exec("DELETE FROM elastic_keys").Error; err != nil {
				return err
			}

			if err := tx.state.gormTx.Exec("DELETE FROM barrier_content_keys").Error; err != nil {
				return err
			}

			if err := tx.state.gormTx.Exec("DELETE FROM barrier_intermediate_keys").Error; err != nil {
				return err
			}

			if err := tx.state.gormTx.Exec("DELETE FROM barrier_root_keys").Error; err != nil {
				return err
			}

			return nil
		})
		cryptoutilAppErr.RequireNoError(err, "failed to cleanup database tables")
	})
}
