// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"testing"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"

	"gorm.io/gorm"
)

// RequireNewForTest creates a new ORM repository from GORM for testing and panics on error.
func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, gormDB *gorm.DB, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, verboseMode bool) *OrmRepository {
	ormRepository, err := NewOrmRepository(ctx, telemetryService, gormDB, jwkGenService, verboseMode)
	cryptoutilSharedApperr.RequireNoError(err, "failed to create new ORM repository")

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
		cryptoutilSharedApperr.RequireNoError(err, "failed to cleanup database tables")
	})
}
