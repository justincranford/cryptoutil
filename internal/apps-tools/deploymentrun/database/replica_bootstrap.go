// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"database/sql"
	json "encoding/json"
	"errors"
	"strings"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"

	"github.com/jackc/pgx/v5"
)

// PrepareReplica uses actual merged migrations in a caller-owned transaction.
// Only a schema observed empty at transaction start may have its freshly
// generated seed/metadata rows cleared. Existing schemas are verify-only.
func (client *SQLClient) PrepareReplica(ctx context.Context, psid, expected string) error {
	if ctx == nil || client.DB == nil || client.backend != Postgres {
		return errors.New("PostgreSQL replica context required")
	}

	files, err := Migrations(psid)
	if err != nil {
		return err
	}

	return client.prepareReplicaWith(ctx, psid, expected, func(ctx context.Context, tx *sql.Tx, _ string) error {
		return cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(files, "migrations").ApplyTransaction(ctx, tx, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL)
	})
}

func replicaSchemaShape(value string) (string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &fields); err != nil || fields["columns"] == nil || fields["constraints"] == nil {
		return "", errors.New("replica schema description rejected")
	}

	delete(fields, "migration")

	raw, err := json.Marshal(fields)
	if err != nil {
		return "", errors.New("replica schema shape rejected")
	}

	return string(raw), nil
}

func (client *SQLClient) prepareReplicaWith(ctx context.Context, psid, expected string, apply func(context.Context, *sql.Tx, string) error) (result error) {
	expectedShape, err := replicaSchemaShape(expected)
	if err != nil {
		return err
	}

	tx, err := client.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.New("replica bootstrap transaction rejected")
	}

	defer func() { _ = tx.Rollback() }()

	var count int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM pg_tables WHERE schemaname='public'").Scan(&count); err != nil {
		return errors.New("replica initial schema inspection rejected")
	}

	if count == 0 {
		if err := apply(ctx, tx, psid); err != nil {
			return errors.New("transactional replica migrations rejected")
		}

		actual, err := replicaSchema(ctx, tx, true)
		if err != nil || actual != expected {
			return errors.New("fresh replica migrated schema mismatch")
		}

		rows, err := tx.QueryContext(ctx, "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename LIMIT 1025")
		if err != nil {
			return errors.New("fresh replica table inventory rejected")
		}

		names := []string{}

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				_ = rows.Close()

				return errors.New("fresh replica table identity rejected")
			}

			names = append(names, pgx.Identifier{replicaPublicSchema, name}.Sanitize())
		}

		rowErr := rows.Err()

		closeErr := rows.Close()
		if rowErr != nil || closeErr != nil || len(names) == 0 || len(names) > maximumReplicaTables {
			return errors.New("fresh replica table inventory incomplete")
		}

		if _, err := tx.ExecContext(ctx, "TRUNCATE TABLE "+strings.Join(names, ",")+" RESTART IDENTITY"); err != nil {
			return errors.New("fresh replica seed clearing rejected")
		}
	} else {
		actual, err := replicaSchema(ctx, tx, false)
		if err != nil {
			return err
		}

		actualShape, err := replicaSchemaShape(actual)
		if err != nil || actualShape != expectedShape {
			return errors.New("existing replica schema collision")
		}

		if actual != expected {
			if err := requireEmptyReplica(ctx, tx); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.New("replica bootstrap commit rejected")
	}

	return contextFailure(ctx)
}
