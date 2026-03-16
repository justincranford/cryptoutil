// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestMigrateClientSecrets_Success validates client secret migration from legacy to PBKDF2.
func TestMigrateClientSecrets_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:         fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		AutoMigrate: true,
	}

	cfg := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	// Create test client with legacy bcrypt hash (simulated with PBKDF2 for testing)
	clientUUID := googleUuid.Must(googleUuid.NewV7())
	legacyHash, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2("legacy-secret")
	require.NoError(t, err, "Failed to hash legacy secret")

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      clientUUID,
		ClientID:                fmt.Sprintf("test-client-%s", clientUUID.String()),
		Name:                    "Test Client Migration",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		ClientSecret:            legacyHash,
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
	}

	clientRepo := repoFactory.ClientRepository()
	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	err = svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	// Run migration
	err = svc.MigrateClientSecrets(ctx)
	require.NoError(t, err, "Migration should succeed")

	// Verify client secret was migrated (remains PBKDF2 in this test)
	migratedClient, err := clientRepo.GetByClientID(ctx, testClient.ClientID)
	require.NoError(t, err, "Failed to retrieve migrated client")
	require.NotEmpty(t, migratedClient.ClientSecret, "Client secret should not be empty")
	require.Contains(t, migratedClient.ClientSecret, "$"+cryptoutilSharedMagic.PBKDF2DefaultHashName+"$", "Client secret should use PBKDF2 format")
}

// TestMigrateClientSecrets_NoClients validates migration with empty database.
func TestMigrateClientSecrets_NoClients(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:         fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		AutoMigrate: true,
	}

	cfg := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	err = svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	// Run migration on empty database
	err = svc.MigrateClientSecrets(ctx)
	require.NoError(t, err, "Migration should succeed with no clients")
}
