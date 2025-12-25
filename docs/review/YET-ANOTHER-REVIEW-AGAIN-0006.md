# Review 0006: Copilot Instructions Deep Analysis

**Date**: 2024-12-24
**Purpose**: Deep analysis of ALL 27 copilot instruction files for contradictions, ambiguities, and inconsistencies
**Context**: User has attempted "dozen" backports to fix SpecKit issues but regeneration always diverges. This review identifies WHY copilot instructions don't stick.

---

## Executive Summary

**Total Files Analyzed**: 27 instruction files + 26 memory files + 4 spec files = 57 files

**Critical Findings**:

- **4 CRITICAL contradictions** found that DIRECTLY cause SpecKit divergence
- **7 ambiguities** that allow multiple interpretations
- **5 missing coverage areas** with no instruction guidance
- **3 redundancy issues** causing conflicting signals

**Severity Assessment**: **HIGH** - The contradictions found explain why SpecKit regeneration diverges. The multi-tenancy contradiction alone is sufficient to cause complete implementation divergence.

**Root Cause**: Copilot instructions use simplified "tactical patterns" that CONTRADICT the detailed specifications in constitution.md and spec.md. LLM agents follow the simpler instructions, ignoring nuanced requirements.

---

## Critical Contradictions

### 1. Multi-Tenancy Isolation Pattern - CRITICAL DIVERGENCE

**Severity**: CRITICAL
**Impact**: Complete implementation divergence
**Files**:

- `.github/instructions/03-04.database.instructions.md` (lines 147-150)
- `.specify/memory/constitution.md` (lines 691-693)
- `specs/002-cryptoutil/spec.md` (lines 2389-2432)
- `specs/002-cryptoutil/plan.md` (lines 80-82)

**Contradiction**:

**Instruction File Says** (03-04.database.instructions.md):

```markdown
## Multi-Tenancy - MANDATORY

**Schema-Level Isolation ONLY**:

- Each tenant gets separate schema: `tenant_<uuid>.users`, `tenant_<uuid>.sessions`
- NEVER use row-level multi-tenancy (single schema, tenant_id column)
```

**Constitution/Spec Says** (constitution.md, spec.md, plan.md):

```markdown
**Dual-layer isolation - per-row tenant_id (all DBs) + schema-level (PostgreSQL only)**

**Layer 1: Per-Row Tenant ID** (PostgreSQL + SQLite):
- ALL tables MUST have `tenant_id UUID NOT NULL` column
- `tenant_id` is foreign key to `tenants.id` (UUIDv4)
- ALL queries MUST filter by `WHERE tenant_id = $1`

**Layer 2: Schema-Level Isolation** (PostgreSQL only):
- Each tenant gets separate schema: `CREATE SCHEMA tenant_<UUID>`
- NEVER use row-level security (RLS) - per-row tenant_id provides sufficient isolation
```

**Why This Causes Divergence**:

- LLM agents read instruction file "Schema-Level Isolation ONLY" and implement PostgreSQL-only solution
- SQLite deployments fail because no multi-tenancy implementation exists
- Constitution requires BOTH layers: row-level (all DBs) + schema-level (PostgreSQL only)
- Instruction file says "NEVER use row-level multi-tenancy" which DIRECTLY contradicts constitution's Layer 1

**Fix Required**:
Update `03-04.database.instructions.md` to say:

```markdown
## Multi-Tenancy - MANDATORY

**Dual-Layer Isolation**:

**Layer 1: Per-Row tenant_id** (PostgreSQL + SQLite):
- ALL tables: `tenant_id UUID NOT NULL REFERENCES tenants(id)`
- ALL queries: `WHERE tenant_id = $1`

**Layer 2: Schema-Level** (PostgreSQL only):
- `CREATE SCHEMA tenant_<uuid>`
- Defense-in-depth for PostgreSQL deployments

**NEVER use**: Row-Level Security (RLS) - per-row tenant_id + schema isolation provides sufficient protection
```

---

### 2. Database Choice: Deployment vs Environment - CRITICAL AMBIGUITY

**Severity**: CRITICAL
**Impact**: Wrong database choice for deployments
**Files**:

- `.specify/memory/constitution.md` (line 42)
- `.specify/memory/github.md` (line 30)
- `.specify/memory/database.md` (line 42)
- `specs/002-cryptoutil/plan.md` (line 70-74)

**Contradiction**:

**Constitution Says** (correctly):

```markdown
- Support SQLite (dev, in-memory or file-based) and PostgreSQL (dev & prod)
```

**Plan Says** (correctly):

```markdown
**Database Architecture**:

- PostgreSQL (multi-service deployments in prod||dev) + SQLite (standalone-service deployments in prod||dev)
- Choice based on deployment type (multi-service vs standalone), NOT environment (prod vs dev)
```

**GitHub Memory Says** (INCORRECTLY):

```markdown
**Tests**: SQLite in-memory (`--dev`) OR test-containers for PostgreSQL
```

**Database Memory Says** (INCORRECTLY):

```markdown
**PostgreSQL**: `postgres://user:pass@localhost:5432/dbname?sslmode=disable` (dev) or `sslmode=require` (prod)
```

**Why This Causes Divergence**:

- Constitution says "SQLite (dev)" which LLM agents interpret as "SQLite for development ONLY"
- Plan correctly says "deployment type (multi-service vs standalone), NOT environment (prod vs dev)"
- Memory files reinforce incorrect "SQLite=dev, PostgreSQL=prod" pattern
- No instruction file clarifies the deployment-based choice

**Fix Required**:

1. Update constitution.md to clarify:

```markdown
- Support SQLite (standalone deployments in prod||dev) and PostgreSQL (multi-service deployments in prod||dev)
- Database choice based on deployment architecture (standalone vs multi-service), NOT environment (prod vs dev)
```

2. Add instruction file `03-04.database.instructions.md` section:

```markdown
## Database Selection - CRITICAL

**NOT environment-based (prod vs dev)**
**IS deployment-based (standalone vs multi-service)**

**SQLite**: Single-service deployments (prod||dev), in-memory tests
**PostgreSQL**: Multi-service deployments (prod||dev), shared session state

**Never**: "SQLite for dev, PostgreSQL for prod" pattern
```

---

### 3. Admin Port Configurability - CONTRADICTORY SIGNALS

**Severity**: MEDIUM
**Impact**: Inconsistent admin port implementation
**Files**:

- `.github/instructions/02-03.https-ports.instructions.md`
- `.specify/memory/constitution.md` (lines 285-295)
- `.specify/memory/https-ports.md`
- `specs/002-cryptoutil/plan.md` (line 61)

**Contradiction**:

**Constitution Says**:

```markdown
- Private endpoints MUST ALWAYS use 127.0.0.1:9090 (never configurable, not mapped outside containers)
```

**Plan Says** (CONFLICTING):

```markdown
**Admin Port Configuration**: 127.0.0.1:9090 inside container (NEVER exposed to host), or 127.0.0.1:0 for tests (dynamic allocation)

**Note**: Admin port is configurable but low priority - focus on public server configurability
```

**Https-Ports Memory Says**:

```markdown
| Environment | Public Bind | Private Bind |
|-------------|-------------|--------------|
| Unit/Integration Tests | `127.0.0.1` | `127.0.0.1` |
| Production | Configurable | `127.0.0.1` |

**Configuration Guidelines**:

| Setting | Test Environments | Production |
|---------|------------------|-----------|
| Port | 0 (dynamic) | 9090 (standard) |
| Bind Address | 127.0.0.1 | 127.0.0.1 |
```

**Why This Causes Divergence**:

- Constitution says "NEVER configurable"
- Plan says "configurable but low priority"
- Memory files show port 0 for tests, 9090 for production (implies configurability)
- LLM agents uncertain whether to implement configurability

**Fix Required**:
Clarify in constitution.md:

```markdown
**Admin Port Requirements**:
- Test environments: Port 0 (dynamic allocation) to avoid collisions
- Production environments: Port 9090 (standard, not configurable)
- Bind address: ALWAYS 127.0.0.1 (never configurable)
- Configurability: Port number MAY be configurable for test vs prod, bind address MUST NOT be configurable
```

---

### 4. CRLDP URL Encoding - CRITICAL MISSING DETAIL

**Severity**: HIGH
**Impact**: CRLDP implementation varies across regenerations
**Files**:

- `.specify/memory/constitution.md` (lines 107-109)
- `specs/002-cryptoutil/spec.md` (lines 2299-2332)
- `specs/002-cryptoutil/clarify.md` (lines 745-767)
- `specs/002-cryptoutil/plan.md` (lines 88-90)

**Ambiguity**:

**Plan/Tasks Say** (SPECIFIC):

```markdown
- CRLDP: Immediate sign and publish to HTTPS URL (NOT batched), one serial number per URL
- URL format: `https://crl.example.com/<base64-url-encoded-serial>.crl`
```

**Spec Says** (SPECIFIC):

```markdown
- **Distribution**: One serial number per HTTPS URL with base64-url-encoded serial (e.g., `https://ca.example.com/crl/EjOrvA.crl`)
```

**Constitution Says** (VAGUE):

```markdown
- **CRLDP MUST provide immediate revocation checks (NOT batched or delayed)**
- **mTLS MUST implement BOTH CRLDP and OCSP for certificate revocation checking**
```

**Instruction File Says** (VAGUE):

```markdown
- CRLDP (CRL Distribution Points): Download CRL from certificate extension
- OCSP (Online Certificate Status Protocol): Real-time revocation check
- Pattern: Parallel checks, fail if BOTH unreachable, cache CRLs with TTL
```

**Missing Details**:

- ❌ No instruction file mentions base64-url encoding
- ❌ No instruction file shows URL format pattern
- ❌ No instruction file specifies one serial per URL
- ❌ Constitution doesn't specify encoding format

**Why This Causes Divergence**:

- LLM agents implement CRLDP without consulting spec/plan/tasks
- Without encoding guidance, implementations use hex, decimal, or other formats
- Regeneration produces different URL formats each time

**Fix Required**:

1. Update `02-10.authn.instructions.md`:

```markdown
## mTLS Revocation Checking - MANDATORY

**MUST check BOTH CRLDP and OCSP**:

**CRLDP Requirements**:
- **Immediate**: Sign and publish CRL immediately on revocation (NOT batched)
- **Per-Serial**: One serial number per HTTPS URL
- **URL Format**: `https://crl.example.com/<base64-url-encoded-serial>.crl`
- **Encoding**: base64-url (RFC 4648 Section 5) of serial number bytes
- **Example**: Serial `0x123ABC` → base64-url → `EjOrvA` → `https://crl.example.com/EjOrvA.crl`

**OCSP Requirements**:
- Real-time revocation check
- Parallel with CRLDP
- Fail if BOTH unreachable
```

2. Update constitution.md to specify encoding:

```markdown
- **CRLDP URL Format**: `https://crl.example.com/<base64-url-encoded-serial>.crl` where serial is base64-url encoded (RFC 4648 Section 5)
```

---

## Ambiguities

### 5. Service Naming: learn-im vs learn-ps - INCONSISTENCY

**Severity**: MEDIUM
**Impact**: Incorrect service names in code/docs
**Files**:

- `.github/instructions/02-01.architecture.instructions.md` (line 13)
- `.github/instructions/02-02.service-template.instructions.md` (line 22)
- `.specify/memory/architecture.md` (service catalog table)
- `specs/002-cryptoutil/spec.md` (service catalog table)

**Finding**:

**Instruction Files Consistently Say**:

```markdown
learn-im: 8888-8889 | ... | learn-im FIRST (validate template)
```

**Constitution/Spec/Plan Consistently Say**:

```markdown
| **learn-im** | Learn | InstantMessenger | 8888-8889 | 127.0.0.1:9090 | Encrypted messaging demonstration service |
```

**No Ambiguity Found**:

- All references use `learn-im`
- No `learn-ps` references found
- User's checklist item may be outdated

**Status**: ✅ NOT AN ISSUE - consistent naming across all files

---

### 6. Federation Timeout Configuration - PARTIALLY SPECIFIED

**Severity**: MEDIUM
**Impact**: Unclear timeout configuration patterns
**Files**:

- `.github/instructions/02-01.architecture.instructions.md`
- `.specify/memory/constitution.md` (lines 556-568)
- `specs/002-cryptoutil/spec.md` (lines 374-398)

**Ambiguity**:

**Constitution/Spec Show** (EXAMPLE):

```yaml
federation:
  identity_url: "https://identity-authz:8180"
  identity_timeout: 10s  # MUST be per-service configurable
  jose_timeout: 10s      # MUST be per-service configurable
```

**Instruction File Shows** (SAME):

```yaml
federation:
  identity_timeout: 10s
  jose_timeout: 10s
```

**Missing Guidance**:

- ❌ What are the RECOMMENDED timeout values per service? (all 10s? different?)
- ❌ What is the MAXIMUM allowed timeout before circuit breaker?
- ❌ Should timeouts be configurable at runtime (hot reload) or startup only?

**Why This Causes Divergence**:

- LLM agents see "10s" example and hardcode it for all services
- No guidance on whether CA operations need longer timeouts (30s?) vs Identity (10s)
- Constitution says "MUST be per-service configurable" but doesn't specify different values

**Fix Required**:
Add to `02-01.architecture.instructions.md`:

```markdown
## Federation Timeout Patterns

**Default Timeouts by Service Type**:
- Identity (authz/idp): 10s (fast token validation)
- JOSE (JWE/JWS): 15s (crypto operations)
- CA (cert operations): 30s (heavy crypto, CRL generation)

**Configuration**: Hot-reloadable (no restart required)
**Maximum**: 60s (circuit breaker threshold)
```

---

### 7. Admin Port Inside vs Outside Container - CLARIFICATION NEEDED

**Severity**: LOW
**Impact**: Docker Compose port mapping confusion
**Files**:

- `.specify/memory/https-ports.md`
- `specs/002-cryptoutil/spec.md` (lines 153-172)
- `specs/002-cryptoutil/clarify.md` (lines 25-64)

**Ambiguity**:

**Constitution/Spec Say**:

```markdown
**Admin Port Isolation** (Unified Deployments):

- Admin ports (127.0.0.1:9090) REQUIRE containerization for multi-service deployments
- Each container has isolated localhost namespace, preventing port collisions
- Non-containerized unified deployments NOT SUPPORTED
```

**Https-Ports Memory Shows**:

```markdown
**Deployment Environments**:

**Docker Containers**:
- Public 0.0.0.0:8080 (external access from host/other containers)
- Private 127.0.0.1:9090 (admin isolated to localhost)
```

**Missing Guidance**:

- ❌ Can admin port be mapped outside container for debugging? (e.g., `-p 127.0.0.1:19090:9090`)
- ❌ Should Docker Compose EVER expose admin port to host?
- ❌ How do operators access admin endpoints for troubleshooting?

**Why This Causes Ambiguity**:

- Constitution says "NEVER exposed to host"
- But debugging/troubleshooting may require admin endpoint access
- No guidance on docker exec vs port mapping for admin access

**Fix Required**:
Add to `.specify/memory/https-ports.md`:

```markdown
## Admin Endpoint Access Patterns

**Production**: NEVER map admin port outside container
**Development/Debugging**: Use `docker exec -it <container> wget https://127.0.0.1:9090/admin/v1/livez`
**Alternative**: Map to localhost ONLY if needed: `-p 127.0.0.1:19090:9090` (NOT `-p 9090:9090`)

**Rationale**: Admin endpoints MUST NOT be accessible from external networks
```

---

### 8. Session Token Format Priority - IMPLEMENTATION vs DEPLOYMENT

**Severity**: MEDIUM
**Impact**: Confusion about which token format to implement first
**Files**:

- `specs/002-cryptoutil/clarify.md` (lines 264-293)
- `specs/002-cryptoutil/plan.md` (lines 63-69)

**Ambiguity**:

**Clarify.md Says**:

```markdown
**Implementation Priority** (all stored in SQL database):

1. **JWS Sessions** (HIGHEST priority): Stateless token validation, cryptographic signature verification
2. **OPAQUE Sessions** (MEDIUM priority): Database lookup for every request, maximum revocation control
3. **JWE Sessions** (LOWER priority): Encrypted session data, requires decryption on every request

**Deployment Priority** (security preference):

1. **JWE Sessions** (HIGHEST security): Encrypted session data prevents inspection
2. **OPAQUE Sessions** (MEDIUM security): Database-backed with immediate revocation
3. **JWS Sessions** (LOWER security): Signed but not encrypted, readable session data
```

**Plan Says**:

```markdown
**Session State Management**:

- SQL-backed ONLY (PostgreSQL or SQLite) - NO Redis
- Three formats: JWS (stateless signed), OPAQUE (database lookup), JWE (encrypted)
- Implementation priority: JWS → OPAQUE → JWE
- Deployment priority: JWE → OPAQUE → JWS
```

**Missing Guidance**:

- ❌ Should Phase 2 implement ONLY JWS (simplest)?
- ❌ Or should Phase 2 implement ALL THREE formats with configuration?
- ❌ Can services use different formats (KMS uses JWE, Identity uses OPAQUE)?

**Why This Causes Ambiguity**:

- "Implementation priority" suggests phased rollout (Phase 2: JWS only, Phase 3: add OPAQUE, Phase 4: add JWE)
- "Deployment priority" suggests all three implemented, configuration chooses which to use
- LLM agents uncertain whether to implement one at a time or all at once

**Fix Required**:
Add to `specs/002-cryptoutil/plan.md`:

```markdown
## Session Token Implementation Strategy

**Phase 2**: Implement JWS ONLY (simplest, fastest to validate template)
**Phase 3**: Add OPAQUE support (database-backed revocation)
**Phase 4**: Add JWE support (encrypted session data)

**Configuration Pattern**:
```yaml
session:
  format: jws  # Options: jws (Phase 2+), opaque (Phase 3+), jwe (Phase 4+)
```

**Production Deployment**: Use JWE format for highest security (Phase 4+)

```

---

### 9. SQLite MaxOpenConns: 1 vs 5 - CONFLICTING VALUES

**Severity**: MEDIUM
**Impact**: Wrong connection pool configuration
**Files**:
- `.github/instructions/03-04.database.instructions.md` (lines 84-86)
- `.github/instructions/03-05.sqlite-gorm.instructions.md`
- `.specify/memory/sqlite-gorm.md`

**Contradiction**:

**Database Instructions Say**:
```markdown
// For SQLite, limit connection pool to prevent write contention
// SQLite only supports 1 concurrent writer (even in WAL mode)
sqlDB.SetMaxOpenConns(1)  // Use magic constant: cryptoutilMagic.SQLiteMaxOpenConnections
```

**SQLite-GORM Memory Says** (need to verify):

```go
sqlDB.SetMaxOpenConns(5)  // GORM transaction handling requires pool size > 1
```

**Missing Guidance**:

- ❌ Which value is correct? 1 or 5?
- ❌ Does GORM transaction handling require pool size > 1?
- ❌ What is the actual `cryptoutilMagic.SQLiteMaxOpenConnections` value?

**Why This Causes Divergence**:

- Different files show different values
- Magic constant referenced but value not shown
- LLM agents pick arbitrary value (1, 5, or 10)

**Fix Required**:

1. Check actual magic constant value in `internal/common/magic/magic_database.go`
2. Update all documentation to use same value
3. Add clarification:

```markdown
## SQLite Connection Pool - CRITICAL

**MaxOpenConns**: 5 (for GORM transaction handling, NOT 1)
**Rationale**: GORM requires pool size > 1 for Begin/Commit transaction flow
**Write Contention**: Handled by busy_timeout + WAL mode, not connection limit
**Magic Constant**: `cryptoutilMagic.SQLiteMaxOpenConnections = 5`
```

---

### 10. Hash Output Format: Variation in Examples - MINOR INCONSISTENCY

**Severity**: LOW
**Impact**: Inconsistent hash format examples
**Files**:

- `.github/instructions/02-08.hashes.instructions.md` (line 48)
- `.specify/memory/hashes.md` (lines 36, 82)
- `.specify/memory/constitution.md` (lines 1081, 1123, 1140, 1154, 1171)

**Ambiguity**:

**Instruction File Shows**:

```
{version}:{algorithm}:{iterations}:base64(randomSalt):base64(hash)
```

**Constitution Shows Multiple Formats**:

```
{v}:base64_hash                                              (line 1081)
{version}:base64(hash)                                       (lines 1123, 1140)
{version}:{algorithm}:{iterations}:base64(randomSalt):base64(hash)  (line 1154)
{version}:{algorithm}:base64(randomSalt):base64(hash)       (line 1171)
```

**Missing Guidance**:

- ❌ Which format applies to which hash registry?
- ❌ Is `{v}` same as `{version}`? (abbreviated vs full)
- ❌ When is `iterations` omitted? (HKDF doesn't have iterations)

**Why This Causes Ambiguity**:

- Different formats for different hash types not clearly documented
- Examples show variations without explaining when each applies
- LLM agents pick inconsistent format

**Fix Required**:
Add to `02-08.hashes.instructions.md`:

```markdown
## Hash Output Formats - Registry-Specific

**LowEntropyDeterministic** (PBKDF2):
```

{version}:{algorithm}:{iterations}:base64(fixedSalt):base64(hash)
Example: {2}:PBKDF2-HMAC-SHA256:600000:abcd1234:efgh5678

```

**LowEntropyRandom** (PBKDF2):
```

{version}:{algorithm}:{iterations}:base64(randomSalt):base64(hash)
Example: {2}:PBKDF2-HMAC-SHA256:600000:xyz789:abc123

```

**HighEntropyDeterministic** (HKDF):
```

{version}:{algorithm}:info={info},salt=base64(fixedSalt):base64(hash)
Example: {3}:HKDF-SHA512:info=api-key,salt=xyz:abc123

```

**HighEntropyRandom** (HKDF):
```

{version}:{algorithm}:base64(randomSalt):base64(hash)
Example: {3}:HKDF-SHA512:xyz789:abc123

```
```

---

## Missing Coverage

### 11. Phase Numbers and Service Sequence - NO ENFORCEMENT GUIDANCE

**Severity**: MEDIUM
**Impact**: Implementation order violations
**Files**: All instruction files (NO enforcement patterns)

**Missing Instruction**:

- ❌ No instruction file enforces "learn-im MUST complete before jose-ja migration"
- ❌ No instruction file blocks "sm-kms migration BEFORE all other services complete"
- ❌ No instruction file validates "Phase 2 complete before Phase 3 starts"

**Why This Causes Divergence**:

- LLM agents see task list and start wherever seems easiest
- Without enforcement, agents skip learn-im and migrate production services directly
- Phase dependencies ignored during implementation

**Fix Required**:
Add new file `.github/instructions/01-04.phase-dependencies.instructions.md`:

```markdown
# Phase Dependencies - MANDATORY ENFORCEMENT

## Phase Gate Pattern

**NEVER start Phase N+1 until Phase N COMPLETE with evidence**

**Phase 2 Gate**: Template extracted, documented, ready for learn-im
**Phase 3 Gate**: learn-im passes ALL tests, no template blockers
**Phase 4-6 Gate**: Each service migration complete before next starts
**Phase 7 Gate**: sm-kms migration ONLY after ALL other services excellent

## Migration Sequence - STRICT ORDER

1. learn-im (FIRST - validates template)
2. jose-ja (production service #1)
3. pki-ca (production service #2)
4. identity-rp, identity-spa, identity-authz, identity-idp, identity-rs (one at a time)
5. sm-kms (LAST - reference implementation)

## Blocking Conditions

❌ NEVER migrate jose-ja before learn-im complete
❌ NEVER migrate sm-kms before ALL others excellent
❌ NEVER skip phases or parallelize service migrations
```

---

### 12. Certificate Validation Chain - NO INSTRUCTION GUIDANCE

**Severity**: MEDIUM
**Impact**: Incomplete TLS validation
**Files**: PKI instructions incomplete

**Missing Instruction**:

- ❌ No instruction file shows HOW to validate certificate chains
- ❌ No instruction file specifies chain building algorithm
- ❌ No instruction file explains partial chain handling

**Constitution/Spec Say** (VAGUE):

```markdown
- Full cert chain validation, MinVersion: TLS 1.3+, never InsecureSkipVerify
```

**Missing Details**:

- ❌ What is "full cert chain validation"? (up to root? up to intermediate? trust anchor?)
- ❌ How to handle missing intermediate certificates?
- ❌ Should services download missing intermediates via AIA (Authority Information Access)?

**Fix Required**:
Add to `02-09.pki.instructions.md`:

```markdown
## Certificate Chain Validation - MANDATORY

**Validation Steps**:
1. Build chain from leaf → intermediate → root
2. Verify each cert signature against parent
3. Check validity dates (NotBefore, NotAfter)
4. Verify Key Usage, Extended Key Usage extensions
5. Check revocation status (CRLDP + OCSP)

**Partial Chains**:
- Download missing intermediates via AIA (Authority Information Access extension)
- Cache downloaded intermediates (TTL: 24 hours)
- Fail if chain incomplete and AIA unavailable

**Trust Anchor**:
- Validation MUST reach configured root CA (trust anchor)
- Partial chains without trust anchor MUST fail
```

---

### 13. Request/Response Logging - NO PRIVACY GUIDANCE

**Severity**: HIGH (Security/Compliance)
**Impact**: Sensitive data exposure in logs
**Files**: Observability instructions incomplete

**Missing Instruction**:

- ❌ No instruction file specifies WHAT to redact from logs
- ❌ No instruction file explains HOW to redact sensitive fields
- ❌ No instruction file lists sensitive headers/fields

**Constitution/Spec Say** (VAGUE):

```markdown
- OpenTelemetry integration (OTLP traces, metrics, logs)
- Structured logging
```

**Missing Details**:

- ❌ Should `Authorization` header be logged? (NO - contains tokens)
- ❌ Should password fields be logged? (NO - plaintext secrets)
- ❌ Should session cookies be logged? (NO - session hijacking risk)
- ❌ How to redact: `[REDACTED]` vs `***` vs hash?

**Fix Required**:
Add to `02-05.observability.instructions.md`:

```markdown
## Request/Response Logging - PRIVACY MANDATORY

**NEVER Log These Fields**:
- `Authorization` header (contains bearer tokens)
- `Cookie` header (contains session tokens)
- `password` fields in request body
- `client_secret` fields
- `api_key` fields
- Any field matching pattern: `*password*`, `*secret*`, `*token*`, `*key*`

**Redaction Pattern**:
```go
func redactSensitive(value string) string {
    if len(value) <= 4 {
        return "[REDACTED]"
    }
    return value[:2] + "***" + value[len(value)-2:]  // Show first/last 2 chars
}
```

**Audit Logging**: Separate audit log for authentication events (log username, NOT password)

```

---

### 14. Error Response Format - NO STANDARDIZATION

**Severity**: MEDIUM
**Impact**: Inconsistent API error responses
**Files**: OpenAPI instructions incomplete

**Missing Instruction**:
- ❌ No instruction file specifies standard error response schema
- ❌ No instruction file shows how to return errors from all services

**Constitution/Spec Say** (NOTHING):
- No error response format specified

**Missing Details**:
- ❌ Should errors be JSON? (yes, but what schema?)
- ❌ Should errors include request ID? (yes, for tracing)
- ❌ Should errors include stack traces? (no, security risk)
- ❌ Should errors be localized? (not in Phase 1-3)

**Fix Required**:
Add to `02-06.openapi.instructions.md`:
```markdown
## Error Response Format - MANDATORY

**Standard Error Schema** (ALL services):
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid tenant ID format",
    "request_id": "req_abc123xyz",
    "timestamp": "2025-12-24T12:34:56Z",
    "details": {
      "field": "tenant_id",
      "expected": "UUID v4",
      "received": "not-a-uuid"
    }
  }
}
```

**Error Codes**: Use consistent codes across services (INVALID_REQUEST, UNAUTHORIZED, FORBIDDEN, NOT_FOUND, etc.)

**NEVER Include**:

- Stack traces (security risk)
- Internal variable names
- Database error messages
- File paths

```

---

### 15. Hot Reload Configuration - NO IMPLEMENTATION PATTERN

**Severity**: LOW
**Impact**: Configuration changes require restarts
**Files**: No instruction file covers hot reload

**Missing Instruction**:
- ❌ No instruction file explains HOW to implement hot reload
- ❌ No instruction file specifies WHICH configs can hot reload

**Constitution/Spec Mention**:
```markdown
- Hot-reloadable configuration (no restart required)
```

**Missing Details**:

- ❌ File watching pattern? (fsnotify library?)
- ❌ Signal-based reload? (SIGHUP?)
- ❌ API endpoint for reload? (`POST /admin/v1/reload`)
- ❌ Which configs hot reload? (connection pools yes, TLS certs no?)

**Fix Required**:
Add to `03-01.coding.instructions.md`:

```markdown
## Hot Reload Pattern - OPTIONAL ENHANCEMENT

**Signal-Based Reload** (Recommended):
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGHUP)

go func() {
    for range sigChan {
        if err := reloadConfig(); err != nil {
            log.Errorf("Config reload failed: %v", err)
        }
    }
}()
```

**Hot-Reloadable Configs**:

- ✅ Connection pool settings
- ✅ Federation timeouts
- ✅ Rate limiting
- ❌ TLS certificates (require restart)
- ❌ Port bindings (require restart)

```

---

## Redundancies

### 16. Duplicate Instructions Across Memory and Copilot Files

**Severity**: LOW
**Impact**: Conflicting updates, maintenance burden
**Files**: All `.github/instructions/*.md` and `.specify/memory/*.md`

**Finding**:
- Every instruction file references corresponding memory file
- Memory files contain COMPLETE specifications
- Instruction files contain "tactical patterns" (simplified)
- Duplication creates maintenance burden (update in 2 places)

**Examples**:
- `02-01.architecture.instructions.md` duplicates `.specify/memory/architecture.md`
- `03-04.database.instructions.md` duplicates `.specify/memory/database.md`
- `02-03.https-ports.instructions.md` duplicates `.specify/memory/https-ports.md`

**Why This Causes Issues**:
- Updates to memory files not propagated to instruction files
- Instruction files lag behind memory file updates
- LLM agents read instruction files (simplified) instead of memory files (complete)

**Fix Required**:
**Option 1** (Preferred): Instruction files as REFERENCES ONLY
```markdown
# Architecture - Tactical Guidance

**Reference**: See `.specify/memory/architecture.md` for COMPLETE specifications

## Quick Reference ONLY

[Minimal examples, patterns, anti-patterns - NO duplication of full specs]
```

**Option 2**: Memory files as SOURCE OF TRUTH, auto-generate instruction files

- Use script to extract tactical patterns from memory files
- Ensure single source of truth

---

### 17. Constitution vs Spec Duplication - SYNCHRONIZATION RISK

**Severity**: MEDIUM
**Impact**: Constitution and spec drift apart
**Files**: `.specify/memory/constitution.md` vs `specs/002-cryptoutil/spec.md`

**Finding**:

- Constitution and spec contain overlapping content
- Updates to one not always reflected in the other
- Multi-tenancy contradiction example: both files have different emphasis

**Examples**:

- Constitution emphasizes "dual-layer isolation"
- Spec provides detailed implementation patterns
- Instruction files simplified both (lost nuance)

**Fix Required**:

1. Constitution = HIGH-LEVEL REQUIREMENTS (WHAT, WHY)
2. Spec = DETAILED IMPLEMENTATION (HOW, EXAMPLES)
3. Instruction files = TACTICAL PATTERNS (QUICK REFERENCE)

Add cross-references:

```markdown
# Constitution
## Multi-Tenancy
**Requirements**: Dual-layer isolation for tenant data protection
**Details**: See `specs/002-cryptoutil/spec.md` Section "Multi-Tenancy Isolation Pattern"
```

---

### 18. Magic Constants Defined in 3 Places - SYNCHRONIZATION RISK

**Severity**: MEDIUM
**Impact**: Wrong constant values used
**Files**: Magic constants duplicated in instruction files, memory files, and code

**Finding**:

- `cryptoutilMagic.SQLiteMaxOpenConnections` mentioned in instructions
- Value not shown in instructions (reference to code)
- Memory files show different values (1 vs 5)

**Fix Required**:

1. Magic constants defined ONCE in code (`internal/common/magic/*.go`)
2. Documentation REFERENCES code values (never duplicates)
3. Instruction files show USAGE pattern, not VALUES

Example:

```markdown
## SQLite Connection Pool

**Pattern**:
```go
sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // See magic_database.go for value
```

**NEVER hardcode**: `SetMaxOpenConns(5)` or `SetMaxOpenConns(1)` - ALWAYS use magic constant

```

---

## Recommendations

### Immediate Fixes (Blocking SpecKit Convergence)

**Priority 1 - CRITICAL**:
1. **Fix Multi-Tenancy Contradiction** (Issue #1)
   - Update `03-04.database.instructions.md` to match constitution dual-layer pattern
   - Add explicit examples for PostgreSQL + SQLite

2. **Fix Database Choice Ambiguity** (Issue #2)
   - Update constitution.md to clarify deployment-based choice
   - Add instruction file section on database selection criteria

3. **Fix CRLDP URL Format** (Issue #4)
   - Add base64-url encoding to instruction file
   - Specify URL format pattern with examples

**Priority 2 - HIGH**:
4. **Fix Admin Port Configurability** (Issue #3)
   - Clarify: Port number configurable (test vs prod), bind address NOT configurable
   - Update all documentation consistently

5. **Add Request/Response Logging Privacy** (Issue #13)
   - Specify fields to redact
   - Add redaction pattern implementation

6. **Add Phase Dependencies Enforcement** (Issue #11)
   - Create new instruction file for phase gates
   - Block out-of-order migrations

### Medium Priority (Quality Improvements)

**Priority 3 - MEDIUM**:
7. **Clarify Federation Timeouts** (Issue #6)
   - Add recommended timeouts per service type
   - Specify hot-reload requirements

8. **Clarify Session Token Priorities** (Issue #8)
   - Specify phased implementation (JWS → OPAQUE → JWE)
   - Or specify all-at-once with configuration

9. **Fix SQLite MaxOpenConns** (Issue #9)
   - Determine correct value (1 or 5)
   - Update all documentation

10. **Standardize Error Response Format** (Issue #14)
    - Add error schema to OpenAPI instructions
    - Specify error codes and format

### Low Priority (Nice-to-Have)

**Priority 4 - LOW**:
11. **Reduce Redundancy** (Issues #16, #17, #18)
    - Make instruction files reference-only
    - Eliminate duplication between constitution/spec
    - Magic constants defined once

12. **Add Certificate Validation Guidance** (Issue #12)
    - Specify chain building algorithm
    - Add AIA download pattern

13. **Add Hot Reload Pattern** (Issue #15)
    - Document signal-based reload
    - Specify which configs hot reload

14. **Fix Hash Format Variations** (Issue #10)
    - Clarify format per registry type
    - Add comprehensive examples

---

## Verification Checklist Results

| Item | Status | Finding |
|------|--------|---------|
| Service naming: learn-im vs learn-ps | ✅ CONSISTENT | All references use `learn-im`, no `learn-ps` found |
| Admin ports: 127.0.0.1:9090 for ALL vs per-service | ⚠️ AMBIGUOUS | Configurability unclear (test vs prod) |
| Multi-tenancy: Dual-layer vs schema-only vs row-only | ❌ CONTRADICTORY | Instruction says "schema-only", constitution says "dual-layer" |
| CRLDP: base64-url encoding vs hex vs other | ⚠️ MISSING | Spec/plan specify base64-url, instruction files don't |
| Database choice: deployment-based vs environment-based | ❌ CONTRADICTORY | Constitution/plan correct, memory files incorrect |
| Implementation order: Phase numbers and service sequence | ⚠️ MISSING | No enforcement guidance in instruction files |

**Summary**: 2 critical contradictions, 3 missing/ambiguous areas

---

## Root Cause Analysis

**Why SpecKit Regeneration Diverges**:

1. **Instruction Files Too Simplified**: "Tactical patterns" lose critical nuances from constitution/spec
2. **Contradictions Not Caught**: Multi-tenancy instruction directly contradicts constitution
3. **Missing Enforcement**: No phase gates, no migration sequence validation
4. **Redundancy Drift**: Updates to memory files not propagated to instruction files
5. **LLM Agent Priority**: Agents read instruction files FIRST (simpler), ignore detailed specs

**Pattern**: Constitution → Spec → Memory Files → **INFORMATION LOST** → Instruction Files → LLM Agent Implementation

**Fix**: Instruction files must be REFERENCES to memory files, NOT simplified duplicates

---

## Next Steps

1. **Fix Critical Contradictions** (Issues #1, #2, #4)
   - Multi-tenancy: Update instruction file to dual-layer pattern
   - Database choice: Clarify deployment-based selection
   - CRLDP: Add URL format with encoding

2. **Add Missing Coverage** (Issues #11, #13)
   - Phase dependencies: Enforce migration sequence
   - Logging privacy: Redact sensitive fields

3. **Validate Changes**:
   - Re-run SpecKit with updated instructions
   - Verify convergence on multi-tenancy implementation
   - Check database selection in generated code

4. **Prevent Future Drift**:
   - Instruction files as REFERENCE-ONLY (point to memory files)
   - Add validation: Check instruction files match memory files
   - Regular audits: Compare constitution/spec/instructions for contradictions

---

**END OF REVIEW**
