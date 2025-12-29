// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"

	cryptoutilTemplateServer "cryptoutil/internal/template/server"
)

// DatabaseType represents supported database types for learn-im.
type DatabaseType = cryptoutilTemplateServer.DatabaseType

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite = cryptoutilTemplateServer.DatabaseTypeSQLite
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL = cryptoutilTemplateServer.DatabaseTypePostgreSQL
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// ApplyMigrations runs database migrations for learn-im service.
//
// Migrations are embedded in the binary and applied automatically on startup.
// Compatible with both PostgreSQL and SQLite (using TEXT type for cross-DB compatibility).
//
// 3-table design:
// - users: User accounts with PBKDF2-HMAC-SHA256 password hashes
// - messages: Encrypted messages with JWE JSON format (multi-recipient)
// - messages_recipient_jwks: Per-recipient decryption keys (encrypted JWK).
func ApplyMigrations(db *sql.DB, dbType DatabaseType) error {
	runner := cryptoutilTemplateServer.NewMigrationRunner(migrationsFS, "migrations")

	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return runner.Apply(db, dbType)
}
