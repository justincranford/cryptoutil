// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInitPostgreSQL_InvalidDatabaseURL tests PostgreSQL initialization with invalid database URL.
func TestInitPostgreSQL_InvalidDatabaseURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Invalid database URL (malformed connection string).
	db, err := InitPostgreSQL(ctx, "invalid://\x00malformed", testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
	// Error message should mention failed to ping (DSN parsing happens in ping, not open).
	require.Contains(t, err.Error(), "failed to ping PostgreSQL database")
}

// TestInitPostgreSQL_UnreachableHost tests PostgreSQL initialization with unreachable host.
func TestInitPostgreSQL_UnreachableHost(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Unreachable host (should fail on ping).
	db, err := InitPostgreSQL(ctx, "postgres://user:pass@999.999.999.999:5432/db", testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
	// Error message should mention ping failure.
	require.Contains(t, err.Error(), "failed to")
}

// TestInitSQLite_WALModeFailure tests SQLite initialization when WAL mode fails.
// Note: This is difficult to trigger in practice without mocking, but we can test with invalid pragma.
func TestInitSQLite_InvalidPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Invalid path that cannot be created (contains null byte).
	db, err := InitSQLite(ctx, "file:/tmp/\x00invalid.db", testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
}
