# Implementation Plan - Framework V17: internal/apps/ Structure Fitness Linters

**Status**: Historical Snapshot
**Created**: 2026-04-26
**Last Updated**: 2026-04-26
**Purpose**: Expand `lint-fitness` to enforce structural consistency across all 16 directories under
`internal/apps/` — 1 suite (`cryptoutil`), 5 products (`identity`, `jose`, `pki`, `skeleton`, `sm`),
and 10 PS-IDs. Currently `lint-fitness` enforces `deployments/` and `configs/` structure but has no
systematic enforcement of `internal/apps/` layout, file naming, or required file presence. New fitness
linters will detect-and-error on violations, NOT generate files.

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

V17 adds new architecture fitness linters to `internal/apps-tools/cicd_lint/lint_fitness/` that
enforce structural invariants for `internal/apps/{SUITE|PRODUCT|PS-ID}/`. Like existing linters,
these detect deviations and return errors — they do NOT generate files.

The plan is organized into three phases:
- **Phase 1**: Research, gap analysis, and target structure definition (linter design only)
- **Phase 2**: Implement 6 new fitness linters covering PS-ID, product, and suite structure
- **Phase 3**: Register all new linters + update lint-fitness-registry.yaml + knowledge propagation

---

## Background

### What Already Exists

`lint-fitness` currently has partial coverage of `internal/apps/` structure:

| Existing Linter | What It Checks |
|-----------------|----------------|
| `service-structure` | 8 of 10 PS-IDs have `{SERVICE}.go`, `{SERVICE}_usage.go`, `server/server.go`, `server/config/config.go` — excludes `identity-authz` and `identity-idp` as "legacy" |
| `product-structure` | All 5 products have `{PRODUCT}.go` and `{PRODUCT}_test.go` |
| `product-wiring` | All products and PS-IDs have `cmd/{PS-ID}/main.go` entry points |
| `subcommand-completeness` | All 10 PS-IDs call `RouteService` in their entry point |
| `require-framework-naming` | No Go files import banned `internal/apps/template/` path |
| `legacy-dir-detection` | No `internal/apps/cipher/` legacy directories exist |

**Critical gap**: No linter currently enforces:
1. PS-IDs must have `swagger.go` (8 of 10 have it; `identity-rp`, `identity-spa` missing)
2. PS-IDs must have `testmain_test.go` at package root (5 of 10 have it; 5 missing)
3. PS-IDs must have `*_lifecycle_test.go` (7 of 10 have it; 3 missing)
4. PS-IDs must have `*_port_conflict_test.go` (5 of 10; 5 missing)
5. Products must NOT contain service-named subdirectories
6. Suite must have `cryptoutil.go` and `cryptoutil_test.go` (no suite-specific linter)
7. `server/testmain_test.go` pattern inconsistency across PS-IDs

### Research Findings — Current State vs Target

#### Suite (`internal/apps/cryptoutil/`)

| File/Dir | Current State | Target | Delta |
|----------|--------------|--------|-------|
| `cryptoutil.go` | EXISTS | REQUIRED | OK |
| `cryptoutil_test.go` | EXISTS | REQUIRED | OK |
| `e2e/` | MISSING | OPTIONAL (ENG-HANDBOOK G.1.1) | gap (informational) |

**Assessment**: Suite directory is minimal and structurally correct. No violations.

#### Products — Critical Violation Discovered

ENG-HANDBOOK §4.4.4 and G.1.1 are explicit: product directories contain ONLY product-level code
(`{PRODUCT}.go`, shared packages) — **NO service subdirectories**.

| Product | Service Subdirs Found | Violation? |
|---------|----------------------|------------|
| `sm/` | `im/`, `kms/` | YES — service code should be at `internal/apps/sm-im/`, `internal/apps/sm-kms/` |
| `jose/` | `ja/`, `unified/` | YES — `ja/` is service code; `unified/` is shared and acceptable |
| `pki/` | `ca/` | YES — service code should be at `internal/apps/pki-ca/` |
| `skeleton/` | `template/` | YES — service code should be at `internal/apps/skeleton-template/` |
| `identity/` | `apperr/`, `config/`, `domain/`, `email/`, `issuer/`, `jobs/`, `mfa/`, `repository/`, `rotation/` | OK — these are shared packages (documented exception in ENG-HANDBOOK G.1.1) |

**Root cause**: `sm/im/`, `sm/kms/`, `jose/ja/`, `pki/ca/`, `skeleton/template/` are service-named
subdirectories inside product directories. ENG-HANDBOOK explicitly documents this as a known
historical layout that predates the flat PS-ID pattern. These directories coexist with the
proper `internal/apps/sm-im/` etc. flat directories (which are the real service entry points).

**Enforcement approach**: The new `ps-id-service-dir-isolation` linter will flag service-named
subdirectories inside product directories. EXCEPTION: `identity/` shared packages are legitimate
(documented in ENG-HANDBOOK §G.1.1 as "shared packages: optional, varies").

#### Architectural Decision: PS-ID Root = CLI Only (Added 2026-04-26)

**DECISION**: The PS-ID root directory contains ONLY CLI integration files. All server
implementation, swagger, TestMain, lifecycle tests, and port conflict tests live in `server/`.

**Root file rule**: ALL files at the PS-ID root MUST start with the `{SERVICE}_` prefix.
- ALLOWED at root: `{SERVICE}.go`, `{SERVICE}_usage.go`, `{SERVICE}_cli_test.go`
- FORBIDDEN at root: `swagger.go`, `testmain_test.go`, `http_test.go`, `handlers_*.go`, etc.

**Rationale**: The PS-ID root package is the CLI entry point — it should only contain code that
tests and exercises the CLI dispatcher. All service implementation belongs in `server/` (which is
already the separate Go package for the admin/public server). Mixing CLI and server test code at
root confuses package boundaries and makes it unclear which tests require a running server.

**Impact on Phase 2 linters**: Linters checking `swagger.go`, `testmain_test.go`, lifecycle, and
port conflict tests now check `server/` paths, not PS-ID root paths.

**Impact on Phase 5**: Includes file MOVES (from root → `server/`) in addition to creating missing
files. Eight PS-IDs need swagger/swagger_test.go moved. Nine PS-IDs need `testmain_test.go` root
copy removed (most already have server/ version). Seven to nine PS-IDs need lifecycle/port_conflict
moved. See target-structure.md G.1.2 for the complete gap matrix.

#### PS-IDs — Detailed Gap Matrix (target: `server/` location for all non-CLI files)

| Invariant | sm-kms | sm-im | jose-ja | pki-ca | id-authz | id-idp | id-rs | id-rp | id-spa | skel-tmpl |
|-----------|--------|-------|---------|--------|----------|--------|-------|-------|--------|-----------|
| root `{SERVICE}.go` | OK | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| root `{SERVICE}_usage.go` | OK | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| root `{SERVICE}_cli_test.go` | OK | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| `server/server.go` | OK | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| `server/swagger.go` | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | **MISS** | **MISS** | MOVE |
| `server/swagger_test.go` | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | **MISS** | **MISS** | MOVE |
| `server/testmain_test.go` | **MISS** | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| `server/{SVC}_lifecycle_test.go` | MOVE | MOVE | MOVE | MOVE¹ | MOVE² | MOVE² | **MISS** | **MISS** | **MISS** | MOVE |
| `server/{SVC}_port_conflict_test.go` | MOVE | MOVE | MOVE | MOVE | **MISS** | **MISS** | **MISS** | **MISS** | **MISS** | MOVE |
| `e2e/` dir | OK | OK | OK | **MISS** | OK | **MISS** | **MISS** | **MISS** | **MISS** | OK |
| `*_contract_test.go` | **MISS** | **MISS** | **MISS** | **MISS** | OK | OK | OK | **MISS** | **MISS** | **MISS** |

¹ pki-ca has `server/server_lifecycle_test.go` — must rename to `ca_lifecycle_test.go`.
² identity-authz/idp have `service_lifecycle_test.go` at root — rename on move.

**MOVE** = file exists at PS-ID root, must be relocated to `server/`.
**MISS** = file does not exist anywhere, must be created in `server/`.

#### Severity Classification

| Category | Invariant | Severity | Rationale |
|----------|-----------|----------|-----------|
| REQUIRED | `{SERVICE}.go` | ERROR | Primary CLI entry point |
| REQUIRED | `{SERVICE}_usage.go` | ERROR | CLI usage string |
| REQUIRED | `server/server.go` | ERROR | Core admin server |
| REQUIRED | `server/swagger.go` | ERROR | OpenAPI serving; must live in server/ (not root) |
| REQUIRED | `server/testmain_test.go` | ERROR | Integration TestMain; missing from sm-kms |
| REQUIRED | `server/{SVC}_lifecycle_test.go` | ERROR | Lifecycle tests; 3 missing, 7 need move |
| REQUIRED | `server/{SVC}_port_conflict_test.go` | ERROR | Port conflict tests; 5 missing, 5 need move |
| INFORMATIONAL | `e2e/` directory | WARN | E2E not complete for 5 PS-IDs; log but don't block |
| INFORMATIONAL | `*_contract_test.go` | WARN | Contract tests incomplete for 7 PS-IDs; in progress |

**Initial linter scope**: Start with ERROR-level invariants only. WARN-level items are tracked as
GAP tasks in tasks.md and will be promoted to ERROR in a future plan once all PS-IDs conform.

---

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: None (pure stdlib + gopkg.in/yaml.v3 for YAML parsing, already used)
- **Pattern**: detect-and-error linters, never generators
- **Implementation home**: `internal/apps-tools/cicd_lint/lint_fitness/`
- **Registry**: `internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go` (AllProductServices, AllProducts, AllSuites)
- **Registration**: `internal/apps-tools/cicd_lint/lint_fitness/lint_fitness.go`
- **Manifest**: `internal/apps-tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml`

### Affected Files

New linter packages (6 new directories = 6 × 2 files each = 12 new files):
```
internal/apps-tools/cicd_lint/lint_fitness/
├── apps_ps_id_required_files/       # NEW: 2 files
│   ├── apps_ps_id_required_files.go
│   └── apps_ps_id_required_files_test.go
├── apps_ps_id_server_package/       # NEW: 2 files
│   ├── apps_ps_id_server_package.go
│   └── apps_ps_id_server_package_test.go
├── apps_product_no_service_dirs/    # NEW: 2 files
│   ├── apps_product_no_service_dirs.go
│   └── apps_product_no_service_dirs_test.go
├── apps_suite_required_files/       # NEW: 2 files
│   ├── apps_suite_required_files.go
│   └── apps_suite_required_files_test.go
├── apps_ps_id_test_patterns/        # NEW: 2 files
│   ├── apps_ps_id_test_patterns.go
│   └── apps_ps_id_test_patterns_test.go
└── apps_ps_id_swagger_presence/     # NEW: 2 files
    ├── apps_ps_id_swagger_presence.go
    └── apps_ps_id_swagger_presence_test.go
```

Modified files (4 existing):
```
internal/apps-tools/cicd_lint/lint_fitness/lint_fitness.go     # register 6 new linters
internal/apps-tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml  # add 6 entries
internal/apps-tools/cicd_lint/lint_fitness/service_structure/service_structure.go  # extend to all 10 PS-IDs
docs/ENG-HANDBOOK.md               # document new invariants in §9.11.1
```

**Total**: 12 new files + 4 modified files = 16 file changes.

---

## Phases

**Phase Status Legend**: `☐ TODO` | `🔄 IN PROGRESS` | `✅ COMPLETE` | `⏳ BLOCKED`

### Phase 1: Gap Analysis & Linter Design (2h) [Status: ☐ TODO]

**Objective**: Finalize the gap matrix, define exact acceptance criteria for each new linter,
and verify the test fixture data needed for linter tests.

- Enumerate all expected violations in the current codebase per linter
- Define allowed exceptions (e.g., `pki-ca`'s legacy subdir layout is intentional)
- Confirm which invariants are ERROR vs WARN level
- Document the exact file patterns each linter will check
- Create test fixture directories in `test-output/phase1/`

**Success**: Gap analysis complete; all 6 linter specifications written; no ambiguities remain.

**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned — what worked,
what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix tasks
immediately.

### Phase 2: Implement 6 New Fitness Linters (8h) [Status: ☐ TODO]

**Objective**: Implement and fully test all 6 new linters. Each linter follows the established
`Check(logger) error` / `CheckInDir(logger, rootDir) error` seam pattern used by all existing
fitness linters.

**Linter 1: `apps-ps-id-required-files`** (`apps_ps_id_required_files/`)
- Checks every PS-ID in `AllProductServices()` has: `{SERVICE}.go`, `{SERVICE}_usage.go`
- Extends current `service-structure` scope by using the registry instead of a hardcoded list
- Replaces/supersedes the hardcoded `knownServices` list in `service_structure.go`
- Tests: pass case (all required files present), fail case (missing entry file), fail case (missing usage file)

**Linter 2: `apps-ps-id-server-package`** (`apps_ps_id_server_package/`)
- Checks every PS-ID has `server/server.go` and `server/public_server.go`
- Exception: `sm-kms` uses `server/server.go` only (no `public_server.go` — legacy)
- Tests: pass case, fail case (missing server dir), fail case (missing public_server.go), exception case

**Linter 3: `apps-ps-id-swagger-presence`** (`apps_ps_id_swagger_presence/`)
- Checks every PS-ID has `swagger.go` and `swagger_test.go` in the `server/` subdirectory
- (Per architectural decision 2026-04-26: swagger lives in `server/`, NOT at PS-ID root)
- Current state: files exist at PS-ID root for 8 PS-IDs (MOVE needed); `identity-rp`, `identity-spa` missing entirely
- Tests: pass case, fail case (missing server/swagger.go), fail case (missing server/swagger_test.go)

**Linter 4: `apps-ps-id-test-patterns`** (`apps_ps_id_test_patterns/`)
- Checks every PS-ID has `testmain_test.go` in the `server/` subdirectory (NOT at PS-ID root)
- Checks every PS-ID has at least one `*_lifecycle_test.go` in `server/`
- Checks every PS-ID has at least one `*_port_conflict_test.go` in `server/`
- (Per architectural decision 2026-04-26: these tests live in `server/`, NOT at PS-ID root)
- Current state: `testmain_test.go` missing from `server/` for sm-kms; lifecycle missing for 3; port_conflict missing for 5
- Tests: pass case, fail case per invariant

**Linter 5: `apps-product-no-service-dirs`** (`apps_product_no_service_dirs/`)
- Checks each product directory does NOT contain a subdirectory whose name matches a
  known PS-ID service suffix (e.g., `sm/kms/` is banned because `kms` is a PS-ID service name)
- Exception: `identity/` shared packages (`apperr/`, `config/`, `domain/`, `email/`, `issuer/`,
  `jobs/`, `mfa/`, `repository/`, `rotation/`) are explicitly whitelisted (they are shared libs,
  not service code copies — documented in ENG-HANDBOOK §G.1.1)
- Uses `AllProductServices()` to derive known service suffixes per product
- Tests: pass case (identity shared packages pass), fail case (sm/kms/ detected), exception case

**Linter 6: `apps-suite-required-files`** (`apps_suite_required_files/`)
- Checks the suite directory (`internal/apps/cryptoutil/`) has `cryptoutil.go` and
  `cryptoutil_test.go` (the `AllSuites()` registry accessor drives this)
- Tests: pass case, fail case (missing cryptoutil.go), fail case (missing test file)

**Per-linter quality gates**:
- ≥98% line coverage (infrastructure/utility standard)
- Build clean: `go build ./...`
- Lint clean: `golangci-lint run ./internal/apps-tools/cicd_lint/lint_fitness/...`
- `go run ./cmd/cicd-lint lint-fitness` runs all linters clean against current codebase

**Success**: All 6 linters implemented; each has ≥98% coverage; lint-fitness passes against current
codebase (violations are expected and captured as known gaps, not linter failures).

**CRITICAL NOTE**: Linters MUST NOT fail against the current codebase at launch time. Violations
in the current codebase MUST be pre-registered as known exceptions via an allowlist or suppressed
via a "baseline" mode. The preferred pattern (used by existing linters) is to document current gaps
as informational log messages and return success — making linters immediately runnable without
requiring all gaps to be fixed first. Alternatively, define a `knownGaps` list of PS-IDs that are
exempt during the migration period. Task 2.1 specifies which approach to use for each linter.

**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

### Phase 3: Registration, Integration & Linter Retirement (3h) [Status: ☐ TODO]

**Objective**: Register all 6 new linters, update the YAML manifest, extend the existing
`service-structure` linter, retire obsolete linters, and validate end-to-end.

- Register all 6 linters in `lint_fitness.go` `registeredLinters` slice
- Add 6 entries to `lint-fitness-registry.yaml` manifest
- Extend `service_structure.go` to cover all 10 PS-IDs (remove "legacy" exclusion, use registry)
  OR deprecate it in favor of the new `apps-ps-id-required-files` linter
- Retire `service_structure` linter (superseded by `apps-ps-id-required-files` + `apps-ps-id-server-package`)
- Retire `product_structure` linter (superseded by `apps-product-template`)
- Retire `product_wiring` linter (superseded by `cmd-ps-id-template` + `cmd-product-template` + `cmd-suite-template`)
- Retire `subcommand_completeness` linter (superseded by `apps-ps-id-required-files`)
- Run `go run ./cmd/cicd-lint lint-fitness` — must pass with 0 errors
- Update `fitness_registry_completeness` test expectations (count changes after retirements)
- Update ENG-HANDBOOK.md §9.11.1 fitness linter catalog (add new, mark retired)

**Success**: `go run ./cmd/cicd-lint lint-fitness` passes; YAML manifest updated; redundant linters retired.

**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

### Phase 4: Template-Compliance Linters for cmd/ and internal/apps/ (6h) [Status: ☐ TODO]

**Objective**: Implement template-compliance linters that use the new MANIFEST.yaml and main.go
template files to enforce structural invariants across all cmd/ and internal/apps/ directories.
These replace the ad-hoc linters retired in Phase 3 with template-driven enforcement.

**What was done in planning** (already on disk):
- Added `CICDTemplateExpansionKeyService = "__SERVICE__"` to `internal/shared/magic/magic_template.go`
- Created `api/cryptosuite-registry/templates/cmd/__PS_ID__/main.go` — structural invariant documentation
- Created `api/cryptosuite-registry/templates/cmd/__PRODUCT__/main.go`
- Created `api/cryptosuite-registry/templates/cmd/__SUITE__/main.go`
- Created `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
- Created `api/cryptosuite-registry/templates/internal/apps/__PRODUCT__/MANIFEST.yaml`
- Created `api/cryptosuite-registry/templates/internal/apps/__SUITE__/MANIFEST.yaml`
- Updated `docs/target-structure.md` C, G.1.1, G.1.2 with rigid structure tables and gap matrix

**New linters to implement** (6 linters, each with `{name}.go` + `{name}_test.go`):

**Linter 7: `apps-ps-id-template`** (`apps_ps_id_template/`)
- Reads `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
- For each PS-ID in `AllProductServices()`, substitutes `__SERVICE__` → `ProductService.Service`
- Checks that all `required_root_files` exist in `internal/apps/{PS-ID}/`
- Checks that all `required_dirs` exist in `internal/apps/{PS-ID}/`
- Errors on any missing required file or dir
- Tests: pass case, fail for each missing file type, fail for missing dir

**Linter 8: `apps-product-template`** (`apps_product_template/`)
- Reads `api/cryptosuite-registry/templates/internal/apps/__PRODUCT__/MANIFEST.yaml`
- For each product in `AllProducts()`, substitutes `__PRODUCT__` → `Product.ID`
- Checks that all `required_root_files` exist in `internal/apps/{PRODUCT}/`
- Checks `forbidden_dir_patterns`: no subdirectory whose name matches a PS-ID service component
  (using `AllProductServices()` to build the forbidden-name set; identity/ exception whitelisted)
- Tests: pass case (identity/ passes), fail for missing required files, fail for forbidden dir

**Linter 9: `apps-suite-template`** (`apps_suite_template/`)
- Reads `api/cryptosuite-registry/templates/internal/apps/__SUITE__/MANIFEST.yaml`
- For the suite in `AllSuites()`, substitutes `__SUITE__` → `Suite.ID`
- Checks that all `required_root_files` exist in `internal/apps/{SUITE}/`
- Tests: pass case, fail for missing suite.go, fail for missing suite_test.go

**Linter 10: `cmd-ps-id-template`** (`cmd_ps_id_template/`)
- Checks structural invariants for `cmd/{PS-ID}/main.go` for all 10 PS-IDs:
  - File exists
  - `package main` declaration
  - Imports `cryptoutil/internal/apps/{PS-ID}` (contains the PS-ID import path)
  - Calls `os.Exit(...)` with `os.Args[1:]` (NOT `os.Args`)
- Tests: pass case, fail for missing file, fail for wrong package, fail for missing import

**Linter 11: `cmd-product-template`** (`cmd_product_template/`)
- Same structure as `cmd-ps-id-template` but for 5 products
- Tests: pass case, fail cases for missing file and structural invariants

**Linter 12: `cmd-suite-template`** (`cmd_suite_template/`)
- Same structure but for 1 suite; checks `os.Args` (not `os.Args[1:]`)
- Tests: pass case, fail for wrong args pattern

**Per-linter quality gates**: ≥98% line coverage; lint clean; `go run ./cmd/cicd-lint lint-fitness` passes.

**Success**: All 6 template-compliance linters implemented; lint-fitness count updated to match.

**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

### Phase 5: Conformance Migration — Fill All Gaps (20h) [Status: ☐ TODO]

**Objective**: Transform all 10 PS-IDs, 5 products, and 1 suite to fully conform to the canonical
templates. This phase has two categories of work:

1. **FILE MOVES** — Relocate files from PS-ID root to `server/` (architectural decision 2026-04-26)
2. **FILE CREATES** — Add missing files in `server/` for PS-IDs that lack them entirely

After this phase, all `knownExclusions` are empty and every linter runs in block-immediately mode.

**PS-ID file moves** (root → `server/`; see gap matrix in docs/target-structure.md G.1.2):

| PS-ID | Files to MOVE from root to `server/` | Notes |
|-------|--------------------------------------|-------|
| `sm-kms` | `swagger.go`, `swagger_test.go`, `kms_lifecycle_test.go`, `kms_port_conflict_test.go` | Also CREATE `server/testmain_test.go` |
| `sm-im` | `swagger.go`, `swagger_test.go`, `im_lifecycle_test.go`, `im_port_conflict_test.go` | Also remove redundant root `testmain_test.go` |
| `jose-ja` | `swagger.go`, `swagger_test.go`, `ja_lifecycle_test.go`, `ja_port_conflict_test.go` | Also remove redundant root `testmain_test.go` |
| `pki-ca` | `swagger.go`, `swagger_test.go`, `ca_lifecycle_test.go`, `ca_port_conflict_test.go` | Also rename `server/server_lifecycle_test.go` → `ca_lifecycle_test.go`; remove root `testmain_test.go` |
| `identity-authz` | `swagger.go`, `swagger_test.go`, `service_lifecycle_test.go` → rename `authz_lifecycle_test.go` | Also remove root `testmain_test.go`; CREATE `authz_port_conflict_test.go` |
| `identity-idp` | `swagger.go`, `swagger_test.go`, `service_lifecycle_test.go` → rename `idp_lifecycle_test.go` | Also remove root `testmain_test.go`; CREATE `idp_port_conflict_test.go` |
| `identity-rs` | `swagger.go`, `swagger_test.go`, `service.go`, `service_admin_test.go`, `service_test.go`, `validator.go` | Also remove root `testmain_test.go`; CREATE `rs_lifecycle_test.go`, `rs_port_conflict_test.go` |
| `skeleton-template` | `swagger.go`, `swagger_test.go`, `template_lifecycle_test.go`, `template_port_conflict_test.go` | Also remove redundant root `testmain_test.go` |

**PS-ID files to CREATE** (do not exist anywhere):

| PS-ID | Files to CREATE in `server/` |
|-------|------------------------------|
| `sm-kms` | `testmain_test.go` |
| `identity-authz` | `authz_port_conflict_test.go` |
| `identity-idp` | `idp_port_conflict_test.go` |
| `identity-rs` | `rs_lifecycle_test.go`, `rs_port_conflict_test.go` |
| `identity-rp` | `swagger.go`, `swagger_test.go`, `rp_lifecycle_test.go`, `rp_port_conflict_test.go` |
| `identity-spa` | `swagger.go`, `swagger_test.go`, `spa_lifecycle_test.go`, `spa_port_conflict_test.go` |

**Non-CLI root files to move** (identity-authz/idp have large amounts of server code at root):

| PS-ID | Root files to move to `server/` |
|-------|----------------------------------|
| `identity-authz` | All `handlers_*.go`, `routes.go`, `service.go`, `middleware.go`, `cleanup.go`, and associated `*_test.go` |
| `identity-idp` | All `handlers_*.go`, `routes.go`, `service.go`, `middleware.go`, `random.go`, and associated `*_test.go` |
| `identity-rs` | `service.go`, `validator.go`, `service_admin_test.go`, `service_test.go` |
| `sm-im` | `http_test.go`, `http_errors_test.go`, `response_body_test.go`, `im_database_test.go`, `im_server_lifecycle_test.go` |

**Product gaps to fix** (service-named subdirectories — forbidden by `apps-product-template`):

| Product | Action |
|---------|--------|
| `sm/im/`, `sm/kms/` | Audit for unique code not in `internal/apps/sm-im/`, `sm-kms/`; delete if redundant |
| `jose/ja/` | Audit and delete if redundant with `internal/apps/jose-ja/` |
| `pki/ca/` | Audit and delete if redundant with `internal/apps/pki-ca/` |
| `skeleton/template/` | Audit and delete if redundant with `internal/apps/skeleton-template/` |

**After all gaps fixed**:
- Remove all `knownExclusions` from Phase 2-4 linters
- Verify `go run ./cmd/cicd-lint lint-fitness` passes with zero exceptions
- All 12 new linters run in block-immediately mode

**Answers to quizme-v1.md required before starting**: Q1 (canonical vs duplicate product dirs)
answers Task 5.7 (delete vs move); Q2 (identity-rp/spa OpenAPI) answers Task 5.6; Q3 (identity
TestMain pattern) answers Tasks 5.2-5.6.

**Success**: Zero `knownExclusions` remain; all 16 directories conform to canonical templates;
PS-ID roots contain ONLY `{SERVICE}_`-prefixed CLI files.

**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

### Phase 6: Knowledge Propagation (2h) [Status: ☐ TODO]

**Objective**: Apply lessons learned to permanent artifacts. NEVER skip this phase.

- Review `lessons.md` from all prior phases
- Update ENG-HANDBOOK.md §9.11.1 fitness linter catalog with all 12 new entries
- Update ENG-HANDBOOK.md §G.1.1 and G.1.2 to reference template enforcement
- Update `.github/instructions/03-01.coding.instructions.md` with `__SERVICE__` expansion key note
- Update `.github/skills/fitness-function-gen/SKILL.md` with MANIFEST.yaml pattern
- Verify propagation integrity: `go run ./cmd/cicd-lint lint-docs`
- Commit all artifact updates with separate semantic commits per artifact type

**Success**: All artifact updates committed; propagation check passes; lessons permanently codified.

---

## Executive Decisions

### Decision 1: Baseline Mode vs Exception Allowlist

**Options**:
- A: Baseline mode — linters log violations as INFO but always return nil (never block CI)
- B: Exception allowlist — known gaps listed as `knownExceptions []string`; linter errors only for non-listed violations
- C: Block-immediately — linters fail on any violation, requiring immediate fix before merge ✓ **SELECTED**
- D: Graduated — new linters start in warn mode, promoted to error after all PS-IDs are fixed

**Decision**: Option C selected — block-immediately.

**Rationale**: The existing linters all block immediately. Adopting a soft mode creates two categories
of linters which complicates the mental model. The current gaps (swagger.go missing in identity-rp/spa;
testmain missing in 5 PS-IDs) SHOULD be fixed as part of this plan's follow-on work, or the linters
SHOULD be scoped to only the PS-IDs that already conform. See Decision 2.

**Impact**: Linters will be scoped to the PS-IDs that currently conform at launch; as other PS-IDs
are fixed, they are added to the scope. This matches how `service_structure.go` currently works
(it excludes identity-authz and identity-idp as "legacy").

### Decision 2: Scope at Launch vs Full Coverage

**Options**:
- A: Enforce only the PS-IDs that already conform (conservative launch) ✓ **SELECTED**
- B: Enforce all 10 PS-IDs immediately, accepting that CI will fail until all gaps are fixed
- C: Enforce all 10 PS-IDs with an allowlist of known-gap suppressions
- D: Skip enforcement entirely; document gaps only

**Decision**: Option A selected — conservative launch, expand scope as gaps are fixed.

**Rationale**: Same pattern as existing `service_structure.go` which excludes identity-authz/idp.
Each linter documents its current scope in a `knownExclusions` list. Tasks in Phase 2 explicitly
specify which PS-IDs are included at launch for each linter. A follow-on plan (V18 or later) will
expand scope as gaps are fixed.

**Impact**: New linters immediately useful without breaking CI; gap matrix is documented in tasks.md
for follow-on work.

### Decision 3: Replace or Extend service_structure.go

**Options**:
- A: Extend `service_structure.go` to use registry and cover all 10 PS-IDs ✓ **SELECTED**
- B: Deprecate `service_structure.go` and replace with new `apps_ps_id_required_files`
- C: Leave `service_structure.go` as-is; new linters add complementary checks
- D: Merge all PS-ID checks into a single monolithic linter

**Decision**: Option A selected — extend `service_structure.go` to use `AllProductServices()` from
the registry instead of a hardcoded list, removing the manual sync burden.

**Rationale**: The registry is the SSOT. Using `AllProductServices()` ensures new PS-IDs are
automatically included in structural checks without manual code updates.

**Impact**: `service_structure.go` becomes registry-driven; `knownServices` hardcoded list is removed.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Linters fail against current codebase, blocking CI | Medium | High | Use conservative scope (Decision 2); only check PS-IDs that already conform |
| `fitness_registry_completeness` fails after adding 6 linters | Low | Medium | Update YAML manifest and Go registration atomically in Phase 3 |
| `pki-ca` has highly non-standard layout (16+ subdirs vs 2-3 for others) | High | Low | Document pki-ca as a permanent exception; apply only applicable checks |
| `identity-authz/idp` flat-file structure won't pass server package checks | Medium | Medium | Exclude from server package linter scope at launch; address in follow-on |
| Go test parallelism causes flaky file-system checks | Low | Low | All linters use `os.Stat` (read-only, parallel-safe) |

---

## Quality Gates — MANDATORY

### Per-Task Quality Gates (ALL must pass before marking any task complete)

- ✅ All tests pass: `go test ./internal/apps-tools/cicd_lint/lint_fitness/...`
- ✅ Build clean: `go build ./...`
- ✅ Lint clean: `golangci-lint run ./...` AND `golangci-lint run --build-tags e2e,integration ./...`
- ✅ Coverage: `go test -coverprofile=coverage.out ./internal/apps-tools/cicd_lint/lint_fitness/...` shows ≥98% per new package
- ✅ `go run ./cmd/cicd-lint lint-fitness` passes (zero errors against current codebase)
- ✅ `go run ./cmd/cicd-lint lint-fitness -q` passes (summary output only)

### Per-Phase Quality Gates

- ✅ Phase 1: Gap analysis document complete in `test-output/phase1/gap-analysis.md`
- ✅ Phase 2: All 6 new linter packages have ≥98% coverage; lint-fitness passes
- ✅ Phase 3: `go run ./cmd/cicd-lint lint-fitness` passes end-to-end; YAML manifest updated; ENG-HANDBOOK updated
- ✅ Phase 4: All 6 template-compliance linters have ≥98% coverage; lint-fitness passes with new count
- ✅ Phase 5: All PS-ID gaps filled; all product service dirs resolved; zero knownExclusions remain
- ✅ Phase 6: All 12 new linters documented in ENG-HANDBOOK §9.11.1; lint-docs passes; lessons.md complete

### Overall Project Quality Gates

- ✅ `go run ./cmd/cicd-lint lint-fitness` passes with all ~80 linters (was 68; +12 new in Phases 2-4)
- ✅ `go run ./cmd/cicd-lint lint-docs` passes (no propagation drift)
- ✅ Coverage ≥98% for all new packages (12 packages)
- ✅ Mutation testing ≥98% for all new packages (run via `gremlins unleash`)
- ✅ Race detector clean: `go test -race -count=2 ./internal/apps-tools/cicd_lint/lint_fitness/...`
- ✅ Zero `knownExclusions` entries after Phase 5 (confirmed by grep)
- ✅ Zero service-named subdirectories in product dirs after Phase 5

---

## Success Criteria

- [ ] 12 new fitness linter packages created with ≥98% coverage and ≥98% mutation (Phases 2-4)
- [ ] All 12 linters registered in `lint_fitness.go` and `lint-fitness-registry.yaml`
- [ ] `service_structure.go` refactored to use `AllProductServices()` (registry-driven)
- [ ] `go run ./cmd/cicd-lint lint-fitness` passes with zero errors and zero exceptions
- [ ] All 10 PS-IDs fully conform to `__PS_ID__/MANIFEST.yaml` (no missing required files)
- [ ] All 5 products free of service-named subdirectories (no `sm/kms/`, `jose/ja/`, etc.)
- [ ] ENG-HANDBOOK.md §9.11.1 updated with all 12 new linter entries
- [ ] `docs/target-structure.md` §G.1.1 and §G.1.2 reference enforcement mechanism
- [ ] `lessons.md` populated with post-mortem from all 6 phases

---

## Quizme Round 1 (2026-04-26)

**Note**: Questions researched and answered autonomously via codebase inspection (no human input required — all answers derivable from current source).

### Q1: Are product service dirs canonical or duplicates? (Was: Blocks Task 5.8)

**Findings** (from `ls internal/apps/{sm/kms,sm/im,jose/ja,pki/ca,skeleton/template}/`):

| Directory | Contents | Verdict |
|-----------|----------|---------|
| `internal/apps/sm/kms/` | `kms_usage.go` only | **DUPLICATE** — thin usage stub; `sm-kms/` is canonical |
| `internal/apps/sm/im/` | `im_usage.go` only | **DUPLICATE** — thin usage stub; `sm-im/` is canonical |
| `internal/apps/jose/ja/` | `ja_usage.go` only | **DUPLICATE** — thin usage stub; `jose-ja/` is canonical |
| `internal/apps/pki/ca/` | `ca_usage.go` only | **DUPLICATE** — thin usage stub; `pki-ca/` is canonical |
| `internal/apps/skeleton/template/` | `template_usage.go` only | **DUPLICATE** — thin usage stub; `skeleton-template/` is canonical |

**Implication for Task 5.8**: All 5 product service dirs contain ONLY a `*_usage.go` file (duplicate of the same file already at the PS-ID root). Task 5.8 is a **safe delete** — no code moves, no import updates, low risk.

### Q2: Do identity-rp and identity-spa have OpenAPI specs? (Was: Blocks Tasks 5.5, 5.6)

**Findings** (from `public_server.go` inspection):

| Service | Endpoints Found | Verdict |
|---------|----------------|---------|
| `identity-rp` | `/health`, livez, readyz (health only; BFF proxies OAuth flows) | **Stub spec** — no domain API; needs placeholder `swagger.go` |
| `identity-spa` | `/health`, livez, readyz, `/config.json`, `/*` (static files) | **Stub spec** — static file server; needs placeholder `swagger.go` |

**Implication**: Both need a stub `swagger.go` serving a minimal placeholder OpenAPI spec (empty `paths: {}` with service metadata). Neither has oapi-codegen generated code.

### Q3: Which identity services use the shared server builder? (Was: Informs Tasks 5.2-5.6)

**Findings**: All 5 identity services already have `testmain_test.go` in `server/`. They use **custom per-service setup** (`NewFromConfig` + `waitForReady`), not the shared `testserver.StartAndWait` builder. Question is moot for Tasks 5.2-5.6 — no new testmain files needed for identity services.

For sm-kms (Task 5.1): use the **standard shared builder** (`testserver.StartAndWait`) since sm-kms uses the framework builder pattern.

## ENG-HANDBOOK.md Cross-References — MANDATORY

| Topic | ENG-HANDBOOK.md Section | Relevance |
|-------|------------------------|-----------|
| Testing Strategy | [Section 10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | ALL plans with implementation phases |
| Unit Testing | [Section 10.2](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy) | Linter unit tests |
| Coverage Targets | [Section 10.2.3](../../docs/ENG-HANDBOOK.md#1023-coverage-targets) | ≥98% infrastructure/utility |
| Test Seam Injection | [Section 10.2.4](../../docs/ENG-HANDBOOK.md#1024-test-seam-injection-pattern) | Injecting os.Stat/os.ReadDir in tests |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL plans |
| Code Quality | [Section 11.3](../../docs/ENG-HANDBOOK.md#113-code-quality-standards) | New linter code |
| Coding Standards | [Section 14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Go patterns |
| Version Control | [Section 14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | Commit strategy |
| Fitness Functions | [Section 9.11](../../docs/ENG-HANDBOOK.md#911-architecture-fitness-functions) | Implementation home |
| Fitness Linter Catalog | [Section 9.11.1](../../docs/ENG-HANDBOOK.md#9111-fitness-sub-linter-catalog) | Add 6 new entries |
| Service Architecture | [Section 5.1](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern) | PS-ID required structure |
| Plan Lifecycle | [Section 14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | ALL plans |
| Post-Mortem & Knowledge Propagation | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | ALL plans every phase |
| internal/apps/ layout | [Section 4.4.4, G.1.1, G.1.2](../../docs/ENG-HANDBOOK.md#444-internal-apps) | Target structure being enforced |
