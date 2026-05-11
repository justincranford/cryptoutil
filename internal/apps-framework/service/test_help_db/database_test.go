// Copyright (c) 2025-2026 Justin Cranford.

package test_help_db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	postgresContainerModule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testUnusedPostgresURL = "postgres://unused"

func TestNewInMemorySQLiteDBForTestMain_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		depsFn    func() dbDeps
		wantErr   string
		wantPanic bool
	}{
		{name: "success", depsFn: defaultDBDeps},
		{
			name: "uuid error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.newUUIDv7Fn = func() (googleUuid.UUID, error) { return googleUuid.UUID{}, errors.New("uuid failure") }

				return deps
			},
			wantErr: "failed to generate UUID",
		},
		{
			name: "open error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantErr: "sql.Open",
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
			cleanup()
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
		{name: "success", depsFn: defaultDBDeps},
		{
			name: "uuid panic",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.newUUIDv7Fn = func() (googleUuid.UUID, error) { return googleUuid.UUID{}, errors.New("uuid failure") }

				return deps
			},
			wantPanic: "failed to generate UUID",
		},
		{
			name: "open panic",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantPanic: "sql.Open",
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
			name: "open error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantErr: "sql.Open",
		},
		{
			name: "wal error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.execSQLitePragmaFn = func(context.Context, *sql.DB, string) error { return errors.New("pragma failure") }

				return deps
			},
			wantErr: "WAL pragma",
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
			name: "gorm open error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.openGormSQLiteFn = func(*sql.DB) (*gorm.DB, error) { return nil, errors.New("gorm open failure") }

				return deps
			},
			wantErr: "gorm.Open",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, sqlDB, err := buildInMemorySQLiteDB(context.Background(), tc.depsFn(), "file:test-build-memory?mode=memory&cache=shared")
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
		{name: "success", depsFn: defaultDBDeps},
		{name: "success with migration", depsFn: defaultDBDeps, applyMigrate: func(*sql.DB) error { return nil }},
		{
			name: "uuid panic",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.newUUIDv7Fn = func() (googleUuid.UUID, error) { return googleUuid.UUID{}, errors.New("uuid failure") }

				return deps
			},
			wantPanicText: "uuid",
		},
		{
			name:   "migration panic",
			depsFn: defaultDBDeps,
			applyMigrate: func(*sql.DB) error {
				return errors.New("migration failure")
			},
			wantPanicText: "apply migrations",
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
			name: "open error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.sqlOpenFn = func(_, _ string) (*sql.DB, error) { return nil, errors.New("open failure") }

				return deps
			},
			wantErrSub: "sql.Open",
		},
		{
			name: "wal error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.execSQLitePragmaFn = func(context.Context, *sql.DB, string) error { return errors.New("wal failure") }

				return deps
			},
			wantErrSub: "WAL pragma",
		},
		{
			name:       "migration error",
			depsFn:     defaultDBDeps,
			migrateFn:  func(*sql.DB) error { return errors.New("migration failure") },
			wantErrSub: "apply migrations",
		},
		{
			name: "gorm open error",
			depsFn: func() dbDeps {
				deps := defaultDBDeps()
				deps.openGormSQLiteFn = func(*sql.DB) (*gorm.DB, error) { return nil, errors.New("gorm open failure") }

				return deps
			},
			wantErrSub: "gorm.Open",
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

			db, err := buildClosedSQLiteDB(context.Background(), tc.depsFn(), "file:test-build-closed?mode=memory&cache=shared", tc.migrateFn)
			require.Nil(t, db)
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErrSub)
		})
	}
}

func TestSafeNewPostgresTestContainer_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		factoryFn func(context.Context) (*postgresContainerModule.PostgresContainer, error)
		wantErr   string
	}{
		{
			name: "factory error",
			factoryFn: func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				return nil, errors.New("factory error")
			},
			wantErr: "postgres container",
		},
		{
			name: "factory panic",
			factoryFn: func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				panic("factory panic")
			},
			wantErr: "docker unavailable (panic)",
		},
		{
			name: "success",
			factoryFn: func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				return &postgresContainerModule.PostgresContainer{}, nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			container, err := safeNewPostgresTestContainer(context.Background(), tc.factoryFn)
			if tc.wantErr != "" {
				require.Nil(t, container)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, container)
		})
	}
}

func TestNewPostgresTestContainer_WithDeps(t *testing.T) {
	t.Parallel()

	t.Run("skip path", func(t *testing.T) {
		t.Parallel()

		deps := defaultDBDeps()
		deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
			return nil, fmt.Errorf("forced skip")
		}

		_ = newPostgresTestContainerWithDeps(context.Background(), t, deps)
	})

	t.Run("connection string panic", func(t *testing.T) {
		t.Parallel()

		deps := defaultDBDeps()
		deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
			return &postgresContainerModule.PostgresContainer{}, nil
		}
		deps.containerConnectionStringFn = func(context.Context, *postgresContainerModule.PostgresContainer) (string, error) {
			return "", errors.New("connection string failure")
		}
		deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
			return nil
		}

		require.Panics(t, func() {
			_ = newPostgresTestContainerWithDeps(context.Background(), t, deps)
		})
	})

	t.Run("gorm open panic", func(t *testing.T) {
		t.Parallel()

		deps := defaultDBDeps()
		deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
			return &postgresContainerModule.PostgresContainer{}, nil
		}
		deps.containerConnectionStringFn = func(context.Context, *postgresContainerModule.PostgresContainer) (string, error) {
			return testUnusedPostgresURL, nil
		}
		deps.openGormPostgresFn = func(string) (*gorm.DB, error) {
			return nil, errors.New("gorm open failure")
		}
		deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
			return nil
		}

		require.Panics(t, func() {
			_ = newPostgresTestContainerWithDeps(context.Background(), t, deps)
		})
	})

	t.Run("success with cleanup error paths", func(t *testing.T) {
		t.Parallel()

		sqlDB, openErr := sql.Open("sqlite", "file:test-postgres-seam-db?mode=memory&cache=shared")
		require.NoError(t, openErr)

		deps := defaultDBDeps()
		deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
			return &postgresContainerModule.PostgresContainer{}, nil
		}
		deps.containerConnectionStringFn = func(context.Context, *postgresContainerModule.PostgresContainer) (string, error) {
			return testUnusedPostgresURL, nil
		}
		deps.openGormPostgresFn = func(string) (*gorm.DB, error) {
			return gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
		}
		deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
			return errors.New("terminate failure")
		}
		deps.closeSQLDBFn = func(*sql.DB) error {
			return errors.New("close failure")
		}

		db := newPostgresTestContainerWithDeps(context.Background(), t, deps)
		require.NotNil(t, db)
	})
}

func TestPublicWrappers_Coverage(t *testing.T) {
	t.Parallel()

	db, cleanup, err := NewInMemorySQLiteDBForTestMain()
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, cleanup)
	cleanup()

	require.NotNil(t, NewInMemorySQLiteDB(t))
	require.NotNil(t, NewClosedSQLiteDB(t, nil))

	t.Run("postgres wrapper skip-or-success", func(t *testing.T) {
		t.Parallel()
		_ = NewPostgresTestContainer(context.Background(), t)
	})
}

func TestDefaultSeamFunctions_Coverage(t *testing.T) {
	t.Parallel()

	deps := defaultDBDeps()
	sqlDB, err := deps.sqlOpenFn("sqlite", "file:test-default-seams?mode=memory&cache=shared")
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

		closedDB, openErr := sql.Open("sqlite", "file:test-default-seams-closed?mode=memory&cache=shared")
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
