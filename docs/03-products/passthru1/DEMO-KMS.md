# DEMO-KMS: KMS-Only Working Demo

**Purpose**: Refactored KMS demo without breaking manual implementation
**Priority**: HIGHEST - This is working code, protect it
**Timeline**: Day 1-2

---

## Current State

The KMS server was manually implemented and is the most stable component:

- Server starts and serves HTTPS
- Swagger UI works
- Browser API has CORS/CSRF support
- Service API works for machine-to-machine
- 3-tier key hierarchy (root → intermediate → content)
- Key pools with configurable algorithms
- Encrypt/decrypt operations
- Sign/verify operations

---

## Demo Goals

### Minimum Viable Demo

```plaintext
User Action                          Expected Result
-----------                          ---------------
docker compose up -d                 → KMS starts on https://localhost:8080
Open https://localhost:8080/ui/swagger → Swagger UI loads
Try /livez endpoint                  → Returns healthy
Try /readyz endpoint                 → Returns ready
Create key pool (AES-256)            → Pool created successfully
Create key in pool                   → Key created with ID
Encrypt "hello world"                → Ciphertext returned
Decrypt ciphertext                   → "hello world" returned
```

### Enhanced Demo (Time Permitting)

```plaintext
Pre-seeded demo data:
- Root key pool (RSA-4096)
- Intermediate key pool (EC-P256)
- Content key pool (AES-256-GCM)
- Sample keys in each pool

Demo walkthrough:
- Show key hierarchy visualization
- Demonstrate key rotation
- Show audit logs
- Demonstrate access control
```

---

## Implementation Tasks

### T1.1: Verify KMS Server Starts

**Steps:**

1. Run `docker compose -f deployments/compose/compose.yml up -d`
2. Check container health: `docker compose ps`
3. Test health endpoint: `curl -k https://localhost:8080/livez`
4. Test readiness: `curl -k https://localhost:8080/readyz`

**Success Criteria:**

- [ ] Container shows as healthy
- [ ] `/livez` returns 200
- [ ] `/readyz` returns 200

### T1.2: Verify Swagger UI

**Steps:**

1. Open `https://localhost:8080/ui/swagger` in browser
2. Verify all endpoints listed
3. Verify "Try it out" works
4. Verify CSRF token handling

**Success Criteria:**

- [ ] Swagger UI loads without errors
- [ ] All API endpoints visible
- [ ] Can execute requests from UI

### T1.3: Verify Key Operations

**Steps:**

1. Create key pool via Swagger UI
2. Create key in pool
3. Use key to encrypt data
4. Use key to decrypt data
5. Create signing key
6. Sign data and verify

**Success Criteria:**

- [ ] Key pool creation works
- [ ] Key creation works
- [ ] Encryption returns ciphertext
- [ ] Decryption returns plaintext
- [ ] Signing returns signature
- [ ] Verification confirms signature

### T1.4: Document Current API

**Steps:**

1. Export OpenAPI spec
2. Document key pool algorithms supported
3. Document key operations available
4. Document error responses
5. Create example requests/responses

**Deliverable:** Updated API documentation in README or dedicated doc

---

## Demo Script

### Quick Demo (2 minutes)

```bash
# 1. Start KMS (30 seconds)
docker compose -f deployments/compose/compose.yml up -d
echo "Waiting for KMS to start..."
sleep 10

# 2. Verify health (10 seconds)
curl -k https://localhost:8080/livez
curl -k https://localhost:8080/readyz

# 3. Open Swagger UI
echo "Open https://localhost:8080/ui/swagger in browser"

# 4. In Swagger UI:
#    - Create key pool (algorithm: AES-256-GCM)
#    - Create key in pool
#    - Encrypt "Hello, Demo!"
#    - Decrypt the result

# 5. Cleanup
docker compose -f deployments/compose/compose.yml down -v
```

### Detailed Demo (5 minutes)

Includes all of quick demo plus:

- Show 3-tier key hierarchy concept
- Demonstrate key versioning
- Show multiple algorithm support
- Demonstrate browser vs service API difference

---

## Risk Mitigation

### Don't Break What Works

- **Rule 1**: No refactoring of working code
- **Rule 2**: Test after every change
- **Rule 3**: Commit frequently with descriptive messages
- **Rule 4**: Keep a rollback plan (git reset)

### If Something Breaks

1. Stop immediately
2. Check git diff for recent changes
3. Revert to last working commit
4. Document what caused the break
5. Plan a safer approach

---

## Verification Checklist

Before marking KMS Demo complete:

- [ ] `docker compose up -d` starts successfully
- [ ] Health endpoints return 200
- [ ] Swagger UI loads and is interactive
- [ ] Can create key pools
- [ ] Can create keys
- [ ] Can encrypt/decrypt
- [ ] Can sign/verify
- [ ] Demo script runs without errors
- [ ] Documentation is accurate

---

## Files Involved

### Core Server

- `internal/server/` - Server implementation
- `cmd/cryptoutil/` - CLI entry point
- `configs/` - Configuration files

### API

- `api/openapi_spec_*.yaml` - OpenAPI specification
- `api/server/` - Generated server code
- `api/client/` - Generated client code

### Deployment

- `deployments/compose/compose.yml` - Docker Compose
- `deployments/Dockerfile` - Container build

---

**Status**: NOT STARTED
**Blocks**: Identity Demo (T2.x), Integration Demo (T5.x)
