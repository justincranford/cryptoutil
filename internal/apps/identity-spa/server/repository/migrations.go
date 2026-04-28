// Copyright (c) 2025 Justin Cranford
//
//
// SPDX-License-Identifier: MIT

package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
)

// DatabaseType represents supported database types for identity-spa.
type DatabaseType = cryptoutilAppsFrameworkServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns framework + identity-spa migrations.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
}

// ApplyIdentitySPAMigrations runs framework and identity-spa migrations.
func ApplyIdentitySPAMigrations(db *sql.DB, dbType DatabaseType) error {
	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply identity-spa migrations (1001-1999 + 10001+): %w", err)
	}

	return nil
}
