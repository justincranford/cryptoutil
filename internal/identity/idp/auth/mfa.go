// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"
	"fmt"
	"time"

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
	telemetry       *MFATelemetry
}

// NewMFAOrchestrator creates a new MFA orchestrator.
func NewMFAOrchestrator(
	mfaRepo cryptoutilIdentityRepository.MFAFactorRepository,
	otpService *OTPService,
	profileRegistry *ProfileRegistry,
	telemetry *MFATelemetry,
) *MFAOrchestrator {
	return &MFAOrchestrator{
		mfaRepo:         mfaRepo,
		otpService:      otpService,
		profileRegistry: profileRegistry,
		telemetry:       telemetry,
	}
}

// GetRequiredFactors returns the required MFA factors for an authentication profile.
func (o *MFAOrchestrator) GetRequiredFactors(ctx context.Context, authProfileID googleUuid.UUID) ([]string, error) {
	ctx, span := o.telemetry.StartGetRequiredFactorsSpan(ctx, authProfileID)
	defer span.End()

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

	o.telemetry.RecordRequiredFactors(ctx, authProfileID, len(factorTypes))

	return factorTypes, nil
}

// ValidateFactor validates a specific MFA factor with replay prevention.
func (o *MFAOrchestrator) ValidateFactor(ctx context.Context, authProfileID googleUuid.UUID, factorType string, credentials map[string]string) error {
	ctx, span := o.telemetry.StartValidationSpan(ctx, factorType, authProfileID)
	defer span.End()

	startTime := time.Now()
	isReplay := false

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
		o.telemetry.RecordValidation(ctx, factorType, false, time.Since(startTime), false)
		return fmt.Errorf("%w: MFA factor not configured", cryptoutilIdentityAppErr.ErrMFAFactorNotFound)
	}

	// Validate nonce for replay prevention.
	if !matchingFactor.IsNonceValid() {
		isReplay = true
		o.telemetry.RecordValidation(ctx, factorType, false, time.Since(startTime), isReplay)
		return fmt.Errorf("%w: nonce already used or expired", cryptoutilIdentityAppErr.ErrInvalidCredentials)
	}

	// Validate factor based on type.
	switch factorType {
	case string(OTPMethodTOTP):
		otpCode, ok := credentials["otp_code"]
		if !ok || otpCode == "" {
			o.telemetry.RecordValidation(ctx, factorType, false, time.Since(startTime), false)
			return fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
		}
		// TODO: Validate TOTP using library (e.g., pquerna/otp).
		_ = otpCode

	case "email_otp":
		otpCode, ok := credentials["otp_code"]
		if !ok || otpCode == "" {
			o.telemetry.RecordValidation(ctx, factorType, false, time.Since(startTime), false)
			return fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
		}
		// TODO: Get user from context for OTP validation.
		// For now, this is a placeholder.
		_ = otpCode

	case "sms_otp":
		otpCode, ok := credentials["otp_code"]
		if !ok || otpCode == "" {
			o.telemetry.RecordValidation(ctx, factorType, false, time.Since(startTime), false)
			return fmt.Errorf("%w: missing otp_code", cryptoutilIdentityAppErr.ErrInvalidCredentials)
		}
		// TODO: Get user from context for OTP validation.
		// For now, this is a placeholder.
		_ = otpCode

	default:
		o.telemetry.RecordValidation(ctx, factorType, false, time.Since(startTime), false)
		return fmt.Errorf("%w: unsupported MFA factor type: %s", cryptoutilIdentityAppErr.ErrServerError, factorType)
	}

	// Mark nonce as used (replay prevention).
	matchingFactor.MarkNonceAsUsed()
	if err := o.mfaRepo.Update(ctx, matchingFactor); err != nil {
		o.telemetry.RecordValidation(ctx, factorType, false, time.Since(startTime), false)
		return fmt.Errorf("failed to mark nonce as used: %w", err)
	}

	o.telemetry.RecordValidation(ctx, factorType, true, time.Since(startTime), false)

	return nil
}

// RequiresMFA checks if authentication profile requires MFA.
func (o *MFAOrchestrator) RequiresMFA(ctx context.Context, authProfileID googleUuid.UUID) (bool, error) {
	ctx, span := o.telemetry.StartRequiresMFASpan(ctx, authProfileID)
	defer span.End()

	factors, err := o.mfaRepo.GetByAuthProfileID(ctx, authProfileID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch MFA factors: %w", err)
	}

	requiresMFA := len(factors) > 0
	o.telemetry.RecordRequiresMFA(ctx, authProfileID, requiresMFA)

	return requiresMFA, nil
}
