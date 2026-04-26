// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed test_migrations/*.sql
var testDBMigrationsFS embed.FS

// TestInitSQLite_InvalidDatabaseURL tests SQLite initialization with invalid database URL.
func TestInitSQLite_InvalidDatabaseURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Invalid database URL (malformed path).
	db, err := InitSQLite(ctx, "\x00invalid", testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
}
