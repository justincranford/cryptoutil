// Copyright (c) 2025 Justin Cranford

package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailOTPRepositoryGORM implements EmailOTPRepository using GORM.
type EmailOTPRepositoryGORM struct {
	db *gorm.DB
}

// NewEmailOTPRepository creates a new GORM-based email OTP repository.
func NewEmailOTPRepository(db *gorm.DB) *EmailOTPRepositoryGORM {
	return &EmailOTPRepositoryGORM{db: db}
}

// Create creates a new email OTP.
func (r *EmailOTPRepositoryGORM) Create(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(otp).Error; err != nil {
		return fmt.Errorf("failed to create email OTP: %w", err)
	}

	return nil
}

// GetByUserID retrieves the most recent email OTP for a user.
func (r *EmailOTPRepositoryGORM) GetByUserID(ctx context.Context, userID googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error) {
	var otp cryptoutilIdentityDomain.EmailOTP

	err := getDB(ctx, r.db).WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrEmailOTPNotFound
		}

		return nil, fmt.Errorf("failed to get email OTP by user ID: %w", err)
	}

	return &otp, nil
}

// GetByID retrieves an email OTP by ID.
func (r *EmailOTPRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.EmailOTP, error) {
	var otp cryptoutilIdentityDomain.EmailOTP

	err := getDB(ctx, r.db).WithContext(ctx).First(&otp, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrEmailOTPNotFound
		}

		return nil, fmt.Errorf("failed to get email OTP by ID: %w", err)
	}

	return &otp, nil
}

// Update updates an existing email OTP.
func (r *EmailOTPRepositoryGORM) Update(ctx context.Context, otp *cryptoutilIdentityDomain.EmailOTP) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Save(otp).Error; err != nil {
		return fmt.Errorf("failed to update email OTP: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all email OTPs for a user.
func (r *EmailOTPRepositoryGORM) DeleteByUserID(ctx context.Context, userID googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("user_id = ?", userID).Delete(&cryptoutilIdentityDomain.EmailOTP{}).Error; err != nil {
		return fmt.Errorf("failed to delete email OTPs by user ID: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired email OTPs.
func (r *EmailOTPRepositoryGORM) DeleteExpired(ctx context.Context) (int64, error) {
	result := getDB(ctx, r.db).WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.EmailOTP{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired email OTPs: %w", result.Error)
	}

	return result.RowsAffected, nil
}
