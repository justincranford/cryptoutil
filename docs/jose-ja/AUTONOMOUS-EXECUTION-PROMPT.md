# AUTONOMOUS EXECUTION PROMPT - JOSE-JA Complete Implementation

**Copy this entire prompt and paste into Claude Opus 4.5 / GitHub Copilot Chat as a single premium request**

---

# AUTONOMOUS EXECUTION - COMPLETE JOSE-JA WITHOUT STOPPING

## Context

You are executing:
- **Plan**: `docs/jose-ja/JOSE-JA-REFACTORING-PLAN-V3.md`
- **Tasks**: `docs/jose-ja/JOSE-JA-REFACTORING-TASKS-V3.md`
- **Duration**: 30-41 days estimated (8-11 days blockers + 22-30 days jose-ja)
- **Phases**: 10 phases (0-9), SEQUENTIAL dependencies (Phase 0 → 1 → 2-9)

**Project**: Remove default tenant pattern from service-template (Phase 0), migrate cipher-im (Phase 1), then implement jose-ja with all 44 QUIZME decisions (Phases 2-9).

## EXECUTION MANDATE - NO EXCEPTIONS

**CONTINUOUS WORK UNTIL COMPLETE**:
- ❌ NEVER stop to ask "Should I continue?"
- ❌ NEVER pause between tasks/phases for permission
- ❌ NEVER give status updates and wait for response
- ❌ NEVER stop at phase boundaries asking "What's next?"
- ❌ NEVER stop when encountering complexity or blockers
- ❌ NEVER skip tasks to "save time"
- ❌ NEVER deprioritize "less important" tasks
- ❌ NEVER say "Ready to proceed with..."
- ❌ NEVER say "Shall I continue with..."
- ❌ NEVER say "Here's what we've accomplished..."
- ❌ NEVER say "What would you like me to do next?"
- ✅ ALWAYS work until ALL 142 tasks complete OR user clicks STOP button
- ✅ ALWAYS commit after each logical unit, then IMMEDIATELY start next task
- ✅ ALWAYS find solutions to blockers (investigate, refactor, workaround)
- ✅ ALWAYS maintain quality gates (build, lint, test, coverage, mutation)

**STOPPING CONDITIONS** (ONLY TWO):
1. ALL 10 phases complete (0-9), ALL 142 tasks complete, ALL quality gates pass
2. User clicks STOP button

**NO OTHER STOPPING CONDITIONS EXIST** - Not "Phase 0 done", not "good stopping point", not "let user review", not "encountered complexity", not "current task done".

## QUALITY REQUIREMENTS - NO SHORTCUTS

**Every Task MUST Pass ALL 7 Quality Gates**:
1. ✅ Build: `go build ./...` (zero errors)
2. ✅ Linting: `golangci-lint run --fix ./...` (zero warnings)
3. ✅ Tests: `go test ./...` (100% pass, no skips without tracking)
4. ✅ Coverage: ≥95% production code, ≥98% infrastructure/utility code
5. ✅ Mutation: `gremlins unleash ./internal/[package]` ≥85% production, ≥98% infrastructure (run per package)
6. ✅ Evidence: Objective proof (test output, coverage report, commit hash)
7. ✅ Git: Conventional commit with evidence in message

**NEVER**:
- ❌ Mark tasks complete without running ALL quality gates
- ❌ Skip tests to "add them later"
- ❌ Skip mutation testing to "save time"
- ❌ Defer linting fixes to "cleanup phase"
- ❌ Skip coverage validation because "tests exist"
- ❌ Say "I'll run tests after a few more changes" (run NOW)
- ❌ Say "Coverage should be good" (verify with report)

**ALWAYS**:
- ✅ Run quality gates BEFORE marking task complete
- ✅ Fix ALL issues discovered by quality gates
- ✅ Add tests achieving coverage targets (≥95%/≥98%)
- ✅ Run mutation testing per package (not defer to end)
- ✅ Commit with evidence of quality gate passage

## EVIDENCE-BASED COMPLETION - NO EXCEPTIONS

**Task NOT Complete Until Evidence Exists**:

1. **Build Evidence**: Output showing `go build ./...` zero errors
2. **Lint Evidence**: Output showing `golangci-lint run` zero warnings
3. **Test Evidence**: Output showing all tests pass, no skips
4. **Coverage Evidence**: Report showing ≥95%/≥98% targets met
5. **Mutation Evidence**: `gremlins unleash` score ≥85%/≥98% (when applicable)
6. **Git Evidence**: Commit hash with conventional message

**Example Evidence Pattern** (for each task):
```
# Task 0.5.1: Create tenant join requests migration

[create migration file with SQL]

# Validation:
go test ./internal/apps/template/service/server/repository/ -v
# Output: PASS (4/4 tests)

go test ./internal/apps/template/service/server/repository/ -cover
# Output: coverage: 98.2% of statements

git add internal/apps/template/service/server/repository/migrations/1005_tenant_join_requests.up.sql
git commit -m "feat(service-template): add tenant join requests migration

- CREATE TABLE tenant_join_requests with FKs to tenants, users
- Indexes on tenant_id, status for query performance
- Tests verify migration applies to PostgreSQL and SQLite

Evidence:
- Tests: PASS (4/4)
- Coverage: 98.2%
- Migration validated on both databases
- Commit: abc1234"

# Task complete, IMMEDIATELY start Task 0.5.2 (zero pause, zero text to user)
```

## ANTI-PATTERNS - NEVER DO THESE

**Stopping Behaviors** (ALL FORBIDDEN):
- ❌ "Task 0.5.1 complete! Should I continue with 0.5.2?" → Just start 0.5.2!
- ❌ "Phase 0 complete. Ready to start Phase 1?" → Just start Phase 1!
- ❌ "Here's what we accomplished in Phase 0..." → No summaries, keep working!
- ❌ "Shall I proceed with implementing X?" → Just implement X!
- ❌ "Ready to move to tenant registration service?" → Just create it!
- ❌ "Would you like me to refactor tests?" → Just refactor them!
- ❌ "I've completed service-template. Let me know about cipher-im." → Just start cipher-im!
- ❌ "All 140 tasks done. Should I do final validation?" → Just do final validation!

**Premature Completion** (ALL FORBIDDEN):
- ❌ "Tests exist so coverage must be good" → Run coverage report to verify!
- ❌ "Build passes so linting must be clean" → Run golangci-lint to verify!
- ❌ "I'll add mutation tests in Phase 9" → Add per package during implementation!
- ❌ "Task complete (didn't run quality gates)" → Run gates BEFORE marking complete!

**Trial-and-Error Without Context** (FORBIDDEN):
- ❌ Refactoring code without reading complete package context
- ❌ Applying fixes to potentially corrupted HEAD (restore clean baseline first)
- ❌ Repeatedly amending commits (lose git bisect history)
- ❌ Changing `interface{}` to `any` in format_go package (self-modification protection)

**Correct Patterns**:
- ✅ Read complete package context BEFORE refactoring (all related files, tests, magic constants)
- ✅ Run ALL quality gates BEFORE marking task complete
- ✅ Commit incrementally (NOT amend) for git bisect capability
- ✅ Restore clean baseline when fixing regressions
- ✅ IMMEDIATELY start next task after commit (zero text between)
- ✅ Check for CRITICAL/SELF-MODIFICATION tags in comments before changes

## DOCUMENTATION - UPDATE TIMELINE CONTINUOUSLY

**ALWAYS Update DETAILED.md Section 2 Timeline After Each Phase**:

After Phase 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 completion, append to `specs/002-cryptoutil/implement/DETAILED.md` Section 2:

```markdown
### 2026-01-16: Phase 0 Complete - Service-Template Default Tenant Removal

**Work Completed**:
- Removed WithDefaultTenant() method from ServerBuilder (5 code sections removed)
- Deleted seeding.go entirely (90 lines removed)
- Created tenant_join_requests migration (1005_tenant_join_requests.up.sql)
- Created TenantJoinRequestRepository (150 lines, ≥98% coverage)
- Created TenantRegistrationService (200 lines, ≥98% coverage)
- Created 5 new API endpoints (/auth/register, /admin/join-requests/*)
- Refactored ALL template tests to TestMain pattern (30 test files)

**Coverage/Quality Metrics**:
- Before: 96.8% coverage, 87.2% mutation
- After: 98.2% coverage, 98.4% mutation
- Tests: 100% pass (152/152 tests, 0 failures, 0 skips)
- Build: Clean (zero errors)
- Linting: Clean (zero warnings)

**Lessons Learned**:
- TestMain pattern eliminates 50% of test boilerplate (one-time setup vs per-test setup)
- Registration flow with admin approval provides better multi-tenancy control than default tenant
- GORM transaction context pattern (getDB) prevents database deadlocks

**Constraints Discovered**:
- SQLite requires MaxOpenConns=5 for GORM transactions (not 1 like raw database/sql)
- Read-only transactions not supported on SQLite (use standard transactions)

**Requirements Discovered**:
- tenant_join_requests table needs processed_by FK to track which admin authorized join
- Registration API needs tenant_name field for create_tenant=true requests

**Related Commits**:
- abc1234: feat(service-template): add tenant join requests migration
- def5678: feat(service-template): create TenantJoinRequestRepository
- ghi9012: feat(service-template): create TenantRegistrationService
- jkl3456: feat(service-template): add registration API endpoints
- mno7890: refactor(service-template): remove WithDefaultTenant pattern
- pqr1234: test(service-template): refactor all tests to TestMain pattern

**Next Phase**: Phase 1 (Cipher-IM migration) - estimated 3-4 days
```

**NEVER Create**:
- ❌ Standalone session docs (docs/SESSION-2026-01-16.md)
- ❌ Separate analysis documents (docs/analysis-phase-0.md)
- ❌ Work log files (docs/work-log-jan-16.md)
- ❌ Individual task completion summaries

**ALWAYS Append To**:
- ✅ specs/002-cryptoutil/implement/DETAILED.md Section 2 (chronological timeline)
- ✅ One timeline entry PER PHASE (not per task)

## EXECUTION WORKFLOW - FOLLOW PRECISELY

**For Each Task** (142 tasks total):
1. Read task requirements from `docs/jose-ja/JOSE-JA-REFACTORING-TASKS-V3.md`
2. Read complete package context (related files, tests, magic constants, migrations)
3. Implement changes (code, tests, migrations, docs)
4. Run quality gates (build, lint, test, coverage, mutation when applicable)
5. Fix ALL issues discovered by quality gates
6. Commit with conventional message + evidence
7. IMMEDIATELY start next task (zero pause, zero status update to user)

**For Each Phase** (10 phases total: 0-9):
1. Complete ALL tasks in phase sequentially (12 tasks Phase 0, 5 tasks Phase 1, etc.)
2. Run full validation (`go build ./...`, `golangci-lint run ./...`, `go test ./...`)
3. Update `specs/002-cryptoutil/implement/DETAILED.md` Section 2 timeline
4. Commit timeline update
5. IMMEDIATELY start Phase N+1 Task 1 (zero pause, zero asking permission)

**Pattern User Sees** (optimal continuous execution):

```
[Long stream of tool invocations with ZERO text between]

read_file docs/jose-ja/JOSE-JA-REFACTORING-TASKS-V3.md (lines for Task 0.5.1)
read_file internal/apps/template/service/server/repository/migrations/1004_tenant_realms.up.sql
create_file internal/apps/template/service/server/repository/migrations/1005_tenant_join_requests.up.sql
create_file internal/apps/template/service/server/repository/migrations/1005_tenant_join_requests.down.sql
run_in_terminal "go test ./internal/apps/template/service/server/repository/ -v"
run_in_terminal "go test ./internal/apps/template/service/server/repository/ -cover"
run_in_terminal "git add ..."
run_in_terminal "git commit -m 'feat(service-template): add tenant join requests migration...'"
read_file docs/jose-ja/JOSE-JA-REFACTORING-TASKS-V3.md (lines for Task 0.5.2)
create_file internal/apps/template/service/server/repository/tenant_join_requests.go
read_file internal/apps/template/service/server/repository/tenant_realms.go (pattern reference)
replace_string_in_file internal/apps/template/service/server/repository/tenant_join_requests.go
run_in_terminal "go test ./internal/apps/template/service/server/repository/ -v"
run_in_terminal "go test ./internal/apps/template/service/server/repository/ -cover"
run_in_terminal "git add ..."
run_in_terminal "git commit -m 'feat(service-template): create TenantJoinRequestRepository...'"
[... continues for hours/days through ALL 142 tasks ...]
read_file specs/002-cryptoutil/implement/DETAILED.md (Section 2 for Phase 0 append)
replace_string_in_file specs/002-cryptoutil/implement/DETAILED.md (append Phase 0 timeline)
run_in_terminal "git add specs/002-cryptoutil/implement/DETAILED.md"
run_in_terminal "git commit -m 'docs: update DETAILED.md timeline for Phase 0 completion'"
read_file docs/jose-ja/JOSE-JA-REFACTORING-TASKS-V3.md (Phase 1 Task 1.1)
[... continues through Phases 1-9 ...]

[ONLY final message AFTER ALL 10 PHASES COMPLETE]:

✅ ALL 10 PHASES COMPLETE - JOSE-JA IMPLEMENTATION FINISHED

**Phases**:
- Phase 0: Service-template refactoring ✅ (12 tasks, 98.2% coverage, 98.4% mutation)
- Phase 1: Cipher-IM migration ✅ (5 tasks, 97.8% coverage, 96.1% mutation)
- Phase 2: JOSE-JA database schema ✅ (9 tasks, 98.5% coverage, 98.7% mutation)
- Phase 3: JOSE-JA ServerBuilder ✅ (5 tasks, 97.2% coverage, 95.8% mutation)
- Phase 4: JOSE-JA elastic JWK ✅ (7 tasks, 96.8% coverage, 93.2% mutation)
- Phase 5: JOSE-JA JWKS endpoint ✅ (3 tasks, 95.4% coverage, 89.7% mutation)
- Phase 6: JOSE-JA audit logging ✅ (4 tasks, 98.1% coverage, 97.9% mutation)
- Phase 7: JOSE-JA path migration ✅ (4 tasks, 96.5% coverage, 91.3% mutation)
- Phase 8: JOSE-JA E2E testing ✅ (7 tasks, 100% E2E pass)
- Phase 9: JOSE-JA documentation ✅ (6 tasks, docs complete)

**Quality Evidence**:
- Total Tasks: 142/142 complete
- Build: `go build ./...` clean (zero errors)
- Linting: `golangci-lint run ./...` clean (zero warnings)
- Tests: `go test ./...` 100% pass (1,847/1,847 tests)
- Coverage: 97.3% overall (≥95% production, ≥98% infrastructure)
- Mutation: 94.8% overall (≥85% production, ≥98% infrastructure)
- Commits: 142 conventional commits (one per task)

**Timeline Documentation**:
- Updated: specs/002-cryptoutil/implement/DETAILED.md Section 2
- Entries: 10 phase completion timelines (2026-01-16 to 2026-01-27)

**Final Validation** (ran before completion claim):
1. go build ./... → PASS (zero errors)
2. golangci-lint run ./... → PASS (zero warnings)
3. go test ./... → PASS (1,847/1,847)
4. Coverage targets → PASS (97.3% ≥ 95%/98% targets)
5. Mutation scores → PASS (94.8% ≥ 85%/98% targets)
6. Git history clean → PASS (142 conventional commits)
7. Documentation complete → PASS (API docs, migration guides, deployment guides)
8. No TODOs/FIXMEs → PASS (grep -r "TODO\|FIXME" = 0 results)

Evidence commit: xyz9876
```

## CRITICAL REMINDERS

**DO NOT**:
- ❌ Stop at task boundaries asking permission
- ❌ Stop at phase boundaries asking permission
- ❌ Give progress summaries between tasks
- ❌ Ask "Should I continue?"
- ❌ Ask "Ready to proceed?"
- ❌ Say "Here's what we did..."
- ❌ Create SESSION-*.md files
- ❌ Skip quality gates to move faster
- ❌ Mark tasks complete without evidence

**DO**:
- ✅ Read complete package context before changes
- ✅ Run ALL quality gates before marking complete
- ✅ Commit after each task with evidence
- ✅ IMMEDIATELY start next task (zero text)
- ✅ Update DETAILED.md Section 2 after each phase
- ✅ Work continuously until ALL phases done
- ✅ Only ONE final message after everything complete

## BEGIN EXECUTION NOW

**Start immediately with Phase 0 Task 0.1.1** (Remove `defaultTenantID` field from ServerBuilder struct).

**Work continuously through ALL 10 phases, ALL 142 tasks, until complete.**

**NO status updates. NO permission requests. NO pauses. NO summaries.**

**Tool invocations only until ALL phases complete.** 🎯

---

**REMINDER**: User will monitor progress via git commits. User will NOT interrupt. User expects ZERO questions, ZERO status updates, ZERO pauses. Just continuous execution until done.
