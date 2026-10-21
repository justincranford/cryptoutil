// Copyright (c) 2025-2026 Justin Cranford.
//

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

const preflightUnsupportedDatabaseTypeError = "unsupported database type"

func TestValidateCertificateRequestIDUniqueness_CleanData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		createTable bool
	}{
		{name: "table not created yet"},
		{name: "nonconflicting request IDs", createTable: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := newPreflightSQLiteDB(t)
			if !tc.createTable {
				require.NoError(t, ValidateCertificateRequestIDUniqueness(
					context.Background(),
					db,
					cryptoutilSharedMagic.TestDatabaseSQLite,
				))

				return
			}

			createCertificateRequestTable(t, db)

			tenantID, err := googleUuid.NewV7()
			require.NoError(t, err)
			otherTenantID, err := googleUuid.NewV7()
			require.NoError(t, err)
			requestID, err := googleUuid.NewV7()
			require.NoError(t, err)
			otherRequestID, err := googleUuid.NewV7()
			require.NoError(t, err)

			insertCertificateRequest(t, db, newPreflightTestUUID(t), tenantID, requestID)
			insertCertificateRequest(t, db, newPreflightTestUUID(t), tenantID, otherRequestID)
			insertCertificateRequest(t, db, newPreflightTestUUID(t), otherTenantID, requestID)
			insertCertificateRequest(t, db, newPreflightTestUUID(t), tenantID, googleUuid.Nil)
			insertCertificateRequest(t, db, newPreflightTestUUID(t), tenantID, googleUuid.Nil)

			require.NoError(t, ValidateCertificateRequestIDUniqueness(
				context.Background(),
				db,
				cryptoutilSharedMagic.TestDatabaseSQLite,
			))
		})
	}
}

func TestValidateCertificateRequestIDUniqueness_DuplicateRows(t *testing.T) {
	t.Parallel()

	db := newPreflightSQLiteDB(t)
	createCertificateRequestTable(t, db)

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)
	requestID, err := googleUuid.NewV7()
	require.NoError(t, err)
	firstCertificateID, err := googleUuid.NewV7()
	require.NoError(t, err)
	secondCertificateID, err := googleUuid.NewV7()
	require.NoError(t, err)

	insertCertificateRequest(t, db, firstCertificateID, tenantID, requestID)
	insertCertificateRequest(t, db, secondCertificateID, tenantID, requestID)

	err = ValidateCertificateRequestIDUniqueness(context.Background(), db, cryptoutilSharedMagic.TestDatabaseSQLite)
	require.Error(t, err)
	require.ErrorContains(t, err, "migration 2013 cannot create")
	require.ErrorContains(t, err, firstCertificateID.String())
	require.ErrorContains(t, err, secondCertificateID.String())
	require.ErrorContains(t, err, tenantID.String())
	require.ErrorContains(t, err, requestID.String())
	require.ErrorContains(t, err, "no records were modified")

	var rowCount int
	require.NoError(t, db.QueryRowContext(
		context.Background(),
		"SELECT COUNT(*) FROM pki_certificates",
	).Scan(&rowCount))
	require.Equal(t, 2, rowCount)
}

func TestValidateCertificateRequestIDUniqueness_QueryErrors(t *testing.T) {
	t.Parallel()

	db := newPreflightSQLiteDB(t)
	_, err := db.ExecContext(
		context.Background(),
		fmt.Sprintf("CREATE TABLE %s (id TEXT NOT NULL)", cryptoutilSharedMagic.PKICertificatesTable),
	)
	require.NoError(t, err)

	err = ValidateCertificateRequestIDUniqueness(context.Background(), db, cryptoutilSharedMagic.TestDatabaseSQLite)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to inspect duplicate certificate request IDs")
}

func TestValidateCertificateRequestIDUniqueness_InvalidArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ctx          context.Context
		createDB     bool
		databaseType string
		wantErr      string
	}{
		{name: "nil context", createDB: true, databaseType: cryptoutilSharedMagic.TestDatabaseSQLite, wantErr: "context is required"},
		{name: "nil database", ctx: context.Background(), databaseType: cryptoutilSharedMagic.TestDatabaseSQLite, wantErr: "database is required"},
		{name: preflightUnsupportedDatabaseTypeError, ctx: context.Background(), createDB: true, databaseType: "unsupported", wantErr: preflightUnsupportedDatabaseTypeError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var db *sql.DB
			if tc.createDB {
				db = newPreflightSQLiteDB(t)
			}

			err := ValidateCertificateRequestIDUniqueness(tc.ctx, db, tc.databaseType)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func newPreflightSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=private", dbID.String())
	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	return db
}

func createCertificateRequestTable(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		fmt.Sprintf(
			"CREATE TABLE %s (id TEXT NOT NULL, tenant_id TEXT NOT NULL, request_id TEXT NOT NULL)",
			cryptoutilSharedMagic.PKICertificatesTable,
		),
	)
	require.NoError(t, err)
}

func insertCertificateRequest(t *testing.T, db *sql.DB, certificateID, tenantID, requestID googleUuid.UUID) {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		fmt.Sprintf("INSERT INTO %s (id, tenant_id, request_id) VALUES (?, ?, ?)", cryptoutilSharedMagic.PKICertificatesTable),
		certificateID.String(),
		tenantID.String(),
		requestID.String(),
	)
	require.NoError(t, err)
}

func newPreflightTestUUID(t *testing.T) googleUuid.UUID {
	t.Helper()

	id, err := googleUuid.NewV7()
	require.NoError(t, err)

	return id
}
