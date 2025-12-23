# cryptoutil CLARIFY-QUIZME-02.md

**Last Updated**: 2025-12-22
**Purpose**: Multiple choice questions for UNKNOWN answers requiring user input
**Format**: A-D options + E write-in for each question

## Instructions

**CRITICAL**: This file contains ONLY questions with UNKNOWN answers that require user clarification.

**Questions with KNOWN answers belong in clarify.md, NOT here.**

**When adding questions**:

1. Search copilot instructions, constitution.md, spec.md, codebase FIRST
2. If answer is KNOWN: Add Q&A to clarify.md and update constitution/spec as needed
3. If answer is UNKNOWN: Add question HERE with NO pre-filled answers
4. After user answers: Refactor clarify.md to cover answered questions, update constitution.md with architecture decisions, and update spec.md with finalized requirements

---

## Open Questions Requiring User Input

### Q1: Additional Authentication Methods Beyond Listed 17

**Context**: spec.md line 184 states "and MORE TO BE CLARIFIED" after listing 17 authentication methods for browser-based MFA in Federated Identity mode. AUTH-AUTHZ-MATRIX.md documents those 17 methods with deterministic priority.

**Question**: What additional authentication methods beyond the documented 17 methods should be supported for browser-based MFA authentication?

**Options**:

A. No additional methods - the 17 listed methods are comprehensive and complete  
B. Add biometric authentication (fingerprint, facial recognition, iris scan) via WebAuthn Biometric extension  
C. Add risk-based/adaptive authentication (device trust scores, geolocation, behavioral analysis)  
D. Add Zero Trust continuous verification (re-authentication on sensitive operations, step-up authentication)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q2: Additional Authentication Methods for Headless Clients

**Context**: spec.md line 187 states "and MORE TO BE CLARIFIED" after listing 7 authentication methods for headless-based MFA in Federated Identity mode. AUTH-AUTHZ-MATRIX.md documents those 7 methods with deterministic priority.

**Question**: What additional authentication methods beyond the documented 7 methods should be supported for headless-based MFA authentication?

**Options**:

A. No additional methods - the 7 listed methods are comprehensive and complete  
B. Add service mesh identity (Istio, Linkerd workload identity tokens)  
C. Add cloud provider managed identity (AWS IAM roles, Azure Managed Identity, GCP Workload Identity)  
D. Add Kubernetes ServiceAccount tokens with projected volume-mounted JWT  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q3: Session Token Format Decision

**Context**: spec.md lines 189-190 specify session tokens can be "opaque||JWE||JWS non-OAuth 2.1" but don't specify WHEN to use each format or if all three must be supported simultaneously.

**Question**: What is the session token format strategy for browser-based and headless-based clients?

**Options**:

A. Support all 3 formats simultaneously with client negotiation via `Prefer: session-format=<opaque|jwe|jws>` request header  
B. Use opaque tokens ONLY (database lookup required, maximum flexibility for revocation)  
C. Use JWS tokens ONLY (stateless validation, no database lookup, cannot revoke until expiration)  
D. Use JWE tokens ONLY (stateless validation with encrypted payload, no database lookup, cannot revoke until expiration)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q4: Session Token Storage Backend

**Context**: Opaque session tokens require server-side storage for validation. If using opaque tokens, must choose storage backend.

**Question**: What storage backend should be used for opaque session tokens (if opaque format chosen in Q3)?

**Options**:

A. Redis ONLY (in-memory, fast, supports TTL, distributed sessions)  
B. PostgreSQL/SQLite ONLY (persistent, ACID guarantees, complex queries)  
C. Redis PRIMARY with PostgreSQL fallback (Redis for speed, PostgreSQL for persistence/recovery)  
D. In-process memory cache ONLY (no external dependencies, single-instance deployments only, no persistence)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q5: Session Cookie Security Attributes

**Context**: spec.md line 200 specifies "HttpOnly, Secure, SameSite=Strict" but doesn't address session lifetime, renewal, or additional attributes.

**Question**: What are the complete session cookie security requirements?

**Options**:

A. HttpOnly, Secure, SameSite=Strict, Max-Age=3600 (1 hour), Domain=service.example.com, Path=/  
B. HttpOnly, Secure, SameSite=Strict, Max-Age=86400 (24 hours), sliding window renewal (extend on activity)  
C. HttpOnly, Secure, SameSite=Strict, Max-Age=900 (15 minutes), absolute expiration (no renewal), __Host- prefix  
D. HttpOnly, Secure, SameSite=Lax (allow cross-site navigation), Max-Age=3600 (1 hour), sliding window renewal  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q6: MFA Step-Up Authentication Triggers

**Context**: AUTH-AUTHZ-MATRIX.md lists 17 MFA methods but doesn't specify when step-up authentication (re-authentication for sensitive operations) is required.

**Question**: When should step-up authentication be triggered even with valid session token?

**Options**:

A. NEVER - once authenticated with MFA, session is valid until expiration (trust session token)  
B. For high-value operations ONLY (key deletion, admin user creation, security settings changes)  
C. For ALL operations requiring scope elevation (e.g., read:keys → write:keys within same session)  
D. Time-based re-authentication (every 30 minutes for sensitive resources, regardless of operation)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q7: MFA Enrollment Workflow for New Users

**Context**: clarify.md Q4 lists 9 MFA factors in priority order but doesn't specify if users MUST enroll in multiple factors or can choose one.

**Question**: What is the MFA enrollment requirement for new users in Federated Identity mode?

**Options**:

A. MANDATORY enrollment in top 3 factors (Passkey + TOTP + Hardware Key) before account activation  
B. MANDATORY enrollment in at least 1 primary factor (Passkey OR TOTP OR Hardware Key) + recovery codes  
C. OPTIONAL enrollment (user can skip MFA during initial setup, enforced later via admin policy)  
D. Adaptive enrollment (require MFA enrollment based on user's assigned roles/scopes)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q8: MFA Factor Fallback Strategy

**Context**: clarify.md Q4 lists 9 MFA factors but doesn't specify what happens when a user's primary factor is unavailable (e.g., lost phone for TOTP, broken hardware key).

**Question**: What is the fallback strategy when user's primary MFA factor is unavailable?

**Options**:

A. Auto-fallback to next enrolled factor in priority order (e.g., if TOTP unavailable, try Recovery Codes automatically)  
B. User MUST explicitly select fallback factor from "Having trouble?" link (no auto-fallback)  
C. Admin-assisted recovery ONLY (user contacts admin to reset MFA, no self-service fallback)  
D. Time-delayed self-service recovery (user requests recovery link via email, 24-hour delay before activation)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q9: OAuth 2.1 Access Token vs Session Token Distinction

**Context**: spec.md mentions both "opaque||JWE||JWS OAuth 2.1 Access Token" (authentication method) and "session cookie (opaque||JWE||JWS non-OAuth 2.1)" (session token after authentication). Unclear if these are the same or different tokens.

**Question**: What is the relationship between OAuth 2.1 Access Tokens and session cookies in Federated Identity mode?

**Options**:

A. SAME TOKEN - OAuth 2.1 Access Token IS the session cookie (set as HttpOnly cookie after token endpoint response)  
B. SEPARATE TOKENS - OAuth 2.1 Access Token exchanged for internal session cookie (backend-for-frontend pattern)  
C. CONDITIONAL - Browser clients use session cookies, headless clients use OAuth 2.1 Access Tokens (no cookies)  
D. NESTED TOKENS - Session cookie contains encrypted OAuth 2.1 Access Token (JWE wrapper around JWS access token)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q10: Realm Type Failover Behavior

**Context**: AUTH-AUTHZ-MATRIX.md specifies "File Realm Type (YAML) is higher priority than Database Realm Type (GORM/SQL)" but doesn't specify failover semantics.

**Question**: How should authentication realm failover work when database is unavailable?

**Options**:

A. Auto-failover to File Realm when database connection fails (transparent to user, log warning)  
B. Reject authentication when database unavailable, even if File Realm configured (fail-secure, prevent split-brain)  
C. Health check-based failover (switch to File Realm only after 3 consecutive database health check failures)  
D. Manual failover only (admin must explicitly enable File Realm mode via config change + service restart)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q11: Authorization Decision Caching Strategy

**Context**: AUTH-AUTHZ-MATRIX.md lists 8 authorization methods but doesn't specify if authorization decisions should be cached or evaluated on every request.

**Question**: What is the authorization decision caching strategy?

**Options**:

A. NO CACHING - Evaluate authorization on EVERY request (most accurate, highest latency)  
B. Session-scoped caching - Cache authorization decisions for session lifetime (fast, stale on permission changes)  
C. Time-limited caching - Cache authorization decisions for 60 seconds (balanced accuracy vs performance)  
D. Layered caching - Cache static decisions (scopes, roles) in session, evaluate dynamic decisions (resource ownership) on every request  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q12: Cross-Service Authorization Propagation

**Context**: spec.md describes 9 services with service federation but doesn't specify how authorization context propagates in service-to-service calls.

**Question**: How should authorization context propagate when KMS calls JOSE, or Identity calls KMS?

**Options**:

A. Propagate original user's scopes/roles (KMS → JOSE call uses end-user's permissions, not KMS service identity)  
B. Use service identity scopes ONLY (KMS → JOSE call uses KMS's service account scopes, ignoring end-user)  
C. Combined scopes (intersection of end-user scopes AND service identity scopes - most restrictive)  
D. Delegated authorization (KMS exchanges user token for service-specific token via token exchange RFC 8693)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q13: Rate Limiting Scope and Thresholds

**Context**: AUTH-AUTHZ-MATRIX.md lists Rate Limiting as LOW priority authorization method but doesn't specify thresholds or scope.

**Question**: What are the rate limiting thresholds and enforcement scope?

**Options**:

A. Per client_id: 1000 req/min globally across all endpoints  
B. Per endpoint: 100 req/min per client_id per endpoint (read:keys 100/min, write:keys 100/min independently)  
C. Tiered by role: admin 10000/min, operator 1000/min, viewer 100/min  
D. Adaptive by resource type: cheap operations 1000/min (list keys), expensive operations 100/min (encrypt 1GB)  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q14: IP Allowlist Configuration Flexibility

**Context**: AUTH-AUTHZ-MATRIX.md lists IP Allowlist for both browser and headless clients but doesn't specify if allowlist is global or per-client.

**Question**: What is the IP allowlist configuration granularity?

**Options**:

A. Global allowlist - ALL clients must come from allowed IP ranges (strict, simple)  
B. Per-client allowlist - Each client_id has its own allowed IP ranges (flexible, complex)  
C. Per-role allowlist - IP restrictions based on user's role (admin from office IPs only, operators from VPN)  
D. No IP allowlist - Use alternative controls (mTLS, geofencing, device trust) instead  
E. _______________________________________________________________________________

**Your Answer**: ______

---

### Q15: Consent Tracking Granularity

**Context**: AUTH-AUTHZ-MATRIX.md lists Consent Tracking as MEDIUM priority authorization method but doesn't specify granularity (per-scope vs per-resource).

**Question**: What is the consent tracking granularity?

**Options**:

A. Per-scope consent - User consents once to "read:keys" scope, applies to all keys  
B. Per-resource consent - User consents individually to each key ID (key-123, key-456)  
C. Time-limited consent - User consents to scope for 30 days, must re-consent after expiration  
D. Purpose-specific consent - User consents to "read:keys for backup purposes" vs "read:keys for testing"  
E. _______________________________________________________________________________

**Your Answer**: ______

---

**Status**: 15 open questions requiring user input as of 2025-12-22. After user answers, integrate into clarify.md, update constitution.md with architectural decisions, and update spec.md with finalized requirements.
