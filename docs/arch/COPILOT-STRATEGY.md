# Copilot Enhancement Strategy

**Version**: 1.0.0  
**Last Updated**: 2026-01-18  
**Purpose**: Identify concrete prompts, agents, and instruction improvements from awesome-copilot best practices

---

## What I DON'T Want

- ❌ Collections (unnecessary organizational overhead)
- ❌ MCP servers (overcomplicated for my needs)
- ❌ Skills directory (adds complexity without clear benefit)
- ❌ Reams of documentation (want concise, actionable guidance)

## What I DO Want

**Focus Areas**:
1. **Instructions best practices** - What patterns from awesome-copilot can improve my 28 instruction files?
2. **Useful prompts** - Concrete examples I can adapt and use
3. **Useful agents** - Concrete examples showing when to use agents vs prompts
4. **Clear guidance** - When to use prompts vs agents vs instructions

---

## Instructions Best Practices

### Current State (28 files)

**Strengths**:
- Good organization (01-methodology, 02-architecture, 03-implementation, 04-cicd, 05-platform, 06-quality)
- Tactical guidance format with quick reference sections
- Cross-references between related files
- Comprehensive coverage of technology stack

**Gaps** (from awesome-copilot patterns):

**1. Missing `applyTo` Glob Patterns**:
- **Problem**: All 28 instructions apply globally (no conditional application)
- **awesome-copilot pattern**: Use `applyTo` frontmatter to scope instructions
- **Examples**:
  ```yaml
  ---
  description: "Go testing patterns"
  applyTo: "**/*_test.go"
  ---
  ```
  ```yaml
  ---
  description: "Docker configuration"
  applyTo: "Dockerfile|docker-compose.yml|deployments/**"
  ---
  ```

**Action**: Add `applyTo` patterns to all 28 instruction files (see QUIZME-v2 for patterns)

**2. Quick Reference Format Inconsistency**:
- **Problem**: Some files have tables, some have lists, some have code blocks
- **awesome-copilot pattern**: Consistent format (tables for complex, lists for simple)
- **Example**:
  ```md
  ## Quick Reference
  
  | Pattern | When to Use | Example |
  |---------|-------------|---------|
  | Table-driven tests | 3+ similar cases | `tests := []struct{...}` |
  | Parallel tests | Independent test cases | `t.Parallel()` |
  ```

**Action**: Standardize quick reference sections across all instructions

**3. Missing Anti-Pattern Sections**:
- **Problem**: Only 1 file (06-02) has comprehensive anti-patterns
- **awesome-copilot pattern**: Every instruction file should include "NEVER DO" sections
- **Example**:
  ```md
  ## Common Mistakes
  
  ❌ NEVER use `0.0.0.0` in tests (triggers Windows Firewall)
  ❌ NEVER skip `defer resp.Body.Close()` (resource leak)
  ✅ ALWAYS use `127.0.0.1` for loopback binding
  ```

**Action**: Add anti-pattern sections to high-risk instruction files (testing, database, docker, security)

**4. Missing Code-Inline Examples**:
- **Problem**: Some patterns described but no code examples
- **awesome-copilot pattern**: Show concrete code, not just descriptions
- **Example** (current):
  ```md
  Use constructor injection for dependencies
  ```
- **Example** (improved):
  ```md
  Use constructor injection for dependencies:
  
  ```go
  // ✅ CORRECT
  func NewService(db *gorm.DB, logger *zap.Logger) *Service {
      return &Service{db: db, logger: logger}
  }
  
  // ❌ WRONG
  func NewService() *Service {
      db := getGlobalDB()  // Hidden dependency
      return &Service{db: db}
  }
  ```
  ```

**Action**: Add code examples to abstract patterns (10-15 instruction files need updates)

---

## Useful Prompts (From awesome-copilot)

### Prompts I Should Create

**1. Code Review Prompt** (`.github/prompts/code-review.prompt.md`):
```yaml
---
description: "Perform comprehensive code review"
argument-hint: "file paths to review"
tools: [read_file, grep_search, list_code_usages]
---

# Code Review Workflow

**Input**: ${file:filesToReview}

**Steps**:
1. Check code quality (naming, structure, duplication)
2. Verify test coverage exists
3. Check for security issues (SQL injection, XSS, secrets)
4. Verify error handling patterns
5. Check for race conditions (shared state, goroutines)
6. Verify proper resource cleanup (defer, Close())

**Output**: Markdown report with findings categorized by severity
```

**2. Refactor Extract Function** (`.github/prompts/refactor-extract.prompt.md`):
```yaml
---
description: "Extract function from large function body"
argument-hint: "file path and line range"
tools: [read_file, replace_string_in_file, runTests]
---

# Extract Function Refactor

**Input**: 
- ${file:targetFile}
- ${input:startLine} to ${input:endLine}
- ${input:newFunctionName}

**Workflow**:
1. Read target file and identify code block
2. Analyze dependencies (params, return values, side effects)
3. Create new function with extracted code
4. Replace original code with function call
5. Run tests to verify behavior unchanged
6. Update documentation if needed

**Quality Gates**:
- ✅ Tests pass after extraction
- ✅ No new linting warnings
- ✅ Function complexity reduced
```

**3. Generate Tests** (`.github/prompts/test-generate.prompt.md`):
```yaml
---
description: "Generate comprehensive test suite for function/type"
argument-hint: "function or type name"
tools: [read_file, semantic_search, create_file, runTests]
---

# Test Generation Workflow

**Input**: ${input:functionOrTypeName}

**Steps**:
1. Find function/type definition
2. Analyze signature (params, returns, errors)
3. Generate table-driven test with cases:
   - Happy path (valid inputs)
   - Error cases (nil, empty, invalid)
   - Edge cases (boundary values)
4. Generate property-based tests if applicable
5. Run tests and verify coverage ≥95%

**Output**: 
- `*_test.go` file with comprehensive tests
- Coverage report
```

**4. Fix Bug** (`.github/prompts/fix-bug.prompt.md`):
```yaml
---
description: "Systematic bug investigation and fix"
argument-hint: "bug description or error message"
tools: [semantic_search, read_file, grep_search, get_errors, runTests]
---

# Bug Fix Workflow

**Input**: ${input:bugDescription}

**Steps**:
1. Reproduce bug (create failing test)
2. Identify root cause (trace execution, read related code)
3. Propose fix with minimal changes
4. Implement fix
5. Verify tests pass
6. Add regression test to prevent recurrence

**Quality Gates**:
- ✅ Failing test demonstrates bug
- ✅ Fix resolves test failure
- ✅ All existing tests still pass
- ✅ Regression test added
```

**5. Optimize Performance** (`.github/prompts/optimize-performance.prompt.md`):
```yaml
---
description: "Identify and fix performance bottlenecks"
argument-hint: "package or function to optimize"
tools: [read_file, semantic_search, run_in_terminal]
---

# Performance Optimization Workflow

**Input**: ${input:targetPackage}

**Steps**:
1. Run benchmarks: `go test -bench=. -benchmem ${targetPackage}`
2. Identify hot paths (allocations, execution time)
3. Analyze algorithmic complexity
4. Propose optimizations:
   - Reduce allocations (preallocate, reuse buffers)
   - Improve algorithms (O(n²) → O(n log n))
   - Add caching for expensive operations
5. Implement changes
6. Re-run benchmarks, verify improvement
7. Ensure tests still pass

**Quality Gates**:
- ✅ Benchmarks show measurable improvement
- ✅ No functionality regressions
- ✅ Code readability maintained
```

**6. Generate Documentation** (`.github/prompts/generate-docs.prompt.md`):
```yaml
---
description: "Generate comprehensive package/function documentation"
argument-hint: "package path or function name"
tools: [read_file, list_dir, create_file]
---

# Documentation Generation

**Input**: ${input:packagePath}

**Steps**:
1. Read package files
2. Extract exported types, functions, constants
3. Generate godoc comments for undocumented items
4. Create package-level documentation
5. Generate examples for complex functions
6. Create README if missing

**Output**:
- Godoc comments added to code
- README.md with usage examples
- Example tests demonstrating usage
```

---

## Useful Agents (From awesome-copilot)

### When to Use Agents vs Prompts

**Use Prompts When**:
- Task is well-defined and sequential
- No complex decision trees
- Can complete in single execution
- Example: "Generate tests for function X"

**Use Agents When**:
- Task requires multiple phases with decisions
- Need to hand off to specialized sub-agents
- Long-running investigation needed
- Example: "Refactor entire package to use new pattern"

### Agents I Should Create

**1. Security Audit Agent** (`.github/agents/expert.security.agent.md`):
```yaml
---
description: "Comprehensive security audit of codebase"
handoffs:
  - label: "Found SQL injection risk"
    target: "expert.database"
    prompt: "Review parameterized query usage in ${affectedFiles}"
  - label: "Found XSS risk"
    target: "expert.testing"
    prompt: "Create security tests for input sanitization in ${affectedFiles}"
---

# Security Audit Agent

## Setup
- Load all .go files
- Load all SQL files
- Load all YAML configs

## Workflow
1. Scan for common vulnerabilities:
   - SQL injection (raw string concatenation in queries)
   - XSS (unescaped HTML output)
   - Secrets in code (hardcoded passwords, API keys)
   - Weak crypto (MD5, SHA-1, weak random)
2. Analyze authentication flows
3. Check authorization patterns (zero-trust violations)
4. Review input validation
5. Check error messages (info disclosure)
6. Verify HTTPS/TLS configuration

## Output
- Security report with severity ratings
- Concrete code examples of violations
- Recommended fixes with patches
```

**2. Performance Analyst Agent** (`.github/agents/expert.performance.agent.md`):
```yaml
---
description: "Analyze and improve code performance"
handoffs:
  - label: "Found database N+1 query"
    target: "expert.database"
    prompt: "Optimize query patterns in ${affectedFiles}"
  - label: "Found excessive allocations"
    target: "expert.testing"
    prompt: "Add benchmarks and verify optimization in ${affectedFiles}"
---

# Performance Analyst Agent

## Setup
- Run benchmarks across all packages
- Collect profiling data (CPU, memory, allocations)

## Workflow
1. Identify hot paths (CPU time)
2. Identify allocation hotspots (memory pressure)
3. Analyze algorithmic complexity (O(n²) loops)
4. Check for common anti-patterns:
   - String concatenation in loops (use strings.Builder)
   - Unbounded goroutines (use worker pools)
   - N+1 database queries (use joins/batching)
   - Excessive reflection (cache reflect.Type)
5. Propose optimizations with benchmarks
6. Implement changes
7. Verify improvements

## Output
- Performance report with before/after benchmarks
- Specific code changes with evidence
```

**3. Database Expert Agent** (`.github/agents/expert.database.agent.md`):
```yaml
---
description: "Database design, optimization, and migration expert"
handoffs:
  - label: "Schema needs indexing"
    target: "expert.performance"
    prompt: "Benchmark query performance before/after index in ${affectedFiles}"
---

# Database Expert Agent

## Setup
- Load all GORM models
- Load all migration files
- Load all repository files

## Workflow
1. Review schema design:
   - Normalization (appropriate denormalization)
   - Indexes (missing, unused, redundant)
   - Constraints (FK, CHECK, UNIQUE)
   - Data types (appropriate for usage)
2. Analyze query patterns:
   - N+1 queries (suggest eager loading)
   - Full table scans (suggest indexes)
   - Inefficient joins (suggest query rewrite)
3. Review transaction usage:
   - Missing transactions (data consistency)
   - Unnecessary transactions (performance)
   - Deadlock risks (lock ordering)
4. Migration quality:
   - Rollback compatibility
   - Zero-downtime migrations
   - Data migration correctness

## Output
- Database optimization report
- Index recommendations with rationale
- Query rewrites with benchmarks
- Migration improvements
```

**4. Testing Expert Agent** (`.github/agents/expert.testing.agent.md`):
```yaml
---
description: "Comprehensive testing strategy and implementation"
handoffs:
  - label: "Need integration tests"
    target: "expert.database"
    prompt: "Review database test patterns in ${affectedFiles}"
  - label: "Need E2E tests"
    target: "speckit.implement"
    prompt: "Implement E2E tests for ${feature}"
---

# Testing Expert Agent

## Setup
- Analyze current test coverage
- Identify untested code paths
- Load mutation testing reports

## Workflow
1. Coverage analysis:
   - Identify RED lines in HTML coverage report
   - Categorize gaps (error paths, edge cases, integration)
2. Generate tests:
   - Unit tests (≥95% production, ≥98% infrastructure)
   - Integration tests (database, external services)
   - E2E tests (both /service/** and /browser/** paths)
   - Property-based tests (gopter for complex functions)
3. Mutation testing:
   - Run gremlins per package
   - Fix surviving mutants (weak tests)
   - Target ≥85% production, ≥98% infrastructure
4. Test quality:
   - Remove flaky tests (race conditions, timing)
   - Fix test data isolation (UUIDv7 for uniqueness)
   - Verify no hardcoded passwords (use magic constants)

## Output
- Comprehensive test suite
- Coverage reports (≥95%/≥98%)
- Mutation scores (≥85%/≥98%)
- Test quality improvements
```

---

## Prompts vs Agents - Decision Matrix

| Scenario | Use Prompt | Use Agent | Rationale |
|----------|-----------|-----------|-----------|
| Generate tests for 1 function | ✅ | ❌ | Single-step, well-defined |
| Generate tests for entire package | ❌ | ✅ | Requires analysis, prioritization, iteration |
| Fix 1 known bug | ✅ | ❌ | Reproduce, fix, test - linear workflow |
| Investigate performance regression | ❌ | ✅ | Profiling, analysis, multiple fixes, validation |
| Extract 1 function | ✅ | ❌ | Single refactor with tests |
| Refactor package to new pattern | ❌ | ✅ | Multiple files, interdependencies, migration |
| Review 1 PR file | ✅ | ❌ | Checklist-based review |
| Comprehensive security audit | ❌ | ✅ | Multiple vulnerability types, context-dependent |
| Generate godoc for 1 type | ✅ | ❌ | Template-based generation |
| Generate full package docs | ❌ | ✅ | Package overview, examples, cross-references |
| Add 1 index to database | ✅ | ❌ | Write migration, test |
| Optimize database schema | ❌ | ✅ | Analyze all queries, propose changes, benchmarks |

**Rule of Thumb**:
- **Prompts** = Single file, single purpose, <30 min work, linear steps
- **Agents** = Multiple files, investigation needed, >30 min work, complex decisions

---

## Implementation Priority

### Phase 1: High-Value Prompts (Week 1)
1. `code-review.prompt.md` (daily use)
2. `test-generate.prompt.md` (daily use)
3. `fix-bug.prompt.md` (weekly use)
4. `refactor-extract.prompt.md` (weekly use)

### Phase 2: Specialized Agents (Week 2-3)
1. `expert.testing.agent.md` (coverage gaps, mutation testing)
2. `expert.security.agent.md` (security audits)
3. `expert.performance.agent.md` (optimization work)
4. `expert.database.agent.md` (schema design, query optimization)

### Phase 3: Instruction Improvements (Week 4)
1. Add `applyTo` patterns to all 28 instruction files
2. Standardize quick reference sections
3. Add anti-pattern sections to high-risk files
4. Add code examples to abstract patterns

---

## Success Metrics

**Prompts**:
- ✅ Used ≥5 times per week (daily workflow integration)
- ✅ Save ≥30 minutes per use (efficiency gain)
- ✅ Consistent output quality (90%+ of runs useful)

**Agents**:
- ✅ Reduce manual investigation time by 50%
- ✅ Improve code quality (fewer bugs, better performance)
- ✅ Increase test coverage (≥95%/≥98% targets met consistently)

**Instructions**:
- ✅ Reduce Copilot mistakes (fewer corrections needed)
- ✅ Faster onboarding (new team members productive faster)
- ✅ Consistent code style (fewer PR review cycles)

---

## Next Steps

1. Answer QUIZME-v2 questions (applyTo patterns, prompt priorities, agent specialization)
2. Implement Phase 1 prompts (4 high-value prompts)
3. Test prompts with real work (collect feedback)
4. Implement Phase 2 agents (4 expert agents)
5. Update instructions with patterns (Phase 3)
