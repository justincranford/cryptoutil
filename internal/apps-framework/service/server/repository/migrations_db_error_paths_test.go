// Copyright (c) 2025-2026 Justin Cranford.
package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io/fs"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/stretchr/testify/require"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	testPostgresURL                 = "postgres://test.invalid:5432/test"
	testFailedToOpenSQLiteDatabase  = "failed to open SQLite database"
	testFailedToApplyMigrations     = "failed to apply migrations"
	testFailedToGetDatabaseInstance = "failed to get database instance"
)

type initializationTestConnector struct {
	connection *initializationTestConnection
}

func (c initializationTestConnector) Connect(context.Context) (driver.Conn, error) {
	return c.connection, nil
}

func (c initializationTestConnector) Driver() driver.Driver {
	return initializationTestDriver(c)
}

type initializationTestDriver struct {
	connection *initializationTestConnection
}

func (d initializationTestDriver) Open(string) (driver.Conn, error) {
	return d.connection, nil
}

type initializationTestConnection struct {
	pingErr   error
	execErrAt int
	execErr   error
	closeErr  error
	execCalls int
}

type initializationTestDBConnector struct{}

func (initializationTestDBConnector) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}

func (initializationTestDBConnector) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, errors.New("unexpected exec")
}

func (initializationTestDBConnector) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (initializationTestDBConnector) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil
}

func (initializationTestDBConnector) GetDBConn() (*sql.DB, error) {
	return nil, errors.New("injected database handle error")
}

func (c *initializationTestConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}

func (c *initializationTestConnection) Close() error {
	return c.closeErr
}

func (c *initializationTestConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected begin")
}

func (c *initializationTestConnection) Ping(context.Context) error {
	return c.pingErr
}

func (c *initializationTestConnection) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	c.execCalls++
	if c.execCalls == c.execErrAt {
		return nil, c.execErr
	}

	return driver.RowsAffected(0), nil
}

func newInitializationTestSQLDB(connection *initializationTestConnection) *sql.DB {
	return sql.OpenDB(initializationTestConnector{connection: connection})
}

func TestDatabaseInitialization_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configure  func(*databaseInitializationOperations)
		initialize func(databaseInitializationOperations) (*gorm.DB, error)
		wantErr    string
	}{
		{
			name: "SQLite SQL open error",
			configure: func(operations *databaseInitializationOperations) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					return nil, errors.New("injected SQL open error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: testFailedToOpenSQLiteDatabase,
		},
		{
			name: "SQLite GORM open error",
			configure: func(operations *databaseInitializationOperations) {
				operations.openGORM = func(gorm.Dialector, ...gorm.Option) (*gorm.DB, error) {
					return nil, errors.New("injected GORM open error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: "failed to initialize GORM for SQLite",
		},
		{
			name: "SQLite migration error",
			configure: func(operations *databaseInitializationOperations) {
				operations.applyMigrations = func(*sql.DB, DatabaseType, fs.FS) error {
					return errors.New("injected migration error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: testFailedToApplyMigrations,
		},
		{
			name: "PostgreSQL SQL open error",
			configure: func(operations *databaseInitializationOperations) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					return nil, errors.New("injected SQL open error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initPostgreSQLWithOperations(context.Background(), "postgres://invalid:5432/test", testDBMigrationsFS, operations)
			},
			wantErr: "failed to open PostgreSQL database",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := defaultDatabaseInitializationOperations()
			tc.configure(&operations)

			db, err := tc.initialize(operations)
			require.ErrorContains(t, err, tc.wantErr)
			require.Nil(t, db)
		})
	}
}

func TestDatabaseInitialization_InjectedFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		initialize func(databaseInitializationOperations) (*gorm.DB, error)
		configure  func(*databaseInitializationOperations, **sql.DB)
		wantErr    string
	}{
		{
			name: "PostgreSQL ping failure",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					*sqlDB = newInitializationTestSQLDB(&initializationTestConnection{pingErr: errors.New("injected ping error")})

					return *sqlDB, nil
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initPostgreSQLWithOperations(context.Background(), testPostgresURL, testDBMigrationsFS, operations)
			},
			wantErr: "failed to ping PostgreSQL database",
		},
		{
			name: "PostgreSQL GORM failure",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					*sqlDB = newInitializationTestSQLDB(&initializationTestConnection{})

					return *sqlDB, nil
				}
				operations.openGORM = func(gorm.Dialector, ...gorm.Option) (*gorm.DB, error) {
					return nil, errors.New("injected GORM error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initPostgreSQLWithOperations(context.Background(), testPostgresURL, testDBMigrationsFS, operations)
			},
			wantErr: "failed to initialize GORM for PostgreSQL",
		},
		{
			name: "PostgreSQL database handle failure",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					*sqlDB = newInitializationTestSQLDB(&initializationTestConnection{})

					return *sqlDB, nil
				}
				operations.openGORM = func(gorm.Dialector, ...gorm.Option) (*gorm.DB, error) {
					return &gorm.DB{Config: &gorm.Config{ConnPool: initializationTestDBConnector{}}}, nil
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initPostgreSQLWithOperations(context.Background(), testPostgresURL, testDBMigrationsFS, operations)
			},
			wantErr: testFailedToGetDatabaseInstance,
		},
		{
			name: "PostgreSQL migration failure",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					var err error

					*sqlDB, err = sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)

					return *sqlDB, err
				}
				operations.openGORM = func(_ gorm.Dialector, options ...gorm.Option) (*gorm.DB, error) {
					return gorm.Open(sqliteTestDialector(*sqlDB), options...)
				}
				operations.applyMigrations = func(*sql.DB, DatabaseType, fs.FS) error {
					return errors.New("injected PostgreSQL migration error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initPostgreSQLWithOperations(context.Background(), testPostgresURL, testDBMigrationsFS, operations)
			},
			wantErr: testFailedToApplyMigrations,
		},
		{
			name: "SQLite WAL pragma failure",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					*sqlDB = newInitializationTestSQLDB(&initializationTestConnection{
						execErrAt: 1,
						execErr:   errors.New("injected WAL error"),
					})

					return *sqlDB, nil
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: "failed to enable WAL mode",
		},
		{
			name: "SQLite busy timeout pragma failure",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					*sqlDB = newInitializationTestSQLDB(&initializationTestConnection{
						execErrAt: 2,
						execErr:   errors.New("injected timeout error"),
					})

					return *sqlDB, nil
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: "failed to set busy timeout",
		},
		{
			name: "SQLite database handle failure",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					connection := &initializationTestConnection{}
					*sqlDB = newInitializationTestSQLDB(connection)

					return *sqlDB, nil
				}
				operations.openGORM = func(gorm.Dialector, ...gorm.Option) (*gorm.DB, error) {
					return &gorm.DB{Config: &gorm.Config{ConnPool: initializationTestDBConnector{}}}, nil
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: testFailedToGetDatabaseInstance,
		},
		{
			name: "SQLite open returns database with error",
			configure: func(operations *databaseInitializationOperations, sqlDB **sql.DB) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					*sqlDB = newInitializationTestSQLDB(&initializationTestConnection{})

					return *sqlDB, errors.New("injected open error after allocation")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: testFailedToOpenSQLiteDatabase,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := defaultDatabaseInitializationOperations()

			var sqlDB *sql.DB
			tc.configure(&operations, &sqlDB)

			db, err := tc.initialize(operations)
			require.Nil(t, db)
			require.ErrorContains(t, err, tc.wantErr)

			if sqlDB != nil {
				require.NoError(t, sqlDB.Close())
			}
		})
	}
}

func sqliteTestDialector(db *sql.DB) gorm.Dialector {
	return gormsqlite.Dialector{Conn: db}
}

func TestCloseSQLDBAfterInitializationError_CloseFailure(t *testing.T) {
	t.Parallel()

	connection := &initializationTestConnection{closeErr: errors.New("injected close failure")}
	sqlDB := newInitializationTestSQLDB(connection)
	require.NoError(t, sqlDB.Ping())

	err := closeSQLDBAfterInitializationError(sqlDB, errors.New("initialization failed"))
	require.ErrorContains(t, err, "initialization failed")
	require.ErrorContains(t, err, "failed to close database")
	require.ErrorContains(t, err, "injected close failure")
}

// TestMigrationRunner_MigrateNewWithInstanceError covers the migrate.NewWithInstance error path.
func TestMigrationRunner_MigrateNewWithInstanceError(t *testing.T) {
	t.Parallel()

	// Open a real SQLite database so that iofs and sqlite driver creation succeed,
	// then fail at the migrate.NewWithInstance step.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, sqlDB.Close())
	}()

	stubMigrateNewInstanceFn := func(_ string, _ source.Driver, _ string, _ database.Driver) (*migrate.Migrate, error) {
		return nil, errors.New("injected migrate instance error")
	}

	runner := newMigrationRunnerInternal(testDBMigrationsFS, "test_migrations", stubMigrateNewInstanceFn)
	applyErr := runner.Apply(sqlDB, DatabaseTypeSQLite)
	require.Error(t, applyErr)
	require.Contains(t, applyErr.Error(), "failed to create migrate instance")
}
