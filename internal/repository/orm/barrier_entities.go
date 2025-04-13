package orm

import (
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

// BarrierKey is an interface for all 3 of the above Keys.
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
