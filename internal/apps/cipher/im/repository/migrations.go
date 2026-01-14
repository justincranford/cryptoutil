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

// MigrationsFS contains all embedded migrations for cipher-im service (1001-1006).
//
// Migration version numbering convention:
//   - 1001-1004: Service-template base infrastructure (session mgmt, barrier, realms template, multi-tenancy)
//   - 1005-1006: Cipher-im app-specific tables (users, messages, cipher_im_realms)
//
// Note: Service-template migrations are duplicated here for self-contained deployment.
// Each service maintains complete migration history for independent deployment.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// ApplyCipherIMMigrations runs database migrations for cipher-im service.
//
// Applies all migrations in sequence (1001-1006):
//
// Service-template infrastructure (1001-1004):
// - browser_session_jwks, service_session_jwks: Session JWK storage
// - browser_sessions, service_sessions: Session tracking
// - barrier_root_keys, barrier_intermediate_keys, barrier_content_keys: Barrier encryption
// - template_realms: Realm configuration template
// - tenants: Multi-tenancy support
//
// Cipher-im specific (1005-1006):
// - users: User accounts with PBKDF2-HMAC-SHA256 password hashes
// - messages: Encrypted messages with JWE JSON format (multi-recipient)
// - messages_recipient_jwks: Per-recipient decryption keys (encrypted JWK)
// - cipher_im_realms: Cipher-IM specific realm configuration.
func ApplyCipherIMMigrations(db *sql.DB, dbType DatabaseType) error {
	if err := cryptoutilTemplateServerRepository.ApplyMigrations(db, dbType, MigrationsFS); err != nil {
		return fmt.Errorf("failed to apply cipher-im migrations: %w", err)
	}

	return nil
}
