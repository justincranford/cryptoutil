// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"

	cryptoutilSharedContainer "cryptoutil/internal/shared/container"

	"github.com/stretchr/testify/require"
)

// TestInitPostgreSQL_Success verifies that InitPostgreSQL succeeds with a live
// PostgreSQL container, covering the full happy path including GORM init,
// connection pool configuration, and migration application.
func TestInitPostgreSQL_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	postgresContainer, err := cryptoutilSharedContainer.NewPostgresTestContainer(ctx)
	require.NoError(t, err)

	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	databaseURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := InitPostgreSQL(ctx, databaseURL, testSuccessMigrationsFS)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	defer func() {
		_ = sqlDB.Close()
	}()
}
