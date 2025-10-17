# Cryptoutil Code Quality & Linting TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 16, 2025
**Status**: Active code quality enhancements in progress - Pre-commit hook automation analysis added

---

## 🟡 MEDIUM - Code Quality & Linting Enhancements

### Task CQ1: Address TODO Comments in Codebase
- **Description**: Multiple TODO/FIXME comments found throughout codebase requiring attention
- **godox Issues**:
  - `internal/client/client_oam_mapper.go:89` - "TODO nil allowed if import not nil"
  - `internal/common/crypto/jose/jwkgen_service.go:46` - "TODO read from settings"
  - `internal/common/crypto/jose/jws_message_util.go:148` - "TODO support multiple signatures"
  - `internal/pool/pool.go:43` - "TODO change generateCounter and getCounter from uint64 to telemetryService.MetricsProvider.Counter()"
  - `internal/server/application/application_listener.go` - Multiple TODOs:
    - Line 54: "TODO Add separate timeouts for different shutdown phases (drain, force close, etc.)"
    - Line 93: "TODO Only use InsecureSkipVerify for DevMode"
    - Line 195: "TODO Replace this with improved otelFiberTelemetryMiddleware; unstructured logs and no OpenTelemetry are undesirable"
    - Line 200: "TODO Limit this to Swagger GET APIs, not Swagger UI static content"
    - Line 239: "TODO Disable Swagger UI in production environments (check settings.DevMode or add settings.Environment)"
    - Line 240: "TODO Add authentication middleware for Swagger UI access"
- **Action Items**:
  - Review each TODO comment for relevance and priority
  - Implement high-priority TODOs or convert to proper issues
  - Remove obsolete TODOs
  - Add proper documentation for complex TODOs
- **Files**: Multiple files across codebase
- **Expected Outcome**: Clean codebase with actionable TODOs only
- **Priority**: LOW - Code maintainability improvement
- **Note**: godox linter disabled in favor of manual tracking in this file

### Task CQ3: Enable Additional Quality Linters
- **Description**: Add more golangci-lint linters for enhanced code quality
- **Current State**: Additional linters enabled incrementally
- **Action Items**:
  - Evaluate and enable:
    - gocognit (cyclomatic complexity analysis)
      - Analysis: 10 functions exceed complexity threshold of 30
      - Files affected: config.go (1), config_test_util.go (1), certificates_server_test_util.go (1), jwe_jwk_util.go (1), jwk_util_test.go (1), jws_jwk_util.go (1), telemetry_service.go (1), application_listener.go (1), sql_provider.go (1), cicd_utils.go (2)
      - Highest complexity: 157 (jwk_util_test.go)
  - Configure appropriate settings for each linter
  - Test CI performance impact
- **Files**: `.golangci.yml`
- **Expected Outcome**: Enhanced code quality and consistency checks
- **Priority**: Medium - Code quality improvement

### Task CQ5: Analyze Instruction Files for Pre-commit Hook Automation ✅ COMPLETED
- **Description**: Evaluate all instruction files to identify rules that could be replaced by fast, reliable pre-commit hooks instead of AI-dependent validation
- **Implementation Results**:
  - ✅ **P0 - CRITICAL**: Enhanced formatting.instructions.md automation - Added UTF-8 BOM fixing, comprehensive file encoding checks
  - ✅ **P1 - HIGH**: Enhanced imports.instructions.md - Added cryptoutil-specific import alias rules to golangci-lint importas configuration
  - ✅ **P2 - HIGH**: Enhanced testing.instructions.md - Created custom test pattern enforcement script for UUIDv7 and testify usage
  - ✅ **Infrastructure**: Created .gofumpt.toml configuration, enhanced pre-commit hooks, made Docker linting optional
- **Files Created/Modified**: `.gofumpt.toml`, `.golangci.yml`, `.pre-commit-config.yaml`, `scripts/enforce_test_patterns.py`
- **Automation Achieved**: UTF-8 encoding, import aliases, test patterns, enhanced Go formatting
- **Expected Outcome**: Reduced reliance on AI for routine validation, faster feedback, more reliable enforcement ✅ ACHIEVED
- **Completion Date**: October 16, 2025

### Task CQ6: Add Recommended Pre-commit Hooks
- **Description**: Add additional pre-commit hooks from pre-commit-hooks repository for enhanced code quality and security
- **Recommended Hooks** (ordered by priority):
  1. **detect-private-key** - Check for existence of private keys (critical for crypto project)
  2. **detect-aws-credentials** - Check for AWS secrets with --allow-missing-credentials flag
  3. **check-case-conflict** - Check for files with names that would conflict on case-insensitive filesystems
  4. **check-illegal-windows-names** - Check for files that cannot be created on Windows
  5. **check-toml** - Validate TOML file syntax
  6. **check-symlinks** - Check for symlinks which do not point to anything
  7. **check-executables-have-shebangs** - Check that non-binary executables have proper shebang
  8. **check-shebang-scripts-are-executable** - Check that scripts with shebangs are executable
  9. **mixed-line-ending** - Fix mixed line endings with --fix=auto flag
  10. **pretty-format-json** - Pretty format JSON files with --autofix and --indent=2 flags
  11. **check-vcs-permalinks** - Ensure that links to VCS websites are permalinks
  12. **no-commit-to-branch** - Prevent commits to protected branches with --branch main flag
  13. **forbid-new-submodules** - Prevent addition of new git submodules
- **Implementation Plan**: Add hooks incrementally based on priority and testing
- **Files**: `.pre-commit-config.yaml`
- **Expected Outcome**: Enhanced security, cross-platform compatibility, and code quality checks
- **Priority**: MEDIUM - Code quality and security enhancement

---

## Common Magic Values to Watch For

- HTTP status codes: `200`, `404`, `500`
- Timeouts: `30`, `60`, `300` (seconds)
- Buffer sizes: `1024`, `4096`
- Retry counts: `3`, `5`, `10`
- Port numbers: `8080`, `5432`
- String literals: `"localhost"`, `"admin"`, `"default"`
