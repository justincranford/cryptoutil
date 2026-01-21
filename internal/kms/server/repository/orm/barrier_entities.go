// Package orm provides database entity definitions and ORM operations for the KMS server.
package orm

import (
	googleUuid "github.com/google/uuid"
)

// UUIDv7 = timestamp (48-bits) + version (4-bits) + rand_a (12-bits) + var (2-bits) + rand_b (62-bits)
// JWK/JWKs = JWE wrapping of JWK/JWKs, stored as JSON (PostgreSQL JSONB, SQLite JSON)

// BarrierRootKey represents root keys that are unsealed by HSM, KMS, Shamir Key Shares, etc. Rotation is possible but infrequent.
type BarrierRootKey struct {
	UUID      googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null"`
}

// BarrierIntermediateKey represents intermediate keys that are wrapped by root keys. Rotation is encouraged and can be frequent.
type BarrierIntermediateKey struct {
	UUID      googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:RootKEKUUID;references:UUID"`
}

// BarrierContentKey represents leaf keys that are wrapped by intermediate keys. Rotation is encouraged and can be very frequent.
type BarrierContentKey struct {
	UUID googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	// Name       string          `gorm:"type:string;unique;not null" validate:"required,min=3,max=50"`
	Encrypted string          `gorm:"type:text;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:IntermediateKEKUUID;references:UUID"`
}

// BarrierKey is an interface for all 3 of the above Keys.
type BarrierKey interface {
	GetUUID() googleUuid.UUID
	SetUUID(googleUuid.UUID)
	GetEncrypted() string
	SetEncrypted(string)
	GetKEKUUID() googleUuid.UUID
	SetKEKUUID(googleUuid.UUID)
}

// GetUUID returns the UUID of the root key.
func (r *BarrierRootKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

// SetUUID sets the UUID of the root key.
func (r *BarrierRootKey) SetUUID(uuidV7 googleUuid.UUID) {
	r.UUID = uuidV7
}

// GetEncrypted returns the encrypted key material.
func (r *BarrierRootKey) GetEncrypted() string {
	return r.Encrypted
}

// SetEncrypted sets the encrypted key material.
func (r *BarrierRootKey) SetEncrypted(encrypted string) {
	r.Encrypted = encrypted
}

// GetKEKUUID returns the UUID of the key encryption key.
func (r *BarrierRootKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

// SetKEKUUID sets the UUID of the key encryption key.
func (r *BarrierRootKey) SetKEKUUID(kekUUIDV7 googleUuid.UUID) {
	r.KEKUUID = kekUUIDV7
}

// GetUUID returns the UUID of the intermediate key.
func (r *BarrierIntermediateKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

// SetUUID sets the UUID of the intermediate key.
func (r *BarrierIntermediateKey) SetUUID(uuidV7 googleUuid.UUID) {
	r.UUID = uuidV7
}

// GetEncrypted returns the encrypted value of the intermediate key.
func (r *BarrierIntermediateKey) GetEncrypted() string {
	return r.Encrypted
}

// SetEncrypted sets the encrypted value of the intermediate key.
func (r *BarrierIntermediateKey) SetEncrypted(encrypted string) {
	r.Encrypted = encrypted
}

// GetKEKUUID returns the UUID of the key encryption key.
func (r *BarrierIntermediateKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

// SetKEKUUID sets the UUID of the key encryption key.
func (r *BarrierIntermediateKey) SetKEKUUID(kekUUIDV7 googleUuid.UUID) {
	r.KEKUUID = kekUUIDV7
}

// GetUUID returns the UUID of the content key.
func (r *BarrierContentKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

// SetUUID sets the UUID of the content key.
func (r *BarrierContentKey) SetUUID(uuidV7 googleUuid.UUID) {
	r.UUID = uuidV7
}

// GetEncrypted returns the encrypted value of the content key.
func (r *BarrierContentKey) GetEncrypted() string {
	return r.Encrypted
}

// SetEncrypted sets the encrypted value of the content key.
func (r *BarrierContentKey) SetEncrypted(encrypted string) {
	r.Encrypted = encrypted
}

// GetKEKUUID returns the UUID of the key encryption key.
func (r *BarrierContentKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

// SetKEKUUID sets the UUID of the key encryption key.
func (r *BarrierContentKey) SetKEKUUID(kekUUIDV7 googleUuid.UUID) {
	r.KEKUUID = kekUUIDV7
}
