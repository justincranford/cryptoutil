// Copyright (c) 2025 Justin Cranford
//
//

package bootstrap_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityBootstrap "cryptoutil/internal/apps/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestCreateDemoClient_NewClient(t *testing.T) {
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

	// Create demo client (should succeed).
	clientID, secret, created, err := cryptoutilIdentityBootstrap.CreateDemoClient(ctx, repoFactory)

	require.NoError(t, err, "CreateDemoClient should succeed")
	require.True(t, created, "Client should be created")
	require.Equal(t, cryptoutilSharedMagic.DemoClientID, clientID, "Client ID should be demo-client")
	require.Equal(t, cryptoutilSharedMagic.DemoClientSecret, secret, "Secret should be demo-secret")
}

func TestCreateDemoClient_ExistingClient(t *testing.T) {
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

	// Create demo client first time.
	_, _, created1, err := cryptoutilIdentityBootstrap.CreateDemoClient(ctx, repoFactory)
	require.NoError(t, err, "First CreateDemoClient should succeed")
	require.True(t, created1, "First call should create client")

	// Try to create again (should return existing, no secret).
	clientID, secret, created2, err := cryptoutilIdentityBootstrap.CreateDemoClient(ctx, repoFactory)
	require.NoError(t, err, "Second CreateDemoClient should succeed")
	require.False(t, created2, "Second call should not create client")
	require.Equal(t, cryptoutilSharedMagic.DemoClientID, clientID, "Client ID should be demo-client")
	require.Empty(t, secret, "Secret should be empty for existing client")
}

func TestBootstrapClients(t *testing.T) {
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

	// Create config.
	config := cryptoutilIdentityConfig.DefaultConfig()

	// Bootstrap clients.
	err = cryptoutilIdentityBootstrap.BootstrapClients(ctx, config, repoFactory)
	require.NoError(t, err, "BootstrapClients should succeed")

	// Verify demo-client exists.
	clientRepo := repoFactory.ClientRepository()
	client, err := clientRepo.GetByClientID(ctx, cryptoutilSharedMagic.DemoClientID)
	require.NoError(t, err, "GetByClientID should succeed")
	require.NotNil(t, client, "Demo client should exist")
	require.Equal(t, cryptoutilSharedMagic.DemoClientID, client.ClientID, "Client ID should match")
	require.Equal(t, cryptoutilSharedMagic.DemoClientName, client.Name, "Client name should match")
	require.NotNil(t, client.Enabled, "Enabled field should not be nil")
	require.True(t, *client.Enabled, "Client should be enabled")
}
