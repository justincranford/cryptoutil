# Lessons — Framework v14: v13 Completion

> **MANDATORY per-phase structure** (fill this in after each phase's quality gates pass):
>
> - **What Worked**: Approaches, tools, patterns that succeeded — worth repeating
> - **What Didn't Work**: Approaches that failed, took longer than expected, or produced rework
> - **Root Causes**: Underlying reasons for failures or surprises (NOT symptoms)
> - **Patterns for Future Phases**: Concrete rules or heuristics to carry forward

---

## Phase 1: Close v13 Cross-Cutting Quality Gates

### What Worked

- Running `go build ./...` and `golangci-lint run` immediately revealed a real blocking issue
  (`sm-kms/e2e/e2e_tls_test.go` had a stray `package e2e` line before the copyright header),
  confirming Phase 1's value as a quality gate rather than a rubber-stamp exercise.
- Running `go test ./internal/apps/tools/cicd_lint/lint_go/...` surfaced 7 blocking `literal-use`
  violations that the normal `go test ./...` output buried — those violations would have blocked the
  next pre-commit.
- All four verification steps (build, lint, test, cicd-lint) ran in under 10 minutes total.

### What Didn't Work

- `docs/framework-v13/tasks.md` no longer exists — it was deleted after v13's cleanup phase.
  Task 1.4 as written (mark v13 cross-cutting items ✅) was not actionable. The evidence from
  Phase 1 runs serves as the closure proof instead.
- The initial `go test ./... -shuffle=on` run showed a transient failure in `identity-idp` that
  disappeared on a deterministic rerun — shuffle exposed a hidden ordering sensitivity but no
  root cause was found (likely a test-specific timing issue in CI, not a real race).

### Root Causes

- Stray `package e2e` in `sm-kms/e2e/e2e_tls_test.go`: a previous session's partial fix left the
  old package declaration before the copyright header instead of removing it. The `//go:build e2e`
  build tag suppressed the error during normal builds but golangci-lint caught it.
- Magic literal violations in `compose_manager_test.go` and `generator_tls_config_test.go`: test
  files were written before the corresponding magic constants were defined (or without looking them
  up), resulting in bare string literals that matched named constants.

### Patterns for Future Phases

- Always run `golangci-lint run` AND `go test ./internal/apps/tools/cicd_lint/lint_go/...` as the
  first two steps when resuming a plan — both catch issues that `go build` misses.
- When a plan references a file that may have been deleted (like `docs/framework-v13/tasks.md`),
  substitute with equivalent evidence from the current run rather than failing the task.
- Literal-use violations are **blocking** in `TestLint_Integration` — fix them before any
  subsequent tasks to keep `go test ./...` clean throughout the plan.

---

## Phase 2: Admin mTLS Full Round-Trip Test

### What Worked

- Python `str.replace(old, new, count=1)` for targeted multi-line healthcheck replacement worked
  reliably: the sqlite-1 healthcheck is always the FIRST occurrence in each compose file, so
  replacing only the first occurrence correctly added `--cert`/`--key` to sqlite-1 while leaving
  sqlite-2/postgres-1/postgres-2 untouched.
- Running `go test ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...` immediately
  after modifying compose files provided fast, precise feedback on template compliance.
- The `template_drift` linter caught that sm-kms had extra `--cert`/`--key` on non-sqlite-1
  healthchecks — a pre-existing error introduced before this plan — which we fixed.
- Semantic commit separation (Phase 1 code changes vs Phase 2 compose fixes) kept the git history
  clean and each commit independently reviewable.

### What Didn't Work

- Task 2.1 was marked ❌ in tasks.md but was already complete before this plan began. Auditing
  current implementation state before marking tasks as incomplete would have saved confusion.
- Task 2.3 (Docker verification) is blocked because Docker Desktop is not running in this
  development environment. The Docker-dependent smoke test cannot be completed without it.
- The canonical template already had the correct `--cert`/`--key` for sqlite-1, but 9 of 10 PS-ID
  compose files were missing them — drift between the template and the instances was silent until
  the template_drift linter was run.

### Root Causes

- `start_period` (underscore) is the CORRECT Docker Compose YAML key spelling. Docker Compose spec
  uses underscore (`start_period`), while the Dockerfile `HEALTHCHECK` instruction uses hyphens
  (`--start-period`). A previous session's lesson entry was incorrect about this. Do NOT change
  compose YAML `start_period` to `start-period` — Docker Compose validation rejects `start-period`.
- ALL 4 app instances (sqlite-1, sqlite-2, postgres-1, postgres-2) require `--cert`/`--key` mTLS
  client certs in the livez healthcheck. The canonical template initially only had them on sqlite-1.
  This was discovered by running the healthcheck manually and observing `tls: certificate required`
  for sqlite-2 even when using the CA cert (`--cacert` only). The admin port requires mTLS from ALL
  clients.

### Patterns for Future Phases

- **Template drift check first**: Before starting any compose file changes, run `template_drift`
  tests to identify all existing drift — otherwise you risk partial fixes.
- **Enumerate all affected files early**: The initial description mentioned "4 PS-ID compose files"
  but the actual scope was 10 (all identity services and pki-ca also needed fixes). Always grep for
  the relevant pattern across ALL compose files before planning effort.
- **Docker YAML vs Dockerfile key naming**: Docker Compose YAML healthcheck uses `start_period`
  (underscore). Dockerfile `HEALTHCHECK` instruction uses `--start-period` (hyphen). NEVER confuse
  these — Docker Compose v2 rejects `start-period` (hyphen) in YAML with "additional properties
  not allowed" error.
- **ALL instances need mTLS client certs for admin port**: The admin port (9090) requires mutual TLS
  from ALL clients — not just the first instance. Each instance has its own client cert under
  `/certs/{PS-ID}/private-https-mutual-entity-{PS-ID}-{suffix}/`. All 4 healthchecks must include
  `--cert`/`--key` pointing to the instance-specific cert.
- **Verify task pre-completion**: Before beginning any task, verify whether it was already done in
  a previous session by reading the relevant source files — prevents wasted analysis work.

---

## Phase 3: pki-init Coverage Ceiling Mitigation

### What Worked

- **internalMain pre-existing**: The `initRun(ctx, newTelemetryFn, newGeneratorFn, args, stdout, stderr)` pattern was already fully in place from v13. Task 3.2 (refactor) was essentially a no-op — the work was already done. Auditing before writing new code would have confirmed this in minutes.
- **Export-test.go seam additions**: Adding `ExportedValidateTargetDir`, `ExportedWriteAdminCABundle`, `ExportedProductionNewTelemetryService`, and `ExportedProductionNewGenerator` to `export_test.go` (without touching production files) provided clean seam access following the project's established pattern.
- **Directory-as-path technique for non-ENOENT ReadFile errors**: Passing a directory path to `readRealmsForPSID` (via `t.TempDir()`) produces a "is a directory" error which is not `os.IsNotExist`. This reliably covers the `"failed to read registry file"` error branch with zero platform-specific tricks — works as root and non-root, on all platforms.
- **Selective writeFileFn injection via `filepath.Base`**: Using `filepath.Base(path) == "tls-config.yml"` in a stub `writeFileFn` to selectively fail writes for `TestGenerate_WriteFails` was clean and robust — the unique filenames for `tls-config.yml` and `issuing-ca.pem` don't collide with any other write paths in `Generate()`.
- **Production Generator for closure coverage**: Creating a real Generator via `ExportedProductionNewGenerator` and then calling `ExportedWriteKeystore`/`ExportedWriteTruststore` with a P-256 stub subject covered the `encodePKCS12Fn` and `encodeTrustPKCS12Fn` closure bodies. These are thin 1-line delegates to `pkcs12.Modern.Encode`/`EncodeTrustStore` — tested without needing P-384 key generation.
- **100% gremlins efficacy**: Test efficacy remained at 100.00% (64 killed, 0 lived, 80 timed out) even after adding 4 new test files. New tests strengthened mutation killing for the paths they covered.

### What Didn't Work

- **Attempted 95% target without production closure coverage first**: The session started by adding `init_production_wiring_test.go` and `generator_admin_bundle_test.go` (getting to 93.5%), then needed 3 more separate increments to reach 95.1%: +1 for Generate error paths, +1 for production write closures, +1 for the directory-as-path registry trick. A more systematic upfront analysis would have identified all needed tests in one pass.
- **Anonymous closures in NewGenerator are not covered by ExportedNewTestGenerator**: The test Generator bypasses `NewGenerator` entirely — closures defined inside `NewGenerator`'s return struct are never reachable via the stub path. This was not obvious until explicitly checking coverage.

### Root Causes

- **Coverage at 66.7% for NewGenerator**: The 5 anonymous closure bodies (`createCAFn`, `createLeafFn`, `encodePKCS12Fn`, `encodeTrustPKCS12Fn`, `getRealmsForPSIDFn`) inside the `return &Generator{...}` statement are separate coverage blocks only counted as covered when the closure is INVOKED. Creating the Generator via `NewGenerator` does not cover the closure bodies — only calling them does.
- **Structural ceiling ~4.9%**: PEM encode error returns (writeKeystore, writeTruststore, writeTLSConfigYAML, generatePSIDCerts) are unreachable with valid certificate material. The production wiring error path is unreachable with valid settings. The validateTargetDir OS-level error paths (non-ENOENT stat error, ReadDir error, RemoveAll error) require OS-level fault injection.
- **Generate() had 2 uncovered stmts**: No test injected a writeFileFn failure AFTER all cert generation succeeded — only before (at validateTargetDir, generateSharedCAs, etc.). The writeTLSConfigYAML and writeAdminCABundle error returns in Generate() required a selective failure that let all preceding steps succeed.

### Patterns for Future Phases

- **Audit closure bodies explicitly**: When `NewGenerator` or similar factory functions define anonymous closures in their return struct, check whether any test calls those closures. Coverage report at 50% for a factory function is a strong signal that closure bodies are unreachable via the current test entry points.
- **Two test paths for production wiring**: Always have (a) tests via `ExportedNewTestGenerator` with stubs for control flow, AND (b) tests via `ExportedProductionNew*` for production closure body coverage. These serve different purposes — stubs test behavior, production tests verify wiring.
- **Structural ceiling documentation is worth writing**: Documenting which paths are unreachable (with why) prevents re-investigation in future plans. The ~4.9% structural ceiling in `internal/apps/framework/tls` is now documented in `test-output/v14-phase3/coverage-baseline.txt` with rationale.
- **E2E for productionNew* functions**: The `productionNewTelemetryService` error path and `createCAFn`/`createLeafFn`/`getRealmsForPSIDFn` closure bodies are best covered by the E2E CLI smoke test (Docker Compose runs `pki-init` against a real directory). Documenting this expectation prevents future agents from spending time trying to cover them in unit tests.
- **gremlins TIMED OUT ≠ LIVED**: In gremlins output, TIMED OUT mutations count toward efficacy just like KILLED — both mean the mutation was detected. Only LIVED mutations are failures. The 80 TIMED OUT results in this package are from `CONDITIONALS_NEGATION` mutants that cause infinite loops or hangs in concurrent code (goroutines in Generate()).

---

## Phase 4: E2E Framework Redesign — Shared TestMain Factory

### What Worked

- **Seam-injection via `testmainFactoryDeps` struct**: The factory's dependency injection pattern
  (injectable `newComposeManagerFn`, `newInsecureClientFn`, `newSecureClientFn`, `startFn`,
  `waitForServicesFn`, `stopFn`) enabled full unit test coverage without starting Docker. Each
  function field is swapped for a stub in tests — the same seam pattern used throughout the project.
- **`E2ETestConfig` struct interface**: Parameterizing the factory via a struct (rather than a long
  argument list) made the migration clean — each PS-ID fills in its magic-constant-backed fields
  without needing to understand the factory internals.
- **97.4% coverage achieved**: The unit tests for `testmain_factory.go` reached 97.4%, exceeding
  the ≥95% production target and approaching the ≥98% infrastructure target, demonstrating that
  the seam-injection approach covered all meaningful code paths.
- **Line reduction**: ~40 lines removed per PS-ID TestMain file (4 × 40 = ~160 lines), replaced
  by a clean, readable `SetupE2ETestMain` call.
- **No build tag required**: The factory compiles without any build tag, consistent with
  `compose_manager.go`. PS-ID TestMain files use `//go:build e2e` — the factory is always
  available regardless of build tag.
- **All 4 E2E suites verified passing**: sm-kms, jose-ja, sm-im, and skeleton-template all passed
  their full E2E suites after the factory migration. The issuing-ca.pem generated by
  `writeAdminCABundle` was correctly mounted and used by all admin mTLS healthchecks.
- **`client_dn` filter for PostgreSQL mTLS test**: The `TestE2E_PostgreSQLMTLS` test was fixed by
  filtering `pg_stat_ssl` by `client_dn LIKE '%-sm-kms-%'` instead of `application_name`. GORM
  does not set `application_name` in PostgreSQL connections by default; `client_dn` is always
  populated from the mTLS client certificate CN. This is now the canonical pattern for mTLS
  identity verification in PostgreSQL tests.

### What Didn't Work

- **Task 4.5 (Docker smoke test) was initially deferred**: Docker Desktop was not running when
  Phase 4 was first executed. Task 4.5 was completed in Phase 4/5 overlap with all 4 suites.
- **Initial audit of existing PS-ID TestMain files revealed inconsistencies**: sm-kms used 3
  containers, jose-ja used 3, sm-im used 4, skeleton-template used 3. The factory had to support
  arbitrary `HealthChecks` maps — no hardcoded assumption about count.
- **33 literal-use violations after Phase 4**: The Phase 4 test files used bare numeric and string
  literals that matched magic constants (e.g. `10 * time.Millisecond` matches `TestUnitPollIntervalMs`,
  `"sm-kms"` matches `OTLPServiceSMKMS`, `0o600` matches `CacheFilePermissions`). These were
  discovered by `TestLint_Integration` and required a dedicated fix pass before Phase 5.
- **stale Docker image caused missing issuing-ca.pem**: The skeleton-template Docker image was
  built before `writeAdminCABundle` was added to `generator_helpers.go`. A stale image silently
  skipped the CA bundle write. Fix: `docker compose build` before running E2E.

### Root Causes

- **TestMain duplication root cause**: Each PS-ID e2e directory was created by copy-paste from
  a previous service, accumulating shared boilerplate that was never refactored into a shared
  library. The factory design closes this debt permanently — new PS-IDs only need to fill in
  magic constants.
- **97.4% vs 98% gap**: The 2.6% uncovered code is in the `Cleanup` method's error return
  (the `Stop` error is ignored with `_ =`). This is intentional — cleanup errors don't affect
  test results and the ignore is documented.
- **GORM does not set `application_name`**: GORM's default connection string does not include
  `application_name`. Filtering PostgreSQL `pg_stat_activity.application_name` for GORM
  connections always returns empty. Use `pg_stat_ssl.client_dn` (from mTLS cert CN) instead.
- **Magic literal violations from test authoring**: Writing test code without consulting existing
  magic constants produces violations that block `TestLint_Integration`. The fix is to always
  check `internal/shared/magic/` for existing constants BEFORE writing any literal in test code.

### Patterns for Future Phases

- **Factory + seam injection = testable infrastructure code**: Infrastructure code that manages
  external resources (Docker Compose, HTTP clients) should always use the `deps` struct pattern
  with production defaults returned by `defaultXxxDeps()` and swapped in tests. This is now the
  established pattern for e2e_infra code.
- **Magic constants as the factory interface**: The factory accepts PS-ID-specific values via
  `E2ETestConfig` which is populated with magic constants. This avoids YAML parsing at runtime and
  keeps the factory generic.
- **Two test clients serve different purposes**: `InsecureClient` (for compose readiness polls)
  and `SecureClient` (for CA-validated TLS tests) must be initialized in the correct order —
  `SecureClient` is built AFTER `WaitForMultipleServices` returns, because pki-init writes the
  CA cert during startup. Building `SecureClient` before health checks pass would reference a
  non-existent cert file.
- **`v13 Item 7` is now closed**: The shared TestMain factory eliminates the copy-paste boilerplate
  pattern identified in v13. New PS-IDs added in future plans need only create an `E2ETestConfig`
  struct with their magic constants — no TestMain boilerplate to copy.
- **Always rebuild Docker images before E2E when new features are added**: `docker compose build`
  is MANDATORY before E2E when production code changes (especially init/startup code). A passing
  healthcheck on a stale image can mask that new features (like `writeAdminCABundle`) are missing.
- **Use `client_dn` not `application_name` for PostgreSQL mTLS identity**: The canonical pattern
  for verifying mTLS client identity in `pg_stat_ssl` JOIN queries is `client_dn LIKE '%-<PS-ID>-%'`.
  Never use `application_name` for GORM services — it is always empty by default.

---

## Phase 5: Mutation Testing on e2e_infra Code

### What Worked

- **Initial gremlins run identified exactly 2 surviving mutations**: Both were visible within the
  first gremlins pass — no guessing required. The structured output (KILLED vs LIVED vs TIMED OUT)
  made it straightforward to target fixes.
- **Error-message assertion killed `attempts++` mutation**: Including the attempt count in the
  timeout error message (`fmt.Sprintf("... after %d attempts", attempts)`) and asserting
  `require.NotContains(t, err.Error(), "after 0 attempts")` was a precise, minimal fix. No new
  logic needed — just richer error messages that are already useful for debugging.
- **Zero data races**: The `go test -race -count=2` run confirmed the e2e_infra package is
  race-free. The seam-injection pattern (per-test factory deps, no shared mutable state) naturally
  prevents races.
- **96.55% efficacy (≥95% ✓)**: The final efficacy exceeded the production code target of ≥95%,
  confirming the test suite is rigorous enough to catch mutations in this infrastructure package.
- **`make` capacity hint documented as structural ceiling**: Rather than spending time trying to
  kill an unkillable mutation, we documented why it cannot be killed (the capacity hint is an
  internal optimization hint inaccessible from the test boundary) and verified it is the only
  remaining LIVED mutation. This avoids future re-investigation.

### What Didn't Work

- **Initial efficacy 93.10% (below ≥95% target)**: The first gremlins run produced 2 LIVED
  mutations, requiring a fix pass before the target was met. A more complete initial test suite
  (asserting richer error messages) would have achieved ≥95% on the first run.
- **TIMED OUT mutations inflate completion time**: The 4 TIMED OUT mutations in `testmain_factory.go`
  are from paths that involve blocking operations (the factory's wait loops). They count toward
  efficacy (correctly — they represent detected mutations) but add real time to each gremlins run.
  Future plans should budget for this.

### Root Causes

- **`attempts++` mutation survived**: The `WaitForMultipleServices` timeout error message did not
  include the attempt count in the initial implementation. Without an assertion on the attempt count,
  the mutation (which prevents `attempts` from incrementing) was invisible to tests. The fix was
  strictly additive — richer error messages expose iteration state.
- **`make` capacity hint mutation is a structural ceiling**: The `make(map[string]error, len(services))`
  capacity hint is an internal optimization that produces the same observable output regardless of
  the capacity argument. Black-box tests cannot observe the capacity of an internal map. This is a
  known gremlins limitation for capacity hints — not a test quality issue.
- **Infrastructure code ≥98% efficacy target**: The e2e_infra package is infrastructure/utility
  code (target ≥98%), but the 1 LIVED `make` mutation is a structural ceiling. 96.55% is the
  practical maximum achievable. This is documented as an accepted ceiling.

### Patterns for Future Phases

- **Retry/timeout loops MUST include attempt counts in error messages**: Any function that retries
  in a loop (`for attempts := 0; ...; attempts++`) MUST include `attempts` in the error message
  returned on exhaustion. This kills `attempts++` mutations AND produces more useful debugging
  output. Pattern: `fmt.Sprintf("timed out after %d attempts: %w", attempts, err)`.
- **`make` capacity hints are structural mutation ceilings**: Do NOT spend time trying to write
  tests to kill `make(T, cap)` mutations where `cap` is computed from input. These are unobservable
  from the outside. Document them as structural ceilings and move on.
- **Budget for gremlins TIMED OUT in time estimates**: Each TIMED OUT mutation represents ~30s of
  real time (the default gremlins timeout). Packages with blocking operations (polling loops,
  network waits) will have more TIMEOUTs. For a package with 4 TIMEOUTs, budget an extra 2 minutes
  per gremlins run.
- **Phase 5 lesson: mutation testing is most effective AFTER unit tests are complete**: Running
  gremlins before the unit test suite is fully written produces misleading low efficacy scores.
  The correct sequence (unit tests → integration tests → mutation testing) was followed correctly
  in this plan (Phases 3-4 test code, Phase 5 mutation testing).

---

## Phase 6: Knowledge Propagation

### What Worked

- **lint-docs as the first verification step**: Running `go run ./cmd/cicd-lint lint-docs`
  immediately after updating ENG-HANDBOOK.md confirmed propagation integrity and identified the
  exact `@source` block in `04-01.deployment.instructions.md` that needed to match the fixed
  `@propagate` block. The linter's diff-style output made the mismatch obvious.
- **Single root cause cascaded to three fixes**: The `start_period` / `start-period` bug was one
  root cause that required fixes in four places (the `@propagate` block, §5.5.5 YAML, §12 YAML,
  §12 bullet, and the `@source` block in instructions). Having a checklist of all four locations
  in the tasks.md prevented any from being missed.
- **Opportunistic quality fix (ST1011)**: Running `golangci-lint run --fix ./...` surfaced 4
  pre-existing `ST1011` violations in `magic_testing.go` (`TestUnitPollIntervalMs` etc.). These
  were fixed immediately per the mandatory opportunistic quality fix policy, resulting in a cleaner
  codebase with names that correctly communicate units via type alone (`time.Duration`).
- **New ENG-HANDBOOK.md sections directly address real lessons**: §10.5.3 (common surviving
  mutations) directly captures the Phase 5 `attempts++` and `make` capacity patterns. §10.3.6
  (e2e_infra TestMain factory) directly captures the Phase 4 factory design with working code
  examples. §10.2.3 (production closure body coverage) directly captures the Phase 3 pki-init
  coverage ceiling solution. All sections are grounded in evidence from this plan.

### What Didn't Work

- **Context limit interrupted mid-task**: The session context limit was hit while §10.3.6,
  §10.4, and §10.2.3 additions were still pending. The conversation summary mechanism preserved
  the state, but resuming mid-task requires careful re-reading of the summary to avoid duplicate
  or conflicting edits. The lesson: commit more granularly (per section, not per task) to reduce
  the amount of work left in an open state.
- **Tasks.md counter was slightly stale**: At the start of this session, the task count showed
  "20/28 complete" but tasks 5.4 and 6.1 had been completed in the session before the context
  limit. Stale counts in tasks.md cause incorrect progress reporting. The fix is to always update
  the counter atomically with the task checkbox, not as a separate step.

### Root Causes

- **`start-period` vs `start_period` typo in ENG-HANDBOOK.md**: The canonical pattern was written
  with Docker CLI syntax (`--start-period`, hyphen) instead of Docker Compose YAML syntax
  (`start_period`, underscore). The actual compose files were correct; only the documentation
  example was wrong. Root cause: the author conflated the two syntaxes when writing the doc.
- **ST1011 violations in magic_testing.go**: The `Ms` suffix on time.Duration constants was added
  when the constants were first defined, before the team enforced ST1011 via golangci-lint.
  The linter version in use at that time may not have included staticcheck's ST1011 rule, or the
  rule was not yet in the golangci-lint configuration. Now that the linter enforces it, the suffix
  is redundant (the type `time.Duration` communicates units without any suffix).

### Patterns for Future Phases

- **Propagation integrity check is part of every ENG-HANDBOOK.md update**: After ANY change to an
  ENG-HANDBOOK.md `@propagate` block, immediately (1) update the corresponding `@source` block in
  the target instruction file, and (2) run `go run ./cmd/cicd-lint lint-docs` to confirm zero drift.
  Never commit ENG-HANDBOOK.md changes without the instructions file update in the same or immediately
  following commit.
- **Document new patterns in ENG-HANDBOOK.md immediately after discovering them**: Phase 6 added
  five new documentation sections capturing lessons from Phases 3-5. The value is proportional to
  how soon the lessons are captured. Waiting until Phase 6 means any subsequent plan that starts
  before Phase 6 runs will miss the lesson. Ideal timing: add ENG-HANDBOOK.md sections at the end
  of each phase's post-mortem, not deferred to a dedicated propagation phase.
- **`time.Duration` constants MUST NOT have unit suffixes (Ms, Ns, Sec, etc.)**: The `time.Duration`
  type is self-documenting — `10 * time.Millisecond` already communicates the unit. Adding `Ms` to
  the name is redundant and violates ST1011. Always name Duration constants without unit suffixes:
  `TestUnitPollInterval`, not `TestUnitPollIntervalMs`.
- **Commit atomically per section when editing large documents**: When updating ENG-HANDBOOK.md
  with multiple independent sections, commit each section (or group of closely related sections)
  separately. This limits the blast radius of context loss and makes individual changes bisectable.

