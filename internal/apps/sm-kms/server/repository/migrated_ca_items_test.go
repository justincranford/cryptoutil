package repository

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigratedCAItemsTable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(
		cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS),
		"migrations",
	)
	err = runner.Apply(db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)

	var tableName string

	err = db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
		cryptoutilSharedMagic.CA_ITEMS,
	).Scan(&tableName)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.CA_ITEMS, tableName)
}
