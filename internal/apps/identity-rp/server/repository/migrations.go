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

// DatabaseType represents supported database types for identity-rp.
type DatabaseType = cryptoutilAppsFrameworkServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns framework + identity-rp migrations.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
}

// ApplyIdentityRPMigrations runs framework and identity-rp migrations.
func ApplyIdentityRPMigrations(db *sql.DB, dbType DatabaseType) error {
	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply identity-rp migrations (1001-1999 + 9001+): %w", err)
	}

	return nil
}
