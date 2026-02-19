// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
	cryptoutilIdentityTestTestutils "cryptoutil/internal/apps/identity/test/testutils"
)

// TestDatabaseSetup tests basic database setup and teardown.
func TestDatabaseSetup(t *testing.T) {
	// Setup test database.
	db := cryptoutilIdentityTestTestutils.SetupTestDatabase(t)
	require.NotNil(t, db, "database should be initialized")

	// Cleanup test database.
	cryptoutilIdentityTestTestutils.CleanupTestDatabase(t, db)
}

// TestConfigCreation tests test configuration creation.
func TestConfigCreation(t *testing.T) {
	config := cryptoutilIdentityTestTestutils.CreateTestConfig(t, 8100, 8100, 8110)

	require.NotNil(t, config, "config should be created")
	require.Equal(t, 8100, config.AuthZ.Port, "AuthZ port should match")
	require.Equal(t, 8100, config.IDP.Port, "IDP port should match")
	require.Equal(t, 8110, config.RS.Port, "RS port should match")
	require.Equal(t, "sqlite", config.Database.Type, "database type should be sqlite")
	require.Equal(t, ":memory:", config.Database.DSN, "database DSN should be in-memory")
}

// TestUserRepository_CRUD tests User repository CRUD operations.
func TestUserRepository_CRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := cryptoutilIdentityTestTestutils.SetupTestDatabase(t)

	userRepo := cryptoutilIdentityORM.NewUserRepository(db)

	tests := []struct {
		name      string
		operation func(t *testing.T)
	}{
		{
			name: "Create user",
			operation: func(t *testing.T) {
				user := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test-" + googleUuid.Must(googleUuid.NewV7()).String() + "@example.com",
					Name:              "Test User",
					PreferredUsername: "testuser-" + googleUuid.Must(googleUuid.NewV7()).String(),
				}

				err := userRepo.Create(ctx, user)
				require.NoError(t, err)
				require.NotEmpty(t, user.ID)
			},
		},
		{
			name: "GetByID user",
			operation: func(t *testing.T) {
				user := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test-" + googleUuid.Must(googleUuid.NewV7()).String() + "@example.com",
					Name:              "Test User",
					PreferredUsername: "testuser-" + googleUuid.Must(googleUuid.NewV7()).String(),
				}
				err := userRepo.Create(ctx, user)
				require.NoError(t, err)

				retrieved, err := userRepo.GetByID(ctx, user.ID)
				require.NoError(t, err)
				require.Equal(t, user.ID, retrieved.ID)
				require.Equal(t, user.Sub, retrieved.Sub)
			},
		},
		{
			name: "GetBySub user",
			operation: func(t *testing.T) {
				user := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test-" + googleUuid.Must(googleUuid.NewV7()).String() + "@example.com",
					Name:              "Test User",
					PreferredUsername: "testuser-" + googleUuid.Must(googleUuid.NewV7()).String(),
				}
				err := userRepo.Create(ctx, user)
				require.NoError(t, err)

				retrieved, err := userRepo.GetBySub(ctx, user.Sub)
				require.NoError(t, err)
				require.Equal(t, user.Sub, retrieved.Sub)
			},
		},
		{
			name: "Update user",
			operation: func(t *testing.T) {
				user := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test-" + googleUuid.Must(googleUuid.NewV7()).String() + "@example.com",
					Name:              "Test User",
					PreferredUsername: "testuser-" + googleUuid.Must(googleUuid.NewV7()).String(),
				}
				err := userRepo.Create(ctx, user)
				require.NoError(t, err)

				user.Name = "Updated Name"
				err = userRepo.Update(ctx, user)
				require.NoError(t, err)

				retrieved, err := userRepo.GetByID(ctx, user.ID)
				require.NoError(t, err)
				require.Equal(t, "Updated Name", retrieved.Name)
			},
		},
		{
			name: "Delete user",
			operation: func(t *testing.T) {
				user := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test-" + googleUuid.Must(googleUuid.NewV7()).String() + "@example.com",
					Name:              "Test User",
					PreferredUsername: "testuser-" + googleUuid.Must(googleUuid.NewV7()).String(),
				}
				err := userRepo.Create(ctx, user)
				require.NoError(t, err)

				err = userRepo.Delete(ctx, user.ID)
				require.NoError(t, err)
			},
		},
		{
			name: "List users",
			operation: func(t *testing.T) {
				user1 := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-1-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test1@example.com",
					Name:              "Test User 1",
					PreferredUsername: "testuser1",
				}
				user2 := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-2-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test2@example.com",
					Name:              "Test User 2",
					PreferredUsername: "testuser2",
				}

				err := userRepo.Create(ctx, user1)
				require.NoError(t, err)
				err = userRepo.Create(ctx, user2)
				require.NoError(t, err)

				users, err := userRepo.List(ctx, 0, 10)
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(users), 2)
			},
		},
		{
			name: "Count users",
			operation: func(t *testing.T) {
				user := &cryptoutilIdentityDomain.User{
					Sub:               "test-user-" + googleUuid.Must(googleUuid.NewV7()).String(),
					Email:             "test-" + googleUuid.Must(googleUuid.NewV7()).String() + "@example.com",
					Name:              "Test User",
					PreferredUsername: "testuser-" + googleUuid.Must(googleUuid.NewV7()).String(),
				}
				err := userRepo.Create(ctx, user)
				require.NoError(t, err)

				count, err := userRepo.Count(ctx)
				require.NoError(t, err)
				require.GreaterOrEqual(t, count, int64(1))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.operation(t)
		})
	}
}

// TestClientRepository_CRUD tests Client repository CRUD operations.
func TestClientRepository_CRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := cryptoutilIdentityTestTestutils.SetupTestDatabase(t)

	clientRepo := cryptoutilIdentityORM.NewClientRepository(db)

	tests := []struct {
		name      string
		operation func(t *testing.T)
	}{
		{
			name: "Create client",
			operation: func(t *testing.T) {
				client := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client",
				}

				err := clientRepo.Create(ctx, client)
				require.NoError(t, err)
				require.NotEmpty(t, client.ID)
			},
		},
		{
			name: "GetByID client",
			operation: func(t *testing.T) {
				client := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client",
				}
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err)

				retrieved, err := clientRepo.GetByID(ctx, client.ID)
				require.NoError(t, err)
				require.Equal(t, client.ID, retrieved.ID)
				require.Equal(t, client.ClientID, retrieved.ClientID)
			},
		},
		{
			name: "GetByClientID client",
			operation: func(t *testing.T) {
				client := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client",
				}
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err)

				retrieved, err := clientRepo.GetByClientID(ctx, client.ClientID)
				require.NoError(t, err)
				require.Equal(t, client.ClientID, retrieved.ClientID)
			},
		},
		{
			name: "Update client",
			operation: func(t *testing.T) {
				client := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client",
				}
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err)

				client.Name = "Updated Client"
				err = clientRepo.Update(ctx, client)
				require.NoError(t, err)

				retrieved, err := clientRepo.GetByID(ctx, client.ID)
				require.NoError(t, err)
				require.Equal(t, "Updated Client", retrieved.Name)
			},
		},
		{
			name: "Delete client",
			operation: func(t *testing.T) {
				client := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client",
				}
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err)

				err = clientRepo.Delete(ctx, client.ID)
				require.NoError(t, err)
			},
		},
		{
			name: "List clients",
			operation: func(t *testing.T) {
				client1 := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-1-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client 1",
				}
				client2 := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-2-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client 2",
				}

				err := clientRepo.Create(ctx, client1)
				require.NoError(t, err)
				err = clientRepo.Create(ctx, client2)
				require.NoError(t, err)

				clients, err := clientRepo.List(ctx, 0, 10)
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(clients), 2)
			},
		},
		{
			name: "Count clients",
			operation: func(t *testing.T) {
				client := &cryptoutilIdentityDomain.Client{
					ClientID:   "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
					ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
					Name:       "Test Client",
				}
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err)

				count, err := clientRepo.Count(ctx)
				require.NoError(t, err)
				require.GreaterOrEqual(t, count, int64(1))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.operation(t)
		})
	}
}

// TestTokenRepository_CRUD tests Token repository CRUD operations.
