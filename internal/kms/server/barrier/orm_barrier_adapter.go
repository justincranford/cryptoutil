// Copyright (c) 2025 Justin Cranford
//
//

// Package barrier provides adapter types to allow KMS OrmRepository/OrmTransaction to implement
// the template barrier Repository/Transaction interfaces.
package barrier

import (
	"context"

	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilKmsServerRepositoryOrm "cryptoutil/internal/kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
)

// OrmRepositoryAdapter wraps KMS OrmRepository to implement template barrier.Repository.
type OrmRepositoryAdapter struct {
	ormRepo *cryptoutilKmsServerRepositoryOrm.OrmRepository
}

// NewOrmRepositoryAdapter creates a new adapter wrapping KMS OrmRepository.
func NewOrmRepositoryAdapter(ormRepo *cryptoutilKmsServerRepositoryOrm.OrmRepository) *OrmRepositoryAdapter {
	return &OrmRepositoryAdapter{ormRepo: ormRepo}
}

// WithTransaction implements barrier.Repository.WithTransaction.
// It wraps KMS transaction handling to provide template barrier Transaction interface.
func (a *OrmRepositoryAdapter) WithTransaction(ctx context.Context, function func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error) error {
	return a.ormRepo.WithTransaction(ctx, cryptoutilKmsServerRepositoryOrm.ReadWrite, func(ormTx *cryptoutilKmsServerRepositoryOrm.OrmTransaction) error {
		txAdapter := &OrmTransactionAdapter{ormTx: ormTx}
		return function(txAdapter)
	})
}

// Shutdown implements barrier.Repository.Shutdown.
func (a *OrmRepositoryAdapter) Shutdown() {
	a.ormRepo.Shutdown()
}

// OrmTransactionAdapter wraps KMS OrmTransaction to implement template barrier.Transaction.
type OrmTransactionAdapter struct {
	ormTx *cryptoutilKmsServerRepositoryOrm.OrmTransaction
}

// Context implements barrier.Transaction.Context.
func (a *OrmTransactionAdapter) Context() context.Context {
	return a.ormTx.Context()
}

// GetRootKeyLatest implements barrier.Transaction.GetRootKeyLatest.
func (a *OrmTransactionAdapter) GetRootKeyLatest() (*cryptoutilAppsTemplateServiceServerBarrier.RootKey, error) {
	ormKey, err := a.ormTx.GetRootKeyLatest()
	if err != nil {
		return nil, err
	}
	return convertOrmRootKeyToBarrier(ormKey), nil
}

// GetRootKey implements barrier.Transaction.GetRootKey.
func (a *OrmTransactionAdapter) GetRootKey(uuid *googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerBarrier.RootKey, error) {
	ormKey, err := a.ormTx.GetRootKey(uuid)
	if err != nil {
		return nil, err
	}
	return convertOrmRootKeyToBarrier(ormKey), nil
}

// AddRootKey implements barrier.Transaction.AddRootKey.
func (a *OrmTransactionAdapter) AddRootKey(key *cryptoutilAppsTemplateServiceServerBarrier.RootKey) error {
	ormKey := convertBarrierRootKeyToOrm(key)
	return a.ormTx.AddRootKey(ormKey)
}

// GetIntermediateKeyLatest implements barrier.Transaction.GetIntermediateKeyLatest.
func (a *OrmTransactionAdapter) GetIntermediateKeyLatest() (*cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey, error) {
	ormKey, err := a.ormTx.GetIntermediateKeyLatest()
	if err != nil {
		return nil, err
	}
	return convertOrmIntermediateKeyToBarrier(ormKey), nil
}

// GetIntermediateKey implements barrier.Transaction.GetIntermediateKey.
func (a *OrmTransactionAdapter) GetIntermediateKey(uuid *googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey, error) {
	ormKey, err := a.ormTx.GetIntermediateKey(uuid)
	if err != nil {
		return nil, err
	}
	return convertOrmIntermediateKeyToBarrier(ormKey), nil
}

// AddIntermediateKey implements barrier.Transaction.AddIntermediateKey.
func (a *OrmTransactionAdapter) AddIntermediateKey(key *cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey) error {
	ormKey := convertBarrierIntermediateKeyToOrm(key)
	return a.ormTx.AddIntermediateKey(ormKey)
}

// GetContentKey implements barrier.Transaction.GetContentKey.
func (a *OrmTransactionAdapter) GetContentKey(uuid *googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerBarrier.ContentKey, error) {
	ormKey, err := a.ormTx.GetContentKey(uuid)
	if err != nil {
		return nil, err
	}
	return convertOrmContentKeyToBarrier(ormKey), nil
}

// AddContentKey implements barrier.Transaction.AddContentKey.
func (a *OrmTransactionAdapter) AddContentKey(key *cryptoutilAppsTemplateServiceServerBarrier.ContentKey) error {
	ormKey := convertBarrierContentKeyToOrm(key)
	return a.ormTx.AddContentKey(ormKey)
}

// Type conversion helpers - types are structurally identical between packages.

func convertOrmRootKeyToBarrier(ormKey *cryptoutilKmsServerRepositoryOrm.RootKey) *cryptoutilAppsTemplateServiceServerBarrier.RootKey {
	if ormKey == nil {
		return nil
	}
	return &cryptoutilAppsTemplateServiceServerBarrier.RootKey{
		UUID:      ormKey.UUID,
		Encrypted: ormKey.Encrypted,
		KEKUUID:   ormKey.KEKUUID,
		CreatedAt: ormKey.CreatedAt,
		UpdatedAt: ormKey.UpdatedAt,
	}
}

func convertBarrierRootKeyToOrm(barrierKey *cryptoutilAppsTemplateServiceServerBarrier.RootKey) *cryptoutilKmsServerRepositoryOrm.RootKey {
	if barrierKey == nil {
		return nil
	}
	return &cryptoutilKmsServerRepositoryOrm.RootKey{
		UUID:      barrierKey.UUID,
		Encrypted: barrierKey.Encrypted,
		KEKUUID:   barrierKey.KEKUUID,
		CreatedAt: barrierKey.CreatedAt,
		UpdatedAt: barrierKey.UpdatedAt,
	}
}

func convertOrmIntermediateKeyToBarrier(ormKey *cryptoutilKmsServerRepositoryOrm.IntermediateKey) *cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey {
	if ormKey == nil {
		return nil
	}
	return &cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{
		UUID:      ormKey.UUID,
		Encrypted: ormKey.Encrypted,
		KEKUUID:   ormKey.KEKUUID,
		CreatedAt: ormKey.CreatedAt,
		UpdatedAt: ormKey.UpdatedAt,
	}
}

func convertBarrierIntermediateKeyToOrm(barrierKey *cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey) *cryptoutilKmsServerRepositoryOrm.IntermediateKey {
	if barrierKey == nil {
		return nil
	}
	return &cryptoutilKmsServerRepositoryOrm.IntermediateKey{
		UUID:      barrierKey.UUID,
		Encrypted: barrierKey.Encrypted,
		KEKUUID:   barrierKey.KEKUUID,
		CreatedAt: barrierKey.CreatedAt,
		UpdatedAt: barrierKey.UpdatedAt,
	}
}

func convertOrmContentKeyToBarrier(ormKey *cryptoutilKmsServerRepositoryOrm.ContentKey) *cryptoutilAppsTemplateServiceServerBarrier.ContentKey {
	if ormKey == nil {
		return nil
	}
	return &cryptoutilAppsTemplateServiceServerBarrier.ContentKey{
		UUID:      ormKey.UUID,
		Encrypted: ormKey.Encrypted,
		KEKUUID:   ormKey.KEKUUID,
		CreatedAt: ormKey.CreatedAt,
		UpdatedAt: ormKey.UpdatedAt,
	}
}

func convertBarrierContentKeyToOrm(barrierKey *cryptoutilAppsTemplateServiceServerBarrier.ContentKey) *cryptoutilKmsServerRepositoryOrm.ContentKey {
	if barrierKey == nil {
		return nil
	}
	return &cryptoutilKmsServerRepositoryOrm.ContentKey{
		UUID:      barrierKey.UUID,
		Encrypted: barrierKey.Encrypted,
		KEKUUID:   barrierKey.KEKUUID,
		CreatedAt: barrierKey.CreatedAt,
		UpdatedAt: barrierKey.UpdatedAt,
	}
}

// Compile-time interface assertions.
var (
	_ cryptoutilAppsTemplateServiceServerBarrier.Repository  = (*OrmRepositoryAdapter)(nil)
	_ cryptoutilAppsTemplateServiceServerBarrier.Transaction = (*OrmTransactionAdapter)(nil)
)
