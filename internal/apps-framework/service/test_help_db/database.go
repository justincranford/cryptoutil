// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_db provides database fixture creation, schema migrations, and DB failure-path helpers
// for integration and E2E test suites. It handles SQLite in-memory setup, PostgreSQL containers (E2E only),
// and deterministic DB error creation for error-path testing.
//
// Consumed by:
//   - test_orch_integration: database fixture creation and migration
//   - test_orch_e2e: PostgreSQL test container setup
//   - Repository test suites: DB fixtures
package test_help_db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	postgresContainerModule "github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type dbDeps struct {
	newUUIDv7Fn                   func() (googleUuid.UUID, error)
	sqlOpenFn                     func(driver, dsn string) (*sql.DB, error)
	newPostgresContainerFactoryFn func(context.Context) (*postgresContainerModule.PostgresContainer, error)
	execSQLitePragmaFn            func(context.Context, *sql.DB, string) error
	openGormSQLiteFn              func(*sql.DB) (*gorm.DB, error)
	openGormPostgresFn            func(string) (*gorm.DB, error)
	getSQLDBFn                    func(*gorm.DB) (*sql.DB, error)
	containerConnectionStringFn   func(context.Context, *postgresContainerModule.PostgresContainer) (string, error)
	containerTerminateFn          func(context.Context, *postgresContainerModule.PostgresContainer) error
	closeSQLDBFn                  func(*sql.DB) error
}

func defaultExecSQLitePragma(ctx context.Context, sqlDB *sql.DB, query string) error {
	_, err := sqlDB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("ExecContext: %w", err)
	}

	return nil
}

func defaultOpenGormSQLite(sqlDB *sql.DB) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})

	return db, wrapIfErr("gorm.Open sqlite", err)
}

func defaultOpenGormPostgres(connStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})

	return db, wrapIfErr("gorm.Open postgres", err)
}

func defaultGetSQLDB(db *gorm.DB) (*sql.DB, error) {
	sqlDB, err := db.DB()

	return sqlDB, wrapIfErr("db.DB", err)
}

func defaultContainerConnectionString(ctx context.Context, container *postgresContainerModule.PostgresContainer) (string, error) {
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")

	return connStr, wrapIfErr("ConnectionString", err)
}

func defaultContainerTerminate(ctx context.Context, container *postgresContainerModule.PostgresContainer) error {
	err := container.Terminate(ctx)

	return wrapIfErr("terminate", err)
}

func defaultCloseSQLDB(sqlDB *sql.DB) error {
	err := sqlDB.Close()

	return wrapIfErr("close", err)
}

func wrapIfErr(op string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", op, err)
}

func defaultDBDeps() dbDeps {
	return dbDeps{
		newUUIDv7Fn:                   googleUuid.NewV7,
		sqlOpenFn:                     sql.Open,
		newPostgresContainerFactoryFn: cryptoutilSharedContainer.NewPostgresTestContainer,
		execSQLitePragmaFn:            defaultExecSQLitePragma,
		openGormSQLiteFn:              defaultOpenGormSQLite,
		openGormPostgresFn:            defaultOpenGormPostgres,
		getSQLDBFn:                    defaultGetSQLDB,
		containerConnectionStringFn:   defaultContainerConnectionString,
		containerTerminateFn:          defaultContainerTerminate,
		closeSQLDBFn:                  defaultCloseSQLDB,
	}
}

// NewInMemorySQLiteDBForTestMain creates a unique in-memory SQLite database for use in TestMain functions.
// Configures WAL mode, busy timeout, and connection pool.
// Returns the db, a cleanup function that reports close errors, and any setup error.
// Unlike NewInMemorySQLiteDB, this function does not require a *testing.T.
func NewInMemorySQLiteDBForTestMain() (*gorm.DB, func() error, error) {
	return newInMemorySQLiteDBForTestMainWithDeps(defaultDBDeps())
}

func newInMemorySQLiteDBForTestMainWithDeps(deps dbDeps) (*gorm.DB, func() error, error) {
	dbID, err := deps.newUUIDv7Fn()
	if err != nil {
		return nil, nil, fmt.Errorf("test_help_db: failed to generate UUID: %w", err)
	}

	dsn := fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, dbID.String())

	db, sqlDB, err := buildInMemorySQLiteDB(context.Background(), deps, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("test_help_db: %w", err)
	}

	cleanup := func() error {
		if closeErr := deps.closeSQLDBFn(sqlDB); closeErr != nil {
			return fmt.Errorf("close SQLite TestMain database: %w", closeErr)
		}

		return nil
	}

	return db, cleanup, nil
}

// NewInMemorySQLiteDB creates a unique in-memory SQLite database configured for testing.
// Configures WAL mode, busy timeout, and connection pool.
// Registers cleanup via t.Cleanup() to close the database after the test.
func NewInMemorySQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	return newInMemorySQLiteDBWithDeps(t, defaultDBDeps())
}

func newInMemorySQLiteDBWithDeps(t *testing.T, deps dbDeps) *gorm.DB {
	t.Helper()

	dbID, err := deps.newUUIDv7Fn()
	if err != nil {
		panic(fmt.Sprintf("test_help_db: failed to generate UUID: %v", err))
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	db, sqlDB, err := buildInMemorySQLiteDB(context.Background(), deps, dsn)
	if err != nil {
		panic(fmt.Sprintf("test_help_db: %v", err))
	}

	t.Cleanup(func() {
		if closeErr := deps.closeSQLDBFn(sqlDB); closeErr != nil {
			t.Logf("test_help_db: failed to close SQLite DB: %v", closeErr)
		}
	})

	return db
}

// buildInMemorySQLiteDB constructs a SQLite GORM database from a DSN with WAL mode configured.
func buildInMemorySQLiteDB(ctx context.Context, deps dbDeps, dsn string) (*gorm.DB, *sql.DB, error) {
	sqlDB, err := deps.sqlOpenFn(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err = deps.execSQLitePragmaFn(ctx, sqlDB, "PRAGMA journal_mode=WAL;"); err != nil {
		return nil, nil, closeSQLDBAfterSetupFailure(deps, sqlDB, fmt.Errorf("WAL pragma: %w", err))
	}

	if err = deps.execSQLitePragmaFn(ctx, sqlDB, "PRAGMA busy_timeout = 30000;"); err != nil {
		return nil, nil, closeSQLDBAfterSetupFailure(deps, sqlDB, fmt.Errorf("busy_timeout pragma: %w", err))
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)
	sqlDB.SetConnMaxLifetime(0)

	db, err := deps.openGormSQLiteFn(sqlDB)
	if err != nil {
		return nil, nil, closeSQLDBAfterSetupFailure(deps, sqlDB, fmt.Errorf("gorm.Open: %w", err))
	}

	return db, sqlDB, nil
}

func closeSQLDBAfterSetupFailure(deps dbDeps, sqlDB *sql.DB, cause error) error {
	if closeErr := deps.closeSQLDBFn(sqlDB); closeErr != nil {
		return errors.Join(cause, fmt.Errorf("close SQLite DB after setup failure: %w", closeErr))
	}

	return cause
}

// NewPostgresTestContainer creates a PostgreSQL test container and returns a configured GORM DB.
// Requires Docker to be running. Fails the test if the container cannot be started.
// Registers cleanup via t.Cleanup() to terminate the container and close the DB after the test.
func NewPostgresTestContainer(ctx context.Context, t *testing.T) *gorm.DB {
	t.Helper()

	db, cleanup, err := newPostgresTestContainerWithDeps(ctx, defaultDBDeps())
	if err != nil {
		t.Fatalf("test_help_db: failed to create PostgreSQL test container: %v", err)

		return nil
	}

	t.Cleanup(func() {
		if cleanupErr := cleanup(); cleanupErr != nil {
			t.Errorf("test_help_db: failed to clean up PostgreSQL test container: %v", cleanupErr)
		}
	})

	return db
}

func newPostgresTestContainerWithDeps(ctx context.Context, deps dbDeps) (*gorm.DB, func() error, error) {
	if ctx == nil {
		return nil, nil, errors.New("context cannot be nil")
	}

	container, err := safeNewPostgresTestContainer(ctx, deps.newPostgresContainerFactoryFn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	if container == nil {
		return nil, nil, errors.New("PostgreSQL container factory returned nil container")
	}

	terminateAfterSetupFailure := func(cause error) error {
		if terminateErr := deps.containerTerminateFn(ctx, container); terminateErr != nil {
			return errors.Join(cause, fmt.Errorf("failed to terminate PostgreSQL container after setup failure: %w", terminateErr))
		}

		return cause
	}

	connStr, err := deps.containerConnectionStringFn(ctx, container)
	if err != nil {
		return nil, nil, terminateAfterSetupFailure(fmt.Errorf("failed to get PostgreSQL connection string: %w", err))
	}

	db, err := deps.openGormPostgresFn(connStr)
	if err != nil {
		return nil, nil, terminateAfterSetupFailure(fmt.Errorf("failed to open GORM PostgreSQL DB: %w", err))
	}

	if db == nil {
		return nil, nil, terminateAfterSetupFailure(errors.New("GORM PostgreSQL opener returned nil database"))
	}

	cleanup := func() error {
		var cleanupErr error

		sqlDB, sqlErr := deps.getSQLDBFn(db)
		if sqlErr != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("failed to get sql.DB for cleanup: %w", sqlErr))
		} else if sqlDB == nil {
			cleanupErr = errors.Join(cleanupErr, errors.New("GORM PostgreSQL database returned nil sql.DB"))
		} else if closeErr := deps.closeSQLDBFn(sqlDB); closeErr != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("failed to close PostgreSQL DB: %w", closeErr))
		}

		if terminateErr := deps.containerTerminateFn(ctx, container); terminateErr != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("failed to terminate PostgreSQL container: %w", terminateErr))
		}

		return cleanupErr
	}

	return db, cleanup, nil
}

// NewClosedSQLiteDB creates an in-memory SQLite DB, applies optional migrations,
// then closes the underlying connection before returning.
// All subsequent GORM operations on the returned DB will fail.
// Used to test repository and service database error paths without hand-rolled setup.
func NewClosedSQLiteDB(t *testing.T, applyMigrations func(*sql.DB) error) *gorm.DB {
	t.Helper()

	return newClosedSQLiteDBWithDeps(t, applyMigrations, defaultDBDeps())
}

func newClosedSQLiteDBWithDeps(t *testing.T, applyMigrations func(*sql.DB) error, deps dbDeps) *gorm.DB {
	t.Helper()

	dbID, err := deps.newUUIDv7Fn()
	if err != nil {
		panic(fmt.Sprintf("test_help_db: uuid: %v", err))
	}

	dsn := fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, dbID.String())

	db, err := buildClosedSQLiteDB(context.Background(), deps, dsn, applyMigrations)
	if err != nil {
		panic(fmt.Sprintf("test_help_db: %v", err))
	}

	return db
}

// buildClosedSQLiteDB creates then closes a DB for error-path testing.
func buildClosedSQLiteDB(ctx context.Context, deps dbDeps, dsn string, applyMigrations func(*sql.DB) error) (*gorm.DB, error) {
	sqlDB, err := deps.sqlOpenFn(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err = deps.execSQLitePragmaFn(ctx, sqlDB, "PRAGMA journal_mode=WAL;"); err != nil {
		return nil, closeSQLDBAfterSetupFailure(deps, sqlDB, fmt.Errorf("WAL pragma: %w", err))
	}

	db, err := deps.openGormSQLiteFn(sqlDB)
	if err != nil {
		return nil, closeSQLDBAfterSetupFailure(deps, sqlDB, fmt.Errorf("gorm.Open: %w", err))
	}

	if applyMigrations != nil {
		if err = applyMigrations(sqlDB); err != nil {
			return nil, closeSQLDBAfterSetupFailure(deps, sqlDB, fmt.Errorf("apply migrations: %w", err))
		}
	}

	if closeErr := deps.closeSQLDBFn(sqlDB); closeErr != nil {
		return nil, fmt.Errorf("close DB after migrations: %w", closeErr)
	}

	return db, nil
}

// safeNewPostgresTestContainer wraps container factory creation to recover from panics
// that occur when Docker is unavailable (testcontainers panics on Windows with no Docker).
func safeNewPostgresTestContainer(ctx context.Context, factoryFn func(context.Context) (*postgresContainerModule.PostgresContainer, error)) (c *postgresContainerModule.PostgresContainer, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("docker unavailable (panic): %v", r)
		}
	}()

	c, err := factoryFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres container: %w", err)
	}

	return c, nil
}
