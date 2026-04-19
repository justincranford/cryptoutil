# Framework v13: v10-v12 Cleanup — TLS Verification, Mutation Testing, Doc Sync

**Status**: Planning
**Created**: 2026-06-30
**Last Updated**: 2026-06-30
**Purpose**: Resolve specific unfinished work from Framework v10, v11, and v12 identified in RETROSPECTIVE-10-11-12.md.

---

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

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

---

## Overview

Framework v13 is a **cleanup plan** that resolves specific unfinished work from v10, v11, and v12. It addresses 7 of the 15 issues identified in [RETROSPECTIVE-10-11-12.md](RETROSPECTIVE-10-11-12.md). The remaining issues were handled separately: 3 deleted (items 8, 13, 15 — not actionable), 5 fixed immediately outside this plan (items 1partial, 2, 9, 10, 14), and 1 deferred to v14 (item 7 — scope creep pattern).

**Scope (retrospective items IN this plan)**:

| Item | Summary | Phase |
|------|---------|-------|
| 1 (partial) | v12 Docker-deferred TLS tasks — Docker verification | Phase 1 |
| 3 | No E2E verification for TLS/mTLS wiring | Phases 1, 2 |
| 4 | PostgreSQL mTLS wiring untested end-to-end | Phases 1, 2 |
| 5 | Admin mTLS trust verification never executed | Phases 1, 2 |
| 6 | v11 mutation/race testing permanently deferred | Phase 3 |
| 11 | Cross-version documentation drift (tls-structure.md vs ENG-HANDBOOK.md) | Phase 4 |
| 12 | v10 template directory never validated via Docker | Phase 5 |

**Out of scope** (handled separately):

| Item | Handling | Rationale |
|------|----------|-----------|
| 2 | Immediate fix (agent directive) | Added lessons.md mandate to implementation-execution agents |
| 7 | Deferred to v14 | Scope creep pattern — v14 plan already exists |
| 8 | Deleted | Over-engineered planning observation, not actionable |
| 9 | Immediate fix (apply v12 lessons) | Sparse lessons applied to permanent artifacts |
| 10 | Immediate fix (coverage ceiling directive) | Added mitigation plan directive to instructions/agents |
| 13 | Deleted | Estimation accuracy observation, not actionable |
| 14 | Immediate fix (standardize templates) | Task notation standardized in agents/instructions |
| 15 | Deleted | Estimation bias observation, not actionable |

**Definition of E2E test** (for this plan): A Go test that orchestrates `docker compose` for start/stop, and validates everything is working according to design intent while services are up, including happy path and sad path table-driven testing.

---

## Background

### Prior Work

- **v10** (Canonical Template Registry): Created ~63 template files, pki-init pflag rewrite, template-compliance linter. 33/33 tasks complete. Lessons: empty (complete knowledge loss).
- **v11** (PKI-Init Cert Structure): Implemented 14-category cert generation, PKCS#12 keystores. 25/26 tasks complete. Deferred: mutation testing, race detection (Linux-only).
- **v12** (PostgreSQL mTLS + Admin mTLS): Configured PostgreSQL server TLS, replication TLS, client mTLS, admin mTLS. 38/43 tasks complete. Deferred: 5 Docker verification tasks across Phases 3, 6, 9, 10.5.

### Consolidated Lessons

25 lessons consolidated in [RETROSPECTIVE-10-11-12.md](RETROSPECTIVE-10-11-12.md). Top 5:
1. Docker verification MUST be in-scope, never deferred
2. Capture lessons during each phase, not after
3. Every config change needs runtime verification
4. Deferred work must be explicitly assigned to a future version
5. E2E tests mandatory for CLI entry points with productionNew* functions

### Prior Analysis (Preserved)

The original framework-v13 plan.md contained deep E2E gap analysis (Flaws F1-F20, Decisions D1-D10) for a comprehensive E2E overhaul of all 16 Docker Compose deployments. That analysis remains valid but is **out of scope** for this cleanup plan. The comprehensive E2E overhaul is a separate future effort. Original analysis recoverable via: `git log --all --oneline -- docs/framework-v13/plan.md`

---

## Technical Context

- **Language**: Go 1.26.1
- **E2E Infrastructure**: `internal/apps/framework/service/testing/e2e_infra/` — `ComposeManager` struct with `Start()`, `Stop()`, `WaitForHealth()`, `WaitForMultipleServices()`
- **Existing E2E Tests**: 4 PS-IDs have compose-based E2E tests: `sm-kms`, `sm-im`, `jose-ja`, `skeleton-template`
- **Missing E2E Tests**: `pki-ca`, `identity-authz`, `identity-idp`, `identity-rs`, `identity-rp`, `identity-spa`
- **TLS Structure**: 14 cert categories generated by pki-init. See `docs/tls-structure.md`.
- **Compose Deployments**: 10 PS-ID + 5 Product + 1 Suite = 16 total (each with 4 instances: sqlite-1, sqlite-2, postgres-1, postgres-2)
- **Docker Requirement**: Docker Desktop MUST be running for Phases 1, 2, 5
- **OS**: Linux (required for Phase 3 — gremlins and race detector)

### Current E2E State

| PS-ID | Has e2e/ dir | Has testmain | Has magic constants | TLS validated |
|-------|:-:|:-:|:-:|:-:|
| skeleton-template | ✅ | ✅ | ✅ | ❌ (InsecureSkipVerify) |
| jose-ja | ✅ | ✅ | ✅ | ❌ (InsecureSkipVerify) |
| sm-im | ✅ | ✅ | ✅ | ❌ (InsecureSkipVerify) |
| sm-kms | ✅ | ✅ | ✅ | ❌ (InsecureSkipVerify) |
| identity-authz | ✅ (partial) | ❌ | ✅ (PRODUCT) | ❌ |
| identity-idp | ❌ | ❌ | ✅ (PRODUCT) | ❌ |
| identity-rs | ❌ | ❌ | ✅ (PRODUCT) | ❌ |
| identity-rp | ❌ | ❌ | ✅ (PRODUCT) | ❌ |
| identity-spa | ❌ | ❌ | ✅ (PRODUCT) | ❌ |
| pki-ca | ❌ | ❌ | ❌ | ❌ |

---

## Phase Status Legend

`☐ TODO` | `🔄 IN PROGRESS` | `✅ COMPLETE` | `⏳ BLOCKED`

## Phases

### Phase 1: v12 Docker-Deferred TLS Smoke Test (4h) [Status: ☐ TODO]

**Objective**: Verify that ALL v12 TLS/mTLS configuration actually works in running Docker Compose containers. This is the prerequisite smoke test that v12 deferred.

**Addresses**: Retrospective items 1 (partial), 3, 4, 5

**What this phase does**:
- Start Docker Desktop
- Run `docker compose up` for a PS-ID deployment with PostgreSQL (sm-kms is the reference service)
- Verify PostgreSQL server TLS is active (Cat 10 server cert mounted and used)
- Verify PostgreSQL client mTLS works (Cat 12 client cert presented by app)
- Verify PostgreSQL replication TLS between leader/follower
- Verify admin mTLS endpoint responds (Cat 3 admin cert)
- Verify public TLS endpoint responds (Cat 2 public cert)
- Fix any configuration issues found (cert paths, HBA rules, volume mounts, GORM DSN)

**Success Criteria**:
- `docker compose up --wait` succeeds for sm-kms deployment
- All 4 app instances (sqlite-1, sqlite-2, postgres-1, postgres-2) reach healthy status
- `psql` connects to PostgreSQL leader with `sslmode=verify-full`
- PostgreSQL replication uses TLS (`pg_stat_ssl` shows SSL=true for replication)
- Admin endpoint `/admin/api/v1/livez` responds over mTLS
- Public endpoint `/service/api/v1/health` responds over TLS

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 2: TLS/mTLS E2E Go Tests (6h) [Status: ☐ TODO]

**Objective**: Write Go E2E tests that programmatically validate TLS certificate chains and mTLS authentication in running Docker Compose deployments. Replace `InsecureSkipVerify: true` with real CA trust validation.

**Addresses**: Retrospective items 3, 4, 5

**What this phase does**:
- Add CA-validated TLS client to `e2e_infra` (use `NewClientForTestWithCA()` or equivalent)
- Update existing 4 PS-ID E2E tests to validate TLS chain instead of InsecureSkipVerify
- Add table-driven tests for public TLS chain validation (happy path: correct CA, sad path: wrong CA, expired cert)
- Add table-driven tests for admin mTLS validation (happy path: correct client cert, sad path: no client cert, wrong client cert)
- Add PostgreSQL mTLS connection test (verify app connects to PostgreSQL with client cert)
- All tests orchestrate `docker compose` for start/stop via `ComposeManager`

**Success Criteria**:
- All 4 existing PS-ID E2E tests pass with real CA trust (no InsecureSkipVerify)
- New TLS chain validation tests pass (happy + sad paths)
- New admin mTLS tests pass (happy + sad paths)
- PostgreSQL mTLS connection verified programmatically
- Tests are table-driven with t.Parallel() where applicable

**Post-Mortem**: After quality gates pass, update lessons.md.

---

### Phase 3: v11 Mutation & Race Testing (2h) [Status: ☐ TODO]

**Objective**: Execute the mutation testing and race detection that v11 deferred to Linux.

**Addresses**: Retrospective item 6

**What this phase does**:
- Run `gremlins unleash` on pki-init packages (`internal/apps/tools/pki_init/`)
- Run `go test -race -count=2` on pki-init packages
- Fix any mutation survivors or race conditions found
- Document mutation testing efficacy score

**Success Criteria**:
- gremlins mutation score ≥95% for pki-init packages
- Race detector clean (zero data races)
- Any survivors analyzed and documented (true survivors vs equivalent mutations)

**Post-Mortem**: After quality gates pass, update lessons.md.

---

### Phase 4: Documentation Synchronization (2h) [Status: ☐ TODO]

**Objective**: Resolve documentation drift between `tls-structure.md`, `tls-structure-suggestions.md`, and `ENG-HANDBOOK.md`.

**Addresses**: Retrospective item 11

**What this phase does**:
- Review `docs/tls-structure-suggestions.md` for pending ENG-HANDBOOK.md updates
- Merge applicable suggestions into ENG-HANDBOOK.md §6.11.3
- Verify `docs/tls-structure.md` reflects post-v12 cert structure (Cat 9, Cat 14 changes)
- Run `go run ./cmd/cicd-lint lint-docs` to verify propagation integrity
- Delete `docs/tls-structure-suggestions.md` after merging (git history preserves content)

**Success Criteria**:
- ENG-HANDBOOK.md §6.11.3 reflects the current 14-category cert structure
- `tls-structure.md` and ENG-HANDBOOK.md are consistent
- `lint-docs` passes with zero errors
- No documentation references stale v11 cert structure

**Post-Mortem**: After quality gates pass, update lessons.md.

---

### Phase 5: Template Docker Validation (2h) [Status: ☐ TODO]

**Objective**: Verify that v10's template directory produces functionally correct Docker Compose deployments (not just structurally correct).

**Addresses**: Retrospective item 12

**What this phase does**:
- Run `docker compose up --wait` for skeleton-template deployment (the reference service)
- Verify all 4 instances reach healthy status
- Run template-compliance linter (`go run ./cmd/cicd-lint lint-fitness`) to confirm structural match
- Compare template output against actual deployment files for skeleton-template
- Document any discrepancies between structural compliance and functional correctness

**Success Criteria**:
- skeleton-template Docker Compose deployment starts successfully
- All 4 instances (sqlite-1, sqlite-2, postgres-1, postgres-2) healthy
- template-compliance linter passes
- No functional issues discovered that the structural linter missed

**Post-Mortem**: After quality gates pass, update lessons.md.

---

### Phase 6: Knowledge Propagation (1h) [Status: ☐ TODO]

**Objective**: Apply lessons learned from Phases 1-5 to permanent artifacts.

**What this phase does**:
- Review lessons.md from all prior phases
- Update ENG-HANDBOOK.md with new patterns and decisions discovered
- Update agents, skills, instructions as warranted
- Update code, tests, workflows where plan work exposed gaps
- Verify propagation integrity (`go run ./cmd/cicd-lint lint-docs validate-propagation`)

**Success Criteria**:
- All artifact updates committed with separate semantic commits per artifact type
- Propagation check passes
- No lessons remain unextracted

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Docker Desktop not available | Low | High (blocks Phases 1, 2, 5) | Phase 3, 4 can proceed without Docker; verify Docker before starting |
| v12 TLS wiring has configuration bugs | Medium | Medium | Phase 1 smoke test catches these before writing E2E Go tests |
| gremlins not installed on Linux | Low | Low | Install via `go install` or download binary |
| PostgreSQL replication TLS hard to verify | Medium | Medium | Use `psql` with `sslinfo` extension; check `pg_stat_ssl` |
| tls-structure-suggestions.md references stale code | Low | Low | Cross-reference with actual pki-init generator code |

---

## Quality Gates - MANDATORY

**Per-Phase Quality Gates**:
- ✅ All tests pass (`go test ./...`) — 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) — zero errors
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) — zero warnings
- ✅ No new TODOs without tracking in tasks.md
- ✅ lessons.md updated with phase post-mortem (definition of done for every phase)

**Coverage Targets**:
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code: ≥98% line coverage
- ✅ Generated code: Excluded from coverage

**Docker-Dependent Phases (1, 2, 5)**:
- ✅ Docker Desktop running
- ✅ `docker compose up --wait` succeeds
- ✅ All containers reach healthy status
- ✅ Health endpoints respond correctly

**Overall**:
- ✅ All 6 phases complete with evidence
- ✅ All retrospective items addressed (1partial, 3, 4, 5, 6, 11, 12)
- ✅ CI/CD clean (build, lint, test)

---

## Success Criteria

- [ ] Phase 1: v12 TLS wiring verified in Docker — all endpoints respond correctly
- [ ] Phase 2: E2E Go tests validate TLS chains — no InsecureSkipVerify in E2E tests
- [ ] Phase 3: pki-init mutation score ≥95%, race detector clean
- [ ] Phase 4: ENG-HANDBOOK.md §6.11.3 current, lint-docs passes
- [ ] Phase 5: skeleton-template Docker Compose functional, template-compliance passes
- [ ] Phase 6: Lessons extracted to permanent artifacts, propagation passes
- [ ] All quality gates passing
- [ ] Evidence archived in test-output/

---

## ENG-HANDBOOK.md Cross-References - MANDATORY

| Topic | ENG-HANDBOOK.md Section | Phases |
|-------|------------------------|--------|
| Testing Strategy | [Section 10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | ALL |
| E2E Testing | [Section 10.4](../../docs/ENG-HANDBOOK.md#104-e2e-testing-strategy) | 1, 2, 5 |
| Mutation Testing | [Section 10.5](../../docs/ENG-HANDBOOK.md#105-mutation-testing-strategy) | 3 |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL |
| Security Architecture | [Section 6](../../docs/ENG-HANDBOOK.md#6-security-architecture) | 1, 2 |
| PKI Architecture | [Section 6.5](../../docs/ENG-HANDBOOK.md#65-pki-architecture--strategy) | 1, 2 |
| Deployment Architecture | [Section 12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) | 1, 2, 5 |
| Plan Lifecycle | [Section 14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | ALL |
| Post-Mortem & Knowledge Propagation | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | ALL |
| Coding Standards | [Section 14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | 2 |
| Version Control | [Section 14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | ALL |

---

## Evidence Archive

- `test-output/phase0-research/` — Phase 0 research findings (from plan creation)
- `test-output/v13-phase1/` — Docker TLS smoke test logs
- `test-output/v13-phase2/` — E2E TLS test results
- `test-output/v13-phase3/` — Mutation and race testing results
- `test-output/v13-phase4/` — Documentation sync verification
- `test-output/v13-phase5/` — Template Docker validation logs
