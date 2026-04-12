# Tasks — Framework v10: Canonical Template Relocation

**Status**: 0 of 30 tasks complete (0%)
**Created**: 2026-04-12
**Last Updated**: 2026-04-12

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥98% coverage/mutation for infrastructure)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**
- ✅ **Fix issues immediately** — when unknowns discovered, blockers identified, any tests fail, or quality gates not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **NEVER defer**: No "fix later", no "non-critical", no shortcuts

---

## Phase 1: Relocate Templates to Canonical Location

**Phase Objective**: Move the 6 existing template files from `internal/.../template_drift/templates/`
to the canonical `api/cryptosuite-registry/templates/`. Create the Go package that exports
`TemplatesFS embed.FS`. Update `template_drift.go` to use the imported FS instead of its own embed.

### Task 1.1: Create Go package at `api/cryptosuite-registry/`

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Create `api/cryptosuite-registry/templates.go` with `package cryptosuiteregistry`
  and an exported `TemplatesFS embed.FS` using `//go:embed templates/*`.
- **Acceptance Criteria**:
  - [ ] `api/cryptosuite-registry/templates.go` created with correct package declaration
  - [ ] `//go:embed templates/*` directive and `var TemplatesFS embed.FS` exported
  - [ ] `api/cryptosuite-registry/templates/` directory exists (initially empty — files moved in Task 1.2)
  - [ ] `go build ./api/...` succeeds (even with empty templates dir the embed will fail until files added)
- **Files**:
  - `api/cryptosuite-registry/templates.go` (CREATE)
  - `api/cryptosuite-registry/templates/` (CREATE directory)

### Task 1.2: Move template files to canonical location

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.1
- **Description**: Copy all 6 template files from `internal/.../template_drift/templates/` to
  `api/cryptosuite-registry/templates/`. Verify content is identical.
- **Acceptance Criteria**:
  - [ ] `api/cryptosuite-registry/templates/Dockerfile.tmpl` exists (identical content)
  - [ ] `api/cryptosuite-registry/templates/compose.yml.tmpl` exists (identical content)
  - [ ] `api/cryptosuite-registry/templates/config-common.yml.tmpl` exists (identical content)
  - [ ] `api/cryptosuite-registry/templates/config-sqlite.yml.tmpl` exists (identical content)
  - [ ] `api/cryptosuite-registry/templates/config-postgresql.yml.tmpl` exists (identical content)
  - [ ] `api/cryptosuite-registry/templates/standalone-config.yml.tmpl` exists (identical content)
  - [ ] `go build ./api/...` succeeds (embed resolves all 6 files)
- **Files**:
  - `api/cryptosuite-registry/templates/` (POPULATE with 6 files)

### Task 1.3: Update `template_drift.go` to use external FS

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.2
- **Description**: Remove the `//go:embed templates/*` directive and `var templatesFS embed.FS` from
  `template_drift.go`. Import `cryptoutilRegistryTemplates "cryptoutil/api/cryptosuite-registry"`.
  Replace all `templatesFS.ReadFile(...)` calls with `cryptoutilRegistryTemplates.TemplatesFS.ReadFile(...)`.
- **Acceptance Criteria**:
  - [ ] `template_drift.go` no longer has `//go:embed templates/*`
  - [ ] `template_drift.go` no longer has `var templatesFS embed.FS`
  - [ ] Import `cryptoutilRegistryTemplates "cryptoutil/api/cryptosuite-registry"` added
  - [ ] All `templatesFS.ReadFile(...)` calls updated to use `cryptoutilRegistryTemplates.TemplatesFS.ReadFile(...)`
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/template_drift.go` (MODIFY)

### Task 1.4: Delete old templates directory

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 1.3
- **Description**: Delete `internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/` directory.
  This directory is obsolete — templates now live in `api/cryptosuite-registry/templates/`.
- **Acceptance Criteria**:
  - [ ] `internal/.../template_drift/templates/` directory does NOT exist
  - [ ] `go build ./...` succeeds (no dangling embed directive)
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...` passes
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/` (DELETE)

### Task 1.5: Verify tests still pass

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.4
- **Description**: Run full test suite for template_drift package. The seam injection tests should
  still work because they use `instantiateFn` injection (not `templatesFS` directly).
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...` passes
  - [ ] Coverage still ≥98% for `template_drift` package
  - [ ] `go test ./...` (full suite) passes with zero failures
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (template linters still validate actual files)
- **Evidence**:
  - `test-output/framework-v10/phase1/test-results.txt`
  - `test-output/framework-v10/phase1/coverage-template-drift.out`

## Phase 2: Add Missing Templates and Linters

**Phase Objective**: Create the 7 missing template files in `api/cryptosuite-registry/templates/`
(5 per-product compose + 1 suite compose; suite Dockerfile reuses `Dockerfile.tmpl` per Decision 3).
Implement new linters for product/suite validation.

### Task 2.1: Create product compose templates (5 files)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.2 (templates directory exists)
- **Description**: Create one template per product compose file. Each template is based on the
  actual `deployments/{PRODUCT}/compose.yml` content, with `cryptoutil` → `__SUITE__` and
  `dev` image tag → `__IMAGE_TAG__` substituted as placeholders.

  Products and their source files:
  - `sm`: `deployments/sm/compose.yml` → `product-sm-compose.yml.tmpl`
  - `jose`: `deployments/jose/compose.yml` → `product-jose-compose.yml.tmpl`
  - `pki`: `deployments/pki/compose.yml` → `product-pki-compose.yml.tmpl`
  - `identity`: `deployments/identity/compose.yml` → `product-identity-compose.yml.tmpl`
  - `skeleton`: `deployments/skeleton/compose.yml` → `product-skeleton-compose.yml.tmpl`

  Parameterization: Replace `cryptoutil` with `__SUITE__`, replace image tag `dev` with `__IMAGE_TAG__`.
  Everything else (PS-ID names, port numbers, includes, service lists) is product-specific and static.
- **Acceptance Criteria**:
  - [ ] All 5 product compose templates created in `api/cryptosuite-registry/templates/`
  - [ ] Each template uses `__SUITE__` and `__IMAGE_TAG__` as placeholders
  - [ ] `go build ./api/...` succeeds (embed resolves all 12 files)
- **Files**:
  - `api/cryptosuite-registry/templates/product-sm-compose.yml.tmpl` (CREATE)
  - `api/cryptosuite-registry/templates/product-jose-compose.yml.tmpl` (CREATE)
  - `api/cryptosuite-registry/templates/product-pki-compose.yml.tmpl` (CREATE)
  - `api/cryptosuite-registry/templates/product-identity-compose.yml.tmpl` (CREATE)
  - `api/cryptosuite-registry/templates/product-skeleton-compose.yml.tmpl` (CREATE)

### Task 2.2: Create suite compose template

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.2
- **Description**: Create `suite-compose.yml.tmpl` based on `deployments/cryptoutil/compose.yml`.
  Parameterize `__SUITE__` (for binary names), `__IMAGE_TAG__` (for image tags), `__BUILD_DATE__`.
- **Acceptance Criteria**:
  - [ ] `api/cryptosuite-registry/templates/suite-compose.yml.tmpl` created
  - [ ] Template uses `__SUITE__` and `__IMAGE_TAG__` placeholders
  - [ ] `go build ./api/...` succeeds
- **Files**:
  - `api/cryptosuite-registry/templates/suite-compose.yml.tmpl` (CREATE)

### Task 2.3: Implement `CheckProductCompose` linter

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 2.1, Task 1.3
- **Description**: Add `CheckProductCompose(logger)` in `template_drift/checks.go`. Iterates the 5
  products, loads `product-{PRODUCT}-compose.yml.tmpl`, substitutes `__SUITE__`/`__IMAGE_TAG__`/
  `__BUILD_DATE__`, compares against `deployments/{PRODUCT}/compose.yml`.

  Uses `compareExact` for most products; `compareSupersetOrdered` for `identity` (or exact if
  files actually match).
- **Acceptance Criteria**:
  - [ ] `CheckProductCompose` function added to `checks.go`
  - [ ] Iterates over all 5 products
  - [ ] Correct comparison mode per product
  - [ ] Returns descriptive diff errors if any product compose file differs
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/checks.go` (MODIFY)

### Task 2.4: Implement `CheckSuiteCompose` and `CheckSuiteDockerfile` linters

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 2.2, Task 1.3
- **Description**:
  - `CheckSuiteCompose(logger)`: loads `suite-compose.yml.tmpl`, substitutes suite params,
    compares against `deployments/cryptoutil/compose.yml` (use `compareExact` or superset as needed).
  - `CheckSuiteDockerfile(logger)`: reuses `Dockerfile.tmpl` with suite-specific params
    (`__PS_ID__` = `cryptoutil`, `__PRODUCT_DISPLAY_NAME__` = `Cryptoutil`,
    `__SERVICE_DISPLAY_NAME__` = `Suite`), compares against `deployments/cryptoutil/Dockerfile`.

  Suite params map includes:
  - `__SUITE__` = `cryptoutil`
  - `__IMAGE_TAG__` = `dev`
  - `__BUILD_DATE__` = `2026-02-17T00:00:00Z`
  - `__PS_ID__` = `cryptoutil`
  - `__PS_ID_UPPER__` = `CRYPTOUTIL`
  - `__PRODUCT_DISPLAY_NAME__` = `Cryptoutil`
  - `__SERVICE_DISPLAY_NAME__` = `Suite`
  - All standard build/container params
- **Acceptance Criteria**:
  - [ ] `CheckSuiteCompose` added to `checks.go`
  - [ ] `CheckSuiteDockerfile` added to `checks.go`
  - [ ] Neither function requires a new template file (reuse existing)
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/checks.go` (MODIFY)

### Task 2.5: Register new linters in registry

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 2.3, Task 2.4
- **Description**: Add all 3 new linters to `lint-fitness-registry.yaml` and ensure they are
  wired into `lint_fitness.go` dispatch table.
- **Acceptance Criteria**:
  - [ ] `template-product-compose` linter registered in `lint-fitness-registry.yaml`
  - [ ] `template-suite-compose` linter registered
  - [ ] `template-suite-dockerfile` linter registered
  - [ ] `go run ./cmd/cicd-lint lint-fitness` includes all 3 new checks
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml` (MODIFY)
  - `internal/apps/tools/cicd_lint/lint_fitness/lint_fitness.go` (MODIFY — add dispatch entries)

### Task 2.6: Write tests for new linters (≥98% coverage)

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 2.3, Task 2.4
- **Description**: Add test cases for `CheckProductCompose`, `CheckSuiteCompose`,
  `CheckSuiteDockerfile` in the appropriate test files. Use seam injection pattern
  (`instantiateFn`) to test error paths without real file I/O.

  Test cases needed per linter:
  - Happy path: actual file matches template (no errors returned)
  - Template load error: inject failing instantiateFn returns error
  - File not found: provide non-existent path
  - Drift detected: inject instantiateFn returning modified content, verify diff in error
- **Acceptance Criteria**:
  - [ ] Tests cover all happy paths and error paths for each new linter
  - [ ] `go test -coverprofile=coverage.out ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...` ≥98%
  - [ ] `t.Parallel()` on all test functions and subtests
  - [ ] No hardcoded UUIDs
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/checks_error_test.go` (MODIFY)
  - or `template_drift_test.go` (MODIFY as appropriate)

### Task 2.7: Fix actual product/suite files if needed

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 2.5, Task 2.6
- **Description**: Run `go run ./cmd/cicd-lint lint-fitness` to detect any drift between the new
  templates and the actual product/suite compose files. If drift exists, fix the actual files
  (not the templates — templates ARE the spec). Fix the suite Dockerfile if it differs.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (zero template violations)
  - [ ] Any product/suite compose file differences resolved
  - [ ] `go test ./...` passes after any file fixes
- **Evidence**:
  - `test-output/framework-v10/phase2/lint-fitness-output.txt`

## Phase 3: Update Documentation

**Phase Objective**: Make `deployment-templates.md` and `target-structure.md` point to the canonical
template files instead of duplicating template content inside the documentation. The template files
ARE the source of truth; docs describe them.

### Task 3.1: Update `deployment-templates.md` template content sections

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 2 complete (all template files exist)
- **Description**: For Sections B.1, C.1, D.1-D.5, E, G.1, I, J: replace the embedded
  template content (```` ```dockerfile ``` ``` / ```` ```yaml``` ```` blocks) with a short
  reference to the canonical template file. Keep all rule tables, parameter tables, and rationale.

  Replacement format for each template section:
  ```
  ### X.1 Canonical Template

  Canonical template: [`api/cryptosuite-registry/templates/FILENAME.tmpl`](../api/cryptosuite-registry/templates/FILENAME.tmpl)

  The template uses `__KEY__` placeholder syntax. See [Section A](#a-parameterization-table) for all parameters.
  Linter `template-LINTERNAME` instantiates this template for each PS-ID and compares byte-for-byte
  against the actual file on disk.
  ```

  Retain the `### X.2 Template Rules` tables and all other non-template-content sections.
- **Acceptance Criteria**:
  - [ ] No embedded template content (code blocks) remain in Sections B.1, C.1, D.1-D.5, E.1, G.1, I, J
  - [ ] Each section references `api/cryptosuite-registry/templates/FILENAME.tmpl`
  - [ ] All rule tables intact
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/deployment-templates.md` (MODIFY)

### Task 3.2: Update `target-structure.md` for `api/cryptosuite-registry/templates/`

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 3.1
- **Description**: Add `api/cryptosuite-registry/` section to `target-structure.md` listing the
  Go package file and all 12 template files in the `templates/` subdirectory.
- **Acceptance Criteria**:
  - [ ] `target-structure.md` lists `api/cryptosuite-registry/templates.go`
  - [ ] All 12 template files listed under `api/cryptosuite-registry/templates/`
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/target-structure.md` (MODIFY)

### Task 3.3: Update Section O.2 of `deployment-templates.md` (template catalog)

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 3.1
- **Description**: Update the template file catalog in Section O.2 to reflect the complete 12-file list
  (6 original PS-ID templates + 5 product templates + 1 suite compose). Note Decision 3 (no separate
  suite-Dockerfile.tmpl needed — `Dockerfile.tmpl` reused with suite params).
- **Acceptance Criteria**:
  - [ ] Section O.2 catalog lists all 12 template files and their purpose
  - [ ] Decision 3 (suite Dockerfile reuses Dockerfile.tmpl) noted in Section J
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/deployment-templates.md` (MODIFY)

### Task 3.4: Update v9 plan to note v10 correction

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None
- **Description**: Add a brief note to v9 `plan.md` Item 23 and Task 8.1 indicating the implementation
  error (templates stored in wrong location) and that v10 corrects this.
- **Acceptance Criteria**:
  - [ ] v9 `plan.md` Item 23 has a note: "⚠️ v9 IMPLEMENTATION ERROR: templates placed in `template_drift/templates/` instead of `api/cryptosuite-registry/templates/`. Corrected by v10."
  - [ ] v9 `tasks.md` Task 8.1 has the same correction note
- **Files**:
  - `docs/framework-v9/plan.md` (MODIFY)
  - `docs/framework-v9/tasks.md` (MODIFY)

## Phase 4: Quality Gates

**Phase Objective**: Full validation that all quality gates pass end-to-end.

### Task 4.1: Full build validation

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: All Phase 1-3 tasks
- **Description**: Run all build variants.
- **Acceptance Criteria**:
  - [ ] `go build ./...` — clean (no errors)
  - [ ] `go build -tags e2e,integration ./...` — clean
- **Evidence**: `test-output/framework-v10/phase4/build.txt`

### Task 4.2: Full linting validation

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 4.1
- **Acceptance Criteria**:
  - [ ] `golangci-lint run` — zero warnings/errors
  - [ ] `golangci-lint run --build-tags e2e,integration` — zero warnings/errors
  - [ ] `go run ./cmd/cicd-lint lint-fitness` — all linters pass
  - [ ] `go run ./cmd/cicd-lint lint-deployments` — all validators pass
  - [ ] `go run ./cmd/cicd-lint lint-docs` — all checks pass
- **Evidence**: `test-output/framework-v10/phase4/lint.txt`

### Task 4.3: Full test suite

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 4.1
- **Acceptance Criteria**:
  - [ ] `go test ./...` — 100% pass, zero skips
  - [ ] Coverage ≥98%: `go test -coverprofile=... ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...`
  - [ ] `go test -race -count=2 ./internal/apps/tools/...` — race-detector clean
- **Evidence**:
  - `test-output/framework-v10/phase4/test-results.txt`
  - `test-output/framework-v10/phase4/coverage-template-drift.out`

## Phase 5: Knowledge Propagation

**Phase Objective**: Apply lessons learned. NEVER skip.

### Task 5.1: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: All Phase 1-4 tasks
- **Description**: Update sections that reference the canonical template location and linter catalog.
- **Acceptance Criteria**:
  - [ ] Section 9.11.1 (fitness linter catalog) updated with new product/suite template linters
  - [ ] Section 13.6 (template enforcement) references `api/cryptosuite-registry/templates/`
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

### Task 5.2: Update instruction files

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 5.1
- **Description**: Review `04-01.deployment.instructions.md` for any references to wrong template location.
- **Acceptance Criteria**:
  - [ ] `04-01.deployment.instructions.md` references correct template location
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes

### Task 5.3: Final commit and clean worktree

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 5.2
- **Acceptance Criteria**:
  - [ ] All changes committed with conventional commit messages
  - [ ] `git status --porcelain` returns empty
  - [ ] Final `go run ./cmd/cicd-lint lint-fitness` passes

---

## Cross-Cutting Tasks

### Testing (enforced per task)

- [ ] `t.Parallel()` on all new tests and subtests
- [ ] Seam injection pattern for all new linter tests
- [ ] Table-driven tests for multi-case scenarios
- [ ] ≥98% line coverage for `template_drift` package

### Code Quality (enforced per task)

- [ ] No new `//nolint:` directives
- [ ] Import alias `cryptoutilRegistryTemplates` for new package import
- [ ] All constants in `internal/shared/magic/` (no new magic literals in Go code)
- [ ] `gofumpt -w .` before committing

---

## Notes / Deferred Work

None — all issues must be resolved in v10. No deferral permitted.

---

## Evidence Archive

- `test-output/framework-v10/phase1/` — Phase 1 test results and coverage
- `test-output/framework-v10/phase2/` — Phase 2 lint-fitness output
- `test-output/framework-v10/phase4/` — Full quality gate evidence
