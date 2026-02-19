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
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ErrNoActiveRealm is returned by GetFirstActiveRealm when no active realm exists for the tenant.
var ErrNoActiveRealm = errors.New("no active realm found for tenant")

// RealmType represents the type of authentication realm.
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
	// Returns the realm if found, or (nil, ErrNoActiveRealm) if no active realms exist.
	// Callers MUST check errors.Is(err, ErrNoActiveRealm) to distinguish from real errors.
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
// Returns the realm if found, or (nil, ErrNoActiveRealm) if no active realms exist.
// Callers MUST check errors.Is(err, ErrNoActiveRealm) to distinguish from real errors.
func (s *RealmServiceImpl) GetFirstActiveRealm(ctx context.Context, tenantID googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	realms, err := s.realmRepo.ListByTenant(ctx, tenantID, true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("failed to list realms: %w", err)
	}

	// Return sentinel when no active realms exist (not a real error, callers check errors.Is).
	if len(realms) == 0 {
		return nil, ErrNoActiveRealm
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
