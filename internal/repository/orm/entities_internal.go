package orm

import (
	"fmt"

	googleUuid "github.com/google/uuid"
)

// UUIDv7 = timestamp (48-bits) + version (4-bits) + rand_a (12-bits) + var (2-bits) + rand_b (62-bits)
// JWK/JWKs = JWE wrapping of JWK/JWKs, stored as JSON (PostgreSQL JSONB, SQLite JSON)

// Root Keys are unsealed by HSM, KMS, Shamir, etc. Rotation is infrequent.
type RootKey struct {
	UUID          googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Serialized    string          `gorm:"type:json;not null"`
	UnsealKeyUUID googleUuid.UUID `gorm:"type:uuid;not null"`
}

// Intermediate Keys are wrapped by root Keys. Rotation can be more frequent than Root Keys.
type IntermediateKey struct {
	UUID        googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Serialized  string          `gorm:"type:json;not null"`
	RootKeyUUID googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:RootKeyUUID;references:UUID"`
}

// Leaf Keys are wrapped by Intermediate Keys.
type LeafKey struct {
	UUID                googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Name                string          `gorm:"type:string;unique;not null" validate:"required,min=3,max=50"`
	Serialized          string          `gorm:"type:json;not null"`
	IntermediateKeyUUID googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:IntermediateKeyUUID;references:UUID"`
}

// BarrierKey is an interface for all Keys.
type BarrierKey interface {
	GetUUID() googleUuid.UUID
	SetUUID(googleUuid.UUID)
	GetSerialized() string
	SetSerialized(string)
	GetParentUUID() googleUuid.UUID
	SetParentUUID(googleUuid.UUID)
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
func (r *RootKey) GetParentUUID() googleUuid.UUID {
	return r.UnsealKeyUUID
}
func (r *RootKey) SetParentUUID(parentUUID googleUuid.UUID) {
	r.UnsealKeyUUID = parentUUID
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
func (r *IntermediateKey) GetParentUUID() googleUuid.UUID {
	return r.RootKeyUUID
}
func (r *IntermediateKey) SetParentUUID(parentUUID googleUuid.UUID) {
	r.RootKeyUUID = parentUUID
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
func (r *LeafKey) GetParentUUID() googleUuid.UUID {
	return r.IntermediateKeyUUID
}
func (r *LeafKey) SetParentUUID(parentUUID googleUuid.UUID) {
	r.IntermediateKeyUUID = parentUUID
}

// Root KEKs

func (r *RepositoryProvider) AddRootKey(rootKek *RootKey) error {
	if err := r.gormDB.Create(rootKek).Error; err != nil {
		return fmt.Errorf("failed to add root KEK: %w", err)
	}
	return nil
}

func (r *RepositoryProvider) GetRootKeys() ([]RootKey, error) {
	var rootKeks []RootKey
	if err := r.gormDB.Order("uuid DESC").Find(&rootKeks).Error; err != nil {
		return nil, fmt.Errorf("failed to get root KEKs: %w", err)
	}
	return rootKeks, nil
}

func (r *RepositoryProvider) GetRootKeyLatest() (*RootKey, error) {
	var rootKek RootKey
	if err := r.gormDB.Order("uuid DESC").First(&rootKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest root KEK: %w", err)
	}
	return &rootKek, nil
}

func (r *RepositoryProvider) GetRootKeyVersioned(uuid googleUuid.UUID) (*RootKey, error) {
	var rootKek RootKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&rootKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get root KEK with UUID %s: %w", uuid, err)
	}
	return &rootKek, nil
}

// Namespace KEKs

func (r *RepositoryProvider) AddIntermediateKey(namespaceKek *IntermediateKey) error {
	if err := r.gormDB.Create(namespaceKek).Error; err != nil {
		return fmt.Errorf("failed to add namespace KEK: %w", err)
	}
	return nil
}

func (r *RepositoryProvider) GetIntermediateKeys(rootKekUUID googleUuid.UUID) ([]IntermediateKey, error) {
	var namespaceKeks []IntermediateKey
	if err := r.gormDB.Where("root_kek_uuid=?", rootKekUUID).Order("uuid DESC").Find(&namespaceKeks).Error; err != nil {
		return nil, fmt.Errorf("failed to get namespace KEKs: %w", err)
	}
	return namespaceKeks, nil
}

func (r *RepositoryProvider) GetIntermediateKeyLatest(rootKekUUID googleUuid.UUID) (*IntermediateKey, error) {
	var namespaceKek IntermediateKey
	if err := r.gormDB.Where("root_kek_uuid=?", rootKekUUID).Order("uuid DESC").First(&namespaceKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest namespace KEK: %w", err)
	}
	return &namespaceKek, nil
}

func (r *RepositoryProvider) GetIntermediateKeyVersion(uuid googleUuid.UUID) (*IntermediateKey, error) {
	var namespaceKek IntermediateKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&namespaceKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get namespace KEK with UUID %s: %w", uuid, err)
	}
	return &namespaceKek, nil
}

// Namespace CEKs

func (r *RepositoryProvider) AddLeafKey(namespaceKeys *LeafKey) error {
	if err := r.gormDB.Create(namespaceKeys).Error; err != nil {
		return fmt.Errorf("failed to add namespace keys: %w", err)
	}
	return nil
}

func (r *RepositoryProvider) GetNamespaceCeks(namespaceKekUUID string) ([]LeafKey, error) {
	var namespaceKeys []LeafKey
	if err := r.gormDB.Where("namespace_kek_uuid=?", namespaceKekUUID).Order("uuid DESC").Find(&namespaceKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to get namespace keys: %w", err)
	}
	return namespaceKeys, nil
}

func (r *RepositoryProvider) GetNamespaceCekLatest(namespaceKekUUID string) (*LeafKey, error) {
	var namespaceKey LeafKey
	if err := r.gormDB.Where("namespace_kek_uuid=?", namespaceKekUUID).Order("uuid DESC").First(&namespaceKey).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest namespace keys: %w", err)
	}
	return &namespaceKey, nil
}

func (r *RepositoryProvider) GetNamespaceCekVersion(uuid string) (*LeafKey, error) {
	var namespaceKey LeafKey
	if err := r.gormDB.Where("uuid=?", uuid).First(&namespaceKey).Error; err != nil {
		return nil, fmt.Errorf("failed to get namespace keys with UUID %s: %w", uuid, err)
	}
	return &namespaceKey, nil
}
