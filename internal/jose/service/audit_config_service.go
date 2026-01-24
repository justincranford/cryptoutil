// Copyright (c) 2025 Justin Cranford
//
//

// Package service provides JOSE-JA business logic services.
package service

import (
	"context"
	"fmt"
	"strings"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"

	googleUuid "github.com/google/uuid"
)

// Default audit configuration values.
const (
	// DefaultAuditSamplingRate is 1% sampling (0.01).
	DefaultAuditSamplingRate = 0.01

	// DefaultAuditEnabled is true - audit enabled by default.
	DefaultAuditEnabled = true
)

// Supported audit operations.
const (
	AuditOperationEncrypt  = "encrypt"
	AuditOperationDecrypt  = "decrypt"
	AuditOperationSign     = "sign"
	AuditOperationVerify   = "verify"
	AuditOperationKeyGen   = "keygen"
	AuditOperationRotate   = "rotate"
	AuditOperationGetJWKS  = "get_jwks"
	AuditOperationGetKey   = "get_key"
	AuditOperationListKeys = "list_keys"
)

// AllAuditOperations lists all supported audit operations.
var AllAuditOperations = []string{
	AuditOperationEncrypt,
	AuditOperationDecrypt,
	AuditOperationSign,
	AuditOperationVerify,
	AuditOperationKeyGen,
	AuditOperationRotate,
	AuditOperationGetJWKS,
	AuditOperationGetKey,
	AuditOperationListKeys,
}

// AuditConfigService provides business logic for audit configuration management.
type AuditConfigService struct {
	repo cryptoutilJoseRepository.AuditConfigRepository
}

// NewAuditConfigService creates a new audit config service.
func NewAuditConfigService(repo cryptoutilJoseRepository.AuditConfigRepository) *AuditConfigService {
	return &AuditConfigService{
		repo: repo,
	}
}

// GetConfig retrieves audit config for a tenant and operation.
// Returns the config if found, or creates a default config if not found.
func (s *AuditConfigService) GetConfig(ctx context.Context, tenantID googleUuid.UUID, operation string) (*cryptoutilJoseDomain.AuditConfig, error) {
	if !isValidOperation(operation) {
		return nil, fmt.Errorf("invalid operation: %s", operation)
	}

	config, err := s.repo.Get(ctx, tenantID, operation)
	if err != nil {
		// If not found error, return default values.
		if isNotFoundAuditConfigError(err) {
			return &cryptoutilJoseDomain.AuditConfig{
				TenantID:     tenantID,
				Operation:    operation,
				Enabled:      DefaultAuditEnabled,
				SamplingRate: DefaultAuditSamplingRate,
			}, nil
		}

		return nil, fmt.Errorf("failed to get audit config: %w", err)
	}

	return config, nil
}

// isNotFoundAuditConfigError checks if the error is a "not found" error.
func isNotFoundAuditConfigError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "not found")
}

// GetAllConfigs retrieves all audit configs for a tenant.
// For operations not configured, returns defaults.
func (s *AuditConfigService) GetAllConfigs(ctx context.Context, tenantID googleUuid.UUID) ([]cryptoutilJoseDomain.AuditConfig, error) {
	// Get existing configs.
	existing, err := s.repo.GetAll(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit configs: %w", err)
	}

	// Build a map of existing configs.
	existingMap := make(map[string]cryptoutilJoseDomain.AuditConfig)

	for _, config := range existing {
		existingMap[config.Operation] = config
	}

	// Build result with defaults for missing operations.
	result := make([]cryptoutilJoseDomain.AuditConfig, 0, len(AllAuditOperations))

	for _, op := range AllAuditOperations {
		if config, found := existingMap[op]; found {
			result = append(result, config)
		} else {
			result = append(result, cryptoutilJoseDomain.AuditConfig{
				TenantID:     tenantID,
				Operation:    op,
				Enabled:      DefaultAuditEnabled,
				SamplingRate: DefaultAuditSamplingRate,
			})
		}
	}

	return result, nil
}

// SetConfig creates or updates audit config for a tenant and operation.
func (s *AuditConfigService) SetConfig(ctx context.Context, tenantID googleUuid.UUID, operation string, enabled bool, samplingRate float64) error {
	if !isValidOperation(operation) {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	if samplingRate < 0 || samplingRate > 1 {
		return fmt.Errorf("sampling rate must be between 0.0 and 1.0, got %f", samplingRate)
	}

	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      enabled,
		SamplingRate: samplingRate,
	}

	if err := s.repo.Upsert(ctx, config); err != nil {
		return fmt.Errorf("failed to upsert audit config: %w", err)
	}

	return nil
}

// InitializeDefaults creates default audit configs for all operations for a tenant.
// This is typically called when a new tenant is created.
func (s *AuditConfigService) InitializeDefaults(ctx context.Context, tenantID googleUuid.UUID) error {
	for _, op := range AllAuditOperations {
		config := &cryptoutilJoseDomain.AuditConfig{
			TenantID:     tenantID,
			Operation:    op,
			Enabled:      DefaultAuditEnabled,
			SamplingRate: DefaultAuditSamplingRate,
		}

		if err := s.repo.Upsert(ctx, config); err != nil {
			return fmt.Errorf("failed to initialize audit config for operation %s: %w", op, err)
		}
	}

	return nil
}

// IsEnabled checks if audit is enabled for a tenant and operation.
// Returns the enabled status and sampling rate.
func (s *AuditConfigService) IsEnabled(ctx context.Context, tenantID googleUuid.UUID, operation string) (enabled bool, samplingRate float64, err error) {
	if !isValidOperation(operation) {
		return false, 0, fmt.Errorf("invalid operation: %s", operation)
	}

	enabled, samplingRate, err = s.repo.IsEnabled(ctx, tenantID, operation)
	if err != nil {
		return false, 0, fmt.Errorf("failed to check audit enabled: %w", err)
	}

	// If not found in DB, return defaults.
	if !enabled && samplingRate == 0 {
		return DefaultAuditEnabled, DefaultAuditSamplingRate, nil
	}

	return enabled, samplingRate, nil
}

// isValidOperation checks if the operation is a valid audit operation.
func isValidOperation(operation string) bool {
	for _, op := range AllAuditOperations {
		if op == operation {
			return true
		}
	}

	return false
}
