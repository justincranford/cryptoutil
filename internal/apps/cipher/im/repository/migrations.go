// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"

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

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// ApplyCipherIMMigrations runs database migrations for cipher-im service.
//
// Migrations are embedded in the binary and applied automatically on startup.
// Compatible with both PostgreSQL and SQLite (using TEXT type for cross-DB compatibility).
//
// 3-table design:
// - users: User accounts with PBKDF2-HMAC-SHA256 password hashes
// - messages: Encrypted messages with JWE JSON format (multi-recipient)
// - messages_recipient_jwks: Per-recipient decryption keys (encrypted JWK).
func ApplyCipherIMMigrations(db *sql.DB, dbType DatabaseType) error {
	return cryptoutilTemplateServerRepository.ApplyMigrations(db, dbType, MigrationsFS)
}
