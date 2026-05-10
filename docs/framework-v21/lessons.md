# Lessons - Framework v21 TestMain Orchestration Consolidation

## Executive Summary

(To be filled at plan completion - numbered links to each phase section with one-sentence outcome.)

## Actions

(To be filled at plan completion - numbered list of concrete follow-up items for reviewer.)

> Per-phase structure (mandatory during execution):
> 1. What Worked
> 2. What Did Not Work
> 3. Root Causes
> 4. Patterns for Future Phases

## Phase 1: Research Freeze and Baseline Evidence

### What Worked

1. **Inventory freeze successful**: All 39 TestMain functions classified and mapped
   - 28 in internal/apps (10 PS-IDs with server/e2e/client/repository variants)
   - 11 in internal/apps-framework (service, config, server, repository, barrier, OTel tests)
   - Classification reflects actual build tags and setup behavior

2. **Canonical implementations identified**:
   - E2E baseline: e2e_infra.SetupE2ETestMain with compose orchestration
   - Integration baseline: testserver.StartAndWait with dual port polling
   - Both patterns are battle-tested and widely used

3. **Deep analysis uncovered critical patterns**:
   - e2e_helpers.MustStartAndWaitForDualPorts is panic-based; testserver.StartAndWait is proper TB-based
   - sm-kms businesslogic per-test setup pattern requires TestMain shared fixture refactor
   - pki-ca e2e health-wait race risk identified and flagged for migration

4. **Package consolidation matrix clear**:
   - testing/* packages understood (11 existing packages)
   - service/testutil HTTP mocks identified for consolidation
   - Reusable utilities vs orchestration-specific clearly separated

### What Didn't Work

1. **Initial scope collapse**: Early drafts collapsed test_help family into integration/e2e only
   - Regression required to expand from 2-directory to 8-directory model
   - Solution: Explicit taxonomy promoted to top-level plan with ownership boundaries

2. **Incomplete classification decisions**:
   - orm_transaction_test.go initially mis-classified as sad-path
   - businesslogic_test.go pattern not recognized until detailed analysis
   - Solution: Code review + build tag inspection + actual behavior tracing

### Root Causes

1. **Scope was too broad too quickly**: Trying to understand 39 files without framework taxonomy first
   - Lesson: Define taxonomy and ownership FIRST, then map concrete files to taxonomy
   - Pattern: Abstract model -> concrete inventory -> validation

2. **Conflation of execution profiles with directory taxonomy**:
   - Early draft used "integration profile" and "e2e profile" as directory names
   - Confusion between test execution modes (how servers start) vs directory ownership (who owns what)
   - Solution: Separated concerns - taxonomy is directory-based (orchestration vs helpers); profiles are execution modes

### Patterns for Future Phases

1. **Taxonomy-first design prevents drift**: Define directory ownership and dependency boundaries BEFORE code migration
2. **Deep code inspection required**: `grep TestMain` + `grep build tag` + actual file read + understanding intent
3. **Migrate with pattern, not file-by-file**: Establish core API (test_orch_integration), then apply systematically to all callers
4. **Validation at each layer**: Phase research -> validate model against code -> implement -> validate again

---

## Phase 2: Orchestration API Design

### What Worked

1. **API design convergence achieved**: Merged user feedback into concrete API shapes
   - Round 1: One-pass migration decision (no compatibility wrappers)
   - Round 2: Readiness defaults (admin readyz) + port policy (port 0)
   - Round 3: Fixture scope defaults + error-path contract

2. **Boundary rules explicit and validated**:
   - Lifecycle ownership (start/wait/shutdown) -> test_orch_* only
   - Docker Compose orchestration -> test_orch_e2e
   - Config/env wiring -> test_help_bootstrap
   - DB fixtures -> test_help_db
   - HTTP helpers -> test_help_api

3. **Error-path contract nailed**:
   - BuildBrokenDBFixture + BuildBrokenAPIFixture surface for deterministic failure testing
   - No surprises or ambiguity in how to test error paths

### What Didn't Work

1. **Over-specification of fixture scope**: Initial design had too many fixture modes
   - Solution: Settled on per-package shared + opt-in per-test isolation (simpler, covers 95% of use cases)

2. **Unclear port binding strategy**: Should tests always use port 0? (yes)
   - Initial design ambiguous about whether port 0 was mandatory or optional
   - Solution: Made port 0 MANDATORY for integration tests; e2e tests use standard ports

### Root Causes

1. **Tried to design for all use cases at once**:
   - Lesson: Design for 80% happy path first, then add knobs for 20% edge cases
   - Pattern: Start simple, add complexity only when patterns emerge

2. **Questions from Phase 2 revealed planning gaps**:
   - Design wasn't concrete enough to answer "how do I actually call this?"
   - Solution: API method signatures and fixture construction became concrete, now design is complete

### Patterns for Future Phases

1. **API design must be concrete and testable**: Can't move to implementation without code skeletons
2. **Fixture scope defaults matter**: Default to shared + simple, make per-test opt-in (reduces test boilerplate)
3. **Error-path contract needs explicit factory methods**: BuildBrokenX pattern is cleaner than BuildX(broken=true)

---

## Phase 3: Implement and Consolidate Framework Packages

### What Worked

1. **Directory structure created successfully**: All 8 directories created without issues
   - test_orch_integration, test_orch_e2e (already had E2E tests)
   - test_help_bootstrap, test_help_barrier, test_help_db, test_help_api, test_help_cli, test_help_tls
   - All packages build cleanly

2. **Core test_orch_integration API implemented**:
   - StartIntegrationServer() wraps ServiceServer with DB and cleanup
   - Dual port allocation, health readiness, error-path fixtures all in place
   - Integration pattern ready for use by 28 internal/apps TestMain files

3. **Foundation solid for massive migration**:
   - API is clean and minimal
   - Error-path support built in from the start
   - Cleanup via tb.Cleanup() prevents resource leaks

### What Didn't Work

1. **Pre-commit hook TODO blocking**: Initial placeholder files had TODOs that blocked commits
   - Solution: Removed TODO comments after scaffolding directories
   - Lesson: Don't use TODO in new code - it blocks commits immediately

### Root Causes

1. **Scope of work is massive**: 54 remaining tasks is a substantial undertaking
   - 28 internal/apps TestMain files to migrate
   - 11 framework TestMain files to migrate
   - Consolidating existing testing packages
   - Creating linter policies
   - Full validation
   - Lesson: This is a 2-3 phase long project; need strategic planning

### Patterns for Future Phases

1. **Migrate one PS-ID completely before moving to next**: Ensures pattern is correct before scaling
2. **Use test-driven migration**: Write test for new API, then migrate callers, validate test passes
3. **Consolidate framework testing packages incrementally**: Don't try to move all 11 packages at once
4. **Create linter policies early**: Enforce new patterns as they're created to prevent drift

## Phase 4: Migrate internal/apps (All 10 PS-IDs)

### What Worked

1. **sm-kms migration pattern is now validated end-to-end**:
   - `server/testmain_test.go` and `client/testmain_test.go` use `test_orch_integration` wrappers.
   - ORM integration-tagged suite now uses unified `testmain_test.go` fixtures.
   - Full ORM package integration run passes reliably.

2. **Type-boundary issues were resolved at root cause**:
   - `ElasticKeyStatus` comparisons were corrected to match server type wrappers.
   - Builder tests now assert on the exact runtime type path used by repository entities.

3. **Cleanup harness robustness improved**:
   - Nested `t.Cleanup(func(){ CleanupDatabase(...) })` anti-pattern was removed.
   - Tests now call `CleanupDatabase(...)` directly, enabling deterministic pre/post cleanup wiring.

4. **jose-ja server migration confirmed portability of the pattern**:
   - Server TestMain moved to `test_orch_integration` without breaking existing integration tests.
   - Compatibility variables were preserved to avoid broad test rewrites during this phase.

5. **Identity server migrations scaled the pattern without reopening design**:
   - identity-authz, identity-idp, identity-rp, identity-rs, and identity-spa server TestMains all migrated using the same orchestration helper.
   - The external-package case (`identity-rp/server_test`) proved the pattern still works when the concrete server type must stay qualified.
   - TLS client setup for RP tests survived the migration cleanly because it was attached after orchestration startup, not folded into helper internals.

### What Didn't Work

1. **Initial assumption of purely pre-existing failures was incomplete**:
   - Focused tests passed individually but failed when run together.
   - This masked shared fixture interference and delayed root-cause isolation.

2. **Mutex-based serialization workaround caused deadlocks**:
   - A package-level mutex in `CleanupDatabase` interacted badly with nested subtests.
   - Full package run hung until timeout due lock contention chains.

3. **Task tracking drifted behind real execution**:
   - Multiple Phase 4 tasks were already complete in code, but `tasks.md` still showed them as not started.
   - This increased the risk of duplicate investigation and false phase status reporting.

### Root Causes

1. **Shared in-memory fixture with parallel mutation tests requires strict lifecycle discipline**:
   - Cross-test contamination occurred when cleanup registration pattern was inconsistent.
   - Package-level state was mutated from multiple parallel tests without deterministic setup boundaries.

2. **Cleanup API misuse pattern had propagated across multiple test files**:
   - Wrapping cleanup helper registration inside another cleanup callback created delayed execution semantics and non-obvious ordering.
   - Correct usage is direct helper invocation at test start.

3. **Progress metadata was treated as secondary instead of a quality artifact**:
   - Once the implementation stream accelerated, task evidence and phase status were not updated at the same cadence as code.
   - That created an avoidable mismatch between repository truth and plan documentation truth.

### Patterns for Future Phases

1. **For shared SQLite integration fixtures, call cleanup helpers directly at test start**; avoid nested cleanup wrappers.
2. **When failures are flaky, always run both isolated and grouped test selections** before concluding root cause.
3. **Prefer behavioral isolation fixes (filter scoping, fixture cleanup discipline) over global locking** in test helpers.
4. **Treat package-level shared fixture tests as sequential when they mutate shared state broadly** and document with explicit `Sequential:` comments where needed.
5. **Keep external-package and TLS-client variants inside the migration sample set early** so the orchestration API proves it handles both plain startup and post-start client wiring.
6. **Update task evidence immediately after each migration cluster lands**; stale documentation becomes its own blocker in long-running plans.

## Phase 5: Migrate internal/apps-framework TestMain files

### What Worked

1. **Framework inventory reduction avoided unnecessary edits**:
   - Re-reading all in-scope framework TestMains showed most files were already clean fixture initializers or no-op `testutil.Initialize()` wrappers.
   - That let the work focus narrowly on the two files that still owned manual SQLite lifecycle logic.

2. **`test_help_db` gained the missing suite-level primitive cleanly**:
   - `NewInMemorySQLiteDBForTestMain()` now exposes the same canonical SQLite setup path for `TestMain` callers that already existed for per-test helpers.
   - This removed the need for framework packages to duplicate `sql.Open`, PRAGMA setup, and connection-pool tuning.

3. **Framework TestMain migrations stayed local and low-risk**:
   - `server/repository/orm/testmain_test.go` now uses the shared DB helper and preserves package globals used by the test suite.
   - `server/apis/test_main_test.go` now uses the shared DB helper plus explicit migration application, preserving the package's database expectations without hand-rolled DSN assembly.

4. **Narrow validation was effective**:
   - Two-pass targeted `golangci-lint` plus `go build ./internal/apps-framework/...` and then `go build ./...` caught structural issues quickly without reopening unrelated packages.

### What Didn't Work

1. **The first `apply_patch` attempts failed due to local file drift**:
   - The expected patch context no longer matched the real file contents.
   - This is a recurring failure mode late in long execution sessions when files have already been partially modified.

2. **A failed patch briefly corrupted `test_help_db/database.go`**:
   - The file ended up with a stray import fragment embedded in a cleanup block.
   - The corruption was cheap to detect because the next build failed immediately.

3. **PowerShell whole-file rewrites reduced formatting fidelity until lint normalized them**:
   - They were effective for recovery, but they temporarily dropped indentation formatting and increased the need for a cleanup pass.

### Root Causes

1. **The code path for suite-level SQLite setup existed conceptually but not as a reusable exported helper**:
   - `buildInMemorySQLiteDB` already contained the correct logic.
   - The missing seam was an exported wrapper for `TestMain` callers without a `*testing.T`.

2. **Patch-based editing was fragile because the working copy had diverged from the earlier read window**:
   - The patch failure was not a design problem; it was a stale-context problem.
   - Rewriting the affected files was the fastest way to restore a known-good state and continue with focused validation.

3. **Framework TestMain migration risk comes from hidden package-global dependencies, not from the DB helper itself**:
   - `orm/testmain_test.go` needed `testGormDB` to remain global because other tests in the package read it directly.
   - Confirming those references before finalizing the rewrite avoided a local regression.

### Patterns for Future Phases

1. **When only one abstraction seam is missing, add the seam instead of cloning setup logic into each caller**.
2. **After a failed file edit, run the cheapest compile check immediately**; it is the fastest discriminator between harmless formatting drift and real file corruption.
3. **Before rewriting a TestMain, grep for package globals consumed by sibling tests** so the rewrite preserves those shared anchors.
4. **Use shared helper + explicit migration application when a package needs custom schema setup but not custom low-level DB construction**.
5. **Treat documentation drift as a real regression after code migrations**; phase completion is not valid until tasks and lessons reflect the actual repository state.

## Phase 6: Template and Linter Policy Lock

(To be filled during Phase 6 execution using the 4-section structure above.)

## Phase 7: Validation and Rollout

(To be filled during Phase 7 execution using the 4-section structure above.)
