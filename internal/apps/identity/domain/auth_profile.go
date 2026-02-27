// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthProfileType represents the type of authentication profile.
type AuthProfileType string

// Authentication profile type constants.
const (
	// AuthProfileTypeUsernamePassword is username/password authentication.
	AuthProfileTypeUsernamePassword AuthProfileType = "username_password"
	// AuthProfileTypeEmailPassword is email/password authentication.
	AuthProfileTypeEmailPassword AuthProfileType = "email_password"
	// AuthProfileTypeMobilePassword is mobile/password authentication.
	AuthProfileTypeMobilePassword AuthProfileType = "mobile_password"
	// AuthProfileTypePasskey is passkey authentication.
	AuthProfileTypePasskey AuthProfileType = "passkey"
	// AuthProfileTypeMTLS is mTLS authentication.
	AuthProfileTypeMTLS AuthProfileType = "mtls"
)

// AuthProfile represents a user authentication profile with configurable MFA factors.
type AuthProfile struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Profile metadata.
	Name        string          `gorm:"uniqueIndex;not null" json:"name"` // Profile name.
	Description string          `json:"description,omitempty"`            // Profile description.
	ProfileType AuthProfileType `gorm:"not null" json:"profile_type"`     // Profile type.

	// MFA configuration.
	RequireMFA bool     `gorm:"default:false" json:"require_mfa"` // Require MFA for authentication.
	MFAChain   []string `gorm:"serializer:json" json:"mfa_chain"` // Ordered list of MFA factors.

	// mTLS configuration.
	MTLSDomains []string `gorm:"column:mtls_domains;serializer:json" json:"mtls_domains"` // Allowed client certificate domains.

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
