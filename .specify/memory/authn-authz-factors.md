# Authentication and Authorization Factors Reference

**Version**: 1.1.0
**Last Updated**: 2025-12-24
**Authority**: Single source of truth for authentication/authorization methods across all cryptoutil products
**Referenced By**: constitution.md, spec.md, copilot instructions (02-10.authn.instructions.md)

---

## Purpose

This document provides the authoritative list of all authentication and authorization methods supported by cryptoutil products, using compact table format for quick reference.

---

## Headless Authentication Methods (10 Total)

**Request Path**: `/service/**` paths only

### Non-Federated (3 Methods)

| # | Method | Storage Realm | Protocol | Use Case |
|---|--------|---------------|----------|----------|
| 1 | Basic (Client ID/Secret) | YAML + SQL (Config > DB) | HTTP Basic with client credentials | Service-to-service API access |
| 2 | Bearer (API Token) | YAML + SQL (Config > DB) | Bearer token (API keys) | Long-lived service credentials |
| 3 | HTTPS Client Certificate | YAML + SQL (Config > DB) | mTLS client certificate | High-security service-to-service |

### Federated (7 Methods)

| # | Method | Storage Realm | Protocol | Use Case |
|---|--------|---------------|----------|----------|
| 4 | Basic (Client ID/Secret) - OAuth 2.1 | YAML + SQL (Config > DB) | HTTP Basic with OAuth 2.1 client credentials | OAuth 2.1 client authentication |
| 5 | Bearer (API Token) - OAuth 2.1 | YAML + SQL (Config > DB) | Bearer from external OAuth 2.1 provider | Federated API access tokens |
| 6 | HTTPS Client Certificate - OAuth 2.1 | YAML + SQL (Config > DB) | mTLS with OAuth 2.1 service-to-service | Certificate-bound OAuth tokens |
| 7 | JWE Access Token | YAML + SQL (Config > DB) | JSON Web Encryption access tokens | Encrypted access tokens |
| 8 | JWS Access Token | YAML + SQL (Config > DB) | JSON Web Signature access tokens | Signed access tokens |
| 9 | Opaque Access Token | YAML + SQL (Config > DB) | Non-JWT access tokens | Server-side session lookup |
| 10 | Opaque Refresh Token | YAML + SQL (Config > DB) | Non-JWT refresh tokens | Long-lived credential refresh |

---

## Browser Authentication Methods (28 Total)

**Request Path**: `/browser/**` paths only

### Non-Federated (6 Methods)

| # | Method | Storage Realm | Protocol | Use Case |
|---|--------|---------------|----------|----------|
| 1 | JWE Session Cookie | YAML + SQL (Config > DB) | Encrypted session cookie (JWE) | Stateless encrypted browser sessions |
| 2 | JWS Session Cookie | YAML + SQL (Config > DB) | Signed session cookie (JWS) | Stateless signed browser sessions |
| 3 | Opaque Session Cookie | YAML + SQL (Config > DB) | Non-JWT session cookie | Server-side session storage |
| 4 | Basic (Username/Password) | YAML + SQL (Config > DB) | HTTP Basic with user credentials | Simple username/password login |
| 5 | Bearer (API Token) | YAML + SQL (Config > DB) | Bearer token authentication | API access from browser |
| 6 | HTTPS Client Certificate | YAML + SQL (Config > DB) | mTLS client certificate | High-security browser access |

### Federated (22 Methods - All non-federated PLUS)

| # | Method | Storage Realm | Protocol | Use Case |
|---|--------|---------------|----------|----------|
| 7 | TOTP (Authenticator App) | SQL ONLY (DB-only) | Time-based One-Time Password | Google Authenticator, Authy |
| 8 | HOTP (Hardware Token) | SQL ONLY (DB-only) | HMAC-based One-Time Password | YubiKey, RSA SecurID |
| 9 | Recovery Codes | SQL ONLY (DB-only) | Backup single-use codes | Account recovery |
| 10 | WebAuthn with Passkeys | SQL ONLY (DB-only) | FIDO2 platform authenticators | Face ID, Touch ID, Windows Hello |
| 11 | WebAuthn without Passkeys | SQL ONLY (DB-only) | FIDO2 security keys | YubiKey, Titan Key |
| 12 | Push Notification | SQL ONLY (DB-only) | Mobile app push-based | Mobile authentication app |
| 13 | Email/Password | YAML + SQL (Config > DB) | Email address + password | Email-based login |
| 14 | Magic Link (Email) | SQL ONLY (DB-only) | Passwordless email-based | Email magic link authentication |
| 15 | Magic Link (SMS) | SQL ONLY (DB-only) | Passwordless SMS-based | SMS magic link authentication |
| 16 | Random OTP (Email) | SQL ONLY (DB-only) | One-time password via email | Email OTP verification |
| 17 | Random OTP (SMS) | SQL ONLY (DB-only) | One-time password via SMS | SMS OTP verification |
| 18 | Random OTP (Phone) | SQL ONLY (DB-only) | One-time password via voice | Phone call OTP verification |
| 19 | Social Login (Google) | YAML + SQL (Config > DB) | OAuth 2.0 | Google Identity Platform |
| 20 | Social Login (Microsoft) | YAML + SQL (Config > DB) | OAuth 2.0 | Microsoft Identity Platform |
| 21 | Social Login (GitHub) | YAML + SQL (Config > DB) | OAuth 2.0 | GitHub authentication |
| 22 | Social Login (Facebook) | YAML + SQL (Config > DB) | OAuth 2.0 | Facebook authentication |
| 23 | Social Login (Apple) | YAML + SQL (Config > DB) | OAuth 2.0 | Sign in with Apple |
| 24 | Social Login (LinkedIn) | YAML + SQL (Config > DB) | OAuth 2.0 | LinkedIn authentication |
| 25 | Social Login (Twitter/X) | YAML + SQL (Config > DB) | OAuth 2.0 | Twitter/X authentication |
| 26 | Social Login (Amazon) | YAML + SQL (Config > DB) | OAuth 2.0 | Amazon Login |
| 27 | Social Login (Okta) | YAML + SQL (Config > DB) | OAuth 2.0 | Okta Identity Cloud |
| 28 | SAML 2.0 | YAML + SQL (Config > DB) | SAML 2.0 federated | Enterprise SSO |

---

## Storage Realm Specifications

### YAML + SQL (Config > DB Priority)

**Purpose**: Disaster recovery - service must start even if database unavailable

**Pattern**:

- Credentials stored in BOTH configuration files AND database
- Service attempts YAML first, falls back to SQL if YAML unavailable
- Updates written to BOTH realms for consistency

**Applicable Methods**: Static credentials (Basic auth, Bearer tokens, Client certificates), Provider configurations (Social Login, SAML)

**Rationale**: Pre-configured credentials enable service bootstrap without database dependency

### SQL ONLY

**Purpose**: Dynamic user-specific data requiring persistence

**Pattern**:

- Credentials stored ONLY in database (no YAML support)
- Cannot be pre-configured in YAML (dynamic per-user enrollment)
- Updates written to database only

**Applicable Methods**: User-specific enrollment data (TOTP secrets, WebAuthn credentials, Recovery codes), One-time tokens/codes (Magic Links, Random OTPs)

**Rationale**: User enrollment is dynamic and cannot be predetermined in configuration files

---

## Multi-Factor Authentication (MFA)

### Common MFA Combinations

- Browser + TOTP: Password + Authenticator App (most common)
- Browser + WebAuthn: Password + Security Key
- Browser + Push: Password + Mobile App Push
- Browser + OTP: Password + SMS/Email OTP
- Headless + mTLS: Client ID/Secret + TLS Client Certificate
- Headless + Bearer: Client ID/Secret + API Token

### MFA Step-Up Authentication

- Re-authentication MANDATORY every 30 minutes for high-sensitivity operations
- Session remains valid for low-sensitivity operations
- Configurable per resource sensitivity level

### MFA Enrollment Workflow

- OPTIONAL enrollment during initial setup
- Access LIMITED to low-sensitivity resources until additional factors enrolled
- Only one identifying factor required for initial login

---

## Session Token Formats

| Format | Description | Use Case |
|--------|-------------|----------|
| Opaque | Random UUID, server-side lookup | Maximum security, no token inspection |
| JWE | Encrypted JWT | Stateless, encrypted claims |
| JWS | Signed JWT | Stateless, inspectable claims |

**Session Storage Backend**: SQLite (single-node), PostgreSQL (distributed). NO Redis/Memcached.

---

## Authorization Methods

### Headless Authorization (2 Methods)

1. **Scope-Based Authorization**: OAuth 2.1 scopes (read:keys, write:keys)
2. **RBAC**: Role assignments (admin, operator, viewer)

### Browser Authorization (4 Methods)

1. **Scope-Based Authorization**: OAuth 2.1 scopes
2. **RBAC**: Role assignments
3. **Resource-Level Access Control**: Per-resource ownership and ACLs
4. **Consent Tracking**: (scope, resource) tuples

**Zero Trust**: NO caching of authorization decisions, always re-evaluate permissions

---

## Key Takeaways

1. **10 headless + 28 browser authentication methods** - Comprehensive factor support
2. **YAML + SQL storage with Config > DB priority** - Disaster recovery pattern
3. **Zero Trust authorization** - NO caching, always re-evaluate permissions
4. **MFA step-up every 30 minutes** - Time-based re-authentication for sensitive resources
5. **Session format configurable** - Opaque, JWE, or JWS tokens
6. **PostgreSQL/SQLite session storage** - NO Redis/Memcached
7. **Consent tracking at scope+resource granularity** - Fine-grained user consent
