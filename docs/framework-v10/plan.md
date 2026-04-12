# Implementation Plan — Framework v10: Canonical Template Registry

**Status**: Planning
**Created**: 2026-04-12
**Last Updated**: 2026-04-12
**Purpose**: Correct the v9 architecture failure where canonical deployment templates were placed
in the wrong location AND implemented with the wrong mechanism (Go `embed.FS`). Implement the
correct architecture: `api/cryptosuite-registry/templates/` is a plain directory of parameterized
configuration files (Dockerfiles, compose.yml, config YAML) with `__KEY__` placeholders in BOTH
directory paths AND file contents. `cicd-lint` reads these files at runtime, loops through all
PS-ID/product/suite combinations, generates an in-memory expected filesystem, and recursively
compares it against the actual `./configs/` and `./deployments/` directories on disk.

**PREREQUISITE — quizme-v1 MERGED, quizme-v2 MERGED, quizme-v3 PENDING**: quizme-v1 (10 questions)
answered and merged (8 decisions applied). quizme-v2 (7 questions) answered and merged — resolved
Decisions 7/8/10/12 (previously BLOCKED/UNANSWERED) and added Decisions 13-14. Deep analysis
uncovered critical issues (pki-init code mismatch, shared-postgres documentation gaps,
Docker Compose env var expansion limitation) requiring user decisions before Phase 1 begins.
See [`docs/framework-v10/quizme-v3.md`](quizme-v3.md).

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL documentation must be accurate and complete
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥98% coverage for infrastructure/utility)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**
- ✅ **Fix issues immediately** — blockers, test failures, lint errors: STOP and address
- ✅ **NEVER defer**: No "fix later", no "non-critical", no shortcuts
- ✅ **Root cause**: Address root cause, not symptoms

---

## Overview

Framework v9 Task 8.1 had two compounding failures:

1. **Wrong location**: templates placed at `internal/.../template_drift/templates/` instead of
   `api/cryptosuite-registry/templates/`.
2. **Wrong mechanism**: v9 used Go `//go:embed` + `embed.FS`. The correct mechanism is: cicd-lint
   reads template files directly from disk at runtime, generates an in-memory expected filesystem
   by expanding all PS-ID/product/suite combinations, and recursively compares it against the
   actual `./configs/` and `./deployments/` directories.

`api/cryptosuite-registry/` is **NOT a Go package** and must NOT contain `.go` files.
It is a plain directory of actual configuration files (Dockerfiles, compose.yml, config YAML)
with `__KEY__` placeholders in BOTH directory paths AND file contents. No build tooling, no
import aliases, no `embed.FS`.

Framework v10 implements the correct architecture by:
1. Creating the parameterized canonical directory structure under `api/cryptosuite-registry/templates/`
   mirroring the `./deployments/` and `./configs/` trees, with `__KEY__` in paths and content
2. Rewriting the `template_drift` linter to: walk the templates directory at runtime, expand
   all `__PS_ID__`/`__PRODUCT__`/`__SUITE__` combinations, build an in-memory expected FS,
   and recursively compare against actual files on disk
3. Removing the incorrect `//go:embed` implementation from `template_drift.go`
4. Registering a single comprehensive `template-compliance` linter (replacing the per-file approach)
5. Updating documentation to reference canonical template files

---

## Background

### v9 Architecture Failures

**Failure 1 — Wrong location**: v9 chose `template_drift/templates/` (same-package embed path)
instead of the required `api/cryptosuite-registry/templates/`.

**Failure 2 — Wrong mechanism**: v9 used Go `//go:embed` + `embed.FS`. This was architecturally
wrong. `api/cryptosuite-registry/` is not Go source code. It is a plain directory of configuration
files that represents the canonical required content for every deployment artifact. cicd-lint reads
these files from disk using `os.WalkDir`, not via Go's embed system.

### Correct Architecture: Parameterized Template Directory

`api/cryptosuite-registry/templates/` mirrors the structure of `./configs/` and `./deployments/`
with `__KEY__` placeholders substituted for values that vary per PS-ID, product, or suite.
**Both the directory paths AND the file contents are parameterized.**

**Example — PS-ID level (`__PS_ID__` in path)**:
```
templates/deployments/__PS_ID__/Dockerfile      → deployments/sm-kms/Dockerfile (×10)
templates/deployments/__PS_ID__/compose.yml     → deployments/sm-kms/compose.yml (×10)
templates/configs/__PS_ID__/__PS_ID__.yml        → configs/sm-kms/sm-kms.yml (×10)
```

**Example — Product level (`__PRODUCT__` in path, content parameterized)**:
```
templates/deployments/__PRODUCT__/compose.yml   → deployments/sm/compose.yml (×5)
                                                → deployments/jose/compose.yml
                                                → deployments/pki/compose.yml
                                                → deployments/identity/compose.yml
                                                → deployments/skeleton/compose.yml
```

**Example — Suite level (`__SUITE__` in path)**:
```
templates/deployments/__SUITE__/Dockerfile      → deployments/cryptoutil/Dockerfile (×1)
templates/deployments/__SUITE__/compose.yml     → deployments/cryptoutil/compose.yml (×1)
```

**Example — Shared telemetry (static path, content parameterized)**:
```
templates/deployments/shared-telemetry/compose.yml                    → deployments/shared-telemetry/compose.yml
templates/deployments/shared-telemetry/otel/otel-collector-config.yaml → deployments/shared-telemetry/otel/otel-collector-config.yaml
```

### How cicd-lint Processes Templates

1. Walk all files under `api/cryptosuite-registry/templates/`
2. For each template file, inspect its path for expansion keys:
   - Path contains `__PS_ID__`: expand × 10 (one per PS-ID from registry.yaml)
   - Path contains `__PRODUCT__`: expand × 5 (sm, jose, pki, identity, skeleton); uses per-product
     params derived from registry.yaml (e.g. `__PRODUCT_INCLUDE_LIST__` built from each product's PS-IDs)
   - Path contains `__SUITE__`: expand × 1 (currently `cryptoutil`); parameterized for future renames
- Path contains `__INFRA_TOOL__`: expand for each infrastructure tool (currently unused — shared-telemetry is static)
- No expansion key in path: substitute generic params in content; use template-relative path directly
     (e.g. `deployments/shared-telemetry/compose.yml` is compared as-is with `__SUITE__` content substitution)
1. For each expansion: substitute ALL `__KEY__` params in both the resolved path and file content
2. Collect all (resolvedPath → expectedContent) pairs → in-memory expected filesystem
3. Recursively compare in-memory FS against actual `./configs/` and `./deployments/` on disk:
   every expected file must exist at exactly the resolved relative path with identical content

### Complete Template Directory Structure

```
api/cryptosuite-registry/templates/
  deployments/
    __PS_ID__/                          ← expands for each of 10 PS-IDs
      Dockerfile                        ← __PS_ID__ + build/label/healthcheck params
      compose.yml                       ← __PS_ID__ + port params; pki-init with --domain=__PS_ID__
      config/
        __PS_ID__-app-common.yml        ← __PS_ID__ + shared config params (PS-ID prefix naming)
        __PS_ID__-app-sqlite-1.yml      ← __PS_ID__ + __SERVICE_APP_PORT_SQLITE_1__
        __PS_ID__-app-sqlite-2.yml      ← __PS_ID__ + __SERVICE_APP_PORT_SQLITE_2__
        __PS_ID__-app-postgresql-1.yml   ← __PS_ID__ + __SERVICE_APP_PORT_PG_1__
        __PS_ID__-app-postgresql-2.yml   ← __PS_ID__ + __SERVICE_APP_PORT_PG_2__
      secrets/                          ← Decision 14: secrets directory template
        unseal-1of5.secret              ← __PS_ID__-unseal-key-1-of-5-BASE64_CHAR43
        unseal-2of5.secret              ← (through unseal-5of5.secret)
        unseal-3of5.secret
        unseal-4of5.secret
        unseal-5of5.secret
        hash-pepper-v3.secret           ← __PS_ID__-hash-pepper-v3-BASE64_CHAR43
        postgres-url.secret             ← postgres connection string
        postgres-username.secret        ← __PS_ID___database_user
        postgres-password.secret        ← BASE64_CHAR43
        postgres-database.secret        ← __PS_ID___database
        browser-username.secret         ← __PS_ID__-browser-user
        browser-password.secret         ← BASE64_CHAR43
        service-username.secret         ← __PS_ID__-service-user
        service-password.secret         ← BASE64_CHAR43
    __PRODUCT__/                        ← expands × 5 (sm, jose, pki, identity, skeleton)
      compose.yml                       ← __PRODUCT__ + per-product params; pki-init with --domain=__PRODUCT__;
                                          all 4 postgres secrets; NO image: on service overrides
      secrets/                          ← Product-level secrets (Decision 14)
        unseal-1of5.secret              ← __PRODUCT__-unseal-key-1-of-5-BASE64_CHAR43
        (... same 14 files, browser/service use .secret.never suffix ...)
    __SUITE__/                          ← expands × 1 (cryptoutil); parameterized for future renames
      Dockerfile                        ← 4-stage build pattern; uses __SUITE__ + suite-level display params
      compose.yml                       ← __SUITE__ + all product/PS-ID references; pki-init with --domain=__SUITE__;
                                          all 4 postgres secrets
      secrets/                          ← Suite-level secrets (Decision 14)
        unseal-1of5.secret              ← __SUITE__-unseal-key-1-of-5-BASE64_CHAR43
        (... same 14 files, browser/service use .secret.never suffix ...)
    shared-postgres/                    ← Decision 10: full shared-postgres template
      compose.yml                       ← __SUITE__ substitution for service names
      init-leader-databases.sql         ← 30 databases from __PS_ID_DATABASE_LIST__
      init-follower-databases.sql       ← follower schema setup
      setup-logical-replication.sh      ← replication subscriptions for all 10 PS-IDs
      secrets/                          ← shared-postgres secrets
        postgres-username.secret        ← cryptoutil_admin
        postgres-password.secret        ← BASE64_CHAR43
    shared-telemetry/                   ← static path (Decision 11: compose + otel are templated)
      compose.yml                       ← __SUITE__ substitution for telemetry service names
      otel/
        otel-collector-config.yaml      ← __SUITE__ substitution for OTLP endpoints
  configs/
    __PS_ID__/                          ← expands for each of 10 PS-IDs
      __PS_ID__-framework.yml           ← Decision 7/13: framework config (templateable)
                                          (domain config is __PS_ID__-domain.yml — NOT templated)
```

**Template file count**: ~60 physical files (estimate pending quizme-v3 resolution):
- 8 deployment files + 14 secrets under `deployments/__PS_ID__/` = 22
- 1 compose + 14 secrets under `deployments/__PRODUCT__/` = 15
- 2 deployment files + 14 secrets under `deployments/__SUITE__/` = 16
- 4 files + 2 secrets under `deployments/shared-postgres/` = 6
- 2 under `deployments/shared-telemetry/` (compose.yml + otel-collector-config.yaml)
- 1 under `configs/__PS_ID__/` (`__PS_ID__-framework.yml`)

**Expected file count after expansion**: ~232 (estimate):
- 22 PS-ID × 10 = 220 + 15 product × 5 = 75 + 16 suite × 1 = 16 + 6 shared-postgres + 2 shared-telemetry + 1 config × 10 = 10 ≈ 329
- ⚠️ Exact count depends on quizme-v3 resolution of shared-postgres gaps and secrets structure

### Existing Code That Changes

`template_drift.go` (currently uses embed):
- REMOVE: `//go:embed templates/*` and `var templatesFS embed.FS`
- REMOVE: `instantiate(templateName, params)` (replaced by runtime reader)
- REMOVE: per-file check wrappers (`CheckDockerfile`, `CheckCompose`, etc.)
- ADD: `LoadTemplatesDir(root string) (map[string]string, error)` — `os.WalkDir` from project root
- ADD: `BuildExpectedFS(templates map[string]string, registry *Registry) (map[string]string, error)` —
  for each template path: detect expansion key (`__PS_ID__`, `__PRODUCT__`, `__SUITE__`, `__INFRA_TOOL__`),
  iterate all values from registry, substitute in both path and content using per-value param sets
- ADD: `CompareExpectedFS(expectedFS map[string]string, projectRoot string) error` — recursive diff
- KEEP: `buildParams(psID)`, `normalizeCommentAlignment`, `normalizeLineEndings`

`lint-fitness-registry.yaml`:
- REMOVE: per-file linter entries (`template-dockerfile`, `template-compose`, `template-config-*`, etc.)
- ADD: single entry `template-compliance` wired to `CheckTemplateCompliance(logger)`

### Template Content Parameters (unchanged from v9)

All existing `__KEY__` parameters remain valid. The `buildParams(psID)` function builds the
full substitution map for a given PS-ID. Product/suite files use a subset of params
(`__SUITE__`, `__IMAGE_TAG__`, `__BUILD_DATE__`, plus product-specific display names).

---

## Technical Context

- **Language**: Go 1.26.1
- **`api/cryptosuite-registry/`**: Plain directory — NO Go package, NO `.go` files, NO `embed.FS`
- **Template reading**: `os.WalkDir("api/cryptosuite-registry/templates")` at runtime (relative to project root)
- **Expansion logic**: path contains `__PS_ID__` → loop registry PS-IDs; static paths → compare directly
- **In-memory FS type**: `map[string]string` (resolved relative path → expected file content)
- **Registry source**: `api/cryptosuite-registry/registry.yaml` (already read by `AllProductServices()`)
- **Linter registry**: `internal/apps/tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml`
- **Related files (changed)**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/template_drift.go` — engine rewrite
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/checks.go` — replace per-file checks
  - `internal/apps/tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml` — registry update
- **Related files (deleted)**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/` — wrong location, deleted
- **Related files (created)**:
  - `api/cryptosuite-registry/templates/` — 14 parameterized template files (see Structure above)
- **Documentation**:
  - `docs/deployment-templates.md` — update to reference canonical template files
  - `docs/target-structure.md` — add `api/cryptosuite-registry/templates/` listing

---

## Phases

### Phase 1: Create Canonical Template Directory (4h) [Status: ☐ TODO]

**Objective**: Create `api/cryptosuite-registry/templates/` with all 14 parameterized template
files. These are PLAIN FILES — not `.go`, not embedded. Structure mirrors `./deployments/` and
`./configs/` with `__KEY__` in both paths and content. Also fix actual deployment files to
match decisions from quizme-v1 (pki-init at all levels, postgres secrets, image removal).

**1A — PS-ID level templates** (8 files, each expands × 10):
- `deployments/__PS_ID__/Dockerfile` — based on current `Dockerfile.tmpl` content (same params)
- `deployments/__PS_ID__/compose.yml` — based on current `compose.yml.tmpl` content;
  MUST include `pki-init` service with `--domain=__PS_ID__` (Decision 4)
- `deployments/__PS_ID__/config/__PS_ID__-app-common.yml` — from `config-common.yml.tmpl`;
  filename uses PS-ID prefix (matches actual naming: `sm-kms-app-common.yml`)
- `deployments/__PS_ID__/config/__PS_ID__-app-sqlite-1.yml` — from `config-sqlite.yml.tmpl` + instance-1
- `deployments/__PS_ID__/config/__PS_ID__-app-sqlite-2.yml` — from `config-sqlite.yml.tmpl` + instance-2
- `deployments/__PS_ID__/config/__PS_ID__-app-postgresql-1.yml` — from `config-postgresql.yml.tmpl` + instance-1
- `deployments/__PS_ID__/config/__PS_ID__-app-postgresql-2.yml` — from `config-postgresql.yml.tmpl` + instance-2

**1B — PS-ID standalone config** (1 file, expands × 10):
- `configs/__PS_ID__/__PS_ID__-framework.yml` — framework settings only (Decision 7/13);
  full-file comparison. Domain config (`__PS_ID__-domain.yml`) is PS-ID-specific and NOT templated.

**1C — Product compose file** (1 physical file, expands × 5):
- `deployments/__PRODUCT__/compose.yml` — template content uses `__PRODUCT__`, `__SUITE__`,
  `__IMAGE_TAG__`, and `__PRODUCT_INCLUDE_LIST__`. Template MUST include:
  - `pki-init` service with `--domain=__PRODUCT__` (Decision 4)
  - All 4 postgres secrets: url, username, password, database (Decision 6)
  - NO `image:` on service overrides — PS-ID level is single source (Decision 5)

**1D — Suite files** (2 files, `__SUITE__` in path, expands × 1):
- `deployments/__SUITE__/Dockerfile` — 4-stage build pattern with `__SUITE__` params
- `deployments/__SUITE__/compose.yml` — suite compose with `__SUITE__` params; MUST include:
  - `pki-init` with `--domain=__SUITE__` (Decision 4)
  - All 4 postgres secrets (Decision 6)

**1E — shared-telemetry templates** (2 files, static path):
- `deployments/shared-telemetry/compose.yml` — with `__SUITE__` substitution (Decision 11)
- `deployments/shared-telemetry/otel/otel-collector-config.yaml` — with `__SUITE__` substitution
  (Grafana dashboards and alerts are NOT templated — too complex/fragile)

**1F — Fix actual deployment files** (bring actual files into compliance with decisions):
- Add `pki-init` with `--domain=<product>` to jose, pki, identity, skeleton product compose files
- Add `--domain=<ps-id>` to all 10 PS-ID compose `pki-init` commands
- Add `postgres-username.secret`, `postgres-password.secret`, `postgres-database.secret`
  to jose, pki, identity, skeleton product compose secrets sections
- Remove `image:` from sm product compose service overrides
- Split all 10 standalone configs into `<ps-id>-framework.yml` + `<ps-id>-domain.yml` (Decision 7)

**1G — Secrets directory templates** (Decision 14):
- PS-ID: `deployments/__PS_ID__/secrets/` — 14 template files with `BASE64_CHAR43` placeholders
- Product: `deployments/__PRODUCT__/secrets/` — 14 template files (`.secret.never` for browser/service)
- Suite: `deployments/__SUITE__/secrets/` — 14 template files (same as product)
- shared-postgres: `deployments/shared-postgres/secrets/` — 2-3 template files

**1H — shared-postgres templates** (Decision 10):
- `deployments/shared-postgres/compose.yml` — with `__SUITE__` substitution
- `deployments/shared-postgres/init-leader-databases.sql` — 30 databases from registry
- `deployments/shared-postgres/init-follower-databases.sql` — follower schema setup
- `deployments/shared-postgres/setup-logical-replication.sh` — all 10 PS-ID subscriptions
- ⚠️ BLOCKED on quizme-v3 resolution of shared-postgres documentation/implementation gaps

- **Success**: ~60 template files exist under `api/cryptosuite-registry/templates/` with
  `__PLACEHOLDER__` in ALL paths that vary by PS-ID, product, or suite. Secrets templates
  include `BASE64_CHAR43` placeholders. Standalone configs split into framework/domain.
  No `.go` files in `api/cryptosuite-registry/`. All actual deployment files match decisions.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 2: Rewrite Template Linter (4h) [Status: ☐ TODO]

**Objective**: Rewrite `template_drift.go` and `checks.go` to implement the runtime OS-walk +
in-memory FS comparison approach. Remove the `//go:embed` implementation entirely.

**2A — Core engine rewrite** (`template_drift.go`):
- Remove `//go:embed templates/*`, `var templatesFS embed.FS`, `instantiate()` function
- Remove all `Check*` per-file check functions from `checks.go`
- Add `LoadTemplatesDir(projectRoot string) (map[string]string, error)` — `os.WalkDir` over
  `api/cryptosuite-registry/templates/`, returns map of template relative path → raw content
- Add `BuildExpectedFS(templates map[string]string, registry *Registry) (map[string]string, error)` — for each template path detect expansion key and iterate all values:
  - `__PS_ID__` in path: loop all 10 PS-IDs from registry; substitute in path + content via `buildParams(psID)`
  - `__PRODUCT__` in path: loop all 5 products from registry; substitute in path + content via `buildProductParams(product, registry)` which includes `__PRODUCT_INCLUDE_LIST__`
  - `__SUITE__` in path: loop suite name(s) from registry (currently 1: `cryptoutil`); substitute via `buildSuiteParams(registry)`
  - `__INFRA_TOOL__` in path: loop infra tools from registry; substitute accordingly
  - Returns fully expanded map of (file relative path from project root) → (expected content)
- Add `CompareExpectedFS(expected map[string]string, projectRoot string) error` — for each expected
  file, read actual file, compare; collect all diffs; return aggregated error
- KEEP: `buildParams(psID string) map[string]string`, `normalizeCommentAlignment`, `normalizeLineEndings`
- ADD: `buildProductParams(product string, registry *Registry) map[string]string` — product-level params including `__PRODUCT_INCLUDE_LIST__`
- ADD: `buildSuiteParams(registry *Registry) map[string]string` — suite-level params

**2B — Single comprehensive linter** (`checks.go`):
- Replace all individual `Check*` functions with ONE: `CheckTemplateCompliance(logger *Logger) error`
- `CheckTemplateCompliance` calls `LoadTemplatesDir → BuildExpectedFS → CompareExpectedFS`
- Seam injection: `type templateComplianceFn func(projectRoot string) (map[string]string, error)`;
  `CheckTemplateCompliance` accepts this as parameter for testing

**2C — Registry update** (`lint-fitness-registry.yaml`):
- Remove all individual template linter entries (`template-dockerfile`, `template-compose`, etc.)
- Add single entry: `template-compliance` → wired to `CheckTemplateCompliance`

**2D — Delete old templates directory**:
- `git rm -r internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/`
- Verify `go build ./...` succeeds and `go run ./cmd/cicd-lint lint-fitness` passes

**2E — Tests** (≥98% coverage, seam injection pattern):
- Happy path: build expected FS from small test template dir, compare against temp dir with matching files
- Drift detected: one file has wrong content → error with diff
- Missing file: expected file does not exist on disk → error
- Extra file on disk: no match in expected FS → allowed (one-directional check)
- `__PS_ID__` expansion: verify 10 expansions from 1 template file
- `LoadTemplatesDir` error paths: non-existent root, unreadable file
- `BASE64_CHAR43` placeholder filtering: verify random content differences are excluded from comparison

**2F — Secrets compliance sublinter** (Decision 14):
- New lint-fitness sublinter validates content patterns inside secret files
- Validates prefix format, base64 length, naming conventions per tier
- Enforces uniqueness across unseal shards and secret files

- **Success**: `go run ./cmd/cicd-lint lint-fitness` passes; ≥98% coverage; per-file checks deleted;
  `LoadTemplatesDir` discovers all template files; secrets compliance sublinter validates patterns.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: Update Documentation (1.5h) [Status: ☐ TODO]

**Objective**: Ensure docs accurately describe the new template architecture.

- `deployment-templates.md`: Replace embedded template code blocks (Sections B.1, C.1, D.1-5, E, G.1, I, J)
  with references to `api/cryptosuite-registry/templates/{path}`. Keep all rule tables.
- `deployment-templates.md` Section O.2: Update template file catalog to list all 15 files with
  their parameterized paths and expansion behavior.
- `target-structure.md`: Add `api/cryptosuite-registry/templates/` directory listing.
- Note v9 implementation error in both v9 `plan.md` and `tasks.md`.
- Run `go run ./cmd/cicd-lint lint-docs` — passes.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 4: Quality Gates (1h) [Status: ☐ TODO]

**Objective**: Validate all quality gates pass end-to-end.

- `go build ./...` and `go build -tags e2e,integration ./...` — clean
- `golangci-lint run` and `golangci-lint run --build-tags e2e,integration` — zero warnings
- `go test ./...` — 100% pass, zero skips
- `go test -race -count=2 ./internal/apps/tools/...` — race-free
- Coverage ≥98%: `go test -coverprofile=... ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...`
- `go run ./cmd/cicd-lint lint-fitness` — passes (template-compliance and all other linters)
- `go run ./cmd/cicd-lint lint-deployments` — passes
- `go run ./cmd/cicd-lint lint-docs` — passes

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 5: Knowledge Propagation (0.5h) [Status: ☐ TODO]

**Objective**: Apply lessons learned to permanent artifacts. NEVER skip.

- Update `ENG-HANDBOOK.md` Section 9.11.1 (fitness linter catalog) — remove per-file entries, add
  `template-compliance`; update Section 13.6 to describe parameterized directory + in-memory FS approach
- Update relevant instruction files if new patterns were discovered
- Run `go run ./cmd/cicd-lint lint-docs validate-propagation` — passes
- Commit all updates with separate semantic commits per artifact type

**Post-Mortem**: After quality gates pass, update lessons.md (final phase).

---

## Architecture Decisions

### Decision 1: Runtime OS Walk — No Go Package, No Embed

**Options**:
- A: Keep `//go:embed` in `template_drift` package (current wrong state)
- B: Create Go package at `api/cryptosuite-registry/` exporting `embed.FS` (still wrong — Go code in api/)
- C: Use `os.WalkDir` at runtime to read templates from `api/cryptosuite-registry/templates/` ✓ **SELECTED**
- D: Hard-code expected content in Go test cases (no template files at all)

**Decision**: Option C — cicd-lint reads template files from disk at runtime using `os.WalkDir`.

**Rationale**: `api/cryptosuite-registry/` is the canonical location for machine-readable
API/registry data. It should contain only YAML, JSON, and configuration files — not Go source.
Runtime reading is appropriate for a linter tool that is always run from the project root;
the templates are part of the project's source tree, readable like any other config file.
This approach keeps the `api/` directory free of Go code and makes templates inspectable
and editable without any build-system knowledge.

**Impact**: `template_drift.go` uses `os.WalkDir`/`os.ReadFile` instead of `embed.FS.ReadFile`.
No new Go imports or packages. Tests use temp directories with sample template files.

### Decision 2: `__PRODUCT__` in Path — Single Template File with Per-Product Params

**Options**:
- A: Per-product files with static literal paths (`deployments/sm/compose.yml` etc.) ✗ NOT parameterized
- B: Single `deployments/__PRODUCT__/compose.yml` with per-product params from registry ✓ **SELECTED**

**Decision**: Option B — single `deployments/__PRODUCT__/compose.yml` template with `__PRODUCT__`
in the directory path, expanded × 5 using per-product param sets derived from registry.yaml.

**Rationale**: ALL variable directory names MUST use `__PLACEHOLDER__` syntax.

**Impact**: 1 product compose template file at `templates/deployments/__PRODUCT__/compose.yml`.

### Decision 3: Suite Files Use `__SUITE__` in Path (Parameterized, Not Literal)

**Options**:
- A: Suite template files at static literal path `deployments/cryptoutil/...` ✗ NOT parameterized
- B: Suite template files at `deployments/__SUITE__/...`; expanded × 1 (currently `cryptoutil`) ✓ **SELECTED**

**Decision**: Option B — suite template files stored at `deployments/__SUITE__/Dockerfile` and
`deployments/__SUITE__/compose.yml`. The path contains `__SUITE__` and expands to
`deployments/cryptoutil/...` using the suite name from registry.

**Impact**: Suite template directory is `deployments/__SUITE__/` (2 files).

### Decision 4: Three-Level pki-init with --domain (Q1 → E)

**Source**: quizme-v1 Q1 — Product Compose `pki-init` Override

**Decision**: There are 3 compose.yml templates (SUITE, PRODUCT, PS-ID). Each MUST include
a `pki-init` service with `--domain=<LEVEL_NAME>`:
- PS-ID template: `pki-init` with `--domain=__PS_ID__`
- PRODUCT template: `pki-init` with `--domain=__PRODUCT__`
- SUITE template: `pki-init` with `--domain=__SUITE__`

**Current State**:
- Suite compose: `pki-init` with `--domain=cryptoutil` ✓
- SM product compose: `pki-init` with `--domain=sm` ✓ (only product with it)
- Jose/pki/identity/skeleton products: NO `pki-init` ✗ (need to add)
- PS-ID composes: `pki-init` with NO `--domain` flag ✗ (need to add `--domain=<PS-ID>`)

**Impact**: Product template includes `pki-init`. All 5 actual product compose files must have
`pki-init` with `--domain=<product>`. PS-ID template updated to add `--domain=__PS_ID__`.
Actual PS-ID compose files must add `--domain=<ps-id>` to `pki-init` commands.

**Resolved (quizme-v2 Q6)**: pki-init is framework code — every PS-ID binary bundles
pki-init capability. All compose levels override pki-init command with `--domain=<level>`.

**⚠️ CRITICAL ISSUE for quizme-v3**: Deep analysis discovered that the Go implementation
(`internal/apps/framework/tls/init.go`) expects **2 positional args** (`<tier-id> <target-dir>`)
with NO flag parsing (no pflag, no cobra). But all compose files pass `["init", "--output-dir=/certs"]`
which is a **single** arg. This is a runtime bug — pki-init containers will ALWAYS fail with
`"pki-init: usage: pki-init <tier-id> <target-dir>"`. See quizme-v3.

### Decision 5: Remove `image:` from Product Service Overrides (Q2 → B)

**Source**: quizme-v1 Q2 — Product Compose `image:` References

**Decision**: Option B — remove `image:` from `sm` product compose service overrides to match
the other 4 products. PS-ID-level compose files are the single source of the image reference.
Product-level overrides should NOT duplicate it.

**Current State**:
- SM product: specifies `image: cryptoutil-sm-*:dev` on every service override ✗ (remove)
- Jose/pki/identity/skeleton products: correctly omit `image:` ✓

**Impact**: Fix `deployments/sm/compose.yml` to remove `image:` from all service override lines.
Product template uses NO `image:` in service port overrides.

### Decision 6: All 4 PostgreSQL Secrets at All Levels (Q3 → E)

**Source**: quizme-v1 Q3 — Product Compose PostgreSQL Secrets

**Decision**: All SUITE, PRODUCT, and PS-ID compose files MUST include all 4 PostgreSQL secrets:
`postgres-url.secret`, `postgres-username.secret`, `postgres-password.secret`,
`postgres-database.secret`.

**Current State**:
- Suite compose: all 4 ✓
- SM product: all 4 ✓
- Jose/pki/identity/skeleton products: only `postgres-url.secret` ✗ (add 3 more)
- PS-ID composes: all 4 (via shared-postgres include) ✓

**Impact**: Add `postgres-username.secret`, `postgres-password.secret`, `postgres-database.secret`
to jose, pki, identity, skeleton product compose secrets sections. All templates use 4 postgres
secrets uniformly.

### Decision 7: Standalone Config Framework/Domain Split (Q4 → E, resolved by quizme-v2 Q1)

**Source**: quizme-v1 Q4 → quizme-v2 Q1

**Decision**: Split each PS-ID standalone config into TWO files:
- `configs/<ps-id>/<ps-id>-framework.yml` — shared framework settings (templateable)
- `configs/<ps-id>/<ps-id>-domain.yml` — PS-ID-specific domain settings (NOT templateable)

**Framework settings** (in template): `bind-*`, `tls-*`, `cors-*`, `otlp-*`, `log-level`,
`database-url`, `browser-session-algorithm`, `service-session-algorithm`,
`enable-dynamic-registration`.

**Domain settings** (NOT in template): `issuer`, `token-lifetime`, `refresh-token-lifetime`,
`authorization-code-ttl`, `enable-discovery`, `ca:`, `storage:`, `revocation:`, `tsa:`,
`est:`, `profiles:`, and other PS-ID-specific settings.

**Breaking change**: Old single-file `<ps-id>.yml` configs replaced by two-file split.
No backward compatibility required.

**Impact**: Template file `__PS_ID__-framework.yml` replaces `__PS_ID__.yml` in
`templates/configs/__PS_ID__/`. Full-file comparison (not prefix-only). Domain files are
per-PS-ID and NOT compared against any template. Go app config loading uses multiple
`--config=` flags (already supported via pflag StringSlice + viper MergeInConfig).

**⚠️ OPEN QUESTION for quizme-v3**: Exact classification of edge-case settings
(pki-ca `storage.type` vs `database-url`, identity-rp `authz-server-url`/`client-id`,
identity-spa `static-files-path`). See quizme-v3.

### Decision 8: Config Hierarchy via Environment Variables (Q5 → E, resolved by quizme-v2 Q2)

**Source**: quizme-v1 Q5 → quizme-v2 Q2

**Decision**: Config hierarchy achieved via `SUITE_ARGS` and `PRODUCT_ARGS` environment
variables. PS-ID compose command becomes:
`command: ["server", "--config=/app/config/<ps-id>-app-*.yml", "$SUITE_ARGS", "$PRODUCT_ARGS"]`

Product compose defines `PRODUCT_ARGS` env var with product-level config path.
Suite compose defines `SUITE_ARGS` env var with suite-level config path.

**Breaking change**: PS-ID compose `command:` modified. No backward compatibility required.

**Impact**: v10 scope. Requires creating product-level and suite-level config files, modifying
PS-ID compose templates to include `$SUITE_ARGS`/`$PRODUCT_ARGS` in command, modifying
product/suite compose templates to set the env vars.

**⚠️ CRITICAL ISSUE for quizme-v3**: Docker Compose `command:` uses exec form (JSON array),
NOT shell form. `$SUITE_ARGS` in a JSON array is NOT expanded by a shell — it is passed as
a literal string `"$SUITE_ARGS"`. This requires either: (a) shell-form command
(`command: /bin/sh -c "exec /app/<ps-id> server --config=... $SUITE_ARGS $PRODUCT_ARGS"`)
or (b) Docker Compose `environment:` mechanism with entrypoint script. See quizme-v3.

### Decision 9: Domain Directory Pattern for PS-ID-Specific Files (Q6/Q7 → E)

**Source**: quizme-v1 Q6 (pki-ca profiles) and Q7 (identity-authz policies)

**Decision**: Templates include an empty `domain/` subdirectory placeholder. Each PS-ID
instance is free to add domain-specific, non-parameterized files in its own `domain/` directory.
The template linter only compares files that exist in the template tree; extra files in the
actual `domain/` directories are allowed (one-directional check).

**Concrete Application**:
- `configs/pki-ca/profiles/` (25 files) — not compared against any template; allowed as extra files
- `configs/identity-authz/domain/policies/` (3 files) — same treatment

**Impact**: No template files created for domain-specific content. `CompareExpectedFS` is
one-directional: every expected file must exist and match; extra actual files are silently allowed.
No exclusion mechanism needed — the engine design naturally handles this.

### Decision 10: shared-postgres Full Template with 30 Databases (Q8 → E, resolved by quizme-v2 Q3)

**Source**: quizme-v1 Q8 → quizme-v2 Q3

**Decision**: Full template for all shared-postgres files with `__SUITE__` substitution.
shared-postgres creates 30 logical databases: 3 deployment levels (SUITE, PRODUCT, PS-ID) ×
10 PS-IDs. Leader postgres has 30 databases; follower postgres has 1 database with 30 schemas.

Each deployment level uses DIFFERENT credentials, unseal secrets, and hash peppers to enforce
authentication and cryptographic isolation between levels.

**Template Files Created**:
- `templates/deployments/shared-postgres/compose.yml` — with `__SUITE__` substitution
- `templates/deployments/shared-postgres/init-leader-databases.sql` — generated from registry
  using `__PS_ID_DATABASE_LIST__` (30 databases = 3 levels × 10 PS-IDs)
- `templates/deployments/shared-postgres/init-follower-databases.sql` — follower schema setup
- `templates/deployments/shared-postgres/setup-logical-replication.sh` — replication subscriptions

**⚠️ CRITICAL ISSUE for quizme-v3**: Deep analysis revealed SERIOUS documentation gaps in
ENG-HANDBOOK.md Section 3.4.2 and IMPLEMENTATION gaps in actual shared-postgres files:

1. DDL/DML user separation documented but NOT implemented (single admin user for all 30 DBs)
2. `postgresql.conf` referenced in handbook but does NOT exist in deployments/
3. `setup-logical-replication.sh` only implements pki-ca subscription (1 of 10)
4. Follower creates 16 databases, not 30 schemas
5. `postgres-url.secret` in PS-ID secrets dirs references stale per-PS-ID hostnames

These gaps are BLOCKING for accurate template creation. See quizme-v3.

### Decision 11: shared-telemetry Template Scope (Q9 → A)

**Source**: quizme-v1 Q9 — Shared Telemetry Template Scope

**Decision**: Option A — compose.yml and otel-collector-config.yaml get templates with `__SUITE__`
substitution. Grafana JSON dashboards, provisioning YAML, and alert rules are static (too
complex/fragile to template).

**Template Files Created**:
- `templates/deployments/shared-telemetry/compose.yml` — with `__SUITE__` substitution
- `templates/deployments/shared-telemetry/otel/otel-collector-config.yaml` — with `__SUITE__` substitution

**Excluded from Template Compliance** (extra files, allowed by one-directional check):
- `alerts/cryptoutil.yml`
- `grafana-otel-lgtm/dashboards/*.json`
- `grafana-otel-lgtm/provisioning/**`
- `otel/cryptoutil-otel.yml`

**Impact**: 2 new template files for shared-telemetry. Static paths (no `__INFRA_TOOL__` expansion
for now). Phase 1 adds these files.

### Decision 12: Registry Expansion — shared_services + infra_tools (Q10 → E, resolved by quizme-v2 Q4)

**Source**: quizme-v1 Q10 → quizme-v2 Q4

**Decision**: Add BOTH `shared_services:` and `infra_tools:` sections to registry.yaml.
ALL parts of the repository MUST eventually be modeled in the registry for enforcing
parameterized linting for maximum consistency and anti-slop.

**shared_services**: `shared-postgres`, `shared-telemetry` — deployment infrastructure
with compose files, related to the suite.

**infra_tools**: `cicd-lint`, `cicd-workflow` — build/CI tools, NOT deployment artifacts
but still critical repository components.

**Impact**: registry.yaml schema expansion. Template expansion uses `shared_services` section
for shared-postgres/shared-telemetry template expansion (if parameterized paths warranted).
infra_tools modeled for completeness and future linting/validation.

### Decision 13: Config File PS-ID Prefix + Framework/Domain Suffix (quizme-v2 Q5 → E)

**Source**: quizme-v2 Q5

**Decision**: Config filenames use PS-ID prefix (matches actual files on disk) PLUS
framework/domain suffix per Decision 7:
- Deployment configs: `__PS_ID__-app-common.yml` (unchanged — deployment configs are framework)
- Standalone configs: `__PS_ID__-framework.yml` + `__PS_ID__-domain.yml` (replaces `__PS_ID__.yml`)

**Impact**: Template for standalone configs changes from `__PS_ID__.yml` to
`__PS_ID__-framework.yml`. Domain file is per-PS-ID, not templateable.

### Decision 14: Secrets Directory Template with BASE64_CHAR43 (quizme-v2 Q7 → E)

**Source**: quizme-v2 Q7

**Decision**: Full secrets directory template with different structures per deployment tier.
Template files include a `BASE64_CHAR43` placeholder for the random content portion of
secret files. During comparison, differences in `BASE64_CHAR43` placeholder positions are
filtered out (expected and required to differ).

**Template Structures Per Tier**:
- **PS-ID**: 14 files — 5 unseal, hash-pepper, 4 postgres, 2 browser, 2 service (all `.secret`)
- **Product**: 14 files — 5 unseal, hash-pepper, 4 postgres (`.secret`); 4 browser/service (`.secret.never`)
- **Suite**: 14 files — same as product (`.secret` for crypto/postgres; `.secret.never` for browser/service)

**Complementary lint-fitness sublinter**: Validates content patterns inside secret files
(prefix format, base64 length) AND enforces uniqueness across unseal shards and secret files.

**Impact**: Significant scope expansion. Adds ~42 template files (14 per tier × 3 tiers)
plus shared-postgres secrets. Adds a new lint-fitness sublinter (`secrets-compliance` or similar).

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Actual deployment files don't match new templates | Medium | Medium | Fix files (not templates) during Phase 2; templates ARE the spec |
| `buildParams()` missing suite-specific params | Low | Low | Add `buildSuiteParams()` helper for suite/product file expansion |
| Test coverage drops below 98% after embed removal | Low | Medium | Run coverage immediately after each Phase 2 step |
| `lint-fitness-registry.yaml` wiring for new single linter | Low | Low | Verify registration and dispatch before marking Phase 2 complete |
| `deployment-templates.md` lint-docs failures after edit | Low | Low | Run `cicd-lint lint-docs` after every doc section edit |
| Product compose files have undocumented content | Low | Low | Read actual files and reverse-engineer params before writing templates |
| **pki-init code/compose mismatch** | **HIGH** | **HIGH** | Go code expects positional args; compose uses flags. Runtime bug. See quizme-v3 |
| **Docker Compose env var expansion** | **HIGH** | **HIGH** | `$SUITE_ARGS` in exec-form command is NOT shell-expanded. See quizme-v3 |
| **shared-postgres documentation gaps** | **HIGH** | **HIGH** | ENG-HANDBOOK §3.4.2 has 5 doc/impl gaps. See quizme-v3 |
| **Scope explosion (14→~60 templates)** | **Medium** | **HIGH** | Phase restructuring needed; may defer secrets/shared-postgres to later phases |
| **Config classification edge cases** | **Medium** | **Medium** | pki-ca uses `storage.type` not `database-url`; identity-rp has unique settings. See quizme-v3 |

---

## Quality Gates - MANDATORY

**Per-Task Gates**:
- ✅ `go build ./...` clean (no errors)
- ✅ `go test ./...` passes (100%, zero skips)
- ✅ `golangci-lint run` zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ `template_drift` package: ≥98% line coverage
- ✅ `api/cryptosuite-registry/`: plain non-Go directory (no coverage applicable)

**Mutation Testing** (Phase 4):
- ✅ `template_drift` package: ≥98% mutation efficacy (infrastructure tool)

**Per-Phase Gates**:
- ✅ `go run ./cmd/cicd-lint lint-fitness` passes after each phase
- ✅ `go run ./cmd/cicd-lint lint-deployments` passes
- ✅ `go run ./cmd/cicd-lint lint-docs` passes

**ENG-HANDBOOK.md Cross-References**:

| Topic | Section |
|-------|---------|
| Testing Architecture | [Section 10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) |
| Coverage Targets | [Section 10.2.3](../../docs/ENG-HANDBOOK.md#1023-coverage-targets) |
| Test Seam Injection | [Section 10.2.4](../../docs/ENG-HANDBOOK.md#1024-test-seam-injection-pattern) |
| Integration Testing | [Section 10.3](../../docs/ENG-HANDBOOK.md#103-integration-testing-strategy) |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) |
| Code Quality | [Section 11.3](../../docs/ENG-HANDBOOK.md#113-code-quality-standards) |
| Coding Standards | [Section 14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) |
| Version Control | [Section 14.2](../../docs/ENG-HANDBOOK.md#142-version-control) |
| Deployment Architecture | [Section 12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) |
| API Architecture (api/ dir) | [Section 8](../../docs/ENG-HANDBOOK.md#8-api-architecture) |
| Fitness Linter Catalog | [Section 9.11.1](../../docs/ENG-HANDBOOK.md#9111-fitness-sub-linter-catalog) |
| Template Enforcement | [Section 13.6](../../docs/ENG-HANDBOOK.md#136-template-enforcement--drift-detection) |
| Plan Lifecycle | [Section 14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) |
| Post-Mortem | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) |

---

## Success Criteria

- [ ] `api/cryptosuite-registry/templates/` contains ~60 parameterized template files (NO `.go` files)
- [ ] Template directory structure mirrors `./deployments/` and `./configs/` with `__KEY__` in paths
- [ ] Secrets templates use `BASE64_CHAR43` placeholders for random content
- [ ] Standalone configs split into `<ps-id>-framework.yml` + `<ps-id>-domain.yml`
- [ ] shared-postgres fully templated with 30-database model from registry
- [ ] `internal/.../template_drift/templates/` directory deleted
- [ ] `template_drift.go` uses `os.WalkDir` (no `//go:embed`, no `embed.FS`)
- [ ] Single `CheckTemplateCompliance` linter replaces all per-file check functions
- [ ] Secrets compliance sublinter validates content patterns and uniqueness
- [ ] `go run ./cmd/cicd-lint lint-fitness` template-compliance check passes against all actual files
- [ ] All quality gates passing (build, lint, test, ≥98% coverage, race-free)
- [ ] Documentation updated to reference `api/cryptosuite-registry/templates/`
- [ ] Evidence archived in `test-output/framework-v10/`
