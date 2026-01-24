// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestService_Creation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupConfig func(*testing.T) *cryptoutilIdentityConfig.Config
		setupRepo   func(*testing.T) *cryptoutilIdentityRepository.RepositoryFactory
		setupToken  func(*testing.T) *cryptoutilIdentityIssuer.TokenService
		wantErr     bool
		validate    func(*testing.T, *cryptoutilIdentityAuthz.Service)
	}{
		{
			name: "valid service creation",
			setupConfig: func(t *testing.T) *cryptoutilIdentityConfig.Config {
				return createServiceTestConfig(t)
			},
			setupRepo: func(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
				return createServiceTestRepoFactory(t)
			},
			setupToken: func(t *testing.T) *cryptoutilIdentityIssuer.TokenService {
				return createServiceTestTokenService(t)
			},
			wantErr: false,
			validate: func(t *testing.T, svc *cryptoutilIdentityAuthz.Service) {
				require.NotNil(t, svc, "Service should not be nil")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := tc.setupConfig(t)
			repoFactory := tc.setupRepo(t)
			tokenSvc := tc.setupToken(t)

			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

			if tc.wantErr {
				require.Nil(t, svc, "Service should be nil on error")
			} else {
				tc.validate(t, svc)
			}
		})
	}
}

func TestService_StartStop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupRepo func(*testing.T) *cryptoutilIdentityRepository.RepositoryFactory
		startErr  bool
		stopErr   bool
	}{
		{
			name: "successful start and stop",
			setupRepo: func(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
				return createServiceTestRepoFactory(t)
			},
			startErr: false,
			stopErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createServiceTestConfig(t)
			repoFactory := tc.setupRepo(t)
			tokenSvc := createServiceTestTokenService(t)

			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
			require.NotNil(t, svc, "Service should not be nil")

			ctx := context.Background()

			// Test Start.
			err := svc.Start(ctx)
			if tc.startErr {
				require.Error(t, err, "Start should fail")
			} else {
				require.NoError(t, err, "Start should succeed")
			}

			// Test Stop.
			err = svc.Stop(ctx)
			if tc.stopErr {
				require.Error(t, err, "Stop should fail")
			} else {
				require.NoError(t, err, "Stop should succeed")
			}
		})
	}
}

func TestService_MigrateClientSecrets(t *testing.T) {
	t.Parallel()

	testSecret := "test-secret-plaintext"

	tests := []struct {
		name         string
		setupClients func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory) []*cryptoutilIdentityDomain.Client
		wantErr      bool
		errContains  string
	}{
		{
			name: "migrate clients with legacy plaintext secrets",
			setupClients: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) []*cryptoutilIdentityDomain.Client {
				ctx := context.Background()
				clientRepo := repoFactory.ClientRepository()

				clientID1 := googleUuid.New()
				client1 := &cryptoutilIdentityDomain.Client{
					ID:                      clientID1,
					ClientID:                "client1",
					Name:                    "Test Client 1",
					ClientSecret:            testSecret,
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					RedirectURIs:            []string{"https://example.com/callback"},
					AllowedGrantTypes:       []string{"authorization_code"},
					AllowedResponseTypes:    []string{"code"},
					AllowedScopes:           []string{"openid", "profile"},
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
				}
				err := clientRepo.Create(ctx, client1)
				require.NoError(t, err, "Failed to create test client 1")

				clientID2 := googleUuid.New()
				client2 := &cryptoutilIdentityDomain.Client{
					ID:                      clientID2,
					ClientID:                "client2",
					Name:                    "Test Client 2",
					ClientSecret:            testSecret,
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					RedirectURIs:            []string{"https://example2.com/callback"},
					AllowedGrantTypes:       []string{"client_credentials"},
					AllowedResponseTypes:    []string{},
					AllowedScopes:           []string{"api:read"},
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				}
				err = clientRepo.Create(ctx, client2)
				require.NoError(t, err, "Failed to create test client 2")

				return []*cryptoutilIdentityDomain.Client{client1, client2}
			},
			wantErr: false,
		},
		{
			name: "no clients to migrate",
			setupClients: func(_ *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) []*cryptoutilIdentityDomain.Client {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createServiceTestConfig(t)
			repoFactory := createServiceTestRepoFactory(t)
			tokenSvc := createServiceTestTokenService(t)

			// Setup test clients if provided.
			var originalClients []*cryptoutilIdentityDomain.Client
			if tc.setupClients != nil {
				originalClients = tc.setupClients(t, repoFactory)
			}

			// Create service and migrate.
			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
			require.NotNil(t, svc, "Service should not be nil")

			ctx := context.Background()
			err := svc.MigrateClientSecrets(ctx)

			if tc.wantErr {
				require.Error(t, err, "MigrateClientSecrets should fail")

				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains, "Error message mismatch")
				}
			} else {
				require.NoError(t, err, "MigrateClientSecrets should succeed")

				// Verify secrets were migrated for test clients.
				if len(originalClients) > 0 {
					clientRepo := repoFactory.ClientRepository()
					for _, originalClient := range originalClients {
						updatedClient, err := clientRepo.GetByID(ctx, originalClient.ID)
						require.NoError(t, err, "Failed to retrieve updated client")
						require.NotNil(t, updatedClient, "Updated client should not be nil")

						// Verify secret is now hashed (PBKDF2 format: $pbkdf2-sha256$...).
						require.NotEqual(t, testSecret, updatedClient.ClientSecret, "Secret should be hashed")
						require.Contains(t, updatedClient.ClientSecret, "$"+cryptoutilSharedMagic.PBKDF2DefaultHashName+"$", "Secret should use PBKDF2 format")
					}
				}
			}
		})
	}
}

func createServiceTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         "file::memory:?cache=private",
		AutoMigrate: true,
	}

	return &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}

func createServiceTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createServiceTestConfig(t)
	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	// Run migrations to create tables.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}

func createServiceTestTokenService(t *testing.T) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	return nil
}
