# Tasks - Framework V17: internal/apps/ Structure Fitness Linters

**Status**: 0 of 21 tasks complete (0%)
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
- The linter count in `fitness_registry_completeness` tests will need updating in Task 3.3.
  The current count is 68; it becomes 74 after V17.
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
