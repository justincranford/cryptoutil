# Identity V2 Master Remediation Plan - Passthru4

**Plan Date**: November 24, 2025
**Status**: ⏳ IN PROGRESS - 0/8 tasks complete (0%)
**Progress**: Evidence-based remediation with validation gates
**Goal**: Production-ready OAuth 2.1 / OIDC identity platform with verified completeness

---

## Executive Summary

### Current Status: ❌ NOT READY FOR PRODUCTION

**Context**: Passthru3 remediation claimed "100% complete" but evidence showed:

- Original plan (Tasks 01-20): 45% complete
- Requirements coverage: 58.5% (38/65 validated)
- TODO comments: 37 found (0 CRITICAL, 4 HIGH, 12 MEDIUM, 21 LOW)
- Production blockers: 8 identified

**This Plan (Passthru4)**: Address ALL identified gaps with evidence-based validation gates.

### Completion Metrics

| Status | Tasks | Percentage | Details |
|--------|-------|------------|---------|
| ✅ Complete | 0 | 0% | None yet |
| ⏳ In Progress | 0 | 0% | Starting now |
| 🔜 Pending | 8 | 100% | All tasks pending |

**Target Completion**: 8/8 tasks (100%) with ≥90% requirements coverage

### Production Readiness Targets

**Foundation Layer** (MUST complete):

- OAuth 2.1 authorization code flow (complete, verified)
- OIDC core endpoints (login, consent, logout, userinfo) functional
- Token-user association using real user IDs (no placeholders)
- Client secret hashing (PBKDF2-HMAC-SHA256)
- Token lifecycle cleanup jobs
- CRL/OCSP revocation checking
- Requirements coverage ≥90%
- Zero CRITICAL/HIGH TODOs

**Validation Gates** (enforced at each task):

- ✅ Zero TODO comments in modified files
- ✅ All tests passing (runTests)
- ✅ Requirements coverage ≥90% for task
- ✅ Integration tests validate end-to-end flows
- ✅ Post-mortem created with evidence

---

## 🤖 LLM Agent Quick Reference

### For Autonomous Implementation Sessions

**READ THIS SECTION BEFORE STARTING WORK**

### Primary Directive: Continuous Work Until Complete

**Token Budget**: Work until 950k/1M tokens used (95% utilization)

**Stop Conditions**: ONLY when:

1. Tokens ≥950,000 (95% of budget), OR
2. User explicitly says "stop"

**Not Stop Conditions** (these are NOT reasons to stop):

- ❌ Time elapsed
- ❌ Tasks complete (unless ALL tasks + validation complete)
- ❌ Commits made (commits are checkpoints, not endpoints)
- ❌ "Made good progress"
- ❌ Analysis documents created

**Pattern**: tool call → IMMEDIATELY invoke next tool → tool call → repeat

**ZERO TEXT between tool calls** - no summaries, no status updates, no announcements

### Token Budget Math - CRITICAL

**Formula**: `Percentage Used = (Tokens Used / 1,000,000) × 100`

**Examples**:

- 10,000 used → 1.0% used → KEEP WORKING ✅
- 75,000 used → 7.5% used → KEEP WORKING ✅
- 100,000 used → 10.0% used → KEEP WORKING ✅
- 500,000 used → 50.0% used → KEEP WORKING ✅
- 940,000 used → 94.0% used → KEEP WORKING ✅
- 950,000 used → 95.0% used → STOP ❌

### FIPS 140-3 Compliance - CRITICAL

**Password Hashing**: MUST use PBKDF2-HMAC-SHA256 (FIPS-approved)

**BANNED Algorithms** (NEVER use):

- ❌ bcrypt, scrypt, Argon2 (password hashing)
- ❌ MD5, SHA-1 (digests)

### Test Value Policy - CRITICAL

**Option A**: Magic values from `internal/identity/magic` package

```go
username: identityMagic.TestUsername
```

**Option B**: Random values generated at runtime (generate once, reuse)

```go
id := googleUuid.NewV7()  // Generate once
sessionID: id,            // Reuse in test
```

**NEVER**:

- ❌ Hardcoded UUIDs
- ❌ Generating random value twice expecting same result

### Evidence-Based Task Completion - MANDATORY

**Before marking ANY task complete, ALL of these MUST be TRUE**:

**Code Evidence**:

- [ ] Zero compilation errors: `go build ./...`
- [ ] Zero linting errors: `golangci-lint run ./...`
- [ ] Zero TODOs in task files: `grep -r "TODO\|FIXME" <files>` = 0
- [ ] Coverage met: ≥85% (infrastructure) or ≥80% (features)

**Test Evidence**:

- [ ] All tests pass: `runTests ./path` = PASS (0 failures)
- [ ] Integration tests pass: E2E flow validated
- [ ] Manual validation: curl/Swagger UI test successful

**Requirements Evidence**:

- [ ] Requirements coverage: ≥90% for this task
- [ ] Acceptance criteria: All checkboxes checked with evidence
- [ ] Task deliverables: All promised items exist and functional

**Documentation Evidence**:

- [ ] Post-mortem created: `RXX-POSTMORTEM.md` exists
- [ ] Corrective actions: All gaps converted to tasks OR immediate fixes
- [ ] PROJECT-STATUS.md updated: Latest metrics, limitations, blockers

**Git Evidence**:

- [ ] Commit message: Conventional commit format
- [ ] All files staged: `git status` clean
- [ ] No uncommitted changes

### Per-Task Loop (For EACH task P4.01-P4.08)

1. **Pre-Implementation** (5 min)
   - Read task acceptance criteria
   - Review related code from passthru3
   - Understand validation gates

2. **Implementation** (varies)
   - Create/modify files per deliverables
   - Follow coding standards (.github/instructions/*.md)
   - Run progressive validation after each file

3. **Testing** (10-15 min)
   - Write table-driven tests with `t.Parallel()`
   - Test happy + sad paths + edge cases
   - Run: `runTests` tool (NEVER `go test`)
   - Achieve ≥85% coverage

4. **Quality** (10 min)
   - Auto-fix: `golangci-lint run --fix`
   - Fix remaining issues manually
   - Verify: `grep -r "TODO\|FIXME" <files>` = 0

5. **Validation** (10 min)
   - Run integration tests: validate E2E flow
   - Run requirements check: coverage ≥90%
   - Manual test: curl or Swagger UI
   - Update PROJECT-STATUS.md

6. **Commit** (2 min)
   - Stage: `git add <files>`
   - Commit: `git commit --no-verify -m "feat(identity): P4.XX - description"`

7. **Post-Mortem** (10 min)
   - Create: `P4.XX-POSTMORTEM.md`
   - Document bugs/fixes, omissions, corrective actions
   - Convert gaps to new tasks (if needed)

8. **Handoff** (0 min - IMMEDIATE)
   - Mark complete: `manage_todo_list`
   - **IMMEDIATELY** read next task
   - **NO STOPPING, NO SUMMARY**

### Anti-Patterns to Avoid

**NEVER**:

- ❌ Stop after commits
- ❌ Provide status updates between tasks
- ❌ Ask "Should I continue?"
- ❌ Use `go test` (use `runTests` tool)
- ❌ Remove `t.Parallel()` to fix failures
- ❌ Create analysis docs without implementing fixes
- ❌ Stop after creating 3-5 files

**ALWAYS**:

- ✅ Tool calls only (zero text between)
- ✅ Work continuously until 950k tokens OR all tasks complete
- ✅ Create post-mortem for EVERY task
- ✅ Fix failing tests before moving on
- ✅ Implement fixes immediately after analysis

---

## Foundation-First Task Ordering

**Phase 1: Core OAuth/OIDC Foundation** (MUST complete before Phase 2)

- P4.01: Fix 8 production blockers (login UI, consent, logout, userinfo, token association, secrets, cleanup, revocation)
- P4.02: Increase requirements coverage to ≥90%
- P4.03: Resolve all HIGH severity TODOs (4 items)
- **Exit Criteria**: Core flows work, 0 HIGH TODOs, ≥90% coverage, 8 blockers resolved

**Phase 2: Quality and Testing** (MUST complete before Phase 3)

- P4.04: Achieve ≥85% test coverage
- P4.05: Zero test failures (all tests passing)
- P4.06: OpenAPI synchronization (specs match implementation)
- **Exit Criteria**: Tests pass, coverage ≥85%, OpenAPI synced

**Phase 3: Production Readiness** (Only after Phase 1+2)

- P4.07: Resolve MEDIUM TODOs (12 items - structured logging, auth profiles)
- P4.08: Final verification with production readiness checklist
- **Exit Criteria**: Production ready, all quality gates passed, documentation complete

---

## Remediation Tasks

### 🔴 PHASE 1: Core Foundation (Days 1-3)

#### P4.01: Fix 8 Production Blockers

**Priority**: 🔴 CRITICAL
**Effort**: 2 days (16 hours)
**Dependencies**: None
**Files**: Multiple (see deliverables)
**Status**: 🔜 PENDING

**Context**: From GAP-ANALYSIS.md - 8 blockers prevent production deployment

**Objectives**:

1. Implement login/consent/logout/userinfo UI (4 blockers)
2. Fix token-user association (1 blocker)
3. Implement client secret hashing (1 blocker)
4. Add token lifecycle cleanup (1 blocker)
5. Implement CRL/OCSP checking (1 blocker)

**Deliverables**:

**D1.1: Login UI Implementation** (4 hours)

- File: `internal/identity/idp/handlers_login.go`
- Implement HTML login form (not JSON response)
- CSRF protection
- Session creation on successful authentication
- Redirect to consent with request context
- **Evidence**: Manual test shows HTML form rendered
- **Blocker Resolved**: BLOCK-01 (no login UI)

**D1.2: Consent UI Implementation** (4 hours)

- File: `internal/identity/idp/handlers_consent.go`
- Fetch client details from repository
- Render consent page with scopes and client info
- Store consent decision
- Generate authorization code
- Redirect to client callback
- **Evidence**: Manual test shows consent page rendered
- **Blocker Resolved**: BLOCK-02 (no consent UI)

**D1.3: Logout Implementation** (2 hours)

- File: `internal/identity/idp/handlers_logout.go`
- Validate session exists
- Revoke tokens for session
- Delete session from repository
- Clear session cookie
- Redirect to logout confirmation
- **Evidence**: Manual test shows logout works
- **Blocker Resolved**: BLOCK-03 (logout doesn't work)

**D1.4: Userinfo Endpoint** (2 hours)

- File: `internal/identity/idp/handlers_userinfo.go`
- Extract token from Authorization header
- Introspect token (validate, check expiration)
- Fetch user from repository
- Map user to OIDC claims
- Return JSON response
- **Evidence**: curl test shows user profile returned
- **Blocker Resolved**: BLOCK-04 (userinfo non-functional)

**D1.5: Token-User Association Fix** (1 hour)

- File: `internal/identity/authz/handlers_token.go`
- Remove `userIDPlaceholder := googleUuid.Must(googleUuid.NewV7())`
- Use `authRequest.UserID` instead
- Validate user ID exists
- **Evidence**: Integration test shows real user ID in token
- **Blocker Resolved**: BLOCK-05 (token uses placeholder)

**D1.6: Client Secret Hashing** (2 hours)

- File: `internal/identity/authz/clientauth/secret_hash.go`, `basic.go`, `post.go`
- Implement PBKDF2-HMAC-SHA256 hashing
- Update basic.go/post.go to use `CompareSecret()`
- Add migration to hash existing secrets
- **Evidence**: Test shows hashed secret validation works
- **Blocker Resolved**: BLOCK-06 (plain text secrets)

**D1.7: Token Cleanup Jobs** (1 hour)

- File: `internal/identity/authz/service/cleanup.go`
- Implement background job to delete expired tokens
- Run every 1 hour (configurable)
- Add graceful shutdown
- **Evidence**: Test shows expired tokens deleted
- **Blocker Resolved**: BLOCK-07 (no cleanup jobs)

**D1.8: CRL/OCSP Revocation Checking** (2 hours)

- File: `internal/identity/authz/clientauth/cert_validation.go`
- Check CRL distribution points
- Check OCSP responder
- Cache revocation results (1 hour TTL)
- **Evidence**: Test shows revoked cert rejected
- **Blocker Resolved**: BLOCK-08 (no revocation checking)

**Acceptance Criteria**:

- [ ] Login UI renders HTML form (not JSON)
- [ ] Consent UI renders with client/scope details
- [ ] Logout revokes tokens and clears session
- [ ] Userinfo returns authenticated user profile
- [ ] Tokens use real user IDs (no placeholders)
- [ ] Client secrets hashed with PBKDF2-HMAC-SHA256
- [ ] Token cleanup job runs every hour
- [ ] CRL/OCSP revocation checking functional
- [ ] All 8 production blockers resolved
- [ ] Zero TODO comments in modified files
- [ ] All tests passing (runTests)
- [ ] Integration test validates E2E flow
- [ ] Post-mortem created: `P4.01-POSTMORTEM.md`

**Validation Commands**:

```bash
# Zero TODOs in modified files
grep -r "TODO\|FIXME" internal/identity/idp/handlers_{login,consent,logout,userinfo}.go
grep -r "TODO\|FIXME" internal/identity/authz/handlers_token.go
grep -r "TODO\|FIXME" internal/identity/authz/clientauth/

# All tests passing
runTests ./internal/identity/idp
runTests ./internal/identity/authz

# Requirements coverage
identity-requirements-check --task P4.01
```

---

#### P4.02: Increase Requirements Coverage to ≥90%

**Priority**: 🔴 CRITICAL
**Effort**: 1 day (8 hours)
**Dependencies**: P4.01 (blockers fixed)
**Files**: Tests across identity packages
**Status**: 🔜 PENDING

**Context**: Current coverage 58.5% (38/65), need ≥90% (59/65)

**Objectives**:

1. Identify 27 uncovered requirements
2. Create tests/implementations for each
3. Run validation to confirm ≥90%

**Deliverables**:

**D2.1: Requirements Gap Analysis** (1 hour)

- Run: `identity-requirements-check --uncovered-only`
- Categorize by package (OIDC, client auth, token lifecycle)
- Prioritize CRITICAL (7) and HIGH (13) uncovered requirements

**D2.2: OIDC Requirements** (3 hours)

- R02-03: Discovery endpoint (metadata JSON)
- R02-01: UserInfo endpoint (DONE in P4.01)
- R02-06: Discovery metadata completeness
- R02-05: UserInfo response claims
- R02-07: Integration tests for OIDC
- **Evidence**: 5/7 R02 requirements validated

**D2.3: Token Lifecycle Requirements** (2 hours)

- R05-06: Token expiration enforcement
- R05-04: Token revocation endpoint
- **Evidence**: 6/6 R05 requirements validated

**D2.4: OpenAPI Requirements** (2 hours)

- R08-03: Swagger UI reflects API (deferred in passthru3)
- R08-06: Security schemes documented
- R08-02: Generated clients functional
- R08-01: Specs match implementation
- R08-05: Schema validation tests
- R08-04: No placeholder endpoints
- **Evidence**: 6/6 R08 requirements validated

**Acceptance Criteria**:

- [ ] Requirements coverage ≥90% (59/65 validated)
- [ ] All CRITICAL requirements covered (22/22)
- [ ] All HIGH requirements covered (26/26)
- [ ] Validation report shows ≥90%
- [ ] Zero TODO comments in new test files
- [ ] All new tests passing
- [ ] Post-mortem created: `P4.02-POSTMORTEM.md`

**Validation Commands**:

```bash
# Requirements coverage
identity-requirements-check --format json | jq '.overall_coverage_percentage'
# Should output: 90.0 or higher

# Coverage by priority
identity-requirements-check --by-priority
# CRITICAL: 100% (22/22)
# HIGH: 100% (26/26)
```

---

#### P4.03: Resolve All HIGH Severity TODOs

**Priority**: ⚠️ HIGH
**Effort**: 4 hours
**Dependencies**: P4.01 (may resolve some)
**Files**: Multiple (see deliverables)
**Status**: 🔜 PENDING

**Context**: From GAP-ANALYSIS.md - 4 HIGH TODOs identified

**Objectives**:

1. Fix database health checks (2 TODOs)
2. Fix cleanup logic for sessions/challenges (1 TODO)
3. Fix server startup/shutdown (2 TODOs)
4. Fix user repository integration (1 TODO) - may be resolved by P4.01

**Deliverables**:

**D3.1: Database Health Checks** (1 hour)

- File: `internal/identity/idp/handlers_health.go`
- File: `internal/identity/authz/handlers_health.go`
- Implement: `db.Ping()` to validate connectivity
- Return 503 if database unreachable
- **Evidence**: Test shows health check fails when DB down
- **TODO Resolved**: IMP-01, IMP-08

**D3.2: Session/Challenge Cleanup** (1 hour)

- File: `internal/identity/idp/service.go`
- Implement cleanup logic for expired sessions/challenges
- Run on service shutdown
- **Evidence**: Test shows cleanup runs on shutdown
- **TODO Resolved**: IMP-02

**D3.3: Server Lifecycle** (2 hours)

- File: `internal/identity/authz/service.go`
- Implement graceful startup (health check, migration)
- Implement graceful shutdown (cleanup, connection close)
- **Evidence**: Test shows graceful shutdown works
- **TODO Resolved**: IMP-03

**Acceptance Criteria**:

- [ ] Database health checks functional
- [ ] Session/challenge cleanup on shutdown
- [ ] Graceful server startup/shutdown
- [ ] Zero HIGH severity TODOs remain
- [ ] All tests passing
- [ ] Post-mortem created: `P4.03-POSTMORTEM.md`

**Validation Commands**:

```bash
# Zero HIGH TODOs
grep -r "TODO\|FIXME" internal/identity/ | grep -i "high\|critical" | wc -l
# Should output: 0
```

---

### 🟡 PHASE 2: Quality and Testing (Days 4-5)

#### P4.04: Achieve ≥85% Test Coverage

**Priority**: 🟡 MEDIUM
**Effort**: 1 day (8 hours)
**Dependencies**: P4.01, P4.02, P4.03
**Files**: Test files across identity packages
**Status**: 🔜 PENDING

**Context**: Ensure all packages meet coverage thresholds

**Objectives**:

1. Identify packages below 85% coverage
2. Add tests for uncovered code paths
3. Verify overall coverage ≥85%

**Deliverables**:

**D4.1: Coverage Analysis** (1 hour)

- Run: `go test ./internal/identity/... -cover -coverprofile=coverage.out`
- Identify packages <85%
- Generate HTML report: `go tool cover -html=coverage.out`

**D4.2: Add Missing Tests** (6 hours)

- Write table-driven tests for uncovered functions
- Test error paths and edge cases
- Use `t.Parallel()` for all tests

**D4.3: Verification** (1 hour)

- Re-run coverage
- Verify ≥85% overall
- Update PROJECT-STATUS.md

**Acceptance Criteria**:

- [ ] Overall coverage ≥85% for identity packages
- [ ] No package <80% coverage
- [ ] All new tests passing
- [ ] Post-mortem created: `P4.04-POSTMORTEM.md`

---

#### P4.05: Zero Test Failures

**Priority**: 🟡 MEDIUM
**Effort**: 4 hours
**Dependencies**: P4.04
**Files**: Test files with failures
**Status**: 🔜 PENDING

**Context**: Passthru3 had 23 test failures (77.9% pass rate)

**Objectives**:

1. Identify all failing tests
2. Fix or remove failing tests
3. Achieve 100% pass rate

**Deliverables**:

**D5.1: Test Failure Analysis** (1 hour)

- Run: `runTests ./internal/identity/...`
- Categorize failures (flaky, deferred features, bugs)

**D5.2: Fix Failing Tests** (3 hours)

- Fix flaky tests (timing, race conditions)
- Skip deferred feature tests with `t.Skip("deferred to future")`
- Fix bug-related failures

**Acceptance Criteria**:

- [ ] All tests passing (100% pass rate)
- [ ] Deferred tests skipped with clear reason
- [ ] No flaky tests remain
- [ ] Post-mortem created: `P4.05-POSTMORTEM.md`

---

#### P4.06: OpenAPI Synchronization

**Priority**: 🟡 MEDIUM
**Effort**: 4 hours
**Dependencies**: P4.01 (APIs stable)
**Files**: `api/identity/*.yaml`
**Status**: 🔜 PENDING

**Context**: Passthru3 deferred Phase 3 OpenAPI sync

**Objectives**:

1. Update OpenAPI specs to match implementation
2. Regenerate client libraries
3. Validate Swagger UI

**Deliverables**:

**D6.1: Spec Updates** (2 hours)

- Add missing endpoints (login, consent, logout, userinfo)
- Update request/response schemas
- Add security schemes (OAuth 2.1)

**D6.2: Client Regeneration** (1 hour)

- Run: `make generate-openapi-clients`
- Verify generated code compiles
- Test client library usage

**D6.3: Swagger UI Validation** (1 hour)

- Start server: `cryptoutil server start --dev`
- Navigate: `https://127.0.0.1:8080/ui/swagger`
- Test endpoints via Swagger UI

**Acceptance Criteria**:

- [ ] OpenAPI specs match implementation
- [ ] Client libraries regenerated and functional
- [ ] Swagger UI reflects all endpoints
- [ ] All R08 requirements validated
- [ ] Post-mortem created: `P4.06-POSTMORTEM.md`

---

### 🟢 PHASE 3: Production Readiness (Days 6-7)

#### P4.07: Resolve MEDIUM TODOs

**Priority**: 🟢 LOW
**Effort**: 4 hours
**Dependencies**: P4.01-P4.06
**Files**: Multiple (12 MEDIUM TODOs)
**Status**: 🔜 PENDING

**Context**: From GAP-ANALYSIS.md - 12 MEDIUM TODOs for polish

**Objectives**:

1. Add structured logging (2 TODOs)
2. Register additional auth profiles (1 TODO)
3. Complete TOTP/passkey validation (7 TODOs)
4. Add MFA context retrieval (2 TODOs)

**Acceptance Criteria**:

- [ ] Structured logging added
- [ ] Auth profiles registered
- [ ] TOTP/passkey validation complete
- [ ] Zero MEDIUM TODOs remain
- [ ] Post-mortem created: `P4.07-POSTMORTEM.md`

---

#### P4.08: Final Verification and Production Readiness

**Priority**: 🔴 CRITICAL
**Effort**: 1 day (8 hours)
**Dependencies**: P4.01-P4.07
**Files**: Documentation, validation reports
**Status**: 🔜 PENDING

**Objectives**:

1. Run all validation checks
2. Update PROJECT-STATUS.md
3. Create production readiness report
4. Document remaining LOW TODOs

**Deliverables**:

**D8.1: Validation Checks** (4 hours)

- Requirements coverage: ≥90%
- Test coverage: ≥85%
- Test pass rate: 100%
- TODO audit: 0 CRITICAL, 0 HIGH, 0 MEDIUM
- Linting: Zero errors
- Security scan: Zero CRITICAL/HIGH findings

**D8.2: Documentation** (2 hours)

- Update PROJECT-STATUS.md (final metrics)
- Create PRODUCTION-READINESS-REPORT.md
- Document known limitations (21 LOW TODOs)

**D8.3: Production Deployment Approval** (2 hours)

- Review all acceptance criteria
- Verify all quality gates passed
- Update status: ✅ PRODUCTION READY

**Acceptance Criteria**:

- [ ] Requirements coverage ≥90%
- [ ] Test coverage ≥85%
- [ ] Test pass rate 100%
- [ ] Zero CRITICAL/HIGH/MEDIUM TODOs
- [ ] All production blockers resolved
- [ ] PROJECT-STATUS.md updated to PRODUCTION READY
- [ ] PRODUCTION-READINESS-REPORT.md created
- [ ] Post-mortem created: `P4.08-POSTMORTEM.md`

---

## Post-Mortem Template

**For each task**, create `P4.XX-POSTMORTEM.md`:

```markdown
# Task P4.XX Post-Mortem

**Task**: [Task Name]
**Date Completed**: YYYY-MM-DD
**Time Spent**: X hours (estimate vs actual)

## What Went Well
- Bullet list of successes

## What Went Wrong
- Bullet list of bugs/issues discovered
- Include code citations (file:line)

## Omissions / Gaps
- What was missed during implementation
- Why it was missed

## Corrective Actions
### Immediate (Applied in this task)
- Fixes applied
- Tests added

### Deferred (Future tasks)
- New task created: P4.XX - Description
- Issue tracking: Link to task doc

## Lessons Learned
- Patterns to avoid
- Patterns to follow
```

---

## Timeline

**Total Effort**: 7 days (assuming full-time focus, continuous work)

| Phase | Tasks | Days | Dependencies |
|-------|-------|------|--------------|
| Phase 1 | P4.01-P4.03 | 3 | None → P4.01 → P4.02 → P4.03 |
| Phase 2 | P4.04-P4.06 | 2 | P4.03 → P4.04 → P4.05 → P4.06 |
| Phase 3 | P4.07-P4.08 | 2 | P4.06 → P4.07 → P4.08 |

**Critical Path**: P4.01 (foundation) → P4.02 (requirements) → P4.08 (verification)

---

## Success Criteria

**Production Ready When**:

- ✅ All 8 tasks complete (P4.01-P4.08)
- ✅ Requirements coverage ≥90% (59/65)
- ✅ Test coverage ≥85%
- ✅ Test pass rate 100%
- ✅ Zero CRITICAL/HIGH/MEDIUM TODOs
- ✅ All production blockers resolved
- ✅ PROJECT-STATUS.md shows PRODUCTION READY
- ✅ PRODUCTION-READINESS-REPORT.md approved

---

**Plan Version**: Passthru4 v1.0
**Last Updated**: November 24, 2025
**Next Review**: After P4.01 completion
