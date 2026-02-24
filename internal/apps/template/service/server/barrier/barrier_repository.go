// Copyright (c) 2025 Justin Cranford
//
//

// Package barrier provides hierarchical key management for root, intermediate, and content encryption keys.
package barrier

import (
	"context"

	googleUuid "github.com/google/uuid"
)

// Repository defines the interface for barrier key storage operations.
// This abstraction allows Service to work with different database implementations
// (KMS OrmRepository, sm-im gorm.DB, etc.) without coupling to specific repository types.
type Repository interface {
	// WithTransaction executes the provided function within a database transaction.
	// The transaction will be automatically committed on success or rolled back on error.
	WithTransaction(ctx context.Context, function func(tx Transaction) error) error

	// Shutdown releases any resources held by the repository.
	Shutdown()
}

// Transaction defines the interface for transactional barrier key operations.
// Implementations must provide ACID guarantees for barrier key lifecycle operations.
type Transaction interface {
	// Context returns the transaction context.
	Context() context.Context

	// Root Key Operations

	// GetRootKeyLatest retrieves the most recently created root key.
	// Returns (nil, nil) if no root keys exist.
	GetRootKeyLatest() (*RootKey, error)

	// GetRootKey retrieves a specific root key by UUID.
	// Returns error if key not found.
	GetRootKey(uuid *googleUuid.UUID) (*RootKey, error)

	// AddRootKey persists a new root key to storage.
	AddRootKey(key *RootKey) error

	// Intermediate Key Operations

	// GetIntermediateKeyLatest retrieves the most recently created intermediate key.
	// Returns (nil, nil) if no intermediate keys exist.
	GetIntermediateKeyLatest() (*IntermediateKey, error)

	// GetIntermediateKey retrieves a specific intermediate key by UUID.
	// Returns error if key not found.
	GetIntermediateKey(uuid *googleUuid.UUID) (*IntermediateKey, error)

	// AddIntermediateKey persists a new intermediate key to storage.
	AddIntermediateKey(key *IntermediateKey) error

	// Content Key Operations

	// GetContentKey retrieves a specific content key by UUID.
	// Returns error if key not found.
	GetContentKey(uuid *googleUuid.UUID) (*ContentKey, error)

	// AddContentKey persists a new content key to storage.
	AddContentKey(key *ContentKey) error
}

// RootKey represents a root-level encryption key in the barrier hierarchy.
// Root keys are encrypted by the unseal key (HSM/KMS/Shamir).
type RootKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted root key
	KEKUUID   googleUuid.UUID `gorm:"type:text"`                              // KEK UUID (nil for root keys)
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
	// TODO: Add RotatedAt *int64 after fixing migration 0004 discovery issue
}

// TableName specifies the database table name for barrier root keys.
func (RootKey) TableName() string {
	return "barrier_root_keys"
}

// IntermediateKey represents an intermediate-level encryption key in the barrier hierarchy.
// Intermediate keys are encrypted by root keys.
type IntermediateKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted intermediate key
	KEKUUID   googleUuid.UUID `gorm:"type:text;not null"`                     // Parent root key UUID
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
	// TODO: Add RotatedAt *int64 after fixing migration 0004 discovery issue
}

// TableName specifies the database table name for barrier intermediate keys.
func (IntermediateKey) TableName() string {
	return "barrier_intermediate_keys"
}

// ContentKey represents a content-level encryption key in the barrier hierarchy.
// Content keys are encrypted by intermediate keys and used for actual data encryption.
type ContentKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted content key
	KEKUUID   googleUuid.UUID `gorm:"type:text;not null"`                     // Parent intermediate key UUID
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
	// TODO: Add RotatedAt *int64 after fixing migration 0004 discovery issue
}

// TableName specifies the database table name for barrier content keys.
func (ContentKey) TableName() string {
	return "barrier_content_keys"
}
