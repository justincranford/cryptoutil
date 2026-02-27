// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthFlowType represents the type of authorization flow.
type AuthFlowType string

// Authorization flow type constants.
const (
	// AuthFlowTypeAuthorizationCode is the authorization code flow.
	AuthFlowTypeAuthorizationCode AuthFlowType = "authorization_code"
	// AuthFlowTypeClientCredentials is the client credentials flow.
	AuthFlowTypeClientCredentials AuthFlowType = "client_credentials"
	// AuthFlowTypeRefreshToken is the refresh token flow.
	AuthFlowTypeRefreshToken AuthFlowType = "refresh_token"
)

// AuthFlow represents an authorization code flow configuration with PKCE.
type AuthFlow struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Flow metadata.
	Name        string       `gorm:"uniqueIndex;not null" json:"name"` // Flow name.
	Description string       `json:"description,omitempty"`            // Flow description.
	FlowType    AuthFlowType `gorm:"not null" json:"flow_type"`        // Flow type.

	// PKCE configuration.
	RequirePKCE         bool   `gorm:"default:true" json:"require_pkce"`            // Require PKCE for this flow.
	PKCEChallengeMethod string `gorm:"default:'S256'" json:"pkce_challenge_method"` // PKCE challenge method (S256 or plain).

	// Scope configuration.
	// Flow configuration.
	AllowedScopes []string `gorm:"serializer:json" json:"allowed_scopes"` // Allowed scopes for this flow.

	// Consent configuration.
	RequireConsent     bool `gorm:"default:true" json:"require_consent"`   // Require user consent.
	ConsentScreenCount int  `gorm:"default:1" json:"consent_screen_count"` // Number of consent screens (1 or 2).
	RememberConsent    bool `gorm:"default:false" json:"remember_consent"` // Remember user consent decisions.

	// State parameter configuration.
	RequireState bool `gorm:"default:true" json:"require_state"` // Require state parameter for CSRF protection.

	// Client profile reference (optional).
	ClientProfileID NullableUUID `gorm:"type:text;index" json:"client_profile_id,omitempty"` // Associated client profile.

	// Account status.
	Enabled   bool       `gorm:"default:true" json:"enabled"`       // Flow enabled status.
	CreatedAt time.Time  `json:"created_at"`                        // Creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                        // Last update timestamp.
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp.
}

// BeforeCreate generates UUID for new auth flows.
func (af *AuthFlow) BeforeCreate(_ *gorm.DB) error {
	if af.ID == googleUuid.Nil {
		af.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for AuthFlow entities.
func (AuthFlow) TableName() string {
	return "auth_flows"
}
