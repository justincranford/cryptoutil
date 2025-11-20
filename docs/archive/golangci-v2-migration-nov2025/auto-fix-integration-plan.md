# golangci-lint Auto-Fix Integration Plan

## Overview

This document outlines the plan to integrate automated linting fixes that are NOT covered by `golangci-lint --fix` into the `cicd` utility. These auto-fixes were developed during the session to resolve 242 lint issues systematically.

## Analysis of Chat Session

### Auto-Fix Scripts Identified

From the conversation summary, the following PowerShell batch replacements were used successfully:

#### 1. **errcheck** - Defer Closures Pattern
**Pattern**: Convert resource cleanup to defer with error handling
```powershell
# Pattern: resp.Body.Close() → defer func() { _ = resp.Body.Close() }()
# Applied to: HTTP response bodies, file handles
```

**Potential cicd command**: `go-fix-errcheck-defer-closures`
- Target: HTTP response bodies (`.Body.Close()`), file handles
- Complexity: HIGH (requires understanding of function context, variable scopes)
- Value: MEDIUM (common pattern, but context-sensitive)
- **Decision**: SKIP - Too context-dependent for automated fix

#### 2. **wsl_v5** - Whitespace Consistency
**Pattern**: Already handled by `golangci-lint run --fix`
- **Decision**: NO ACTION NEEDED (golangci-lint handles this)

#### 3. **wrapcheck** - File-Level Nolint
**Pattern**: Add `//nolint:wrapcheck` before package declaration
```powershell
# Pattern: Insert after copyright, before package declaration
```

**Potential cicd command**: `go-fix-wrapcheck-handlers`
- Target: Fiber handler files (handlers_*.go in identity/)
- Complexity: LOW (simple text insertion at specific location)
- Value: LOW (project-specific pattern, rare need)
- **Decision**: SKIP - Too project-specific, manual placement is clearer

#### 4. **thelper** - Test Helper Annotation
**Pattern**: Add `t.Helper()` as first line in test helper functions
```go
func setupTest(t *testing.T) {
    t.Helper() // <-- Add this
    // ... setup code
}
```

**Potential cicd command**: `go-fix-thelper`
- Target: Test helper functions (named setup*, check*, assert*, verify*)
- Complexity: MEDIUM (requires AST parsing to identify helper functions)
- Value: HIGH (common pattern, improves test failure reporting)
- **Decision**: CANDIDATE - High value, well-defined pattern

#### 5. **tparallel** - Parallel Test Cleanup
**Pattern**: Convert `defer cleanup()` → `t.Cleanup(func() { cleanup() })`
```go
// Before:
defer cleanup()

// After:
t.Cleanup(func() { cleanup() })
```

**Potential cicd command**: `go-fix-tparallel-cleanup`
- Target: Test files with `defer` calls inside parallel tests
- Complexity: HIGH (requires understanding of test structure, parallel context)
- Value: MEDIUM (important for parallel tests, but requires context)
- **Decision**: SKIP - Too context-dependent (needs to verify t.Parallel() exists)

#### 6. **mnd** - Magic Number Detection
**Pattern**: Extract magic numbers to named constants
```powershell
# Pattern: Identify repeated numbers, create constants
```

**Potential cicd command**: `go-fix-mnd`
- Target: Repeated numeric literals
- Complexity: VERY HIGH (semantic analysis, constant naming, scope determination)
- Value: HIGH (improves code clarity)
- **Decision**: SKIP - Requires semantic understanding, manual naming is better

#### 7. **goconst** - Constant Extraction
**Pattern**: Extract repeated strings to named constants
```go
// Before:
if mode == "container-mode-disabled" { ... }
if mode == "container-mode-disabled" { ... }

// After:
const containerModeDisabled = "container-mode-disabled"
if mode == containerModeDisabled { ... }
```

**Potential cicd command**: `go-fix-goconst`
- Target: Repeated string literals (3+ occurrences)
- Complexity: VERY HIGH (semantic analysis, constant naming, scope determination)
- Value: HIGH (reduces duplication)
- **Decision**: SKIP - Requires semantic understanding, manual naming is better

#### 8. **copyloopvar** - Remove Unnecessary Loop Variable Captures
**Pattern**: Remove `tc := tc` in Go 1.25+ parallel tests
```go
// Before (Go 1.22):
for _, tc := range tests {
    tc := tc // Capture for closure
    t.Run(tc.name, func(t *testing.T) { ... })
}

// After (Go 1.25+):
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

**Potential cicd command**: `go-fix-copyloopvar`
- Target: Remove `varName := varName` inside range loops
- Complexity: MEDIUM (AST parsing, verify Go version >= 1.25)
- Value: MEDIUM (cleanup, but not critical)
- **Decision**: CANDIDATE - Clean pattern, version-specific optimization

#### 9. **staticcheck** - Lowercase Error Strings
**Pattern**: Lowercase error message first word
```go
// Before:
return fmt.Errorf("Missing openapi spec")

// After:
return fmt.Errorf("missing openapi spec")
```

**Potential cicd command**: `go-fix-staticcheck-error-strings`
- Target: Error strings starting with uppercase
- Complexity: LOW (regex-based replacement)
- Value: HIGH (Go style convention)
- **Decision**: CANDIDATE - Simple, well-defined, high value

#### 10. **noctx** - Add Context Parameters
**Pattern**: Convert operations to use context
```go
// Database:
db.Exec() → db.ExecContext(ctx)
db.Query() → db.QueryContext(ctx)
db.Ping() → db.PingContext(ctx)

// Network:
net.Listen() → (&net.ListenConfig{}).Listen(context.Background())
tls.Dial() → (&tls.Dialer{}).DialContext(context.Background())

// Exec:
exec.Command() → exec.CommandContext(context.Background())
```

**Potential cicd command**: `go-fix-noctx`
- Target: Database, network, exec operations without context
- Complexity: VERY HIGH (requires understanding function signatures, adding imports, context source)
- Value: HIGH (improves cancellation, timeouts)
- **Decision**: SKIP - Too complex, context source is semantic decision

## Selected Commands for Implementation

Based on complexity, value, and automation feasibility:

### 1. `go-fix-staticcheck-error-strings` (Priority: HIGH)
**What**: Lowercase the first word of error strings
**Why**: Simple, well-defined, follows Go conventions
**Implementation**:
- Use regex to find `fmt.Errorf("`, `errors.New("` patterns
- Check if first character after quote is uppercase letter (excluding acronyms)
- Lowercase the first character
- Preserve acronyms (HTTP, URL, etc.)

### 2. `go-fix-copyloopvar` (Priority: MEDIUM)
**What**: Remove unnecessary `varName := varName` in loops (Go 1.25+)
**Why**: Cleanup obsolete pattern for modern Go versions
**Implementation**:
- Check `go.mod` for Go version >= 1.25
- Use AST to find range loops with closure calls
- Identify `varName := varName` statements
- Remove if inside range loop with closure

### 3. `go-fix-thelper` (Priority: MEDIUM)
**What**: Add `t.Helper()` to test helper functions
**Why**: Improves test failure reporting
**Implementation**:
- Use AST to find functions in `*_test.go` files
- Identify helper patterns: `setup*`, `check*`, `assert*`, `verify*`, `create*` with `*testing.T` parameter
- Check if first statement is `t.Helper()`
- Add if missing

### 4. `go-fix-all` (Priority: HIGH)
**What**: Run all auto-fix commands in sequence
**Why**: Convenience command for comprehensive auto-fixing
**Implementation**:
- Execute all go-fix-* commands in dependency order
- Report results for each command

## Implementation Phases

### Phase 1: Infrastructure (Estimated: 2-4 hours)
1. Create `internal/cmd/cicd/cicd_go_fix.go` - Main auto-fix dispatcher
2. Create `internal/cmd/cicd/cicd_go_fix_test.go` - Test infrastructure
3. Add command constants and registration in `cicd.go`
4. Update `magic_cicd.go` with new command names and usage

### Phase 2: `go-fix-staticcheck-error-strings` (Estimated: 2-3 hours)
1. Implement in `internal/cmd/cicd/cicd_go_fix_staticcheck.go`
2. Create tests in `cicd_go_fix_staticcheck_test.go`
3. Test with examples from the session (api/server/util.go)
4. Document edge cases (acronyms, proper nouns)

### Phase 3: `go-fix-copyloopvar` (Estimated: 4-6 hours)
1. Implement in `internal/cmd/cicd/cicd_go_fix_copyloopvar.go`
2. Use `go/ast` and `go/parser` for AST analysis
3. Create tests in `cicd_go_fix_copyloopvar_test.go`
4. Test with examples from session (sql_edge_cases_test.go, sql_final_coverage_test.go)
5. Add Go version check logic

### Phase 4: `go-fix-thelper` (Estimated: 4-6 hours)
1. Implement in `internal/cmd/cicd/cicd_go_fix_thelper.go`
2. Use `go/ast` for function analysis
3. Create tests in `cicd_go_fix_thelper_test.go`
4. Test with examples from session (sql_final_coverage_test.go, sql_postgres_coverage_test.go)
5. Document helper function naming conventions

### Phase 5: `go-fix-all` (Estimated: 1-2 hours)
1. Implement command orchestration
2. Add result aggregation and reporting
3. Create integration tests
4. Document recommended usage workflow

### Phase 6: Integration (Estimated: 2-3 hours)
1. Update `.pre-commit-config.yaml` with new cicd commands
2. Update `docs/pre-commit-hooks.md` with command documentation
3. Update copilot instructions (`.github/instructions/01-06.linting.instructions.md`)
4. Update README.md with new commands
5. Add to `docs/DEV-SETUP.md` developer workflow

### Phase 7: Testing & Documentation (Estimated: 2-3 hours)
1. Run all new commands against full codebase
2. Verify no regressions
3. Update test coverage reports
4. Create usage examples
5. Document limitations and edge cases

## Total Estimated Time: 17-27 hours

## Success Criteria

1. **Functionality**:
   - All three auto-fix commands work correctly on real codebase
   - `go-fix-all` orchestrates all commands successfully
   - Pre-commit hooks integrate seamlessly

2. **Quality**:
   - 95%+ test coverage for cicd auto-fix code
   - No false positives in auto-fixes
   - Idempotent operations (running twice produces no changes)

3. **Documentation**:
   - Clear usage examples in README
   - Integration documented in pre-commit-hooks.md
   - Copilot instructions updated
   - Edge cases and limitations documented

4. **Performance**:
   - Commands complete in <10 seconds for full codebase
   - Caching prevents redundant work
   - Pre-commit hooks remain fast (<30 seconds total)

## Future Enhancements (Post-MVP)

1. **Auto-fix for noctx** (complex, requires context source analysis)
2. **Auto-fix for goconst/mnd** (requires semantic analysis and naming)
3. **Auto-fix for errcheck defer closures** (requires scope analysis)
4. **Interactive mode** for ambiguous cases
5. **Dry-run mode** for preview without changes
6. **Statistics reporting** (files processed, fixes applied)

## Non-Goals

- **NOT implementing**: Fixes that require semantic understanding (noctx, goconst, mnd)
- **NOT implementing**: Context-dependent fixes (errcheck defer, tparallel cleanup, wrapcheck handlers)
- **NOT implementing**: Fixes already handled by `golangci-lint --fix` (wsl_v5, nlreturn)

## References

- Session conversation: 242 lint issues resolved systematically
- Existing cicd commands: `go-enforce-test-patterns`, `go-enforce-any`, `go-check-identity-imports`
- golangci-lint v2.6.2 configuration: `.golangci.yml`
- Test infrastructure: `internal/cmd/cicd/cicd_test.go`
