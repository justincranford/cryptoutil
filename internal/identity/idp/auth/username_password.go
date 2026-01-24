// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"
	"fmt"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
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
		return nil, fmt.Errorf("%w: missing username", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	password, ok := credentials["password"]
	if !ok || password == "" {
		return nil, fmt.Errorf("%w: missing password", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	// Fetch user by username.
	user, err := p.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%w: user lookup failed: %w", cryptoutilIdentityAppErr.ErrInvalidCredentials, err)
	}

	// Check if account is enabled.
	if !user.Enabled {
		return nil, fmt.Errorf("%w: account disabled", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	// Check if account is locked.
	if user.Locked {
		return nil, fmt.Errorf("%w: account locked", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	// Validate password hash using configured crypto wrapper (PBKDF2 default). Supports legacy hashes.
	ok, verr := cryptoutilSharedCryptoDigests.VerifySecret(user.PasswordHash, password)
	if verr != nil {
		return nil, fmt.Errorf("%w: password verification error: %w", cryptoutilIdentityAppErr.ErrInvalidCredentials, verr)
	}

	if !ok {
		return nil, fmt.Errorf("%w: invalid password", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

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
		return fmt.Errorf("%w: missing username", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	password, ok := credentials["password"]
	if !ok || password == "" {
		return fmt.Errorf("%w: missing password", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	return nil
}
