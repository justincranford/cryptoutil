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

// DatabaseType represents supported database types for identity-idp.
type DatabaseType = cryptoutilAppsFrameworkServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns framework + identity-idp migrations.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
}

// ApplyIdentityIDPMigrations runs framework and identity-idp migrations.
func ApplyIdentityIDPMigrations(db *sql.DB, dbType DatabaseType) error {
	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply identity-idp migrations (1001-1999 + 7001+): %w", err)
	}

	return nil
}
