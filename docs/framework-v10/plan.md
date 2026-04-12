# Implementation Plan — Framework v10: Canonical Template Relocation

**Status**: Planning
**Created**: 2026-04-12
**Last Updated**: 2026-04-12
**Purpose**: Correct the v9 architecture failure where canonical deployment templates were placed
in the wrong location (`internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/`)
instead of the mandated canonical location (`api/cryptosuite-registry/templates/`). Complete the
missing product and suite templates. Update documentation to reference canonical files instead
of embedding shorthand template content.

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

Framework v9 Task 8.1 claimed to store canonical templates at `api/cryptosuite-registry/templates/`
but the implementation put them at `internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/`.
This violated the explicit architecture decision in:
- `docs/deployment-templates.md` Section N.1 ("All canonical templates are stored as parameterized
  template files in `api/cryptosuite-registry/templates/`")
- `docs/deployment-templates.md` Section O.2 (canonical template file catalog)
- v9 plan Item 23 (explicit architecture decision — `api/cryptosuite-registry/templates/`)

Framework v10 corrects this by:
1. Creating a Go package at `api/cryptosuite-registry/` that exports `TemplatesFS embed.FS`
2. Moving all 6 existing template files to `api/cryptosuite-registry/templates/`
3. Adding the 7 missing templates (5 per-product compose, 1 suite compose, 1 suite Dockerfile)
4. Updating the `template_drift` linter to use the canonical FS via import (not own embed)
5. Registering new product/suite template-comparison linters in `lint-fitness`
6. Updating documentation to reference canonical files instead of embedding shorthand content

---

## Background

### Why v9 Got This Wrong

v9 Task 8.1 implementation notes say: "Templates stored in `internal/apps/tools/cicd_lint/
lint_fitness/template_drift/templates/` using `__KEY__` placeholder format." This contradicts
the architecture decision in v9 plan Item 23 (stored in `api/cryptosuite-registry/templates/`).
The developer chose a simpler path (same-package embed) but violated the canonical location
requirement. v10 corrects this.

### Go `//go:embed` Constraint

Go's `//go:embed` directive can only reference paths within the package's own directory tree
(no `../` parent directory traversal). To have templates both in `api/cryptosuite-registry/templates/`
AND embedded in the linter binary, we need a Go package at `api/cryptosuite-registry/` that:
1. Contains the `templates/` subdirectory  
2. Has a `.go` file with `//go:embed templates/*` that exports `TemplatesFS embed.FS`
3. Is imported by `template_drift` (which then removes its own embed directive)

### Current Template Inventory

**Exists (wrong location)**:

| File | Current Location | Target Location |
|------|-----------------|----------------|
| `Dockerfile.tmpl` | `internal/.../template_drift/templates/` | `api/cryptosuite-registry/templates/` |
| `compose.yml.tmpl` | `internal/.../template_drift/templates/` | `api/cryptosuite-registry/templates/` |
| `config-common.yml.tmpl` | `internal/.../template_drift/templates/` | `api/cryptosuite-registry/templates/` |
| `config-sqlite.yml.tmpl` | `internal/.../template_drift/templates/` | `api/cryptosuite-registry/templates/` |
| `config-postgresql.yml.tmpl` | `internal/.../template_drift/templates/` | `api/cryptosuite-registry/templates/` |
| `standalone-config.yml.tmpl` | `internal/.../template_drift/templates/` | `api/cryptosuite-registry/templates/` |

**Missing (must be created)**:

| File | Template For | Based On |
|------|-------------|---------|
| `product-sm-compose.yml.tmpl` | `deployments/sm/compose.yml` | Section G of deployment-templates.md + actual file |
| `product-jose-compose.yml.tmpl` | `deployments/jose/compose.yml` | Section G |
| `product-pki-compose.yml.tmpl` | `deployments/pki/compose.yml` | Section G |
| `product-identity-compose.yml.tmpl` | `deployments/identity/compose.yml` | Section G |
| `product-skeleton-compose.yml.tmpl` | `deployments/skeleton/compose.yml` | Section G |
| `suite-compose.yml.tmpl` | `deployments/cryptoutil/compose.yml` | Section I |
| `suite-Dockerfile.tmpl` | `deployments/cryptoutil/Dockerfile` | Section J (same as B but suite binary) |

---

## Technical Context

- **Language**: Go 1.26.1
- **New Go package**: `api/cryptosuite-registry/` (package `cryptosuiteregistry`)
- **Import path**: `cryptoutil/api/cryptosuite-registry`
- **Import alias**: `cryptoutilRegistryTemplates` (per §11.1.3 convention)
- **Templates**: Use `__KEY__` placeholder syntax (double-underscore delimiters)
- **Template engine**: Existing in `template_drift.go` — minimal changes needed
- **Linter registry**: `internal/apps/tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml`
- **Related files**:
  - `api/cryptosuite-registry/registry.yaml` — parameter source
  - `api/cryptosuite-registry/registry-schema.json` — schema
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/template_drift.go` — embed + engine
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/checks.go` — per-file checks
  - `docs/deployment-templates.md` — human documentation (references canonical files)
  - `docs/target-structure.md` — directory layout spec (lists canonical files)

---

## Phases

### Phase 1: Relocate Templates to Canonical Location (3h) [Status: ☐ TODO]

**Objective**: Move 6 existing template files from `template_drift/templates/` to
`api/cryptosuite-registry/templates/`. Create the Go package that exports `TemplatesFS`.
Update `template_drift.go` to use the imported FS.

- Create `api/cryptosuite-registry/templates.go` with `package cryptosuiteregistry` and exported `TemplatesFS embed.FS`
- Copy all 6 template files to `api/cryptosuite-registry/templates/`
- Update `template_drift.go`: remove own `//go:embed templates/*` + `var templatesFS embed.FS`; import `cryptoutilRegistryTemplates`; replace all `templatesFS.ReadFile(...)` calls with `cryptoutilRegistryTemplates.TemplatesFS.ReadFile(...)`
- Delete `internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/` directory
- Update tests (seam injection already handles FS — verify tests still pass)
- **Success**: `go build ./...` clean; `go test ./...` passes; actual `template_drift/templates/` directory deleted

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 2: Add Missing Templates and Linters (5h) [Status: ☐ TODO]

**Objective**: Create the 7 missing template files in `api/cryptosuite-registry/templates/`.
Implement linters for product-compose and suite-compose/Dockerfile validation.

**2A — Template Files (in `api/cryptosuite-registry/templates/`)**:
- Create `product-sm-compose.yml.tmpl` — based on actual `deployments/sm/compose.yml` + `__SUITE__`/`__IMAGE_TAG__` params
- Create `product-jose-compose.yml.tmpl` — based on `deployments/jose/compose.yml`
- Create `product-pki-compose.yml.tmpl` — based on `deployments/pki/compose.yml`
- Create `product-identity-compose.yml.tmpl` — based on `deployments/identity/compose.yml`
- Create `product-skeleton-compose.yml.tmpl` — based on `deployments/skeleton/compose.yml`
- Create `suite-compose.yml.tmpl` — based on `deployments/cryptoutil/compose.yml`
- Create `suite-Dockerfile.tmpl` — same 4-stage pattern as `Dockerfile.tmpl` but for `cryptoutil` suite binary

**2B — Linters (in `template_drift/checks.go`)**:
- Add `CheckProductCompose(logger)` — for each product, load `product-{PRODUCT}-compose.yml.tmpl`, substitute `__SUITE__`/`__IMAGE_TAG__`/`__BUILD_DATE__`, compare against `deployments/{PRODUCT}/compose.yml`
- Add `CheckSuiteCompose(logger)` — load `suite-compose.yml.tmpl`, substitute suite params, compare against `deployments/cryptoutil/compose.yml`
- Add `CheckSuiteDockerfile(logger)` — load `suite-Dockerfile.tmpl`, substitute suite params, compare against `deployments/cryptoutil/Dockerfile`
- Register all 3 new linters in `lint-fitness-registry.yaml`

**2C — Tests**:
- Add tests for all 3 new linters (seam injection pattern, ≥98% coverage)
- Verify actual product/suite files match templates (fix files if needed)

**Success**: All 13 template files exist in `api/cryptosuite-registry/templates/`; new linters pass against actual files; ≥98% coverage in `template_drift` package.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: Update Documentation (2h) [Status: ☐ TODO]

**Objective**: Update `deployment-templates.md` and `target-structure.md` to reference canonical
template files instead of embedding shorthand/incomplete template content. Docs describe the
templates; the files ARE the templates.

- `deployment-templates.md` Sections B.1, C.1, D.1-D.5, E: Replace embedded template content with a reference to `api/cryptosuite-registry/templates/{FILE}.tmpl`. Keep all rule tables (B.2, C.2, etc.), parameter tables, and rationale text.
- `deployment-templates.md` Sections G.1, I-J: Same — replace shorthand sketch templates with references to canonical files.
- `deployment-templates.md` Section O.2: Verify catalog is complete (13 files now).
- `target-structure.md` `api/cryptosuite-registry/` section: List the new `templates/` directory and all 13 template files.
- Update v9 `plan.md` to note the implementation error and v10 correction.
- Run `go run ./cmd/cicd-lint lint-docs` — all checks pass.

**Success**: Docs reference `api/cryptosuite-registry/templates/` as canonical source; no embedded template content duplicating the actual template files.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 4: Quality Gates (1h) [Status: ☐ TODO]

**Objective**: Validate all quality gates pass end-to-end.

- `go build ./...` — clean (production code)
- `go build -tags e2e,integration ./...` — clean
- `golangci-lint run` — zero warnings
- `golangci-lint run --build-tags e2e,integration` — zero warnings
- `go test ./...` — 100% pass, zero skips
- `go test -race -count=2 ./internal/apps/tools/...` — race-detector clean
- Coverage ≥98%: `go test -coverprofile=coverage.out ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...`
- `go run ./cmd/cicd-lint lint-fitness` passes (all template linters green)
- `go run ./cmd/cicd-lint lint-deployments` passes
- `go run ./cmd/cicd-lint lint-docs` passes

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 5: Knowledge Propagation (1h) [Status: ☐ TODO]

**Objective**: Apply lessons learned to permanent artifacts. NEVER skip this phase.

- Review all lessons from Phases 1-4
- Update `docs/ENG-HANDBOOK.md` Section O.2 canonical template catalog to reflect 13 files
- Update `docs/ENG-HANDBOOK.md` Section 9.11 (fitness linter catalog) for new product/suite linters
- Update instruction files if new patterns were discovered
- Run `go run ./cmd/cicd-lint lint-docs validate-propagation` — passes
- Commit all updates

**Success**: ENG-HANDBOOK.md updated; propagation intact; all commits clean.

**Post-Mortem**: After quality gates pass, update lessons.md (final phase).

---

## Architecture Decisions

### Decision 1: Single Go Package with embed.FS at api/cryptosuite-registry/

**Options**:
- A: Keep templates in `template_drift/templates/` (current wrong location, cheapest)
- B: Create a separate `internal/.../registry_templates/` package (still internal, not canonical)
- C: Create Go package at `api/cryptosuite-registry/` with exported `TemplatesFS embed.FS` ✓ **SELECTED**
- D: Use `os.ReadFile` at runtime reading from project root path (no embed, runtime dependency)
- E:

**Decision**: Option C — Go package at `api/cryptosuite-registry/` exports embedded FS

**Rationale**: The architecture decision (v9 Item 23) explicitly requires templates at `api/cryptosuite-registry/templates/`. The `api/` directory is the correct place for machine-readable API definitions and canonical data. `os.ReadFile` at runtime (Option D) would make linters fragile (depends on working directory). A new internal package (Option B) would still be wrong location. The compile-time embed approach ensures the linter binary is self-contained.

**Impact**: `api/cryptosuite-registry/` becomes a real Go package. Import alias `cryptoutilRegistryTemplates "cryptoutil/api/cryptosuite-registry"`.

### Decision 2: Per-Product Template Files (Not Parameterized by Product)

**Options**:
- A: Single `product-compose.yml.tmpl` with complex iteration placeholders for child PS-IDs
- B: Per-product template files (`product-sm-compose.yml.tmpl` etc.) ✓ **SELECTED**
- C: Code-generated expected content (Go code computes expected product compose from registry)
- D: Skip product/suite template linting (deferred)
- E:

**Decision**: Option B — 5 per-product template files + 1 suite template

**Rationale**: Product compose files differ significantly per product (differing numbers of PS-ID children, different port ranges, different service names). A single parameterized template would require complex iteration logic in the template engine. Per-product static templates (with only `__SUITE__`, `__IMAGE_TAG__`, `__BUILD_DATE__` substitutions) are simple, readable, and maintainable. When a product's list of services changes, the corresponding template is updated directly.

**Impact**: `api/cryptosuite-registry/templates/` contains 13 files total (6 original + 7 new).

### Decision 3: `suite-Dockerfile.tmpl` Is Independent (Not Derived from `Dockerfile.tmpl`)

**Options**:
- A: `suite-Dockerfile.tmpl` is a separate complete template for the `cryptoutil` suite binary
- B: Derive suite Dockerfile from `Dockerfile.tmpl` with a known psID substitution
- C: Re-use `Dockerfile.tmpl` with `__PS_ID__ = __SUITE__` (same file, different params) ✓ **SELECTED**
- D: No suite Dockerfile template (manual validation only)
- E:

**Decision**: Option C — `CheckSuiteDockerfile` reuses `Dockerfile.tmpl` with suite-specific params

**Rationale**: The suite Dockerfile follows the EXACT same 4-stage pattern as PS-ID Dockerfiles (as documented in deployment-templates.md Section J). The only differences are name substitutions: binary path uses `cryptoutil` not a PS-ID, label title is `cryptoutil`, description is `Cryptoutil Suite`. We can reuse `Dockerfile.tmpl` by providing params where `__PS_ID__` = `cryptoutil`, `__PRODUCT_DISPLAY_NAME__` = `Cryptoutil`, `__SERVICE_DISPLAY_NAME__` = `Suite`. This avoids duplication and ensures any future Dockerfile template changes automatically apply to the suite. No separate `suite-Dockerfile.tmpl` file is needed — `CheckSuiteDockerfile` uses `Dockerfile.tmpl` with suite params.

**Impact**: Only 12 total template files (not 13). `suite-Dockerfile.tmpl` is removed from the missing list.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Go `//go:embed` path issue | Low | High | Test build immediately after Phase 1 step 1 |
| Actual product/suite files don't match new templates | Medium | Medium | Fix files (not templates) in Phase 2 if needed; templates ARE the spec |
| Test coverage drops below 98% | Low | Medium | Run coverage immediately after implementing each linter |
| lint-fitness-registry.yaml missing new linters | Low | Low | Verify registration before marking Phase 2 complete |
| `deployment-templates.md` breaks lint-docs | Low | Low | Run `cicd-lint lint-docs` after every doc change |

---

## Quality Gates - MANDATORY

**Per-Task Gates**:
- ✅ `go build ./...` clean (no errors)
- ✅ `go test ./...` passes (100%, zero skips)
- ✅ `golangci-lint run` zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ `template_drift` package: ≥98% line coverage
- ✅ `api/cryptosuite-registry` Go package: 0% acceptable (embedded FS, no executable logic)

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

- [ ] All 12 template files exist in `api/cryptosuite-registry/templates/` (see Decision 3 — no `suite-Dockerfile.tmpl` needed)
- [ ] `api/cryptosuite-registry/` is a valid Go package exporting `TemplatesFS embed.FS`
- [ ] `template_drift/templates/` directory deleted
- [ ] `template_drift.go` imports and uses external TemplatesFS
- [ ] All 3 new linters (CheckProductCompose, CheckSuiteCompose, CheckSuiteDockerfile) implemented and registered
- [ ] All quality gates passing
- [ ] Documentation updated to reference canonical templates
- [ ] Evidence archived in `test-output/`
