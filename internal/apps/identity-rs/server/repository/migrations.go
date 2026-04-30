// Copyright (c) 2025-2026 Justin Cranford.
//
// SPDX-License-Identifier: AGPL-3.0-only
package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
)

// DatabaseType represents supported database types for identity-rs.
type DatabaseType = cryptoutilAppsFrameworkServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns framework + identity-rs migrations.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
}

// ApplyIdentityRSMigrations runs framework and identity-rs migrations.
func ApplyIdentityRSMigrations(db *sql.DB, dbType DatabaseType) error {
	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply identity-rs migrations (1001-1999 + 8001+): %w", err)
	}

	return nil
}
