// Copyright (c) 2025 Justin Cranford
//
//

package bootstrap_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestResetDemoData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory database.
	cfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg)
	require.NoError(t, err, "Repository factory creation should succeed")

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	// Run migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "AutoMigrate should succeed")

	// Create demo data first.
	_, _, _, err = cryptoutilIdentityBootstrap.CreateDemoClient(ctx, repoFactory)
	require.NoError(t, err)

	_, _, _, err = cryptoutilIdentityBootstrap.CreateDemoUser(ctx, repoFactory)
	require.NoError(t, err)

	// Verify demo data exists.
	client, err := repoFactory.ClientRepository().GetByClientID(ctx, cryptoutilSharedMagic.DemoClientID)
	require.NoError(t, err)
	require.NotNil(t, client)

	user, err := repoFactory.UserRepository().GetBySub(ctx, "demo-user")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Reset demo data.
	err = cryptoutilIdentityBootstrap.ResetDemoData(ctx, repoFactory)
	require.NoError(t, err)

	// Verify demo data is deleted.
	_, err = repoFactory.ClientRepository().GetByClientID(ctx, cryptoutilSharedMagic.DemoClientID)
	require.Error(t, err, "Demo client should be deleted")

	_, err = repoFactory.UserRepository().GetBySub(ctx, "demo-user")
	require.Error(t, err, "Demo user should be deleted")
}

func TestResetDemoData_NoData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory database without seeding demo data.
	cfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg)
	require.NoError(t, err, "Repository factory creation should succeed")

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	// Run migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "AutoMigrate should succeed")

	// Reset should not error when there's no demo data.
	err = cryptoutilIdentityBootstrap.ResetDemoData(ctx, repoFactory)
	require.NoError(t, err)
}

func TestResetAndReseedDemo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory database.
	cfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg)
	require.NoError(t, err, "Repository factory creation should succeed")

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	// Run migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "AutoMigrate should succeed")

	// Create demo user first (ResetAndReseedDemo resets and recreates users).
	_, _, _, err = cryptoutilIdentityBootstrap.CreateDemoUser(ctx, repoFactory)
	require.NoError(t, err)

	// Reset and reseed.
	err = cryptoutilIdentityBootstrap.ResetAndReseedDemo(ctx, repoFactory)
	require.NoError(t, err)

	// Verify user was recreated.
	user, err := repoFactory.UserRepository().GetBySub(ctx, "demo-user")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "demo-user", user.Sub)
}
