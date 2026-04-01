// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
)

// DatabaseType represents supported database types for sm-im.
type DatabaseType = cryptoutilAppsFrameworkServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL
)

// MigrationsFS contains embedded sm-im specific migrations (3001+ only).
//
// Migration version numbering convention:
//   - 1001-1999: Service-template base infrastructure (reserved range, loaded from template package)
//   - 3001-3999: Sm-im app-specific tables (messages, messages_recipient_jwks)
//
// CRITICAL: This embed ONLY contains sm-im specific migrations (3001+).
// CRITICAL: Template can add migrations 1005-1999 without conflicts with sm-im migrations.
// Service-template base infrastructure migrations (1001-1004) are loaded from template package first.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns a filesystem combining template and sm-im migrations.
// This is used by tests to access all migrations (1001-1999 template + 3001+ sm-im) in sequence.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
}

// ApplySmIMMigrations runs database migrations for sm-im service.
//
// Two-phase migration loading:
//
// Phase 1 - Service-template base infrastructure (1001-1004):
// - 1001_session_management: browser_session_jwks, service_session_jwks, browser_sessions, service_sessions
// - 1002_barrier_tables: barrier_root_keys, barrier_intermediate_keys, barrier_content_keys
// - 1003_realms_template: template_realms table structure (services create their own <service>_realms tables)
// - 1004_add_multi_tenancy: tenants, users, clients, unverified_users, unverified_clients, roles, user_roles, client_roles
//
// Phase 2 - Sm-im specific tables (3001+):
// - 3001_init: messages (multi-recipient JWE), messages_recipient_jwks (per-recipient decryption keys)
//
// NOTE: users table comes from template 1004_add_multi_tenancy (NOT sm-im 2001).
// NOTE: sm-im uses template_realms from template 1003_realms_template (NOT custom sm_im_realms table).
func ApplySmIMMigrations(db *sql.DB, dbType DatabaseType) error {
	// Apply all migrations in sequence (1001-1999 template + 3001+ sm-im) using merged filesystem.
	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply sm-im migrations (1001-1999 + 3001+): %w", err)
	}

	return nil
}
