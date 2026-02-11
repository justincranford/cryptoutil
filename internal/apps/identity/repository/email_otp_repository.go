// Copyright (c) 2025 Justin Cranford

package repository

import (
	"context"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"

	googleUuid "github.com/google/uuid"
)

// EmailOTPRepository defines persistence operations for email OTPs.
type EmailOTPRepository interface {
	// Create creates a new email OTP.
	Create(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error

	// GetByUserID retrieves the most recent email OTP for a user.
	GetByUserID(ctx context.Context, userID googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error)

	// GetByID retrieves an email OTP by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error)

	// Update updates an existing email OTP.
	Update(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error

	// DeleteByUserID deletes all email OTPs for a user.
	DeleteByUserID(ctx context.Context, userID googleUuid.UUID) error

	// DeleteExpired deletes all expired email OTPs.
	DeleteExpired(ctx context.Context) (int64, error)
}
