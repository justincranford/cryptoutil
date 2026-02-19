// Copyright (c) 2025 Justin Cranford
//
//

package tests

import (
	"context"
	"testing"
	"time"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Register CGO-free SQLite driver
)

func TestUserRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	userRepo := repoFactory.UserRepository()

	// Test Create
	uuidSuffix := googleUuid.Must(googleUuid.NewV7()).String()
	user := &cryptoutilIdentityDomain.User{
		Sub:               "test-user-" + uuidSuffix,
		Email:             "test-" + uuidSuffix + "@example.com",
		Name:              "Test User",
		PreferredUsername: "test-" + uuidSuffix,
	}

	err := userRepo.Create(ctx, user)
	require.NoError(t, err)
	require.NotEmpty(t, user.ID)

	// Test GetByID
	retrievedUser, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedUser)
	require.Equal(t, user.ID, retrievedUser.ID)

	// Test GetBySub
	userBySub, err := userRepo.GetBySub(ctx, user.Sub)
	require.NoError(t, err)
	require.NotNil(t, userBySub)
	require.Equal(t, user.Sub, userBySub.Sub)

	// Test Update
	updatedUser := *retrievedUser
	updatedUser.Name = "Updated Test User"
	err = userRepo.Update(ctx, &updatedUser)
	require.NoError(t, err)

	// Verify update
	retrievedUpdated, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Test User", retrievedUpdated.Name)

	// Test List
	users, err := userRepo.List(ctx, 0, 10)
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, user.ID, users[0].ID)

	// Test Count
	count, err := userRepo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Test Delete
	err = userRepo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = userRepo.GetByID(ctx, user.ID)
	require.Error(t, err) // Should return error for not found
}

func TestClientRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	clientRepo := repoFactory.ClientRepository()

	uuidSuffix := googleUuid.Must(googleUuid.NewV7()).String()

	// Test Create
	client := &cryptoutilIdentityDomain.Client{
		ClientID:   "test-client-" + uuidSuffix,
		ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:       "Test Client",
	}

	err := clientRepo.Create(ctx, client)
	require.NoError(t, err)
	require.NotEmpty(t, client.ID)

	// Test GetByID
	retrievedClient, err := clientRepo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, client.ID, retrievedClient.ID)

	// Test GetByClientID
	clientByID, err := clientRepo.GetByClientID(ctx, client.ClientID)
	require.NoError(t, err)
	require.Equal(t, client.ClientID, clientByID.ClientID)

	// Test Update
	updatedClient := *retrievedClient
	updatedClient.Name = "Updated Test Client"
	err = clientRepo.Update(ctx, &updatedClient)
	require.NoError(t, err)

	// Test List
	clients, err := clientRepo.List(ctx, 0, 10)
	require.NoError(t, err)
	require.Len(t, clients, 1)

	// Test Count
	count, err := clientRepo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Test Delete
	err = clientRepo.Delete(ctx, client.ID)
	require.NoError(t, err)
}

func TestTokenRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	tokenRepo := repoFactory.TokenRepository()

	uuidSuffix := googleUuid.Must(googleUuid.NewV7()).String()

	// Create a test client first
	client := &cryptoutilIdentityDomain.Client{
		ClientID:   "token-test-client-" + uuidSuffix,
		ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:       "Test Client",
	}
	clientRepo := repoFactory.ClientRepository()
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err)

	// Test Create
	token := &cryptoutilIdentityDomain.Token{
		TokenValue:  "test-token-" + uuidSuffix,
		TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
		ClientID:    client.ID,
		Scopes:      []string{"openid", "profile"},
		IssuedAt:    time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
	}

	err = tokenRepo.Create(ctx, token)
	require.NoError(t, err)
	require.NotEmpty(t, token.ID)

	// Test GetByID
	retrievedToken, err := tokenRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	require.Equal(t, token.ID, retrievedToken.ID)

	// Test GetByTokenValue
	tokenByValue, err := tokenRepo.GetByTokenValue(ctx, token.TokenValue)
	require.NoError(t, err)
	require.Equal(t, token.TokenValue, tokenByValue.TokenValue)

	// Test Update
	updatedToken := *retrievedToken
	updatedToken.Scopes = []string{"openid", "profile", "email"}
	err = tokenRepo.Update(ctx, &updatedToken)
	require.NoError(t, err)

	// Test RevokeByID
	err = tokenRepo.RevokeByID(ctx, token.ID)
	require.NoError(t, err)

	// Verify revocation
	retrievedRevoked, err := tokenRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	require.True(t, retrievedRevoked.Revoked.Bool())

	// Test Delete
	err = tokenRepo.Delete(ctx, token.ID)
	require.NoError(t, err)
}

func TestSessionRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	sessionRepo := repoFactory.SessionRepository()

	// Create test user first (foreign key requirement)
	uuidSuffix := googleUuid.Must(googleUuid.NewV7()).String()
	user := &cryptoutilIdentityDomain.User{
		Sub:               "session-test-user-" + uuidSuffix,
		Email:             "session-test-" + uuidSuffix + "@example.com",
		Name:              "Session Test User",
		PreferredUsername: "session-test-" + uuidSuffix,
	}
	userRepo := repoFactory.UserRepository()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Test Create
	session := &cryptoutilIdentityDomain.Session{
		SessionID: "test-session-" + uuidSuffix,
		UserID:    user.ID,
		Active:    boolPtr(true),
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)
	require.NotEmpty(t, session.ID)

	// Test GetByID
	retrievedSession, err := sessionRepo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	require.Equal(t, session.ID, retrievedSession.ID)

	// Test GetBySessionID
	sessionByID, err := sessionRepo.GetBySessionID(ctx, session.SessionID)
	require.NoError(t, err)
	require.Equal(t, session.SessionID, sessionByID.SessionID)

	// Test Update
	updatedSession := *retrievedSession
	updatedSession.Active = boolPtr(false)
	err = sessionRepo.Update(ctx, &updatedSession)
	require.NoError(t, err)

	// Test List
	sessions, err := sessionRepo.List(ctx, 0, 10)
	require.NoError(t, err)
	require.Len(t, sessions, 1)

	// Test Count
	count, err := sessionRepo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Test Delete
	err = sessionRepo.Delete(ctx, session.ID)
	require.NoError(t, err)
}

func TestClientProfileRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	profileRepo := repoFactory.ClientProfileRepository()

	// Test Create
	profile := &cryptoutilIdentityDomain.ClientProfile{
		Name:        "test-profile",
		Description: "Test profile",
	}

	err := profileRepo.Create(ctx, profile)
	require.NoError(t, err)
	require.NotEmpty(t, profile.ID)

	// Test GetByID
	retrievedProfile, err := profileRepo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	require.Equal(t, profile.ID, retrievedProfile.ID)

	// Test GetByName
	profileByName, err := profileRepo.GetByName(ctx, profile.Name)
	require.NoError(t, err)
	require.Equal(t, profile.Name, profileByName.Name)

	// Test Update
	updatedProfile := *retrievedProfile
	updatedProfile.Description = "Updated test profile"
	err = profileRepo.Update(ctx, &updatedProfile)
	require.NoError(t, err)

	// Test Delete
	err = profileRepo.Delete(ctx, profile.ID)
	require.NoError(t, err)
}

func TestAuthFlowRepositoryCRUD(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	repoFactory := setupTestRepositoryFactory(ctx, t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	flowRepo := repoFactory.AuthFlowRepository()

	// Test Create
	flow := &cryptoutilIdentityDomain.AuthFlow{
		Name:     "test-flow",
		FlowType: cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
	}

	err := flowRepo.Create(ctx, flow)
	require.NoError(t, err)
	require.NotEmpty(t, flow.ID)

	// Test GetByID
	retrievedFlow, err := flowRepo.GetByID(ctx, flow.ID)
	require.NoError(t, err)
	require.Equal(t, flow.ID, retrievedFlow.ID)

	// Test GetByName
	flowByName, err := flowRepo.GetByName(ctx, flow.Name)
	require.NoError(t, err)
	require.Equal(t, flow.Name, flowByName.Name)

	// Test Update
	updatedFlow := *retrievedFlow
	updatedFlow.Description = "Updated test flow"
	err = flowRepo.Update(ctx, &updatedFlow)
	require.NoError(t, err)

	// Test Delete
	err = flowRepo.Delete(ctx, flow.ID)
	require.NoError(t, err)
}

