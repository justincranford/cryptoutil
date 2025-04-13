package orm

import (
	googleUuid "github.com/google/uuid"
)

// UUIDv7 = timestamp (48-bits) + version (4-bits) + rand_a (12-bits) + var (2-bits) + rand_b (62-bits)
// JWK/JWKs = JWE wrapping of JWK/JWKs, stored as JSON (PostgreSQL JSONB, SQLite JSON)

// Root Keys are unsealed by HSM, KMS, Shamir Key Shares, etc. Rotation is posible but infrequent.
type BarrierRootKey struct {
	UUID      googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Encrypted string          `gorm:"type:json;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null"`
}

// Intermediate Keys are wrapped by root Keys. Rotation is encouraged and can be frequent.
type BarrierIntermediateKey struct {
	UUID      googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	Encrypted string          `gorm:"type:json;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
	KEKUUID   googleUuid.UUID `gorm:"type:uuid;not null;foreignKey:RootKEKUUID;references:UUID"`
}

// Leaf Keys are wrapped by Intermediate Keys. Rotation is encouraged and can be very frequent.
type BarrierContentKey struct {
	UUID googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	// Name       string          `gorm:"type:string;unique;not null" validate:"required,min=3,max=50"`
	Encrypted string          `gorm:"type:json;not null"` // Encrypted column contains JWEs (JOSE Encrypted JSON doc)
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

func (r *BarrierRootKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

func (r *BarrierRootKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}

func (r *BarrierRootKey) GetEncrypted() string {
	return r.Encrypted
}

func (r *BarrierRootKey) SetEncrypted(serialized string) {
	r.Encrypted = serialized
}

func (r *BarrierRootKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

func (r *BarrierRootKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}

func (r *BarrierIntermediateKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

func (r *BarrierIntermediateKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}

func (r *BarrierIntermediateKey) GetEncrypted() string {
	return r.Encrypted
}

func (r *BarrierIntermediateKey) SetEncrypted(serialized string) {
	r.Encrypted = serialized
}

func (r *BarrierIntermediateKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

func (r *BarrierIntermediateKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}

func (r *BarrierContentKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

func (r *BarrierContentKey) SetUUID(uuid googleUuid.UUID) {
	r.UUID = uuid
}

func (r *BarrierContentKey) GetEncrypted() string {
	return r.Encrypted
}

func (r *BarrierContentKey) SetEncrypted(serialized string) {
	r.Encrypted = serialized
}

func (r *BarrierContentKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

func (r *BarrierContentKey) SetKEKUUID(kekUUID googleUuid.UUID) {
	r.KEKUUID = kekUUID
}
