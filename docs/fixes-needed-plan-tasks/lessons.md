# Lessons Learned - JOSE Refactoring Blockers

## Root Cause Analysis

### What Went Wrong

Agent completed Phases 0-2 successfully but encountered blockers in Phase X (High Coverage) and did not create resolution phases.

**Primary Blocker**: Tests were creating new GORM DB per test instead of using TestMain pattern with shared server instance per package.

**Secondary Blockers**:
1. Docker Desktop dependency (cipher-im tests fail when Docker not running)
2. JOSE repositories stuck at 82.8% coverage (needs TestMain refactoring)
3. JOSE services stuck at 82.7% coverage (needs TestMain refactoring)

### Specific Failures

**Phase X - Task X.2.1**:
- **Issue**: TestInitDatabase_HappyPaths/PostgreSQL_Container fails
- **Root Cause**: Docker Desktop not running on Windows
- **Agent Action**: Documented failure but did NOT create resolution phase
- **Correct Action**: Create Phase Z.1 to fix Docker dependency

**Phase X - Task X.3.1**:
- **Issue**: JOSE repositories coverage 82.8% (target 98%)
- **Root Cause**: Tests creating new DB per test instead of shared TestMain server
- **Blocker**: TestMain pattern not used - each test calls setupTestDB()
- **Coverage Gap**: 15.2 percentage points
- **Agent Action**: Marked BLOCKED but did NOT create resolution phase
- **Correct Action**: Create Phase Z.2 (TestMain refactoring) + Phase Z.3 (comprehensive testing)

**Phase X - Task X.5.1**:
- **Issue**: JOSE services coverage 82.7% (target 95%)
- **Root Cause**: Tests creating new DB per test instead of shared TestMain server
- **Blocker**: Same TestMain pattern gap
- **Coverage Gap**: 12.3 percentage points  
- **Agent Action**: Marked BLOCKED but did NOT create resolution phase
- **Correct Action**: Create Phase Z.2 (TestMain refactoring) + Phase Z.4 (comprehensive testing)

**Phase Y - Mutation Testing**:
- **Issue**: 25 mutation testing tasks documented but not started
- **Root Cause**: Blocked on Phase X completion
- **Agent Action**: Documented Phase Y but did not resolve Phase X blockers
- **Correct Action**: Resolve Phase X blockers  Start Phase Y automatically

---

## Technical Deep Dive

### TestMain Pattern Gap Analysis

**CORRECT Pattern** (Used in most packages):
```go
var (
    testDB     *gorm.DB
    testServer *Server
    testRepo   Repository
)

func TestMain(m *testing.M) {
    // Setup heavyweight resources ONCE per package
    testDB = setupDatabase()
    testServer = setupServer(testDB)
    testRepo = NewRepository(testDB)
    go testServer.Start()
    defer testServer.Shutdown()

    exitCode := m.Run()
    os.Exit(exitCode)
}

func TestSomething(t *testing.T) {
    // Use shared testDB, testRepo - already initialized
    // Create orthogonal test data with UUIDv7
    result := testRepo.Get(ctx, googleUuid.NewV7())
}
```

**WRONG Pattern** (Found in 5 packages):
```go
func TestSomething(t *testing.T) {
    // Creates NEW database for EVERY test - SLOW and WASTEFUL
    db := setupTestDB(t)
    repo := NewRepository(db)

    result := repo.Get(ctx, id)
}
```

**Why TestMain Pattern is Mandatory**:
1. ✅ **Performance**: Database setup runs ONCE (not N times per test)
2. ✅ **Integration Testing**: Tests use real GORM DB with full error path coverage
3. ✅ **No Mocking Needed**: Real database handles all error scenarios
4. ✅ **Test Isolation**: UUIDv7 for orthogonal test data prevents conflicts
5. ✅ **Production Parity**: Same patterns as E2E tests

**Violations Found**:
- ❌ internal/apps/template/service/server/businesslogic/session_manager_test.go
- ❌ internal/apps/template/service/server/businesslogic/tenant_registration_service_test.go
- ❌ internal/identity/repository/orm/test_helpers_test.go
- ❌ internal/jose/repository/elastic_jwk_gorm_repository_test.go
- ❌ internal/infra/tenant/tenant_test.go

**Database Error Path Testing (NO MOCKING NEEDED)**:
```go
func TestCreate_DuplicateKey(t *testing.T) {
    // Use real testDB with constraint violations
    id := googleUuid.NewV7()
    testRepo.Create(ctx, &Model{ID: id})

    // Real database error - no mocking needed
    err := testRepo.Create(ctx, &Model{ID: id})
    require.Error(t, err)
    require.Contains(t, err.Error(), "UNIQUE constraint")
}
```

---

## What SHOULD Have Happened

### Phase X.3.1 Example (JOSE Repositories Coverage)

**What Agent DID** (WRONG):
1. ✅ Analyze coverage gap: 82.8% → 98% (15.2 points)
2. ✅ Identify root cause: Database error paths untestable
3. ❌ Document blocker: "P2.4 GORM mocking not implemented"
4. ❌ Mark task BLOCKED
5. ❌ Continue to next task (X.5.1)
6. ❌ Stop execution without resolution

**What Agent SHOULD HAVE DONE** (CORRECT):
1. ✅ Analyze coverage gap
2. ✅ Identify root cause: per-test setupTestDB() pattern
3. ✅ Recognize TestMain pattern exists elsewhere
4. ✅ Create Phase Z: "Resolve Phase X Blockers"
5. ✅ Add Phase Z.2: "Refactor TestMain Pattern Violations"
   - Z.2.1: Refactor session_manager_test.go
   - Z.2.2: Refactor tenant_registration_service_test.go
   - Z.2.3: Refactor test_helpers_test.go
   - Z.2.4: Refactor elastic_jwk_gorm_repository_test.go
   - Z.2.5: Refactor tenant_test.go
   - Z.2.6-Z.2.10: Validation, testing, commit
6. ✅ Add Phase Z.3: "Unblock X.3.1 - JOSE Repositories Coverage"
   - Z.3.1-Z.3.9: Use real GORM DB to test database error paths
   - Target: 82.8% → 98% coverage
7. ✅ Update tasks.md with Phase Z
8. ✅ CONTINUE execution into Phase Z.2 immediately
9. ✅ Complete Z.2 → Complete Z.3 → Mark X.3.1 [x] with evidence
10. ✅ THEN proceed to X.5.1 (now unblocked by same TestMain pattern)

---

## Fix Applied

### Created Phase Z to Resolve Phase X Blockers

Created comprehensive blocker resolution phase with 36 tasks:

**Z.1: Fix Docker Desktop Dependency** (8 tasks):
- Start Docker Desktop via PowerShell
- Wait for initialization (60 seconds)
- Run cipher-im tests: `go test -v ./internal/apps/cipher/...`
- Verify TestInitDatabase_HappyPaths/PostgreSQL_Container passes
- Update README.md with Docker Desktop prerequisite
- Create pre-test verification script
- Document workaround in test comments

**Z.2: Refactor TestMain Pattern Violations** (10 tasks):
- Convert 5 packages from per-test setupTestDB() to TestMain pattern
- Expose package-level testDB variables
- Update all tests to use shared resources
- Verify test execution faster (no repeated setup overhead)
- Files: session_manager_test.go, tenant_registration_service_test.go, test_helpers_test.go, elastic_jwk_gorm_repository_test.go, tenant_test.go

**Z.3: Unblock X.3.1 - JOSE Repositories Coverage** (9 tasks):
- Run baseline: `go test -coverprofile=jose_repo_baseline.out ./internal/apps/jose/ja/repository/`
- Analyze uncovered lines: `go tool cover -func=... | grep -v "100.0%"`
- Create database error tests for each uncovered path
- Run coverage again
- Verify ≥98% coverage (target: 82.8% → 98%)
- All tests pass
- Test execution <15 seconds per package

**Z.4: Unblock X.5.1 - JOSE Services Coverage** (9 tasks):
- Run baseline: `go test -coverprofile=jose_service_baseline.out ./internal/apps/jose/ja/service/`
- Analyze uncovered lines (business logic validation ALREADY tested)
- Create database error tests (after validation succeeds, repository fails)
- Run coverage again
- Verify 95% coverage (target: 82.7%  95%)
- All tests pass

**Z.5: Complete Phase X Validation** (10 tasks):
- Complete X.2.2: Cipher-IM coverage 85%  95%
- Mark X.2.3, X.3.2, X.5.2 [x]
- Run X.6.1-X.6.5: Build, lint, test, coverage verification
- All quality gates pass

---

## Prevention Strategies

### For Future Agent Executions

1. **When P2.4 blocker identified, create Phase Z.2 immediately**:
   - NEVER defer infrastructure implementation
   - Blocked coverage tasks = infrastructure gap
   - Create resolution phase before continuing

2. **Docker Desktop prerequisite must be documented**:
   - README.md must list Docker Desktop as prerequisite
   - Tests must verify Docker running before starting
   - Create helper script: scripts/verify-docker.ps1

3. **Coverage targets require appropriate tools**:
   - 98% infrastructure = need mocking for error paths
   - Database error paths = need GORM mocking
   - Identify tool gaps early, create infrastructure phase

### For Developers

1. **Identify missing prerequisites early**:
   - Check README.md for complete prerequisite list
   - Verify all tools available before starting tests
   - Document prerequisites in DEV-SETUP.md

2. **Coverage analysis patterns**:
   - 66.7% = success + not-found covered, errors NOT covered
   - 80-90% = validation covered, DB errors NOT covered
   - >95% = comprehensive including error paths

3. **Infrastructure-first approach**:
   - When blocked on missing infrastructure, implement IMMEDIATELY
   - NEVER defer infrastructure to future phases
   - Infrastructure gaps block multiple tasks

---

## Metrics

### Before Fix

- **Phase X tasks**: 53 unchecked (X.2.1, X.2.2, X.3.1, X.5.1 blocked)
- **Phase Y tasks**: 25 unchecked (blocked on Phase X)
- **P2.4 status**: NOT IMPLEMENTED (prerequisite for X.3.1, X.5.1)
- **Coverage gaps**: JOSE repositories 82.8%, services 82.7%

### After Fix

- **Phase Z created**: 49 resolution tasks
- **Infrastructure phase**: Z.2 implements P2.4 (13 tasks)
- **Blocker resolution**: Z.3 (repositories), Z.4 (services)
- **Target coverage**: Z.3  98%, Z.4  95%

---

## Related Commits

- 3450ca43: fix(agent): add mandatory blocker resolution to plan-tasks-implement
- a5efd645: docs(tasks-v1): add Phase Z to resolve Phase X blockers
- 390b5352: style(lint): fix importas and wsl_v5 violations
- aa976e91: style(wsl): add blank lines after t.Helper() calls

---

## Cross-References

- Agent file: .github/agents/plan-tasks-implement.agent.md
- Beast mode: .github/instructions/01-02.beast-mode.instructions.md
- Database: .github/instructions/03-04.database.instructions.md
- Testing: .github/instructions/03-02.testing.instructions.md
- Task document: docs/fixes-needed-plan-tasks/tasks.md
