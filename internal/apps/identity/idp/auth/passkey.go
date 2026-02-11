// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// PasskeyProfile implements WebAuthn passkey authentication.
type PasskeyProfile struct {
	userRepo cryptoutilIdentityRepository.UserRepository
	mfaRepo  cryptoutilIdentityRepository.MFAFactorRepository
}

// NewPasskeyProfile creates a new WebAuthn passkey authentication profile.
func NewPasskeyProfile(userRepo cryptoutilIdentityRepository.UserRepository, mfaRepo cryptoutilIdentityRepository.MFAFactorRepository) *PasskeyProfile {
	return &PasskeyProfile{
		userRepo: userRepo,
		mfaRepo:  mfaRepo,
	}
}

// Name returns the profile name.
func (p *PasskeyProfile) Name() string {
	return "passkey"
}

// Authenticate performs WebAuthn passkey authentication.
func (p *PasskeyProfile) Authenticate(_ context.Context, credentials map[string]string) (*cryptoutilIdentityDomain.User, error) {
	credentialID, ok := credentials["credential_id"]
	if !ok || credentialID == "" {
		return nil, fmt.Errorf("%w: missing credential_id", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	assertion, ok := credentials["assertion"]
	if !ok || assertion == "" {
		return nil, fmt.Errorf("%w: missing assertion", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	// TODO: Validate WebAuthn assertion using library (e.g., go-webauthn/webauthn).
	// TODO: Fetch user by credential ID.
	// TODO: Verify signature and challenge.
	// TODO: Return user object if validation succeeds.

	_ = credentialID
	_ = assertion

	return nil, fmt.Errorf("%w: passkey validation not implemented", cryptoutilIdentityAppErr.ErrServerError)
}

// RequiresMFA indicates whether this profile requires multi-factor authentication.
func (p *PasskeyProfile) RequiresMFA() bool {
	return false // Passkeys are inherently multi-factor (possession + biometrics/PIN).
}

// ValidateCredentials validates the credential format.
func (p *PasskeyProfile) ValidateCredentials(credentials map[string]string) error {
	credentialID, ok := credentials["credential_id"]
	if !ok || credentialID == "" {
		return fmt.Errorf("%w: missing credential_id", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	assertion, ok := credentials["assertion"]
	if !ok || assertion == "" {
		return fmt.Errorf("%w: missing assertion", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	return nil
}
