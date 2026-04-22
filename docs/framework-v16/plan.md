# Implementation Plan - Framework V16: LLM Token Efficiency + Framework Lifecycle Helpers + V15 Knowledge Propagation

**Status**: Planning
**Created**: 2026-04-22
**Last Updated**: 2026-04-22
**Purpose**: (1) P0-CRITICAL: Reduce LLM agent token consumption — GitHub Copilot Pro + Claude Code
Pro rate limits became a productivity blocker. (2) P1: Extract signal/shutdown/usage patterns into
shared framework helpers — currently copy-pasted across all 10 PS-ID entry points. (3) P2:
Propagate V15 lessons to ENG-HANDBOOK.md, magic constants, and instruction files. (4) P3:
Complete GAP-0.5 (usage string deduplication, deferred from V15 Task 0.5). (5) P4: Identify and
fix work V15 marked complete but did not fully deliver.

**V15 Status**: All 46 tasks marked ✅ — but Task 0.5 (usage deduplication) was deferred with a
GAP file while remaining marked ✅. V16 carries that forward as Phase 3.

---

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL code functionally correct; comprehensive tests |
| Completeness | NO phases/tasks/steps skipped; NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Reliability | ≥95% production coverage; ≥98% infrastructure coverage; ≥95% mutation |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Overview

V16 is organized into four sequential phases, each addressing a distinct concern.

**Section A — Phase 0: LLM Token Efficiency (P0 CRITICAL)**
GitHub Copilot Pro and Claude Code Pro introduced draconian rate limits without warning. V16
Phase 0 audits ALL surfaces that contribute to LLM agent token consumption and implements targeted
reductions. This phase is prerequisite to all other work — if rate limits remain binding, other
phases become unexecutable.

**Section B — Phase 1: Framework Lifecycle Helpers (P1 HIGH)**
Signal handling (`sigChan` + `close(sigChan)`) and graceful shutdown (`context.WithTimeout` +
error log + cancel) are copy-pasted identically across all 10 PS-ID entry points. Extract to
`internal/apps/framework/service/lifecycle/` so new services get correct patterns automatically.
Also complete GAP-0.5 usage string deduplication as part of this phase.

**Section C — Phase 2: V15 Knowledge Propagation (P2 MEDIUM)**
V15 lessons.md contains actionable patterns for ENG-HANDBOOK.md, magic constants, instruction
files, and code comments. These must be applied before V15 is archived.

**Section D — Phase 3: V15 Incomplete Work (P3 MEDIUM)**
Task 0.5 in V15 tasks.md is marked ✅ but a GAP file explicitly records it as deferred. Fix the
discrepancy, surface any other partial completions, and resolve them.

---

## Background

### V15 Completion Status

V15 delivered:
- OTel Collector server TLS (Cat 2 cert + Cat 8 client CA)
- App→OTel client mTLS (Cat 9 per-variant certs)
- Grafana LGTM HTTPS UI + OTLP ingest mTLS (`grafana.ini` + `OTELCOL_EXTRA_ARGS`)
- OTel→Grafana client mTLS (Cat 9 infra cert)
- Public PS-ID app server TLS (Cat 3 server cert + Cat 4 client CA)
- Full pipeline E2E verification via Go TLS dial tests
- Phase 0 pre-flight gap fixes (CI/CD, signal handling, permission scoping)

### V16 Carries Forward

- `gap-0.5-usage-deduplication.md`: usage.go deduplication (8 files, 4 pairs)
- Framework lifecycle helpers: not yet extracted to shared package
- V15 lessons: actionable patterns not yet written to ENG-HANDBOOK.md

---

## Technical Context

**Language**: Go 1.26.1 | **Test DB**: SQLite in-memory (unit/integration), PostgreSQL (E2E)

**Affected Files — Phase 0 (LLM Token Efficiency)**:
```
.github/instructions/03-01.coding.instructions.md      (add tool efficiency section)
.github/instructions/06-02.agent-format.instructions.md (add tool discovery rules)
.github/instructions/*.instructions.md                 (prune redundant cross-refs: 197 lines)
.github/agents/{beast-mode,implementation-execution,implementation-planning}.agent.md (compact)
.claude/agents/{beast-mode,implementation-execution,implementation-planning}.md       (sync)
internal/apps/tools/cicd_lint/                         (add --summary mode: ~8 linter files)
.github/workflows/ci-quality.yml                       (::group:: + --quiet flags)
.github/workflows/ci-test.yml                          (::group:: + -count=1)
.github/workflows/ci-coverage.yml                      (::group:: annotations)
.pre-commit-config.yaml                                (golangci-lint -q flag)
docs/ENG-HANDBOOK.md                                   (add §14.9 LLM agent token efficiency)
```

**Affected Files — Phase 1 (Framework Lifecycle Helpers)**:
```
internal/apps/framework/service/lifecycle/lifecycle.go        (NEW)
internal/apps/framework/service/lifecycle/lifecycle_test.go   (NEW)
internal/apps/framework/service/usage/usage.go               (NEW — from GAP-0.5)
internal/apps/framework/service/usage/usage_test.go          (NEW)
internal/apps/{sm-kms,sm-im,jose-ja,pki-ca,skeleton-template}/   (5 files, entry points)
internal/apps/{identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/  (5 files, entry points)
internal/apps/{sm,jose,pki}/usage.go                          (3 product-level files)
internal/apps/{sm-kms,sm-im,jose-ja,pki-ca}/usage.go         (4 PS-ID-level files)
# Total: 2 new packages + 17 updated files
```

**Affected Files — Phase 2 (Knowledge Propagation)**:
```
docs/ENG-HANDBOOK.md                                   (§9.4, §10.4, §12.3, §6.5 updates)
internal/shared/magic/magic_pki.go                     (Cat 3 CN constants)
internal/apps/framework/tls/generator.go               (// Cat N: <name> comments verified)
.github/instructions/02-03.observability.instructions.md (OTel mTLS patterns)
.github/instructions/04-01.deployment.instructions.md  (template sync, port conventions)
.github/instructions/03-02.testing.instructions.md     (TLS rejection test pattern, e2e build tag)
```

**Affected Files — Phase 3 (V15 Incomplete Work)**:
```
docs/framework-v15/tasks.md                            (fix Task 0.5 status: ✅ → deferred)
docs/framework-v15/gap-0.5-usage-deduplication.md      (delete when Phase 1 completes)
```

---

## Phase 0: LLM Token Efficiency Audit & Implementation

**Status**: ☐ TODO | **Estimated**: 8h

### P0 Problem Statement

GitHub Copilot Pro and Claude Code Pro rate limits are based on token consumption per hour/day/week.
Every LLM agent session ingests:
1. **Instructions** (auto-loaded): 3046 lines across 18 instruction files — ~75,000 tokens per session
2. **Agent files** (on-demand): 586–1291 lines each — self-contained by design but extremely verbose
3. **cicd-lint output**: verbose per-file logging even when all checks pass — forces reading large outputs
4. **GitHub Actions logs**: steps stream full output even for passing checks — inflates context
5. **Plan/task documents**: quality mandates repeated in every plan template — adds no new info after first read

### P0 Improvements — Complete Catalogue

#### Category A: Instruction File Compaction (highest impact)

| Improvement | Detail | Lines Saved (est.) |
|-------------|--------|-------------------|
| A1: Remove redundant trailing cross-refs | 197 "See ENG-HANDBOOK.md Section X for..." lines at end of each bullet — glue text adding no value when propagated content is already present | 197 |
| A2: Convert long prose bullets to tables | Sections with 5+ identical-structure bullets (e.g., HTTP status codes, log levels, registry types) are already tables — verify all converted | ~30 |
| A3: Trim "Cross-References" sections | End-of-file "Cross-References" sections repeat all section-level refs already distributed through the file | ~40 |
| **Total** | | **~267 lines = ~9% reduction** |

**Constraint**: `@source` propagated blocks are verbatim and MUST NOT be modified.
Glue text (section headings, `See` lines, intro paragraphs) is safe to remove.

#### Category B: Agent File Compaction

| Improvement | Detail | Lines Saved (est.) |
|-------------|--------|-------------------|
| B1: Compact Prohibited Stop Behaviors list | 8-item list with 1-line each can be single-sentence per item | ~20 |
| B2: Compact Evidence Collection Pattern | Lengthy examples in planning agent; move long examples to an appendix | ~40 |
| B3: Dedup workflow step tables | `implementation-planning.agent.md` has 2 identical "Cross-Cutting" tables | ~30 |
| **Total** | | **~90 lines** |

**Constraint**: Agents are self-contained by design (do not inherit copilot instructions). The
quality mandate and continuous execution sections MUST remain in full — they are required for
correct agent behavior. Only non-behavioral prose can be trimmed.

#### Category C: cicd-lint Output Compaction (highest runtime impact)

| Improvement | Detail |
|-------------|--------|
| C1: Add `--summary` flag | On success: one line per linter ("lint-text: PASS (1247 files)"). On failure: errors only. Default remains verbose for backward compat. |
| C2: Compact error format | Group errors by file: `path/to/file.go: [rule1, rule2]` instead of one line per violation |
| C3: Add `--quiet` / `-q` shorthand | Alias for `--summary` for pre-commit / CI use |
| C4: Pre-commit uses `-q` | `.pre-commit-config.yaml` passes `-q` to all cicd-lint calls |
| C5: CI uses `-q` | All CI workflow steps pass `-q` to cicd-lint |

**Implementation**: Add `--summary bool` flag to `RegisterLinterFlags()` in
`internal/apps/tools/cicd_lint/` runner. Pass flag through to each linter's `Lint()` function.
Affected linters: all 14 linters in `lint_*/` subdirectories.

#### Category D: GitHub Actions Step Compaction

| Improvement | Detail |
|-------------|--------|
| D1: `::group::` annotations | Wrap verbose steps (golangci-lint, go test, docker build) in `echo "::group::<name>"` / `echo "::endgroup::"` — collapses by default in GitHub UI |
| D2: `--quiet` on golangci-lint | Add `--quiet` flag to golangci-lint invocations in CI (hides passing linters) |
| D3: Drop `-v` from go test | Remove `-v` from `go test` in CI — only failures print with default verbosity |
| D4: `--progress=quiet` on docker build | `docker build --progress=quiet` suppresses build layer output in CI |

**Files**: `.github/workflows/ci-quality.yml`, `ci-test.yml`, `ci-coverage.yml`, `ci-race.yml`

#### Category E: Pre-Commit Hook Verbosity

| Improvement | Detail |
|-------------|--------|
| E1: golangci-lint `-q` in pre-commit | Current: streams all linter output. With `-q`: only failures. |
| E2: gofumpt suppress no-change output | gofumpt outputs nothing for files with no changes by default — verify this is already the case |
| E3: cicd-lint pre-commit uses `-q` | After C4 implemented, passes through to summary mode |

**File**: `.pre-commit-config.yaml`

#### Category F: New Instruction File — 06-03.tool-efficiency.instructions.md

Creates explicit, codified guidance for LLM agents on token-efficient tool use. Currently no
such guidance exists; agents default to semantic_search (most expensive) when any search suffices.

| Rule | Guidance |
|------|----------|
| F1: Prefer grep_search over semantic_search | `grep_search` returns targeted matches. `semantic_search` scans entire workspace. Use `semantic_search` only when query cannot be expressed as regex. |
| F2: Targeted read_file ranges | Always specify `startLine`/`endLine`. Never read entire files (especially large ones) unless absolutely required. Read 50-100 line windows. |
| F3: multi_replace_string_in_file for batch edits | Always batch ≤10 independent edits. Never chain sequential `replace_string_in_file` calls. |
| F4: Session memory for recurring context | Store frequently re-read constants (magic package values, port assignments) in `/memories/session/` to avoid re-reading files. |
| F5: file_search before read_file | Always use `file_search` to confirm file path before `read_file`. Avoids 404 errors and wasted reads. |
| F6: list_dir before file_search | When unsure of directory structure, `list_dir` first (lightweight). |
| F7: grep_search with isRegexp=false for literals | Plain text match is faster and avoids regex escape issues. |
| F8: Avoid parallel semantic_search | `semantic_search` is explicitly flagged as non-parallelizable in tool instructions. |

#### Category G: Docker Build Output (CI-only)

| Improvement | Detail |
|-------------|--------|
| G1: `BUILDKIT_PROGRESS=plain` → `quiet` | Reduces Docker build output from verbose layer-by-layer to summary only |
| G2: Upload build logs as artifacts | Redirect verbose docker output to file; upload as artifact on failure only |

#### Category H: Plan/Task Document Format

| Improvement | Detail |
|-------------|--------|
| H1: Compact Quality Mandate in templates | Replace 8-bullet quality mandate in plan templates with 8-row table (saves ~25% space) — already done in this V16 plan.md |
| H2: Remove per-task cross-cutting evidence sections | The "Evidence Archive" at end of tasks.md rarely used during execution; compress to one-line reference |
| H3: Gap file format compaction | Gap files should be ≤30 lines total using a table format |

**Phase 0 Success Criteria**:
- Instruction files: ≥200 lines removed (verified by `wc -l`)
- cicd-lint: `go run ./cmd/cicd-lint lint-text -q` outputs single summary line when passing
- GitHub Actions: all verbose steps wrapped in `::group::` and use `--quiet` where available
- New instruction file `06-03.tool-efficiency.instructions.md` created and passes `lint-docs`
- All agent file changes pass `lint-agent-drift` check
- **Post-Mortem**: Update lessons.md Phase 0 section.

---

## Phase 1: Framework Lifecycle Helpers + Usage Deduplication

**Status**: ☐ TODO | **Estimated**: 6h | **Dependencies**: Phase 0 complete

**Phase Objective**: Extract the signal handling + graceful shutdown pattern from all 10 PS-ID
entry points into a shared `lifecycle` package. Complete GAP-0.5 (usage string deduplication).
Result: new services get correct patterns automatically; regressions (like V15's `sm-kms` missing
shutdown timeout) become structurally impossible.

### Signal Handling — Current State

Each of the 10 PS-ID entry points contains ~25 lines of identical signal handling:
```go
errChan := make(chan error, 1)
go func() { errChan <- srv.Start(ctx) }()
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
select {
case err := <-errChan: ...
case sig := <-sigChan:
    shutdownCtx, cancel := context.WithTimeout(ctx, magic.DefaultDataServerShutdownTimeout)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil { ... }
}
signal.Stop(sigChan)
close(sigChan)
```

This pattern MUST be identical across all 10 services. Any deviation (e.g., V15's `sm-kms` missing
the timeout, any service missing `close(sigChan)`) is a bug that only fitness linters currently catch.

### Target State

```go
// Each PS-ID entry point reduces to:
return cryptoutilFrameworkLifecycle.RunService(ctx, stdout, stderr, srv)
// OR for explicit control:
return cryptoutilFrameworkLifecycle.RunWithGracefulShutdown(ctx, stdout, stderr,
    func() error { return srv.Start(ctx) },
    func(ctx context.Context) error { return srv.Shutdown(ctx) },
)
```

### lifecycle Package API

```go
package lifecycle

// Starter is implemented by all framework service servers.
type Starter interface {
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
}

// RunService is the canonical entrypoint for PS-ID server subcommands.
// It starts the server, waits for SIGINT/SIGTERM or server error,
// performs graceful shutdown with DefaultDataServerShutdownTimeout, and returns exit code.
func RunService(ctx context.Context, stdout, stderr io.Writer, srv Starter) int

// RunWithGracefulShutdown is the lower-level variant for callers with custom start/shutdown funcs.
func RunWithGracefulShutdown(ctx context.Context, stdout, stderr io.Writer,
    startFn func() error, shutdownFn func(context.Context) error) int
```

### Usage Package API (GAP-0.5 Completion)

```go
package usage

// BuildUsageMain returns the multi-line usage string for the product-level main command.
func BuildUsageMain(productCmd, serviceCmd, serviceName, configFile string) string

// BuildUsageServer returns the usage string for the server subcommand.
func BuildUsageServer(productCmd, serviceCmd, configFile string) string

// BuildUsageClient returns the usage string for the client subcommand.
func BuildUsageClient(productCmd, serviceCmd string) string

// BuildUsageHealth/Livez/Readyz/Shutdown: analogous builders for each subcommand.
```

**Phase 1 Success Criteria**:
- `internal/apps/framework/service/lifecycle/` package: `go test ./...` passes; ≥98% coverage
- `internal/apps/framework/service/usage/` package: `go test ./...` passes; ≥98% coverage
- All 10 PS-ID entry point files updated to use `lifecycle.RunService()`
- All 8 usage.go files updated to use `usage.Build*()` functions
- `go build ./...` clean; `golangci-lint run ./...` clean; no new TODOs
- GAP-0.5 file deleted (work complete)
- **Post-Mortem**: Update lessons.md Phase 1 section.

---

## Phase 2: V15 Knowledge Propagation

**Status**: ☐ TODO | **Estimated**: 4h | **Dependencies**: Phase 1 complete

**Phase Objective**: Apply all actionable V15 lessons to permanent artifacts. Items are grouped by
target artifact. NEVER defer any item — knowledge propagation is the payoff for detailed lessons.

### 2.1: ENG-HANDBOOK.md Updates

| Section | Item | Source Lesson |
|---------|------|---------------|
| §9.4.1 | OTel `client_ca_file` = server-side mTLS enforcement | Phase 2 lesson |
| §9.4.3 | `OTELCOL_EXTRA_ARGS` pattern for Grafana LGTM embedded OTel | Phase 5 lesson |
| §9.4.4 | Container endpoint vs host endpoint naming convention | Phase 6 lesson |
| §10.4 | TLS rejection tests MUST assert `err.Error()` contains `"tls"` | Phase 7 lesson |
| §10.1 | `//go:build e2e` tag MUST be on every file in an E2E package | Phase 4 lesson |
| §12.3 | Canonical template sync in same commit as deploy config change | Phase 2 lesson |
| §3.4 | Port offset +10000 convention: E2E test-expose ports avoid Grafana range | Phase 4/11 lesson |
| §6.5 | Cat 4 CA scope: shared per trust domain (postgres variants), isolated per SQLite | Phase 8 lesson |
| §12.3 | `./certs:/certs:ro` bind mount is structural requirement for TLS in compose | Phase 8 lesson |
| §11 | `lint-deployments` as post-phase gate for deployments/configs changes | Phase 3/10 lesson |
| §11 | `lint-fitness` as post-template-change gate | Phase 9 lesson |
| §14.1 | After `golangci-lint --fix`, always run `golangci-lint run` again (auto-fix creates new violations) | Phase 11 lesson |
| §14.1 | Go file header order: copyright → build tag → package declaration | Phase 11 lesson |
| §9.7 | `continue-on-error: true` on quality gates = suppressor debt; no automated check for removal | Phase 0 lesson |
| §9.7 | `pull-requests: write` at workflow level = over-scope; use per-job minimum | Phase 0 lesson |
| §14.9 | LLM agent token efficiency strategy (from Phase 0 work) | New |

### 2.2: Magic Package — Cat 3 CN Constants

Add named constants to `internal/shared/magic/magic_pki.go` (or new `magic_pki_tls.go`) for all
Cat 3 server cert CNs. These are already used in pki-init tests and E2E tests; the constants must
be in magic to satisfy `literal-use` linter requirements.

```go
// Cat 3: Public HTTPS server entity cert CNs, one per PS-ID per variant.
const (
    Cat3SmKmsSQLite1ServerCN      = "public-https-server-entity-sm-kms-sqlite-1"
    Cat3SmKmsSQLite2ServerCN      = "public-https-server-entity-sm-kms-sqlite-2"
    Cat3SmKmsPostgres1ServerCN    = "public-https-server-entity-sm-kms-postgres-1"
    Cat3SmKmsPostgres2ServerCN    = "public-https-server-entity-sm-kms-postgres-2"
    // ... repeated for all 10 PS-IDs × 4 variants = 40 constants
)
```

### 2.3: Generator Call Site Comments

Verify `internal/apps/framework/tls/generator.go` has `// Cat N: <name>` comments at all 14
category call sites. These comments were established in V15 Phase 1 but must be verified to be
present in the committed code.

### 2.4: Instruction File Updates

| File | Item |
|------|------|
| `02-03.observability.instructions.md` | OTel `client_ca_file` pattern; `OTELCOL_EXTRA_ARGS`; container vs host port naming |
| `04-01.deployment.instructions.md` | Template sync rule; `lint-deployments` as post-phase gate; Cat 4 CA scope |
| `03-02.testing.instructions.md` | TLS rejection test pattern; `//go:build e2e` package-wide rule; `golangci-lint --fix` two-pass |
| `03-05.linting.instructions.md` | `golangci-lint --fix` creates new violations → always re-run |

**Phase 2 Success Criteria**:
- `go run ./cmd/cicd-lint lint-docs` passes (no propagation drift)
- All 14 ENG-HANDBOOK.md sections updated with new patterns
- Cat 3 CN constants present in magic package; `golangci-lint run` clean
- All 4 instruction files updated
- **Post-Mortem**: Update lessons.md Phase 2 section.

---

## Phase 3: V15 Incomplete Work

**Status**: ☐ TODO | **Estimated**: 1h | **Dependencies**: Phase 1 complete

**Phase Objective**: Correct V15 plan integrity issues; surface any remaining gaps from
`docs/framework-v15/gaps.md` not yet addressed.

### 3.1: Fix V15 tasks.md Task 0.5 Status

Task 0.5 in `docs/framework-v15/tasks.md` is marked `✅` but `gap-0.5-usage-deduplication.md`
explicitly records the work as deferred. This is a data integrity violation in the task tracking.

**Fix**:
1. Update Task 0.5 status in `docs/framework-v15/tasks.md` from `✅` to `⏳ Blocked (deferred to V16)`
2. Update overall task count: 13 of 46 complete (28%) instead of 14 of 46
3. Add note referencing `gap-0.5` file

### 3.2: Gaps.md Audit — Remaining Items

Items from `docs/framework-v15/gaps.md` that were not addressed in V15 and are not scheduled in
V16 Phases 0-2:

| Gap ID | Severity | Description | V16 Phase | Status |
|--------|----------|-------------|-----------|--------|
| 1.5 | MEDIUM | `ci-race.yml` missing build tag exclusions | — | Under investigation in V15 (intentional design) |
| 2.3 | MEDIUM | Port type inconsistency in `identity-authz` | Verify fixed | V15 Task 0.6 |
| 6.1–6.5 | MEDIUM | TLS documentation gaps in `tls-structure.md` | Verify fixed | V15 Task 0.8 |

**Action for each**: Verify fix was applied; if not, add as new task in V16 Phase 3.

### 3.3: V15 Lessons.md Completeness Audit

V15 `lessons.md` Phase 12 notes that lessons are only valuable if specific. Verify all 12 phase
sections are filled with substantive content (not just placeholder text). Mark any empty sections.

**Phase 3 Success Criteria**:
- V15 tasks.md Task 0.5 status corrected
- Gaps 2.3, 6.1–6.5 verified as fixed or new tasks created
- V15 lessons.md verified complete
- `gap-0.5-usage-deduplication.md` deleted (Phase 1 delivers the fix)
- **Post-Mortem**: Update lessons.md Phase 3 section.

---

## Phase 4: Knowledge Propagation

**Status**: ☐ TODO | **Estimated**: 2h | **Dependencies**: Phases 0–3 complete

**Phase Objective**: Apply V16 lessons to permanent artifacts. NEVER skip this phase.

- Review all V16 lessons.md sections
- Update ENG-HANDBOOK.md with new lifecycle helper patterns, token efficiency patterns
- Update agents and skills where V16 work exposed improvements
- Verify `go run ./cmd/cicd-lint lint-docs` passes
- Commit all updates

**Post-Mortem**: Update lessons.md Phase 4 section.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| `@source` block drift during instruction pruning | Medium | High | Run `lint-docs` after every edit; prune only glue text |
| lifecycle package API breaks identity services (larger codebases) | Low | Medium | Run full `go build ./...` and `go test ./...` after each service migration |
| Cat 3 CN constant count (40) bloats magic package | Low | Low | Use a sub-file `magic_pki_tls.go`; excluded from coverage threshold |
| cicd-lint `--summary` flag disrupts existing pre-commit users | Low | Low | Default remains verbose; `-q` is opt-in |
| Rate limits hit during V16 execution | High | High | P0 Phase itself reduces consumption; execute P0 first to reduce risk for P1–4 |

---

## Quality Gates — Per Phase

**Before marking any phase complete**:
- `go build ./...` AND `go build -tags e2e,integration ./...` — zero errors
- `golangci-lint run ./...` AND `golangci-lint run --build-tags e2e,integration ./...` — zero warnings
- `go test ./...` — 100% passing, zero skips
- Coverage: ≥95% production, ≥98% infrastructure (lifecycle, usage packages)
- No new TODOs without tracking in tasks.md
- `go run ./cmd/cicd-lint lint-docs` — passes (Phase 2 and Phase 4 only)
- `go run ./cmd/cicd-lint lint-deployments` — passes (if deployments/ changed)

---

## Success Criteria

- [ ] Phase 0: ≥200 lines removed from instruction files; cicd-lint has `-q` mode; new `06-03.tool-efficiency.instructions.md`
- [ ] Phase 1: 10 PS-ID entry points use `lifecycle.RunService()`; 8 usage.go files use `usage.Build*()`; GAP-0.5 deleted
- [ ] Phase 2: ENG-HANDBOOK.md §§9.4, 10.4, 12.3, 6.5, 3.4, 14.9 updated; Cat 3 CN constants added; instruction files updated
- [ ] Phase 3: V15 tasks.md corrected; gaps verified; V15 lessons.md complete
- [ ] Phase 4: V16 lessons propagated; `lint-docs` passes; clean working tree
- [ ] All quality gates passing at end of each phase

---

## ENG-HANDBOOK.md Cross-References

| Topic | Section | When Applied |
|-------|---------|--------------|
| Testing Strategy | [§10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | Phases 1, 2 |
| Unit Testing / Coverage | [§10.2](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy) | Phase 1 |
| Quality Gates | [§11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL phases |
| Coding Standards | [§14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Phases 0, 1 |
| Version Control | [§14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | ALL phases |
| CI/CD Workflow Architecture | [§9.7](../../docs/ENG-HANDBOOK.md#97-cicd-workflow-architecture) | Phase 0 |
| Telemetry Strategy | [§9.4](../../docs/ENG-HANDBOOK.md#94-telemetry-strategy) | Phase 2 |
| CICD Command Architecture | [§9.10](../../docs/ENG-HANDBOOK.md#910-cicd-command-architecture) | Phase 0 |
| Deployment Architecture | [§12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) | Phase 2 |
| Service Framework Pattern | [§5.1](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern) | Phase 1 |
| Pre-Commit Hook Architecture | [§9.9](../../docs/ENG-HANDBOOK.md#99-pre-commit-hook-architecture) | Phase 0 |
| Infrastructure Blockers | [§14.7](../../docs/ENG-HANDBOOK.md#147-infrastructure-blocker-escalation) | ALL phases |
| Post-Mortem & Knowledge Propagation | [§14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | ALL phases |
| Plan Lifecycle | [§14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | ALL phases |
