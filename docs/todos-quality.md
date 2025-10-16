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

### Task CQ5: Analyze Instruction Files for Pre-commit Hook Automation
- **Description**: Evaluate all instruction files to identify rules that could be replaced by fast, reliable pre-commit hooks instead of AI-dependent validation
- **Analysis Results**:
  - **HIGH PRIORITY - Fast & Reliable Automation Candidates**:
    1. **formatting.instructions.md**: File encoding, line endings, trailing whitespace - ALREADY partially automated, expand coverage
    2. **code-quality.instructions.md**: Linting rules (errcheck, wrapcheck, noctx, etc.) - MOST of these are already automated via golangci-lint
    3. **imports.instructions.md**: Import alias naming conventions - COULD be automated with custom goimports rules or custom linter
    4. **commits.instructions.md**: Conventional commit message validation - CAN be automated with commit-msg hooks
  - **MEDIUM PRIORITY - Partial Automation Candidates**:
    5. **security.instructions.md**: Security scanning (run scripts/security-scan.{ps1,sh}) - CAN be automated pre-commit for high-risk changes
    6. **testing.instructions.md**: Test patterns (UUIDv7 usage, testify assertions) - COULD be partially automated with custom linters
  - **LOW PRIORITY - Manual Process Candidates**:
    7. **crypto.instructions.md**: Algorithm compliance - Manual review required, not automatable
    8. **documentation.instructions.md**: Documentation organization - Manual process, not automatable
    9. **architecture.instructions.md**: Application architecture - Manual design decisions
    10. **database.instructions.md**: ORM patterns - Manual implementation choices
    11. **observability.instructions.md**: Telemetry configuration - Manual setup required
    12. **openapi.instructions.md**: API specification patterns - Manual design process
    13. **powershell.instructions.md**: PowerShell usage guidelines - Manual coding standards
    14. **project-layout.instructions.md**: Go project structure - Manual organization
    15. **scripts.instructions.md**: Script development patterns - Manual implementation
    16. **docker.instructions.md**: Docker configuration - Manual setup
    17. **cicd.instructions.md**: CI/CD workflow configuration - Manual pipeline design
    18. **act-testing.instructions.md**: Local workflow testing - Manual testing process
    19. **copilot-customization.instructions.md**: VS Code Copilot setup - Manual configuration
    20. **errors.instructions.md**: Error handling patterns - Manual implementation
    21. **git.instructions.md**: Git workflow patterns - Manual processes
    22. **go-dependencies.instructions.md**: Go dependency management - Manual decisions
    23. **linting-exclusions.instructions.md**: Linting exclusion patterns - ALREADY automated
    24. **todo-maintenance.instructions.md**: TODO list maintenance - Manual process
- **Priority Order for Implementation**:
  1. **P0 - CRITICAL**: Expand formatting.instructions.md automation (file encoding, line endings, whitespace)
  2. **P1 - HIGH**: Implement commit-msg hook for commits.instructions.md (conventional commits)
  3. **P2 - HIGH**: Create custom linter for imports.instructions.md (import alias naming)
  4. **P3 - MEDIUM**: Add security scanning hook for security.instructions.md (high-risk file changes)
  5. **P4 - MEDIUM**: Custom linter for testing.instructions.md patterns (UUIDv7, testify usage)
  6. **P5 - LOW**: Audit remaining instruction files for any missed automation opportunities
- **Action Items**:
  - Review each instruction file for automatable rules vs manual processes
  - Implement pre-commit hooks for high-priority automation candidates
  - Update instruction files to reference automated validation where applicable
  - Remove or update instructions that become fully automated
  - Test automation reliability vs AI-dependent validation
- **Files**: `.pre-commit-config.yaml`, `.golangci.yml`, custom linter scripts, all `.github/instructions/*.md` files
- **Expected Outcome**: Reduced reliance on AI for routine validation, faster feedback, more reliable enforcement
- **Priority**: HIGH - Development workflow efficiency and reliability improvement
- **Timeline**: Q4 2025 - Implement P0-P2, evaluate P3-P4
- **Success Criteria**: 80%+ of routine code quality checks automated and reliable

---

## Common Magic Values to Watch For

- HTTP status codes: `200`, `404`, `500`
- Timeouts: `30`, `60`, `300` (seconds)
- Buffer sizes: `1024`, `4096`
- Retry counts: `3`, `5`, `10`
- Port numbers: `8080`, `5432`
- String literals: `"localhost"`, `"admin"`, `"default"`
