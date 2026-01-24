// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"fmt"

	googleUuid "github.com/google/uuid"
)

// Root KEKs

// AddRootKey adds a new root key to the database.
func (tx *OrmTransaction) AddRootKey(rootKey *RootKey) error {
	if err := tx.state.gormTx.Create(rootKey).Error; err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}

	return nil
}

// GetRootKeys retrieves all root keys from the database.
func (tx *OrmTransaction) GetRootKeys() ([]RootKey, error) {
	var rootKeys []RootKey
	if err := tx.state.gormTx.Order("uuid DESC").Find(&rootKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load root keys: %w", err)
	}

	return rootKeys, nil
}

// GetRootKeyLatest retrieves the most recent root key from the database.
func (tx *OrmTransaction) GetRootKeyLatest() (*RootKey, error) {
	var rootKey RootKey
	if err := tx.state.gormTx.Order("uuid DESC").First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest root key: %w", err)
	}

	return &rootKey, nil
}

// GetRootKey retrieves a root key by UUID from the database.
func (tx *OrmTransaction) GetRootKey(uuid *googleUuid.UUID) (*RootKey, error) {
	var rootKey RootKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}

	return &rootKey, nil
}

// DeleteRootKey deletes a root key by UUID from the database.
func (tx *OrmTransaction) DeleteRootKey(uuid *googleUuid.UUID) (*RootKey, error) {
	var rootKey RootKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).Delete(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete root key with UUID %s: %w", uuid, err)
	}

	return &rootKey, nil
}

// Intermediate Keys

// AddIntermediateKey adds a new intermediate key to the database.
func (tx *OrmTransaction) AddIntermediateKey(intermediateKey *IntermediateKey) error {
	if err := tx.state.gormTx.Create(intermediateKey).Error; err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}

	return nil
}

// GetIntermediateKeys retrieves all intermediate keys from the database.
func (tx *OrmTransaction) GetIntermediateKeys() ([]IntermediateKey, error) {
	var intermediateKeys []IntermediateKey
	if err := tx.state.gormTx.Order("uuid DESC").Find(&intermediateKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load intermediate keys: %w", err)
	}

	return intermediateKeys, nil
}

// GetIntermediateKeyLatest retrieves the most recent intermediate key from the database.
func (tx *OrmTransaction) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := tx.state.gormTx.Order("uuid DESC").First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest intermediate key: %w", err)
	}

	return &intermediateKey, nil
}

// GetIntermediateKey retrieves an intermediate key by UUID from the database.
func (tx *OrmTransaction) GetIntermediateKey(uuid *googleUuid.UUID) (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}

	return &intermediateKey, nil
}

// DeleteIntermediateKey deletes an intermediate key by UUID from the database.
func (tx *OrmTransaction) DeleteIntermediateKey(uuid *googleUuid.UUID) (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).Delete(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete intermediate key with UUID %s: %w", uuid, err)
	}

	return &intermediateKey, nil
}

// Leaf Keys

// AddContentKey adds a new content key to the database.
func (tx *OrmTransaction) AddContentKey(contentKey *ContentKey) error {
	if err := tx.state.gormTx.Create(contentKey).Error; err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}

	return nil
}

// GetContentKeys retrieves all content keys from the database.
func (tx *OrmTransaction) GetContentKeys() ([]ContentKey, error) {
	var contentKeys []ContentKey
	if err := tx.state.gormTx.Order("uuid DESC").Find(&contentKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load content keys: %w", err)
	}

	return contentKeys, nil
}

// GetContentKeyLatest retrieves the most recent content key from the database.
func (tx *OrmTransaction) GetContentKeyLatest() (*ContentKey, error) {
	var contentKey ContentKey
	if err := tx.state.gormTx.Order("uuid DESC").First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest content key: %w", err)
	}

	return &contentKey, nil
}

// GetContentKey retrieves a content key by UUID from the database.
func (tx *OrmTransaction) GetContentKey(uuid *googleUuid.UUID) (*ContentKey, error) {
	var contentKey ContentKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}

	return &contentKey, nil
}

// DeleteContentKey deletes a content key by UUID from the database.
func (tx *OrmTransaction) DeleteContentKey(uuid *googleUuid.UUID) (*ContentKey, error) {
	var contentKey ContentKey
	if err := tx.state.gormTx.Where("uuid=?", uuid).Delete(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete content key with UUID %s: %w", uuid, err)
	}

	return &contentKey, nil
}
