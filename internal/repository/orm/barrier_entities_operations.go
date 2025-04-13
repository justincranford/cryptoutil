package orm

import (
	"fmt"

	googleUuid "github.com/google/uuid"
)

// Root KEKs

func (r *OrmRepository) AddRootKey(rootKey *BarrierRootKey) error {
	if err := r.gormDB.Create(rootKey).Error; err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}
	return nil
}

func (r *OrmRepository) GetRootKeys() ([]BarrierRootKey, error) {
	var rootKeys []BarrierRootKey
	if err := r.gormDB.Order("uuid DESC").Find(&rootKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load root keys: %w", err)
	}
	return rootKeys, nil
}

func (r *OrmRepository) GetRootKeyLatest() (*BarrierRootKey, error) {
	var rootKey BarrierRootKey
	if err := r.gormDB.Order("uuid DESC").First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest root key: %w", err)
	}
	return &rootKey, nil
}

func (r *OrmRepository) GetRootKey(uuid googleUuid.UUID) (*BarrierRootKey, error) {
	var rootKey BarrierRootKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

func (r *OrmRepository) DeleteRootKey(uuid googleUuid.UUID) (*BarrierRootKey, error) {
	var rootKey BarrierRootKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete root key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

// Intermediate Keys

func (r *OrmRepository) AddIntermediateKey(intermediateKey *BarrierIntermediateKey) error {
	if err := r.gormDB.Create(intermediateKey).Error; err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}
	return nil
}

func (r *OrmRepository) GetIntermediateKeys() ([]BarrierIntermediateKey, error) {
	var intermediateKeys []BarrierIntermediateKey
	if err := r.gormDB.Order("uuid DESC").Find(&intermediateKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load intermediate keys: %w", err)
	}
	return intermediateKeys, nil
}

func (r *OrmRepository) GetIntermediateKeyLatest() (*BarrierIntermediateKey, error) {
	var intermediateKey BarrierIntermediateKey
	if err := r.gormDB.Order("uuid DESC").First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest intermediate key: %w", err)
	}
	return &intermediateKey, nil
}

func (r *OrmRepository) GetIntermediateKey(uuid googleUuid.UUID) (*BarrierIntermediateKey, error) {
	var intermediateKey BarrierIntermediateKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

func (r *OrmRepository) DeleteIntermediateKey(uuid googleUuid.UUID) (*BarrierIntermediateKey, error) {
	var intermediateKey BarrierIntermediateKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete intermediate key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

// Leaf Keys

func (r *OrmRepository) AddContentKey(contentKey *BarrierContentKey) error {
	if err := r.gormDB.Create(contentKey).Error; err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}
	return nil
}

func (r *OrmRepository) GetContentKeys() ([]BarrierContentKey, error) {
	var contentKeys []BarrierContentKey
	if err := r.gormDB.Order("uuid DESC").Find(&contentKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load content keys: %w", err)
	}
	return contentKeys, nil
}

func (r *OrmRepository) GetContentKeyLatest() (*BarrierContentKey, error) {
	var contentKey BarrierContentKey
	if err := r.gormDB.Order("uuid DESC").First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest content key: %w", err)
	}
	return &contentKey, nil
}

func (r *OrmRepository) GetContentKey(uuid googleUuid.UUID) (*BarrierContentKey, error) {
	var contentKey BarrierContentKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &contentKey, nil
}

func (r *OrmRepository) DeleteContentKey(uuid googleUuid.UUID) (*BarrierContentKey, error) {
	var contentKey BarrierContentKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete content key with UUID %s: %w", uuid, err)
	}
	return &contentKey, nil
}
