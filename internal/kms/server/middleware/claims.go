// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	json "encoding/json"
	"errors"
	"time"
)

// OIDCClaims represents comprehensive OIDC token claims.
// Includes all standard OIDC claims plus common extensions.
type OIDCClaims struct {
	// Standard JWT claims.
	Subject   string    `json:"sub"`
	Issuer    string    `json:"iss"`
	Audience  []string  `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	NotBefore time.Time `json:"nbf"`
	JTI       string    `json:"jti"`

	// OIDC standard claims (profile scope).
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	MiddleName        string `json:"middle_name,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Profile           string `json:"profile,omitempty"`
	Picture           string `json:"picture,omitempty"`
	Website           string `json:"website,omitempty"`
	Gender            string `json:"gender,omitempty"`
	Birthdate         string `json:"birthdate,omitempty"`
	ZoneInfo          string `json:"zoneinfo,omitempty"`
	Locale            string `json:"locale,omitempty"`
	UpdatedAt         int64  `json:"updated_at,omitempty"`

	// OIDC email claims (email scope).
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`

	// OIDC phone claims (phone scope).
	PhoneNumber         string `json:"phone_number,omitempty"`
	PhoneNumberVerified bool   `json:"phone_number_verified,omitempty"`

	// OIDC address claim (address scope).
	Address *AddressClaim `json:"address,omitempty"`

	// OAuth2 claims.
	Scope       string   `json:"scope,omitempty"`
	Scopes      []string `json:"-"` // Parsed from scope claim.
	ClientID    string   `json:"client_id,omitempty"`
	TokenType   string   `json:"token_type,omitempty"`
	ActiveUntil int64    `json:"active_until,omitempty"`

	// Authorization claims.
	Groups      []string `json:"groups,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`

	// Multi-tenancy claims.
	TenantID   string   `json:"tenant_id,omitempty"`
	TenantName string   `json:"tenant_name,omitempty"`
	TenantIDs  []string `json:"tenant_ids,omitempty"` // For multi-tenant access.

	// Service identity claims.
	ServiceName    string `json:"service_name,omitempty"`
	ServiceVersion string `json:"service_version,omitempty"`
	ServiceType    string `json:"service_type,omitempty"`

	// Session claims.
	SessionID string `json:"sid,omitempty"`
	AuthTime  int64  `json:"auth_time,omitempty"`
	ACR       string `json:"acr,omitempty"` // Authentication Context Class Reference.
	AMR       string `json:"amr,omitempty"` // Authentication Methods Reference.
	Nonce     string `json:"nonce,omitempty"`
	AtHash    string `json:"at_hash,omitempty"` // Access token hash.
	CHash     string `json:"c_hash,omitempty"`  // Code hash.

	// Custom claims map for extension.
	Custom map[string]any `json:"-"`

	// Raw claims for debugging.
	RawClaims map[string]any `json:"-"`
}

// AddressClaim represents the OIDC address claim structure.
type AddressClaim struct {
	Formatted     string `json:"formatted,omitempty"`
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}

// OIDCClaimsContextKey is the context key for OIDC claims.
type OIDCClaimsContextKey struct{}

// ClaimsExtractor extracts claims from various token types.
type ClaimsExtractor struct {
	// KnownClaims are claim names that will be mapped to struct fields.
	KnownClaims []string

	// CustomClaimPrefix is the prefix for custom claims (e.g., "urn:cryptoutil:").
	CustomClaimPrefix string
}

// NewClaimsExtractor creates a new claims extractor with default configuration.
func NewClaimsExtractor() *ClaimsExtractor {
	return &ClaimsExtractor{
		KnownClaims: []string{
			// Standard JWT.
			"sub", "iss", "aud", "exp", "iat", "nbf", "jti",
			// OIDC profile.
			"name", "given_name", "family_name", "middle_name", "nickname",
			"preferred_username", "profile", "picture", "website",
			"gender", "birthdate", "zoneinfo", "locale", "updated_at",
			// OIDC email.
			"email", "email_verified",
			// OIDC phone.
			"phone_number", "phone_number_verified",
			// OIDC address.
			"address",
			// OAuth2.
			"scope", "client_id", "token_type", "active_until",
			// Authorization.
			"groups", "roles", "permissions",
			// Multi-tenancy.
			"tenant_id", "tenant_name", "tenant_ids",
			// Service identity.
			"service_name", "service_version", "service_type",
			// Session.
			"sid", "auth_time", "acr", "amr", "nonce", "at_hash", "c_hash",
		},
	}
}

// ExtractFromMap extracts OIDC claims from a generic map.
func (e *ClaimsExtractor) ExtractFromMap(rawClaims map[string]any) (*OIDCClaims, error) {
	if rawClaims == nil {
		return nil, errors.New("claims map is nil")
	}

	// Marshal and unmarshal to handle type conversions.
	data, err := json.Marshal(rawClaims)
	if err != nil {
		return nil, errors.New("failed to marshal claims: " + err.Error())
	}

	claims := &OIDCClaims{
		RawClaims: rawClaims,
		Custom:    make(map[string]any),
	}

	if err := json.Unmarshal(data, claims); err != nil {
		return nil, errors.New("failed to unmarshal claims: " + err.Error())
	}

	// Parse scope string into slice.
	if claims.Scope != "" {
		claims.Scopes = ParseScopeString(claims.Scope)
	}

	// Extract custom claims.
	e.extractCustomClaims(rawClaims, claims)

	return claims, nil
}

// extractCustomClaims extracts claims not in the known list.
func (e *ClaimsExtractor) extractCustomClaims(rawClaims map[string]any, claims *OIDCClaims) {
	knownMap := make(map[string]bool)
	for _, known := range e.KnownClaims {
		knownMap[known] = true
	}

	for key, value := range rawClaims {
		if !knownMap[key] {
			claims.Custom[key] = value
		}
	}
}

// ConvertFromJWTClaims converts JWTClaims to OIDCClaims.
func ConvertFromJWTClaims(jwtClaims *JWTClaims) *OIDCClaims {
	if jwtClaims == nil {
		return nil
	}

	return &OIDCClaims{
		Subject:           jwtClaims.Subject,
		Issuer:            jwtClaims.Issuer,
		Audience:          jwtClaims.Audience,
		ExpiresAt:         jwtClaims.ExpiresAt,
		IssuedAt:          jwtClaims.IssuedAt,
		NotBefore:         jwtClaims.NotBefore,
		JTI:               jwtClaims.JTI,
		Name:              jwtClaims.Name,
		PreferredUsername: jwtClaims.PreferredUsername,
		Email:             jwtClaims.Email,
		EmailVerified:     jwtClaims.EmailVerified,
		Scope:             jwtClaims.Scope,
		Scopes:            jwtClaims.Scopes,
		Custom:            jwtClaims.Custom,
	}
}

// GetOIDCClaims extracts OIDC claims from request context.
func GetOIDCClaims(ctx context.Context) *OIDCClaims {
	// Try direct OIDC claims first.
	if claims, ok := ctx.Value(OIDCClaimsContextKey{}).(*OIDCClaims); ok {
		return claims
	}

	// Try to convert from JWT claims.
	if jwtClaims, ok := ctx.Value(JWTContextKey{}).(*JWTClaims); ok {
		return ConvertFromJWTClaims(jwtClaims)
	}

	return nil
}

// HasScope checks if claims include a specific scope.
func (c *OIDCClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}

	return false
}

// HasAnyScope checks if claims include any of the specified scopes.
func (c *OIDCClaims) HasAnyScope(scopes ...string) bool {
	for _, scope := range scopes {
		if c.HasScope(scope) {
			return true
		}
	}

	return false
}

// HasAllScopes checks if claims include all specified scopes.
func (c *OIDCClaims) HasAllScopes(scopes ...string) bool {
	for _, scope := range scopes {
		if !c.HasScope(scope) {
			return false
		}
	}

	return true
}

// HasGroup checks if claims include a specific group.
func (c *OIDCClaims) HasGroup(group string) bool {
	for _, g := range c.Groups {
		if g == group {
			return true
		}
	}

	return false
}

// HasRole checks if claims include a specific role.
func (c *OIDCClaims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}

	return false
}

// HasPermission checks if claims include a specific permission.
func (c *OIDCClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// GetCustomClaim retrieves a custom claim by key.
func (c *OIDCClaims) GetCustomClaim(key string) (any, bool) {
	if c.Custom == nil {
		return nil, false
	}

	value, exists := c.Custom[key]

	return value, exists
}

// GetCustomString retrieves a custom claim as string.
func (c *OIDCClaims) GetCustomString(key string) string {
	if value, exists := c.GetCustomClaim(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}

	return ""
}

// IsExpired checks if the token is expired.
func (c *OIDCClaims) IsExpired() bool {
	if c.ExpiresAt.IsZero() {
		return false
	}

	return time.Now().UTC().After(c.ExpiresAt)
}

// TimeUntilExpiry returns the duration until token expiry.
func (c *OIDCClaims) TimeUntilExpiry() time.Duration {
	if c.ExpiresAt.IsZero() {
		return time.Duration(0)
	}

	return time.Until(c.ExpiresAt)
}
