// Copyright (c) 2025 Justin Cranford
//
//

package fixtures

import (
	"context"
	"fmt"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestDataHelper provides utilities for setting up and cleaning test data.
type TestDataHelper struct {
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	ctx         context.Context
}

// NewTestDataHelper creates a new test data helper.
func NewTestDataHelper(ctx context.Context, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *TestDataHelper {
	return &TestDataHelper{
		repoFactory: repoFactory,
		ctx:         ctx,
	}
}

// CreateTestUser creates a test user using the builder pattern.
func (h *TestDataHelper) CreateTestUser(builder *TestUserBuilder) (*cryptoutilIdentityDomain.User, error) {
	user := builder.Build()
	userRepo := h.repoFactory.UserRepository()

	err := userRepo.Create(h.ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}

	return user, nil
}

// CreateTestClient creates a test client using the builder pattern.
func (h *TestDataHelper) CreateTestClient(builder *TestClientBuilder) (*cryptoutilIdentityDomain.Client, error) {
	client := builder.Build()
	clientRepo := h.repoFactory.ClientRepository()

	err := clientRepo.Create(h.ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create test client: %w", err)
	}

	return client, nil
}

// CreateTestToken creates a test token using the builder pattern.
func (h *TestDataHelper) CreateTestToken(builder *TestTokenBuilder) (*cryptoutilIdentityDomain.Token, error) {
	token := builder.Build()
	tokenRepo := h.repoFactory.TokenRepository()

	err := tokenRepo.Create(h.ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create test token: %w", err)
	}

	return token, nil
}

// CreateTestSession creates a test session using the builder pattern.
func (h *TestDataHelper) CreateTestSession(builder *TestSessionBuilder) (*cryptoutilIdentityDomain.Session, error) {
	session := builder.Build()
	sessionRepo := h.repoFactory.SessionRepository()

	err := sessionRepo.Create(h.ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create test session: %w", err)
	}

	return session, nil
}

// CleanupTestData removes all test data created during testing.
func (h *TestDataHelper) CleanupTestData() error {
	// Note: In a real implementation, you might want to track created entities
	// and clean them up specifically. For now, this is a placeholder.
	return nil
}

// TestScenario represents a complete test scenario with related entities.
type TestScenario struct {
	User    *cryptoutilIdentityDomain.User
	Client  *cryptoutilIdentityDomain.Client
	Token   *cryptoutilIdentityDomain.Token
	Session *cryptoutilIdentityDomain.Session
}

// CreateTestScenario creates a complete test scenario with related entities.
func (h *TestDataHelper) CreateTestScenario() (*TestScenario, error) {
	// Create a test user
	userBuilder := NewTestUserBuilder().
		WithSub("test-scenario-user").
		WithEmail("scenario@example.com").
		WithName("Test Scenario User")

	user, err := h.CreateTestUser(userBuilder)
	if err != nil {
		return nil, err
	}

	// Create a test client
	clientBuilder := NewTestClientBuilder().
		WithClientID("test-scenario-client").
		WithName("Test Scenario Client")

	client, err := h.CreateTestClient(clientBuilder)
	if err != nil {
		return nil, err
	}

	// Create a test token
	tokenBuilder := NewTestTokenBuilder().
		WithClientID(client.ID).
		WithUserID(&user.ID).
		WithTokenValue("test-scenario-token")

	token, err := h.CreateTestToken(tokenBuilder)
	if err != nil {
		return nil, err
	}

	// Create a test session
	sessionBuilder := NewTestSessionBuilder().
		WithUserID(user.ID).
		WithSessionID("test-scenario-session")

	session, err := h.CreateTestSession(sessionBuilder)
	if err != nil {
		return nil, err
	}

	return &TestScenario{
		User:    user,
		Client:  client,
		Token:   token,
		Session: session,
	}, nil
}

// CleanupTestScenario removes all entities from a test scenario.
func (h *TestDataHelper) CleanupTestScenario(scenario *TestScenario) error {
	if scenario.Token != nil {
		tokenRepo := h.repoFactory.TokenRepository()
		if err := tokenRepo.Delete(h.ctx, scenario.Token.ID); err != nil {
			return fmt.Errorf("failed to delete test token: %w", err)
		}
	}

	if scenario.Session != nil {
		sessionRepo := h.repoFactory.SessionRepository()
		if err := sessionRepo.Delete(h.ctx, scenario.Session.ID); err != nil {
			return fmt.Errorf("failed to delete test session: %w", err)
		}
	}

	if scenario.Client != nil {
		clientRepo := h.repoFactory.ClientRepository()
		if err := clientRepo.Delete(h.ctx, scenario.Client.ID); err != nil {
			return fmt.Errorf("failed to delete test client: %w", err)
		}
	}

	if scenario.User != nil {
		userRepo := h.repoFactory.UserRepository()
		if err := userRepo.Delete(h.ctx, scenario.User.ID); err != nil {
			return fmt.Errorf("failed to delete test user: %w", err)
		}
	}

	return nil
}
