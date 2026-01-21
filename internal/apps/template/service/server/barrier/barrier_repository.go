// Copyright (c) 2025 Justin Cranford
//
//

// Package barrier provides hierarchical key management for root, intermediate, and content encryption keys.
package barrier

import (
	"context"

	googleUuid "github.com/google/uuid"
)

// BarrierRepository defines the interface for barrier key storage operations.
// This abstraction allows BarrierService to work with different database implementations
// (KMS OrmRepository, cipher-im gorm.DB, etc.) without coupling to specific repository types.
type BarrierRepository interface {
	// WithTransaction executes the provided function within a database transaction.
	// The transaction will be automatically committed on success or rolled back on error.
	WithTransaction(ctx context.Context, function func(tx BarrierTransaction) error) error

	// Shutdown releases any resources held by the repository.
	Shutdown()
}

// BarrierTransaction defines the interface for transactional barrier key operations.
// Implementations must provide ACID guarantees for barrier key lifecycle operations.
type BarrierTransaction interface {
	// Context returns the transaction context.
	Context() context.Context

	// Root Key Operations

	// GetRootKeyLatest retrieves the most recently created root key.
	// Returns (nil, nil) if no root keys exist.
	GetRootKeyLatest() (*BarrierRootKey, error)

	// GetRootKey retrieves a specific root key by UUID.
	// Returns error if key not found.
	GetRootKey(uuid *googleUuid.UUID) (*BarrierRootKey, error)

	// AddRootKey persists a new root key to storage.
	AddRootKey(key *BarrierRootKey) error

	// Intermediate Key Operations

	// GetIntermediateKeyLatest retrieves the most recently created intermediate key.
	// Returns (nil, nil) if no intermediate keys exist.
	GetIntermediateKeyLatest() (*BarrierIntermediateKey, error)

	// GetIntermediateKey retrieves a specific intermediate key by UUID.
	// Returns error if key not found.
	GetIntermediateKey(uuid *googleUuid.UUID) (*BarrierIntermediateKey, error)

	// AddIntermediateKey persists a new intermediate key to storage.
	AddIntermediateKey(key *BarrierIntermediateKey) error

	// Content Key Operations

	// GetContentKey retrieves a specific content key by UUID.
	// Returns error if key not found.
	GetContentKey(uuid *googleUuid.UUID) (*BarrierContentKey, error)

	// AddContentKey persists a new content key to storage.
	AddContentKey(key *BarrierContentKey) error
}

// BarrierRootKey represents a root-level encryption key in the barrier hierarchy.
// Root keys are encrypted by the unseal key (HSM/KMS/Shamir).
type BarrierRootKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted root key
	KEKUUID   googleUuid.UUID `gorm:"type:text"`                              // KEK UUID (nil for root keys)
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
	// TODO: Add RotatedAt *int64 after fixing migration 0004 discovery issue
}

// TableName specifies the database table name for barrier root keys.
func (BarrierRootKey) TableName() string {
	return "barrier_root_keys"
}

// BarrierIntermediateKey represents an intermediate-level encryption key in the barrier hierarchy.
// Intermediate keys are encrypted by root keys.
type BarrierIntermediateKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted intermediate key
	KEKUUID   googleUuid.UUID `gorm:"type:text;not null"`                     // Parent root key UUID
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
	// TODO: Add RotatedAt *int64 after fixing migration 0004 discovery issue
}

// TableName specifies the database table name for barrier intermediate keys.
func (BarrierIntermediateKey) TableName() string {
	return "barrier_intermediate_keys"
}

// BarrierContentKey represents a content-level encryption key in the barrier hierarchy.
// Content keys are encrypted by intermediate keys and used for actual data encryption.
type BarrierContentKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted content key
	KEKUUID   googleUuid.UUID `gorm:"type:text;not null"`                     // Parent intermediate key UUID
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
	// TODO: Add RotatedAt *int64 after fixing migration 0004 discovery issue
}

// TableName specifies the database table name for barrier content keys.
func (BarrierContentKey) TableName() string {
	return "barrier_content_keys"
}
