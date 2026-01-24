# Documentation Clarification and Enhancement Tasks

## P0: Critical Tasks (COMPLETED)

### P0.1: Fix Terminology Confusion in plan.md ✅

**Description**: Correct misleading "SQLite + 0.0.0.0" terminology that conflates database choice with bind address validation.

**Acceptance Criteria**:
- [x] Issue #1 title changed from "SQLite Container Mode" to "Container Mode - Explicit Database URL Support"
- [x] Problem statement clarifies orthogonal concerns (database vs bind address)
- [x] Added "Key Insight" section explaining independence
- [x] Updated Cross-Cutting Issues section with correct terminology
- [x] Committed with conventional commit message

**Verification**:
```bash
# Check plan.md contains corrected terminology
grep -n "Container Mode - Explicit Database URL Support" docs/fixes-needed-plan-tasks-v2/plan.md
grep -n "orthogonal" docs/fixes-needed-plan-tasks-v2/plan.md
```

**Status**: ✅ COMPLETED (commit ca718194)

---

### P0.2: Create Session Tracking Infrastructure ✅

**Description**: Establish docs/fixes-needed-plan-tasks-v3/ with standardized tracking files.

**Acceptance Criteria**:
- [x] Created issues.md with 3 issues documented
- [x] Created categories.md with pattern analysis
- [x] Created lessons-extraction-checklist.md with 6-step workflow
- [x] Created plan.md (session overview)
- [x] Created tasks.md (this file)
- [x] Committed with conventional commit message

**Files Created**:
- docs/fixes-needed-plan-tasks-v3/issues.md (150+ lines)
- docs/fixes-needed-plan-tasks-v3/categories.md (200+ lines)
- docs/fixes-needed-plan-tasks-v3/lessons-extraction-checklist.md (311 lines)
- docs/fixes-needed-plan-tasks-v3/plan.md (this session)
- docs/fixes-needed-plan-tasks-v3/tasks.md (this file)

**Status**: ✅ COMPLETED (commits ca718194, 13fe43bb)

---

## P1: High Priority Tasks (IN PROGRESS)

### P1.1: Add Lessons to Docker Instructions

**File**: .github/instructions/04-02.docker.instructions.md

**Lessons to Add**:

1. **Docker healthcheck Syntax - CRITICAL** (after "Healthcheck" section):
```markdown
### Healthcheck Syntax Pitfall - CRITICAL

**ALWAYS use `--start-period` (hyphen), NEVER `--start_period` (underscore)**

**Problem**: Docker silently ignores `--start_period` (underscore), causing premature health checks during startup.

**Symptom**: Service marked unhealthy before initialization complete.

**Correct Pattern**:
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
  interval: 10s
  timeout: 5s
  retries: 3
  start_period: 40s  # ✅ CORRECT (hyphen)
```

**WRONG Pattern**:
```yaml
  start_period: 40s  # ❌ WRONG (underscore) - silently ignored
```

**Detection**: Add `docker inspect <container>` verification to CI/CD.

**Reference**: Docker Compose specification v3.8+
```

2. **.dockerignore Optimization - MANDATORY** (after "Multi-Stage Dockerfile Patterns" section):
```markdown
### .dockerignore Optimization - MANDATORY

**ALWAYS exclude development/test artifacts from Docker build context**

**Purpose**: Reduce build context size, prevent development secrets leakage, faster builds.

**Required Exclusions** (42+ patterns):
```dockerignore
# Version Control
.git
.gitignore
.gitattributes

# Development
.vscode/
.idea/
*.swp
*.swo
*~

# Testing
test-output/
testdata/
coverage*.out
coverage*.html
*_test.go
*_bench_test.go
*_fuzz_test.go

# Documentation
docs/
*.md
!README.md

# Build Artifacts
*.exe
*.dll
*.so
*.dylib
dist/
build/

# Dependencies
vendor/
node_modules/

# Configuration
*.env
*.local
secrets/

# Temporary
tmp/
temp/
*.tmp
*.log
```

**Verification**:
```bash
# Check build context size (should be <10MB for Go projects)
docker build --no-cache --progress=plain . 2>&1 | grep "transferring context"
```

**Impact**: Reduces build context from 200MB+ → <10MB, prevents leaking `.env` files into images.
```

**Acceptance Criteria**:
- [ ] Both sections added to 04-02.docker.instructions.md
- [ ] Code examples formatted correctly with syntax highlighting
- [ ] Verification commands tested
- [ ] File compiles/renders correctly

**Verification**:
```bash
grep -n "healthcheck Syntax Pitfall" .github/instructions/04-02.docker.instructions.md
grep -n ".dockerignore Optimization" .github/instructions/04-02.docker.instructions.md
```

---

### P1.2: Add Lessons to Testing Instructions

**File**: .github/instructions/03-02.testing.instructions.md

**Lessons to Add**:

1. **Coverage Targets by Package Type** (after "Coverage Targets - NO EXCEPTIONS" section):
```markdown
### Coverage Targets by Package Type

| Package Type | Minimum Coverage | Rationale |
|--------------|------------------|-----------|
| Production (internal/{jose,identity,kms,ca}) | 95% | Business logic requires comprehensive testing |
| Infrastructure (internal/cmd/cicd/*) | 98% | Build tools must be highly reliable |
| Utility (internal/shared/*, pkg/*) | 98% | Shared code used across services |
| Main Functions (cmd/*/main.go) | 0% acceptable IF internalMain() ≥95% | Delegate to testable internal function |
| Generated Code (api/*/openapi_gen_*.go) | 0% acceptable | Stable generated code, mutation testing adds no value |

**Pattern for Main Functions**:
```go
// cmd/app/main.go
func main() {
    os.Exit(internalMain(os.Args, os.Stdin, os.Stdout, os.Stderr))  // ≤5 lines, untested
}

// cmd/app/main_internal.go
func internalMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
    // 100+ lines of testable business logic
    // Target: ≥95% coverage
}
```

**Configuration-Specific Expectations**:
- Config structs: 20-40% coverage (field validation only)
- Config loading: ≥80% coverage (error paths critical)
- Config defaults: 100% coverage (must validate all defaults work)
```

2. **SQLite DateTime UTC Comparison - CRITICAL** (after "Test Data Isolation Requirements" section):
```markdown
### SQLite DateTime UTC Comparison - CRITICAL

**ALWAYS call `.UTC()` before comparing timestamps with SQLite**

**Problem**: SQLite stores timestamps in UTC but Go's `time.Now()` includes local timezone offset.

**Symptom**: Time comparisons fail with "off by N hours" in non-UTC timezones.

**WRONG Pattern**:
```go
beforeCreate := time.Now()
db.Create(&model)
db.First(&retrieved, model.ID)
assert.True(t, retrieved.CreatedAt.After(beforeCreate))  // ❌ FAILS in PST/EST/etc
```

**CORRECT Pattern**:
```go
beforeCreate := time.Now().UTC()
db.Create(&model)
db.First(&retrieved, model.ID)
assert.True(t, retrieved.CreatedAt.After(beforeCreate))  // ✅ WORKS in all timezones
```

**Alternative - Time Range Validation**:
```go
beforeCreate := time.Now().UTC()
db.Create(&model)
afterCreate := time.Now().UTC()
db.First(&retrieved, model.ID)

// Verify timestamp within creation window
assert.True(t, retrieved.CreatedAt.After(beforeCreate) || retrieved.CreatedAt.Equal(beforeCreate))
assert.True(t, retrieved.CreatedAt.Before(afterCreate) || retrieved.CreatedAt.Equal(afterCreate))
```

**Root Cause**: SQLite's datetime functions (`datetime('now')`) always return UTC, but Go's `time.Now()` respects `$TZ`.

**Detection**: Run tests with `TZ=America/Los_Angeles go test ./...` to catch timezone bugs.
```

**Acceptance Criteria**:
- [ ] Both sections added to 03-02.testing.instructions.md
- [ ] Table formatting correct
- [ ] Code examples tested
- [ ] Cross-references to other sections added

**Verification**:
```bash
grep -n "Coverage Targets by Package Type" .github/instructions/03-02.testing.instructions.md
grep -n "SQLite DateTime UTC Comparison" .github/instructions/03-02.testing.instructions.md
```

---

### P1.3: Add Lessons to Linting Instructions

**File**: .github/instructions/03-07.linting.instructions.md

**Lesson to Add**:

**golangci-lint Major Version Upgrades** (after "golangci-lint v2 Configuration" section):
```markdown
### golangci-lint Major Version Upgrades - CRITICAL

**ALWAYS review breaking changes before upgrading major versions**

**v1.x → v2.x Breaking Changes**:

| v1.x | v2.x | Migration |
|------|------|-----------|
| `wsl` | `wsl_v5` | Rename config key |
| `wsl.force-err-cuddling` | Removed (always enabled) | Delete config key |
| `misspell.ignore-words` | Removed | Use `allowlist` instead |
| `wrapcheck.ignoreSigs` | Removed | Use `ignorePackageGlobs` instead |
| External formatters | Built-in gofumpt/goimports | Use `--fix` flag |

**Migration Workflow**:
1. Read release notes: https://github.com/golangci/golangci-lint/releases
2. Update .golangci.yml config keys
3. Run `golangci-lint run --fix` to apply formatters
4. Fix remaining linting errors manually
5. Update CI/CD workflows to use new version
6. Update pre-commit hooks `.pre-commit-config.yaml`
7. Update docs/pre-commit-hooks.md

**Version Pinning**:
```yaml
# .github/workflows/ci-quality.yml
- name: golangci-lint
  uses: golangci/golangci-lint-action@v4
  with:
    version: v2.7.2  # ✅ Pin to specific version
```

**Pre-commit Hook**:
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v2.7.2  # ✅ Pin to match CI/CD
    hooks:
      - id: golangci-lint
```

**Common Migration Issues**:
- `wsl` errors increase after upgrade (v2 stricter)
- `godot` auto-fix requires `--fix` flag
- `importas` enforcement may require alias updates

**Testing After Upgrade**:
```bash
golangci-lint run --fix  # Apply auto-fixes
golangci-lint run        # Verify no remaining errors
go test ./...            # Verify tests still pass
```
```

**Acceptance Criteria**:
- [ ] Section added to 03-07.linting.instructions.md
- [ ] Table formatting correct
- [ ] Migration workflow steps clear
- [ ] Version pinning examples accurate

**Verification**:
```bash
grep -n "golangci-lint Major Version Upgrades" .github/instructions/03-07.linting.instructions.md
```

---

### P1.4: Add Lessons to Dev Setup Docs

**File**: docs/DEV-SETUP.md

**Lesson to Add**:

**gopls Installation and Configuration** (after "Prerequisites" section):
```markdown
### gopls Installation and Configuration

**ALWAYS install gopls separately for better performance and stability**

#### Installation

**Via Go Install** (recommended):
```bash
go install golang.org/x/tools/gopls@latest
```

**Verify Installation**:
```bash
gopls version
# Expected output: gopls v0.17.1 or newer
```

#### VS Code Configuration

**Add to `.vscode/settings.json`**:
```json
{
  "go.useLanguageServer": true,
  "gopls": {
    "ui.diagnostic.analyses": {
      "composites": false,
      "unusedparams": true,
      "unusedwrite": true
    },
    "ui.completion.usePlaceholders": true,
    "formatting.gofumpt": true
  }
}
```

**Enable Auto-Import**:
```json
{
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

#### Troubleshooting

**Issue**: gopls not found or not running

**Solutions**:
1. Verify `$GOPATH/bin` in `$PATH`: `echo $PATH | grep "$(go env GOPATH)/bin"`
2. Restart VS Code: `Ctrl+Shift+P` → "Developer: Reload Window"
3. Check gopls logs: `Ctrl+Shift+P` → "Go: Show Language Server Output Channel"

**Issue**: Slow performance or high memory usage

**Solutions**:
1. Increase memory limit in settings.json:
```json
{
  "gopls": {
    "ui.diagnostic.staticcheck": false,
    "build.buildFlags": ["-tags="]
  }
}
```
2. Exclude large directories:
```json
{
  "files.watcherExclude": {
    "**/vendor/**": true,
    "**/node_modules/**": true,
    "**/test-output/**": true
  }
}
```

**Issue**: Import errors or unresolved symbols

**Solutions**:
1. Run `go mod tidy` to fix dependencies
2. Clear gopls cache: `Ctrl+Shift+P` → "Go: Reset gopls Workspace"
3. Verify module initialization: check `go.mod` exists at workspace root
```

**Acceptance Criteria**:
- [ ] Section added to docs/DEV-SETUP.md
- [ ] Installation commands tested
- [ ] VS Code config examples valid JSON
- [ ] Troubleshooting steps verified

**Verification**:
```bash
grep -n "gopls Installation and Configuration" docs/DEV-SETUP.md
```

---

### P1.5: Create Agent Prompt Best Practices Doc

**File**: docs/agent-prompt-best-practices.md (NEW FILE)

**Content**:
```markdown
# Agent Prompt Best Practices

## Overview

This document captures best practices for creating and maintaining AI agent prompt files used in this repository. These patterns ensure consistent, effective agent behavior across different workflows.

---

## YAML Frontmatter - REQUIRED

**ALWAYS include YAML frontmatter at top of prompt files**

### Required Format

```yaml
---
description: "Brief one-line description of prompt purpose"
tools:
  - tool_name_1
  - tool_name_2
  - tool_name_3
---
```

### Purpose

1. **Agent Recognition**: VS Code Copilot and other LLM agents use frontmatter to identify prompt files
2. **Tool Discovery**: Lists required tools for agent capability verification
3. **Documentation**: Provides quick reference for prompt purpose

### Examples

**Workflow Fixing Prompt**:
```yaml
---
description: "Autonomous workflow execution with issue tracking and analysis"
tools:
  - file_search
  - read_file
  - replace_string_in_file
  - run_in_terminal
  - grep_search
---
```

**Beast Mode Prompt**:
```yaml
---
description: "7-step autonomous workflow with internet research and quality gates"
tools:
  - fetch_webpage
  - semantic_search
  - multi_replace_string_in_file
  - run_in_terminal
---
```

### Common Tools

| Tool | Purpose |
|------|---------|
| `file_search` | Find files by pattern |
| `read_file` | Read file contents |
| `grep_search` | Text search in files |
| `replace_string_in_file` | Edit files |
| `multi_replace_string_in_file` | Batch edits |
| `run_in_terminal` | Execute commands |
| `fetch_webpage` | Internet research |
| `semantic_search` | AI-powered code search |

---

## Autonomous Execution Patterns

**ALWAYS structure prompts for fully autonomous execution**

### Core Pattern

```markdown
1. Initial Analysis Phase
   - Gather context
   - Identify scope
   - Create tracking documents

2. Planning Phase
   - Create plan.md
   - Create tasks.md
   - (Optional) Create QUIZME.md

3. Execution Phase
   - Work through tasks sequentially
   - Update tracking as you go
   - Commit incrementally

4. Verification Phase
   - Run quality gates
   - Verify all criteria met
   - Document completion

5. Analysis Phase
   - Create issues.md
   - Create categories.md
   - Identify patterns
```

### Autonomous Directives

**Include these directives in prompts**:

```markdown
- Work continuously until ALL tasks complete OR user clicks STOP
- NEVER ask permission between tasks
- NEVER pause for status updates
- Make decisions autonomously based on context
- Commit after each logical unit of work
- Document issues as encountered
```

### Quality Gates

**Define clear completion criteria**:

```markdown
Task NOT complete until:
- [ ] Code builds: `go build ./...`
- [ ] Linting passes: `golangci-lint run`
- [ ] Tests pass: `go test ./...`
- [ ] Coverage maintained: ≥95% production, ≥98% infrastructure
- [ ] Documentation updated
- [ ] Conventional commit made
```

---

## Session Tracking Integration

**ALWAYS include session tracking workflow in prompts**

### Standard Location

```
docs/fixes-needed-plan-tasks-v#/
├── issues.md           # Specific problems encountered
├── categories.md       # Pattern analysis
├── plan.md            # Session overview
├── tasks.md           # Actionable checklist
└── QUIZME.md          # (Optional) Questions for user
```

### Tracking Workflow

**During Execution**:
```markdown
1. Encounter issue → Add to issues.md
2. Multiple similar issues → Update categories.md
3. Complete task → Check off in tasks.md
4. Session end → Create plan.md summarizing work
```

**Post-Execution**:
```markdown
1. Analyze issues.md for patterns
2. Update categories.md with root causes
3. Create systematic fixes section
4. Identify prevention strategies
```

### Issue Template

```markdown
## Issue #N: [Title]

**Date**: YYYY-MM-DD
**Category**: [Documentation Clarity | Process Improvement | Tooling Enhancement]
**Severity**: [Low | Medium | High | Critical]
**Status**: [Identified | In Progress | Resolved]

**Problem Description**: ...

**Root Cause**: ...

**Impact**: ...

**Fix Required**: ...

**Files Affected**: ...

**Lessons Learned**: ...
```

---

## Memory Management

**ALWAYS manage prompt context efficiently**

### Todo List Pattern

**Use structured todo lists for tracking**:

```markdown
## Implementation Checklist

- [ ] Task 1: Description (Status: Not Started)
- [ ] Task 2: Description (Status: Not Started)
- [x] Task 3: Description (Status: Complete)
- [ ] Task 4: Description (Status: Blocked - waiting for X)
```

### Progress Updates

**Update task status inline**:
```markdown
- [x] P0.1: Fix terminology ✅ (commit abc1234)
- [~] P1.2: Add lessons to docs (3 of 5 files complete)
- [ ] P2.1: Enhance prompts (not started)
```

### Context Preservation

**Reference previous work explicitly**:
```markdown
Building on commit abc1234 where we fixed X...
Continuing from Task 3 where we identified Y...
As discovered in categories.md, pattern Z occurs...
```

---

## Prompt Enhancement Guidelines

### When to Create QUIZME.md

**Create QUIZME.md when**:
- Ambiguous requirements (need clarification)
- Multiple implementation options (user must choose)
- Missing domain knowledge (external input required)

**Skip QUIZME.md when**:
- Tasks straightforward (clear implementation path)
- Conventions established (follow existing patterns)
- Research possible (can fetch_webpage for answers)

### Quality Over Speed

**ALWAYS prioritize**:
- ✅ Correctness over completion speed
- ✅ Comprehensive testing over minimal coverage
- ✅ Thorough validation over quick fixes
- ✅ Evidence-based completion over claimed completion

### Continuous Work Directive

**Pattern to include**:
```markdown
CRITICAL: Work continuously until user clicks STOP button

- Complete task → Commit → IMMEDIATELY start next task
- No status updates between tasks
- No asking permission to continue
- No celebrations followed by stopping
```

---

## Common Anti-Patterns

### ❌ NEVER Do

**Stopping Behaviors**:
- Asking "Should I proceed with X?"
- Pausing after each task completion
- Presenting options and waiting for user choice
- Summarizing progress instead of continuing

**Documentation Issues**:
- Missing YAML frontmatter
- Creating session-specific docs outside tracking system
- Leaving TODOs without tracking
- Incomplete extraction before deleting temp docs

**Quality Shortcuts**:
- Skipping quality gates to finish faster
- Marking tasks complete without evidence
- Assuming tests pass without running them
- Committing without verification

### ✅ ALWAYS Do

**Continuous Execution**:
- Work through entire task list
- Commit after each logical unit
- Document issues as encountered
- Update tracking incrementally

**Quality Assurance**:
- Run all quality gates
- Verify with objective evidence
- Test across platforms if applicable
- Update documentation alongside code

**Session Management**:
- Create tracking directory first
- Update issues.md when problems arise
- Analyze patterns in categories.md
- Create comprehensive plan.md at end

---

## Workflow Integration Examples

### Beast Mode with Session Tracking

```markdown
1. Internet Research Phase
   - Fetch requirements from URLs
   - Document findings in research.md
   - Create issues.md if blockers found

2. Planning Phase
   - Create docs/fixes-needed-plan-tasks-v#/
   - Generate plan.md from research
   - Generate tasks.md with priorities
   - Generate QUIZME.md if needed

3. Implementation Phase
   - Execute tasks sequentially
   - Update issues.md as problems arise
   - Commit after each task
   - Track progress in tasks.md

4. Analysis Phase
   - Update categories.md with patterns
   - Document lessons learned
   - Create systematic fixes
   - Commit final summary
```

### Workflow Fixing with Quality Gates

```markdown
1. Analysis Phase
   - Read workflow logs
   - Identify failure patterns
   - Create issues.md tracking

2. Fix Phase
   - Apply fixes incrementally
   - Verify each fix: `act -j workflow-name`
   - Document in issues.md
   - Commit with conventional message

3. Verification Phase
   - Run all affected workflows
   - Check artifacts uploaded
   - Verify logs show success
   - Update issues.md status

4. Documentation Phase
   - Update tracking documents
   - Create categories.md analysis
   - Document prevention strategies
```

---

## References

- [01-02.continuous-work.instructions.md](.github/instructions/01-02.continuous-work.instructions.md)
- [06-01.evidence-based.instructions.md](.github/instructions/06-01.evidence-based.instructions.md)
- [01-03.speckit.instructions.md](.github/instructions/01-03.speckit.instructions.md)
```

**Acceptance Criteria**:
- [ ] File created as docs/agent-prompt-best-practices.md
- [ ] All 5 major sections complete (Frontmatter, Autonomous, Tracking, Memory, Enhancement)
- [ ] Examples tested for accuracy
- [ ] Cross-references to copilot instructions added
- [ ] Formatting renders correctly in Markdown

**Verification**:
```bash
ls -la docs/agent-prompt-best-practices.md
grep -n "YAML Frontmatter" docs/agent-prompt-best-practices.md
grep -n "Session Tracking Integration" docs/agent-prompt-best-practices.md
```

---

## P2: Medium Priority Tasks

### P2.1: Verify All Lessons Covered

**Description**: Systematic verification that all 11 lessons from checklist are documented.

**Verification Checklist**:

#### From maintenance-session-2026-01-23.md:
- [ ] SQLite datetime UTC comparison (in 03-02.testing.instructions.md)
- [ ] Docker healthcheck --start-period syntax (in 04-02.docker.instructions.md)
- [ ] E2E test infrastructure gaps (in 03-02.testing.instructions.md)
- [ ] .dockerignore 42+ exclusion patterns (in 04-02.docker.instructions.md)
- [ ] golangci-lint v2 migration workflow (in 03-07.linting.instructions.md)
- [ ] importas enforcement patterns (in 03-07.linting.instructions.md)
- [ ] Coverage expectations by package type (in 03-02.testing.instructions.md)

#### From workflow-fixing-prompt-fixes.md:
- [ ] gopls installation via `go install` (in docs/DEV-SETUP.md)
- [ ] YAML frontmatter required for agents (in docs/agent-prompt-best-practices.md)
- [ ] Autonomous execution patterns (in docs/agent-prompt-best-practices.md)
- [ ] Todo list tracking (in docs/agent-prompt-best-practices.md)

**Acceptance Criteria**:
- [ ] All 11 checklist items verified in permanent homes
- [ ] No unique information remains in temp docs
- [ ] Cross-references added where applicable

**Verification**:
```bash
# Check each permanent file contains expected content
grep -c "SQLite DateTime UTC" .github/instructions/03-02.testing.instructions.md  # Should be >0
grep -c "healthcheck Syntax" .github/instructions/04-02.docker.instructions.md    # Should be >0
grep -c "gopls Installation" docs/DEV-SETUP.md                                   # Should be >0
grep -c "YAML Frontmatter" docs/agent-prompt-best-practices.md                   # Should be >0
```

---

### P2.2: Delete Temporary Maintenance Docs

**Description**: Remove temporary docs after verifying lessons extracted.

**Files to Delete**:
1. docs/maintenance-session-2026-01-23.md
2. docs/workflow-fixing-prompt-fixes.md

**Prerequisites**:
- [x] All lessons identified (P1.5 checklist)
- [ ] All lessons added to permanent homes (P1.1-P1.5)
- [ ] Verification complete (P2.1)

**Acceptance Criteria**:
- [ ] Files deleted from filesystem
- [ ] Committed with audit trail message
- [ ] Commit message references extraction checklist

**Commit Message Template**:
```
docs: remove temporary maintenance docs after lesson extraction

All lessons from these documents have been extracted and documented
in permanent locations:

From maintenance-session-2026-01-23.md:
- Docker healthcheck syntax → 04-02.docker.instructions.md
- .dockerignore optimization → 04-02.docker.instructions.md
- SQLite UTC comparison → 03-02.testing.instructions.md
- Coverage targets → 03-02.testing.instructions.md
- golangci-lint v2 → 03-07.linting.instructions.md

From workflow-fixing-prompt-fixes.md:
- gopls installation → docs/DEV-SETUP.md
- YAML frontmatter → docs/agent-prompt-best-practices.md
- Autonomous patterns → docs/agent-prompt-best-practices.md

See docs/fixes-needed-plan-tasks-v3/lessons-extraction-checklist.md
for complete extraction workflow and verification.

No information lost - all lessons preserved in permanent documentation.
```

**Commands**:
```bash
git rm docs/maintenance-session-2026-01-23.md
git rm docs/workflow-fixing-prompt-fixes.md
git commit -F commit_message.txt
```

---

### P2.3: Enhance workflow-fixing.prompt.md

**File**: .github/prompts/workflow-fixing.prompt.md

**Enhancements to Add**:

1. **Session Tracking Section** (after frontmatter):
```markdown
## Session Tracking - MANDATORY

**ALWAYS create session tracking directory before starting work**

### Directory Structure

```
docs/fixes-needed-plan-tasks-v#/
├── issues.md           # Specific problems encountered during execution
├── categories.md       # Pattern analysis across issues
├── plan.md            # Session overview (created after completion)
├── tasks.md           # Actionable checklist
└── QUIZME.md          # (Optional) Questions requiring user input
```

### Workflow

1. **Start**: Create `docs/fixes-needed-plan-tasks-v#/` (increment version number)
2. **During Work**: Update `issues.md` when problems encountered
3. **Pattern Analysis**: Update `categories.md` when multiple similar issues
4. **End**: Create `plan.md` and `tasks.md` summarizing session
5. **Optional**: Create `QUIZME.md` if clarifications needed

### Issue Template

```markdown
## Issue #N: [Title]

**Date**: YYYY-MM-DD
**Category**: [Documentation | Process | Tooling | Testing | Build]
**Severity**: [Low | Medium | High | Critical]
**Status**: [Identified | In Progress | Resolved]

**Problem**: ...
**Root Cause**: ...
**Impact**: ...
**Fix**: ...
**Files Affected**: ...
**Lessons**: ...
```
```

2. **Quality Gates Section** (before "Todo List" section):
```markdown
## Quality Gates - MANDATORY

**NEVER mark workflow fixes complete without verification**

### Verification Checklist

For each workflow fix:
- [ ] Fix applied to workflow file
- [ ] Workflow syntax validated: `act -l -W .github/workflows/workflow-name.yml`
- [ ] Workflow runs successfully: `act -j job-name`
- [ ] Artifacts uploaded correctly (if applicable)
- [ ] Logs show expected behavior
- [ ] Committed with conventional message

### Evidence Requirements

**Build/Lint Workflows**:
- `go build ./...` clean
- `golangci-lint run` clean
- Exit code 0

**Test Workflows**:
- All tests pass
- Coverage maintained/improved
- Artifacts contain expected files

**E2E Workflows**:
- Services start successfully
- Health checks pass
- Test scenarios complete
- Containers cleaned up

### Post-Fix Analysis

After fixing workflows:
1. Document root cause in issues.md
2. Identify pattern (one-off vs systemic)
3. Update categories.md if pattern found
4. Create prevention strategy
5. Update documentation if needed
```

**Acceptance Criteria**:
- [ ] Session tracking section added
- [ ] Quality gates section added
- [ ] Examples updated to reference tracking
- [ ] File structure valid

**Verification**:
```bash
grep -n "Session Tracking - MANDATORY" .github/prompts/workflow-fixing.prompt.md
grep -n "Quality Gates - MANDATORY" .github/prompts/workflow-fixing.prompt.md
```

---

### P2.4: Enhance beast-mode-3.1.prompt.md

**File**: .github/prompts/beast-mode-3.1.prompt.md

**Enhancements to Add**:

1. **Post-Completion Analysis** (after Step 7):
```markdown
## Step 8: Post-Completion Analysis - MANDATORY

**ALWAYS analyze session after all tasks complete**

### Create Session Tracking

1. **Create directory**: `docs/fixes-needed-plan-tasks-v#/` (increment version)

2. **Create issues.md**:
   - List all problems encountered during implementation
   - Categorize by type (Documentation, Process, Tooling, etc.)
   - Document root causes and lessons learned

3. **Create categories.md**:
   - Analyze patterns across issues
   - Identify systemic problems
   - Propose prevention strategies

4. **Create plan.md**:
   - Session overview
   - Issues addressed
   - Key insights
   - Success criteria checklist

5. **Create tasks.md**:
   - Actionable checklist for next session
   - Priority levels (P0, P1, P2, P3)
   - Acceptance criteria
   - Verification commands

6. **(Optional) Create QUIZME.md**:
   - Questions requiring user clarification
   - Multiple choice options
   - Context for each question

### Analysis Template

```markdown
# Session Analysis v#

## What Went Well
- Successfully completed X
- Discovered pattern Y
- Improved process Z

## Challenges Encountered
1. Issue: ...
   - Root Cause: ...
   - Solution: ...
   - Prevention: ...

## Patterns Identified
- Pattern 1: ... (occurred N times)
- Pattern 2: ... (systemic issue)

## Lessons Learned
1. ...
2. ...

## Next Steps
- Task 1: ... (Priority P1)
- Task 2: ... (Priority P2)
```

### Quality Verification

Before ending session:
- [ ] All todo items checked off
- [ ] All quality gates passed
- [ ] Documentation updated
- [ ] Session tracking created
- [ ] Conventional commits made
- [ ] Lessons documented
```

**Acceptance Criteria**:
- [ ] Step 8 added after existing 7 steps
- [ ] Analysis template included
- [ ] Integration with existing workflow clear

**Verification**:
```bash
grep -n "Step 8: Post-Completion Analysis" .github/prompts/beast-mode-3.1.prompt.md
```

---

### P2.5: Enhance autonomous-execution.prompt.md

**File**: .github/prompts/autonomous-execution.prompt.md

**Enhancements to Add**:

1. **Session Tracking Integration** (after "Continuous Work Directive" section):
```markdown
## Session Tracking System - MANDATORY

**ALWAYS maintain incremental documentation during execution**

### Tracking Location

Standard location: `docs/fixes-needed-plan-tasks-v#/` (version incremented per session)

### Required Files

1. **issues.md** (created FIRST, updated throughout session):
   ```markdown
   ## Issue #1: [Title]
   
   **Date**: YYYY-MM-DD
   **Status**: [Identified | In Progress | Resolved]
   
   Problem: ...
   Root Cause: ...
   Fix: ...
   Lessons: ...
   ```

2. **categories.md** (updated when patterns emerge):
   ```markdown
   ## Category: [Name]
   
   **Frequency**: N occurrences
   **Pattern**: ...
   **Root Causes**: ...
   **Systematic Fix**: ...
   **Prevention**: ...
   ```

3. **plan.md** (created at END of session):
   - Session overview
   - Issues addressed
   - Key insights
   - Success criteria

4. **tasks.md** (created at END of session):
   - Actionable checklist
   - Priority levels
   - Acceptance criteria
   - Verification commands

### Workflow Integration

```
1. Encounter Issue
   ↓
2. Add to issues.md immediately
   ↓
3. Continue work (don't pause)
   ↓
4. Multiple similar issues?
   YES → Update categories.md
   NO → Continue
   ↓
5. Fix applied?
   YES → Update issue status, commit
   NO → Document blocker, continue other work
   ↓
6. All tasks complete?
   YES → Create plan.md + tasks.md
   NO → Continue execution
```

### Continuous Execution Rules

**Session tracking does NOT pause work**:
- ✅ Update issues.md while continuing execution
- ✅ Commit incremental changes with tracking updates
- ✅ Pattern analysis happens during natural breaks
- ❌ NEVER pause to ask "Should I update tracking?"
- ❌ NEVER stop after updating issues.md
```

2. **Analysis Phase Definition** (before "Quality Gates" section):
```markdown
## Analysis Phase - POST-EXECUTION ONLY

**ALWAYS create comprehensive analysis after ALL tasks complete**

### When to Trigger

Analysis phase begins when:
- [x] All tasks in checklist complete
- [x] All quality gates passed
- [x] All verification commands successful
- [x] All commits made
- [x] Literally nothing left to implement

### Analysis Deliverables

1. **Final plan.md**:
   - Executive summary
   - Issues addressed (from issues.md)
   - Key insights
   - Success criteria checklist
   - Metrics (files modified, commits made, etc.)

2. **Final tasks.md**:
   - All tasks with completion status
   - Priority levels (P0, P1, P2, P3)
   - Acceptance criteria
   - Verification evidence

3. **Pattern Analysis**:
   - Review categories.md
   - Identify cross-cutting themes
   - Propose systemic fixes
   - Document prevention strategies

4. **(Optional) QUIZME.md**:
   - Create ONLY if genuinely unknown answers
   - Multiple choice + write-in format
   - Provide context for each question

### Analysis Anti-Patterns

❌ **Creating analysis instead of doing work**
❌ **Pausing execution to analyze mid-session**
❌ **Asking "Should I create plan.md now?"**
❌ **Stopping after partial task completion to analyze**

✅ **Work through ALL tasks FIRST**
✅ **Analyze comprehensively AFTER completion**
✅ **Update tracking incrementally DURING work**
```

**Acceptance Criteria**:
- [ ] Session tracking section added
- [ ] Analysis phase clearly defined as post-execution
- [ ] Workflow integration diagram included
- [ ] Anti-patterns documented

**Verification**:
```bash
grep -n "Session Tracking System" .github/prompts/autonomous-execution.prompt.md
grep -n "Analysis Phase - POST-EXECUTION" .github/prompts/autonomous-execution.prompt.md
```

---

## P3: Low Priority Tasks (Optional)

### P3.1: Consider QUIZME.md Creation

**Description**: Evaluate whether clarifying questions needed for this session.

**Evaluation Criteria**:

**Create QUIZME.md IF**:
- Ambiguous requirements (how should X work?)
- Multiple valid approaches (user must choose)
- Missing domain knowledge (what is standard for Y?)

**Skip QUIZME.md IF** (current situation):
- Tasks straightforward (add sections to existing files)
- Clear examples provided (lessons already written)
- No ambiguity (exact content specified in checklist)

**Decision**: SKIP QUIZME.md for this session

**Rationale**:
- All lesson content already written in extraction checklist
- Permanent home locations clearly identified
- No ambiguous requirements or choices needed
- Straightforward implementation of P1 tasks

---

## Summary

**Total Tasks**: 12 (2 P0, 6 P1, 4 P2, 0 P3)

**Status**:
- ✅ P0: Completed (2/2)
- 🔄 P1: In Progress (0/6 complete)
- ⏳ P2: Pending (0/4)
- ⏸️ P3: Skipped (0/0)

**Next Immediate Actions**:
1. Add Docker lessons to 04-02.docker.instructions.md (P1.1)
2. Add Testing lessons to 03-02.testing.instructions.md (P1.2)
3. Add Linting lessons to 03-07.linting.instructions.md (P1.3)
4. Add Dev setup lessons to docs/DEV-SETUP.md (P1.4)
5. Create docs/agent-prompt-best-practices.md (P1.5)
6. Verify all lessons covered (P2.1)
7. Delete temp docs (P2.2)
8. Enhance 3 prompt files (P2.3-P2.5)

**Estimated Completion**: After completing P1-P2 tasks (approximately 8 files to modify/create)
