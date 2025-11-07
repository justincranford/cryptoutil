package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// ClientType represents the type of OAuth 2.1 client.
type ClientType string

const (
	ClientTypeConfidential ClientType = "confidential" // Confidential client (can securely store secrets).
	ClientTypePublic       ClientType = "public"       // Public client (cannot store secrets).
	ClientTypeSPA          ClientType = "spa"          // Single Page Application.
)

// ClientAuthMethod represents the client authentication method.
type ClientAuthMethod string

const (
	ClientAuthMethodSecretBasic       ClientAuthMethod = "client_secret_basic"         // HTTP Basic authentication.
	ClientAuthMethodSecretPost        ClientAuthMethod = "client_secret_post"          // POST body authentication.
	ClientAuthMethodSecretJWT         ClientAuthMethod = "client_secret_jwt"           // JWT signed with client secret.
	ClientAuthMethodPrivateKeyJWT     ClientAuthMethod = "private_key_jwt"             // JWT signed with private key.
	ClientAuthMethodTLSClientAuth     ClientAuthMethod = "tls_client_auth"             // mTLS with CA-issued certificate.
	ClientAuthMethodSelfSignedTLSAuth ClientAuthMethod = "self_signed_tls_client_auth" // mTLS with self-signed certificate.
	ClientAuthMethodBearerToken       ClientAuthMethod = "bearer_token"                // Bearer token authentication.
	ClientAuthMethodNone              ClientAuthMethod = "none"                        // No authentication (public clients).
)

// Client represents an OAuth 2.1 client configuration.
type Client struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Client identification.
	ClientID     string     `gorm:"uniqueIndex;not null" json:"client_id"` // OAuth 2.1 client identifier.
	ClientSecret string     `json:"-"`                                     // Client secret (hashed).
	ClientType   ClientType `gorm:"not null" json:"client_type"`           // Client type (confidential, public, SPA).

	// Client JWK Set (for private_key_jwt authentication).
	JWKs string `gorm:"type:text" json:"jwks,omitempty"` // JSON Web Key Set (RFC 7517).

	// Client metadata.
	Name        string `gorm:"not null" json:"name"`    // Human-readable client name.
	Description string `json:"description,omitempty"`   // Client description.
	LogoURI     string `json:"logo_uri,omitempty"`      // Logo URL.
	HomePageURI string `json:"home_page_uri,omitempty"` // Home page URL.
	PolicyURI   string `json:"policy_uri,omitempty"`    // Privacy policy URL.
	TOSURI      string `json:"tos_uri,omitempty"`       // Terms of service URL.

	// OAuth 2.1 configuration.
	RedirectURIs            []string         `gorm:"type:json" json:"redirect_uris"`             // Allowed redirect URIs.
	AllowedGrantTypes       []string         `gorm:"type:json" json:"allowed_grant_types"`       // Allowed grant types.
	AllowedResponseTypes    []string         `gorm:"type:json" json:"allowed_response_types"`    // Allowed response types.
	AllowedScopes           []string         `gorm:"type:json" json:"allowed_scopes"`            // Allowed scopes.
	TokenEndpointAuthMethod ClientAuthMethod `gorm:"not null" json:"token_endpoint_auth_method"` // Authentication method.

	// PKCE configuration.
	RequirePKCE         bool   `gorm:"default:true" json:"require_pkce"`            // Require PKCE for authorization code flow.
	PKCEChallengeMethod string `gorm:"default:'S256'" json:"pkce_challenge_method"` // PKCE challenge method (S256 or plain).

	// Token configuration.
	AccessTokenLifetime  int `gorm:"default:3600" json:"access_token_lifetime"`   // Access token lifetime (seconds).
	RefreshTokenLifetime int `gorm:"default:86400" json:"refresh_token_lifetime"` // Refresh token lifetime (seconds).
	IDTokenLifetime      int `gorm:"default:3600" json:"id_token_lifetime"`       // ID token lifetime (seconds).

	// Client profile reference (optional).
	ClientProfileID *googleUuid.UUID `gorm:"type:uuid;index" json:"client_profile_id,omitempty"` // Associated client profile.

	// Account status.
	Enabled   bool       `gorm:"default:true" json:"enabled"`       // Client enabled status.
	CreatedAt time.Time  `json:"created_at"`                        // Creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                        // Last update timestamp.
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp.

	// GORM timestamps.
	gorm.Model `json:"-"`
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
