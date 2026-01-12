// Copyright (c) 2025 Justin Cranford

package im

import (
	"context"
	"testing"
	"time"

	cipherIMRepository "cryptoutil/internal/apps/cipher/im/repository"
	serverTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"

	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for test-containers.
)

// TestInitDatabase_HappyPaths tests successful database initialization for PostgreSQL and SQLite.
func TestInitDatabase_HappyPaths(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, ctx context.Context) (tableCount int, cleanup func())
	}{
		{
			name: "PostgreSQL Container",
			setupFunc: func(t *testing.T, ctx context.Context) (int, func()) {
				t.Parallel()

				// Start PostgreSQL container with randomized credentials.
				sqlDB, closeDB, err := cryptoutilTemplateServerTestutil.NewInitializedPostgresTestDatabase(ctx, cipherIMRepository.MigrationsFS)
				require.NoError(t, err)

				// Verify schema migration (tables exist).
				var tableCount int

				err = sqlDB.QueryRowContext(ctx,
					"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'messages', 'messages_recipient_jwks')",
				).Scan(&tableCount)
				require.NoError(t, err)

				return tableCount, closeDB
			},
		},
		{
			name: "SQLite In-Memory",
			setupFunc: func(t *testing.T, ctx context.Context) (int, func()) {
				// No t.Parallel() - prevent cross-test pollution with shared in-memory SQLite.

				// Start SQLite in-memory database.
				sqlDB, err := cryptoutilTemplateServerTestutil.NewInitializedSQLiteTestDatabase(ctx, cipherIMRepository.MigrationsFS, true)
				require.NoError(t, err)

				// Verify schema migration (tables exist).
				var tableCount int

				err = sqlDB.QueryRowContext(ctx,
					"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'messages', 'messages_recipient_jwks')",
				).Scan(&tableCount)
				require.NoError(t, err)

				return tableCount, func() {}
			},
		},
		{
			name: "SQLite File-Based",
			setupFunc: func(t *testing.T, ctx context.Context) (int, func()) {
				// No t.Parallel() - SQLite file locking issues with concurrent tests.

				// Start SQLite file database.
				sqlDB, err := cryptoutilTemplateServerTestutil.NewInitializedSQLiteTestDatabase(ctx, cipherIMRepository.MigrationsFS, false)
				require.NoError(t, err)

				// Verify schema migration (tables exist).
				var tableCount int

				err = sqlDB.QueryRowContext(ctx,
					"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'messages', 'messages_recipient_jwks')",
				).Scan(&tableCount)
				require.NoError(t, err)

				// Close database before test cleanup (Windows file locking).
				require.NoError(t, sqlDB.Close())

				return tableCount, func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tableCount, cleanup := tt.setupFunc(t, ctx)
			defer cleanup()

			require.Equal(t, 3, tableCount, "Expected 3 tables (users, messages, messages_recipient_jwks) to be created")
		})
	}
}

// TestInitDatabase_ErrorPaths tests error handling for database initialization failures.
func TestInitDatabase_ErrorPaths(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(ctx context.Context) error
		expectedErrMsg string
	}{
		{
			name: "Invalid Database Type",
			setupFunc: func(ctx context.Context) error {
				// Initialize database (should fail with unsupported scheme error).
				gormDB, err := cryptoutilTemplateServerTestutil.InitDatabase(ctx, "mysql://user:pass@localhost:3306/dbname", cipherIMRepository.MigrationsFS)
				require.Nil(t, gormDB)

				return err
			},
			expectedErrMsg: "unsupported database URL scheme",
		},
		{
			name: "PostgreSQL Connection Error",
			setupFunc: func(ctx context.Context) error {
				// Use 1-second timeout for fast failure (was 5.4s with no timeout).
				ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()

				// Use invalid connection string (nonexistent server).
				gormDB, err := serverTemplateRepository.InitPostgreSQL(ctx, "postgres://user:pass@nonexistent:5432/dbname", cipherIMRepository.MigrationsFS)
				require.Nil(t, gormDB)

				return err
			},
			expectedErrMsg: "ping",
		},
		{
			name: "SQLite Invalid Path",
			setupFunc: func(ctx context.Context) error {
				// Use invalid file path (directory doesn't exist).
				gormDB, err := serverTemplateRepository.InitSQLite(ctx, "file:/nonexistent/invalid/path.db", cipherIMRepository.MigrationsFS)
				require.Nil(t, gormDB)

				return err
			},
			expectedErrMsg: "", // Error message varies by platform, just check it's an error.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := tt.setupFunc(ctx)

			require.Error(t, err)
			if tt.expectedErrMsg != "" {
				require.Contains(t, err.Error(), tt.expectedErrMsg)
			}
		})
	}
}
