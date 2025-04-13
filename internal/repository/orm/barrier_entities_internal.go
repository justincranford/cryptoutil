package orm

import (
	"fmt"

	googleUuid "github.com/google/uuid"
)

// UUIDv7 = timestamp (48-bits) + version (4-bits) + rand_a (12-bits) + var (2-bits) + rand_b (62-bits)
// JWK/JWKs = JWE wrapping of JWK/JWKs, stored as JSON (PostgreSQL JSONB, SQLite JSON)

// Root Keys are unsealed by HSM, KMS, Shamir Key Shares, etc. Rotation is posible but infrequent.
type RootKey struct {
	UUID      googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Encrypted string          `gorm:"type:json;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null"`
}

// Intermediate Keys are wrapped by root Keys. Rotation is encouraged and can be frequent.
type IntermediateKey struct {
	UUID      googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Encrypted string          `gorm:"type:json;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:RootKEKUUID;references:UUID"`
}

// Leaf Keys are wrapped by Intermediate Keys. Rotation is encouraged and can be very frequent.
type ContentKey struct {
	UUID googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	// Name       string          `gorm:"type:string;unique;not null" validate:"required,min=3,max=50"`
	Encrypted string          `gorm:"type:json;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:IntermediateKEKUUID;references:UUID"`
}

// BarrierKey is an interface for all Keys.
type BarrierKey interface {
	GetUUID() googleUuid.UUID
	SetUUID(googleUuid.UUID)
	GetEncrypted() string
	SetEncrypted(string)
	GetKEKUUID() googleUuid.UUID
	SetKEKUUID(googleUuid.UUID)
}

func (r *RootKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

func (r *RootKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}

func (r *RootKey) GetEncrypted() string {
	return r.Encrypted
}

func (r *RootKey) SetEncrypted(serialized string) {
	r.Encrypted = serialized
}

func (r *RootKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

func (r *RootKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}

func (r *IntermediateKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

func (r *IntermediateKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}

func (r *IntermediateKey) GetEncrypted() string {
	return r.Encrypted
}

func (r *IntermediateKey) SetEncrypted(serialized string) {
	r.Encrypted = serialized
}

func (r *IntermediateKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

func (r *IntermediateKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}

func (r *ContentKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

func (r *ContentKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}

func (r *ContentKey) GetEncrypted() string {
	return r.Encrypted
}

func (r *ContentKey) SetEncrypted(serialized string) {
	r.Encrypted = serialized
}

func (r *ContentKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

func (r *ContentKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}

// Root KEKs

func (r *OrmRepository) AddRootKey(rootKey *RootKey) error {
	if err := r.gormDB.Create(rootKey).Error; err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}
	return nil
}

func (r *OrmRepository) GetRootKeys() ([]RootKey, error) {
	var rootKeys []RootKey
	if err := r.gormDB.Order("uuid DESC").Find(&rootKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load root keys: %w", err)
	}
	return rootKeys, nil
}

func (r *OrmRepository) GetRootKeyLatest() (*RootKey, error) {
	var rootKey RootKey
	if err := r.gormDB.Order("uuid DESC").First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest root key: %w", err)
	}
	return &rootKey, nil
}

func (r *OrmRepository) GetRootKey(uuid googleUuid.UUID) (*RootKey, error) {
	var rootKey RootKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

func (r *OrmRepository) DeleteRootKey(uuid googleUuid.UUID) (*RootKey, error) {
	var rootKey RootKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete root key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

// Intermediate Keys

func (r *OrmRepository) AddIntermediateKey(intermediateKey *IntermediateKey) error {
	if err := r.gormDB.Create(intermediateKey).Error; err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}
	return nil
}

func (r *OrmRepository) GetIntermediateKeys() ([]IntermediateKey, error) {
	var intermediateKeys []IntermediateKey
	if err := r.gormDB.Order("uuid DESC").Find(&intermediateKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load intermediate keys: %w", err)
	}
	return intermediateKeys, nil
}

func (r *OrmRepository) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := r.gormDB.Order("uuid DESC").First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest intermediate key: %w", err)
	}
	return &intermediateKey, nil
}

func (r *OrmRepository) GetIntermediateKey(uuid googleUuid.UUID) (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

func (r *OrmRepository) DeleteIntermediateKey(uuid googleUuid.UUID) (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete intermediate key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

// Leaf Keys

func (r *OrmRepository) AddContentKey(contentKey *ContentKey) error {
	if err := r.gormDB.Create(contentKey).Error; err != nil {
		return fmt.Errorf("failed to add content key: %w", err)
	}
	return nil
}

func (r *OrmRepository) GetContentKeys() ([]ContentKey, error) {
	var contentKeys []ContentKey
	if err := r.gormDB.Order("uuid DESC").Find(&contentKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load content keys: %w", err)
	}
	return contentKeys, nil
}

func (r *OrmRepository) GetContentKeyLatest() (*ContentKey, error) {
	var contentKey ContentKey
	if err := r.gormDB.Order("uuid DESC").First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest content key: %w", err)
	}
	return &contentKey, nil
}

func (r *OrmRepository) GetContentKey(uuid googleUuid.UUID) (*ContentKey, error) {
	var contentKey ContentKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &contentKey, nil
}

func (r *OrmRepository) DeleteContentKey(uuid googleUuid.UUID) (*ContentKey, error) {
	var contentKey ContentKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&contentKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete content key with UUID %s: %w", uuid, err)
	}
	return &contentKey, nil
}
