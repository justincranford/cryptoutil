# Cryptoutil Code Quality & Linting TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 26, 2025
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
      - Files affected: config.go (1), config_test_util.go (1), certificates_server_test_util.go (1), jwe_jwk_util.go (1), jwk_util_test.go (1), jws_jwk_util.go (1), telemetry_service.go (1), application_listener.go (1), sql_provider.go (1), cicd_checks.go (2)
      - Highest complexity: 157 (jwk_util_test.go)
  - Configure appropriate settings for each linter
  - Test CI performance impact
- **Files**: `.golangci.yml`
- **Expected Outcome**: Enhanced code quality and consistency checks
- **Priority**: Medium - Code quality improvement

### Task CQ4: Add Initialism Comments in Configuration Files
- **Description**: Add clarifying comments for initialisms/acronyms in .golangci.yml and OpenAPI configuration files
- **Action Items**:
  - Review .golangci.yml for initialisms that need long-form explanations
  - Add comments to OpenAPI spec files (openapi_spec_*.yaml) explaining initialisms
  - Append comment to relevant lines specifying the long form of the initialism
  - Focus on crypto/security initialisms (JWE, JWK, JWS, AES, RSA, etc.)
- **Files**: `.golangci.yml`, `api/openapi_spec_*.yaml`
- **Expected Outcome**: Improved code readability and maintainability for future developers
- **Priority**: LOW - Documentation enhancement

### ✅ COMPLETED: Task CQ5: Review cicd_checks.go and cicd_checks_test.go Linter Exemptions
- **Description**: Review the comprehensive golangci-lint exemptions for scripts/cicd_checks.go and scripts/cicd_checks_test.go to identify which ones can be safely removed
- **Current State**: Both files were excluded from all 25+ enabled linters due to containing deliberate violations for testing purposes
- **Action Items**:
  - ✅ Analyzed each linter exemption to determine if it's actually needed
  - ✅ Tested removing individual linter exemptions one by one (bulk testing approach used for efficiency)
  - ✅ Verified that cicd_checks.go functionality still works after exemption removal
  - ✅ Updated .golangci.yml exclude-rules to remove unnecessary exemptions
  - ✅ Documented rationale for remaining exemptions
- **Files**: `.golangci.yml` (exclude-rules section), `scripts/cicd_checks.go`, `scripts/cicd_checks_test.go`
- **Expected Outcome**: Minimal but sufficient linter exemptions for cicd_checks files
- **Results**: 
  - `scripts/cicd_checks.go`: **NO exemptions needed** - all 25+ linters pass
  - `scripts/cicd_checks_test.go`: Only `gofumpt` and `goimports` exemptions needed (contains deliberate formatting violations for testing)
  - Reduced exemptions from 25+ linters each to 0 for cicd_checks.go and 2 for cicd_checks_test.go
- **Priority**: LOW - Optimization opportunity ✅ **COMPLETED**
