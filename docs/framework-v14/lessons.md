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

*(To be filled during Phase 3 execution using the 4-section structure above)*

---

## Phase 4: E2E Framework Redesign — Shared TestMain Factory

*(To be filled during Phase 4 execution using the 4-section structure above)*

---

## Phase 5: Mutation Testing on e2e_infra Code

*(To be filled during Phase 5 execution using the 4-section structure above)*

---

## Phase 6: Knowledge Propagation

*(To be filled during Phase 6 execution using the 4-section structure above)*
