// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

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

// mergedFS combines template and cipher-im migrations into a single filesystem view.
// This allows golang-migrate to see all migrations (1001-1006) in sequence.
type mergedFS struct {
	templateFS embed.FS
	cipherIMFS embed.FS
}

func (m *mergedFS) Open(name string) (fs.File, error) {
	// Try cipher-im filesystem first (1005-1006).
	file, err := m.cipherIMFS.Open(name)
	if err == nil {
		return file, nil
	}

	// Fall back to template filesystem (1001-1004).
	return m.templateFS.Open(name)
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
	return fs.ReadFile(m.templateFS, name)
}

func (m *mergedFS) Stat(name string) (fs.FileInfo, error) {
	// Try cipher-im filesystem first.
	info, err := fs.Stat(m.cipherIMFS, name)
	if err == nil {
		return info, nil
	}

	// Fall back to template filesystem.
	return fs.Stat(m.templateFS, name)
}

// GetMergedMigrationsFS returns a filesystem combining template and cipher-im migrations.
// This is used by tests to access all migrations (1001-1006) in sequence.
func GetMergedMigrationsFS() fs.FS {
	return &mergedFS{
		templateFS: cryptoutilTemplateServerRepository.MigrationsFS,
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
// Phase 2 - Cipher-im specific tables (1005-1006):
// - 1005_init: messages (multi-recipient JWE), messages_recipient_jwks (per-recipient decryption keys)
// - 1006_add_cipher_im_realms: cipher_im_realms configuration (6 non-federated authn methods)
//
// NOTE: users table comes from template 1004_add_multi_tenancy (NOT cipher-im 1005).
func ApplyCipherIMMigrations(db *sql.DB, dbType DatabaseType) error {
	// Apply all migrations in sequence (1001-1006) using merged filesystem.
	runner := cryptoutilTemplateServerRepository.NewMigrationRunner(GetMergedMigrationsFS(), "migrations")

	if err := runner.Apply(db, dbType); err != nil {
		return fmt.Errorf("failed to apply cipher-im migrations (1001-1006): %w", err)
	}

	return nil
}


