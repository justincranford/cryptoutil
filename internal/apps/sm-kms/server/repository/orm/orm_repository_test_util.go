// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"testing"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"gorm.io/gorm"
)

// RequireNewForTest creates a new ORM repository from GORM for testing and panics on error.
func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, gormDB *gorm.DB, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, verboseMode bool) *OrmRepository {
	ormRepository, err := NewOrmRepository(ctx, telemetryService, gormDB, jwkGenService, verboseMode)
	cryptoutilSharedApperr.RequireNoError(err, "failed to create new ORM repository")

	return ormRepository
}

// KMSCleanupTables is the ordered list of sm-kms tables for test cleanup (reverse FK order).
var KMSCleanupTables = []string{
	"material_keys",
	"elastic_keys",
	"barrier_content_keys",
	"barrier_intermediate_keys",
	"barrier_root_keys",
}

// CleanupDatabase removes all data from the specified tables to ensure test isolation.
// Tables are deleted in the order provided (caller must provide reverse FK order).
// Should be called via t.Cleanup() at the start of each test that modifies database state.
func CleanupDatabase(t *testing.T, repo *OrmRepository, tables []string) {
	t.Helper()
	t.Cleanup(func() {
		err := repo.WithTransaction(context.Background(), ReadWrite, func(tx *OrmTransaction) error {
			for _, table := range tables {
				if err := tx.state.gormTx.Exec("DELETE FROM " + table).Error; err != nil { //nolint:gosec // Table names are internal constants, not user input
					return err
				}
			}

			return nil
		})
		cryptoutilSharedApperr.RequireNoError(err, "failed to cleanup database tables")
	})
}
