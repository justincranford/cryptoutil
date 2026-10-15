//go:build e2e

// Copyright (c) 2025-2026 Justin Cranford.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestPKIMigrations_PostgreSQLUpgradeAndRollback(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	postgresContainer, err := cryptoutilSharedContainer.NewPostgresTestContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, postgresContainer.Terminate(context.Background()))
	})

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	require.NoError(t, db.PingContext(ctx))

	_, err = db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE %s (id TEXT NOT NULL PRIMARY KEY, tenant_id TEXT NOT NULL, request_id TEXT NOT NULL);
CREATE INDEX %s ON %s (tenant_id, request_id);
CREATE TABLE %s (
	id TEXT NOT NULL PRIMARY KEY,
	certificate_der BYTEA NOT NULL,
	private_key_encrypted BYTEA NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,
		cryptoutilSharedMagic.PKICertificatesTable,
		cryptoutilSharedMagic.PKICertificateRequestLegacyIndex,
		cryptoutilSharedMagic.PKICertificatesTable,
		cryptoutilSharedMagic.PKIAuthoritiesTable,
	))
	require.NoError(t, err)

	authorityID, err := googleUuid.NewV7()
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO pki_authorities (id, certificate_der, private_key_encrypted) VALUES ($1, $2, $3)",
		authorityID.String(),
		[]byte("certificate"),
		[]byte("encrypted-key"),
	)
	require.NoError(t, err)

	executePostgreSQLPKIMigration(t, ctx, db, "2012_pki_crl_sequences.up.sql")
	require.Equal(t, 1, postgreSQLObjectCount(t, ctx, db, "table", cryptoutilSharedMagic.PKICRLSequencesTable))

	executePostgreSQLPKIMigration(t, ctx, db, "2013_pki_certificate_request_unique.up.sql")
	require.Equal(t, 0, postgreSQLObjectCount(t, ctx, db, "index", cryptoutilSharedMagic.PKICertificateRequestLegacyIndex))
	require.Equal(t, 1, postgreSQLObjectCount(t, ctx, db, "index", cryptoutilSharedMagic.PKICertificateRequestIndex))

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)
	requestID, err := googleUuid.NewV7()
	require.NoError(t, err)
	firstCertificateID, err := googleUuid.NewV7()
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO pki_certificates (id, tenant_id, request_id) VALUES ($1, $2, $3)",
		firstCertificateID.String(),
		tenantID.String(),
		requestID.String(),
	)
	require.NoError(t, err)

	duplicateCertificateID, err := googleUuid.NewV7()
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO pki_certificates (id, tenant_id, request_id) VALUES ($1, $2, $3)",
		duplicateCertificateID.String(),
		tenantID.String(),
		requestID.String(),
	)
	require.Error(t, err)

	for range 2 {
		certificateID, idErr := googleUuid.NewV7()
		require.NoError(t, idErr)

		_, insertErr := db.ExecContext(ctx,
			"INSERT INTO pki_certificates (id, tenant_id, request_id) VALUES ($1, $2, $3)",
			certificateID.String(),
			tenantID.String(),
			googleUuid.Nil.String(),
		)
		require.NoError(t, insertErr)
	}

	executePostgreSQLPKIMigration(t, ctx, db, "2014_pki_certificate_issuer_authority.up.sql")
	require.Equal(t, 1, postgreSQLColumnCount(t, ctx, db, cryptoutilSharedMagic.PKICertificatesTable, "issuer_authority_id"))

	var issuerAuthorityID string

	err = db.QueryRowContext(ctx,
		"SELECT issuer_authority_id FROM pki_certificates WHERE id = $1",
		firstCertificateID.String(),
	).Scan(&issuerAuthorityID)
	require.NoError(t, err)
	require.Empty(t, issuerAuthorityID)

	executePostgreSQLPKIMigration(t, ctx, db, "2015_pki_authority_key_envelope_version.up.sql")
	require.Equal(t, 1, postgreSQLColumnCount(t, ctx, db, cryptoutilSharedMagic.PKIAuthoritiesTable, "private_key_envelope_version"))

	var envelopeVersion int

	err = db.QueryRowContext(ctx,
		"SELECT private_key_envelope_version FROM pki_authorities WHERE id = $1",
		authorityID.String(),
	).Scan(&envelopeVersion)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1, envelopeVersion)

	executePostgreSQLPKIMigration(t, ctx, db, "2015_pki_authority_key_envelope_version.down.sql")
	require.Equal(t, 0, postgreSQLColumnCount(t, ctx, db, cryptoutilSharedMagic.PKIAuthoritiesTable, "private_key_envelope_version"))

	var authorityCount int

	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pki_authorities WHERE id = $1",
		authorityID.String(),
	).Scan(&authorityCount)
	require.NoError(t, err)
	require.Equal(t, 1, authorityCount)

	executePostgreSQLPKIMigration(t, ctx, db, "2014_pki_certificate_issuer_authority.down.sql")
	require.Equal(t, 0, postgreSQLColumnCount(t, ctx, db, cryptoutilSharedMagic.PKICertificatesTable, "issuer_authority_id"))

	executePostgreSQLPKIMigration(t, ctx, db, "2013_pki_certificate_request_unique.down.sql")
	require.Equal(t, 0, postgreSQLObjectCount(t, ctx, db, "index", cryptoutilSharedMagic.PKICertificateRequestIndex))
	require.Equal(t, 1, postgreSQLObjectCount(t, ctx, db, "index", cryptoutilSharedMagic.PKICertificateRequestLegacyIndex))

	duplicateAfterRollbackID, err := googleUuid.NewV7()
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO pki_certificates (id, tenant_id, request_id) VALUES ($1, $2, $3)",
		duplicateAfterRollbackID.String(),
		tenantID.String(),
		requestID.String(),
	)
	require.NoError(t, err)

	executePostgreSQLPKIMigration(t, ctx, db, "2012_pki_crl_sequences.down.sql")
	require.Equal(t, 0, postgreSQLObjectCount(t, ctx, db, "table", cryptoutilSharedMagic.PKICRLSequencesTable))
}

func executePostgreSQLPKIMigration(t *testing.T, ctx context.Context, db *sql.DB, name string) {
	t.Helper()

	migration, err := MigrationsFS.ReadFile(fmt.Sprintf("migrations/%s", name))
	require.NoError(t, err)

	sqlLines := make([]string, 0)

	for _, line := range strings.Split(string(migration), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			sqlLines = append(sqlLines, line)
		}
	}

	for _, statement := range strings.Split(strings.Join(sqlLines, "\n"), ";") {
		query := strings.TrimSpace(statement)
		if query == "" {
			continue
		}

		_, err = db.ExecContext(ctx, query)
		require.NoErrorf(t, err, "migration %s statement: %s", name, query)
	}
}

func postgreSQLObjectCount(t *testing.T, ctx context.Context, db *sql.DB, objectType, objectName string) int {
	t.Helper()

	var count int

	switch objectType {
	case "table":
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1",
			objectName,
		).Scan(&count)
		require.NoError(t, err)
	case "index":
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM pg_indexes WHERE schemaname = current_schema() AND indexname = $1",
			objectName,
		).Scan(&count)
		require.NoError(t, err)
	default:
		require.FailNowf(t, "unsupported PostgreSQL object type", "object type %q", objectType)
	}

	return count
}

func postgreSQLColumnCount(t *testing.T, ctx context.Context, db *sql.DB, tableName, columnName string) int {
	t.Helper()

	var count int

	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2",
		tableName,
		columnName,
	).Scan(&count)
	require.NoError(t, err)

	return count
}
