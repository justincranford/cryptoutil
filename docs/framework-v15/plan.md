# Implementation Plan - Framework V15: Pre-Flight Gap Fixes + OTel/Grafana mTLS + Public App TLS Trust

**Status**: Active
**Created**: 2026-04-16
**Last Updated**: 2026-04-22
**Purpose**: First, close all CRITICAL/HIGH gaps from the deep analysis (`gaps.md`) that undermine
CI/CD integrity and code correctness. Then wire OTel Collector receiver mTLS, Grafana LGTM HTTPS
UI + OTLP ingest mTLS, and public PS-ID app server TLS (Cat 3/4) into all deployment templates.
V15 directly continues V12 (PostgreSQL mTLS + private admin mTLS) and V13/V14 (completion cycles),
completing the full TLS wiring across the cryptoutil suite.

**Prerequisites**:
- `docs/framework-v14/` is COMPLETE — all 28 tasks across 6 phases ✅. Directory deleted.
- V12 Phase 0 generated Cat 9 infra cert (`otel-collector-contrib-https-client-entity-infra`).
  V15 Phase 1 generates Cat 2, Cat 3, Cat 4, Cat 8, Cat 9 app — independent of V12 Phases 1-11.

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

**ALL issues are blockers — NO exceptions:**
- ✅ **Fix issues immediately** — When unknowns discovered, blockers identified, any tests fail, or
  quality gates not met, STOP and address
- ✅ **Treat as BLOCKING** — ALL issues block progress to next phase
- ✅ **Document root causes** — Root cause analysis is mandatory; planning blockers resolved during
  planning, implementation blockers resolved during implementation
- ✅ **NEVER defer**: No "fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase/task/step complete with known issues

---

## Overview

V15 has two distinct sections:

**Section A — Phase 0: Pre-Flight Gap Fixes (6h)**

Before any TLS wiring begins, fix the CRITICAL and HIGH gaps identified in `gaps.md`. These gaps
undermine CI/CD integrity (`lint-docs`/`lint-deployments` not enforced in CI, coverage gates
bypassed) and code correctness (`sm-kms` missing shutdown timeout). Fixing them FIRST ensures the
TLS implementation phases operate on trustworthy test and lint infrastructure.

**Section B — Phases 1–12: OTel/Grafana mTLS + Public App TLS Trust (36h)**

Wire TLS across the entire telemetry pipeline and application public endpoints:
- **Phase 1**: pki-init patch — generate Cat 2 (OTel/Grafana server), Cat 3 (PS-ID public server),
  Cat 4 (PS-ID public client CA), Cat 8 (OTel/Grafana client CA), Cat 9 app (app→OTel client)
- **Phases 2–7**: OTel Collector server TLS → app→OTel client mTLS → Grafana HTTPS + OTLP ingest
  → OTel→Grafana client mTLS → full pipeline verification
- **Phase 8**: Public PS-ID app server TLS (Cat 3) with client CA enforcement (Cat 4)
- **Phases 9–11**: Deployment templates, linting, and full-stack deployment verification
- **Phase 12**: Knowledge propagation to ENG-HANDBOOK.md and permanent artifacts

---

## Background

### V12 — PostgreSQL mTLS + Private Admin mTLS

V12 completed all PostgreSQL TLS wiring (leader server TLS → replication TLS → app client mTLS →
replication client mTLS) and private admin mTLS trust (Cat 7+14 certs on admin port 9090). V12
Phase 0 generated Cat 9 infra cert (`otel-collector-contrib-https-client-entity-infra`) in advance.

### V13/V14 — Completion and Quality Passes

V13 focused on quality improvements and deferred items from V10–V12. V14 closed all V13 deferred
items across 6 phases (28 tasks, all ✅ complete). V15 begins from a clean baseline.

---

## Background Lessons from V14 — Carried Into V15

These lessons from V14 execution are explicitly applied as constraints and patterns in V15.

### CI/CD and Quality Infrastructure

- **Run `go run ./cmd/cicd-lint lint-go ./...` FIRST** before any linting work — this catches
  `literal-use` violations that block `TestLint_Integration`. V14 Phase 4 discovered 33 violations
  mid-phase because the baseline was not checked upfront.
- **Magic constants: ALWAYS search `internal/shared/magic/` BEFORE writing ANY literal** — bare
  string or numeric literals violate `literal-use` and block CI. This applies to both test and
  production code equally.
- **Propagation integrity: run `go run ./cmd/cicd-lint lint-docs` IMMEDIATELY after any
  ENG-HANDBOOK.md change** — before every commit touching that file.

### Docker Compose YAML Syntax

- **`start_period` uses underscore in YAML; `--start-period` uses hyphen in Dockerfiles** — these
  are different syntaxes. V14 Phase 2 caused a rework pass because `start-period` (hyphen) was
  written in YAML where `start_period` (underscore) is required.
- **ALL 4 deployment variants need explicit entries** — do not configure sqlite-1 only and assume
  others inherit. V14 Phase 2 initially missed sqlite-2, postgres-1, postgres-2, requiring a second
  pass. All 4 variants = separate entries in compose files and config files.
- **Rebuild Docker images before E2E phases** — stale images cause spurious E2E failures that waste
  investigation time. Add `docker compose build` as the first step in Phases 4, 7, and 11.

### Coverage and Testing

- **Two test paths required for production closure coverage**:
  1. Stub tests (`ExportedNewTestXxx`) — test control flow and error paths
  2. Production wiring tests (`ExportedProductionNewXxx`) — invoke real closures to cover their
     bodies (creating the struct does NOT cover closure bodies)
- **`attempts++` mutation pattern** — kill the mutation by including the count in the error message
  AND asserting the error string does NOT contain `"after 0 attempts"`.
- **`make` capacity hints are structural ceilings** — `make(map[K]V, len(xs))` capacity mutations
  (`len(xs)→0`) are invisible to black-box tests. Document as ceiling; do NOT chase.
- **Budget ~30s per TIMED OUT mutation** when estimating gremlins run time. TIMED OUT ≠ LIVED.

### PostgreSQL mTLS Identity

- **Use `client_dn` (NOT `application_name`) for PostgreSQL mTLS identity** in `pg_stat_ssl`.
  GORM does not set `application_name` by default — it is always empty. Query:
  `WHERE pg_stat_ssl.client_dn LIKE '%-sm-kms-%'`

### Commit and Documentation Discipline

- **`time.Duration` constants MUST NOT have unit suffixes** (`Ms`, `Ns`, `Sec`, `Min`) — violates
  staticcheck ST1011. Correct: `DefaultPollInterval = 5 * time.Second`. Wrong: `DefaultPollIntervalMs = 5000`.
- **Atomic commits per section for large docs** — for ENG-HANDBOOK.md edits across multiple
  sections, commit per section, not per task. Avoids massive single-commit diffs impossible to bisect.
- **Enumerate ALL affected files early with derivation formula** — write
  `deployments/{sm-kms,...}/compose.yml (10 files = 1 per PS-ID)` not just "compose files". Raw
  counts without formulas are unverifiable during review.

---

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: `internal/apps/framework/service/` (dual HTTPS listeners, builder pattern)
- **TLS Generator**: `internal/apps/framework/tls/generator.go` (14 certificate categories)
- **Database**: PostgreSQL + SQLite with GORM
- **Telemetry**: OTel Collector (`otel/opentelemetry-collector-contrib`) → Grafana LGTM
- **Cert Categories Used in V15**:
  - Cat 2: `public-https-server-entity-{otel-collector-contrib,grafana-otel-lgtm}` — OTel + Grafana server TLS
  - Cat 3: `public-https-server-entity-{PS-ID}-{variant}` — PS-ID public server TLS (40 = 4×10)
  - Cat 4: `public-https-client-issuing-ca-{PS-ID}-{sqlite,postgres}/truststore/` — PS-ID public client CA (20 = 2×10)
  - Cat 8: `{otel-collector-contrib,grafana-otel-lgtm}-https-client-issuing-ca/truststore/` — OTel/Grafana client CA
  - Cat 9 app: `otel-collector-contrib-https-client-entity-{PS-ID}-{variant}` — app→OTel client cert (40 = 4×10)
  - Cat 9 infra (from V12): `otel-collector-contrib-https-client-entity-infra` — OTel→Grafana
- **Affected Files** (enumerated):
  - Phase 0: `.github/workflows/ci-*.yml` (~7 files), `internal/apps/sm-kms/server/server.go`,
    `internal/apps/identity-authz/server/server.go`, `internal/apps/{sm,sm-kms,sm-im,jose,jose-ja,pki,pki-ca}/usage.go` (7 files + new shared util),
    `internal/apps/pki-ca/server/testmain_test.go`, service entry points for signal handling (10 files),
    `.pre-commit-config.yaml`, `docs/tls-structure.md`
  - Phase 1: `internal/apps/framework/tls/generator.go` + generator tests
  - Phases 2–3: `deployments/shared-telemetry/otel-collector-contrib/config.yml`,
    `internal/apps/framework/service/config/` (OTLP TLS fields),
    `deployments/{PS-ID}/config/*-app-framework-{variant}.yml` (40 files = 4×10),
    `deployments/{PS-ID}/compose.yml` (10 files = 1 per PS-ID)
  - Phase 5: `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini` (new),
    `deployments/shared-telemetry/compose.yml`
  - Phase 8: `internal/apps/framework/service/config/` (public TLS fields),
    `deployments/{PS-ID}/config/*-app-framework-{variant}.yml` (40 files again)
  - Phases 9–11: `docs/deployment-templates.md`,
    `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`,
    `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`

---

## Phases

**Phase Status Legend**: `☐ TODO` | `🔄 IN PROGRESS` | `✅ COMPLETE` | `⏳ BLOCKED`

---

### Phase 0: Pre-Flight Gap Fixes (6h) [Status: ☐ TODO]

**Objective**: Fix all CRITICAL and HIGH gaps from `gaps.md` before TLS work begins. These gaps
undermine CI/CD integrity and code correctness and MUST be resolved before adding new TLS complexity.

**Priority order within phase** (fix in this order):
1. CRITICAL CI/CD gaps (Tasks 0.1–0.3): Gate all CI/CD quality enforcement
2. HIGH code correctness bugs (Task 0.4): `sm-kms` shutdown hangs on bug
3. HIGH refactoring (Task 0.5): `usage.go` duplication; can defer to V16 if blocked
4. Medium fixes (Tasks 0.6–0.8): Small, low-risk, batch together
5. Documentation (Task 0.9): `tls-structure.md` updates from doc gap analysis

**V14 anti-pattern warning**: Run `go run ./cmd/cicd-lint lint-go ./...` FIRST to establish
baseline before any code changes. V14 Phase 4 discovered 33 `literal-use` violations mid-phase
because the baseline was not checked upfront.

- **Success**: `lint-docs`/`lint-deployments` run in CI; `ci-coverage.yml` actually blocks on
  coverage failure; `sm-kms` shutdown is bounded; all changes pass `golangci-lint run` and
  `go test ./...` clean.
- **Post-Mortem**: After quality gates pass, update `lessons.md` Phase 0 section.

---

### Phase 1: pki-init Patch — Cat 2, Cat 3, Cat 4, Cat 8, Cat 9 app (4h) [Status: ☐ TODO]

**Objective**: Add generator calls for all V15 certificate categories to `pki-init`.

**V14 lessons applied**:
- Read `internal/apps/framework/tls/generator.go` fully (2000+ lines) BEFORE editing.
- Add `// Cat N: <name>` comments at each new call site for cross-reference.
- BOTH stub tests (ExportedNewTestXxx) AND production closure tests (ExportedProductionNewXxx)
  are required — creating the struct does NOT cover closure bodies.

**Cert Entities Generated** (from `internal/apps/framework/tls/generator.go`):
- Cat 2: `public-https-server-entity-otel-collector-contrib`, `public-https-server-entity-grafana-otel-lgtm` (2 global entities)
- Cat 3: `public-https-server-entity-{PS-ID}-{sqlite-1,sqlite-2,postgres-1,postgres-2}` (40 = 4×10)
- Cat 4: `public-https-client-issuing-ca-{PS-ID}-{sqlite-domain,postgres-domain}/truststore/` (20 = 2×10)
- Cat 8: `otel-collector-contrib-https-client-issuing-ca`, `grafana-otel-lgtm-https-client-issuing-ca` (2 global entities)
- Cat 9 app: `otel-collector-contrib-https-client-entity-{PS-ID}-{sqlite-1,sqlite-2,postgres-1,postgres-2}` (40 = 4×10)

- **Success**: `go run ./cmd/pki-ca pki-init` generates all V15 cert dirs; integration test verifies
  directory count; generator coverage ≥98%; mutation score ≥98%.
- **Post-Mortem**: Update `lessons.md` Phase 1.

---

### Phase 2: OTel Collector Server TLS (3h) [Status: ☐ TODO]

**Objective**: Add TLS to OTel Collector OTLP receivers (gRPC :4317 + HTTP :4318).

- **OTel config**: `deployments/shared-telemetry/otel-collector-contrib/config.yml` — add `tls:`
  block to `receivers.otlp.protocols.grpc` and `receivers.otlp.protocols.http` pointing to Cat 2
  `public-https-server-entity-otel-collector-contrib/` cert+key.
- **Cat 8 truststore**: `otel-collector-contrib-https-client-issuing-ca/truststore/` as
  `client_ca_file` to enforce client cert verification.
- **Compose mounts**: `deployments/shared-telemetry/compose.yml` — mount Cat 2 server cert dir +
  Cat 8 client CA truststore (least privilege; NO Cat 9 yet).
- **V14 lesson**: `start_period` (underscore) in YAML; `--start-period` (hyphen) in Dockerfiles.
- **Success**: OTel Collector starts with TLS; `openssl s_client -connect localhost:4317` shows Cat
  2 cert; HTTP :4318 also shows TLS.
- **Post-Mortem**: Update `lessons.md` Phase 2.

---

### Phase 3: App→OTel Client mTLS (4h) [Status: ☐ TODO]

**Objective**: Configure framework OTLP exporter to present Cat 9 app client certs to OTel.

- **Framework config fields**: `internal/apps/framework/service/config/` — add `otlp.tls.cert-file`,
  `otlp.tls.key-file`, `otlp.tls.ca-file` to ServerSettings (optional; absent = insecure or
  unauthenticated TLS fallback).
- **Deployment configs**: `deployments/{PS-ID}/config/*-app-framework-{variant}.yml` — add OTLP TLS
  fields per variant (40 files = 4 variants × 10 PS-IDs). Change `otlp.endpoint` from `http://` to
  `https://` for all variants.
- **Compose mounts**: `deployments/{PS-ID}/compose.yml` — mount Cat 9 app keystore per variant (10
  files = 1 per PS-ID).
- **V14 lesson**: ALL 4 variants need separate entries. Do NOT configure sqlite-1 and assume others
  inherit — they do not.
- **Success**: App instances connect to OTel via mTLS; non-mTLS connections rejected.
- **Post-Mortem**: Update `lessons.md` Phase 3.

---

### Phase 4: Verify OTel Standalone (2h) [Status: ☐ TODO]

**Objective**: Start pki-init + OTel + 1 PS-ID; verify server TLS and app→OTel mTLS before Grafana
adds new variables.

- **Start stack**: `docker compose build && docker compose up pki-init otel-collector-contrib {PS-ID}-sqlite-1 {PS-ID}-postgres-1`
- **Verify**: OTel gRPC :4317 TLS active; HTTP :4318 TLS active; app OTLP exporter connects
  successfully with Cat 9 app cert; connections without client cert rejected.
- **V14 lesson**: `docker compose build` BEFORE starting — stale images cause spurious failures.
- **Post-Mortem**: Update `lessons.md` Phase 4.

---

### Phase 5: Grafana LGTM HTTPS + OTLP Ingest TLS (3h) [Status: ☐ TODO]

**Objective**: Enable HTTPS on Grafana UI (:3000) using D1 approach, and configure OTLP ingest TLS
(D6 verification step).

- **D1 (grafana.ini)**: Create `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini` with
  `[server]` block: `protocol = https`, `cert_file`, `cert_key` pointing to Cat 2
  `public-https-server-entity-grafana-otel-lgtm/` paths.
- **D6 (OTLP ingest TLS)**: Configure Grafana OTLP ingest (:14317/:14318) with Cat 8
  `grafana-otel-lgtm-https-client-issuing-ca/truststore/` as `client_ca_file`. If Grafana image
  does not support this, document finding and pivot to D6=C (OTel sidecar) — create fix task.
- **Compose mounts**: Cat 2 server cert dir + Cat 8 client CA truststore + `grafana.ini` config.
  Healthcheck: `https://127.0.0.1:3000/api/health` with `--cacert`.
- **V14 lesson**: `start_period` underscore in YAML. Test BOTH Grafana UI HTTPS AND OTLP ingest in
  this phase — they use the same Cat 2 cert, so verify both before moving on.
- **Post-Mortem**: Update `lessons.md` Phase 5.

---

### Phase 6: OTel→Grafana Client mTLS (2h) [Status: ☐ TODO]

**Objective**: Configure OTel Collector exporter to present Cat 9 infra cert when forwarding to
Grafana (Cat 9 infra was generated in V12 Phase 0).

- **OTel config**: `otel-collector-contrib/config.yml` exporters section — add `tls:` block with
  `cert_file`/`key_file` = Cat 9 infra `otel-collector-contrib-https-client-entity-infra/`,
  `ca_file` = Cat 1 `public-https-server-issuing-ca/truststore/` (to verify Grafana server cert),
  `endpoint` changed to `https://grafana-otel-lgtm:14317`.
- **Compose mounts**: Add Cat 9 infra keystore + Cat 1 truststore to OTel compose. Total OTel mounts
  after this phase: Cat 1 truststore + Cat 2 keystore + Cat 8 truststore + Cat 9 infra keystore
  (exactly 4 dirs; verify no extras).
- **Success**: OTel→Grafana pipeline active; non-mTLS OTel→Grafana forwarding rejected.
- **Post-Mortem**: Update `lessons.md` Phase 6.

---

### Phase 7: Verify OTel→Grafana Pipeline (2h) [Status: ☐ TODO]

**Objective**: Full telemetry pipeline verification: app→OTel→Grafana mTLS chain working end-to-end.

- **Verify**: Traces visible in Grafana Tempo; metrics in Mimir/Prometheus; logs in Loki; Grafana
  HTTPS UI accessible at `https://localhost:3000`; OTel→Grafana mTLS rejection test passes.
- **V14 lesson**: `docker compose build` before starting if production code changed since last build.
- **Post-Mortem**: Update `lessons.md` Phase 7.

---

### Phase 8: Public PS-ID App Server TLS (3h) [Status: ☐ TODO]

**Objective**: Configure framework public listener (:8080) to serve Cat 3 cert and verify Cat 4
client CAs for optional client cert enforcement.

- **Framework config fields**: `internal/apps/framework/service/config/` — add
  `server.public-tls-cert-file`, `server.public-tls-key-file`, `server.public-tls-client-ca-file`
  to ServerSettings (all optional; absent = existing auto-TLS behavior unchanged).
- **Deployment configs**: `deployments/{PS-ID}/config/*-app-framework-{variant}.yml` — add public
  TLS fields per variant (40 files = 4 variants × 10 PS-IDs, same 40 files as Phase 3).
- **Compose mounts**: `deployments/{PS-ID}/compose.yml` — mount Cat 3 + Cat 4 dirs per variant
  (combined with V12 Cat 6+7+10+14 mounts and Phase 3 Cat 9 app mounts).
- **V14 lesson**: ALL 4 variants need entries. Write test verifying fallback (no config → auto-TLS).
- **Success**: `GET /service/api/v1/health` via HTTPS shows Cat 3 cert for all 4 variants;
  connections without valid client cert rejected when Cat 4 CA configured.
- **Post-Mortem**: Update `lessons.md` Phase 8.

---

### Phase 9: Deployment Templates (5h) [Status: ☐ TODO]

**Objective**: Update canonical deployment templates to encode V15 cert mounts and config fields.

- **V14 lesson**: Config correctness verified in a RUNNING stack (Phases 4, 7, 8) BEFORE encoding
  into canonical templates. Do NOT write templates based on untested configs.
- **Files**:
  - `docs/deployment-templates.md` — combined V12+V15 least privilege table; OTel/Grafana cert dirs; App cert dirs per variant
  - `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml` — OTel + Grafana cert mounts, grafana.ini mount
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` — Cat 3+4+9app mounts per variant (combined with V12 mounts)
- **Success**: Template compliance linter accepts all files; `__PS_ID__` placeholders consistent;
  `go run ./cmd/cicd-lint lint-deployments` passes.
- **Post-Mortem**: Update `lessons.md` Phase 9.

---

### Phase 10: Deployment Linting (1h) [Status: ☐ TODO]

**Objective**: All updated deployment files pass `lint-deployments` validators.

- **V14 lesson**: Lint BEFORE deployment verification — catch structural errors before starting
  Docker Compose. Cost is low; value is catching errors early.
- **Run**: `go run ./cmd/cicd-lint lint-deployments` — must exit 0 with all 8 validators passing.
- **Also**: `go test ./internal/apps/tools/cicd_lint/lint_deployments/...` passes.
- **Post-Mortem**: Update `lessons.md` Phase 10.

---

### Phase 11: Deployment Verification — Full Telemetry Stack (3h) [Status: ☐ TODO]

**Objective**: Start complete deployment and verify all TLS chain endpoints function correctly.

- **Stack**: pki-init + shared-telemetry (OTel + Grafana) + one PS-ID (all 4 variants).
- **V14 lesson**: `docker compose build` BEFORE `docker compose up` when production code changed.
- **Verify**:
  - All 4 PS-ID variants healthy (public HTTPS on :8080 with Cat 3; admin on :9090)
  - App→OTel mTLS working (Cat 9 app certs verified in OTel logs)
  - OTel→Grafana mTLS working (Cat 9 infra cert)
  - Grafana HTTPS UI accessible at `https://localhost:3000`
  - Non-mTLS connections rejected at ALL 3 enforcement points (OTel receiver, Grafana OTLP ingest,
    app public Cat 4 CA if configured)
  - Traces, metrics, logs visible in Grafana dashboards
- **Post-Mortem**: Update `lessons.md` Phase 11.

---

### Phase 12: Knowledge Propagation (2h) [Status: ☐ TODO]

**Objective**: Apply all lessons to permanent artifacts. NEVER skip this phase.

1. Review `lessons.md` entries from ALL prior phases
2. Update `ENG-HANDBOOK.md`:
   - §9.4: OTel Collector mTLS receiver config (gRPC `tls:` block pattern)
   - §9.4: Grafana HTTPS `grafana.ini` approach (D1) documented
   - §9.4: OTel→Grafana mTLS forwarding pattern
   - §12/§13: Public PS-ID app server TLS (Cat 3/4) deployment pattern
   - §13: Combined V12+V15 cert mount least privilege table
3. Update agents/skills/instructions where new patterns were discovered during implementation
4. Run `go run ./cmd/cicd-lint lint-docs` — propagation integrity MUST pass
5. Update `docs/deployment-templates.md` final combined V12+V15 table (from Phase 9)
6. Commit per section for ENG-HANDBOOK.md changes (V14 lesson: atomic section commits)
- **Post-Mortem**: Update `lessons.md` Phase 12.

---

## Decisions

### Decision 1 (D1): Grafana HTTPS UI Configuration Approach

**Options**:
- A: Custom `grafana.ini` with `[server]` block ✓ **SELECTED**
- B: Environment variables (banned per project policy — NO)
- C: Grafana provisioning files

**Decision**: Option A — mount a custom `grafana.ini` with explicit TLS cert/key paths.

**Rationale**: Clean, idiomatic Grafana configuration. Environment variables are banned per
ENG-HANDBOOK.md Section 9.2. Provisioning files are for datasource/dashboard provisioning, not
server TLS.

---

### Decision 6 (D6): Grafana OTLP Ingest TLS Support

**Options**:
- A: Configure Grafana OTLP ingest TLS directly (Cat 8 truststore, `client_ca_file`) ✓ **ASSUMED**
- B: Grafana OTLP ingest does not support TLS — keep existing behavior (accept plaintext from OTel)
- C: OTel sidecar receives from apps over mTLS; forwards to Grafana OTLP over internal loopback

**Decision**: Option A assumed. Verify empirically in Phase 5 Task 5.2. If grafana/otel-lgtm image
does not support OTLP ingest mTLS configuration, document finding and pivot to Option C (sidecar
wraps Grafana OTLP endpoint).

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| D6: Grafana OTLP ingest does not support TLS | Medium | Medium | Pivot to OTel sidecar (D6=C) in Phase 5 |
| Phase 0 Task 0.5 (usage.go refactor) too large | Medium | Low | Defer to V16 with GAP file; Phase 0 still complete |
| Phases 3/8: 40 deployment config files required | Medium | High | Enumerate all 40 explicitly; CI file-count check |
| OTel TLS receiver config not documented in image | Low | High | Test with real container in Phase 4 before Phase 5 |
| Docker image rebuild forgotten before E2E | High | Medium | Explicit `docker compose build` step in Phases 4, 7, 11 |
| Cat 9 infra cert path wrong (V12 generated) | Low | High | Verify Cat 9 infra path in Phase 6 before configuring OTel |

---

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- ✅ All tests pass (`go test ./...`) — 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) — zero errors
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) — zero
  warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code (pki-init generator, cicd_lint): ≥98% line coverage
- ✅ main() functions: 0% acceptable if internalMain() ≥95%
- ✅ Generated code: Excluded from coverage

**Mutation Testing Targets**:
- ✅ pki-init generator changes: ≥98% mutation efficacy
- ✅ Framework config changes: ≥95% mutation efficacy

**Per-Phase Quality Gates**:
- ✅ Unit + integration tests complete before moving to next phase
- ✅ `docker compose build` before E2E phases (Phases 4, 7, 11) — V14 lesson
- ✅ Docker Compose health checks pass (Phases 4, 7, 11)
- ✅ `go run ./cmd/cicd-lint lint-deployments` passes (Phase 10)
- ✅ `go run ./cmd/cicd-lint lint-docs` passes (Phase 12)
- ✅ Race detector clean: `go test -race -count=2 ./...`

**V15-Specific Quality Gates**:
- ✅ ALL 4 deployment variants (sqlite-1, sqlite-2, postgres-1, postgres-2) verified — no skipping
- ✅ Baseline `lint-go` run BEFORE any code changes (Phase 0 anti-pattern prevention)
- ✅ Template compliance linter passes after Phase 9 changes

---

## Success Criteria

- [ ] Phase 0: All CRITICAL/HIGH gaps from `gaps.md` fixed; CI/CD enforces `lint-docs` +
  `lint-deployments`; `ci-coverage.yml` blocks on failures; `sm-kms` shutdown bounded
- [ ] Phase 1: pki-init generates all V15 cert categories (Cat 2, 3, 4, 8, 9 app); ≥98% generator
  coverage; ≥98% mutation score
- [ ] Phase 2: OTel Collector serves TLS on :4317 (gRPC) and :4318 (HTTP); Cat 8 client CA enforced
- [ ] Phase 3: App instances present Cat 9 app client cert to OTel; non-mTLS rejected
- [ ] Phase 4: OTel standalone verified — server TLS + app→OTel mTLS + rejection test
- [ ] Phase 5: Grafana serves HTTPS UI (Cat 2 cert); Grafana OTLP ingest accepts OTel mTLS (Cat 8)
- [ ] Phase 6: OTel presents Cat 9 infra cert to Grafana; OTel→Grafana mTLS rejection verified
- [ ] Phase 7: Full telemetry pipeline verified end-to-end (traces/metrics/logs in Grafana)
- [ ] Phase 8: PS-ID apps serve Cat 3 cert on public :8080 (all 4 variants); Cat 4 CA enforced
- [ ] Phase 9: Deployment templates updated for combined V12+V15 cert mounts; compliance linter passes
- [ ] Phase 10: `lint-deployments` passes (all 8 validators)
- [ ] Phase 11: Full deployment stack verified (all 4 variants + OTel + Grafana)
- [ ] Phase 12: ENG-HANDBOOK.md updated; `lint-docs` propagation passes; clean working tree
- [ ] Evidence archived in `test-output/v15-*/`

---

## Notes

- **Least Privilege Enforcement**: Every compose template task MUST list exactly which Cat dirs are
  mounted and explicitly note what is NOT mounted.
- **`./configs/` isolation**: No changes to `./configs/` files — they continue using auto-TLS only.
- **V15 depends on V12 Phase 0**: Cat 9 infra cert was generated in V12 Phase 0. If V12 is
  incomplete, generate Cat 9 infra manually before starting Phase 6.
- **D6 contingency**: If grafana/otel-lgtm does not support OTLP ingest mTLS, pivot to OTel sidecar
  (D6=C) in Phase 5 — document the finding, create the sidecar task.
- **V14 carry-forward (CLI YAML config)**: `internal/apps/framework/service/cli/` subcommands
  (`livez`, `readyz`, `shutdown`) are CLI-args-only. They should be extended to support YAML config
  files (priority: Docker secrets > YAML > CLI), allowing PS-ID instances to reuse existing config
  files. Track in V16 as a dedicated phase.
- **Phase 0 Task 0.5 carry-forward**: If usage.go refactor proves too large for Phase 0, defer to
  V16 with a GAP file documenting current state, target state, and acceptance criteria.

---

## ENG-HANDBOOK.md Cross-References - MANDATORY

| Topic | ENG-HANDBOOK.md Section | Applicability |
|-------|------------------------|----|
| Testing Strategy | [§10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | Phase 1 pki-init tests, Phases 3/8 framework tests |
| Coverage Targets | [§10.2.3](../../docs/ENG-HANDBOOK.md#1023-coverage-targets) | ≥98% pki-init, ≥95% framework; production closure ceiling |
| Test Seam Injection | [§10.2.4](../../docs/ENG-HANDBOOK.md#1024-test-seam-injection-pattern) | Testing cert loading in framework (Phases 3, 8) |
| Mutation Testing | [§10.5](../../docs/ENG-HANDBOOK.md#105-mutation-testing-strategy) | ≥98% pki-init; attempts++ pattern; capacity hint ceilings |
| Quality Gates | [§11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL phases |
| Coding Standards | [§14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Phases 0, 1, 3, 8 |
| Version Control | [§14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | Atomic section commits for ENG-HANDBOOK.md |
| Deployment Architecture | [§12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) | Phases 9–11 |
| Service Framework | [§5.1](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern) | Phases 3/8 framework config changes |
| OTel/Telemetry | [§9.4](../../docs/ENG-HANDBOOK.md#94-telemetry-strategy) | Phases 2–7 OTel/Grafana wiring |
| OTel Processor Constraints | [§9.4.1](../../docs/ENG-HANDBOOK.md#941-otel-collector-processor-constraints) | Phase 2 OTel TLS config (`detectors: [env, system]`) |
| CI/CD Workflow Architecture | [§9.7](../../docs/ENG-HANDBOOK.md#97-cicd-workflow-architecture) | Phase 0 CI/CD gap fixes |
| Pre-Commit Architecture | [§9.9](../../docs/ENG-HANDBOOK.md#99-pre-commit-hook-architecture) | Phase 0 golangci-lint pin; lint-docs enforcement |
| Infrastructure Blockers | [§14.7](../../docs/ENG-HANDBOOK.md#147-infrastructure-blocker-escalation) | D6 fallback; all Phase 0 CI/CD gaps are BLOCKING |
| Plan Lifecycle | [§14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | ALL phases |
| Post-Mortem & Propagation | [§14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | Every phase post-mortem + Phase 12 knowledge propagation |
