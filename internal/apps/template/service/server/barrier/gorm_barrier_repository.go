// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"context"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrNoRootKeyFound indicates no root keys exist in the database.
	ErrNoRootKeyFound = errors.New("no root key found")
	// ErrNoIntermediateKeyFound indicates no intermediate keys exist in the database.
	ErrNoIntermediateKeyFound = errors.New("no intermediate key found")
)

// GormRepository implements Repository using gorm.DB.
// This adapter allows barrier encryption to work with any service using gorm.DB
// (cipher-im, future services) without depending on KMS-specific OrmRepository.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new gorm.DB-based barrier repository.
func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db must be non-nil")
	}

	return &GormRepository{db: db}, nil
}

// WithTransaction executes the provided function within a database transaction.
func (r *GormRepository) WithTransaction(ctx context.Context, function func(tx Transaction) error) error {
	err := r.db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
		tx := &GormTransaction{gormDB: gormTx}

		return function(tx)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// Shutdown releases any resources held by the repository.
func (r *GormRepository) Shutdown() {
	// No resources to release for gorm.DB adapter
}

// GormTransaction implements Transaction using gorm.DB transaction.
type GormTransaction struct {
	gormDB *gorm.DB
}

// Context returns the transaction context.
func (tx *GormTransaction) Context() context.Context {
	return tx.gormDB.Statement.Context
}

// GetRootKeyLatest retrieves the most recently created root key.
func (tx *GormTransaction) GetRootKeyLatest() (*RootKey, error) {
	var key RootKey

	err := tx.gormDB.Order("uuid DESC").First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoRootKeyFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get latest root key: %w", err)
	}

	return &key, nil
}

// GetRootKey retrieves a specific root key by UUID.
func (tx *GormTransaction) GetRootKey(uuid *googleUuid.UUID) (*RootKey, error) {
	if uuid == nil {
		return nil, fmt.Errorf("uuid must be non-nil")
	}

	var key RootKey

	err := tx.gormDB.Where("uuid = ?", uuid.String()).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get root key %s: %w", uuid, err)
	}

	return &key, nil
}

// AddRootKey persists a new root key to storage.
func (tx *GormTransaction) AddRootKey(key *RootKey) error {
	if key == nil {
		return fmt.Errorf("key must be non-nil")
	}

	if err := tx.gormDB.Create(key).Error; err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}

	return nil
}

// GetIntermediateKeyLatest retrieves the most recently created intermediate key.
func (tx *GormTransaction) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	var key IntermediateKey

	err := tx.gormDB.Order("uuid DESC").First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoIntermediateKeyFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get latest intermediate key: %w", err)
	}

	return &key, nil
}

// GetIntermediateKey retrieves a specific intermediate key by UUID.
func (tx *GormTransaction) GetIntermediateKey(uuid *googleUuid.UUID) (*IntermediateKey, error) {
	if uuid == nil {
		return nil, fmt.Errorf("uuid must be non-nil")
	}

	var key IntermediateKey

	err := tx.gormDB.Where("uuid = ?", uuid.String()).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get intermediate key %s: %w", uuid, err)
	}

	return &key, nil
}

// AddIntermediateKey persists a new intermediate key to storage.
func (tx *GormTransaction) AddIntermediateKey(key *IntermediateKey) error {
	if key == nil {
		return fmt.Errorf("key must be non-nil")
	}

	if err := tx.gormDB.Create(key).Error; err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}

	return nil
}

// GetContentKey retrieves a specific content key by UUID.
func (tx *GormTransaction) GetContentKey(uuid *googleUuid.UUID) (*ContentKey, error) {
	if uuid == nil {
		return nil, fmt.Errorf("uuid must be non-nil")
	}

	var key ContentKey

	err := tx.gormDB.Where("uuid = ?", uuid.String()).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get content key %s: %w", uuid, err)
	}

	return &key, nil
}

// AddContentKey persists a new content key to storage.
func (tx *GormTransaction) AddContentKey(key *ContentKey) error {
	if key == nil {
		return fmt.Errorf("key must be non-nil")
	}

	if err := tx.gormDB.Create(key).Error; err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}

	return nil
}
