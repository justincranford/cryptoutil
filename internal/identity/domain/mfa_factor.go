// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// MFAFactorType represents the type of multi-factor authentication factor.
type MFAFactorType string

// Multi-factor authentication factor type constants.
const (
	// MFAFactorTypePassword is a password factor.
	MFAFactorTypePassword MFAFactorType = "password"
	// MFAFactorTypeEmailOTP is an email OTP factor.
	MFAFactorTypeEmailOTP MFAFactorType = "email_otp"
	// MFAFactorTypeSMSOTP is an SMS OTP factor.
	MFAFactorTypeSMSOTP MFAFactorType = "sms_otp"
	// MFAFactorTypeTOTP is a TOTP (Time-based OTP) factor.
	MFAFactorTypeTOTP MFAFactorType = "totp"
	// MFAFactorTypeHOTP is an HOTP (HMAC-based OTP) factor.
	MFAFactorTypeHOTP MFAFactorType = "hotp"
	// MFAFactorTypePasskey is a passkey (WebAuthn) factor.
	MFAFactorTypePasskey MFAFactorType = "passkey"
	// MFAFactorTypeMagicLink is a magic link factor.
	MFAFactorTypeMagicLink MFAFactorType = "magic_link"
	// MFAFactorTypeMTLS is an mTLS certificate factor.
	MFAFactorTypeMTLS MFAFactorType = "mtls"
	// MFAFactorTypeHardwareToken is a hardware security key factor.
	MFAFactorTypeHardwareToken MFAFactorType = "hardware_token"
	// MFAFactorTypeBiometric is a biometric factor.
	MFAFactorTypeBiometric MFAFactorType = "biometric"
)

// MFAFactor represents a multi-factor authentication factor configuration.
type MFAFactor struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Factor metadata.
	Name        string        `gorm:"uniqueIndex;not null" json:"name"` // Factor name.
	Description string        `json:"description,omitempty"`            // Factor description.
	FactorType  MFAFactorType `gorm:"not null" json:"factor_type"`      // Factor type.

	// Factor ordering.
	Order int `gorm:"not null" json:"order"` // Factor order in MFA chain (1-N).

	// Factor configuration.
	Required IntBool `gorm:"type:integer;default:0" json:"required"` // Factor is required (INTEGER for cross-DB compatibility).

	// TOTP/HOTP configuration.
	TOTPAlgorithm string `json:"totp_algorithm,omitempty"` // TOTP algorithm (SHA1, SHA256, SHA512).
	TOTPDigits    int    `json:"totp_digits,omitempty"`    // TOTP digits (6 or 8).
	TOTPPeriod    int    `json:"totp_period,omitempty"`    // TOTP period (seconds).

	// Authentication profile reference (foreign key).
	AuthProfileID googleUuid.UUID `gorm:"type:text;index;not null" json:"auth_profile_id"` // Associated auth profile.

	// Replay prevention (time-bound nonces).
	Nonce          string     `gorm:"uniqueIndex" json:"-"`                    // One-time nonce for replay prevention.
	NonceExpiresAt *time.Time `gorm:"index" json:"nonce_expires_at,omitempty"` // Nonce expiration timestamp.
	NonceUsedAt    *time.Time `gorm:"index" json:"nonce_used_at,omitempty"`    // Timestamp when nonce was consumed.

	// Account status.
	Enabled   bool       `gorm:"default:true" json:"enabled"`       // Factor enabled status.
	CreatedAt time.Time  `json:"created_at"`                        // Creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                        // Last update timestamp.
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp.
}

// BeforeCreate generates UUID and nonce for new MFA factors.
func (mf *MFAFactor) BeforeCreate(_ *gorm.DB) error {
	if mf.ID == googleUuid.Nil {
		mf.ID = googleUuid.Must(googleUuid.NewV7())
	}

	if mf.Nonce == "" {
		mf.Nonce = googleUuid.Must(googleUuid.NewV7()).String()
	}

	return nil
}

// TableName returns the table name for MFAFactor entities.
func (MFAFactor) TableName() string {
	return "mfa_factors"
}

// IsNonceValid checks if nonce is valid and not expired.
func (mf *MFAFactor) IsNonceValid() bool {
	if mf.NonceUsedAt != nil {
		return false
	}

	if mf.NonceExpiresAt != nil && time.Now().After(*mf.NonceExpiresAt) {
		return false
	}

	return true
}

// MarkNonceAsUsed marks nonce as consumed with current timestamp.
func (mf *MFAFactor) MarkNonceAsUsed() {
	now := time.Now()
	mf.NonceUsedAt = &now
}
