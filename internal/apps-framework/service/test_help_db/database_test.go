// Copyright (c) 2025-2026 Justin Cranford.

package test_help_db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	testDBCaseSuccess                 = "success"
	testDBErrorDockerUnavailablePanic = "docker unavailable (panic)"
	testDBErrorFailedGenerateUUID     = "failed to generate UUID"
	testDBCaseOpenError               = "open error"
	testDBErrorSQLOpen                = "sql.Open"
	testDBCaseUUIDPanic               = "uuid panic"
	testDBCaseWALError                = "wal error"
	testDBErrorWALPragma              = "WAL pragma"
	testDBCaseGormOpenError           = "gorm open error"
	testDBErrorGormOpen               = "gorm.Open"
	testDBErrorApplyMigrations        = "apply migrations"
)

func newIsolatedSQLiteDSN(t *testing.T) string {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	return fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, dbID.String())
}

func TestNewInMemorySQLiteDBForTestMain_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		depsFn         func() dbDeps
		wantErr        string
		wantCleanupErr string
	}{
		{name: testDBCaseSuccess, depsFn: defaultDBDeps},
		{
			name: "uuid error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.newUUIDv7Fn = func() (googleUuid.UUID, error) { return googleUuid.UUID{}, errors.New("uuid failure") }

				return deps
			},
			wantErr: testDBErrorFailedGenerateUUID,
		},
		{
			name: testDBCaseOpenError,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantErr: testDBErrorSQLOpen,
		},
		{
			name: "cleanup error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.closeSQLDBFn = func(*sql.DB) error { return errors.New("close failure") }

				return deps
			},
			wantCleanupErr: "close SQLite TestMain database",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, cleanup, err := newInMemorySQLiteDBForTestMainWithDeps(tc.depsFn())
			if tc.wantErr != "" {
				require.Nil(t, db)
				require.Nil(t, cleanup)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, db)
			require.NotNil(t, cleanup)

			sqlDB, sqlErr := db.DB()
			require.NoError(t, sqlErr)
			require.NoError(t, sqlDB.Ping())

			cleanupErr := cleanup()
			if tc.wantCleanupErr != "" {
				require.ErrorContains(t, cleanupErr, tc.wantCleanupErr)
			} else {
				require.NoError(t, cleanupErr)
			}
		})
	}
}

func TestNewInMemorySQLiteDB_Wrapper_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		depsFn    func() dbDeps
		wantPanic string
	}{
		{name: testDBCaseSuccess, depsFn: defaultDBDeps},
		{
			name: testDBCaseUUIDPanic,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.newUUIDv7Fn = func() (googleUuid.UUID, error) { return googleUuid.UUID{}, errors.New("uuid failure") }

				return deps
			},
			wantPanic: testDBErrorFailedGenerateUUID,
		},
		{
			name: "open panic",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantPanic: testDBErrorSQLOpen,
		},
		{
			name: "cleanup close error path",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.closeSQLDBFn = func(*sql.DB) error { return errors.New("close failure") }

				return deps
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.wantPanic != "" {
				require.Panics(t, func() { _ = newInMemorySQLiteDBWithDeps(t, tc.depsFn()) })

				return
			}

			db := newInMemorySQLiteDBWithDeps(t, tc.depsFn())
			require.NotNil(t, db)
		})
	}
}

func TestBuildInMemorySQLiteDB_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		depsFn  func() dbDeps
		wantErr string
	}{
		{
			name: testDBCaseOpenError,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantErr: testDBErrorSQLOpen,
		},
		{
			name: testDBCaseWALError,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.execSQLitePragmaFn = func(context.Context, *sql.DB, string) error { return errors.New("pragma failure") }

				return deps
			},
			wantErr: testDBErrorWALPragma,
		},
		{
			name: "busy timeout error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				calls := 0
				deps.execSQLitePragmaFn = func(context.Context, *sql.DB, string) error {
					calls++
					if calls == 2 {
						return errors.New("busy timeout failure")
					}

					return nil
				}

				return deps
			},
			wantErr: "busy_timeout pragma",
		},
		{
			name: testDBCaseGormOpenError,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.openGormSQLiteFn = func(*sql.DB) (*gorm.DB, error) { return nil, errors.New("gorm open failure") }

				return deps
			},
			wantErr: testDBErrorGormOpen,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, sqlDB, err := buildInMemorySQLiteDB(context.Background(), tc.depsFn(), newIsolatedSQLiteDSN(t))
			require.Nil(t, db)
			require.Nil(t, sqlDB)
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestNewClosedSQLiteDB_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		depsFn        func() dbDeps
		applyMigrate  func(*sql.DB) error
		wantPanicText string
	}{
		{name: testDBCaseSuccess, depsFn: defaultDBDeps},
		{name: "success with migration", depsFn: defaultDBDeps, applyMigrate: func(*sql.DB) error { return nil }},
		{
			name: testDBCaseUUIDPanic,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.newUUIDv7Fn = func() (googleUuid.UUID, error) { return googleUuid.UUID{}, errors.New("uuid failure") }

				return deps
			},
			wantPanicText: cryptoutilSharedMagic.IdentityTokenFormatUUID,
		},
		{
			name:   "migration panic",
			depsFn: defaultDBDeps,
			applyMigrate: func(*sql.DB) error {
				return errors.New("migration failure")
			},
			wantPanicText: testDBErrorApplyMigrations,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.wantPanicText != "" {
				require.Panics(t, func() {
					_ = newClosedSQLiteDBWithDeps(t, tc.applyMigrate, tc.depsFn())
				})

				return
			}

			db := newClosedSQLiteDBWithDeps(t, tc.applyMigrate, tc.depsFn())
			require.NotNil(t, db)
		})
	}
}

func TestBuildClosedSQLiteDB_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		depsFn     func() dbDeps
		migrateFn  func(*sql.DB) error
		wantErrSub string
	}{
		{
			name: testDBCaseOpenError,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantErrSub: testDBErrorSQLOpen,
		},
		{
			name: testDBCaseWALError,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.execSQLitePragmaFn = func(context.Context, *sql.DB, string) error { return errors.New("wal failure") }

				return deps
			},
			wantErrSub: testDBErrorWALPragma,
		},
		{
			name:       "migration error",
			depsFn:     defaultDBDeps,
			migrateFn:  func(*sql.DB) error { return errors.New("migration failure") },
			wantErrSub: testDBErrorApplyMigrations,
		},
		{
			name: testDBCaseGormOpenError,
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.openGormSQLiteFn = func(*sql.DB) (*gorm.DB, error) { return nil, errors.New("gorm open failure") }

				return deps
			},
			wantErrSub: testDBErrorGormOpen,
		},
		{
			name: "close error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.closeSQLDBFn = func(*sql.DB) error { return errors.New("close failure") }

				return deps
			},
			wantErrSub: "close DB after migrations",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, err := buildClosedSQLiteDB(context.Background(), tc.depsFn(), newIsolatedSQLiteDSN(t), tc.migrateFn)
			require.Nil(t, db)
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErrSub)
		})
	}
}

func TestSQLiteSetupErrorsReportCloseFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(dbDeps) error
	}{
		{
			name: "in-memory WAL setup",
			build: func(deps dbDeps) error {
				deps.execSQLitePragmaFn = func(context.Context, *sql.DB, string) error { return errors.New("setup failure") }
				_, _, err := buildInMemorySQLiteDB(context.Background(), deps, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)

				return err
			},
		},
		{
			name: "in-memory GORM setup",
			build: func(deps dbDeps) error {
				deps.openGormSQLiteFn = func(*sql.DB) (*gorm.DB, error) { return nil, errors.New("setup failure") }
				_, _, err := buildInMemorySQLiteDB(context.Background(), deps, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)

				return err
			},
		},
		{
			name: "closed DB migration",
			build: func(deps dbDeps) error {
				_, err := buildClosedSQLiteDB(context.Background(), deps, cryptoutilSharedMagic.SQLiteMemoryPlaceholder, func(*sql.DB) error {
					return errors.New("setup failure")
				})

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := defaultDBDeps()
			deps.closeSQLDBFn = func(*sql.DB) error { return errors.New("close failure") }

			err := tc.build(deps)
			require.Error(t, err)
			require.ErrorContains(t, err, "setup failure")
			require.ErrorContains(t, err, "close SQLite DB after setup failure")
			require.ErrorContains(t, err, "close failure")
		})
	}
}

func TestSQLitePublicWrappers(t *testing.T) {
	t.Parallel()

	db, cleanup, err := NewInMemorySQLiteDBForTestMain()
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, cleanup)
	require.NoError(t, cleanup())

	require.NotNil(t, NewInMemorySQLiteDB(t))
	require.NotNil(t, NewClosedSQLiteDB(t, nil))
}

func TestDefaultSeamFunctions_Coverage(t *testing.T) {
	t.Parallel()

	deps := defaultDBDeps()
	sqlDB, err := deps.sqlOpenFn(cryptoutilSharedMagic.TestDatabaseSQLite, newIsolatedSQLiteDSN(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, deps.execSQLitePragmaFn(context.Background(), sqlDB, "PRAGMA journal_mode=WAL;"))

	gormDB, gormErr := deps.openGormSQLiteFn(sqlDB)
	require.NoError(t, gormErr)
	require.NotNil(t, gormDB)

	handle, handleErr := deps.getSQLDBFn(gormDB)
	require.NoError(t, handleErr)
	require.NotNil(t, handle)

	t.Run("default pragma error wrapping", func(t *testing.T) {
		t.Parallel()

		closedDB, openErr := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, newIsolatedSQLiteDSN(t))
		require.NoError(t, openErr)
		require.NoError(t, closedDB.Close())

		err := defaultExecSQLitePragma(context.Background(), closedDB, "PRAGMA journal_mode=WAL;")
		require.Error(t, err)
		require.ErrorContains(t, err, "ExecContext")
	})

	t.Run("wrapIfErr table", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			err     error
			wantErr string
		}{
			{name: "nil error", err: nil},
			{name: "wrapped error", err: errors.New("boom"), wantErr: "operation"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				wrappedErr := wrapIfErr("operation", tc.err)
				if tc.wantErr == "" {
					require.NoError(t, wrappedErr)

					return
				}

				require.Error(t, wrappedErr)
				require.ErrorContains(t, wrappedErr, tc.wantErr)
			})
		}
	})
}
