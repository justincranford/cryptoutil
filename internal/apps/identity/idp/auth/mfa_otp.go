// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

const (
	totpPeriod30Seconds  = 30
	totpPeriod300Seconds = 300
	totpPeriod600Seconds = 600
)

// TOTPValidator handles Time-based One-Time Password validation.
type TOTPValidator struct {
	secretStore OTPSecretStore
}

// OTPSecretStore defines interface for retrieving TOTP secrets.
type OTPSecretStore interface {
	GetTOTPSecret(ctx context.Context, userID string) (string, error)
	GetEmailOTPSecret(ctx context.Context, userID string) (string, error)
	GetSMSOTPSecret(ctx context.Context, userID string) (string, error)
}

// NewTOTPValidator creates a new TOTP validator.
func NewTOTPValidator(secretStore OTPSecretStore) *TOTPValidator {
	return &TOTPValidator{
		secretStore: secretStore,
	}
}

// ValidateTOTP validates a TOTP code against the user's stored secret.
func (v *TOTPValidator) ValidateTOTP(ctx context.Context, userID string, code string) (bool, error) {
	secret, err := v.secretStore.GetTOTPSecret(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve TOTP secret for user %s: %w", userID, err)
	}

	// Validate TOTP code with current time.
	valid := totp.Validate(code, secret)
	if !valid {
		return false, nil
	}

	return true, nil
}

// ValidateTOTPWithWindow validates a TOTP code with a time window (allows clock skew).
func (v *TOTPValidator) ValidateTOTPWithWindow(ctx context.Context, userID string, code string, windowSize uint) (bool, error) {
	secret, err := v.secretStore.GetTOTPSecret(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve TOTP secret for user %s: %w", userID, err)
	}

	// Validate with time window (e.g., windowSize=1 allows 30s before/after current time).
	valid, err := totp.ValidateCustom(code, secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    totpPeriod30Seconds, // Standard TOTP period (30 seconds)
		Skew:      windowSize,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false, fmt.Errorf("TOTP validation failed: %w", err)
	}

	return valid, nil
}

// ValidateEmailOTP validates an email-based OTP code.
func (v *TOTPValidator) ValidateEmailOTP(ctx context.Context, userID string, code string) (bool, error) {
	secret, err := v.secretStore.GetEmailOTPSecret(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve email OTP secret for user %s: %w", userID, err)
	}

	// Email OTP typically uses longer period (e.g., 5-10 minutes).
	valid, err := totp.ValidateCustom(code, secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    totpPeriod300Seconds, // 5 minutes
		Skew:      1,                    // Allow 1 period before/after
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA256, // More secure for email
	})
	if err != nil {
		return false, fmt.Errorf("email OTP validation failed: %w", err)
	}

	return valid, nil
}

// ValidateSMSOTP validates an SMS-based OTP code.
func (v *TOTPValidator) ValidateSMSOTP(ctx context.Context, userID string, code string) (bool, error) {
	secret, err := v.secretStore.GetSMSOTPSecret(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve SMS OTP secret for user %s: %w", userID, err)
	}

	// SMS OTP typically uses longer period (e.g., 10 minutes).
	valid, err := totp.ValidateCustom(code, secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    totpPeriod600Seconds, // 10 minutes
		Skew:      1,                    // Allow 1 period before/after
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA256, // More secure for SMS
	})
	if err != nil {
		return false, fmt.Errorf("SMS OTP validation failed: %w", err)
	}

	return valid, nil
}

// IntegrateTOTPValidation integrates TOTP validation into MFAOrchestrator.ValidateFactor.
func (o *MFAOrchestrator) IntegrateTOTPValidation(ctx context.Context, factor *cryptoutilIdentityDomain.MFAFactor, credentials map[string]string) (bool, error) {
	// Extract OTP code from credentials.
	code, ok := credentials["otp_code"]
	if !ok {
		return false, fmt.Errorf("missing otp_code in credentials")
	}

	// Determine factor type and validate accordingly.
	switch factor.FactorType {
	case cryptoutilIdentityDomain.MFAFactorTypeTOTP:
		// TOTP validation with 1-period window (30s before/after).
		if o.totpValidator == nil {
			return false, fmt.Errorf("TOTP validator not configured")
		}

		// Extract user ID from auth profile (factor doesn't have UserID).
		// TODO: Retrieve user ID from authentication context.
		userID := factor.AuthProfileID.String() // Placeholder: use auth profile ID.

		valid, err := o.totpValidator.ValidateTOTPWithWindow(ctx, userID, code, 1)
		if err != nil {
			return false, fmt.Errorf("TOTP validation error: %w", err)
		}

		return valid, nil

	case cryptoutilIdentityDomain.MFAFactorTypeEmailOTP:
		// Email OTP validation with 5-minute period.
		if o.totpValidator == nil {
			return false, fmt.Errorf("TOTP validator not configured")
		}

		userID := factor.AuthProfileID.String()

		valid, err := o.totpValidator.ValidateEmailOTP(ctx, userID, code)
		if err != nil {
			return false, fmt.Errorf("email OTP validation error: %w", err)
		}

		return valid, nil

	case cryptoutilIdentityDomain.MFAFactorTypeSMSOTP:
		// SMS OTP validation with 10-minute period.
		if o.totpValidator == nil {
			return false, fmt.Errorf("TOTP validator not configured")
		}

		userID := factor.AuthProfileID.String()

		valid, err := o.totpValidator.ValidateSMSOTP(ctx, userID, code)
		if err != nil {
			return false, fmt.Errorf("SMS OTP validation error: %w", err)
		}

		return valid, nil

	case cryptoutilIdentityDomain.MFAFactorTypePassword,
		cryptoutilIdentityDomain.MFAFactorTypeHOTP,
		cryptoutilIdentityDomain.MFAFactorTypePasskey,
		cryptoutilIdentityDomain.MFAFactorTypeMagicLink,
		cryptoutilIdentityDomain.MFAFactorTypeMTLS,
		cryptoutilIdentityDomain.MFAFactorTypeHardwareToken,
		cryptoutilIdentityDomain.MFAFactorTypeBiometric:
		return false, fmt.Errorf("unsupported OTP factor type: %s", factor.FactorType)
	}

	return false, fmt.Errorf("unknown factor type: %s", factor.FactorType)
}
