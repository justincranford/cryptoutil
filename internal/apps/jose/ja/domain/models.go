// Copyright (c) 2025 Justin Cranford
//
//

// Package domain provides domain models for the JOSE-JA service.
// These models represent Elastic JWKs (key containers) and Material JWKs (key versions).
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// ElasticJWK represents a logical key container that supports key rotation.
// Each ElasticJWK can have multiple MaterialJWKs (key versions).
// Only one MaterialJWK is active at a time (used for signing/encrypting).
// Retired MaterialJWKs remain available for verification/decryption.
// CRITICAL: TenantID for data scoping only - realms are authentication-only, NOT data scope.
type ElasticJWK struct {
	ID                   googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID             googleUuid.UUID `gorm:"type:text;not null;index:idx_elastic_jwks_tenant"`
	KID                  string          `gorm:"type:text;not null;uniqueIndex:idx_elastic_jwks_unique_kid;column:kid"`
	KeyType              string          `gorm:"type:text;not null;column:kty"`              // RSA, EC, OKP, oct.
	Algorithm            string          `gorm:"type:text;not null;column:alg;index"`        // RS256, ES256, EdDSA, A256GCM, etc.
	Use                  string          `gorm:"type:text;not null;column:use;index"`        // sig or enc.
	MaxMaterials         int             `gorm:"not null;default:1000;column:max_materials"` // Maximum material versions.
	CurrentMaterialCount int             `gorm:"not null;default:0;column:current_material_count"`
	CreatedAt            time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName specifies the database table name for ElasticJWK.
func (ElasticJWK) TableName() string {
	return "elastic_jwks"
}

// MaterialJWK represents a specific key version within an ElasticJWK.
// Private and public JWKs are encrypted at rest using barrier encryption.
type MaterialJWK struct {
	ID             googleUuid.UUID `gorm:"type:text;primaryKey"`
	ElasticJWKID   googleUuid.UUID `gorm:"type:text;not null;index:idx_material_jwks_elastic"`
	MaterialKID    string          `gorm:"type:text;not null;uniqueIndex:idx_material_jwks_unique_kid;column:material_kid"`
	PrivateJWKJWE  string          `gorm:"type:text;not null;column:private_jwk_jwe"` // JWE-encrypted private JWK.
	PublicJWKJWE   string          `gorm:"type:text;not null;column:public_jwk_jwe"`  // JWE-encrypted public JWK.
	Active         bool            `gorm:"not null;default:false;index:idx_material_jwks_active"`
	CreatedAt      time.Time       `gorm:"not null;autoCreateTime"`
	RetiredAt      *time.Time      `gorm:"index"`
	BarrierVersion int             `gorm:"not null;column:barrier_version;index"`

	// Relations.
	ElasticJWK ElasticJWK `gorm:"foreignKey:ElasticJWKID"`
}

// TableName specifies the database table name for MaterialJWK.
func (MaterialJWK) TableName() string {
	return "material_jwks"
}

// AuditConfig represents per-tenant audit settings for operations.
type AuditConfig struct {
	TenantID     googleUuid.UUID `gorm:"type:text;primaryKey"`
	Operation    string          `gorm:"type:text;primaryKey"` // generate, sign, verify, encrypt, decrypt, rotate.
	Enabled      bool            `gorm:"not null;default:true"`
	SamplingRate float64         `gorm:"not null;default:0.01;column:sampling_rate"` // 0.0 to 1.0.
}

// TableName specifies the database table name for AuditConfig.
func (AuditConfig) TableName() string {
	return "tenant_audit_config"
}

// AuditLogEntry represents a single audit log entry for cryptographic operations.
// CRITICAL: TenantID for data scoping only - realms are authentication-only, NOT data scope.
// SessionID added per task requirements for traceability.
type AuditLogEntry struct {
	ID           googleUuid.UUID  `gorm:"type:text;primaryKey"`
	TenantID     googleUuid.UUID  `gorm:"type:text;not null;index:idx_audit_log_tenant"`
	SessionID    *googleUuid.UUID `gorm:"type:text;index:idx_audit_log_session"`     // Session context for operation.
	ElasticJWKID *googleUuid.UUID `gorm:"type:text;index:idx_audit_log_elastic_jwk"` // NULL for non-key operations.
	MaterialKID  *string          `gorm:"type:text;column:material_kid"`             // NULL for non-material operations.
	Operation    string           `gorm:"type:text;not null;index:idx_audit_log_operation"`
	Success      bool             `gorm:"not null;index:idx_audit_log_success"`
	ErrorMessage *string          `gorm:"type:text"`
	UserID       *googleUuid.UUID `gorm:"type:text"`
	ClientID     *googleUuid.UUID `gorm:"type:text"`
	RequestID    string           `gorm:"type:text;not null;index:idx_audit_log_request_id"`
	IPAddress    *string          `gorm:"type:text"`
	UserAgent    *string          `gorm:"type:text"`
	CreatedAt    time.Time        `gorm:"not null;autoCreateTime;index:idx_audit_log_created_at"`
}

// TableName specifies the database table name for AuditLogEntry.
func (AuditLogEntry) TableName() string {
	return "audit_log"
}

// Operation constants for audit logging.
const (
	OperationGenerate = "generate"
	OperationSign     = "sign"
	OperationVerify   = "verify"
	OperationEncrypt  = "encrypt"
	OperationDecrypt  = "decrypt"
	OperationRotate   = "rotate"
)

// KeyType constants.
const (
	KeyTypeRSA = "RSA"
	KeyTypeEC  = "EC"
	KeyTypeOKP = "OKP"
	KeyTypeOct = "oct"
)

// KeyUse constants.
const (
	KeyUseSig = "sig"
	KeyUseEnc = "enc"
)
