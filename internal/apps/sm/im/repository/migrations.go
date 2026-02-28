// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// DatabaseType represents supported database types for sm-im.
type DatabaseType = cryptoutilAppsTemplateServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypePostgreSQL
)

// MigrationsFS contains embedded sm-im specific migrations (1005-1006 only).
//
// Migration version numbering convention:
//   - 1001-1999: Service-template base infrastructure (reserved range, loaded from template package)
//   - 2001+: Sm-im app-specific tables (messages, messages_recipient_jwks)
//
// CRITICAL: This embed ONLY contains sm-im specific migrations (2001+).
// CRITICAL: Template can add migrations 1005-1999 without conflicts with sm-im migrations.
// Service-template base infrastructure migrations (1001-1004) are loaded from template package first.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns a filesystem combining template and sm-im migrations.
// This is used by tests to access all migrations (1001-1999 template + 2001+ sm-im) in sequence.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsTemplateServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
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
// Phase 2 - Sm-im specific tables (2001+):
// - 2001_init: messages (multi-recipient JWE), messages_recipient_jwks (per-recipient decryption keys)
//
// NOTE: users table comes from template 1004_add_multi_tenancy (NOT sm-im 2001).
// NOTE: sm-im uses template_realms from template 1003_realms_template (NOT custom sm_im_realms table).
func ApplySmIMMigrations(db *sql.DB, dbType DatabaseType) error {
	// Apply all migrations in sequence (1001-1999 template + 2001+ sm-im) using merged filesystem.
	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply sm-im migrations (1001-1999 + 2001+): %w", err)
	}

	return nil
}
