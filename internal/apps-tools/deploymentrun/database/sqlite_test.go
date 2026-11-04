// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetSharedSQLite(t *testing.T) {
	t.Helper()

	rows, err := sharedSQLite.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	require.NoError(t, err)

	names := []string{}

	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}

	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())

	for _, name := range names {
		_, err := sharedSQLite.Exec(`DROP TABLE "` + strings.ReplaceAll(name, `"`, `""`) + `"`)
		require.NoError(t, err)
	}
}

func TestRealSQLiteProductionMigrations(t *testing.T) {
	t.Parallel()

	for _, psid := range []string{psidAuthz, psidIDP, psidRP, psidRS, psidSPA, psidSkeleton, psidKMS} {
		t.Run(psid, func(t *testing.T) {
			t.Parallel()
			sqliteMu.Lock()
			defer sqliteMu.Unlock()

			resetSharedSQLite(t)
			require.NoError(t, Apply(context.Background(), sharedSQLite, psid, SQLite))

			var (
				version int
				dirty   bool
			)
			require.NoError(t, sharedSQLite.QueryRow("SELECT version,dirty FROM schema_migrations").Scan(&version, &dirty))
			require.Greater(t, version, 1999)
			require.False(t, dirty)
			require.NoError(t, Apply(context.Background(), sharedSQLite, psid, SQLite))
			client := &SQLClient{DB: sharedSQLite, backend: SQLite}
			require.NoError(t, client.RequireDatabase(context.Background(), "explicit_sqlite", testRole))
			require.NoError(t, client.Migrate(context.Background(), psid, SQLite))
		})
	}
}

func TestApplyRejectsInvalidOrCancelledInput(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, psid, backend string
		cancel              bool
	}{
		{testUnknownService, testUnknown, SQLite, false}, {testUnknownBackend, psidKMS, testUnknown, false}, {testCancelled, psidKMS, SQLite, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if test.cancel {
				cancel()
			}

			require.Error(t, Apply(ctx, sharedSQLite, test.psid, test.backend))
		})
	}
}
