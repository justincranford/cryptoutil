package orm

import (
	"fmt"

	googleUuid "github.com/google/uuid"
)

// Root KEKs

func (tx *OrmTransaction) AddRootKey(rootKey *BarrierRootKey) error {
	if err := tx.state.gormTx.Create(rootKey).Error; err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}
	return nil
}

func (tx *OrmTransaction) GetRootKeys() ([]BarrierRootKey, error) {
	var rootKeys []BarrierRootKey
	if err := tx.state.gormTx.Order("uuid DESC").Find(&rootKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load root keys: %w", err)
	}
	return rootKeys, nil
}

func (tx *OrmTransaction) GetRootKeyLatest() (*BarrierRootKey, error) {
	var rootKey BarrierRootKey
	if err := tx.state.gormTx.Order("uuid DESC").First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest root key: %w", err)
	}
	return &rootKey, nil
}

func (tx *OrmTransaction) GetRootKey(uuid *googleUuid.UUID) (*BarrierRootKey, error) {
	var rootKey BarrierRootKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

func (tx *OrmTransaction) DeleteRootKey(uuid *googleUuid.UUID) (*BarrierRootKey, error) {
	var rootKey BarrierRootKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).Delete(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete root key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

// Intermediate Keys

func (tx *OrmTransaction) AddIntermediateKey(intermediateKey *BarrierIntermediateKey) error {
	if err := tx.state.gormTx.Create(intermediateKey).Error; err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}
	return nil
}

func (tx *OrmTransaction) GetIntermediateKeys() ([]BarrierIntermediateKey, error) {
	var intermediateKeys []BarrierIntermediateKey
	if err := tx.state.gormTx.Order("uuid DESC").Find(&intermediateKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load intermediate keys: %w", err)
	}
	return intermediateKeys, nil
}

func (tx *OrmTransaction) GetIntermediateKeyLatest() (*BarrierIntermediateKey, error) {
	var intermediateKey BarrierIntermediateKey
	if err := tx.state.gormTx.Order("uuid DESC").First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest intermediate key: %w", err)
	}
	return &intermediateKey, nil
}

func (tx *OrmTransaction) GetIntermediateKey(uuid *googleUuid.UUID) (*BarrierIntermediateKey, error) {
	var intermediateKey BarrierIntermediateKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

func (tx *OrmTransaction) DeleteIntermediateKey(uuid *googleUuid.UUID) (*BarrierIntermediateKey, error) {
	var intermediateKey BarrierIntermediateKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).Delete(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete intermediate key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

// Leaf Keys

func (tx *OrmTransaction) AddContentKey(contentKey *BarrierContentKey) error {
	if err := tx.state.gormTx.Create(contentKey).Error; err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}
	return nil
}

func (tx *OrmTransaction) GetContentKeys() ([]BarrierContentKey, error) {
	var contentKeys []BarrierContentKey
	if err := tx.state.gormTx.Order("uuid DESC").Find(&contentKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load content keys: %w", err)
	}
	return contentKeys, nil
}

func (tx *OrmTransaction) GetContentKeyLatest() (*BarrierContentKey, error) {
	var contentKey BarrierContentKey
	if err := tx.state.gormTx.Order("uuid DESC").First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest content key: %w", err)
	}
	return &contentKey, nil
}

func (tx *OrmTransaction) GetContentKey(uuid *googleUuid.UUID) (*BarrierContentKey, error) {
	var contentKey BarrierContentKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &contentKey, nil
}

func (tx *OrmTransaction) DeleteContentKey(uuid *googleUuid.UUID) (*BarrierContentKey, error) {
	var contentKey BarrierContentKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).Delete(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete content key with UUID %s: %w", uuid, err)
	}
	return &contentKey, nil
}
