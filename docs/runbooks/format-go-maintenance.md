# Format_go Maintenance Runbook

## Overview

The `format_go` command enforces the replacement of `interface{}` with `any` across the codebase. This runbook documents preventative measures against self-modification regressions and maintenance procedures.

## Self-Modification History

The `enforce_any.go` file has experienced multiple self-modification regressions where LLM agents inadvertently modified the file during refactoring:

1. **b934879b (Nov 17, 2025)**: Comments modified - backticks added to prevent pattern replacement
2. **b0e4b6ef (Dec 16, 2025)**: Infinite loop bug - counting logic incorrectly counted "any" instead of "interface{}"
3. **8c855a6e (Dec 16, 2025)**: Test data corruption - test expectations used "any" instead of "interface{}"
4. **71b0e90d (Nov 20, 2025)**: Added comprehensive self-exclusion patterns for all cicd commands

## Root Cause Analysis

**Primary Issue**: LLM agents (GitHub Copilot, Grok, Claude) lose exclusion context during narrow-focus refactoring tasks.

**Secondary Issue**: When reviewing only the function being modified, agents don't see:

- File-level exclusion patterns in `magic_cicd.go`
- Filter logic in `common/filter.go`
- Self-referential nature of the pattern replacement logic

## Protection Mechanisms

### 1. File-Level Exclusion Pattern

**Location**: `internal/shared/magic/magic_cicd.go`

```go
CICDSelfExclusionPatterns = map[string]string{
    "format-go": `internal[/\\]cmd[/\\]cicd[/\\]format_go[/\\].*\.go$`,
    // ... other commands
}
```

**Purpose**: Prevents `format_go` command from processing its own source files.

**Verification**:

```powershell
go run ./cmd/cicd format-go
# Should report "0 files modified" - never processes format_go/*.go
```

### 2. CRITICAL Comment Blocks

**Location**: `internal/cmd/cicd/format_go/enforce_any.go`

Two CRITICAL comment blocks document self-modification risks:

1. Function-level comment in `enforceAny()` (lines 16-22)
2. Inline comment in `processGoFile()` (lines 92-101)

**Purpose**: Warn LLM agents about self-modification context before making changes.

**Verification**: Grep for "CRITICAL" and "SELF-MODIFICATION PROTECTION" comments.

### 3. Test Data Pattern

**Location**: `internal/cmd/cicd/format_go/format_go_test.go`

**Pattern**: Tests MUST use `interface{}` in input data, verify replacement to `any`.

**Examples**:

```go
// CORRECT - input uses interface{}, expects any after replacement
testGoContentWithInterfaceEmpty = "var x interface{}"
require.Contains(t, result, "any", "File should contain 'any' after replacement")
require.NotContains(t, result, "interface{}", "File should not contain 'interface{}' after replacement")

// WRONG - input already uses any, breaks test
testGoContentWithInterfaceEmpty = "var x any"  // ❌ NEVER DO THIS
```

### 4. Counting Logic Pattern

**Location**: `internal/cmd/cicd/format_go/enforce_any.go` (line 100)

**Correct Pattern**:

```go
replacements := strings.Count(originalContent, "interface{}")  // ✅ Count source pattern
```

**Incorrect Pattern** (causes infinite loop):

```go
replacements := strings.Count(originalContent, "any")  // ❌ Counts result, not source
```

**Why**: Counting "any" in original content doesn't detect replacements - it counts existing "any" keywords, causing false positives.

### 5. Self-Modification Prevention Test

**Location**: `internal/cmd/cicd/format_go/self_modification_test.go`

**Purpose**: Automated verification that protection mechanisms remain intact.

**Checks**:

- CRITICAL comment blocks present
- Counting logic uses `interface{}` not `any`
- Test data uses `interface{}` as input
- Test expectations verify `any` after replacement

**Run Test**:

```powershell
go test ./internal/cmd/cicd/format_go -run TestEnforceAnyDoesNotModifyItself -v
```

## Warning Signs of Impending Self-Modification

When reviewing or modifying `format_go` code, watch for these red flags:

1. **Comment simplification**: Removing "verbose" CRITICAL comments about self-modification
2. **Pattern changes**: Changing comments from "`interface{}`" (with backticks) to "any"
3. **Test data changes**: Updating test constants from `interface{}` to `any`
4. **Counting logic changes**: "Fixing" the count to use "any" instead of "interface{}"
5. **Narrow-focus refactoring**: Modifying functions without reading full file context

## Maintenance Procedures

### Before Modifying format_go Code

1. **Read entire context**: Open and read `enforce_any.go`, `filter.go`, and `magic_cicd.go`
2. **Verify exclusion patterns**: Check `CICDSelfExclusionPatterns["format-go"]` still exists
3. **Run self-modification test**: Verify protection mechanisms intact
4. **Document changes**: Update this runbook if protection mechanisms change

### After Modifying format_go Code

1. **Run self-modification test**: `go test ./internal/cmd/cicd/format_go -run TestEnforceAnyDoesNotModifyItself -v`
2. **Verify no self-modifications**: `git diff internal/cmd/cicd/format_go/` should show only intended changes
3. **Check comment preservation**: CRITICAL comment blocks must remain intact
4. **Verify test data**: Test constants must still use `interface{}` as input

### If Self-Modification Detected

1. **Immediately revert**: `git checkout internal/cmd/cicd/format_go/enforce_any.go internal/cmd/cicd/format_go/format_go_test.go`
2. **Re-read context**: Read entire `enforce_any.go`, `filter.go`, `magic_cicd.go` before re-attempting
3. **Verify exclusion patterns**: Ensure `CICDSelfExclusionPatterns["format-go"]` not deleted
4. **Add more comments**: If regression persists, add more explicit warning comments
5. **Update this runbook**: Document new incident with commit hash and date

## Copilot Instructions Integration

The `.github/copilot-instructions.md` file contains a dedicated "Format_go Self-Modification Prevention" section documenting:

- Historical incidents with commit hashes
- Root cause analysis
- Protection mechanisms
- MANDATORY rules (what NEVER to do)
- Warning signs of impending self-modification
- Recovery procedures

**Purpose**: Ensure all LLM agents (GitHub Copilot, Grok, Claude) are aware of self-modification risks before modifying format_go code.

## Testing Strategy

### Unit Tests

- `TestEnforceAny_NoFiles`: Verify no-op when no files to process
- `TestEnforceAny_WithModifications`: Verify replacement works on external files
- `TestProcessGoFile_WithChanges`: Verify counting and replacement logic
- `TestEnforceAnyDoesNotModifyItself`: Verify protection mechanisms intact

### Integration Tests

Run `format_go` command against entire codebase:

```powershell
go run ./cmd/cicd format-go
```

**Expected**: Reports modifications to other files, but NEVER modifies `internal/cmd/cicd/format_go/*.go`.

### Manual Verification

1. Check git status after running format_go: `git status internal/cmd/cicd/format_go/`
2. Verify no changes to `enforce_any.go` or `format_go_test.go`
3. Verify CRITICAL comments still present: `grep -n "CRITICAL" internal/cmd/cicd/format_go/enforce_any.go`

## References

- **Exclusion Pattern Definition**: `internal/shared/magic/magic_cicd.go` lines 170-187
- **Filter Logic**: `internal/cmd/cicd/common/filter.go` lines 14-56
- **Enforcement Logic**: `internal/cmd/cicd/format_go/enforce_any.go`
- **Test Suite**: `internal/cmd/cicd/format_go/format_go_test.go`
- **Self-Modification Test**: `internal/cmd/cicd/format_go/self_modification_test.go`
- **Copilot Instructions**: `.github/copilot-instructions.md` lines 201-247

## Incident Log

| Date | Commit | Issue | Resolution |
|------|--------|-------|------------|
| Nov 17, 2025 | b934879b | Comments modified by pattern replacement | Added backticks to protect pattern name in comments |
| Nov 20, 2025 | 71b0e90d | Self-exclusion patterns incomplete | Added comprehensive patterns for all 12 cicd commands |
| Dec 16, 2025 | b0e4b6ef | Infinite loop - counted "any" not "interface{}" | Fixed counting logic to use source pattern |
| Dec 16, 2025 | 8c855a6e | Test data used "any" instead of "interface{}" | Fixed test constants to use source pattern as input |
| Dec 16, 2025 | 303babba | Copilot instructions missing warnings | Added format_go self-modification prevention section |
| Dec 16, 2025 | 3d94c4c6 | No automated test for protection mechanisms | Created TestEnforceAnyDoesNotModifyItself |

## Future Improvements

1. **Pre-commit hook**: Add validation to prevent format_go self-modifications from being committed
2. **CI/CD check**: Add workflow step to verify self-modification test passes
3. **Linter rule**: Custom golangci-lint rule to detect self-modification attempts
4. **File watching**: Pre-commit hook to block any changes to `format_go/*.go` files unless explicitly allowed
