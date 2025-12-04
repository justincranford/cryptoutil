# Specification Clarifications

## Purpose

This document resolves ambiguities identified in `spec.md` during `/speckit.clarify`.

**Date**: December 4, 2025
**Iteration**: 1 (completing skipped clarify step)

---

## Authentication Methods Clarifications

### 1. client_secret_jwt (⚠️ Partial → Defined)

**Status**: ⚠️ 70% Complete

**What's Implemented**:
- JWT assertion parsing
- Client secret verification structure
- Basic token endpoint support

**What's Missing (30%)**:
- Assertion lifetime validation (exp, iat, nbf claims)
- `jti` claim uniqueness verification (replay protection)
- Proper error messages per RFC 7523
- Test coverage for edge cases

**Priority**: HIGH - Required for OAuth 2.1 compliance

**Resolution**: Update spec.md to show "⚠️ 70% (missing: jti replay protection, assertion lifetime validation)"

---

### 2. private_key_jwt (⚠️ Partial → Defined)

**Status**: ⚠️ 50% Complete

**What's Implemented**:
- JWT parsing infrastructure
- Public key retrieval from JWKS

**What's Missing (50%)**:
- Client JWKS registration endpoint
- Client public key storage
- `jti` claim uniqueness verification
- `kid` header matching to client keys
- Test coverage

**Priority**: HIGH - Required for OAuth 2.1 compliance with confidential clients

**Resolution**: Update spec.md to show "⚠️ 50% (missing: client JWKS registration, jti replay, kid matching)"

---

## MFA Factors Clarifications

### 3. Email OTP (⚠️ Partial → Defined)

**Status**: ⚠️ 30% Complete

**What's Implemented**:
- OTP generation infrastructure (shared with TOTP)
- Email template structure (placeholder)

**What's Missing (70%)**:
- Email delivery service integration
- Email delivery provider abstraction (SMTP, SendGrid, etc.)
- OTP storage with expiration
- Rate limiting per email address
- Email verification flow

**Priority**: MEDIUM - Not required for core OAuth 2.1 flows

**Resolution**: Update spec.md to show "⚠️ 30% (missing: email delivery service, rate limiting)"

---

### 4. SMS OTP (⚠️ Partial → Defined)

**Status**: ⚠️ 20% Complete

**What's Implemented**:
- OTP generation infrastructure (shared with TOTP)

**What's Missing (80%)**:
- SMS delivery service integration
- SMS provider abstraction (Twilio, AWS SNS, etc.)
- Phone number verification flow
- OTP storage with expiration
- Rate limiting per phone number
- Country code validation

**Priority**: LOW - NIST discourages SMS OTP; prefer TOTP/Passkey

**Resolution**: Update spec.md to show "⚠️ 20% (missing: SMS provider integration, rate limiting)"

---

## MFA Priority Order

Per constitution and security requirements:

1. **Passkey/WebAuthn** (HIGHEST) - Phishing-resistant, FIDO2 compliant
2. **Hardware Security Keys** (HIGH) - U2F/FIDO hardware tokens
3. **TOTP** (HIGH) - Time-based OTP apps (Google Authenticator, Authy)
4. **Email OTP** (MEDIUM) - Fallback for password reset
5. **SMS OTP** (LOW) - NIST deprecated, use only as last resort

---

## Status Legend Update

| Symbol | Meaning | Percentage |
|--------|---------|------------|
| ✅ Working | Feature complete and tested | 100% |
| ⚠️ XX% | Partially implemented | 1-99% |
| ❌ Not Implemented | Not started | 0% |
| ❌ Not Required | Explicitly out of scope | N/A |

---

## Action Items

1. ✅ Update spec.md with percentage completions
2. ✅ Add missing features to tasks.md
3. ⬜ Implement client_secret_jwt replay protection
4. ⬜ Implement private_key_jwt client JWKS
5. ⬜ Design email/SMS provider abstraction

---

*Clarification Version: 1.0.0*
*Resolved By: /speckit.clarify*
