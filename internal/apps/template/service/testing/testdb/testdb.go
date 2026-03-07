// Copyright (c) 2025 Justin Cranford
//

// Package testdb provides shared database test helpers for cryptoutil services.
// Centralizes test DB setup patterns to eliminate boilerplate across TestMain functions.
package testdb

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

// NewInMemorySQLiteDB creates a unique in-memory SQLite database configured for testing.
// Configures WAL mode, busy timeout, and connection pool.
// Registers cleanup via t.Cleanup() to close the database after the test.
func NewInMemorySQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	if err != nil {
		t.Fatalf("testdb: failed to generate UUID: %v", err)
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	db, sqlDB, err := buildInMemorySQLiteDB(context.Background(), sql.Open, dsn)
	if err != nil {
		t.Fatalf("testdb: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Logf("testdb: failed to close SQLite DB: %v", closeErr)
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
		t.Skipf("testdb: skipping PostgreSQL test - container unavailable: %v", err)

		return nil
	}

	t.Cleanup(func() {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("testdb: failed to terminate PostgreSQL container: %v", termErr)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("testdb: failed to get PostgreSQL connection string: %v", err)
	}

	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("testdb: failed to open GORM PostgreSQL DB: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, sqlErr := db.DB()
		if sqlErr != nil {
			t.Logf("testdb: failed to get sql.DB for cleanup: %v", sqlErr)

			return
		}

		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Logf("testdb: failed to close PostgreSQL DB: %v", closeErr)
		}
	})

	return db
}

// RequireNewInMemorySQLiteDB creates an in-memory SQLite DB and auto-migrates the given models.
// Convenience wrapper for tests that use GORM AutoMigrate.
func RequireNewInMemorySQLiteDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	db := NewInMemorySQLiteDB(t)

	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("testdb: failed to auto-migrate: %v", err)
		}
	}

	return db
}

// RequireNewPostgresTestContainer creates a PostgreSQL test container DB with auto-migrated models.
// Skips if Docker is unavailable. Registers cleanup automatically.
func RequireNewPostgresTestContainer(ctx context.Context, t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	db := NewPostgresTestContainer(ctx, t)
	if db == nil {
		return nil
	}

	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("testdb: failed to auto-migrate PostgreSQL: %v", err)
		}
	}

	return db
}

// FormatDSN formats a DSN string for PostgreSQL connections (exported for test use).
func FormatDSN(host, port, user, pass, dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbName)
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
