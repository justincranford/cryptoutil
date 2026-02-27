// Package orm provides database entity definitions and ORM operations for the KMS server.
package orm

import (
	googleUuid "github.com/google/uuid"
)

// UUIDv7 = timestamp (48-bits) + version (4-bits) + rand_a (12-bits) + var (2-bits) + rand_b (62-bits)
// JWK/JWKs = JWE wrapping of JWK/JWKs, stored as JSON (PostgreSQL JSONB, SQLite JSON)

// RootKey represents root keys that are unsealed by HSM, KMS, Shamir Key Shares, etc. Rotation is possible but infrequent.
type RootKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted root key
	KEKUUID   googleUuid.UUID `gorm:"type:text"`                              // KEK UUID (nil for root keys)
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
}

// TableName returns the table name for RootKey.
func (RootKey) TableName() string {
	return "barrier_root_keys"
}

// IntermediateKey represents intermediate keys that are wrapped by root keys. Rotation is encouraged and can be frequent.
type IntermediateKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted intermediate key
	KEKUUID   googleUuid.UUID `gorm:"type:text;not null"`                     // Parent root key UUID
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
}

// TableName returns the table name for IntermediateKey.
func (IntermediateKey) TableName() string {
	return "barrier_intermediate_keys"
}

// ContentKey represents leaf keys that are wrapped by intermediate keys. Rotation is encouraged and can be very frequent.
type ContentKey struct {
	UUID      googleUuid.UUID `gorm:"type:text;primaryKey"`
	Encrypted string          `gorm:"type:text;not null"`                     // JWE-encrypted content key
	KEKUUID   googleUuid.UUID `gorm:"type:text;not null"`                     // Parent intermediate key UUID
	CreatedAt int64           `gorm:"autoCreateTime:milli" json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64           `gorm:"autoUpdateTime:milli" json:"updated_at"` // Unix epoch milliseconds
}

// TableName returns the table name for ContentKey.
func (ContentKey) TableName() string {
	return "barrier_content_keys"
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
func (r *RootKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

// SetUUID sets the UUID of the root key.
func (r *RootKey) SetUUID(uuidV7 googleUuid.UUID) {
	r.UUID = uuidV7
}

// GetEncrypted returns the encrypted key material.
func (r *RootKey) GetEncrypted() string {
	return r.Encrypted
}

// SetEncrypted sets the encrypted key material.
func (r *RootKey) SetEncrypted(encrypted string) {
	r.Encrypted = encrypted
}

// GetKEKUUID returns the UUID of the key encryption key.
func (r *RootKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

// SetKEKUUID sets the UUID of the key encryption key.
func (r *RootKey) SetKEKUUID(kekUUIDV7 googleUuid.UUID) {
	r.KEKUUID = kekUUIDV7
}

// GetUUID returns the UUID of the intermediate key.
func (r *IntermediateKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

// SetUUID sets the UUID of the intermediate key.
func (r *IntermediateKey) SetUUID(uuidV7 googleUuid.UUID) {
	r.UUID = uuidV7
}

// GetEncrypted returns the encrypted value of the intermediate key.
func (r *IntermediateKey) GetEncrypted() string {
	return r.Encrypted
}

// SetEncrypted sets the encrypted value of the intermediate key.
func (r *IntermediateKey) SetEncrypted(encrypted string) {
	r.Encrypted = encrypted
}

// GetKEKUUID returns the UUID of the key encryption key.
func (r *IntermediateKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

// SetKEKUUID sets the UUID of the key encryption key.
func (r *IntermediateKey) SetKEKUUID(kekUUIDV7 googleUuid.UUID) {
	r.KEKUUID = kekUUIDV7
}

// GetUUID returns the UUID of the content key.
func (r *ContentKey) GetUUID() googleUuid.UUID {
	return r.UUID
}

// SetUUID sets the UUID of the content key.
func (r *ContentKey) SetUUID(uuidV7 googleUuid.UUID) {
	r.UUID = uuidV7
}

// GetEncrypted returns the encrypted value of the content key.
func (r *ContentKey) GetEncrypted() string {
	return r.Encrypted
}

// SetEncrypted sets the encrypted value of the content key.
func (r *ContentKey) SetEncrypted(encrypted string) {
	r.Encrypted = encrypted
}

// GetKEKUUID returns the UUID of the key encryption key.
func (r *ContentKey) GetKEKUUID() googleUuid.UUID {
	return r.KEKUUID
}

// SetKEKUUID sets the UUID of the key encryption key.
func (r *ContentKey) SetKEKUUID(kekUUIDV7 googleUuid.UUID) {
	r.KEKUUID = kekUUIDV7
}
