package orm

import (
	"fmt"

	googleUuid "github.com/google/uuid"
)

// UUIDv7 = timestamp (48-bits) + version (4-bits) + rand_a (12-bits) + var (2-bits) + rand_b (62-bits)
// JWK/JWKs = JWE wrapping of JWK/JWKs, stored as JSON (PostgreSQL JSONB, SQLite JSON)

// Root Keys are unsealed by HSM, KMS, Shamir, etc. Rotation is infrequent.
type RootKey struct {
	UUID       googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Serialized string          `gorm:"type:json;not null"`
	KEKUUID    googleUuid.UUID `gorm:"type:uuid;not null"`
}

// Intermediate Keys are wrapped by root Keys. Rotation can be more frequent than Root Keys.
type IntermediateKey struct {
	UUID       googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Serialized string          `gorm:"type:json;not null"`
	KEKUUID    googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:RootKEKUUID;references:UUID"`
}

// Leaf Keys are wrapped by Intermediate Keys.
type LeafKey struct {
	UUID googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	// Name       string          `gorm:"type:string;unique;not null" validate:"required,min=3,max=50"`
	Serialized string          `gorm:"type:json;not null"`
	KEKUUID    googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:IntermediateKEKUUID;references:UUID"`
}

// BarrierKey is an interface for all Keys.
type BarrierKey interface {
	GetUUID() googleUuid.UUID
	SetUUID(googleUuid.UUID)
	GetSerialized() string
	SetSerialized(string)
	GetKEKUUID() googleUuid.UUID
	SetKEKUUID(googleUuid.UUID)
}

func (r *RootKey) GetUUID() googleUuid.UUID {
	return r.UUID
}
func (r *RootKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}
func (r *RootKey) GetSerialized() string {
	return r.Serialized
}
func (r *RootKey) SetSerialized(serialized string) {
	r.Serialized = serialized
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
func (r *IntermediateKey) GetSerialized() string {
	return r.Serialized
}
func (r *IntermediateKey) SetSerialized(serialized string) {
	r.Serialized = serialized
}
func (r *IntermediateKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}
func (r *IntermediateKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}

func (r *LeafKey) GetUUID() googleUuid.UUID {
	return r.UUID
}
func (r *LeafKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}
func (r *LeafKey) GetSerialized() string {
	return r.Serialized
}
func (r *LeafKey) SetSerialized(serialized string) {
	r.Serialized = serialized
}
func (r *LeafKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}
func (r *LeafKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}

// Root KEKs

func (r *RepositoryProvider) AddRootKey(rootKey *RootKey) error {
	if err := r.gormDB.Create(rootKey).Error; err != nil {
		return fmt.Errorf("failed to add root key: %w", err)
	}
	return nil
}

func (r *RepositoryProvider) GetRootKeys() ([]RootKey, error) {
	var rootKeys []RootKey
	if err := r.gormDB.Order("uuid DESC").Find(&rootKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load root keys: %w", err)
	}
	return rootKeys, nil
}

func (r *RepositoryProvider) GetRootKeyLatest() (*RootKey, error) {
	var rootKey RootKey
	if err := r.gormDB.Order("uuid DESC").First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest root key: %w", err)
	}
	return &rootKey, nil
}

func (r *RepositoryProvider) GetRootKey(uuid googleUuid.UUID) (*RootKey, error) {
	var rootKey RootKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

func (r *RepositoryProvider) DeleteRootKey(uuid googleUuid.UUID) (*RootKey, error) {
	var rootKey RootKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&rootKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete root key with UUID %s: %w", uuid, err)
	}
	return &rootKey, nil
}

// Intermediate Keys

func (r *RepositoryProvider) AddIntermediateKey(intermediateKey *IntermediateKey) error {
	if err := r.gormDB.Create(intermediateKey).Error; err != nil {
		return fmt.Errorf("failed to add intermediate key: %w", err)
	}
	return nil
}

func (r *RepositoryProvider) GetIntermediateKeys() ([]IntermediateKey, error) {
	var intermediateKeys []IntermediateKey
	if err := r.gormDB.Order("uuid DESC").Find(&intermediateKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load intermediate keys: %w", err)
	}
	return intermediateKeys, nil
}

func (r *RepositoryProvider) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := r.gormDB.Order("uuid DESC").First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest intermediate key: %w", err)
	}
	return &intermediateKey, nil
}

func (r *RepositoryProvider) GetIntermediateKey(uuid googleUuid.UUID) (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

func (r *RepositoryProvider) DeleteIntermediateKey(uuid googleUuid.UUID) (*IntermediateKey, error) {
	var intermediateKey IntermediateKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&intermediateKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete intermediate key with UUID %s: %w", uuid, err)
	}
	return &intermediateKey, nil
}

// Leaf Keys

func (r *RepositoryProvider) AddLeafKey(leafKey *LeafKey) error {
	if err := r.gormDB.Create(leafKey).Error; err != nil {
		return fmt.Errorf("failed to add leaf key: %w", err)
	}
	return nil
}

func (r *RepositoryProvider) GetLeafKeys() ([]LeafKey, error) {
	var leafKeys []LeafKey
	if err := r.gormDB.Order("uuid DESC").Find(&leafKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to load leaf keys: %w", err)
	}
	return leafKeys, nil
}

func (r *RepositoryProvider) GetLeafKeyLatest() (*LeafKey, error) {
	var leafKey LeafKey
	if err := r.gormDB.Order("uuid DESC").First(&leafKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load latest leaf key: %w", err)
	}
	return &leafKey, nil
}

func (r *RepositoryProvider) GetLeafKey(uuid googleUuid.UUID) (*LeafKey, error) {
	var leafKey LeafKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&leafKey).Error; err != nil {
		return nil, fmt.Errorf("failed to load key key with UUID %s: %w", uuid, err)
	}
	return &leafKey, nil
}

func (r *RepositoryProvider) DeleteLeafKey(uuid googleUuid.UUID) (*LeafKey, error) {
	var leafKey LeafKey
	if err := r.gormDB.Where("uuid=?", uuid).Delete(&leafKey).Error; err != nil {
		return nil, fmt.Errorf("failed to delete leaf key with UUID %s: %w", uuid, err)
	}
	return &leafKey, nil
}
