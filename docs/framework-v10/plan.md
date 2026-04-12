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

**Example — Product level (static path, content parameterized)**:
```
templates/deployments/sm/compose.yml            → deployments/sm/compose.yml (compare directly)
templates/deployments/identity/compose.yml      → deployments/identity/compose.yml
```

**Example — Suite level (static path)**:
```
templates/deployments/cryptoutil/Dockerfile     → deployments/cryptoutil/Dockerfile
templates/deployments/cryptoutil/compose.yml    → deployments/cryptoutil/compose.yml
```

### How cicd-lint Processes Templates

1. Walk all files under `api/cryptosuite-registry/templates/`
2. For each template file, inspect its path for expansion keys:
   - Path contains `__PS_ID__`: expand × 10 (one per PS-ID from registry.yaml)
   - Path contains `__PRODUCT__`: expand × 5 (sm, jose, pki, identity, skeleton); uses per-product
     params derived from registry.yaml (e.g. `__PRODUCT_INCLUDE_LIST__` built from each product's PS-IDs)
   - Path contains `__SUITE__`: expand × 1 (currently `cryptoutil`); parameterized for future renames
   - Path contains `__INFRA_TOOL__`: expand for each infrastructure tool (shared-postgres, shared-telemetry, …)
   - No expansion key in path: compare directly
3. For each expansion: substitute ALL `__KEY__` params in both the resolved path and file content
4. Collect all (resolvedPath → expectedContent) pairs → in-memory expected filesystem
5. Recursively compare in-memory FS against actual `./configs/` and `./deployments/` on disk:
   every expected file must exist at exactly the resolved relative path with identical content

### Complete Template Directory Structure

```
api/cryptosuite-registry/templates/
  deployments/
    __PS_ID__/                          ← expands for each of 10 PS-IDs
      Dockerfile                        ← __PS_ID__ + build/label/healthcheck params
      compose.yml                       ← __PS_ID__ + port params
      config/
        config-common.yml               ← __PS_ID__ + shared config params
        config-sqlite-1.yml             ← __PS_ID__ + __SERVICE_APP_PORT_SQLITE_1__
        config-sqlite-2.yml             ← __PS_ID__ + __SERVICE_APP_PORT_SQLITE_2__
        config-postgresql-1.yml         ← __PS_ID__ + __SERVICE_APP_PORT_PG_1__
        config-postgresql-2.yml         ← __PS_ID__ + __SERVICE_APP_PORT_PG_2__
    __PRODUCT__/                        ← expands × 5 (sm, jose, pki, identity, skeleton)
      compose.yml                       ← uses __PRODUCT__ + per-product params from registry
                                          (e.g. __PRODUCT_INCLUDE_LIST__ for each product's PS-ID includes)
    __SUITE__/                          ← expands × 1 (cryptoutil); parameterized for future renames
      Dockerfile                        ← 4-stage build pattern; uses __SUITE__ + suite-level display params
      compose.yml                       ← uses __SUITE__ + all product/PS-ID references
    __INFRA_TOOL__/                     ← expands for each infra tool (shared-postgres, shared-telemetry, …)
      compose.yml
  configs/
    __PS_ID__/                          ← expands for each of 10 PS-IDs
      __PS_ID__.yml                     ← both dirname and filename are parameterized
```

**Template file count**: 12 physical files (8 under `deployments/__PS_ID__/config/ + Dockerfile + compose.yml`,
1 under `deployments/__PRODUCT__/`, 2 under `deployments/__SUITE__/`, 1 under `configs/__PS_ID__/`).
Infra-tool templates added as needed during implementation.
**Expected file count after expansion**: 80 PS-ID deployment files + 10 PS-ID config files +
5 product compose files + 2 suite files = 97+ (plus infra-tool expansions).

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
  - `api/cryptosuite-registry/templates/` — 15 parameterized template files (see Structure above)
- **Documentation**:
  - `docs/deployment-templates.md` — update to reference canonical template files
  - `docs/target-structure.md` — add `api/cryptosuite-registry/templates/` listing

---

## Phases

### Phase 1: Create Canonical Template Directory (3h) [Status: ☐ TODO]

**Objective**: Create `api/cryptosuite-registry/templates/` with all 15 parameterized template
files. These are PLAIN FILES — not `.go`, not embedded. Structure mirrors `./deployments/` and
`./configs/` with `__KEY__` in both paths and content.

**1A — PS-ID level templates** (7 files, each expands to 10):
- `deployments/__PS_ID__/Dockerfile` — based on current `Dockerfile.tmpl` content (same params)
- `deployments/__PS_ID__/compose.yml` — based on current `compose.yml.tmpl` content
- `deployments/__PS_ID__/config/config-common.yml` — from `config-common.yml.tmpl`
- `deployments/__PS_ID__/config/config-sqlite-1.yml` — from `config-sqlite.yml.tmpl` + instance-1 params
- `deployments/__PS_ID__/config/config-sqlite-2.yml` — from `config-sqlite.yml.tmpl` + instance-2 params
- `deployments/__PS_ID__/config/config-postgresql-1.yml` — from `config-postgresql.yml.tmpl` + instance-1
- `deployments/__PS_ID__/config/config-postgresql-2.yml` — from `config-postgresql.yml.tmpl` + instance-2

**1B — PS-ID standalone config** (1 file, expands to 10):
- `configs/__PS_ID__/__PS_ID__.yml` — based on `standalone-config.yml.tmpl`

**1C — Product compose file** (1 physical file, expands × 5):
- `deployments/__PRODUCT__/compose.yml` — template content uses `__PRODUCT__`, `__SUITE__`, `__IMAGE_TAG__`,
  and `__PRODUCT_INCLUDE_LIST__` (multi-line include entries generated from registry per product).
  Phase 1 must verify whether actual product compose files are structurally uniform enough for one template;
  if not, product-specific sidecars in registry.yaml define the varying content.

**1D — Suite files** (2 files, `__SUITE__` in path, expands × 1):
- `deployments/__SUITE__/Dockerfile` — 4-stage build pattern with `__SUITE__` params
- `deployments/__SUITE__/compose.yml` — based on actual suite compose with `__SUITE__` params

- **Success**: 12+ files exist under `api/cryptosuite-registry/templates/` with `__PLACEHOLDER__` in ALL
  paths that vary by PS-ID, product, suite, or infra tool. Manually verify a sample expansion
  (e.g., sm-kms Dockerfile) matches actual `deployments/sm-kms/Dockerfile` with params filled in.
  No `.go` files in `api/cryptosuite-registry/`.

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

- **Success**: `go run ./cmd/cicd-lint lint-fitness` passes; ≥98% coverage; per-file checks deleted.

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
- E:

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
- C: Generate expected product compose content entirely from registry.yaml (no template file)
- D: Skip product/suite template linting
- E:

**Decision**: Option B — single `deployments/__PRODUCT__/compose.yml` template with `__PRODUCT__`
in the directory path, expanded × 5 using per-product param sets derived from registry.yaml.

**Rationale**: ALL variable directory names MUST use `__PLACEHOLDER__` syntax — including product
names. Hardcoding `sm`, `jose`, etc. in template paths defeats the parameterization requirement.
Product compose files differ in their service include lists, but this variation is capturable via
a `__PRODUCT_INCLUDE_LIST__` param computed from registry.yaml (each product's PS-IDs generate
multi-line include entries). Phase 1 implementation must verify structural uniformity and define
product-specific params; if product compose files have additional non-parameterizable differences,
add those values to registry.yaml as product-level metadata.

**Impact**: 1 product compose template file at `templates/deployments/__PRODUCT__/compose.yml`.
Registry must supply per-product params during expansion.

### Decision 3: Suite Files Use `__SUITE__` in Path (Parameterized, Not Literal)

**Options**:
- A: Suite template files at static literal path `deployments/cryptoutil/...` ✗ NOT parameterized
- B: Suite template files at `deployments/__SUITE__/...`; expanded × 1 (currently `cryptoutil`) ✓ **SELECTED**
- C: Suite Dockerfile derived at runtime from PS-ID template with `__PS_ID__`=`__SUITE__`
- D: No suite Dockerfile template (manual validation only)
- E:

**Decision**: Option B — suite template files stored at `deployments/__SUITE__/Dockerfile` and
`deployments/__SUITE__/compose.yml`. The path contains `__SUITE__` and expands to
`deployments/cryptoutil/...` using the suite name from registry. Template content uses `__SUITE__`
and other suite-level params throughout.

**Rationale**: ALL variable directory names MUST use `__PLACEHOLDER__` syntax. Even though there
is currently only one suite (`cryptoutil`), hardcoding it violates the parameterization requirement
and makes future renames (if any) require template edits. `__SUITE__` in the path costs nothing
and ensures model consistency with PS-ID and product template paths.

**Impact**: Suite template directory is `deployments/__SUITE__/` (2 files). BuildExpectedFS expands
`__SUITE__` using the suite name from registry (currently `cryptoutil`).

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

- [ ] `api/cryptosuite-registry/templates/` contains 15 parameterized template files (NO `.go` files)
- [ ] Template directory structure mirrors `./deployments/` and `./configs/` with `__KEY__` in paths
- [ ] `internal/.../template_drift/templates/` directory deleted
- [ ] `template_drift.go` uses `os.WalkDir` (no `//go:embed`, no `embed.FS`)
- [ ] Single `CheckTemplateCompliance` linter replaces all per-file check functions
- [ ] `go run ./cmd/cicd-lint lint-fitness` template-compliance check passes against all actual files
- [ ] All quality gates passing (build, lint, test, ≥98% coverage, race-free)
- [ ] Documentation updated to reference `api/cryptosuite-registry/templates/`
- [ ] Evidence archived in `test-output/framework-v10/`
