# DEMO-INTEGRATION: KMS + Identity Working Demo

**Purpose**: Combine KMS and Identity into unified authentication demo
**Priority**: ULTIMATE GOAL
**Timeline**: Day 6-7 (only after KMS and Identity demos work)

---

## Prerequisites

**DO NOT START THIS UNTIL:**

- [ ] KMS Demo (DEMO-KMS.md) verified working
- [ ] Identity Demo (DEMO-IDENTITY.md) verified working
- [ ] Both can run independently without errors

---

## Integration Goals

### Primary Goal: KMS Protected by Identity

```plaintext
Flow:
1. Client authenticates with Identity → gets access token
2. Client calls KMS API with access token
3. KMS validates token with Identity
4. KMS authorizes based on scopes
5. KMS performs operation
6. Returns result to client
```

### Secondary Goal: Embedded Identity Option

```plaintext
Option A: Standalone (Production)
┌─────────────┐     ┌─────────────┐
│   Client    │────►│  Identity   │
└─────────────┘     └──────┬──────┘
       │                   │
       │ (with token)      │ (validate)
       ▼                   ▼
┌─────────────────────────────────┐
│              KMS                │
└─────────────────────────────────┘

Option B: Embedded (Development/Simple)
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ (auth + ops)
       ▼
┌─────────────────────────────────┐
│     KMS (with embedded Identity)│
│  ┌─────────────────────────┐    │
│  │   Identity (in-process)  │    │
│  └─────────────────────────┘    │
└─────────────────────────────────┘
```

---

## Integration Architecture

### Scopes for KMS Operations

| Scope | Description | Operations |
|-------|-------------|------------|
| `kms:admin` | Full KMS administration | All operations |
| `kms:read` | Read key metadata | List pools, list keys, get key info |
| `kms:write` | Create/modify keys | Create pool, create key, rotate |
| `kms:encrypt` | Encrypt data | Encrypt operation |
| `kms:decrypt` | Decrypt data | Decrypt operation |
| `kms:sign` | Sign data | Sign operation |
| `kms:verify` | Verify signatures | Verify operation |

### Demo Clients

| Client | Type | Grant Types | Scopes |
|--------|------|-------------|--------|
| `kms-admin-app` | confidential | client_credentials | kms:admin |
| `kms-web-app` | public | authorization_code | kms:read, kms:encrypt, kms:decrypt |
| `kms-service` | confidential | client_credentials | kms:encrypt, kms:decrypt, kms:sign, kms:verify |

### Demo Users

| User | Role | Accessible Scopes |
|------|------|-------------------|
| admin@kms.local | Admin | kms:admin, kms:* |
| user@kms.local | User | kms:read, kms:encrypt, kms:decrypt |
| auditor@kms.local | Auditor | kms:read |

---

## Implementation Tasks

### T5.1: KMS Authentication Setup

**Goal:** Add token validation middleware to KMS

**Steps:**

1. Create token validation middleware
2. Configure Identity token endpoint
3. Add middleware to protected routes
4. Test authenticated requests
5. Test rejection of invalid tokens

**Implementation:**

```go
// internal/server/middleware/auth.go
func TokenValidation(identityURL string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c)
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }

        // Validate with Identity introspection endpoint
        active, claims, err := introspectToken(identityURL, token)
        if err != nil || !active {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }

        c.Set("claims", claims)
        c.Next()
    }
}
```

**Success Criteria:**

- [ ] Requests without token rejected (401)
- [ ] Requests with invalid token rejected (401)
- [ ] Requests with expired token rejected (401)
- [ ] Requests with valid token proceed

### T5.2: Scope-Based Authorization

**Goal:** Implement scope checking for KMS operations

**Steps:**

1. Define scope requirements per endpoint
2. Create scope validation middleware
3. Extract scopes from token claims
4. Enforce scope requirements
5. Test authorization failures

**Implementation:**

```go
// internal/server/middleware/authz.go
func RequireScopes(requiredScopes ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := c.MustGet("claims").(TokenClaims)
        tokenScopes := strings.Split(claims.Scope, " ")

        for _, required := range requiredScopes {
            if !contains(tokenScopes, required) {
                c.AbortWithStatusJSON(403, gin.H{"error": "insufficient scope"})
                return
            }
        }
        c.Next()
    }
}

// Route registration
router.POST("/api/v1/pools",
    TokenValidation(identityURL),
    RequireScopes("kms:write"),
    handlers.CreatePool)
```

**Success Criteria:**

- [ ] Operations require correct scopes
- [ ] Missing scopes return 403
- [ ] Admin scope grants all access
- [ ] Scope inheritance works (kms:admin → all)

### T5.3: Embedded Identity Option

**Goal:** Allow Identity to run in-process with KMS

**Steps:**

1. Create embeddable Identity package
2. Add `--embedded-identity` flag to KMS
3. Initialize Identity in KMS process
4. Route internal auth calls in-process
5. Test embedded mode works

**Implementation:**

```go
// internal/identity/embedded/embedded.go
type EmbeddedIdentity struct {
    config *config.Config
    server *authz.Server
}

func New(cfg *config.Config) (*EmbeddedIdentity, error) {
    // Initialize Identity server in-process
    server, err := authz.NewServer(cfg)
    if err != nil {
        return nil, err
    }
    return &EmbeddedIdentity{config: cfg, server: server}, nil
}

func (e *EmbeddedIdentity) ValidateToken(token string) (*Claims, error) {
    // Direct in-process validation (no HTTP call)
    return e.server.IntrospectToken(token)
}
```

**KMS Configuration:**

```yaml
# configs/kms/kms-embedded.yml
identity:
  mode: embedded  # or "external"
  # If external:
  url: https://identity.local:9000
  # If embedded:
  embedded:
    issuer: https://kms.local:8080
    signing_key: <auto-generated or configured>
```

**Success Criteria:**

- [ ] KMS starts with embedded Identity
- [ ] Token generation works in-process
- [ ] Token validation works in-process
- [ ] No external Identity server needed

### T5.4: Integration Demo Polish

**Goal:** Single command demo experience

**Docker Compose:**

```yaml
# deployments/compose/compose-integration.yml
services:
  identity:
    image: cryptoutil/identity:latest
    ports:
      - "9000:9000"
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "-", "https://127.0.0.1:9000/healthz"]

  kms:
    image: cryptoutil/kms:latest
    ports:
      - "8080:8080"
    environment:
      - IDENTITY_URL=https://identity:9000
    depends_on:
      identity:
        condition: service_healthy
```

**Success Criteria:**

- [ ] `docker compose up -d` starts both services
- [ ] Identity health check passes
- [ ] KMS waits for Identity to be ready
- [ ] End-to-end auth flow works
- [ ] Demo script completes successfully

---

## Demo Scripts

### Integration Demo (5 minutes)

```bash
# 1. Start integrated services
docker compose -f deployments/compose/compose-integration.yml up -d

# 2. Wait for services
echo "Waiting for services to start..."
sleep 15

# 3. Get service token
TOKEN=$(curl -sk -X POST https://localhost:9000/oauth2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=kms-service" \
  -d "client_secret=secret" \
  -d "scope=kms:encrypt kms:decrypt" | jq -r '.access_token')

echo "Got token: ${TOKEN:0:50}..."

# 4. Create key pool (requires kms:write - should fail)
echo "Attempting to create pool without kms:write scope..."
curl -sk -X POST https://localhost:8080/api/v1/pools \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "demo-pool", "algorithm": "AES-256-GCM"}'
# Expected: 403 Forbidden

# 5. Get admin token
ADMIN_TOKEN=$(curl -sk -X POST https://localhost:9000/oauth2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=kms-admin-app" \
  -d "client_secret=admin-secret" \
  -d "scope=kms:admin" | jq -r '.access_token')

# 6. Create key pool with admin token
echo "Creating pool with admin token..."
curl -sk -X POST https://localhost:8080/api/v1/pools \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "demo-pool", "algorithm": "AES-256-GCM"}'
# Expected: 201 Created

# 7. Encrypt with service token
echo "Encrypting data with service token..."
curl -sk -X POST https://localhost:8080/api/v1/encrypt \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"pool": "demo-pool", "plaintext": "SGVsbG8gV29ybGQh"}'
# Expected: 200 OK with ciphertext

# 8. Cleanup
docker compose -f deployments/compose/compose-integration.yml down -v
```

### Embedded Mode Demo (3 minutes)

```bash
# 1. Start KMS with embedded Identity
./cryptoutil server start --config configs/kms/kms-embedded.yml

# 2. Get token from embedded Identity endpoint
TOKEN=$(curl -sk -X POST https://localhost:8080/oauth2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=demo" \
  -d "client_secret=secret" \
  -d "scope=kms:encrypt" | jq -r '.access_token')

# 3. Use KMS with token
curl -sk -X POST https://localhost:8080/api/v1/encrypt \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"plaintext": "SGVsbG8gV29ybGQh"}'

# 4. Notice: Single service, single port, integrated auth
```

---

## Architecture Decisions

### ADR-1: Token Validation Strategy

**Decision:** Use token introspection for validation

**Rationale:**

- Works with both embedded and external Identity
- Supports token revocation
- Provides consistent security model
- Can cache results for performance

### ADR-2: Scope Granularity

**Decision:** Fine-grained scopes (per-operation)

**Rationale:**

- Enables least-privilege access
- Supports different client types
- Allows audit trail per operation
- Flexible for future expansion

### ADR-3: Embedded vs External Default

**Decision:** External by default, embedded opt-in

**Rationale:**

- Production deployments should use external
- Embedded simplifies development/testing
- Clear separation of concerns
- Easier to secure external Identity

---

## Verification Checklist

Before marking Integration Demo complete:

- [ ] Both services start with docker compose
- [ ] Identity issues tokens correctly
- [ ] KMS validates tokens correctly
- [ ] Scope enforcement works
- [ ] Unauthorized requests rejected
- [ ] Embedded mode works
- [ ] Demo scripts run successfully
- [ ] Documentation complete

---

## Files to Create/Modify

### KMS Changes

- `internal/server/middleware/auth.go` - Token validation
- `internal/server/middleware/authz.go` - Scope checking
- `internal/server/config/config.go` - Identity configuration
- `internal/server/routes.go` - Middleware registration

### Identity Changes

- `internal/identity/embedded/embedded.go` - Embeddable package
- `internal/identity/embedded/client.go` - In-process client

### Shared

- `deployments/compose/compose-integration.yml` - Docker Compose
- `configs/kms/kms-integrated.yml` - KMS with external Identity
- `configs/kms/kms-embedded.yml` - KMS with embedded Identity

---

**Status**: NOT STARTED
**Depends On**: DEMO-KMS.md, DEMO-IDENTITY.md (both must be complete)
**Blocks**: Nothing (this is the ultimate goal)
