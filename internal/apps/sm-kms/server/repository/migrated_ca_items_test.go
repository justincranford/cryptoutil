package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigratedPKITables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	databaseID, err := googleUuid.NewV7()
	require.NoError(t, err)
	db, err := sql.Open(
		cryptoutilSharedMagic.TestDatabaseSQLite,
		fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(
		cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS),
		"migrations",
	)
	err = runner.Apply(db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)

	for _, expectedTable := range []string{
		cryptoutilSharedMagic.CA_ITEMS,
		cryptoutilSharedMagic.PKICertificatesTable,
		cryptoutilSharedMagic.PKIAuthoritiesTable,
	} {
		var tableName string

		err = db.QueryRowContext(ctx,
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			expectedTable,
		).Scan(&tableName)
		require.NoError(t, err)
		require.Equal(t, expectedTable, tableName)
	}
}

func TestRequestUniqueMigration_DuplicateRowsPreserveExistingIndex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE %s (tenant_id TEXT NOT NULL, request_id TEXT NOT NULL);
CREATE INDEX %s ON %s (tenant_id, request_id);`,
		cryptoutilSharedMagic.PKICertificatesTable,
		cryptoutilSharedMagic.PKICertificateRequestLegacyIndex,
		cryptoutilSharedMagic.PKICertificatesTable,
	))
	require.NoError(t, err)

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)
	requestID, err := googleUuid.NewV7()
	require.NoError(t, err)

	for range 2 {
		_, err = db.ExecContext(ctx,
			"INSERT INTO pki_certificates (tenant_id, request_id) VALUES (?, ?)",
			tenantID.String(),
			requestID.String(),
		)
		require.NoError(t, err)
	}

	migration, err := MigrationsFS.ReadFile("migrations/2013_pki_certificate_request_unique.up.sql")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.Error(t, err)

	var indexCount int

	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?",
		cryptoutilSharedMagic.PKICertificateRequestLegacyIndex,
	).Scan(&indexCount)
	require.NoError(t, err)
	require.Equal(t, 1, indexCount)
}
