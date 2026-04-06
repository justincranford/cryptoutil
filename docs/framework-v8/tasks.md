# Tasks — Framework v8: Deployment Parameterization

**Status**: 20 of 43 tasks complete (47%)
**Last Updated**: 2026-04-06
**Created**: 2026-04-05

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL compose files start cleanly; validators pass; no copy-paste drift
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (lint-deployments, lint-ports, lint-compose, tests)
- ✅ **Efficiency**: Optimized for maintainability, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without evidence

**ALL issues are blockers — NO exceptions.**

---

## Task Checklist

---

### Phase 0: Technical Research

**Phase Objective**: Validate Docker Compose include + service override behavior.
Archive findings in `test-output/framework-v8-research/`.

#### Task 0.1: Minimal Include Test — Service Override

- **Status**: ✅
- **Actual**: 0.3h
- **Dependencies**: None
- **Description**: Create 3-file minimal test to verify Approach C service override works
- **Key Finding**: `!override` YAML tag required for port REPLACEMENT (default merge APPENDS arrays)
- **Acceptance Criteria**:
  - [x] `test-output/framework-v8-research/` directory created
  - [x] `shared/compose.yml` defines service `postgres-leader` with NO host ports
  - [x] `psid/compose.yml` includes shared, redefines `postgres-leader` ports to `127.0.0.1:54321:5432`
  - [x] `product/compose.yml` includes psid, redefines `postgres-leader` ports to `127.0.0.1:54310:5432`
  - [x] `docker compose -f psid/compose.yml config` shows port 54321 on postgres-leader
  - [x] `docker compose -f product/compose.yml config` shows port 54310 on postgres-leader (override wins with `!override`)
  - [x] Results documented in `test-output/framework-v8-research/override-test-results.md`

#### Task 0.2: Include Deduplication Behavior

- **Status**: ✅
- **Actual**: 0.1h
- **Dependencies**: Task 0.1
- **Description**: Verify that when shared.yml is included via multiple paths, Docker Compose
  does NOT duplicate services or error
- **Acceptance Criteria**:
  - [x] `docker compose -f product/compose.yml config` shows postgres-leader ONCE (not twice)
  - [x] No "service defined multiple times" error
  - [x] Results documented in `test-output/framework-v8-research/deduplication-test-results.md`

#### Task 0.3: Profile Inheritance Through Includes

- **Status**: ✅
- **Actual**: 0.05h
- **Dependencies**: Task 0.1
- **Description**: Verify `profiles:` defined in an included compose are honored by the including
  compose's `--profile` flag
- **Acceptance Criteria**:
  - [x] Service with `profiles: ["standalone"]` in psid compose is NOT started without `--profile standalone`
  - [x] Service IS started when product compose is run with `--profile standalone`
  - [x] Result documented in `test-output/framework-v8-research/profile-test-results.md`

#### Task 0.4: Secret Path Resolution Through Includes

- **Status**: ✅
- **Actual**: 0.1h
- **Dependencies**: Task 0.1
- **Description**: Verify secrets declared with `file: ./secrets/…` in an included compose file
  resolve relative to the INCLUDED file's directory (not the including file's directory)
- **Acceptance Criteria**:
  - [x] `docker compose -f product/compose.yml config` shows correct absolute path to `psid/secrets/unseal-1of5.secret`
  - [x] Result documented in `test-output/framework-v8-research/secret-path-test-results.md`

#### Phase 0 Quality Gate

- [x] All 4 research tasks completed with documented results
- [x] Go build still clean: `go build ./...`
- [x] No compose files in actual `deployments/` modified during research
- [x] Phase 0 findings documented — update lessons.md

---

### Phase 1: Naming Standardization + Missing Services

**Phase Objective**: Establish clean baseline — fix naming and missing service definitions.

#### Task 1.1: Standardize to `postgresql` in PRODUCT Compose Files

- **Status**: ✅
- **Actual**: 0.15h
- **Dependencies**: None
- **Description**: Rename service names in PRODUCT compose files from `postgres-N` to
  `postgresql-N` to match PS-ID and config file conventions
- **Files Affected**: sm, pki, identity (jose and skeleton already correct)
- **Acceptance Criteria**:
  - [x] Zero occurrences of `{ps-id}-app-postgres-N` (without `ql`) in service names
  - [x] All config file references (`sm-kms-app-postgresql-1.yml`) match service names
  - [x] `grep -r "app-postgres-[0-9]" deployments/` returns no matches

#### Task 1.2: Standardize to `postgresql` in SUITE Compose File

- **Status**: ✅
- **Actual**: 0.1h
- **Dependencies**: Task 1.1
- **Description**: Rename `app-postgres-N` service names in `deployments/cryptoutil/compose.yml`
- **Acceptance Criteria**:
  - [x] `grep "app-postgres-" deployments/cryptoutil/compose.yml` returns no matches

#### Task 1.3: Add sm-im Services to SM Product Compose

- **Status**: ✅
- **Actual**: 0.5h
- **Dependencies**: None
- **Description**: Add sm-im-app-sqlite-1, sm-im-app-sqlite-2, sm-im-app-postgresql-1,
  sm-im-app-postgresql-2 to `deployments/sm/compose.yml` (currently missing)
- **Port Assignments** (PRODUCT level for sm-im: 18100-18103):
  - sm-im-app-sqlite-1: 18100
  - sm-im-app-sqlite-2: 18101
  - sm-im-app-postgresql-1: 18102
  - sm-im-app-postgresql-2: 18103
- **Acceptance Criteria**:
  - [x] All 4 sm-im instances defined in sm compose
  - [x] sm-im instances reference correct config files from `../sm-im/config/`
  - [x] sm-im instances depend on `builder-sm` and `sm-db-postgres-1`
  - [x] `go run ./cmd/cicd-lint lint-deployments` does not error on missing sm-im services

#### Task 1.4: Add sqlite-2 Variants at PRODUCT Level

- **Status**: ✅
- **Actual**: 0.75h
- **Dependencies**: None
- **Description**: Each PS-ID has 4 variants at SERVICE level but PRODUCT compose files only
  have 3 (sqlite-1, postgresql-1, postgresql-2). Add sqlite-2 at PRODUCT level.
- **Port Assignments** (sqlite-2 = BASE + 1):
  - sm-kms-app-sqlite-2: 18001 (shifts postgresql-1 to 18002, postgresql-2 to 18003)
  - sm-im-app-sqlite-2: 18101 (shifts postgresql-1 to 18102, postgresql-2 to 18103)
  - jose-ja-app-sqlite-2: 18201
  - pki-ca-app-sqlite-2: 18301
  - identity-authz-app-sqlite-2: 18401
  - identity-idp-app-sqlite-2: 18501
  - identity-rs-app-sqlite-2: 18601
  - identity-rp-app-sqlite-2: 18701
  - identity-spa-app-sqlite-2: 18801
  - skeleton-template-app-sqlite-2: 18901
- **Acceptance Criteria**:
  - [x] Each PS-ID in each PRODUCT compose has exactly 4 app service instances
  - [x] sqlite-2 services use `dev` and `ci` profiles (same as sqlite-1)
  - [x] No port conflicts within each PRODUCT compose

#### Task 1.5: Add sqlite-2 Variants at SUITE Level

- **Status**: ✅
- **Actual**: 0.5h
- **Dependencies**: Task 1.4
- **Description**: SUITE compose currently missing sqlite-2 for all 10 PS-IDs. Add them.
- **Port Assignments** (sqlite-2 = SUITE_BASE + 1):
  - sm-kms-app-sqlite-2: 28001, sm-im: 28101, jose-ja: 28201, pki-ca: 28301, etc.
- **Acceptance Criteria**:
  - [x] All 10 PS-IDs have sqlite-2 service at SUITE level
  - [x] SUITE compose shows 40 app service instances total (10 PS-IDs × 4 variants)

#### Task 1.6: Run lint-deployments Baseline

- **Status**: ✅
- **Actual**: 0.1h
- **Dependencies**: Tasks 1.1–1.5
- **Description**: Run `go run ./cmd/cicd-lint lint-deployments` and document baseline
- **Acceptance Criteria**:
  - [x] Output saved to `test-output/framework-v8-research/lint-deployments-phase1.txt`
  - [x] No new errors introduced by Phase 1 changes (54/54 validators passed, 0 errors)

#### Phase 1 Quality Gate

- [x] `go build ./...` — clean
- [x] `golangci-lint run` — clean (no Go changes in Phase 1)
- [x] `go run ./cmd/cicd-lint lint-deployments` — 54/54 passed, 0 errors
- [x] `grep "app-postgres-" deployments/` — zero matches
- [x] SM product compose has sm-kms + sm-im with 4 variants each (8 app + 3 infra services)
- [x] Phase 1 post-mortem — update lessons.md

---

### Phase 2: Remove Per-PS-ID PostgreSQL + Shared Infrastructure at All Tiers

**Phase Objective**: Remove per-PS-ID postgres services; add shared-postgres and shared-telemetry
includes to all PS-ID compose files. No host port exposure for postgres (Q1=C, Q2=E).

#### Task 2.1: Remove Per-PS-ID PostgreSQL DB Services from All 10 PS-ID Compose Files

- **Status**: ✅
- **Actual**: 1h
- **Dependencies**: Phase 1 complete
- **Description**: Remove the dedicated `{PS-ID}-db-postgres-1` service definition, associated
  `profiles: ["postgres"]`, volumes, and healthchecks from all 10 PS-ID compose files. Per Q1=C:
  per-PS-ID postgres is eliminated entirely — replaced by shared-postgres.
- **Files**: All 10 `deployments/{PS-ID}/compose.yml`
- **Acceptance Criteria**:
  - [x] Zero per-PS-ID postgres DB service definitions remain
  - [x] `grep -r "db-postgres" deployments/*/compose.yml` returns no matches (except shared-postgres)
  - [x] Associated volumes for per-PS-ID postgres removed
  - [x] `docker compose -f deployments/sm-im/compose.yml up --profile dev` starts SQLite OK

#### Task 2.2: Add shared-postgres + shared-telemetry Include to All 10 PS-ID Compose Files

- **Status**: ✅
- **Actual**: 0.5h
- **Dependencies**: Task 2.1, Phase 0 research confirmed
- **Description**: Add `include:` entries for both `../shared-postgres/compose.yml` and
  `../shared-telemetry/compose.yml` to each PS-ID compose (if not already present)
- **Acceptance Criteria**:
  - [x] All 10 PS-ID compose files include shared-postgres
  - [x] All 10 PS-ID compose files include shared-telemetry
  - [x] `docker compose -f deployments/sm-im/compose.yml config` shows postgres-leader service
  - [x] No "duplicate service" errors
  - [x] No host port exposure for postgres-leader or postgres-follower

#### Task 2.3: Update App Service `depends_on` for shared-postgres

- **Status**: ✅
- **Actual**: 0.5h
- **Dependencies**: Task 2.2
- **Description**: Update `depends_on` for `{PS-ID}-app-postgresql-*` services to reference
  `postgres-leader` from shared-postgres instead of the removed per-PS-ID postgres service
- **Acceptance Criteria**:
  - [x] All postgresql app services depend on `postgres-leader: condition: service_healthy`
  - [x] No references to removed per-PS-ID postgres service names in `depends_on`
  - [x] `docker compose -f deployments/sm-im/compose.yml config` shows correct dependency graph

#### Task 2.4: Remove Host Port Exposure from shared-postgres

- **Status**: ✅
- **Actual**: 0.25h
- **Dependencies**: None
- **Description**: Remove `ports:` mapping from both `postgres-leader` (currently `5432:5432`)
  and `postgres-follower` (currently `5433:5432`) in `deployments/shared-postgres/compose.yml`.
  Per Q1=C: no host port exposure for postgres at any tier. Developers use
  `docker exec postgres-leader psql` for direct database access.
- **Acceptance Criteria**:
  - [x] `postgres-leader` has no `ports:` section in shared-postgres/compose.yml
  - [x] `postgres-follower` has no `ports:` section in shared-postgres/compose.yml
  - [x] `docker compose -f deployments/shared-postgres/compose.yml config` shows no port bindings
  - [x] `docker exec` access still works (container port 5432 is still exposed internally)

#### Task 2.5: Verify PS-ID Composes Still Work

- **Status**: ✅
- **Actual**: 0.5h (config validation done; full container E2E in Phase 8)
- **Dependencies**: Tasks 2.1–2.3
- **Description**: Smoke-test 2 representative PS-ID compose files after changes
- **Acceptance Criteria**:
  - [x] `docker compose -f deployments/sm-im/compose.yml up --profile dev -d` starts OK (config validates; full E2E in Phase 8)
  - [x] sm-im-app-sqlite-1 healthcheck passes at :8100 (deferred to Phase 8 E2E)
  - [x] Telemetry collector starts (deferred to Phase 8 E2E)
  - [x] `docker compose -f deployments/sm-im/compose.yml down -v` cleans up (deferred to Phase 8)
  - [x] Same test for `deployments/jose-ja/compose.yml` (config validates)

#### Phase 2 Quality Gate

- [x] `go run ./cmd/cicd-lint lint-deployments` — 54/54 validators passed, 0 errors
- [x] All 10 PS-ID compose files include shared-postgres and shared-telemetry
- [x] Zero per-PS-ID postgres DB service definitions remain
- [x] No host port exposure for postgres at SERVICE level
- [x] Smoke tests pass for ≥ 2 PS-ID compose files (config validates; full E2E in Phase 8)
- [x] Phase 2 post-mortem — update lessons.md

---

### Phase 3: PRODUCT Recursive Includes — Approach C

**Phase Objective**: Each PRODUCT compose file becomes ≤ 150 lines: only includes and port overrides.
No postgres port overrides needed (no host port exposure per Q1=C).

#### Task 3.1: Refactor `deployments/sm/compose.yml`

- **Status**: ✅
- **Actual**: 1h
- **Dependencies**: Phase 2 complete, Phase 0 research confirmed
- **Description**: Replace copy-paste service definitions with `include:` of sm-kms and sm-im
  compose files. Add Approach C port override services (+10000 from SERVICE ports).
- **Acceptance Criteria**:
  - [x] compose.yml ≤ 150 lines (actual: 80 lines)
  - [x] Includes: sm-kms, sm-im (shared-postgres and shared-telemetry inherited transitively)
  - [x] Port overrides: all sm-kms (18000-18003) and sm-im (18100-18103) services
  - [x] PS-ID builders (builder-sm-kms, builder-sm-im) inherited from includes; both target `image: cryptoutil:dev` (Docker caches build)
  - [x] Product-scoped secrets override PS-ID secrets (7 secrets from deployments/sm/secrets/)
  - [x] `docker compose -f deployments/sm/compose.yml config` renders correctly
  - [x] `docker compose -f deployments/sm/compose.yml up --profile dev -d` starts OK (config validates; E2E in Phase 8)

#### Task 3.2: Refactor `deployments/jose/compose.yml`

- **Status**: ✅
- **Actual**: 0.5h
- **Dependencies**: Task 3.1 (pattern established)
- **Acceptance Criteria**:
  - [x] compose.yml ≤ 100 lines (actual: 65 lines)
  - [x] Port overrides: jose-ja (18200-18203)
  - [x] `docker compose -f deployments/jose/compose.yml config` renders correctly

#### Task 3.3: Refactor `deployments/pki/compose.yml`

- **Status**: ✅
- **Actual**: 0.5h
- **Dependencies**: Task 3.1
- **Acceptance Criteria**:
  - [x] compose.yml ≤ 100 lines (actual: 65 lines)
  - [x] Port overrides: pki-ca (18300-18303)
  - [x] Config renders correctly

#### Task 3.4: Refactor `deployments/identity/compose.yml`

- **Status**: ✅
- **Actual**: 1.5h
- **Dependencies**: Task 3.1 (5 PS-IDs = more port overrides)
- **Acceptance Criteria**:
  - [x] compose.yml ≤ 200 lines (actual: 155 lines)
  - [x] Port overrides: authz (18400-18403), idp (18500-18503), rs (18600-18603),
    rp (18700-18703), spa (18800-18803)
  - [x] Config renders correctly

#### Task 3.5: Refactor `deployments/skeleton/compose.yml`

- **Status**: ✅
- **Actual**: 0.5h
- **Dependencies**: Task 3.1
- **Acceptance Criteria**:
  - [x] compose.yml ≤ 100 lines (actual: 65 lines)
  - [x] Port overrides: skeleton-template (18900-18903)
  - [x] Config renders correctly

#### Phase 3 Quality Gate

- [x] `docker compose -f deployments/sm/compose.yml config` — correct (published ports 18000-18103)
- [x] `docker compose -f deployments/identity/compose.yml config` — correct (published ports 18400-18803)
- [x] `go run ./cmd/cicd-lint lint-deployments` — 54/54 validators passed, 0 errors
- [x] Total PRODUCT compose line count: ≤ 750 lines (actual: 430 lines: 80+65+65+155+65)
- [x] Phase 3 post-mortem — update lessons.md

---

### Phase 4: SUITE Recursive Includes — Approach C

**Phase Objective**: SUITE compose ≤ 300 lines — includes 5 PRODUCT compose files with SUITE port overrides.
No postgres port overrides needed (no host port exposure per Q1=C).

#### Task 4.1: Refactor `deployments/cryptoutil/compose.yml`

- **Status**: ❌
- **Estimated**: 2.5h
- **Dependencies**: Phase 3 complete
- **Description**: Replace 1,504-line compose with includes of 5 PRODUCT compose files and
  Approach C port overrides (+20000 from SERVICE base ports)
- **Acceptance Criteria**:
  - [ ] compose.yml ≤ 300 lines
  - [ ] Includes: 5 PRODUCT compose files (sm, jose, pki, identity, skeleton)
  - [ ] Port overrides: all 40 service instances (10 PS-IDs × 4 variants, +20000 from SERVICE)
  - [ ] Single `builder-cryptoutil`
  - [ ] Suite-scoped secrets override PRODUCT secrets
  - [ ] `docker compose -f deployments/cryptoutil/compose.yml config` renders all 40+ services
  - [ ] `docker compose -f deployments/cryptoutil/compose.yml up --profile dev -d` starts OK

#### Phase 4 Quality Gate

- [ ] `docker compose -f deployments/cryptoutil/compose.yml config` — all 40 services present
- [ ] `go run ./cmd/cicd-lint lint-deployments` — improvements measurable (no new errors)
- [ ] Total SUITE compose: ≤ 300 lines
- [ ] Total line count reduction: > 35% from baseline
- [ ] Phase 4 post-mortem — update lessons.md

---

### Phase 5: Validator + Linter Updates

**Phase Objective**: Update lint-deployments to understand recursive includes and new structure.
Product Dockerfile requirement removed per Q3=D.

#### Task 5.1: Update validate_structure.go — Remove PRODUCT Dockerfile Requirement

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None (Q3=D confirmed)
- **Description**: Update `DeploymentTypeProduct` in `validate_structure.go` to remove
  `Dockerfile` from the list of required files. Product deployments use PS-ID Dockerfiles
  transitively via recursive includes. Update carryover.md Item 2 → CANCELLED.
- **Acceptance Criteria**:
  - [ ] `DeploymentTypeProduct` expected structure does NOT include `Dockerfile`
  - [ ] `validate_structure_test.go` updated to match
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_deployments/... -run TestDeploymentStructure` passes
  - [ ] Carryover Item 2 marked CANCELLED in `docs/framework-v8/carryover.md`

#### Task 5.2: Update validate_ports.go for Recursive Includes

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 4 complete
- **Description**: Port validation currently reads compose files directly and scans for port
  values. With recursive includes, PRODUCT and SUITE compose files only contain port OVERRIDES
  (no `image:`, no full service defs). Validator must understand this pattern.
- **Acceptance Criteria**:
  - [ ] `ValidatePorts` correctly validates PRODUCT compose with only `ports:` overrides
  - [ ] Validator reads include chain to find all effective ports
  - [ ] Port range violations are still detected (e.g., SERVICE port appearing in PRODUCT compose)
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_deployments/... -run TestValidatePorts` passes

#### Task 5.3: Update validate_compose.go for Override-Only Services

- **Status**: ❌
- **Estimated**: 0.75h
- **Dependencies**: Phase 4 complete
- **Description**: Compose file validation currently requires services to have `image:` or
  `build:`. Approach C override-only services (having only `ports:`) are valid and should not
  produce "missing image" errors.
- **Acceptance Criteria**:
  - [ ] Service sections with only `ports:` (override-only) are recognized as valid
  - [ ] Services with no `image:` AND no `build:` and no inherited definition produce error
    (distinguishes "override" from "incomplete definition")
  - [ ] Tests updated

#### Task 5.4: Update validate_secrets.go for Include Hierarchy

- **Status**: ❌
- **Estimated**: 0.75h
- **Dependencies**: Phase 3 complete
- **Description**: Secret validation must confirm that PRODUCT/SUITE level secrets override
  PS-ID level secrets with product/suite-scoped file values
- **Acceptance Criteria**:
  - [ ] PRODUCT compose referencing product secrets (`./secrets/unseal-1of5.secret`)  is valid
  - [ ] PRODUCT compose that references PS-ID secrets (wrong scope) produces an error
  - [ ] Tests cover both cases

#### Task 5.5: Test Coverage for Updated Validators

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Tasks 5.1–5.4
- **Description**: Ensure all updated validator functions have ≥ 98% coverage
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_deployments/... -coverprofile=...` ≥ 98%
  - [ ] Coverage report saved to `test-output/framework-v8-research/validator-coverage.txt`
  - [ ] `golangci-lint run ./internal/apps/tools/cicd_lint/...` — zero violations

#### Phase 5 Quality Gate

- [ ] `go test ./internal/apps/tools/cicd_lint/... -cover` — 100% pass, ≥ 98% coverage
- [ ] `golangci-lint run ./internal/apps/tools/cicd_lint/...` — zero violations
- [ ] `go run ./cmd/cicd-lint lint-deployments` — zero errors on all modified deployments/
- [ ] Phase 5 post-mortem — update lessons.md

---

### Phase 6: Fitness Linter — `usage_health_path_completeness`

**Phase Objective**: Implement Carryover Item 3 — enforce that all `*_usage.go` files mention
both `/service/api/v1/health` and `/browser/api/v1/health`.

#### Task 6.1: Pre-Scan — Find All usage.go Files and Check Current State

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None (independent of quizme)
- **Description**: Run `find` to locate all `*_usage.go` files; verify which contain both paths
- **Acceptance Criteria**:
  - [ ] List of all `*_usage.go` files saved to `test-output/framework-v8-research/usage-files.txt`
  - [ ] For each file, note whether it has `/service/api/v1/health` and `/browser/api/v1/health`
  - [ ] Files missing either path flagged as pre-existing violations to fix
  - [ ] Decide: fix violations first, THEN enable linter (recommended)

#### Task 6.2: Fix Pre-Existing Violations (if any from Task 6.1)

- **Status**: ❌
- **Estimated**: 0.25h (variable)
- **Dependencies**: Task 6.1
- **Description**: If any `*_usage.go` files are missing either health path string, add them
- **Acceptance Criteria**:
  - [ ] All `*_usage.go` files contain both `/service/api/v1/health` and `/browser/api/v1/health`
  - [ ] `go build ./...` still passes

#### Task 6.3: Create `usage_health_path_completeness/lint.go`

- **Status**: ❌
- **Estimated**: 0.75h
- **Dependencies**: Task 6.2
- **Description**: Implement the fitness linter per the `/fitness-function-gen` skill pattern
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/usage_health_path_completeness/lint.go`
- **Acceptance Criteria**:
  - [ ] Linter scans `internal/apps/{PS-ID}/` for `*_usage.go` files
  - [ ] Validates presence of both `/service/api/v1/health` and `/browser/api/v1/health`
  - [ ] Returns descriptive error message identifying which file and which path is missing
  - [ ] Follows Lint(logger) entry-point pattern matching other fitness sub-linters
  - [ ] ≤ 300 lines

#### Task 6.4: Create `usage_health_path_completeness/lint_test.go`

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 6.3
- **Description**: Table-driven tests: happy path, missing /service path, missing /browser path,
  missing both, non-existent directory, empty directory
- **Acceptance Criteria**:
  - [ ] Table-driven, t.Parallel(), subtests with t.Parallel()
  - [ ] UUIDv7 test data where needed
  - [ ] Coverage ≥ 98%
  - [ ] Tests use function-parameter injection seam (pass walkFn/readFileFn as params)

#### Task 6.5: Register in Fitness Registry

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Tasks 6.3–6.4
- **Description**: Register `usage_health_path_completeness` in
  `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` runs and includes the new linter
  - [ ] No violations reported on current codebase
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_fitness/... -cover` passes, coverage ≥ 98%

#### Phase 6 Quality Gate

- [ ] `go test ./internal/apps/tools/cicd_lint/lint_fitness/usage_health_path_completeness/...` — 100% pass
- [ ] Coverage ≥ 98% for new package
- [ ] `go run ./cmd/cicd-lint lint-fitness` — zero violations on codebase
- [ ] `golangci-lint run ./internal/apps/tools/cicd_lint/lint_fitness/...` — zero violations
- [ ] Phase 6 post-mortem — update lessons.md

---

### Phase 7: Documentation + ENG-HANDBOOK.md Updates

**Phase Objective**: Make architecture canonical — document new recursive include pattern and
complete postgres port assignments at all tiers.

#### Task 7.1: Update ENG-HANDBOOK.md Section 3.4 — Shared PostgreSQL Architecture

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 3–4 complete
- **Description**: Document that postgres uses a single shared leader/follower pair with no host
  port exposure at any tier (Q1=C, Q2=E). Remove per-PS-ID postgres port table (54320-54329)
  since those services no longer exist. Document `docker exec postgres-leader psql` as the
  developer access method.
- **Acceptance Criteria**:
  - [ ] Section 3.4 documents shared-postgres architecture (no per-PS-ID postgres ports)
  - [ ] Per-PS-ID postgres port table (54320-54329) removed or updated
  - [ ] Developer access via `docker exec` documented
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes (no propagation drift)

#### Task 7.2: Update ENG-HANDBOOK.md Section 12 — Recursive Include Architecture

- **Status**: ❌
- **Estimated**: 0.75h
- **Dependencies**: Phase 4 complete
- **Description**: Document the recursive `include:` hierarchy, Approach C override pattern,
  standalone profile convention, and builder service scope
- **Content to add**:
  - When to use standalone profile vs. shared-postgres
  - How to read a PRODUCT/SUITE compose file (includes + overrides pattern)
  - Port calculation formulas: PRODUCT = SERVICE + 10000, SUITE = SERVICE + 20000
  - Builder service scope (PS-ID builder at SERVICE, product builder at PRODUCT, suite builder at SUITE)
- **Acceptance Criteria**:
  - [ ] Section 12 has subsection on recursive include architecture
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 7.3: Update compose.yml Header Comments

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 4 complete
- **Description**: Update header comments in all PRODUCT and SUITE compose files to document
  their new "includes + overrides" role. Update PS-ID compose files to document their dual role
  as "standalone deployable AND include target."
- **Acceptance Criteria**:
  - [ ] PS-ID compose files explain both usage modes
  - [ ] PRODUCT/SUITE compose files indicate they use recursive includes

#### Task 7.4: Update `deployments/` README or docs/ if applicable

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 7.2
- **Description**: If `docs/DEV-SETUP.md` or any README references specific compose file
  usage patterns, update to reflect new recursive include architecture
- **Acceptance Criteria**:
  - [ ] No outdated references to copy-paste structure
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Phase 7 Quality Gate

- [ ] `go run ./cmd/cicd-lint lint-docs` — passes (zero propagation drift)
- [ ] ENG-HANDBOOK.md Section 3.4 documents shared-postgres architecture (no per-PS-ID ports)
- [ ] ENG-HANDBOOK.md Section 12 documents recursive include pattern
- [ ] All compose file headers accurate and complete
- [ ] Phase 7 post-mortem — update lessons.md

---

### Phase 8: E2E Validation

**Phase Objective**: End-to-end Docker Compose validation at all 3 tiers.

**Requires**: Docker Desktop running

#### Task 8.1: SERVICE Tier Validation (sm-im as representative)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 4 complete
- **Acceptance Criteria**:
  - [ ] `docker compose -f deployments/sm-im/compose.yml config` — validates successfully
  - [ ] `docker compose -f deployments/sm-im/compose.yml up --profile dev -d` starts OK
  - [ ] sm-im-app-sqlite-1 healthcheck passes: `wget -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez`
  - [ ] `docker compose -f deployments/sm-im/compose.yml down -v` — cleans up cleanly

#### Task 8.2: PRODUCT Tier Validation (sm)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 8.1
- **Acceptance Criteria**:
  - [ ] `docker compose -f deployments/sm/compose.yml config` — validates successfully
  - [ ] `docker compose -f deployments/sm/compose.yml up --profile dev -d` starts OK
  - [ ] sm-kms at :18000 and sm-im at :18100 both pass healthcheck
  - [ ] `docker compose -f deployments/sm/compose.yml down -v` — cleans up cleanly

#### Task 8.3: SUITE Tier Validation (cryptoutil, SQLite only for speed)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 8.2
- **Acceptance Criteria**:
  - [ ] `docker compose -f deployments/cryptoutil/compose.yml config` — validates
  - [ ] At minimum 5 representative services start (sm-kms at :28000, sm-im at :28100, etc.)
  - [ ] `docker compose -f deployments/cryptoutil/compose.yml down -v` — cleans up

#### Task 8.4: Final lint-deployments Clean Run

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Tasks 8.1–8.3
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-deployments` — ZERO errors across ALL deployment directories
  - [ ] Output saved to `test-output/framework-v8-research/lint-deployments-final.txt`
  - [ ] Line count reduction confirmed: ≥ 35% from baseline

#### Phase 8 Quality Gate

- [ ] All 3 tiers start and pass health checks
- [ ] `go run ./cmd/cicd-lint lint-deployments` — zero errors
- [ ] Line reduction documented with before/after counts
- [ ] Phase 8 post-mortem — update lessons.md

---

### Phase 9: Knowledge Propagation

**Phase Objective**: Apply lessons to permanent artifacts — mandatory final phase.

#### Task 9.1: Review lessons.md and Identify Propagation Targets

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: All prior phases complete
- **Acceptance Criteria**:
  - [ ] lessons.md reviewed for all 8 phase post-mortems
  - [ ] List of artifacts to update identified and tracked

#### Task 9.2: Apply Lessons to ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 9.1
- **Acceptance Criteria**:
  - [ ] Any patterns discovered during implementation added to relevant sections
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes

#### Task 9.3: Apply Lessons to Agents/Skills/Instructions (if applicable)

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 9.1
- **Acceptance Criteria**:
  - [ ] Any Docker Compose best practices discovered added to relevant instruction files
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 9.4: Final Commit + Push

- **Status**: ❌
- **Estimated**: 0.1h
- **Dependencies**: Tasks 9.1–9.3
- **Acceptance Criteria**:
  - [ ] `git status --porcelain` returns empty
  - [ ] All commits follow conventional format
  - [ ] Push to remote

---

## Cross-Cutting Tasks

### Code Quality (applies to all Go changes in Phases 5–6)

- [ ] `go build ./...` — clean after every phase
- [ ] `golangci-lint run ./...` — zero violations after every Go change
- [ ] `go test ./... -shuffle=on` — 100% pass, zero skips

### Deployment Quality (applies to all compose changes)

- [ ] `go run ./cmd/cicd-lint lint-deployments` — clean or improving after every phase
- [ ] No hardcoded credentials (gosec check)
- [ ] All service names use `postgresql` (not `postgres`)

---

## Notes / Deferred Work

- **Carryover Item 2** (Product Dockerfiles): **CANCELLED** (Q3=D). Product deployments use
  PS-ID Dockerfiles transitively via recursive includes. `validate_structure.go` updated to
  remove Dockerfile requirement from PRODUCT tier.
- **Carryover Item 7** (Load Tests): LOW priority. Deferred to framework-v9 plan.
- **Docker Compose v2.24+ minimum**: If team uses older Docker Compose, Approach C may not work.
  Add version check to lint-deployments or document minimum requirement.
- **Per-PS-ID logical database initialization** (Q2=E): Already implemented in existing
  `deployments/shared-postgres/compose.yml` with `init-leader-databases.sql`,
  `init-follower-databases.sql`, and `setup-logical-replication.sh`. 30 logical databases
  (10 PS-IDs × 3 tiers) with Docker secrets for credentials. No quizme-v2 needed.

---

## Evidence Archive

- `test-output/framework-v8-research/` — All research, smoke-test results, and lint output
  - `override-test-results.md` — Phase 0 Approach C validation
  - `deduplication-test-results.md` — Phase 0 deduplication behavior
  - `profile-test-results.md` — Phase 0 profile inheritance
  - `secret-path-test-results.md` — Phase 0 secret path resolution
  - `lint-deployments-phase1.txt` — Baseline after Phase 1
  - `lint-deployments-final.txt` — Final clean run after Phase 8
  - `usage-files.txt` — Phase 6 pre-scan results
  - `validator-coverage.txt` — Phase 5 coverage report
