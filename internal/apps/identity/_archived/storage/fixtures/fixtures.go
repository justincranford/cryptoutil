// Copyright (c) 2025 Justin Cranford
//
//

// Package fixtures provides test fixtures for identity storage testing.
package fixtures

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"time"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"

	googleUuid "github.com/google/uuid"
)

// TestUserBuilder builds test user entities with fluent API.
type TestUserBuilder struct {
	user *cryptoutilIdentityDomain.User
}

// NewTestUserBuilder creates a new test user builder with default values.
func NewTestUserBuilder() *TestUserBuilder {
	return &TestUserBuilder{
		user: &cryptoutilIdentityDomain.User{
			Sub:          googleUuid.Must(googleUuid.NewV7()).String(),
			Email:        "test@example.com",
			Name:         "Test User",
			Enabled:      true,
			PasswordHash: "pbkdf2$210000$Gpdnumx30ru2iTk0hkEdvQ$4KSkexHlyfwwlhVHm2f/1KqZWewwlQmy0GMvAoeFxsQ", // PBKDF2 hash for "password"
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		},
	}
}

// WithSub sets the subject identifier.
func (b *TestUserBuilder) WithSub(sub string) *TestUserBuilder {
	b.user.Sub = sub

	return b
}

// WithEmail sets the email address.
func (b *TestUserBuilder) WithEmail(email string) *TestUserBuilder {
	b.user.Email = email

	return b
}

// WithName sets the full name.
func (b *TestUserBuilder) WithName(name string) *TestUserBuilder {
	b.user.Name = name

	return b
}

// WithEnabled sets the enabled status.
func (b *TestUserBuilder) WithEnabled(enabled bool) *TestUserBuilder {
	b.user.Enabled = enabled

	return b
}

// Build returns the built user entity.
func (b *TestUserBuilder) Build() *cryptoutilIdentityDomain.User {
	return b.user
}

// TestClientBuilder builds test client entities with fluent API.
type TestClientBuilder struct {
	client *cryptoutilIdentityDomain.Client
}

// NewTestClientBuilder creates a new test client builder with default values.
func NewTestClientBuilder() *TestClientBuilder {
	enabled := true

	return &TestClientBuilder{
		client: &cryptoutilIdentityDomain.Client{
			ClientID:   googleUuid.Must(googleUuid.NewV7()).String(),
			ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:       "Test Client",
			Enabled:    &enabled,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		},
	}
}

// WithClientID sets the client identifier.
func (b *TestClientBuilder) WithClientID(clientID string) *TestClientBuilder {
	b.client.ClientID = clientID

	return b
}

// WithClientType sets the client type.
func (b *TestClientBuilder) WithClientType(clientType cryptoutilIdentityDomain.ClientType) *TestClientBuilder {
	b.client.ClientType = clientType

	return b
}

// WithName sets the client name.
func (b *TestClientBuilder) WithName(name string) *TestClientBuilder {
	b.client.Name = name

	return b
}

// WithEnabled sets the enabled status.
func (b *TestClientBuilder) WithEnabled(enabled bool) *TestClientBuilder {
	b.client.Enabled = &enabled

	return b
}

// Build returns the built client entity.
func (b *TestClientBuilder) Build() *cryptoutilIdentityDomain.Client {
	return b.client
}

// TestTokenBuilder builds test token entities with fluent API.
type TestTokenBuilder struct {
	token *cryptoutilIdentityDomain.Token
}

// NewTestTokenBuilder creates a new test token builder with default values.
func NewTestTokenBuilder() *TestTokenBuilder {
	return &TestTokenBuilder{
		token: &cryptoutilIdentityDomain.Token{
			TokenValue:  googleUuid.Must(googleUuid.NewV7()).String(),
			TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
			TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
			ClientID:    googleUuid.Must(googleUuid.NewV7()), // Will be overridden in tests
			Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
			IssuedAt:    time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(time.Hour),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}
}

// WithTokenValue sets the token value.
func (b *TestTokenBuilder) WithTokenValue(tokenValue string) *TestTokenBuilder {
	b.token.TokenValue = tokenValue

	return b
}

// WithClientID sets the client ID.
func (b *TestTokenBuilder) WithClientID(clientID googleUuid.UUID) *TestTokenBuilder {
	b.token.ClientID = clientID

	return b
}

// WithUserID sets the user ID.
func (b *TestTokenBuilder) WithUserID(userID *googleUuid.UUID) *TestTokenBuilder {
	b.token.UserID = cryptoutilIdentityDomain.NewNullableUUID(userID)

	return b
}

// WithScopes sets the token scopes.
func (b *TestTokenBuilder) WithScopes(scopes []string) *TestTokenBuilder {
	b.token.Scopes = scopes

	return b
}

// WithExpiresAt sets the expiration time.
func (b *TestTokenBuilder) WithExpiresAt(expiresAt time.Time) *TestTokenBuilder {
	b.token.ExpiresAt = expiresAt

	return b
}

// Build returns the built token entity.
func (b *TestTokenBuilder) Build() *cryptoutilIdentityDomain.Token {
	return b.token
}

// TestSessionBuilder builds test session entities with fluent API.
type TestSessionBuilder struct {
	session *cryptoutilIdentityDomain.Session
}

// NewTestSessionBuilder creates a new builder for session test data.
func NewTestSessionBuilder() *TestSessionBuilder {
	active := true

	return &TestSessionBuilder{
		session: &cryptoutilIdentityDomain.Session{
			SessionID: googleUuid.Must(googleUuid.NewV7()).String(),
			UserID:    googleUuid.Must(googleUuid.NewV7()), // Will be overridden in tests
			Active:    &active,
			IssuedAt:  time.Now().UTC(),
			ExpiresAt: time.Now().UTC().Add(time.Hour),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
}

// WithSessionID sets the session identifier.
func (b *TestSessionBuilder) WithSessionID(sessionID string) *TestSessionBuilder {
	b.session.SessionID = sessionID

	return b
}

// WithUserID sets the user ID.
func (b *TestSessionBuilder) WithUserID(userID googleUuid.UUID) *TestSessionBuilder {
	b.session.UserID = userID

	return b
}

// WithActive sets the active status.
func (b *TestSessionBuilder) WithActive(active bool) *TestSessionBuilder {
	b.session.Active = &active

	return b
}

// Build returns the built session entity.
func (b *TestSessionBuilder) Build() *cryptoutilIdentityDomain.Session {
	return b.session
}
