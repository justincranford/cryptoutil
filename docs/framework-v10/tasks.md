# Tasks — Framework v10: Canonical Template Registry

**Status**: 0 of 22 tasks complete (0%)
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

## Phase 1: Create Canonical Template Directory

**Phase Objective**: Populate `api/cryptosuite-registry/templates/` with all 15 parameterized
template files. These are plain configuration files — NOT Go code. They mirror the structure of
`./deployments/` and `./configs/` with `__KEY__` placeholders in both directory paths and content.
Delete the obsolete `template_drift/templates/` directory.

### Task 1.1: Read and understand all existing template content

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Before creating any template files, read all 6 existing `.tmpl` files in
  `internal/.../template_drift/templates/` to understand exact content and placeholder usage.
  Also read the actual product/suite compose files to understand product template content.
- **Acceptance Criteria**:
  - [ ] All 6 `.tmpl` files read and understood
  - [ ] Actual `deployments/{sm,jose,pki,identity,skeleton}/compose.yml` files read
  - [ ] Actual `deployments/cryptoutil/Dockerfile` and `compose.yml` read
  - [ ] Param usage per template type noted

### Task 1.2: Create `api/cryptosuite-registry/templates/` directory tree

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 1.1
- **Description**: Create the directory structure. No `.go` files. No Go package. Plain directories only.
- **Acceptance Criteria**:
  - [ ] `api/cryptosuite-registry/templates/deployments/__PS_ID__/` exists
  - [ ] `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/` exists
  - [ ] `api/cryptosuite-registry/templates/deployments/__PRODUCT__/` exists (single product template dir)
  - [ ] `api/cryptosuite-registry/templates/deployments/__SUITE__/` exists (single suite template dir)
  - [ ] `api/cryptosuite-registry/templates/configs/__PS_ID__/` exists
  - [ ] NO `.go` files anywhere in `api/cryptosuite-registry/`

### Task 1.3: Create PS-ID level template files (8 files)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.2
- **Description**: Convert existing `.tmpl` files into parameterized template files.
  The content is identical to current `.tmpl` content but placed in parameterized subdirectory
  paths using `__PS_ID__` (no `.tmpl` extension). `config-sqlite.yml.tmpl` becomes TWO files
  (config-sqlite-1.yml and config-sqlite-2.yml); same for config-postgresql.

  Files to create in `api/cryptosuite-registry/templates/deployments/__PS_ID__/`:
  - `Dockerfile` (from `Dockerfile.tmpl`)
  - `compose.yml` (from `compose.yml.tmpl`)
  - `config/config-common.yml` (from `config-common.yml.tmpl`)
  - `config/config-sqlite-1.yml` (from `config-sqlite.yml.tmpl`, instance-1 port param)
  - `config/config-sqlite-2.yml` (from `config-sqlite.yml.tmpl`, instance-2 port param)
  - `config/config-postgresql-1.yml` (from `config-postgresql.yml.tmpl`, instance-1 port param)
  - `config/config-postgresql-2.yml` (from `config-postgresql.yml.tmpl`, instance-2 port param)

  File to create in `api/cryptosuite-registry/templates/configs/__PS_ID__/`:
  - `__PS_ID__.yml` (from `standalone-config.yml.tmpl`)
- **Acceptance Criteria**:
  - [ ] 7 files in `templates/deployments/__PS_ID__/` (Dockerfile + compose + 5 configs)
  - [ ] 1 file in `templates/configs/__PS_ID__/` (`__PS_ID__.yml`)
  - [ ] Content matches existing `.tmpl` files; instance-1 vs instance-2 files differ only in port param name
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/Dockerfile` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/config-common.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/config-sqlite-1.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/config-sqlite-2.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/config-postgresql-1.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/config-postgresql-2.yml` (CREATE)
  - `api/cryptosuite-registry/templates/configs/__PS_ID__/__PS_ID__.yml` (CREATE)

### Task 1.4: Create product compose template file (1 file, `__PRODUCT__` in path)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.2
- **Description**: Create ONE parameterized template file at
  `api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose.yml`.
  The directory name `__PRODUCT__` is the expansion placeholder; the linter expands it × 5
  (sm, jose, pki, identity, skeleton) using per-product params from registry.yaml.

  The template content uses:
  - `__PRODUCT__` — substituted with product name (sm, jose, etc.)
  - `__SUITE__` — substituted with suite name
  - `__IMAGE_TAG__` — substituted with image tag
  - `__PRODUCT_INCLUDE_LIST__` — multi-line YAML include entries generated from registry.yaml
    (each product's PS-IDs produce include lines like `- path: ../sm-kms/compose.yml`)

  **Phase 1 implementation must**: read all 5 actual product compose files, verify they share
  a common structure, then write the single template. If any product compose file has
  non-parameterizable differences beyond the include list, add product-level metadata to
  registry.yaml to capture that variation.
- **Acceptance Criteria**:  
  - [ ] Exactly ONE template file: `templates/deployments/__PRODUCT__/compose.yml`
  - [ ] Template uses `__PRODUCT__`, `__SUITE__`, `__IMAGE_TAG__`, `__PRODUCT_INCLUDE_LIST__`
  - [ ] No literal product names (`sm`, `jose`, `pki`, `identity`, `skeleton`) in the path
  - [ ] `go build ./...` succeeds
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose.yml` (CREATE)

### Task 1.5: Create suite template files (2 files, `__SUITE__` in path)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.2
- **Description**: Create suite-level template files in `templates/deployments/__SUITE__/`.
  The directory name `__SUITE__` is the expansion placeholder; the linter expands it × 1
  (currently `cryptoutil`). Parameterized even though there is only one suite value today,
  because the suite name may change in the future.

  Template content uses `__SUITE__` throughout (not the hard-coded literal `cryptoutil`).
  These are compared against `deployments/cryptoutil/{Dockerfile,compose.yml}` (after expansion).
- **Acceptance Criteria**:
  - [ ] `templates/deployments/__SUITE__/Dockerfile` created with `__SUITE__` and suite-level params
  - [ ] `templates/deployments/__SUITE__/compose.yml` created with `__SUITE__` and suite-level params
  - [ ] Literal `cryptoutil` does NOT appear in either template file content (use `__SUITE__`)
  - [ ] Manually verifying: substituting `__SUITE__`=`cryptoutil` produces content identical to actual files
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__SUITE__/Dockerfile` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__SUITE__/compose.yml` (CREATE)

### Task 1.6: Delete the obsolete `template_drift/templates/` directory

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Tasks 1.3, 1.4, 1.5 (all canonical templates exist first)
- **Description**: Remove the 6 old `.tmpl` files from `internal/.../template_drift/templates/`.
  Their content is superseded by the canonical files in `api/cryptosuite-registry/templates/`.
- **Acceptance Criteria**:
  - [ ] `internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/` does NOT exist
  - [ ] `go build ./...` succeeds (no dangling references to removed directory)

---

## Phase 2: Rewrite Template Linter

**Phase Objective**: Rewrite `template_drift.go` to read templates from disk at runtime using
`os.WalkDir`, build an in-memory expected filesystem, and compare against actual `./deployments/`
and `./configs/` directories. Replace all per-file `Check*` functions with a single
`CheckTemplateCompliance` linter.

### Task 2.1: Rewrite core engine in `template_drift.go`

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 1 complete
- **Description**: Rewrite `template_drift.go` to implement the new in-memory FS approach:

  1. `LoadTemplatesDir(projectRoot string) (map[string]string, error)` — walks
     `{projectRoot}/api/cryptosuite-registry/templates/` using `os.WalkDir`, returns map
     of template-relative path → raw file content (e.g., `"deployments/__PS_ID__/Dockerfile"` → content)

  2. `BuildExpectedFS(templates map[string]string, registry *Registry) (map[string]string, error)` —
     for each template path in the map, detect the expansion key:
     - `__PS_ID__` in path: iterate over all 10 PS-IDs from registry; substitute `__PS_ID__` in
       BOTH the path AND call `buildParams(psID)` for full content substitution
     - `__PRODUCT__` in path: iterate over all 5 products from registry; substitute `__PRODUCT__`
       in path AND call `buildProductParams(product)` for content (includes `__PRODUCT_INCLUDE_LIST__`
       computed from registry PS-IDs for that product)
     - `__SUITE__` in path: iterate over suite name(s) from registry (currently just `cryptoutil`);
       substitute `__SUITE__` in path AND call `buildSuiteParams()` for content
     - `__INFRA_TOOL__` in path: iterate over infra tools from registry; substitute accordingly
     - No expansion key: substitute generic params in content; use template-relative path directly

  3. `CompareExpectedFS(expected map[string]string, projectRoot string) error` — for each entry
     in expected FS: resolve `{projectRoot}/{resolvedPath}`, read actual file, compare after
     normalization; collect all diffs; return aggregated error listing all mismatches

  4. REMOVE: `//go:embed templates/*`, `var templatesFS embed.FS`, `instantiate()` function,
     all individual `Check*` per-file wrappers

  5. KEEP: `buildParams(psID string) map[string]string`, `normalizeCommentAlignment`,
     `normalizeLineEndings`

  6. ADD: `buildProductParams(product string, registry *Registry) map[string]string` —
     builds product-level substitution map including `__PRODUCT_INCLUDE_LIST__`
     (multi-line include entries for each of that product's PS-IDs)

  7. ADD: `buildSuiteParams(registry *Registry) map[string]string` —
     builds suite-level substitution map (`__SUITE__`, display names, etc.)
- **Acceptance Criteria**:
  - [ ] `template_drift.go` has NO `//go:embed` directive, NO `embed.FS`, NO `embed` import
  - [ ] `LoadTemplatesDir` correctly discovers all 12 template files
  - [ ] `BuildExpectedFS` correctly expands `__PS_ID__` templates × 10: 80 deployment files + 10 config files
  - [ ] `BuildExpectedFS` correctly expands `__PRODUCT__` template × 5 product compose files
  - [ ] `BuildExpectedFS` correctly expands `__SUITE__` templates × 1 (Dockerfile + compose.yml)
  - [ ] `CompareExpectedFS` returns aggregated error with description for each mismatch
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/template_drift.go` (REWRITE)

### Task 2.2: Implement `CheckTemplateCompliance` single linter in `checks.go`

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 2.1
- **Description**: Rewrite `checks.go` to remove all per-file `Check*` functions and add a single
  `CheckTemplateCompliance(logger *Logger) error`. Use seam injection for testing:

  ```
  type templateComplianceFn func(projectRoot string) (map[string]string, error)
  func CheckTemplateCompliance(logger *Logger) error {
      return checkTemplateComplianceInDir(logger, ".", defaultComplianceFn)
  }
  func checkTemplateComplianceInDir(logger *Logger, projectRoot string, fn templateComplianceFn) error {
      ...
  }
  ```

  `defaultComplianceFn` calls `LoadTemplatesDir → BuildExpectedFS → CompareExpectedFS`.
- **Acceptance Criteria**:
  - [ ] All old `Check*` functions removed from `checks.go`
  - [ ] `CheckTemplateCompliance` added
  - [ ] `checkTemplateComplianceInDir` is the seam-injectable private function
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/checks.go` (REWRITE)

### Task 2.3: Update `lint-fitness-registry.yaml` and dispatch

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 2.2
- **Description**: Replace all per-file template linter entries with a single `template-compliance`
  entry. Update dispatch in `lint_fitness.go` to wire `CheckTemplateCompliance`.
- **Acceptance Criteria**:
  - [ ] All old per-file entries removed (`template-dockerfile`, `template-compose`, etc.)
  - [ ] Single `template-compliance` entry added, wired to `CheckTemplateCompliance`
  - [ ] `go run ./cmd/cicd-lint lint-fitness` runs without unknown command errors
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml` (MODIFY)
  - `internal/apps/tools/cicd_lint/lint_fitness/lint_fitness.go` (MODIFY)

### Task 2.4: Write tests (≥98% coverage, seam injection pattern)

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 2.2
- **Description**: Write comprehensive tests using `t.TempDir()` for file system test isolation.
  All tests must call `t.Parallel()`. Use table-driven pattern for multi-variant cases.

  Test cases needed:
  - `TestLoadTemplatesDir_Happy`: temp dir with a few template files → correct map returned
  - `TestLoadTemplatesDir_NonExistentRoot`: error returned for missing directory
  - `TestBuildExpectedFS_PSIDExpansion`: 1 template with `__PS_ID__` path → N expansions
  - `TestBuildExpectedFS_StaticPath`: 1 static template → 1 entry without expansion
  - `TestBuildExpectedFS_ContentSubstitution`: verify `__PS_ID__` substituted in content too
  - `TestCompareExpectedFS_AllMatch`: expected FS matches temp dir → nil error
  - `TestCompareExpectedFS_ContentMismatch`: one file has wrong content → error with diff description
  - `TestCompareExpectedFS_MissingFile`: expected file does not exist → error
  - `TestCheckTemplateComplianceInDir_Success`: inject fn returning valid expected FS → passes
  - `TestCheckTemplateComplianceInDir_ComplianceError`: inject fn returning mismatch → error
  - `TestCheckTemplateComplianceInDir_LoadError`: inject fn returning error → propagated
- **Acceptance Criteria**:
  - [ ] All test functions have `t.Parallel()` at top
  - [ ] Table-driven for multi-variant tests
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...` → 100% pass
  - [ ] Coverage ≥98% for `template_drift` package
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/template_drift_test.go` (MODIFY/ADD)

### Task 2.5: Fix actual deployment/config files if templates reveal drift

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 2.3
- **Description**: Run `go run ./cmd/cicd-lint lint-fitness` to detect any drift between the new
  templates and the actual `./deployments/` and `./configs/` files. Templates ARE the canonical
  spec — if a file on disk differs from the template-expanded expected content, fix the file on
  disk (not the template).
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes with zero template-compliance violations
- **Evidence**:
  - `test-output/framework-v10/phase2/lint-fitness-output.txt`

---

## Phase 3: Update Documentation

**Phase Objective**: Ensure docs accurately describe the new parameterized template architecture.
Template files are the source of truth; docs describe them.

### Task 3.1: Update `deployment-templates.md` template content sections

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 1 complete
- **Description**: For Sections B.1, C.1, D.1-5, E, G.1, I, J: replace embedded template code
  blocks with a reference to the canonical file in `api/cryptosuite-registry/templates/{path}`.
  Keep all rule tables, parameter tables, and rationale text.
- **Acceptance Criteria**:
  - [ ] No embedded template code blocks remain in those sections
  - [ ] Each section references its canonical template file path
  - [ ] All rule tables and rationale text intact
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/deployment-templates.md` (MODIFY)

### Task 3.2: Update Section O.2 template catalog

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 3.1
- **Description**: Update the template file catalog in Section O.2 to list all 15 canonical
  template files with their parameterized paths, expansion behavior (PS-ID × 10, product × 5,
  suite × 1), and the actual files they validate.
- **Acceptance Criteria**:
  - [ ] Section O.2 correctly lists all 15 template files
  - [ ] Expansion behavior documented for each file type
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/deployment-templates.md` (MODIFY)

### Task 3.3: Update `target-structure.md`

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 3.2
- **Description**: Add `api/cryptosuite-registry/templates/` tree to `target-structure.md`
  listing all 15 template files and the full directory structure.
- **Acceptance Criteria**:
  - [ ] `target-structure.md` shows `api/cryptosuite-registry/templates/` with full subtree
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/target-structure.md` (MODIFY)

### Task 3.4: Annotate v9 plan and tasks with correction note

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None
- **Description**: Add a `⚠️ v9 IMPLEMENTATION ERROR` note to v9 Task 8.1 and v9 plan Item 23
  indicating templates were placed at wrong location; v10 corrects this.
- **Acceptance Criteria**:
  - [ ] `docs/framework-v9/plan.md` Item 23 has correction note
  - [ ] `docs/framework-v9/tasks.md` Task 8.1 has correction note
- **Files**:
  - `docs/framework-v9/plan.md` (MODIFY)
  - `docs/framework-v9/tasks.md` (MODIFY)

---

## Phase 4: Quality Gates

**Phase Objective**: Full validation that all quality gates pass end-to-end.

### Task 4.1: Full build and lint validation

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: All Phase 1–3 tasks
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `golangci-lint run` zero warnings/errors
  - [ ] `golangci-lint run --build-tags e2e,integration` zero warnings/errors
- **Evidence**: `test-output/framework-v10/phase4/build-lint.txt`

### Task 4.2: Full test suite and coverage

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 4.1
- **Acceptance Criteria**:
  - [ ] `go test ./...` 100% pass, zero skips
  - [ ] `go test -race -count=2 ./internal/apps/tools/...` race-free
  - [ ] ≥98% coverage for `template_drift` package
- **Evidence**:
  - `test-output/framework-v10/phase4/test-results.txt`
  - `test-output/framework-v10/phase4/coverage-template-drift.out`

### Task 4.3: cicd-lint end-to-end validation

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 4.1
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (template-compliance green)
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Evidence**: `test-output/framework-v10/phase4/cicd-lint-output.txt`

---

## Phase 5: Knowledge Propagation

**Phase Objective**: Apply lessons learned. NEVER skip.

### Task 5.1: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: All Phase 1–4 tasks
- **Acceptance Criteria**:
  - [ ] Section 9.11.1 fitness linter catalog: old per-file entries removed, `template-compliance` added
  - [ ] Section 13.6 describes parameterized template directory structure and `os.WalkDir` approach
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes

### Task 5.2: Update instruction files if new patterns discovered

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 5.1
- **Acceptance Criteria**:
  - [ ] Any new patterns documented in relevant instruction files
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes

### Task 5.3: Final clean commit

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 5.2
- **Acceptance Criteria**:
  - [ ] All changes committed with conventional commit messages (one commit per semantic group)
  - [ ] `git status --porcelain` returns empty
  - [ ] Final `go run ./cmd/cicd-lint lint-fitness` passes

---

## Cross-Cutting Tasks

### Code Quality (enforced per task)

- [ ] No new `//nolint:` directives without GitHub issue reference
- [ ] All magic constants in `internal/shared/magic/` (no new magic literals in Go code)
- [ ] `gofumpt -w .` before committing any Go file
- [ ] No `.go` files added to `api/cryptosuite-registry/` (CRITICAL — it MUST NOT be a Go package)

### Testing (enforced per task)

- [ ] `t.Parallel()` on all new tests and subtests
- [ ] Seam injection pattern for all new linter functions
- [ ] Table-driven for multi-variant tests
- [ ] `t.TempDir()` for file system test isolation
- [ ] ≥98% line coverage for `template_drift` package

---

## Notes

**Critical constraint**: `api/cryptosuite-registry/` MUST NOT contain any `.go` files.
The templates are plain configuration files, not Go source code. Any `.go` file discovered
there during implementation is an error and must be removed immediately.

**Expansion rules**:
- `__PS_ID__` in path → expand for all 10 PS-IDs from `registry.yaml`
- No `__KEY__` in path → compare directly (product/suite files; substitute content params only)
- `__PRODUCT__` in path: not used in v10 (per-product templates have static paths)

**Total template files**: 15 physical files → 87 expected files after full expansion
(8 PS-ID templates × 10 PS-IDs = 80 expanded, + 5 product compose + 2 suite files = 87)

---

## Evidence Archive

- `test-output/framework-v10/phase2/` — lint-fitness output after Phase 2
- `test-output/framework-v10/phase4/` — full quality gate evidence
