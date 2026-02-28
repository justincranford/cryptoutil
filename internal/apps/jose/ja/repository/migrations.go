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

// DatabaseType represents supported database types for jose-ja.
type DatabaseType = cryptoutilAppsTemplateServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypePostgreSQL
)

// MigrationsFS contains embedded jose-ja specific migrations (2001+ only).
//
// Migration version numbering convention:
//   - 1001-1999: Service-template base infrastructure (reserved range, loaded from template package)
//   - 2001+: JOSE-JA app-specific tables (elastic_jwks, material_jwks, audit_config, audit_log)
//
// CRITICAL: This embed ONLY contains jose-ja specific migrations (2001+).
// CRITICAL: Template can add migrations 1005-1999 without conflicts with jose-ja migrations.
// Service-template base infrastructure migrations (1001-1004) are loaded from template package first.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// GetMergedMigrationsFS returns a filesystem combining template and JOSE-JA migrations.
// This is used by tests to access all migrations (1001-1999 template + 2001+ JOSE-JA) in sequence.
func GetMergedMigrationsFS() fs.FS {
	return cryptoutilAppsTemplateServiceServerRepository.NewMergedMigrationsFS(MigrationsFS)
}

// ApplyJoseJAMigrations runs database migrations for jose-ja service.
//
// Two-phase migration loading:
//
// Phase 1 - Service-template base infrastructure (1001-1004):
// - 1001_session_management: browser_session_jwks, service_session_jwks, browser_sessions, service_sessions
// - 1002_barrier_tables: barrier_root_keys, barrier_intermediate_keys, barrier_content_keys
// - 1003_realms_template: template_realms table structure
// - 1004_add_multi_tenancy: tenants, users, clients, unverified_users, unverified_clients, roles, user_roles, client_roles
//
// Phase 2 - JOSE-JA specific tables (2001+):
// - 2001_elastic_jwks: Elastic JWK containers for key rotation
// - 2002_material_jwks: Encrypted key material versions
// - 2003_audit_config: Per-tenant audit configuration
// - 2004_audit_log: Cryptographic operation audit entries.
func ApplyJoseJAMigrations(db *sql.DB, dbType DatabaseType) error {
	// Apply all migrations in sequence (1001-1999 template + 2001+ jose-ja) using merged filesystem.
	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply jose-ja migrations (1001-1999 + 2001+): %w", err)
	}

	return nil
}
