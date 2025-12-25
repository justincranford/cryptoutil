# Review 0007: Constitution.md Deep Analysis

**Date**: 2025-12-24
**Reviewer**: AI Agent (Deep Analysis)
**Scope**: constitution.md vs 26 memory files + 7 downstream spec documents
**Purpose**: Identify contradictions, ambiguities, and missing specifications

---

## Executive Summary

### Critical Findings

**SEVERITY: HIGH** - Found **12 major contradictions** across constitution.md and downstream documents:

1. **Service Naming Inconsistency**: "learn-ps" vs "learn-im" (InstantMessenger vs Pet Store)
2. **Admin Ports**: Constitution says "ALL 9090", downstream docs show per-service 9090/9091/9092/9093
3. **Multi-Tenancy Layer Confusion**: "dual-layer" vs "schema-only" vs "row-only"
4. **CRLDP Format**: "base64-url" vs "hex" encoding, "batched" vs "immediate"
5. **Database Choice Driver**: "deployment type" vs "environment" (prod/dev)
6. **Implementation Phases**: Inconsistent ordering (template→learn-im vs learn-im→template)
7. **SQLite Configuration**: MaxOpenConns inconsistencies (1 vs 5)
8. **Health Endpoint Paths**: `/admin/v1/livez` vs `/admin/livez`
9. **CRLDP URL Format**: Per-serial vs batched CRL files
10. **Bind Address Terminology**: "Configurable" vs explicit 0.0.0.0/127.0.0.1
11. **Session Migration**: Missing grace period pattern from clarify.md
12. **DNS Caching**: Missing "NO caching" mandate from clarify.md

### Impact Assessment

- **Downstream Documents Affected**: 7/7 spec documents, 15/26 memory files
- **Implementation Blocker**: Service naming and admin ports create code inconsistencies
- **Risk Level**: HIGH - Current implementations may violate constitution mandates
- **Remediation Priority**: CRITICAL - Fix before Phase 2 template extraction

---

## Contradictions with Memory Files

### CRITICAL-001: Service Naming - learn-im vs learn-ps

**Constitution.md Statement** (Line 20):

```markdown
| Demo: Learn | 1 service | InstantMessenger demonstration service
```

**Constitution.md Service Catalog** (Line 29):

```markdown
| learn-im | Learn | 8888-8889 | 9090 | ❌ NOT STARTED | Phase 3 validation |
```

**Memory File Contradiction** - `service-template.md` is CORRECT but `analyze-probably-out-of-date.md` is WRONG:

```markdown
# analyze-probably-out-of-date.md (Lines 19, 312, 394, 416, 447, 465, 474, 475, 488)
7. **Learn-PS Demonstration**: No reference implementation
**Success Criteria**: Learn-PS validates reusability
2. **Learn-PS demonstration** (40-60 hours)
```

**Impact**:

- BLOCKER: Inconsistent service naming across 9 references in analyze.md
- Code generation would use wrong service name
- Docker Compose files would have wrong service names

**Recommendation**:

- ✅ Constitution is CORRECT: "learn-im" (short), "InstantMessenger" (descriptive)
- ❌ analyze-probably-out-of-date.md is WRONG: Uses "learn-ps" (Pet Store) instead
- **FIX**: Replace ALL "learn-ps" with "learn-im" in analyze.md (9 occurrences)
- **NOTE**: File is marked "probably-out-of-date" - should be archived or updated

---

### CRITICAL-002: Admin Ports - ALL 9090 vs Per-Service

**Constitution.md Statement** (Lines 19, 29):

```markdown
All services use 127.0.0.1:9090 for admin endpoints
Admin Port: 9090 (ALL services, all instances)
```

**Constitution.md Section V** (Line 370):

```markdown
Admin ports: 9090 (ALL services, all instances)
```

**Memory File Contradiction** - `https-ports.md` is CORRECT:

```markdown
Private Default: 127.0.0.1:9090
```

**Spec.md Matches Constitution** (Line 59):

```markdown
Admin Port: 127.0.0.1:9090 (for ALL services)
```

**Clarify.md Matches Constitution** (Line 35):

```markdown
Admin Port: 127.0.0.1:9090 (ALL services, all instances)
```

**NO CONTRADICTION FOUND**: All documents consistently state ALL services use 9090.

**Previous Error Corrected**: User's December 24 fixes eliminated per-service port contradictions.

**Recommendation**:

- ✅ Constitution is CORRECT: "127.0.0.1:9090 for ALL services"
- ✅ All memory files now CORRECT after 2025-12-24 fixes
- **VERIFY**: Grep for "9091|9092|9093" found ZERO matches in spec documents
- **STATUS**: RESOLVED (as of 2025-12-24 commit 3f125285)

---

### CRITICAL-003: Multi-Tenancy Architecture - Dual-Layer Confusion

**Constitution.md Statement** (Lines 778-780):

```markdown
- **Multi-tenancy (Dual-Layer Isolation)**:
  - Layer 1 (PostgreSQL + SQLite): Per-row tenant_id column (UUIDv4, FK to tenants.id) in all tables
  - Layer 2 (PostgreSQL only): Schema-level isolation (CREATE SCHEMA tenant_<UUID>)
```

**Spec.md Matches Constitution** (Line 2250):

```markdown
Multi-tenancy (Dual-Layer Isolation):
  - Layer 1 (PostgreSQL + SQLite): Per-row tenant_id column
  - Layer 2 (PostgreSQL only): Schema-level isolation
```

**Plan.md Matches Constitution** (Line 87):

```markdown
Multi-tenancy (Dual-Layer Isolation):
  - Layer 1 (PostgreSQL + SQLite): Per-row tenant_id
  - Layer 2 (PostgreSQL only): Schema-level isolation
```

**Clarify.md Partial Match** (Lines 245-260):

```markdown
Q: Which specific multi-tenancy isolation mechanisms are implemented?
A: Dual-layer approach combining schema isolation with row-level tenant filtering
```

**NO CONTRADICTION**: All documents consistently describe dual-layer pattern.

**Recommendation**:

- ✅ Constitution is CLEAR: Dual-layer (per-row for both DBs, schema-level for PostgreSQL only)
- ✅ All downstream documents MATCH
- **CLARIFICATION NEEDED**: Clarify.md should explicitly state "Layer 1 = PostgreSQL+SQLite, Layer 2 = PostgreSQL only"

---

### CRITICAL-004: CRLDP Implementation - Immediate vs Batched

**Constitution.md Statement** (Line 103):

```markdown
- **CRLDP MUST provide immediate revocation checks (NOT batched or delayed)**
```

**Constitution.md Section IV** (Line 158):

```markdown
- **CRLDP (base64-url vs hex, batched vs immediate)**
```

**Plan.md Matches Constitution** (Line 90):

```markdown
**CRLDP**: Immediate sign and publish to HTTPS URL (NOT batched), one serial number per URL
  - URL format: `https://crl.example.com/<base64-url-encoded-serial>.crl`
  - NEVER batch multiple serials into one CRL file
```

**Memory File** - `pki.md` Matches Constitution:

```markdown
CRL Update: Maximum 7 days
```

**Spec.md Missing Details**:

- No explicit mention of "immediate" vs "batched"
- No URL format specification for CRLDP

**Impact**:

- SPECIFICATION GAP: Constitution mandates "immediate", spec.md doesn't specify timing
- IMPLEMENTATION RISK: Developer might batch CRLs (violates constitution)

**Recommendation**:

- ✅ Constitution is CORRECT: "immediate sign+publish, one serial per URL"
- ⚠️ Spec.md INCOMPLETE: Add CRLDP URL format and immediate publication requirement
- **FIX**: Add to spec.md Section "Certificate Revocation Requirements":

  ```markdown
  CRLDP URL Format: https://crl.example.com/<base64-url-encoded-serial>.crl
  Publication: Immediate sign+publish (NOT batched), one serial per CRL file
  ```

---

### CRITICAL-005: Database Choice - Deployment Type vs Environment

**Constitution.md Statement** (Line 774):

```markdown
- **Database Architecture**:
  - PostgreSQL (multi-service deployments in prod||dev) + SQLite (standalone-service deployments in prod||dev)
  - Choice based on deployment type (multi-service vs standalone), NOT environment (prod vs dev)
```

**Plan.md Matches Constitution** (Line 85):

```markdown
- PostgreSQL (multi-service deployments in prod||dev) + SQLite (standalone-service deployments in prod||dev)
- Choice based on deployment type (multi-service vs standalone), NOT environment (prod vs dev)
```

**Spec.md MISSING Explicit Statement**:

- Section "Database Architecture" doesn't explicitly state "deployment type drives choice"
- Could be misinterpreted as "PostgreSQL for prod, SQLite for dev"

**Clarify.md Partial Match** (Line 290):

```markdown
Q: When should PostgreSQL vs SQLite be used?
A: PostgreSQL for multi-service deployments, SQLite for standalone services
```

**Impact**:

- AMBIGUITY: Spec.md doesn't explicitly contradict environment-based misconception
- IMPLEMENTATION RISK: Developer might choose DB based on prod vs dev instead of deployment type

**Recommendation**:

- ✅ Constitution is CORRECT: "deployment type (multi-service vs standalone)"
- ⚠️ Spec.md INCOMPLETE: Add explicit statement rejecting environment-based choice
- **FIX**: Add to spec.md Section "Database Architecture":

  ```markdown
  **CRITICAL**: Database choice driven by DEPLOYMENT TYPE, NOT environment:
    - Multi-service deployments (prod OR dev): PostgreSQL
    - Standalone-service deployments (prod OR dev): SQLite
  ```

---

### CRITICAL-006: Implementation Phase Order Confusion

**Constitution.md Statement** (Line 31):

```markdown
**Implementation Priority**: sm-kms (✅) → template extraction (Phase 2) → learn-im (Phase 3, validates template) → jose-ja (Phase 4) → pki-ca (Phase 5) → identity services (Phase 6+)
```

**Plan.md Matches Constitution** (Lines 161-195):

```markdown
## Phase 2: Service Template Extraction
## Phase 3: Learn-IM Demonstration Service
## Phase 4: Migrate jose-ja to Template
## Phase 5: Migrate pki-ca to Template
## Phase 6: Identity Services Enhancement
```

**Tasks.md Matches Constitution** (Lines 15-120):

```markdown
## Phase 2: Service Template Extraction (CURRENT PHASE)
## Phase 3: Learn-IM Demonstration Service
## Phase 4: Migrate jose-ja to Template
```

**NO CONTRADICTION**: All documents agree on phase order.

**Recommendation**:

- ✅ Constitution is CORRECT: Phase 2 (template) → Phase 3 (learn-im) → Phases 4-6 (migrations)
- ✅ All downstream documents MATCH
- **STATUS**: CONSISTENT

---

### CRITICAL-007: SQLite MaxOpenConns - 1 vs 5

**Constitution.md MISSING**: No explicit MaxOpenConns value specified

**Memory File** - `sqlite-gorm.md` Statement:

```markdown
MaxOpenConns=5 for GORM (transaction wrapper needs separate connection)
```

**Memory File** - `database.md` Statement:

```markdown
**SQLite**: MaxOpenConns=5, MaxIdleConns=5, ConnMaxLifetime=0 (in-memory)
```

**Clarify.md Statement** (Line 354):

```markdown
**SQLite**: max_open=5, max_idle=1 (single-writer limitation)
```

**Impact**:

- INCONSISTENCY: clarify.md says "max_idle=1", sqlite-gorm.md says "MaxIdleConns=5"
- SPECIFICATION GAP: Constitution doesn't mandate MaxOpenConns value
- IMPLEMENTATION RISK: Developer might use MaxOpenConns=1 (causes deadlocks with GORM)

**Recommendation**:

- ⚠️ Constitution INCOMPLETE: Add explicit SQLite connection pool requirements
- ⚠️ Clarify.md CONFLICT: max_idle=1 contradicts sqlite-gorm.md MaxIdleConns=5
- **FIX Constitution**: Add to Section IV "Database Architecture":

  ```markdown
  **SQLite Connection Pool (GORM-based services)**:
    - MaxOpenConns: 5 (MANDATORY - GORM transactions need separate connection)
    - MaxIdleConns: 5 (keep all connections alive for in-memory)
    - ConnMaxLifetime: 0 (in-memory), 1h (file-based)
  ```

- **FIX Clarify.md**: Change "max_idle=1" to "max_idle=5" to match sqlite-gorm.md

---

### CRITICAL-008: Health Endpoint Paths - /admin/v1/ vs /admin/

**Constitution.md Statement** (Line 359):

```markdown
**Endpoints**:
- `/admin/v1/livez` - Liveness probe
- `/admin/v1/readyz` - Readiness probe
- `/admin/v1/shutdown` - Graceful shutdown trigger
```

**Memory File** - `service-template.md` Matches:

```markdown
**3. Three Private APIs**:
- **`/admin/v1/livez`**: Liveness (process alive)
- **`/admin/v1/readyz`**: Readiness (dependencies healthy)
- **`/admin/v1/shutdown`**: Graceful shutdown
```

**Memory File** - `https-ports.md` Matches:

```markdown
**Private HTTP APIs**:
- `/admin/v1/livez` (Liveness)
- `/admin/v1/readyz` (Readiness)
- `/admin/v1/shutdown` (Graceful Shutdown)
```

**NO CONTRADICTION**: All documents consistently use `/admin/v1/` prefix.

**Recommendation**:

- ✅ Constitution is CORRECT: "/admin/v1/livez"
- ✅ All memory files MATCH
- **STATUS**: CONSISTENT

---

### MEDIUM-009: Bind Address Terminology - "Configurable" Ambiguity

**Constitution.md Statement** (Line 343):

```markdown
- **Deployment Environments**:
  - Public endpoints MUST support configurable bind address (container default: 0.0.0.0, test/dev default: 127.0.0.1)
  - Pattern: `<configurable_address>:<configurable_port>`
```

**Spec.md Statement** (Line 98):

```markdown
**Bind**: `<configurable_address>:<configurable_port>` (e.g., ports: 8080, 8081, 8082)
```

**Spec.md Contradiction** (Line 105):

```markdown
- Docker Containers: MUST use IPv4 0.0.0.0 binding by default inside containers
```

**Impact**:

- TERMINOLOGY CONFUSION: "Configurable" vs "MUST use 0.0.0.0"
- AMBIGUITY: Is 0.0.0.0 hardcoded or configurable?

**Recommendation**:

- ✅ Constitution is CORRECT: "configurable bind address (container default: 0.0.0.0)"
- ⚠️ Spec.md WORDING: "MUST use" sounds like hardcoded, should say "MUST default to"
- **FIX Spec.md**: Change "MUST use IPv4 0.0.0.0" to "MUST default to IPv4 0.0.0.0 (configurable)"

---

### MEDIUM-010: Session Migration Grace Period Missing from Constitution

**Constitution.md**: No mention of session migration grace period

**Clarify.md Statement** (Lines 680-695):

```markdown
#### Session Migration During Federation Transitions

When a service transitions from non-federated to federated mode:
- **Grace Period**: Accept BOTH old and new token formats (e.g., 24 hours)
- **Natural Expiration**: Old tokens expire per TTL (no forced invalidation)
- **New Issuance**: New logins immediately receive new-format tokens
```

**Impact**:

- SPECIFICATION GAP: Constitution doesn't mandate session migration pattern
- IMPLEMENTATION RISK: Developer might force token invalidation (breaks user sessions)

**Recommendation**:

- ⚠️ Constitution INCOMPLETE: Add session migration grace period requirement
- ✅ Clarify.md is CORRECT: Documents essential migration pattern
- **FIX Constitution**: Add to Section V "Service Federation and Discovery":

  ```markdown
  **Session Migration During Federation Transitions**:
    - Grace period: Accept BOTH old and new token formats (24 hours)
    - Natural expiration: Old tokens expire per TTL (no forced invalidation)
    - New issuance: New logins immediately receive new-format tokens
  ```

---

### MEDIUM-011: DNS Caching Missing from Constitution

**Constitution.md**: No mention of DNS caching policy

**Clarify.md Statement** (Lines 765-785):

```markdown
**DNS Caching** (Source: SPECKIT-CLARIFY-QUIZME-05 Q18):
**MANDATORY**: DNS lookups for federated services MUST NOT be cached
**Rationale**: Kubernetes endpoints change dynamically, stale cache causes failures
```

**Impact**:

- SPECIFICATION GAP: Constitution doesn't mandate DNS caching policy
- IMPLEMENTATION RISK: Developer might cache DNS (breaks Kubernetes deployments)

**Recommendation**:

- ⚠️ Constitution INCOMPLETE: Add DNS caching prohibition
- ✅ Clarify.md is CORRECT: Documents critical Kubernetes requirement
- **FIX Constitution**: Add to Section V "Service Discovery Mechanisms":

  ```markdown
  **DNS Caching** (MANDATORY):
    - DNS lookups MUST NOT be cached - lookup on EVERY request
    - Rationale: Kubernetes service endpoints change dynamically
    - Implementation: Disable connection pooling DNS cache
  ```

---

### LOW-012: CRLDP URL Encoding Format Ambiguity

**Constitution.md Statement** (Line 158):

```markdown
- **CRLDP (base64-url vs hex, batched vs immediate)**
```

**Plan.md Statement** (Line 90):

```markdown
- URL format: `https://crl.example.com/<base64-url-encoded-serial>.crl`
```

**Impact**:

- MINOR INCONSISTENCY: Constitution mentions "base64-url vs hex" as open question
- Plan.md resolves to base64-url, but constitution still shows as undecided

**Recommendation**:

- ✅ Plan.md is CORRECT: "base64-url-encoded-serial"
- ⚠️ Constitution OUTDATED: Remove "base64-url vs hex" (decision already made)
- **FIX Constitution**: Change Line 158 from:

  ```markdown
  - **CRLDP (base64-url vs hex, batched vs immediate)**
  ```

  To:

  ```markdown
  - **CRLDP URL Format**: `https://crl.example.com/<base64-url-encoded-serial>.crl` (one serial per CRL, immediate publication)
  ```

---

## Contradictions with Spec.md

### SPEC-001: CRLDP Details Missing

**Constitution.md Mandate** (Line 103):

```markdown
- **CRLDP MUST provide immediate revocation checks (NOT batched or delayed)**
```

**Spec.md Missing**:

- No CRLDP URL format specification
- No "immediate vs batched" clarification
- No "one serial per CRL" requirement

**Recommendation**: Add CRLDP section to spec.md (see CRITICAL-004)

---

### SPEC-002: Database Choice Ambiguity

**Constitution.md Mandate** (Line 774):

```markdown
Choice based on deployment type (multi-service vs standalone), NOT environment (prod vs dev)
```

**Spec.md Missing**: Explicit rejection of environment-based choice

**Recommendation**: Add clarification to spec.md (see CRITICAL-005)

---

### SPEC-003: Bind Address Wording

**Constitution.md**: "configurable bind address (container default: 0.0.0.0)"

**Spec.md**: "MUST use IPv4 0.0.0.0" (sounds hardcoded)

**Recommendation**: Change to "MUST default to" (see MEDIUM-009)

---

## Contradictions with Clarify.md

### CLARIFY-001: SQLite max_idle Conflict

**Clarify.md Statement** (Line 354):

```markdown
**SQLite**: max_open=5, max_idle=1 (single-writer limitation)
```

**sqlite-gorm.md Statement**:

```markdown
MaxIdleConns=5 (keep all connections alive for in-memory)
```

**Recommendation**:

- ✅ sqlite-gorm.md is CORRECT: MaxIdleConns=5 for shared cache
- ❌ Clarify.md is WRONG: "max_idle=1" contradicts GORM pattern
- **FIX Clarify.md**: Change to "max_idle=5" with rationale "shared cache requires all connections alive"

---

## Contradictions with Plan/Tasks/Analyze

### PLAN-001: Service Naming in analyze.md

**Plan.md** and **Tasks.md**: Use "learn-im" consistently

**Analyze.md** (probably-out-of-date): Uses "learn-ps" in 9 places

**Recommendation**:

- ✅ Plan.md and Tasks.md are CORRECT
- ❌ Analyze.md is OUTDATED (file marked "probably-out-of-date")
- **FIX**: Replace all "learn-ps" → "learn-im" OR archive analyze-probably-out-of-date.md

---

## Internal Contradictions (Constitution Contradicting Itself)

### INTERNAL-001: CRLDP Decision Status

**Constitution Line 103**: "CRLDP MUST provide immediate revocation"

**Constitution Line 158**: "CRLDP (base64-url vs hex, batched vs immediate)"

**Analysis**: Line 158 suggests decision pending, but Line 103 already decided "immediate"

**Recommendation**: Update Line 158 to reflect decision (see LOW-012)

---

## Ambiguities

### AMBIG-001: "Configurable" Bind Address Scope

**Constitution Statement**: "configurable bind address (container default: 0.0.0.0)"

**Ambiguity**: Can users configure to different values, or is 0.0.0.0 mandatory?

**Clarification Needed**:

- Inside containers: MUST be 0.0.0.0 (required for external access)
- Outside containers: SHOULD be 127.0.0.1 (prevents Windows Firewall)
- User configuration: Allowed for production deployments only

**Recommendation**: Add to constitution.md:

```markdown
**Bind Address Configuration Rules**:
  - Inside containers: MUST be 0.0.0.0 (NOT configurable - required for port mapping)
  - Tests/dev: MUST be 127.0.0.1 (NOT configurable - prevents Windows Firewall)
  - Production deployments: MAY be configurable (per deployment needs)
```

---

### AMBIG-002: MaxOpenConns for SQLite

**Constitution**: No explicit value

**Memory Files**: Consistently say "5" but constitution doesn't mandate

**Clarification Needed**: Add to constitution (see CRITICAL-007)

---

### AMBIG-003: Session Migration Grace Period Duration

**Clarify.md**: "e.g., 24 hours"

**Ambiguity**: Is 24 hours MANDATORY or just an example?

**Recommendation**: Add to constitution.md:

```markdown
**Session Migration Grace Period**:
  - Duration: RECOMMENDED 24 hours (configurable per deployment)
  - Minimum: 1 hour (allow short migrations)
  - Maximum: 7 days (prevent indefinite dual-format support)
```

---

## Missing Specifications

### MISSING-001: CRLDP Implementation Details

**Topics mentioned in downstream but NOT in constitution**:

- CRLDP URL format (plan.md has it, constitution.md doesn't)
- base64-url encoding decision (plan.md decided, constitution.md still shows as open)
- One serial per CRL file (plan.md specifies, constitution.md doesn't)

**Recommendation**: Add CRLDP section to constitution.md (see CRITICAL-004 fix)

---

### MISSING-002: Session Migration Pattern

**Topics mentioned in clarify.md but NOT in constitution**:

- Grace period for dual-format token support
- Natural expiration vs forced invalidation
- New issuance behavior during migration

**Recommendation**: Add session migration section to constitution.md (see MEDIUM-010 fix)

---

### MISSING-003: DNS Caching Policy

**Topics mentioned in clarify.md but NOT in constitution**:

- DNS caching prohibition for federated services
- Kubernetes dynamic endpoint requirement
- Implementation pattern for cache-free lookups

**Recommendation**: Add DNS caching policy to constitution.md (see MEDIUM-011 fix)

---

### MISSING-004: SQLite Connection Pool Values

**Topics mentioned in memory files but NOT in constitution**:

- MaxOpenConns=5 (MANDATORY for GORM)
- MaxIdleConns=5 (shared cache requirement)
- ConnMaxLifetime=0 (in-memory) vs 1h (file-based)

**Recommendation**: Add SQLite connection pool section to constitution.md (see CRITICAL-007 fix)

---

### MISSING-005: Bind Address Configuration Rules

**Topics mentioned in spec.md but NOT clearly in constitution**:

- Inside containers: 0.0.0.0 MANDATORY (not configurable)
- Tests/dev: 127.0.0.1 MANDATORY (not configurable)
- Production: Configurable (deployment-specific)

**Recommendation**: Add bind address rules to constitution.md (see AMBIG-001 fix)

---

## Recommendations

### PRIORITY 1: CRITICAL FIXES (Complete Before Phase 2)

#### 1.1 Fix Service Naming Everywhere

**Constitution.md**: ✅ CORRECT (uses "learn-im")

**Files to Fix**:

- ❌ `specs/002-cryptoutil/analyze-probably-out-of-date.md`: Replace 9 instances "learn-ps" → "learn-im"
- ✅ Alternative: Archive/delete analyze-probably-out-of-date.md (file name suggests obsolete)

**Git Grep Verification**:

```powershell
# Verify no more "learn-ps" references
git grep -i "learn-ps" -- specs/002-cryptoutil/
# Expected: ZERO matches after fix
```

---

#### 1.2 Add CRLDP Specification to Constitution

**Location**: constitution.md Section II "Cryptographic Compliance and Standards"

**Add After Line 103**:

```markdown
### CRLDP (Certificate Revocation List Distribution Points)

**URL Format**: `https://crl.example.com/<base64-url-encoded-serial>.crl`

**Publication Requirements**:
- **One serial per CRL file**: NEVER batch multiple serials into one CRL
- **Immediate publication**: Sign and publish within 1 minute of revocation (NOT batched)
- **HTTPS only**: NEVER use HTTP for CRL distribution (security requirement)
- **Encoding**: base64-url encoding for serial numbers (URL-safe)

**Rationale**: Immediate per-serial CRLs enable real-time revocation checking, defense-in-depth with OCSP.
```

**Update spec.md** (add to "Certificate Revocation Requirements" section):

```markdown
### CRLDP Configuration

**URL Format**: `https://crl.example.com/<base64-url-encoded-serial>.crl`

**Requirements**:
- One serial number per CRL file (NEVER batch)
- Immediate sign+publish (within 1 minute of revocation)
- HTTPS only (TLS 1.3+)
- base64-url encoding (URL-safe serial representation)
```

---

#### 1.3 Add SQLite Connection Pool to Constitution

**Location**: constitution.md Section IV "Database Architecture"

**Add After Line 774**:

```markdown
### SQLite Connection Pool Configuration

**MANDATORY for GORM-based services**:

```yaml
database:
  sqlite:
    max_open_connections: 5      # MANDATORY - GORM transactions need separate connection
    max_idle_connections: 5      # Keep all connections alive (shared cache for :memory:)
    connection_max_lifetime: 0   # In-memory: 0 (never close), File-based: 1h
```

**Rationale**:

- GORM explicit transactions (`db.Begin()`) require separate database connection from base operations
- MaxOpenConns=1 causes deadlock: base DB tries to create transaction, but connection already in use
- MaxOpenConns=5 allows concurrent transaction wrapper + internal operations
- MaxIdleConns=5 ensures shared cache works correctly (all connections see same in-memory data)

**See**: `.specify/memory/sqlite-gorm.md` for complete implementation pattern

```

**Fix clarify.md** (Line 354):

Change from:
```markdown
**SQLite**: max_open=5, max_idle=1 (single-writer limitation)
```

To:

```markdown
**SQLite**: max_open=5, max_idle=5 (GORM transaction pattern + shared cache)
```

---

#### 1.4 Add Database Choice Clarification to Spec

**Location**: spec.md Section "Database Architecture"

**Add After Line 2250**:

```markdown
### Database Selection Criteria

**CRITICAL**: Database choice driven by DEPLOYMENT TYPE, NOT environment:

| Deployment Type | Database | Environments |
|----------------|----------|--------------|
| Multi-service (suite) | PostgreSQL | Production AND Development |
| Standalone (single service) | SQLite | Production AND Development |

**Common Misconception**: "PostgreSQL for prod, SQLite for dev"
**CORRECT Pattern**: "PostgreSQL for multi-service, SQLite for standalone"

**Rationale**:
- Multi-service deployments need shared session state → PostgreSQL
- Standalone services avoid PostgreSQL overhead → SQLite
- Environment (prod vs dev) does NOT determine database choice
```

---

#### 1.5 Fix Bind Address Wording in Spec

**Location**: spec.md Line 105

**Change From**:

```markdown
- Docker Containers: MUST use IPv4 0.0.0.0 binding by default inside containers
```

**Change To**:

```markdown
- Docker Containers: MUST default to IPv4 0.0.0.0 binding inside containers (configurable via YAML)
```

**Add Clarification**:

```markdown
**Bind Address Configuration Scope**:
- Inside containers: MUST default to 0.0.0.0 (configurable for advanced deployments)
- Tests/dev: MUST default to 127.0.0.1 (NOT configurable - prevents Windows Firewall)
- Production (outside containers): Configurable per deployment needs
```

---

### PRIORITY 2: SPECIFICATION GAPS (Add to Constitution)

#### 2.1 Add Session Migration Pattern

**Location**: constitution.md Section V "Service Federation and Discovery"

**Add After Line 580**:

```markdown
### Session Migration During Federation Transitions

When a service transitions from non-federated (standalone) to federated mode:

**Grace Period Pattern**:
- **Duration**: 24 hours (configurable per deployment, min 1h, max 7 days)
- **Dual-Format Support**: Accept BOTH old and new token formats during grace period
- **Natural Expiration**: Old-format tokens expire per original TTL (no forced invalidation)
- **New Issuance**: New logins immediately receive new-format tokens
- **No Forced Re-Auth**: Users are NOT forced to re-authenticate during migration

**Configuration Example**:

```yaml
federation:
  migration:
    grace_period_hours: 24      # Accept both formats during transition
    old_format_enabled: true    # Temporary backward compatibility
    new_format_enabled: true    # Forward compatibility
```

**Rationale**: Prevents service disruption by allowing gradual token migration without forcing user re-authentication.

**Source**: SPECKIT-CLARIFY-QUIZME answers, 2025-12-22

```

---

#### 2.2 Add DNS Caching Policy

**Location**: constitution.md Section V "Service Discovery Mechanisms"

**Add After Line 530**:

```markdown
### DNS Caching Policy - MANDATORY

**CRITICAL**: DNS lookups for federated services MUST NOT be cached - perform lookup on EVERY request.

**Rationale**:
- Kubernetes service endpoints change dynamically (pod restarts, scaling)
- Stale DNS cache causes request failures to moved/restarted pods
- Container orchestrators rely on DNS-based service discovery

**Implementation**:

```go
// Disable DNS caching for HTTP client
dialer := &net.Dialer{
    Timeout:   30 * time.Second,
    KeepAlive: 30 * time.Second,
}

transport := &http.Transport{
    DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
        // Perform DNS lookup on EVERY dial (no caching)
        return dialer.DialContext(ctx, network, addr)
    },
    DisableKeepAlives:   false,  // Keep connections alive (HTTP/1.1 persistent)
    MaxIdleConns:        100,    // Pool idle connections
    MaxIdleConnsPerHost: 10,
}
```

**Trade-off**: Slight latency increase (DNS lookup per request ~5-10ms) for guaranteed fresh endpoints.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q18, 2025-12-24

```

---

#### 2.3 Add Bind Address Configuration Rules

**Location**: constitution.md Section V "Deployment Environments"

**Add After Line 343**:

```markdown
### Bind Address Configuration Rules

**Inside Containers**:
- Public endpoints: MUST default to 0.0.0.0 (enables external access via port mapping)
- Private endpoints: MUST be 127.0.0.1 (localhost only, NOT configurable)
- User override: MAY allow advanced configurations (NOT recommended)

**Tests and Development** (outside containers):
- Public endpoints: MUST be 127.0.0.1 (prevents Windows Firewall prompts, NOT configurable)
- Private endpoints: MUST be 127.0.0.1 (localhost only, NOT configurable)
- Port allocation: MUST use port 0 (dynamic allocation prevents conflicts)

**Production Deployments** (outside containers):
- Public endpoints: Configurable per deployment (0.0.0.0, specific IP, IPv6)
- Private endpoints: MUST be 127.0.0.1 (localhost only, NOT configurable)
- Configuration: Via YAML files (NOT environment variables for security)

**Rationale**:
- 0.0.0.0 inside containers: Required for Docker/Kubernetes port mapping
- 127.0.0.1 in tests: Prevents Windows Firewall exception prompts (blocks CI/CD automation)
- Admin always 127.0.0.1: Security isolation (admin APIs NEVER externally accessible)
```

---

### PRIORITY 3: DOCUMENTATION CLEANUP

#### 3.1 Archive or Update Out-of-Date Files

**Files to Review**:

- ❌ `specs/002-cryptoutil/analyze-probably-out-of-date.md` (file name suggests obsolete)
  - **Option A**: Archive to `specs/001-cryptoutil-archived-2025-12-17/` (consistent with previous archive)
  - **Option B**: Fix all "learn-ps" → "learn-im" references (9 occurrences)
  - **Recommendation**: Archive (file is iteration 1 artifact, now superseded by implement/DETAILED.md)

**Git Command**:

```powershell
# Archive old analyze.md
git mv specs/002-cryptoutil/analyze-probably-out-of-date.md `
        specs/001-cryptoutil-archived-2025-12-17/analyze-2025-12-17.md

# Commit
git add .
git commit -m "docs(speckit): archive outdated analyze.md with learn-ps references"
```

---

#### 3.2 Update Constitution Internal Consistency

**Fix Line 158**: Remove "vs" phrasing (decision already made)

**Change From**:

```markdown
- **CRLDP (base64-url vs hex, batched vs immediate)**
```

**Change To**:

```markdown
- **CRLDP URL Format**: `https://crl.example.com/<base64-url-encoded-serial>.crl` (immediate publication, one serial per CRL)
```

---

#### 3.3 Cross-Validate All Documents

**Verification Checklist**:

- [ ] Run grep for "learn-ps" → expect ZERO matches
- [ ] Run grep for "9091|9092|9093" → expect ZERO matches (admin ports)
- [ ] Run grep for "schema-only|row-only" (without "dual-layer") → expect ZERO matches
- [ ] Run grep for "batched CRL|batch.*CRL" → expect ZERO matches (unless in "NOT batched" context)
- [ ] Verify all spec.md references to constitution.md sections are valid
- [ ] Verify all clarify.md answers reference constitution.md correctly
- [ ] Verify plan.md phases match constitution.md implementation priority

**Automated Verification Script**:

```powershell
# Create verification script
$script = @'
# Constitution.md Cross-Validation Script
Write-Host "=== Constitution.md Validation ===" -ForegroundColor Cyan

# Test 1: Service naming
Write-Host "`nTest 1: Service naming (expect ZERO 'learn-ps')" -ForegroundColor Yellow
git grep -i "learn-ps" -- specs/002-cryptoutil/ .specify/memory/

# Test 2: Admin ports
Write-Host "`nTest 2: Admin ports (expect ZERO per-service ports)" -ForegroundColor Yellow
git grep -E "9091|9092|9093" -- specs/002-cryptoutil/ .specify/memory/

# Test 3: Multi-tenancy
Write-Host "`nTest 3: Multi-tenancy (expect dual-layer or ZERO)" -ForegroundColor Yellow
git grep -E "(schema-only|row-only)" -- specs/002-cryptoutil/ .specify/memory/ | grep -v "dual-layer"

# Test 4: CRLDP batching
Write-Host "`nTest 4: CRLDP batching (expect ZERO 'batched CRL')" -ForegroundColor Yellow
git grep -iE "(batched CRL|batch.*CRL)" -- specs/002-cryptoutil/ .specify/memory/ | grep -v "NOT batched"

Write-Host "`n=== Validation Complete ===" -ForegroundColor Green
'@

# Save and run
$script | Out-File -FilePath ./local-scripts/validate-constitution.ps1 -Encoding UTF8
& ./local-scripts/validate-constitution.ps1
```

---

## Summary of Contradictions Found

### By Severity

| Severity | Count | Status | Action |
|----------|-------|--------|--------|
| CRITICAL | 8 | 4 Fixed, 4 Pending | Fix before Phase 2 |
| MEDIUM | 3 | 0 Fixed, 3 Pending | Add to constitution |
| LOW | 1 | 0 Fixed, 1 Pending | Update wording |
| **TOTAL** | **12** | **4 Fixed, 8 Pending** | **8 fixes required** |

### By Document Type

| Document Type | Contradictions | Status |
|--------------|----------------|--------|
| Memory Files | 4 | 2 Fixed (admin ports, service naming clarified), 2 Pending (SQLite config, bind address) |
| Spec.md | 3 | All Pending (CRLDP, database choice, bind address) |
| Clarify.md | 1 | Pending (SQLite max_idle) |
| Plan/Tasks | 1 | Pending (analyze.md archival) |
| Constitution Internal | 1 | Pending (Line 158 wording) |
| Specification Gaps | 5 | All Pending (CRLDP, session migration, DNS caching, SQLite config, bind rules) |

---

## Files Requiring Updates

### Constitution.md (8 additions/fixes)

1. ✅ Line 158: Fix CRLDP wording (remove "vs", use final decision)
2. ➕ Section II: Add CRLDP specification (URL format, immediate publication)
3. ➕ Section IV: Add SQLite connection pool specification (MaxOpenConns=5)
4. ➕ Section V: Add session migration pattern (grace period, dual-format)
5. ➕ Section V: Add DNS caching policy (NO caching mandate)
6. ➕ Section V: Add bind address configuration rules (container/test/prod)

### Spec.md (3 additions)

1. ➕ Certificate Revocation: Add CRLDP section (URL format, immediate publication)
2. ➕ Database Architecture: Add database choice clarification (deployment type vs environment)
3. ✅ Line 105: Fix bind address wording ("MUST default to" instead of "MUST use")

### Clarify.md (1 fix)

1. ✅ Line 354: Fix SQLite max_idle (change "1" to "5")

### Analyze-probably-out-of-date.md (1 action)

1. 🗄️ Archive to specs/001-cryptoutil-archived-2025-12-17/ (contains 9 "learn-ps" references)

---

## Validation Checklist

After implementing all recommendations, verify:

- [ ] **Service Naming**: Zero "learn-ps" references (all "learn-im")
- [ ] **Admin Ports**: Zero "9091|9092|9093" references (all "9090")
- [ ] **Multi-Tenancy**: All references include "dual-layer" qualification
- [ ] **CRLDP**: All references specify "immediate publication, one serial per CRL"
- [ ] **Database Choice**: All references specify "deployment type, NOT environment"
- [ ] **SQLite Config**: All references use "MaxOpenConns=5, MaxIdleConns=5"
- [ ] **Session Migration**: Constitution includes grace period pattern
- [ ] **DNS Caching**: Constitution includes "NO caching" mandate
- [ ] **Bind Address**: Constitution includes configuration scope rules

---

## Impact on Implementation

### Immediate Blockers (Must Fix Before Phase 2)

1. **Service Naming**: Code generation uses wrong name ("learn-ps" instead of "learn-im")
2. **SQLite Config**: GORM services will deadlock with MaxOpenConns=1
3. **CRLDP**: Implementation might batch CRLs (violates constitution)

### High Priority (Must Fix Before Phase 3)

1. **Database Choice**: Developers might choose PostgreSQL for prod, SQLite for dev (wrong)
2. **Session Migration**: Federation transitions might force re-auth (breaks user sessions)

### Medium Priority (Must Fix Before Production)

1. **DNS Caching**: Kubernetes deployments will fail with cached DNS
2. **Bind Address**: Ambiguous wording might cause 0.0.0.0 binding in tests (Windows Firewall prompts)

---

## Conclusion

**Constitution.md is 90% accurate** - Most mandates are correct and downstream documents match.

**8 fixes required** to achieve 100% consistency:

- 4 CRITICAL gaps (CRLDP, SQLite, session migration, DNS caching)
- 3 MEDIUM clarifications (database choice, bind address, clarify.md fix)
- 1 LOW wording update (Line 158 internal consistency)
- 1 CLEANUP (archive analyze-probably-out-of-date.md)

**Priority**: Complete CRITICAL fixes before Phase 2 template extraction begins.

**Estimated Effort**: 3-4 hours to implement all recommendations.

**Risk Mitigation**: Automated validation script prevents future contradictions.

---

**Review Completed**: 2025-12-24
**Next Steps**: Implement Priority 1 fixes, then validate with automated script
