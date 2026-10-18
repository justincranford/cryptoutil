package repository

import (
	"context"
	"database/sql"
	"io/fs"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetMergedMigrationsFS(t *testing.T) {
	t.Parallel()

	merged := GetMergedMigrationsFS()
	entries, err := fs.ReadDir(merged, "migrations")
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	_, err = fs.ReadFile(merged, "migrations/10001_identity_spa_init.up.sql")
	require.NoError(t, err)
	_, err = fs.ReadFile(merged, "migrations/1001_session_management.up.sql")
	require.NoError(t, err)
	_, err = fs.ReadFile(merged, "migrations/9999_missing.up.sql")
	require.Error(t, err)
}

func TestApplyIdentitySPAMigrations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		closed  bool
		wantErr bool
	}{
		{name: "apply migrations"},
		{name: "closed database", closed: true, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
			require.NoError(t, err)
			db.SetMaxOpenConns(1)

			if tc.closed {
				require.NoError(t, db.Close())
			} else {
				t.Cleanup(func() { require.NoError(t, db.Close()) })
			}

			err = ApplyIdentitySPAMigrations(db, DatabaseTypeSQLite)
			if tc.wantErr {
				require.ErrorContains(t, err, "failed to apply identity-spa migrations")

				return
			}

			require.NoError(t, err)

			var count int

			err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM identity_spa_sessions").Scan(&count)
			require.NoError(t, err)
			require.Zero(t, count)
		})
	}
}
