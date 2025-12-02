# CA (Certificate Authority) Product

This directory contains the Certificate Authority (CA) product implementation.

## Current Status

**Note**: CA functionality is planned but not yet implemented.
This directory is a placeholder for the future CA product as part of the 4-products architecture.

## Planned Structure

```text
internal/ca/
├── application/     # Server application lifecycle
├── config/          # CA-specific configuration
├── crypto/          # Crypto provider interface
│   ├── memory/      # In-memory key storage
│   ├── filesystem/  # File-based key storage
│   └── hsm/         # HSM integration stubs
├── handler/         # HTTP handlers
├── middleware/      # HTTP middleware
├── profile/         # Certificate profiles
│   ├── subject/     # Subject template resolution
│   └── certificate/ # Certificate policy rendering
├── repository/      # Data repository
├── revocation/      # CRL and OCSP services
│   ├── crl/         # CRL generation
│   └── ocsp/        # OCSP responder
└── service/         # CA business logic
```

## Compliance Requirements

### Standards

- **RFC 5280** - X.509 PKI Certificate and CRL Profile
- **CA/Browser Forum** - Baseline Requirements for TLS Server Certificates
- **ACME Protocol** (RFC 8555) - Automated Certificate Management Environment

### Certificate Serial Numbers

- Minimum 64 bits CSPRNG
- Non-sequential
- Greater than 0
- Less than 2^159

### Validity Periods

- Maximum 398 days for subscriber certificates (post-2020-09-01)

### Algorithms (FIPS 140-3 Compliant)

- RSA ≥ 2048 bits
- ECDSA P-256, P-384, P-521
- EdDSA (Ed25519)

## Planned Certificate Profiles

1. Root CA
2. Intermediate CA
3. Issuing CA
4. TLS Server
5. TLS Client
6. Code Signing
7. Email (S/MIME)
8. VPN
9. Device
10. Timestamp Authority
11. OCSP Responder
12. Cross-Certificate
13. CA Repository
14. Delta CRL
15. User Authentication
16. Smart Card Logon
17. Document Signing
18. EV TLS Server
19. OV TLS Server
20. DV TLS Server
21. Wildcard TLS
22. SAN (Subject Alternative Name)
23. Key Escrow
24. Key Recovery
25. Attribute Certificate

## Phase 4 Deliverables

Per the implementation plan, Phase 4 includes:

1. Domain Charter
2. Configuration Schema
3. Crypto Provider Abstractions
4. Subject Profile Engine
5. Certificate Profile Engine
6. Root CA Bootstrap
7. Intermediate CA Provisioning
8. Issuing CA Lifecycle
9. Enrollment API
10. Revocation Services (CRL, OCSP)
