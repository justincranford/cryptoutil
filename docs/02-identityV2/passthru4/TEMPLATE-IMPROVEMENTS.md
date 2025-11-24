# SDLC Feature Template Improvements

**Based On**: Analysis of Identity V2 documentation contradictions and implementation gaps
**Date**: 2025-01-XX
**Purpose**: Recommend improvements to prevent future template usage issues

---

## Executive Summary

Identity V2 implementation revealed critical gaps in the SDLC feature template that allowed:
- Agent completion claims without evidence validation
- Documentation contradictions (100% complete vs. 45% complete)
- Multiple truth sources causing confusion
- Gaps accumulating across multiple tasks without detection

This document proposes **8 major template improvements** to enforce evidence-based validation.

---

## Improvement 1: Evidence-Based Acceptance Criteria

### Current Problem
Acceptance criteria are vague, allowing subjective "complete" claims:
```markdown
- [ ] OAuth 2.1 authorization code flow functional
```

### Recommended Improvement
Add **"Evidence Required"** subsection to every acceptance criterion:

```markdown
- [ ] OAuth 2.1 authorization code flow functional
  - **Evidence Required**:
    - [ ] Zero TODO comments in handlers_authorize.go (automated scan)
    - [ ] Integration test passes: TestAuthorizationCodeFlow (CI run)
    - [ ] Manual validation: curl flow authorize → login → consent → token succeeds
    - [ ] Requirements coverage: R01-01 through R01-06 all validated
    - [ ] Code coverage: ≥85% for authorization package
```

### Template Addition Location
Section: "Quality Gates and Acceptance Criteria" → "Task-Specific Acceptance Criteria"

### Rationale
- Makes completion criteria objective and verifiable
- Prevents "looks done" vs. "actually done" discrepancies
- Provides clear checklist for validation

---

## Improvement 2: Automated Quality Gates

### Current Problem
Quality gates rely on manual checks, easy to skip:
- Agent may not run TODO scans
- Test failures may go unnoticed
- Requirements coverage unchecked

### Recommended Improvement
Add **automated quality gates** to template with specific commands:

```markdown
### Automated Quality Gates (Pre-Task Completion)

**Code Quality** (run these commands before marking task complete):
- [ ] `go build ./...` → Zero compilation errors
- [ ] `golangci-lint run ./...` → Zero linting errors
- [ ] `grep -r "TODO\|FIXME" <modified_files> | wc -l` → Zero TODOs in modified files
- [ ] `golangci-lint run --enable-only=importas ./...` → Import aliases correct

**Testing**:
- [ ] `runTests ./path/to/package` → All tests passing (0 failures)
- [ ] `go test ./... -cover | grep "coverage:"` → Coverage ≥85% infrastructure, ≥80% features
- [ ] `grep -r "t.Skip" <modified_files>` → Zero test skips without issue tracking

**Requirements** (if validation tool exists):
- [ ] `identity-requirements-check` → Coverage ≥90% for this task's requirements
- [ ] All task requirements mapped to tests (automated check)
- [ ] Acceptance criteria met: all checkboxes checked with evidence links

**Documentation**:
- [ ] README updated with user-facing changes
- [ ] OpenAPI specs synced (if API changes): `make generate-openapi-clients`
- [ ] Post-mortem created: `##-<TASK>-POSTMORTEM.md` exists
```

### Template Addition Location
Section: "Quality Gates and Acceptance Criteria" → New subsection "Automated Quality Gates"

### Rationale
- Removes subjectivity from quality validation
- Provides exact commands to run
- Makes quality gates enforceable by automation

---

## Improvement 3: Post-Mortem Corrective Action Enforcement

### Current Problem
Post-mortems identify gaps but don't enforce fixes:
- Corrective actions documented but not converted to tasks
- Gaps accumulate without resolution
- No follow-up mechanism

### Recommended Improvement
Make corrective actions **mandatory with enforcement**:

```markdown
### Corrective Action Enforcement (MANDATORY)

**For EVERY gap identified in post-mortem**:

**1. Immediate Fixes** (applied in current task):
- [ ] Fix documented in post-mortem "Immediate Corrective Actions" section
- [ ] Test added to prevent regression
- [ ] Acceptance criteria updated if new pattern discovered

**2. Deferred Fixes** (future tasks):
- [ ] **MUST create new task document**: `##.##-<GAP_NAME>.md` (NOT optional)
- [ ] **MUST add to manage_todo_list**: Specific task with clear acceptance criteria
- [ ] **MUST add to dependency graph**: Update MASTER-PLAN with new task dependencies

**3. Pattern Improvements** (affect all future tasks):
- [ ] Pattern documented in LESSONS-LEARNED.md
- [ ] Template/instructions updated (if applicable)
- [ ] Code review checklist updated

**Validation**: Task is NOT complete until:
- ✅ All immediate fixes verified in code
- ✅ All deferred fixes converted to new task documents
- ✅ All pattern improvements documented and applied

**Example Violation**: Identifying "client secret plain text comparison" gap but NOT creating follow-up task.
```

### Template Addition Location
Section: "Post-Mortem and Corrective Actions" → Add new "Corrective Action Enforcement" subsection

### Rationale
- Ensures gaps don't accumulate
- Makes post-mortem actionable, not just documentation
- Prevents "known issues" from lingering unaddressed

---

## Improvement 4: Single Source of Truth Documentation Pattern

### Current Problem
Multiple conflicting status documents:
- MASTER-PLAN.md: "100% complete"
- README.md: "45% complete"
- STATUS-REPORT.md: Evidence-based gaps
Users don't know which to trust.

### Recommended Improvement
Enforce **single source of truth (SSOT)** pattern:

```markdown
### Single Source of Truth (SSOT) Documentation Pattern

**Primary Status Document**: `PROJECT-STATUS.md`
- **Purpose**: ONLY authoritative source for project status
- **Location**: `docs/<FEATURE_ID>/PROJECT-STATUS.md`
- **Mandatory Sections**:
  1. Current Status: ❌ NOT READY | ⚠️ CONDITIONAL | ✅ PRODUCTION READY
  2. Completion Metrics: X/Y tasks complete, Z% requirements coverage, W TODOs remaining
  3. Known Limitations: Documented gaps, deferred features (with issue/task references)
  4. Production Blockers: Critical gaps preventing deployment
  5. Last Updated: Timestamp, commit hash, responsible agent/person

**All Other Documents REFERENCE the SSOT**:
- MASTER-PLAN.md: "See PROJECT-STATUS.md for current status"
- README.md: "See PROJECT-STATUS.md for completion metrics"
- Task docs: "Update PROJECT-STATUS.md when this task completes"

**Update Triggers** (automatic):
- After every task completion (via manage_todo_list)
- After every requirements validation run
- After every TODO scan
- Before any "production ready" claim

**Enforcement**: CI/CD fails if PROJECT-STATUS.md last-updated > 24 hours old.
```

### Template Addition Location
Section: "Documentation" → New subsection "Single Source of Truth Pattern"

### Rationale
- Eliminates confusion about "which document is correct?"
- Provides single update point for status
- Makes status tracking consistent across all features

---

## Improvement 5: Progressive Validation Pattern

### Current Problem
Gaps accumulate across multiple tasks:
- Agent completes Task 01-05 without validation
- TODOs, test failures, coverage regressions pile up
- Issues discovered only at end during final verification

### Recommended Improvement
Add **progressive validation after every task**:

```markdown
### Progressive Validation (After Each Task)

**Validation Checklist** (run after EVERY task completion, before starting next):

1. **TODO Scan**:
   - Command: `grep -r "TODO\|FIXME" <package> | tee todo-scan.txt`
   - Requirement: Zero new TODOs introduced (compare with baseline)
   - Action: Document any new TODOs in post-mortem, create follow-up tasks

2. **Test Run**:
   - Command: `runTests ./path/to/package`
   - Requirement: All tests passing (0 failures)
   - Action: Fix test failures before proceeding

3. **Coverage Check**:
   - Command: `go test ./... -cover -coverprofile=coverage.out`
   - Requirement: Coverage maintained or improved (not regressed)
   - Action: Add tests if coverage dropped

4. **Requirements Validation**:
   - Command: `identity-requirements-check` (or equivalent)
   - Requirement: Coverage maintained/improved from previous task
   - Action: Map new functionality to requirements

5. **Integration Smoke Test**:
   - Command: Run E2E flow for affected component
   - Requirement: Core flow still works end-to-end
   - Action: Fix broken integrations immediately

6. **Documentation Sync**:
   - Command: Update PROJECT-STATUS.md with latest metrics
   - Requirement: Status doc reflects current reality
   - Action: Commit updated status with task completion

**Quality Gate**: Task is NOT complete until all 6 validation checks pass.

**Rationale**: Catch issues incrementally instead of accumulating debt across multiple tasks.

**Example Violation**: Completing Tasks 01-05 without running integration tests, discovering OAuth flow broken in Task 06.
```

### Template Addition Location
Section: "Task Execution Instructions" → New subsection "Progressive Validation"

### Rationale
- Prevents gap accumulation
- Catches regressions early when context is fresh
- Maintains quality continuously instead of "big bang" validation at end

---

## Improvement 6: Foundation-Before-Features Enforcement

### Current Problem
Identity V2 implemented advanced features (MFA, WebAuthn) before foundation (OAuth flows) worked:
- Result: Production-ready advanced features on broken foundation
- Wasted effort implementing features that can't be used

### Recommended Improvement
Add **strict phase ordering with dependency checks**:

```markdown
### Foundation-Before-Features Pattern (STRICT ENFORCEMENT)

**Phase Ordering** (MUST complete in sequence):

**Phase 1: Foundation** (MUST complete before Phase 2)
- Domain models
- Database schema and migrations
- Repository layer (CRUD operations)
- **Exit Criteria**:
  - [ ] All repository CRUD operations work
  - [ ] Unit tests pass (≥85% coverage)
  - [ ] Integration tests pass (database operations validated)
  - [ ] Zero TODOs in foundation packages
- **Blocked Tasks**: Phase 2/3 tasks CANNOT start until Phase 1 exit criteria met

**Phase 2: Core Features** (MUST complete before Phase 3)
- Business logic layer
- API endpoints (HTTP handlers)
- Authentication/authorization flows
- **Exit Criteria**:
  - [ ] Core flows work end-to-end (e.g., OAuth authorization code flow)
  - [ ] Integration tests pass (full flow validation)
  - [ ] API endpoints functional (Swagger UI accessible)
  - [ ] Zero CRITICAL/HIGH TODOs
- **Blocked Tasks**: Phase 3 tasks CANNOT start until Phase 2 exit criteria met

**Phase 3: Advanced Features** (ONLY after Phase 1+2 complete)
- MFA, WebAuthn, hardware credentials, adaptive authentication, etc.
- **Exit Criteria**:
  - [ ] Advanced features work ON TOP OF solid foundation
  - [ ] Integration with Phase 2 core flows validated
  - [ ] Performance requirements met

**Enforcement**:
- Add dependency checks to each task acceptance criteria:
  - [ ] All Phase 1 tasks complete (if this is Phase 2 task)
  - [ ] All Phase 2 tasks complete (if this is Phase 3 task)
- CI/CD fails if attempting to merge Phase 3 PR before Phase 2 complete

**Violation Example**: Identity V2 implemented MFA/WebAuthn (Phase 3) before OAuth flows (Phase 2) worked correctly.

**Consequence**: Advanced features unusable because foundation broken; rework required.
```

### Template Addition Location
Section: "Implementation Phases" → Replace existing phase structure with strict ordering

### Rationale
- Prevents building on unstable foundation
- Ensures advanced features have working base to integrate with
- Avoids rework when foundation issues discovered late

---

## Improvement 7: Evidence-Based Task Completion Checklist

### Current Problem
Task completion is subjective:
- Agent claims "complete" without objective validation
- No standard checklist for what "complete" means
- Easy to skip validation steps

### Recommended Improvement
Add **mandatory evidence-based completion checklist**:

```markdown
### Evidence-Based Task Completion Checklist

**Before marking any task complete, ALL of these must be TRUE**:

**Code Evidence**:
- [ ] Zero compilation errors: `go build ./...` output clean
- [ ] Zero linting errors: `golangci-lint run ./...` output clean
- [ ] Zero TODOs in task files: `grep -r "TODO\|FIXME" <task_files>` shows 0 results
- [ ] Code coverage met: Coverage report shows ≥85% (infrastructure) or ≥80% (features)

**Test Evidence**:
- [ ] All tests pass: `runTests ./...` shows "PASS" (no failures)
- [ ] Integration tests pass: E2E flow validated (manual or automated)
- [ ] Performance benchmarks: No regressions (if applicable)
- [ ] Load tests: Meets requirements (if applicable)

**Requirements Evidence**:
- [ ] Requirements coverage: `identity-requirements-check` shows ≥90% for this task
- [ ] Acceptance criteria met: All checkboxes checked with evidence links
- [ ] Task deliverables: All promised deliverables exist and functional

**Documentation Evidence**:
- [ ] Post-mortem created: `##-<TASK>-POSTMORTEM.md` exists
- [ ] Corrective actions: All gaps converted to new tasks OR immediate fixes
- [ ] PROJECT-STATUS.md updated: Latest metrics, known limitations, blockers
- [ ] OpenAPI synced: Specs match implementation (if API changes)

**Git Evidence**:
- [ ] Commit message: Conventional commit format `type(scope): description`
- [ ] Commit includes: All modified files staged and committed
- [ ] Branch clean: `git status` shows working tree clean

**Quality Gate**: If ANY checkbox is unchecked, task is NOT complete.

**Enforcement**: Automated script checks these items before allowing task completion.
```

### Template Addition Location
Section: "Task Execution Instructions" → New subsection "Evidence-Based Task Completion Checklist"

### Rationale
- Removes subjectivity from completion decisions
- Provides objective, verifiable completion criteria
- Makes it impossible to claim completion without evidence

---

## Improvement 8: Requirements Coverage Threshold Enforcement

### Current Problem
Requirements coverage can drift low without detection:
- Identity V2: 58.5% coverage (38/65 requirements)
- No enforcement mechanism
- Agent claims "complete" despite uncovered requirements

### Recommended Improvement
Add **hard requirements coverage threshold**:

```markdown
### Requirements Coverage Threshold (MANDATORY)

**Per-Task Threshold**:
- Every task MUST validate ≥90% of its assigned requirements
- Requirements validation tool: `identity-requirements-check` (or equivalent)
- Run after every task completion

**Overall Threshold**:
- Project MUST maintain ≥85% overall requirements coverage
- Check before any "production ready" claim
- Generate coverage report in PROJECT-STATUS.md

**Enforcement**:
```bash
# Per-task enforcement
coverage=$(identity-requirements-check --task R01 --format json | jq '.coverage_percentage')
if (( $(echo "$coverage < 90" | bc -l) )); then
  echo "ERROR: Requirements coverage $coverage% < 90% threshold"
  exit 1
fi

# Overall enforcement (before production claim)
coverage=$(identity-requirements-check --format json | jq '.overall_coverage_percentage')
if (( $(echo "$coverage < 85" | bc -l) )); then
  echo "ERROR: Overall coverage $coverage% < 85% threshold"
  exit 1
fi
```

**Acceptance Criteria Addition**:
- [ ] Requirements coverage: ≥90% for this task (verified via automated check)
- [ ] Overall coverage maintained: ≥85% (no regression from previous task)
- [ ] Uncovered requirements documented: If <100%, list uncovered requirements with justification

**CI/CD Integration**:
- Pre-commit hook: Warn if coverage drops
- PR checks: Fail if coverage < threshold
- Production deployment gate: Block if overall coverage < 85%

**Example Violation**: Identity V2 with 58.5% coverage allowed to claim "production ready".
```

### Template Addition Location
Section: "Quality Gates and Acceptance Criteria" → New subsection "Requirements Coverage Threshold"

### Rationale
- Prevents incomplete implementations from claiming completion
- Enforces evidence-based validation
- Provides objective metric for production readiness

---

## Summary of Template Improvements

| Improvement | Problem Solved | Template Section | Enforcement |
|-------------|----------------|------------------|-------------|
| 1. Evidence-Based Acceptance Criteria | Vague completion criteria | Quality Gates | Checklist with evidence links |
| 2. Automated Quality Gates | Manual validation skipped | Quality Gates | Automated commands |
| 3. Post-Mortem Enforcement | Gaps not converted to tasks | Post-Mortem | Mandatory task doc creation |
| 4. Single Source of Truth | Conflicting status docs | Documentation | SSOT pattern (PROJECT-STATUS.md) |
| 5. Progressive Validation | Gap accumulation | Task Execution | After-each-task validation |
| 6. Foundation-Before-Features | Advanced features on broken base | Implementation Phases | Strict phase ordering |
| 7. Evidence-Based Completion | Subjective completion claims | Task Execution | Mandatory checklist |
| 8. Requirements Coverage Threshold | Low coverage accepted | Quality Gates | 90% per-task, 85% overall |

---

## Implementation Recommendations

### Immediate (Apply to Passthru4 Plan)
1. Use all 8 improvements in new MASTER-PLAN-V4.md
2. Add evidence requirements to every acceptance criterion
3. Include automated quality gate commands
4. Create PROJECT-STATUS.md as SSOT

### Short-Term (Next Feature Implementation)
5. Create requirements validation tool if not exists
6. Add automated TODO scanning to pre-commit
7. Implement progressive validation script
8. Add requirements coverage threshold to CI/CD

### Long-Term (Template Standardization)
9. Update feature-template.md with all 8 improvements
10. Create validation script for template compliance
11. Document lessons learned from Identity V2
12. Share improvements with team

---

## Conclusion

Identity V2 revealed critical gaps in SDLC template enforcement. These 8 improvements transform the template from **descriptive guidelines** to **prescriptive, enforceable requirements**.

**Key Principle**: Evidence-based validation at every step, with objective criteria and automated enforcement.

**Impact**: Prevents future projects from repeating Identity V2's pattern of claiming completion without evidence-based validation.

---

**Improvements Generated**: 2025-01-XX
**Based On**: Identity V2 gap analysis and documentation contradiction review
**Next Step**: Apply improvements to passthru4 MASTER-PLAN-V4.md
