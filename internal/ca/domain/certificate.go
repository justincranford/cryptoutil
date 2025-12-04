// Copyright (c) 2025 Justin Cranford
//
//

// Package domain defines the core domain models for the Certificate Authority subsystem.
package domain

import (
	"crypto"
	"crypto/x509"
	"time"

	googleUuid "github.com/google/uuid"
)

// CAType represents the type of Certificate Authority.
type CAType string

const (
	// CATypeRoot is a self-signed root CA (trust anchor).
	CATypeRoot CAType = "root"
	// CATypeIntermediate is an intermediate CA signed by a root or another intermediate.
	CATypeIntermediate CAType = "intermediate"
	// CATypeIssuing is an issuing CA that signs end-entity certificates.
	CATypeIssuing CAType = "issuing"
)

// CAStatus represents the operational status of a CA.
type CAStatus string

const (
	// CAStatusActive indicates the CA is operational and can issue certificates.
	CAStatusActive CAStatus = "active"
	// CAStatusPending indicates the CA is awaiting activation.
	CAStatusPending CAStatus = "pending"
	// CAStatusSuspended indicates the CA is temporarily non-operational.
	CAStatusSuspended CAStatus = "suspended"
	// CAStatusRevoked indicates the CA has been revoked.
	CAStatusRevoked CAStatus = "revoked"
	// CAStatusExpired indicates the CA certificate has expired.
	CAStatusExpired CAStatus = "expired"
)

// CertificateAuthority represents a Certificate Authority in the PKI hierarchy.
type CertificateAuthority struct {
	ID          googleUuid.UUID `gorm:"type:text;primaryKey"`
	Name        string          `gorm:"type:text;not null;uniqueIndex"`
	Description string          `gorm:"type:text"`
	Type        CAType          `gorm:"type:text;not null"`
	Status      CAStatus        `gorm:"type:text;not null;default:pending"`

	// Certificate chain from this CA to the root.
	CertificateChainPEM []byte `gorm:"type:blob"`

	// Parent CA (nil for root CAs).
	ParentID *googleUuid.UUID `gorm:"type:text;index"`

	// CA certificate details.
	SubjectCN   string    `gorm:"type:text;not null"`
	IssuerCN    string    `gorm:"type:text;not null"`
	SerialHex   string    `gorm:"type:text;not null;uniqueIndex"`
	NotBefore   time.Time `gorm:"not null"`
	NotAfter    time.Time `gorm:"not null"`
	MaxPathLen  int       `gorm:"not null;default:0"`
	KeyAlgOID   string    `gorm:"type:text;not null"` // e.g., "1.2.840.10045.2.1" for ECDSA
	KeySizeBits int       `gorm:"not null"`

	// Operational settings.
	CRLDistributionPoint string `gorm:"type:text"`
	OCSPResponderURL     string `gorm:"type:text"`

	// Timestamps.
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

// Certificate represents an issued certificate.
type Certificate struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	ProfileID googleUuid.UUID `gorm:"type:text;not null;index"`
	IssuerID  googleUuid.UUID `gorm:"type:text;not null;index"` // Issuing CA ID

	// Certificate content.
	CertificatePEM []byte `gorm:"type:blob;not null"`
	CertificateDER []byte `gorm:"type:blob;not null"`

	// Certificate details.
	SubjectCN    string    `gorm:"type:text;not null;index"`
	IssuerCN     string    `gorm:"type:text;not null"`
	SerialHex    string    `gorm:"type:text;not null;uniqueIndex"`
	NotBefore    time.Time `gorm:"not null"`
	NotAfter     time.Time `gorm:"not null;index"`
	KeyAlgOID    string    `gorm:"type:text;not null"`
	KeySizeBits  int       `gorm:"not null"`
	DNSNames     string    `gorm:"type:text"` // JSON array of DNS names
	IPAddresses  string    `gorm:"type:text"` // JSON array of IP addresses
	EmailAddrs   string    `gorm:"type:text"` // JSON array of email addresses
	KeyUsage     int       `gorm:"not null"`
	ExtKeyUsages string    `gorm:"type:text"` // JSON array of OIDs

	// Revocation status.
	RevokedAt        *time.Time `gorm:"index"`
	RevocationReason int        `gorm:"default:0"` // RFC 5280 CRLReason

	// Timestamps.
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

// CertificateProfile defines a certificate issuance policy.
type CertificateProfile struct {
	ID          googleUuid.UUID `gorm:"type:text;primaryKey"`
	Name        string          `gorm:"type:text;not null;uniqueIndex"`
	Description string          `gorm:"type:text"`

	// Validity constraints.
	MaxValidityDays int `gorm:"not null;default:365"`
	MinKeySize      int `gorm:"not null;default:2048"`

	// Key usage settings.
	KeyUsage        int    `gorm:"not null"`           // x509.KeyUsage flags
	ExtKeyUsages    string `gorm:"type:text"`          // JSON array of x509.ExtKeyUsage
	AllowedKeyTypes string `gorm:"type:text;not null"` // JSON array: ["RSA", "ECDSA", "EdDSA"]
	AllowedKeyAlgs  string `gorm:"type:text"`          // JSON array of specific algorithms

	// Subject constraints.
	RequireCN          bool   `gorm:"not null;default:true"`
	AllowedSubjectOIDs string `gorm:"type:text"` // JSON array of allowed subject OIDs
	RequiredExtensions string `gorm:"type:text"` // JSON array of required extension OIDs

	// SAN constraints.
	AllowDNSNames    bool `gorm:"not null;default:true"`
	AllowIPAddresses bool `gorm:"not null;default:true"`
	AllowEmailAddrs  bool `gorm:"not null;default:false"`
	AllowURIs        bool `gorm:"not null;default:false"`

	// Active status.
	IsActive bool `gorm:"not null;default:true"`

	// Timestamps.
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

// RevocationEntry represents a revoked certificate entry for CRL generation.
type RevocationEntry struct {
	ID               googleUuid.UUID `gorm:"type:text;primaryKey"`
	CertificateID    googleUuid.UUID `gorm:"type:text;not null;uniqueIndex"`
	IssuerID         googleUuid.UUID `gorm:"type:text;not null;index"` // CA that issued the certificate
	SerialHex        string          `gorm:"type:text;not null;index"`
	RevokedAt        time.Time       `gorm:"not null;index"`
	RevocationReason int             `gorm:"not null;default:0"` // RFC 5280 CRLReason
	InvalidityDate   *time.Time      `gorm:""`                   // Optional invalidity date
	CreatedAt        time.Time       `gorm:"not null;autoCreateTime"`
}

// CertificateRequest represents a pending certificate signing request.
type CertificateRequest struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	ProfileID googleUuid.UUID `gorm:"type:text;not null;index"`
	IssuerID  googleUuid.UUID `gorm:"type:text;not null;index"` // Target issuing CA

	// Request content.
	CSRPEM      []byte `gorm:"type:blob;not null"`
	SubjectCN   string `gorm:"type:text;not null"`
	DNSNames    string `gorm:"type:text"` // JSON array
	IPAddresses string `gorm:"type:text"` // JSON array
	EmailAddrs  string `gorm:"type:text"` // JSON array

	// Request status.
	Status       string     `gorm:"type:text;not null;default:pending"` // pending, approved, rejected, issued
	ApprovedAt   *time.Time `gorm:""`
	ApprovedBy   string     `gorm:"type:text"`
	RejectedAt   *time.Time `gorm:""`
	RejectedBy   string     `gorm:"type:text"`
	RejectReason string     `gorm:"type:text"`

	// Issued certificate (if approved and issued).
	CertificateID *googleUuid.UUID `gorm:"type:text;index"`

	// Timestamps.
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

// KeyMaterial holds cryptographic key material for a CA.
// Note: This is a transient struct, not persisted. Keys are stored encrypted.
type KeyMaterial struct {
	PublicKey        crypto.PublicKey
	PrivateKey       crypto.PrivateKey
	CertificateChain []*x509.Certificate
}
