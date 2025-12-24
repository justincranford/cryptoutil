# Coding Patterns and Standards Specifications

**Version**: 1.0.0
**Last Updated**: 2025-12-24
**Referenced By**: `.github/instructions/03-01.coding.instructions.md`

## File Size Limits

**Soft limit**: 300 lines (ideal target)
**Medium limit**: 400 lines (acceptable with justification)
**Hard limit**: 500 lines → refactor required

**Why Size Limits Matter**:
- Faster LLM processing and token usage
- Easier human review and maintenance
- Better code organization and discoverability
- Forces logical component grouping

## Code Patterns

### Default Values

**ALWAYS declare default values as named variables** rather than inline literals

**Example**:
```go
var defaultConfigFiles = []string{"config.yaml", "app.yaml"}
var defaultPort = 8080

// CORRECT
server.Start(defaultPort, defaultConfigFiles)

// WRONG
server.Start(8080, []string{"config.yaml", "app.yaml"})
```

### Pass-through Calls

**Prefer same parameter and return value order** as helper functions

Maintains API consistency and reduces confusion when chaining calls.

### Context Reading Before Refactoring - CRITICAL

**CRITICAL: ALWAYS read complete context before refactoring or modifying code**

**Why This Matters**:
- LLM agents lose exclusion context during narrow-focus refactoring
- "Verbose comments" may be intentional protection (e.g., self-modification warnings)
- Design patterns may not be obvious from single file view
- Simplifications can break critical invariants

**NEVER refactor in isolation**:

```bash
# ❌ WRONG: Just read the target file
read_file enforce_any.go
# Agent sees "verbose comments" and "simplifies" them
# Agent sees `interface{}` and "modernizes" to "any"
# Result: Self-modification protection bypassed
```

**ALWAYS read complete package context**:

```bash
# ✅ CORRECT: Read ALL related context files
read_file enforce_any.go              # Target file
read_file filter.go                   # Self-exclusion patterns
read_file magic_cicd.go               # Exclusion constants
read_file format_go_test.go           # Test data patterns
read_file self_modification_test.go   # Validation patterns
# Now understand WHY verbose comments exist
# Now understand WHY `interface{}` is intentional
```

**Key Questions Before Refactoring**:

1. Why does this code exist? (Read README, post-mortems, git log)
2. What protections are in place? (Read self-exclusion patterns, test files)
3. Are "verbose" comments intentional? (Check for CRITICAL/SELF-MODIFICATION tags)
4. What tests validate this behavior? (Read test files, understand assertions)
5. Has this failed before? (Check docs/P0.* post-mortems, LESSONS-LEARNED)

**Pattern Recognition**:

- **CRITICAL comments**: NEVER simplify or remove without understanding purpose
- **SELF-MODIFICATION PROTECTION**: NEVER change `interface{}` to `any` in format_go package
- **Backticked strings**: Intentional protection against replacement (e.g., `` `interface{}` ``)
- **Test data patterns**: May use "wrong" values intentionally (e.g., `interface{}` as input to test replacement)

**Common Refactoring Mistakes**:

- Changing `` `interface{}` `` to `` `any` `` in comments (breaks replacement logic)
- Removing "verbose" CRITICAL comments (loses protection context)
- Modernizing test data from `interface{}` to `any` (breaks test intent)
- Simplifying exclusion patterns without understanding full context

### Format_go Self-Modification Prevention - CRITICAL

**CRITICAL WARNING: enforce_any.go self-modification regression has occurred MULTIPLE times**

**Historical Incidents**:

1. **b934879b (Nov 17)**: Added backticks to comments to prevent pattern replacement
2. **b0e4b6ef (Dec 16)**: Fixed infinite loop (counted "any" instead of "interface{}")
3. **8c855a6e (Dec 16)**: Fixed test data (used "any" instead of "interface{}")
4. **71b0e90d (Nov 20)**: Added comprehensive self-exclusion patterns

**Root Cause**: LLM agents lose exclusion context during narrow-focus refactoring

**MANDATORY Rules**:

- ❌ **NEVER modify comments/test data in enforce_any.go or format_go_test.go**
- ❌ **NEVER change `interface{}` to `any` in format_go package without reading full context**
- ✅ **ALWAYS read complete context before refactoring self-modifying code**

## Conditional Statement Chaining

### CRITICAL: Pattern for Mutually Exclusive Conditions

**PREFER SWITCH STATEMENTS** over `if/else if/else` chains for cleaner, more maintainable code

**ALWAYS prefer switch statements** when possible:

```go
// ✅ CORRECT: Switch statement for multiple exclusive conditions
switch {
case ctx == nil:
    return nil, fmt.Errorf("nil context")
case logger == nil:
    return nil, fmt.Errorf("nil logger")
case description == "":
    return nil, fmt.Errorf("empty description")
default:
    return processValid(ctx, logger, description)
}
```

**ALWAYS prefer chained if/else if/else for mutually exclusive conditions**:

```go
// ✅ CORRECT: Chained if/else if/else for mutually exclusive conditions
if ctx == nil {
    return nil, fmt.Errorf("nil context")
} else if logger == nil {
    return nil, fmt.Errorf("nil logger")
} else if description == "" {
    return nil, fmt.Errorf("empty description")
}
```

**Avoid separate if statements for mutually exclusive conditions**:

```go
// ❌ WRONG: Separate if statements for mutually exclusive conditions
if ctx == nil {
    return nil, fmt.Errorf("nil context")
}
if logger == nil {
    return nil, fmt.Errorf("nil logger")
}
```

### When NOT to Chain

- Independent conditions (not mutually exclusive)
- Error accumulation patterns
- Cases with early returns that don't overlap

## Key Takeaways

1. **File Size Limits**: 300 (soft) / 400 (medium) / 500 (hard) lines → refactor at 500
2. **Default Values**: Always declare as named variables, never inline literals
3. **Context Reading**: ALWAYS read complete package context before refactoring
4. **Self-Modification Protection**: CRITICAL warnings exist for historical regression prevention
5. **Switch Statements**: Prefer over if/else if/else chains for cleaner code
6. **Refactoring Anti-Patterns**: Never simplify CRITICAL comments, backticked strings, test data patterns
