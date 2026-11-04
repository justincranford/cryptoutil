// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"testing"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"

	"github.com/stretchr/testify/require"
)

func TestActualMergedMigrationsRemainInOwnerTransaction(t *testing.T) {
	t.Parallel()

	for _, psid := range []string{psidAuthz, psidIDP, psidRP, psidRS, psidSPA, psidSkeleton, psidKMS} {
		t.Run(psid, func(t *testing.T) {
			t.Parallel()
			sqliteMu.Lock()
			defer sqliteMu.Unlock()

			resetSharedSQLite(t)

			tx, err := sharedSQLite.BeginTx(context.Background(), nil)
			require.NoError(t, err)

			defer func() { _ = tx.Rollback() }()

			files, err := Migrations(psid)
			require.NoError(t, err)
			require.NoError(t, cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(files, "migrations").ApplyTransaction(context.Background(), tx, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite))

			var count int
			require.NoError(t, tx.QueryRow("SELECT count(*) FROM tenants").Scan(&count))
			require.Equal(t, 1, count)
			require.NoError(t, tx.Rollback())
			require.NoError(t, sharedSQLite.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&count))
			require.Zero(t, count, "the migration library must not commit or close the owner's transaction")
		})
	}
}
