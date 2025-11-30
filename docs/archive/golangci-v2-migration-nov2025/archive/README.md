# golangci-lint v2 Migration Archive

## Overview

This directory contains historical documentation from the golangci-lint v1 → v2 migration completed in November 2025.

## Migration Summary

**Date**: November 19, 2025
**Version**: golangci-lint v2.6.2
**Status**: ✅ Complete and stable

## Archived Files

### Migration Documentation

- **migrate-v2-todos.md**: Post-migration task tracking (all 10 tasks completed)
- **migrate-v2-summary.md**: Comprehensive migration summary and lessons learned
- **migrate-v2-problems.md**: Identified problems and v1/v2 differences
- **migrate-v2-performance.md**: Performance analysis and benchmarks
- **migrate-v2-functionality.md**: Detailed functionality changes and workarounds
- **migrate-v2-completed.md**: Completion checklist and validation results
- **session-summary-nov19-2025.md**: Session-by-session implementation log

### Original v1 Configuration

**File**: `.golangci.yml.backup` (deleted after archiving)
**Commit**: Prior to migration (hash: see git history)
**Final v2 config commit**: 9f90313f (November 19, 2025)

To view original v1 configuration:

```bash
git log --all --oneline -- .golangci.yml.backup
git show <commit-hash>:.golangci.yml.backup
```

## Key Migration Changes

### Removed Settings (v2 API)

- `wsl.force-err-cuddling` → Use `wsl.error-variable-names` instead
- `misspell.ignore-words` → No longer needed (v2 handles crypto terms correctly)
- `wrapcheck.ignoreSigs` → Use file-level `//nolint:wrapcheck` for HTTP handlers
- `depguard` file-scoped rules → Use custom cicd checks

### Renamed Linters

- `wsl` → `wsl_v5` (configuration key, linter name unchanged)

### Built-in Formatters

- **gofumpt** and **goimports** are built into golangci-lint v2
- No separate installation required
- Configuration files (`.gofumpt.toml`) still respected
- Both run automatically with `--fix` flag

### Custom Solutions

**Domain Isolation Enforcement**: v2 removed file-scoped depguard rules

**Solution**: Custom cicd check (`go-check-identity-imports`)

- File: `internal/cmd/cicd/cicd_check_identity_imports.go`
- Integration: Pre-commit hook (cicd-checks-internal)
- Purpose: Blocks identity module from importing KMS infrastructure packages

**Blocked Packages** (9 total):

1. `cryptoutil/internal/server` - KMS server domain
2. `cryptoutil/internal/client` - KMS client
3. `cryptoutil/api` - OpenAPI generated code
4. `cryptoutil/cmd/cryptoutil` - CLI command
5. `cryptoutil/internal/common/crypto` - Use stdlib instead
6. `cryptoutil/internal/common/pool` - KMS infrastructure
7. `cryptoutil/internal/common/container` - KMS infrastructure
8. `cryptoutil/internal/common/telemetry` - KMS infrastructure
9. `cryptoutil/internal/common/util` - KMS infrastructure

## Validation Results

### Pre-commit Hooks

- **Command**: `pre-commit run --all-files`
- **Result**: ✅ All hooks passed (~1.0s total execution)
- **Commit**: `5f665028`

### Pre-push Hooks

- **Command**: `pre-commit run --hook-stage pre-push --all-files`
- **Result**: ✅ All hooks passed (~1.0s total execution)
- **Commit**: `68c3cd60`

### Full Linter Run

- **Command**: `golangci-lint run --timeout=10m`
- **Result**: 242 lines output (75 errcheck, 20 noctx, 6 goconst, 3 trivial)
- **Status**: All warnings acceptable (test code, tools, cosmetic issues)

## Migration Completion

**All 10 post-migration tasks completed**:

1. ✅ Monitor Misspell False Positives
2. ✅ Monitor Wrapcheck Noise
3. ✅ Restore Domain Isolation Enforcement (custom cicd check)
4. ✅ Line Length Enforcement Decision (documented)
5. ✅ Inline Comments Decision (documented)
6. ✅ Formatter Documentation (updated)
7. ✅ Update Instruction Files (linting.instructions.md, pre-commit-hooks.md)
8. ✅ Test CI/CD Pipeline (pre-commit, pre-push validated)
9. ✅ Monitor Linter Behavior Changes (full run analysis)
10. ✅ Cleanup Migration Artifacts (this archive)

## Current Configuration

**Active file**: `.golangci.yml` (v2 configuration)
**Enabled linters**: 22 linters
**Performance**: <2s incremental, acceptable for full runs
**Stability**: No rollback needed since migration

## References

- **v2 Documentation**: <https://golangci-lint.run/>
- **Migration Guide**: See archived files in this directory
- **Instruction Files**: `.github/instructions/01-06.linting.instructions.md`
- **Pre-commit Docs**: `docs/pre-commit-hooks.md`

## Archive Policy

These files are preserved for:

- Historical reference
- Lessons learned for future migrations
- Understanding v1 → v2 differences
- Troubleshooting migration-related issues

**Do not delete** - Minimal storage impact, high reference value.
