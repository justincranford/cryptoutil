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

// TOTPProfile implements TOTP/HOTP authentication.
type TOTPProfile struct {
	mfaRepo cryptoutilIdentityRepository.MFAFactorRepository
}

// NewTOTPProfile creates a new TOTP/HOTP authentication profile.
func NewTOTPProfile(mfaRepo cryptoutilIdentityRepository.MFAFactorRepository) *TOTPProfile {
	return &TOTPProfile{
		mfaRepo: mfaRepo,
	}
}

// Name returns the profile name.
func (p *TOTPProfile) Name() string {
	return string(OTPMethodTOTP)
}

// Authenticate performs TOTP/HOTP authentication.
func (p *TOTPProfile) Authenticate(_ context.Context, credentials map[string]string) (*cryptoutilIdentityDomain.User, error) {
	userID, ok := credentials["user_id"]
	if !ok || userID == "" {
		return nil, fmt.Errorf("%w: missing user_id", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	otpCode, ok := credentials["otp_code"]
	if !ok || otpCode == "" {
		return nil, fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	// TODO: Fetch MFA factors for user.
	// TODO: Validate TOTP/HOTP code using library (e.g., pquerna/otp).
	// TODO: Return user object if validation succeeds.

	_ = userID
	_ = otpCode

	return nil, fmt.Errorf("%w: TOTP validation not implemented", cryptoutilIdentityAppErr.ErrServerError)
}

// RequiresMFA indicates whether this profile requires multi-factor authentication.
func (p *TOTPProfile) RequiresMFA() bool {
	return false // TOTP itself is an MFA factor.
}

// ValidateCredentials validates the credential format.
func (p *TOTPProfile) ValidateCredentials(credentials map[string]string) error {
	userID, ok := credentials["user_id"]
	if !ok || userID == "" {
		return fmt.Errorf("%w: missing user_id", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	otpCode, ok := credentials["otp_code"]
	if !ok || otpCode == "" {
		return fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	return nil
}
