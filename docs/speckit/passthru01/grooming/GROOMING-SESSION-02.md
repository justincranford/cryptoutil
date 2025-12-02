# Grooming Session 02: Identity OAuth 2.1 and OIDC Deep Dive

## Overview

- **Focus Area**: OAuth 2.1 Authorization Server, OIDC Identity Provider, authentication flows
- **Related Spec Section**: Spec P2: Identity, OpenAPI authz/idp specifications
- **Prerequisites**: Session 01 completed, understanding of OAuth 2.0/2.1, OIDC concepts

---

## Questions

### Q1: What is the REQUIRED parameter for OAuth 2.1 authorization requests that was optional in OAuth 2.0?

A) state
B) nonce
C) code_challenge (PKCE)
D) redirect_uri

**Answer**: C
**Explanation**: OAuth 2.1 mandates PKCE (code_challenge and code_challenge_method) for ALL authorization code flows. State is recommended but not required. Nonce is OIDC-specific. Redirect_uri was already required in practice.

---

### Q2: Which grant type is explicitly removed in OAuth 2.1?

A) authorization_code
B) client_credentials
C) implicit
D) refresh_token

**Answer**: C
**Explanation**: OAuth 2.1 removes the implicit grant and resource owner password credentials (ROPC) grant. Authorization code with PKCE replaces implicit for SPAs.

---

### Q3: What is the correct endpoint for exchanging an authorization code for tokens?

A) `/oauth2/v1/authorize`
B) `/oauth2/v1/token`
C) `/oauth2/v1/exchange`
D) `/oidc/v1/token`

**Answer**: B
**Explanation**: The token endpoint `/oauth2/v1/token` handles code exchange, refresh token grants, and client_credentials grants.

---

### Q4: What authentication method sends credentials in the HTTP Authorization header?

A) client_secret_post
B) client_secret_basic
C) client_secret_jwt
D) private_key_jwt

**Answer**: B
**Explanation**: client_secret_basic uses HTTP Basic Auth: `Authorization: Basic base64(client_id:client_secret)`. client_secret_post sends credentials in the request body.

---

### Q5: What is the minimum length for PKCE code_challenge?

A) 32 characters
B) 43 characters
C) 64 characters
D) 128 characters

**Answer**: B
**Explanation**: code_challenge must be 43-128 characters (base64url-encoded SHA256 hash of code_verifier).

---

### Q6: What does the `/oauth2/v1/introspect` endpoint do?

A) Revokes tokens
B) Validates and returns token metadata
C) Issues new tokens
D) Refreshes tokens

**Answer**: B
**Explanation**: Token introspection (RFC 7662) validates tokens and returns metadata including active status, scopes, and expiration.

---

### Q7: Which endpoint is used for token revocation?

A) `/oauth2/v1/introspect`
B) `/oauth2/v1/revoke`
C) `/oauth2/v1/token/delete`
D) `/oauth2/v1/logout`

**Answer**: B
**Explanation**: `/oauth2/v1/revoke` implements RFC 7009 token revocation for access and refresh tokens.

---

### Q8: What HTTP status code indicates a successful redirect in OAuth authorization?

A) 200 OK
B) 201 Created
C) 302 Found
D) 400 Bad Request

**Answer**: C
**Explanation**: 302 Found redirects the user agent to the IdP login (if not authenticated) or to the redirect_uri with the authorization code.

---

### Q9: What is the current status of the login UI endpoint?

A) ✅ Fully Working with HTML UI
B) ⚠️ API Only (No UI)
C) ❌ Not Implemented
D) 🔄 In Progress

**Answer**: B
**Explanation**: `/oidc/v1/login` is ⚠️ API Only (No UI) - it processes credentials but returns JSON, not HTML forms.

---

### Q10: Which scope is required for OpenID Connect to return an ID token?

A) profile
B) email
C) openid
D) offline_access

**Answer**: C
**Explanation**: The `openid` scope is REQUIRED for OIDC flows. Without it, the flow is pure OAuth 2.1 without ID tokens.

---

### Q11: What does the `/.well-known/jwks.json` endpoint return?

A) OpenID Connect discovery document
B) JSON Web Key Set for token verification
C) List of valid redirect URIs
D) Client registration information

**Answer**: B
**Explanation**: JWKS endpoint returns the public keys used to verify JWT signatures (access tokens, ID tokens).

---

### Q12: Which grant type is used for machine-to-machine authentication?

A) authorization_code
B) implicit
C) client_credentials
D) refresh_token

**Answer**: C
**Explanation**: client_credentials grant is for server-to-server/machine-to-machine authentication without user involvement.

---

### Q13: What happens when an authorization request is received for an unauthenticated user?

A) 401 Unauthorized response
B) Redirect to IdP login
C) 400 Bad Request response
D) Empty 200 OK response

**Answer**: B
**Explanation**: The authorization endpoint redirects (302) to the IdP login URL with a return_url parameter for flow continuation.

---

### Q14: What is the purpose of the `state` parameter in OAuth?

A) Store user preferences
B) CSRF protection
C) Session management
D) Token caching

**Answer**: B
**Explanation**: The state parameter is opaque value for CSRF protection. The client should verify it matches when receiving the callback.

---

### Q15: Which authentication flow is verified as fully working in cryptoutil?

A) Implicit flow
B) Authorization code flow with PKCE
C) Resource owner password credentials
D) Device authorization grant

**Answer**: B
**Explanation**: Authorization code flow with PKCE is ✅ verified working through the complete flow: authorize → login → consent → token exchange.

---

### Q16: What type of tokens does the token endpoint return?

A) Only access_token
B) access_token and refresh_token
C) Only id_token
D) access_token, refresh_token, and id_token (when openid scope)

**Answer**: D
**Explanation**: Token endpoint returns access_token and refresh_token. When openid scope is included, id_token is also returned.

---

### Q17: What is the current status of the consent UI endpoint?

A) ✅ Fully Working with HTML UI
B) ⚠️ API Only (No UI)
C) ❌ Not Implemented
D) ✅ Working with Remembered Consent

**Answer**: B
**Explanation**: `/oidc/v1/consent` is ⚠️ API Only (No UI) - similar to login, it processes consent but lacks HTML interface.

---

### Q18: Which endpoint provides user claims based on access token?

A) `/oidc/v1/login`
B) `/oidc/v1/consent`
C) `/oidc/v1/userinfo`
D) `/oauth2/v1/introspect`

**Answer**: C
**Explanation**: `/oidc/v1/userinfo` returns user claims (name, email, etc.) based on the scopes in the access token.

---

### Q19: What is the demo user credentials for testing?

A) admin/admin
B) test/test
C) demo/demo-password
D) user/password

**Answer**: C
**Explanation**: Demo user bootstrap creates demo/demo-password for testing the OAuth/OIDC flows.

---

### Q20: Which secret rotation feature ensures smooth client transitions?

A) Immediate rotation
B) Grace period support
C) Hard cutover
D) Parallel secrets

**Answer**: B
**Explanation**: Grace period support allows configurable overlap where both old and new secrets are valid during rotation.

---

### Q21: What does the ClientSecretVersion model enable?

A) Single secret per client
B) Multiple secret versions per client
C) Encrypted storage only
D) Secret expiration only

**Answer**: B
**Explanation**: ClientSecretVersion enables multiple secret versions per client for rotation without service interruption.

---

### Q22: Which NIST standard does the secret rotation system demonstrate compliance with?

A) NIST SP 800-53
B) NIST SP 800-57
C) NIST SP 800-63
D) NIST SP 800-171

**Answer**: B
**Explanation**: NIST SP 800-57 covers key lifecycle management including rotation, which the secret rotation system demonstrates.

---

### Q23: What does the KeyRotationEvent model track?

A) Encryption operations
B) Audit trail for rotation events
C) Key generation timestamps
D) Client authentication attempts

**Answer**: B
**Explanation**: KeyRotationEvent provides audit trail for rotation events for compliance and troubleshooting.

---

### Q24: What is the current status of client_secret_jwt authentication?

A) ✅ Working
B) ⚠️ Not Tested
C) ❌ Not Implemented
D) 🔄 In Progress

**Answer**: B
**Explanation**: client_secret_jwt implementation exists but is ⚠️ Not Tested/Validated for production use.

---

### Q25: Which endpoint does NOT require authentication?

A) `/oauth2/v1/token` (authorization_code grant)
B) `/oauth2/v1/introspect`
C) `/oauth2/v1/authorize` (initial request)
D) `/oauth2/v1/revoke`

**Answer**: C
**Explanation**: The initial `/oauth2/v1/authorize` request doesn't require authentication - it initiates the flow that will authenticate the user.

---

### Q26: What is the current port for identity-authz service?

A) 8080
B) 8081
C) 8090
D) 8091

**Answer**: C
**Explanation**: identity-authz runs on port 8090, identity-idp on 8091, to avoid conflicts with KMS services on 8080-8082.

---

### Q27: What error code indicates an invalid PKCE code_verifier?

A) invalid_request
B) invalid_grant
C) invalid_client
D) server_error

**Answer**: B
**Explanation**: invalid_grant indicates the authorization code is invalid or expired, or the code_verifier doesn't match the original code_challenge.

---

### Q28: Which response type is required for authorization code flow?

A) token
B) code
C) id_token
D) code id_token

**Answer**: B
**Explanation**: response_type=code is required for authorization code flow. "token" was implicit flow (removed in OAuth 2.1).

---

### Q29: What is the purpose of the IntBool type in Identity models?

A) Store integers as booleans
B) Cross-database bool↔INTEGER compatibility
C) Optimize storage space
D) Enable JSON serialization

**Answer**: B
**Explanation**: IntBool provides cross-database bool↔INTEGER compatibility between PostgreSQL (native bool) and SQLite (INTEGER 0/1).

---

### Q30: What must the redirect_uri in token request match?

A) Any registered redirect URI
B) The redirect_uri from the authorization request
C) The client's primary redirect URI
D) Any URI on the same domain

**Answer**: B
**Explanation**: The redirect_uri in token request must EXACTLY match the one used in the authorization request for security.

---

### Q31: Which MFA factor uses time-based codes that change every 30 seconds?

A) Email OTP
B) SMS OTP
C) TOTP
D) Passkey

**Answer**: C
**Explanation**: TOTP (Time-based One-Time Password) generates codes that typically change every 30 seconds using a shared secret.

---

### Q32: What is required for OIDC logout to work correctly?

A) Clear server-side session only
B) Revoke tokens only
C) Clear session AND revoke tokens AND redirect
D) Send logout notification only

**Answer**: C
**Explanation**: Complete logout requires clearing server-side session, revoking associated tokens, and redirecting to post-logout URI.

---

### Q33: Which standard defines the token introspection endpoint?

A) RFC 6749
B) RFC 7009
C) RFC 7662
D) RFC 8628

**Answer**: C
**Explanation**: RFC 7662 defines OAuth 2.0 Token Introspection. RFC 7009 is Token Revocation. RFC 6749 is OAuth 2.0 core.

---

### Q34: What is the current production readiness status of Identity?

A) ❌ NOT READY
B) ⚠️ PARTIAL (with documented limitations)
C) ✅ PRODUCTION READY
D) 🔄 BETA

**Answer**: C
**Explanation**: Identity is ✅ PRODUCTION READY (Core OAuth 2.1 + Secret Rotation + Demo Working) with documented limitations.

---

### Q35: What happens when token cleanup job is not running?

A) Tokens never expire
B) Database grows unbounded with expired tokens
C) Authentication fails
D) New tokens cannot be issued

**Answer**: B
**Explanation**: Without cleanup jobs, expired and revoked tokens accumulate in the database, causing storage growth and potential performance issues.

---

### Q36: Which scope grants access to user's email address?

A) openid
B) profile
C) email
D) address

**Answer**: C
**Explanation**: The `email` scope grants access to email and email_verified claims. `profile` grants name, nickname, etc.

---

### Q37: What authentication is required for the token endpoint with authorization_code grant?

A) No authentication required
B) Client authentication (secret or JWT)
C) User authentication only
D) Bearer token only

**Answer**: B
**Explanation**: Token endpoint requires client authentication via client_secret_basic, client_secret_post, or JWT-based methods.

---

### Q38: What is the status of front-channel logout?

A) ✅ Implemented
B) ⚠️ Partial
C) ❌ Not Implemented (in plan)
D) 🔄 In Progress

**Answer**: C
**Explanation**: Front-channel logout is in the plan (Phase 1, task 1.3.4) but not yet implemented.

---

### Q39: Which error is returned for invalid client credentials?

A) invalid_request
B) invalid_client
C) invalid_grant
D) unauthorized_client

**Answer**: B
**Explanation**: invalid_client indicates client authentication failed - wrong client_id or client_secret.

---

### Q40: What is the purpose of the return_url parameter in login redirect?

A) Store the original page URL
B) Enable flow continuation after login
C) Track referrer information
D) Cache authorization parameters

**Answer**: B
**Explanation**: return_url enables the flow to continue after login - the user is redirected back to complete authorization.

---

### Q41: Which claims are included with the profile scope?

A) email, email_verified
B) name, nickname, preferred_username, profile, picture, website, gender, birthdate, zoneinfo, locale, updated_at
C) address
D) phone_number, phone_number_verified

**Answer**: B
**Explanation**: The profile scope includes name, nickname, preferred_username, profile, picture, website, gender, birthdate, zoneinfo, locale, and updated_at claims.

---

### Q42: What is the current status of WebAuthn/FIDO2 Passkey MFA?

A) ✅ Fully Working
B) ⚠️ Partial
C) ❌ Not Implemented
D) 🔄 In Progress

**Answer**: B
**Explanation**: Passkey/WebAuthn is ⚠️ Partial - some implementation exists but not fully complete.

---

### Q43: What token type prefix is used for Bearer tokens?

A) Token
B) Bearer
C) JWT
D) OAuth

**Answer**: B
**Explanation**: Bearer tokens use the "Bearer" prefix in the Authorization header: `Authorization: Bearer <access_token>`.

---

### Q44: Which endpoint should be checked to verify IdP health?

A) `/health`
B) `/livez`
C) `/readyz`
D) All of the above

**Answer**: D
**Explanation**: Health can be checked via `/health` (public), `/livez` (liveness), or `/readyz` (readiness) depending on the use case.

---

### Q45: What is the expected behavior when consent is denied?

A) Redirect with error=access_denied
B) Return 403 Forbidden
C) Return to login page
D) Ignore and continue

**Answer**: A
**Explanation**: When user denies consent, the flow redirects to redirect_uri with error=access_denied per OAuth spec.

---

### Q46: Which database stores Identity tokens and sessions?

A) PostgreSQL only
B) SQLite only
C) Both PostgreSQL and SQLite supported
D) Redis

**Answer**: C
**Explanation**: Identity supports both PostgreSQL (production) and SQLite (development/testing) with identical behavior.

---

### Q47: What does the offline_access scope enable?

A) Offline user authentication
B) Refresh token issuance
C) Long-lived access tokens
D) Cached consent

**Answer**: B
**Explanation**: The offline_access scope indicates the client wants a refresh token for obtaining new access tokens without user interaction.

---

### Q48: What is the port conflict that was fixed for identity-idp?

A) 8080 → 8081
B) 8081 → 8091
C) 8090 → 8091
D) 5432 → 5433

**Answer**: B
**Explanation**: IDP port was changed from 8081 to 8091 to avoid conflict with KMS cryptoutil-postgres-1 on 8081.

---

### Q49: Which authentication profile registration is a known TODO?

A) client_secret_basic
B) client_secret_post
C) Auth profile registration for custom profiles
D) None, all are complete

**Answer**: C
**Explanation**: Auth profile registration is a documented MEDIUM severity TODO for allowing custom authentication profiles.

---

### Q50: What percentage of original Identity tasks (01-20) are fully complete?

A) 25%
B) 45%
C) 65%
D) 85%

**Answer**: B
**Explanation**: 9/20 tasks fully complete = 45% completion. 5 are partial, 6 are incomplete.

---

## Session Summary

**Topics Covered**:

- OAuth 2.1 authorization code flow with PKCE
- Token endpoint operations (exchange, refresh, client_credentials)
- OIDC endpoints (login, consent, logout, userinfo)
- Authentication methods (basic, post, jwt)
- Secret rotation system
- MFA factors and status
- Production readiness assessment

**Next Session**: GROOMING-SESSION-03 - KMS Key Management Deep Dive
