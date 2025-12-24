# Anti-Patterns and Lessons Learned - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/06-03.anti-patterns.instructions.md`

**Purpose**: Document historically regression-prone areas, common mistakes, and lessons learned from P0 incidents to prevent future regressions.

---

## CRITICAL Regression-Prone Areas

### Format_go Self-Modification - P0 INCIDENTS

**Historical Context**: Multiple P0 regressions (Nov 17, Nov 20, Dec 16 × 2) where LLM agents inadvertently modified self-exclusion patterns during refactoring.

#### Root Cause

**LLM Context Loss**: During narrow-focus refactoring (e.g., "modernize code to Go 1.25"), agents lose exclusion context:

1. Read single file: `enforce_any.go`
2. See "verbose comments" and `` `interface{}` `` usage
3. Apply modernization: Change `` `interface{}` `` → `` `any` ``
4. **Miss critical context**: These are intentional protection patterns
5. **Result**: Self-modification protection bypassed

#### Historical Incidents

**Incident 1 (commit b934879b, Nov 17)**:
- Added backticks to comments to prevent pattern replacement
- Symptom: Replacement logic caught comment text as code

**Incident 2 (commit 71b0e90d, Nov 20)**:
- Added comprehensive self-exclusion patterns
- Symptom: `filter.go` didn't exclude `enforce_any.go` itself

**Incident 3 (commit b0e4b6ef, Dec 16)**:
- Fixed infinite loop (counted "any" instead of "interface{}")
- Symptom: Replacement counter logic broken

**Incident 4 (commit 8c855a6e, Dec 16)**:
- Fixed test data (used "any" instead of "interface{}")
- Symptom: Test expectations didn't match replacement behavior

#### MANDATORY Prevention Rules

**NEVER DO**:
- ❌ Modify comments or test data in `enforce_any.go` without reading full package context
- ❌ Change `` `interface{}` `` to `` `any` `` in format_go package without verification
- ❌ Refactor code in isolation (single-file view)
- ❌ Simplify "verbose" CRITICAL comments without understanding purpose

**ALWAYS DO**:
- ✅ Read complete package context before refactoring self-modifying code
- ✅ Check for CRITICAL/SELF-MODIFICATION tags in comments
- ✅ Verify self-exclusion patterns exist and are respected
- ✅ Run tests after ANY changes to format_go package
- ✅ Grep for exclusion constants in `magic_cicd.go`

#### Pattern Recognition

**Indicators of Intentional Protection**:
- **CRITICAL comments**: High-priority annotations requiring preservation
- **Backticked strings** in code: `` `interface{}` `` → Prevents replacement by pattern matching
- **Test data patterns**: May use "wrong" values intentionally (e.g., `interface{}` as input to test replacement)
- **Self-exclusion constants**: `MagicCICDFilterExcludeEnforceAny` in `magic_cicd.go`

#### Code Archaeology Checklist (Before Refactoring)

**Required Reading**:
1. `enforce_any.go` (target file)
2. `filter.go` (self-exclusion patterns)
3. `magic_cicd.go` (exclusion constants)
4. `format_go_test.go` (test data patterns)
5. `self_modification_test.go` (validation patterns)
6. Post-mortems: `docs/P0.*` files (historical lessons)
7. Git log: `git log --oneline enforce_any.go` (change history)

**Key Questions**:
1. Why does this code exist? (Read README, post-mortems)
2. What protections are in place? (Check self-exclusion patterns)
3. Are "verbose" comments intentional? (Look for CRITICAL tags)
4. What tests validate this behavior? (Read test files)
5. Has this failed before? (Check post-mortems, git log)

**Example**:
```bash
# Correct archaeology pattern
read_file enforce_any.go              # Target file
read_file filter.go                   # Self-exclusion patterns
read_file magic_cicd.go               # Exclusion constants
read_file format_go_test.go           # Test data
read_file self_modification_test.go   # Validation
git log --oneline enforce_any.go      # Change history
```

---

### Windows Firewall Exception Prevention - CRITICAL

**Problem**: Binding to `0.0.0.0` triggers Windows Firewall exception prompts, blocking CI/CD automation.

**Impact**: Each `0.0.0.0` binding = 1 Windows Firewall popup = blocked test execution

#### NEVER DO

❌ **Bind to 0.0.0.0 in unit tests or integration tests**:
```go
// WRONG - triggers firewall prompt
listener, _ := net.Listen("tcp", "0.0.0.0:8080")
```

❌ **Use "localhost" (may resolve to IPv6 `::1`)**:
```go
// WRONG - ambiguous, may resolve to ::1
listener, _ := net.Listen("tcp", "localhost:8080")
```

#### ALWAYS DO

✅ **Bind to 127.0.0.1 (IPv4 loopback) in tests**:
```go
// CORRECT - no firewall prompt
addr := fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, port)  // "127.0.0.1"
listener, _ := net.Listen("tcp", addr)
```

✅ **Use 0.0.0.0 ONLY in Docker containers**:
```yaml
# docker-compose.yml
services:
  kms:
    ports:
      - "8080:8080"  # Maps external 8080 to container's 0.0.0.0:8080
    command: ["--bind-address=0.0.0.0"]
```

**Rationale**: Docker containers run in isolated network namespace → no Windows Firewall prompt

#### Detection Pattern

**Symptoms**:
- CI/CD hangs during test execution
- Windows Firewall dialog appears during local test runs
- `net.Listen` calls timeout in GitHub Actions

**Fix**:
1. Grep for `0.0.0.0` in test files: `grep -r "0.0.0.0" **/*_test.go`
2. Replace with `127.0.0.1` or `cryptoutilMagic.IPv4Loopback`
3. Verify tests pass locally without firewall prompts

---

### SQLite Connection Pool Deadlocks - P0 INCIDENT

**Problem**: GORM explicit transactions require multiple connections, but `MaxOpenConns=1` causes deadlock.

**Symptom**: Tests hang indefinitely when using `db.Begin()` with SQLite

#### Root Cause

**GORM Transaction Wrapper Pattern**:
```go
// GORM transaction acquires connection #1
tx := db.Begin()

// Repository method uses getDB(ctx, r.db)
// If ctx has no transaction, uses base db.WithContext(ctx)
// Tries to acquire connection #2 → DEADLOCK (MaxOpenConns=1)
result := getDB(ctx, r.db).WithContext(ctx).Create(user).Error
```

**MaxOpenConns=1**: Prevents second connection acquisition → indefinite hang

#### NEVER DO

❌ **Set MaxOpenConns=1 with GORM transactions**:
```go
// WRONG - causes deadlock with GORM transactions
sqlDB.SetMaxOpenConns(1)
sqlDB.SetMaxIdleConns(1)
```

❌ **Use sql.TxOptions{ReadOnly: true} with SQLite** (not supported):
```go
// WRONG - SQLite doesn't support read-only transactions
tx := db.Begin(&sql.TxOptions{ReadOnly: true})
```

#### ALWAYS DO

✅ **Set MaxOpenConns=5 for GORM transaction support**:
```go
// CORRECT - allows GORM transaction wrapper + repository operations
sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
```

✅ **Enable WAL mode**:
```go
// CORRECT - allows multiple concurrent readers + 1 writer
sqlDB.Exec("PRAGMA journal_mode=WAL;")
```

✅ **Set busy timeout**:
```go
// CORRECT - retry on lock contention instead of immediate failure
sqlDB.Exec("PRAGMA busy_timeout = 30000;")  // 30 seconds
```

✅ **Use standard transactions for read operations** (SQLite doesn't support read-only):
```go
// CORRECT - standard transaction works for reads
tx := db.Begin()  // NOT db.Begin(&sql.TxOptions{ReadOnly: true})
```

#### Rationale

**Why MaxOpenConns=5**:
- GORM transaction wrapper uses 1 connection
- Repository operations use separate connection (from pool)
- SQLite still enforces 1 concurrent writer (via WAL mode + busy timeout)
- Connection pool prevents contention, busy timeout handles retries

**Reference Implementation**: `internal/server/repository/sqlrepository/sql_provider.go` lines 201-213

#### Magic Constants

Use from `internal/common/magic/magic_database.go`:
- `cryptoutilMagic.DBSQLiteBusyTimeout` = 30 seconds
- `cryptoutilMagic.SQLiteMaxOpenConnections` = 5

---

### Docker Compose Port Conflicts - E2E FAILURES

**Problem**: Multiple services include same telemetry compose file, causing port conflicts on host machine.

**Historical Incident**: 2025-12-19 session - CA and JOSE deployments both tried to bind OTEL collector ports 4317, 4318, 8888, 8889, 13133 to host.

#### Root Cause

**Naive Compose Include Pattern**:
```yaml
# ca/compose.yml
include:
  - path: ../telemetry/otel-collector.yml  # Binds ports to host

# jose/compose.yml
include:
  - path: ../telemetry/otel-collector.yml  # CONFLICT - same ports
```

**Result**: Second `docker compose up` fails with "port already allocated"

#### NEVER DO

❌ **Expose container ports to host if multiple instances may run simultaneously**:
```yaml
# WRONG - exposes to host, conflicts with other services
services:
  opentelemetry-collector:
    ports:
      - "4317:4317"  # Host port 4317 → container port 4317
      - "4318:4318"
```

❌ **Use same compose include in multiple services without considering port conflicts**

#### ALWAYS DO

✅ **Use container-to-container networking (no host port mappings)**:
```yaml
# CORRECT - no host ports, container-to-container only
services:
  opentelemetry-collector:
    # NO ports section (no host mapping)
    networks:
      - cryptoutil-network
```

✅ **Services communicate via Docker network using container names**:
```go
// Application configuration
otlpEndpoint: "opentelemetry-collector-contrib:4317"  // Container name, not localhost
```

✅ **Test E2E workflows with sequential deployments to catch port conflicts**:
```bash
# E2E test pattern
docker compose -f deployments/ca/compose.yml up -d
docker compose -f deployments/jose/compose.yml up -d  # Should work (no port conflicts)
```

#### Detection Pattern

**Symptoms**:
- `docker compose up` for service B fails after service A succeeds
- Error: "bind: address already in use" or "port is already allocated"
- E2E workflow fails with "cannot start service X"

**Diagnosis**:
```bash
# Check what's using the port
netstat -ano | findstr "4317"  # Windows
lsof -i :4317                  # Linux/Mac

# Check included compose files
grep -r "ports:" deployments/*/compose.yml
```

**Fix**:
1. Remove host port mappings from shared compose files (e.g., `telemetry/otel-collector.yml`)
2. Use container-to-container networking only
3. Update application configs to use container names (not `localhost:4317`)
4. Test all E2E workflows in sequence to verify no conflicts

---

### Incomplete Service Implementation - WORKFLOW DEBUGGING

**Problem**: Missing public HTTP servers cause cascading configuration errors that mask root cause.

**Historical Incident**: 2025-12-20 WORKFLOW-FIXES - Identity services missing `server.go` files, only admin servers implemented.

#### Root Cause

**Assumption Bias**: Assumed container crashes were ALWAYS configuration problems, not missing code.

#### Symptom Pattern

**Progressive Configuration Fixes with Zero Symptom Change**:
```
Round 3: 331 bytes log - "TLS cert file required"
Round 4: 313 bytes log - "database DSN required"
Round 5: 196 bytes log - "Starting AuthZ server..." (no error logged)
Round 6: 196 bytes log - SAME BYTES (zero change after valid fix)
```

**Pattern**: Decreasing log bytes = earlier crash = deeper problem (NOT configuration)

**Zero symptom change** despite valid fixes = implementation issue, NOT configuration

#### NEVER DO

❌ **Keep applying configuration fixes when symptoms don't change**:
```bash
# WRONG - 6 rounds of config fixes with zero symptom improvement
# Round 1: Fix TLS paths
# Round 2: Fix database DSN
# Round 3: Fix credentials
# Round 4-6: More config changes, SAME 196-byte log output
```

❌ **Assume container crash is always a configuration problem**

❌ **Debug configuration before verifying complete architecture exists**

#### ALWAYS DO

✅ **Code archaeology FIRST - compare with working service before debugging config**:
```bash
# CORRECT - 9 minutes to identify root cause
# 1. Download container logs from failed workflow
# 2. Compare with working service (KMS) file structure
# 3. Notice missing server.go in identity services
# 4. Check Application.Start() - only initializes admin server
# 5. Root cause: Public HTTP server not implemented
```

✅ **Verify all required files exist** (server.go, application.go, admin.go)

✅ **Check Application.Start() initializes both public + admin servers**

✅ **Compare container log byte counts across fix attempts**:
- Decreasing bytes = earlier crash = deeper problem
- Same bytes = no symptom change = implementation issue

#### Detection Pattern

**Indicators of Missing Implementation**:

1. **Byte count pattern**:
   - Config fix Round 1: 331 bytes
   - Config fix Round 2: 313 bytes
   - Config fix Round 3: 196 bytes
   - Config fix Round 4: 196 bytes ← SAME (zero improvement)

2. **Error message pattern**:
   - Early rounds: Specific errors ("TLS cert file required", "database DSN required")
   - Later rounds: Generic startup ("Starting AuthZ server...") then silence

3. **File structure pattern**:
   - Working service (KMS): `server.go`, `application.go`, `admin.go` all present
   - Broken service (Identity): Only `application.go`, `admin.go` (missing `server.go`)

#### Time Wasted Pattern

**Configuration Debugging**: 40-60 minutes (4-6 rounds × 10 minutes each)

**Code Archaeology Upfront**: 9 minutes (download logs + compare architecture)

**Lesson**: Code archaeology should be FIRST step, NOT last resort

#### Reference

See `docs/WORKFLOW-FIXES-CONSOLIDATED.md` for complete timeline of 2025-12-20 workflow debugging session.

---

## Testing Anti-Patterns

### Coverage Improvement Without Baseline Analysis

**Problem**: Writing 60+ tests without analyzing baseline coverage HTML = 0% improvement.

**Symptom**: Massive test file (1000+ lines), coverage unchanged

#### NEVER DO

❌ **Write tests without checking baseline coverage first**:
```bash
# WRONG - trial and error
# Write TestFunc1, run coverage → 60%
# Write TestFunc2, run coverage → 60% (no improvement)
# Write TestFunc3, run coverage → 60% (still no improvement)
# Result: 60+ tests, 0% improvement
```

❌ **Add tests randomly hoping to hit uncovered code**

❌ **Trial-and-error test writing cycles**

#### ALWAYS DO

✅ **Generate baseline coverage**:
```bash
go test ./pkg -coverprofile=./test-output/coverage_pkg.out
```

✅ **Analyze HTML to identify RED (uncovered) lines**:
```bash
go tool cover -html=./test-output/coverage_pkg.out -o ./test-output/coverage_pkg.html
# Open in browser, find RED lines
```

✅ **Identify specific functions with coverage gaps**:
```bash
go tool cover -func=./test-output/coverage_pkg.out | grep "0.0%"
# Focus on 0% coverage functions first
```

✅ **Write targeted tests for identified gaps**:
```go
// Test covers specific RED line: error path in ParseKey()
func TestParseKey_InvalidFormat(t *testing.T) {
    _, err := ParseKey("invalid")
    require.Error(t, err)  // Covers RED line
}
```

✅ **Verify improvement with new coverage report**:
```bash
go test ./pkg -coverprofile=./test-output/coverage_pkg_new.out
go tool cover -func=./test-output/coverage_pkg_new.out | grep total
# Compare: 60% → 95% (35 percentage point improvement)
```

#### Lesson

**Coverage ≠ Test Count**: Many tests can add 0% if exercising already-covered code paths.

**Efficient Pattern**: Baseline → Analyze RED lines → Targeted tests → Verify

---

### Individual Test Functions vs Table-Driven

**Problem**: Creating `TestFunc_Variant1`, `TestFunc_Variant2`, `TestFunc_Variant3` as separate functions.

**Result**: 1371-line test file (2.7× hard limit of 500 lines), maintenance nightmare, slower LLM processing

#### NEVER DO

❌ **Separate test functions for algorithm/key size variants**:
```go
// WRONG - separate functions
func TestGenerateKey_RSA2048(t *testing.T) { /* ... */ }
func TestGenerateKey_RSA3072(t *testing.T) { /* ... */ }
func TestGenerateKey_RSA4096(t *testing.T) { /* ... */ }
func TestGenerateKey_ECDSAP256(t *testing.T) { /* ... */ }
// Result: 20 separate functions × 60 lines each = 1200 lines
```

❌ **Duplicate setup code across multiple test functions**

#### ALWAYS DO

✅ **Use table-driven tests with variants as rows**:
```go
// CORRECT - table-driven
func TestGenerateKey(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name    string
        keyType KeyType
        keySize int
    }{
        {name: "RSA_2048", keyType: RSA, keySize: 2048},
        {name: "RSA_3072", keyType: RSA, keySize: 3072},
        {name: "RSA_4096", keyType: RSA, keySize: 4096},
        {name: "ECDSA_P256", keyType: ECDSA, keySize: 256},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // Test logic (shared setup, minimal duplication)
        })
    }
}
// Result: 1 function × 80 lines = 80 lines (15× reduction)
```

✅ **Group related test cases in single function with `t.Run(tt.name, ...)`**

✅ **Keep test files under 500 lines (hard limit)**

#### File Size Limits

| Limit | Lines | Action Required |
|-------|-------|-----------------|
| Soft | 300 | Ideal target |
| Medium | 400 | Acceptable with justification |
| Hard | 500 | NEVER EXCEED - refactor required |

---

### Race Condition Testing Patterns

**Problem**: Race detector overhead (~10×) causes test timeouts with short deadlines.

**Symptom**: Tests pass normally, fail with `context deadline exceeded` under `-race`

#### NEVER DO

❌ **Hardcode 2-second timeouts for network operations**:
```go
// WRONG - fails under race detector
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
client := &http.Client{Timeout: 2 * time.Second}
```

❌ **Assume race detector runs at normal speed**

#### ALWAYS DO

✅ **Use 10+ second timeouts for network operations in race mode**:
```go
// CORRECT - accounts for race detector overhead
timeout := 10 * time.Second
if testing.Short() {
    timeout = 2 * time.Second  // Fast mode without race detector
}
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
```

✅ **Increase test timeouts 10× when race detector enabled**

✅ **Add thread-safe accessor methods (RLock/RUnlock) for shared state**:
```go
// CORRECT - mutex-protected map access
func (s *SessionStore) Get(sessionID string) (*Session, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    session, ok := s.sessions[sessionID]
    return session, ok
}
```

✅ **Never access shared maps/slices without mutex protection**

#### Pattern

`context deadline exceeded` errors under `-race` = insufficient timeout, NOT actual bug

**Fix**: Increase timeout 10× or use dynamic timeout based on race detector flag

---

## Git Workflow Anti-Patterns

### Amending Repeatedly (Loses History)

**Problem**: Using `git commit --amend` repeatedly loses history, masks mistakes, breaks bisect.

**Symptom**: Commit history shows 1 commit with 50 changes, impossible to identify when specific bug introduced

#### NEVER DO

❌ **Amend commit after push (breaks shared history)**:
```bash
# WRONG - breaks collaborators' repos
git commit -m "initial"
git push origin main
# ... discover bug, fix it
git commit --amend  # BREAKS shared history
git push --force    # DANGEROUS
```

❌ **Amend repeatedly during debugging session**:
```bash
# WRONG - loses incremental context
git commit -m "fix bug"
# ... test fails, fix again
git commit --amend
# ... test fails again, fix again
git commit --amend
# Result: 1 commit, 3 attempts hidden, can't bisect
```

❌ **Use amend to hide incremental fixes**

#### ALWAYS DO

✅ **Commit each logical unit independently**:
```bash
# CORRECT - preserve full timeline
git commit -m "fix(format_go): restore clean baseline from 07192eac"
git commit -m "fix(format_go): add defensive check with filepath.Abs()"
git commit -m "test(format_go): verify self_modification_test catches regressions"
# Result: 3 commits, full context, bisect-friendly
```

✅ **Preserve full timeline of changes and decisions**

✅ **Enable git bisect to identify when bugs were introduced**

✅ **Use amend ONLY for immediate typo fixes (within 1 minute, before push)**:
```bash
# ACCEPTABLE - immediate typo fix
git commit -m "fix(auth): add mising validation"  # Typo in message
git commit --amend -m "fix(auth): add missing validation"  # Fix within 1 minute
```

#### Rationale

**Incremental Commits**:
- Preserve context (why each change was made)
- Enable selective revert (revert specific fix without losing others)
- Show thought process (debugging steps visible)
- Support bisect (identify exact commit that introduced bug)

---

### Applying Fixes to Corrupted HEAD

**Problem**: Assuming HEAD is correct when it may be corrupted from previous failed attempts.

**Symptom**: Apply "one more fix" on top of corrupted code → fails again → more fixes → fails again (cycle)

#### NEVER DO

❌ **Apply "one more fix" on top of corrupted code**:
```bash
# WRONG - HEAD may be corrupted from previous attempts
# Attempt 1: Apply fix A → test fails
# Attempt 2: Apply fix B on top of fix A → test fails
# Attempt 3: Apply fix C on top of fix A+B → test fails
# Problem: Fix A may have introduced corruption
```

❌ **Mix baseline restoration with new fixes in same commit**:
```bash
# WRONG - can't isolate which change fixed the bug
git checkout 07192eac -- enforce_any.go
# Apply new fix in same file
git commit -m "fix everything"  # Too broad
```

❌ **Assume HEAD is always clean**

#### ALWAYS DO

✅ **Restore clean baseline from known-good commit FIRST**:
```bash
# CORRECT - start from known-good state
# 1. Find last known-good commit
git log --oneline --grep="baseline" | head -5

# 2. Restore entire package from clean commit
git checkout 07192eac -- internal/cmd/cicd/format_go/

# 3. Verify baseline works
go test ./internal/cmd/cicd/format_go/
# PASS (baseline confirmed clean)

# 4. Commit baseline restoration
git commit -m "fix(format_go): restore clean baseline from 07192eac"
```

✅ **Verify baseline works (tests pass)**

✅ **Apply ONLY the new fix (minimal change)**:
```bash
# 5. Apply targeted fix ONLY
# Edit enforce_any.go: Add filepath.Abs() check

# 6. Verify fix independently
go test ./internal/cmd/cicd/format_go/
# PASS (fix works on clean baseline)
```

✅ **Commit as NEW commit with clear description**:
```bash
# 7. Commit fix as new commit (not amend)
git commit -m "fix(format_go): add defensive check with filepath.Abs()"
```

#### Pattern

**Find Last Known-Good** → **Restore Baseline** → **Verify Baseline** → **Apply Targeted Fix** → **Verify Fix** → **Commit Separately**

**Why**: HEAD corruption accumulates from failed attempts. Start fresh from verified clean state.

---

## Documentation Anti-Patterns

### Creating Standalone Session Documentation

**Problem**: Creating `docs/SESSION-2025-12-14-*.md` files leads to documentation bloat.

**Result**: 50+ session docs scattered across `docs/`, difficult to find historical context

#### NEVER DO

❌ **Create dated session documentation files**:
```
docs/SESSION-2025-12-14-coverage-improvement.md
docs/SESSION-2025-12-15-mutation-testing.md
docs/SESSION-2025-12-16-e2e-workflows.md
# Result: 50+ session files, no single source of truth
```

❌ **Create standalone analysis documents for session work**

❌ **Create separate work log files per session**

#### ALWAYS DO

✅ **Append to `specs/*/implement/DETAILED.md` Section 2 timeline**:
```markdown
## Section 2: Implementation Timeline

### 2025-12-14: Coverage Improvement Session
- Work completed: Commit abc1234 (20 new tests)
- Coverage: 60% → 95% (+35 percentage points)
- Key findings: Needed baseline HTML analysis before writing tests
- Next steps: Mutation testing

### 2025-12-15: Mutation Testing Session
- Work completed: Commit def5678 (gremlins configuration)
- Mutation score: 78% → 85% (+7 percentage points)
- Key findings: Probabilistic execution for test speed
- Next steps: E2E workflow validation
```

✅ **Single source of truth for implementation timeline**

✅ **Create separate docs ONLY for permanent reference material** (ADRs, post-mortems, user guides)

#### Rule of Thumb

**Session-specific work** → Append to `DETAILED.md`

**Permanent reference** → Create dedicated doc:
- `docs/ADR-001-database-choice.md` (architectural decision)
- `docs/P0.1-format-go-regression.md` (post-mortem)
- `docs/USER-GUIDE.md` (user documentation)

---

## Architecture Anti-Patterns

### Missing Service Federation Configuration

**Problem**: Services don't know how to discover or communicate with federated services.

**Symptom**: Hardcoded service URLs in code, fails when service moves

#### NEVER DO

❌ **Hardcode service URLs in application code**:
```go
// WRONG - hardcoded URL
identityURL := "https://identity-authz:8180"
```

❌ **Assume services are always co-located**

#### ALWAYS DO

✅ **Use configuration for service discovery** (YAML, environment, DNS):
```yaml
# config.yaml
federation:
  identity_url: "${IDENTITY_SERVICE_URL}"  # Configurable
  identity_enabled: true
  identity_timeout: 10s
```

✅ **Support multiple federation patterns**:
- DNS-based discovery (Kubernetes)
- Config file URLs (Docker Compose)
- Service mesh integration (Consul, Istio)

✅ **Implement graceful degradation when federated services unavailable**:
```go
// Graceful degradation pattern
if federationEnabled {
    result, err := callFederatedService(ctx)
    if err != nil {
        log.Warn("federated service unavailable, using fallback")
        return localFallback(ctx)
    }
    return result
}
return localFallback(ctx)
```

#### Pattern

**Service A depends on Service B** → Configure B's URL in A's config, NOT hardcode

---

## Performance Anti-Patterns

### Mutation Testing Timeout (>45 minutes)

**Problem**: Running gremlins on entire codebase sequentially causes 45-minute timeouts.

**Symptom**: CI workflow exceeds job timeout, incomplete mutation coverage

#### NEVER DO

❌ **Run mutation testing on all packages sequentially**:
```bash
# WRONG - 45+ minutes sequential
gremlins unleash
# Runs all packages one-by-one: pkg1 (10min) → pkg2 (8min) → ... → timeout
```

❌ **Include test utilities and generated code in mutation scope**:
```yaml
# WRONG - wastes time on non-production code
gremlins unleash --tags=""  # Tests EVERYTHING including testutil/
```

#### ALWAYS DO

✅ **Parallelize by package using GitHub Actions matrix strategy**:
```yaml
# CORRECT - parallel execution
strategy:
  matrix:
    package:
      - internal/kms
      - internal/identity
      - internal/jose
steps:
  - run: gremlins unleash --tags="~integration,~e2e" ./{{ matrix.package }}
    timeout-minutes: 15
# Result: 4-6 packages in parallel, <20 minutes total
```

✅ **Exclude tests, generated code, vendor directories**:
```yaml
# CORRECT - focus on business logic only
gremlins unleash --tags="~integration,~e2e" --exclude="*_test.go,**/testutil/**,**/vendor/**"
```

✅ **Set per-job timeout (15 minutes max)**

✅ **Target <20 minutes total with parallel execution**

#### Optimization

**4-6 packages per parallel job**, focus on business logic only

**Expected Result**: Sequential 45 minutes → Parallel 15-20 minutes (2-3× speedup)

---

### Test Timing Violations (>15s per package)

**Problem**: Test packages taking >15 seconds due to exhaustive algorithm variant testing.

**Symptom**: `go test ./...` takes >180 seconds (violates target)

#### NEVER DO

❌ **Test every key size variant (RSA 2048/3072/4096) every time**:
```go
// WRONG - tests all variants every run
tests := []struct {
    name string
    size int
}{
    {name: "RSA_2048", size: 2048},  // ALWAYS runs
    {name: "RSA_3072", size: 3072},  // ALWAYS runs
    {name: "RSA_4096", size: 4096},  // ALWAYS runs
}
// Result: 3× test time
```

❌ **Use `TestProbAlways` for redundant variants**

#### ALWAYS DO

✅ **Use `TestProbTenth` (10%) or `TestProbQuarter` (25%) for algorithm variants**:
```go
// CORRECT - statistical sampling
tests := []struct {
    name string
    size int
    prob int
}{
    {name: "RSA_2048", size: 2048, prob: cryptoutilMagic.TestProbAlways},   // 100% (base)
    {name: "RSA_3072", size: 3072, prob: cryptoutilMagic.TestProbQuarter},  // 25%
    {name: "RSA_4096", size: 4096, prob: cryptoutilMagic.TestProbTenth},    // 10%
}
for _, tt := range tests {
    if rand.Intn(100) >= tt.prob {
        t.Skip("probabilistic skip")
    }
    // Test logic
}
// Result: Average 1.35 runs instead of 3 (2.2× speedup)
```

✅ **Reserve `TestProbAlways` (100%) for base algorithms only**

✅ **Target <15s per unit test package, <180s full unit test suite**

#### Rationale

**Statistical Sampling**: Bugs eventually caught without running all variants every time

**Magic Constants**:
- `TestProbAlways = 100` (100%) - Base algorithms
- `TestProbQuarter = 25` (25%) - Important variants
- `TestProbTenth = 10` (10%) - Redundant variants

---

## Key Takeaways

### Context Reading - CRITICAL
**ALWAYS read complete context before refactoring self-modifying code**. Check for CRITICAL tags, self-exclusion patterns, test validation patterns.

### Windows Firewall - CRITICAL
**ALWAYS bind to 127.0.0.1 in tests** (NEVER 0.0.0.0). Use 0.0.0.0 ONLY in Docker containers (isolated namespace).

### Coverage Analysis - MANDATORY
**ALWAYS analyze baseline HTML before writing tests**. Identify RED lines, write targeted tests, verify improvement.

### Incremental Commits - BEST PRACTICE
**NEVER amend repeatedly** - preserve history for bisect. Commit each logical unit independently.

### Restore from Clean - BEST PRACTICE
**ALWAYS restore clean baseline before applying fixes**. HEAD may be corrupted from previous attempts.

### Port Conflicts - CRITICAL
**Remove host port mappings for shared services in Docker Compose**. Use container-to-container networking only.

### Mutation Parallelization - PERFORMANCE
**NEVER run sequentially** - use GitHub Actions matrix. 4-6 packages per job, <20 minutes total.

### Test Timeouts - COMPATIBILITY
**ALWAYS increase timeouts 10× for race detector mode**. Race detector overhead ~10× normal execution.

---

## Cross-References

**Related Documentation**:
- Format_go protection patterns: `.specify/memory/coding.md`
- Windows Firewall prevention: `.specify/memory/security.md`
- SQLite configuration: `.specify/memory/sqlite-gorm.md`
- Docker Compose patterns: `.specify/memory/docker.md`
- Testing standards: `.specify/memory/testing.md`
- Git workflow: `.specify/memory/git.md`
- Workflow debugging: `docs/WORKFLOW-FIXES-CONSOLIDATED.md`
- Post-mortems: `docs/P0.*` files
