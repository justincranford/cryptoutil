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

// DatabaseType represents supported database types for cipher-im.
type DatabaseType = cryptoutilAppsTemplateServiceServerRepository.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilAppsTemplateServiceServerRepository.DatabaseTypePostgreSQL
)

// MigrationsFS contains embedded cipher-im specific migrations (1005-1006 only).
//
// Migration version numbering convention:
//   - 1001-1999: Service-template base infrastructure (reserved range, loaded from template package)
//   - 2001+: Cipher-im app-specific tables (messages, messages_recipient_jwks)
//
// CRITICAL: This embed ONLY contains cipher-im specific migrations (2001+).
// CRITICAL: Template can add migrations 1005-1999 without conflicts with cipher-im migrations.
// Service-template base infrastructure migrations (1001-1004) are loaded from template package first.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// mergedFS combines template and cipher-im migrations into a single filesystem view.
// This allows golang-migrate to see all migrations (1001-1006) in sequence.
type mergedFS struct {
	templateFS embed.FS
	cipherIMFS embed.FS
}

func (m *mergedFS) Open(name string) (fs.File, error) {
	// Try cipher-im filesystem first (2001+).
	file, err := m.cipherIMFS.Open(name)
	if err == nil {
		return file, nil
	}

	// Fall back to template filesystem (1001-1004).
	file, err = m.templateFS.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open from template: %w", err)
	}

	return file, nil
}

func (m *mergedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	// Read both directories and merge results.
	var entries []fs.DirEntry

	// Read template migrations (1001-1004).
	templateEntries, err := m.templateFS.ReadDir(name)
	if err == nil {
		entries = append(entries, templateEntries...)
	}

	// Read cipher-im migrations (1005-1006).
	cipherIMEntries, err := m.cipherIMFS.ReadDir(name)
	if err == nil {
		entries = append(entries, cipherIMEntries...)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("directory not found: %s", name)
	}

	return entries, nil
}

func (m *mergedFS) ReadFile(name string) ([]byte, error) {
	// Try cipher-im filesystem first (1005-1006).
	data, err := fs.ReadFile(m.cipherIMFS, name)
	if err == nil {
		return data, nil
	}

	// Fall back to template filesystem (1001-1004).
	data, err = fs.ReadFile(m.templateFS, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read from template: %w", err)
	}

	return data, nil
}

func (m *mergedFS) Stat(name string) (fs.FileInfo, error) {
	// Try cipher-im filesystem first.
	info, err := fs.Stat(m.cipherIMFS, name)
	if err == nil {
		return info, nil
	}

	// Fall back to template filesystem.
	info, err = fs.Stat(m.templateFS, name)
	if err != nil {
		return nil, fmt.Errorf("failed to stat from template: %w", err)
	}

	return info, nil
}

// GetMergedMigrationsFS returns a filesystem combining template and cipher-im migrations.
// This is used by tests to access all migrations (1001-1999 template + 2001+ cipher-im) in sequence.
func GetMergedMigrationsFS() fs.FS {
	return &mergedFS{
		templateFS: cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
		cipherIMFS: MigrationsFS,
	}
}

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
// Phase 2 - Cipher-im specific tables (2001+):
// - 2001_init: messages (multi-recipient JWE), messages_recipient_jwks (per-recipient decryption keys)
//
// NOTE: users table comes from template 1004_add_multi_tenancy (NOT cipher-im 2001).
// NOTE: cipher-im uses template_realms from template 1003_realms_template (NOT custom cipher_im_realms table).
func ApplyCipherIMMigrations(db *sql.DB, dbType DatabaseType) error {
	// Apply all migrations in sequence (1001-1999 template + 2001+ cipher-im) using merged filesystem.
	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply cipher-im migrations (1001-1999 + 2001+): %w", err)
	}

	return nil
}
