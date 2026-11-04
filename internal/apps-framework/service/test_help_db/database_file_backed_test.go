// Copyright (c) 2025-2026 Justin Cranford.

package test_help_db

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNewFileBackedSQLiteDB_ReopenPreservesData(t *testing.T) {
	t.Parallel()

	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "durable.sqlite")) +
		"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db := NewFileBackedSQLiteDB(t, dsn)
	require.NoError(t, db.Exec("CREATE TABLE durable_values (value TEXT NOT NULL)").Error)
	require.NoError(t, db.Exec("INSERT INTO durable_values (value) VALUES (?)", "preserved").Error)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	reopenedDB := NewFileBackedSQLiteDB(t, dsn)

	var value string
	require.NoError(t, reopenedDB.Raw("SELECT value FROM durable_values").Scan(&value).Error)
	require.Equal(t, "preserved", value)
}

func TestNewFileBackedSQLiteDB_RejectsNonFileBackedDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dsn  string
	}{
		{name: "empty DSN"},
		{name: "non-file DSN", dsn: cryptoutilSharedMagic.SQLiteMemoryPlaceholder},
		{name: "empty file name", dsn: "file:?cache=shared"},
		{name: "anonymous in-memory file URI", dsn: cryptoutilSharedMagic.SQLiteInMemoryDSN},
		{
			name: "named in-memory file URI",
			dsn:  fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, "memory-probe"),
		},
		{name: "memory mode with file name", dsn: "file:database?mode=memory&cache=shared"},
		{name: "escaped anonymous in-memory file URI", dsn: "file:%3Amemory%3A?cache=shared"},
		{name: "malformed query", dsn: "file:database?mode=%zz"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.PanicsWithValue(t, "test_help_db: file-backed SQLite DSN must name a persistent file", func() {
				NewFileBackedSQLiteDB(t, tc.dsn)
			})
		})
	}
}

func TestIsFileBackedSQLiteDSN_AcceptsPersistentFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dsn  string
	}{
		{name: "file without query", dsn: "file:durable.sqlite"},
		{name: "file with query", dsn: "file:durable.sqlite?cache=shared"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.True(t, isFileBackedSQLiteDSN(tc.dsn))
		})
	}
}

func TestNewFileBackedSQLiteDB_SetupErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		depsFn     func(*bool) dbDeps
		wantPanic  string
		wantClosed bool
	}{
		{
			name: "SQL open error",
			depsFn: func(_ *bool) dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) {
					return nil, errors.New("file-backed open failure")
				}

				return deps
			},
			wantPanic: "test_help_db: sql.Open: file-backed open failure",
		},
		{
			name: "GORM open error closes SQL DB",
			depsFn: func(closed *bool) dbDeps {
				deps := defaultDBDeps()
				deps.openGormSQLiteFn = func(*sql.DB) (*gorm.DB, error) {
					return nil, errors.New("file-backed GORM failure")
				}
				closeSQLDB := deps.closeSQLDBFn
				deps.closeSQLDBFn = func(sqlDB *sql.DB) error {
					*closed = true

					return closeSQLDB(sqlDB)
				}

				return deps
			},
			wantPanic:  "test_help_db: file-backed GORM failure",
			wantClosed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			closed := false
			deps := tc.depsFn(&closed)
			dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "setup-error.sqlite"))

			require.PanicsWithValue(t, tc.wantPanic, func() {
				_ = newFileBackedSQLiteDBWithDeps(t, dsn, deps)
			})
			require.Equal(t, tc.wantClosed, closed)
		})
	}
}

func TestNewFileBackedSQLiteDB_ReportsCleanupError(t *testing.T) {
	t.Parallel()

	closed := false

	t.Cleanup(func() {
		require.True(t, closed)
	})

	deps := defaultDBDeps()
	deps.closeSQLDBFn = func(sqlDB *sql.DB) error {
		if err := defaultCloseSQLDB(sqlDB); err != nil {
			return err
		}

		closed = true

		return errors.New("file-backed cleanup failure")
	}

	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "cleanup-error.sqlite"))
	db := newFileBackedSQLiteDBWithDeps(t, dsn, deps)
	require.NotNil(t, db)
}
