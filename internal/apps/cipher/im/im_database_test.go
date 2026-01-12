// Copyright (c) 2025 Justin Cranford

package im

import (
	"context"
	"fmt"
	"os"
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
func TestInitDatabase_PostgreSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start PostgreSQL postgresContainer with randomized credentials.
	postgresContainer, err := cryptoutilContainer.NewPostgresTestContainer(ctx)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, postgresContainer.Terminate(ctx))
	}()

	databaseURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Initialize database.
	db, err := initDatabase(ctx, databaseURL)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify database connection.
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.PingContext(ctx)
	require.NoError(t, err)

	// Verify schema migration (tables exist).
	var tableCount int

	err = sqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'messages', 'messages_recipient_jwks')",
	).Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 3, tableCount, "Expected 3 tables (users, messages, messages_recipient_jwks) to be created")
}

// TestInitDatabase_SQLite tests SQLite in-memory database initialization.
func TestInitDatabase_SQLite(t *testing.T) {
	// Remove t.Parallel() - prevent cross-test pollution with shared in-memory SQLite.
	ctx := context.Background()

	// Use unique in-memory database URL to prevent conflicts between parallel tests.
	originalEnv := os.Getenv("DATABASE_URL")
	uniqueDB := fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.NewString())
	require.NoError(t, os.Setenv("DATABASE_URL", uniqueDB))

	defer func() { _ = os.Setenv("DATABASE_URL", originalEnv) }()

	// Initialize database (should use unique in-memory database).
	db, err := initDatabase(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify database connection.
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.PingContext(ctx)
	require.NoError(t, err)

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

	// Create temporary database file path with unique name.
	tmpFile := fmt.Sprintf("file:%s/test_%s.db?cache=shared", t.TempDir(), googleUuid.NewString())

	// Set environment variable for database URL.
	originalEnv := os.Getenv("DATABASE_URL")

	require.NoError(t, os.Setenv("DATABASE_URL", tmpFile))

	defer func() { _ = os.Setenv("DATABASE_URL", originalEnv) }()

	// Initialize database.
	db, err := initDatabase(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify database connection.
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.PingContext(ctx)
	require.NoError(t, err)

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

// TestInitDatabase_InvalidScheme tests error handling for unsupported database URL schemes.
func TestInitDatabase_InvalidScheme(t *testing.T) {
	// Remove t.Parallel() - prevent cross-test pollution with shared in-memory SQLite.
	ctx := context.Background()

	// Set environment variable for invalid database URL.
	originalEnv := os.Getenv("DATABASE_URL")

	require.NoError(t, os.Setenv("DATABASE_URL", "mysql://user:pass@localhost:3306/dbname"))

	defer func() { _ = os.Setenv("DATABASE_URL", originalEnv) }()

	// Initialize database (should fail with unsupported scheme error).
	db, err := initDatabase(ctx, "")
	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "unsupported database URL scheme")
}

// TestInitPostgreSQL_ConnectionError tests PostgreSQL connection error handling.
func TestInitPostgreSQL_ConnectionError(t *testing.T) {
	// Remove t.Parallel() - prevent cross-test pollution.
	// Use 1-second timeout for fast failure (was 5.4s with no timeout).
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Use invalid connection string (nonexistent server).
	db, err := serverTemplateRepository.InitPostgreSQL(ctx, "postgres://user:pass@nonexistent:5432/dbname", cipherIMRepository.MigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "ping")
}

// TestInitSQLite_InvalidPath tests SQLite invalid path error handling.
func TestInitSQLite_InvalidPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use invalid file path (directory doesn't exist).
	db, err := initSQLite(ctx, "file:/nonexistent/invalid/path.db")
	require.Error(t, err)
	require.Nil(t, db)
}
