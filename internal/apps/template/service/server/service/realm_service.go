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
	"encoding/json"
	"fmt"

	googleUuid "github.com/google/uuid"

	"cryptoutil/internal/apps/template/service/server/repository"
)

// RealmType represents the type of authentication realm.
type RealmType string

const (
	// RealmTypeUsernamePassword is a database-based username/password realm.
	RealmTypeUsernamePassword RealmType = "username_password"

	// RealmTypeLDAP is an LDAP-based authentication realm.
	RealmTypeLDAP RealmType = "ldap"

	// RealmTypeOAuth2 is an OAuth2/OIDC-based authentication realm.
	RealmTypeOAuth2 RealmType = "oauth2"

	// RealmTypeSAML is a SAML-based authentication realm.
	RealmTypeSAML RealmType = "saml"
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
	URL            string `json:"url"`
	BindDN         string `json:"bind_dn"`
	BindPassword   string `json:"bind_password"`
	BaseDN         string `json:"base_dn"`
	UserFilter     string `json:"user_filter"`
	GroupFilter    string `json:"group_filter"`
	UseTLS         bool   `json:"use_tls"`
	SkipTLSVerify  bool   `json:"skip_tls_verify"`
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
	ProviderURL    string   `json:"provider_url"`
	ClientID       string   `json:"client_id"`
	ClientSecret   string   `json:"client_secret"`
	Scopes         []string `json:"scopes"`
	RedirectURI    string   `json:"redirect_uri"`
	UseDiscovery   bool     `json:"use_discovery"`
	AuthorizeURL   string   `json:"authorize_url"`
	TokenURL       string   `json:"token_url"`
	UserInfoURL    string   `json:"userinfo_url"`
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

// RealmService defines operations for managing tenant realms.
type RealmService interface {
	// CreateRealm creates a new realm for a tenant.
	CreateRealm(ctx context.Context, tenantID googleUuid.UUID, realmType string, config RealmConfig) (*repository.TenantRealm, error)

	// GetRealm retrieves a realm by ID.
	GetRealm(ctx context.Context, tenantID, realmID googleUuid.UUID) (*repository.TenantRealm, error)

	// ListRealms lists all realms for a tenant.
	ListRealms(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*repository.TenantRealm, error)

	// UpdateRealm updates realm configuration.
	UpdateRealm(ctx context.Context, tenantID, realmID googleUuid.UUID, config RealmConfig, active *bool) (*repository.TenantRealm, error)

	// DeleteRealm deactivates a realm (soft delete).
	DeleteRealm(ctx context.Context, tenantID, realmID googleUuid.UUID) error

	// GetRealmConfig parses and returns the typed configuration for a realm.
	GetRealmConfig(ctx context.Context, tenantID, realmID googleUuid.UUID) (RealmConfig, error)
}

// RealmServiceImpl implements RealmService.
type RealmServiceImpl struct {
	realmRepo repository.TenantRealmRepository
}

// NewRealmService creates a new RealmService instance.
func NewRealmService(realmRepo repository.TenantRealmRepository) RealmService {
	return &RealmServiceImpl{
		realmRepo: realmRepo,
	}
}

// CreateRealm creates a new realm for a tenant.
func (s *RealmServiceImpl) CreateRealm(ctx context.Context, tenantID googleUuid.UUID, realmType string, config RealmConfig) (*repository.TenantRealm, error) {
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

	realm := &repository.TenantRealm{
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
func (s *RealmServiceImpl) GetRealm(ctx context.Context, tenantID, realmID googleUuid.UUID) (*repository.TenantRealm, error) {
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
func (s *RealmServiceImpl) ListRealms(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*repository.TenantRealm, error) {
	return s.realmRepo.ListByTenant(ctx, tenantID, activeOnly)
}

// UpdateRealm updates realm configuration.
func (s *RealmServiceImpl) UpdateRealm(ctx context.Context, tenantID, realmID googleUuid.UUID, config RealmConfig, active *bool) (*repository.TenantRealm, error) {
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
	case RealmTypeUsernamePassword, RealmTypeLDAP, RealmTypeOAuth2, RealmTypeSAML:
		return nil
	default:
		return fmt.Errorf("unsupported realm type: %s", realmType)
	}
}

// parseRealmConfig parses JSON configuration into the appropriate typed config.
func (s *RealmServiceImpl) parseRealmConfig(realmType, configJSON string) (RealmConfig, error) {
	if configJSON == "" {
		return nil, nil
	}

	var config RealmConfig
	switch RealmType(realmType) {
	case RealmTypeUsernamePassword:
		config = &UsernamePasswordConfig{}
	case RealmTypeLDAP:
		config = &LDAPConfig{}
	case RealmTypeOAuth2:
		config = &OAuth2Config{}
	case RealmTypeSAML:
		config = &SAMLConfig{}
	default:
		return nil, fmt.Errorf("unsupported realm type: %s", realmType)
	}

	if err := json.Unmarshal([]byte(configJSON), config); err != nil {
		return nil, fmt.Errorf("failed to parse realm configuration: %w", err)
	}

	return config, nil
}
