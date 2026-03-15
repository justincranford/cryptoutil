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

// TestCreateDemoUser_NewUser tests CreateDemoUser when user doesn't exist.
func TestCreateDemoUser_NewUser(t *testing.T) {
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

	// Create demo user (should succeed).
	sub, password, created, err := cryptoutilIdentityBootstrap.CreateDemoUser(ctx, repoFactory)

	require.NoError(t, err, "CreateDemoUser should succeed")
	require.True(t, created, "User should be created")
	require.Equal(t, "demo-user", sub, "Sub should be demo-user")
	require.Equal(t, "demo-password", password, "Password should be demo-password")

	// Verify user exists in database.
	userRepo := repoFactory.UserRepository()
	user, err := userRepo.GetBySub(ctx, "demo-user")
	require.NoError(t, err, "GetBySub should succeed")
	require.NotNil(t, user, "User should exist")
	require.Equal(t, "demo-user", user.Sub, "User sub should match")
	require.Equal(t, "Demo User", user.Name, "User name should match")
	require.Equal(t, "demo", user.PreferredUsername, "Username should match")
	require.Equal(t, "demo@example.com", user.Email, "Email should match")
	require.True(t, bool(user.EmailVerified), "Email should be verified")
	require.True(t, bool(user.Enabled), "User should be enabled")
	require.False(t, bool(user.Locked), "User should not be locked")
	require.NotEmpty(t, user.PasswordHash, "Password hash should be set")
}

// TestCreateDemoUser_ExistingUser tests CreateDemoUser when user already exists.
func TestCreateDemoUser_ExistingUser(t *testing.T) {
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

	// Create demo user first time.
	_, _, created1, err := cryptoutilIdentityBootstrap.CreateDemoUser(ctx, repoFactory)
	require.NoError(t, err, "First CreateDemoUser should succeed")
	require.True(t, created1, "First call should create user")

	// Try to create again (should return existing, no password).
	sub, password, created2, err := cryptoutilIdentityBootstrap.CreateDemoUser(ctx, repoFactory)
	require.NoError(t, err, "Second CreateDemoUser should succeed")
	require.False(t, created2, "Second call should not create user")
	require.Equal(t, "demo-user", sub, "Sub should be demo-user")
	require.Empty(t, password, "Password should be empty for existing user")
}

// TestBootstrapUsers tests BootstrapUsers creates demo user successfully.
func TestBootstrapUsers(t *testing.T) {
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

	// Bootstrap users (should create demo user).
	err = cryptoutilIdentityBootstrap.BootstrapUsers(ctx, repoFactory)
	require.NoError(t, err, "BootstrapUsers should succeed")

	// Verify demo user exists.
	userRepo := repoFactory.UserRepository()
	user, err := userRepo.GetBySub(ctx, "demo-user")
	require.NoError(t, err, "GetBySub should succeed")
	require.NotNil(t, user, "Demo user should exist")
	require.Equal(t, "demo-user", user.Sub, "User sub should match")
	require.Equal(t, "Demo User", user.Name, "User name should match")
}

// TestBootstrapUsers_ExistingUser tests BootstrapUsers when user already exists.
func TestBootstrapUsers_ExistingUser(t *testing.T) {
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

	// Bootstrap users first time.
	err = cryptoutilIdentityBootstrap.BootstrapUsers(ctx, repoFactory)
	require.NoError(t, err, "First BootstrapUsers should succeed")

	// Bootstrap users second time (should succeed, no duplicate).
	err = cryptoutilIdentityBootstrap.BootstrapUsers(ctx, repoFactory)
	require.NoError(t, err, "Second BootstrapUsers should succeed")

	// Verify still only one demo user.
	userRepo := repoFactory.UserRepository()
	users, err := userRepo.List(ctx, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err, "List should succeed")
	require.Len(t, users, 1, "Should have exactly one user")
	require.Equal(t, "demo-user", users[0].Sub, "User should be demo-user")
}
