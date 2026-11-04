// Copyright (c) 2025-2026 Justin Cranford.

package test_help_db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewInMemorySQLiteDB_SharedCacheAcrossConnections(t *testing.T) {
	t.Parallel()

	db := NewInMemorySQLiteDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)

	const tableName = "test_helper_connection_probe"

	ctx := context.Background()

	firstConnection, err := sqlDB.Conn(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, firstConnection.Close()) })

	_, err = firstConnection.ExecContext(ctx, "CREATE TABLE "+tableName+" (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	// Hold the creating connection so the named in-memory DB persists while
	// the pool opens a second connection for the visibility check.
	secondConnection, err := sqlDB.Conn(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, secondConnection.Close()) })

	var tableCount int64

	err = secondConnection.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?",
		"table",
		tableName,
	).Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, int64(1), tableCount)
}
