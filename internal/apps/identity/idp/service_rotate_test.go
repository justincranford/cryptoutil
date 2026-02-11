// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestService_RotateClientSecret_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Initialize database.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	// Create test client.
	clientRepo := repoFactory.ClientRepository()

	testClient := &cryptoutilIdentityDomain.Client{
		ID:           googleUuid.New(),
		ClientID:     "test-rotate-client",
		ClientSecret: "original-secret-hash",
		Name:         "Test Rotate Client",
		RedirectURIs: []string{"https://example.com/callback"},
	}

	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err)

	// Create IDP service.
	config := &cryptoutilIdentityConfig.Config{}
	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)

	// Rotate client secret.
	newSecret, err := service.RotateClientSecret(ctx, "test-rotate-client", "test-user", "rotation test")
	require.NoError(t, err)
	require.NotEmpty(t, newSecret, "New secret should not be empty")
	require.NotEqual(t, "original-secret-hash", newSecret, "New secret should differ from original")
}

func TestService_RotateClientSecret_ClientNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Initialize database.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	// Create IDP service.
	config := &cryptoutilIdentityConfig.Config{}
	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)

	// Attempt to rotate secret for non-existent client.
	_, err = service.RotateClientSecret(ctx, "non-existent-client", "test-user", "rotation test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to find client")
}
