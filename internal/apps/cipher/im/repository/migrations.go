// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"
	"fmt"

	cryptoutilTemplateServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// DatabaseType represents supported database types for cipher-im.
type DatabaseType = cryptoutilTemplateServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilTemplateServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilTemplateServerRepository.DatabaseTypePostgreSQL
)

// MigrationsFS contains embedded cipher-im specific migrations (1005-1006 only).
//
// Migration version numbering convention:
//   - 1001-1004: Service-template base infrastructure (loaded from template package)
//   - 1005-1006: Cipher-im app-specific tables (messages, messages_recipient_jwks, cipher_im_realms)
//
// CRITICAL: This embed ONLY contains cipher-im specific migrations (1005-1006).
// Service-template base infrastructure migrations (1001-1004) are loaded from template package first.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// ApplyCipherIMMigrations runs database migrations for cipher-im service.
//
// Two-phase migration loading:
//
// Phase 1 - Service-template base infrastructure (1001-1004):
// - 1001_session_management: browser_session_jwks, service_session_jwks, browser_sessions, service_sessions
// - 1002_barrier_tables: barrier_root_keys, barrier_intermediate_keys, barrier_content_keys
// - 1003_realms_template: template_realms table structure (services create their own <service>_realms tables)
// - 1004_add_multi_tenancy: tenants, users, clients, unverified_users, unverified_clients, roles, user_roles, client_roles
//
// Phase 2 - Cipher-im specific tables (1005-1006):
// - 1005_init: messages (multi-recipient JWE), messages_recipient_jwks (per-recipient decryption keys)
// - 1006_add_cipher_im_realms: cipher_im_realms configuration (6 non-federated authn methods)
//
// NOTE: users table comes from template 1004_add_multi_tenancy (NOT cipher-im 1005).
func ApplyCipherIMMigrations(db *sql.DB, dbType DatabaseType) error {
	// Phase 1: Apply template base infrastructure migrations (1001-1004).
	templateRunner := cryptoutilTemplateServerRepository.NewMigrationRunner(
		cryptoutilTemplateServerRepository.MigrationsFS,
		"migrations",
	)

	if err := templateRunner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply template base migrations (1001-1004): %w", err)
	}

	// Phase 2: Apply cipher-im specific migrations (1005-1006).
	cipherIMRunner := cryptoutilTemplateServerRepository.NewMigrationRunner(MigrationsFS, "migrations")

	if err := cipherIMRunner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply cipher-im specific migrations (1005-1006): %w", err)
	}

	return nil
}

