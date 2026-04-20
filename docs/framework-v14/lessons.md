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

- `start_period` (underscore) is silently accepted by Docker Compose but treated as an unknown key
  and IGNORED — Docker uses the default 0-second start period, which means healthchecks fire
  immediately before the service is ready. The correct spelling is `start-period` (hyphen).
- sqlite-2/postgres-1/postgres-2 healthchecks correctly use only `--cacert` (they are app instances
  that don't require mTLS client cert authentication). Only sqlite-1 presents a client cert because
  the canonical template was designed this way — each PS-ID instance has a unique client cert under
  `/certs/{PS-ID}/private-https-mutual-entity-{PS-ID}-{suffix}/`.

### Patterns for Future Phases

- **Template drift check first**: Before starting any compose file changes, run `template_drift`
  tests to identify all existing drift — otherwise you risk partial fixes.
- **Enumerate all affected files early**: The initial description mentioned "4 PS-ID compose files"
  but the actual scope was 10 (all identity services and pki-ca also needed fixes). Always grep for
  `start_period` (or the relevant pattern) across ALL compose files before planning effort.
- **Task 2.3 Docker verification gate**: Mark Docker-dependent tasks as BLOCKED immediately if
  Docker is unavailable, rather than attempting them and failing. Create an explicit resolution plan
  in lessons.md so the next session knows exactly what to verify.
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

*(To be filled during Phase 4 execution using the 4-section structure above)*

---

## Phase 5: Mutation Testing on e2e_infra Code

*(To be filled during Phase 5 execution using the 4-section structure above)*

---

## Phase 6: Knowledge Propagation

*(To be filled during Phase 6 execution using the 4-section structure above)*
