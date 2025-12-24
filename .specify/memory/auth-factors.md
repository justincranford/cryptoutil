# Authentication Factors Reference

**Version**: 1.0.0
**Last Updated**: 2025-12-23
**Authority**: Single source of truth for authentication methods across all cryptoutil products
**Referenced By**: constitution.md, spec.md, copilot instructions (02-10.authentication.instructions.md)

---

## Purpose

This document provides the authoritative list of all authentication methods supported by cryptoutil products, including per-factor storage realm specifications, MFA patterns, and session management requirements.

---

## Headless Authentication Methods (10 Total)

### Non-Federated (3 Methods)

1. **Basic (Client ID/Secret)**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: HTTP Basic authentication with client credentials
   - Use Case: Service-to-service API access

2. **Bearer (API Token)**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: Bearer token authentication (API keys)
   - Use Case: Long-lived service credentials

3. **HTTPS Client Certificate**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: mTLS client certificate authentication
   - Use Case: High-security service-to-service communication

### Federated (7 Methods)

4. **Basic (Client ID/Secret) - OAuth 2.1**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: HTTP Basic with OAuth 2.1 client credentials flow
   - Use Case: OAuth 2.1 client authentication

5. **Bearer (API Token) - OAuth 2.1**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: Bearer token from external OAuth 2.1 provider
   - Use Case: Federated API access tokens

6. **HTTPS Client Certificate - OAuth 2.1**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: mTLS with OAuth 2.1 service-to-service
   - Use Case: Certificate-bound OAuth tokens

7. **JWE Access Token**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: JSON Web Encryption access tokens
   - Use Case: Encrypted access tokens with confidential claims

8. **JWS Access Token**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: JSON Web Signature access tokens
   - Use Case: Signed access tokens with inspectable claims

9. **Opaque Access Token**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: Non-JWT access tokens (random identifiers)
   - Use Case: Server-side session lookup tokens

10. **Opaque Refresh Token**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: Non-JWT refresh tokens
    - Use Case: Long-lived credential refresh mechanism

---

## Browser Authentication Methods (28 Total)

### Non-Federated (6 Methods)

1. **JWE Session Cookie**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: Encrypted session cookie (JSON Web Encryption)
   - Use Case: Stateless encrypted browser sessions

2. **JWS Session Cookie**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: Signed session cookie (JSON Web Signature)
   - Use Case: Stateless signed browser sessions

3. **Opaque Session Cookie**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: Non-JWT session cookie (random session ID)
   - Use Case: Server-side session storage with cookie reference

4. **Basic (Username/Password)**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: HTTP Basic authentication with user credentials
   - Use Case: Simple username/password login

5. **Bearer (API Token)**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: Bearer token authentication
   - Use Case: Browser-based API token usage

6. **HTTPS Client Certificate**
   - Storage: YAML + SQL (Config > DB priority)
   - Protocol: mTLS client certificate authentication
   - Use Case: Certificate-based browser authentication

### Federated (22 Methods - All Non-Federated PLUS)

7. **TOTP (Authenticator App)**
   - Storage: SQL ONLY (DB-only)
   - Protocol: Time-based One-Time Password (Google Authenticator, Authy)
   - Use Case: Second factor authentication with mobile app
   - Enrollment: User-specific, requires QR code or secret key setup

8. **HOTP (Hardware Token)**
   - Storage: SQL ONLY (DB-only)
   - Protocol: HMAC-based One-Time Password (YubiKey, RSA SecurID)
   - Use Case: Hardware-based one-time passwords
   - Enrollment: User-specific, requires hardware token registration

9. **Recovery Codes**
   - Storage: SQL ONLY (DB-only)
   - Protocol: Backup single-use recovery codes
   - Use Case: Account recovery when primary MFA unavailable
   - Enrollment: User-specific, generated during MFA setup

10. **WebAuthn with Passkeys**
    - Storage: SQL ONLY (DB-only)
    - Protocol: FIDO2 WebAuthn with platform authenticators (Face ID, Touch ID, Windows Hello)
    - Use Case: Passwordless authentication with device biometrics
    - Enrollment: User-specific, requires WebAuthn registration ceremony

11. **WebAuthn without Passkeys**
    - Storage: SQL ONLY (DB-only)
    - Protocol: FIDO2 WebAuthn with security keys (YubiKey, Titan Key)
    - Use Case: Hardware security key authentication
    - Enrollment: User-specific, requires physical key registration

12. **Push Notification**
    - Storage: SQL ONLY (DB-only)
    - Protocol: Mobile app push-based authentication
    - Use Case: Approve/deny push notifications on mobile device
    - Enrollment: User-specific, requires mobile app installation

13. **Email/Password**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: Email address + password authentication
    - Use Case: Standard email-based user accounts

14. **Magic Link (Email)**
    - Storage: SQL ONLY (DB-only)
    - Protocol: Passwordless email-based authentication
    - Use Case: One-click email login links
    - Enrollment: User-specific, sent to registered email

15. **Magic Link (SMS)**
    - Storage: SQL ONLY (DB-only)
    - Protocol: Passwordless SMS-based authentication
    - Use Case: One-click SMS login links
    - Enrollment: User-specific, sent to registered phone

16. **Random OTP (Email)**
    - Storage: SQL ONLY (DB-only)
    - Protocol: One-time password sent via email
    - Use Case: Email-based verification codes
    - Enrollment: User-specific, sent to registered email

17. **Random OTP (SMS)**
    - Storage: SQL ONLY (DB-only)
    - Protocol: One-time password sent via SMS
    - Use Case: SMS-based verification codes
    - Enrollment: User-specific, sent to registered phone

18. **Random OTP (Phone)**
    - Storage: SQL ONLY (DB-only)
    - Protocol: One-time password sent via voice call
    - Use Case: Voice-based verification codes
    - Enrollment: User-specific, sent to registered phone

19. **Social Login (Google)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with Google Identity Platform
    - Use Case: Sign in with Google account

20. **Social Login (Microsoft)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with Microsoft Identity Platform
    - Use Case: Sign in with Microsoft account

21. **Social Login (GitHub)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with GitHub
    - Use Case: Sign in with GitHub account

22. **Social Login (Facebook)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with Facebook
    - Use Case: Sign in with Facebook account

23. **Social Login (Apple)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with Sign in with Apple
    - Use Case: Sign in with Apple ID

24. **Social Login (LinkedIn)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with LinkedIn
    - Use Case: Sign in with LinkedIn account

25. **Social Login (Twitter/X)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with Twitter/X
    - Use Case: Sign in with Twitter/X account

26. **Social Login (Amazon)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with Amazon Login
    - Use Case: Sign in with Amazon account

27. **Social Login (Okta)**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: OAuth 2.0 with Okta Identity Cloud
    - Use Case: Enterprise SSO with Okta

28. **SAML 2.0**
    - Storage: YAML + SQL (Config > DB priority)
    - Protocol: SAML 2.0 federated authentication
    - Use Case: Enterprise SSO with SAML identity providers

---

## Storage Realm Specifications

### YAML + SQL (Config > DB Priority)

**Purpose**: Disaster recovery - service must start even if database unavailable

**Pattern**:

- Credentials stored in BOTH configuration files AND database
- Service attempts YAML first, falls back to SQL if YAML unavailable
- Updates written to BOTH realms for consistency

**Applicable Methods**:

- Static credentials (Basic auth, Bearer tokens, Client certificates)
- Provider configurations (Social Login client IDs/secrets, SAML metadata)
- Session token encryption keys (JWE/JWS signing keys)

**Rationale**: Pre-configured credentials enable service bootstrap without database dependency

### SQL ONLY

**Purpose**: Dynamic user-specific data requiring persistence

**Pattern**:

- Credentials stored ONLY in database (no YAML support)
- Cannot be pre-configured in YAML (dynamic per-user enrollment)
- Updates written to database only

**Applicable Methods**:

- User-specific enrollment data (TOTP secrets, WebAuthn credentials, Recovery codes)
- One-time tokens/codes (Magic Link tokens, Random OTP codes)
- Push notification device registrations

**Rationale**: User enrollment is dynamic and cannot be predetermined in configuration files

---

## Multi-Factor Authentication (MFA)

### Common MFA Combinations

- **Browser + TOTP**: Password + Authenticator App (most common)
- **Browser + WebAuthn**: Password + Security Key
- **Browser + Push**: Password + Mobile App Push
- **Browser + OTP**: Password + SMS/Email OTP
- **Headless + mTLS**: Client ID/Secret + TLS Client Certificate
- **Headless + Bearer**: Client ID/Secret + API Token

### MFA Step-Up Authentication

**Trigger**: Time-based re-authentication for sensitive resources

**Pattern**:

- Re-authentication MANDATORY every 30 minutes for high-sensitivity operations
- Applies regardless of operation type
- Session remains valid for low-sensitivity operations
- Configurable per resource sensitivity level

**Configuration Example**:

```yaml
mfa:
  step_up_interval: 1800  # 30 minutes in seconds
  required_factors: 2
  resource_sensitivity:
    high: ["delete:account", "update:billing", "export:data"]
    medium: ["update:profile", "create:api-key"]
    low: ["read:profile", "list:resources"]
```

### MFA Enrollment Workflow

**Optional Enrollment** (Limited Access Pattern):

- OPTIONAL enrollment during initial setup
- Access LIMITED to low-sensitivity resources until additional factors enrolled
- Only one identifying factor required for initial login
- Admin configures minimum factors required per resource sensitivity

---

## Session Token Formats

### Configuration-Driven Selection

**Non-Federated Mode**: Product-specific config determines format (opaque, JWE, JWS)
**Federated Mode**: Identity Provider config determines format

**Supported Formats**:

| Format | Description | Use Case |
|--------|-------------|----------|
| **Opaque** | Random UUID, server-side lookup | Maximum security, no token inspection |
| **JWE** | Encrypted JWT | Stateless, encrypted claims |
| **JWS** | Signed JWT | Stateless, inspectable claims |

**Example Configuration**:

```yaml
identity:
  session:
    token_format: "jwe"  # opaque, jwe, or jws
    token_ttl: 3600      # 1 hour
    refresh_ttl: 604800  # 7 days
```

### Session Storage Backend

**Supported Backends**:

- **SQLite**: Single-node deployments, development, testing
- **PostgreSQL**: Distributed/high-availability deployments with shared session data

**NEVER SUPPORTED**:

- ❌ Redis (BANNED - adds operational complexity)
- ❌ Memcached (BANNED - volatile storage unsuitable for sessions)
- ❌ In-memory (BANNED - lost on restart, breaks multi-instance deployments)

**Rationale**: SQLite/PostgreSQL provide ACID guarantees, persistence, and consistency.

---

## Authorization Methods

### Headless Authorization (2 Methods)

1. **Scope-Based Authorization**
   - Protocol: OAuth 2.1 scopes (read:keys, write:keys)
   - Use Case: API access control

2. **RBAC (Role-Based Access Control)**
   - Protocol: Role assignments (admin, operator, viewer)
   - Use Case: Service-to-service permissions

### Browser Authorization (4 Methods)

1. **Scope-Based Authorization**
   - Protocol: OAuth 2.1 scopes
   - Use Case: API access control

2. **RBAC (Role-Based Access Control)**
   - Protocol: Role assignments
   - Use Case: User role permissions

3. **Resource-Level Access Control**
   - Protocol: Per-resource ownership and ACLs
   - Use Case: User-owned resources

4. **Consent Tracking**
   - Protocol: (scope, resource) tuples
   - Use Case: User consent management

### Zero Trust Authorization

**MANDATORY**: Authorization MUST be evaluated on EVERY request

- ❌ **NO caching of authorization decisions** (prevents stale permissions)
- ✅ **Always fetch latest permissions from database**
- ✅ **Performance via efficient policy evaluation, NOT caching**

**Cross-Service Authorization**:

- Session Token passed between federated services
- Each service independently validates token AND enforces authorization
- NO token transformation or delegation
- NO trust-on-first-use patterns

**Consent Tracking Granularity**:

- Tracked as `(scope, resource)` tuples
- Example: `("read:keys", "key-123")` separate from `("read:keys", "key-456")`
- User grants consent per scope+resource combination

---

## Testing Requirements

### Authentication Tests MUST

- Test all 10 headless + 28 browser authentication methods
- Verify storage realm failover (YAML → PostgreSQL → SQLite)
- Verify MFA combinations (2-factor, 3-factor)
- Test step-up authentication timing (30-minute intervals)
- Test enrollment workflows (optional with limited access)

### Authorization Tests MUST

- Test Zero Trust evaluation (no caching)
- Test cross-service token validation
- Test consent tracking (scope+resource tuples)
- Test realm failover behavior

### Coverage Targets

- Authentication handlers: ≥95% coverage
- Authorization middleware: ≥95% coverage
- Realm failover logic: ≥98% coverage (infrastructure)

---

## Migration Checklist

When implementing authentication/authorization:

- [ ] Define supported authentication factors (headless + browser)
- [ ] Configure storage realms (YAML + SQL with failover)
- [ ] Implement middleware stacks (`/service/**` vs `/browser/**`)
- [ ] Configure session token format and storage backend
- [ ] Implement MFA step-up and enrollment workflows
- [ ] Configure realm failover priority list
- [ ] Implement Zero Trust authorization (no caching)
- [ ] Add consent tracking for scope+resource tuples
- [ ] Write tests for all authentication methods
- [ ] Write tests for authorization policies
- [ ] Document authentication factor configuration in deployment guides

---

## Key Takeaways

1. **10 headless + 28 browser authentication methods** - Comprehensive factor support
2. **YAML + SQL storage with Config > DB priority** - Disaster recovery pattern
3. **Zero Trust authorization** - NO caching, always re-evaluate permissions
4. **MFA step-up every 30 minutes** - Time-based re-authentication for sensitive resources
5. **Session format configurable** - Opaque, JWE, or JWS tokens
6. **PostgreSQL/SQLite session storage** - NO Redis/Memcached
7. **Realm failover with priority list** - Try YAML → PostgreSQL → SQLite
8. **Consent tracking at scope+resource granularity** - Fine-grained user consent
