# DEMO-IDENTITY: Identity-Only Working Demo

**Purpose**: Fix LLM-generated Identity code to working state
**Priority**: HIGH - After 6 passthrus, this MUST work
**Timeline**: Day 2-6

---

## Current State Assessment

Identity server has been through 6 LLM passthrus with mixed results:

### Package Structure (Exists)

```plaintext
internal/identity/
├── apperr/           # Application errors
├── authz/            # Authorization server (OAuth2.1)
├── bootstrap/        # Bootstrap/initialization
├── config/           # Configuration
├── demo/             # Demo setup
├── domain/           # Domain models
├── healthcheck/      # Health endpoints
├── idp/              # Identity provider
├── integration/      # Integration tests
├── issuer/           # Token issuer
├── jobs/             # Background jobs
├── jwks/             # JWKS management
├── magic/            # Magic constants
├── notifications/    # Notifications
├── process/          # Process management
├── repository/       # Data access
├── rotation/         # Key rotation
├── rs/               # Resource server
├── security/         # Security utilities
├── server/           # Server management
├── storage/          # Storage abstractions
└── test/             # Test utilities
```

### Known Issues (From 6 Passthrus)

- Mix of working and broken code
- Some compilation errors
- Runtime errors in flows
- Missing implementations
- Incomplete test coverage

---

## Demo Goals

### Minimum Viable Demo

```plaintext
User Action                              Expected Result
-----------                              ---------------
Start identity server                    → Server running on https://localhost:9000
GET /.well-known/openid-configuration    → OIDC discovery document
GET /oauth2/authorize?...                → Login page or redirect
POST /oauth2/token (client_credentials)  → Access token returned
POST /oauth2/token (authorization_code)  → Access + refresh tokens
POST /oauth2/introspect                  → Token info returned
POST /oauth2/revoke                      → Token revoked successfully
```

### OAuth2.1 Compliance Demo

```plaintext
Grant Types:
- [ ] authorization_code (with PKCE)
- [ ] client_credentials
- [ ] refresh_token

Endpoints:
- [ ] /oauth2/authorize
- [ ] /oauth2/token
- [ ] /oauth2/introspect
- [ ] /oauth2/revoke
- [ ] /.well-known/openid-configuration
- [ ] /oauth2/jwks

Security:
- [ ] PKCE required for public clients
- [ ] Token binding
- [ ] Refresh token rotation
```

---

## Implementation Tasks

### Phase 2: Assessment (Day 2-3)

#### T2.1: Code Audit

**Steps:**

1. Run `go build ./internal/identity/...`
2. Document all compilation errors
3. Run `go test ./internal/identity/...`
4. Document all test failures
5. Prioritize fixes by dependency order

**Deliverable:** Prioritized fix list

#### T2.2: Database Setup Verification

**Steps:**

1. Check `repository/database.go` implementation
2. Verify SQLite in-memory configuration
3. Verify PostgreSQL configuration
4. Test migrations run successfully
5. Test basic CRUD operations

**Success Criteria:**

- [ ] Database connects without errors
- [ ] Migrations run successfully
- [ ] Can create/read/update/delete records

#### T2.3: Domain Model Verification

**Models to verify:**

| Model | File | Status |
|-------|------|--------|
| User | `domain/user.go` | TBD |
| Client | `domain/client.go` | TBD |
| Session | `domain/session.go` | TBD |
| Token | `domain/token.go` | TBD |
| AuthorizationRequest | `domain/authorization_request.go` | TBD |
| ConsentDecision | `domain/consent_decision.go` | TBD |
| Key | `domain/key.go` | TBD |

**For each model:**

- Verify struct fields match database schema
- Verify GORM annotations correct
- Verify relationships work
- Test create/read operations

#### T2.4: Repository Layer Verification

**Repositories to verify:**

| Repository | Location | Status |
|------------|----------|--------|
| UserRepository | `repository/orm/` | TBD |
| ClientRepository | `repository/orm/` | TBD |
| SessionRepository | `repository/orm/` | TBD |
| TokenRepository | `repository/orm/` | TBD |

**For each repository:**

- Verify interface implementation
- Test CRUD operations
- Test transactions work
- Test error handling

---

### Phase 3: Core Flows (Day 3-5)

#### T3.1: Authorization Endpoint

**File:** `authz/handlers_authorize.go`

**Steps:**

1. Verify endpoint routing works
2. Test authorization code flow
3. Test PKCE challenge/verifier
4. Test redirect handling
5. Test error responses

**Success Criteria:**

- [ ] GET /oauth2/authorize returns login page
- [ ] Authorization code generated on success
- [ ] Redirect to client with code
- [ ] PKCE validation works

#### T3.2: Token Endpoint

**File:** `authz/handlers_token.go`

**Steps:**

1. Verify endpoint routing works
2. Test client_credentials grant
3. Test authorization_code grant
4. Test refresh_token grant
5. Verify JWT format correct

**Success Criteria:**

- [ ] POST /oauth2/token works
- [ ] client_credentials returns access token
- [ ] authorization_code exchanges for tokens
- [ ] refresh_token returns new tokens
- [ ] Tokens are valid JWTs

#### T3.3: Token Introspection

**File:** `authz/handlers_introspect_revoke.go`

**Steps:**

1. Verify endpoint routing works
2. Test active token introspection
3. Test expired token introspection
4. Test invalid token introspection

**Success Criteria:**

- [ ] POST /oauth2/introspect works
- [ ] Active tokens return `active: true`
- [ ] Expired tokens return `active: false`
- [ ] Invalid tokens return `active: false`

#### T3.4: Token Revocation

**File:** `authz/handlers_introspect_revoke.go`

**Steps:**

1. Verify endpoint routing works
2. Test access token revocation
3. Test refresh token revocation
4. Verify revoked tokens fail introspection

**Success Criteria:**

- [ ] POST /oauth2/revoke works
- [ ] Access tokens can be revoked
- [ ] Refresh tokens can be revoked
- [ ] Revoked tokens show `active: false`

#### T3.5: Discovery Endpoint

**File:** `authz/handlers_discovery.go`

**Steps:**

1. Verify endpoint works
2. Verify all required fields present
3. Verify JWKS endpoint accessible

**Success Criteria:**

- [ ] GET /.well-known/openid-configuration works
- [ ] Returns valid JSON
- [ ] Contains issuer, token_endpoint, etc.
- [ ] JWKS URI accessible

---

### Phase 4: Demo Polish (Day 5-6)

#### T4.1: Server Startup

**Steps:**

1. Create simple startup command
2. Verify health endpoint works
3. Verify graceful shutdown
4. Test with docker compose

**Success Criteria:**

- [ ] Single command starts server
- [ ] Health endpoint returns healthy
- [ ] Swagger/OpenAPI UI works
- [ ] Docker compose starts correctly

#### T4.2: Demo Data Seeding

**Steps:**

1. Create demo users (admin, user, service)
2. Create demo clients (web app, service)
3. Create demo scopes
4. Seed on server startup

**Demo Data:**

```plaintext
Users:
- admin@demo.local (password: admin123)
- user@demo.local (password: user123)
- service@demo.local (for service accounts)

Clients:
- demo-web-app (public client, authorization_code)
- demo-service (confidential client, client_credentials)

Scopes:
- openid, profile, email (OIDC standard)
- read:keys, write:keys (KMS integration)
```

#### T4.3: Demo Documentation

**Deliverable:** Step-by-step walkthrough

```plaintext
1. Start identity server
2. Access discovery endpoint
3. Register a client (if not pre-seeded)
4. Get client_credentials token
5. Introspect the token
6. Start authorization_code flow
7. Complete login
8. Exchange code for tokens
9. Refresh the tokens
10. Revoke the tokens
```

---

## Demo Script

### Quick Demo (3 minutes)

```bash
# 1. Start Identity Server
cd cmd/identity
go run . start --config configs/identity/identity-sqlite.yml

# 2. Test discovery (in another terminal)
curl -k https://localhost:9000/.well-known/openid-configuration | jq .

# 3. Get client_credentials token
curl -k -X POST https://localhost:9000/oauth2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=demo-service" \
  -d "client_secret=demo-secret" \
  -d "scope=read:keys" | jq .

# 4. Introspect the token
TOKEN=<token from step 3>
curl -k -X POST https://localhost:9000/oauth2/introspect \
  -d "token=$TOKEN" \
  -d "client_id=demo-service" \
  -d "client_secret=demo-secret" | jq .

# 5. Revoke the token
curl -k -X POST https://localhost:9000/oauth2/revoke \
  -d "token=$TOKEN" \
  -d "client_id=demo-service" \
  -d "client_secret=demo-secret"
```

### Full OAuth Flow Demo (5 minutes)

Includes authorization_code flow with PKCE:

1. Generate PKCE code_verifier and code_challenge
2. Navigate to authorization endpoint
3. Login as demo user
4. Consent to scopes
5. Receive authorization code
6. Exchange code for tokens
7. Use refresh token
8. Introspect and revoke

---

## Files to Fix (Priority Order)

### Critical Path

1. `repository/database.go` - Database connection
2. `repository/migrations/` - Schema setup
3. `domain/*.go` - Domain models
4. `repository/orm/*.go` - Data access
5. `authz/service.go` - Authorization service
6. `authz/handlers_*.go` - HTTP handlers
7. `authz/routes.go` - Route registration

### Secondary

- `config/config.go` - Configuration
- `healthcheck/` - Health endpoints
- `jwks/` - JWKS management
- `issuer/` - Token generation

### Can Wait

- `idp/` - Full IdP features
- `rs/` - Resource server
- `notifications/` - Email/SMS
- `rotation/` - Key rotation

---

## Risk Assessment

### High Risk

- **Authorization flows broken**: Core OAuth2.1 flows may have subtle bugs
- **Token validation broken**: JWT generation/validation issues
- **Database issues**: ORM relationships may be incorrect

### Medium Risk

- **PKCE not working**: PKCE implementation may be incomplete
- **Session management**: User sessions may not persist correctly
- **Refresh tokens**: Rotation may not work properly

### Mitigation

- Test each component in isolation first
- Use test fixtures for predictable inputs
- Compare against OAuth2.1 spec for correctness
- May need to rewrite some components

---

## Verification Checklist

Before marking Identity Demo complete:

- [ ] Server starts without errors
- [ ] Discovery endpoint returns valid config
- [ ] JWKS endpoint returns valid keys
- [ ] client_credentials grant works
- [ ] authorization_code grant works
- [ ] Token introspection works
- [ ] Token revocation works
- [ ] Demo accounts seeded
- [ ] Demo script runs successfully
- [ ] Documentation accurate

---

**Status**: NOT STARTED
**Depends On**: T1.x (KMS verification - optional)
**Blocks**: T5.x (Integration Demo)
