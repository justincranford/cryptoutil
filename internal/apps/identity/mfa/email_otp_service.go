// Copyright (c) 2025 Justin Cranford

// Package mfa provides multi-factor authentication services.
package mfa

import (
	"context"
	"fmt"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityEmail "cryptoutil/internal/apps/identity/email"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRatelimit "cryptoutil/internal/apps/identity/ratelimit"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"

	googleUuid "github.com/google/uuid"
)

// EmailOTPService handles email-based OTP generation and verification.
type EmailOTPService struct {
	emailOTPRepo EmailOTPRepository
	emailService cryptoutilIdentityEmail.EmailService
	rateLimiter  *cryptoutilIdentityRatelimit.RateLimiter
}

// EmailOTPRepository defines minimal repository interface needed by EmailOTPService.
type EmailOTPRepository interface {
	Create(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error
	GetByUserID(ctx context.Context, userID googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error)
	Update(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error
}

// NewEmailOTPService creates a new email OTP service.
func NewEmailOTPService(
	emailOTPRepo EmailOTPRepository,
	emailService cryptoutilIdentityEmail.EmailService,
) *EmailOTPService {
	return &EmailOTPService{
		emailOTPRepo: emailOTPRepo,
		emailService: emailService,
		rateLimiter: cryptoutilIdentityRatelimit.NewRateLimiter(
			cryptoutilIdentityMagic.DefaultEmailOTPRateLimit,
			cryptoutilIdentityMagic.DefaultEmailOTPRateLimitWindow,
		),
	}
}

// SendOTP generates and sends an OTP to the user's email.
func (s *EmailOTPService) SendOTP(ctx context.Context, userID googleUuid.UUID, email string) error {
	// Check rate limit.
	if err := s.rateLimiter.Allow(userID.String()); err != nil {
		return fmt.Errorf("%w: %w", cryptoutilIdentityAppErr.ErrRateLimitExceeded, err)
	}

	// Generate OTP.
	plainOTP, err := GenerateEmailOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash OTP with PBKDF2 (FIPS-compliant).
	hashedOTP, err := cryptoutilSharedCryptoPassword.HashPassword(plainOTP)
	if err != nil {
		return fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Create OTP record.
	otp := &cryptoutilIdentityDomain.EmailOTP{
		ID:        googleUuid.New(),
		UserID:    userID,
		CodeHash:  hashedOTP,
		Used:      false,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultEmailOTPLifetime),
	}

	if err := s.emailOTPRepo.Create(ctx, otp); err != nil {
		return fmt.Errorf("failed to create OTP record: %w", err)
	}

	// Send OTP via email.
	subject := "Your One-Time Password"
	body := fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 10 minutes.", plainOTP)

	if err := s.emailService.SendEmail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// VerifyOTP verifies an OTP for the user.
func (s *EmailOTPService) VerifyOTP(ctx context.Context, userID googleUuid.UUID, code string) error {
	// Fetch most recent OTP for user.
	otp, err := s.emailOTPRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %w", cryptoutilIdentityAppErr.ErrInvalidOTP, err)
	}

	// Check if OTP is expired.
	if otp.IsExpired() {
		return cryptoutilIdentityAppErr.ErrExpiredOTP
	}

	// Check if OTP is already used.
	if otp.IsUsed() {
		return cryptoutilIdentityAppErr.ErrOTPAlreadyUsed
	}

	// Verify OTP code (PBKDF2-HMAC-SHA256, FIPS-compliant).
	match, _, err := cryptoutilSharedCryptoPassword.VerifyPassword(code, otp.CodeHash)
	if err != nil {
		return fmt.Errorf("%w: verification failed", cryptoutilIdentityAppErr.ErrInvalidOTP)
	}

	if !match {
		return cryptoutilIdentityAppErr.ErrInvalidOTP
	}

	// Mark OTP as used.
	otp.MarkAsUsed()

	if err := s.emailOTPRepo.Update(ctx, otp); err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	return nil
}
