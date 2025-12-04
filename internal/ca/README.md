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
├── domain/                # Domain models and interfaces ✅ Task 1
│   ├── certificate.go     # Certificate domain model
│   └── repository.go      # Repository interfaces
├── config/                # CA configuration ✅ Task 2
│   └── config.go          # YAML config loading/validation
├── crypto/                # Cryptographic operations ✅ Task 3
│   └── provider.go        # Provider interface (RSA, ECDSA, EdDSA)
├── profile/               # Certificate profile engines
│   ├── subject/           # Subject template resolution ✅ Task 4
│   └── certificate/       # Certificate policy rendering ✅ Task 5
├── bootstrap/             # Root CA creation ✅ Task 6
│   └── bootstrap.go       # Offline root CA bootstrap workflow
├── intermediate/          # Intermediate CA provisioning ✅ Task 7
│   └── intermediate.go    # Intermediate CA signing workflow
├── service/               # Business logic services (TODO)
│   ├── issuer.go          # Certificate issuance service
│   ├── revocation.go      # Revocation management
│   └── lifecycle.go       # CA lifecycle management
└── repository/            # Persistence layer (TODO)
    └── orm/               # GORM-based repository
```

## Implementation Progress

| Task | Status | Package | Tests |
|------|--------|---------|-------|
| 1. Domain Charter | ✅ | `domain/` | - |
| 2. Configuration Schema | ✅ | `config/` | 10 |
| 3. Crypto Provider | ✅ | `crypto/` | 8 |
| 4. Subject Profile Engine | ✅ | `profile/subject/` | 4 |
| 5. Certificate Profile Engine | ✅ | `profile/certificate/` | 7 |
| 6. Root CA Bootstrap | ✅ | `bootstrap/` | 7 |
| 7. Intermediate CA Provisioning | ✅ | `intermediate/` | 8 |
| 8. Issuing CA Lifecycle | 🔲 | `service/` | - |
| 9. End-Entity Enrollment API | 🔲 | `api/` | - |
| 10. Revocation Services | 🔲 | `service/` | - |

**Total Tests: 40+**

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
