# Retrospective — Framework v10, v11, v12

**Created**: 2026-06-29
**Purpose**: Identify work that was not completed correctly, or has issues, inefficiencies, mistakes, omissions, and all other problems across Framework v10, v11, and v12 implementation plans.

---

## Executive Summary

Across three framework plans (v10: Canonical Template Registry, v11: PKI-Init Cert Structure, v12: PostgreSQL mTLS + Private App mTLS Trust), 96 of 102 tasks were completed (94%). However, completion percentage obscures significant structural problems: v10's lessons were never captured (empty placeholder), v12 has 5 Docker-dependent tasks that remain indefinitely deferred with no concrete plan to complete them, and several cross-cutting issues compound across versions. The issues below are prioritized from highest to lowest impact.

1. [v12 Docker-deferred tasks permanently unresolved](#1-v12-docker-deferred-tasks-permanently-unresolved)
2. [v10 lessons.md never filled — complete knowledge loss](#2-v10-lessonsmd-never-filled--complete-knowledge-loss)
3. [No E2E verification for any TLS/mTLS wiring (v12)](#3-no-e2e-verification-for-any-tlsmtls-wiring-v12)
4. [PostgreSQL mTLS wiring untested end-to-end (v12)](#4-postgresql-mtls-wiring-untested-end-to-end-v12)
5. [Admin mTLS trust verification never executed (v12 Phase 10.5)](#5-admin-mtls-trust-verification-never-executed-v12-phase-105)
6. [v11 mutation/race testing permanently deferred](#6-v11-mutationrace-testing-permanently-deferred)
7. [Scope creep pattern — deferral cascades across versions](#7-scope-creep-pattern--deferral-cascades-across-versions)
8. [v10 over-engineered planning (4 quizme rounds, 18 decisions)](#8-v10-over-engineered-planning-4-quizme-rounds-18-decisions)
9. [v12 lessons.md is sparse — minimal root cause analysis](#9-v12-lessonsmd-is-sparse--minimal-root-cause-analysis)
10. [v11 coverage ceiling accepted without mitigation plan](#10-v11-coverage-ceiling-accepted-without-mitigation-plan)
11. [Cross-version documentation drift — tls-structure.md vs ENG-HANDBOOK.md](#11-cross-version-documentation-drift--tls-structuremd-vs-eng-handbookmd)
12. [v10 template directory never validated against live deployments via Docker](#12-v10-template-directory-never-validated-against-live-deployments-via-docker)
13. [v12 Phase estimation accuracy was poor](#13-v12-phase-estimation-accuracy-was-poor)
14. [Inconsistent task status notation across versions](#14-inconsistent-task-status-notation-across-versions)
15. [v11 Phase 6 over-estimated — systemic estimation bias toward overcount](#15-v11-phase-6-over-estimated--systemic-estimation-bias-toward-overcount)

---

## Details

### 1. v12 Docker-deferred tasks permanently unresolved

**Severity**: HIGH — Blocks confidence in production readiness
**Version**: v12
**Tasks affected**: Phase 3 (Verify PostgreSQL Standalone), Phase 6 (Verify PostgreSQL Full Stack), Phase 9 (Deployment Verification — PostgreSQL TLS), Phase 10.5 (Admin mTLS Docker verification)

v12 completed 38 of 43 tasks (88%). The 5 remaining tasks all require Docker Compose to run and were marked "⏳ DEFERRED (requires Docker)." No concrete plan exists to resolve these deferrals — they were not assigned to a future framework version, not tracked in a backlog, and not given a deadline. This means the PostgreSQL TLS configuration was written but never verified to actually work in a running deployment.

**Impact**: All PostgreSQL mTLS wiring (Phases 1-2, 4-5) is configuration-only with zero runtime verification. If any cert path, HBA rule, or GORM SSL parameter is wrong, the error will only surface in production or during manual Docker testing.

**Recommended fix**: Framework v13 MUST include Docker-based verification of all v12 TLS wiring as a prerequisite Phase 0 task, before writing any new E2E tests.

---

### 2. v10 lessons.md never filled — complete knowledge loss

**Severity**: HIGH — Prevents learning from v10 implementation
**Version**: v10

v10's `lessons.md` contains only empty phase placeholders (`*(To be filled during Phase N execution)*`). Despite completing 33/33 tasks, zero lessons were captured. This is the only framework version with a completely empty lessons file.

v10 involved significant work: ~63 template files created, pki-init pflag rewrite, framework/domain config split, 4 rounds of quizme design questions (18 decisions). The implementation certainly encountered patterns, bugs, and tradeoffs worth documenting.

**Impact**: Future implementers cannot learn from v10's experience. The 18 decisions are documented in plan.md but the implementation experience (what worked, what broke, what was harder than expected) is permanently lost.

**Recommended fix**: Cannot be retroactively filled (context is gone). For v13, mandate that lessons.md is updated at the end of each phase (not deferred to "Phase N: Knowledge Propagation").

---

### 3. No E2E verification for any TLS/mTLS wiring (v12)

**Severity**: HIGH — Zero confidence that TLS works in containers
**Version**: v12

v12 implemented PostgreSQL server TLS (Phase 1), replication TLS (Phase 2), client mTLS (Phase 4), replication mTLS (Phase 5), and admin app mTLS (Phase 10). All of this was done as configuration file changes (postgresql.conf, pg_hba.conf, compose volumes, GORM DSN parameters). None of it was verified in a running Docker Compose environment.

The verification phases (3, 6, 9) were designed to run `docker compose up`, execute `psql` commands, check `pg_stat_ssl`, and verify replication. These phases were entirely deferred.

**Impact**: The TLS wiring may be correct on paper but broken in practice. Common failure modes: wrong cert file paths after volume mounting, permission errors on cert files inside containers, HBA rule ordering issues, GORM parameter mismatches with actual PostgreSQL SSL expectations.

**Recommended fix**: v13 Phase 0 MUST include a Docker-based TLS smoke test that validates all v12 TLS wiring before proceeding to E2E test implementation.

---

### 4. PostgreSQL mTLS wiring untested end-to-end (v12)

**Severity**: HIGH — Related to #3 but specifically about the mTLS chain
**Version**: v12

The full mTLS chain requires: (1) PostgreSQL server presents a cert signed by Cat 10 Issuing CA, (2) client presents a cert signed by Cat 12 Issuing CA, (3) both sides verify the peer cert against their respective CA truststore. This chain was configured across 6 different files (leader conf, follower conf, leader HBA, follower HBA, app GORM DSN, template compose volumes) but never tested as an integrated system.

**Impact**: Any single misconfiguration in the 6-file chain breaks the entire mTLS handshake. The failure mode is a connection refused or TLS handshake error at runtime, not a build-time error.

**Recommended fix**: Include an mTLS connection verification test in v13's Docker-based verification phase.

---

### 5. Admin mTLS trust verification never executed (v12 Phase 10.5)

**Severity**: MEDIUM — Admin endpoint is internal-only but still critical
**Version**: v12

Phase 10 implemented `applyAdminMTLS()` in the framework's admin server configuration (server-admin-tls-cert-file, server-admin-tls-key-file, server-admin-tls-ca-file). Phase 10.5 was supposed to verify this in Docker but was deferred. The unit tests use seam-injected `osReadFileFn` stubs, so the actual file-reading paths are untested.

**Impact**: Admin health checks (`/admin/api/v1/livez`, `/admin/api/v1/readyz`) may fail if the mTLS configuration is incorrect, causing Docker health checks to fail and containers to restart in a loop.

**Recommended fix**: v13 E2E tests MUST verify admin endpoint accessibility with the correct client cert.

---

### 6. v11 mutation/race testing permanently deferred

**Severity**: MEDIUM — Quality gate not fully enforced
**Version**: v11

Tasks 5.3 (mutation testing) and 5.4 (race detection) were deferred to "Linux CI/CD" because gremlins v0.6.0 panics on Windows and `go test -race` requires CGO_ENABLED=1. No CI/CD workflow was created to run these checks, so they remain unexecuted indefinitely.

**Impact**: The pki-init generator code (14 categories, PKCS#12, file I/O) has 92.4% coverage but unknown mutation testing efficacy. Race conditions in concurrent cert generation are unverified.

**Recommended fix**: Add gremlins and race detection to Linux CI/CD pipeline, or run manually on a Linux machine for v11 packages.

---

### 7. Scope creep pattern — deferral cascades across versions

**Severity**: MEDIUM — Systemic process issue
**Version**: v11 → v12 → v13

v11 deferred PostgreSQL TLS, OTel TLS, and Grafana TLS to v12 (Decisions 2-4). v12 then deferred OTel TLS, Grafana TLS, and Public App TLS to v13 (plan.md overview). v12 also deferred its own Docker verification phases. Each version completes its scope by deferring boundary work to the next version, but the deferred work accumulates and is never fully resolved.

**Pattern**: v11 deferred 4 TLS concerns → v12 picked up 2 (PostgreSQL), deferred 3 (OTel, Grafana, Public App) + created 5 new deferrals (Docker verification) → v13 now inherits 8 deferred items before starting its own work.

**Impact**: Each framework version starts with an unpaid debt from prior versions. The deferred Docker verification from v12 is particularly dangerous because it means the foundation for v13 (PostgreSQL TLS) has never been validated.

**Recommended fix**: v13 plan.md (already created) addresses this by including Docker verification in Phase 0. Future plans MUST include a "deferred debt" section that explicitly lists inherited work.

---

### 8. v10 over-engineered planning (4 quizme rounds, 18 decisions)

**Severity**: LOW — Time spent on planning that could have been implementation
**Version**: v10

v10 had 4 rounds of quizme questions (v1: 10 questions, v2: 7 questions, v3: 8 questions, v4: 7 questions) resulting in 18 formal decisions before any code was written. While design questions are valuable, 32 total questions for a template registry seems disproportionate. Some decisions (e.g., Decision 18: "Docker Compose profiles BANNED") could have been established as a project-wide rule rather than a per-plan decision.

**Impact**: Planning overhead. The 4 quizme rounds represent approximately 8-12 hours of design discussion before implementation began. Some decisions were refinements of earlier decisions (e.g., Decision 4 was updated 3 times across quizme v1, v3, v4).

**Recommended fix**: Limit quizme rounds to 2 (initial + clarification). Decisions that are project-wide policies should be added to ENG-HANDBOOK.md, not embedded in plan-specific decision lists.

---

### 9. v12 lessons.md is sparse — minimal root cause analysis

**Severity**: LOW — Knowledge partially captured but shallow
**Version**: v12

v12's `lessons.md` exists and covers all phases, but each phase entry is 2-4 bullet points with no "What Worked" / "What Didn't Work" / "Root Causes" structure (compare v11's detailed 3-section format per phase). For example, Phase 4's entry mentions `stripQueryParam` and `allowedInstanceKeys` but does not explain why these were needed or what bug they fixed.

**Impact**: Future implementers get hints but not actionable patterns from v12's experience.

**Recommended fix**: v13 lessons.md should follow v11's per-phase structure: What Worked, What Didn't Work, Root Causes, Patterns for Future Phases.

---

### 10. v11 coverage ceiling accepted without mitigation plan

**Severity**: LOW — Accepted deviation from quality gate
**Version**: v11

v11 achieved 92.4% coverage on pki-init against a 95% target, citing a "coverage ceiling" of ~93% due to `productionNew*` functions that are only exercisable via E2E. The ceiling analysis is well-documented but no mitigation was planned — no E2E CI/CD integration, no `internalMain` pattern to wrap the production init code.

**Impact**: The 92.4% coverage is accepted as permanent. The `productionNew*` functions will remain untested unless someone explicitly adds E2E coverage.

**Recommended fix**: Apply the `internalMain` pattern to pki-init CLI entry points to raise the ceiling, or add pki-init E2E tests to the CI/CD pipeline.

---

### 11. Cross-version documentation drift — tls-structure.md vs ENG-HANDBOOK.md

**Severity**: LOW — Documentation inconsistency
**Version**: v11 → v12

v11 extensively updated `docs/tls-structure.md` with the 14-category cert structure. v11 Phase 6 added a section to ENG-HANDBOOK.md (§6.11.3). However, subsequent v12 changes to cert categories (Cat 9 infra entity, Cat 14 postgres-only scoping) updated the generator code and tls-structure.md but may not have propagated to ENG-HANDBOOK.md.

**Impact**: ENG-HANDBOOK.md may describe the v11 cert structure rather than the post-v12 structure. This violates the "ENG-HANDBOOK.md is the single source of truth" principle.

**Recommended fix**: Part 3 of the current task (deep analysis of docs) will address this.

---

### 12. v10 template directory never validated against live deployments via Docker

**Severity**: LOW — Template linter runs but Docker Compose not tested
**Version**: v10

v10 created ~63 template files and a `template-compliance` linter that compares templates against actual deployment files. The linter runs via `golangci-lint` / pre-commit but does not validate that the deployment files actually work (i.e., `docker compose up` succeeds). The templates could be structurally correct but functionally broken.

**Impact**: Template compliance checks structural match, not functional correctness.

**Recommended fix**: v13 E2E tests will indirectly validate this — if Docker Compose deployments work, the templates are functionally correct.

---

### 13. v12 Phase estimation accuracy was poor

**Severity**: LOW — Process improvement opportunity
**Version**: v12

v12's plan.md estimated phase durations (e.g., Phase 1: 5h, Phase 3: 2h, Phase 4: 5h). No actual durations were recorded in tasks.md or lessons.md. Without actuals, estimation calibration is impossible.

**Impact**: Cannot improve estimation accuracy for v13 planning.

**Recommended fix**: v13 tasks.md should track estimated vs actual hours per phase.

---

### 14. Inconsistent task status notation across versions

**Severity**: LOW — Readability issue
**Version**: v10, v11, v12

v10 uses `✅` only (all complete). v11 uses `✅` and `⚠️ PARTIAL`. v12 uses `✅`, `⏳ DEFERRED`, and `☐ TODO`. The status format, phase header format, and checkbox style vary across versions.

**Impact**: No consistent way to scan across all three versions for incomplete work.

**Recommended fix**: Standardize on v12's notation (`✅ COMPLETE`, `⏳ DEFERRED`, `☐ TODO`) for v13.

---

### 15. v11 Phase 6 over-estimated — systemic estimation bias toward overcount

**Severity**: LOW — Calibration data
**Version**: v11

Phase 6 (Knowledge Propagation) was estimated at 2h but took 0.5h. Phase 4 (Template & Deployment Updates) was estimated at 3h but took 0.5h. v11 lessons.md explicitly notes this pattern. The root cause: estimates assumed agent/skill/instruction updates would be needed, but implementation used existing patterns with no new artifacts.

**Impact**: Inflated total estimates reduce credibility and make planning less useful.

**Recommended fix**: For v13, estimate Knowledge Propagation phases at ≤1h unless lessons explicitly identify new patterns requiring artifact updates. Estimate verification/documentation phases at 50% of implementation phase estimates.
