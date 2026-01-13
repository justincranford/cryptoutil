// Copyright (c) 2025 Justin Cranford

package businesslogic

import (
	"context"
	"encoding/json"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilCipherImRepository "cryptoutil/internal/apps/cipher/im/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// Default token expiry for Bearer tokens (1 hour).
	defaultBearerTokenExpiry = 3600

	// Enterprise token expiry for Bearer tokens (30 minutes).
	enterpriseBearerTokenExpiry = 1800
)

// RealmService provides business logic for realm management.
type RealmService interface {
	CreateRealm(ctx context.Context, realmType cryptoutilCipherImRepository.RealmType, name string, config *cryptoutilCipherImRepository.RealmConfig) (*cryptoutilCipherImRepository.Realm, error)
	GetRealm(ctx context.Context, realmID googleUuid.UUID) (*cryptoutilCipherImRepository.Realm, error)
	GetRealmByName(ctx context.Context, name string) (*cryptoutilCipherImRepository.Realm, error)
	GetRealmConfig(ctx context.Context, realmID googleUuid.UUID) (*cryptoutilCipherImRepository.RealmConfig, error)
	ListRealms(ctx context.Context, activeOnly bool) ([]*cryptoutilCipherImRepository.Realm, error)
	UpdateRealm(ctx context.Context, realm *cryptoutilCipherImRepository.Realm) error
	DeleteRealm(ctx context.Context, realmID googleUuid.UUID) error
	GetActiveByPriority(ctx context.Context) ([]*cryptoutilCipherImRepository.Realm, error)
}

// RealmServiceImpl implements RealmService.
type RealmServiceImpl struct {
	realmRepo cryptoutilCipherImRepository.RealmRepository
}

// NewRealmService creates a new RealmService.
func NewRealmService(realmRepo cryptoutilCipherImRepository.RealmRepository) RealmService {
	return &RealmServiceImpl{
		realmRepo: realmRepo,
	}
}

// CreateRealm creates a new authentication realm.
func (s *RealmServiceImpl) CreateRealm(ctx context.Context, realmType cryptoutilCipherImRepository.RealmType, name string, config *cryptoutilCipherImRepository.RealmConfig) (*cryptoutilCipherImRepository.Realm, error) {
	// Validate realm type.
	if err := s.validateRealmType(realmType); err != nil {
		return nil, err
	}

	// Validate name.
	if name == "" {
		return nil, fmt.Errorf("realm name cannot be empty")
	}

	// Serialize config to JSON.
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal realm config: %w", err)
	}

	// Create realm.
	realm := &cryptoutilCipherImRepository.Realm{
		ID:       googleUuid.Must(googleUuid.NewV7()),
		RealmID:  googleUuid.Must(googleUuid.NewV7()),
		Type:     realmType,
		Name:     name,
		Config:   string(configJSON),
		Active:   true,
		Source:   "db",
		Priority: 0,
	}

	if err := s.realmRepo.Create(ctx, realm); err != nil {
		return nil, fmt.Errorf("failed to create realm: %w", err)
	}

	return realm, nil
}

// GetRealm retrieves a realm by ID.
func (s *RealmServiceImpl) GetRealm(ctx context.Context, realmID googleUuid.UUID) (*cryptoutilCipherImRepository.Realm, error) {
	realm, err := s.realmRepo.GetByRealmID(ctx, realmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get realm: %w", err)
	}

	return realm, nil
}

// GetRealmByName retrieves a realm by name.
func (s *RealmServiceImpl) GetRealmByName(ctx context.Context, name string) (*cryptoutilCipherImRepository.Realm, error) {
	realm, err := s.realmRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get realm by name: %w", err)
	}

	return realm, nil
}

// GetRealmConfig parses and returns the typed configuration for a realm.
func (s *RealmServiceImpl) GetRealmConfig(ctx context.Context, realmID googleUuid.UUID) (*cryptoutilCipherImRepository.RealmConfig, error) {
	realm, err := s.realmRepo.GetByRealmID(ctx, realmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get realm: %w", err)
	}

	// Parse configuration.
	var config cryptoutilCipherImRepository.RealmConfig

	if err := json.Unmarshal([]byte(realm.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal realm config: %w", err)
	}

	return &config, nil
}

// ListRealms retrieves all realms.
func (s *RealmServiceImpl) ListRealms(ctx context.Context, activeOnly bool) ([]*cryptoutilCipherImRepository.Realm, error) {
	realms, err := s.realmRepo.ListAll(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list realms: %w", err)
	}

	return realms, nil
}

// GetActiveByPriority retrieves active realms ordered by priority.
func (s *RealmServiceImpl) GetActiveByPriority(ctx context.Context) ([]*cryptoutilCipherImRepository.Realm, error) {
	realms, err := s.realmRepo.GetActiveByPriority(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active realms by priority: %w", err)
	}

	return realms, nil
}

// UpdateRealm updates a realm.
func (s *RealmServiceImpl) UpdateRealm(ctx context.Context, realm *cryptoutilCipherImRepository.Realm) error {
	// Validate realm type.
	if err := s.validateRealmType(realm.Type); err != nil {
		return err
	}

	if err := s.realmRepo.Update(ctx, realm); err != nil {
		return fmt.Errorf("failed to update realm: %w", err)
	}

	return nil
}

// DeleteRealm deletes a realm.
func (s *RealmServiceImpl) DeleteRealm(ctx context.Context, realmID googleUuid.UUID) error {
	realm, err := s.realmRepo.GetByRealmID(ctx, realmID)
	if err != nil {
		return fmt.Errorf("failed to get realm: %w", err)
	}

	if err := s.realmRepo.Delete(ctx, realm.ID); err != nil {
		return fmt.Errorf("failed to delete realm: %w", err)
	}

	return nil
}

// validateRealmType validates that the realm type is supported.
func (s *RealmServiceImpl) validateRealmType(realmType cryptoutilCipherImRepository.RealmType) error {
	switch realmType {
	case cryptoutilCipherImRepository.RealmTypeJWESessionCookie,
		cryptoutilCipherImRepository.RealmTypeJWSSessionCookie,
		cryptoutilCipherImRepository.RealmTypeOpaqueSessionCookie,
		cryptoutilCipherImRepository.RealmTypeBasicAuth,
		cryptoutilCipherImRepository.RealmTypeBearerToken,
		cryptoutilCipherImRepository.RealmTypeHTTPSClientCert:
		return nil
	default:
		return fmt.Errorf("unsupported realm type: %s", string(realmType))
	}
}

// DefaultRealmConfig returns the default realm configuration.
// Used when no specific realm is configured or as fallback.
func DefaultRealmConfig() *cryptoutilCipherImRepository.RealmConfig {
	return &cryptoutilCipherImRepository.RealmConfig{
		PasswordMinLength:        cryptoutilSharedMagic.CipherDefaultPasswordMinLength,
		PasswordRequireUppercase: true,
		PasswordRequireLowercase: true,
		PasswordRequireDigits:    true,
		PasswordRequireSpecial:   true,
		PasswordMinUniqueChars:   cryptoutilSharedMagic.CipherDefaultPasswordMinUniqueChars,
		PasswordMaxRepeatedChars: cryptoutilSharedMagic.CipherDefaultPasswordMaxRepeatedChars,
		SessionTimeout:           cryptoutilSharedMagic.CipherDefaultSessionTimeout,
		SessionAbsoluteMax:       cryptoutilSharedMagic.CipherDefaultSessionAbsoluteMax,
		SessionRefreshEnabled:    true,
		TokenExpiry:              defaultBearerTokenExpiry, // 1 hour for Bearer tokens.
		MFARequired:              false,
		MFAMethods:               []string{},
		LoginRateLimit:           cryptoutilSharedMagic.CipherDefaultLoginRateLimit,
		MessageRateLimit:         cryptoutilSharedMagic.CipherDefaultMessageRateLimit,
		RequireClientCert:        false,
		TrustedCAs:               []string{},
	}
}

// EnterpriseRealmConfig returns a more restrictive realm configuration for enterprise deployments.
func EnterpriseRealmConfig() *cryptoutilCipherImRepository.RealmConfig {
	return &cryptoutilCipherImRepository.RealmConfig{
		PasswordMinLength:        cryptoutilSharedMagic.CipherEnterprisePasswordMinLength,
		PasswordRequireUppercase: true,
		PasswordRequireLowercase: true,
		PasswordRequireDigits:    true,
		PasswordRequireSpecial:   true,
		PasswordMinUniqueChars:   cryptoutilSharedMagic.CipherEnterprisePasswordMinUniqueChars,
		PasswordMaxRepeatedChars: cryptoutilSharedMagic.CipherEnterprisePasswordMaxRepeatedChars,
		SessionTimeout:           cryptoutilSharedMagic.CipherEnterpriseSessionTimeout,
		SessionAbsoluteMax:       cryptoutilSharedMagic.CipherEnterpriseSessionAbsoluteMax,
		SessionRefreshEnabled:    true,
		TokenExpiry:              enterpriseBearerTokenExpiry, // 30 minutes for Bearer tokens.
		MFARequired:              true,
		MFAMethods:               []string{"totp", "webauthn"},
		LoginRateLimit:           cryptoutilSharedMagic.CipherEnterpriseLoginRateLimit,
		MessageRateLimit:         cryptoutilSharedMagic.CipherEnterpriseMessageRateLimit,
		RequireClientCert:        true,
		TrustedCAs:               []string{},
	}
}
