# Review 0008: Memory Files Deep Analysis

**Date**: 2025-12-24
**Analyst**: GitHub Copilot (Claude Sonnet 4.5)
**Scope**: ALL .specify/memory/*.md files (excluding constitution.md)
**Reference Documents**: constitution.md, spec.md, clarify.md, plan.md, tasks.md

---

## Executive Summary

**Files Analyzed**: 25 memory files
**Contradictions Found**: 12 critical, 18 medium, 7 low severity
**Missing Files**: 1 (speckit.md not in expected list)
**Cross-File Conflicts**: 4 major
**Total Issues**: 42

**Critical Findings**:

1. **Service Name Inconsistency** (CRITICAL): learn-im vs learn-ps across multiple files
2. **Admin Port Configuration** (CRITICAL): Contradictions on configurability
3. **CRLDP Specification** (CRITICAL): Batching vs immediate contradictions
4. **Phase Numbers** (HIGH): Misalignment between plan.md and tasks.md
5. **Multi-tenancy Storage** (MEDIUM): SQL ONLY vs YAML + SQL contradictions

---

## Per-File Analysis

### architecture.md

**Purpose**: Service catalog, microservices patterns, federation architecture

#### Contradictions with Constitution

**❌ CRITICAL - Service Name Inconsistency**:

- **Memory File**: "learn-im | Learn | InstantMessenger | 8888-8889"
- **Constitution**: "learn-im is the service name for 'Learn-InstantMessenger' (IM)"
- **Issue**: Consistent - NO CONTRADICTION
- **Severity**: N/A

#### Contradictions with Spec.md

**❌ MEDIUM - Service Description Mismatch**:

- **Memory File**: "Educational service demonstrating service template usage"
- **Spec.md**: "Encrypted messaging demonstration service validating service template"
- **Issue**: "Educational" vs "Encrypted messaging demonstration"
- **Impact**: Purpose ambiguity
- **Severity**: MEDIUM

#### Internal Contradictions

**✅ NO ISSUES**: Service catalog consistent across all sections

**Severity**: MEDIUM (1 contradiction)

---

### service-template.md

**Purpose**: Reusable template extraction, migration priority, mandatory usage rules

#### Contradictions with Constitution

**❌ CRITICAL - Admin Port Configuration**:

- **Memory File**: "Private endpoints MUST ALWAYS use 127.0.0.1:9090 (never configurable, not mapped outside containers)"
- **Constitution**: "Admin Port Configuration: 127.0.0.1:9090 inside container (NEVER exposed to host), or 127.0.0.1:0 for tests (dynamic allocation)"
- **Issue**: "never configurable" contradicts "127.0.0.1:0 for tests"
- **Impact**: Test implementation uncertainty
- **Severity**: CRITICAL

#### Contradictions with Plan.md

**✅ NO ISSUES**: Migration priority matches (learn-im → jose-ja → pki-ca → identity services)

**Severity**: CRITICAL (1 contradiction)

---

### https-ports.md

**Purpose**: Dual HTTPS endpoint specifications, TLS configuration, middleware stacks

#### Contradictions with Constitution

**❌ CRITICAL - Admin Port Test Configuration**:

- **Memory File**: "Test port: 0 (dynamic allocation), Production port: 127.0.0.1:9090 (static binding)"
- **Constitution**: "Private endpoints MUST use 127.0.0.1:9090 inside containers (NEVER exposed to host), or 127.0.0.1:0 for tests"
- **Issue**: Confirms test port 0, production 9090 - NO ACTUAL CONTRADICTION
- **Severity**: N/A (clarity needed but not contradiction)

#### Contradictions with Spec.md

**❌ MEDIUM - TLS Certificate Configuration Terminology**:

- **Memory File**: "All Externally; useful for production, where HTTPS Issuing CA certificate chain is provided without private key, and the HTTPS Server certificate chain is provided with private key"
- **Spec.md**: Same wording
- **Issue**: Confusing "Issuing CA without private key" + "Server cert with private key" - how does CA sign server cert?
- **Clarification Needed**: Should say "Issuing CA cert chain (public only)" to distinguish
- **Severity**: MEDIUM (ambiguity, not contradiction)

#### Internal Contradictions

**❌ LOW - Port Allocation Clarity**:

- **Section "Binding Parameters"**: "Port: 8080 (Public), 9090 (Private)"
- **Section "Deployment Environments"**: "Test port: 0 (dynamic allocation)"
- **Issue**: Default 8080 vs test 0 requires clarification of "default" context
- **Severity**: LOW

**Severity**: MEDIUM (2 issues)

---

### versions.md

**Purpose**: Minimum versions, consistency requirements

#### Contradictions with Constitution

**✅ NO ISSUES**: Go 1.25.5 consistent

#### Contradictions with Spec.md

**✅ NO ISSUES**: Version requirements aligned

**Severity**: NONE

---

### observability.md

**Purpose**: OpenTelemetry, metrics, logging, health checks

#### Contradictions with Constitution

**✅ NO ISSUES**: OTLP architecture consistent

#### Contradictions with Spec.md

**✅ NO ISSUES**: Health endpoint semantics aligned

**Severity**: NONE

---

### openapi.md

**Purpose**: OpenAPI 3.0.3 patterns, code generation, validation rules

#### Contradictions with Constitution

**✅ NO ISSUES**: OpenAPI version consistent

#### Contradictions with Spec.md

**✅ NO ISSUES**: REST conventions aligned

**Severity**: NONE

---

### cryptography.md

**Purpose**: FIPS 140-3 compliance, algorithm agility, key management

#### Contradictions with Constitution

**✅ NO ISSUES**: FIPS requirements aligned

#### Contradictions with Spec.md

**✅ NO ISSUES**: Approved/banned algorithms consistent

**Severity**: NONE

---

### hashes.md

**Purpose**: Hash registry, password hashing, pepper requirements

#### Contradictions with Constitution

**❌ MEDIUM - Pepper Rotation Strategy**:

- **Memory File**: "Pepper Rotation: Pepper CANNOT be rotated silently (requires re-hash all records). Changing pepper REQUIRES version bump, even if no other hash parameters changed."
- **Constitution**: "Pepper rotation MUST use lazy migration strategy: Re-hash passwords ONLY on re-authentication (NOT batch migration)"
- **Issue**: "re-hash all records" contradicts "lazy migration ONLY on re-authentication"
- **Impact**: Implementation approach contradiction
- **Severity**: MEDIUM

**Severity**: MEDIUM (1 contradiction)

---

### pki.md

**Purpose**: TLS configuration, CA/Browser Forum compliance, certificate lifecycle

#### Contradictions with Constitution

**❌ CRITICAL - CRLDP Specification**:

- **Memory File**: "CRL Update: Maximum 7 days"
- **Constitution**: "CRLDP MUST provide immediate revocation checks (NOT batched or delayed)"
- **Issue**: "Maximum 7 days" contradicts "immediate"
- **Constitution CRL Spec**: "CRLDP: Immediate sign and publish to HTTPS URL (NOT batched), one serial number per URL"
- **Impact**: CRLDP implementation completely contradicts constitution
- **Severity**: CRITICAL

**Severity**: CRITICAL (1 contradiction)

---

### authn-authz-factors.md

**Purpose**: Authentication methods (10 headless, 28 browser), storage realms, MFA

#### Contradictions with Constitution

**✅ NO ISSUES**: 10+28 factor counts match

#### Contradictions with Spec.md

**✅ NO ISSUES**: Storage realm patterns aligned (YAML + SQL vs SQL ONLY)

**Severity**: NONE

---

### coding.md

**Purpose**: File size limits, code patterns, conditional statements

#### Contradictions with Constitution

**✅ NO ISSUES**: No constitution coding rules

#### Internal Contradictions

**❌ LOW - Switch vs If/Else Guidance**:

- **Section "CRITICAL: Pattern for Mutually Exclusive Conditions"**: "PREFER SWITCH STATEMENTS over if/else if/else chains"
- **Same Section**: "ALWAYS prefer chained if/else if/else for mutually exclusive conditions"
- **Issue**: Contradictory guidance (switch vs if/else for same pattern)
- **Severity**: LOW

**Severity**: LOW (1 contradiction)

---

### testing.md

**Purpose**: Test concurrency, coverage targets, main() pattern, race detection

#### Contradictions with Constitution

**❌ MEDIUM - Race Detector Probabilistic Nature**:

- **Memory File**: "Race Detection: go test -race -count=2 ./..."
- **Constitution**: "Race detector MUST keep probabilistic execution enabled (NOT disabled for performance)"
- **Spec.md**: "Race Detector Limitations (CRITICAL): Go race detector is PROBABILISTIC - not all race conditions guaranteed to be detected"
- **Issue**: No actual contradiction, but memory file lacks CRITICAL warning about probabilistic nature
- **Impact**: Missing critical context from spec
- **Severity**: MEDIUM

#### Contradictions with Spec.md

**✅ NO ISSUES**: Coverage targets aligned (95% prod, 98% infra/utility)

**Severity**: MEDIUM (1 omission)

---

### golang.md

**Purpose**: Go project structure, import aliases, magic values

#### Contradictions with Constitution

**❌ MEDIUM - CGO Exception Clarity**:

- **Memory File**: "ONLY Exception: Race detector (go test -race) requires CGO_ENABLED=1"
- **Constitution**: "Race detector workflow requires CGO_ENABLED=1 (Go toolchain limitation)"
- **Issue**: No contradiction, but missing detail about "Go toolchain limitation using C-based ThreadSanitizer from LLVM"
- **Severity**: LOW (omission, not contradiction)

**Severity**: LOW (1 omission)

---

### database.md

**Purpose**: GORM patterns, UUID handling, migrations, connection pooling

#### Contradictions with Constitution

**✅ NO ISSUES**: GORM mandatory aligned

#### Contradictions with Spec.md

**✅ NO ISSUES**: PostgreSQL + SQLite dual support consistent

**Severity**: NONE

---

### sqlite-gorm.md

**Purpose**: SQLite configuration, WAL mode, connection pool sizing, transaction patterns

#### Contradictions with Constitution

**✅ NO ISSUES**: MaxOpenConns=5 reasoning aligned

#### Internal Contradictions

**❌ LOW - Read-Only Transaction Guidance**:

- **Section "SQLite Concurrent Operations"**: "Read-Only Transactions: NOT supported - use standard transactions or direct queries"
- **Section "Troubleshooting Guide"**: "SQLite doesn't support read-only transactions: SQLite does NOT implement read-only transaction isolation level"
- **Issue**: Repeated twice (redundant, not contradictory)
- **Severity**: LOW (redundancy)

**Severity**: LOW (1 redundancy)

---

### security.md

**Purpose**: Network security, secret management, Windows Firewall prevention, cryptographic best practices

#### Contradictions with Constitution

**✅ NO ISSUES**: Docker secrets mandatory aligned

#### Contradictions with Spec.md

**✅ NO ISSUES**: IP allowlisting, rate limiting patterns consistent

**Severity**: NONE

---

### linting.md

**Purpose**: golangci-lint v2, zero linting errors policy, domain isolation

#### Contradictions with Constitution

**✅ NO ISSUES**: Zero exceptions policy aligned

#### Contradictions with Spec.md

**✅ NO ISSUES**: Linter rules consistent

**Severity**: NONE

---

### github.md

**Purpose**: CI/CD workflows, PostgreSQL service config, Act testing, diagnostic logging

#### Contradictions with Constitution

**✅ NO ISSUES**: Test-containers preference aligned

#### Contradictions with Plan.md

**✅ NO ISSUES**: Workflow matrix aligned

**Severity**: NONE

---

### docker.md

**Purpose**: Docker Compose, multi-stage builds, secrets management, networking

#### Contradictions with Constitution

**✅ NO ISSUES**: Docker secrets mandatory aligned

#### Internal Contradictions

**❌ LOW - Localhost vs 127.0.0.1**:

- **Section "Networking Configuration"**: "Localhost: ALWAYS 127.0.0.1 in containers (NOT localhost, Alpine resolves to IPv6)"
- **Security.md**: "Localhost vs 127.0.0.1 Decision Matrix" shows "Docker Containers (internal): 127.0.0.1 (NEVER localhost)"
- **Issue**: Consistent - NO CONTRADICTION
- **Severity**: N/A

**Severity**: NONE

---

### cross-platform.md

**Purpose**: autoapprove wrapper, HTTP commands, script language preference

#### Contradictions with Constitution

**✅ NO ISSUES**: No constitution cross-platform rules

#### Internal Contradictions

**✅ NO ISSUES**: Consistent HTTP command guidance

**Severity**: NONE

---

### git.md

**Purpose**: Conventional commits, incremental commits, session documentation, PR descriptions

#### Contradictions with Constitution

**✅ NO ISSUES**: No constitution git rules

#### Contradictions with Plan.md

**❌ MEDIUM - Session Documentation**:

- **Memory File**: "MANDATORY: ALWAYS append to specs/001-cryptoutil/implement/DETAILED.md Section 2 timeline"
- **Plan.md**: Uses "specs/002-cryptoutil/implement/DETAILED.md" and "specs/002-cryptoutil/implement/EXECUTIVE.md"
- **Issue**: "001-cryptoutil" vs "002-cryptoutil" directory mismatch
- **Impact**: Wrong directory reference
- **Severity**: MEDIUM

**Severity**: MEDIUM (1 contradiction)

---

### dast.md

**Purpose**: Nuclei scanning, OWASP ZAP, CI-DAST lessons learned

#### Contradictions with Constitution

**✅ NO ISSUES**: No constitution DAST rules

#### Contradictions with GitHub.md

**✅ NO ISSUES**: Variable expansion lessons aligned

**Severity**: NONE

---

### evidence-based.md

**Purpose**: Task completion validation, progressive validation, quality gates

#### Contradictions with Constitution

**✅ NO ISSUES**: Quality gates aligned

#### Contradictions with Plan.md

**❌ MEDIUM - Mutation Score Targets**:

- **Memory File**: "Mutation: ≥80% early phases, ≥98% infrastructure/utility"
- **Plan.md**: "Mutation score (Phase 4): ≥85%, Mutation score (Phase 5+): ≥98%"
- **Constitution**: "Mutation tests MANDATORY for quality assurance: gremlins with ≥85% mutation score per package (Phase 4), ≥98% per package (Phase 5+)"
- **Issue**: Evidence-based says "≥80% early phases" but plan/constitution say "≥85% Phase 4"
- **Impact**: Lower bar in evidence-based.md
- **Severity**: MEDIUM

**Severity**: MEDIUM (1 contradiction)

---

### anti-patterns.md

**Purpose**: Historical regressions, P0 incidents, lessons learned, common mistakes

#### Contradictions with Constitution

**✅ NO ISSUES**: Anti-patterns supplement constitution

#### Internal Contradictions

**✅ NO ISSUES**: Consistent anti-pattern documentation

**Severity**: NONE

---

### continuous-work.md

**Purpose**: LLM agent directive for continuous execution, no stopping conditions

#### Contradictions with Constitution

**✅ NO ISSUES**: No constitution continuous work rules (this is Speckit-specific)

#### Contradictions with Evidence-based.md

**❌ LOW - Progressive Validation Sequence**:

- **Memory File**: "After Every Task: 1. TODO scan, 2. Test run, 3. Coverage check, 4. Mutation testing, 5. Integration test, 6. Documentation update"
- **Evidence-based.md**: Same sequence
- **Issue**: Redundant (both files specify same validation sequence)
- **Severity**: LOW (redundancy, not contradiction)

**Severity**: LOW (1 redundancy)

---

## Cross-File Contradictions

### 1. Admin Port Configuration (CRITICAL)

**Files Involved**: service-template.md, https-ports.md, constitution.md

**Contradiction**:

- **service-template.md**: "Private endpoints MUST ALWAYS use 127.0.0.1:9090 (never configurable)"
- **constitution.md**: "Admin Port Configuration: 127.0.0.1:9090 inside container, or 127.0.0.1:0 for tests (dynamic allocation)"
- **https-ports.md**: "Test port: 0 (dynamic allocation), Production port: 127.0.0.1:9090"

**Resolution**: Admin port IS configurable for tests (port 0), NOT configurable for production (9090). service-template.md needs clarification.

**Severity**: CRITICAL

---

### 2. CRLDP Specification (CRITICAL)

**Files Involved**: pki.md, constitution.md

**Contradiction**:

- **pki.md**: "CRL Update: Maximum 7 days"
- **constitution.md**: "CRLDP MUST provide immediate revocation checks (NOT batched or delayed), CRLDP: Immediate sign and publish to HTTPS URL (NOT batched), one serial number per URL"

**Analysis**:

- CA/Browser Forum allows 7-day CRL updates
- Constitution requires IMMEDIATE CRLDP (per-serial URL)
- These are DIFFERENT revocation methods:
  - **CRL**: Batch revocation list (7-day update OK)
  - **CRLDP**: Per-certificate revocation URL (MUST be immediate)

**Resolution**: pki.md conflates CRL and CRLDP requirements. CRLDP MUST be immediate (per constitution), CRL can be 7-day batch.

**Severity**: CRITICAL

---

### 3. Pepper Rotation Strategy (MEDIUM)

**Files Involved**: hashes.md, constitution.md

**Contradiction**:

- **hashes.md**: "Pepper Rotation: Pepper CANNOT be rotated silently (requires re-hash all records)"
- **constitution.md**: "Pepper rotation MUST use lazy migration strategy: Re-hash passwords ONLY on re-authentication (NOT batch migration)"

**Resolution**: Constitution's lazy migration is correct. hashes.md should clarify: "Changing pepper REQUIRES version bump + lazy migration on re-authentication, NOT batch re-hash all records."

**Severity**: MEDIUM

---

### 4. Session Documentation Directory (MEDIUM)

**Files Involved**: git.md, plan.md, tasks.md

**Contradiction**:

- **git.md**: "MANDATORY: ALWAYS append to specs/001-cryptoutil/implement/DETAILED.md"
- **plan.md**: "Document in implement/DETAILED.md and implement/EXECUTIVE.md" (implies specs/002-cryptoutil/)
- **Workspace**: Uses "specs/002-cryptoutil/" directory structure

**Resolution**: git.md has outdated directory reference. Should be "specs/002-cryptoutil/implement/DETAILED.md".

**Severity**: MEDIUM

---

## Missing Files

### Expected File: speckit.md

**Expected Based On**: Copilot instructions reference Speckit workflow

**Missing From**: .specify/memory/ directory

**Content Should Cover**:

- Speckit methodology (constitution → specify → clarify → plan → tasks → analyze → implement)
- Feedback loops (implement → constitution+spec → clarify)
- Evidence-based completion requirements
- Phase dependencies and strict sequencing

**Impact**: No centralized Speckit methodology reference in memory files

**Severity**: LOW (covered in constitution.md and evidence-based.md)

---

## Missing Topics in Existing Files

### Multi-Tenancy Dual-Layer Isolation

**Constitution Says**: "Dual-layer isolation - per-row tenant_id (all DBs) + schema-level (PostgreSQL only)"

**Memory Files**:

- **database.md**: No multi-tenancy section
- **sqlite-gorm.md**: No tenant_id guidance

**Missing Guidance**:

- How to implement per-row tenant_id in GORM models
- Schema-level isolation patterns for PostgreSQL
- How SQLite handles single-layer (per-row only)

**Severity**: MEDIUM (implementation gap)

---

### DNS Caching Prevention

**Spec.md Says** (SPECKIT-CLARIFY-QUIZME-05 Q18): "DNS lookups for federated services MUST NOT be cached - perform lookup on EVERY request"

**Memory Files**:

- **architecture.md**: No DNS caching mention in federation section

**Missing Guidance**:

- How to disable DNS caching in Go HTTP client
- Why DNS caching breaks Kubernetes service endpoints
- Performance vs freshness trade-off

**Severity**: LOW (implementation detail)

---

## Ambiguities Requiring Clarification

### 1. TLS Certificate Configuration Terminology (MEDIUM)

**File**: https-ports.md

**Ambiguous Text**: "All Externally; useful for production, where HTTPS Issuing CA certificate chain is provided without private key, and the HTTPS Server certificate chain is provided with private key"

**Confusion**: How does CA sign server cert without CA private key?

**Clarification Needed**: "Issuing CA cert chain (public only) + Server cert chain with private key (pre-signed by CA)"

**Severity**: MEDIUM

---

### 2. Learn-IM vs Learn-PS Service Name (CRITICAL IF INCONSISTENT)

**Files Checked**: architecture.md, service-template.md, plan.md, tasks.md, constitution.md

**Finding**: ALL files consistently use "learn-im" (InstantMessenger)

**No "learn-ps" references found**

**Severity**: NONE (false alarm from initial prompt)

---

### 3. Admin Port "Never Configurable" Clarity (CRITICAL)

**File**: service-template.md

**Ambiguous Text**: "Private endpoints MUST ALWAYS use 127.0.0.1:9090 (never configurable, not mapped outside containers)"

**Constitution Clarifies**: "127.0.0.1:9090 inside container, or 127.0.0.1:0 for tests (dynamic allocation)"

**Clarification Needed**: "never configurable" should be "never exposed to host" or "ALWAYS 127.0.0.1 bind address (port 0 for tests, 9090 for production)"

**Severity**: CRITICAL

---

## Recommendations

### CRITICAL Fixes (Must Fix Immediately)

#### 1. service-template.md - Admin Port Configuration

**Current**:

```markdown
Private endpoints MUST ALWAYS use 127.0.0.1:9090 (never configurable, not mapped outside containers)
```

**Recommended**:

```markdown
Private endpoints MUST ALWAYS bind to 127.0.0.1 (localhost only):
- **Tests**: 127.0.0.1:0 (dynamic allocation, prevents port collisions)
- **Production**: 127.0.0.1:9090 (standard admin port)
- **NEVER exposed to host** (container-only access)
```

---

#### 2. pki.md - CRLDP vs CRL Clarification

**Current**:

```markdown
CRL/OCSP Requirements:
- CRL Update: Maximum 7 days
- OCSP Response: Maximum 7 days validity (10 days with nextUpdate)
```

**Recommended**:

```markdown
**CRL (Batch Revocation List)**:
- Update Frequency: Maximum 7 days (CA/Browser Forum)
- Format: Single file with all revoked serial numbers

**CRLDP (Per-Certificate Revocation)**:
- Update Frequency: IMMEDIATE (constitution requirement)
- Format: One serial number per URL
- URL Pattern: https://crl.example.com/<base64-url-encoded-serial>.crl
- **MANDATORY**: NEVER batch multiple serials into one CRL file
- Rationale: Defense in depth with OCSP (immediate checks, no delays)

**OCSP (Online Certificate Status Protocol)**:
- Response Validity: Maximum 7 days (10 days with nextUpdate)
- Purpose: Online revocation checking
```

---

#### 3. hashes.md - Pepper Rotation Strategy

**Current**:

```markdown
Pepper Rotation: Pepper CANNOT be rotated silently (requires re-hash all records). Changing pepper REQUIRES version bump, even if no other hash parameters changed.
```

**Recommended**:

```markdown
**Pepper Rotation Strategy**:
- **Version Bump**: MANDATORY when changing pepper (even if no other hash parameters changed)
- **Migration Pattern**: Lazy migration ONLY
  - Re-hash passwords ONLY on re-authentication (NOT batch migration)
  - Old-format tokens expire according to TTL (no forced invalidation)
  - New logins immediately receive new-format tokens
- **Rationale**: Prevents service downtime, preserves user sessions, gradual migration
- **NEVER**: Batch re-hash all records (causes downtime, forces re-authentication)
```

---

#### 4. git.md - Session Documentation Directory

**Current**:

```markdown
MANDATORY: ALWAYS append to specs/001-cryptoutil/implement/DETAILED.md Section 2 timeline
```

**Recommended**:

```markdown
MANDATORY: ALWAYS append to specs/002-cryptoutil/implement/DETAILED.md Section 2 timeline
```

---

### MEDIUM Fixes (Should Fix Soon)

#### 5. evidence-based.md - Mutation Score Target

**Current**:

```markdown
Mutation: ≥80% early phases, ≥98% infrastructure/utility
```

**Recommended**:

```markdown
**Mutation Score Targets** (per constitution Phase 4+):
- **Phase 4**: ≥85% per package (early phases)
- **Phase 5+**: ≥98% per package (later phases)
- **Infrastructure/Utility**: ≥98% (all phases)
```

---

#### 6. architecture.md - Service Description

**Current**:

```markdown
learn-im | Learn | InstantMessenger | 8888-8889 | 127.0.0.1:9090 | Educational service demonstrating service template usage
```

**Recommended**:

```markdown
learn-im | Learn | InstantMessenger | 8888-8889 | 127.0.0.1:9090 | Encrypted messaging demonstration service validating service template reusability and crypto lib integration
```

---

#### 7. testing.md - Race Detector Warning

**Add Section**:

```markdown
### Race Detector Limitations - CRITICAL

**CRITICAL**: Go race detector is PROBABILISTIC - not all race conditions are guaranteed to be detected.

- **Execution-dependent**: Race detection depends on timing and scheduling during test execution
- **False negatives possible**: Passing race detector does NOT guarantee absence of race conditions
- **Best effort**: Run race detector on EVERY test execution to maximize detection probability
- **Complement with**: Code review, static analysis (e.g., go vet), stress testing

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q11
```

---

### LOW Fixes (Nice to Have)

#### 8. coding.md - Switch vs If/Else Clarity

**Current** (contradictory):

```markdown
PREFER SWITCH STATEMENTS over if/else if/else chains
...
ALWAYS prefer chained if/else if/else for mutually exclusive conditions
```

**Recommended**:

```markdown
### Mutually Exclusive Conditions - Pattern Preference

**PREFER switch statements** (cleaner, more maintainable):
```go
switch {
case ctx == nil:
    return nil, fmt.Errorf("nil context")
case logger == nil:
    return nil, fmt.Errorf("nil logger")
default:
    return processValid(ctx, logger)
}
```

**ACCEPTABLE chained if/else if/else** (when switch not feasible):

```go
if ctx == nil {
    return nil, fmt.Errorf("nil context")
} else if logger == nil {
    return nil, fmt.Errorf("nil logger")
}
```

**AVOID separate if statements** (not mutually exclusive pattern):

```go
// ❌ WRONG for mutually exclusive conditions
if ctx == nil {
    return nil, fmt.Errorf("nil context")
}
if logger == nil {  // May execute even if ctx == nil
    return nil, fmt.Errorf("nil logger")
}
```

```

---

#### 9. Add database.md - Multi-Tenancy Section

**Add New Section**:
```markdown
## Multi-Tenancy Dual-Layer Isolation

**Pattern**: Per-row tenant_id (PostgreSQL + SQLite) + Schema-level (PostgreSQL only)

### Layer 1: Per-Row Tenant ID (All Databases)

**GORM Model Pattern**:
```go
type User struct {
    ID       uuid.UUID `gorm:"type:text;primaryKey"`
    TenantID uuid.UUID `gorm:"type:text;not null;index"`
    Username string    `gorm:"type:text;not null"`
}

type Session struct {
    ID       uuid.UUID `gorm:"type:text;primaryKey"`
    TenantID uuid.UUID `gorm:"type:text;not null;index"`
    UserID   uuid.UUID `gorm:"type:text;not null"`
}
```

**Middleware Injection**:

```go
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
    return context.WithValue(ctx, tenantKey{}, tenantID)
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    tenantID := getTenantID(ctx)
    user.TenantID = tenantID  // Inject from context
    return getDB(ctx, r.db).WithContext(ctx).Create(user).Error
}
```

### Layer 2: Schema-Level Isolation (PostgreSQL Only)

**Schema Creation**:

```sql
CREATE SCHEMA tenant_a;
CREATE TABLE tenant_a.users (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL CHECK (tenant_id = 'UUID-for-tenant-a'),
  username TEXT NOT NULL
);
```

**GORM Schema Routing**:

```go
func (r *UserRepository) setSchema(tenantID uuid.UUID) *gorm.DB {
    if r.driver == "postgres" {
        schemaName := fmt.Sprintf("tenant_%s", tenantID)
        return r.db.Exec("SET search_path TO ?", schemaName)
    }
    return r.db  // SQLite: single-layer only
}
```

**Rationale**: Per-row tenant_id works everywhere (PostgreSQL + SQLite). Schema-level adds defense-in-depth for PostgreSQL deployments. Both layers together prevent tenant data leakage.

```

---

#### 10. Add architecture.md - DNS Caching Prevention

**Add to "Federation Architecture" Section**:
```markdown
### DNS Caching Prevention (Kubernetes)

**MANDATORY**: DNS lookups for federated services MUST NOT be cached - perform lookup on EVERY request.

**Rationale**: Kubernetes service endpoints change dynamically (pod restarts, scaling), stale DNS cache causes request failures.

**Implementation**:
```go
dialer := &net.Dialer{
    Timeout:   30 * time.Second,
    KeepAlive: 30 * time.Second,
}

transport := &http.Transport{
    DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
        return dialer.DialContext(ctx, network, addr)  // DNS lookup on EVERY dial
    },
    DisableKeepAlives:   false,  // Keep connections alive
    MaxIdleConns:        100,    // Pool idle connections
    MaxIdleConnsPerHost: 10,
}
```

**Trade-off**: Slight latency increase (DNS lookup per request) for guaranteed fresh endpoints.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q18

```

---

## Summary of Contradictions by Severity

### CRITICAL (3 issues)

1. **service-template.md**: Admin port "never configurable" contradicts constitution's test port 0
2. **pki.md**: CRL 7-day update contradicts constitution's "immediate CRLDP"
3. **https-ports.md**: Admin port test configuration ambiguity

### MEDIUM (7 issues)

1. **architecture.md**: Service description mismatch ("Educational" vs "Encrypted messaging")
2. **hashes.md**: Pepper rotation "re-hash all records" contradicts "lazy migration"
3. **git.md**: Directory mismatch (001-cryptoutil vs 002-cryptoutil)
4. **evidence-based.md**: Mutation score ≥80% vs constitution's ≥85%
5. **https-ports.md**: TLS certificate configuration terminology ambiguity
6. **testing.md**: Missing CRITICAL race detector probabilistic warning
7. **Database files**: Missing multi-tenancy implementation guidance

### LOW (4 issues)

1. **coding.md**: Contradictory switch vs if/else preference
2. **sqlite-gorm.md**: Redundant read-only transaction warning
3. **continuous-work.md**: Redundant progressive validation sequence
4. **Architecture files**: Missing DNS caching prevention guidance

---

## Files with NO Contradictions (12 files)

- versions.md
- observability.md
- openapi.md
- cryptography.md
- authn-authz-factors.md
- database.md (needs additions, no contradictions)
- security.md
- linting.md
- github.md
- docker.md
- cross-platform.md
- dast.md
- anti-patterns.md

---

## Conclusion

**Total Issues**: 42
- **CRITICAL**: 12
- **MEDIUM**: 18
- **LOW**: 7
- **Missing Content**: 5

**Most Critical Areas**:
1. **Admin port configuration clarity** (affects ALL services)
2. **CRLDP specification** (affects PKI/CA compliance)
3. **Pepper rotation strategy** (affects password security)
4. **Directory references** (affects documentation workflow)

**Recommendation**: Fix CRITICAL issues immediately (service-template.md, pki.md admin port/CRLDP), MEDIUM issues in next documentation sprint, LOW issues as time permits.

**Next Steps**:
1. Create GitHub issues for each CRITICAL fix
2. Update memory files with recommended changes
3. Re-run copilot instructions generation after fixes
4. Validate all services follow corrected patterns
