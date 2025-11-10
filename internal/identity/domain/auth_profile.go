package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthProfileType represents the type of authentication profile.
type AuthProfileType string

const (
	AuthProfileTypeUsernamePassword AuthProfileType = "username_password" // Username/password authentication.
	AuthProfileTypeEmailPassword    AuthProfileType = "email_password"    // Email/password authentication.
	AuthProfileTypeMobilePassword   AuthProfileType = "mobile_password"   // Mobile/password authentication.
	AuthProfileTypePasskey          AuthProfileType = "passkey"           // Passkey authentication.
	AuthProfileTypeMTLS             AuthProfileType = "mtls"              // mTLS authentication.
)

// AuthProfile represents a user authentication profile with configurable MFA factors.
type AuthProfile struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Profile metadata.
	Name        string          `gorm:"uniqueIndex;not null" json:"name"` // Profile name.
	Description string          `json:"description,omitempty"`            // Profile description.
	ProfileType AuthProfileType `gorm:"not null" json:"profile_type"`     // Profile type.

	// MFA configuration.
	RequireMFA bool     `gorm:"default:false" json:"require_mfa"` // Require multi-factor authentication.
	MFAChain   []string `gorm:"type:json" json:"mfa_chain"`       // Ordered list of MFA factors.

	// mTLS configuration.
	MTLSDomains []string `gorm:"type:json" json:"mtls_domains"` // Allowed client certificate domains.

	// Account status.
	Enabled   bool       `gorm:"default:true" json:"enabled"`       // Profile enabled status.
	CreatedAt time.Time  `json:"created_at"`                        // Creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                        // Last update timestamp.
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp.
}

// BeforeCreate generates UUID for new auth profiles.
func (ap *AuthProfile) BeforeCreate(_ *gorm.DB) error {
	if ap.ID == googleUuid.Nil {
		ap.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for AuthProfile entities.
func (AuthProfile) TableName() string {
	return "auth_profiles"
}
