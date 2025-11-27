// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// validateSchema checks that all required tables exist after migration.
func validateSchema(db *sql.DB, ctx context.Context) error {
	requiredTables := []string{
		"users",
		"clients",
		"client_secret_versions",
		"key_rotation_events",
		"tokens",
		"sessions",
		"authorization_requests",
		"schema_migrations",
	}

	for _, table := range requiredTables {
		var exists int
		query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
		if err := db.QueryRowContext(ctx, query, table).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		if exists == 0 {
			return fmt.Errorf("required table missing: %s", table)
		}
	}

	return nil
}
