package auth

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// MFAOrchestrator manages multi-factor authentication flows.
type MFAOrchestrator struct {
	mfaRepo         cryptoutilIdentityRepository.MFAFactorRepository
	otpService      *OTPService
	profileRegistry *ProfileRegistry
}

// NewMFAOrchestrator creates a new MFA orchestrator.
func NewMFAOrchestrator(
	mfaRepo cryptoutilIdentityRepository.MFAFactorRepository,
	otpService *OTPService,
	profileRegistry *ProfileRegistry,
) *MFAOrchestrator {
	return &MFAOrchestrator{
		mfaRepo:         mfaRepo,
		otpService:      otpService,
		profileRegistry: profileRegistry,
	}
}

// GetRequiredFactors returns the required MFA factors for an authentication profile.
func (o *MFAOrchestrator) GetRequiredFactors(ctx context.Context, authProfileID googleUuid.UUID) ([]string, error) {
	// Fetch MFA factors for authentication profile.
	factors, err := o.mfaRepo.GetByAuthProfileID(ctx, authProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MFA factors: %w", err)
	}

	// Extract factor types.
	factorTypes := make([]string, 0, len(factors))
	for _, factor := range factors {
		factorTypes = append(factorTypes, string(factor.FactorType))
	}

	return factorTypes, nil
}

// ValidateFactor validates a specific MFA factor.
func (o *MFAOrchestrator) ValidateFactor(ctx context.Context, authProfileID googleUuid.UUID, factorType string, credentials map[string]string) error {
	// Fetch MFA factors for authentication profile.
	factors, err := o.mfaRepo.GetByAuthProfileID(ctx, authProfileID)
	if err != nil {
		return fmt.Errorf("failed to fetch MFA factors: %w", err)
	}

	// Find matching factor.
	var matchingFactor *cryptoutilIdentityDomain.MFAFactor

	for _, factor := range factors {
		if string(factor.FactorType) == factorType {
			matchingFactor = factor

			break
		}
	}

	if matchingFactor == nil {
		return fmt.Errorf("%w: MFA factor not configured", cryptoutilIdentityAppErr.ErrMFAFactorNotFound)
	}

	// Validate factor based on type.
	switch factorType {
	case string(OTPMethodTOTP):
		otpCode, ok := credentials["otp_code"]
		if !ok || otpCode == "" {
			return fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
		}
		// TODO: Validate TOTP using library (e.g., pquerna/otp).
		_ = otpCode

	case "email_otp":
		otpCode, ok := credentials["otp_code"]
		if !ok || otpCode == "" {
			return fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
		}
		// TODO: Get user from context for OTP validation.
		// For now, this is a placeholder.
		_ = otpCode

	case "sms_otp":
		otpCode, ok := credentials["otp_code"]
		if !ok || otpCode == "" {
			return fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
		}
		// TODO: Get user from context for OTP validation.
		// For now, this is a placeholder.
		_ = otpCode

	default:
		return fmt.Errorf("%w: unsupported MFA factor type: %s", cryptoutilIdentityAppErr.ErrServerError, factorType)
	}

	return nil
}

// RequiresMFA checks if authentication profile requires MFA.
func (o *MFAOrchestrator) RequiresMFA(ctx context.Context, authProfileID googleUuid.UUID) (bool, error) {
	factors, err := o.mfaRepo.GetByAuthProfileID(ctx, authProfileID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch MFA factors: %w", err)
	}

	return len(factors) > 0, nil
}
