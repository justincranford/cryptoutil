# Deep Research: Service Architecture & Product Taxonomy

**Created**: 2026-02-23
**Purpose**: Exhaustive analysis of all possible service groupings, product boundaries, and adjacency gaps for cryptoutil

---

## 1. Current Service Domain Analysis

| Service | Domain | Primary Data | External Consumers | LOC |
|---------|--------|-------------|--------------------|----|
| sm-kms | Key Material | Elastic key rings (AES/RSA), encrypted with barrier | Other services needing encrypt/sign | 9,536 |
| sm-im | Messaging | JWE-encrypted messages, per-recipient JWKs | End users (plaintext send/receive) | 2,309 |
| jose-ja | JOSE Protocol | Elastic JWK rings (RSA/EC), material rotation, audit logs | Services needing JWT/JWS/JWE via API | 4,406 |
| pki-ca | Certificates | X.509 certs, CRLs, OCSP (in-memory) | TLS clients, cert requestors | 11,418 |
| identity-* | Auth/Authz | Users, sessions, realms, tenants, MFA | End users, service-to-service auth | ~18K |
| template/barrier | Infrastructure | Unseal/Root/Intermediate/Content keys | Internal to each service | 322 files |

### 1.1 What the Barrier Layer Actually Provides

The barrier layer is INTERNAL to EACH service. It is NOT accessible via HTTP (no external API). It provides:
- **Unseal key**: derived from Docker secrets via HKDF; never stored
- **Root key**: encrypted with unseal key; rotated annually
- **Intermediate keys**: encrypted with root key; rotated quarterly
- **Content keys**: encrypted with intermediate keys; rotated per-operation

**Critical distinction**: The barrier handles disk-level encryption for each service's own data. It is not a cross-service key authority. Jose-ja's barrier protects jose-ja's own JWK storage. sm-kms's barrier protects sm-kms's own key storage.

### 1.2 The jose-ja / sm-kms Overlap Problem

Both services implement the same "elastic key ring" concept:

| Capability | sm-kms | jose-ja |
|------------|--------|---------|
| Elastic key ring management | ✅ (AES + RSA) | ✅ (RSA + EC only) |
| Sign payload | ✅ PostSignByElasticKeyID | ✅ JWSService.Sign |
| Verify signature | ✅ PostVerifyByElasticKeyID | ✅ JWSService.Verify |
| Encrypt plaintext | ✅ PostEncryptByElasticKeyID | ✅ JWEService.Encrypt |
| Decrypt ciphertext | ✅ PostDecryptByElasticKeyID | ✅ JWEService.Decrypt |
| Barrier-protected storage | ✅ | ✅ |
| JWT create/validate | ❌ | ✅ JWTService |
| JWKS endpoint (public keys) | ❌ | ✅ JWKSService |
| Material rotation scheduling | ❌ | ✅ MaterialRotationService |
| Audit log | ❌ | ✅ AuditLogService |

**Finding**: sm-kms and jose-ja are nearly duplicate at the business logic level. They differ in:
1. jose-ja provides TOKEN-oriented operations (JWT, JWKS discovery endpoint)
2. jose-ja has richer audit and rotation features
3. sm-kms was historically focused on AES/symmetric keys; jose-ja on asymmetric/JOSE RFCs

### 1.3 Inter-Service Dependency Map (Current)

```
identity-* ──(JWT validation)──→ jose-ja (JWKS endpoint for token signing keys)
sm-im ────(JWE per-message)──→ barrier (internal; jose-ja not used)
pki-ca ────────(cert signing)────→ pki-ca internal crypto (X.509/ASN.1; jose-ja not used)
sm-kms ────────(encryption)──────→ barrier (internal; jose-ja not used)

barrier ─────────────────────────→ independent per service (no cross-service calls)
```

**Finding**: Currently, only identity-* depends on jose-ja externally. sm-im, pki-ca, and sm-kms all handle crypto internally via barrier. This is intentional but means jose-ja is only serving 1 of 8 non-template services.

---

## 2. Product Taxonomy Schemas

### Schema A: "Vault-Inspired" — Two Products

Inspired by HashiCorp Vault: one platform for all non-identity crypto operations.

```
Product 1: SM (Secure Materials Platform)
  ├── sm-kms   (symmetric key management, encrypt/decrypt)
  ├── sm-im    (encrypted messaging)*
  ├── sm-jwk   (elastic JWK authority = jose-ja renamed)
  └── sm-pki   (certificate authority = pki-ca renamed)

Product 2: Identity
  ├── identity-authz
  ├── identity-idp
  ├── identity-rp
  ├── identity-rs
  └── identity-spa
```

*sm-im moved to SM product as sm-im

**Pros**: Maximum cohesion; all crypto lives in SM; similar to Vault architecture; easy to reason about ("if it's crypto, it's SM")
**Cons**: SM becomes a very wide product (4 services spanning keys, messages, certs, JWKs); CA/BF compliance may require PKI isolation
**Services**: 9 (unchanged), products: 2 (from 5)
**Rating**: ⭐⭐⭐⭐ (if PKI isolation requirements allow)

---

### Schema B: "Three Products" — Crypto, PKI, Identity

```
Product 1: Crypto (all non-PKI non-identity crypto)
  ├── crypto-kms   (symmetric keys + AES operations)
  ├── crypto-im    (encrypted messaging)
  └── crypto-jwk   (JWK authority + JOSE operations)

Product 2: PKI (certificate authority, standalone for CA/BF compliance)
  └── pki-ca

Product 3: Identity
  └── identity-*
```

**Pros**: PKI isolation respected (CA/BF compliance argument); consolidates remaining crypto
**Cons**: "Crypto" product spans very different concerns (messaging vs key mgmt vs JWK authority)
**Services**: 8 (jose-ja renamed, sm-im moved)
**Rating**: ⭐⭐⭐

---

### Schema C: "Four Products — SM, JOSE, PKI, Identity" (current + sm-im addition)

```
SM:       sm-kms, sm-im (sm-im moved here)
JOSE:     jose-ja
PKI:      pki-ca
Identity: identity-*
```

**Comment on former Cipher product**: If sm-im moves to SM, the Cipher product disappears. JOSE and PKI remain separate. This is the minimal change from current state.
**Pros**: Former Cipher product eliminated (never had more than 1 service); SM now has 2 cohesive services (key mgmt + encrypted messaging); other products unchanged
**Cons**: JOSE and SM overlap (both manage elastic key rings); jose-ja purpose is unclear when sm-kms also provides sign/encrypt
**Services**: 9 (same), products: 4 (from 5, former Cipher eliminated)
**Rating**: ⭐⭐⭐⭐ (clean, minimal disruption)

---

### Schema D: "SM + JOSE Merged" — Three Products

```
SM (unified secret materials + JOSE operations):
  ├── sm-kms     (symmetric/AES key management + encrypt/decrypt)
  ├── sm-jwk     (jose-ja merged here: elastic JWK authority, JWT, JWKS, JWS, JWE)
  └── sm-im      (sm-im moved here: encrypted messaging)

PKI: pki-ca
Identity: identity-*
```

**Key argument**: sm-kms and jose-ja are both "key material management" — combining their PRODUCT reduces confusion about which service to use for a given crypto operation.
**Differentiation**: sm-kms handles AES symmetric encryption and key-as-material (other services call to encrypt data). sm-jwk handles asymmetric JWK lifecycle for token/signature purposes.
**Pros**: Eliminates the sm-kms/jose-ja overlap confusion; SM becomes the clear home for all key-related operations; PKI and Identity remain separate
**Cons**: SM product is wider (3 services); jose-ja rename required
**Services**: 9 (same), products: 3
**Rating**: ⭐⭐⭐⭐

---

### Schema E: "Collapse sm-kms into jose-ja" — Single Key Service

```
jose-ja (absorbed sm-kms functionality, becomes "Key Authority"):
  - ElasticKey management (symmetric AES + asymmetric RSA/EC)
  - Encrypt/decrypt (AES via elastic key)
  - Sign/verify (via elastic JWK)
  - JWT/JWKS
  - Barrier-protected storage

Products: JA, PKI, SM-IM, Identity
  ├── ja (merged jose-ja + sm-kms)
  ├── pki-ca
  ├── sm-im (sm-im moved here, standalone product)
  └── identity-*
```

**Pros**: Eliminates duplicate elastic key ring implementations; single API surface for all key operations; consumers query one service for all crypto needs
**Cons**: jose-ja is already PARTIALLY migrated and has critical TODOs; merging sm-kms into it is high effort; sm-kms is the LAST in migration priority; creating a very large service
**Services**: 8 (sm-kms absorbed into ja, sm-im becomes sm-im standalone)
**Rating**: ⭐⭐⭐ (architecturally elegant but high execution risk)

---

### Schema F: "Protocol-Aligned Microservices" — Granular Split

Break each service into its protocol-specific concerns:

```
Key Primitives:
  ├── kp-aes    (AES key generation, AES encrypt/decrypt)
  ├── kp-rsa    (RSA key generation, RSA sign/verify/encrypt)
  └── kp-ecdsa  (EC key generation, ECDSA sign/verify, ECDH)

JOSE Standard:
  ├── jose-jwk  (JWK CRUD, elastic key ring management)
  ├── jose-jwt  (JWT create/validate)
  └── jose-jwe  (JWE encrypt/decrypt via key reference)

PKI Standard:
  ├── pki-ca    (cert issuance)
  ├── pki-ocsp  (OCSP responder, standalone)
  └── pki-tsp   (timestamp authority, standalone)

Applications:
  └── sm-im     (encrypted messaging)

Identity:
  └── identity-*
```

**Pros**: Maximum single-responsibility; each service is small and focused; independent scaling per protocol
**Cons**: Explosion of services (13+); massive operational complexity; clients must call multiple services; cross-Service auth between key primitives and jose adds latency
**Rating**: ⭐⭐ (too granular for a platform product; better as internal library design)

---

### Schema G: "Domain-Driven Bounded Contexts"

```
Bounded Context 1: Key Lifecycle Management
  → sm-kms (AES key lifecycle) + jose-ja (JWK lifecycle)
  → Merge into: key-authority service
  → Responsibilities: ALL key material (symmetric + asymmetric), elastic ring, sign/verify/encrypt/decrypt

Bounded Context 2: Certificate Lifecycle Management
  → pki-ca
  → Responsibilities: X.509 issuance, revocation, CRL, OCSP, TSP

Bounded Context 3: Secure Communication
  → sm-im (current) + potential sm-group, sm-file
  → Responsibilities: End-to-end encrypted data transfer

Bounded Context 4: Identity & Access Management
  → identity-*
  → Responsibilities: AuthN, AuthZ, sessions, OAuth2/OIDC
```

3 non-identity products, 1 merged key service + 1 PKI + 1 communication = more coherent than current
**Rating**: ⭐⭐⭐⭐

---

### Schema H: "Consumer-First" — Who Calls What

Group by who the primary consumer is:

```
Service-to-Service Crypto Layer:
  → sm-kms (other services call to encrypt their data)
  → jose-ja (identity services call for JWT signing keys)
  → pki-ca (all services need TLS certs)

User-Facing Data Layer:
  → sm-im (end users send/receive messages)
  → identity-* (end users authenticate)
```

This is more of an analysis tool than a product boundary. Both groups span multiple current products.
**Rating**: Framework for analysis, not a product schema.

---

## 3. The jose-ja Role: Three Possible Positions

Given that jose-ja's functionality is currently only consumed externally by identity-*, there are three strategic positions:

### Position 1: Jose-ja as a SHARED INFRASTRUCTURE service
Like a service mesh crypto layer — all services call jose-ja for JWK operations instead of managing their own keys.
- Pros: Centralized key management, single audit trail, consistent rotation across all tenants
- Cons: Single point of failure; latency for every crypto op; all 8 services become dependent on jose-ja; jose-ja must never go down

### Position 2: Jose-ja as IDENTITY-ADJACENT (current)
Jose-ja exists primarily to serve identity services (JWT signing key management and JWKS endpoint discovery for token validation).
- Pros: Limited blast radius; jose-ja can be co-deployed with identity services
- Cons: sm-kms duplicates its functionality; confusing to have two "key management" services

### Position 3: Jose-ja EMBEDDED within sm-kms
Jose-ja's functionality becomes part of a unified key authority service. No separate jose-ja process.
- Pros: Eliminates duplicate elastic key ring; single key authority
- Cons: High migration effort; sm-kms is already the most complex service

**Current best choice**: Position 2 (keep separated) until sm-kms migration debt is resolved, then evaluate Position 3.

---

## 4. sm-im Product Placement Analysis

### Why sm-im in former "Cipher" product doesn't work well
The "Cipher" product name suggests a protocol (encryption algorithm), not a use case. Having only one service (sm-im = instant messenger) in a "Cipher" product is an odd match. 

### Why sm-im fits better in "SM" (Secrets/Secure Materials)
- SM = "Secure Materials" (keys + protected data)
- sm-im stores PROTECTED MESSAGES (JWE-encrypted data at rest)
- sm-kms stores PROTECTED KEYS
- Both services serve the overarching goal of "keeping sensitive material secure"
- Industry analogy: AWS has KMS (keys) + end-to-end encrypted services under the same security brand

### sm-im standalone vs sm-kms merged

| Dimension | sm-im (standalone) | merged into sm-kms |
|-----------|-------------------|-------------------|
| Data model | Messages + RecipientJWKs (separate domain) | Combined with key ring tables |
| Scaling | Independent (message volume-driven) | Locked to KMS scaling |
| Blast radius | IM outage doesn't affect KMS | Single failure takes both |
| API surface | Clean, purpose-focused | Bloated (keys + messages = very different APIs) |
| Code size | 2,309 LOC (manageable) | ~12K LOC combined (getting large) |
| Multi-tenancy | Per-tenant message isolation | Per-tenant key isolation (same pattern) |
| Migration risk | Low (just rename/move) | High (merge GORM models, route conflicts) |
| Maintenance | Independent deployability | Deploy both for any change |

**Recommendation**: sm-im as STANDALONE service under SM product (Schema C or D). Do NOT merge into sm-kms.

---

## 5. Adjacent/Missing Services Catalog

### 5.1 Secret Storage (Static Secrets) — HIGH PRIORITY MISSING
**sm-secrets** (or vault-kv)
- **What**: Key-value store for static secrets (DB passwords, API keys, connection strings, TLS private keys)
- **Why missing**: sm-kms manages ENCRYPTION KEY MATERIAL, not arbitrary secret values. There is currently no place to store "my PostgresDB password" securely.
- **Precedent**: HashiCorp Vault KV v2, AWS Secrets Manager (distinct from KMS), Azure Key Vault Secrets
- **API**: GET/SET/DELETE/VERSION secret; automatic rotation; access policies; audit log
- **Multi-tenant**: per-tenant secret namespaces, barrier-encrypted
- **Priority**: HIGH — this is a common first use case for any secrets platform

### 5.2 SSH Certificate Authority — MEDIUM PRIORITY MISSING
**sm-ssh** (or pki-ssh)
- **What**: Issues short-lived SSH certificates for passwordless machine authentication
- **Why useful**: SSH key management is operationally painful; short-lived certs solve access revocation
- **Precedent**: HashiCorp Vault SSH, BeyondCorp, Teleport
- **API**: Sign SSH public key → short-lived cert; define principals and TTL
- **Depends on**: pki-ca or standalone CA

### 5.3 PKI Registration Authority — IN pki-ca README (unimplemented)
**pki-ra** (Registration Authority)
- **What**: Validates certificate requests before forwarding to issuing CA
- **Why**: Separates validation logic (RA) from signing (CA) — important for large deployments
- **Status in pki-ca**: README Task 16 "RA Workflows" marked as TODO

### 5.4 OCSP Responder — IN pki-ca README (unimplemented)
**pki-ocsp** (Standalone OCSP Responder)
- **What**: Responds to certificate revocation status queries
- **Why standalone**: High-availability requirement (must be always-on), different scaling from CA
- **Status in pki-ca**: Currently embedded in pki-ca, README says make standalone

### 5.5 Timestamp Authority — IN pki-ca README (unimplemented)
**pki-tsp** (Timestamp Authority)
- **What**: RFC 3161 timestamps for document signing
- **Status in pki-ca**: README Task 18 "TSA Full Implementation" marked as TODO

### 5.6 ACME Certificate Protocol — MISSING
**pki-acme** 
- **What**: RFC 8555 ACME protocol (Let's Encrypt compatible)
- **Why**: Enables automated certificate issuance and renewal for clients
- **Depends on**: pki-ca

### 5.7 Encrypted Group Messaging — MISSING
**sm-group** (or sm-group if sm-im moves to SM)
- **What**: Group encrypted channels (1:N persistent, not just per-message recipients)
- **Why**: sm-im is designed for 1:N per-message; persistent groups require group key management
- **Extends**: sm-im domain

### 5.8 Encrypted File Storage — MISSING
**sm-file** (or sm-file)
- **What**: Encrypted blob/file storage with per-tenant keys and access control
- **Why**: Natural extension of the encryption platform — not just messages but files
- **Precedent**: AWS S3 with KMS server-side encryption, but as a standalone service

### 5.9 Token Service (STS) — PARTIALLY IN identity
**identity-sts** (or jose-sts if jose-ja is SMS-adjacent)
- **What**: Short-lived credential issuance (temporary access tokens, AWS STS equivalent)
- **Why**: Enables service-to-service auth with scoped, time-limited tokens
- **Status**: Currently embedded in identity-authz; could be standalone

### 5.10 Centralized Audit Log — MISSING
**audit-server**
- **What**: WORM (write-once-read-many) audit log storage across all services
- **Why**: Currently each service logs independently; compliance requires centralized immutable audit
- **API**: Append event, query by tenant/service/time; immutable (no delete)
- **Priority**: HIGH — needed for CA/BF compliance (7-year cert audit retention)

### 5.11 Secret Scanning Service — MISSING
**audit-scanner** (or cicd-scanner)
- **What**: Scans code/configs/containers for leaked secrets
- **Why**: Currently gosec in golangci-lint does static analysis; no runtime scanning
- **Precedent**: GitLeaks (already in CI), but a persistent service-side scanner

### 5.12 HSM Gateway — ADVANCED
**sm-hsm** or **hsm-gateway**
- **What**: Abstract interface to hardware security modules (AWS CloudHSM, Azure HSM, Thales, SoftHSM2)
- **Why**: FIPS 140-3 Level 3 compliance requires HSM for key storage in some contexts
- **Impact**: sm-kms, jose-ja, pki-ca could all route key operations through hsm-gateway

### 5.13 Directory/SCIM Service — MISSING FROM IDENTITY
**identity-directory** or **identity-scim**
- **What**: LDAP/SCIM 2.0 user provisioning and directory service
- **Why**: Currently identity services manage users internally; enterprise deployments need LDAP integration
- **Precedent**: Keycloak has LDAP federation; Auth0 has connections

### 5.14 Backup/Disaster Recovery Service — MISSING
**sm-backup**
- **What**: Encrypted backup of barrier keys, certificates, and secrets
- **Why**: If unseal key is lost, all data is unrecoverable; disaster recovery needs structured backup
- **Complexity**: Must handle key material without exposing it

---

## 6. Industry Reference Architecture Comparisons

### HashiCorp Vault
| Vault Component | cryptoutil Equivalent | Status |
|----------------|----------------------|--------|
| Transit (encrypt/decrypt) | sm-kms | ✅ exists |
| PKI Secrets Engine | pki-ca | ✅ partial |
| KV v2 (static secrets) | **MISSING** | ❌ |
| SSH Secrets Engine | **MISSING** | ❌ |
| JWT/OIDC Auth | jose-ja + identity | ✅/partial |
| AWS/Azure/GCP Auth | **MISSING** | ❌ |
| AppRole Auth | identity-* | ✅ partial |
| Seal/Unseal | barrier (per-service) | ✅ |
| Audit Backends | logging (per-service) | partial |

### AWS Security Services
| AWS Service | cryptoutil Equivalent | Status |
|-------------|----------------------|--------|
| KMS | sm-kms | ✅ exists |
| Secrets Manager (static) | **MISSING** | ❌ |
| ACM (cert mgmt) | pki-ca | ✅ partial |
| Cognito (identity) | identity-* | ✅ partial |
| STS (token service) | identity-authz | ✅ partial |
| CloudHSM | **MISSING** | ❌ |
| End-to-end encryption | sm-im | ✅ |

### Google BeyondCorp / Zero Trust
| BeyondCorp Component | cryptoutil Equivalent | Status |
|--------------------|-----------------------|--------|
| Certificate Authority Service | pki-ca | ✅ partial |
| Cloud KMS | sm-kms | ✅ |
| IAM | identity-* | ✅ partial |
| Access Context Manager | **MISSING** | ❌ |

---

## 7. Recommended Target Architecture (New Proposal)

Based on the analysis above, the following service grouping is recommended:

### Proposed Products (4 products, 11 services)

```
SM (Secure Materials):
  ├── sm-kms     (AES/symmetric key management + encrypt/decrypt/sign/verify)
  ├── sm-jwk     (jose-ja renamed: elastic JWK authority, JWT, JWKS, material rotation)
  ├── sm-im      (sm-im moved: encrypted messaging)
  └── sm-secrets (NEW: static secret KV store — HIGH PRIORITY)

PKI (Public Key Infrastructure):
  ├── pki-ca     (certificate authority — X.509 issuance, CRL)
  └── pki-ocsp   (FUTURE: standalone OCSP responder)

Identity:
  ├── identity-authz
  ├── identity-idp
  ├── identity-rp
  ├── identity-rs
  └── identity-spa

Audit (FUTURE, single service):
  └── audit-server   (centralized WORM audit log)
```

**Key changes from current**:
1. Former Cipher product → dissolved; sm-im becomes sm-im under SM
2. JOSE product → jose-ja renamed sm-jwk, moved under SM
3. SM now has coherent 3-service (or 4-service with sm-secrets) family
4. PKI stands alone (CA/BF compliance justification)
5. Identity unchanged
6. sm-secrets added (high-priority gap)
7. audit-server added (future, compliance gap)

**Why sm-jwk belongs in SM (not standalone JOSE product)**:
- sm-kms already provides sign/verify/encrypt/decrypt — sm-jwk is the RFC JOSE version of the same
- Both are "key material management" — natural product siblings
- SM becomes: "everything that manages sensitive material (keys, messages, secrets)"
- PKI is separate because X.509 certificates have distinct compliance and trust chain requirements

---

## 8. Decision Matrix: sm-im Placement

| Criterion | Stay in former Cipher | Move to SM-IM standalone | Merge into sm-kms |
|-----------|---------------|--------------------------|-------------------|
| Product coherence | ❌ (1-service product) | ✅ (SM = keys + messages + secrets) | ⚠️ (forces different domains together) |
| Independent scaling | ✅ | ✅ | ❌ |
| Blast radius isolation | ✅ | ✅ | ❌ |
| Migration effort | n/a | ~2h (rename/move) | ~8h (merge domain + routes) |
| Future extensibility | ❌ (Former Cipher = sm-im only?) | ✅ (can add sm-group, sm-file) | ❌ |
| CA/BF compliance impact | none | none | none |
| Port assignment change | none | yes (needs SM range) | no |
| **Recommendation** | ❌ | ✅ ⭐⭐⭐⭐⭐ | ⚠️ ⭐⭐ |

---

## 9. Decision Matrix: jose-ja Placement

| Criterion | Keep as JOSE product (jose-ja) | Move to SM product (sm-jwk) | Merge into sm-kms |
|-----------|--------------------------------|------------------------------|-------------------|
| Product coherence | ⚠️ (1-service product) | ✅ (SM = all key material) | ✅ (single key authority) |
| Service independence | ✅ | ✅ | ❌ |
| Eliminates sm-kms overlap | ❌ | ❌ | ✅ |
| Migration effort | n/a (just rename) | ~1h (rename/move) | ~20h (merge business logic) |
| Risk | none | low | high |
| Recommended timing | Now: rename sm-jwk? | With sm-im move | Only after sm-kms fully migrated |

---

## 10. Summary Recommendations

### Immediate (low effort, high value):
1. **Move sm-im → sm-im** (standalone service under SM product) — 2h rename/move
2. **Rename jose-ja → sm-jwk** and move to SM product — 1h rename/move
3. **Add sm-secrets** (new static secret KV service) — significant new feature but addresses the highest-priority gap

### Medium term (requires migration prerequisites):
4. **After sm-kms fully migrated**: evaluate merging sm-jwk into sm-kms as single KMS+JWK service
5. **Add pki-ocsp** as standalone service (once pki-ca is template-migrated)

### Long term:
6. **Add audit-server** (centralized WORM audit log)
7. **Add sm-ssh** (SSH certificate authority)
8. **Add identity-directory** (SCIM/LDAP federation)

### NOT recommended:
- Merging sm-im directly into sm-kms (different domains, different scaling, blast radius concern)
- Schema F (too granular, operational explosion)
- Schema A pure (PKI isolation concern for CA/BF compliance)
