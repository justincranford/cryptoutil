// Copyright 2025 Cisco Systems, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	json "encoding/json"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RealmType represents the type of authentication realm.
type RealmType string

const (
	// Federated realm types (external identity providers).

	// RealmTypeUsernamePassword is a database-based username/password realm.
	RealmTypeUsernamePassword RealmType = "username_password"

	// RealmTypeLDAP is an LDAP-based authentication realm.
	RealmTypeLDAP RealmType = "ldap"

	// RealmTypeOAuth2 is an OAuth2/OIDC-based authentication realm.
	RealmTypeOAuth2 RealmType = "oauth2"

	// RealmTypeSAML is a SAML-based authentication realm.
	RealmTypeSAML RealmType = "saml"

	// Non-federated browser realm types (session-based, /browser/** paths).

	// RealmTypeJWESessionCookie uses JSON Web Encryption for stateless encrypted session cookies.
	RealmTypeJWESessionCookie RealmType = "jwe-session-cookie"

	// RealmTypeJWSSessionCookie uses JSON Web Signature for stateless signed session cookies.
	RealmTypeJWSSessionCookie RealmType = "jws-session-cookie"

	// RealmTypeOpaqueSessionCookie uses server-side session storage with opaque token cookies.
	RealmTypeOpaqueSessionCookie RealmType = "opaque-session-cookie"

	// RealmTypeBasicUsernamePassword uses HTTP Basic authentication with username/password.
	RealmTypeBasicUsernamePassword RealmType = "basic-username-password"

	// RealmTypeBearerAPIToken uses Bearer token authentication from browser clients.
	RealmTypeBearerAPIToken RealmType = "bearer-api-token"

	// RealmTypeHTTPSClientCert uses mTLS client certificate authentication from browsers.
	RealmTypeHTTPSClientCert RealmType = "https-client-cert"

	// RealmTypeJWESessionToken uses JSON Web Encryption for stateless encrypted service tokens.
	RealmTypeJWESessionToken RealmType = "jwe-session-token"

	// RealmTypeJWSSessionToken uses JSON Web Signature for stateless signed service tokens.
	RealmTypeJWSSessionToken RealmType = "jws-session-token"

	// RealmTypeOpaqueSessionToken uses server-side token storage with opaque tokens.
	RealmTypeOpaqueSessionToken RealmType = "opaque-session-token"

	// RealmTypeBasicClientIDSecret uses HTTP Basic authentication with client_id/client_secret.
	RealmTypeBasicClientIDSecret RealmType = "basic-client-id-secret"

	// Note: bearer-api-token and https-client-cert are shared between browser and service realms.
	// The realm configuration (BrowserRealms vs ServiceRealms) determines the request path enforcement.
)

// RealmConfig holds configuration for a specific realm type.
type RealmConfig interface {
	// GetType returns the realm type.
	GetType() RealmType

	// Validate validates the configuration.
	Validate() error
}

// UsernamePasswordConfig configures a username/password realm.
type UsernamePasswordConfig struct {
	MinPasswordLength int  `json:"min_password_length"`
	RequireUppercase  bool `json:"require_uppercase"`
	RequireLowercase  bool `json:"require_lowercase"`
	RequireDigit      bool `json:"require_digit"`
	RequireSpecial    bool `json:"require_special"`
}

// GetType returns RealmTypeUsernamePassword.
func (c *UsernamePasswordConfig) GetType() RealmType {
	return RealmTypeUsernamePassword
}

// Validate validates the configuration.
func (c *UsernamePasswordConfig) Validate() error {
	if c.MinPasswordLength < 1 {
		return fmt.Errorf("min_password_length must be at least 1")
	}

	return nil
}

// LDAPConfig configures an LDAP realm.
type LDAPConfig struct {
	URL           string `json:"url"`
	BindDN        string `json:"bind_dn"`
	BindPassword  string `json:"bind_password"`
	BaseDN        string `json:"base_dn"`
	UserFilter    string `json:"user_filter"`
	GroupFilter   string `json:"group_filter"`
	UseTLS        bool   `json:"use_tls"`
	SkipTLSVerify bool   `json:"skip_tls_verify"`
}

// GetType returns RealmTypeLDAP.
func (c *LDAPConfig) GetType() RealmType {
	return RealmTypeLDAP
}

// Validate validates the configuration.
func (c *LDAPConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url is required")
	}

	if c.BaseDN == "" {
		return fmt.Errorf("base_dn is required")
	}

	return nil
}

// OAuth2Config configures an OAuth2/OIDC realm.
type OAuth2Config struct {
	ProviderURL  string   `json:"provider_url"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
	RedirectURI  string   `json:"redirect_uri"`
	UseDiscovery bool     `json:"use_discovery"`
	AuthorizeURL string   `json:"authorize_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"userinfo_url"`
}

// GetType returns RealmTypeOAuth2.
func (c *OAuth2Config) GetType() RealmType {
	return RealmTypeOAuth2
}

// Validate validates the configuration.
func (c *OAuth2Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	if c.UseDiscovery && c.ProviderURL == "" {
		return fmt.Errorf("provider_url is required when use_discovery is true")
	}

	if !c.UseDiscovery && (c.AuthorizeURL == "" || c.TokenURL == "") {
		return fmt.Errorf("authorize_url and token_url are required when use_discovery is false")
	}

	return nil
}

// SAMLConfig configures a SAML realm.
type SAMLConfig struct {
	MetadataURL     string `json:"metadata_url"`
	MetadataXML     string `json:"metadata_xml"`
	EntityID        string `json:"entity_id"`
	AssertionURL    string `json:"assertion_consumer_url"`
	SignRequests    bool   `json:"sign_requests"`
	SigningCertPath string `json:"signing_cert_path"`
	SigningKeyPath  string `json:"signing_key_path"`
}

// GetType returns RealmTypeSAML.
func (c *SAMLConfig) GetType() RealmType {
	return RealmTypeSAML
}

// Validate validates the configuration.
func (c *SAMLConfig) Validate() error {
	if c.MetadataURL == "" && c.MetadataXML == "" {
		return fmt.Errorf("either metadata_url or metadata_xml is required")
	}

	if c.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	return nil
}

// JWESessionCookieConfig configures a JWE session cookie realm (browser, /browser/** paths).
type JWESessionCookieConfig struct {
	EncryptionAlgorithm  string `json:"encryption_algorithm"`   // e.g., "dir+A256GCM"
	SessionExpiryMinutes int    `json:"session_expiry_minutes"` // e.g., 15
}

// GetType returns RealmTypeJWESessionCookie.
func (c *JWESessionCookieConfig) GetType() RealmType {
	return RealmTypeJWESessionCookie
}

// Validate validates the configuration.
func (c *JWESessionCookieConfig) Validate() error {
	if c.SessionExpiryMinutes < 1 {
		return fmt.Errorf("session_expiry_minutes must be at least 1")
	}

	return nil
}

// JWSSessionCookieConfig configures a JWS session cookie realm (browser, /browser/** paths).
type JWSSessionCookieConfig struct {
	SigningAlgorithm     string `json:"signing_algorithm"`      // e.g., "RS256", "ES256", "EdDSA"
	SessionExpiryMinutes int    `json:"session_expiry_minutes"` // e.g., 15
}

// GetType returns RealmTypeJWSSessionCookie.
func (c *JWSSessionCookieConfig) GetType() RealmType {
	return RealmTypeJWSSessionCookie
}

// Validate validates the configuration.
func (c *JWSSessionCookieConfig) Validate() error {
	if c.SessionExpiryMinutes < 1 {
		return fmt.Errorf("session_expiry_minutes must be at least 1")
	}

	return nil
}

// OpaqueSessionCookieConfig configures an opaque session cookie realm (browser, /browser/** paths).
type OpaqueSessionCookieConfig struct {
	TokenLengthBytes     int    `json:"token_length_bytes"`     // e.g., 32 bytes
	SessionExpiryMinutes int    `json:"session_expiry_minutes"` // e.g., 15
	StorageType          string `json:"storage_type"`           // "database" or "redis"
}

// GetType returns RealmTypeOpaqueSessionCookie.
func (c *OpaqueSessionCookieConfig) GetType() RealmType {
	return RealmTypeOpaqueSessionCookie
}

// Validate validates the configuration.
func (c *OpaqueSessionCookieConfig) Validate() error {
	if c.TokenLengthBytes < cryptoutilSharedMagic.RealmMinTokenLengthBytes {
		return fmt.Errorf("token_length_bytes must be at least %d", cryptoutilSharedMagic.RealmMinTokenLengthBytes)
	}

	if c.SessionExpiryMinutes < 1 {
		return fmt.Errorf("session_expiry_minutes must be at least 1")
	}

	if c.StorageType != cryptoutilSharedMagic.RealmStorageTypeDatabase && c.StorageType != cryptoutilSharedMagic.RealmStorageTypeRedis {
		return fmt.Errorf("storage_type must be '%s' or '%s'", cryptoutilSharedMagic.RealmStorageTypeDatabase, cryptoutilSharedMagic.RealmStorageTypeRedis)
	}

	return nil
}

// BasicUsernamePasswordConfig configures a Basic HTTP authentication realm (browser, /browser/** paths).
// Note: This is different from RealmTypeUsernamePassword which is federated.
type BasicUsernamePasswordConfig struct {
	MinPasswordLength int  `json:"min_password_length"`
	RequireUppercase  bool `json:"require_uppercase"`
	RequireLowercase  bool `json:"require_lowercase"`
	RequireDigit      bool `json:"require_digit"`
	RequireSpecial    bool `json:"require_special"`
}

// GetType returns RealmTypeBasicUsernamePassword.
func (c *BasicUsernamePasswordConfig) GetType() RealmType {
	return RealmTypeBasicUsernamePassword
}

// Validate validates the configuration.
func (c *BasicUsernamePasswordConfig) Validate() error {
	if c.MinPasswordLength < 1 {
		return fmt.Errorf("min_password_length must be at least 1")
	}

	return nil
}

// BearerAPITokenConfig configures a Bearer token authentication realm.
// Used for both browser (/browser/**) and service (/service/**) paths.
type BearerAPITokenConfig struct {
	TokenExpiryDays   int  `json:"token_expiry_days"`   // e.g., 30 for long-lived service tokens
	TokenLengthBytes  int  `json:"token_length_bytes"`  // e.g., 64 bytes
	AllowRefreshToken bool `json:"allow_refresh_token"` // Allow token refresh
}

// GetType returns RealmTypeBearerAPIToken.
func (c *BearerAPITokenConfig) GetType() RealmType {
	return RealmTypeBearerAPIToken
}

// Validate validates the configuration.
func (c *BearerAPITokenConfig) Validate() error {
	if c.TokenExpiryDays < 1 {
		return fmt.Errorf("token_expiry_days must be at least 1")
	}

	if c.TokenLengthBytes < cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes {
		return fmt.Errorf("token_length_bytes must be at least %d", cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	}

	return nil
}

// HTTPSClientCertConfig configures mTLS client certificate authentication.
// Used for both browser (/browser/**) and service (/service/**) paths.
type HTTPSClientCertConfig struct {
	RequireClientCert bool     `json:"require_client_cert"` // Require client certificate
	TrustedCAs        []string `json:"trusted_cas"`         // PEM-encoded CA certificates
	ValidateOCSP      bool     `json:"validate_ocsp"`       // Check OCSP revocation
	ValidateCRL       bool     `json:"validate_crl"`        // Check CRL revocation
}

// GetType returns RealmTypeHTTPSClientCert.
func (c *HTTPSClientCertConfig) GetType() RealmType {
	return RealmTypeHTTPSClientCert
}

// Validate validates the configuration.
func (c *HTTPSClientCertConfig) Validate() error {
	if c.RequireClientCert && len(c.TrustedCAs) == 0 {
		return fmt.Errorf("trusted_cas is required when require_client_cert is true")
	}

	return nil
}

// JWESessionTokenConfig configures a JWE session token realm (service, /service/** paths).
type JWESessionTokenConfig struct {
	EncryptionAlgorithm string `json:"encryption_algorithm"` // e.g., "dir+A256GCM"
	TokenExpiryMinutes  int    `json:"token_expiry_minutes"` // e.g., 60
}

// GetType returns RealmTypeJWESessionToken.
func (c *JWESessionTokenConfig) GetType() RealmType {
	return RealmTypeJWESessionToken
}

// Validate validates the configuration.
func (c *JWESessionTokenConfig) Validate() error {
	if c.TokenExpiryMinutes < 1 {
		return fmt.Errorf("token_expiry_minutes must be at least 1")
	}

	return nil
}

// JWSSessionTokenConfig configures a JWS session token realm (service, /service/** paths).
type JWSSessionTokenConfig struct {
	SigningAlgorithm   string `json:"signing_algorithm"`    // e.g., "RS256", "ES256", "EdDSA"
	TokenExpiryMinutes int    `json:"token_expiry_minutes"` // e.g., 60
}

// GetType returns RealmTypeJWSSessionToken.
func (c *JWSSessionTokenConfig) GetType() RealmType {
	return RealmTypeJWSSessionToken
}

// Validate validates the configuration.
func (c *JWSSessionTokenConfig) Validate() error {
	if c.TokenExpiryMinutes < 1 {
		return fmt.Errorf("token_expiry_minutes must be at least 1")
	}

	return nil
}

// OpaqueSessionTokenConfig configures an opaque session token realm (service, /service/** paths).
type OpaqueSessionTokenConfig struct {
	TokenLengthBytes   int    `json:"token_length_bytes"`   // e.g., 32 bytes
	TokenExpiryMinutes int    `json:"token_expiry_minutes"` // e.g., 60
	StorageType        string `json:"storage_type"`         // "database" or "redis"
}

// GetType returns RealmTypeOpaqueSessionToken.
func (c *OpaqueSessionTokenConfig) GetType() RealmType {
	return RealmTypeOpaqueSessionToken
}

// Validate validates the configuration.
func (c *OpaqueSessionTokenConfig) Validate() error {
	if c.TokenLengthBytes < cryptoutilSharedMagic.RealmMinTokenLengthBytes {
		return fmt.Errorf("token_length_bytes must be at least %d", cryptoutilSharedMagic.RealmMinTokenLengthBytes)
	}

	if c.TokenExpiryMinutes < 1 {
		return fmt.Errorf("token_expiry_minutes must be at least 1")
	}

	if c.StorageType != cryptoutilSharedMagic.RealmStorageTypeDatabase && c.StorageType != cryptoutilSharedMagic.RealmStorageTypeRedis {
		return fmt.Errorf("storage_type must be '%s' or '%s'", cryptoutilSharedMagic.RealmStorageTypeDatabase, cryptoutilSharedMagic.RealmStorageTypeRedis)
	}

	return nil
}

// BasicClientIDSecretConfig configures HTTP Basic authentication with client_id/client_secret (service, /service/** paths).
type BasicClientIDSecretConfig struct {
	MinSecretLength  int  `json:"min_secret_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireDigit     bool `json:"require_digit"`
	RequireSpecial   bool `json:"require_special"`
}

// GetType returns RealmTypeBasicClientIDSecret.
func (c *BasicClientIDSecretConfig) GetType() RealmType {
	return RealmTypeBasicClientIDSecret
}

// Validate validates the configuration.
func (c *BasicClientIDSecretConfig) Validate() error {
	if c.MinSecretLength < 1 {
		return fmt.Errorf("min_secret_length must be at least 1")
	}

	return nil
}

// RealmService defines operations for managing tenant realms.
type RealmService interface {
	// CreateRealm creates a new realm for a tenant.
	CreateRealm(ctx context.Context, tenantID googleUuid.UUID, realmType string, config RealmConfig) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error)

	// GetRealm retrieves a realm by ID.
	GetRealm(ctx context.Context, tenantID, realmID googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error)

	// ListRealms lists all realms for a tenant.
	ListRealms(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error)

	// GetFirstActiveRealm retrieves the first active realm for a tenant.
	// Returns the realm if found, or nil with no error if no active realms exist.
	// This is used for single-realm tenants or as a default realm selection strategy.
	GetFirstActiveRealm(ctx context.Context, tenantID googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error)

	// UpdateRealm updates realm configuration.
	UpdateRealm(ctx context.Context, tenantID, realmID googleUuid.UUID, config RealmConfig, active *bool) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error)

	// DeleteRealm deactivates a realm (soft delete).
	DeleteRealm(ctx context.Context, tenantID, realmID googleUuid.UUID) error

	// GetRealmConfig parses and returns the typed configuration for a realm.
	GetRealmConfig(ctx context.Context, tenantID, realmID googleUuid.UUID) (RealmConfig, error)
}

// RealmServiceImpl implements RealmService.
type RealmServiceImpl struct {
	realmRepo cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository
}

// NewRealmService creates a new RealmService instance.
func NewRealmService(realmRepo cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository) RealmService {
	return &RealmServiceImpl{
		realmRepo: realmRepo,
	}
}

// CreateRealm creates a new realm for a tenant.
func (s *RealmServiceImpl) CreateRealm(ctx context.Context, tenantID googleUuid.UUID, realmType string, config RealmConfig) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	// Validate the realm type.
	if err := s.validateRealmType(realmType); err != nil {
		return nil, err
	}

	// Validate the configuration if provided.
	if config != nil {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid realm configuration: %w", err)
		}
	}

	// Serialize configuration to JSON.
	var configJSON string

	if config != nil {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize realm configuration: %w", err)
		}

		configJSON = string(configBytes)
	}

	realm := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenantID,
		RealmID:  googleUuid.New(),
		Type:     realmType,
		Config:   configJSON,
		Active:   true,
		Source:   "db",
	}

	if err := s.realmRepo.Create(ctx, realm); err != nil {
		return nil, fmt.Errorf("failed to create realm: %w", err)
	}

	return realm, nil
}

// GetRealm retrieves a realm by ID.
func (s *RealmServiceImpl) GetRealm(ctx context.Context, tenantID, realmID googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	realm, err := s.realmRepo.GetByRealmID(ctx, tenantID, realmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get realm: %w", err)
	}

	// Verify tenant ownership.
	if realm.TenantID != tenantID {
		return nil, fmt.Errorf("realm does not belong to the specified tenant")
	}

	return realm, nil
}

// ListRealms lists all realms for a tenant.
func (s *RealmServiceImpl) ListRealms(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	realms, err := s.realmRepo.ListByTenant(ctx, tenantID, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list realms: %w", err)
	}

	return realms, nil
}

// GetFirstActiveRealm retrieves the first active realm for a tenant.
// Returns the realm if found, or nil with no error if no active realms exist.
// This is used for single-realm tenants or as a default realm selection strategy.
func (s *RealmServiceImpl) GetFirstActiveRealm(ctx context.Context, tenantID googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	realms, err := s.realmRepo.ListByTenant(ctx, tenantID, true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("failed to list realms: %w", err)
	}

	// Return nil if no active realms exist (not an error condition)
	if len(realms) == 0 {
		return nil, nil
	}

	// Return the first active realm
	return realms[0], nil
}

// UpdateRealm updates realm configuration.
func (s *RealmServiceImpl) UpdateRealm(ctx context.Context, tenantID, realmID googleUuid.UUID, config RealmConfig, active *bool) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	// Get existing realm.
	realm, err := s.realmRepo.GetByRealmID(ctx, tenantID, realmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get realm: %w", err)
	}

	// Verify tenant ownership.
	if realm.TenantID != tenantID {
		return nil, fmt.Errorf("realm does not belong to the specified tenant")
	}

	if config != nil {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid realm configuration: %w", err)
		}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize realm configuration: %w", err)
		}

		realm.Config = string(configBytes)
	}

	if active != nil {
		realm.Active = *active
	}

	if err := s.realmRepo.Update(ctx, realm); err != nil {
		return nil, fmt.Errorf("failed to update realm: %w", err)
	}

	return realm, nil
}

// DeleteRealm deactivates a realm (soft delete).
func (s *RealmServiceImpl) DeleteRealm(ctx context.Context, tenantID, realmID googleUuid.UUID) error {
	// Get existing realm.
	realm, err := s.realmRepo.GetByRealmID(ctx, tenantID, realmID)
	if err != nil {
		return fmt.Errorf("failed to get realm: %w", err)
	}

	// Verify tenant ownership.
	if realm.TenantID != tenantID {
		return fmt.Errorf("realm does not belong to the specified tenant")
	}

	// Soft delete by marking inactive.
	realm.Active = false
	if err := s.realmRepo.Update(ctx, realm); err != nil {
		return fmt.Errorf("failed to deactivate realm: %w", err)
	}

	return nil
}

// GetRealmConfig parses and returns the typed configuration for a realm.
func (s *RealmServiceImpl) GetRealmConfig(ctx context.Context, tenantID, realmID googleUuid.UUID) (RealmConfig, error) {
	realm, err := s.realmRepo.GetByRealmID(ctx, tenantID, realmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get realm: %w", err)
	}

	// Verify tenant ownership.
	if realm.TenantID != tenantID {
		return nil, fmt.Errorf("realm does not belong to the specified tenant")
	}

	// Parse configuration based on realm type.
	return s.parseRealmConfig(realm.Type, realm.Config)
}

// validateRealmType validates that the realm type is supported.
func (s *RealmServiceImpl) validateRealmType(realmType string) error {
	switch RealmType(realmType) {
	// Federated realm types.
	case RealmTypeUsernamePassword, RealmTypeLDAP, RealmTypeOAuth2, RealmTypeSAML:
		return nil
	// Non-federated browser realm types.
	case RealmTypeJWESessionCookie, RealmTypeJWSSessionCookie, RealmTypeOpaqueSessionCookie:
		return nil
	case RealmTypeBasicUsernamePassword, RealmTypeBearerAPIToken, RealmTypeHTTPSClientCert:
		return nil
	// Non-federated service realm types.
	case RealmTypeJWESessionToken, RealmTypeJWSSessionToken, RealmTypeOpaqueSessionToken:
		return nil
	case RealmTypeBasicClientIDSecret:
		return nil
	default:
		return fmt.Errorf("unsupported realm type: %s", realmType)
	}
}

// parseRealmConfig parses JSON configuration into the appropriate typed config.
func (s *RealmServiceImpl) parseRealmConfig(realmType, configJSON string) (RealmConfig, error) {
	if configJSON == "" {
		return nil, nil //nolint:nilnil // Empty config is valid - returns nil value with nil error
	}

	var config RealmConfig

	switch RealmType(realmType) {
	// Federated realm types.
	case RealmTypeUsernamePassword:
		config = &UsernamePasswordConfig{}
	case RealmTypeLDAP:
		config = &LDAPConfig{}
	case RealmTypeOAuth2:
		config = &OAuth2Config{}
	case RealmTypeSAML:
		config = &SAMLConfig{}
	// Non-federated browser realm types.
	case RealmTypeJWESessionCookie:
		config = &JWESessionCookieConfig{}
	case RealmTypeJWSSessionCookie:
		config = &JWSSessionCookieConfig{}
	case RealmTypeOpaqueSessionCookie:
		config = &OpaqueSessionCookieConfig{}
	case RealmTypeBasicUsernamePassword:
		config = &BasicUsernamePasswordConfig{}
	case RealmTypeBearerAPIToken:
		config = &BearerAPITokenConfig{}
	case RealmTypeHTTPSClientCert:
		config = &HTTPSClientCertConfig{}
	// Non-federated service realm types.
	case RealmTypeJWESessionToken:
		config = &JWESessionTokenConfig{}
	case RealmTypeJWSSessionToken:
		config = &JWSSessionTokenConfig{}
	case RealmTypeOpaqueSessionToken:
		config = &OpaqueSessionTokenConfig{}
	case RealmTypeBasicClientIDSecret:
		config = &BasicClientIDSecretConfig{}
	default:
		return nil, fmt.Errorf("unsupported realm type: %s", realmType)
	}

	if err := json.Unmarshal([]byte(configJSON), config); err != nil {
		return nil, fmt.Errorf("failed to parse realm configuration: %w", err)
	}

	return config, nil
}
