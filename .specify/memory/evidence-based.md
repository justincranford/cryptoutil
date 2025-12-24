# Evidence-Based Task Completion - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/06-01.evidence-based.instructions.md`

## Core Principle - CRITICAL

**NEVER mark tasks complete without objective evidence**

**Why This Matters**:

- Prevents premature completion claims
- Ensures quality gates are met
- Enables reproducible validation
- Builds trust in completion status
- Catches regressions early

---

## Mandatory Evidence Checklist

### Code Quality Evidence

**Build Clean**:

```bash
go build ./...
# Expected: No errors, all packages compile successfully
```

**Linting Clean**:

```bash
golangci-lint run
# Expected: No warnings or errors, all linters pass
```

**No New TODOs**:

```bash
grep -r "TODO\|FIXME" <package>
# Expected: 0 new TODOs compared to baseline
```

**Coverage ≥95% (Production)**:

```bash
go test ./internal/kms -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# Expected: total: (statements) ≥95.0%
```

**Coverage ≥98% (Infrastructure/Utility)**:

```bash
go test ./internal/shared/magic -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# Expected: total: (statements) ≥98.0%
```

### Test Quality Evidence

**All Tests Passing**:

```bash
go test ./...
# Expected: PASS for all packages, 0 failures
```

**No Skipped Tests Without Tracking**:

```bash
grep -r "t.Skip\|t.SkipNow" <package>
# Expected: All skips documented in DETAILED.md with reason + timeline
```

**Mutation Score ≥80%**:

```bash
gremlins unleash
# Expected: Killed ≥80% for early phases, ≥98% for infrastructure/utility
```

### Git Quality Evidence

**Conventional Commit Format**:

```bash
git log -1 --pretty=%B
# Expected: Matches pattern: type(scope): description
```

**Clean Working Tree**:

```bash
git status --short
# Expected: No uncommitted changes (or documented in-progress work)
```

---

## Progressive Validation (After Every Task)

**Execute these steps IMMEDIATELY after completing each task**:

### Step 1: TODO Scan

```bash
# Check for new TODOs
grep -r "TODO\|FIXME" <package> | wc -l
# Compare with baseline count
```

**Pass Criteria**: 0 new TODOs introduced by task

### Step 2: Test Run

```bash
# Run tests for affected package
go test ./path/to/package -v
```

**Pass Criteria**: All tests passing, 0 failures

### Step 3: Coverage Check

```bash
# Generate coverage report
go test ./path/to/package -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

**Pass Criteria**:

- Production packages: ≥95%
- Infrastructure/utility: ≥98%

### Step 4: Mutation Testing

```bash
# Run gremlins on package
gremlins unleash --tags="~integration,~e2e"
```

**Pass Criteria**:

- Early phases: ≥80% mutation score
- Infrastructure/utility: ≥98% mutation score

### Step 5: Integration Test

```bash
# Run E2E smoke test
docker compose -f deployments/compose/compose.yml up -d
autoapprove curl -k https://127.0.0.1:8080/admin/v1/livez
```

**Pass Criteria**: Core E2E workflow functional

### Step 6: Documentation Update

**Update `specs/*/implement/DETAILED.md` Section 2 timeline**:

```markdown
### YYYY-MM-DD: Task Description
- Work completed: [commit hash] description
- Coverage: Before X% → After Y%
- Mutation score: Before X% → After Y%
- Key findings: Any discoveries or blockers
- Violations found: Any anti-patterns discovered
```

**Pass Criteria**: Timeline entry added with metrics

---

## Quality Gate - MANDATORY

**Task is NOT complete until ALL checks pass**:

1. ✅ Build clean (`go build ./...`)
2. ✅ Linting clean (`golangci-lint run`)
3. ✅ No new TODOs (baseline maintained)
4. ✅ Tests passing (0 failures)
5. ✅ Coverage ≥95% production, ≥98% infrastructure/utility
6. ✅ Mutation score ≥80% early phases, ≥98% infrastructure/utility
7. ✅ Integration test passing
8. ✅ Documentation updated (DETAILED.md Section 2)

**NO EXCEPTIONS**: If any check fails, task is incomplete

---

## Single Source of Truth - CRITICAL

### implement/DETAILED.md

**Authoritative status source for spec kit iterations**:

**Section 1: Task Checklist**:

- Maintains TASKS.md order for cross-reference
- Status symbols: ❌ (not started), ⚠️ (in progress), ✅ (complete)
- Each task links to completion evidence (commit hashes, coverage %, mutation %)

**Section 2: Append-Only Timeline**:

- Time-ordered implementation log
- Each entry: Date, task description, metrics, findings, violations, next steps
- NEVER delete entries (append-only for audit trail)

**Example**:

```markdown
## Section 1: Task Status

- ✅ Task 1.1: Domain models ← Commit abc1234, Coverage 96%, Mutation 82%
- ⚠️ Task 1.2: Database schema ← In progress, blocked by migration tooling
- ❌ Task 1.3: CRUD operations ← Not started, depends on Task 1.2

## Section 2: Timeline

### 2025-12-20: Domain Models Implementation
- Completed: Task 1.1 (commit abc1234)
- Coverage: 91% → 96% (+5 percentage points)
- Mutation score: 78% → 82% (+4 percentage points)
- Key findings: Needed custom NullableUUID type for cross-DB compatibility
- Violations found: None
- Next steps: Unblock Task 1.2 (research migration tooling)

### 2025-12-21: Database Schema Research
- Work completed: Evaluated golang-migrate, goose, Atlas
- Key findings: golang-migrate best fit (version control, rollback support)
- Next steps: Implement Task 1.2 with golang-migrate
```

### implement/EXECUTIVE.md

**Stakeholder-facing status summary**:

**Sections**:

1. **Stakeholder Overview**: Product status, key achievements, high-level roadmap
2. **Customer Demonstrability**: Docker commands, E2E demos, video walkthroughs
3. **Risk Tracking**: Known issues, limitations, missing features with severity
4. **Post Mortem**: Lessons learned, suggestions for future improvements
5. **Last Updated**: Timestamp of most recent update

**Update Frequency**: After each phase completion, major milestone, or stakeholder request

---

## Phase Dependencies - Strict Sequence

**MANDATORY: Complete phases in order, NO skipping ahead**

### Phase 1: Foundation

- **Tasks**: Domain models, database schema, CRUD operations
- **Completion Criteria**: ≥95% coverage, ≥80% mutation, zero TODOs
- **Blockers**: None (can start immediately)

### Phase 2: Core Logic

- **Tasks**: Business logic, API endpoints, authentication/authorization
- **Completion Criteria**: E2E works, zero CRITICAL TODOs
- **Blockers**: Phase 1 MUST be 100% complete

### Phase 3: Advanced Features

- **Tasks**: MFA, WebAuthn, federation, advanced security
- **Completion Criteria**: Full feature parity, ≥98% coverage, ≥98% mutation
- **Blockers**: Phase 1+2 MUST be 100% complete

**Rationale**:

- Foundation bugs cascade into later phases
- Refactoring Phase 1 breaks Phase 2+3 implementations
- Quality gates prevent technical debt accumulation

---

## Post-Mortem Enforcement - MANDATORY

**Every gap discovered → Immediate action required**

### Gap Response Pattern

**Option 1: Immediate Fix** (< 30 minutes):

- Fix gap in current session
- Update DETAILED.md Section 2 with fix details
- Commit with reference to gap

**Option 2: New Task Doc** (> 30 minutes):

- Create `specs/*/##.##-GAP_NAME.md` task document
- Document gap, impact, fix approach, completion criteria
- Link from DETAILED.md Section 1 task list
- Schedule for future sprint/iteration

**NEVER**: Leave gaps undocumented or untracked

### Post-Mortem Template

**File**: `docs/P0.X-INCIDENT_NAME.md` (for critical issues)

**Sections**:

1. **Incident Summary**: What happened, when, severity
2. **Root Cause Analysis**: Why it happened (5 Whys technique)
3. **Timeline**: Event sequence with timestamps
4. **Impact**: Affected systems, users, data
5. **Lessons Learned**: What worked, what didn't
6. **Action Items**: Preventive measures with owners and deadlines
7. **References**: Commits, PRs, related documentation

---

## Common Violations

### ❌ Violations (NEVER DO)

**Mark complete without validation**:

- Claiming task done without running tests
- Skipping coverage/mutation checks
- Assuming code compiles without verification

**Skip post-mortem**:

- Finding gap but not documenting
- "Will fix later" without creating task doc
- Ignoring failed quality gates

**Phase 3 before Phase 2**:

- Implementing advanced features before core logic
- Building on unstable foundation
- Skipping prerequisite phases

### ✅ Correct Pattern (ALWAYS DO)

**Run all checks**:

- Execute full quality gate checklist
- Verify every metric meets threshold
- No manual overrides or exceptions

**Create post-mortems**:

- Document every gap discovered
- Create task docs for future work
- Link from DETAILED.md for traceability

**Respect phase dependencies**:

- Complete Phase 1 before Phase 2
- Complete Phase 2 before Phase 3
- Maintain strict sequence

---

## Cross-References

**Related Documentation**:

- Testing standards: `.specify/memory/testing.md`
- Git workflow: `.specify/memory/git.md`
- Coverage requirements: `.specify/memory/testing.md`
- Mutation testing: `.specify/memory/testing.md`
