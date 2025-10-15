# Cryptoutil Code Quality & Linting TODOs

**Last Updated**: October 14, 2025
**Status**: Code quality enhancements planned for ongoing maintenance

---

## 🟡 MEDIUM - Code Quality & Linting Enhancements

### Task CQ1: Address TODO Comments in Codebase
- **Description**: Multiple TODO/FIXME comments found throughout codebase requiring attention
- **godox Issues**:
  - `internal/client/client_oam_mapper.go:89` - "TODO nil allowed if import not nil"
  - `internal/client/client_test.go:334` - "TODO validate public key does not contain any private key or secret key material"
  - `internal/client/client_test_util.go:22` - "TODO Add error checking for https with rootCAsPool=nil"
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

### Task CQ2: Enhanced Magic Number Detection ✅ COMPLETED
- **Description**: Improve detection of repeated strings and numeric constants
- **Current State**: Updated goconst configuration to maximum sensitivity
- **Action Items**:
  - ✅ Update goconst settings (min-len: 1, min-occurrences: 2, numbers: true)
  - ❌ Add gomnd (Go Magic Number Detector) linter - INCOMPATIBLE: requires typecheck which is disabled due to import resolution issues
    - **Root Cause**: gomnd performs semantic analysis requiring full type information and import resolution
    - **Conflict**: Project disables typecheck due to circular dependencies and complex module structure
    - **Impact**: Cannot use gomnd within golangci-lint pipeline
    - **Workaround**: Could run gomnd standalone, but not integrated with current linting workflow
  - Enable gocritic linter with magic number detection
  - Configure appropriate thresholds for hugeParam and rangeValCopy
- **Files**: `.golangci.yml`
- **Results**:
  - Previous run detected 4 magic string constants (fixed with named constants)
  - Current aggressive settings (min-len: 1) detect no additional issues - codebase is clean
  - Settings can be relaxed if needed, but current configuration provides maximum detection
- **Expected Outcome**: Better detection of magic strings and numeric literals
- **Priority**: Medium - Code quality improvement

### Task CQ3: Enable Additional Quality Linters
- **Description**: Add more golangci-lint linters for enhanced code quality
- **Current State**: Core linters enabled
- **Action Items**:
  - Evaluate and enable: exportloopref, gocognit, goheader, gomoddirectives, gomodguard, importas, lll, nlreturn, testpackage, wsl
  - Configure appropriate settings for each linter
  - Test CI performance impact
- **Files**: `.golangci.yml`
- **Expected Outcome**: Enhanced code quality and consistency checks
- **Priority**: Medium - Code quality improvement

### Task CQ4: Resolve Circular Dependencies
- **Description**: Fix circular import dependencies that prevent enabling typecheck and advanced linters
- **Current State**: Typecheck disabled in golangci-lint due to import resolution issues
- **Impact**: Prevents use of semantic analysis linters like gomnd
- **Action Items**:
  - Identify all circular dependency chains in the codebase
  - Refactor code to break circular dependencies
  - Test that typecheck can be re-enabled after fixes
  - Verify all existing functionality still works
- **Files**: All Go packages in internal/, pkg/, cmd/
- **Expected Outcome**: Clean dependency graph with no circular imports
- **Priority**: High - Enables advanced linting and improves code maintainability
- **Note**: This is a prerequisite for re-enabling gomnd and other semantic analysis tools

### Task CQ5: Add Circular Dependency Prevention
- **Description**: Implement automated detection and prevention of circular dependencies
- **Current State**: No automated checks for circular dependencies
- **Action Items**:
  - Research and evaluate Go circular dependency detection tools
  - Add appropriate linter or tool to CI/CD pipeline
  - Configure to fail builds on circular dependency detection
  - Document circular dependency prevention guidelines
- **Files**: CI/CD configuration, linting setup
- **Expected Outcome**: Automated prevention of future circular dependency issues
- **Priority**: Medium - Proactive code quality maintenance
- **Note**: Should integrate with existing golangci-lint or run as separate check

### Task CQ6: Re-enable gomnd After Circular Dependencies Fixed
- **Description**: Attempt to re-enable gomnd linter once circular dependencies are resolved
- **Current State**: gomnd removed due to typecheck incompatibility
- **Prerequisites**: CQ4 (circular dependencies) and CQ5 (prevention) completed
- **Action Items**:
  - Re-enable typecheck in golangci-lint configuration
  - Add gomnd back to enabled linters
  - Configure gomnd settings appropriately
  - Test that gomnd works with the fixed dependency structure
  - Fix any new magic number issues detected by gomnd
- **Files**: .golangci.yml, all Go source files
- **Expected Outcome**: Enhanced magic number detection with semantic analysis
- **Priority**: Medium - Advanced code quality improvement
- **Note**: Dependent on successful completion of CQ4 and CQ5

---

## Common Magic Values to Watch For

- HTTP status codes: `200`, `404`, `500`
- Timeouts: `30`, `60`, `300` (seconds)
- Buffer sizes: `1024`, `4096`
- Retry counts: `3`, `5`, `10`
- Port numbers: `8080`, `5432`
- String literals: `"localhost"`, `"admin"`, `"default"`
