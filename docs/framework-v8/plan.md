# Implementation Plan — Framework v8: Deployment Parameterization

**Status**: Complete
**Created**: 2026-04-05
**Last Updated**: 2026-04-06
**Purpose**: Eliminate ~2,800 lines of copy-paste across PRODUCT and SUITE compose files by
implementing recursive Docker Compose `include:` with Approach C (inline service redefinition
for port and secret overrides). Simultaneously resolve naming inconsistencies, missing service
definitions, and shared-infrastructure import gaps. PostgreSQL architecture: single shared
leader/follower pair at all tiers with no host port exposure (Q1=C, Q2=E). Product-level
Dockerfiles cancelled — PS-ID Dockerfiles used transitively (Q3=D).

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL compose files must start cleanly; validators must pass
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation (docker compose up + lint-deployments) at every step
- ✅ **Reliability**: Quality gates enforced (lint-deployments, lint-ports, lint-compose)
- ✅ **Efficiency**: Optimized for maintainability (canonical definitions in PS-ID only)
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers — NO exceptions.**

---

## Overview

The `deployments/` directory currently has ~6,636 lines across 16 compose files. PRODUCT and
SUITE levels copy-paste all service definitions from PS-ID levels, with only port numbers changed.
Any change to a PS-ID service (healthcheck, resource limit, dependency) must be replicated 2-3
times. This is error-prone and has already produced drift (SM product missing sm-im, inconsistent
naming, missing sqlite-2 at PRODUCT/SUITE).

The solution is Docker Compose `include:` with **Approach C** (inline service redefinition):
PRODUCT includes PS-ID compose files and redefines port mappings inline. SUITE includes PRODUCT
compose files and redefines ports again. Service definitions live ONLY in PS-ID compose files.

**Three open design decisions** were captured in `quizme-v1.md` and have been **resolved**:
- **Q1=C**: shared-postgres included at all tiers; per-PS-ID postgres services removed entirely;
  no host port exposure; developers use `docker exec postgres-leader psql`
- **Q2=E**: Single PostgreSQL leader/follower pair for all tiers; each PS-ID connects with
  different username/password/logical database; follower replicates all logical databases to
  separate schemas. No host ports at any tier (consistent with Q1=C).
- **Q3=D**: Carryover Item 2 (product Dockerfiles) permanently cancelled; `validate_structure.go`
  updated to remove Dockerfile requirement from PRODUCT level

---

## Background

**Framework v7 outcome**: Service framework adoption complete for sm-kms, sm-im, jose-ja,
skeleton-template; in progress for pki-ca and identity services.

**Carryover from v7**:
- Item 2: Product-level Dockerfiles (HIGH) — **CANCELLED** (Q3=D). Product deployments use
  PS-ID Dockerfiles transitively via recursive includes.
- Item 3: Fitness linter `usage_health_path_completeness` (MEDIUM) — included in Phase 6
- Item 7: Load test multi-tier scenarios (LOW) — deferred to future plan

---

## Technical Context

- **Language**: Go 1.26.1 (validators)
- **Tool**: Docker Compose v2.24+ (required for full include + service override support)
- **Pattern**: Approach C — service redefinition in including compose file overrides included service
- **Validator**: `go run ./cmd/cicd-lint lint-deployments` — must pass after every phase
- **Related Dirs**: `deployments/`, `configs/`, `internal/apps/tools/cicd_lint/lint_deployments/`,
  `internal/apps/tools/cicd_lint/lint_ports/`, `internal/apps/tools/cicd_lint/lint_fitness/`

---

## Phases

### Phase 0: Technical Research (0.5h) [Status: ☑ COMPLETE]

**Objective**: Validate Docker Compose include + service override behavior before committing to
the full recursive include design. Resolve work.md Open Questions #1–5.

**Research Tasks**:
1. Create a minimal 3-file test (shared/compose.yml, psid/compose.yml, product/compose.yml)
   - Verify: service defined in shared included from psid can be overridden in psid compose
   - Verify: when product includes psid (which includes shared), shared is not double-included
   - Verify: PRODUCT's service redefinition overrides PSID's redefinition
2. Verify `profiles:` in included files interact correctly with parent `--profile` flag
3. Verify secret `file:` paths resolve from the included file's directory (not the including file)
4. Verify network definitions merge across include hierarchies without errors

**Success**: All 4 behaviors confirmed with working minimal examples (archived in `test-output/framework-v8-research/`).

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 1: Naming Standardization + Missing Services (2h) [Status: ☑ COMPLETE]

**Objective**: Fix all naming inconsistencies and missing service definitions before touching the
include hierarchy. Establishes a clean baseline for subsequent phases.

**Tasks**:
1. Standardize service names to `postgresql` everywhere (PRODUCT and SUITE compose files
   currently mix `postgres` and `postgresql` service names; PS-ID and config files already use
   `postgresql`)
2. Add sm-im services to `deployments/sm/compose.yml` (sm-im-app-sqlite-1, sm-im-app-sqlite-2,
   sm-im-app-postgresql-1, sm-im-app-postgresql-2)
3. Fix SUITE compose.yml per-service port allocations to include sqlite-2 at 28101, 28201, etc.
   (currently only sqlite-1 and postgresql-1/2 at SUITE level)
4. Add sqlite-2 to all PRODUCT compose files (currently only 3 instances per PS-ID at PRODUCT)
5. Correct any other drift found by `go run ./cmd/cicd-lint lint-deployments`

**Success**:
- `go run ./cmd/cicd-lint lint-deployments` passes with ≤ same number of errors as before (no
  new violations introduced)
- SM product compose includes all sm-kms AND sm-im service variants (8 service instances total)
- All compose files use `postgresql` consistently (no occurrences of `postgres-1` or `postgres-2`
  in service names)
- sqlite-2 instances present at every tier for every PS-ID

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 2: Remove Per-PS-ID PostgreSQL + Shared Infrastructure at All Tiers (2h) [Status: ☑ COMPLETE]

**Objective**: Prepare PS-ID compose files to serve as include targets by (a) removing per-PS-ID
PostgreSQL DB services entirely and (b) adding `shared-postgres` and `shared-telemetry` includes
to all PS-ID compose files. Per Q1=C: no host port exposure for postgres at any tier. Per Q2=E:
single shared leader/follower pair serves all PS-IDs with per-PS-ID logical database isolation.

**Tasks**:
1. Remove all per-PS-ID PostgreSQL DB service definitions (e.g., `sm-im-db-postgres-1`) from
   all 10 PS-ID compose files. Remove associated profiles, volumes, and healthchecks.
2. Add `include: ../shared-postgres/compose.yml` to all 10 PS-ID compose files (if not already)
3. Add `include: ../shared-telemetry/compose.yml` to all 10 PS-ID compose files (if not already)
4. Update app service `depends_on` references: change from per-PS-ID postgres to
   `postgres-leader` from shared-postgres
5. Verify PS-ID compose files work: `docker compose up --profile dev` (SQLite mode, no postgres)
6. Verify shared-postgres starts when PS-ID compose is used: `docker compose up --profile postgres`

**Success**:
- Each PS-ID compose starts correctly with `--profile dev` (SQLite) and `--profile postgres`
  (shared-postgres leader/follower)
- No per-PS-ID PostgreSQL DB service definitions remain in any PS-ID compose file
- No host port exposure for postgres-leader or postgres-follower at SERVICE level
- `go run ./cmd/cicd-lint lint-deployments` still passes

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 3: PRODUCT Recursive Includes — Approach C (5h) [Status: ☑ COMPLETE]

**Objective**: Replace each PRODUCT compose file's copy-pasted PS-ID service definitions with
`include:` references to the PS-ID compose files, adding inline port override redefinitions to
remap ports from SERVICE range (8XXX) to PRODUCT range (18XXX).

**Tasks (per product — 5 products)**:

For each PRODUCT (`sm`, `jose`, `pki`, `identity`, `skeleton`):
1. Replace SERVICE-level service definitions with `include:` blocks for each PS-ID
2. Add `services:` section with ONLY port override redefinitions for each included service
3. Keep a single `builder-{PRODUCT}` (remove PS-ID builders from PRODUCT level)
4. Redefine `secrets:` to product-scoped file paths (override PS-ID secrets)
5. Add `include: ../shared-postgres/compose.yml` (if not inherited from PS-ID includes)
6. Test: `docker compose -f deployments/{PRODUCT}/compose.yml up --profile dev`
7. Test: `docker compose -f deployments/{PRODUCT}/compose.yml up --profile postgres`

**Port Override Pattern (Approach C, PRODUCT level)**:
```yaml
# deployments/sm/compose.yml
include:
  - path: ../shared-telemetry/compose.yml
  - path: ../shared-postgres/compose.yml
  - path: ../sm-kms/compose.yml
  - path: ../sm-im/compose.yml

services:
  sm-kms-app-sqlite-1:
    ports:
      - "18000:8080"   # Override SERVICE port 8000 → PRODUCT port 18000
  sm-kms-app-sqlite-2:
    ports:
      - "18001:8080"
  sm-kms-app-postgresql-1:
    ports:
      - "18002:8080"
  sm-kms-app-postgresql-2:
    ports:
      - "18003:8080"
  sm-im-app-sqlite-1:
    ports:
      - "18100:8080"
  # ... (all sm-im instances)
  # NOTE: No postgres-leader port override — no host port exposure per Q1=C
```

**Success**:
- Each PRODUCT compose line count ≤ 150 lines (down from 261-818)
- `docker compose config` renders correctly (no missing service errors)
- All 3-4 variant profiles work at PRODUCT level
- `go run ./cmd/cicd-lint lint-deployments` passes

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 4: SUITE Recursive Includes — Approach C (3h) [Status: ☑ COMPLETE]

**Objective**: Replace the SUITE (cryptoutil) compose file's 1,504 lines with `include:` of 5
PRODUCT compose files, adding inline port override redefinitions to remap from PRODUCT range
(18XXX) to SUITE range (28XXX). No postgres port overrides needed (no host port exposure per Q1=C).

**Tasks**:
1. Replace all SERVICE-level service definitions in `deployments/cryptoutil/compose.yml` with
   `include:` of 5 PRODUCT compose files
2. Add `services:` section with ONLY port override redefinitions (+20000 from SERVICE base)
3. Keep single `builder-cryptoutil`
4. Redefine secrets to suite-scoped file paths
5. Remove redundant direct includes (shared-telemetry, shared-postgres already inherited
   through PRODUCT includes — may need explicit re-include if deduplication causes issues)
6. Test: `docker compose -f deployments/cryptoutil/compose.yml up --profile dev`
7. Test: `docker compose -f deployments/cryptoutil/compose.yml up --profile postgres`

**Expected Port Overrides (SUITE level)**:
```yaml
services:
  sm-kms-app-sqlite-1:
    ports:
      - "28000:8080"   # Override PRODUCT port 18000 → SUITE port 28000
  sm-kms-app-sqlite-2:
    ports:
      - "28001:8080"
  # ... (all 40 service instances across 10 PS-IDs × 4 variants)
```

**Success**:
- SUITE compose ≤ 300 lines (down from 1,504)
- `docker compose config` renders 40+ services correctly
- `go run ./cmd/cicd-lint lint-deployments` passes for cryptoutil

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 5: Validator + Linter Updates (4h) [Status: ☑ COMPLETE]

**Objective**: Update `lint-deployments` and `lint-ports` to understand recursive include
structure. Remove Dockerfile requirement from PRODUCT level (Q3=D confirmed). These validators
currently check compose files directly; they need to traverse includes to validate the full
resolved configuration.

**Tasks**:
1. Update `validate_structure.go`: Remove Dockerfile requirement from `DeploymentTypeProduct`
   (Q3=D confirmed — product deployments use PS-ID Dockerfiles transitively)
2. Update `validate_ports.go`: Validate port ranges by resolving include hierarchy, not just
   the top-level compose file (PRODUCT/SUITE compose files now contain only overrides)
3. Update `validate_compose.go`: Recognize `include:` blocks as valid structure (don't error
   on service sections with only port overrides and no `image:`)
4. Update `validate_secrets.go`: Trail secret file paths through include hierarchy to verify
   product and suite-scoped secrets exist
5. Add test: `validate_ports_test.go` covers recursive include port validation
6. Add test: `validate_structure_test.go` covers PRODUCT compose without Dockerfile (if Q3=defer)
7. Run `go test ./internal/apps/tools/cicd_lint/lint_deployments/... -cover`

**Success**:
- All lint-deployments tests pass: `go test ./internal/apps/tools/cicd_lint/...`
- Coverage ≥ 98% for lint_deployments package (infrastructure/utility threshold)
- `go run ./cmd/cicd-lint lint-deployments` produces no errors on modified deployments/
- No new `//nolint:` suppressions introduced

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 6: Fitness Linter — `usage_health_path_completeness` (2h) [Status: ☑ COMPLETE]

**Objective**: Implement Carryover Item 3. All `{ps-id}_usage.go` files must mention both
`/service/api/v1/health` and `/browser/api/v1/health` paths. Enforce via lint-fitness.

**Tasks**:
1. Create `internal/apps/tools/cicd_lint/lint_fitness/usage_health_path_completeness/lint.go`
2. Create `usage_health_path_completeness/lint_test.go` (table-driven, t.Parallel, UUIDv7)
3. Register in fitness registry (`internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`)
4. Scan all `internal/apps/{PS-ID}/` for `*_usage.go` files
5. For each found `*_usage.go`, verify both `/service/api/v1/health` and `/browser/api/v1/health`
   appear as string literals or const references
6. Run `go test ./internal/apps/tools/cicd_lint/lint_fitness/usage_health_path_completeness/...`

**Success**:
- Fitness linter runs as part of `go run ./cmd/cicd-lint lint-fitness`
- Coverage ≥ 98% for new linter package
- Zero violations on current codebase (all usage files already have both paths — verify first)
- If violations found, fix usage files before enabling linter

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 7: Documentation + ENG-HANDBOOK.md Updates (2h) [Status: ☑ COMPLETE]

**Objective**: Update authoritative documentation to reflect the new recursive include
architecture and the resolved port strategy from quizme Q2.

**Tasks**:
1. Update ENG-HANDBOOK.md Section 3.4 (Port Assignments & Networking): Document that postgres
   uses a single shared leader/follower pair with no host port exposure at any tier (Q1=C, Q2=E).
   Remove per-PS-ID postgres port table (54320-54329) since those services no longer exist.
   Document `docker exec postgres-leader psql` as the developer access method.
2. Update ENG-HANDBOOK.md Section 12 (Deployment Architecture): Document recursive include
   hierarchy, Approach C override pattern, shared-postgres inclusion at all tiers
3. Update ENG-HANDBOOK.md Section 3.4.1 (Port Design Principles): Document that app ports follow
   +10000/+20000 offset; postgres uses shared infrastructure with no host port exposure
4. Update `deployments/{PS-ID}/compose.yml` header comments to document new usage:
   - Standalone (SERVICE): `docker compose up`
   - As include target: `include: - path: ../ps-id/compose.yml` with service overrides
5. Update `lint-deployments` config-overlay-templates.yaml if templates changed
6. Run `go run ./cmd/cicd-lint lint-docs` to verify propagation integrity

**Success**:
- `go run ./cmd/cicd-lint lint-docs` passes (no drift errors)
- Section 3.4 documents shared-postgres architecture (no per-PS-ID postgres ports)
- Section 12 explains recursive include pattern with examples

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 8: E2E Validation (2h) [Status: ☑ COMPLETE]

**Objective**: Confirm all 3 deployment tiers start correctly with Docker Compose after the
recursive include refactoring.

**Tasks** (requires Docker Desktop running):
1. SERVICE level: `docker compose -f deployments/sm-im/compose.yml config` (dry-run)
2. SERVICE level: `docker compose -f deployments/sm-im/compose.yml up --profile dev -d`
   - Verify: sm-im-app-sqlite-1 at :8100 passes healthcheck
   - Verify: shared telemetry starts
   - Tear down: `docker compose down -v`
3. PRODUCT level: `docker compose -f deployments/sm/compose.yml config`
4. PRODUCT level: `docker compose -f deployments/sm/compose.yml up --profile dev -d`
   - Verify: sm-kms at :18000 and sm-im at :18100 both pass healthcheck
   - Tear down: `docker compose down -v`
5. SUITE level: `docker compose -f deployments/cryptoutil/compose.yml config`
6. Run `go run ./cmd/cicd-lint lint-deployments` — MUST report zero errors

**Success**:
- All 3 tiers start and pass health checks
- `lint-deployments` reports zero errors across all deployment directories
- Line count reduction achieved: PRODUCT ≤ 150 lines each, SUITE ≤ 300 lines

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't, root causes, patterns.

---

### Phase 9: Knowledge Propagation (1h) [Status: ☑ COMPLETE]

**Objective**: Apply lessons learned to permanent artifacts — NEVER skip this phase.

**Tasks**:
1. Review lessons.md from all prior phases
2. Update ENG-HANDBOOK.md with new patterns and architectural decisions discovered
3. Update agents/skills/instructions where warranted by lessons
4. Update `configs/{PS-ID}/` header comments if postgres port strategy changed
5. Verify propagation: `go run ./cmd/cicd-lint lint-docs validate-propagation`
6. Commit all artifact updates with semantic commits

**Success**: All artifact updates committed; propagation check passes; lessons applied.

---

## Executive Decisions

### Decision 1: Docker Compose Port Override Approach

**Options**:
- A: Environment variable substitution (`${PORT:-default}`) in PS-ID compose files
- B: Multiple `-f` flags at runtime (`docker compose -f base.yml -f override.yml up`)
- C: Inline service redefinition in the including compose file ✓ **SELECTED**
- D: Separate port-map `.env` files per deployment tier

**Decision**: Option C — inline service redefinition.

**Rationale**: Cleanest approach. No env files, no multi-file flags. Docker Compose v2.24+
supports this natively. Conforms to "no environment variables" policy in ENG-HANDBOOK.md.
Including compose file redefines only `ports:` for each service; all other service attributes
(image, command, volumes, healthcheck, resource limits) are inherited from the PS-ID definition.

### Decision 2: Service Naming Convention

**Options**:
- A: Use `postgres` (abbreviated) everywhere
- B: Use `postgresql` (full word) everywhere ✓ **SELECTED**
- C: Leave mixed (per current state)

**Decision**: Option B — `postgresql` everywhere.

**Rationale**: Config files are already named `{PS-ID}-app-postgresql-{N}.yml`. Service names
in compose files must match to avoid confusion. PS-ID compose files already use `postgresql`.
PRODUCT and SUITE (currently using `postgres`) must be updated to match.

### Decision 3: Builder Service Scope

**Options**:
- A: One `builder-{PS-ID}` per PS-ID, inherits to PRODUCT/SUITE via includes
- B: PRODUCT/SUITE define their own `builder-{PRODUCT}` replacing included PS-ID builders ✓ **SELECTED**
- C: No builder services (pre-build outside compose)

**Decision**: Option B — PS-ID builders at SERVICE level; PRODUCT/SUITE define own builders.

**Rationale**: Each PS-ID continues to build its own image at SERVICE level. At PRODUCT/SUITE,
a single `builder-{PRODUCT}` or `builder-cryptoutil` builds the shared image once for the tier.
Prevents duplicate build steps when multiple PS-IDs are included.

### Decision 4: PostgreSQL Import Strategy (Q1=C)

**Options**:
- A: Full shared-postgres include at all tiers with standalone profile for per-PS-ID postgres
- B: shared-postgres at PRODUCT/SUITE only; PS-ID keeps dedicated postgres
- C: shared-postgres at ALL tiers; per-PS-ID postgres REMOVED entirely; no host port exposure ✓ **SELECTED**
- D: shared-postgres at PRODUCT/SUITE; PS-ID uses single-node-postgres

**Decision**: Option C — shared-postgres included at all tiers; per-PS-ID postgres services
removed entirely from PS-ID compose files.

**Rationale**: Simplest architecture. One postgres infrastructure for everything. Developers
access postgres via `docker exec postgres-leader psql` (no host port needed). Eliminates 10
per-PS-ID postgres service definitions and associated volumes/healthchecks. Consistent with
the "single source of truth" principle — postgres config lives only in shared-postgres.

### Decision 5: PostgreSQL Architecture (Q2=E)

**Options**:
- A: Compact port ranges (54320-54329 SERVICE, 54420-54424 PRODUCT, 54530-54531 SUITE)
- B: No host port exposure above SERVICE level
- C: Reuse ports at all tiers (one tier at a time on developer machine)
- D: Dynamic host ports (Docker assigns ephemeral)
- E: Single PostgreSQL leader/follower pair for all tiers; per-PS-ID logical database isolation ✓ **SELECTED**

**Decision**: Option E — single shared PostgreSQL leader/follower pair.

**Rationale**: Each PS-ID connects to postgres-leader with a different username, password, and
logical database name. All connect to the same address (container port 5432). No host port
exposure (consistent with Q1=C). The PostgreSQL follower replicates all logical leader databases
to separate schemas in a single logical follower database. This eliminates port allocation
complexity entirely — no per-PS-ID, per-PRODUCT, or per-SUITE postgres ports needed.

**Note**: The existing `deployments/shared-postgres/compose.yml` already implements this
architecture with 30 logical databases (10 PS-IDs × 3 tiers), init scripts
(`init-leader-databases.sql`, `init-follower-databases.sql`), and replication setup
(`setup-logical-replication.sh`). The only change needed is removing host port exposure
(currently `5432:5432` on leader and `5433:5432` on follower) per Q1=C.

### Decision 6: Product-Level Dockerfiles (Q3=D)

**Options**:
- A: Defer/supersede — remove Dockerfile requirement from validator
- B: Create lightweight product Dockerfiles with OCI labels only
- C: Create product binary Dockerfiles (multi-service binary scope expansion)
- D: Mark as permanently cancelled — PS-ID Dockerfiles used transitively ✓ **SELECTED**

**Decision**: Option D — permanently cancelled.

**Rationale**: Recursive-include architecture (framework-v8) renders product Dockerfiles
unnecessary. PRODUCT compose files include PS-ID compose files, which reference PS-ID
Dockerfiles. Each PS-ID image is built by its own `builder-{PS-ID}` service. No separate
PRODUCT-level build is needed. `validate_structure.go` updated to remove Dockerfile requirement
from `DeploymentTypeProduct`. Carryover Item 2 marked CANCELLED.

---

## Resolved Decisions (from Quizme v1)

All three design decisions have been resolved. Answers merged from `quizme-v1.md`.

| Decision | Quizme | Answer | Affects Phases |
|----------|--------|--------|---------------|
| PostgreSQL import at SERVICE level | Q1 | **C** — shared-postgres at all tiers; per-PS-ID postgres removed; no host ports | 2, 3, 4 |
| PostgreSQL architecture for all tiers | Q2 | **E** — single leader/follower pair; per-PS-ID logical DB isolation; no host ports | 2, 3, 4, 7 |
| Product-level Dockerfiles (Carryover Item 2) | Q3 | **D** — permanently cancelled; PS-ID Dockerfiles used transitively | 5 |

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Docker Compose include deduplication bug (same file via multiple paths) | Medium | High | Phase 0 research validates this before full implementation |
| `profiles:` in included files not honored by parent `--profile` | Medium | High | Phase 0 research validates profile inheritance |
| Approach C port override fails (arrays not merged as expected) | Low | High | Phase 0 minimal test validates array replacement |
| lint-deployments false-positive errors on service-override-only sections | High | Medium | Phase 5 validator updates resolve before full deployment |
| Secret path resolution breaks across include boundary | Medium | High | Phase 0 validates that included compose secret paths resolve correctly |

---

## Quality Gates — MANDATORY

**Per-Phase**:
- ✅ `go run ./cmd/cicd-lint lint-deployments` — zero new errors (may reduce existing ones)
- ✅ `go build ./...` — clean build
- ✅ `golangci-lint run` / `golangci-lint run --build-tags e2e,integration` — zero violations
- ✅ `go test ./internal/apps/tools/cicd_lint/...` — 100% passing
- ✅ Coverage ≥ 98% for validator packages (infrastructure threshold)

**E2E (Phase 8)**:
- ✅ All 3 tiers pass `docker compose config` (valid rendered output)
- ✅ SERVICE and PRODUCT tiers start and pass health checks
- ✅ `lint-deployments` reports zero errors

---

## Success Criteria

- [x] All 9 phases complete with evidence
- [x] Line count: PRODUCT compose files ≤ 150 lines each, SUITE ≤ 300 lines
- [x] Zero copy-paste service definitions across tiers (service defs only in PS-ID compose)
- [x] All 10 PS-IDs have 4 variants (sqlite-1, sqlite-2, postgresql-1, postgresql-2) at all 3 tiers
- [x] All naming standardized to `postgresql` throughout
- [x] SM product includes both sm-kms and sm-im services
- [x] shared-postgres and shared-telemetry imported at all tiers (no per-PS-ID postgres services)
- [x] `usage_health_path_completeness` fitness linter active and passing
- [x] ENG-HANDBOOK.md documents shared-postgres architecture (no per-PS-ID postgres ports)
- [x] `go run ./cmd/cicd-lint lint-deployments` clean end-to-end

---

## ENG-HANDBOOK.md Cross-References

| Topic | Section | Applies |
|-------|---------|---------|
| Deployment Architecture | [Section 12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) | ALL phases |
| Port Assignments | [Section 3.4](../../docs/ENG-HANDBOOK.md#34-port-assignments--networking) | Phases 2–4, 7 |
| Port Design Principles | [Section 3.4.1](../../docs/ENG-HANDBOOK.md#341-port-design-principles) | Phases 2–4, 7 |
| Secrets Management | [Section 12.3.3](../../docs/ENG-HANDBOOK.md#1233-secrets-coordination-strategy) | Phases 2–4 |
| CICD Command Architecture | [Section 9.10](../../docs/ENG-HANDBOOK.md#910-cicd-command-architecture) | Phase 5 |
| Deployment Validators | [Section 13.1.11](../../docs/ENG-HANDBOOK.md#13111-validation-pipeline-architecture) | Phase 5 |
| Config File Architecture | [Section 13.2](../../docs/ENG-HANDBOOK.md#132-config-file-architecture) | Phases 1, 7 |
| Fitness Functions | [Section 9.11](../../docs/ENG-HANDBOOK.md#911-architecture-fitness-functions) | Phase 6 |
| Testing Strategy | [Section 10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | Phase 5 |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL phases |
| Knowledge Propagation | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | Phase 9 |
| Plan Lifecycle | [Section 14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | ALL phases |
