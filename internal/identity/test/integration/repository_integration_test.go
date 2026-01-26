// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/identity/repository/orm"
	cryptoutilIdentityTestTestutils "cryptoutil/internal/identity/test/testutils"
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
	config := cryptoutilIdentityTestTestutils.CreateTestConfig(t, 8443, 8444, 8445)

	require.NotNil(t, config, "config should be created")
	assert.Equal(t, 8443, config.AuthZ.Port, "AuthZ port should match")
	assert.Equal(t, 8444, config.IDP.Port, "IDP port should match")
	assert.Equal(t, 8445, config.RS.Port, "RS port should match")
	assert.Equal(t, "sqlite", config.Database.Type, "database type should be sqlite")
	assert.Equal(t, ":memory:", config.Database.DSN, "database DSN should be in-memory")
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
				assert.NotEmpty(t, user.ID)
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
				assert.Equal(t, user.ID, retrieved.ID)
				assert.Equal(t, user.Sub, retrieved.Sub)
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
				assert.Equal(t, user.Sub, retrieved.Sub)
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
				assert.Equal(t, "Updated Name", retrieved.Name)
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
				assert.GreaterOrEqual(t, len(users), 2)
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
				assert.GreaterOrEqual(t, count, int64(1))
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
				assert.NotEmpty(t, client.ID)
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
				assert.Equal(t, client.ID, retrieved.ID)
				assert.Equal(t, client.ClientID, retrieved.ClientID)
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
				assert.Equal(t, client.ClientID, retrieved.ClientID)
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
				assert.Equal(t, "Updated Client", retrieved.Name)
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
				assert.GreaterOrEqual(t, len(clients), 2)
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
				assert.GreaterOrEqual(t, count, int64(1))
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
func TestTokenRepository_CRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := cryptoutilIdentityTestTestutils.SetupTestDatabase(t)

	tokenRepo := cryptoutilIdentityORM.NewTokenRepository(db)
	clientRepo := cryptoutilIdentityORM.NewClientRepository(db)

	// Create test client for foreign key.
	testClient := &cryptoutilIdentityDomain.Client{
		ClientID:   "token-test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:       "Test Client",
	}
	err := clientRepo.Create(ctx, testClient)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation func(t *testing.T)
	}{
		{
			name: "Create token",
			operation: func(t *testing.T) {
				token := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}

				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)
				assert.NotEmpty(t, token.ID)
			},
		},
		{
			name: "GetByID token",
			operation: func(t *testing.T) {
				token := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				retrieved, err := tokenRepo.GetByID(ctx, token.ID)
				require.NoError(t, err)
				assert.Equal(t, token.ID, retrieved.ID)
				assert.Equal(t, token.TokenValue, retrieved.TokenValue)
			},
		},
		{
			name: "GetByTokenValue token",
			operation: func(t *testing.T) {
				token := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				retrieved, err := tokenRepo.GetByTokenValue(ctx, token.TokenValue)
				require.NoError(t, err)
				assert.Equal(t, token.TokenValue, retrieved.TokenValue)
			},
		},
		{
			name: "Update token",
			operation: func(t *testing.T) {
				token := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				token.Scopes = []string{"openid", "profile", "email"}
				err = tokenRepo.Update(ctx, token)
				require.NoError(t, err)

				retrieved, err := tokenRepo.GetByID(ctx, token.ID)
				require.NoError(t, err)
				assert.Equal(t, token.Scopes, retrieved.Scopes)
			},
		},
		{
			name: "Delete token",
			operation: func(t *testing.T) {
				token := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				err = tokenRepo.Delete(ctx, token.ID)
				require.NoError(t, err)
			},
		},
		{
			name: "List tokens",
			operation: func(t *testing.T) {
				token1 := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-1-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				token2 := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-2-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}

				err := tokenRepo.Create(ctx, token1)
				require.NoError(t, err)
				err = tokenRepo.Create(ctx, token2)
				require.NoError(t, err)

				tokens, err := tokenRepo.List(ctx, 0, 10)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(tokens), 2)
			},
		},
		{
			name: "Count tokens",
			operation: func(t *testing.T) {
				token := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				count, err := tokenRepo.Count(ctx)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, count, int64(1))
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

// TestSessionRepository_CRUD tests Session repository CRUD operations.
func TestSessionRepository_CRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := cryptoutilIdentityTestTestutils.SetupTestDatabase(t)

	sessionRepo := cryptoutilIdentityORM.NewSessionRepository(db)
	userRepo := cryptoutilIdentityORM.NewUserRepository(db)

	// Create test user for foreign key.
	testUser := &cryptoutilIdentityDomain.User{
		Sub:               "session-test-user-" + googleUuid.Must(googleUuid.NewV7()).String(),
		Email:             "sessiontest@example.com",
		Name:              "Session Test User",
		PreferredUsername: "sessiontestuser",
	}
	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation func(t *testing.T)
	}{
		{
			name: "Create session",
			operation: func(t *testing.T) {
				session := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}

				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)
				assert.NotEmpty(t, session.ID)
			},
		},
		{
			name: "GetByID session",
			operation: func(t *testing.T) {
				session := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)

				retrieved, err := sessionRepo.GetByID(ctx, session.ID)
				require.NoError(t, err)
				assert.Equal(t, session.ID, retrieved.ID)
				assert.Equal(t, session.SessionID, retrieved.SessionID)
			},
		},
		{
			name: "GetBySessionID session",
			operation: func(t *testing.T) {
				session := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)

				retrieved, err := sessionRepo.GetBySessionID(ctx, session.SessionID)
				require.NoError(t, err)
				assert.Equal(t, session.SessionID, retrieved.SessionID)
			},
		},
		{
			name: "Update session",
			operation: func(t *testing.T) {
				session := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)

				session.Active = boolPtr(false)
				err = sessionRepo.Update(ctx, session)
				require.NoError(t, err)

				retrieved, err := sessionRepo.GetByID(ctx, session.ID)
				require.NoError(t, err)
				assert.False(t, *retrieved.Active)
			},
		},
		{
			name: "Delete session",
			operation: func(t *testing.T) {
				session := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)

				err = sessionRepo.Delete(ctx, session.ID)
				require.NoError(t, err)
			},
		},
		{
			name: "List sessions",
			operation: func(t *testing.T) {
				session1 := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-1-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}
				session2 := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-2-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}

				err := sessionRepo.Create(ctx, session1)
				require.NoError(t, err)
				err = sessionRepo.Create(ctx, session2)
				require.NoError(t, err)

				sessions, err := sessionRepo.List(ctx, 0, 10)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(sessions), 2)
			},
		},
		{
			name: "Count sessions",
			operation: func(t *testing.T) {
				session := &cryptoutilIdentityDomain.Session{
					SessionID: "test-session-" + googleUuid.Must(googleUuid.NewV7()).String(),
					UserID:    testUser.ID,
					Active:    boolPtr(true),
					IssuedAt:  time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)

				count, err := sessionRepo.Count(ctx)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, count, int64(1))
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
