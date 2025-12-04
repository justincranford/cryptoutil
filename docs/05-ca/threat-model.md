# CA Security Threat Model

## Overview

This document provides a STRIDE-based threat model for the Certificate Authority (CA) subsystem. It identifies assets, threats, and controls to ensure secure CA operations.

## Document Information

| Field | Value |
|-------|-------|
| Version | 1.0.0 |
| Status | Active |
| Last Updated | 2025-01-16 |
| Classification | Internal |

## Scope

This threat model covers:

- Root CA operations (offline)
- Intermediate CA operations
- Issuing CA operations
- Certificate enrollment and issuance
- Revocation services (CRL/OCSP)
- Time-stamping services
- Key management and storage

## Assets

### Critical Assets

| Asset ID | Asset | Type | Sensitivity | Description |
|----------|-------|------|-------------|-------------|
| ASSET-001 | Root CA Private Key | Cryptographic Key | Critical | Private key for the root certificate authority. Compromise would invalidate entire PKI hierarchy. |
| ASSET-002 | Intermediate CA Private Keys | Cryptographic Key | High | Private keys for intermediate certificate authorities. Compromise allows unauthorized certificate issuance. |
| ASSET-003 | Issuing CA Private Keys | Cryptographic Key | High | Private keys for issuing CAs. Compromise allows unauthorized end-entity certificate issuance. |

### High Value Assets

| Asset ID | Asset | Type | Sensitivity | Description |
|----------|-------|------|-------------|-------------|
| ASSET-004 | Certificate Database | Database | High | Database storing issued certificates, status, and metadata. |
| ASSET-005 | CRL Signing Key | Cryptographic Key | High | Key used to sign Certificate Revocation Lists. |
| ASSET-006 | OCSP Responder Key | Cryptographic Key | High | Key used for OCSP response signing. |
| ASSET-007 | TSA Signing Key | Cryptographic Key | High | Time-stamping authority signing key. |

### Medium Value Assets

| Asset ID | Asset | Type | Sensitivity | Description |
|----------|-------|------|-------------|-------------|
| ASSET-008 | Audit Logs | Log Data | Medium | Logs of all CA operations for compliance and forensics. |
| ASSET-009 | Configuration Files | Configuration | Medium | CA configuration including policies and profiles. |
| ASSET-010 | Certificate Profiles | Configuration | Medium | Templates defining certificate policies and extensions. |

## STRIDE Threat Analysis

### S - Spoofing

| Threat ID | Title | Asset | Severity | Likelihood | Impact |
|-----------|-------|-------|----------|------------|--------|
| THREAT-S-001 | Unauthorized Certificate Issuance | ASSET-002, ASSET-003 | Critical | Medium | Critical |
| THREAT-S-002 | Impersonation of CA Administrator | ASSET-002, ASSET-003 | High | Low | High |
| THREAT-S-003 | Forged Certificate Requests | ASSET-004 | Medium | Medium | Medium |

#### THREAT-S-001: Unauthorized Certificate Issuance

**Description**: An attacker issues certificates without proper authorization, enabling impersonation attacks.

**Attack Vectors**:

- Compromised CA credentials
- Bypassed approval workflows
- Exploited API vulnerabilities

**Mitigations**:

- Multi-party approval for certificate issuance
- HSM-protected private keys
- Comprehensive audit logging
- Role-based access control (RBAC)
- Certificate Transparency (CT) logging

### T - Tampering

| Threat ID | Title | Asset | Severity | Likelihood | Impact |
|-----------|-------|-------|----------|------------|--------|
| THREAT-T-001 | Certificate Database Tampering | ASSET-004 | High | Low | High |
| THREAT-T-002 | CRL Manipulation | ASSET-005 | High | Low | High |
| THREAT-T-003 | Audit Log Tampering | ASSET-008 | Medium | Medium | Medium |
| THREAT-T-004 | Configuration Modification | ASSET-009 | Medium | Low | Medium |

#### THREAT-T-001: Certificate Database Tampering

**Description**: An attacker modifies certificate records to change validity status or metadata.

**Attack Vectors**:

- SQL injection
- Direct database access
- Compromised database credentials

**Mitigations**:

- Database integrity checks (checksums, MACs)
- Strict access controls
- Immutable audit logging
- Database activity monitoring
- Parameterized queries (prevent SQL injection)

### R - Repudiation

| Threat ID | Title | Asset | Severity | Likelihood | Impact |
|-----------|-------|-------|----------|------------|--------|
| THREAT-R-001 | Denied Certificate Operations | ASSET-008 | Medium | Medium | Medium |
| THREAT-R-002 | Disputed Revocation Actions | ASSET-005, ASSET-008 | Medium | Low | Medium |

#### THREAT-R-001: Denied Certificate Operations

**Description**: An operator or system denies performing certificate operations (issuance, revocation).

**Mitigations**:

- Immutable, append-only audit logs
- Digital signatures on log entries
- Centralized log aggregation with integrity protection
- Cryptographic timestamps on all operations

### I - Information Disclosure

| Threat ID | Title | Asset | Severity | Likelihood | Impact |
|-----------|-------|-------|----------|------------|--------|
| THREAT-I-001 | Private Key Exposure | ASSET-001, ASSET-002, ASSET-003 | Critical | Low | Critical |
| THREAT-I-002 | Audit Log Disclosure | ASSET-008 | Medium | Medium | Medium |
| THREAT-I-003 | Configuration Leakage | ASSET-009 | Low | Medium | Low |

#### THREAT-I-001: Private Key Exposure

**Description**: CA private keys are disclosed to unauthorized parties, enabling complete PKI compromise.

**Attack Vectors**:

- Memory dump attacks
- Backup media theft
- Insider threats
- Side-channel attacks

**Mitigations**:

- Hardware Security Module (HSM) storage
- Formal key ceremony procedures
- Strict access controls
- Key encryption at rest
- Secure key backup procedures

### D - Denial of Service

| Threat ID | Title | Asset | Severity | Likelihood | Impact |
|-----------|-------|-------|----------|------------|--------|
| THREAT-D-001 | CA Service Unavailability | ASSET-004 | Medium | Medium | Medium |
| THREAT-D-002 | OCSP Responder Overload | ASSET-006 | Medium | High | Medium |
| THREAT-D-003 | CRL Distribution Disruption | ASSET-005 | Medium | Medium | Medium |

#### THREAT-D-001: CA Service Unavailability

**Description**: CA services become unavailable, preventing certificate issuance and revocation checking.

**Mitigations**:

- High availability deployment
- Rate limiting and throttling
- DDoS protection
- Caching for OCSP and CRL
- Geographic distribution

### E - Elevation of Privilege

| Threat ID | Title | Asset | Severity | Likelihood | Impact |
|-----------|-------|-------|----------|------------|--------|
| THREAT-E-001 | Privilege Escalation to CA Admin | ASSET-002, ASSET-003 | Critical | Low | Critical |
| THREAT-E-002 | Bypass of Approval Workflow | ASSET-004 | High | Low | High |

#### THREAT-E-001: Privilege Escalation to CA Admin

**Description**: An attacker gains CA administrator privileges, enabling full control over certificate operations.

**Mitigations**:

- Role-based access control (RBAC)
- Multi-factor authentication (MFA)
- Separation of duties
- Principle of least privilege
- Regular access reviews

## Security Controls

### Technical Controls

| Control ID | Control | Type | Mitigates | Status |
|------------|---------|------|-----------|--------|
| CTRL-001 | HSM Key Storage | Technical | THREAT-I-001, THREAT-S-001 | Implemented |
| CTRL-002 | Audit Logging | Technical | THREAT-R-001, THREAT-T-001 | Implemented |
| CTRL-003 | Rate Limiting | Technical | THREAT-D-001, THREAT-D-002 | Implemented |
| CTRL-004 | Input Validation | Technical | THREAT-T-001, THREAT-S-003 | Implemented |
| CTRL-005 | TLS/mTLS | Technical | THREAT-I-002, THREAT-T-003 | Implemented |
| CTRL-006 | Database Encryption | Technical | THREAT-I-001, THREAT-T-001 | Implemented |

### Procedural Controls

| Control ID | Control | Type | Mitigates | Status |
|------------|---------|------|-----------|--------|
| CTRL-007 | Multi-Party Approval | Procedural | THREAT-S-001, THREAT-E-001 | Implemented |
| CTRL-008 | Key Ceremony | Procedural | THREAT-I-001 | Documented |
| CTRL-009 | Access Reviews | Procedural | THREAT-E-001 | Planned |
| CTRL-010 | Incident Response | Procedural | All | Documented |

## Risk Matrix

| Likelihood \ Impact | Low | Medium | High | Critical |
|---------------------|-----|--------|------|----------|
| **High** | Low | Medium | High | Critical |
| **Medium** | Low | Medium | High | High |
| **Low** | Info | Low | Medium | High |

## Compliance Requirements

This threat model supports compliance with:

- **CA/Browser Forum Baseline Requirements** - Section 5.1 (Security)
- **WebTrust for CAs** - Physical and logical security requirements
- **RFC 5280** - Internet X.509 PKI Certificate and CRL Profile
- **NIST SP 800-57** - Key Management Guidelines

## Security Hardening Checklist

### Cryptographic Security

- [ ] Use HSM for root CA key storage
- [ ] Use HSM or software cryptostore for intermediate/issuing CA keys
- [ ] Enforce minimum key sizes (RSA ≥ 2048, EC ≥ 256)
- [ ] Use only FIPS 140-3 approved algorithms
- [ ] Implement key rotation procedures
- [ ] Enable algorithm agility for future updates

### Access Control

- [ ] Implement RBAC with least privilege
- [ ] Require MFA for administrative access
- [ ] Enable separation of duties for sensitive operations
- [ ] Conduct regular access reviews
- [ ] Implement session timeouts

### Audit and Monitoring

- [ ] Enable comprehensive audit logging
- [ ] Protect audit logs from tampering
- [ ] Forward logs to centralized SIEM
- [ ] Set up alerting for security events
- [ ] Retain logs per compliance requirements

### Network Security

- [ ] Deploy behind firewall/WAF
- [ ] Use TLS 1.3 for all communications
- [ ] Implement mTLS for internal services
- [ ] Enable rate limiting
- [ ] Configure DDoS protection

### Operational Security

- [ ] Document key ceremony procedures
- [ ] Create and test incident response plan
- [ ] Establish disaster recovery procedures
- [ ] Conduct regular security assessments
- [ ] Maintain vulnerability management program

## Implementation Notes

The security package at `internal/ca/security/` provides:

1. **SecurityValidator** - Validates certificates, keys, and CSRs against security policies
2. **ThreatModelBuilder** - Programmatic construction of threat models
3. **SecurityScanner** - Scans certificate chains for security issues
4. **SecurityReport** - Generates comprehensive security reports

### Usage Example

```go
import "cryptoutil/internal/ca/security"

// Create validator with default security config
validator := security.NewSecurityValidator(nil)

// Validate a certificate
result, err := validator.ValidateCertificate(ctx, cert)
if !result.Valid {
    for _, e := range result.Errors {
        log.Printf("Security error: %s", e)
    }
}

// Generate threat model
model := security.CAThreatModel()

// Create security report
report := security.GenerateSecurityReport(model, validations)
```

## References

- [STRIDE Threat Model](https://docs.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats)
- [CA/Browser Forum Baseline Requirements](https://cabforum.org/baseline-requirements-documents/)
- [RFC 5280 - X.509 PKI Certificate Profile](https://tools.ietf.org/html/rfc5280)
- [NIST SP 800-57 - Key Management](https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final)
