# golangci-lint v2 Migration - Post-Migration Tasks

## Overview

Tasks to address functionality lost or degraded in v2 migration.

**Status**: 📋 Active tracking
**Created**: November 19, 2025
**Context**: Post-v2 migration cleanup and enhancement

---

## High Priority Tasks

### 1. Monitor Misspell False Positives ✅ COMPLETE

**Problem**: v2 removed `misspell.ignore-words` setting

**Lost Words** (crypto/technical terms):

- cryptoutil, keygen, jwa, jwk, jwe, jws
- ecdsa, ecdh, rsa, hmac, aes
- pkcs, pkix, x509, pem, der, ikm

**Impact**: NONE - No false positives found

**Validation Results** (November 19, 2025):

```bash
golangci-lint run --enable-only=misspell
```

**Findings**: 8 legitimate spelling issues, ZERO crypto term false positives

**Issues Found** (all legitimate American English corrections):

- `cancelled` → `canceled` (8 occurrences)
  - internal/common/telemetry/telemetry_service.go
  - internal/identity/jobs/cleanup_test.go
  - internal/identity/server/server_manager.go
  - internal/identity/test/e2e/mock_services.go
  - internal/server/repository/sqlrepository/* (4 files)

**Conclusion**: ✅ **NO ACTION NEEDED** - v2's misspell linter correctly handles crypto terminology without ignore-words setting

**Acceptance Criteria**: ✅ Zero false positives for legitimate crypto terminology

---

### 2. Monitor Wrapcheck Noise ✅ COMPLETE

**Problem**: v2 removed `wrapcheck.ignoreSigs` setting

**Lost Exemptions**:

- `.Errorf(` - fmt.Errorf and similar
- `errors.New(` - stdlib error creation
- `errors.Unwrap(` - error unwrapping
- `.Wrap(`, `.Wrapf(` - third-party error wrapping
- `(*github.com/gofiber/fiber/v2.Ctx).JSON(` - Fiber HTTP context methods
- `(*github.com/gofiber/fiber/v2.Ctx).SendStatus(` - Fiber HTTP responses

**Impact**: ALL 22 warnings are Fiber HTTP handlers (100% false positive rate for error wrapping pattern)

**Validation Results** (November 19, 2025):

```bash
golangci-lint run --enable-only=wrapcheck
```

**Findings**: 22 wrapcheck warnings, ALL are Fiber HTTP response methods

**Issues Breakdown**:

- **Fiber ctx.JSON()**: 20 warnings
  - internal/identity/authz/handlers_authorize.go (15 warnings)
  - internal/identity/authz/handlers_introspect_revoke.go (0 warnings for JSON)
  - internal/identity/idp/handlers_*.go (5 warnings across multiple handlers)
  - internal/identity/rs/service.go (5 warnings)

- **Fiber ctx.SendStatus()**: 2 warnings
  - internal/identity/authz/handlers_introspect_revoke.go (2 warnings)

**Analysis**:

1. **100% False Positive Rate**: ALL warnings are for HTTP response methods (ctx.JSON, ctx.SendStatus)
2. **Pattern**: HTTP handlers returning Fiber framework errors don't need wrapping (framework handles error responses)
3. **Standard Practice**: Fiber handlers return raw ctx.JSON/ctx.SendStatus errors - wrapping adds no value
4. **Lost ignoreSigs**: v1 explicitly exempted these exact signatures

**Decision**: ✅ **SUPPRESS WITH PATTERN-BASED NOLINT** - Use file-level or package-level wrapcheck suppression

**Solution Options**:

1. **Recommended**: Add file-level `//nolint:wrapcheck` to all HTTP handler files with justification
2. **Alternative**: Disable wrapcheck for `internal/identity/**/*handlers*.go` files in .golangci.yml
3. **Not Recommended**: Wrap all Fiber errors (adds no value, violates framework patterns)

**Acceptance Criteria**: ✅ 100% false positive rate justifies systematic suppression

---

### 3. Restore Domain Isolation Enforcement 🔧 IN PROGRESS

**Problem**: v2 removed complex depguard file-scoped rules

**Lost Capability**: Identity module domain isolation (10+ blocked imports)

**v1 Behavior**:

```yaml
depguard:
  rules:
    identity-domain-isolation:
      files: ["internal/identity/**/*.go"]
      deny:
        - cryptoutil/internal/server (KMS server)
        - cryptoutil/internal/client (KMS client)
        - cryptoutil/api (OpenAPI generated)
        - cryptoutil/cmd/cryptoutil (CLI)
        - cryptoutil/internal/common/crypto (use stdlib)
        - cryptoutil/internal/common/pool
        - cryptoutil/internal/common/container
        - cryptoutil/internal/common/telemetry
        - cryptoutil/internal/common/util
```

**Impact**: Identity module can now import from KMS domain (breaks architectural boundaries)

**Research Results** (November 19, 2025):

- **v2 depguard status**: File-scoped rules NOT SUPPORTED in v2.6.2
- **v1 config lost**: Complex identity-domain-isolation rule with file patterns
- **v2 limitation**: Only global deny rules supported (single main rule in .golangci.yml)
- **Reference**: `docs/golangci/migrate-v2-functionality.md` (Complex Depguard Rules section)

**Solution**: ✅ **CUSTOM CICD CHECK** - Follow existing pattern

**Implementation Plan**:

1. **Create check command**: `internal/cmd/cicd/cicd_check_identity_imports.go`
   - Pattern: Follow `cicd_check_circular_deps.go` structure (275 lines)
   - Input: Scan `internal/identity/**/*.go` files
   - Logic: Parse imports, check against blocked package list
   - Cache: Use cache file pattern (similar to circular deps check)
   - Output: Error if forbidden imports detected, success otherwise

2. **Add test coverage**: `internal/cmd/cicd/cicd_check_identity_imports_test.go`
   - Unit tests: Blocked import detection, cache validation
   - Integration tests: Real file scanning

3. **Integrate with pre-commit hook**: `.pre-commit-config.yaml`
   - Add hook: `cicd check-identity-imports` (similar to circular deps hook)
   - Stage: Run after go-any check, before golangci-lint

4. **Add to cicd.go command registry**:
   - Command: `cicd check-identity-imports`
   - Alias: `cicd all` should include this check

**Blocked Imports** (from v1 config):

```go
blockedPackages := []string{
    "cryptoutil/internal/server",      // KMS server domain
    "cryptoutil/internal/client",      // KMS client
    "cryptoutil/api",                  // OpenAPI generated code
    "cryptoutil/cmd/cryptoutil",       // CLI command
    "cryptoutil/internal/common/crypto", // Use stdlib instead
    "cryptoutil/internal/common/pool",
    "cryptoutil/internal/common/container",
    "cryptoutil/internal/common/telemetry",
    "cryptoutil/internal/common/util",
}
```

**Cache Strategy** (like circular deps):

- Cache file: `.cicd/identity-imports-cache.json`
- Invalidation: When `internal/identity/**/*.go` files change OR go.mod changes
- Validity: 5 minutes (same as circular deps check)

**Acceptance Criteria**:

- ✅ Command `cicd check-identity-imports` detects forbidden imports
- ✅ Pre-commit hook runs check automatically
- ✅ Cache prevents expensive scanning on every commit
- ✅ Clear error messages showing file + line + forbidden import

**Status**: 🔧 READY FOR IMPLEMENTATION (pattern identified, design complete)

---

## Medium Priority Tasks

### 4. Consider Line Length Enforcement 📏

**Problem**: v2 config doesn't enable `lll` linter (line length)

**v1 Behavior**: `lll.line-length: 190` (enforced 190 character maximum)

**Impact**: No automatic line length enforcement (relies on developer discipline)

**Action Items**:

- [ ] Survey codebase for long lines: `grep -r ".{191,}" --include="*.go" .`
- [ ] Decide if line length enforcement valuable:
  - Yes → Re-enable lll linter with 190 character limit
  - No → Document style guide in README (manual enforcement)
- [ ] If enabling, configure in .golangci.yml:

  ```yaml
  linters:
    enable:
      - lll
  settings:
    lll:
      line-length: 190
  ```

**Acceptance Criteria**: Documented decision (enable or manual) in this task

---

### 5. Restore Helpful Inline Comments ✅ COMPLETED

**Problem**: Migration removed verbose inline comments

**Impact**: LLM agents (Grok, Claude) may have difficulty understanding linter purposes

**Action Items**:

- [x] Review .golangci.yml.backup for valuable comments
- [x] Restore comments explaining:
  - Linter purposes (what each linter checks)
  - Setting rationales (why specific values chosen)
  - Exclusion reasons (why specific rules excluded)
  - Cross-references to instruction files
- [x] Focus on comments that aid LLM understanding (not just human developers)

**Completion**: Comments restored in commit 42c84697

---

### 6. Clarify Formatter Enforcement ✅ COMPLETED

**Problem**: gofumpt and goimports not explicitly configured in .golangci.yml

**Impact**: Unclear if formatters applied when using `--fix` flag

**Action Items**:

- [x] Add explanatory comments to .golangci.yml clarifying formatter behavior
- [x] Document that formatters are built-in to golangci-lint v2 (no separate config needed)
- [x] Update functionality doc to clarify formatters still enforced via --fix
- [x] Verify pre-commit hooks use --fix flag (already confirmed in .pre-commit-config.yaml)
- [x] Test: `golangci-lint run --fix` applies gofumpt and goimports automatically

**Completion**: Documentation updated in commit 42c84697

**Note**: golangci-lint v2 has gofumpt and goimports built-in (no configuration section needed)

---

## Low Priority Tasks

### 7. Update Instruction Files 📝

**Problem**: Instruction files reference v1 configuration

**Files to Update**:

- `.github/instructions/01-06.linting.instructions.md`
  - Document v2 API changes
  - Update removed properties list
  - Document wsl → wsl_v5 migration
  - Add depguard rules configuration
  - Clarify gofumpt/goimports are built-in formatters (no config needed)

- `docs/pre-commit-hooks.md`
  - Update golangci-lint configuration section
  - Document v2-specific settings
  - Clarify formatter vs linter distinction
  - Document that gofumpt/goimports are automatically applied with --fix

**Action Items**:

- [ ] Read migrate-v2-summary.md for v2 API changes
- [ ] Update linting.instructions.md with v2 specifics
- [ ] Update pre-commit-hooks.md with v2 changes
- [ ] Document built-in formatter behavior

**Acceptance Criteria**: Instruction files accurately reflect v2 configuration

---

### 8. Test CI/CD Pipeline 🧪

**Problem**: v2 config not validated in full CI/CD workflows

**Workflows to Test**:

- `ci-quality.yml` - Runs golangci-lint on every PR
- Pre-commit hooks - Runs golangci-lint locally
- Pre-push hooks - Runs full validation

**Action Items**:

- [ ] Trigger ci-quality workflow: Create test PR with intentional linting issues
- [ ] Verify pre-commit hook: Make local changes, attempt commit
- [ ] Verify pre-push hook: Push to feature branch
- [ ] Monitor for:
  - Unexpected linter failures
  - Missing issues (false negatives)
  - Excessive warnings (false positives)
  - Performance regressions (execution time >15 seconds)

**Acceptance Criteria**: All workflows execute without v2-related errors

---

### 9. Monitor Linter Behavior Changes 👀

**Problem**: v2 merged linters may have different behavior

**Merged Linters**:

- `staticcheck` now includes `gosimple` + `stylecheck`
- `wsl` replaced by `wsl_v5`

**Action Items**:

- [ ] Run lint on full codebase: `golangci-lint run --timeout=10m > lint-v2.txt`
- [ ] Compare with v1 baseline (if available)
- [ ] Document unexpected issues:
  - New warnings not present in v1
  - Missing warnings that were in v1
  - Different error messages
- [ ] For each unexpected issue:
  - Determine if legitimate code issue → fix code
  - Determine if linter configuration issue → adjust .golangci.yml
  - Determine if linter bug → report to golangci-lint project

**Acceptance Criteria**: All unexpected issues documented and resolved/accepted

---

### 10. Cleanup Migration Artifacts 🧹

**Problem**: Backup files and migration docs clutter repository

**Artifacts**:

- `.golangci.yml.backup` (489 lines, original v1 config)
- `docs/golangci/migrate-v2-*.md` (migration documentation)

**Action Items**:

- [ ] Validate v2 config stable for 30+ days (no rollback needed)
- [ ] Archive migration documentation:
  - Move `docs/golangci/migrate-v2-*.md` to `docs/golangci/archive/`
  - Update README to reference archive location
- [ ] Delete `.golangci.yml.backup` after archiving commit hash in migration docs
- [ ] Update .gitignore if needed

**Acceptance Criteria**: Repository cleaned, migration history preserved in git and docs/archive

---

## Completion Tracking

**Total Tasks**: 10
**Completed**: 2 ✅
**In Progress**: 0 🔄
**Blocked**: 0 🚫
**Not Started**: 8 📋

**Next Actions**:

1. Monitor misspell false positives (run lint, review warnings)
2. Monitor wrapcheck noise (run lint, categorize warnings)
3. Research v2 depguard file-scoped rules (restore domain isolation)

---

## References

- Migration summary: `docs/golangci/migrate-v2-summary.md`
- Functionality comparison: `docs/golangci/migrate-v2-functionality.md`
- Performance analysis: `docs/golangci/migrate-v2-performance.md`
- Remaining problems: `docs/golangci/migrate-v2-problems.md` (currently none)
- golangci-lint v2 docs: <https://golangci-lint.run/docs/configuration/file/>
- depguard v2 syntax: <https://golangci-lint.run/usage/linters/#depguard>
