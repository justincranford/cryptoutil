// Copyright (c) 2025 Justin Cranford
//

// Package builder provides fluent API for constructing service applications.
package builder

import (
	"context"
	"errors"
)

// ErrBarrierConfigRequired is returned when barrier config validation fails.
var ErrBarrierConfigRequired = errors.New("barrier config is required")

// BarrierConfig configures the barrier service for ServerBuilder.
// Uses template barrier (context-based, GORM transactions) by default.
type BarrierConfig struct {
	// EnableRotationEndpoints enables the /admin/api/v1/barrier/* rotation endpoints.
	EnableRotationEndpoints bool

	// EnableStatusEndpoints enables the /admin/api/v1/barrier/status endpoint.
	EnableStatusEndpoints bool
}

// NewBarrierConfig creates a default BarrierConfig with all endpoints enabled.
func NewBarrierConfig() *BarrierConfig {
	return &BarrierConfig{
		EnableRotationEndpoints: true,
		EnableStatusEndpoints:   true,
	}
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
	if c == nil {
		return ErrBarrierConfigRequired
	}

	return nil
}

// BarrierEncryptor provides a common interface for encryption operations.
type BarrierEncryptor interface {
	// EncryptContent encrypts content bytes using the active content key.
	// ctx should contain the GORM transaction.
	EncryptContent(ctx context.Context, clearBytes []byte) ([]byte, error)

	// DecryptContent decrypts JWE message bytes to recover the original content.
	// ctx should contain the GORM transaction.
	DecryptContent(ctx context.Context, encryptedJWEBytes []byte) ([]byte, error)

	// Shutdown gracefully shuts down the barrier service.
	Shutdown()
}
