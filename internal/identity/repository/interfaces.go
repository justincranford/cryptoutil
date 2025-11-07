package repository

import (
	"context"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// UserRepository defines operations for user persistence.
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, user *cryptoutilIdentityDomain.User) error

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.User, error)

	// GetBySub retrieves a user by subject identifier.
	GetBySub(ctx context.Context, sub string) (*cryptoutilIdentityDomain.User, error)

	// GetByUsername retrieves a user by preferred username.
	GetByUsername(ctx context.Context, username string) (*cryptoutilIdentityDomain.User, error)

	// GetByEmail retrieves a user by email address.
	GetByEmail(ctx context.Context, email string) (*cryptoutilIdentityDomain.User, error)

	// Update updates an existing user.
	Update(ctx context.Context, user *cryptoutilIdentityDomain.User) error

	// Delete deletes a user by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// List lists users with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.User, error)

	// Count returns the total number of users.
	Count(ctx context.Context) (int64, error)
}

// ClientRepository defines operations for OAuth client persistence.
type ClientRepository interface {
	// Create creates a new client.
	Create(ctx context.Context, client *cryptoutilIdentityDomain.Client) error

	// GetByID retrieves a client by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Client, error)

	// GetByClientID retrieves a client by client ID.
	GetByClientID(ctx context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error)

	// Update updates an existing client.
	Update(ctx context.Context, client *cryptoutilIdentityDomain.Client) error

	// Delete deletes a client by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// List lists clients with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Client, error)

	// Count returns the total number of clients.
	Count(ctx context.Context) (int64, error)
}

// TokenRepository defines operations for token persistence.
type TokenRepository interface {
	// Create creates a new token.
	Create(ctx context.Context, token *cryptoutilIdentityDomain.Token) error

	// GetByID retrieves a token by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Token, error)

	// GetByTokenValue retrieves a token by its value.
	GetByTokenValue(ctx context.Context, tokenValue string) (*cryptoutilIdentityDomain.Token, error)

	// Update updates an existing token.
	Update(ctx context.Context, token *cryptoutilIdentityDomain.Token) error

	// Delete deletes a token by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// RevokeByID revokes a token by ID.
	RevokeByID(ctx context.Context, id googleUuid.UUID) error

	// RevokeByTokenValue revokes a token by its value.
	RevokeByTokenValue(ctx context.Context, tokenValue string) error

	// DeleteExpired deletes all expired tokens.
	DeleteExpired(ctx context.Context) error

	// List lists tokens with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Token, error)

	// Count returns the total number of tokens.
	Count(ctx context.Context) (int64, error)
}

// SessionRepository defines operations for session persistence.
type SessionRepository interface {
	// Create creates a new session.
	Create(ctx context.Context, session *cryptoutilIdentityDomain.Session) error

	// GetByID retrieves a session by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Session, error)

	// GetBySessionID retrieves a session by session ID.
	GetBySessionID(ctx context.Context, sessionID string) (*cryptoutilIdentityDomain.Session, error)

	// Update updates an existing session.
	Update(ctx context.Context, session *cryptoutilIdentityDomain.Session) error

	// Delete deletes a session by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// TerminateByID terminates a session by ID.
	TerminateByID(ctx context.Context, id googleUuid.UUID) error

	// TerminateBySessionID terminates a session by session ID.
	TerminateBySessionID(ctx context.Context, sessionID string) error

	// DeleteExpired deletes all expired sessions.
	DeleteExpired(ctx context.Context) error

	// List lists sessions with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Session, error)

	// Count returns the total number of sessions.
	Count(ctx context.Context) (int64, error)
}

// ClientProfileRepository defines operations for client profile persistence.
type ClientProfileRepository interface {
	// Create creates a new client profile.
	Create(ctx context.Context, profile *cryptoutilIdentityDomain.ClientProfile) error

	// GetByID retrieves a client profile by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.ClientProfile, error)

	// GetByName retrieves a client profile by name.
	GetByName(ctx context.Context, name string) (*cryptoutilIdentityDomain.ClientProfile, error)

	// Update updates an existing client profile.
	Update(ctx context.Context, profile *cryptoutilIdentityDomain.ClientProfile) error

	// Delete deletes a client profile by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// List lists client profiles with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.ClientProfile, error)

	// Count returns the total number of client profiles.
	Count(ctx context.Context) (int64, error)
}

// AuthFlowRepository defines operations for authorization flow persistence.
type AuthFlowRepository interface {
	// Create creates a new authorization flow.
	Create(ctx context.Context, flow *cryptoutilIdentityDomain.AuthFlow) error

	// GetByID retrieves an authorization flow by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.AuthFlow, error)

	// GetByName retrieves an authorization flow by name.
	GetByName(ctx context.Context, name string) (*cryptoutilIdentityDomain.AuthFlow, error)

	// Update updates an existing authorization flow.
	Update(ctx context.Context, flow *cryptoutilIdentityDomain.AuthFlow) error

	// Delete deletes an authorization flow by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// List lists authorization flows with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.AuthFlow, error)

	// Count returns the total number of authorization flows.
	Count(ctx context.Context) (int64, error)
}

// AuthProfileRepository defines operations for authentication profile persistence.
type AuthProfileRepository interface {
	// Create creates a new authentication profile.
	Create(ctx context.Context, profile *cryptoutilIdentityDomain.AuthProfile) error

	// GetByID retrieves an authentication profile by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.AuthProfile, error)

	// GetByName retrieves an authentication profile by name.
	GetByName(ctx context.Context, name string) (*cryptoutilIdentityDomain.AuthProfile, error)

	// Update updates an existing authentication profile.
	Update(ctx context.Context, profile *cryptoutilIdentityDomain.AuthProfile) error

	// Delete deletes an authentication profile by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// List lists authentication profiles with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.AuthProfile, error)

	// Count returns the total number of authentication profiles.
	Count(ctx context.Context) (int64, error)
}

// MFAFactorRepository defines operations for MFA factor persistence.
type MFAFactorRepository interface {
	// Create creates a new MFA factor.
	Create(ctx context.Context, factor *cryptoutilIdentityDomain.MFAFactor) error

	// GetByID retrieves an MFA factor by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.MFAFactor, error)

	// GetByAuthProfileID retrieves all MFA factors for an authentication profile.
	GetByAuthProfileID(ctx context.Context, authProfileID googleUuid.UUID) ([]*cryptoutilIdentityDomain.MFAFactor, error)

	// Update updates an existing MFA factor.
	Update(ctx context.Context, factor *cryptoutilIdentityDomain.MFAFactor) error

	// Delete deletes an MFA factor by ID (soft delete).
	Delete(ctx context.Context, id googleUuid.UUID) error

	// List lists MFA factors with pagination.
	List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.MFAFactor, error)

	// Count returns the total number of MFA factors.
	Count(ctx context.Context) (int64, error)
}
