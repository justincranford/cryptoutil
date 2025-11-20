# CICD Utility Refactoring Plan

**Date Created**: November 19, 2025
**Current Coverage**: 80.1%
**Target Coverage**: 95%+
**Status**: 🟡 **PLANNING PHASE**

---

## Executive Summary

This document outlines the comprehensive refactoring plan for the `internal/cmd/cicd` utility to improve maintainability, testability, and code organization. The refactoring will create a modular, well-tested command structure with clear separation of concerns.

---

## Current State Analysis

### File Structure (As of Nov 19, 2025)

```
internal/cmd/cicd/
├── .cicd/                                           # Cache directory
├── cicd.go                                          # Main command dispatcher (240 lines)
├── cicd_check_circular_deps.go                      # Circular dependency checker
├── cicd_check_circular_deps_test.go
├── cicd_check_identity_imports.go                   # Identity domain isolation checker
├── cicd_coverage_boost_test.go
├── cicd_edge_cases_test.go
├── cicd_enforce_any.go                              # Enforce Go 'any' vs 'interface{}'
├── cicd_enforce_any_test.go
├── cicd_enforce_test_patterns.go                    # Enforce UUIDv7, testify patterns
├── cicd_enforce_test_patterns_integration_test.go
├── cicd_enforce_test_patterns_test.go
├── cicd_enforce_utf8.go                             # UTF-8 encoding enforcement
├── cicd_enforce_utf8_test.go
├── cicd_final_coverage_test.go
├── cicd_github_api_cache.go                         # GitHub API caching utility
├── cicd_github_api_cache_test.go
├── cicd_go_fix_copyloopvar.go                       # Auto-fix loop variable capture
├── cicd_go_fix_copyloopvar_test.go
├── cicd_go_fix_staticcheck.go                       # Auto-fix staticcheck ST1005
├── cicd_go_fix_staticcheck_test.go
├── cicd_go_fix_thelper.go                           # Auto-fix missing t.Helper()
├── cicd_go_fix_thelper_test.go
├── cicd_integration_test.go
├── cicd_log.go                                      # Logging utility
├── cicd_run_integration_test.go
├── cicd_test.go
├── cicd_update_deps.go                              # Dependency update checker
├── cicd_update_deps_test.go
├── cicd_util_test.go
├── cicd_workflow_functions_test.go
├── cicd_workflow_lint.go                            # GitHub workflow linting
├── cicd_workflow_lint_integration_test.go
└── cicd_workflow_lint_test.go
```

### Command Categories

**Enforcement Commands** (Validation, fail on violation):
1. `all-enforce-utf8` - UTF-8 encoding enforcement
2. `go-enforce-any` - Enforce Go 'any' vs 'interface{}'
3. `go-enforce-test-patterns` - Enforce UUIDv7, testify patterns

**Checking Commands** (Analysis, informational):
1. `go-check-circular-package-dependencies` - Circular dependency detection
2. `go-check-identity-imports` - Identity domain isolation validation

**Auto-Fix Commands** (Code transformation):
1. `go-fix-staticcheck-error-strings` - Fix ST1005 violations
2. `go-fix-copyloopvar` - Fix loop variable capture
3. `go-fix-thelper` - Add missing t.Helper()
4. `go-fix-all` - Run all auto-fix commands

**Update Commands** (Dependency management):
1. `go-update-direct-dependencies` - Check direct dependency updates
2. `go-update-all-dependencies` - Check all dependency updates

**Linting Commands** (External file validation):
1. `github-workflow-lint` - GitHub workflow validation

### Code Metrics

| Category | Files | Test Files | Lines (approx) | Coverage |
|----------|-------|------------|----------------|----------|
| Main Dispatcher | 1 | 1 | 240 | High |
| Enforcement | 3 | 3 | 600 | Good |
| Checking | 2 | 1 | 400 | Medium |
| Auto-Fix | 3 | 3 | 900 | Good |
| Update | 1 | 1 | 300 | Good |
| Linting | 1 | 2 | 400 | Good |
| Utilities | 2 | 1 | 200 | High |
| **Total** | **13** | **12** | **~3040** | **80.1%** |

---

## golangci-lint v2 Overlap Analysis

### Task 4: Identify Redundant cicd Commands

**Analysis Needed**: Compare cicd command functionality with golangci-lint v2 capabilities

#### Commands with Potential Overlap

**1. `go-enforce-any` (interface{} → any)**
- **golangci-lint v2**: No built-in linter for this pattern
- **Decision**: **KEEP** - Project-specific requirement, not covered by v2
- **Justification**: Enforces Go 1.18+ type parameter syntax

**2. `go-fix-staticcheck-error-strings` (ST1005 auto-fix)**
- **golangci-lint v2**: staticcheck linter detects, but doesn't auto-fix ST1005
- **Decision**: **KEEP** - Provides auto-fix capability v2 lacks
- **Justification**: Saves manual editing, preserves acronyms

**3. `go-fix-copyloopvar` (loop variable capture)**
- **golangci-lint v2**: copyloopvar linter detects issues
- **Go 1.22+**: Automatic loop variable capture makes this obsolete
- **Decision**: **DEPRECATE for Go 1.25+** - No-op for current Go version
- **Migration Path**: Keep for backwards compatibility, mark as deprecated

**4. `go-fix-thelper` (missing t.Helper())**
- **golangci-lint v2**: thelper linter detects, doesn't auto-fix
- **Decision**: **KEEP** - Provides auto-fix capability v2 lacks
- **Justification**: Common pattern in test code, saves manual editing

**5. `go-enforce-test-patterns` (UUIDv7, testify)**
- **golangci-lint v2**: No built-in support for these patterns
- **Decision**: **KEEP** - Project-specific testing standards
- **Justification**: Enforces test quality and consistency

**6. `all-enforce-utf8` (UTF-8 encoding)**
- **golangci-lint v2**: No encoding enforcement
- **pre-commit hooks**: `fix-byte-order-marker` removes BOM, doesn't enforce UTF-8
- **Decision**: **KEEP** - Critical for cross-platform compatibility
- **Justification**: PowerShell UTF-16 LE breaks Docker secrets

**7. `github-workflow-lint` (workflow validation)**
- **golangci-lint v2**: Go-only, doesn't lint YAML
- **pre-commit**: actionlint covers basic workflow syntax
- **cicd check**: Adds version pinning, naming conventions
- **Decision**: **KEEP** - Complements actionlint with project standards

**8. `go-check-circular-package-dependencies`**
- **golangci-lint v2**: No circular dependency detection
- **Decision**: **KEEP** - Important architectural validation

**9. `go-check-identity-imports`**
- **golangci-lint v2**: depguard removed file-scoped rules in v2
- **Decision**: **KEEP** - Replaces v2 missing functionality
- **Justification**: Critical for domain isolation (identity vs KMS)

**10. `go-update-direct-dependencies` / `go-update-all-dependencies`**
- **golangci-lint v2**: No dependency update functionality
- **Decision**: **KEEP** - Unique capability, no overlap

#### Summary: Overlap Analysis Results

| Command | golangci-lint v2 Overlap | Decision | Reason |
|---------|--------------------------|----------|--------|
| go-enforce-any | None | **KEEP** | Project-specific requirement |
| go-fix-staticcheck | Detects, doesn't fix | **KEEP** | Auto-fix capability |
| go-fix-copyloopvar | Detects, auto in Go 1.22+ | **DEPRECATE** | Obsolete for Go 1.25+ |
| go-fix-thelper | Detects, doesn't fix | **KEEP** | Auto-fix capability |
| go-enforce-test-patterns | None | **KEEP** | Project-specific standards |
| all-enforce-utf8 | None | **KEEP** | Critical cross-platform |
| github-workflow-lint | None (not Go) | **KEEP** | Workflow-specific validation |
| go-check-circular-deps | None | **KEEP** | Architecture validation |
| go-check-identity-imports | v2 removed file-scoped | **KEEP** | Replaces v2 missing feature |
| go-update-*-dependencies | None | **KEEP** | Dependency management |

**Conclusion**: **NO REDUNDANT COMMANDS** - All cicd commands provide unique value not covered by golangci-lint v2.

---

## Proposed Refactoring Structure

### Target Directory Layout

```
internal/cmd/cicd/
├── cicd.go                          # Main command dispatcher (simplified)
├── cicd_test.go                     # Integration tests for dispatcher
│
├── common/                          # Shared utilities (extracted)
│   ├── logger.go                    # Logging utility (from cicd_log.go)
│   ├── logger_test.go
│   ├── files.go                     # File collection helpers
│   ├── files_test.go
│   ├── cache.go                     # Generic caching (from github_api_cache)
│   ├── cache_test.go
│   └── summary.go                   # Execution summary formatting
│
├── enforce/                         # Enforcement commands
│   ├── utf8/
│   │   ├── utf8.go                  # UTF-8 encoding enforcement
│   │   └── utf8_test.go
│   ├── any/
│   │   ├── any.go                   # Go 'any' enforcement
│   │   └── any_test.go
│   └── testpatterns/
│       ├── testpatterns.go          # UUIDv7, testify enforcement
│       ├── testpatterns_test.go
│       └── testpatterns_integration_test.go
│
├── check/                           # Analysis/checking commands
│   ├── circuitdeps/
│   │   ├── circuitdeps.go           # Circular dependency checker
│   │   └── circuitdeps_test.go
│   └── identityimports/
│       ├── identityimports.go       # Identity domain isolation
│       ├── identityimports_test.go
│       └── cache.go                 # Import check caching
│
├── fix/                             # Auto-fix commands
│   ├── staticcheck/
│   │   ├── staticcheck.go           # ST1005 auto-fix
│   │   └── staticcheck_test.go
│   ├── copyloopvar/
│   │   ├── copyloopvar.go           # Loop variable capture fix
│   │   └── copyloopvar_test.go
│   ├── thelper/
│   │   ├── thelper.go               # t.Helper() auto-fix
│   │   └── thelper_test.go
│   └── all/
│       ├── all.go                   # Run all auto-fix commands
│       └── all_test.go
│
├── update/                          # Dependency management
│   ├── deps/
│   │   ├── deps.go                  # Dependency update checker
│   │   ├── deps_test.go
│   │   └── github_cache.go          # GitHub API caching
│
└── lint/                            # External file linting
    └── workflow/
        ├── workflow.go              # GitHub workflow linting
        ├── workflow_test.go
        └── workflow_integration_test.go
```

### File Size Targets

**Large Files to Split** (>300 lines):
- `cicd_enforce_test_patterns.go` → Split into multiple focused files
- `cicd_go_fix_staticcheck.go` → Split validation, transformation, formatting
- `cicd_update_deps.go` → Split API calls, parsing, analysis

**Target**: All files <200 lines for easy comprehension and maintenance

---

## Refactoring Tasks Breakdown

### Phase 1: Common Code Extraction (Task 6)

**Goal**: Extract shared utilities into `common/` package

**Files to Create**:
1. `common/logger.go` - Extract from `cicd_log.go`
2. `common/files.go` - File collection and filtering utilities
3. `common/cache.go` - Generic caching abstraction
4. `common/summary.go` - Execution summary formatting

**Rationale**: Reduces duplication, improves testability

**Estimated Effort**: 3-4 hours

---

### Phase 2: Subdirectory Structure (Task 5)

**Goal**: Create command-specific subdirectories

**Commands to Migrate**:

**2.1 Enforcement Commands** (2 hours):
- `enforce/utf8/` ← `cicd_enforce_utf8.go`
- `enforce/any/` ← `cicd_enforce_any.go`
- `enforce/testpatterns/` ← `cicd_enforce_test_patterns.go`

**2.2 Checking Commands** (1.5 hours):
- `check/circuitdeps/` ← `cicd_check_circular_deps.go`
- `check/identityimports/` ← `cicd_check_identity_imports.go`

**2.3 Auto-Fix Commands** (2.5 hours):
- `fix/staticcheck/` ← `cicd_go_fix_staticcheck.go`
- `fix/copyloopvar/` ← `cicd_go_fix_copyloopvar.go`
- `fix/thelper/` ← `cicd_go_fix_thelper.go`
- `fix/all/` ← New file for `go-fix-all`

**2.4 Update Commands** (1.5 hours):
- `update/deps/` ← `cicd_update_deps.go` + `cicd_github_api_cache.go`

**2.5 Lint Commands** (1 hour):
- `lint/workflow/` ← `cicd_workflow_lint.go`

**Estimated Effort**: 8-9 hours

---

### Phase 3: File Splitting (Task 7)

**Goal**: Break large files into smaller, focused modules

**Priority Files**:

**3.1 `enforce/testpatterns/`** (Currently ~400 lines):
- `testpatterns.go` - Core validation logic
- `uuid_checker.go` - UUIDv7 validation
- `testify_checker.go` - Testify assertion validation
- `file_checker.go` - Test file organization

**3.2 `fix/staticcheck/`** (Currently ~350 lines):
- `staticcheck.go` - Main orchestration
- `validator.go` - Error string validation
- `transformer.go` - AST transformation
- `formatter.go` - Output formatting

**3.3 `update/deps/`** (Currently ~300 lines):
- `deps.go` - Main update logic
- `github_api.go` - GitHub API calls
- `parser.go` - go.mod parsing
- `analyzer.go` - Version comparison

**Estimated Effort**: 4-5 hours

---

### Phase 4: Test Coverage Enhancement (Task 8)

**Goal**: Increase coverage from 80.1% to 95%+

**Coverage Gaps by File** (from coverage report):

**High Priority** (Biggest impact):
1. `cicd_check_identity_imports.go` - Add edge case tests
2. `cicd_update_deps.go` - Mock GitHub API error scenarios
3. `cicd_workflow_lint.go` - Add workflow parsing edge cases

**Medium Priority**:
1. `cicd_enforce_test_patterns.go` - Add regex matching edge cases
2. `cicd_go_fix_staticcheck.go` - Add AST transformation edge cases
3. `cicd_enforce_utf8.go` - Add encoding detection edge cases

**Low Priority**:
1. Utility functions already well-tested
2. Integration tests provide good coverage

**Test Additions Needed**:
- Error path coverage: ~50 new test cases
- Edge case coverage: ~30 new test cases
- Integration tests: ~10 new test scenarios

**Estimated Effort**: 6-7 hours

---

### Phase 5: Documentation & Integration

**Goal**: Update documentation and ensure smooth integration

**5.1 Update Main Dispatcher** (1 hour):
- Simplify `cicd.go` to delegate to subdirectory commands
- Update command routing
- Preserve backwards compatibility

**5.2 Update Documentation** (1.5 hours):
- Update README with new structure
- Document each command's subdirectory
- Add architecture diagram

**5.3 Update Build & CI** (1 hour):
- Verify pre-commit hooks work with new structure
- Test CI workflows
- Update any import paths

**Estimated Effort**: 3.5 hours

---

## Implementation Plan

### Timeline & Sequencing

**Week 1**: Common Code Extraction (Phase 1)
- Days 1-2: Extract and test common utilities
- Day 3: Update all commands to use common code

**Week 2**: Subdirectory Structure (Phase 2)
- Days 1-2: Migrate enforcement and checking commands
- Days 3-4: Migrate auto-fix commands
- Day 5: Migrate update and lint commands

**Week 3**: File Splitting (Phase 3)
- Days 1-2: Split large enforcement files
- Days 3-4: Split large auto-fix files
- Day 5: Split update/deps files

**Week 4**: Test Coverage (Phase 4)
- Days 1-3: Add error path and edge case tests
- Days 4-5: Add integration tests and verify coverage

**Week 5**: Documentation & Integration (Phase 5)
- Days 1-2: Update documentation
- Days 3-4: Update build and CI
- Day 5: Final testing and validation

**Total Estimated Effort**: 25-30 hours over 5 weeks

---

## Success Criteria

**Refactoring Complete When**:
- [ ] All commands moved to subdirectories
- [ ] Common code extracted and shared
- [ ] All files <200 lines
- [ ] Test coverage ≥95%
- [ ] All pre-commit hooks passing
- [ ] All CI workflows passing
- [ ] Documentation updated
- [ ] Code review approved

---

## Risks & Mitigations

**Risk 1**: Breaking pre-commit hooks during refactoring
- **Mitigation**: Test hooks after each phase, maintain backwards compatibility

**Risk 2**: Import path changes breaking external code
- **Mitigation**: Keep public API stable, only refactor internal structure

**Risk 3**: Test coverage regression during refactoring
- **Mitigation**: Run coverage checks after each file migration

**Risk 4**: Large files split incorrectly, reducing cohesion
- **Mitigation**: Review file splits with team, ensure logical grouping

---

## Appendix A: Missed Changes Tracker

**Purpose**: Track items discovered during implementation that weren't in the original plan

**Format**: Add entries as discovered, review after each phase

### Phase 1 Discoveries
- [ ] TBD

### Phase 2 Discoveries
- [ ] TBD

### Phase 3 Discoveries
- [ ] TBD

### Phase 4 Discoveries
- [ ] TBD

### Phase 5 Discoveries
- [ ] TBD

---

## Appendix B: Command Reference

| Command | Current File | Target Location | Category |
|---------|-------------|-----------------|----------|
| all-enforce-utf8 | cicd_enforce_utf8.go | enforce/utf8/ | Enforcement |
| go-enforce-any | cicd_enforce_any.go | enforce/any/ | Enforcement |
| go-enforce-test-patterns | cicd_enforce_test_patterns.go | enforce/testpatterns/ | Enforcement |
| go-check-circular-package-dependencies | cicd_check_circular_deps.go | check/circuitdeps/ | Checking |
| go-check-identity-imports | cicd_check_identity_imports.go | check/identityimports/ | Checking |
| go-fix-staticcheck-error-strings | cicd_go_fix_staticcheck.go | fix/staticcheck/ | Auto-Fix |
| go-fix-copyloopvar | cicd_go_fix_copyloopvar.go | fix/copyloopvar/ | Auto-Fix |
| go-fix-thelper | cicd_go_fix_thelper.go | fix/thelper/ | Auto-Fix |
| go-fix-all | cicd.go (inline) | fix/all/ | Auto-Fix |
| go-update-direct-dependencies | cicd_update_deps.go | update/deps/ | Update |
| go-update-all-dependencies | cicd_update_deps.go | update/deps/ | Update |
| github-workflow-lint | cicd_workflow_lint.go | lint/workflow/ | Linting |
