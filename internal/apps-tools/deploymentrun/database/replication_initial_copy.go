// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

const (
	replicaPublicSchema  = cryptoutilSharedMagic.PublicLabel
	maximumReplicaTables = 1024
)

// Initial copy inserts source rows; it does not upsert migration metadata or
// seeded application rows. Never start that copy over populated follower data.
func (client *SQLClient) requireEmptyReplica(ctx context.Context) error {
	return requireEmptyReplica(ctx, client.DB)
}

type replicaQueries interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func requireEmptyReplica(ctx context.Context, query replicaQueries) error {
	rows, err := query.QueryContext(ctx, "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename LIMIT 1025")
	if err != nil {
		return errors.New("replica initial-data safety lookup rejected")
	}
	defer func() { _ = rows.Close() }()

	tables := []string{}

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return errors.New("replica table identity rejected")
		}

		tables = append(tables, name)
	}

	if err := rows.Err(); err != nil {
		return errors.New("replica table enumeration rejected")
	}

	if err := rows.Close(); err != nil {
		return errors.New("replica table enumeration closure rejected")
	}

	if len(tables) > maximumReplicaTables {
		return errors.New("replica table inventory exceeds bound")
	}

	for _, table := range tables {
		var populated bool
		if err := query.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM "+pgx.Identifier{replicaPublicSchema, table}.Sanitize()+")").Scan(&populated); err != nil {
			return errors.New("replica initial-data safety read rejected")
		}

		if populated {
			return errors.New("initial replication copy requires an empty schema-only follower; embedded migrations populate metadata and seeds")
		}
	}

	return nil
}
