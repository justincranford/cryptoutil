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

const sqliteTableColumnCountQuery = "SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?"

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

func TestPKIMigrations_Rollback(t *testing.T) {
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

	runMigration := func(name string) {
		t.Helper()

		migration, readErr := MigrationsFS.ReadFile(fmt.Sprintf("migrations/%s", name))
		require.NoError(t, readErr)

		_, execErr := db.ExecContext(ctx, string(migration))
		require.NoError(t, execErr)
	}
	assertSQLiteObjectCount := func(objectType, objectName string, want int) {
		t.Helper()

		var count int

		query := "SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?"
		queryErr := db.QueryRowContext(ctx, query, objectType, objectName).Scan(&count)
		require.NoError(t, queryErr)
		require.Equal(t, want, count)
	}
	assertIssuerAuthorityColumnCount := func(want int) {
		t.Helper()

		var count int

		queryErr := db.QueryRowContext(ctx, sqliteTableColumnCountQuery, cryptoutilSharedMagic.PKICertificatesTable, "issuer_authority_id").Scan(&count)
		require.NoError(t, queryErr)
		require.Equal(t, want, count)
	}
	assertAuthorityEnvelopeVersionColumnCount := func(want int) {
		t.Helper()

		var count int

		queryErr := db.QueryRowContext(ctx, sqliteTableColumnCountQuery, cryptoutilSharedMagic.PKIAuthoritiesTable, "private_key_envelope_version").Scan(&count)
		require.NoError(t, queryErr)
		require.Equal(t, want, count)
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE %s (tenant_id TEXT NOT NULL, request_id TEXT NOT NULL);
CREATE INDEX %s ON %s (tenant_id, request_id);`,
		cryptoutilSharedMagic.PKICertificatesTable,
		cryptoutilSharedMagic.PKICertificateRequestLegacyIndex,
		cryptoutilSharedMagic.PKICertificatesTable,
	))
	require.NoError(t, err)

	authorityID, err := googleUuid.NewV7()
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE %s (
	id TEXT NOT NULL PRIMARY KEY,
	certificate_der BLOB NOT NULL,
	private_key_encrypted BLOB NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO %s (id, certificate_der, private_key_encrypted) VALUES (?, ?, ?);`,
		cryptoutilSharedMagic.PKIAuthoritiesTable,
		cryptoutilSharedMagic.PKIAuthoritiesTable,
	), authorityID.String(), []byte("certificate"), []byte("encrypted-key"))
	require.NoError(t, err)

	runMigration("2012_pki_crl_sequences.up.sql")
	assertSQLiteObjectCount("table", cryptoutilSharedMagic.PKICRLSequencesTable, 1)

	runMigration("2013_pki_certificate_request_unique.up.sql")
	assertSQLiteObjectCount("index", cryptoutilSharedMagic.PKICertificateRequestLegacyIndex, 0)
	assertSQLiteObjectCount("index", cryptoutilSharedMagic.PKICertificateRequestIndex, 1)

	runMigration("2014_pki_certificate_issuer_authority.up.sql")
	assertIssuerAuthorityColumnCount(1)

	runMigration("2015_pki_authority_key_envelope_version.up.sql")
	assertAuthorityEnvelopeVersionColumnCount(1)

	var envelopeVersion int

	err = db.QueryRowContext(ctx,
		"SELECT private_key_envelope_version FROM pki_authorities WHERE id = ?",
		authorityID.String(),
	).Scan(&envelopeVersion)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1, envelopeVersion)

	runMigration("2015_pki_authority_key_envelope_version.down.sql")
	assertAuthorityEnvelopeVersionColumnCount(0)

	var authorityCount int

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pki_authorities WHERE id = ?", authorityID.String()).Scan(&authorityCount)
	require.NoError(t, err)
	require.Equal(t, 1, authorityCount)

	runMigration("2014_pki_certificate_issuer_authority.down.sql")
	assertIssuerAuthorityColumnCount(0)

	runMigration("2013_pki_certificate_request_unique.down.sql")
	assertSQLiteObjectCount("index", cryptoutilSharedMagic.PKICertificateRequestIndex, 0)
	assertSQLiteObjectCount("index", cryptoutilSharedMagic.PKICertificateRequestLegacyIndex, 1)

	runMigration("2012_pki_crl_sequences.down.sql")
	assertSQLiteObjectCount("table", cryptoutilSharedMagic.PKICRLSequencesTable, 0)
}
