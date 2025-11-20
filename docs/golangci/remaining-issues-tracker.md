# golangci-lint v2 Migration - Remaining Issues Tracker

**Date Created**: November 19, 2025
**Version**: golangci-lint v2.6.2
**Status**: 🟡 **TRACKING REMAINING ISSUES**

---

## Purpose

This document tracks all remaining unresolved issues from the golangci-lint v2 migration, pre-commit/pre-push hook updates, workflow changes, copilot instructions updates, and documentation synchronization. This is a living document that should be updated as issues are resolved.

---

## Critical Outstanding Items

### 1. Pre-commit Hook Documentation Synchronization

**Status**: 🔴 **NEEDS REVIEW**

**Issue**: Pre-commit hook documentation in `docs/pre-commit-hooks.md` may not reflect all v2 changes

**Action Items**:
- [ ] Review `docs/pre-commit-hooks.md` for outdated v1 references
- [ ] Update tool ordering diagrams with current pipeline
- [ ] Verify all hook configuration examples match `.pre-commit-config.yaml`
- [ ] Update timing expectations based on v2 performance
- [ ] Add v2-specific troubleshooting guidance

**Files**: `docs/pre-commit-hooks.md`

**Priority**: HIGH - Documentation accuracy for team onboarding

---

### 2. Copilot Instructions Import Alias Validation

**Status**: 🟡 **NEEDS VALIDATION**

**Issue**: `.github/instructions/01-03.golang.instructions.md` importas section must stay synchronized with `.golangci.yml`

**Action Items**:
- [ ] Audit all import aliases in `.golangci.yml` importas section
- [ ] Verify matching aliases in copilot instructions file
- [ ] Check for missing identity-related package aliases
- [ ] Add validation script to detect instruction/config drift

**Files**:
- `.golangci.yml` (importas section)
- `.github/instructions/01-03.golang.instructions.md` (Import Alias Conventions)

**Priority**: MEDIUM - Prevents import naming inconsistencies

---

### 3. Workflow Integration Testing

**Status**: 🟡 **NEEDS VALIDATION**

**Issue**: CI/CD workflows (ci-quality.yml, etc.) may not fully leverage v2 capabilities

**Action Items**:
- [ ] Verify `ci-quality.yml` uses latest golangci-lint action version
- [ ] Check workflow timeout settings are appropriate for v2
- [ ] Validate artifact upload patterns for linter output
- [ ] Test workflow performance with v2 vs v1 baseline

**Files**: `.github/workflows/ci-quality.yml`, `.github/workflows/*.yml`

**Priority**: MEDIUM - CI/CD efficiency and reliability

---

### 4. Linting Instructions Completeness

**Status**: 🟢 **MOSTLY COMPLETE, NEEDS REVIEW**

**Issue**: `.github/instructions/01-06.linting.instructions.md` should cover all v2 migration patterns

**Action Items**:
- [ ] Verify all v2 removed settings are documented
- [ ] Add examples for new wsl_v5 configuration patterns
- [ ] Document depguard limitation workarounds (custom cicd check)
- [ ] Update linter enable/disable decision documentation

**Files**: `.github/instructions/01-06.linting.instructions.md`

**Priority**: MEDIUM - Developer guidance for linting patterns

---

### 5. Auto-Fix Integration Gaps

**Status**: 🟡 **PARTIAL COVERAGE**

**Issue**: Not all auto-fixable patterns are covered by `golangci-lint run --fix`

**Current Coverage**:
- ✅ gofumpt (formatting)
- ✅ goimports (imports)
- ✅ wsl (whitespace)
- ✅ godot (comment periods)
- ✅ staticcheck (error strings via cicd)
- ✅ copyloopvar (via cicd)
- ✅ thelper (via cicd)

**Missing Auto-Fixes** (from `docs/golangci/auto-fix-integration-plan.md`):
- ❌ errcheck defer closures (too context-dependent)
- ❌ wrapcheck file-level nolint (too project-specific)
- ❌ tparallel cleanup conversion (too context-dependent)
- ❌ mnd magic number extraction (requires semantic analysis)
- ❌ goconst constant extraction (requires semantic analysis)

**Action Items**:
- [ ] Document which patterns require manual intervention
- [ ] Create troubleshooting guide for common manual fix scenarios
- [ ] Consider AST-based auto-fix for high-value patterns (thelper already done)

**Files**: `docs/golangci/auto-fix-integration-plan.md`, cicd auto-fix commands

**Priority**: LOW - Current coverage is good, manual fixes are acceptable

---

### 6. VS Code Integration Settings

**Status**: 🟡 **NEEDS VALIDATION**

**Issue**: `.vscode/settings.json` should reflect v2 configuration and capabilities

**Action Items**:
- [ ] Verify `go.lintTool` setting uses golangci-lint
- [ ] Check `go.lintFlags` includes `--fix` for auto-fix on save
- [ ] Validate editor integration with v2 linters
- [ ] Test VSCode problem matcher patterns with v2 output format

**Files**: `.vscode/settings.json`

**Priority**: MEDIUM - Developer experience and IDE integration

---

### 7. Testing Instructions Updates

**Status**: 🟡 **NEEDS REVIEW**

**Issue**: `.github/instructions/01-02.testing.instructions.md` should cover v2 linting integration

**Action Items**:
- [ ] Document test file linting best practices with v2
- [ ] Update pre-commit test execution guidance
- [ ] Add troubleshooting for thelper, tparallel, testpackage linters
- [ ] Document test coverage + linting workflow integration

**Files**: `.github/instructions/01-02.testing.instructions.md`

**Priority**: LOW - Testing practices are stable, minimal v2 impact

---

### 8. Docker/Compose Dockerfile Linting

**Status**: 🟢 **COMPLETE**

**Issue**: hadolint integration for Dockerfile linting

**Current State**:
- ✅ hadolint configured in `.pre-commit-config.yaml`
- ✅ Running on Dockerfile and Dockerfile.* files

**Priority**: COMPLETE - No action needed

---

### 9. GitHub Actions Workflow Linting

**Status**: 🟢 **COMPLETE**

**Issue**: actionlint integration for workflow validation

**Current State**:
- ✅ actionlint configured in `.pre-commit-config.yaml`
- ✅ Custom cicd check (github-workflow-lint) for additional validations
- ✅ Running on `.github/workflows/*.yml` files

**Priority**: COMPLETE - No action needed

---

### 10. Archive Old Migration Docs

**Status**: 🟢 **COMPLETE**

**Issue**: Clean up temporary migration documentation

**Current State**:
- ✅ Migration artifacts archived in `docs/golangci/archive/`
- ✅ `MIGRATION-COMPLETE.md` documents final state
- ✅ `auto-fix-integration-plan.md` preserved for reference

**Priority**: COMPLETE - No action needed

---

## Documentation Synchronization Matrix

| Document | Last Updated | Sync Status | Priority |
|----------|--------------|-------------|----------|
| `.golangci.yml` | Nov 19, 2025 | ✅ Current | - |
| `.pre-commit-config.yaml` | Nov 19, 2025 | ✅ Current | - |
| `docs/pre-commit-hooks.md` | Oct 26, 2025 | 🟡 Needs Review | HIGH |
| `.github/instructions/01-06.linting.instructions.md` | Nov 19, 2025 | 🟢 Mostly Current | MEDIUM |
| `.github/instructions/01-03.golang.instructions.md` | Nov 19, 2025 | 🟡 Needs Validation | MEDIUM |
| `.github/instructions/01-02.testing.instructions.md` | Unknown | 🟡 Needs Review | LOW |
| `.vscode/settings.json` | Unknown | 🟡 Needs Validation | MEDIUM |
| `.github/workflows/ci-quality.yml` | Unknown | 🟡 Needs Validation | MEDIUM |

---

## Next Steps

### Immediate (This Week)
1. Review and update `docs/pre-commit-hooks.md` with v2 specifics
2. Validate import alias synchronization between `.golangci.yml` and copilot instructions
3. Test VS Code integration settings with v2

### Short-term (This Sprint)
1. Validate workflow integration and performance
2. Update linting instructions with v2 migration patterns
3. Review testing instructions for v2-specific guidance

### Long-term (Future Sprints)
1. Consider AST-based auto-fix for additional patterns (low priority)
2. Monitor for new golangci-lint v2 linters and formatters
3. Evaluate v2 performance optimizations

---

## Success Criteria

Migration is considered fully complete when:
- [ ] All documentation accurately reflects v2 configuration
- [ ] Import alias validation script is in place
- [ ] CI/CD workflows validated with v2
- [ ] VS Code integration tested and documented
- [ ] All team members onboarded to v2 workflows
- [ ] No outstanding synchronization gaps between config and docs

---

## Notes

- This tracker should be reviewed and updated weekly during active migration cleanup
- Mark items as complete (✅) when validated, not just when "done"
- Add new items as discovered during code reviews and development
- Archive this document to `docs/golangci/archive/` once all items complete
