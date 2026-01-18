// Copyright (c) 2025 Justin Cranford
//
//

// Package domain contains JOSE-JA domain models for elastic JWKs, material JWKs, and audit.
package domain

import (
	googleUuid "github.com/google/uuid"
)

// ElasticJWK represents a key ring with multiple Material JWKs (key rotation support).
// Each Elastic JWK is a logical key that can have many Material JWKs over time.
type ElasticJWK struct {
	ID                   googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID             googleUuid.UUID `gorm:"column:tenant_id;type:text;not null;index:idx_elastic_jwks_tenant_realm"`
	RealmID              googleUuid.UUID `gorm:"column:realm_id;type:text;not null;index:idx_elastic_jwks_tenant_realm"`
	KID                  string          `gorm:"column:kid;type:text;not null;uniqueIndex:idx_elastic_jwks_tenant_realm_kid,composite:tenant_id,realm_id"`
	KTY                  string          `gorm:"column:kty;type:text;not null"` // RSA, EC, oct.
	ALG                  string          `gorm:"column:alg;type:text;not null"` // RS256, ES256, A256GCM, etc.
	USE                  string          `gorm:"column:use;type:text;not null"` // sig, enc.
	MaxMaterials         int             `gorm:"column:max_materials;not null;default:1000"`
	CurrentMaterialCount int             `gorm:"column:current_material_count;not null;default:0"`
	CreatedAt            int64           `gorm:"column:created_at;autoCreateTime:milli"`
}

// TableName returns the table name for ElasticJWK.
func (ElasticJWK) TableName() string {
	return "elastic_jwks"
}

// MaterialJWK represents actual cryptographic key material for an Elastic JWK.
// Each Material JWK is a versioned key used for encryption/signing operations.
// Active key used for new operations, retired keys used for decryption/verification.
type MaterialJWK struct {
	ID             googleUuid.UUID `gorm:"type:text;primaryKey"`
	ElasticJWKID   googleUuid.UUID `gorm:"column:elastic_jwk_id;type:text;not null;index:idx_material_jwks_elastic;index:idx_material_jwks_active,composite:active;uniqueIndex:idx_material_jwks_elastic_material_kid,composite:material_kid"`
	MaterialKID    string          `gorm:"column:material_kid;type:text;not null;uniqueIndex:idx_material_jwks_elastic_material_kid,composite:elastic_jwk_id"`
	PrivateJWKJWE  string          `gorm:"column:private_jwk_jwe;type:text;not null"` // Private key encrypted with barrier.
	PublicJWKJWE   string          `gorm:"column:public_jwk_jwe;type:text;not null"`  // Public key encrypted with barrier.
	Active         bool            `gorm:"column:active;not null;default:false;index:idx_material_jwks_active,composite:elastic_jwk_id"`
	CreatedAt      int64           `gorm:"column:created_at;autoCreateTime:milli"`
	RetiredAt      *int64          `gorm:"column:retired_at;default:null"`
	BarrierVersion int             `gorm:"column:barrier_version;not null"`
}

// TableName returns the table name for MaterialJWK.
func (MaterialJWK) TableName() string {
	return "material_jwks"
}

// AuditConfig represents per-tenant, per-operation audit settings.
type AuditConfig struct {
	TenantID     googleUuid.UUID `gorm:"type:text;primaryKey;index:idx_tenant_audit_config_tenant"`
	Operation    string          `gorm:"type:text;primaryKey"` // encrypt, decrypt, sign, verify, keygen, rotate.
	Enabled      bool            `gorm:"not null"`
	SamplingRate float64         `gorm:"not null"` // 0.0 to 1.0, 1% sampling = 0.01.
}

// TableName returns the table name for AuditConfig.
func (AuditConfig) TableName() string {
	return "tenant_audit_config"
}

// AuditLogEntry represents a sampled cryptographic operation for compliance.
type AuditLogEntry struct {
	ID           googleUuid.UUID  `gorm:"type:text;primaryKey"`
	TenantID     googleUuid.UUID  `gorm:"type:text;not null;index:idx_audit_log_tenant_realm"`
	RealmID      googleUuid.UUID  `gorm:"type:text;not null;index:idx_audit_log_tenant_realm"`
	UserID       *googleUuid.UUID `gorm:"type:text"`
	Operation    string           `gorm:"type:text;not null;index:idx_audit_log_operation"` // encrypt, decrypt, sign, verify, keygen, rotate.
	ResourceType string           `gorm:"type:text;not null;index:idx_audit_log_resource,composite:resource_id"`
	ResourceID   string           `gorm:"type:text;not null;index:idx_audit_log_resource,composite:resource_type"`
	Success      bool             `gorm:"not null"`
	ErrorMessage *string          `gorm:"type:text"`
	Metadata     *string          `gorm:"type:text"` // JSON blob with operation-specific details.
	CreatedAt    int64            `gorm:"autoCreateTime:milli;index:idx_audit_log_created_at"`
}

// TableName returns the table name for AuditLogEntry.
func (AuditLogEntry) TableName() string {
	return "tenant_audit_log"
}
