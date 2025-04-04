package orm

import "fmt"

// UUIDv7 = timestamp (48-bits) + version (4-bits) + rand_a (12-bits) + var (2-bits) + rand_b (62-bits)
// JWK/JWKs = JWE wrapping of JWK/JWKs, stored as JSON (PostgreSQL JSONB, SQLite JSON)

// Root KEKs are unsealed by HSM, KMS, Shamir, etc. Rotation is infrequent.
type RootKek struct {
	UUID         string `gorm:"primaryKey;type:uuid"`
	EncryptedJWK string `gorm:"type:json;not null"`
}

// Namespace KEKs are wrapped by root KEKs. Rotation can be more frequent than Root KEKs.
type NamespaceKek struct {
	UUID         string `gorm:"primaryKey;type:uuid"`
	EncryptedJWK string `gorm:"type:json;not null"`
	RootKekUUID  string `gorm:"type:uuid;not null;foreignKey:RootKekUUID;references:UUID"`
}

// Namespace JWKSets are wrapped by namespace KEKs; each JWK can have unique use (e.g. enc, sig) + key_ops (e.g. sign, verify, encrypt, decrypt, wrapKey, unwrapKey, deriveKey, deriveBits). Rotation can be more frequent than namespace KEKs.
type NamespaceKeys struct {
	UUID             string `gorm:"primaryKey;type:uuid"`
	Name             string `gorm:"type:string;unique;not null" validate:"required,min=3,max=50"`
	EncryptedJWKSet  string `gorm:"type:json;not null"`
	NamespaceKekUUID string `gorm:"type:uuid;not null;foreignKey:NamespaceKekUUID;references:UUID"`
}

// Root KEKs

func (r *RepositoryProvider) GetRootKeks() ([]RootKek, error) {
	var rootKeks []RootKek
	if err := r.gormDB.Order("uuid DESC").Find(&rootKeks).Error; err != nil {
		return nil, fmt.Errorf("failed to get root KEKs: %w", err)
	}
	return rootKeks, nil
}

func (r *RepositoryProvider) GetRootKek() (*RootKek, error) {
	var rootKek RootKek
	if err := r.gormDB.Order("uuid DESC").First(&rootKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest root KEK: %w", err)
	}
	return &rootKek, nil
}

func (r *RepositoryProvider) GetRootKekVersioned(uuid string) (*RootKek, error) {
	var rootKek RootKek
	if err := r.gormDB.Where("uuid=?", uuid).First(&rootKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get versioned root KEK with UUID %s: %w", uuid, err)
	}
	return &rootKek, nil
}

// Namespace KEKs

func (r *RepositoryProvider) GetNamespaceKeks(rootKekUUID string) ([]NamespaceKek, error) {
	var namespaceKeks []NamespaceKek
	if err := r.gormDB.Where("root_kek_uuid=?", rootKekUUID).Order("uuid DESC").Find(&namespaceKeks).Error; err != nil {
		return nil, fmt.Errorf("failed to get namespace KEKs: %w", err)
	}
	return namespaceKeks, nil
}

func (r *RepositoryProvider) GetNamespaceKek(rootKekUUID string) (*NamespaceKek, error) {
	var namespaceKek NamespaceKek
	if err := r.gormDB.Where("root_kek_uuid=?", rootKekUUID).Order("uuid DESC").First(&namespaceKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest namespace KEK: %w", err)
	}
	return &namespaceKek, nil
}

func (r *RepositoryProvider) GetNamespaceKekVersion(uuid string) (*NamespaceKek, error) {
	var namespaceKek NamespaceKek
	if err := r.gormDB.Where("uuid=?", uuid).First(&namespaceKek).Error; err != nil {
		return nil, fmt.Errorf("failed to get versioned namespace KEK with UUID %s: %w", uuid, err)
	}
	return &namespaceKek, nil
}

// Namespace CEKs

func (r *RepositoryProvider) GetNamespaceCeks(namespaceKekUUID string) ([]NamespaceKeys, error) {
	var namespaceKeys []NamespaceKeys
	if err := r.gormDB.Where("namespace_kek_uuid=?", namespaceKekUUID).Order("uuid DESC").Find(&namespaceKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to get namespace CEKs: %w", err)
	}
	return namespaceKeys, nil
}

func (r *RepositoryProvider) GetNamespaceCek(namespaceKekUUID string) (*NamespaceKeys, error) {
	var namespaceKey NamespaceKeys
	if err := r.gormDB.Where("namespace_kek_uuid=?", namespaceKekUUID).Order("uuid DESC").First(&namespaceKey).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest namespace CEK: %w", err)
	}
	return &namespaceKey, nil
}

func (r *RepositoryProvider) GetNamespaceCekVersion(uuid string) (*NamespaceKeys, error) {
	var namespaceKey NamespaceKeys
	if err := r.gormDB.Where("uuid=?", uuid).First(&namespaceKey).Error; err != nil {
		return nil, fmt.Errorf("failed to get versioned namespace CEK with UUID %s: %w", uuid, err)
	}
	return &namespaceKey, nil
}
