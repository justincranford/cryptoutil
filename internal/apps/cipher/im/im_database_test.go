// Copyright (c) 2025 Justin Cranford

package im

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"testing"
	"time"

	cipherIMRepository "cryptoutil/internal/apps/cipher/im/repository"
	serverTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilContainer "cryptoutil/internal/shared/container"

	"github.com/stretchr/testify/require"

	googleUuid "github.com/google/uuid"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for test-containers.
)

// TestInitDatabase_PostgreSQL tests PostgreSQL database initialization using test-containers.
func TestInitDatabase_PostgreSQLContainer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start PostgreSQL postgresContainer with randomized credentials.
	sqlDB, closeDB, err := NewInitializedPostgresTestDatabase(ctx, cipherIMRepository.MigrationsFS)
	defer closeDB()

	// Verify schema migration (tables exist).
	var tableCount int

	err = sqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'messages', 'messages_recipient_jwks')",
	).Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 3, tableCount, "Expected 3 tables (users, messages, messages_recipient_jwks) to be created")
}

// TestInitDatabase_SQLite tests SQLite in-memory database initialization.
func TestInitDatabase_SQLiteInMemory(t *testing.T) {
	// Remove t.Parallel() - prevent cross-test pollution with shared in-memory SQLite.
	ctx := context.Background()

	// Start SQLite in-memory database with randomized credentials.
	sqlDB, err := NewInitializedSQLiteTestDatabase(ctx, cipherIMRepository.MigrationsFS, true)

	// Verify schema migration (tables exist).
	var tableCount int

	err = sqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'messages', 'messages_recipient_jwks')",
	).Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 3, tableCount, "Expected 3 tables (users, messages, messages_recipient_jwks) to be created")
}

// TestInitDatabase_SQLiteFile tests SQLite file-based database initialization.
func TestInitDatabase_SQLiteFile(t *testing.T) {
	// Remove t.Parallel() - SQLite file locking issues with concurrent tests.
	ctx := context.Background()

	// Start SQLite file database with randomized credentials.
	sqlDB, err := NewInitializedSQLiteTestDatabase(ctx, cipherIMRepository.MigrationsFS, false)

	// Verify schema migration (tables exist).
	var tableCount int

	err = sqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'messages', 'messages_recipient_jwks')",
	).Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 3, tableCount, "Expected 3 tables (users, messages, messages_recipient_jwks) to be created")

	// Close database before test cleanup (Windows file locking).
	require.NoError(t, sqlDB.Close())
}

func NewInitializedPostgresTestDatabase(ctx context.Context, migrationsFS embed.FS) (*sql.DB, func(), error) {
	postgresContainer, err := cryptoutilContainer.NewPostgresTestContainer(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create PostgreSQL container: %w", err)
	}

	closeDB := func() {
		err := postgresContainer.Terminate(ctx)
		if err != nil {
			fmt.Printf("failed to terminate PostgreSQL container: %v\n", err)
		}
	}

	databaseURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		closeDB()
		return nil, nil, fmt.Errorf("failed to get PostgreSQL connection string: %w", err)
	}

	// Initialize database.
	gormDB, err := InitDatabase(ctx, databaseURL, migrationsFS)
	if err != nil {
		closeDB()
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	} else if gormDB == nil {
		closeDB()
		return nil, nil, fmt.Errorf("gormDB must be non-nil: %w", err)
	}

	// Verify database connection.
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("sqlDB must be non-nill: %w", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return sqlDB, closeDB, nil
}

func NewInitializedSQLiteTestDatabase(ctx context.Context, migrationsFS embed.FS, inMemory bool) (*sql.DB, error) {
	var databaseURL string
	if inMemory {
		databaseURL = fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.NewString())
	} else {
		databaseURL = fmt.Sprintf("file:%s/test_%s.db?cache=shared", googleUuid.NewString(), googleUuid.NewString())
	}

	// Initialize unique in-memory SQLite database
	gormDB, err := InitDatabase(ctx, databaseURL, migrationsFS)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	} else if gormDB == nil {
		return nil, fmt.Errorf("gormDB must be non-nill: %w", err)
	}

	// Verify database connection.
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("sqlDB must be non-nill: %w", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return sqlDB, nil
}

// TestInitDatabase_InvalidDbType tests error handling for unsupported database URL schemes.
func TestInitDatabase_InvalidDbType(t *testing.T) {
	// Remove t.Parallel() - prevent cross-test pollution with shared in-memory SQLite.
	ctx := context.Background()

	// Initialize database (should fail with unsupported scheme error).
	gormDB, err := InitDatabase(ctx, "mysql://user:pass@localhost:3306/dbname", cipherIMRepository.MigrationsFS)
	require.Error(t, err)
	require.Nil(t, gormDB)
	require.Contains(t, err.Error(), "unsupported database URL scheme")
}

// TestInitPostgreSQL_ConnectionError tests PostgreSQL connection error handling.
func TestInitPostgreSQL_ConnectionError(t *testing.T) {
	// Remove t.Parallel() - prevent cross-test pollution.
	// Use 1-second timeout for fast failure (was 5.4s with no timeout).
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Use invalid connection string (nonexistent server).
	gormDB, err := serverTemplateRepository.InitPostgreSQL(ctx, "postgres://user:pass@nonexistent:5432/dbname", cipherIMRepository.MigrationsFS)
	require.Error(t, err)
	require.Nil(t, gormDB)
	require.Contains(t, err.Error(), "ping")
}

// TestInitSQLite_InvalidPath tests SQLite invalid path error handling.
func TestInitSQLite_InvalidPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use invalid file path (directory doesn't exist).
	gormDB, err := serverTemplateRepository.InitSQLite(ctx, "file:/nonexistent/invalid/path.db", cipherIMRepository.MigrationsFS)
	require.Error(t, err)
	require.Nil(t, gormDB)
}
