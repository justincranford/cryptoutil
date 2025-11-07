package auth

import (
	"context"
	"fmt"

	cryptoutilIdentityApperr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// UsernamePasswordProfile implements username/password authentication.
type UsernamePasswordProfile struct {
	userRepo cryptoutilIdentityRepository.UserRepository
}

// NewUsernamePasswordProfile creates a new username/password authentication profile.
func NewUsernamePasswordProfile(userRepo cryptoutilIdentityRepository.UserRepository) *UsernamePasswordProfile {
	return &UsernamePasswordProfile{
		userRepo: userRepo,
	}
}

// Name returns the profile name.
func (p *UsernamePasswordProfile) Name() string {
	return "username_password"
}

// Authenticate performs username/password authentication.
func (p *UsernamePasswordProfile) Authenticate(ctx context.Context, credentials map[string]string) (*cryptoutilIdentityDomain.User, error) {
	username, ok := credentials["username"]
	if !ok || username == "" {
		return nil, fmt.Errorf("%w: missing username", cryptoutilIdentityApperr.ErrInvalidCredentials)
	}

	password, ok := credentials["password"]
	if !ok || password == "" {
		return nil, fmt.Errorf("%w: missing password", cryptoutilIdentityApperr.ErrInvalidCredentials)
	}

	// Fetch user by username.
	user, err := p.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%w: user lookup failed: %w", cryptoutilIdentityApperr.ErrInvalidCredentials, err)
	}

	// TODO: Validate password hash using bcrypt or argon2.
	// For now, this is a placeholder that always succeeds if the user exists.
	_ = password

	return user, nil
}

// RequiresMFA indicates whether this profile requires multi-factor authentication.
func (p *UsernamePasswordProfile) RequiresMFA() bool {
	return false
}

// ValidateCredentials validates the credential format.
func (p *UsernamePasswordProfile) ValidateCredentials(credentials map[string]string) error {
	username, ok := credentials["username"]
	if !ok || username == "" {
		return fmt.Errorf("%w: missing username", cryptoutilIdentityApperr.ErrInvalidCredentials)
	}

	password, ok := credentials["password"]
	if !ok || password == "" {
		return fmt.Errorf("%w: missing password", cryptoutilIdentityApperr.ErrInvalidCredentials)
	}

	return nil
}
