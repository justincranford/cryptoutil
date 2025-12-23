# Authentication and Authorization Priority Matrix

**Document Purpose**: Deterministic priority ordering of authentication methods, authorization methods, and realm types for all 9 cryptoutil services

**Created**: 2025-12-22
**Source**: Analysis of `.specify/memory/constitution.md` and `specs/002-cryptoutil/spec.md`
**Context**: All 9 services implement dual HTTPS endpoints with `/browser/*` and `/service/*` request paths

---

## Service List (9 Total Services)

### Product Services (8 Core Services)

| Service Alias | Full Name | Product | Public Port | Admin Port |
|---------------|-----------|---------|-------------|------------|
| **sm-kms** | Secrets Manager - Key Management Service | P3: KMS | 8080-8089 | 9090 |
| **pki-ca** | Public Key Infrastructure - Certificate Authority | P4: CA | 8443-8449 | 9090 |
| **jose-ja** | JOSE - JWK Authority | P1: JOSE | 9443-9449 | 9090 |
| **identity-authz** | Identity - Authorization Server | P2: Identity | 18000-18009 | 9090 |
| **identity-idp** | Identity - Identity Provider | P2: Identity | 18100-18109 | 9090 |
| **identity-rs** | Identity - Resource Server | P2: Identity | 18200-18209 | 9090 |
| **identity-rp** | Identity - Relying Party | P2: Identity | 18300-18309 | 9090 |
| **identity-spa** | Identity - Single Page Application | P2: Identity | 18400-18409 | 9090 |

### Demonstration Service (1 Service)

| Service Alias | Full Name | Product | Public Port | Admin Port |
|---------------|-----------|---------|-------------|------------|
| **learn-ps** | Learn - Pet Store | Demo | 8888-8889 | 9090 |

---

## Authentication Methods Priority

**Deterministic Priority Order** (Higher priority = listed first):

### Browser-Based Clients (`/browser/*` paths)

#### Standalone Product Mode (Single Factor Authentication)

1. **Basic (Username/Password)** - HIGHEST priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: Most common authentication method, works without database
2. **Basic (Email/Password)** - HIGH priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: Email-based login for user-friendly authentication
3. **Bearer (API Token)** - MEDIUM priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: API key authentication for programmatic access

#### Federated Identity Mode (Multi-Factor Authentication)

1. **Passkey (WebAuthn with discoverable credentials)** - HIGHEST priority
   - Realm: Database (GORM/SQL) only (requires challenge persistence)
   - Rationale: FIDO2 standard, phishing-resistant, best user experience
2. **TOTP (Time-based One-Time Password)** - HIGH priority
   - Realm: Database (GORM/SQL) only (requires secret storage)
   - Rationale: RFC 6238, authenticator apps (Google Authenticator, Authy), widely adopted
3. **Hardware Security Keys (WebAuthn without passkeys)** - HIGH priority
   - Realm: Database (GORM/SQL) only (requires challenge persistence)
   - Rationale: FIDO U2F/FIDO2, phishing-resistant, enterprise security
4. **Basic (Username/Password)** - MEDIUM priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: Backward compatibility, disaster recovery access
5. **Basic (Email/Password)** - MEDIUM priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: User-friendly authentication
6. **Email OTP (One-Time Password via email)** - MEDIUM priority
   - Realm: Database (GORM/SQL) only (requires OTP persistence)
   - Rationale: Backup factor when other methods unavailable
7. **Recovery Codes (Pre-generated backup codes)** - MEDIUM priority
   - Realm: Database (GORM/SQL) only (requires code storage)
   - Rationale: Account recovery when primary factor unavailable
8. **Bearer (API Token)** - MEDIUM priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: API key authentication for programmatic access
9. **HTTPS Client Certificate** - MEDIUM priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: mTLS authentication, enterprise PKI integration
10. **Opaque OAuth 2.1 Access Token** - LOW priority
    - Realm: Database (GORM/SQL) only (token introspection required)
    - Rationale: Federation with external OAuth 2.1 providers
11. **JWE OAuth 2.1 Access Token** - LOW priority
    - Realm: Database (GORM/SQL) > File (YAML) (public key validation)
    - Rationale: Federation with external OAuth 2.1 providers, encrypted payload
12. **JWS OAuth 2.1 Access Token** - LOW priority
    - Realm: Database (GORM/SQL) > File (YAML) (public key validation)
    - Rationale: Federation with external OAuth 2.1 providers, signed payload
13. **Opaque OAuth 2.1 Refresh Token** - LOW priority
    - Realm: Database (GORM/SQL) only (token introspection required)
    - Rationale: Long-lived token refresh
14. **SMS OTP (NIST deprecated but MANDATORY)** - LOW priority
    - Realm: Database (GORM/SQL) only (OTP persistence, SMS delivery tracking)
    - Rationale: Backward compatibility, accessibility for non-technical users
15. **Phone Call OTP (NIST deprecated but MANDATORY)** - LOWEST priority
    - Realm: Database (GORM/SQL) only (OTP persistence, call delivery tracking)
    - Rationale: Backward compatibility, accessibility alternative
16. **Magic Link (Time-limited authentication link via email/SMS)** - LOWEST priority
    - Realm: Database (GORM/SQL) only (link token persistence)
    - Rationale: Passwordless alternative
17. **Push Notification (Mobile app push-based approval)** - LOWEST priority
    - Realm: Database (GORM/SQL) only (push token persistence, delivery tracking)
    - Rationale: Requires mobile app integration

### Headless-Based Clients (`/service/*` paths)

#### Standalone Product Mode (Single Factor Authentication)

1. **Basic (Client ID/Client Secret)** - HIGHEST priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: OAuth 2.1 Client Credentials flow standard, works without database
2. **Bearer (API Token)** - MEDIUM priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: API key authentication for service-to-service communication

#### Federated Identity Mode (Multi-Factor Authentication)

1. **HTTPS Client Certificate (mTLS)** - HIGHEST priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: Strongest authentication for service-to-service, no token management needed
2. **Basic (Client ID/Client Secret)** - HIGH priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: OAuth 2.1 Client Credentials flow standard
3. **Bearer (API Token)** - HIGH priority
   - Realm: File (YAML) > Database (GORM/SQL)
   - Rationale: API key authentication
4. **JWS OAuth 2.1 Access Token** - MEDIUM priority
   - Realm: Database (GORM/SQL) > File (YAML) (public key validation)
   - Rationale: Signed JWT tokens for federated services, stateless validation
5. **JWE OAuth 2.1 Access Token** - LOW priority
   - Realm: Database (GORM/SQL) > File (YAML) (decryption key required)
   - Rationale: Encrypted JWT tokens for sensitive payloads
6. **Opaque OAuth 2.1 Access Token** - LOW priority
   - Realm: Database (GORM/SQL) only (token introspection required)
   - Rationale: Federation with external OAuth 2.1 providers
7. **Opaque OAuth 2.1 Refresh Token** - LOWEST priority
   - Realm: Database (GORM/SQL) only (token introspection required)
   - Rationale: Long-lived token refresh for headless clients

---

## Authorization Methods Priority

**Deterministic Priority Order** (Higher priority = listed first):

### Browser-Based Clients (`/browser/*` paths)

1. **Scope-Based Authorization** - HIGHEST priority
   - Mechanism: Session token contains user's granted scopes (e.g., `read:keys`, `write:keys`)
   - Enforcement: Middleware checks scopes against endpoint requirements
   - Storage: Session store (Redis/database) contains scope list per session
   - Rationale: OAuth 2.1 standard, fine-grained access control
2. **Role-Based Access Control (RBAC)** - HIGH priority
   - Mechanism: Session token contains user's assigned roles (e.g., `admin`, `operator`, `viewer`)
   - Enforcement: Middleware checks roles against endpoint requirements
   - Storage: Database stores role assignments, session store caches roles
   - Rationale: Simplified administration, common enterprise pattern
3. **Resource-Level Access Control** - MEDIUM priority
   - Mechanism: Middleware checks user's ownership or permission for specific resource
   - Enforcement: Query database for resource ownership or explicit permissions
   - Storage: Database stores resource ownership (e.g., `user_id` foreign key)
   - Rationale: Fine-grained access control for multi-tenant deployments
4. **Consent Tracking** - MEDIUM priority
   - Mechanism: Session tracks which scopes user explicitly consented to
   - Enforcement: Middleware validates consent before granting access
   - Storage: Session store contains consent timestamp and scope list
   - Rationale: OAuth 2.1 requirement, GDPR compliance
5. **IP Allowlist** - LOW priority
   - Mechanism: Middleware checks client IP against allowlist (CIDR ranges)
   - Enforcement: Reject requests from non-allowlisted IPs
   - Storage: Configuration file (YAML) or database
   - Rationale: Additional security layer, compliance requirements

### Headless-Based Clients (`/service/*` paths)

1. **Scope-Based Authorization** - HIGHEST priority
   - Mechanism: JWT access token contains client's granted scopes in `scope` claim
   - Enforcement: Middleware validates JWT signature and checks scopes against endpoint requirements
   - Storage: JWT token (self-contained), JWKS for signature validation
   - Rationale: OAuth 2.1 standard, stateless validation, scalable
2. **Client-Level Access Control** - HIGH priority
   - Mechanism: JWT token contains `client_id` claim identifying service client
   - Enforcement: Middleware checks client_id against endpoint ACL
   - Storage: Configuration file (YAML) or database stores client ACLs
   - Rationale: Service-to-service authorization, audit trail
3. **IP Allowlist** - MEDIUM priority
   - Mechanism: Middleware checks client IP against allowlist (CIDR ranges)
   - Enforcement: Reject requests from non-allowlisted IPs
   - Storage: Configuration file (YAML) or database
   - Rationale: Network security layer, prevent unauthorized service access
4. **mTLS-Based Authorization** - MEDIUM priority
   - Mechanism: TLS client certificate validation, extract client_id from certificate Subject
   - Enforcement: Middleware validates certificate chain and checks client_id against ACL
   - Storage: CA certificates in configuration, client ACLs in database
   - Rationale: Strongest authentication AND authorization for service-to-service
5. **Rate Limiting** - LOW priority
   - Mechanism: Middleware tracks request count per client_id or IP
   - Enforcement: Reject requests exceeding rate limit (e.g., 1000 req/min)
   - Storage: In-memory rate limiter (e.g., Redis, in-process cache)
   - Rationale: Prevent abuse, ensure fair resource allocation

---

## Realm Types Priority

**Deterministic Priority Order** (File > Database for disaster recovery):

### File Realm Type (YAML Configuration)

**Priority**: HIGHEST (Disaster recovery, zero database dependency)

**Supported Authentication Methods**:

- ✅ Basic (Username/Password) - Hashed passwords in YAML
- ✅ Basic (Email/Password) - Hashed passwords in YAML
- ✅ Basic (Client ID/Client Secret) - Hashed secrets in YAML
- ✅ Bearer (API Token) - Hashed tokens in YAML
- ✅ HTTPS Client Certificate - Trusted CA certificates in YAML
- ✅ JWS/JWE OAuth 2.1 Access Token - Public keys in YAML (stateless validation)
- ❌ Passkey/WebAuthn - Requires database (challenge persistence)
- ❌ TOTP - Requires database (secret storage)
- ❌ Email/SMS OTP - Requires database (OTP persistence)
- ❌ Recovery Codes - Requires database (code storage)
- ❌ Magic Link - Requires database (link token persistence)
- ❌ Push Notification - Requires database (push token persistence)
- ❌ Opaque OAuth 2.1 tokens - Requires database (token introspection)

**Storage Format**:

```yaml
# Example file realm configuration (config.yaml)
authentication:
  file_realm:
    users:
      - username: admin
        email: admin@example.com
        password_hash: $pbkdf2-sha256$600000$salt$hash  # PBKDF2-HMAC-SHA256
        roles: [admin]
        scopes: [read:*, write:*]
      - username: operator
        email: operator@example.com
        password_hash: $pbkdf2-sha256$600000$salt$hash
        roles: [operator]
        scopes: [read:keys, write:keys]

    clients:
      - client_id: service-a
        client_secret_hash: $pbkdf2-sha256$600000$salt$hash
        scopes: [read:keys, encrypt:data]
      - client_id: service-b
        client_secret_hash: $pbkdf2-sha256$600000$salt$hash
        scopes: [read:certs, issue:cert]

    api_tokens:
      - token_hash: $pbkdf2-sha256$600000$salt$hash
        description: "CI/CD automation token"
        scopes: [read:keys]
        expires_at: "2026-12-31T23:59:59Z"
```

**Rationale**:

- **Disaster Recovery**: Database outage doesn't prevent admin access
- **Availability**: File-based authentication works without external dependencies
- **Simplicity**: No database migrations, no complex queries
- **Limitations**: Manual updates required, no dynamic user registration

### Database Realm Type (GORM/SQL)

**Priority**: MEDIUM (Production default, supports all authentication methods)

**Supported Authentication Methods**: ALL methods

**Storage Format**:

- PostgreSQL (production): Relational database with ACID guarantees
- SQLite (development/testing): Embedded database, zero-configuration

**Database Schema**:

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,  -- PBKDF2-HMAC-SHA256
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- User roles (many-to-many)
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id),
    role TEXT NOT NULL,
    PRIMARY KEY (user_id, role)
);

-- User scopes (many-to-many)
CREATE TABLE user_scopes (
    user_id UUID REFERENCES users(id),
    scope TEXT NOT NULL,
    PRIMARY KEY (user_id, scope)
);

-- TOTP secrets
CREATE TABLE totp_secrets (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    secret TEXT NOT NULL,  -- Encrypted with KMS
    created_at TIMESTAMP NOT NULL
);

-- WebAuthn credentials
CREATE TABLE webauthn_credentials (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    credential_id BYTEA UNIQUE NOT NULL,
    public_key BYTEA NOT NULL,
    attestation_type TEXT,
    aaguid BYTEA,
    sign_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL
);

-- Recovery codes
CREATE TABLE recovery_codes (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    code_hash TEXT NOT NULL,  -- PBKDF2-HMAC-SHA256
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

-- OAuth 2.1 clients
CREATE TABLE oauth_clients (
    client_id TEXT PRIMARY KEY,
    client_secret_hash TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    grant_types TEXT[] NOT NULL,
    scopes TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- OAuth 2.1 access tokens (opaque only)
CREATE TABLE oauth_access_tokens (
    token_hash TEXT PRIMARY KEY,
    client_id TEXT REFERENCES oauth_clients(client_id),
    user_id UUID REFERENCES users(id),
    scopes TEXT[] NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Sessions
CREATE TABLE sessions (
    session_id TEXT PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    client_id TEXT REFERENCES oauth_clients(client_id),
    scopes TEXT[] NOT NULL,
    client_type TEXT NOT NULL CHECK (client_type IN ('browser', 'headless')),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

**Rationale**:

- **Full Feature Support**: Supports ALL authentication methods (Passkey, TOTP, etc.)
- **Dynamic Management**: User registration, MFA enrollment, token revocation
- **Scalability**: Handles millions of users, distributed databases
- **Limitations**: Database dependency, requires availability and backups

---

## Per-Service Authentication/Authorization Matrix

### SM-KMS (Secrets Manager - Key Management Service)

**Product**: P3: KMS
**Public Port**: 8080-8089
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Standalone Mode**:

1. Basic (Username/Password) - File > Database
2. Basic (Email/Password) - File > Database
3. Bearer (API Token) - File > Database

**Federated Mode**:

1. Passkey (WebAuthn) - Database only
2. TOTP - Database only
3. Hardware Security Keys - Database only
4. Basic (Username/Password) - File > Database
5. Email OTP - Database only
6. Bearer (API Token) - File > Database
7. HTTPS Client Certificate - File > Database
8. OAuth 2.1 Access Token (JWS/JWE/Opaque) - Database > File

#### Authentication Priority (Headless Clients)

**Standalone Mode**:

1. Basic (Client ID/Secret) - File > Database
2. Bearer (API Token) - File > Database

**Federated Mode**:

1. HTTPS Client Certificate (mTLS) - File > Database
2. Basic (Client ID/Secret) - File > Database
3. Bearer (API Token) - File > Database
4. OAuth 2.1 Access Token (JWS/JWE/Opaque) - Database > File

#### Authorization Priority

1. Scope-Based (read:keys, write:keys, rotate:keys)
2. RBAC (admin, operator, viewer)
3. Resource-Level (tenant isolation, key ownership)
4. IP Allowlist

---

### PKI-CA (Public Key Infrastructure - Certificate Authority)

**Product**: P4: CA
**Public Port**: 8443-8449
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Standalone Mode**:

1. Basic (Username/Password) - File > Database
2. Basic (Email/Password) - File > Database
3. Bearer (API Token) - File > Database

**Federated Mode**:

1. Passkey (WebAuthn) - Database only
2. TOTP - Database only
3. Hardware Security Keys - Database only
4. Basic (Username/Password) - File > Database
5. Email OTP - Database only
6. Bearer (API Token) - File > Database
7. HTTPS Client Certificate - File > Database (PKI integration)
8. OAuth 2.1 Access Token (JWS/JWE/Opaque) - Database > File

#### Authentication Priority (Headless Clients)

**Standalone Mode**:

1. Basic (Client ID/Secret) - File > Database
2. Bearer (API Token) - File > Database

**Federated Mode**:

1. HTTPS Client Certificate (mTLS) - File > Database (PKI integration)
2. Basic (Client ID/Secret) - File > Database
3. Bearer (API Token) - File > Database
4. OAuth 2.1 Access Token (JWS/JWE/Opaque) - Database > File

#### Authorization Priority

1. Scope-Based (read:certs, issue:cert, revoke:cert)
2. RBAC (ca_admin, ca_operator, ca_auditor)
3. Resource-Level (certificate profile enforcement)
4. IP Allowlist

---

### JOSE-JA (JOSE - JWK Authority)

**Product**: P1: JOSE
**Public Port**: 9443-9449
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Standalone Mode**:

1. Basic (Username/Password) - File > Database
2. Basic (Email/Password) - File > Database
3. Bearer (API Token) - File > Database

**Federated Mode**:

1. Passkey (WebAuthn) - Database only
2. TOTP - Database only
3. Hardware Security Keys - Database only
4. Basic (Username/Password) - File > Database
5. Email OTP - Database only
6. Bearer (API Token) - File > Database
7. HTTPS Client Certificate - File > Database
8. OAuth 2.1 Access Token (JWS/JWE/Opaque) - Database > File

#### Authentication Priority (Headless Clients)

**Standalone Mode**:

1. Basic (Client ID/Secret) - File > Database
2. Bearer (API Token) - File > Database

**Federated Mode**:

1. HTTPS Client Certificate (mTLS) - File > Database
2. Basic (Client ID/Secret) - File > Database
3. Bearer (API Token) - File > Database
4. OAuth 2.1 Access Token (JWS/JWE/Opaque) - Database > File

#### Authorization Priority

1. Scope-Based (read:jwks, sign:jwt, verify:jwt)
2. RBAC (jose_admin, jose_operator, jose_viewer)
3. Resource-Level (JWK ownership, algorithm restrictions)
4. IP Allowlist

---

### Identity-Authz (Identity - Authorization Server)

**Product**: P2: Identity
**Public Port**: 18000-18009
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Federated Mode ONLY** (Identity product doesn't have standalone mode):

1. Passkey (WebAuthn) - Database only
2. TOTP - Database only
3. Hardware Security Keys - Database only
4. Basic (Username/Password) - File > Database
5. Basic (Email/Password) - File > Database
6. Email OTP - Database only
7. Recovery Codes - Database only
8. SMS OTP - Database only
9. Magic Link - Database only

#### Authentication Priority (Headless Clients)

**Federated Mode ONLY**:

1. HTTPS Client Certificate (mTLS) - File > Database
2. Basic (Client ID/Secret) - File > Database
3. Bearer (API Token) - File > Database

#### Authorization Priority

1. Scope-Based (openid, profile, email, offline_access)
2. RBAC (authz_admin, authz_operator)
3. Client-Level ACL (client_id allowlist/blocklist)
4. IP Allowlist

**Special Note**: Authorization Server is the **source of truth** for OAuth 2.1 tokens and scopes. All other services validate tokens issued by this service.

---

### Identity-IdP (Identity - Identity Provider)

**Product**: P2: Identity
**Public Port**: 18100-18109
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Federated Mode ONLY**:

1. Passkey (WebAuthn) - Database only
2. TOTP - Database only
3. Hardware Security Keys - Database only
4. Basic (Username/Password) - File > Database
5. Basic (Email/Password) - File > Database
6. Email OTP - Database only
7. Recovery Codes - Database only
8. SMS OTP - Database only
9. Phone Call OTP - Database only
10. Magic Link - Database only
11. Push Notification - Database only

#### Authentication Priority (Headless Clients)

**Federated Mode ONLY**:

1. HTTPS Client Certificate (mTLS) - File > Database
2. Basic (Client ID/Secret) - File > Database
3. Bearer (API Token) - File > Database

#### Authorization Priority

1. Scope-Based (openid, profile, email, phone)
2. RBAC (idp_admin, idp_operator)
3. Consent Tracking (user consent for scopes)
4. IP Allowlist

**Special Note**: Identity Provider handles **all MFA enrollment and verification**. Supports ALL 9 MFA factors from constitution.md.

---

### Identity-RS (Identity - Resource Server)

**Product**: P2: Identity
**Public Port**: 18200-18209
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Federated Mode ONLY** (validates tokens from AuthZ):

1. Session Cookie (opaque/JWE/JWS) - Database validation
2. OAuth 2.1 Access Token (JWS) - Stateless validation (JWKS)
3. OAuth 2.1 Access Token (JWE) - Database > File
4. OAuth 2.1 Access Token (Opaque) - Database introspection

#### Authentication Priority (Headless Clients)

**Federated Mode ONLY**:

1. OAuth 2.1 Access Token (JWS) - Stateless validation (JWKS)
2. OAuth 2.1 Access Token (JWE) - Database > File
3. OAuth 2.1 Access Token (Opaque) - Database introspection
4. HTTPS Client Certificate (mTLS) - File > Database

#### Authorization Priority

1. Scope-Based (read:profile, write:profile)
2. Resource-Level (user owns resource)
3. IP Allowlist

**Special Note**: Resource Server is **reference implementation** demonstrating token validation patterns for protected APIs.

---

### Identity-RP (Identity - Relying Party)

**Product**: P2: Identity
**Public Port**: 18300-18309
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Federated Mode ONLY** (Backend-for-Frontend pattern):

1. Session Cookie (opaque/JWE/JWS) - Server-side session
2. OAuth 2.1 Authorization Code + PKCE - Exchange code for token

#### Authentication Priority (Headless Clients)

**NOT APPLICABLE** (RP is browser-facing only)

#### Authorization Priority

1. Scope-Based (scopes from OAuth 2.1 flow)
2. Session-Based (user authenticated via OAuth 2.1)

**Special Note**: Relying Party is **reference implementation** demonstrating Backend-for-Frontend pattern for SPAs.

---

### Identity-SPA (Identity - Single Page Application)

**Product**: P2: Identity
**Public Port**: 18400-18409
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Federated Mode ONLY** (Static hosting for SPA):

1. OAuth 2.1 Authorization Code + PKCE - SPA redirects to AuthZ
2. Session Cookie (issued by RP backend) - BFF pattern

#### Authentication Priority (Headless Clients)

**NOT APPLICABLE** (SPA is browser-facing only)

#### Authorization Priority

1. Scope-Based (scopes from OAuth 2.1 flow)
2. CORS enforcement (allowlist origins)

**Special Note**: SPA is **reference implementation** demonstrating OAuth 2.1 + PKCE for browser-based clients.

---

### Learn-PS (Learn - Pet Store)

**Product**: Demo
**Public Port**: 8888-8889
**Admin Port**: 9090

#### Authentication Priority (Browser Clients)

**Standalone Mode**:

1. Basic (Username/Password) - File > Database
2. Bearer (API Token) - File > Database

**Federated Mode**:

1. Passkey (WebAuthn) - Database only
2. TOTP - Database only
3. Basic (Username/Password) - File > Database
4. OAuth 2.1 Access Token (JWS) - Stateless validation

#### Authentication Priority (Headless Clients)

**Standalone Mode**:

1. Basic (Client ID/Secret) - File > Database
2. Bearer (API Token) - File > Database

**Federated Mode**:

1. HTTPS Client Certificate (mTLS) - File > Database
2. OAuth 2.1 Access Token (JWS) - Stateless validation

#### Authorization Priority

1. Scope-Based (read:pets, write:pets)
2. RBAC (pet_admin, pet_viewer)

**Special Note**: Pet Store is **educational demonstration service** validating service template reusability (Phase 7).

---

## Implementation Phases

### Phase 1: Standalone Mode - Basic Authentication

**Services**: All 9 services
**Authentication Methods**:

- Browser: Basic (Username/Password), Bearer (API Token)
- Headless: Basic (Client ID/Secret), Bearer (API Token)

**Realm Types**: File (YAML) only
**Authorization**: Scope-Based, RBAC

### Phase 2.1: Federated Mode - Core MFA

**Services**: Identity-Authz, Identity-IdP, Identity-RS
**Authentication Methods**:

- Browser: Passkey, TOTP, Hardware Security Keys (+ Phase 1 methods)
- Headless: mTLS, OAuth 2.1 JWS tokens (+ Phase 1 methods)

**Realm Types**: File + Database
**Authorization**: Scope-Based, RBAC, Consent Tracking

### Phase 2.2: Federated Mode - Extended MFA

**Services**: All 9 services
**Authentication Methods**:

- Browser: Email OTP, Recovery Codes, SMS OTP (+ Phase 2.1 methods)
- Headless: OAuth 2.1 JWE/Opaque tokens (+ Phase 2.1 methods)

**Realm Types**: File + Database
**Authorization**: Scope-Based, RBAC, Resource-Level, Client-Level

### Phase 2.3: Federated Mode - Complete MFA

**Services**: All 9 services
**Authentication Methods**:

- Browser: Phone Call OTP, Magic Link, Push Notification (+ Phase 2.2 methods)
- Headless: All methods from Phase 2.2

**Realm Types**: File + Database
**Authorization**: All methods (Scope, RBAC, Resource, Client, IP, Rate Limiting)

### Phase 3+: Reference Implementations

**Services**: Identity-RP, Identity-SPA, Learn-PS
**Purpose**: Demonstrate integration patterns, best practices
**Optional Deployment**: Not required for core Identity product functionality

---

## Summary Tables

### Authentication Method Support by Service

| Authentication Method | sm-kms | pki-ca | jose-ja | authz | idp | rs | rp | spa | ps |
|----------------------|--------|--------|---------|-------|-----|----|----|-----|-----|
| Basic (User/Pass) - File | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Basic (User/Pass) - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Basic (Client/Secret) - File | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Basic (Client/Secret) - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Bearer (API Token) - File | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Bearer (API Token) - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Passkey (WebAuthn) - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| TOTP - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Hardware Security Keys - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Email OTP - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Recovery Codes - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| SMS OTP - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Phone Call OTP - DB | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Magic Link - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Push Notification - DB | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| mTLS - File/DB | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |
| OAuth 2.1 JWS - File/DB | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| OAuth 2.1 JWE - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| OAuth 2.1 Opaque - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Session Cookie - DB | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**Legend**:

- ✅ Supported
- ❌ Not applicable (service doesn't support this auth method)

### Authorization Method Support by Service

| Authorization Method | sm-kms | pki-ca | jose-ja | authz | idp | rs | rp | spa | ps |
|---------------------|--------|--------|---------|-------|-----|----|----|----|-----|
| Scope-Based | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| RBAC | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |
| Resource-Level | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ |
| Consent Tracking | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ |
| Client-Level ACL | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| IP Allowlist | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |
| Rate Limiting | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| mTLS-Based | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |

---

## Configuration Examples

### Example 1: SM-KMS Standalone Mode (File Realm)

```yaml
# configs/kms/config-standalone.yaml
server:
  public_bind_address: "127.0.0.1"
  public_port: 8080
  admin_bind_address: "127.0.0.1"
  admin_port: 9090

authentication:
  file_realm:
    enabled: true
    users:
      - username: admin
        password_hash: $pbkdf2-sha256$600000$salt$hash
        roles: [admin]
        scopes: [read:*, write:*, rotate:*]
      - username: operator
        password_hash: $pbkdf2-sha256$600000$salt$hash
        roles: [operator]
        scopes: [read:keys, write:keys]

    clients:
      - client_id: backup-service
        client_secret_hash: $pbkdf2-sha256$600000$salt$hash
        scopes: [read:keys, encrypt:data]

    api_tokens:
      - token_hash: $pbkdf2-sha256$600000$salt$hash
        description: "Monitoring system token"
        scopes: [read:health]

  database_realm:
    enabled: false  # Standalone mode

authorization:
  scope_enforcement: true
  rbac_enforcement: true
  ip_allowlist:
    enabled: true
    cidrs: ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
```

### Example 2: SM-KMS Federated Mode (Database Realm)

```yaml
# configs/kms/config-federated.yaml
server:
  public_bind_address: "0.0.0.0"  # Container binding
  public_port: 8080
  admin_bind_address: "127.0.0.1"
  admin_port: 9090

database:
  url: "postgres://kms_user:password@postgres:5432/kms_db"

authentication:
  file_realm:
    enabled: true  # Disaster recovery
    users:
      - username: admin
        password_hash: $pbkdf2-sha256$600000$salt$hash
        roles: [admin]
        scopes: [read:*, write:*, rotate:*]

  database_realm:
    enabled: true  # Primary authentication source
    mfa_required: true
    supported_factors:
      - passkey  # Priority 1
      - totp     # Priority 2
      - hardware_security_key  # Priority 3
      - email_otp  # Priority 4
      - recovery_codes  # Priority 5

federation:
  identity_url: "https://identity-authz:18000"
  identity_enabled: true
  identity_timeout: 10s

  jose_url: "https://jose-ja:9443"
  jose_enabled: true

  federation_fallback:
    identity_fallback_mode: "reject_all"  # MANDATORY for production
    jose_fallback_mode: "internal_crypto"

authorization:
  scope_enforcement: true
  rbac_enforcement: true
  resource_level_acl: true
  ip_allowlist:
    enabled: true
    cidrs: ["10.0.0.0/8"]
```

### Example 3: Identity-IdP MFA Configuration

```yaml
# configs/identity/idp-config.yaml
server:
  public_bind_address: "0.0.0.0"
  public_port: 18100
  admin_bind_address: "127.0.0.1"
  admin_port: 9090

database:
  url: "postgres://idp_user:password@postgres:5432/idp_db"

authentication:
  # File realm for disaster recovery only
  file_realm:
    enabled: true
    users:
      - username: admin
        password_hash: $pbkdf2-sha256$600000$salt$hash
        roles: [idp_admin]

  # Primary authentication: Database realm with ALL MFA factors
  database_realm:
    enabled: true
    mfa_required: true
    mfa_factor_priority:
      - passkey  # HIGHEST priority, FIDO2
      - totp  # HIGH priority, authenticator apps
      - hardware_security_key  # HIGH priority, FIDO U2F
      - email_otp  # MEDIUM priority
      - recovery_codes  # MEDIUM priority
      - sms_otp  # LOW priority (NIST deprecated but MANDATORY)
      - phone_call_otp  # LOWEST priority (NIST deprecated but MANDATORY)
      - magic_link  # LOWEST priority
      - push_notification  # LOWEST priority

    # Enrollment flow
    mfa_enrollment:
      require_on_first_login: true
      minimum_factors: 2  # Require 2 MFA factors enrolled
      backup_factor_required: true  # Require recovery codes

    # Verification flow
    mfa_verification:
      max_attempts: 3
      lockout_duration: 15m
      remember_device: true
      remember_duration: 30d

# External service integrations
integrations:
  email_provider:
    type: smtp
    smtp_host: "smtp.example.com"
    smtp_port: 587
    smtp_from: "noreply@example.com"

  sms_provider:
    type: twilio
    account_sid: "file:///run/secrets/twilio_account_sid"
    auth_token: "file:///run/secrets/twilio_auth_token"

  push_provider:
    type: firebase
    service_account: "file:///run/secrets/firebase_service_account.json"

authorization:
  scope_enforcement: true
  consent_tracking: true
  consent_expiration: 90d
```

---

## References

- Constitution: `.specify/memory/constitution.md` (Version 3.0.0)
- Specification: `specs/002-cryptoutil/spec.md` (1786 lines)
- Clarifications: `specs/002-cryptoutil/clarify.md` (Q4: MFA Factor Priority, Q7: Federation Fallback)
- NIST SP 800-63B Revision 3: Digital Identity Guidelines (Authentication and Lifecycle Management)
- OAuth 2.1: RFC draft (Authorization Framework)
- OIDC 1.0: OpenID Connect Core specification
- FIDO2/WebAuthn: W3C Web Authentication specification
- RFC 6238: TOTP (Time-Based One-Time Password Algorithm)

---

**Document Version**: 1.0.0
**Last Updated**: 2025-12-22
**Maintainer**: Spec Kit AI Agent
