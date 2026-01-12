// Copyright (c) 2025 Justin Cranford

package im

import (
	"testing"

	cipherIMRepository "cryptoutil/internal/apps/cipher/im/repository"
	cryptoutilTemplateServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

const (
	countTablesQueryPostgres = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'messages', 'messages_recipient_jwks')"
	countTablesQuerySQLite   = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'messages', 'messages_recipient_jwks')"
	expectedTableCount       = 3
)

// TestInitDatabase_HappyPaths tests successful database initialization for PostgreSQL and SQLite.
func TestInitDatabase_HappyPaths(t *testing.T) {
	cryptoutilTemplateServerTestutil.HelpTest_InitDatabase_HappyPaths(
		t,
		cipherIMRepository.MigrationsFS,
		expectedTableCount,
		countTablesQueryPostgres,
		countTablesQuerySQLite,
	)
}

// TestInitDatabase_SadPaths tests error handling for database initialization failures.
func TestInitDatabase_SadPaths(t *testing.T) {
	cryptoutilTemplateServerTestutil.HelpTest_InitDatabase_SadPaths(t, cipherIMRepository.MigrationsFS)
}
