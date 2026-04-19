# Implementation Plan — Framework v14: v13 Completion

**Status**: Planning
**Created**: 2026-04-19
**Last Updated**: 2026-04-19
**Purpose**: Complete all deferred, missed, inefficient, and partially-done work from framework-v13
before advancing to framework-v15 (OTel/Grafana mTLS + Public PS-ID App TLS Trust). This plan
carries forward every actionable item from v13's "Patterns for Future Phases" lessons and its
explicit deferred items list.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers — NO exceptions:**

- ✅ **Fix issues immediately** — When unknowns discovered, blockers identified, tests fail, or quality
  gates not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next phase or task
- ✅ **Document root causes** — Root cause analysis is part of implementation, not optional
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task complete with known issues
- ✅ **NEVER de-prioritize quality** — Evidence-based verification is ALWAYS highest priority

---

## Overview

Framework v13 ("v10-v12 Cleanup") completed all 6 phases and 30 tasks. However, several items were
explicitly deferred, acknowledged as incomplete, or identified as "patterns for future phases" that
require concrete implementation. This plan addresses all of them before the suite can advance to
v15's OTel/Grafana mTLS wiring.

**Deferred items being resolved in this plan**:

| # | Source | Item | Phase |
|---|--------|------|-------|
| 1 | v13 tasks.md Item 7 | Full E2E framework redesign: registry-driven config, shared TestMain factory | Phase 4 |
| 2 | v13 Phase 2 lessons | Admin mTLS full round-trip test via docker exec (port isolation only was done) | Phase 2 |
| 3 | v13 Retrospective #10 | pki-init coverage ceiling 92.4% — no mitigation plan existed | Phase 3 |
| 4 | v13 Phase 3 lessons | gremlins run only on framework/tls — not on e2e_infra additions | Phase 5 |
| 5 | v13 cross-cutting | Cross-cutting quality tasks never individually verified/closed | Phase 1 |

---

## Background

### What v13 Completed

v13 ("v10-v12 Cleanup") resolved all 15 retrospective items from v10, v11, v12:

- Phase 1: Verified all v12 TLS/mTLS wiring in Docker (found 3 bugs: `setup-logical-replication.sh`
  used TCP, OTel healthcheck used wget in distroless, pki-init didn't generate `tls-config.yml`)
- Phase 2: Added CA-validated TLS E2E tests for all 4 PS-IDs (sm-kms, jose-ja, sm-im,
  skeleton-template); added `BuildDockerExecArgs` to `e2e_infra/compose_manager.go`
- Phase 3: Ran gremlins mutation testing (100% efficacy, 92% mutator coverage) and race detection
  on `internal/apps/framework/tls/` — zero races, zero survived mutations
- Phase 4: Merged tls-structure-suggestions to ENG-HANDBOOK.md; confirmed tls-structure.md
  consistency; deleted ephemeral suggestions file
- Phase 5: Verified skeleton-template Docker Compose (all 4 instances healthy); fixed 2 canonical
  template drift issues; fixed `e2e_admin_isolation_test.go` `t.Errorf` → `require.Fail`
- Phase 6: Propagated 4 new Docker Compose rules to ENG-HANDBOOK.md and deployment instructions

### What v13 Did NOT Complete (Carried into v14)

**Explicit deferrals** (from v13 tasks.md "Notes / Deferred Work"):
- Item 7: Full E2E framework redesign — registry-driven config, shared TestMain factory,
  16-deployment orchestrator. Scope was too large for v13.

**Partial completions** (from v13 lessons "Patterns for Future Phases"):
- Admin mTLS verification: v13 Phase 2 only tested that port 9090 is NOT exposed to the host
  (isolation test). The "full round-trip mTLS test requires either docker exec or a Go E2E test
  that runs inside the network" — this was called out explicitly but deferred.
- gremlins coverage: Only run on `internal/apps/framework/tls/`. The new `e2e_infra` code
  (`compose_manager.go`, `BuildDockerExecArgs`) was never mutation-tested.

**Acknowledged quality debt** (from v13 RETROSPECTIVE-10-11-12.md):
- Issue #10: v11 coverage ceiling 92.4% accepted without mitigation plan. "Apply the `internalMain`
  pattern to pki-init CLI entry points to raise the ceiling, or add pki-init E2E tests to CI/CD
  pipeline." — never actioned.

**Unclosed cross-cutting tasks** (from v13 tasks.md):
- The cross-cutting quality task checklist (Testing, Code Quality, Documentation, Deployment
  sections) was never individually verified and closed — all items remain unchecked despite the
  phase work completing them.

---

## Technical Context

- **Language**: Go 1.26.1, CGO_ENABLED=0
- **Framework**: `internal/apps/framework/service/`
- **pki-init**: `internal/apps/framework/tls/generator.go` + `generator_helpers.go`
- **E2E infra**: `internal/apps/framework/service/testing/e2e_infra/`
- **E2E test dirs**: `internal/apps/{sm-kms,jose-ja,sm-im,skeleton-template}/e2e/`
- **Registry**: `api/cryptosuite-registry/registry.yaml`

---

## Phases

**Phase Status Legend**: `☐ TODO` | `🔄 IN PROGRESS` | `✅ COMPLETE` | `⏳ BLOCKED`

### Phase 1: Close v13 Cross-Cutting Quality Gates (1h) [Status: ☐ TODO]

**Objective**: Explicitly verify and close every unchecked cross-cutting task from v13's tasks.md.
These were done as side effects of phases but never individually confirmed.

- Run `go build ./...` + `go build -tags e2e,integration ./...` → must be clean
- Run `golangci-lint run` + `golangci-lint run --build-tags e2e,integration` → zero violations
- Run `go test ./... -shuffle=on -count=1` → 100% pass, zero skips
- Run `go run ./cmd/cicd-lint lint-fitness lint-docs` → zero errors
- Confirm: `grep -r "InsecureSkipVerify.*true" internal/apps/*/e2e/` → zero results in test files
  that use CA-validated client (sm-kms/e2e/e2e_tls_test.go confirms chain; other files may still
  use insecure client for health check ping — audit and document intentional usages)
- Update v13 tasks.md: mark all cross-cutting items ✅ with evidence
- **Success**: All checks pass; v13 tasks.md cross-cutting section fully closed
- **Post-Mortem**: After quality gates pass, update lessons.md with lessons learned — what worked,
  what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix
  tasks immediately.

### Phase 2: Admin mTLS Full Round-Trip Test (3h) [Status: ☐ TODO]

**Objective**: Write a Go E2E test that verifies the admin endpoint's mTLS requirement by
connecting FROM INSIDE the Docker container network — filling the gap from v13 Phase 2.

**Context**: v13 Phase 2 established the admin port isolation test (port 9090 is NOT exposed to
host → `net.DialTimeout` fails from test host). This is necessary but not sufficient. The full
correctness test is: when connecting WITH the correct client cert, the admin endpoint accepts the
connection; when connecting WITHOUT a cert, it returns `tls: certificate required`. This requires
`docker exec` to run from inside the container.

- Add `e2e_admin_mtls_test.go` to `internal/apps/sm-kms/e2e/`:
  - Use `composeManager.BuildDockerExecArgs` (added in v13) to run `curl --cert ... --key ... --cacert ...` inside the container
  - Happy path: connect with correct admin client cert → HTTP 200 from `/admin/api/v1/livez`
  - Sad path: connect without client cert → `tls: certificate required` (TLS handshake error)
- Run the new test against the sm-kms Docker Compose stack to verify it passes
- Confirm `golangci-lint run --build-tags e2e` is clean on new file
- **Success**: Both happy and sad paths pass in CI
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: pki-init Coverage Ceiling Mitigation (4h) [Status: ☐ TODO]

**Objective**: Apply the `internalMain` pattern to pki-init CLI entry points to raise test coverage
from 92.4% (accepted ceiling in v11) to ≥95% (mandatory production target). Close Retrospective
Issue #10 definitively.

**Context**: v11 accepted 92.4% coverage citing a "coverage ceiling" of ~93% due to
`productionNew*` functions that are only exercisable via E2E. The retrospective's recommended fix
was to apply `internalMain` pattern — but v13 never actioned this.

- Identify the production wiring functions in `internal/apps/framework/tls/` that block coverage:
  `productionNewLogger`, `productionNewTelemetry`, `productionNewGenerator`, CLI `main()` body
- Refactor CLI entry point `cmd/pki-ca/main.go` (or equivalent pki-init entry) to use the
  `internalMain(args, stdin, stdout, stderr)` pattern per ENG-HANDBOOK.md §10.2.3
- Add unit tests for the `internalMain` function with injected I/O and fake args
- Re-run coverage: `go test -coverprofile=coverage.out ./internal/apps/framework/tls/...`
  → coverage.out must show ≥95%
- Re-run gremlins to confirm mutation efficacy unchanged or improved
- **Success**: `go test -cover ./...` shows ≥95% for pki-init packages; gremlins efficacy ≥95%
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 4: E2E Framework Redesign — Shared TestMain Factory (8h) [Status: ☐ TODO]

**Objective**: Eliminate the copy-paste TestMain pattern across 4 PS-ID e2e directories.
Implement a registry-driven, shared TestMain factory. This is v13 deferred Item 7.

**Context**: All 4 PS-ID e2e test suites (`sm-kms`, `jose-ja`, `sm-im`, `skeleton-template`) share
identical or near-identical `testmain_e2e_test.go` boilerplate:
- `composeManager` setup from PS-ID-specific magic constant
- `WaitForMultipleServices` call with hardcoded health URLs
- `sharedHTTPClient` (InsecureSkipVerify) initialization
- `sharedHTTPClientWithCA` (CA-validated) initialization from pki-init CA cert path

The redesign produces a single factory in `e2e_infra/` that any PS-ID can call, eliminating
~200 lines of duplicated boilerplate per service.

**Scope** (registry-driven):
- `e2e_infra/testmain_factory.go`: `NewPS-IDTestMain(psID string) *E2ETestMain`
  - Reads PS-ID config from `api/cryptosuite-registry/registry.yaml` for ports, service names
  - Builds `ComposeManager` from registry-derived compose file path
  - Initializes both HTTP clients (insecure + CA-validated)
  - Provides `Setup()`, `Teardown()`, `PublicClient()`, `SecureClient()` methods
- Update all 4 PS-ID `testmain_e2e_test.go` files to use the factory (replace boilerplate)
- Unit tests for the factory in `e2e_infra/testmain_factory_test.go` (≥95% coverage)
- `golangci-lint run --build-tags e2e,integration` clean
- **Success**: All 4 PS-ID E2E suites pass using the shared factory; factory unit tests pass
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 5: Mutation Testing on New e2e_infra Code (2h) [Status: ☐ TODO]

**Objective**: Run gremlins on all `e2e_infra` code added in v13 and v14 (Phases 2+4).
v13 Phase 3 only mutation-tested `internal/apps/framework/tls/` — the new `compose_manager.go`
additions (`BuildDockerExecArgs`) and the new TestMain factory were never mutation-tested.

- Run `gremlins unleash --tags=!integration ./internal/apps/framework/service/testing/e2e_infra`
- Target: ≥95% efficacy, ≥90% mutator coverage
- Fix any surviving mutations (add tests or document why the mutation is impossible to kill)
- Re-run race detector: `CGO_ENABLED=1 go test -race -count=2 ./internal/apps/framework/service/testing/e2e_infra/...`
- **Success**: Mutation efficacy ≥95%; zero races detected
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 6: Knowledge Propagation (2h) [Status: ☐ TODO]

**Objective**: Apply lessons learned from Phases 1-5 to permanent artifacts — NEVER skip this phase.

- Review lessons.md from all prior phases
- Update ENG-HANDBOOK.md with patterns from v14:
  - Admin mTLS testing via docker exec (§10.4 E2E Testing Strategy)
  - `internalMain` pattern applicability to pki-init CLI entry points (§10.2.3 Coverage Targets)
  - Shared TestMain factory pattern for E2E suites (§10.3.6 Shared Test Infrastructure)
- Update agents/skills/instructions where v14 work exposed gaps
- Verify propagation: `go run ./cmd/cicd-lint lint-docs` passes
- **Success**: All artifact updates committed; propagation check passes
- **Post-Mortem**: After quality gates pass, update lessons.md — v14 complete.

---

## Decisions

### D1: Admin mTLS Test Approach

**Options**:
- A: docker exec curl with client cert flags inside the container
- B: Go test that directly calls `docker exec` via `exec.Command`
- C: Add a dedicated test helper container that runs the mTLS connection attempt
- D: Accept port isolation test as sufficient; skip full round-trip
- E:

**Decision**: Option B selected — Go test using `exec.Command("docker", ...)` via `composeManager.BuildDockerExecArgs`

**Rationale**: Option B uses the `BuildDockerExecArgs` infrastructure added in v13 Phase 2
(commit `8dae215dd`). The helper was added specifically in anticipation of this use case. Option A
requires manual shell steps. Option C adds unnecessary complexity. Option D was v13's choice and
explicitly called out in lessons as insufficient.

### D2: pki-init `internalMain` Scope

**Options**:
- A: Apply `internalMain` only to the CLI entry point `cmd/pki-ca/main.go` or equivalent
- B: Apply `internalMain` + inject `productionNewLogger`, `productionNewGenerator`, etc. as fn args
- C: Add E2E pki-init smoke test to CI/CD instead of `internalMain` refactor
- D: Accept 92.4% ceiling permanently — document exception
- E:

**Decision**: Option B selected — full function-parameter injection per ENG-HANDBOOK.md §10.2.4

**Rationale**: Option B is the canonical approach per ENG-HANDBOOK.md §10.2.4 "Test Seam Injection
Pattern" and is the mitigation specified in Retrospective Issue #10. Option A partially helps but
leaves `productionNew*` functions untested. Option C is valid but orthogonal — doesn't close the
unit coverage ceiling. Option D was explicitly rejected in the retrospective.

### D3: TestMain Factory Registry Integration

**Options**:
- A: Parse `api/cryptosuite-registry/registry.yaml` at test runtime for port numbers
- B: Use magic constants from `internal/shared/magic/` — no runtime YAML parsing in tests
- C: Registry-driven factory generates magic constants at code generation time (pre-compile)
- D: Hardcode port/path tables in the factory (simpler, less maintenance)
- E:

**Decision**: Option B selected — magic constants as the runtime interface; registry.yaml drives codegen or documentation only

**Rationale**: Magic constants are already the established pattern in this project. Parsing YAML at
test runtime adds dependency on file availability during test execution and adds parser code to the
test path. Option C (codegen) is the ideal long-term solution but is out of scope for v14. Option D
(hardcode in factory) is rejected — the factory would become a maintenance burden as new PS-IDs are
added. Option B: the factory accepts PS-ID magic constants as parameters, keeping the factory
generic while using existing infrastructure.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Admin mTLS docker exec test is flaky (timing) | Medium | Low | Use `composeManager.WaitForMultipleServices` guarantee before test; add retry logic if needed |
| pki-init CLI entry point structure doesn't fit `internalMain` cleanly | Low | Medium | Audit `cmd/` structure first (Phase 3.1); adapt pattern to actual structure before writing tests |
| TestMain factory introduces subtle race conditions across PS-IDs | Low | Medium | Run `go test -race` on all 4 E2E suites after Phase 4 |
| gremlins timeout on e2e_infra code (network/exec paths) | Medium | Low | Use `--tags=!integration` to skip Docker-dependent paths; document NOT_COVERED with justification |

---

## Quality Gates - MANDATORY

**Per-Task**:
- ✅ All tests pass — 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`)
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`)
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ pki-init packages (`internal/apps/framework/tls/`): ≥95% (raised from v11/v13 ceiling)
- ✅ e2e_infra package: ≥95% (production utility code)
- ✅ New TestMain factory: ≥95%
- ✅ Generated code: excluded

**Per-Phase**:
- ✅ Verification step passes before next phase
- ✅ Race detector clean: `CGO_ENABLED=1 go test -race -count=2 ./...`
- ✅ lint-fitness passes after each E2E code change

---

## Success Criteria

- [ ] Phase 1: All v13 cross-cutting tasks explicitly closed with evidence
- [ ] Phase 2: Admin mTLS full round-trip test passes (happy + sad paths)
- [ ] Phase 3: pki-init coverage ≥95%; gremlins efficacy ≥95%
- [ ] Phase 4: Shared TestMain factory in use by all 4 PS-ID E2E suites
- [ ] Phase 5: e2e_infra mutation efficacy ≥95%; zero races
- [ ] Phase 6: ENG-HANDBOOK.md updated; lint-docs passes
- [ ] All phases complete; framework-v15 (OTel/Grafana mTLS) can begin

---

## ENG-HANDBOOK.md Cross-References — MANDATORY

| Topic | Section | When to Reference |
|-------|---------|-------------------|
| Testing Strategy | [§10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | ALL phases |
| Unit Testing | [§10.2](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy) | Phases 1, 3 |
| Coverage Targets + Ceiling | [§10.2.3](../../docs/ENG-HANDBOOK.md#1023-coverage-targets) | Phase 3 |
| Test Seam Injection | [§10.2.4](../../docs/ENG-HANDBOOK.md#1024-test-seam-injection-pattern) | Phase 3 |
| Integration Testing | [§10.3](../../docs/ENG-HANDBOOK.md#103-integration-testing-strategy) | Phase 4 |
| Shared Test Infrastructure | [§10.3.6](../../docs/ENG-HANDBOOK.md#1036-shared-test-infrastructure) | Phase 4 |
| E2E Testing | [§10.4](../../docs/ENG-HANDBOOK.md#104-e2e-testing-strategy) | Phases 2, 4 |
| Mutation Testing | [§10.5](../../docs/ENG-HANDBOOK.md#105-mutation-testing-strategy) | Phase 5 |
| Quality Gates | [§11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL phases |
| Coding Standards | [§14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Phases 3, 4 |
| Version Control | [§14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | ALL phases |
| Infrastructure Blockers | [§14.7](../../docs/ENG-HANDBOOK.md#147-infrastructure-blocker-escalation) | Phase 2 |
| Post-Mortem & Propagation | [§14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | Every phase + Phase 6 |
| Plan Lifecycle | [§14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | ALL phases |
