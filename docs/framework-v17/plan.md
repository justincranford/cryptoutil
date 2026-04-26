# Implementation Plan - Framework V17: internal/apps/ Structure Fitness Linters

**Status**: Planning
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

#### PS-IDs — Detailed Gap Matrix

| Invariant | sm-kms | sm-im | jose-ja | pki-ca | id-authz | id-idp | id-rs | id-rp | id-spa | skel-tmpl |
|-----------|--------|-------|---------|--------|----------|--------|-------|-------|--------|-----------|
| `{SERVICE}.go` | OK | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| `{SERVICE}_usage.go` | OK | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| `swagger.go` | OK | OK | OK | OK | OK | OK | OK | **MISS** | **MISS** | OK |
| `testmain_test.go` | **MISS** | OK | OK | OK | **MISS** | **MISS** | **MISS** | **MISS** | **MISS** | OK |
| `*_lifecycle_test.go` | OK | OK | OK | OK | OK | OK | **MISS** | **MISS** | **MISS** | OK |
| `*_port_conflict_test.go` | OK | OK | OK | OK | **MISS** | **MISS** | **MISS** | **MISS** | **MISS** | OK |
| `server/` dir | OK | OK | OK | OK | OK | OK | OK | OK | OK | OK |
| `e2e/` dir | OK | OK | OK | **MISS** | OK | **MISS** | **MISS** | **MISS** | **MISS** | OK |
| `*_contract_test.go` | **MISS** | **MISS** | **MISS** | **MISS** | OK | OK | OK | **MISS** | **MISS** | **MISS** |

**Notes**:
- `testmain_test.go`: Required per testing instructions (TestMain for heavyweight resources). The
  5 identity services without it use inline setup — a pattern drift that should be enforced.
- `swagger.go`: Present in 8 of 10. `identity-rp` and `identity-spa` missing — likely incomplete
  migration to OpenAPI-based serving.
- `e2e/`: Present in 5 of 10. ENG-HANDBOOK marks it optional but the 5 missing represent work not
  yet done (identity-rp, identity-spa, identity-idp, identity-rs, pki-ca have no E2E tests).
  Plan: enforce as REQUIRED for all PS-IDs, consistent with the framework migration target.
- `*_contract_test.go`: Only 3 of 10 have it (identity-authz, identity-idp, identity-rs). The
  missing 7 represent framework migration work not yet complete.
- `*_lifecycle_test.go` and `*_port_conflict_test.go`: Absent from identity-rs, identity-rp,
  identity-spa. These are framework-standard tests for all services.

#### Severity Classification

| Category | Invariant | Severity | Rationale |
|----------|-----------|----------|-----------|
| REQUIRED | `{SERVICE}.go` | ERROR | Primary entry point; already partially enforced |
| REQUIRED | `{SERVICE}_usage.go` | ERROR | CLI usage string; already partially enforced |
| REQUIRED | `server/server.go` | ERROR | Core service implementation |
| REQUIRED | `swagger.go` | ERROR | OpenAPI serving; 8/10 have it; gap is drift |
| REQUIRED | `testmain_test.go` | ERROR | TestMain for heavyweight resources; 5/10 missing = technical debt |
| REQUIRED | `*_lifecycle_test.go` | ERROR | Server lifecycle tests; 3/10 missing = drift |
| REQUIRED | `*_port_conflict_test.go` | ERROR | Port conflict tests; 5/10 missing = drift |
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
- Checks every PS-ID has `swagger.go` and `swagger_test.go` at the package root
- Current state: `identity-rp` and `identity-spa` are missing these files
- Tests: pass case, fail case (missing swagger.go), fail case (missing swagger_test.go)

**Linter 4: `apps-ps-id-test-patterns`** (`apps_ps_id_test_patterns/`)
- Checks every PS-ID has `testmain_test.go` at the package root
- Checks every PS-ID has at least one `*_lifecycle_test.go` file at package root
- Checks every PS-ID has at least one `*_port_conflict_test.go` file at package root
- Current state: 5 PS-IDs missing `testmain_test.go`; 3 missing lifecycle; 5 missing port conflict
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

### Phase 3: Registration, Integration & Knowledge Propagation (3h) [Status: ☐ TODO]

**Objective**: Register all 6 new linters, update the YAML manifest, extend the existing
`service-structure` linter, validate end-to-end, and propagate lessons.

- Register all 6 linters in `lint_fitness.go` `registeredLinters` slice
- Add 6 entries to `lint-fitness-registry.yaml` manifest
- Extend `service_structure.go` to cover all 10 PS-IDs (remove "legacy" exclusion, use registry)
  OR deprecate it in favor of the new `apps-ps-id-required-files` linter
- Run `go run ./cmd/cicd-lint lint-fitness` — must pass with 0 errors
- Run `go run ./cmd/cicd-lint lint-fitness -q` — must show all linters PASS
- Update ENG-HANDBOOK.md §9.11.1 fitness linter catalog with 6 new entries
- Update `docs/target-structure.md` §G.1.1 to reference the new enforcement
- Update `fitness_registry_completeness` test expectations (count goes from 68 to 74 linters)
- Review `lessons.md` and propagate insights to ENG-HANDBOOK.md, agents, instructions

**Success**: `go run ./cmd/cicd-lint lint-fitness` passes; all 6 linters in YAML manifest; ENG-HANDBOOK
updated; coverage and mutation targets met.

**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

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

### Overall Project Quality Gates

- ✅ `go run ./cmd/cicd-lint lint-fitness` passes with all 74 linters (was 68 + 6 new)
- ✅ `go run ./cmd/cicd-lint lint-docs` passes (no propagation drift)
- ✅ Coverage ≥98% for all new packages
- ✅ Mutation testing ≥98% for all new packages (run via `gremlins unleash`)
- ✅ Race detector clean: `go test -race -count=2 ./internal/apps-tools/cicd_lint/lint_fitness/...`

---

## Success Criteria

- [ ] 6 new fitness linter packages created with ≥98% coverage and ≥98% mutation
- [ ] All 6 linters registered in `lint_fitness.go` and `lint-fitness-registry.yaml`
- [ ] `service_structure.go` refactored to use `AllProductServices()` (registry-driven)
- [ ] `go run ./cmd/cicd-lint lint-fitness` passes with zero errors
- [ ] ENG-HANDBOOK.md §9.11.1 updated with 6 new linter entries
- [ ] `docs/target-structure.md` §G.1.1 references enforcement mechanism
- [ ] `lessons.md` populated with post-mortem from all 3 phases

---

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
