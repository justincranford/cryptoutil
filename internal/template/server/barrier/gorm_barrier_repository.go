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

// GormBarrierRepository implements BarrierRepository using gorm.DB.
// This adapter allows barrier encryption to work with any service using gorm.DB
// (cipher-im, future services) without depending on KMS-specific OrmRepository.
type GormBarrierRepository struct {
	db *gorm.DB
}

// NewGormBarrierRepository creates a new gorm.DB-based barrier repository.
func NewGormBarrierRepository(db *gorm.DB) (*GormBarrierRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db must be non-nil")
	}

	return &GormBarrierRepository{db: db}, nil
}

// WithTransaction executes the provided function within a database transaction.
func (r *GormBarrierRepository) WithTransaction(ctx context.Context, function func(tx BarrierTransaction) error) error {
	err := r.db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
		tx := &GormBarrierTransaction{gormDB: gormTx}

		return function(tx)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// Shutdown releases any resources held by the repository.
func (r *GormBarrierRepository) Shutdown() {
	// No resources to release for gorm.DB adapter
}

// GormBarrierTransaction implements BarrierTransaction using gorm.DB transaction.
type GormBarrierTransaction struct {
	gormDB *gorm.DB
}

// Context returns the transaction context.
func (tx *GormBarrierTransaction) Context() context.Context {
	return tx.gormDB.Statement.Context
}

// GetRootKeyLatest retrieves the most recently created root key.
func (tx *GormBarrierTransaction) GetRootKeyLatest() (*BarrierRootKey, error) {
	var key BarrierRootKey

	err := tx.gormDB.Order("created_at DESC").First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoRootKeyFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get latest root key: %w", err)
	}

	return &key, nil
}

// GetRootKey retrieves a specific root key by UUID.
func (tx *GormBarrierTransaction) GetRootKey(uuid *googleUuid.UUID) (*BarrierRootKey, error) {
	if uuid == nil {
		return nil, fmt.Errorf("uuid must be non-nil")
	}

	var key BarrierRootKey

	err := tx.gormDB.Where("uuid = ?", uuid.String()).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get root key %s: %w", uuid, err)
	}

	return &key, nil
}

// AddRootKey persists a new root key to storage.
func (tx *GormBarrierTransaction) AddRootKey(key *BarrierRootKey) error {
	if key == nil {
		return fmt.Errorf("key must be non-nil")
	}

	if err := tx.gormDB.Create(key).Error; err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}

	return nil
}

// GetIntermediateKeyLatest retrieves the most recently created intermediate key.
func (tx *GormBarrierTransaction) GetIntermediateKeyLatest() (*BarrierIntermediateKey, error) {
	var key BarrierIntermediateKey

	err := tx.gormDB.Order("created_at DESC").First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoIntermediateKeyFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get latest intermediate key: %w", err)
	}

	return &key, nil
}

// GetIntermediateKey retrieves a specific intermediate key by UUID.
func (tx *GormBarrierTransaction) GetIntermediateKey(uuid *googleUuid.UUID) (*BarrierIntermediateKey, error) {
	if uuid == nil {
		return nil, fmt.Errorf("uuid must be non-nil")
	}

	var key BarrierIntermediateKey

	err := tx.gormDB.Where("uuid = ?", uuid.String()).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get intermediate key %s: %w", uuid, err)
	}

	return &key, nil
}

// AddIntermediateKey persists a new intermediate key to storage.
func (tx *GormBarrierTransaction) AddIntermediateKey(key *BarrierIntermediateKey) error {
	if key == nil {
		return fmt.Errorf("key must be non-nil")
	}

	if err := tx.gormDB.Create(key).Error; err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}

	return nil
}

// GetContentKey retrieves a specific content key by UUID.
func (tx *GormBarrierTransaction) GetContentKey(uuid *googleUuid.UUID) (*BarrierContentKey, error) {
	if uuid == nil {
		return nil, fmt.Errorf("uuid must be non-nil")
	}

	var key BarrierContentKey

	err := tx.gormDB.Where("uuid = ?", uuid.String()).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get content key %s: %w", uuid, err)
	}

	return &key, nil
}

// AddContentKey persists a new content key to storage.
func (tx *GormBarrierTransaction) AddContentKey(key *BarrierContentKey) error {
	if key == nil {
		return fmt.Errorf("key must be non-nil")
	}

	if err := tx.gormDB.Create(key).Error; err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}

	return nil
}
