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

// OrmRepository implements Repository using KMS OrmRepository.
// This adapter allows existing KMS barrier encryption to work with the new
// Repository interface without breaking existing KMS code.
type OrmRepository struct {
	ormRepo *cryptoutilOrmRepository.OrmRepository
}

// NewOrmRepository creates a new OrmRepository-based barrier repository.
func NewOrmRepository(ormRepo *cryptoutilOrmRepository.OrmRepository) (*OrmRepository, error) {
	if ormRepo == nil {
		return nil, fmt.Errorf("ormRepo must be non-nil")
	}

	return &OrmRepository{ormRepo: ormRepo}, nil
}

// WithTransaction executes the provided function within a database transaction.
func (r *OrmRepository) WithTransaction(ctx context.Context, function func(tx Transaction) error) error {
	err := r.ormRepo.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(ormTx *cryptoutilOrmRepository.OrmTransaction) error {
		tx := &OrmTransaction{ormTx: ormTx}

		return function(tx)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// Shutdown releases any resources held by the repository.
func (r *OrmRepository) Shutdown() {
	r.ormRepo.Shutdown()
}

// OrmTransaction implements Transaction using KMS OrmTransaction.
type OrmTransaction struct {
	ormTx *cryptoutilOrmRepository.OrmTransaction
}

// Context returns the transaction context.
func (tx *OrmTransaction) Context() context.Context {
	return tx.ormTx.Context()
}

// GetRootKeyLatest retrieves the most recently created root key.
func (tx *OrmTransaction) GetRootKeyLatest() (*RootKey, error) {
	kmsKey, err := tx.ormTx.GetRootKeyLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest root key: %w", err)
	}

	if kmsKey == nil {
		return nil, ErrNoRootKeyFound
	}

	return &RootKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// GetRootKey retrieves a specific root key by UUID.
func (tx *OrmTransaction) GetRootKey(uuid *googleUuid.UUID) (*RootKey, error) {
	kmsKey, err := tx.ormTx.GetRootKey(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get root key: %w", err)
	}

	return &RootKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// AddRootKey persists a new root key to storage.
func (tx *OrmTransaction) AddRootKey(key *RootKey) error {
	// Convert template RootKey to KMS RootKey
	kmsKey := &cryptoutilOrmRepository.RootKey{
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
func (tx *OrmTransaction) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	kmsKey, err := tx.ormTx.GetIntermediateKeyLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest intermediate key: %w", err)
	}

	if kmsKey == nil {
		return nil, ErrNoIntermediateKeyFound
	}

	return &IntermediateKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// GetIntermediateKey retrieves a specific intermediate key by UUID.
func (tx *OrmTransaction) GetIntermediateKey(uuid *googleUuid.UUID) (*IntermediateKey, error) {
	kmsKey, err := tx.ormTx.GetIntermediateKey(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get intermediate key: %w", err)
	}

	return &IntermediateKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// AddIntermediateKey persists a new intermediate key to storage.
func (tx *OrmTransaction) AddIntermediateKey(key *IntermediateKey) error {
	// Convert template IntermediateKey to KMS IntermediateKey
	kmsKey := &cryptoutilOrmRepository.IntermediateKey{
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
func (tx *OrmTransaction) GetContentKey(uuid *googleUuid.UUID) (*ContentKey, error) {
	kmsKey, err := tx.ormTx.GetContentKey(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get content key: %w", err)
	}

	return &ContentKey{
		UUID:      kmsKey.UUID,
		Encrypted: kmsKey.Encrypted,
		KEKUUID:   kmsKey.KEKUUID,
	}, nil
}

// AddContentKey persists a new content key to storage.
func (tx *OrmTransaction) AddContentKey(key *ContentKey) error {
	// Convert template ContentKey to KMS ContentKey
	kmsKey := &cryptoutilOrmRepository.ContentKey{
		UUID:      key.UUID,
		Encrypted: key.Encrypted,
		KEKUUID:   key.KEKUUID,
	}

	if err := tx.ormTx.AddContentKey(kmsKey); err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}

	return nil
}
