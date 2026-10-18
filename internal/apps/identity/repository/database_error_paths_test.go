// Copyright (c) 2025-2026 Justin Cranford.
//

package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	testUnsupportedDatabaseType           = "unsupported"
	testUnsupportedDatabaseTypeMessage    = "unsupported database type"
	testFailedToCreatePostgresDriverError = "failed to create postgres driver"
)

func TestInitializeDatabaseWithDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		databaseType  string
		openSQLError  string
		execErrorAt   int
		openGORMError string
		getDBError    string
		pingError     string
		maxOpenConns  int
		wantError     string
		wantAppError  error
	}{
		{
			name:         testUnsupportedDatabaseTypeMessage,
			databaseType: testUnsupportedDatabaseType,
			wantError:    testUnsupportedDatabaseTypeMessage,
			wantAppError: cryptoutilIdentityAppErr.ErrInvalidConfiguration,
		},
		{
			name:         "sqlite open failure",
			databaseType: dbTypeSQLite,
			openSQLError: "injected SQL open failure",
			wantError:    "failed to open SQLite database",
			wantAppError: cryptoutilIdentityAppErr.ErrDatabaseConnection,
		},
		{
			name:         "sqlite WAL failure",
			databaseType: dbTypeSQLite,
			execErrorAt:  1,
			wantError:    "failed to enable WAL mode",
			wantAppError: cryptoutilIdentityAppErr.ErrDatabaseConnection,
		},
		{
			name:         "sqlite busy timeout failure",
			databaseType: dbTypeSQLite,
			execErrorAt:  2,
			wantError:    "failed to set busy timeout",
			wantAppError: cryptoutilIdentityAppErr.ErrDatabaseConnection,
		},
		{
			name:          "GORM open failure",
			databaseType:  dbTypeSQLite,
			openGORMError: "injected GORM open failure",
			wantError:     "failed to connect to database",
			wantAppError:  cryptoutilIdentityAppErr.ErrDatabaseConnection,
		},
		{
			name:         "GORM database extraction failure",
			databaseType: dbTypeSQLite,
			getDBError:   "injected GORM database extraction failure",
			wantError:    "failed to get database instance",
			wantAppError: cryptoutilIdentityAppErr.ErrDatabaseConnection,
		},
		{
			name:         "database ping failure",
			databaseType: dbTypeSQLite,
			pingError:    "injected database ping failure",
			wantError:    "failed to ping database",
			wantAppError: cryptoutilIdentityAppErr.ErrDatabaseConnection,
		},
		{
			name:         "postgres connection pool configuration",
			databaseType: cryptoutilSharedMagic.DockerServicePostgres,
			maxOpenConns: 7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			connector := &databaseInitializationConnector{
				execErrorAt: tc.execErrorAt,
				pingError:   tc.pingError,
			}
			sqlDB := sql.OpenDB(connector)

			t.Cleanup(func() {
				_ = sqlDB.Close()
			})

			openSQL := func(string, string) (*sql.DB, error) {
				if tc.openSQLError != "" {
					return nil, errors.New(tc.openSQLError)
				}

				return sqlDB, nil
			}
			openGORM := func(gorm.Dialector, ...gorm.Option) (*gorm.DB, error) {
				if tc.openGORMError != "" {
					return nil, errors.New(tc.openGORMError)
				}

				return &gorm.DB{}, nil
			}
			getSQLDB := func(*gorm.DB) (*sql.DB, error) {
				if tc.getDBError != "" {
					return nil, errors.New(tc.getDBError)
				}

				return sqlDB, nil
			}

			cfg := &cryptoutilIdentityConfig.DatabaseConfig{
				Type:         tc.databaseType,
				DSN:          "test",
				MaxOpenConns: tc.maxOpenConns,
			}
			db, err := initializeDatabaseWithDependencies(context.Background(), cfg, openSQL, openGORM, getSQLDB)

			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				require.ErrorIs(t, err, tc.wantAppError)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, db)
			require.Equal(t, tc.maxOpenConns, sqlDB.Stats().MaxOpenConnections)
		})
	}
}

type databaseInitializationConnector struct {
	execErrorAt int
	pingError   string
	execCount   int
}

func (c *databaseInitializationConnector) Connect(context.Context) (driver.Conn, error) {
	return &databaseInitializationConnection{connector: c}, nil
}

func (c *databaseInitializationConnector) Driver() driver.Driver {
	return databaseInitializationDriver{connector: c}
}

type databaseInitializationDriver struct {
	connector *databaseInitializationConnector
}

func (d databaseInitializationDriver) Open(string) (driver.Conn, error) {
	return &databaseInitializationConnection{connector: d.connector}, nil
}

type databaseInitializationConnection struct {
	connector *databaseInitializationConnector
}

func (*databaseInitializationConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is unsupported")
}

func (*databaseInitializationConnection) Close() error {
	return nil
}

func (*databaseInitializationConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are unsupported")
}

func (c *databaseInitializationConnection) Ping(context.Context) error {
	if c.connector.pingError != "" {
		return errors.New(c.connector.pingError)
	}

	return nil
}

func (c *databaseInitializationConnection) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	c.connector.execCount++
	if c.connector.execErrorAt == c.connector.execCount {
		return nil, errors.New("injected SQL execution failure")
	}

	return driver.RowsAffected(0), nil
}

func (*databaseInitializationConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("queries are unsupported")
}
