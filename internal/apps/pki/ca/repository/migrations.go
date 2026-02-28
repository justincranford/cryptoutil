// Copyright (c) 2025 Justin Cranford

package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// DatabaseType represents supported database types for pki-ca.
type DatabaseType = cryptoutilAppsTemplateServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypePostgreSQL
)

// MigrationsFS contains embedded pki-ca specific migrations (2001+ only).
//
// Migration version numbering convention:
//   - 1001-1999: Service-template base infrastructure (reserved range, loaded from template package)
//   - 2001+: pki-ca app-specific tables (ca_items)
//
// CRITICAL: This embed ONLY contains pki-ca specific migrations (2001+).
// CRITICAL: Template can add migrations 1005-1999 without conflicts with pki-ca migrations.
// Service-template base infrastructure migrations (1001-1004) are loaded from template package first.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns a filesystem combining template and pki-ca migrations.
// This is used by tests to access all migrations (1001-1999 template + 2001+ pki-ca) in sequence.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsTemplateServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
}

// ApplyPKICAMigrations runs database migrations for pki-ca service.
//
// Two-phase migration loading:
//
// Phase 1 - Service-template base infrastructure (1001-1004):
// - 1001_session_management: browser_session_jwks, service_session_jwks, browser_sessions, service_sessions
// - 1002_barrier_tables: barrier_root_keys, barrier_intermediate_keys, barrier_content_keys
// - 1003_realms_template: template_realms table structure
// - 1004_add_multi_tenancy: tenants, users, clients, unverified_users, unverified_clients, roles, user_roles, client_roles
//
// Phase 2 - pki-ca specific tables (2001+):
// - 2001_ca_items: Minimal CA demonstration table.
func ApplyPKICAMigrations(db *sql.DB, dbType DatabaseType) error {
	// Apply all migrations in sequence (1001-1999 template + 2001+ pki-ca) using merged filesystem.
	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply pki-ca migrations (1001-1999 + 2001+): %w", err)
	}

	return nil
}
