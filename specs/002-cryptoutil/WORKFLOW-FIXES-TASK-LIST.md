# Workflow Fixes Task List - 2025-12-20

## Critical Issues to Fix

### 1. Speckit Document Updates (HIGH PRIORITY)

#### 1.1 Update constitution.md

- [ ] Add complete 9-service table (8 product + 1 demo) to Section I
- [ ] Include all 5 Identity services (authz, idp, rs, rp, spa) with public/admin ports
- [ ] Verify service status table matches current implementation
- [ ] Add Learn-PS (Pet Store) demo service details

#### 1.2 Update spec-incomplete.md → spec.md

- [ ] Rename spec-incomplete.md to spec.md
- [ ] Add all 5 Identity services (currently missing identity-rs and identity-spa)
- [ ] Add all 3 other services (jose-ja, pki-ca, learn-ps)
- [ ] Sync service architecture section with constitution.md
- [ ] Add detailed service descriptions from copilot instructions
- [ ] Include federation patterns and service discovery mechanisms

#### 1.3 Update clarify-incomplete.md → clarify.md

- [ ] Rename clarify-incomplete.md to clarify.md
- [ ] Add Q&A for all 5 Identity services
- [ ] Add Q&A for federation between all 9 services
- [ ] Sync with updated constitution.md and spec.md
- [ ] Add clarifications from SPECKIT-CONFLICTS-ANALYSIS sessions

#### 1.4 Update PLAN-incomplete.md → PLAN.md

- [ ] Rename PLAN-incomplete.md to PLAN.md
- [ ] Add all 9 services to phase breakdown
- [ ] Update task dependencies based on all services
- [ ] Sync completion criteria with updated spec.md
- [ ] Add Learn-PS Phase 7 implementation plan

#### 1.5 Create CLARIFY-QUIZME.md

- [ ] Comprehensive A-D multiple choice questions
- [ ] Cover constitution.md gaps
- [ ] Cover spec.md gaps
- [ ] Cover clarify.md gaps
- [ ] Cover copilot instructions alignment
- [ ] Optional E write-in for all questions

### 2. Workflow Failures to Fix (IMMEDIATE PRIORITY)

#### 2.1 CI - End-to-End Testing (Run #404) - FAILURE

**URL**: <https://github.com/justincranford/cryptoutil/actions/runs/20388807383>
**Status**: Missing Identity public servers (authz, idp, rs) block E2E tests

- [ ] Investigate specific E2E test failure logs
- [ ] Fix Docker Compose service startup issues
- [ ] Verify all services can communicate
- [ ] Re-run workflow after fixes

#### 2.2 CI - DAST Security Testing (Run #414) - FAILURE

**URL**: <https://github.com/justincranford/cryptoutil/actions/runs/20388807370>
**Status**: DAST scanner can't reach services (likely due to missing public servers)

- [ ] Investigate DAST scanner connectivity
- [ ] Fix service endpoint availability
- [ ] Update DAST scan configuration if needed
- [ ] Re-run workflow after fixes

#### 2.3 CI - Load Testing (Run #393) - FAILURE

**URL**: <https://github.com/justincranford/cryptoutil/actions/runs/20388807357>
**Status**: Load tests fail due to missing service endpoints

- [ ] Investigate Gatling load test failures
- [ ] Fix service endpoint availability
- [ ] Update load test scenarios if needed
- [ ] Re-run workflow after fixes

#### 2.4 CI - Race Condition Detection (Run #370) - FAILURE

**URL**: <https://github.com/justincranford/cryptoutil/actions/runs/20388807362>
**Status**: Race detector errors OR test failures

- [ ] Get detailed race detector output
- [ ] Identify specific race conditions
- [ ] Fix concurrent access issues
- [ ] Add mutex protection where needed
- [ ] Re-run workflow after fixes

#### 2.5 CI - Mutation Testing (Run #110) - FAILURE

**URL**: <https://github.com/justincranford/cryptoutil/actions/runs/20388807354>
**Status**: Mutation testing timeout or insufficient score

- [ ] Get detailed gremlins output
- [ ] Identify packages below 85% mutation score
- [ ] Add missing test cases for uncovered mutations
- [ ] Optimize mutation testing execution time
- [ ] Re-run workflow after fixes

### 3. Workflow Fix Strategy

**Iterative Approach**:

1. Fix all speckit documents first (foundation for all work)
2. Commit and push speckit updates
3. For each failing workflow:
   a. Get detailed logs via `gh run view <run-id> --log-failed`
   b. Identify root cause
   c. Create targeted fix
   d. Commit fix
   e. Push to trigger workflow
   f. Monitor workflow run
   g. If still failing, repeat from step a
4. When all 5 workflows pass, move to next iteration

**Expected Iterations**: 2-3 per workflow (diagnosis → fix → verify)

**Total Estimated Time**: 4-6 hours for all fixes

### 4. Document Alignment Verification

After all updates:

- [ ] Verify constitution.md, spec.md, clarify.md, PLAN.md all consistent
- [ ] Verify copilot instructions align with speckit docs
- [ ] Run `git grep "identity-" | grep -E "(authz|idp|rs|rp|spa)"` to find all references
- [ ] Update any remaining inconsistent references
- [ ] Final commit with comprehensive update message

### 5. Completion Criteria

**Speckit Documents Complete**:

- ✅ constitution.md has all 9 services listed
- ✅ spec.md has all 9 services documented
- ✅ clarify.md has Q&A for all services
- ✅ PLAN.md has all services in phase breakdown
- ✅ CLARIFY-QUIZME.md created with comprehensive questions
- ✅ All -incomplete.md files renamed to proper names

**Workflows All Passing**:

- ✅ CI - End-to-End Testing (Run #404+)
- ✅ CI - DAST Security Testing (Run #414+)
- ✅ CI - Load Testing (Run #393+)
- ✅ CI - Race Condition Detection (Run #370+)
- ✅ CI - Mutation Testing (Run #110+)

**Quality Gates**:

- ✅ All git working tree clean
- ✅ All commits follow conventional commit format
- ✅ All workflows triggered and monitored
- ✅ No remaining TODOs in updated documents

---

## Execution Log

**Start Time**: 2025-12-20 (current session)

### Completed Tasks

- [x] Fetch speckit documentation from GitHub
- [x] Update copilot-instructions.md to remove duplicate accuracy/commit directives
- [x] Create this task list
- [x] Optimize .github/copilot-instructions.md and .github/instructions/* for token limits without sacrificing quality
- [x] Review constitution.md, specs/002-cryptoutil/spec.md, and specs/002-cryptoutil/PLAN.md for completeness/correctness/clarity/alignment with copilot instructions
- [x] Analyze files again and create consolidated CLARIFY-QUIZME.md with multiple-choice questions on problems/omissions/ambiguities/conflicts/risks
- [x] Create docs/WORKFLOW-TEST-GUIDELINE.md documenting local workflow testing methods and identifying gaps
- [x] Delete unnecessary session files (docs\RS-PUBLIC-SERVER-IMPLEMENTATION.md, docs\SESSION-2025-12-21-SUMMARY.md)

### In Progress Tasks

- [ ] Update constitution.md with all 9 services

### Blocked Tasks

- None (all tasks executable now)

---

## Notes

- User confirmed 5 Identity services (authz, idp, rs, rp, spa) missing from specs
- Copilot instructions (.github/instructions/01-01.architecture.instructions.md) already has correct service list
- Constitution.md Section I has service table but incomplete
- All -incomplete.md files need comprehensive updates before renaming
