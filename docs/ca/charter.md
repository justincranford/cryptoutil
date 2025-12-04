# CA Subsystem Domain Charter

**Version**: 1.0
**Date**: December 3, 2025
**Status**: APPROVED

---

## Purpose

The Certificate Authority (CA) subsystem provides enterprise-grade certificate lifecycle management capabilities for cryptoutil, enabling:

1. **Internal PKI**: Self-contained certificate hierarchies for development, testing, and internal services
2. **TLS Automation**: Automated TLS certificate provisioning for cryptoutil services
3. **Code Signing**: Certificate-based code and artifact signing
4. **Identity Binding**: Cryptographic identity for users, services, and devices

---

## Scope

### In Scope

| Capability | Description | Priority |
|------------|-------------|----------|
| CA Hierarchy | Root, Intermediate, Issuing CA management | HIGH |
| TLS Certificates | Server and client TLS certificate issuance | HIGH |
| Code Signing | Software and artifact signing certificates | MEDIUM |
| S/MIME | Email signing and encryption certificates | LOW |
| Certificate Revocation | CRL generation and OCSP responder | HIGH |
| Certificate Profiles | 20+ predefined certificate profiles | HIGH |
| YAML Configuration | Declarative certificate policy definition | HIGH |
| Multi-backend | PostgreSQL and SQLite support | HIGH |
| Audit Logging | Certificate lifecycle audit trail | HIGH |

### Out of Scope

| Capability | Reason | Alternative |
|------------|--------|-------------|
| Public CA | Requires WebTrust audit, not planned | Use Let's Encrypt, DigiCert |
| HSM Integration | Future enhancement | Software-based key storage |
| Certificate Transparency | Public CA requirement | Not applicable for internal PKI |
| ACME Protocol | Future enhancement | REST API enrollment |

---

## Compliance Obligations

### CA/Browser Forum Baseline Requirements

The CA subsystem SHALL comply with relevant sections of the CA/Browser Forum Baseline Requirements for internal certificate profiles that mirror public certificate types:

- **Section 6.1.5**: Key sizes (RSA ≥2048, ECDSA P-256+)
- **Section 7.1**: Certificate serial numbers (64+ bits CSPRNG)
- **Section 7.1.2**: Required certificate extensions
- **Section 7.1.3.2**: Approved signature algorithms
- **Section 7.2**: CRL profile requirements
- **Section 7.3**: OCSP profile requirements

### RFC Compliance

| RFC | Title | Applicability |
|-----|-------|---------------|
| RFC 5280 | X.509 PKI Certificate and CRL Profile | Core certificate format |
| RFC 6960 | OCSP | Online certificate status |
| RFC 3161 | Time-Stamp Protocol | Timestamp authority (future) |
| RFC 5652 | CMS | Signed data structures |

### FIPS 140-3

All cryptographic operations SHALL use FIPS 140-3 approved algorithms:

- RSA (2048, 3072, 4096 bits)
- ECDSA (P-256, P-384, P-521)
- EdDSA (Ed25519, Ed448) - Note: Check FIPS status
- SHA-256, SHA-384, SHA-512

---

## Non-Goals

1. **Not a Public CA**: Will not issue publicly trusted certificates
2. **Not a Full PKI Suite**: Focus on certificate lifecycle, not full PKI policy management
3. **Not a Replacement for External CAs**: Complement, not replace, external certificate providers
4. **Not HSM-Required**: Initial implementation uses software key storage

---

## Dependencies

### Internal Dependencies

| Package | Purpose |
|---------|---------|
| `internal/common/crypto/certificate/` | Core certificate operations |
| `internal/common/crypto/keygen/` | Key generation |
| `internal/jose/` | JWK operations |
| `internal/infra/telemetry/` | Observability |
| `internal/infra/database/` | Persistence |

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `crypto/x509` | stdlib | X.509 certificate handling |
| `crypto/tls` | stdlib | TLS configuration |
| `gorm.io/gorm` | v1.25+ | ORM for persistence |

---

## Success Criteria

### Phase 1 (Q1 2026)

- [ ] CA hierarchy creation (root, intermediate, issuing)
- [ ] TLS server certificate issuance
- [ ] YAML profile configuration
- [ ] SQLite persistence

### Phase 2 (Q2 2026)

- [ ] PostgreSQL persistence
- [ ] CRL generation
- [ ] REST API for enrollment
- [ ] 10+ certificate profiles

### Phase 3 (Q3 2026)

- [ ] OCSP responder
- [ ] Certificate rotation automation
- [ ] Audit logging and compliance evidence
- [ ] 20+ certificate profiles

---

## Glossary

| Term | Definition |
|------|------------|
| CA | Certificate Authority - entity that issues digital certificates |
| CRL | Certificate Revocation List - list of revoked certificates |
| OCSP | Online Certificate Status Protocol - real-time certificate status |
| PKI | Public Key Infrastructure - system for certificate management |
| RA | Registration Authority - validates certificate requests |
| SAN | Subject Alternative Name - additional identities in certificate |

---

## Stakeholder Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Project Owner | Justin Cranford | 2025-12-03 | ✅ Approved |
| Security Review | TBD | - | Pending |
| Architecture Review | TBD | - | Pending |
