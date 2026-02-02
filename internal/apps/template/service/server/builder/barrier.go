// Copyright (c) 2025 Justin Cranford
//

// Package builder provides fluent API for constructing service applications.
package builder

import (
	"context"
	"errors"
)

// BarrierMode defines how the barrier service operates.
type BarrierMode string

const (
	// BarrierModeTemplate uses the template-specific barrier (context-based, GORM transactions).
	// This is the default mode for services using the template pattern.
	BarrierModeTemplate BarrierMode = "template"

	// BarrierModeShared uses the shared barrier (transaction-based, OrmRepository).
	// This mode is used by KMS and services requiring raw SQL transaction support.
	BarrierModeShared BarrierMode = "shared"

	// BarrierModeDisabled disables the barrier service entirely.
	// Use this for services that don't need encryption-at-rest.
	BarrierModeDisabled BarrierMode = "disabled"
)

// ErrBarrierModeRequired is returned when barrier mode is not specified.
var ErrBarrierModeRequired = errors.New("barrier mode is required")

// BarrierConfig configures the barrier service for ServerBuilder.
// This abstraction allows ServerBuilder to work with either the template barrier
// (context-based, GORM transactions) or the shared barrier (transaction-based, OrmRepository).
type BarrierConfig struct {
	// Mode determines which barrier implementation to use.
	Mode BarrierMode

	// EnableRotationEndpoints enables the /admin/api/v1/barrier/* rotation endpoints.
	// Only applicable for BarrierModeTemplate.
	EnableRotationEndpoints bool

	// EnableStatusEndpoints enables the /admin/api/v1/barrier/status endpoint.
	// Only applicable for BarrierModeTemplate.
	EnableStatusEndpoints bool
}

// NewDefaultBarrierConfig creates a default BarrierConfig using template mode.
// Rotation and status endpoints are enabled by default.
func NewDefaultBarrierConfig() *BarrierConfig {
	return &BarrierConfig{
		Mode:                    BarrierModeTemplate,
		EnableRotationEndpoints: true,
		EnableStatusEndpoints:   true,
	}
}

// NewSharedBarrierConfig creates a BarrierConfig for shared barrier mode.
// Rotation and status endpoints are disabled (not available in shared barrier).
func NewSharedBarrierConfig() *BarrierConfig {
	return &BarrierConfig{
		Mode:                    BarrierModeShared,
		EnableRotationEndpoints: false,
		EnableStatusEndpoints:   false,
	}
}

// NewDisabledBarrierConfig creates a BarrierConfig that disables the barrier.
func NewDisabledBarrierConfig() *BarrierConfig {
	return &BarrierConfig{
		Mode:                    BarrierModeDisabled,
		EnableRotationEndpoints: false,
		EnableStatusEndpoints:   false,
	}
}

// WithMode sets the barrier mode.
func (c *BarrierConfig) WithMode(mode BarrierMode) *BarrierConfig {
	c.Mode = mode

	return c
}

// WithRotationEndpoints enables or disables rotation endpoints.
func (c *BarrierConfig) WithRotationEndpoints(enabled bool) *BarrierConfig {
	c.EnableRotationEndpoints = enabled

	return c
}

// WithStatusEndpoints enables or disables status endpoints.
func (c *BarrierConfig) WithStatusEndpoints(enabled bool) *BarrierConfig {
	c.EnableStatusEndpoints = enabled

	return c
}

// Validate checks that the configuration is valid.
func (c *BarrierConfig) Validate() error {
	if c.Mode == "" {
		return ErrBarrierModeRequired
	}

	switch c.Mode {
	case BarrierModeTemplate, BarrierModeShared, BarrierModeDisabled:
		return nil
	default:
		return errors.New("invalid barrier mode: " + string(c.Mode))
	}
}

// IsEnabled returns true if the barrier is enabled.
func (c *BarrierConfig) IsEnabled() bool {
	return c.Mode != BarrierModeDisabled
}

// BarrierEncryptor provides a common interface for encryption operations.
// This allows code to work with either barrier implementation.
type BarrierEncryptor interface {
	// EncryptContent encrypts content bytes using the active content key.
	// For template barrier, ctx should contain the GORM transaction.
	// For shared barrier, ctx should contain the OrmTransaction.
	EncryptContent(ctx context.Context, clearBytes []byte) ([]byte, error)

	// DecryptContent decrypts JWE message bytes to recover the original content.
	// For template barrier, ctx should contain the GORM transaction.
	// For shared barrier, ctx should contain the OrmTransaction.
	DecryptContent(ctx context.Context, encryptedJWEBytes []byte) ([]byte, error)

	// Shutdown gracefully shuts down the barrier service.
	Shutdown()
}
