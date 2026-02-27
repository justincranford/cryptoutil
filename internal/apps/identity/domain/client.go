// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// ClientType represents the type of OAuth 2.1 client.
type ClientType string

// OAuth 2.1 client type constants.
const (
	// ClientTypeConfidential is a confidential client (can securely store secrets).
	ClientTypeConfidential ClientType = "confidential"
	// ClientTypePublic is a public client (cannot store secrets).
	ClientTypePublic ClientType = "public"
	// ClientTypeSPA is a Single Page Application.
	ClientTypeSPA ClientType = "spa"
)

// ClientAuthMethod represents the client authentication method.
type ClientAuthMethod string

// Client authentication method constants.
const (
	// ClientAuthMethodSecretBasic is HTTP Basic authentication.
	ClientAuthMethodSecretBasic ClientAuthMethod = "client_secret_basic"
	// ClientAuthMethodSecretPost is POST body authentication.
	ClientAuthMethodSecretPost ClientAuthMethod = "client_secret_post"
	// ClientAuthMethodSecretJWT is JWT signed with client secret.
	ClientAuthMethodSecretJWT ClientAuthMethod = "client_secret_jwt"
	// ClientAuthMethodPrivateKeyJWT is JWT signed with private key.
	ClientAuthMethodPrivateKeyJWT ClientAuthMethod = "private_key_jwt"
	// ClientAuthMethodTLSClientAuth is mTLS with CA-issued certificate.
	ClientAuthMethodTLSClientAuth ClientAuthMethod = "tls_client_auth"
	// ClientAuthMethodSelfSignedTLSAuth is mTLS with self-signed certificate.
	ClientAuthMethodSelfSignedTLSAuth ClientAuthMethod = "self_signed_tls_client_auth"
	// ClientAuthMethodBearerToken is Bearer token authentication.
	ClientAuthMethodBearerToken ClientAuthMethod = "bearer_token"
	// ClientAuthMethodNone is no authentication (public clients).
	ClientAuthMethodNone ClientAuthMethod = "none"
)

// Client represents an OAuth 2.1 client configuration.
type Client struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Client identification.
	ClientID     string     `gorm:"uniqueIndex;not null" json:"client_id"` // OAuth 2.1 client identifier.
	ClientSecret string     `json:"-"`                                     // Client secret (hashed).
	ClientType   ClientType `gorm:"not null" json:"client_type"`           // Client type (confidential, public, SPA).

	// Client JWK Set (for private_key_jwt authentication).
	JWKs string `gorm:"column:j_w_ks;type:text" json:"jwks,omitempty"` // JSON Web Key Set (RFC 7517).

	// Client metadata.
	Name        string `gorm:"not null" json:"name"`    // Human-readable client name.
	Description string `json:"description,omitempty"`   // Client description.
	LogoURI     string `json:"logo_uri,omitempty"`      // Logo URL.
	HomePageURI string `json:"home_page_uri,omitempty"` // Home page URL.
	PolicyURI   string `json:"policy_uri,omitempty"`    // Privacy policy URL.
	TOSURI      string `json:"tos_uri,omitempty"`       // Terms of service URL.

	// OAuth 2.1 configuration.
	RedirectURIs            []string         `gorm:"serializer:json" json:"redirect_uris"`             // Allowed redirect URIs.
	PostLogoutRedirectURIs  []string         `gorm:"serializer:json" json:"post_logout_redirect_uris"` // Allowed post-logout redirect URIs (OIDC RP-Initiated Logout).
	AllowedGrantTypes       []string         `gorm:"serializer:json" json:"allowed_grant_types"`       // Allowed grant types.
	AllowedResponseTypes    []string         `gorm:"serializer:json" json:"allowed_response_types"`    // Allowed response types.
	AllowedScopes           []string         `gorm:"serializer:json" json:"allowed_scopes"`            // Allowed scopes.
	TokenEndpointAuthMethod ClientAuthMethod `gorm:"not null" json:"token_endpoint_auth_method"`       // Authentication method.

	// OIDC Logout configuration (OpenID Connect Front-Channel Logout 1.0 / Back-Channel Logout 1.0).
	FrontChannelLogoutURI             string `gorm:"column:frontchannel_logout_uri" json:"frontchannel_logout_uri,omitempty"`                                            // URL for front-channel logout iframe.
	FrontChannelLogoutSessionRequired *bool  `gorm:"column:frontchannel_logout_session_required;type:boolean;default:false" json:"frontchannel_logout_session_required"` // Include sid in logout request.
	BackChannelLogoutURI              string `gorm:"column:backchannel_logout_uri" json:"backchannel_logout_uri,omitempty"`                                              // URL for back-channel logout token delivery.
	BackChannelLogoutSessionRequired  *bool  `gorm:"column:backchannel_logout_session_required;type:boolean;default:false" json:"backchannel_logout_session_required"`   // Include sid in logout token.

	// PKCE configuration.
	RequirePKCE         *bool  `gorm:"type:boolean;default:true" json:"require_pkce"` // Require PKCE for authorization code flow.
	PKCEChallengeMethod string `gorm:"default:'S256'" json:"pkce_challenge_method"`   // PKCE challenge method (S256 or plain).

	// Token configuration.
	AccessTokenLifetime  int `gorm:"default:3600" json:"access_token_lifetime"`   // Access token lifetime (seconds).
	RefreshTokenLifetime int `gorm:"default:86400" json:"refresh_token_lifetime"` // Refresh token lifetime (seconds).
	IDTokenLifetime      int `gorm:"default:3600" json:"id_token_lifetime"`       // ID token lifetime (seconds).

	// Client profile reference (optional).
	ClientProfileID NullableUUID `gorm:"type:text;index" json:"client_profile_id,omitempty"` // Associated client profile.

	// Certificate-based authentication fields.
	CertificateSubject     string `gorm:"index" json:"certificate_subject,omitempty"`     // Expected certificate subject (CN).
	CertificateFingerprint string `gorm:"index" json:"certificate_fingerprint,omitempty"` // Expected certificate SHA-256 fingerprint (hex).

	// Account status.
	Enabled   *bool      `gorm:"type:boolean;default:true" json:"enabled"` // Client enabled status.
	CreatedAt time.Time  `json:"created_at"`                               // Creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                               // Last update timestamp.
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`        // Soft delete timestamp.
}

// BeforeCreate generates UUID for new clients.
func (c *Client) BeforeCreate(_ *gorm.DB) error {
	if c.ID == googleUuid.Nil {
		c.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for Client entities.
func (Client) TableName() string {
	return "clients"
}
