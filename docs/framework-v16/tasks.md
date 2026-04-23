# Tasks - Framework V16: LLM Token Efficiency + Framework Lifecycle + V15 Propagation

**Status**: 0 of 34 tasks complete (0%)
**Last Updated**: 2026-04-22
**Created**: 2026-04-22

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL code functionally correct; comprehensive tests |
| Completeness | NO phases/tasks/steps skipped; NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Reliability | ≥95% production coverage; ≥98% infrastructure coverage |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Task Status Legend — MANDATORY

| Symbol | Meaning |
|--------|---------|
| ❌ | Not started |
| 🔄 | In progress |
| ✅ | Complete (evidence required) |
| ⏳ | Blocked (resolution plan required) |

---

## Phase 0: LLM Token Efficiency Audit & Implementation

**Phase Objective**: Reduce token consumption for all LLM agent sessions. Target ≥200 lines removed
from instruction files; cicd-lint gains `-q` summary mode; new tool efficiency instruction file;
GitHub Actions steps gain `::group::` collapse; pre-commit hooks suppress verbose passing output.

### Task 0.1: Instruction File Cross-Reference Pruning

- **Status**: ✅
- **Estimated**: 2h
- **Actual**: 0.5h
- **Dependencies**: None
- **Description**: Remove 197 redundant `See [ENG-HANDBOOK.md Section X.Y...]` trailing lines from
  instruction files. These are glue text OUTSIDE `@source` blocks and can be removed without
  affecting `lint-docs` validation. Also remove end-of-file "Cross-References" sections that
  repeat refs already distributed through the file. Audit for identical-structure bullets convertible
  to tables.
- **Acceptance Criteria**:
  - [x] `wc -l .github/instructions/*.instructions.md` shows ≥200 lines removed vs baseline (3046→2641 = 405 removed)
  - [x] `go run ./cmd/cicd-lint lint-docs` passes (no propagation drift)
  - [x] Agent-drift linter passes (instruction changes don't affect `@source` blocks)
  - [x] No `@source` block content was modified
- **Files**:
  - `.github/instructions/*.instructions.md` (all 18 files)

### Task 0.2: Agent File Boilerplate Compaction

- **Status**: ✅
- **Estimated**: 1h
- **Actual**: 0.25h
- **Dependencies**: None
- **Description**: Compact non-behavioral prose in agent files. Remove duplicate cross-cutting
  tables in `implementation-planning.agent.md`. Compact Prohibited Stop Behaviors from multi-line
  bullets to concise one-liners. Verify dual canonical sync (Copilot `.agent.md` ↔ Claude `.md`).
  NOTE: Quality mandate, continuous execution section, and all workflow steps MUST remain in full
  — agent isolation principle requires self-containment.
- **Acceptance Criteria**:
  - [x] `wc -l .github/agents/*.agent.md` shows measurable reduction (7453→7349 = 104 lines removed)
  - [x] `go run ./cmd/cicd-lint lint-docs` passes (`lint-agent-drift` sub-check)
  - [x] Both Copilot and Claude canonical pairs updated in sync
- **Files**:
  - `.github/agents/implementation-planning.agent.md`
  - `.github/agents/implementation-execution.agent.md`
  - `.github/agents/beast-mode.agent.md`
  - `.claude/agents/implementation-planning.md`
  - `.claude/agents/implementation-execution.md`
  - `.claude/agents/beast-mode.md`

### Task 0.3: New Instruction File — 06-03.tool-efficiency.instructions.md

- **Status**: ✅
- **Estimated**: 0.5h
- **Actual**: 0.2h
- **Dependencies**: None
- **Description**: Create `.github/instructions/06-03.tool-efficiency.instructions.md` with
  codified token-efficient tool use patterns (F1–F8 from plan.md). Add to
  `.github/copilot-instructions.md` instruction table. Add corresponding Claude instruction file.
  Verify `lint-docs` passes.
- **Acceptance Criteria**:
  - [x] File created and valid YAML frontmatter
  - [x] All 8 rules (F1–F8) documented
  - [x] Referenced in `.github/copilot-instructions.md` instruction table
  - [x] `go run ./cmd/cicd-lint lint-docs` passes
  - [x] `CLAUDE.md` updated if instruction files are listed there
- **Files**:
  - `.github/instructions/06-03.tool-efficiency.instructions.md` (NEW)
  - `.github/copilot-instructions.md`
  - `CLAUDE.md`

### Task 0.4: cicd-lint Summary/Quiet Mode

- **Status**: ✅
- **Estimated**: 2h
- **Actual**: 2.5h
- **Dependencies**: None
- **Description**: Add `--summary` (`-q`) flag to cicd-lint runner. On success: one line per linter
  `lint-X: PASS (N files)`. On failure: errors only, no per-file passing output. Default behavior
  (verbose) unchanged. Update all 14 linter `Lint()` signatures to accept a `summary bool`
  parameter. Add tests for quiet mode output format.
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd-lint lint-text -q` outputs single summary line (no per-file output)
  - [x] `go run ./cmd/cicd-lint lint-text -q` with errors still shows error details
  - [x] All 14 linters support `-q` flag (via NewQuietLogger routing)
  - [x] `go test ./internal/apps/tools/cicd_lint/...` passes; 98.3% coverage maintained
  - [x] `golangci-lint run ./internal/apps/tools/cicd_lint/...` clean
- **Files**:
  - `internal/apps/tools/cicd_lint/cicd_lint.go`
  - `internal/apps/tools/cicd_lint/lint_*/` (all 14 linter packages, update Lint signature)
  - `internal/apps/tools/cicd_lint/cicd_lint_test.go` (update tests)

### Task 0.5: GitHub Actions Step Compaction

- **Status**: ✅
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: None
- **Description**: Wrap verbose steps in `::group::` / `::endgroup::` annotations in CI workflows.
  Add `--quiet` to golangci-lint in CI. Remove `-v` from `go test` in CI (failures still print
  without `-v`). Add `--progress=quiet` to docker build. Remove `continue-on-error: true` from
  quality gate steps (raises removal ticket if intentional) and downscope `pull-requests: write`.
- **Acceptance Criteria**:
  - [x] All golangci-lint steps have `--quiet` flag (golangci-lint action updated)
  - [x] All `go test` CI steps use `-count=1` without `-v` (ci-coverage.yml updated)
  - [x] All docker build steps have `--progress=quiet` (ci-quality.yml updated)
  - [x] GitHub Actions log groups collapse verbose output in UI (::group:: added to build/test steps)
  - [x] `continue-on-error: true` on quality gate steps have tracking comments (ci-race.yml documented)
  - [x] Docker Scout steps are optional scanning (continue-on-error acceptable for non-gate steps)
  - [x] cicd-lint steps wrapped with ::group:: in ci-quality.yml
- **Files**:
  - `.github/workflows/ci-quality.yml`
  - `.github/workflows/ci-coverage.yml`
  - `.github/workflows/ci-race.yml`
  - `.github/actions/golangci-lint/action.yml`

### Task 0.6: Pre-Commit Hook Verbosity Reduction

- **Status**: ✅
- **Estimated**: 0.5h
- **Actual**: 0.1h
- **Dependencies**: Task 0.4
- **Description**: After cicd-lint gains `-q` support, update pre-commit hooks to use quiet mode.
  Verify golangci-lint in pre-commit already supports `-q` flag. Add `args: [--quiet]` where
  missing.
- **Acceptance Criteria**:
  - [x] `cicd-lint-all`, `cicd-format-go`, `cicd-format-go-test` hooks use `-q` flag
  - [x] `golangci-lint` and `golangci-lint-full` hooks use `--quiet` flag
  - [x] Failures still show actionable error details (only success messages suppressed)
  - [x] `.pre-commit-config.yaml` updated with quiet args
- **Files**:
  - `.pre-commit-config.yaml`

### Phase 0 Quality Gate

- [ ] All tests pass: `go test ./...` (100%, zero skips)
- [x] Build clean: `go build ./...` AND `go build -tags e2e,integration ./...`
- [x] Lint clean: `golangci-lint run ./...` (0 issues)
- [x] cicd-lint passes: `go run ./cmd/cicd-lint -q lint-text lint-go lint-go-test lint-fitness lint-docs lint-workflow`
- [x] Instruction file size: ≥200 lines removed (405 removed in Task 0.1)
- [x] Update `lessons.md` Phase 0 section with post-mortem

---

## Phase 1: Framework Lifecycle Helpers + Usage Deduplication

**Phase Objective**: Extract signal/shutdown pattern to `lifecycle` package. Complete GAP-0.5
(usage string deduplication). All 10 PS-ID entry points and 8 usage.go files updated. Both
new packages at ≥98% test coverage.

### Task 1.1: Create lifecycle Package

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: None
- **Description**: Create `internal/apps/framework/service/lifecycle/` package. Implement
  `RunService(ctx, stdout, stderr, srv Starter) int` and `RunWithGracefulShutdown(ctx, stdout,
  stderr, startFn, shutdownFn) int`. Include signal.Notify, errChan, select, signal.Stop,
  close(sigChan), context.WithTimeout with `magic.DefaultDataServerShutdownTimeout`, and error
  logging. Table-driven tests with injected start/shutdown functions (seam injection pattern).
- **Acceptance Criteria**:
  - [ ] Package compiles and `go test ./internal/apps/framework/service/lifecycle/...` passes
  - [ ] ≥98% line coverage (infrastructure package)
  - [ ] Handles: server error, SIGINT, SIGTERM, shutdown error
  - [ ] `golangci-lint run` clean on new package
  - [ ] `t.Parallel()` on all tests; table-driven for signal/error combinations
- **Files**:
  - `internal/apps/framework/service/lifecycle/lifecycle.go` (NEW)
  - `internal/apps/framework/service/lifecycle/lifecycle_test.go` (NEW)

### Task 1.2: Migrate 5 Core PS-ID Entry Points to lifecycle.RunService

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.1
- **Description**: Update sm-kms, sm-im, jose-ja, pki-ca, skeleton-template entry points to use
  `lifecycle.RunService()`. Remove duplicate signal handling code (~25 lines per file). Verify
  all tests still pass.
- **Acceptance Criteria**:
  - [ ] `go build ./cmd/sm-kms/...` `./cmd/sm-im/...` `./cmd/jose-ja/...` `./cmd/pki-ca/...` `./cmd/skeleton-template/...` clean
  - [ ] All existing tests pass
  - [ ] No signal handling code remaining in entry point files (grep verified)
  - [ ] `golangci-lint run` clean
- **Files**:
  - `internal/apps/sm-kms/kms.go`
  - `internal/apps/sm-im/im.go`
  - `internal/apps/jose-ja/ja.go`
  - `internal/apps/pki-ca/ca.go`
  - `internal/apps/skeleton-template/template.go`

### Task 1.3: Migrate 5 Identity PS-ID Entry Points to lifecycle.RunService

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.2
- **Description**: Update identity-authz, identity-idp, identity-rp, identity-rs, identity-spa
  entry points. These are larger codebases — verify no behavioral changes.
- **Acceptance Criteria**:
  - [ ] All 5 identity service entry points updated
  - [ ] `go build ./cmd/identity-authz/... ./cmd/identity-idp/... ./cmd/identity-rp/... ./cmd/identity-rs/... ./cmd/identity-spa/...` clean
  - [ ] `go test ./internal/apps/identity-authz/... ./internal/apps/identity-idp/...` passes
  - [ ] `golangci-lint run ./internal/apps/identity-*/...` clean
- **Files**:
  - `internal/apps/identity-authz/authz.go`
  - `internal/apps/identity-idp/idp.go`
  - `internal/apps/identity-rp/rp.go`
  - `internal/apps/identity-rs/rs.go`
  - `internal/apps/identity-spa/spa.go`

### Task 1.4: Create usage Package (GAP-0.5 Completion)

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: None
- **Description**: Create `internal/apps/framework/service/usage/` package with `BuildUsageMain()`,
  `BuildUsageServer()`, `BuildUsageClient()`, `BuildUsageHealth()`, `BuildUsageLivez()`,
  `BuildUsageReadyz()`, `BuildUsageShutdown()`. Table-driven tests with parameterized inputs.
  Each function takes named parameters for product name, service name, config file path.
- **Acceptance Criteria**:
  - [ ] Package compiles; `go test ./internal/apps/framework/service/usage/...` passes
  - [ ] ≥98% coverage (infrastructure package)
  - [ ] Output strings match current usage.go `var` values exactly
  - [ ] `t.Parallel()` on all tests; table-driven for all string builder functions
- **Files**:
  - `internal/apps/framework/service/usage/usage.go` (NEW)
  - `internal/apps/framework/service/usage/usage_test.go` (NEW)

### Task 1.5: Migrate 4 PS-ID usage.go Files to usage.Build*()

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.4
- **Description**: Update sm-kms, sm-im, jose-ja, pki-ca usage.go files to use `usage.Build*()`
  functions. Change `const` → `var` for usage strings (required for non-compile-time values).
  Verify existing usage strings match output of new helpers.
- **Acceptance Criteria**:
  - [ ] 4 usage.go files updated: `sm-kms/usage.go`, `sm-im/usage.go`, `jose-ja/usage.go`, `pki-ca/usage.go`
  - [ ] No more raw usage string literals (grep verified)
  - [ ] `golangci-lint run ./...` clean (no `mnd` or `literal-use` violations)
  - [ ] `go build ./...` clean
- **Files**:
  - `internal/apps/sm-kms/sm_kms_usage.go` (or equivalent filename — verify actual name)
  - `internal/apps/sm-im/sm_im_usage.go`
  - `internal/apps/jose-ja/jose_ja_usage.go`
  - `internal/apps/pki-ca/pki_ca_usage.go`

### Task 1.6: Migrate 3 Product-Level usage.go Files to usage.Build*()

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.5
- **Description**: Update sm, jose, pki product-level usage.go files. These contain usage strings
  for the product-level CLI wiring. Verify actual file names and locations first.
- **Acceptance Criteria**:
  - [ ] 3 product usage.go files updated
  - [ ] `go build ./cmd/sm/... ./cmd/jose/... ./cmd/pki/...` clean
  - [ ] No raw usage string literals in product usage files
- **Files**:
  - `internal/apps/sm/sm_usage.go` (or equivalent — verify actual filename)
  - `internal/apps/jose/jose_usage.go`
  - `internal/apps/pki/pki_usage.go`

### Task 1.7: Confirm GAP-0.5 Work Delivered

- **Status**: ❌
- **Estimated**: 0.1h
- **Dependencies**: Tasks 1.4, 1.5, 1.6
- **Description**: Confirm usage package and all 8 usage.go migrations are complete. Note:
  `docs/framework-v15/` was deleted pre-V16 as part of plan cleanup (all 46 V15 tasks were
  verified ✅ complete before deletion). The gap-0.5 file and V15 tasks.md no longer exist.
- **Acceptance Criteria**:
  - [x] `docs/framework-v15/gap-0.5-usage-deduplication.md` deleted (done pre-V16)
  - [x] `docs/framework-v15/` directory deleted (done pre-V16; all tasks verified complete)
  - [ ] Confirm `go build ./...` clean and `golangci-lint run ./...` clean after usage migrations
- **Files**:
  - Verification only (no file writes)

### Phase 1 Quality Gate

- [ ] All tests pass: `go test ./...` (100%, zero skips)
- [ ] Build clean: `go build ./...` AND `go build -tags e2e,integration ./...`
- [ ] Lint clean: `golangci-lint run ./...`
- [ ] Coverage: `lifecycle` package ≥98%; `usage` package ≥98%
- [ ] No signal handling code in any PS-ID entry point (grep `sigChan\|signal\.Notify` in entry points = 0 results)
- [ ] No raw usage string literals in migrated files (grep `literal-use` violations = 0)
- [ ] Update `lessons.md` Phase 1 section with post-mortem

---

## Phase 2: V15 Knowledge Propagation

**Phase Objective**: Apply V15 lessons to ENG-HANDBOOK.md (14 sections), magic package (Cat 3 CN
constants), instruction files (4 files), and generator call site comments. All changes verified by
`lint-docs`.

### Task 2.1: ENG-HANDBOOK.md — OTel and Grafana Patterns (§9.4)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add to §9.4 (Telemetry Strategy):
  - `client_ca_file` in OTel `tls:` block enables server-side mTLS enforcement (requires client cert)
  - `OTELCOL_EXTRA_ARGS` environment variable pattern for Grafana LGTM embedded OTel Collector config injection
  - Container endpoint naming: `service-name:container-port` (inter-service); host test endpoint: `127.0.0.1:host-port`
- **Acceptance Criteria**:
  - [ ] Three new patterns documented in §9.4
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes (no drift)
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.2: ENG-HANDBOOK.md — Testing Patterns (§10)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add to §10:
  - TLS rejection tests MUST assert `err.Error()` contains `"tls"` (not just `require.Error`)
  - `//go:build e2e` tag MUST be present on ALL files in an E2E package (not just test files)
  - After `golangci-lint --fix`, always re-run `golangci-lint run` (auto-fix creates new violations)
  - `create_file` tool in VS Code agent prepends `package` statement before file content → copyright header must be literally first in file content string
- **Acceptance Criteria**:
  - [ ] Four patterns added to §10
  - [ ] `lint-docs` passes
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.3: ENG-HANDBOOK.md — Deployment Patterns (§12, §3.4, §6.5)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add/update:
  - §12.3: Canonical template sync MUST happen in same commit as deployment config change
  - §3.4: Port offset +10000 for E2E test-expose ports (avoids Grafana 3000 range)
  - §6.5: Cat 4 CA scope — shared across postgres variants (same trust domain), isolated per SQLite variant
  - §12: `./certs:/certs:ro` bind mount is structural requirement for TLS in Docker Compose
  - §11: `lint-deployments` as mandatory post-phase gate when `deployments/` or `configs/` changed
- **Acceptance Criteria**:
  - [ ] Five items added/updated
  - [ ] `lint-docs` passes
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.4: ENG-HANDBOOK.md — CI/CD Quality Gate Patterns (§9.7, §11)

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None
- **Description**: Add:
  - §9.7: `continue-on-error: true` on quality gate steps is a suppressor anti-pattern; requires tracking comment with removal ticket
  - §9.7: `pull-requests: write` at workflow level is over-scope; use per-job minimum permissions
- **Acceptance Criteria**:
  - [ ] Two patterns documented in §9.7
  - [ ] `lint-docs` passes
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.5: ENG-HANDBOOK.md — LLM Token Efficiency (§14.9)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 0 complete
- **Description**: Add new §14.9 "LLM Agent Token Efficiency Strategy" documenting the patterns
  implemented in Phase 0: instruction file cross-ref pruning, `::group::` annotations, cicd-lint
  `-q` mode, tool preference order (grep_search > semantic_search), read_file targeting.
  This section becomes the source of truth for future efficiency improvements.
- **Acceptance Criteria**:
  - [ ] New §14.9 section added
  - [ ] Includes table of tool preferences (F1–F8)
  - [ ] Includes cicd-lint summary mode documentation
  - [ ] `lint-docs` passes; propagation to `06-03.tool-efficiency.instructions.md` (if applicable)
- **Files**: `docs/ENG-HANDBOOK.md`, possibly `06-03.tool-efficiency.instructions.md`

### Task 2.6: Magic Package — Cat 3 CN Constants

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Add Cat 3 public HTTPS server entity cert CN constants to
  `internal/shared/magic/magic_pki_tls.go` (new file to keep magic_pki.go manageable). One
  constant per PS-ID × variant (10 PS-IDs × 4 variants = 40 constants). These are referenced
  in pki-init and E2E tests; must be in magic to satisfy `literal-use` linter.
- **Acceptance Criteria**:
  - [ ] 40 Cat 3 CN constants defined (10 PS-IDs × 4 variants: sqlite-1, sqlite-2, postgres-1, postgres-2)
  - [ ] File is `internal/shared/magic/magic_pki_tls.go` with correct package declaration
  - [ ] `golangci-lint run ./internal/shared/magic/...` clean
  - [ ] `go build ./...` clean (no undefined references)
  - [ ] magic package excluded from coverage threshold (constants only, no executable logic)
- **Files**:
  - `internal/shared/magic/magic_pki_tls.go` (NEW)

### Task 2.7: Generator Call Site Comments Verification

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None
- **Description**: Verify `internal/apps/framework/tls/generator.go` has `// Cat N: <name>`
  comments at all 14 category call sites. If any are missing, add them. These comments are
  required per §14.1.2 (Multi-Category Generator Call Sites).
- **Acceptance Criteria**:
  - [ ] All 14 Cat N call sites have `// Cat N: <name>` comments
  - [ ] `golangci-lint run` clean (godot: comments end with period? Verify format)
  - [ ] grep confirms: `grep -c "// Cat [0-9]" generator.go` = 14
- **Files**: `internal/apps/framework/tls/generator.go`

### Task 2.8: Instruction File Updates (4 Files)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Tasks 2.1–2.5
- **Description**: Update 4 instruction files with patterns from V15:
  - `02-03.observability.instructions.md`: OTel `client_ca_file`, `OTELCOL_EXTRA_ARGS`, container vs host port naming
  - `04-01.deployment.instructions.md`: template sync rule, `lint-deployments` post-phase gate, Cat 4 CA scope
  - `03-02.testing.instructions.md`: TLS rejection test `err.Error()` contains `"tls"`; `//go:build e2e` package-wide rule; `golangci-lint --fix` two-pass
  - `03-05.linting.instructions.md`: `golangci-lint --fix` creates new violations → always re-run
- **Acceptance Criteria**:
  - [ ] All 4 instruction files updated
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes (propagation check)
  - [ ] No `@source` blocks modified (only glue text updates)
- **Files**:
  - `.github/instructions/02-03.observability.instructions.md`
  - `.github/instructions/04-01.deployment.instructions.md`
  - `.github/instructions/03-02.testing.instructions.md`
  - `.github/instructions/03-05.linting.instructions.md`

### Phase 2 Quality Gate

- [ ] All tests pass: `go test ./...`
- [ ] Build clean: `go build ./...`
- [ ] Lint clean: `golangci-lint run ./...`
- [ ] `go run ./cmd/cicd-lint lint-docs` passes (no propagation drift)
- [ ] `wc -l docs/ENG-HANDBOOK.md` shows net increase (new sections added)
- [ ] Update `lessons.md` Phase 2 section with post-mortem

---

## Phase 3: V15 Incomplete Work Cleanup

**Phase Objective**: Verify V15 gaps resolved or tracked. Confirm Phase 1 usage work delivered.
Note: `docs/framework-v15/` was deleted pre-V16 as part of plan cleanup. All 46 V15 tasks were
verified ✅ complete, all 12 lessons.md sections verified substantive, and all gaps.md items
verified fixed/deferred/intentional before deletion.

**Pre-V16 cleanup findings (already verified)**:
- All 46 V15 tasks: ✅ COMPLETE
- V15 lessons.md: All 12 sections substantive
- V15 gaps.md audit:
  - Gaps 1.1, 1.2, 1.3, 1.7, 1.9: ✅ fixed in V15 Phase 0
  - Gap 1.5 (ci-race build tags): intentional design (integration tags optional by design)
  - Gap 2.1 (usage dedup): deferred to V16 Phase 1 (gap-0.5) — delivered in this V16
  - Gaps 2.2, 2.3, 2.4: ✅ fixed in V15 Phase 0
  - Gaps 4.1, 4.2: ✅ fixed in V15 Phase 0
  - Gap 5.1 (pki-ca TestMain): ✅ fixed in V15 Phase 0
  - Gap 5.2: confirmed intentional design
  - Gaps 6.1–6.5: ✅ fixed in V15 Task 0.8
  - Gap 7.1: ✅ fixed in V15

### Task 3.1: Confirm V15 Usage Work Delivered

- **Status**: ❌
- **Estimated**: 0.1h
- **Dependencies**: Phase 1 complete
- **Description**: Confirm V16 Phase 1 delivered all usage deduplication work from V15 gap-0.5.
  V15 directory is deleted; this task simply verifies the work is done via build/lint/test.
- **Acceptance Criteria**:
  - [x] `docs/framework-v15/` deleted pre-V16 (already done)
  - [ ] V16 Phase 1 usage package confirmed delivered (`go test ./internal/apps/framework/service/usage/...` passes)
  - [ ] V16 Phase 1 lifecycle package confirmed delivered (`go test ./internal/apps/framework/service/lifecycle/...` passes)
- **Files**: Verification only

### Task 3.2: Confirm Gap 1.5 Disposition (ci-race Build Tags)

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None
- **Description**: Gap 1.5 from V15 gaps.md — ci-race.yml missing build tag exclusions for bench/fuzz/e2e.
  Pre-V16 audit confirmed this is intentional design (integration tags optional). Verify the current
  state of ci-race.yml and confirm no action is needed, or add `--tags=!bench,!fuzz,!e2e` if appropriate.
- **Acceptance Criteria**:
  - [ ] ci-race.yml reviewed
  - [ ] Decision documented: intentional design (no action) OR add build tag exclusions
  - [ ] If change made: `go run ./cmd/cicd-lint lint-docs` passes
- **Files**: `.github/workflows/ci-race.yml` (review only, change if warranted)

### Task 3.3: Phase 3 Post-Mortem

- **Status**: ❌
- **Estimated**: 0.1h
- **Dependencies**: Tasks 3.1, 3.2
- **Description**: Document Phase 3 findings in V16 lessons.md.
- **Acceptance Criteria**:
  - [ ] `docs/framework-v16/lessons.md` Phase 3 section filled with substantive content
- **Files**: `docs/framework-v16/lessons.md`

### Phase 3 Quality Gate

- [ ] V16 Phase 1 usage + lifecycle packages confirmed delivered
- [ ] Gap 1.5 disposition confirmed (intentional design or fix applied)
- [ ] `go build ./...` clean
- [ ] Update `lessons.md` Phase 3 section with post-mortem

---

## Phase 4: V16 Knowledge Propagation

**Phase Objective**: Apply V16 lessons to permanent artifacts. Review all 4 lessons.md sections.
Update ENG-HANDBOOK.md, agents, skills, instructions. Verify `lint-docs` passes.

### Task 4.1: Review V16 Lessons.md and Identify Propagation Items

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 0–3 complete
- **Description**: Read all 4 V16 lessons.md sections. List all actionable propagation items:
  patterns not yet in ENG-HANDBOOK.md, agent improvements, skill improvements.
- **Files**: `docs/framework-v16/lessons.md`

### Task 4.2: Apply V16 Lessons to ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.1
- **Description**: Apply all propagation items identified in Task 4.1. Focus on lifecycle helper
  patterns, token efficiency patterns, usage deduplication strategy.
- **Acceptance Criteria**:
  - [ ] All actionable items applied
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.3: Apply V16 Lessons to Agents and Skills

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 4.1
- **Description**: Update agents and skills where V16 work exposed improvements or new patterns.
  Sync Copilot ↔ Claude canonical pairs after each update.
- **Acceptance Criteria**:
  - [ ] Agent/skill files updated as needed
  - [ ] `lint-agent-drift` and `lint-skill-command-drift` pass
- **Files**: `.github/agents/*.agent.md`, `.claude/agents/*.md` (as needed)

### Task 4.4: Final Quality Gate Verification

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Tasks 4.1–4.3
- **Description**: Run all quality gates in sequence. Verify clean working tree. Commit.
- **Acceptance Criteria**:
  - [ ] `go test ./...` passes
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run ./...` clean
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] `git status --porcelain` returns empty
- **Files**: N/A (verification only)

### Phase 4 Quality Gate

- [ ] All ENG-HANDBOOK.md updates committed
- [ ] `lint-docs` passes with zero violations
- [ ] All agents/skills synced (Copilot ↔ Claude drift = 0)
- [ ] Clean working tree: `git status --porcelain` empty
- [ ] Update `lessons.md` Phase 4 section with post-mortem

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥98% coverage (lifecycle, usage packages — infrastructure)
- [ ] Unit tests ≥95% coverage (production code)
- [ ] Integration tests pass (`go test -tags integration ./...`)
- [ ] Mutation testing ≥95% minimum (≥98% lifecycle, usage packages)
- [ ] Race detector clean: `go test -race -count=2 ./...`

### Code Quality
- [ ] Linting passes: `golangci-lint run ./...`
- [ ] Linting passes: `golangci-lint run --build-tags e2e,integration ./...`
- [ ] No new TODOs without tracking
- [ ] No security vulnerabilities (gosec clean)

---

## Evidence Archive

- `test-output/v16-phase0/` - Token efficiency measurements (wc -l before/after)
- `test-output/v16-phase1/` - lifecycle and usage package coverage results
- `test-output/v16-phase2/` - lint-docs validation evidence
- `test-output/v16-phase3/` - V15 gaps audit results
- `test-output/v16-phase4/` - Final quality gate verification

---

## Notes / Deferred Work

*(Fill during execution — use for items that arise as blockers during a task)*
