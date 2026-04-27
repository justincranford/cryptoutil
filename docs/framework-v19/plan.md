# Implementation Plan - Framework V19: Prescriptive MANIFEST.yaml + Identity Conformance Migration

**Status**: Planning
**Created**: 2026-04-27
**Last Updated**: 2026-04-27
**Predecessors**: V17 (commit a747ac2ea — 87 linters, partial migration), V18 (ENG-HANDBOOK propagation — unexecuted)
**Purpose**: Two interrelated work streams carried forward from V17 GAPs and new user request:
(1) Expand `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml` to be
fully prescriptive — specifying every required subdirectory, its required files, and purpose —
and extend the `apps_ps_id_template` linter to validate all new MANIFEST fields recursively;
(2) Complete V17 deferred conformance migration for identity services (Tasks 5.2–5.7) plus
sm-im root cleanup.

---

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL code functionally correct; comprehensive tests |
| Completeness | NO phases/tasks/steps skipped; NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Reliability | ≥98% infrastructure/utility coverage; ≥98% mutation (cicd_lint packages) |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Overview

Framework V17 (commit `a747ac2ea`) delivered 12 new architecture fitness linters and partial
conformance migration. Six migration tasks were explicitly deferred as GAPs. A separate V18
plan for ENG-HANDBOOK.md documentation propagation was also pre-created (unexecuted).

V19 addresses two parallel concerns:

**Stream 1 — Prescriptive MANIFEST.yaml (new request)**:
The current `__PS_ID__ MANIFEST.yaml` specifies only 3 shallow fields (`required_root_files`,
`required_dirs: [server]`, `required_server_files`). It does not specify what subdirectories
must exist inside `server/`, what files each subdirectory must contain, or whether `client/`
and `e2e/` are required vs. optional. The user's request is to make it fully recursive — every
required subdirectory, its purpose, and its expected files.

**Stream 2 — V17 Deferred Migration (GAP-A through GAP-G)**:
Identity services (authz, idp, rs, rp, spa) still have domain code at PS-ID root instead of
in `server/`. sm-im has test files at root that belong in `server/`. No lifecycle or port_conflict
tests exist for identity services.

---

## Background

### V17 Deferred GAPs (carried into V19)

| GAP | PS-ID | Description | Current State |
|-----|-------|-------------|---------------|
| A / Task 5.2 | identity-authz | Move 60+ handler/route/service files from root → server/ | ALL at PS-ID root |
| A / Task 5.3 | identity-idp | Move 40+ handler/route/service files from root → server/ | ALL at PS-ID root |
| A / Task 5.4 | identity-rs | Move service.go, validator.go, tests from root → server/ | At PS-ID root |
| A / Task 5.5 | identity-rp | Create lifecycle_test.go and port_conflict_test.go in server/ | Missing entirely |
| A / Task 5.6 | identity-spa | Create lifecycle_test.go and port_conflict_test.go in server/ | Missing entirely |
| A / Task 5.7 | sm-im | Move http_test.go, im_database_test.go, etc. from root → server/ | At PS-ID root |

### Server/ Subdirectory State — All 10 PS-IDs

| PS-ID | server/apis/ | server/config/ | server/model/ | server/repository/ | Notes |
|-------|-------------|---------------|--------------|---------------------|-------|
| sm-kms | ❌ | ❌ | ❌ | ✅ | Legacy: businesslogic/ handler/ instead |
| sm-im | ✅ | ✅ | ✅ | ✅ | Canonical modern pattern |
| jose-ja | ✅ | ✅ | ✅ | ✅ | Canonical + service/ layer |
| pki-ca | ❌ | ✅ | ❌ | ❌ | Complex CA; cmd/ middleware/ instead |
| identity-authz | ❌ | ✅ | ❌ | ❌ | All code at root |
| identity-idp | ❌ | ✅ | ❌ | ❌ | All code at root |
| identity-rs | ❌ | ✅ | ❌ | ❌ | Partial code at root |
| identity-rp | ❌ | ✅ | ❌ | ❌ | Minimal server/ |
| identity-spa | ❌ | ✅ | ❌ | ❌ | Minimal server/ |
| skeleton-template | ✅ | ✅ | ✅ | ✅ | Canonical template |

### client/ and e2e/ State — All 10 PS-IDs

| PS-ID | client/ | e2e/ |
|-------|---------|------|
| sm-kms | ✅ | ✅ |
| sm-im | ✅ | ✅ |
| jose-ja | ❌ | ✅ |
| pki-ca | ❌ | ❌ |
| identity-authz | ❌ | ✅ |
| identity-idp | ❌ | ❌ |
| identity-rs | ❌ | ❌ |
| identity-rp | ❌ | ❌ |
| identity-spa | ❌ | ❌ |
| skeleton-template | ❌ | ✅ |

**Gap summary**: 8 PS-IDs missing client/; 5 PS-IDs missing e2e/.

---

## Target Structure — Canonical PS-ID Layout

The FULL target structure every PS-ID must eventually reach. Current gaps tracked as
`knownExclusions` in the linter; removed as migration completes.

```
internal/apps/{PS-ID}/
├── {SERVICE}.go                              REQUIRED — CLI entry: server/health/livez/readyz/shutdown
├── {SERVICE}_usage.go                        REQUIRED — CLI usage string via BuildUsageMain()
├── {SERVICE}_cli_test.go                     REQUIRED — CLI integration tests: help, version, unknown-subcommand
│                                              EXCEPTION: sm-im uses im_cli_commands_test.go + im_cli_url_test.go
├── client/                                   REQUIRED — typed HTTP client package for this service's API
│   ├── *.go (non-test, domain-named)          REQUIRED — at least one non-test .go file
│   │                                              e.g. client.go, messages.go, client_oam_mapper.go
│   └── *_test.go                              OPTIONAL — tests for client
├── e2e/                                      REQUIRED — E2E Docker Compose integration tests
│   ├── testmain_e2e_test.go                  REQUIRED — TestMain: starts Compose, waits for health endpoint
│   └── {SERVICE}_e2e_test.go                 REQUIRED — primary E2E scenarios
│                                              ALL .go files MUST have //go:build e2e tag
├── testing/ (optional)                       OPTIONAL — shared test helpers for root-package tests
└── server/                                   REQUIRED — ALL server implementation; NOTHING domain at root
    ├── server.go                             REQUIRED — admin HTTPS server: livez/readyz/shutdown + mTLS
    ├── public_server.go                      REQUIRED — public HTTPS server: browser/ + service/ paths
    │                                          EXCEPTION: sm-kms (legacy — no public_server.go)
    ├── swagger.go                            REQUIRED — OpenAPI spec serving (ServeHTTP returns embedded spec)
    ├── swagger_test.go                       REQUIRED — swagger serving tests
    ├── testmain_test.go                      REQUIRED — integration TestMain: shared server+DB for server/ tests
    ├── {SERVICE}_lifecycle_test.go           REQUIRED — dual-port startup, graceful shutdown, signal handling
    ├── {SERVICE}_port_conflict_test.go       REQUIRED — deterministic failure when ports already bound
    ├── apis/                                 REQUIRED — HTTP handler implementations by domain/resource
    │   └── *.go / *_test.go                  (domain-named: messages.go, jwk_handler.go, handler.go)
    │                                          EXCEPTION: sm-kms (legacy businesslogic/handler/ structure)
    │                                          EXCEPTION: pki-ca (complex CA — deferred to V20)
    ├── config/                               REQUIRED — server configuration package
    │   ├── config.go                         REQUIRED — Config struct, Load(), Validate()
    │   ├── config_test.go                    REQUIRED — valid/invalid/missing-field config tests
    │   └── config_test_helper.go             REQUIRED — NewTestConfig() shared fixture builder
    │                                          EXCEPTION for config_test_helper.go: pki-ca, identity services
    ├── model/                                REQUIRED — GORM persistence models
    │   └── *.go / *_test.go                  (domain-named: model.go, models.go, message.go)
    │                                          EXCEPTION: sm-kms, pki-ca (deferred to V20)
    └── repository/                           REQUIRED — database access layer
        ├── migrations/                       REQUIRED subdir — SQL migration files
        │   ├── NNNN_name.up.sql              (e.g. 3001_init.up.sql, range from registry.yaml)
        │   └── NNNN_name.down.sql
        ├── migrations.go                     REQUIRED — //go:embed migrations/*.sql + Migrate()
        └── {domain}_repository.go            (one or more GORM repository implementations)
                                              EXCEPTION: sm-kms, pki-ca (deferred to V20)
```

**Root file rule (MANDATORY)**: ALL files at the PS-ID root MUST start with `{SERVICE}_` prefix
OR be named `testmain_test.go`. FORBIDDEN at root: `swagger.go`, `handlers_*.go`, `routes.go`,
`service.go`, `middleware.go`, `http_test.go`, `validator.go`, any non-CLI implementation file.

---

## Technical Context

- **Language**: Go 1.26.1; CGO_ENABLED=0
- **MANIFEST path**: `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
- **Linter path**: `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/`
- **Pattern**: Detect-and-error (linter emits errors); no code generation by linter

### Affected Files — Phase 1 (MANIFEST + Linter Extension)

```
api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml    MODIFY
internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/
    apps_ps_id_template.go                                                    MODIFY
    apps_ps_id_template_test.go                                               MODIFY
```

### Affected Files — Phase 2 (Identity Service Migration)

```
internal/apps/identity-authz/server/   + swagger.go, lifecycle_test.go, port_conflict_test.go
internal/apps/identity-authz/          [60+ files moved from root to server/]
internal/apps/identity-idp/server/     + swagger.go, lifecycle_test.go, port_conflict_test.go
internal/apps/identity-idp/            [40+ files moved from root to server/]
internal/apps/identity-rs/server/      + swagger.go, service.go, validator.go, lifecycle_test.go, port_conflict_test.go
internal/apps/identity-rp/server/      + lifecycle_test.go, port_conflict_test.go; rp_test.go moved
internal/apps/identity-spa/server/     + lifecycle_test.go, port_conflict_test.go; spa_test.go moved
```

### Affected Files — Phase 3 (sm-im Root Cleanup)

```
internal/apps/sm-im/server/   + lifecycle and test file moves from root
internal/apps/sm-im/          [testmain_test.go root copy deleted]
```

### Affected Files — Phase 4 (Create Missing server/ Subdirs)

```
internal/apps/identity-{authz,idp,rs,rp,spa}/server/apis/        NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/model/       NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/repository/  NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/repository/migrations/  NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/repository/migrations.go  NEW
```

### Affected Files — Phase 5 (Create client/ Packages)

```
internal/apps/{jose-ja,pki-ca,identity-authz,identity-idp,identity-rs,
               identity-rp,identity-spa,skeleton-template}/client/  NEW (8 dirs)
```

### Affected Files — Phase 6 (Create e2e/ Packages)

```
internal/apps/{pki-ca,identity-idp,identity-rs,identity-rp,identity-spa}/e2e/  NEW (5 dirs)
```

---

## Phases

**Phase Status Legend**: `☐ TODO` | `🔄 IN PROGRESS` | `✅ COMPLETE` | `⏳ BLOCKED`

---

### Phase 0: Pre-flight Build Health (0.25h) [Status: ☐ TODO]

**Objective**: Verify clean baseline before any changes.

- `go build ./...` exits 0
- `go build -tags e2e,integration ./...` exits 0
- `go run ./cmd/cicd-lint lint-fitness` exits 0 (all 87 linters pass)
- `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` exits 0

**Success**: Output archived in `test-output/v19-phase0/`.

---

### Phase 1: Prescriptive MANIFEST.yaml + Linter Extension (4h) [Status: ☐ TODO]

**Objective**: Expand `__PS_ID__ MANIFEST.yaml` to fully specify server/ subdirectory structure,
config/ files, repository/ files/dirs, and required e2e/ files. Promote client/ and e2e/ from
optional_dirs to required_dirs. Extend the linter to validate all new MANIFEST fields.

#### 1a: Update **PS_ID** MANIFEST.yaml

Replace the current sparse MANIFEST.yaml with a fully annotated version:

**Fields to ADD**:

```yaml
# Promoted from optional_dirs
required_dirs:
  - server
  - client   # typed HTTP client — knownExclusions: jose-ja, pki-ca, identity-*, skeleton-template
  - e2e      # E2E tests — knownExclusions: pki-ca, identity-idp, identity-rs, identity-rp, identity-spa

# NEW: subdirectories required inside server/
required_server_dirs:
  - apis        # handler implementations — knownExclusions: sm-kms, pki-ca, all 5 identity services
  - config      # server config package — ALL 10 PS-IDs have this; NO exclusions
  - model       # GORM models — knownExclusions: sm-kms, pki-ca, all 5 identity services
  - repository  # database access — knownExclusions: sm-kms, pki-ca, all 5 identity services

# NEW: files required inside server/config/
required_server_config_files:
  - config.go
  - config_test.go
  - config_test_helper.go   # knownExclusions: pki-ca, identity-{authz,idp,rs,rp,spa}

# NEW: files required inside server/repository/
required_server_repository_files:
  - migrations.go           # knownExclusions: pki-ca, identity-{authz,idp,rs,rp,spa} (Phase 4 adds)

# NEW: subdirectories required inside server/repository/
required_server_repository_dirs:
  - migrations              # knownExclusions: pki-ca, identity-{authz,idp,rs,rp,spa} (Phase 4 adds)

# NEW: files required inside e2e/
required_e2e_files:
  - testmain_e2e_test.go    # knownExclusions: pki-ca, identity-idp, identity-rs, identity-rp, identity-spa
  - __SERVICE___e2e_test.go # same knownExclusions
```

Each field has rich YAML comments explaining purpose, package pattern, cross-references to
ENG-HANDBOOK.md sections, and exception rationale.

#### 1b: Extend apps_ps_id_template Linter

Add new fields to `psIDManifest` struct:

```go
type psIDManifest struct {
    RequiredRootFiles             []string `yaml:"required_root_files"`
    RequiredDirs                  []string `yaml:"required_dirs"`
    RequiredServerFiles           []string `yaml:"required_server_files"`
    RequiredServerDirs            []string `yaml:"required_server_dirs"`              // NEW
    RequiredServerConfigFiles     []string `yaml:"required_server_config_files"`      // NEW
    RequiredServerRepositoryFiles []string `yaml:"required_server_repository_files"`  // NEW
    RequiredServerRepositoryDirs  []string `yaml:"required_server_repository_dirs"`   // NEW
    RequiredE2EFiles              []string `yaml:"required_e2e_files"`                // NEW
}
```

New check functions (one per new YAML field category):
1. `checkServerDirs` — verifies `server/{dir}` exists for each entry in RequiredServerDirs
2. `checkServerConfigFiles` — verifies `server/config/{file}` exists
3. `checkServerRepositoryFiles` — verifies `server/repository/{file}` exists
4. `checkServerRepositoryDirs` — verifies `server/repository/{dir}` exists
5. `checkE2EFiles` — verifies `e2e/{file}` exists with `__SERVICE__` substitution applied

knownExclusions maps registered before enabling each check (see table in 1a above).

**Success**: `lint-fitness` exits 0; MANIFEST is fully prescriptive; `apps_ps_id_template`
linter has ≥98% coverage; all 5 new check functions have corresponding test cases.

---

### Phase 2: Identity Services Server Code Migration (20h) [Status: ☐ TODO]

**Objective**: Complete V17 GAPs — move domain code from PS-ID root → server/ for
identity-authz, identity-idp, identity-rs, identity-rp, identity-spa.

**Package declaration rule**: Files moved from `internal/apps/identity-authz/` root declare
`package identity_authz`. Files moved to `server/` declare `package server`. Test files that
use an external test package (`package identity_authz_test`) must be re-declared as
`package server_test`. Update all intra-package imports accordingly.

#### 2a: identity-authz — ~60 files from root → server/

Files to KEEP at root (CLI only): `authz.go`, `authz_usage.go`, `authz_cli_test.go`, `authz_contract_test.go`

Files currently in server/ (keep): `server.go`, `public_server.go`, `admin.go`, `admin_error_test.go`,
`admin_test.go`, `server_integration_test.go`, `testmain_test.go`, `config/`

Files to MOVE to server/ (representative; full list from `git ls-files`):
- `swagger.go`, `swagger_test.go`
- `service.go`, `service_lifecycle_test.go`, `service_test.go`
- `authorization_request.go` + `authorization_request_test.go`
- `cleanup.go`, `cleanup_test.go`
- `client_authentication.go` + tests
- `code_generator.go`, `device_code_generator.go` + tests
- `dpop/` subdir
- `handlers_*.go` (all ~25 handlers files)
- `handlers_*_test.go` (all handler test files)
- `middleware.go` + test
- `pkce/` subdir
- `request_uri_generator.go` + test
- `routes.go` + test
- `performance_bench_test.go`, `test_helpers_test.go`
- `authz_test.go` (HTTP test, not CLI)

Files to CREATE in server/: `authz_lifecycle_test.go`, `authz_port_conflict_test.go`

#### 2b: identity-idp — ~40 files from root → server/

Files to KEEP at root: `idp.go`, `idp_usage.go`, `idp_cli_test.go`, `idp_contract_test.go`

Files to MOVE to server/ (representative):
- `swagger.go`, `swagger_test.go`
- `service.go` + service tests
- `backchannel_logout.go` + test
- `client_secret.go` + test
- `handlers_consent*.go` + tests, `handlers_discovery.go` + test
- `handlers_jwks.go` + test, `handlers_login.go` + test, `handlers_logout.go` + test
- `handlers_oidc_e2e_test.go`, `handlers_openapi_validation_test.go`
- `handlers_parallel_safety_test.go`, `handlers_postgres_test.go`
- `handlers_security_*.go` + tests, `handlers_token_*.go` + tests
- `handlers_userinfo*.go` + tests
- `magic_test_constants.go`, `middleware.go` + tests, `random.go`
- `routes.go` + test
- `auth/` subdir, `templates/` subdir
- test helpers

Files to CREATE in server/: `idp_lifecycle_test.go`, `idp_port_conflict_test.go`

#### 2c: identity-rs — ~7 files from root → server/

Files to KEEP at root: `rs.go`, `rs_usage.go`, `rs_cli_test.go`, `rs_contract_test.go`

Files to MOVE to server/:
- `swagger.go`, `swagger_test.go`
- `service.go`, `service_admin_test.go`, `service_test.go`
- `validator.go`

Files to CREATE in server/: `rs_lifecycle_test.go`, `rs_port_conflict_test.go`

#### 2d: identity-rp — Create missing tests; move rp_test.go

Move from root → server/: `rp_test.go` (HTTP handler test, not a CLI test)
Files to CREATE in server/: `rp_lifecycle_test.go`, `rp_port_conflict_test.go`

#### 2e: identity-spa — Create missing tests; move spa_test.go

Move from root → server/: `spa_test.go` (HTTP handler test, not a CLI test)
Files to CREATE in server/: `spa_lifecycle_test.go`, `spa_port_conflict_test.go`

**Phase 2 success**: All identity service roots have ONLY CLI files; all server/ directories
have swagger.go + testmain_test.go + lifecycle + port_conflict tests; `go test ./internal/apps/identity-.../...` passes for all five.

---

### Phase 3: sm-im Root Cleanup (2h) [Status: ☐ TODO]

**Objective**: Complete V17 GAP Task 5.7 — move all non-CLI test files from sm-im root to server/.

Files to MOVE from root → `server/`:
```
http_test.go
http_errors_test.go
response_body_test.go
im_database_test.go
im_server_lifecycle_test.go
im_lifecycle_test.go        ← lifecycle test (MOVE + remove root copy if duplicate)
im_port_conflict_test.go    ← port conflict test (MOVE + remove root copy if duplicate)
```

Files to DELETE from root (redundant):
```
testmain_test.go   ← server/testmain_test.go already exists
```

Files to KEEP at root:
```
im.go, im_usage.go, im_cli_commands_test.go, im_cli_url_test.go
```

**Phase 3 success**: sm-im root has ONLY `im.go`, `im_usage.go`, `im_cli_commands_test.go`,
`im_cli_url_test.go`; `go test ./internal/apps/sm-im/...` passes.

---

### Phase 4: Create Missing server/ Subdirectory Packages (12h) [Status: ☐ TODO]

**Objective**: Create the canonical `server/apis/`, `server/model/`, `server/repository/` (with
`migrations/`) for all 5 identity services. This allows removing 21 entries from knownExclusions.

**sm-kms and pki-ca are EXPLICITLY DEFERRED to V20.** See Decision 3.

#### Per-PS-ID scope

| PS-ID | server/apis/ | server/model/ | server/repository/ + migrations/ | Notes |
|-------|-------------|--------------|-----------------------------------|-------|
| identity-authz | ✅ Phase 2 populates | CREATE (domain types) | CREATE | Phase 2 handlers → apis/ |
| identity-idp | ✅ Phase 2 populates | CREATE (domain types) | CREATE | Phase 2 handlers → apis/ |
| identity-rs | ✅ Phase 2 populates | CREATE (domain types) | CREATE | Phase 2 handlers → apis/ |
| identity-rp | CREATE (minimal handler) | CREATE | CREATE | Thin proxy service |
| identity-spa | CREATE (minimal handler) | CREATE | CREATE | Thin proxy service |

Each new `repository/` must have:
- `migrations/` subdir
- Migration number from `api/cryptosuite-registry/registry.yaml` range for that PS-ID
- `migrations.go` with `//go:embed migrations/*.sql`

**Phase 4 success**: All 5 identity services have `server/apis/`, `server/model/`,
`server/repository/migrations/`, `migrations.go`; `go build ./...` exits 0;
knownExclusions reduced by 21 entries.

---

### Phase 5: Create Missing client/ Packages (8h) [Status: ☐ TODO]

**Objective**: Create typed HTTP client packages for the 8 PS-IDs currently missing them.

| PS-ID | Minimum client/ content | Reference Pattern |
|-------|------------------------|------------------|
| jose-ja | `client.go` — GetJWKS, RotateJWK | sm-kms/client/ |
| pki-ca | `client.go` — IssueCert, RevokeCert, GetCRL | sm-kms/client/ |
| identity-authz | `client.go` — Authorize, Introspect, Token | sm-kms/client/ |
| identity-idp | `client.go` — Login, Logout, JWKS | sm-kms/client/ |
| identity-rs | `client.go` — ValidateToken, GetResources | sm-kms/client/ |
| identity-rp | `client.go` — Callback, Logout | sm-kms/client/ |
| identity-spa | `client.go` — minimal API surface | sm-kms/client/ |
| skeleton-template | `client.go` — placeholder client | sm-im/client/ |

Each `client/` uses the oapi-generated `api/client/client.gen.go` as the underlying HTTP client
(typed wrapper over the generated interface).

**Phase 5 success**: All 10 PS-IDs have `client/`; `go build ./...` exits 0;
`required_dirs: client` knownExclusions emptied.

---

### Phase 6: Create Missing e2e/ Packages (8h) [Status: ☐ TODO]

**Objective**: Create E2E test packages for 5 PS-IDs that currently lack them.

| PS-ID | Required files | Notes |
|-------|---------------|-------|
| pki-ca | `testmain_e2e_test.go`, `ca_e2e_test.go` | Complex CA — basic smoke test |
| identity-idp | `testmain_e2e_test.go`, `idp_e2e_test.go` | Login/JWKS smoke test |
| identity-rs | `testmain_e2e_test.go`, `rs_e2e_test.go` | Token validation smoke test |
| identity-rp | `testmain_e2e_test.go`, `rp_e2e_test.go` | Callback smoke test |
| identity-spa | `testmain_e2e_test.go`, `spa_e2e_test.go` | SPA smoke test |

ALL `.go` files in `e2e/` MUST have `//go:build e2e` as the first line.
Use `sm-im/e2e/` or `jose-ja/e2e/` as reference patterns.

**Phase 6 success**: All 10 PS-IDs have `e2e/`; `go build -tags e2e ./...` exits 0.

---

### Phase 7: Remove knownExclusions + Final Validation (3h) [Status: ☐ TODO]

**Objective**: After Phases 1–6, all PS-IDs (except sm-kms and pki-ca) conform to the
canonical template. Remove temporary knownExclusions from the `apps_ps_id_template` linter.

**Permanent exceptions (never removed)**:
- `sm-kms`: `public_server.go` optional (legacy — documented permanent exception)
- `sm-im`: `__SERVICE___cli_test.go` exclusion (uses `im_cli_commands_test.go` + `im_cli_url_test.go`)
- `sm-kms`, `pki-ca`: `required_server_dirs: [apis, model, repository]` — deferred V20

**Temporary exclusions to REMOVE** after Phases 1–6 complete:
- All identity service exclusions for `required_server_dirs`
- All identity service exclusions for `required_server_config_files: config_test_helper.go`
- All identity service exclusions for `required_server_repository_files`
- All identity service exclusions for `required_dirs: client`
- All identity service exclusions for `required_dirs: e2e`
- All identity service exclusions for `required_e2e_files`

**Phase 7 success**: `go run ./cmd/cicd-lint lint-fitness` exits 0; only 3 permanent
exceptions remain (sm-kms public_server, sm-im cli_test, sm-kms/pki-ca server/ subdirs).

---

### Phase 8: Knowledge Propagation (2h) [Status: ☐ TODO]

**Objective**: Apply lessons learned to permanent project artifacts. NEVER skip.

- Review `lessons.md` from all prior phases
- Update ENG-HANDBOOK.md: canonical PS-ID structure spec, MANIFEST field catalog, migration patterns
- Update `docs/target-structure.md`: reflect new canonical structure
- Update instruction files where V19 exposes new coding/testing patterns
- Update `.github/skills/fitness-function-gen/SKILL.md` with recursive MANIFEST pattern
- Verify propagation: `go run ./cmd/cicd-lint lint-docs` exits 0
- Commit all artifact updates with separate semantic commits per artifact type

---

## Executive Decisions

### Decision 1: MANIFEST.yaml Granularity Scope

**Options**:
- A: Add `required_server_dirs` (dir names only — no per-dir file specs except config/ and repository/)
- B: Add per-file specs for every file in every subdirectory (too brittle — domain-specific)
- C: A + `required_server_config_files` + `required_server_repository_files` + `required_e2e_files` ✓ **SELECTED**

**Decision**: Option C. dirs with predictable content (config/, repository/, e2e/) get required_files
specs. Dirs with domain-specific content (apis/, model/) get only directory-level requirements.

**Rationale**: config/, repository/, and e2e/ have fixed, non-domain file sets. apis/ and model/
vary by service domain — prescribing filenames would require per-PS-ID MANIFEST variants.

### Decision 2: client/ and e2e/ Promotion from Optional to Required

**Options**:
- A: Promote to required_dirs + large knownExclusions ✓ **SELECTED**
- B: Keep as optional_dirs (no enforcement)
- C: Add required_dirs_with_exceptions section (unnecessary complexity)

**Decision**: Option A. Makes the requirement explicit even with current gaps; knownExclusions
list makes the migration debt visible in every CI run until resolved.

### Decision 3: sm-kms and pki-ca server/ Subdirectory Migration

**Options**:
- A: Include in V19 Phase 4
- B: Defer to V20 as separate sub-plans ✓ **SELECTED**
- C: Permanent exception (never migrate)

**Decision**: Option B. pki-ca's CA architecture is fundamentally different (complex cert issuance,
profiles system, no clean repository layer). sm-kms uses a legacy businesslogic/ ORM structure.
Both require dedicated analysis sub-plans before structural migration.

### Decision 4: identity-authz/idp handler file destination

**Options**:
- A: Move handlers_*.go to `server/` root (single flat server package)
- B: Move handlers_*.go to `server/apis/` subpackage ✓ **SELECTED**
- C: Create domain-specific subdirs (server/authz/, server/token/, etc.)

**Decision**: Option B. handlers_*.go into `server/apis/` as `package apis`, matching the
sm-im canonical pattern (sm-im/server/apis/messages.go). Domain logic (service.go, cleanup.go)
stays at `server/` root as `package server`. This separates HTTP concerns from domain concerns.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| identity-authz circular import after package split | Medium | High | `go build ./...` after EVERY file move; fix cycles immediately |
| Package declaration changes break test package references | High | Medium | All `_test` packages updated before attempting `go test` |
| Large identity-authz migration (~60 files) introduces build errors | High | Medium | Move in batches: swagger first, service next, handlers last |
| sm-im test moves break existing CI | Low | Medium | Run `go test ./internal/apps/sm-im/...` after each file move |
| migrations/ numbering conflicts with framework range | Low | High | Verify against registry.yaml migration ranges before creating files |
| pki-ca server/ conflict with existing cmd/ structure | Low | High | pki-ca DEFERRED — knownExclusions maintained for all pki-ca server checks |

---

## Quality Gates — MANDATORY

**Per-Action**:
- ✅ `go test ./...` — 100% passing, zero skips
- ✅ `go build ./...` AND `go build -tags e2e,integration ./...`
- ✅ `golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ `apps_ps_id_template` package: ≥98% after Phase 1
- ✅ Identity service packages: ≥95% after Phase 2
- ✅ New `client/` packages: ≥95%
- ✅ New `e2e/` packages: excluded (E2E-tagged)

**Per-Phase**:
- ✅ `go run ./cmd/cicd-lint lint-fitness` exits 0 after each phase
- ✅ `go run ./cmd/cicd-lint lint-docs` exits 0 after Phase 8
- ✅ Race detector clean: `go test -race -count=2 ./...`

**ENG-HANDBOOK.md Cross-References**:
- [§5.1 Service Framework Pattern](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern)
- [§10.2 Unit Testing](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy)
- [§10.4 E2E Testing](../../docs/ENG-HANDBOOK.md#104-e2e-testing-strategy)
- [§11.2 Quality Gates](../../docs/ENG-HANDBOOK.md#112-quality-gates)
- [§14.1 Coding Standards](../../docs/ENG-HANDBOOK.md#141-coding-standards)
- [§14.2 Version Control](../../docs/ENG-HANDBOOK.md#142-version-control)
- [§14.8 Phase Post-Mortem](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation)

---

## Success Criteria

- [ ] `__PS_ID__ MANIFEST.yaml` is fully prescriptive: server_dirs, config_files, repository_files/dirs, e2e_files
- [ ] `apps_ps_id_template` linter validates all MANIFEST fields recursively with ≥98% coverage
- [ ] All 10 PS-ID roots contain ONLY `{SERVICE}_`-prefixed CLI files
- [ ] All 10 PS-ID `server/` dirs have `swagger.go`, `testmain_test.go`, lifecycle, port_conflict tests
- [ ] All 5 identity services have `server/apis/`, `server/model/`, `server/repository/migrations/`
- [ ] All 10 PS-IDs have `client/` (except sm-kms, pki-ca — explicitly deferred to V20)
- [ ] All 10 PS-IDs have `e2e/` (except sm-kms, pki-ca — explicitly deferred to V20)
- [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0 with only 3 permanent exceptions
- [ ] All quality gates passing; CI/CD green
- [ ] Evidence archived in `test-output/v19-*/`
