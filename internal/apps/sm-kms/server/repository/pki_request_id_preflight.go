// Copyright (c) 2025-2026 Justin Cranford.
//

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
)

const (
	certificateTablePlaceholder         = "{{certificate_table}}"
	requestIDParameterPlaceholder       = "{{request_id_parameter}}"
	duplicateCertificateRequestIDsQuery = `
SELECT certificate.id, certificate.tenant_id, certificate.request_id
FROM {{certificate_table}} AS certificate
INNER JOIN (
	SELECT tenant_id, request_id
	FROM {{certificate_table}}
	WHERE request_id <> {{request_id_parameter}}
	GROUP BY tenant_id, request_id
	HAVING COUNT(*) > 1
) AS duplicates
	ON certificate.tenant_id = duplicates.tenant_id
	AND certificate.request_id = duplicates.request_id
ORDER BY certificate.tenant_id, certificate.request_id, certificate.id`
)

// ValidateCertificateRequestIDUniqueness reports duplicate nonzero certificate
// request IDs before migration 2013 attempts to enforce their unique index.
func ValidateCertificateRequestIDUniqueness(ctx context.Context, db *sql.DB, databaseType string) error {
	if ctx == nil {
		return fmt.Errorf("context is required for certificate request ID preflight")
	}

	if db == nil {
		return fmt.Errorf("database is required for certificate request ID preflight")
	}

	tableExists, err := certificateTableExists(ctx, db, databaseType)
	if err != nil {
		return fmt.Errorf("failed to check for certificate table before migration 2013: %w", err)
	}

	if !tableExists {
		return nil
	}

	var requestIDPlaceholder string

	switch databaseType {
	case cryptoutilSharedMagic.TestDatabaseSQLite:
		requestIDPlaceholder = "?"
	case cryptoutilSharedMagic.DockerServicePostgres:
		requestIDPlaceholder = "$1"
	default:
		return fmt.Errorf("unsupported database type %q", databaseType)
	}

	query := strings.NewReplacer(
		certificateTablePlaceholder,
		cryptoutilSharedMagic.PKICertificatesTable,
		requestIDParameterPlaceholder,
		requestIDPlaceholder,
	).Replace(duplicateCertificateRequestIDsQuery)

	rows, err := db.QueryContext(ctx, query, googleUuid.Nil.String())
	if err != nil {
		return fmt.Errorf("failed to inspect duplicate certificate request IDs before migration 2013: %w", err)
	}

	conflictingRows, err := collectDuplicateCertificateRequestIDs(rows)
	if err != nil {
		return fmt.Errorf("failed to read duplicate certificate request IDs before migration 2013: %w", err)
	}

	if len(conflictingRows) == 0 {
		return nil
	}

	return fmt.Errorf(
		"migration 2013 cannot create its unique certificate request ID index: duplicate nonzero (tenant_id, request_id) rows found: %s; inspect and manually reconcile the listed legacy records before retrying; no records were modified",
		strings.Join(conflictingRows, "; "),
	)
}

func collectDuplicateCertificateRequestIDs(rows *sql.Rows) (conflictingRows []string, resultErr error) {
	defer func() {
		if err := rows.Close(); err != nil {
			resultErr = errors.Join(
				resultErr,
				fmt.Errorf("failed to close duplicate certificate request ID rows: %w", err),
			)
		}
	}()

	for rows.Next() {
		var certificateID, tenantID, requestID string
		if err := rows.Scan(&certificateID, &tenantID, &requestID); err != nil {
			return nil, fmt.Errorf("failed to scan duplicate certificate request ID row: %w", err)
		}

		conflictingRows = append(conflictingRows, fmt.Sprintf(
			"certificate_id=%s tenant_id=%s request_id=%s",
			certificateID,
			tenantID,
			requestID,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating duplicate certificate request ID rows: %w", err)
	}

	return conflictingRows, nil
}

func certificateTableExists(ctx context.Context, db *sql.DB, databaseType string) (bool, error) {
	var exists bool

	switch databaseType {
	case cryptoutilSharedMagic.TestDatabaseSQLite:
		err := db.QueryRowContext(
			ctx,
			"SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)",
			cryptoutilSharedMagic.PKICertificatesTable,
		).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("query SQLite certificate table existence: %w", err)
		}

		return exists, nil
	case cryptoutilSharedMagic.DockerServicePostgres:
		var tableName sql.NullString

		err := db.QueryRowContext(
			ctx,
			"SELECT to_regclass($1)",
			cryptoutilSharedMagic.PKICertificatesTable,
		).Scan(&tableName)
		if err != nil {
			return false, fmt.Errorf("query PostgreSQL certificate table existence: %w", err)
		}

		return tableName.Valid, nil
	default:
		return false, fmt.Errorf("unsupported database type %q", databaseType)
	}
}
