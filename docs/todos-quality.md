# Cryptoutil Code Quality & Linting TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 16, 2025
**Status**: Active code quality enhancements in progress

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
    - Line 241: "TODO Add specific rate limiting for Swagger UI endpoints"
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

---

## Common Magic Values to Watch For

- HTTP status codes: `200`, `404`, `500`
- Timeouts: `30`, `60`, `300` (seconds)
- Buffer sizes: `1024`, `4096`
- Retry counts: `3`, `5`, `10`
- Port numbers: `8080`, `5432`
- String literals: `"localhost"`, `"admin"`, `"default"`
