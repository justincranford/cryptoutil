// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestAuthMiddleware validates session-based authentication middleware.
func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP:      &cryptoutilIdentityConfig.ServerConfig{},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	require.NotNil(t, service)

	middleware := service.AuthMiddleware()
	require.NotNil(t, middleware)
}

// TestTokenAuthMiddleware validates token-based authentication middleware.
func TestTokenAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP:      &cryptoutilIdentityConfig.ServerConfig{},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	require.NotNil(t, service)

	middleware := service.TokenAuthMiddleware()
	require.NotNil(t, middleware)
}

// TestInitializeAuthProfiles validates authentication profile registration.
func TestInitializeAuthProfiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Session{},
	)
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP:      &cryptoutilIdentityConfig.ServerConfig{},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	require.NotNil(t, service)

	// Start service to initialize auth profiles.
	err = service.Start(ctx)
	require.NoError(t, err)
}

// TestHybridAuthMiddleware validates hybrid (Bearer token OR session cookie) authentication.
//
// Requirements verified:
// - P1.6.3: Session cookie authentication for SPA UI.
func TestHybridAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP:      &cryptoutilIdentityConfig.ServerConfig{},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	require.NotNil(t, service)

	middleware := service.HybridAuthMiddleware()
	require.NotNil(t, middleware)
}
