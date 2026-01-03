// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"context"
	"fmt"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
)

// OrmBarrierRepository implements BarrierRepository using KMS OrmRepository.
// This adapter allows existing KMS barrier encryption to work with the new
// BarrierRepository interface without breaking existing KMS code.
type OrmBarrierRepository struct {
	ormRepo *cryptoutilOrmRepository.OrmRepository
}

// NewOrmBarrierRepository creates a new OrmRepository-based barrier repository.
func NewOrmBarrierRepository(ormRepo *cryptoutilOrmRepository.OrmRepository) (*OrmBarrierRepository, error) {
	if ormRepo == nil {
		return nil, fmt.Errorf("ormRepo must be non-nil")
	}

	return &OrmBarrierRepository{ormRepo: ormRepo}, nil
}

// WithTransaction executes the provided function within a database transaction.
func (r *OrmBarrierRepository) WithTransaction(ctx context.Context, function func(tx BarrierTransaction) error) error {
	err := r.ormRepo.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(ormTx *cryptoutilOrmRepository.OrmTransaction) error {
		tx := &OrmBarrierTransaction{ormTx: ormTx}

		return function(tx)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// Shutdown releases any resources held by the repository.
func (r *OrmBarrierRepository) Shutdown() {
	r.ormRepo.Shutdown()
}

// OrmBarrierTransaction implements BarrierTransaction using KMS OrmTransaction.
type OrmBarrierTransaction struct {
	ormTx *cryptoutilOrmRepository.OrmTransaction
}

// Context returns the transaction context.
func (tx *OrmBarrierTransaction) Context() context.Context {
	return tx.ormTx.Context()
}

// GetRootKeyLatest retrieves the most recently created root key.
func (tx *OrmBarrierTransaction) GetRootKeyLatest() (*BarrierRootKey, error) {
	kmsKey, err := tx.ormTx.GetRootKeyLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest root key: %w", err)
	}

	if kmsKey == nil {
		return nil, ErrNoRootKeyFound
	}

	return &BarrierRootKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// GetRootKey retrieves a specific root key by UUID.
func (tx *OrmBarrierTransaction) GetRootKey(uuid *googleUuid.UUID) (*BarrierRootKey, error) {
	kmsKey, err := tx.ormTx.GetRootKey(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get root key: %w", err)
	}

	return &BarrierRootKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// AddRootKey persists a new root key to storage.
func (tx *OrmBarrierTransaction) AddRootKey(key *BarrierRootKey) error {
	// Convert template BarrierRootKey to KMS BarrierRootKey
	kmsKey := &cryptoutilOrmRepository.BarrierRootKey{
		UUID:      key.UUID,
		Encrypted: key.Encrypted,
		KEKUUID:   key.KEKUUID,
	}

	if err := tx.ormTx.AddRootKey(kmsKey); err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}

	return nil
}

// GetIntermediateKeyLatest retrieves the most recently created intermediate key.
func (tx *OrmBarrierTransaction) GetIntermediateKeyLatest() (*BarrierIntermediateKey, error) {
	kmsKey, err := tx.ormTx.GetIntermediateKeyLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest intermediate key: %w", err)
	}

	if kmsKey == nil {
		return nil, ErrNoIntermediateKeyFound
	}

	return &BarrierIntermediateKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// GetIntermediateKey retrieves a specific intermediate key by UUID.
func (tx *OrmBarrierTransaction) GetIntermediateKey(uuid *googleUuid.UUID) (*BarrierIntermediateKey, error) {
	kmsKey, err := tx.ormTx.GetIntermediateKey(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get intermediate key: %w", err)
	}

	return &BarrierIntermediateKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// AddIntermediateKey persists a new intermediate key to storage.
func (tx *OrmBarrierTransaction) AddIntermediateKey(key *BarrierIntermediateKey) error {
	// Convert template BarrierIntermediateKey to KMS BarrierIntermediateKey
	kmsKey := &cryptoutilOrmRepository.BarrierIntermediateKey{
		UUID:      key.UUID,
		Encrypted: key.Encrypted,
		KEKUUID:   key.KEKUUID,
	}

	if err := tx.ormTx.AddIntermediateKey(kmsKey); err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}

	return nil
}

// GetContentKey retrieves a specific content key by UUID.
func (tx *OrmBarrierTransaction) GetContentKey(uuid *googleUuid.UUID) (*BarrierContentKey, error) {
	kmsKey, err := tx.ormTx.GetContentKey(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get content key: %w", err)
	}

	return &BarrierContentKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// AddContentKey persists a new content key to storage.
func (tx *OrmBarrierTransaction) AddContentKey(key *BarrierContentKey) error {
	// Convert template BarrierContentKey to KMS BarrierContentKey
	kmsKey := &cryptoutilOrmRepository.BarrierContentKey{
		UUID:      key.UUID,
		Encrypted: key.Encrypted,
		KEKUUID:   key.KEKUUID,
	}

	if err := tx.ormTx.AddContentKey(kmsKey); err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}

	return nil
}
