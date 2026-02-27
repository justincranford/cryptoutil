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

// mergedFS combines template and pki-ca migrations into a single filesystem view.
// This allows golang-migrate to see all migrations (1001-1004 + 2001+) in sequence.
type mergedFS struct {
templateFS embed.FS
pkiCaFS    embed.FS
}

func (m *mergedFS) Open(name string) (fs.File, error) {
// Try pki-ca filesystem first (2001+).
file, err := m.pkiCaFS.Open(name)
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

// Read pki-ca migrations (2001+).
pkiCaEntries, err := m.pkiCaFS.ReadDir(name)
if err == nil {
entries = append(entries, pkiCaEntries...)
}

if len(entries) == 0 {
return nil, fmt.Errorf("directory not found: %s", name)
}

return entries, nil
}

func (m *mergedFS) ReadFile(name string) ([]byte, error) {
// Try pki-ca filesystem first (2001+).
data, err := fs.ReadFile(m.pkiCaFS, name)
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
// Try pki-ca filesystem first.
info, err := fs.Stat(m.pkiCaFS, name)
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

// GetMergedMigrationsFS returns a filesystem combining template and pki-ca migrations.
// This is used by tests to access all migrations (1001-1999 template + 2001+ pki-ca) in sequence.
func GetMergedMigrationsFS() fs.FS {
return &mergedFS{
templateFS: cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
pkiCaFS:    MigrationsFS,
}
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
