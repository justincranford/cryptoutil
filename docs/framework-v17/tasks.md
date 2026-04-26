# Tasks - Framework V17: internal/apps/ Structure Fitness Linters

**Status**: 0 of 40 tasks complete (0%)
**Last Updated**: 2026-04-26
**Created**: 2026-04-26

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL code functionally correct; comprehensive tests |
| Completeness | NO phases/tasks/steps skipped; NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Reliability | ≥98% infrastructure/utility coverage; ≥98% mutation |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Task Status Legend — MANDATORY

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| ❌ | Not started | Task not yet begun |
| 🔄 | In progress | Currently being worked on |
| ✅ | Complete | Task finished with evidence |
| ⏳ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Phase 1: Gap Analysis & Linter Design

**Phase Objective**: Validate the research findings in plan.md by running targeted checks against
the live codebase; finalize the violation matrix; write linter acceptance criteria.

### Task 1.1: Build Health Pre-Flight

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Verify the codebase compiles and all existing fitness linters pass before
  starting new work. Establish a clean baseline.
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go build -tags e2e,integration ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/...` exits 0
  - [ ] Output archived in `test-output/phase1/preflight-build.log`
  - [ ] Output archived in `test-output/phase1/preflight-fitness.log`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase1/`

### Task 1.2: Verify PS-ID Gap Matrix

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Write a shell survey script (or use Bash commands) to verify each cell of the
  gap matrix in plan.md against the actual codebase. Confirm which PS-IDs are missing each
  required file. Results determine the `knownExclusions` list for each new linter.
- **Acceptance Criteria**:
  - [ ] Every cell of the gap matrix confirmed (YES/NO for each PS-ID × invariant)
  - [ ] `swagger.go` missing from: `identity-rp`, `identity-spa` — confirmed
  - [ ] `testmain_test.go` missing from: `sm-kms`, `identity-authz`, `identity-idp`, `identity-rs`, `identity-rp`, `identity-spa` — confirmed
  - [ ] `*_lifecycle_test.go` missing from: `identity-rs`, `identity-rp`, `identity-spa` — confirmed
  - [ ] `*_port_conflict_test.go` missing from: `identity-authz`, `identity-idp`, `identity-rs`, `identity-rp`, `identity-spa` — confirmed
  - [ ] `sm/kms/`, `sm/im/`, `jose/ja/`, `pki/ca/`, `skeleton/template/` confirmed as service subdirs in product dirs
  - [ ] `identity/` shared packages (`apperr/`, `config/`, `domain/`, etc.) confirmed as shared libs (NOT service code) — verified via absence of PS-ID routing code
  - [ ] Gap matrix written to `test-output/phase1/gap-matrix.md`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase1/gap-matrix.md`

### Task 1.3: Linter Specification Document

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.2
- **Description**: Write the precise specification for each of the 6 new linters: package name,
  function signatures, input patterns, error messages, test fixture needs, and the exact
  `knownExclusions` list per linter (based on Task 1.2 confirmed gaps).
- **Acceptance Criteria**:
  - [ ] Spec written for all 6 linters
  - [ ] Each spec includes: package name, Check/CheckInDir signatures, required files/patterns, error message format, `knownExclusions` list at launch, test cases (pass + fail + exception)
  - [ ] Spec stored at `test-output/phase1/linter-specs.md`
  - [ ] No open questions remain (all ambiguities resolved)
- **Files**: None (planning artifact only)
- **Evidence**: `test-output/phase1/linter-specs.md`

---

## Phase 2: Implement 6 New Fitness Linters

**Phase Objective**: Implement all 6 linters with ≥98% coverage each. Each linter must pass
`go run ./cmd/cicd-lint lint-fitness` against the current codebase without CI breakage.

### Task 2.1: Linter `apps-ps-id-required-files` (registry-driven entry + usage files)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Implement `apps_ps_id_required_files/apps_ps_id_required_files.go`. Uses
  `AllProductServices()` to iterate all 10 PS-IDs and verify each has `{SERVICE}.go` and
  `{SERVICE}_usage.go`. This is a registry-driven rewrite of the hardcoded portion of the
  existing `service_structure.go`. Includes `knownExclusions` for any PS-IDs that are known
  gaps (none expected here — all 10 have both files per gap matrix).
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Uses `AllProductServices()` from registry (NOT a hardcoded list)
  - [ ] Derives service name from `ps.Service` field (e.g., `"kms"` → checks `kms.go`)
  - [ ] Error message format: `"{PS-ID}: missing required file internal/apps/{PS-ID}/{SERVICE}.go"`
  - [ ] Test: pass case (all 10 PS-IDs have required files)
  - [ ] Test: fail case (synthetic rootDir missing one entry file)
  - [ ] Test: fail case (synthetic rootDir missing one usage file)
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_required_files/...` ≥98%
  - [ ] `golangci-lint run ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_required_files/...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` still passes (this linter finds 0 violations)
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_required_files/apps_ps_id_required_files.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_required_files/apps_ps_id_required_files_test.go` (NEW)
- **Evidence**: `test-output/phase2/task-2.1-coverage.log`

### Task 2.2: Linter `apps-ps-id-server-package` (server/server.go + server/public_server.go)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Implement `apps_ps_id_server_package/apps_ps_id_server_package.go`. Checks
  every PS-ID has `server/server.go`. Also checks `server/public_server.go` EXCEPT for `sm-kms`
  (legacy service — does not follow this pattern yet). `knownExclusions` for `public_server.go`
  check: `sm-kms`. `knownExclusions` for identity-authz/idp server checks: none (both have
  `server/server.go`).
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Checks `server/server.go` for all 10 PS-IDs (no exclusions)
  - [ ] Checks `server/public_server.go` for 9 PS-IDs (excludes `sm-kms`)
  - [ ] `sm-kms` excluded from `public_server.go` check via `knownExclusions` map
  - [ ] Error message format: `"{PS-ID}: missing required file internal/apps/{PS-ID}/server/public_server.go"`
  - [ ] Test: pass case (all included PS-IDs have required server files)
  - [ ] Test: fail case (synthetic root missing server/server.go)
  - [ ] Test: fail case (synthetic root missing server/public_server.go for non-excluded PS-ID)
  - [ ] Test: exception case (sm-kms passes without public_server.go)
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` still passes
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_server_package/apps_ps_id_server_package.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_server_package/apps_ps_id_server_package_test.go` (NEW)
- **Evidence**: `test-output/phase2/task-2.2-coverage.log`

### Task 2.3: Linter `apps-ps-id-swagger-presence` (swagger.go + swagger_test.go)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Implement `apps_ps_id_swagger_presence/apps_ps_id_swagger_presence.go`. Checks
  every PS-ID has `swagger.go` and `swagger_test.go` at the package root. `knownExclusions`:
  `identity-rp` and `identity-spa` (confirmed missing per gap matrix — these are launch-time
  exclusions until the files are created in a follow-on plan).
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Checks `swagger.go` for 8 PS-IDs (excludes `identity-rp`, `identity-spa`)
  - [ ] Checks `swagger_test.go` for same 8 PS-IDs
  - [ ] `identity-rp` and `identity-spa` excluded via `knownExclusions` with TODO comment
  - [ ] Error message format: `"{PS-ID}: missing required file internal/apps/{PS-ID}/swagger.go"`
  - [ ] Test: pass case
  - [ ] Test: fail case (missing swagger.go)
  - [ ] Test: fail case (missing swagger_test.go)
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` still passes
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_swagger_presence/apps_ps_id_swagger_presence.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_swagger_presence/apps_ps_id_swagger_presence_test.go` (NEW)
- **Evidence**: `test-output/phase2/task-2.3-coverage.log`

### Task 2.4: Linter `apps-ps-id-test-patterns` (testmain + lifecycle + port conflict tests)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Implement `apps_ps_id_test_patterns/apps_ps_id_test_patterns.go`. Three
  sub-checks per PS-ID:
  1. `testmain_test.go` at package root
  2. At least one `*_lifecycle_test.go` file at package root
  3. At least one `*_port_conflict_test.go` file at package root

  `knownExclusions` for testmain: `sm-kms`, `identity-authz`, `identity-idp`, `identity-rs`,
  `identity-rp`, `identity-spa` (6 PS-IDs missing — excluded at launch).

  `knownExclusions` for lifecycle: `identity-rs`, `identity-rp`, `identity-spa` (3 PS-IDs).

  `knownExclusions` for port conflict: `identity-authz`, `identity-idp`, `identity-rs`,
  `identity-rp`, `identity-spa` (5 PS-IDs).

  The linter uses `os.ReadDir` to scan for glob-like patterns (`*_lifecycle_test.go`), checking
  if any directory entry name has the required suffix.
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Three separate sub-check functions (testmain, lifecycle, port_conflict), each with own exclusion list
  - [ ] Uses `os.ReadDir` for glob-style suffix matching (NOT filepath.Glob to avoid path issues)
  - [ ] `sm-im`, `jose-ja`, `pki-ca`, `skeleton-template` included in all three checks
  - [ ] `sm-kms` excluded from testmain check only (has lifecycle + port_conflict)
  - [ ] All 5 identity services excluded from port_conflict check
  - [ ] `identity-rs`, `identity-rp`, `identity-spa` excluded from lifecycle check
  - [ ] Test: pass case per sub-check
  - [ ] Test: fail case per sub-check (missing file)
  - [ ] Test: exception case (excluded PS-ID passes without required file)
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` still passes
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_test_patterns/apps_ps_id_test_patterns.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_test_patterns/apps_ps_id_test_patterns_test.go` (NEW)
- **Evidence**: `test-output/phase2/task-2.4-coverage.log`

### Task 2.5: Linter `apps-product-no-service-dirs` (no PS-ID service subdirs in products)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Implement `apps_product_no_service_dirs/apps_product_no_service_dirs.go`. For
  each product in `AllProducts()`, scan its `internal/apps/{PRODUCT}/` directory for subdirectories
  whose names match a known PS-ID service suffix (i.e., `ps.Service` fields for that product's
  PS-IDs). If found, report a violation.

  **Exception whitelist for `identity/`** (per ENG-HANDBOOK §G.1.1): The following subdirectories
  in `internal/apps/identity/` are explicitly whitelisted as shared libraries (NOT service code):
  `apperr`, `config`, `domain`, `email`, `issuer`, `jobs`, `mfa`, `repository`, `rotation`.
  The service suffixes for identity PS-IDs are: `authz`, `idp`, `rs`, `rp`, `spa` — these would
  be violations, but none currently exist as subdirectories in `internal/apps/identity/`.

  **Known violations at launch**: `sm/kms/`, `sm/im/`, `jose/ja/`, `pki/ca/`, `skeleton/template/`
  ARE service-named subdirs inside their respective products. Decision 2 (conservative scope) means
  these are excluded from the initial linter scope via a `knownServiceDirExceptions` map. A TODO
  comment marks each exception for resolution in a follow-on plan.

  The linter derives expected service suffixes per product from `AllProductServices()` by filtering
  on `ps.Product == product.ID`.
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Uses `AllProducts()` and `AllProductServices()` to derive service suffixes per product
  - [ ] `knownServiceDirExceptions` map lists all currently-accepted violations with TODO comments
  - [ ] `identity/` shared packages (`apperr`, `config`, `domain`, `email`, `issuer`, `jobs`, `mfa`, `repository`, `rotation`) are NOT service suffixes — they pass automatically
  - [ ] Test: pass case (identity/ shared packages pass correctly)
  - [ ] Test: fail case (synthetic `sm/kms/` found when not in exceptions list)
  - [ ] Test: exception case (sm/kms/ in exceptions list — passes)
  - [ ] Test: new violation case (hypothetical `identity/authz/` dir is caught)
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` still passes (all current violations are excepted)
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_product_no_service_dirs/apps_product_no_service_dirs.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_product_no_service_dirs/apps_product_no_service_dirs_test.go` (NEW)
- **Evidence**: `test-output/phase2/task-2.5-coverage.log`
- **GAP Note**: `sm/kms/`, `sm/im/`, `jose/ja/`, `pki/ca/`, `skeleton/template/` are current violations
  in `knownServiceDirExceptions`. These are follow-on work — see Gap Tasks section.

### Task 2.6: Linter `apps-suite-required-files` (cryptoutil.go + cryptoutil_test.go)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Implement `apps_suite_required_files/apps_suite_required_files.go`. Uses
  `AllSuites()` to get the suite ID (currently only `cryptoutil`) and verifies that
  `internal/apps/{SUITE}/{SUITE}.go` and `internal/apps/{SUITE}/{SUITE}_test.go` exist. Simple
  registry-driven pattern matching `product_structure.go`.
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Uses `AllSuites()` from registry (future-proof for additional suites)
  - [ ] Checks `{SUITE}.go` and `{SUITE}_test.go` for each suite
  - [ ] Error message format: `"cryptoutil: missing required file internal/apps/cryptoutil/cryptoutil.go"`
  - [ ] Test: pass case (cryptoutil.go and cryptoutil_test.go exist)
  - [ ] Test: fail case (synthetic root missing cryptoutil.go)
  - [ ] Test: fail case (synthetic root missing cryptoutil_test.go)
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` still passes (0 violations in current codebase)
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_suite_required_files/apps_suite_required_files.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_suite_required_files/apps_suite_required_files_test.go` (NEW)
- **Evidence**: `test-output/phase2/task-2.6-coverage.log`

### Task 2.7: Extend service_structure.go to Use Registry

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.1 (to coordinate overlap)
- **Description**: Refactor `service_structure.go` to use `AllProductServices()` instead of the
  hardcoded `knownServices` slice. The hardcoded list currently excludes `identity-authz` and
  `identity-idp` as "legacy". The refactored version will include all 10 PS-IDs via the registry,
  using a `knownExclusions` map to track the identity-authz/idp exclusions explicitly with TODO
  comments. Remove the `ServiceDef` struct and `knownServices` var; replace with registry iteration.
- **Acceptance Criteria**:
  - [ ] `knownServices` hardcoded slice removed
  - [ ] `AllProductServices()` used to iterate PS-IDs
  - [ ] `identity-authz` and `identity-idp` listed in `knownExclusions` with TODO comments
  - [ ] Existing tests updated to reflect registry-driven behavior
  - [ ] New test covers the case where a PS-ID in the registry has no `knownExclusions` entry
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/service_structure/...` passes
  - [ ] Coverage ≥98% for `service_structure` package
  - [ ] `go run ./cmd/cicd-lint lint-fitness` still passes
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/service_structure/service_structure.go` (MODIFIED)
  - `internal/apps-tools/cicd_lint/lint_fitness/service_structure/service_structure_test.go` (MODIFIED)
- **Evidence**: `test-output/phase2/task-2.7-coverage.log`

### Task 2.8: Phase 2 Coverage Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.1-2.7
- **Description**: Run coverage across all new linter packages and verify ≥98% for each. Run
  mutation testing. Run race detector. Archive results.
- **Acceptance Criteria**:
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_required_files/...` ≥98%
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_server_package/...` ≥98%
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_swagger_presence/...` ≥98%
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_test_patterns/...` ≥98%
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/apps_product_no_service_dirs/...` ≥98%
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/apps_suite_required_files/...` ≥98%
  - [ ] `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/service_structure/...` ≥98%
  - [ ] `go test -race -count=2 ./internal/apps-tools/cicd_lint/lint_fitness/...` passes
  - [ ] Coverage results archived in `test-output/phase2/coverage-summary.log`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase2/coverage-summary.log`, `test-output/phase2/race-detector.log`

---

## Phase 3: Registration, Integration & Knowledge Propagation

**Phase Objective**: Register all 6 new linters, update YAML manifest, validate end-to-end, and
propagate lessons to permanent artifacts.

### Task 3.1: Register Linters in lint_fitness.go

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Add 6 new import statements and 6 new entries to the `registeredLinters` slice
  in `lint_fitness.go`. Follow the existing comment style: `// New fitness checks (added in Phase 2
  of framework-v17).` Register in alphabetical order within the new block.
- **Acceptance Criteria**:
  - [ ] 6 import aliases added (snake_case package → camelCase alias per import alias convention)
  - [ ] 6 entries added to `registeredLinters` in correct position with comment
  - [ ] Entry names match YAML manifest names (kebab-case)
  - [ ] `go build ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` runs all 74 linters
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/lint_fitness.go` (MODIFIED)
- **Evidence**: `test-output/phase3/task-3.1-build.log`

### Task 3.2: Update lint-fitness-registry.yaml

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Description**: Add 6 new entries to `lint-fitness-registry.yaml`. Each entry needs: `name`
  (kebab-case), `directory` (snake_case), `description` (one sentence), `category`. Category for
  all 6 new linters is `architecture`.
- **Acceptance Criteria**:
  - [ ] 6 entries added in alphabetical order by name
  - [ ] `apps-ps-id-required-files` entry added
  - [ ] `apps-ps-id-server-package` entry added
  - [ ] `apps-ps-id-swagger-presence` entry added
  - [ ] `apps-ps-id-test-patterns` entry added
  - [ ] `apps-product-no-service-dirs` entry added
  - [ ] `apps-suite-required-files` entry added
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (`fitness-registry-completeness` validates manifest vs filesystem)
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml` (MODIFIED)
- **Evidence**: `test-output/phase3/task-3.2-fitness.log`

### Task 3.3: Update fitness_registry_completeness Tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.2
- **Description**: The `fitness_registry_completeness` tests contain hard-coded expected counts
  (e.g., "all 68 sub-linters"). Update test assertions to expect 74 sub-linters (68 + 6 new).
- **Acceptance Criteria**:
  - [ ] `fitness_registry_completeness_test.go` updated with new expected count (74)
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/fitness_registry_completeness/...` passes
  - [ ] Coverage still ≥98%
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/fitness_registry_completeness/fitness_registry_completeness_test.go` (MODIFIED)
- **Evidence**: `test-output/phase3/task-3.3-coverage.log`

### Task 3.4: End-to-End Fitness Run Validation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1-3.3
- **Description**: Run the full fitness suite and verify all 74 linters pass. Archive output.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness -q` shows "PASS" for all linters
  - [ ] Output shows 74 linters (68 existing + 6 new)
  - [ ] Output archived in `test-output/phase3/fitness-full-run.log`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase3/fitness-full-run.log`

### Task 3.5: Update ENG-HANDBOOK.md §9.11.1 Fitness Linter Catalog

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.4
- **Description**: Add 6 new entries to the fitness linter catalog in ENG-HANDBOOK.md §9.11.1.
  Follow the existing format (bullet list with linter name, directory name, and one-sentence
  description). Also update the linter count in any summary text (e.g., "68 linters" → "74 linters").
- **Acceptance Criteria**:
  - [ ] 6 new entries added in alphabetical position within the catalog
  - [ ] Linter count updated in all occurrences in ENG-HANDBOOK.md
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes (no propagation drift)
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (unchanged)
- **Files**:
  - `docs/ENG-HANDBOOK.md` (MODIFIED)
- **Evidence**: `test-output/phase3/task-3.5-lint-docs.log`

### Task 3.6: Update target-structure.md §G.1.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.5
- **Description**: Update `docs/target-structure.md` section G.1.1 (internal/apps/ layout) to
  explicitly reference the new fitness linters as the enforcement mechanism for the structural
  invariants described there. Add a note: "Enforced by: `apps-ps-id-required-files`,
  `apps-ps-id-server-package`, `apps-ps-id-swagger-presence`, `apps-ps-id-test-patterns`,
  `apps-product-no-service-dirs`, `apps-suite-required-files`."
- **Acceptance Criteria**:
  - [ ] §G.1.1 references all 6 new linters
  - [ ] §G.1.2 references per-file invariants with linter names
  - [ ] `go run ./cmd/cicd-lint lint-docs` still passes
- **Files**:
  - `docs/target-structure.md` (MODIFIED)
- **Evidence**: `test-output/phase3/task-3.6-lint-docs.log`

### Task 3.7: Full Test Suite Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1-3.6
- **Description**: Run the full test suite to confirm no regressions. Archive results.
- **Acceptance Criteria**:
  - [ ] `go test ./...` exits 0 (all packages pass)
  - [ ] `go test -race -count=2 ./internal/apps-tools/cicd_lint/...` exits 0
  - [ ] `golangci-lint run ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] Results archived in `test-output/phase3/full-test-suite.log`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase3/full-test-suite.log`

### Task 3.8: Knowledge Propagation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.7
- **Description**: Review `lessons.md` from all three phase post-mortems. Apply insights to
  permanent artifacts: ENG-HANDBOOK.md, agents, skills, instructions, code, and tests. Run
  propagation check. Commit all updates.
- **Acceptance Criteria**:
  - [ ] `lessons.md` has content in all 3 phase sections (written during phase post-mortems)
  - [ ] `lessons.md` Executive Summary filled with phase links and outcomes
  - [ ] `lessons.md` Actions section filled with concrete follow-up items
  - [ ] ENG-HANDBOOK.md updated where lessons expose omissions (beyond the §9.11.1 catalog update in 3.5)
  - [ ] Instruction files updated if lessons expose instruction gaps
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
- **Files**:
  - `docs/framework-v17/lessons.md` (MODIFIED — execution agent fills content)
  - `docs/ENG-HANDBOOK.md` (possible additional changes)
  - `.github/instructions/*.instructions.md` (if lessons expose gaps)
- **Evidence**: `test-output/phase3/task-3.8-propagation.log`

---

## Phase 4: Template-Compliance Linters for cmd/ and internal/apps/

**Phase Objective**: Implement 6 template-compliance linters driven by the MANIFEST.yaml files and
cmd/ template `.go` files created during planning. These linters supersede the ad-hoc linters
retired in Phase 3 and enforce structural invariants via template substitution.

**Prerequisites** (already on disk from planning):
- `api/cryptosuite-registry/templates/cmd/__PS_ID__/main.go`
- `api/cryptosuite-registry/templates/cmd/__PRODUCT__/main.go`
- `api/cryptosuite-registry/templates/cmd/__SUITE__/main.go`
- `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
- `api/cryptosuite-registry/templates/internal/apps/__PRODUCT__/MANIFEST.yaml`
- `api/cryptosuite-registry/templates/internal/apps/__SUITE__/MANIFEST.yaml`
- `CICDTemplateExpansionKeyService = "__SERVICE__"` in `internal/shared/magic/magic_template.go`

### Task 4.1: Linter `apps-ps-id-template` (MANIFEST.yaml-driven PS-ID structure check)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete; Task 1.2 gap matrix
- **Description**: Implement `apps_ps_id_template/apps_ps_id_template.go`. Reads
  `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml` at runtime.
  For each PS-ID in `AllProductServices()`, substitutes `__SERVICE__` → `ProductService.Service`
  and `__PS_ID__` → `ProductService.PSID` in file name patterns. Checks that all
  `required_root_files` exist in `internal/apps/{PS-ID}/` and all `required_dirs` exist.
  Known exclusions mirror Task 2.4 exclusion lists at launch.
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Reads MANIFEST.yaml using `os.ReadFile` (not embedded FS — template is external)
  - [ ] `__SERVICE__` substitution correct (e.g., `sm-kms` → service=`kms`)
  - [ ] `__PS_ID__` substitution correct
  - [ ] Checks all 8 required_root_files from MANIFEST.yaml
  - [ ] Checks `server` required_dir from MANIFEST.yaml
  - [ ] Error message format: `"{PS-ID}: missing required file internal/apps/{PS-ID}/{file}"`
  - [ ] `knownExclusions` matches Phase 2 exclusion lists (consistent at launch)
  - [ ] Test: pass case (all required files and dirs present)
  - [ ] Test: fail case per each missing required_root_file type
  - [ ] Test: fail case (missing server/ dir)
  - [ ] Test: exception case (excluded PS-ID passes without testmain_test.go)
  - [ ] Coverage ≥98%
  - [ ] `golangci-lint run` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after registration
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template_test.go` (NEW)
- **Evidence**: `test-output/phase4/task-4.1-coverage.log`

### Task 4.2: Linter `apps-product-template` (MANIFEST.yaml-driven product structure check)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete; Task 2.5
- **Description**: Implement `apps_product_template/apps_product_template.go`. Reads
  `api/cryptosuite-registry/templates/internal/apps/__PRODUCT__/MANIFEST.yaml`. For each product
  in `AllProducts()`, substitutes `__PRODUCT__` → `Product.ID`. Checks required_root_files exist.
  For `forbidden_dir_patterns`, uses `AllProductServices()` to derive known service suffixes per
  product, then scans `internal/apps/{PRODUCT}/` for directories matching those suffixes.
  Maintains the identity/ shared-package whitelist from the MANIFEST.yaml EXCEPTION comment.
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Reads MANIFEST.yaml with `__PRODUCT__` substitution
  - [ ] Checks `{PRODUCT}.go` and `{PRODUCT}_test.go` required files
  - [ ] Service-dir detection uses `AllProductServices()` filtered by product
  - [ ] `identity/` shared packages whitelist: `apperr`, `config`, `domain`, `email`, `issuer`, `jobs`, `mfa`, `repository`, `rotation` — NOT flagged
  - [ ] Known violations (`sm/kms/`, `sm/im/`, `jose/ja/`, `pki/ca/`, `skeleton/template/`)
    registered in `knownServiceDirExceptions` with TODO comments
  - [ ] Test: pass case (identity/ shared packages pass correctly)
  - [ ] Test: fail case (missing `{PRODUCT}.go`)
  - [ ] Test: fail case (forbidden service dir detected in non-excepted product)
  - [ ] Test: exception case (known violation in exception list passes)
  - [ ] Test: identity whitelisted dirs do NOT trigger forbidden check
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after registration
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_product_template/apps_product_template.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_product_template/apps_product_template_test.go` (NEW)
- **Evidence**: `test-output/phase4/task-4.2-coverage.log`

### Task 4.3: Linter `apps-suite-template` (MANIFEST.yaml-driven suite structure check)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete
- **Description**: Implement `apps_suite_template/apps_suite_template.go`. Reads
  `api/cryptosuite-registry/templates/internal/apps/__SUITE__/MANIFEST.yaml`. For each suite in
  `AllSuites()`, substitutes `__SUITE__` → `Suite.ID`. Checks `{SUITE}.go` and `{SUITE}_test.go`
  exist. Supersedes `apps_suite_required_files` from Phase 2 (retire that linter after this one
  is registered).
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] Reads MANIFEST.yaml from template directory
  - [ ] `__SUITE__` substitution correct
  - [ ] Checks `cryptoutil.go` and `cryptoutil_test.go` for the suite
  - [ ] Test: pass case
  - [ ] Test: fail case (missing cryptoutil.go)
  - [ ] Test: fail case (missing cryptoutil_test.go)
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after registration
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_suite_template/apps_suite_template.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_suite_template/apps_suite_template_test.go` (NEW)
- **Evidence**: `test-output/phase4/task-4.3-coverage.log`

### Task 4.4: Linter `cmd-ps-id-template` (structural invariants for cmd/{PS-ID}/main.go)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete
- **Description**: Implement `cmd_ps_id_template/cmd_ps_id_template.go`. Checks structural
  invariants for `cmd/{PS-ID}/main.go` for all 10 PS-IDs. Reads file content (does NOT compile
  it). Checks: (1) file exists, (2) contains `package main`, (3) contains import path
  `cryptoutil/internal/apps/{PS-ID}`, (4) calls `os.Args[1:]` (NOT `os.Args`). Uses string
  matching, not AST parsing (string matching is sufficient for these structural invariants).
- **Acceptance Criteria**:
  - [ ] `Check(logger) error` and `CheckInDir(logger, rootDir) error` implemented
  - [ ] All 10 PS-IDs checked via `AllProductServices()`
  - [ ] File-existence check: `cmd/{PS-ID}/main.go`
  - [ ] `package main` check
  - [ ] Import path check: `cryptoutil/internal/apps/{PS-ID}`
  - [ ] `os.Args[1:]` check (CRITICAL: must NOT accept bare `os.Args`)
  - [ ] Error message format: `"{PS-ID}: cmd/{PS-ID}/main.go {reason}"`
  - [ ] Test: pass case (all 10 PS-IDs valid)
  - [ ] Test: fail case (missing main.go)
  - [ ] Test: fail case (wrong package declaration)
  - [ ] Test: fail case (missing import path)
  - [ ] Test: fail case (uses os.Args instead of os.Args[1:])
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after registration
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/cmd_ps_id_template/cmd_ps_id_template.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/cmd_ps_id_template/cmd_ps_id_template_test.go` (NEW)
- **Evidence**: `test-output/phase4/task-4.4-coverage.log`

### Task 4.5: Linter `cmd-product-template` (structural invariants for cmd/{PRODUCT}/main.go)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.4 (same pattern, just 5 products)
- **Description**: Same structure as Task 4.4 but for 5 products. Checks: (1) file exists,
  (2) `package main`, (3) import path `cryptoutil/internal/apps/{PRODUCT}`,
  (4) uses `os.Args[1:]` (NOT `os.Args`).
- **Acceptance Criteria**:
  - [ ] All 5 products checked via `AllProducts()`
  - [ ] Same 4 structural invariants as Task 4.4
  - [ ] Test: pass case, fail cases per invariant
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after registration
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/cmd_product_template/cmd_product_template.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/cmd_product_template/cmd_product_template_test.go` (NEW)
- **Evidence**: `test-output/phase4/task-4.5-coverage.log`

### Task 4.6: Linter `cmd-suite-template` (structural invariants for cmd/{SUITE}/main.go)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.4 (same pattern, 1 suite)
- **Description**: Same structure as Task 4.4 but for the suite. **CRITICAL difference**: suite
  main.go uses `os.Args` (full, including argv[0]) NOT `os.Args[1:]`. The linter must check for
  `os.Args` WITHOUT the `[1:]` slice for the suite only. This matches the suite's need for
  full argv for multi-level routing display.
- **Acceptance Criteria**:
  - [ ] Suite (`cryptoutil`) checked via `AllSuites()`
  - [ ] Checks `os.Args` (NOT `os.Args[1:]`) — OPPOSITE of PS-ID and product linters
  - [ ] Test: pass case (cryptoutil/main.go uses os.Args)
  - [ ] Test: fail case (suite main.go uses os.Args[1:])
  - [ ] Test: fail case (missing main.go)
  - [ ] Coverage ≥98%
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after registration
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/cmd_suite_template/cmd_suite_template.go` (NEW)
  - `internal/apps-tools/cicd_lint/lint_fitness/cmd_suite_template/cmd_suite_template_test.go` (NEW)
- **Evidence**: `test-output/phase4/task-4.6-coverage.log`

### Task 4.7: Register Phase 4 Linters + Update YAML + Update Count

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.1-4.6
- **Description**: Register all 6 Phase 4 linters in `lint_fitness.go` and `lint-fitness-registry.yaml`.
  Update `fitness_registry_completeness` test expected count. Retire `apps-suite-required-files`
  (superseded by `apps-suite-template`).
- **Acceptance Criteria**:
  - [ ] 6 new imports and entries in `lint_fitness.go`
  - [ ] 6 new entries in `lint-fitness-registry.yaml` (alphabetical)
  - [ ] `apps-suite-required-files` retired (removed from `lint_fitness.go` and YAML)
  - [ ] `fitness_registry_completeness` test count updated to reflect net change (80 total = 74 + 6 new - 0 retired; minus `apps-suite-required-files` if it was from Phase 2)
  - [ ] `go build ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/lint_fitness.go` (MODIFIED)
  - `internal/apps-tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml` (MODIFIED)
  - `internal/apps-tools/cicd_lint/lint_fitness/fitness_registry_completeness/fitness_registry_completeness_test.go` (MODIFIED)
- **Evidence**: `test-output/phase4/task-4.7-fitness.log`

### Task 4.8: Phase 4 Coverage + Race Detection Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.1-4.7
- **Description**: Run coverage and race detection across all Phase 4 packages. Archive results.
- **Acceptance Criteria**:
  - [ ] ≥98% coverage for each of: `apps_ps_id_template`, `apps_product_template`, `apps_suite_template`, `cmd_ps_id_template`, `cmd_product_template`, `cmd_suite_template`
  - [ ] `go test -race -count=2 ./internal/apps-tools/cicd_lint/lint_fitness/...` exits 0
  - [ ] Results archived in `test-output/phase4/`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase4/coverage-summary.log`, `test-output/phase4/race-detector.log`

---

## Phase 5: Conformance Migration — Fill All Gaps

**Phase Objective**: Transform all 10 PS-IDs, 5 products, and 1 suite to fully conform to the
canonical templates. After this phase, all `knownExclusions` lists are empty and every linter runs
in block-immediately mode with zero exceptions.

**Work order**: Fill PS-ID file gaps first (5.1-5.6), then resolve product service dirs (5.7),
then remove all knownExclusions (5.8), then validate (5.9).

### Task 5.1: Add testmain_test.go to sm-kms

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 4 complete
- **Description**: Add `internal/apps/sm-kms/testmain_test.go` following the pattern established
  in `sm-im`, `jose-ja`, `pki-ca`, `skeleton-template`. The TestMain must start a shared test
  server and database for all root-package tests in `sm-kms`.
- **Acceptance Criteria**:
  - [ ] `testmain_test.go` created with correct TestMain function
  - [ ] Follows `testdb.NewInMemorySQLiteDB(t)` + `testserver.StartAndWait` pattern
  - [ ] `go test ./internal/apps/sm-kms/...` passes
  - [ ] Existing sm-kms tests still pass (no regressions)
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after sm-kms removed from testmain exclusion
- **Files**:
  - `internal/apps/sm-kms/testmain_test.go` (NEW)
- **Evidence**: `test-output/phase5/task-5.1-sm-kms.log`

### Task 5.2: Add testmain_test.go + port_conflict_test.go to identity-authz

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 4 complete
- **Description**: Add `internal/apps/identity-authz/testmain_test.go` and
  `internal/apps/identity-authz/authz_port_conflict_test.go`. Port conflict test verifies
  deterministic failure when the authz service ports are already in use (pattern from sm-im).
- **Acceptance Criteria**:
  - [ ] `testmain_test.go` created following shared pattern
  - [ ] `authz_port_conflict_test.go` follows sm-im/jose-ja port conflict test pattern
  - [ ] `go test ./internal/apps/identity-authz/...` passes (including new files)
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes after identity-authz removed from exclusions
- **Files**:
  - `internal/apps/identity-authz/testmain_test.go` (NEW)
  - `internal/apps/identity-authz/authz_port_conflict_test.go` (NEW)
- **Evidence**: `test-output/phase5/task-5.2-identity-authz.log`

### Task 5.3: Add testmain_test.go + port_conflict_test.go to identity-idp

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2 (same pattern)
- **Description**: Same as Task 5.2 but for `identity-idp`.
- **Acceptance Criteria**:
  - [ ] `testmain_test.go` created
  - [ ] `idp_port_conflict_test.go` created
  - [ ] Tests pass; fitness linter passes after idp removed from exclusions
- **Files**:
  - `internal/apps/identity-idp/testmain_test.go` (NEW)
  - `internal/apps/identity-idp/idp_port_conflict_test.go` (NEW)
- **Evidence**: `test-output/phase5/task-5.3-identity-idp.log`

### Task 5.4: Add testmain + lifecycle + port_conflict tests to identity-rs

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2 (same patterns)
- **Description**: Add `testmain_test.go`, `rs_lifecycle_test.go`, and `rs_port_conflict_test.go`
  to `identity-rs`. Lifecycle test verifies server start/stop/graceful-shutdown pattern across
  dual public+admin ports (pattern from sm-im or jose-ja).
- **Acceptance Criteria**:
  - [ ] `testmain_test.go`, `rs_lifecycle_test.go`, `rs_port_conflict_test.go` created
  - [ ] Lifecycle test covers: start, health check, shutdown, verify stopped
  - [ ] Port conflict test covers: both public and admin port conflicts
  - [ ] Tests pass; fitness linters pass after rs removed from all three exclusion lists
- **Files**:
  - `internal/apps/identity-rs/testmain_test.go` (NEW)
  - `internal/apps/identity-rs/rs_lifecycle_test.go` (NEW)
  - `internal/apps/identity-rs/rs_port_conflict_test.go` (NEW)
- **Evidence**: `test-output/phase5/task-5.4-identity-rs.log`

### Task 5.5: Add swagger + testmain + lifecycle + port_conflict to identity-rp

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.4 (same patterns); requires confirmation that identity-rp has OpenAPI spec
- **Description**: Add `swagger.go`, `swagger_test.go`, `testmain_test.go`, `rp_lifecycle_test.go`,
  `rp_port_conflict_test.go` to `identity-rp`. If identity-rp does not yet have an OpenAPI spec,
  add a minimal swagger stub (consistent with how other services handle this). Investigate first.
- **Acceptance Criteria**:
  - [ ] Investigation confirms whether identity-rp has OpenAPI spec (or will have one)
  - [ ] `swagger.go` added (or swagger stub documented as intentional placeholder)
  - [ ] `swagger_test.go` added
  - [ ] `testmain_test.go`, `rp_lifecycle_test.go`, `rp_port_conflict_test.go` added
  - [ ] Tests pass; fitness linters pass after rp removed from all exclusion lists
- **Files**:
  - `internal/apps/identity-rp/swagger.go` (NEW)
  - `internal/apps/identity-rp/swagger_test.go` (NEW)
  - `internal/apps/identity-rp/testmain_test.go` (NEW)
  - `internal/apps/identity-rp/rp_lifecycle_test.go` (NEW)
  - `internal/apps/identity-rp/rp_port_conflict_test.go` (NEW)
- **Evidence**: `test-output/phase5/task-5.5-identity-rp.log`

### Task 5.6: Add swagger + testmain + lifecycle + port_conflict to identity-spa

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.5 (same patterns)
- **Description**: Same as Task 5.5 but for `identity-spa`.
- **Acceptance Criteria**:
  - [ ] Same checklist as Task 5.5 (adapted for identity-spa)
  - [ ] Tests pass; fitness linters pass after spa removed from all exclusion lists
- **Files**:
  - `internal/apps/identity-spa/swagger.go` (NEW)
  - `internal/apps/identity-spa/swagger_test.go` (NEW)
  - `internal/apps/identity-spa/testmain_test.go` (NEW)
  - `internal/apps/identity-spa/spa_lifecycle_test.go` (NEW)
  - `internal/apps/identity-spa/spa_port_conflict_test.go` (NEW)
- **Evidence**: `test-output/phase5/task-5.6-identity-spa.log`

### Task 5.7: Audit and Resolve Product Service Directories (GAP-E)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 8h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.1-5.6
- **Description**: Investigate and resolve the 5 service-named subdirectories inside product dirs:
  `sm/kms/`, `sm/im/`, `jose/ja/`, `pki/ca/`, `skeleton/template/`. For each:
  1. Determine if this is canonical code (not yet moved) or duplicated code (already in PS-ID dir)
  2. If duplicated: delete the product-level copy, update any remaining imports
  3. If canonical: move to PS-ID dir, update all imports, delete product-level copy
  This is the largest gap and requires careful investigation before any deletions.
- **Acceptance Criteria**:
  - [ ] Each of the 5 product service dirs investigated (finding documented in `test-output/phase5/gap-e-audit.md`)
  - [ ] Each dir either deleted (if fully mirrored in PS-ID dir) or moved (if canonical)
  - [ ] All imports updated after any moves
  - [ ] `go build ./...` exits 0 after each individual deletion/move
  - [ ] All existing tests pass: `go test ./...`
  - [ ] Zero service-named subdirs remain in any product dir (no more identity/ whitelist needed)
  - [ ] `knownServiceDirExceptions` map emptied in `apps-product-template` linter
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
- **Files**:
  - `internal/apps/sm/kms/` (DELETE or MOVE)
  - `internal/apps/sm/im/` (DELETE or MOVE)
  - `internal/apps/jose/ja/` (DELETE or MOVE)
  - `internal/apps/pki/ca/` (DELETE or MOVE)
  - `internal/apps/skeleton/template/` (DELETE or MOVE)
  - Affected linter files (MODIFIED — remove exceptions)
- **Evidence**: `test-output/phase5/gap-e-audit.md`, `test-output/phase5/task-5.7-build.log`

### Task 5.8: Remove All knownExclusions from Phase 2–4 Linters

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.1-5.7
- **Description**: After all gaps are filled in Tasks 5.1-5.7, remove all `knownExclusions`,
  `knownServiceDirExceptions`, and TODO-marked exception entries from every Phase 2-4 linter.
  Every linter should now run in strict block-immediately mode with no exceptions.
- **Acceptance Criteria**:
  - [ ] `knownExclusions` is empty (or the variable removed) in all Phase 2-4 linters
  - [ ] `knownServiceDirExceptions` map is empty in `apps-product-template`
  - [ ] No TODO comments referencing "gap-" or "follow-on" remain in linter code
  - [ ] `go build ./...` exits 0
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/...` exits 0
- **Files**:
  - All Phase 2-4 linter `.go` files (MODIFIED — remove exclusion maps)
- **Evidence**: `test-output/phase5/task-5.8-build.log`

### Task 5.9: Final Fitness Run — Zero Exceptions Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.8
- **Description**: Run the complete fitness suite. Verify all 80 linters pass with zero exceptions
  and zero known-gap suppressions. Archive final output.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness -q` shows "PASS" for all linters
  - [ ] Zero `knownExclusions` entries remain across all new linters (confirmed by grep)
  - [ ] `grep -r "knownExclusions\|knownServiceDirExceptions" internal/apps-tools/cicd_lint/lint_fitness/` returns only empty-slice/map declarations
  - [ ] Output archived in `test-output/phase5/final-fitness-run.log`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase5/final-fitness-run.log`

---

## Phase 6: Knowledge Propagation

**Phase Objective**: Apply all lessons learned to permanent artifacts. This phase is MANDATORY —
never skip it. Each lesson becomes a lasting improvement to documentation, instructions, or skills.

### Task 6.1: Update ENG-HANDBOOK.md §9.11.1 Fitness Linter Catalog (All 12 New Linters)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 5 complete; `lessons.md` Phase 4-5 sections filled
- **Description**: Add all 12 new linters from Phases 2-4 to the §9.11.1 catalog. Update total
  linter count. Also update §G.1.1 and §G.1.2 to reference the template-compliance linters as
  the enforcement mechanism for the structural invariants described there.
- **Acceptance Criteria**:
  - [ ] All 12 new linter entries added to §9.11.1 in alphabetical order
  - [ ] Linter total count updated everywhere it appears in ENG-HANDBOOK.md
  - [ ] §G.1.1 references: `apps-ps-id-template`, `apps-product-template`, `apps-suite-template`
  - [ ] §G.1.2 references: `cmd-ps-id-template`, `cmd-product-template`, `cmd-suite-template`
  - [ ] Retired linters marked as `[RETIRED in V17]` in §9.11.1 catalog
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/ENG-HANDBOOK.md` (MODIFIED)
- **Evidence**: `test-output/phase6/task-6.1-lint-docs.log`

### Task 6.2: Update Instruction Files + Skills for Template Pattern

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.1
- **Description**: Apply lessons from Phases 2-5 to instruction files and skills:
  1. Add `__SERVICE__` expansion key note to `.github/instructions/03-01.coding.instructions.md`
     (under "CICD Template Parameterization Invariant" or a new section)
  2. Update `.github/skills/fitness-function-gen/SKILL.md` to mention the MANIFEST.yaml-driven
     pattern as the preferred approach for structural linters
  3. Mirror all changes to `.claude/` counterparts
  4. Run `lint-skill-command-drift` and `lint-agent-drift` sub-linters to verify no drift
- **Acceptance Criteria**:
  - [ ] `03-01.coding.instructions.md` updated with `__SERVICE__` expansion key guidance
  - [ ] `.github/skills/fitness-function-gen/SKILL.md` updated with MANIFEST.yaml pattern
  - [ ] `.claude/skills/fitness-function-gen/SKILL.md` updated identically (drift = 0)
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] `golangci-lint run ./...` still exits 0
- **Files**:
  - `.github/instructions/03-01.coding.instructions.md` (MODIFIED)
  - `.github/skills/fitness-function-gen/SKILL.md` (MODIFIED)
  - `.claude/skills/fitness-function-gen/SKILL.md` (MODIFIED)
- **Evidence**: `test-output/phase6/task-6.2-lint-docs.log`

### Task 6.3: Final Full-Suite Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 6.1-6.2
- **Description**: Run the full test suite, linting, fitness, and docs checks. Confirm zero
  regressions and all gates green.
- **Acceptance Criteria**:
  - [ ] `go test ./...` exits 0
  - [ ] `go test -race -count=2 ./internal/apps-tools/cicd_lint/lint_fitness/...` exits 0
  - [ ] `golangci-lint run ./...` exits 0
  - [ ] `golangci-lint run --build-tags e2e,integration ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0 (all linters pass, zero exceptions)
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] Results archived in `test-output/phase6/`
- **Files**: None (verification only)
- **Evidence**: `test-output/phase6/final-verification.log`

### Task 6.4: Populate lessons.md Executive Summary + Actions

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.3
- **Description**: Complete the `lessons.md` Executive Summary (linking all 6 phase post-mortems)
  and Actions sections. Commit all V17 documentation artifacts with a final semantic commit.
- **Acceptance Criteria**:
  - [ ] `lessons.md` Executive Summary section filled (links to all 6 phase sections)
  - [ ] `lessons.md` Actions section lists concrete follow-on items with LOE and priority
  - [ ] `docs/framework-v17/tasks.md` Status header updated: "40 of 40 tasks complete (100%)"
  - [ ] `docs/framework-v17/plan.md` Status updated: "Complete"
  - [ ] Final commit pushed; CI/CD green
- **Files**:
  - `docs/framework-v17/lessons.md` (MODIFIED)
  - `docs/framework-v17/tasks.md` (MODIFIED — final status update)
  - `docs/framework-v17/plan.md` (MODIFIED — final status update)
- **Evidence**: `test-output/phase6/lessons-summary.log`

---

## Cross-Cutting Tasks

### Testing

- [ ] All new linter packages: ≥98% line coverage
- [ ] Race detector clean for all new packages
- [ ] Mutation testing ≥98% (run `gremlins unleash --tags=!integration ./internal/apps-tools/cicd_lint/lint_fitness/...`)
- [ ] `go test ./...` passes with `-shuffle=on`

### Code Quality

- [ ] `golangci-lint run ./...` exits 0
- [ ] `golangci-lint run --build-tags e2e,integration ./...` exits 0
- [ ] No new `//nolint:` directives without GitHub issue reference
- [ ] Import aliases follow `cryptoutil<Package>` convention for internal imports

### Documentation

- [ ] All 6 new linter packages have package-level doc comments explaining purpose and invariant
- [ ] `knownExclusions` lists have TODO comments referencing follow-on work
- [ ] ENG-HANDBOOK.md §9.11.1 updated (Task 3.5)
- [ ] `docs/target-structure.md` §G.1.1 updated (Task 3.6)

---

## Gap Tasks (Follow-On Work — NOT in V17 Scope)

These are known gaps discovered during gap analysis that are excluded from V17 enforcement
via `knownExclusions` lists. They represent technical debt for follow-on plans.

### GAP-A: Add swagger.go and swagger_test.go to identity-rp and identity-spa

- **Scope**: `internal/apps/identity-rp/swagger.go`, `internal/apps/identity-rp/swagger_test.go`,
  `internal/apps/identity-spa/swagger.go`, `internal/apps/identity-spa/swagger_test.go`
- **Blocker**: Requires understanding how these services serve their OpenAPI spec (may not have one yet)
- **LOE**: ~2h
- **Priority**: P2 — once added, remove from `knownExclusions` in `apps-ps-id-swagger-presence`

### GAP-B: Add testmain_test.go to sm-kms and all 5 identity services

- **Scope**: `internal/apps/sm-kms/testmain_test.go`, `internal/apps/identity-authz/testmain_test.go`,
  `internal/apps/identity-idp/testmain_test.go`, `internal/apps/identity-rs/testmain_test.go`,
  `internal/apps/identity-rp/testmain_test.go`, `internal/apps/identity-spa/testmain_test.go`
- **Blocker**: Requires understanding existing test setup for each service (some use inline setup)
- **LOE**: ~4h
- **Priority**: P2 — once added, remove from `knownExclusions` in `apps-ps-id-test-patterns`

### GAP-C: Add *_lifecycle_test.go to identity-rs, identity-rp, identity-spa

- **Scope**: `internal/apps/identity-rs/*_lifecycle_test.go`, etc.
- **Blocker**: Requires implementing lifecycle tests following the sm-kms/jose-ja pattern
- **LOE**: ~3h
- **Priority**: P2

### GAP-D: Add *_port_conflict_test.go to all 5 identity services

- **Scope**: 5 files under `internal/apps/identity-{authz,idp,rs,rp,spa}/`
- **Blocker**: Requires understanding port assignment for identity services
- **LOE**: ~3h
- **Priority**: P2

### GAP-E: Resolve service code in product directories

- **Scope**: `internal/apps/sm/kms/`, `internal/apps/sm/im/`, `internal/apps/jose/ja/`,
  `internal/apps/pki/ca/`, `internal/apps/skeleton/template/`
- **Blocker**: These may be the canonical implementation locations for these services
  (with `internal/apps/sm-kms/` etc. being wrappers). Requires investigation of whether
  the service-level entry point delegates to the product-level code or vice versa.
- **LOE**: ~8h (significant investigation + possible code reorganization)
- **Priority**: P1 — architectural debt; once resolved, tighten `apps-product-no-service-dirs`
- **Note**: This is the largest gap. The existing ENG-HANDBOOK explicitly documents the intended
  flat `internal/apps/{PS-ID}/` layout as the target, but the current codebase has dual locations.

### GAP-F: Add e2e/ directories to 5 PS-IDs

- **Scope**: `internal/apps/pki-ca/e2e/`, `internal/apps/identity-idp/e2e/`,
  `internal/apps/identity-rs/e2e/`, `internal/apps/identity-rp/e2e/`, `internal/apps/identity-spa/e2e/`
- **Blocker**: Requires Docker Compose E2E test implementation (significant work)
- **LOE**: ~20h across all 5 services
- **Priority**: P3 — long-term goal

### GAP-G: Add *_contract_test.go to 7 PS-IDs

- **Scope**: `sm-kms`, `sm-im`, `jose-ja`, `pki-ca`, `identity-rp`, `identity-spa`, `skeleton-template`
- **Blocker**: Requires implementing cross-service contract compliance tests per framework pattern
- **LOE**: ~6h
- **Priority**: P2

---

## Notes / Deferred Work

- GAP-E (service code in product dirs) is the most significant architectural debt item.
  The V17 plan deliberately excludes it from enforcement scope to avoid breaking CI while
  the investigation is pending. A dedicated framework-v18 plan should address it.
- The linter count in `fitness_registry_completeness` tests will need updating in Tasks 3.3 and 4.7.
  The current count is 68; it becomes 74 after Phase 3 (6 new) and ~80 after Phase 4 (6 more new,
  minus any retired linters like `apps-suite-required-files`). Exact count determined at implementation time.
- Mutation testing for new packages should be run on Linux CI/CD (gremlins v0.6.0 panics on Windows).

---

## Evidence Archive

- `test-output/phase1/` — Phase 1 gap analysis artifacts
  - `preflight-build.log` — Build health check output
  - `preflight-fitness.log` — Existing fitness run output
  - `gap-matrix.md` — Confirmed PS-ID gap matrix
  - `linter-specs.md` — Finalized linter specifications
- `test-output/phase2/` — Phase 2 implementation artifacts
  - `task-2.N-coverage.log` — Coverage output per linter (N=1-7)
  - `coverage-summary.log` — Combined coverage across all new packages
  - `race-detector.log` — Race detector output
- `test-output/phase3/` — Phase 3 validation artifacts
  - `task-3.1-build.log` — Build after registration
  - `task-3.2-fitness.log` — Fitness run after YAML update
  - `task-3.3-coverage.log` — Registry completeness test coverage
  - `task-3.4-fitness-full.log` — Full fitness suite output (74 linters)
  - `task-3.5-lint-docs.log` — lint-docs output after ENG-HANDBOOK update
  - `task-3.6-lint-docs.log` — lint-docs output after target-structure update
  - `full-test-suite.log` — `go test ./...` full output
  - `task-3.8-propagation.log` — Knowledge propagation verification
- `test-output/phase4/` — Phase 4 template-compliance linter artifacts
  - `task-4.1-coverage.log` — apps-ps-id-template coverage
  - `task-4.2-coverage.log` — apps-product-template coverage
  - `task-4.3-coverage.log` — apps-suite-template coverage
  - `task-4.4-coverage.log` — cmd-ps-id-template coverage
  - `task-4.5-coverage.log` — cmd-product-template coverage
  - `task-4.6-coverage.log` — cmd-suite-template coverage
  - `task-4.7-fitness.log` — Fitness run after Phase 4 registration
  - `coverage-summary.log` — Combined Phase 4 coverage
  - `race-detector.log` — Race detector output
- `test-output/phase5/` — Phase 5 conformance migration artifacts
  - `gap-e-audit.md` — Product service directory audit findings
  - `task-5.N-{ps-id}.log` — Per-PS-ID gap fill evidence (N=1-6)
  - `task-5.7-build.log` — Build after product service dir resolution
  - `task-5.8-build.log` — Build after knownExclusions removal
  - `final-fitness-run.log` — Final fitness run with zero exceptions
- `test-output/phase6/` — Phase 6 knowledge propagation artifacts
  - `task-6.1-lint-docs.log` — ENG-HANDBOOK update propagation check
  - `task-6.2-lint-docs.log` — Instruction/skill update propagation check
  - `final-verification.log` — Full test + lint + fitness + docs verification
  - `lessons-summary.log` — lessons.md final state
