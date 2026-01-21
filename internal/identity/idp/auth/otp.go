// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// OTPMethod represents the OTP delivery method.
type OTPMethod string

const (
	// OTPMethodEmail delivers OTP via email.
	OTPMethodEmail OTPMethod = "email"

	// OTPMethodSMS delivers OTP via SMS.
	OTPMethodSMS OTPMethod = "sms"

	// OTPMethodTOTP uses time-based OTP (TOTP).
	OTPMethodTOTP OTPMethod = "totp"
)

// OTPService handles OTP generation and validation.
type OTPService struct {
	// TODO: Add dependencies for email/SMS delivery.
}

// NewOTPService creates a new OTP service.
func NewOTPService() *OTPService {
	return &OTPService{}
}

// GenerateOTP generates a one-time password for the specified user.
func (s *OTPService) GenerateOTP(_ context.Context, user *cryptoutilIdentityDomain.User, method OTPMethod) (string, error) {
	// TODO: Generate OTP code (6-digit numeric).
	// TODO: Store OTP with expiration (5 minutes).
	// TODO: Send OTP via email/SMS based on method.
	_ = user
	_ = method

	return "", fmt.Errorf("%w: OTP generation not implemented", cryptoutilIdentityAppErr.ErrServerError)
}

// ValidateOTP validates a one-time password for the specified user.
func (s *OTPService) ValidateOTP(_ context.Context, user *cryptoutilIdentityDomain.User, otpCode string, method OTPMethod) error {
	// TODO: Fetch stored OTP for user.
	// TODO: Validate OTP code matches.
	// TODO: Check OTP not expired.
	// TODO: Invalidate OTP after successful validation.
	_ = user
	_ = otpCode
	_ = method

	return fmt.Errorf("%w: OTP validation not implemented", cryptoutilIdentityAppErr.ErrServerError)
}
