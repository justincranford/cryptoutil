// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIDP "cryptoutil/internal/apps/identity-idp/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const idpOriginalSecretHashForTest = "original-secret-hash"

func TestService_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(*testing.T, context.Context, *cryptoutilIdentityIDP.Service, *cryptoutilIdentityRepository.RepositoryFactory)
	}{
		{
			name: "expired session cleanup fails",
			run: func(t *testing.T, ctx context.Context, service *cryptoutilIdentityIDP.Service, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
				t.Helper()

				user := &cryptoutilIdentityDomain.User{Sub: googleUuid.Must(googleUuid.NewV7()).String()}
				require.NoError(t, repoFactory.UserRepository().Create(ctx, user))
				session := &cryptoutilIdentityDomain.Session{
					UserID:     user.ID,
					SessionID:  googleUuid.Must(googleUuid.NewV7()).String(),
					IssuedAt:   time.Now().UTC().Add(-2 * time.Hour),
					ExpiresAt:  time.Now().UTC().Add(-time.Hour),
					LastSeenAt: time.Now().UTC().Add(-2 * time.Hour),
				}
				require.NoError(t, repoFactory.SessionRepository().Create(ctx, session))
				require.NoError(t, repoFactory.DB().Exec(
					"CREATE TRIGGER fail_session_cleanup BEFORE DELETE ON sessions BEGIN SELECT RAISE(ABORT, 'injected cleanup failure'); END",
				).Error)

				err := service.Stop(ctx)
				require.ErrorContains(t, err, "failed to delete expired sessions during shutdown")
			},
		},
		{
			name: "client secret rotation update fails",
			run: func(t *testing.T, ctx context.Context, service *cryptoutilIdentityIDP.Service, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
				t.Helper()

				clientID, err := googleUuid.NewV7()
				require.NoError(t, err)
				clientOpaqueID, err := googleUuid.NewV7()
				require.NoError(t, err)

				client := &cryptoutilIdentityDomain.Client{
					ID:           clientID,
					ClientID:     "rotation-error-" + clientOpaqueID.String(),
					ClientSecret: idpOriginalSecretHashForTest,
					Name:         "Rotation Error Test Client",
					RedirectURIs: []string{cryptoutilSharedMagic.TestRedirectURI},
				}
				require.NoError(t, repoFactory.ClientRepository().Create(ctx, client))
				require.NoError(t, repoFactory.DB().Exec(
					"CREATE TRIGGER fail_client_secret_rotation BEFORE UPDATE OF client_secret ON clients BEGIN SELECT RAISE(ABORT, 'injected secret update failure'); END",
				).Error)

				_, err = service.RotateClientSecret(ctx, client.ClientID, "test-user", "test rotation")
				require.ErrorContains(t, err, "failed to rotate secret in repository")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
				Type: cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
			}
			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, repoFactory.Close())
			})
			require.NoError(t, repoFactory.AutoMigrate(ctx))

			service := cryptoutilIdentityIDP.NewService(&cryptoutilIdentityConfig.Config{}, repoFactory, nil)
			tc.run(t, ctx, service, repoFactory)
		})
	}
}
