// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
	cryptoutilIdentityTestTestutils "cryptoutil/internal/apps/identity/test/testutils"
)

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
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}

				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)
				require.NotEmpty(t, token.ID)
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
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				retrieved, err := tokenRepo.GetByID(ctx, token.ID)
				require.NoError(t, err)
				require.Equal(t, token.ID, retrieved.ID)
				require.Equal(t, token.TokenValue, retrieved.TokenValue)
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
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				retrieved, err := tokenRepo.GetByTokenValue(ctx, token.TokenValue)
				require.NoError(t, err)
				require.Equal(t, token.TokenValue, retrieved.TokenValue)
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
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				token.Scopes = []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail}
				err = tokenRepo.Update(ctx, token)
				require.NoError(t, err)

				retrieved, err := tokenRepo.GetByID(ctx, token.ID)
				require.NoError(t, err)
				require.Equal(t, token.Scopes, retrieved.Scopes)
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
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
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
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				token2 := &cryptoutilIdentityDomain.Token{
					TokenValue:  "test-token-2-" + googleUuid.Must(googleUuid.NewV7()).String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    testClient.ID,
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}

				err := tokenRepo.Create(ctx, token1)
				require.NoError(t, err)
				err = tokenRepo.Create(ctx, token2)
				require.NoError(t, err)

				tokens, err := tokenRepo.List(ctx, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(tokens), 2)
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
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC(),
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
				}
				err := tokenRepo.Create(ctx, token)
				require.NoError(t, err)

				count, err := tokenRepo.Count(ctx)
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
				require.NotEmpty(t, session.ID)
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
				require.Equal(t, session.ID, retrieved.ID)
				require.Equal(t, session.SessionID, retrieved.SessionID)
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
				require.Equal(t, session.SessionID, retrieved.SessionID)
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
				require.False(t, *retrieved.Active)
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

				sessions, err := sessionRepo.List(ctx, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(sessions), 2)
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
