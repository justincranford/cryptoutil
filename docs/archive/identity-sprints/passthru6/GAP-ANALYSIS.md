# Identity V2 Passthru6 Gap Analysis

## Executive Summary

**CRITICAL FINDING**: Passthru5 achieved 100% requirements coverage and client secret rotation **on paper**, but the OAuth 2.1 Authorization Server is **NOT FUNCTIONAL** for end-to-end flows.

**Root Cause**: Token issuers (JWS, JWE, UUID) are created as empty structs without proper initialization in `cmd/identity/authz/main.go`.

**Impact**: Services start successfully and health checks pass, but ALL OAuth 2.1 endpoints return errors when invoked.

## Gap Categories

### Category 1: Critical Service Initialization Failures

**GAP-001: Token Issuers Not Initialized**

- **Location**: `cmd/identity/authz/main.go` lines 52-54
- **Current Code**:

  ```go
  // TODO: Create JWS, JWE, UUID issuers properly.
  // For now, use placeholders.
  jwsIssuer := &cryptoutilIdentityIssuer.JWSIssuer{}
  jweIssuer := &cryptoutilIdentityIssuer.JWEIssuer{}
  uuidIssuer := &cryptoutilIdentityIssuer.UUIDIssuer{}
  ```

- **Expected Code** (from integration tests):

  ```go
  // Create key rotation manager
  keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
      cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
      &realKeyGenerator{}, // Production key generator
      nil,
  )

  // Initialize signing keys
  err = keyRotationMgr.RotateSigningKey(ctx, "RS256")

  // Create JWS issuer with key rotation
  jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
      config.Tokens.Issuer,
      keyRotationMgr,
      config.Tokens.SigningAlgorithm,
      config.Tokens.AccessTokenTTL,
      config.Tokens.IDTokenTTL,
  )

  // Initialize encryption keys
  err = keyRotationMgr.RotateEncryptionKey(ctx)

  // Create JWE issuer
  jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)

  // Create UUID issuer
  uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
  ```

- **Impact**: Token endpoint returns 401 Unauthorized because issuers cannot sign/encrypt tokens
- **Evidence**: Token test: `POST /oauth2/v1/token` → 401 Unauthorized
- **Blockers**: ALL OAuth flows blocked (authorization code, client credentials, refresh token)

**GAP-002: Missing Production KeyGenerator Implementation**

- **Location**: No production implementation exists
- **Current State**: Only mockKeyGenerator exists in test files
- **Required**: Production implementation that generates real RSA/ECDSA/HMAC keys
- **Reference**: `internal/identity/issuer/key_rotation.go` line 87-90:

  ```go
  type KeyGenerator interface {
      GenerateSigningKey(ctx context.Context, algorithm string) (*SigningKey, error)
      GenerateEncryptionKey(ctx context.Context) (*EncryptionKey, error)
  }
  ```

- **Impact**: Cannot initialize KeyRotationManager in production
- **Blockers**: Cannot create functional JWS/JWE issuers

**GAP-003: OpenAPI Spec Generation Failure**

- **Location**: `internal/identity/authz/routes.go` lines 14-20
- **Current Code**:

  ```go
  swaggerHandler, err := ServeOpenAPISpec()
  if err != nil {
      // Swagger UI is non-critical, skip if spec generation fails.
      // Error already includes context from ServeOpenAPISpec().
      _ = err
  } else {
      app.Get("/ui/swagger/doc.json", swaggerHandler)
  }
  ```

- **Problem**: Error is silently swallowed, no logging
- **Impact**: `/ui/swagger/doc.json` returns 404 with no diagnostic information
- **Evidence**: `GET http://127.0.0.1:8080/ui/swagger/doc.json` → 404 Not Found
- **Required Fix**: Add logging to expose initialization errors

### Category 2: Missing Client Registration/Bootstrap

**GAP-004: No Client Registration Endpoint**

- **OpenAPI Spec**: `/clients` endpoints NOT defined in `api/identity/authz/openapi.yaml`
- **Required Endpoints**:
  - `POST /oauth2/v1/clients` - Register new OAuth client
  - `GET /oauth2/v1/clients/{id}` - Retrieve client details
  - `PUT /oauth2/v1/clients/{id}` - Update client configuration
  - `DELETE /oauth2/v1/clients/{id}` - Deregister client
- **Impact**: No way to create clients for OAuth flows
- **Current Workaround**: Direct database insertion required

**GAP-005: No Bootstrap Client Creation**

- **Location**: No bootstrap logic in service startup
- **Required**: Create default test client on first startup
- **Example Spec**:

  ```yaml
  client_id: demo-client
  client_secret: demo-secret
  redirect_uris: ["http://localhost:3000/callback"]
  grant_types: ["authorization_code", "refresh_token"]
  response_types: ["code"]
  scope: "openid profile email"
  token_endpoint_auth_method: "client_secret_post"
  ```

- **Impact**: Cannot test OAuth flows without manual database operations
- **Blockers**: Demo workflows require working client

### Category 3: Missing Discovery/Metadata Endpoints

**GAP-006: No OAuth 2.1 Authorization Server Metadata**

- **Required Endpoint**: `GET /.well-known/oauth-authorization-server`
- **RFC**: RFC 8414 (OAuth 2.0 Authorization Server Metadata)
- **OpenAPI Spec**: NOT defined in `api/identity/authz/openapi.yaml`
- **Required Response**:

  ```json
  {
    "issuer": "https://authz.cryptoutil.local",
    "authorization_endpoint": "https://authz.cryptoutil.local/oauth2/v1/authorize",
    "token_endpoint": "https://authz.cryptoutil.local/oauth2/v1/token",
    "introspection_endpoint": "https://authz.cryptoutil.local/oauth2/v1/introspect",
    "revocation_endpoint": "https://authz.cryptoutil.local/oauth2/v1/revoke",
    "jwks_uri": "https://authz.cryptoutil.local/oauth2/v1/jwks",
    "grant_types_supported": ["authorization_code", "refresh_token", "client_credentials"],
    "response_types_supported": ["code"],
    "code_challenge_methods_supported": ["S256"],
    "token_endpoint_auth_methods_supported": ["client_secret_post", "client_secret_basic"]
  }
  ```

- **Impact**: Clients cannot auto-discover server capabilities
- **Blockers**: OAuth libraries rely on metadata for configuration

**GAP-007: No JWKS (JSON Web Key Set) Endpoint**

- **Required Endpoint**: `GET /oauth2/v1/jwks`
- **RFC**: RFC 7517 (JSON Web Key - JWK)
- **OpenAPI Spec**: NOT defined in `api/identity/authz/openapi.yaml`
- **Purpose**: Expose public keys for JWT signature verification
- **Required Response**:

  ```json
  {
    "keys": [
      {
        "kty": "RSA",
        "use": "sig",
        "kid": "key-2025-01-15-123456",
        "alg": "RS256",
        "n": "...",
        "e": "AQAB"
      }
    ]
  }
  ```

- **Impact**: Clients cannot verify JWT signatures
- **Blockers**: JWS token validation impossible

**GAP-008: No OpenID Connect Discovery**

- **Required Endpoint**: `GET /.well-known/openid-configuration`
- **RFC**: OpenID Connect Discovery 1.0
- **OpenAPI Spec**: NOT defined in `api/identity/idp/openapi.yaml`
- **Required Response**: Superset of OAuth metadata + OIDC-specific claims
- **Impact**: OIDC clients cannot auto-discover endpoints
- **Blockers**: OIDC libraries require discovery for initialization

### Category 4: Incomplete OAuth Flow Implementation

**GAP-009: Authorization Endpoint Returns Errors**

- **Endpoints**:
  - `GET /oauth2/v1/authorize`
  - `POST /oauth2/v1/authorize`
- **Expected Behavior**: Redirect to login page or return authorization code
- **Likely Status**: Untested (requires functional issuers)
- **Testing Required**:
  1. Valid authorization request with PKCE
  2. Invalid client_id
  3. Missing redirect_uri
  4. Invalid code_challenge_method

**GAP-010: Token Endpoint Client Authentication**

- **Current Status**: Returns 401 Unauthorized
- **Required**: Support all authentication methods from config:
  - `client_secret_post` (credentials in request body)
  - `client_secret_basic` (HTTP Basic auth)
  - `client_secret_jwt` (JWT assertion)
  - `private_key_jwt` (JWT assertion with private key)
  - `tls_client_auth` (mutual TLS)
  - `self_signed_tls_client_auth` (self-signed certificate)
- **Current Config**: All methods listed in `client_auth_methods` but untested
- **Testing Required**: Verify each authentication method works

**GAP-011: PKCE Validation Not Tested**

- **Components**:
  - Code challenge generation
  - Code verifier validation
  - S256 transformation verification
- **Code Exists**: `internal/identity/authz/pkce/` package implemented
- **Status**: Unit tests exist, integration tests missing
- **Testing Required**: End-to-end PKCE flow with real token issuance

### Category 5: Service-to-Service Integration Gaps

**GAP-012: IdP Service Not Connected to AuthZ**

- **Current State**: IdP runs independently on port 8081
- **Missing**: Integration points between IdP and AuthZ:
  - User authentication callback
  - Consent screen integration
  - Session management coordination
- **Impact**: Authorization code flow cannot complete (no user login)
- **Required**: Define integration contract and implement coordination

**GAP-013: Resource Server Token Validation**

- **Current State**: RS service exists but token validation untested
- **Required Tests**:
  1. Valid access token → 200 OK
  2. Expired access token → 401 Unauthorized
  3. Invalid signature → 401 Unauthorized
  4. Missing token → 401 Unauthorized
- **Blockers**: Requires functional JWS issuer to create test tokens

### Category 6: Configuration and Deployment Gaps

**GAP-014: Config Validation Incomplete**

- **Fixed in Passthru5**: Added refresh_token_format, id_token_format, signing_algorithm
- **Still Missing**: Runtime validation of token issuer initialization
- **Required**: Startup checks to verify:
  1. Key rotation manager initialized
  2. At least one signing key available
  3. At least one encryption key available
  4. Token service functional

**GAP-015: No Development Seed Data**

- **Current State**: Empty database on startup
- **Required**: Seed data for development:
  - Bootstrap OAuth client (see GAP-005)
  - Test user accounts
  - Sample scopes and permissions
  - Test sessions
- **Impact**: Manual database operations required for testing

### Category 7: Documentation and Demo Gaps

**GAP-016: Demo Guide References Non-Functional Endpoints**

- **Document**: `docs/02-identityV2/IDENTITY-V2-DEMO.md`
- **Problem**: Guide documents endpoints that return errors
- **Required**: Update guide after endpoints are functional
- **Include**: Working curl examples with expected responses

**GAP-017: No End-to-End OAuth Flow Example**

- **Missing**: Complete walkthrough of authorization code flow:
  1. Client registration
  2. Authorization request with PKCE
  3. User authentication (IdP)
  4. Authorization code issuance
  5. Token exchange with code_verifier
  6. Access token usage
  7. Token refresh
- **Required**: Documented example with real requests/responses

## Test Coverage Analysis

### Existing Test Infrastructure

**Strong Test Coverage** (Unit Tests):

- `internal/identity/authz/*_test.go` - 80%+ coverage for handlers
- `internal/identity/issuer/*_test.go` - 95%+ coverage for token issuers
- `internal/identity/repository/*_test.go` - 85%+ coverage for data access

**Integration Tests** (Partial):

- `internal/identity/integration/integration_test.go` - 848 lines
- Uses legacy JWS issuer for testing (not production KeyRotationManager)
- Tests server startup and basic flows
- **GAP**: Not using same initialization as production

**Missing E2E Tests**:

- No tests invoking `/oauth2/v1/authorize` with real requests
- No tests for full authorization code flow
- No tests for client credentials flow
- No tests for refresh token flow

### Test Evidence Required for Passthru6 Completion

Each task MUST include test evidence:

1. **Token Issuance**: `POST /oauth2/v1/token` → 200 OK with valid JWT
2. **Authorization Code Flow**: Complete flow end-to-end → 200 OK
3. **Client Credentials**: `grant_type=client_credentials` → 200 OK
4. **Token Introspection**: `POST /oauth2/v1/introspect` → 200 OK with active=true
5. **Token Revocation**: `POST /oauth2/v1/revoke` → 200 OK
6. **JWKS Endpoint**: `GET /oauth2/v1/jwks` → 200 OK with valid keys
7. **OAuth Metadata**: `GET /.well-known/oauth-authorization-server` → 200 OK

## Summary Statistics

| Category | Total Gaps | Critical | High | Medium |
|----------|------------|----------|------|--------|
| Service Initialization | 3 | 3 | 0 | 0 |
| Client Registration | 2 | 2 | 0 | 0 |
| Discovery/Metadata | 3 | 0 | 3 | 0 |
| OAuth Flows | 3 | 1 | 2 | 0 |
| Integration | 2 | 1 | 1 | 0 |
| Configuration | 2 | 0 | 1 | 1 |
| Documentation | 2 | 0 | 0 | 2 |
| **TOTAL** | **17** | **7** | **7** | **3** |

## Severity Definitions

- **CRITICAL**: Blocks ALL OAuth flows (e.g., uninitialized issuers)
- **HIGH**: Blocks specific flows or major features (e.g., missing metadata)
- **MEDIUM**: Usability/testing issues (e.g., missing seed data)

## Passthru6 Scope Recommendation

**Phase 1: Core Service Functionality** (CRITICAL - Must Complete)

- GAP-001: Initialize token issuers properly
- GAP-002: Implement production KeyGenerator
- GAP-003: Fix OpenAPI spec logging
- GAP-005: Create bootstrap client

**Phase 2: OAuth Flow Validation** (CRITICAL - Must Complete)

- GAP-010: Test token endpoint with all auth methods
- GAP-011: Validate PKCE end-to-end
- GAP-013: Test resource server token validation
- Create E2E test for authorization code flow

**Phase 3: Discovery and Metadata** (HIGH - Should Complete)

- GAP-006: Implement OAuth metadata endpoint
- GAP-007: Implement JWKS endpoint
- GAP-008: Implement OIDC discovery endpoint

**Phase 4: Client Management** (HIGH - Should Complete if time permits)

- GAP-004: Implement client registration endpoints
- GAP-015: Add development seed data

**Phase 5: Integration and Documentation** (MEDIUM - Nice to Have)

- GAP-012: Connect IdP to AuthZ
- GAP-016: Update demo guide with working examples
- GAP-017: Create comprehensive OAuth flow documentation

## Success Criteria for Passthru6

**Minimum Viable Product**:

1. ✅ Token endpoint returns 200 OK with valid JWT for client_credentials grant
2. ✅ Authorization code flow completes end-to-end (with test client)
3. ✅ JWKS endpoint returns public keys
4. ✅ OAuth metadata endpoint returns server capabilities
5. ✅ All critical endpoints tested with evidence

**Stretch Goals**:

1. ✅ Client registration endpoints functional
2. ✅ IdP integration working for user authentication
3. ✅ Complete demo guide with working curl examples

## Next Steps

1. **Create Passthru6 Master Plan** with prioritized tasks
2. **Implement Phase 1** (Core Service Functionality)
3. **Test each endpoint** with evidence-based validation
4. **Update documentation** with working examples
5. **Create comprehensive E2E test suite**

---

**Document Status**: Gap Analysis Complete  
**Date**: 2025-01-15  
**Author**: GitHub Copilot  
**Next Action**: Create `07-PASSTHRU6-MASTER-PLAN.md`
