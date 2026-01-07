// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// ClientProfile represents a parameterized authorization flow profile for OAuth clients.
type ClientProfile struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Profile metadata.
	Name        string `gorm:"uniqueIndex;not null" json:"name"` // Profile name.
	Description string `json:"description,omitempty"`            // Profile description.

	// Scope configuration.
	RequiredScopes []string `gorm:"serializer:json" json:"required_scopes"` // Required scopes for this profile.
	OptionalScopes []string `gorm:"serializer:json" json:"optional_scopes"` // Optional scopes for this profile.

	// Consent configuration.
	ConsentScreenCount int    `gorm:"default:1" json:"consent_screen_count"`                               // Number of consent screens (1 or 2).
	ConsentScreen1Text string `gorm:"column:consent_screen_1_text" json:"consent_screen_1_text,omitempty"` // First consent screen text.
	ConsentScreen2Text string `gorm:"column:consent_screen_2_text" json:"consent_screen_2_text,omitempty"` // Second consent screen text (if count=2).

	// MFA configuration for client authentication.
	RequireClientMFA bool     `gorm:"default:false" json:"require_client_mfa"` // Require MFA for client authentication.
	ClientMFAChain   []string `gorm:"serializer:json" json:"client_mfa_chain"` // Ordered list of client auth methods.

	// Account status.
	Enabled   bool       `gorm:"default:true" json:"enabled"`       // Profile enabled status.
	CreatedAt time.Time  `json:"created_at"`                        // Creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                        // Last update timestamp.
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp.
}

// BeforeCreate generates UUID for new client profiles.
func (cp *ClientProfile) BeforeCreate(_ *gorm.DB) error {
	if cp.ID == googleUuid.Nil {
		cp.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for ClientProfile entities.
func (ClientProfile) TableName() string {
	return "client_profiles"
}
