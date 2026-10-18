// Copyright (c) 2025-2026 Justin Cranford.
//

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func TestStartPostgresSafely(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		startFn func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, string, string, string) (string, func(), error)
		wantURL string
		wantErr string
	}{
		{
			name: "success",
			startFn: func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, string, string, string) (string, func(), error) {
				return testPostgresLocalhostTest, func() {}, nil
			},
			wantURL: testPostgresLocalhostTest,
		},
		{
			name: "panic is recovered",
			startFn: func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, string, string, string) (string, func(), error) {
				panic("injected container panic")
			},
			wantErr: "panic while starting PostgreSQL testcontainer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			containerURL, cleanup, err := startPostgresSafely(context.Background(), &Basic{}, tc.startFn, "db", "user", "pass")

			require.Equal(t, tc.wantURL, containerURL)

			if tc.wantErr != "" {
				require.Nil(t, cleanup)
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, cleanup)
			cleanup()
		})
	}
}

func TestOpenSQLiteInternal_ExecErrors(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected SQLite PRAGMA failure")
	tests := []struct {
		name        string
		databaseURL string
		wantQuery   string
	}{
		{
			name:        "WAL mode",
			databaseURL: "file:disk.db",
			wantQuery:   "PRAGMA journal_mode=WAL;",
		},
		{
			name:        "busy timeout",
			databaseURL: cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
			wantQuery:   "PRAGMA busy_timeout = 30000;",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var sqlDB *sql.DB

			db, err := openSQLiteInternalWithExec(
				context.Background(),
				tc.databaseURL,
				false,
				func(string, string) (*sql.DB, error) {
					var openErr error

					sqlDB, openErr = sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)

					return sqlDB, openErr
				},
				func(gorm.Dialector, *gorm.Config) (*gorm.DB, error) {
					return nil, errors.New("GORM open must not be called")
				},
				func(_ context.Context, _ *sql.DB, query string) error {
					require.Equal(t, tc.wantQuery, query)

					return injectedErr
				},
			)

			if sqlDB != nil {
				t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
			}

			require.Nil(t, db)
			require.ErrorIs(t, err, injectedErr)
		})
	}
}

func TestOpenSQLiteInternal_GormDBError(t *testing.T) {
	t.Parallel()

	db, err := openSQLiteInternalWithExec(
		context.Background(),
		cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
		false,
		func(string, string) (*sql.DB, error) {
			return sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
		},
		func(gorm.Dialector, *gorm.Config) (*gorm.DB, error) {
			return &gorm.DB{Config: &gorm.Config{ConnPool: &applicationConnPool{}}}, nil
		},
		func(context.Context, *sql.DB, string) error {
			return nil
		},
	)

	require.Nil(t, db)
	require.ErrorContains(t, err, "failed to get database instance")
}

func TestOpenPostgreSQL_DBError(t *testing.T) {
	t.Parallel()

	db, err := openPostgreSQLInternal(
		testPostgresLocalhostTest,
		false,
		"",
		"",
		"",
		"",
		func(gorm.Dialector, *gorm.Config) (*gorm.DB, error) {
			return &gorm.DB{Config: &gorm.Config{ConnPool: &applicationConnPool{}}}, nil
		},
	)

	require.Nil(t, db)
	require.ErrorContains(t, err, "failed to get database instance")
}

func TestOpenPostgreSQLInternal_SSLConfiguration(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected PostgreSQL opener failure")
	tests := []struct {
		name            string
		databaseURL     string
		sslMode         string
		sslCert         string
		sslKey          string
		sslRootCert     string
		debugMode       bool
		wantDatabaseURL string
	}{
		{
			name:            "adds SSL options to URL without a query",
			databaseURL:     testPostgresLocalhostTest,
			sslMode:         "verify-full",
			sslCert:         "/certs/client.pem",
			sslKey:          "/certs/client-key.pem",
			sslRootCert:     "/certs/root.pem",
			debugMode:       true,
			wantDatabaseURL: testPostgresLocalhostTest + "?sslmode=verify-full&sslcert=/certs/client.pem&sslkey=/certs/client-key.pem&sslrootcert=/certs/root.pem",
		},
		{
			name:            "replaces SSL mode in URL with an existing query",
			databaseURL:     testPostgresLocalhostTest + "?application_name=framework&sslmode=disable",
			sslMode:         "require",
			wantDatabaseURL: testPostgresLocalhostTest + "?application_name=framework&sslmode=require",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, err := openPostgreSQLInternal(
				tc.databaseURL,
				tc.debugMode,
				tc.sslMode,
				tc.sslCert,
				tc.sslKey,
				tc.sslRootCert,
				func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
					postgresDialector, ok := dialector.(*postgres.Dialector)
					require.True(t, ok)
					require.Equal(t, cryptoutilSharedMagic.DockerServicePostgres, postgresDialector.Name())
					require.Equal(t, tc.wantDatabaseURL, postgresDialector.DSN)
					require.Equal(t, tc.debugMode, config.Logger != nil)
					require.True(t, config.SkipDefaultTransaction)

					return nil, injectedErr
				},
			)

			require.Nil(t, db)
			require.ErrorContains(t, err, "failed to open PostgreSQL database")
			require.ErrorIs(t, err, injectedErr)
		})
	}
}

func TestOpenPostgreSQLInternal_Success(t *testing.T) {
	t.Parallel()

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	db, err := openPostgreSQLInternal(
		testPostgresLocalhostTest,
		true,
		"",
		"",
		"",
		"",
		func(_ gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
			config.ConnPool = sqlDB

			return &gorm.DB{
				Config: config,
				Statement: &gorm.Statement{
					ConnPool: sqlDB,
				},
			}, nil
		},
	)

	require.NoError(t, err)
	require.NotNil(t, db)
	require.Equal(t, cryptoutilSharedMagic.PostgreSQLMaxOpenConns, sqlDB.Stats().MaxOpenConnections)
}

type applicationConnPool struct{}

func (*applicationConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, nil
}

func (*applicationConnPool) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, nil
}

func (*applicationConnPool) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, nil
}

func (*applicationConnPool) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil
}
