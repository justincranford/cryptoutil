# V8 Analysis Thorough - Detailed Technical Analysis

**Created**: 2026-02-03
**Purpose**: Deep technical analysis for V8 implementation
**Overview**: See [analysis-overview.md](analysis-overview.md) for executive summary

---

## 1. Executive Summary

### 1.1 V7 Claim vs Reality

| V7 Claim | Evidence | Reality |
|----------|----------|---------|
| "Barrier integration addressed in Phase 5" | Tasks 5.3, 5.4 in tasks.md | Both marked "❌ Not Started" |
| "KMS uses template barrier" | `grep -r "shared/barrier" internal/kms/` | 4 files still import shared/barrier |
| "Phase 5 complete" | TODOs in server.go | 3 explicit TODOs about incomplete migration |
| "Adapter created" | orm_barrier_adapter.go exists | True, but UNUSED - KMS still uses shared/barrier directly |

### 1.2 Code Archaeology Evidence

```bash
# Verification commands run 2026-02-03

$ grep -r "shared/barrier" internal/kms/ --include="*.go"
internal/kms/server/businesslogic/businesslogic.go:import cryptoutilBarrier "cryptoutil/internal/shared/barrier"
internal/kms/server/application/application_basic.go:import cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
internal/kms/server/application/application_core.go:import cryptoutilBarrier "cryptoutil/internal/shared/barrier"
internal/kms/server/server.go:// TODO: This imports shared/barrier but should use template barrier

$ grep "TODO" internal/kms/server/server.go | head -5
// TODO(Phase2-5): KMS needs to be migrated to use template's GORM database and barrier.
// TODO(Phase2-5): Replace with template's GORM database and barrier.
// TODO(Phase2-5): Switch to TemplateWithDomain mode once KMS uses template DB.

$ grep "NewServerBuilder" internal/kms/server/server.go
builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, settings)
```

### 1.3 Key Insight

KMS DOES use ServerBuilder (contrary to earlier belief), but the barrier migration within ServerBuilder is incomplete. The builder is configured but KMS still directly imports and uses `shared/barrier` for actual operations.

KMS us

---

## 2. Service Architecture Comparison

### 2.1 KMS (sm-kms)

**Location**: `internal/kms/`

**Structure**:
```
internal/kms/
├── domain/           # Domain models
├── repository/       # Data access (raw database/sql, NOT GORM)
├── server/
│   ├── application/  # App lifecycle
│   ├── barrier/      # UNUSED orm_barrier_adapter.go
│   ├── businesslogic/# Business services
│   └── server.go     # Uses ServerBuilder with TODOs
└── service/          # Service layer
```

**Key Characteristics**:
- Uses ServerBuilder (✅ confirmed via grep)
- Still imports shared/barrier (❌ 4 files)
- Uses raw database/sql (❌ not GORM)
- Has orm_barrier_adapter.go (unused)

**Evidence**:
```bash
$ grep "gorm.io/gorm" internal/kms/ -r | wc -l
11  # GORM imports exist BUT...

$ grep "database/sql" internal/kms/server/repository/*.go | wc -l
8   # Primary data access uses raw database/sql
```

### 2.2 Service-Template

**Location**: `internal/apps/template/`

**Structure**:
```
internal/apps/template/
├── service/
│   └── server/
│       ├── barrier/      # 18 files - GORM-based barrier ✅ TARGET
│       ├── builder/      # ServerBuilder pattern
│       ├── apis/         # HTTP handlers
│       └── application/  # App lifecycle
└── testing/
    └── e2e/              # E2E test helpers
```

**Key Characteristics**:
- Reference implementation (✅ all patterns)
- GORM-based barrier (✅ target for KMS)
- ServerBuilder (✅ provides to other services)
- 98.91% mutation efficacy (✅ exceeds 98% ideal)

### 2.3 Cipher-IM

**Location**: `internal/apps/cipher/im/`

**Structure**:
```
internal/apps/cipher/im/
├── domain/       # Domain models with GORM tags
├── repository/   # GORM repositories
├── server/       # Uses ServerBuilder
└── service/      # Business logic
```

**Key Characteristics**:
- First service using template (✅ validated template)
- Uses ServerBuilder (✅ via builder package)
- Uses template barrier (✅ via ServerBuilder)
- GORM throughout (✅ no database/sql)

### 2.4 JOSE-JA

**Location**: `internal/apps/jose/ja/`

**Key Characteristics**:
- Uses ServerBuilder (✅ confirmed)
- Uses template barrier (✅ via ServerBuilder)
- GORM (✅)
- Migration in progress (⏳ multi-tenancy pending)

---

## 3. Testing Strategy Comparison

### 3.1 Coverage Analysis

| Service | Total | Production | Infrastructure | Target | Gap |
|---------|-------|------------|----------------|--------|-----|
| KMS | 75.2% | ~70% | ~80% | 95%/98% | -20%/-18% |
| Template | 82.5% | ~80% | ~85% | 95%/98% | -15%/-13% |
| Cipher-IM | 78.9% | ~75% | ~82% | 95%/98% | -20%/-16% |
| JOSE-JA | 92.5% | ~90% | ~95% | 95%/98% | -5%/-3% |

**JOSE-JA closest to targets** but still below minimum.

### 3.2 Mutation Testing Results

| Service | Efficacy | Target | Status |
|---------|----------|--------|--------|
| Template | 98.91% | 98% ideal | ✅ EXCEEDS |
| JOSE-JA | 97.20% | 95% min | ✅ Above min, below ideal |
| KMS | Not run | 95% min | ❌ No data |
| Cipher-IM | Not run | 95% min | ❌ Docker issues blocking |

### 3.3 Testing Patterns Used

**Unit Tests**:
- All services: Table-driven tests with t.Parallel()
- KMS: Extensive crypto/key operation tests
- Template: Barrier, session, realm service tests
- Cipher-IM: Message CRUD tests
- JOSE-JA: JWK generation/validation tests

**Integration Tests**:
- KMS: PostgreSQL containers via testcontainers-go
- Template: GORM with SQLite in-memory, PostgreSQL containers
- Cipher-IM: Same as Template
- JOSE-JA: Partial (migration in progress)

**E2E Tests**:
- KMS: Docker Compose with PostgreSQL
- Template: Docker Compose with E2E helpers
- Cipher-IM: Docker Compose
- JOSE-JA: Partial (migration in progress)

### 3.4 Test File Organization

```
*_test.go           # Unit tests
*_integration_test.go  # Integration tests
*_bench_test.go     # Benchmarks
*_fuzz_test.go      # Fuzz tests (sparse)
*_property_test.go  # Property tests (sparse)
```

---

## 4. Barrier Implementation Analysis

### 4.1 Template Barrier (TARGET)

**Location**: `internal/apps/template/service/server/barrier/`

**Files (18 total)**:
```
barrier_service.go           # Main service interface
gorm_barrier_repository.go   # GORM storage
rotation_service.go          # Key rotation
unsealkeysservice/           # Unseal key management
  unsealkeysservice.go
  unsealkeysservice_settings.go
  unsealkeysservice_test.go
... (tests, domain models)
```

**Key Interfaces**:
```go
type BarrierService interface {
    Encrypt(ctx context.Context, keyID uuid.UUID, plaintext []byte) ([]byte, error)
    Decrypt(ctx context.Context, keyID uuid.UUID, ciphertext []byte) ([]byte, error)
    GetOrCreateKey(ctx context.Context, keyType KeyType) (*Key, error)
    RotateKey(ctx context.Context, keyID uuid.UUID) (*Key, error)
}
```

**Storage**: GORM with PostgreSQL/SQLite support

### 4.2 Shared Barrier (LEGACY - TO DELETE)

**Location**: `internal/shared/barrier/`

**Files**:
```
barrier.go                   # Legacy implementation
barrier_repository.go        # Raw SQL storage
unsealkeysservice/           # Shared unseal key service
```

**Current Users**: KMS only (4 files importing)

**V8 Decision (Q2=E)**: DELETE IMMEDIATELY after KMS migration completes

### 4.3 KMS Barrier Adapter (UNUSED)

**Location**: `internal/kms/server/barrier/orm_barrier_adapter.go`

**Purpose**: Was intended to bridge KMS to GORM-based barrier
**Status**: EXISTS but UNUSED - KMS still uses shared/barrier directly
**V8 Action**: DELETE as part of cleanup (Phase 4)

### 4.4 Migration Path

```
Current:  KMS → shared/barrier → raw SQL storage
Target:   KMS → ServerBuilder → template barrier → GORM storage
```

**Steps**:
1. Update KMS imports from shared/barrier to template barrier
2. Update KMS initialization to use barrier from ServiceResources
3. Verify all barrier operations work with GORM backend
4. Delete shared/barrier and orm_barrier_adapter.go

---

## 5. KMS Migration Scope

### 5.1 Files Requiring Changes

**File 1**: `internal/kms/server/businesslogic/businesslogic.go`
```go
// CURRENT
import cryptoutilBarrier "cryptoutil/internal/shared/barrier"

// TARGET
import cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
```

**File 2**: `internal/kms/server/application/application_basic.go`
```go
// CURRENT
import cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"

// TARGET
import cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
```

**File 3**: `internal/kms/server/application/application_core.go`
```go
// CURRENT
import cryptoutilBarrier "cryptoutil/internal/shared/barrier"

// TARGET
import cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
```

**File 4**: `internal/kms/server/server.go`
- Update comment referencing shared/barrier
- Remove 3 TODOs after migration complete

### 5.2 Interface Compatibility Check

**Required Analysis** (Phase 2 Task 2.1):
- Compare shared/barrier.BarrierService vs template barrier.BarrierService
- Identify method signature differences
- Create adapter if needed (or update KMS usage)

### 5.3 Estimated LOE

| Task | Hours | Complexity |
|------|-------|------------|
| Interface analysis | 2h | Low |
| Import updates | 1h | Low |
| Initialization changes | 3h | Medium |
| Test updates | 4h | Medium |
| Verification | 2h | Low |
| **Total** | **12h** | **Medium** |

---

## 6. Quality Gates Summary

### 6.1 Per-Phase Quality Gates

**MANDATORY for each phase completion**:

```
✅ All tests pass
   $ go test ./... -shuffle=on

✅ Coverage maintained/improved
   $ go test ./... -coverprofile=coverage.out
   $ go tool cover -func=coverage.out | tail -1
   # Target: ≥95% production, ≥98% infrastructure

✅ Linting clean
   $ golangci-lint run
   # Zero errors required

✅ No new TODOs without tracking
   $ grep -r "TODO" internal/kms/ | wc -l
   # Should decrease as migration progresses
```

### 6.2 Phase 5 Quality Gate (Mutation Testing)

**Grouped at end per Q3=E decision**:

```
✅ Mutation efficacy ≥95% minimum
   $ gremlins unleash ./internal/kms/...
   # Target: ≥95% minimum, 98% ideal
```

### 6.3 Documentation Updates (Q4=E)

**Only update .github/instructions/*.instructions.md if ACTUALLY-WRONG**:
- Found error: Document in test-output/doc-fixes/
- Update with commit message: `docs: fix incorrect claim about X`
- Do NOT update for style/clarity (scope creep prevention)

---

## 7. Risk Assessment

### 7.1 Risk Matrix

| Risk | Prob | Impact | Score | Mitigation |
|------|------|--------|-------|------------|
| Barrier API mismatch | M | H | 6 | Interface comparison in Task 2.1 |
| Test breakage during migration | H | M | 6 | Run tests after each file change |
| Hidden shared/barrier usage | L | H | 3 | grep -r verification before delete |
| GORM migration side effects | M | M | 4 | Test SQLite + PostgreSQL |
| Coverage regression | M | M | 4 | Track coverage per commit |

**Score**: Probability × Impact (L=1, M=2, H=3)

### 7.2 Mitigation Strategies

**API Mismatch**:
- Task 2.1 compares interfaces explicitly
- Create compatibility layer if methods differ
- Test all barrier operations after migration

**Test Breakage**:
- Commit after each successful file migration
- Run `go test ./internal/kms/...` after each change
- Keep shared/barrier until all tests pass

**Hidden Dependencies**:
- Run `grep -r "shared/barrier"` across entire codebase
- Check transitive imports
- Verify no other services secretly depend on shared/barrier

---

## 8. Phase Summary

### Phase 1: Research & Documentation (8h)

| Task | Focus | Deliverable |
|------|-------|-------------|
| 1.1 | Code archaeology verification | Updated comparison-table.md |
| 1.2 | Interface compatibility analysis | Interface diff document |
| 1.3 | Test inventory | Test coverage baseline |
| 1.4 | Risk assessment update | Risk matrix with mitigations |

**Exit Criteria**: Accurate documentation, interface analysis complete

### Phase 2: KMS Barrier Migration (16h)

| Task | Focus | Deliverable |
|------|-------|-------------|
| 2.1 | Interface compatibility | Adapter if needed |
| 2.2 | Import updates (4 files) | Code changes |
| 2.3 | Initialization changes | ServiceResources integration |
| 2.4 | Integration verification | All tests passing |

**Exit Criteria**: KMS uses template barrier, 0 imports from shared/barrier

### Phase 3: Testing & Validation (12h)

| Task | Focus | Deliverable |
|------|-------|-------------|
| 3.1 | Unit test updates | Tests for new barrier usage |
| 3.2 | Integration test updates | PostgreSQL/SQLite tests |
| 3.3 | E2E test updates | Docker Compose tests |
| 3.4 | Coverage improvement | ≥95% coverage target |

**Exit Criteria**: All tests pass, coverage ≥95%

### Phase 4: Cleanup (4h)

| Task | Focus | Deliverable |
|------|-------|-------------|
| 4.1 | Delete shared/barrier | `rm -rf internal/shared/barrier/` |
| 4.2 | Delete orm_barrier_adapter | `rm internal/kms/server/barrier/` |

**Exit Criteria**: Unused code removed, repo clean

### Phase 5: Mutation Testing (8h)

| Task | Focus | Deliverable |
|------|-------|-------------|
| 5.1 | Run mutation tests | gremlins output |
| 5.2 | Address surviving mutants | ≥95% efficacy |

**Exit Criteria**: Mutation efficacy ≥95% minimum

---

## 9. V8 Decisions from Quizme

### Q1: Barrier Location (Answer: E)

**Question**: Where should the barrier implementation live?
**Options**:
- A: shared/barrier (status quo)
- B: Both shared and template
- C: Template with shared wrapper
- D: New dedicated package
- **E: Template only (CHOSEN)**

**Rationale**: Single source of truth, GORM consistency, ServerBuilder integration

### Q2: shared/barrier Fate (Answer: E)

**Question**: When to delete shared/barrier?
**Options**:
- A: Keep indefinitely
- B: Archive in separate branch
- C: Delete after 6-month deprecation
- D: Delete after all services migrate
- **E: Delete IMMEDIATELY after KMS migration (CHOSEN)**

**Rationale**: No other users, clean codebase, prevents accidental usage

### Q3: Testing Scope (Answer: E)

**Question**: Testing approach for V8?
**Options**:
- A: Unit tests only
- B: Unit + integration
- C: All tests deferred to end
- D: Mutations per phase
- **E: Full testing per phase, mutations grouped at end (CHOSEN)**

**Rationale**: Quality + velocity balance, mutations are slow

### Q4: Documentation Updates (Answer: E)

**Question**: When to update .github/instructions/?
**Options**:
- A: Update all docs comprehensively
- B: Update affected docs only
- C: No doc updates
- D: Separate doc phase after implementation
- **E: Only ACTUALLY-WRONG instructions (CHOSEN)**

**Rationale**: Prevents scope creep, focuses on errors

---

## 10. Success Metrics

### 10.1 Completion Checklist

**Code Changes**:
- [ ] KMS imports template barrier (0 imports from shared/barrier)
- [ ] KMS initializes barrier from ServiceResources
- [ ] 3 TODOs removed from server.go
- [ ] shared/barrier directory deleted
- [ ] orm_barrier_adapter.go deleted

**Testing**:
- [ ] All existing tests pass
- [ ] New barrier tests added
- [ ] Coverage ≥95% for migrated code
- [ ] Mutation efficacy ≥95%

**Verification Commands**:
```bash
# Zero shared/barrier imports
$ grep -r "shared/barrier" internal/kms/ --include="*.go" | wc -l
0

# Zero TODOs about barrier migration
$ grep "TODO.*barrier" internal/kms/server/server.go | wc -l
0

# shared/barrier deleted
$ ls internal/shared/barrier/
ls: cannot access 'internal/shared/barrier/': No such file or directory

# Tests pass
$ go test ./internal/kms/... -v
PASS

# Coverage check
$ go test ./internal/kms/... -coverprofile=coverage.out
$ go tool cover -func=coverage.out | grep total
total:    (statements)    95.0%
```

### 10.2 Tracking Table

| Metric | Baseline | Target | Current |
|--------|----------|--------|---------|
| shared/barrier imports | 4 | 0 | - |
| server.go TODOs | 3 | 0 | - |
| Test coverage | 75.2% | ≥95% | - |
| Mutation efficacy | N/A | ≥95% | - |
| shared/barrier exists | Yes | No | - |

---

## Appendix A: File Inventory

### A.1 Files to Modify

| File | Change Type | Priority |
|------|-------------|----------|
| `internal/kms/server/businesslogic/businesslogic.go` | Import update | High |
| `internal/kms/server/application/application_basic.go` | Import update | High |
| `internal/kms/server/application/application_core.go` | Import update | High |
| `internal/kms/server/server.go` | TODO cleanup | Medium |

### A.2 Files to Delete

| File/Directory | Reason | Phase |
|----------------|--------|-------|
| `internal/shared/barrier/` | Legacy, no longer used | 4 |
| `internal/kms/server/barrier/` | Unused adapter | 4 |

### A.3 Files to Create

| File | Purpose | Phase |
|------|---------|-------|
| New barrier tests | Test template barrier with KMS | 3 |
| Coverage report | Document final coverage | 3 |
| Mutation report | Document mutation results | 5 |

---

## 11. HTTPS Ports Review (All 9 Product-Services)

### 11.1 Data Collection Methodology

Port configurations extracted from:
- `deployments/kms/compose.yml`
- `deployments/ca/compose.yml`
- `deployments/ca/compose/compose.yml`
- `deployments/jose/compose.yml`
- `deployments/identity/compose.advanced.yml`
- `deployments/template/compose.yml`
- `deployments/telemetry/compose.yml`
- `.github/instructions/02-01.architecture.instructions.md`
- `.github/instructions/02-03.https-ports.instructions.md`

### 11.2 Detailed Port Analysis by Service

#### sm-kms (Secrets Manager - Key Management Service)
```yaml
# deployments/kms/compose.yml
kms-sqlite:
  ports:
    - "8080:8080"  # HTTPS public (SQLite profile)
    
kms-postgres-1:
  ports:
    - "8081:8080"  # HTTPS public (PostgreSQL instance 1)
    
kms-postgres-2:
  ports:
    - "8082:8080"  # HTTPS public (PostgreSQL instance 2)
```
- Container port: 8080
- Admin port: 9090 (127.0.0.1 only, NOT exposed)
- Health check: `https://127.0.0.1:9090/admin/api/v1/livez`

#### pki-ca (PKI - Certificate Authority)
```yaml
# deployments/ca/compose.yml
ca-sqlite:
  ports:
    - "8443:8443"  # HTTPS public (SQLite)
    
ca-postgres-1:
  ports:
    - "8444:8443"  # HTTPS public (PG1)
    
ca-postgres-2:
  ports:
    - "8445:8443"  # HTTPS public (PG2)
```
- Container port: 8443
- Admin port: NOT EXPOSED (uses different pattern)
- Health check: `https://127.0.0.1:8443/livez` (NON-STANDARD - no /admin/api/v1/ prefix)

#### jose-ja (JOSE - JWK Authority)
```yaml
# deployments/jose/compose.yml
jose-server:
  ports:
    - "8092:8092"  # Public API (HTTPS)
    - "9092:9092"  # Admin API (HTTPS) - EXPOSED!
```
- Container port: 8092 (public), 9092 (admin)
- **ISSUE**: Admin port 9092 is EXPOSED to host (violates security pattern)
- **DISCREPANCY**: Instructions document 9443-9449, implementation uses 8092

#### identity-authz (OAuth 2.1 Authorization Server)
```yaml
# deployments/identity/compose.advanced.yml
identity-authz:
  ports:
    - "8080-8089:8080"  # Port range allows scaling
```
- Container port: 8080
- Admin port: 9090 (NOT exposed per instructions)
- **DISCREPANCY**: Instructions document 18000-18009, implementation uses 8080-8089

#### identity-idp (OIDC Identity Provider)
```yaml
# deployments/identity/compose.advanced.yml
identity-idp:
  ports:
    - "8100-8109:8081"  # Port range for scaling
```
- Container port: 8081
- Admin port: 9090 (NOT exposed)
- **DISCREPANCY**: Instructions document 18100-18109, implementation uses 8100-8109

#### identity-rs (Resource Server)
```yaml
# deployments/identity/compose.advanced.yml
identity-rs:
  ports:
    - "8200-8209:8082"  # Port range for scaling
```
- Container port: 8082
- Admin port: 9090 (NOT exposed)
- **DISCREPANCY**: Instructions document 18200-18209, implementation uses 8200-8209

#### identity-rp (Relying Party)
```yaml
# deployments/identity/compose.advanced.yml (extrapolated)
```
- Container port: 8083 (expected)
- Host port range: 8300-8309 (expected)
- **DISCREPANCY**: Instructions document 18300-18309

#### identity-spa (Single Page Application)
```yaml
# deployments/identity/compose.advanced.yml (extrapolated)
```
- Container port: 8084 (expected)
- Host port range: 8400-8409 (expected)
- **DISCREPANCY**: Instructions document 18400-18409

#### cipher-im (Cipher - Instant Messenger)
```yaml
# deployments/template/compose.yml (cipher-im binary used)
cipher-im-sqlite:
  ports:
    - "8880:8888"  # Public HTTPS API
    
cipher-im-postgres-1:
  ports:
    - "8881:8888"
    
cipher-im-postgres-2:
  ports:
    - "8882:8888"
```
- Container port: 8888
- Admin port: 9090 (NOT exposed)
- Health check: `https://127.0.0.1:9090/admin/api/v1/livez`

### 11.3 Telemetry Infrastructure Ports

```yaml
# deployments/telemetry/compose.yml
opentelemetry-collector-contrib:
  # ports:  (COMMENTED OUT - internal only)
  #   - "4317:4317"   # OTLP gRPC
  #   - "4318:4318"   # OTLP HTTP
```
- Container ports: 4317 (gRPC), 4318 (HTTP)
- Host ports: NOT exposed (container-to-container only)
- Correct pattern: Telemetry ports should not be exposed to host

### 11.4 PostgreSQL Ports

| Compose File | Service | Host Port | Purpose |
|--------------|---------|-----------|---------|
| kms/compose.yml | postgres | 5432 | KMS database |
| ca/compose.yml | postgres | 5432 | CA database |
| identity/compose.advanced.yml | identity-postgres | 5433 | Identity database |
| template/compose.yml | postgres | 5433 | Template database |

- Port 5432: Default PostgreSQL (KMS, CA)
- Port 5433: Offset PostgreSQL (identity, template to avoid conflicts)

### 11.5 Issues Identified

1. **jose-ja Admin Port Exposure**: Admin port 9092 exposed to host (security violation)
2. **Port Range Discrepancies**: Instructions vs implementation mismatch
3. **pki-ca Health Check Path**: Non-standard path without /admin/api/v1/ prefix
4. **Documentation Drift**: Instructions file not updated when ports changed

### 11.6 Recommendations

1. **Short-term**: Document actual ports in ARCHITECTURE.md (done)
2. **Medium-term**: Standardize jose-ja to NOT expose admin port
3. **Long-term**: Update instructions file OR update implementations to match
4. **Follow-up Task**: Add port standardization phase to V8 or V9 plan

## 12. Realm Design and Implementation Analysis

### Critical Distinction: realm_id vs tenant_id

**tenant_id**: Data isolation boundary
- ALL data queries MUST filter by tenant_id
- Scopes: keys, sessions, audit logs, messages, users
- Cross-tenant access is FORBIDDEN

**realm_id**: Authentication policy context
- Determines HOW users authenticate (not WHAT they access)
- Multiple realms can exist within one tenant
- Users from different realms see SAME tenant data

### 16 Realm Types from realm_service.go

```go
// Federated realm types (external identity providers)
RealmTypeUsernamePassword = "username_password"  // Default, database credentials
RealmTypeLDAP             = "ldap"               // LDAP/Active Directory
RealmTypeOAuth2           = "oauth2"             // OAuth 2.0/OIDC provider
RealmTypeSAML             = "saml"               // SAML 2.0 federation

// Non-federated browser realm types (/browser/** paths)
RealmTypeJWESessionCookie       = "jwe-session-cookie"      // Encrypted JWT cookie
RealmTypeJWSSessionCookie       = "jws-session-cookie"      // Signed JWT cookie
RealmTypeOpaqueSessionCookie    = "opaque-session-cookie"   // Server-side session
RealmTypeBasicUsernamePassword  = "basic-username-password" // HTTP Basic auth
RealmTypeBearerAPIToken         = "bearer-api-token"        // Bearer token
RealmTypeHTTPSClientCert        = "https-client-cert"       // mTLS client cert

// Non-federated service realm types (/service/** paths)
RealmTypeJWESessionToken        = "jwe-session-token"       // Encrypted JWT token
RealmTypeJWSSessionToken        = "jws-session-token"       // Signed JWT token
RealmTypeOpaqueSessionToken     = "opaque-session-token"    // Server-side token
RealmTypeBasicClientIDSecret    = "basic-client-id-secret"  // Client credentials
// bearer-api-token and https-client-cert shared with browser types
```

### RealmConfig Structure from realm_config.go

```go
type RealmConfig struct {
    // Password validation rules
    PasswordMinLength        int   // Default: 12
    PasswordRequireUppercase bool  // Default: true
    PasswordRequireLowercase bool  // Default: true
    PasswordRequireDigits    bool  // Default: true
    PasswordRequireSpecial   bool  // Default: true
    PasswordMinUniqueChars   int   // Default: 8
    PasswordMaxRepeatedChars int   // Default: 3

    // Session configuration
    SessionTimeout        int   // Seconds, default: 3600 (1 hour)
    SessionAbsoluteMax    int   // Seconds, default: 86400 (24 hours)
    SessionRefreshEnabled bool  // Default: true

    // Multi-factor authentication
    MFARequired bool     // Default: false
    MFAMethods  []string // e.g., ["totp", "webauthn", "sms"]

    // Rate limiting overrides
    LoginRateLimit   int // Attempts per minute, default: 5
    MessageRateLimit int // Messages per minute, default: 10
}
```

### Factory Functions

- `DefaultRealm()` - Standard security policies
- `EnterpriseRealm()` - Stricter policies (MFA required, shorter sessions)

### Documentation Updates Applied

| File | Change | Status |
|------|--------|--------|
| ARCHITECTURE.md | Expanded realm section: 16 types, config, tenant relationship | ✅ Done |
| SERVICE-TEMPLATE.md | Added Realm Pattern section after ServiceResources | ✅ Done |
| analysis-overview.md | Section 12 with realm summary | ✅ Done |
| analysis-thorough.md | This detailed analysis | ✅ Done |

### LLM Training vs Implementation

**Common LLM Mistake**: Treating realms as data isolation boundaries (like AWS Organizations)

**Correct Implementation**: Realms are authentication METHOD selectors only:
- Same tenant + different realms = SAME data access
- Different tenants = ISOLATED data (regardless of realm)

### Remaining Work

Phase needed to verify realm implementation in all services matches this design.
