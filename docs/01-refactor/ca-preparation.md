# CA Structure Preparation Plan

## Executive Summary

Prepare skeleton structure for future Certificate Authority (CA) implementation, establishing interfaces, repository patterns, and domain models without full implementation.

**Status**: Planning
**Dependencies**: Tasks 10-11 (identity and KMS extraction complete)
**Risk Level**: Low (skeleton only, no business logic)

## CA Service Vision

From [Service Groups Taxonomy](service-groups.md) Group 3:

### Scope
- Root, Intermediate, and Issuing CA provisioning
- 20+ certificate profile library (TLS server/client, S/MIME, code signing, document signing, VPN, IoT, SAML, JWT, OCSP, RA, TSA, CT log, ACME, SCEP, EST, CMP, enterprise custom)
- Certificate lifecycle management (issuance, renewal, revocation)
- CRL and OCSP responder services
- Time-stamping authority (TSA) and Registration Authority (RA) workflows
- ACME/SCEP/EST/CMP protocol support
- Certificate transparency log integration

### Key Features
- CA/Browser Forum Baseline Requirements compliance
- RFC 5280 strict enforcement
- YAML-driven configuration for crypto, subject, and certificate profiles
- Multi-backend persistence (PostgreSQL, SQLite)
- Observability and audit logging for compliance

### Dependencies
- **Depends On**: KMS (key storage and operations), `internal/common/crypto` (certificate utilities)
- **Used By**: Identity (mTLS), Secrets (TLS), PKI utilities (future)

## Skeleton Structure

```
internal/ca/                    # NEW skeleton package
├── README.md                   # CA service overview and roadmap reference
├── domain/                     # Domain models (skeleton)
│   ├── ca.go                   # CA entity (root, intermediate, issuing)
│   ├── certificate.go          # Certificate entity
│   ├── crl.go                  # CRL entity
│   └── profile.go              # Certificate profile entity
├── repository/                 # Data access layer (interfaces only)
│   ├── interfaces.go           # Repository interfaces (CA, Certificate, CRL)
│   └── README.md               # Repository implementation notes
├── service/                    # Business logic (interfaces only)
│   ├── interfaces.go           # Service interfaces (provisioning, issuance, lifecycle)
│   └── README.md               # Service implementation notes
├── config/                     # Configuration (skeleton)
│   ├── config.go               # CA configuration structure
│   └── defaults.go             # Default CA settings
└── magic/                      # CA-specific constants
    └── magic.go                # Magic values for CA operations
```

## Implementation Details

### Phase 1: Directory Structure

**Create skeleton directories**:

```bash
mkdir -p internal/ca/{domain,repository,service,config,magic}
```

### Phase 2: Domain Models (Skeleton)

**`internal/ca/domain/ca.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
    "time"
    
    googleUuid "github.com/google/uuid"
)

// CAType defines the type of certificate authority.
type CAType string

const (
    CATypeRoot         CAType = "root"
    CATypeIntermediate CAType = "intermediate"
    CATypeIssuing      CAType = "issuing"
)

// CA represents a certificate authority entity.
// Future implementation will integrate with KMS for key storage and signing operations.
type CA struct {
    ID              googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`
    Type            CAType          `gorm:"type:text;not null" json:"type"`
    Subject         string          `gorm:"type:text;not null" json:"subject"`
    Issuer          string          `gorm:"type:text" json:"issuer,omitempty"`
    SerialNumber    string          `gorm:"type:text;uniqueIndex" json:"serial_number"`
    NotBefore       time.Time       `gorm:"not null" json:"not_before"`
    NotAfter        time.Time       `gorm:"not null" json:"not_after"`
    KeyID           googleUuid.UUID `gorm:"type:text;index" json:"key_id"` // Reference to KMS key
    CertificatePEM  string          `gorm:"type:text;not null" json:"certificate_pem"`
    ParentCAID      *googleUuid.UUID `gorm:"type:text;index" json:"parent_ca_id,omitempty"`
    CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (CA) TableName() string {
    return "cas"
}
```

**`internal/ca/domain/certificate.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
    "time"
    
    googleUuid "github.com/google/uuid"
)

// CertificateStatus defines the lifecycle status of a certificate.
type CertificateStatus string

const (
    CertificateStatusPending  CertificateStatus = "pending"
    CertificateStatusIssued   CertificateStatus = "issued"
    CertificateStatusRevoked  CertificateStatus = "revoked"
    CertificateStatusExpired  CertificateStatus = "expired"
)

// Certificate represents an issued certificate.
// Future implementation will support 20+ certificate profiles.
type Certificate struct {
    ID             googleUuid.UUID   `gorm:"type:text;primaryKey" json:"id"`
    CAID           googleUuid.UUID   `gorm:"type:text;not null;index" json:"ca_id"`
    ProfileName    string            `gorm:"type:text;not null" json:"profile_name"` // e.g., "tls-server", "s-mime"
    Subject        string            `gorm:"type:text;not null" json:"subject"`
    SerialNumber   string            `gorm:"type:text;uniqueIndex" json:"serial_number"`
    NotBefore      time.Time         `gorm:"not null" json:"not_before"`
    NotAfter       time.Time         `gorm:"not null" json:"not_after"`
    Status         CertificateStatus `gorm:"type:text;not null;index" json:"status"`
    CertificatePEM string            `gorm:"type:text;not null" json:"certificate_pem"`
    PrivateKeyPEM  string            `gorm:"type:text" json:"private_key_pem,omitempty"` // Encrypted via KMS
    RevokedAt      *time.Time        `gorm:"index" json:"revoked_at,omitempty"`
    CreatedAt      time.Time         `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt      time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (Certificate) TableName() string {
    return "certificates"
}
```

**`internal/ca/domain/profile.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package domain

// Profile defines a certificate profile with cryptographic and extension settings.
// Future implementation will support 20+ profiles (see docs/05-ca/README.md).
type Profile struct {
    Name          string   `yaml:"name" json:"name"`                     // Profile name (e.g., "tls-server")
    KeyType       string   `yaml:"key_type" json:"key_type"`             // "rsa", "ecdsa", "eddsa"
    KeySize       int      `yaml:"key_size" json:"key_size"`             // RSA: 2048/4096; ECDSA: 256/384/521
    SignAlgorithm string   `yaml:"sign_algorithm" json:"sign_algorithm"` // "sha256", "sha384", "sha512"
    ValidityDays  int      `yaml:"validity_days" json:"validity_days"`   // Max 398 for subscriber certs
    KeyUsage      []string `yaml:"key_usage" json:"key_usage"`           // e.g., ["digitalSignature", "keyEncipherment"]
    ExtKeyUsage   []string `yaml:"ext_key_usage" json:"ext_key_usage"`   // e.g., ["serverAuth", "clientAuth"]
}
```

**`internal/ca/domain/crl.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
    "time"
    
    googleUuid "github.com/google/uuid"
)

// CRL represents a Certificate Revocation List.
// Future implementation will support CRL generation and distribution.
type CRL struct {
    ID           googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`
    CAID         googleUuid.UUID `gorm:"type:text;not null;index" json:"ca_id"`
    ThisUpdate   time.Time       `gorm:"not null" json:"this_update"`
    NextUpdate   time.Time       `gorm:"not null" json:"next_update"`
    CRLPEM       string          `gorm:"type:text;not null" json:"crl_pem"`
    CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for GORM.
func (CRL) TableName() string {
    return "crls"
}
```

### Phase 3: Repository Interfaces

**`internal/ca/repository/interfaces.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
    "context"
    
    googleUuid "github.com/google/uuid"
    
    "cryptoutil/internal/ca/domain"
)

// CARepository defines data access operations for certificate authorities.
// Future implementation will use GORM with PostgreSQL/SQLite backends.
type CARepository interface {
    Create(ctx context.Context, ca *domain.CA) error
    GetByID(ctx context.Context, id googleUuid.UUID) (*domain.CA, error)
    GetBySerialNumber(ctx context.Context, serialNumber string) (*domain.CA, error)
    ListByType(ctx context.Context, caType domain.CAType) ([]*domain.CA, error)
    Update(ctx context.Context, ca *domain.CA) error
    Delete(ctx context.Context, id googleUuid.UUID) error
}

// CertificateRepository defines data access operations for certificates.
type CertificateRepository interface {
    Create(ctx context.Context, cert *domain.Certificate) error
    GetByID(ctx context.Context, id googleUuid.UUID) (*domain.Certificate, error)
    GetBySerialNumber(ctx context.Context, serialNumber string) (*domain.Certificate, error)
    ListByCAID(ctx context.Context, caID googleUuid.UUID) ([]*domain.Certificate, error)
    ListByStatus(ctx context.Context, status domain.CertificateStatus) ([]*domain.Certificate, error)
    Update(ctx context.Context, cert *domain.Certificate) error
    Revoke(ctx context.Context, id googleUuid.UUID) error
}

// CRLRepository defines data access operations for certificate revocation lists.
type CRLRepository interface {
    Create(ctx context.Context, crl *domain.CRL) error
    GetLatestByCAID(ctx context.Context, caID googleUuid.UUID) (*domain.CRL, error)
    ListByCAID(ctx context.Context, caID googleUuid.UUID) ([]*domain.CRL, error)
}

// RepositoryFactory creates repository instances.
// Future implementation will support transaction management like identity module.
type RepositoryFactory interface {
    CARepository() CARepository
    CertificateRepository() CertificateRepository
    CRLRepository() CRLRepository
    
    // Transaction executes operations within a database transaction.
    Transaction(ctx context.Context, fn func(context.Context) error) error
}
```

**`internal/ca/repository/README.md`**:

```markdown
# CA Repository Layer

## Implementation Notes

### Database Schema

**Tables**:
- `cas` - Certificate authorities (root, intermediate, issuing)
- `certificates` - Issued certificates with lifecycle status
- `crls` - Certificate revocation lists

### Transaction Patterns

Use transaction context pattern from identity module:
```go
err := repoFactory.Transaction(ctx, func(txCtx context.Context) error {
    ca, err := repoFactory.CARepository().GetByID(txCtx, caID)
    if err != nil {
        return err
    }
    
    cert := &domain.Certificate{...}
    return repoFactory.CertificateRepository().Create(txCtx, cert)
})
```

### Future Implementations

- **ORM Repository**: `repository/orm/` (GORM-based, PostgreSQL/SQLite)
- **Migrations**: `repository/migrations/` (SQL schema files, golang-migrate)
- **Error Mapping**: `repository/orm/errors.go` (GORM → application errors)
```

### Phase 4: Service Interfaces

**`internal/ca/service/interfaces.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
    "context"
    
    googleUuid "github.com/google/uuid"
    
    "cryptoutil/internal/ca/domain"
)

// ProvisioningService handles CA hierarchy provisioning.
// Future implementation will integrate with KMS for key generation and storage.
type ProvisioningService interface {
    // ProvisionRootCA creates a self-signed root CA.
    ProvisionRootCA(ctx context.Context, subject string, validityYears int) (*domain.CA, error)
    
    // ProvisionIntermediateCA creates an intermediate CA signed by a root CA.
    ProvisionIntermediateCA(ctx context.Context, rootCAID googleUuid.UUID, subject string, validityYears int) (*domain.CA, error)
    
    // ProvisionIssuingCA creates an issuing CA signed by an intermediate CA.
    ProvisionIssuingCA(ctx context.Context, intermediateCAID googleUuid.UUID, subject string, validityYears int) (*domain.CA, error)
}

// IssuanceService handles certificate issuance.
type IssuanceService interface {
    // IssueCertificate issues a certificate using the specified profile.
    IssueCertificate(ctx context.Context, caID googleUuid.UUID, profileName string, subject string) (*domain.Certificate, error)
    
    // RenewCertificate renews an existing certificate (new serial number, same key).
    RenewCertificate(ctx context.Context, certID googleUuid.UUID) (*domain.Certificate, error)
}

// LifecycleService handles certificate lifecycle operations.
type LifecycleService interface {
    // RevokeCertificate revokes a certificate and updates CRL.
    RevokeCertificate(ctx context.Context, certID googleUuid.UUID, reason string) error
    
    // GenerateCRL generates a CRL for the specified CA.
    GenerateCRL(ctx context.Context, caID googleUuid.UUID) (*domain.CRL, error)
}
```

**`internal/ca/service/README.md`**:

```markdown
# CA Service Layer

## Implementation Notes

### KMS Integration

CA services will integrate with KMS for:
- **Key Generation**: `POST /browser/api/v1/elastic-keys/generate` (CA keys stored as elastic keys)
- **Key Storage**: CA keys encrypted at rest using KMS barrier
- **Signing Operations**: Use KMS signing APIs for certificate/CRL signing

### Certificate Profiles

Support 20+ profiles (see `docs/05-ca/README.md`):
- TLS server/client, S/MIME, code signing, document signing, VPN, IoT
- SAML, JWT, OCSP, RA, TSA, CT log, ACME, SCEP, EST, CMP
- Enterprise custom profiles

### ACME/SCEP/EST/CMP Protocols

Future services will implement protocol handlers:
- **ACME**: Automated certificate issuance (Let's Encrypt pattern)
- **SCEP**: Simple Certificate Enrollment Protocol (legacy devices)
- **EST**: Enrollment over Secure Transport (RFC 7030)
- **CMP**: Certificate Management Protocol (RFC 4210)
```

### Phase 5: Configuration

**`internal/ca/config/config.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
    "cryptoutil/internal/ca/domain"
)

// Config defines CA service configuration.
// Future implementation will support YAML-driven configuration.
type Config struct {
    Database DatabaseConfig       `yaml:"database" json:"database"`
    KMS      KMSConfig             `yaml:"kms" json:"kms"`
    Profiles map[string]domain.Profile `yaml:"profiles" json:"profiles"`
}

// DatabaseConfig defines database connection settings.
type DatabaseConfig struct {
    Type string `yaml:"type" json:"type"` // "postgres", "sqlite"
    DSN  string `yaml:"dsn" json:"dsn"`   // Connection string
}

// KMSConfig defines KMS integration settings.
type KMSConfig struct {
    BaseURL string `yaml:"base_url" json:"base_url"` // e.g., "https://localhost:8080"
    APIKey  string `yaml:"api_key" json:"api_key"`   // Service API key
}
```

**`internal/ca/config/defaults.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
    "cryptoutil/internal/ca/domain"
)

// DefaultConfig returns default CA configuration.
func DefaultConfig() *Config {
    return &Config{
        Database: DatabaseConfig{
            Type: "sqlite",
            DSN:  ":memory:",
        },
        KMS: KMSConfig{
            BaseURL: "https://localhost:8080",
        },
        Profiles: DefaultProfiles(),
    }
}

// DefaultProfiles returns default certificate profiles.
// Future implementation will load from YAML configuration.
func DefaultProfiles() map[string]domain.Profile {
    return map[string]domain.Profile{
        "tls-server": {
            Name:          "tls-server",
            KeyType:       "rsa",
            KeySize:       2048,
            SignAlgorithm: "sha256",
            ValidityDays:  398, // CA/Browser Forum max for subscriber certs
            KeyUsage:      []string{"digitalSignature", "keyEncipherment"},
            ExtKeyUsage:   []string{"serverAuth"},
        },
        "tls-client": {
            Name:          "tls-client",
            KeyType:       "rsa",
            KeySize:       2048,
            SignAlgorithm: "sha256",
            ValidityDays:  398,
            KeyUsage:      []string{"digitalSignature"},
            ExtKeyUsage:   []string{"clientAuth"},
        },
    }
}
```

### Phase 6: Magic Constants

**`internal/ca/magic/magic.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package magic

import (
    "time"
)

const (
    // DefaultRootCAValidityYears is the default validity period for root CAs.
    DefaultRootCAValidityYears = 20
    
    // DefaultIntermediateCAValidityYears is the default validity period for intermediate CAs.
    DefaultIntermediateCAValidityYears = 10
    
    // DefaultIssuingCAValidityYears is the default validity period for issuing CAs.
    DefaultIssuingCAValidityYears = 5
    
    // DefaultCRLValidityHours is the default validity period for CRLs.
    DefaultCRLValidityHours = 24 * time.Hour
    
    // MinSerialNumberBits is the minimum serial number size per CA/Browser Forum requirements.
    MinSerialNumberBits = 64
    
    // MaxSerialNumberBits is the maximum serial number size (2^159).
    MaxSerialNumberBits = 159
)
```

### Phase 7: README and Roadmap

**`internal/ca/README.md`**:

```markdown
# CA (Certificate Authority) Service

**Status**: 🚧 Skeleton Structure Only

This package contains the skeleton structure for the future Certificate Authority service. Full implementation is planned but not yet started.

## Current State

- ✅ Domain models defined (CA, Certificate, CRL, Profile)
- ✅ Repository interfaces defined (CARepository, CertificateRepository, CRLRepository)
- ✅ Service interfaces defined (ProvisioningService, IssuanceService, LifecycleService)
- ✅ Configuration structure defined
- ✅ Magic constants defined

## Future Implementation

See [`docs/05-ca/README.md`](../../../docs/05-ca/README.md) for full CA implementation roadmap:

- **20 planned tasks** covering:
  - Root/Intermediate/Issuing CA provisioning
  - Certificate issuance with 20+ profiles
  - CRL and OCSP responder services
  - ACME/SCEP/EST/CMP protocol support
  - Certificate transparency log integration
  - Time-stamping authority (TSA)
  - Registration authority (RA) workflows

## Dependencies

- **KMS**: Key generation, storage, and signing operations
- **Common Crypto**: Certificate utilities, ASN.1 parsing, PEM/DER encoding

## Integration Points

### KMS Integration

CA keys will be stored as KMS elastic keys:
```bash
# Generate root CA key
POST /browser/api/v1/elastic-keys/generate
{
  "key_type": "rsa",
  "key_size": 4096,
  "owner_id": "root-ca-001"
}

# Sign certificate
POST /browser/api/v1/crypto/sign
{
  "key_id": "...",
  "data": "base64-encoded-cert-tbs",
  "algorithm": "sha256"
}
```

### Identity Integration (mTLS)

Identity service will use CA for client certificate validation:
```go
// Validate client certificate chain
cert, err := x509.ParseCertificate(clientCertDER)
chains, err := cert.Verify(x509.VerifyOptions{
    Roots: rootCAPool,
    Intermediates: intermediateCAPool,
})
```

## References

- [CA Implementation Roadmap](../../../docs/05-ca/README.md)
- [Service Groups Taxonomy](../../../docs/01-refactor/service-groups.md) - Group 3: CA
- [CA/Browser Forum Baseline Requirements](https://cabforum.org/baseline-requirements-documents/)
- [RFC 5280 - Internet X.509 Public Key Infrastructure](https://tools.ietf.org/html/rfc5280)
```

### Phase 8: Importas Rules

**Update `.golangci.yml`**:

```yaml
# Cryptoutil internal - CA (skeleton)
- pkg: cryptoutil/internal/ca/domain
  alias: cryptoutilCADomain
- pkg: cryptoutil/internal/ca/repository
  alias: cryptoutilCARepository
- pkg: cryptoutil/internal/ca/service
  alias: cryptoutilCAService
- pkg: cryptoutil/internal/ca/config
  alias: cryptoutilCAConfig
- pkg: cryptoutil/internal/ca/magic
  alias: cryptoutilCAMagic
```

**Total**: 5 CA importas rules (skeleton only)

### Phase 9: Testing & Validation

**Validation checklist**:
- [ ] Skeleton structure created (`internal/ca/`)
- [ ] Domain models compile (`domain/*.go`)
- [ ] Repository interfaces compile (`repository/interfaces.go`)
- [ ] Service interfaces compile (`service/interfaces.go`)
- [ ] Configuration compiles (`config/*.go`)
- [ ] Magic constants compile (`magic/magic.go`)
- [ ] README.md created with roadmap reference
- [ ] Importas rules added to `.golangci.yml`
- [ ] golangci-lint passes (no import errors)

**Test commands**:

```bash
# Verify compilation
go build ./internal/ca/...

# Verify imports
go list -m all

# Lint
golangci-lint run --timeout=10m
```

## Risk Assessment

### Low Risks

1. **Skeleton Structure Only**
   - No business logic to test
   - No external dependencies (except googleUuid, standard library)
   - Mitigation: Interfaces define future contracts

2. **Importas Rules**
   - Only 5 new aliases
   - Mitigation: golangci-lint validation

3. **Documentation**
   - README.md references existing roadmap (docs/05-ca/README.md)
   - Mitigation: Keep roadmap reference up-to-date

## Success Metrics

- [ ] Skeleton structure created in `internal/ca/`
- [ ] All files compile without errors
- [ ] Importas rules added (5 CA aliases)
- [ ] golangci-lint passes
- [ ] README.md created with roadmap reference
- [ ] Zero test coverage (no implementation yet - expected)

## Timeline

- **Phase 1**: Directory structure (15 minutes)
- **Phase 2**: Domain models (30 minutes)
- **Phase 3**: Repository interfaces (15 minutes)
- **Phase 4**: Service interfaces (15 minutes)
- **Phase 5**: Configuration (15 minutes)
- **Phase 6**: Magic constants (10 minutes)
- **Phase 7**: README (15 minutes)
- **Phase 8**: Importas rules (10 minutes)
- **Phase 9**: Validation (15 minutes)

**Total**: 2.5 hours

## Cross-References

- [Service Groups Taxonomy](service-groups.md) - Group 3: CA definition
- [CA Implementation Roadmap](../../05-ca/README.md) - Full CA implementation plan (20 tasks)
- [KMS Extraction](kms-extraction.md) - KMS integration points
- [Import Alias Policy](import-aliases.md) - Importas patterns

## Next Steps

After CA skeleton:
1. **Task 13-15**: CLI restructuring (kms, identity, ca commands)
2. **Task 16-18**: Infrastructure updates (workflows, importas, telemetry)
3. **Task 19-20**: Integration testing and handoff

**Future**: Full CA implementation (see `docs/05-ca/README.md` for 20-task roadmap)
