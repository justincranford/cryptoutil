// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestService_StartupHealthcheck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := createServiceComprehensiveTestConfig(t)
	repoFactory := createServiceComprehensiveTestRepoFactory(t)
	tokenSvc := createServiceComprehensiveTestTokenService(t)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, service, "Service should not be nil")

	// Start service and verify database connectivity.
	err := service.Start(ctx)
	require.NoError(t, err, "Service start should succeed with valid database")
}

func TestService_GracefulShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := createServiceComprehensiveTestConfig(t)
	repoFactory := createServiceComprehensiveTestRepoFactory(t)
	tokenSvc := createServiceComprehensiveTestTokenService(t)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, service, "Service should not be nil")

	// Start service.
	err := service.Start(ctx)
	require.NoError(t, err, "Service start should succeed")

	// Stop service gracefully (should clean up expired tokens and close database).
	err = service.Stop(ctx)
	require.NoError(t, err, "Service stop should succeed")
}

func createServiceComprehensiveTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:               "https://identity.example.com",
			AccessTokenLifetime:  15 * time.Minute,
			RefreshTokenLifetime: cryptoutilSharedMagic.HoursPerDay * time.Hour,
		},
	}
}

func createServiceComprehensiveTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:         "file::memory:?cache=private&_id=" + googleUuid.NewString(),
		AutoMigrate: true,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	// Run migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run database migrations")

	return repoFactory
}

func createServiceComprehensiveTestTokenService(t *testing.T) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	config := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:               "https://identity.example.com",
		AccessTokenLifetime:  15 * time.Minute,
		RefreshTokenLifetime: cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	// Service comprehensive tests don't need issuers (nil is acceptable).
	return cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config)
}
