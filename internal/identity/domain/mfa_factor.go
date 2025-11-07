package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// MFAFactorType represents the type of multi-factor authentication factor.
type MFAFactorType string

const (
	MFAFactorTypePassword      MFAFactorType = "password"       // Password factor.
	MFAFactorTypeEmailOTP      MFAFactorType = "email_otp"      // Email OTP factor.
	MFAFactorTypeSMSOTP        MFAFactorType = "sms_otp"        // SMS OTP factor.
	MFAFactorTypeTOTP          MFAFactorType = "totp"           // TOTP (Time-based OTP) factor.
	MFAFactorTypeHOTP          MFAFactorType = "hotp"           // HOTP (HMAC-based OTP) factor.
	MFAFactorTypePasskey       MFAFactorType = "passkey"        // Passkey (WebAuthn) factor.
	MFAFactorTypeMagicLink     MFAFactorType = "magic_link"     // Magic link factor.
	MFAFactorTypeMTLS          MFAFactorType = "mtls"           // mTLS certificate factor.
	MFAFactorTypeHardwareToken MFAFactorType = "hardware_token" // Hardware security key factor.
	MFAFactorTypeBiometric     MFAFactorType = "biometric"      // Biometric factor.
)

// MFAFactor represents a multi-factor authentication factor configuration.
type MFAFactor struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Factor metadata.
	Name        string        `gorm:"uniqueIndex;not null" json:"name"` // Factor name.
	Description string        `json:"description,omitempty"`            // Factor description.
	FactorType  MFAFactorType `gorm:"not null" json:"factor_type"`      // Factor type.

	// Factor ordering.
	Order int `gorm:"not null" json:"order"` // Factor order in MFA chain (1-N).

	// Factor configuration.
	Required bool `gorm:"default:false" json:"required"` // Factor is required.

	// TOTP/HOTP configuration.
	TOTPAlgorithm string `json:"totp_algorithm,omitempty"` // TOTP algorithm (SHA1, SHA256, SHA512).
	TOTPDigits    int    `json:"totp_digits,omitempty"`    // TOTP digits (6 or 8).
	TOTPPeriod    int    `json:"totp_period,omitempty"`    // TOTP period (seconds).

	// Authentication profile reference.
	AuthProfileID googleUuid.UUID `gorm:"type:uuid;index;not null" json:"auth_profile_id"` // Associated auth profile.

	// Account status.
	Enabled   bool       `gorm:"default:true" json:"enabled"`       // Factor enabled status.
	CreatedAt time.Time  `json:"created_at"`                        // Creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                        // Last update timestamp.
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp.

	// GORM timestamps.
	gorm.Model `json:"-"`
}

// BeforeCreate generates UUID for new MFA factors.
func (mf *MFAFactor) BeforeCreate(_ *gorm.DB) error {
	if mf.ID == googleUuid.Nil {
		mf.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for MFAFactor entities.
func (MFAFactor) TableName() string {
	return "mfa_factors"
}
