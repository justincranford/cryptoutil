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

// NewInMemorySQLiteDBForTestMain creates a unique in-memory SQLite database for use in TestMain functions.
// Configures WAL mode, busy timeout, and connection pool.
// Returns the db, a cleanup function, and any error.
// Unlike NewInMemorySQLiteDB, this function does not require a *testing.T.
func NewInMemorySQLiteDBForTestMain() (*gorm.DB, func(), error) {
	dbID, err := googleUuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("test_help_db: failed to generate UUID: %w", err)
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	db, sqlDB, err := buildInMemorySQLiteDB(context.Background(), sql.Open, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("test_help_db: %w", err)
	}

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return db, cleanup, nil
}

// NewInMemorySQLiteDB creates a unique in-memory SQLite database configured for testing.
// Configures WAL mode, busy timeout, and connection pool.
// Registers cleanup via t.Cleanup() to close the database after the test.
func NewInMemorySQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	if err != nil {
		t.Fatalf("test_help_db: failed to generate UUID: %v", err)
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	db, sqlDB, err := buildInMemorySQLiteDB(context.Background(), sql.Open, dsn)
	if err != nil {
		t.Fatalf("test_help_db: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Logf("test_help_db: failed to close SQLite DB: %v", closeErr)
		}
	})

	return db
}

// buildInMemorySQLiteDB constructs a SQLite GORM database from a DSN with WAL mode configured.
// openFn is injected to allow testing of all code paths including error scenarios.
func buildInMemorySQLiteDB(ctx context.Context, openFn func(driver, dsn string) (*sql.DB, error), dsn string) (*gorm.DB, *sql.DB, error) {
	sqlDB, err := openFn(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("sql.Open: %w", err)
	}

	if _, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("WAL pragma: %w", err)
	}

	if _, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("busy_timeout pragma: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)
	sqlDB.SetConnMaxLifetime(0)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("gorm.Open: %w", err)
	}

	return db, sqlDB, nil
}

// NewPostgresTestContainer creates a PostgreSQL test container and returns a configured GORM DB.
// Requires Docker to be running. Skips the test if the container fails to start.
// Registers cleanup via t.Cleanup() to terminate the container and close the DB after the test.
func NewPostgresTestContainer(ctx context.Context, t *testing.T) *gorm.DB {
	t.Helper()

	container, err := safeNewPostgresTestContainer(ctx)
	if err != nil {
		t.Skipf("test_help_db: skipping PostgreSQL test - container unavailable: %v", err)

		return nil
	}

	t.Cleanup(func() {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("test_help_db: failed to terminate PostgreSQL container: %v", termErr)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("test_help_db: failed to get PostgreSQL connection string: %v", err)
	}

	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("test_help_db: failed to open GORM PostgreSQL DB: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, sqlErr := db.DB()
		if sqlErr != nil {
			t.Logf("test_help_db: failed to get sql.DB for cleanup: %v", sqlErr)

			return
		}

		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Logf("test_help_db: failed to close PostgreSQL DB: %v", closeErr)
		}
	})

	return db
}

// NewClosedSQLiteDB creates an in-memory SQLite DB, applies optional migrations,
// then closes the underlying connection before returning.
// All subsequent GORM operations on the returned DB will fail.
// Used to test repository and service database error paths without hand-rolled setup.
func NewClosedSQLiteDB(t *testing.T, applyMigrations func(*sql.DB) error) *gorm.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	if err != nil {
		t.Fatalf("test_help_db: uuid: %v", err)
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	db, err := buildClosedSQLiteDB(context.Background(), sql.Open, dsn, applyMigrations)
	if err != nil {
		t.Fatalf("test_help_db: %v", err)
	}

	return db
}

// buildClosedSQLiteDB creates then closes a DB for error-path testing.
func buildClosedSQLiteDB(ctx context.Context, openFn func(driver, dsn string) (*sql.DB, error), dsn string, applyMigrations func(*sql.DB) error) (*gorm.DB, error) {
	sqlDB, err := openFn(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	if _, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("WAL pragma: %w", err)
	}

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	if applyMigrations != nil {
		if err = applyMigrations(sqlDB); err != nil {
			_ = sqlDB.Close()

			return nil, fmt.Errorf("apply migrations: %w", err)
		}
	}

	if closeErr := sqlDB.Close(); closeErr != nil {
		return nil, fmt.Errorf("close DB after migrations: %w", closeErr)
	}

	return db, nil
}

// safeNewPostgresTestContainer wraps NewPostgresTestContainer to recover from panics
// that occur when Docker is unavailable (testcontainers panics on Windows with no Docker).
func safeNewPostgresTestContainer(ctx context.Context) (c *postgresContainerModule.PostgresContainer, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("docker unavailable (panic): %v", r)
		}
	}()

	c, err := cryptoutilSharedContainer.NewPostgresTestContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres container: %w", err)
	}

	return c, nil
}
