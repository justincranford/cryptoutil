// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestServiceStart validates service startup.
func TestServiceStart(t *testing.T) {
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

	err = service.Start(ctx)
	require.NoError(t, err)
}

// TestServiceStop validates service shutdown.
func TestServiceStop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations to create sessions table.
	db := repoFactory.DB()
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			expires_at INTEGER
		)
	`).Error
	require.NoError(t, err)

	config := &cryptoutilIdentityConfig.Config{
		IDP:      &cryptoutilIdentityConfig.ServerConfig{},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{},
	}

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	require.NotNil(t, service)

	err = service.Stop(ctx)
	require.NoError(t, err)
}
