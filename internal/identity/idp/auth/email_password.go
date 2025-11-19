// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// EmailPasswordProfile implements email/password authentication.
type EmailPasswordProfile struct {
	userRepo cryptoutilIdentityRepository.UserRepository
}

// NewEmailPasswordProfile creates a new email/password authentication profile.
func NewEmailPasswordProfile(userRepo cryptoutilIdentityRepository.UserRepository) *EmailPasswordProfile {
	return &EmailPasswordProfile{
		userRepo: userRepo,
	}
}

// Name returns the profile name.
func (p *EmailPasswordProfile) Name() string {
	return "email_password"
}

// Authenticate performs email/password authentication.
func (p *EmailPasswordProfile) Authenticate(ctx context.Context, credentials map[string]string) (*cryptoutilIdentityDomain.User, error) {
	email, ok := credentials["email"]
	if !ok || email == "" {
		return nil, fmt.Errorf("%w: missing email", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	password, ok := credentials["password"]
	if !ok || password == "" {
		return nil, fmt.Errorf("%w: missing password", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	// Fetch user by email.
	user, err := p.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%w: user lookup failed: %w", cryptoutilIdentityAppErr.ErrInvalidCredentials, err)
	}

	// TODO: Validate password hash using bcrypt or argon2.
	// For now, this is a placeholder that always succeeds if the user exists.
	_ = password

	return user, nil
}

// RequiresMFA indicates whether this profile requires multi-factor authentication.
func (p *EmailPasswordProfile) RequiresMFA() bool {
	return true // Email/password typically requires 2FA.
}

// ValidateCredentials validates the credential format.
func (p *EmailPasswordProfile) ValidateCredentials(credentials map[string]string) error {
	email, ok := credentials["email"]
	if !ok || email == "" {
		return fmt.Errorf("%w: missing email", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	password, ok := credentials["password"]
	if !ok || password == "" {
		return fmt.Errorf("%w: missing password", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	return nil
}
