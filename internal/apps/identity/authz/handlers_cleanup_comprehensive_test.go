// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestCleanupService_DeletesExpiredTokens(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := createCleanupTestConfig(t)
	repoFactory := createCleanupTestRepoFactory(t)
	service := createCleanupTestService(t, config, repoFactory)

	// Create cleanup service.
	cleanup := cryptoutilIdentityAuthz.NewCleanupService(service)
	require.NotNil(t, cleanup, "Cleanup service should not be nil")

	// Create client for token association.
	testClient := createCleanupTestClient(ctx, t, repoFactory)

	// Client.ID (UUID) is the foreign key for tokens, NOT ClientID (string).
	clientUUID := testClient.ID

	// Create expired access token.
	expiredAccessToken := &cryptoutilIdentityDomain.Token{
		ID:          googleUuid.New(),
		ClientID:    clientUUID,
		TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
		TokenValue:  "expired-access-token-" + googleUuid.NewString(),
		Scopes:      []string{"read", "write"},
		ExpiresAt:   time.Now().UTC().Add(-1 * time.Hour), // Expired 1 hour ago
		IssuedAt:    time.Now().UTC().Add(-2 * time.Hour),
		NotBefore:   time.Now().UTC().Add(-2 * time.Hour),
		Revoked:     false,
	}

	tokenRepo := repoFactory.TokenRepository()
	err := tokenRepo.Create(ctx, expiredAccessToken)
	require.NoError(t, err, "Failed to create expired access token")

	// Create valid access token (should NOT be deleted).
	validAccessToken := &cryptoutilIdentityDomain.Token{
		ID:          googleUuid.New(),
		ClientID:    clientUUID,
		TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
		TokenValue:  "valid-access-token-" + googleUuid.NewString(),
		Scopes:      []string{"read"},
		ExpiresAt:   time.Now().UTC().Add(1 * time.Hour), // Valid for 1 hour
		IssuedAt:    time.Now().UTC(),
		NotBefore:   time.Now().UTC(),
		Revoked:     false,
	}

	err = tokenRepo.Create(ctx, validAccessToken)
	require.NoError(t, err, "Failed to create valid access token")

	// Run cleanup manually (simulates cleanup loop iteration).
	err = tokenRepo.DeleteExpired(ctx)
	require.NoError(t, err, "Cleanup should succeed")

	// Verify expired token deleted.
	_, err = tokenRepo.GetByTokenValue(ctx, expiredAccessToken.TokenValue)
	require.Error(t, err, "Expired token should be deleted")

	// Verify valid token still exists.
	retrievedToken, err := tokenRepo.GetByTokenValue(ctx, validAccessToken.TokenValue)
	require.NoError(t, err, "Valid token should still exist")
	require.Equal(t, validAccessToken.TokenValue, retrievedToken.TokenValue, "Token value should match")
}

func TestCleanupService_IntervalConfiguration(t *testing.T) {
	t.Parallel()

	config := createCleanupTestConfig(t)
	repoFactory := createCleanupTestRepoFactory(t)
	service := createCleanupTestService(t, config, repoFactory)

	// Create cleanup service with default interval.
	cleanup := cryptoutilIdentityAuthz.NewCleanupService(service)
	require.NotNil(t, cleanup, "Cleanup service should not be nil")

	// Configure custom interval.
	customInterval := 30 * time.Second
	cleanup = cleanup.WithInterval(customInterval)
	require.NotNil(t, cleanup, "Cleanup service should support custom interval")
}

func TestCleanupService_StartStopLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := createCleanupTestConfig(t)
	repoFactory := createCleanupTestRepoFactory(t)
	service := createCleanupTestService(t, config, repoFactory)

	// Create cleanup service with short interval for testing.
	cleanup := cryptoutilIdentityAuthz.NewCleanupService(service).WithInterval(100 * time.Millisecond)
	require.NotNil(t, cleanup, "Cleanup service should not be nil")

	// Start cleanup service.
	cleanup.Start(ctx)

	// Wait for at least one cleanup cycle.
	time.Sleep(200 * time.Millisecond)

	// Stop cleanup service gracefully.
	cleanup.Stop()
	// Verify cleanup service stopped (no panic, no hang).
}

func createCleanupTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  15 * time.Minute,
			RefreshTokenLifetime: 24 * time.Hour,
		},
	}
}

func createCleanupTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         "file::memory:?cache=private&_id=" + googleUuid.NewString(),
		AutoMigrate: true, // Enable auto-migration for tests
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	// Run migrations explicitly (AutoMigrate field doesn't trigger migration).
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run database migrations")

	return repoFactory
}

func createCleanupTestService(
	t *testing.T,
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) *cryptoutilIdentityAuthz.Service {
	t.Helper()

	// Cleanup tests don't need token service (only testing repository cleanup).
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, service, "Service should not be nil")

	return service
}

func createCleanupTestClient(
	ctx context.Context,
	t *testing.T,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) *cryptoutilIdentityDomain.Client {
	t.Helper()

	clientID := "test-client-" + googleUuid.NewString()
	client := &cryptoutilIdentityDomain.Client{
		ClientID:                clientID,
		ClientSecret:            "cleanup-test-secret",
		Name:                    "Cleanup Test Client",
		RedirectURIs:            []string{"https://example.com/callback"},
		AllowedScopes:           []string{"read", "write"},
		ClientType:              "confidential",
		TokenEndpointAuthMethod: cryptoutilIdentityMagic.ClientAuthMethodSecretPost,
		RequirePKCE:             boolPtr(true),
	}

	clientRepo := repoFactory.ClientRepository()
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err, "Failed to create test client")

	return client
}
