# Cryptoutil Code Quality & Linting TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 26, 2025
**Status**: Active code quality enhancements in progress - File extension pattern review completed

---

## 🟡 MEDIUM - Code Quality & Linting Enhancements

### Task CQ1: Address TODO Comments in Codebase
- **Description**: Multiple TODO/FIXME comments found throughout codebase requiring attention
- **Current TODO Inventory (Excluding Identity Subsystem)**:
  - `internal/common/pool/pool.go:40` - "TODO change generateCounter and getCounter from uint64 to telemetryService.MetricsProvider.Counter()"
  - `internal/common/crypto/jose/jws_message_util.go:170` - "TODO support multiple signatures"
  - `internal/server/application/application_listener.go:630` - "TODO: Add actual dependency health checks here"
  - `internal/server/application/application_listener.go:710` - "TODO Add more readiness checks as needed"
  - `internal/server/repository/sqlrepository/gormdb.go:62` - "TODO : Enable gorm debug mode if needed"
  - `internal/server/repository/sqlrepository/sql_schema_util.go` - Multiple context.TODO() usages (lines 28, 62, 100, 132) for database queries
- **Identity Subsystem TODOs** (40+ items - tracked separately as they represent incomplete features):
  - User authentication flows (passkey, TOTP, OTP)
  - Session management (cleanup, validation)
  - Token operations (introspection, revocation)
  - Authorization flows (consent, logout, userinfo)
  - Repository integrations
- **Action Items**:
  - Review non-identity TODOs for relevance and priority
  - Implement high-priority TODOs or convert to proper issues
  - Document context.TODO() usage patterns for database operations
  - Track identity subsystem TODOs separately as feature work
- **Files**: Multiple files across codebase
- **Expected Outcome**: Clean codebase with actionable TODOs only; identity subsystem TODOs tracked as feature work
- **Priority**: LOW - Code maintainability improvement
- **Note**: godox linter disabled in favor of manual tracking in this file

### Task CQ3: Enable Additional Quality Linters

- **Description**: Add more golangci-lint linters for enhanced code quality
- **Current State**: Additional linters enabled incrementally
- **Action Items**:
  - Evaluate and enable:
    - gocognit (cyclomatic complexity analysis)
      - Analysis: 10 functions exceed complexity threshold of 30
      - Files affected: config.go (1), config_test_util.go (1), certificates_server_test_util.go (1), jwe_jwk_util.go (1), jwk_util_test.go (1), jws_jwk_util.go (1), telemetry_service.go (1), application_listener.go (1), sql_provider.go (1), cicd.go (2)
      - Highest complexity: 157 (jwk_util_test.go)
  - Configure appropriate settings for each linter
  - Test CI performance impact
- **Files**: `.golangci.yml`
- **Expected Outcome**: Enhanced code quality and consistency checks
- **Priority**: Medium - Code quality improvement

### Task CQ4: Investigate linters for EOL/maintenance mode dependencies

- **Description**: Research and evaluate tools that can detect dependencies in end-of-life or maintenance mode
- **Current State**: No automated detection of deprecated/unmaintained dependencies
- **Potential Tools to Investigate**:
  - `go-mod-outdated`: Shows outdated dependency versions (<https://github.com/psampaz/go-mod-outdated>)
  - `govulncheck`: Official Go vulnerability scanner (already in use)
  - Custom scripts to check GitHub repository status/README for maintenance warnings
  - Integration with dependency health services or APIs
- **Action Items**:
  - Research available Go tools for dependency lifecycle detection
  - Evaluate feasibility of integrating EOL detection into CI/CD pipeline
  - Consider custom linter or cicd command for maintenance mode checking
  - Document findings and recommend implementation approach
- **Files**: `.golangci.yml`, `internal/cmd/cicd/cicd.go` (potential new command)
- **Expected Outcome**: Automated detection of unmaintained dependencies to prevent security/technical debt
- **Priority**: LOW - Proactive maintenance improvement
- **Timeline**: Q2 2026
