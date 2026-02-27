// Copyright (c) 2025 Justin Cranford
//
//

// Package auth provides authentication mechanisms including email-password, MFA, and other auth flows.
package auth

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// EmailPasswordProfile implements email/password authentication.
type EmailPasswordProfile struct {
	userRepo cryptoutilIdentityRepository.UserRepository
	hasher   *cryptoutilIdentityClientAuth.PBKDF2Hasher
}

// NewEmailPasswordProfile creates a new email/password authentication profile.
func NewEmailPasswordProfile(userRepo cryptoutilIdentityRepository.UserRepository) *EmailPasswordProfile {
	return &EmailPasswordProfile{
		userRepo: userRepo,
		hasher:   cryptoutilIdentityClientAuth.NewPBKDF2Hasher(),
	}
}

// Name returns the profile name.
func (p *EmailPasswordProfile) Name() string {
	return "email_password"
}

// Authenticate performs email/password authentication.
func (p *EmailPasswordProfile) Authenticate(ctx context.Context, credentials map[string]string) (*cryptoutilIdentityDomain.User, error) {
	email, ok := credentials[cryptoutilSharedMagic.ClaimEmail]
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

	// Validate password hash using PBKDF2-HMAC-SHA256 (FIPS 140-3 approved).
	if err := p.hasher.CompareSecret(user.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("%w: invalid credentials", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	return user, nil
}

// RequiresMFA indicates whether this profile requires multi-factor authentication.
func (p *EmailPasswordProfile) RequiresMFA() bool {
	return true // Email/password typically requires 2FA.
}

// ValidateCredentials validates the credential format.
func (p *EmailPasswordProfile) ValidateCredentials(credentials map[string]string) error {
	email, ok := credentials[cryptoutilSharedMagic.ClaimEmail]
	if !ok || email == "" {
		return fmt.Errorf("%w: missing email", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	password, ok := credentials["password"]
	if !ok || password == "" {
		return fmt.Errorf("%w: missing password", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	return nil
}
