# PKI and Certificate Management - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/02-09.pki.instructions.md`

## Certificate Validation Requirements

### TLS Configuration - MANDATORY

**Full cert chain validation, TLS 1.3+, NEVER InsecureSkipVerify**

```go
import "crypto/tls"

// ✅ CORRECT: Strict TLS configuration
tlsConfig := &tls.Config{
    MinVersion:         tls.VersionTLS13,
    InsecureSkipVerify: false,  // ALWAYS validate certificates
    RootCAs:            certPool,
    ClientCAs:          certPool,
    ClientAuth:         tls.RequireAndVerifyClientCert,
}

// ❌ WRONG: Insecure TLS (bypasses validation)
tlsConfig := &tls.Config{
    InsecureSkipVerify: true,   // NEVER do this
    MinVersion:         tls.VersionTLS12,  // Too old
}
```

**See**: `02-07.cryptography.instructions.md` for complete cryptographic requirements

---

## CA/Browser Forum Baseline Requirements

**MANDATORY: Adhere to latest CA/Browser Forum Baseline Requirements for TLS Server Certificates**

**Reference**: [CA/Browser Forum Baseline Requirements](https://cabforum.org/baseline-requirements/)

### Certificate Profile Requirements (Section 7)

#### Serial Number Generation (Section 7.1)

**Requirements**:
- Minimum 64 bits from CSPRNG (Cryptographically Secure Pseudo-Random Number Generator)
- Non-sequential generation (MUST NOT be predictable)
- Greater than zero (>0)
- Less than 2^159 (maximum value)
- MUST be unique per CA
- MUST NOT be reused within CA lifetime

**Implementation**:

```go
import crand "crypto/rand"

func GenerateSerialNumber() (*big.Int, error) {
    // Generate 20 random bytes (160 bits, well above 64-bit minimum)
    serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
    serialNumber, err := crand.Int(crand.Reader, serialNumberLimit)
    if err != nil {
        return nil, fmt.Errorf("failed to generate serial number: %w", err)
    }
    
    // Ensure >0
    if serialNumber.Cmp(big.NewInt(0)) <= 0 {
        return nil, fmt.Errorf("invalid serial number: must be >0")
    }
    
    return serialNumber, nil
}
```

#### Cryptographic Algorithms and Key Sizes (Section 6.1.5, 6.1.6)

**Approved Algorithms**:

| Algorithm | Minimum Key Size | Recommended | Hash Algorithm |
|-----------|------------------|-------------|----------------|
| RSA | 2048 bits | 3072/4096 bits | SHA-256/384/512 |
| ECDSA | P-256 (256 bits) | P-384 (384 bits) | SHA-256/384/512 |
| EdDSA | Ed25519 (256 bits) | Ed448 (448 bits) | Built-in |

**Prohibited Algorithms**:
- ❌ MD5 with RSA (`md5WithRSAEncryption`)
- ❌ SHA-1 with RSA (`sha1WithRSAEncryption`)
- ❌ Any algorithm using MD2, MD4, MD5, SHA-1
- ❌ RSA < 2048 bits

#### Validity Periods

| Certificate Type | Maximum Validity | Notes |
|------------------|------------------|-------|
| Subscriber Certificates | 398 days | Post-2020-09-01 |
| Intermediate CA Certificates | 5-10 years | Typical industry practice |
| Root CA Certificates | 20-25 years | Typical industry practice |

**MUST NOT exceed maximum validity periods per Baseline Requirements**

### Required Certificate Extensions (Section 7.1.2)

**Mandatory Extensions for Subscriber Certificates**:

1. **Key Usage** (MUST be marked critical):
   - TLS Server: Digital Signature, Key Encipherment
   - CA: Certificate Sign, CRL Sign

2. **Extended Key Usage** (MUST be present):
   - TLS Server: `id-kp-serverAuth` (1.3.6.1.5.5.7.3.1)
   - TLS Client: `id-kp-clientAuth` (1.3.6.1.5.5.7.3.2)

3. **Basic Constraints** (MUST be marked critical for CA certificates):
   - CA: TRUE
   - Path Length Constraint: Appropriate value

4. **Subject Alternative Name (SAN)** (MUST be present for subscriber certificates):
   - DNS names, IP addresses, or other identifiers
   - MUST include all domain names covered by certificate

5. **Authority Key Identifier** (MUST be present)

6. **Subject Key Identifier** (MUST be present)

7. **CRL Distribution Points** (MUST be present if CRL supported)

8. **Authority Information Access** (MUST be present):
   - OCSP responder URI
   - CA Issuers URI

### Subject and Issuer Name Encoding (Section 7.1.4)

**Distinguished Name (DN) Requirements**:

| Component | Description | Format |
|-----------|-------------|--------|
| Country (C) | ISO 3166-1 alpha-2 code | 2 characters (e.g., "US") |
| Organization (O) | Legal organization name | Full legal name |
| Common Name (CN) | FQDN for TLS server | Deprecated in favor of SAN |
| State/Province (ST) | Full name | No abbreviations |
| Locality (L) | City or locality | Full name |

**Encoding Rules**:
- UTF-8 encoding for all name components
- PrintableString or UTF8String encoding types
- NO special characters unless properly escaped
- NO leading or trailing whitespace

### Signature Algorithms (Section 7.1.3.2)

**Approved Algorithms**:
- **RSA with SHA-256/384/512**: `sha256WithRSAEncryption`, `sha384WithRSAEncryption`, `sha512WithRSAEncryption`
- **ECDSA with SHA-256/384/512**: `ecdsa-with-SHA256`, `ecdsa-with-SHA384`, `ecdsa-with-SHA512`
- **EdDSA**: `id-Ed25519`, `id-Ed448`

**Prohibited Algorithms**:
- ❌ MD5 with RSA
- ❌ SHA-1 with RSA
- ❌ Any algorithm using MD2, MD4, MD5, SHA-1

### CRL and OCSP Profile Requirements

#### CRL Requirements (Section 7.2)

- **CRL Distribution Points**: MUST be present in subscriber certificates
- **Update Frequency**: Maximum 7 days for subscriber CRLs
- **Next Update Field**: MUST be present
- **CRL Number**: MUST be present and monotonically increasing
- **Revocation Reason**: SHOULD be included
- **CRL Signing Key**: MUST match Authority Key Identifier

#### OCSP Requirements (Section 7.3)

- **OCSP Responder URI**: MUST be present in Authority Information Access extension
- **Response Validity**: Maximum 7 days (10 days for OCSP responses with nextUpdate)
- **Response Signing**: MUST be signed by authorized responder
- **Revocation Status**: Must return `good`, `revoked`, or `unknown`
- **Nonce Extension**: SHOULD be supported to prevent replay attacks

### Audit Logging Requirements (Section 5.4.1)

**Mandatory Logging Events**:
- Certificate issuance requests and approvals
- Certificate revocations (with reason code)
- Certificate renewals and re-keys
- Key generation and destruction events
- Access to CA private keys
- Configuration changes to CA systems
- Security events (login failures, unauthorized access attempts)

**Log Retention**:
- **Minimum**: 7 years after certificate expiration
- **Audit Trail**: Tamper-evident, append-only logs
- **Access Control**: Restricted to authorized personnel
- **Backup**: Regular backups with offsite storage

### Validation Requirements (Section 3.2.2)

#### Domain Validation (DV)

**Methods**:
- DNS-based validation (DNS TXT/CNAME records)
- HTTP-based validation (`.well-known/pki-validation/`)
- Email-based validation (admin@domain, postmaster@domain)

**Validation MUST occur within 30 days of issuance**

#### Organization Validation (OV)

**Requirements**:
- Legal existence verification (government records, third-party databases)
- Operational existence verification (telephone, physical address)
- Domain control validation (same as DV)

**Validation MUST occur within 13 months of issuance**

#### Extended Validation (EV)

**Requirements**:
- All OV requirements
- Enhanced identity verification (legal opinion, accountant letter)
- Physical address verification (site visit or reliable database)
- Operational presence verification (telephone directory, government records)

**Validation MUST occur within 13 months of issuance**

---

## CA Architecture Patterns

### TLS Issuing CA Configurations

**Examples in order of highest to lowest preference**:

#### 1. Offline Root CA → Online Root CA → Online Issuing CA (Recommended)

```
┌─────────────────┐
│ Offline Root CA │ ← Air-gapped, signs Online Root CA
└────────┬────────┘
         │ Signs
         ▼
┌─────────────────┐
│ Online Root CA  │ ← Signs Issuing CA certificates
└────────┬────────┘
         │ Signs
         ▼
┌─────────────────┐
│ Online Issuing CA│ ← Signs subscriber certificates
└─────────────────┘
```

**Rationale**: Maximum security - offline root compromise requires physical access

#### 2. Online Root CA → Online Issuing CA (Balanced)

```
┌─────────────────┐
│ Online Root CA  │ ← Signs Issuing CA certificates
└────────┬────────┘
         │ Signs
         ▼
┌─────────────────┐
│ Online Issuing CA│ ← Signs subscriber certificates
└─────────────────┘
```

**Rationale**: Simpler operations, acceptable security for most deployments

#### 3. Online Root CA (Simple)

```
┌─────────────────┐
│ Online Root CA  │ ← Directly signs subscriber certificates
└─────────────────┘
```

**Rationale**: Simplest configuration, acceptable for development/testing

#### 4. Online Root CA → Policy Root CA → Online Issuing CA (Policy-Driven)

**Rationale**: Supports multiple certificate policies (DV, OV, EV)

#### 5. Offline Root CA → Online Root CA → Policy CA → Online Issuing CA (Enterprise)

**Rationale**: Maximum security + policy flexibility

### Certificate Lifecycle Management

#### Issuance Workflow

1. Certificate request validation (CSR format, key strength, domain ownership)
2. Subscriber identity verification (per validation level: DV/OV/EV)
3. Certificate generation (apply policy, populate extensions)
4. Certificate signing (use appropriate CA private key)
5. Certificate publication (delivery to subscriber, LDAP/HTTP repository)

#### Renewal Workflow

1. Pre-expiration notification (60/30/7 days before expiration)
2. Renewal request validation (verify subscriber still controls domain)
3. Re-validation if required (13-month rule for OV/EV)
4. New certificate issuance (new serial number, updated validity period)
5. Old certificate retention (keep in archive for audit trail)

#### Revocation Workflow

1. Revocation request validation (verify requester authority)
2. Revocation reason determination (key compromise, superseded, cessation of operation)
3. Certificate status update (mark as revoked in CA database)
4. CRL/OCSP update (publish revocation status within SLA)
5. Notification (alert subscriber and relying parties)

### Key Ceremony Best Practices

**Root CA Key Generation**:
- Multi-person control (minimum 3 custodians)
- Ceremony script (documented step-by-step procedure)
- Witnessed ceremony (independent observers, video recording)
- Split knowledge (key shares distributed to custodians)
- Secure storage (hardware security module or offline vault)

**Key Backup and Escrow**:
- Encrypted backups (split across multiple locations)
- Tamper-evident packaging (sealed envelopes, cryptographic checksums)
- Dual control access (require multiple custodians for recovery)
- Regular testing (verify backups are valid and recoverable)

**Key Destruction**:
- Multi-person authorization (require multiple approvals)
- Secure deletion (cryptographic erasure, physical destruction)
- Audit trail (log destruction event with witnesses)
- Certificate revocation (revoke all certificates signed by destroyed key)

---

## Certificate Transparency (CT)

**MANDATORY for Public CAs**

**CT Log Submission**:
- **Pre-certificates**: Submit before final certificate issuance
- **SCT Embedding**: Embed Signed Certificate Timestamps (SCTs) in certificates
- **SCT Count**: Minimum 2 SCTs from different CT log operators
- **Log Diversity**: Use logs operated by different entities

**CT Monitoring**:
- Monitor CT logs for mis-issuance
- Detect unauthorized certificates for your domains
- Respond to incidents within 24 hours
- Revoke mis-issued certificates immediately

---

## OCSP Stapling and Must-Staple

**OCSP Stapling**:
- TLS server fetches OCSP response from CA
- Server includes OCSP response in TLS handshake
- Reduces client-side OCSP queries
- Improves performance and privacy

**Must-Staple Extension**:
- TLS Feature extension (`id-pe-tlsfeature`, OID 1.3.6.1.5.5.7.1.24)
- Value: `status_request` (5)
- Forces TLS server to staple OCSP response
- Client MUST reject connection if OCSP response missing

**Configuration Example**:

```yaml
certificate:
  extensions:
    tls_feature:
      enabled: true
      must_staple: true  # Forces OCSP stapling
```

---

## Certificate Pinning (Deprecated)

**Public Key Pinning (HPKP)** - ❌ DEPRECATED (DO NOT USE):
- Deprecated due to operational risks
- Replaced by Certificate Transparency
- Can cause catastrophic failures (pin-to-brick)

**Alternatives**:
- Certificate Transparency monitoring
- CAA DNS records (restrict authorized CAs)
- Expect-CT header (enforce CT compliance)

---

## CAA DNS Records - RECOMMENDED

**Certification Authority Authorization (CAA)**:
- DNS record type (CAA RRs, RFC 8659)
- Restricts which CAs can issue certificates for domain
- Reduces risk of mis-issuance by unauthorized CAs

**Example CAA Records**:

```dns
example.com. CAA 0 issue "ca.example.com"
example.com. CAA 0 issuewild "ca.example.com"
example.com. CAA 0 iodef "mailto:security@example.com"
```

**CAA Checking**:
- CAs MUST check CAA records before issuance
- Check full domain hierarchy (example.com, sub.example.com, subsub.sub.example.com)
- Reject issuance if CAA prohibits CA

---

## Compliance Summary

**Subscriber Certificates MUST**:
- ✅ Serial number ≥64 bits from CSPRNG, non-sequential, >0, <2^159
- ✅ Key size: RSA ≥2048 bits, ECDSA P-256/384/521, EdDSA
- ✅ Hash algorithm: SHA-256/384/512 (NEVER MD5/SHA-1)
- ✅ Validity: ≤398 days (post-2020-09-01)
- ✅ Extensions: Key Usage (critical), Extended Key Usage, SAN, AKI, SKI
- ✅ CRL/OCSP: Distribution points present, update ≤7 days
- ✅ Audit logs: 7-year retention, tamper-evident

**CA Certificates MUST**:
- ✅ Basic Constraints: CA=TRUE, path length constraint
- ✅ Key Usage: Certificate Sign, CRL Sign (critical)
- ✅ Validity: Appropriate for CA tier (root: 20-25y, intermediate: 5-10y)
- ✅ AKI/SKI: Present for certificate chain validation

**Validation MUST**:
- ✅ Domain control: Validated within 30 days of issuance
- ✅ Organization identity: Validated within 13 months (OV/EV)
- ✅ CAA records: Checked before issuance
