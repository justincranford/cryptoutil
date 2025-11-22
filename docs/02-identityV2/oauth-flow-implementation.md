# OAuth 2.1 Authorization Code Flow Implementation

## Overview

This document describes the OAuth 2.1 authorization code flow implementation with PKCE (Proof Key for Code Exchange) in the cryptoutil identity server.

## Architecture

### Service Components

1. **AuthZ Server** (Authorization Server)
   - Endpoint: `/oauth2/v1/authorize` - Authorization code initiation
   - Endpoint: `/oauth2/v1/token` - Token exchange
   - Endpoint: `/health` - Health check
   - Port: 8080 (HTTP), 8443 (HTTPS production)

2. **IdP Server** (Identity Provider)
   - Endpoint: `/oidc/v1/login` - User authentication
   - Endpoint: `/oidc/v1/consent` - User consent
   - Endpoint: `/health` - Health check
   - Port: 8081 (HTTP), 8444 (HTTPS production)

3. **RS Server** (Resource Server)
   - Protected endpoints requiring Bearer tokens
   - OAuth 2.0 token validation
   - Scope-based access control
   - Port: 8082 (HTTP), 8445 (HTTPS production)

## OAuth 2.1 Authorization Code Flow

### Step 1: Authorization Request

**Endpoint**: `GET /oauth2/v1/authorize`

**Required Parameters**:
- `response_type=code` - Must be "code" (OAuth 2.1 removes implicit flow)
- `client_id` - Client identifier
- `redirect_uri` - Callback URL (must match registered URI)
- `code_challenge` - PKCE code challenge (SHA-256 hash of verifier)
- `code_challenge_method=S256` - Only S256 supported (OAuth 2.1 requirement)

**Optional Parameters**:
- `scope` - Requested scopes (space-separated)
- `state` - CSRF protection token

**Response**: 302 redirect to `redirect_uri` with:
- `code` - Authorization code (single-use, expires in 5 minutes)
- `state` - Preserved from request (if provided)

**Example Request**:
```http
GET /oauth2/v1/authorize?
  response_type=code&
  client_id=test-client&
  redirect_uri=https://localhost:3000/callback&
  scope=openid profile email&
  state=xyz123&
  code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&
  code_challenge_method=S256
```

**Example Response**:
```http
HTTP/1.1 302 Found
Location: https://localhost:3000/callback?code=ZG5Bb3eVyGxgBSEJXY8ejPLXa9DeizcOUcyFySyxKsc&state=xyz123
```

### Step 2: Token Exchange

**Endpoint**: `POST /oauth2/v1/token`

**Required Parameters** (application/x-www-form-urlencoded):
- `grant_type=authorization_code`
- `code` - Authorization code from Step 1
- `redirect_uri` - Must match original request
- `client_id` - Client identifier
- `client_secret` - Client secret (confidential clients)
- `code_verifier` - PKCE code verifier (proves possession)

**Response** (application/json):
```json
{
  "access_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "019aa8d8-...",
  "id_token": "eyJhbGc..." (if openid scope requested)
}
```

**Example Request**:
```http
POST /oauth2/v1/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=ZG5Bb3eVyGxgBSEJXY8ejPLXa9DeizcOUcyFySyxKsc&
redirect_uri=https://localhost:3000/callback&
client_id=test-client&
client_secret=test-secret&
code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk
```

### Step 3: Protected Resource Access

**Endpoint**: `GET /api/v1/protected/resource`

**Required Header**:
- `Authorization: Bearer <access_token>`

**Response**: Protected resource data (if token valid and scopes sufficient)

## PKCE (Proof Key for Code Exchange)

### Why PKCE?

OAuth 2.1 mandates PKCE for all clients (public and confidential) to prevent authorization code interception attacks.

### Code Verifier Generation

```javascript
// Generate random 43-128 character string
const codeVerifier = base64url(crypto.randomBytes(32))
// Example: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
```

### Code Challenge Generation

```javascript
// SHA-256 hash of verifier, base64url encoded
const codeChallenge = base64url(sha256(codeVerifier))
// Example: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
```

### PKCE Validation Flow

1. Client generates `code_verifier` and `code_challenge`
2. Client sends `code_challenge` in authorization request
3. AuthZ server stores challenge with authorization code
4. Client sends `code_verifier` in token request
5. AuthZ server validates: `sha256(code_verifier) == stored_code_challenge`
6. If valid, issue tokens; if invalid, reject with `invalid_grant` error

## Session Management

### Current Implementation (Task 10.5)

**Auto-consent for integration tests**:
- Authorization codes generated immediately
- No user authentication required
- No consent screen shown

**Storage**:
- In-memory authorization request store
- Thread-safe with sync.RWMutex
- Authorization codes expire in 5 minutes
- Single-use codes (deleted after exchange)

### Future Implementation (Task 10.6+)

**User authentication flow**:
1. Check session cookie for authenticated user
2. If not authenticated: 302 redirect to IdP `/oidc/v1/login`
3. IdP authenticates user (username/password, MFA, etc.)
4. IdP creates session, sets secure HttpOnly cookie
5. IdP redirects back to AuthZ `/oauth2/v1/authorize` with session
6. AuthZ shows consent screen (if required)
7. User approves scopes
8. AuthZ generates authorization code
9. AuthZ redirects to client callback with code

**Session storage**:
- Database-backed session repository
- Secure session tokens (UUIDv7)
- Configurable lifetime (default 1 hour)
- Idle timeout (default 15 minutes)

## CSRF Protection

### State Parameter

**Purpose**: Prevents cross-site request forgery attacks

**Flow**:
1. Client generates random state: `crypto.randomBytes(16).toString('hex')`
2. Client stores state in session storage
3. Client includes state in authorization request
4. AuthZ echoes state in redirect response
5. Client validates state matches stored value
6. If mismatch, reject response as potential attack

**Example**:
```javascript
// Before authorization request
const state = crypto.randomBytes(16).toString('hex')
sessionStorage.setItem('oauth_state', state)

// After redirect callback
const returnedState = new URL(window.location).searchParams.get('state')
const storedState = sessionStorage.getItem('oauth_state')
if (returnedState !== storedState) {
  throw new Error('State mismatch - potential CSRF attack')
}
```

## Error Handling

### OAuth 2.1 Error Codes

#### Authorization Endpoint Errors

**invalid_request**: Missing or invalid required parameter
```json
{
  "error": "invalid_request",
  "error_description": "code_challenge is required (OAuth 2.1 requires PKCE)"
}
```

**unauthorized_client**: Client not authorized for requested grant type
```json
{
  "error": "unauthorized_client",
  "error_description": "Client not authorized to use authorization_code grant"
}
```

**access_denied**: User or authorization server denied request
```json
{
  "error": "access_denied",
  "error_description": "User denied consent"
}
```

**unsupported_response_type**: Response type not supported
```json
{
  "error": "unsupported_response_type",
  "error_description": "Only 'code' response_type is supported (OAuth 2.1)"
}
```

**invalid_scope**: Requested scope invalid or exceeds granted scopes
```json
{
  "error": "invalid_scope",
  "error_description": "Scope 'admin' not allowed for this client"
}
```

**server_error**: Internal server error
```json
{
  "error": "server_error",
  "error_description": "Failed to generate authorization code"
}
```

#### Token Endpoint Errors

**invalid_grant**: Authorization code invalid, expired, or already used
```json
{
  "error": "invalid_grant",
  "error_description": "Authorization code expired or already used"
}
```

**invalid_client**: Client authentication failed
```json
{
  "error": "invalid_client",
  "error_description": "Client authentication failed"
}
```

**unsupported_grant_type**: Grant type not supported
```json
{
  "error": "unsupported_grant_type",
  "error_description": "Only authorization_code, refresh_token, client_credentials supported"
}
```

### Error Response Patterns

**Authorization Endpoint**:
- Parameter validation errors: 400 Bad Request with JSON error
- Redirect URI validation errors: 400 Bad Request (NO redirect for security)
- After successful client validation: Redirect to redirect_uri with error query params

**Token Endpoint**:
- All errors: 400/401 status with JSON error response
- Never redirect (token endpoint is POST, not GET)

## Security Considerations

### Authorization Code Security

- **Single-use**: Codes deleted immediately after exchange
- **Short lifetime**: 5 minute expiration
- **PKCE required**: Prevents code interception attacks
- **Redirect URI validation**: Exact match against registered URIs
- **Secure generation**: crypto/rand with 32 bytes entropy

### Redirect URI Validation

**Exact match required** - no wildcards, no regex:
```go
validRedirectURI := false
for _, uri := range client.RedirectURIs {
    if uri == redirectURI {
        validRedirectURI = true
        break
    }
}
```

**Why strict validation?**
- Prevents open redirect vulnerabilities
- Prevents authorization code theft
- OAuth 2.1 security best practice

### Client Authentication

**Confidential Clients** (backend applications):
- `client_secret_basic` - HTTP Basic authentication
- `client_secret_post` - POST body authentication
- `client_secret_jwt` - JWT signed with client secret
- `private_key_jwt` - JWT signed with private key
- `tls_client_auth` - mTLS with CA certificate
- `self_signed_tls_client_auth` - mTLS with pinned certificate

**Public Clients** (SPAs, mobile apps):
- No client secret
- PKCE mandatory for security
- Registered redirect URIs strictly enforced

## Token Formats

### Access Tokens

**Format**: JWS (JSON Web Signature)

**Claims**:
```json
{
  "sub": "user-uuid",
  "client_id": "client-id",
  "scope": "openid profile email",
  "exp": 1700000000,
  "iat": 1699996400,
  "iss": "https://authz.example.com",
  "aud": "https://api.example.com"
}
```

**Lifetime**: 1 hour (configurable per client)

### Refresh Tokens

**Format**: UUIDv7 (time-ordered UUID)

**Storage**: Database with rotation support

**Lifetime**: 24 hours (configurable per client)

### ID Tokens

**Format**: JWS (OIDC standard)

**Claims**:
```json
{
  "sub": "user-uuid",
  "aud": "client-id",
  "exp": 1700000000,
  "iat": 1699996400,
  "iss": "https://idp.example.com",
  "name": "John Doe",
  "email": "john@example.com",
  "email_verified": true
}
```

**Lifetime**: 1 hour (configurable per client)

## Configuration

### AuthZ Server Configuration

```yaml
authz:
  name: "production-authz"
  bind_address: "0.0.0.0"
  port: 8443
  tls_enabled: true
  tls_cert_file: "/etc/certs/authz.crt"
  tls_key_file: "/etc/certs/authz.key"

tokens:
  access_token_format: "jws"  # jws, jwe, or uuid
  issuer: "https://authz.example.com"
  access_token_lifetime: 3600
  refresh_token_lifetime: 86400
  id_token_lifetime: 3600

database:
  type: "postgres"
  dsn: "postgres://user:pass@localhost:5432/identity"
  max_open_conns: 25
  max_idle_conns: 5
```

### Client Registration

```json
{
  "client_id": "web-app",
  "client_secret": "secret-value",
  "client_type": "confidential",
  "redirect_uris": [
    "https://app.example.com/callback",
    "https://app.example.com/oauth/callback"
  ],
  "allowed_grant_types": [
    "authorization_code",
    "refresh_token"
  ],
  "allowed_response_types": ["code"],
  "allowed_scopes": [
    "openid",
    "profile",
    "email",
    "read:resource",
    "write:resource"
  ],
  "token_endpoint_auth_method": "client_secret_post",
  "require_pkce": true,
  "pkce_challenge_method": "S256",
  "access_token_lifetime": 3600,
  "refresh_token_lifetime": 86400,
  "id_token_lifetime": 3600
}
```

## Testing

### Integration Tests

**File**: `internal/identity/integration/integration_test.go`

**TestOAuth2AuthorizationCodeFlow**:
1. Generates PKCE code verifier and challenge
2. Requests authorization code from `/oauth2/v1/authorize`
3. Validates 302 redirect with code and state
4. Exchanges code for tokens at `/oauth2/v1/token`
5. Validates access_token, refresh_token, id_token
6. Uses access_token to access protected resource

**TestHealthCheckEndpoints**:
- Validates AuthZ, IdP, RS health endpoints return 200 OK
- Validates JSON response format

**TestResourceServerScopeEnforcement**:
- Validates scope-based access control
- Tests read, write, delete, admin scopes
- Validates 403 Forbidden for insufficient scopes

**TestUnauthorizedAccess**:
- Validates 401 Unauthorized for missing Authorization header
- Tests protected and admin endpoints

**TestGracefulShutdown**:
- Validates servers shut down cleanly
- Validates connections fail after shutdown

### Manual Testing with SPA Relying Party

**Start services**:
```bash
docker compose -f deployments/compose/identity-compose.yml up -d
```

**Access SPA**: http://localhost:8446

**Test flow**:
1. Configure AuthZ URL, IdP URL, client ID, redirect URI
2. Click "Login with OAuth 2.1"
3. Verify redirect to callback with authorization code
4. Verify token exchange succeeds
5. Verify access token works for protected resources

## References

- [OAuth 2.1 Draft](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-15)
- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
- [RFC 6749 - OAuth 2.0](https://datatracker.ietf.org/doc/html/rfc6749)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
