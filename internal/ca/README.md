# Certificate Authority (CA) Subsystem

## Overview

The CA subsystem provides cryptographic certificate lifecycle management for cryptoutil. It builds on the existing `internal/common/crypto/certificate/` infrastructure to offer:

- Root, Intermediate, and Issuing CA management
- End-entity certificate issuance (TLS, code signing, S/MIME, etc.)
- Certificate revocation (CRL, OCSP)
- Compliance with CA/Browser Forum Baseline Requirements and RFC 5280

## Architecture

```
internal/ca/
├── README.md              # This file
├── domain/                # Domain models and interfaces
│   ├── certificate.go     # Certificate domain model
│   ├── profile.go         # Certificate profile definitions
│   └── repository.go      # Repository interfaces
├── profile/               # Certificate profile engine
│   ├── subject/           # Subject template resolution
│   └── certificate/       # Certificate policy rendering
├── service/               # Business logic services
│   ├── issuer.go          # Certificate issuance service
│   ├── revocation.go      # Revocation management
│   └── lifecycle.go       # CA lifecycle management
├── repository/            # Persistence layer
│   └── orm/               # GORM-based repository
└── config/                # CA configuration
    └── profiles/          # YAML certificate profiles
```

## Existing Infrastructure

The CA subsystem leverages these existing packages:

| Package | Location | Capabilities |
|---------|----------|--------------|
| Certificate | `internal/common/crypto/certificate/` | CA chain creation, signing, serialization |
| KeyGen | `internal/common/crypto/keygen/` | RSA, ECDSA, ECDH, EdDSA key generation |
| JOSE | `internal/jose/` | JWK generation and management |

## Compliance Requirements

### CA/Browser Forum Baseline Requirements

- Serial number generation: minimum 64 bits CSPRNG, non-sequential, >0, <2^159
- Key sizes: RSA ≥2048, ECDSA P-256/P-384/P-521, Ed25519/Ed448
- Validity period: max 398 days for TLS server certificates
- Required extensions: Subject Key Identifier, Authority Key Identifier, Key Usage
- CRL and OCSP availability

### RFC 5280 Compliance

- X.509 v3 certificate format
- Standard extension profiles
- Certificate path validation
- Name constraints and policy constraints

## Migration Path

### Phase 1: Foundation (Current)

- Use existing `internal/common/crypto/certificate/` for core operations
- Add YAML-based profile configuration
- Implement domain models

### Phase 2: Services

- Certificate issuance service with profile enforcement
- Repository layer with PostgreSQL/SQLite support
- API endpoints for enrollment

### Phase 3: Revocation & Compliance

- CRL generation and distribution
- OCSP responder
- Audit logging and compliance evidence

## Status

| Task | Description | Status |
|------|-------------|--------|
| Task 1 | Domain Charter | ✅ Complete |
| Task 2 | Configuration Schema | 🔄 Planned |
| Task 3 | Crypto Provider Abstractions | 🔄 Planned |
| Task 4 | Subject Profile Engine | 🔄 Planned |
| Task 5 | Certificate Profile Engine | 🔄 Planned |

See `docs/05-ca/README.md` for complete 20-task roadmap.
