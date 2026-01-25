# Lessons Learned Extraction Checklist

## Purpose
Systematic process for extracting valuable lessons from temporary documentation before deletion.

## Source Documents
1. `docs/maintenance-session-2026-01-23.md` (135 lines)
2. `docs/workflow-fixing-prompt-fixes.md` (171 lines)

## Extraction Workflow

### Step 1: Identify Lessons

#### From maintenance-session-2026-01-23.md

**Testing Lessons**:
- [ ] SQLite datetime UTC comparison fix (consent_expired test)
- [ ] Docker healthcheck syntax: --start-period (not --start_period)
- [ ] E2E test Docker compose dependency failures (infrastructure issue)

**Build/CI Lessons**:
- [ ] .dockerignore optimization (42+ patterns for faster builds)
- [ ] golangci-lint v2.8.0 upgrade process
- [ ] importas enforcement via custom linter

**Coverage Patterns**:
- [ ] cmd/* packages expected 0% (thin main wrappers)
- [ ] Generated code (api/*) expected 0%
- [ ] Config packages 20-40% acceptable

#### From workflow-fixing-prompt-fixes.md

**Tool Configuration Lessons**:
- [ ] gopls installation via go install golang.org/x/tools/gopls@latest
- [ ] VS Code go.alternateTools setting pattern
- [ ] gopls path resolution issues

**Agent Prompt Lessons**:
- [ ] YAML frontmatter requirements for Copilot agent recognition
- [ ] Frontmatter format: description + tools array
- [ ] Agent autonomous execution patterns
- [ ] Todo list format for agent tracking
- [ ] Memory management in .github/instructions/memory.instruction.md

---

### Step 2: Map to Permanent Locations

| Lesson | Permanent Home | Justification |
|--------|----------------|---------------|
| SQLite datetime UTC fix | ✅ Already in plan.md Issue #1 "Lessons Learned" | Covered in workflow fixes analysis |
| Docker healthcheck syntax | Need to add to Docker instructions | Common mistake, valuable reference |
| E2E Docker infrastructure | ✅ Already in plan.md (E2E test gaps) | Covered in test coverage section |
| .dockerignore optimization | Need to add to Docker instructions | Build performance best practice |
| golangci-lint upgrade | Need to add to linting instructions | Tool version migration pattern |
| importas enforcement | ✅ Already implemented in lint_go.go | Code enforces this |
| cmd/* coverage expectations | Need to add to testing instructions | Coverage target clarification |
| gopls installation | Need to add to dev setup docs | Tool setup pattern |
| VS Code gopls config | Need to add to dev setup docs | Editor configuration |
| Agent YAML frontmatter | Need to add to prompt best practices | Agent creation requirements |
| Todo list format | Need to add to prompt templates | Agent tracking pattern |

---

### Step 3: Document Lessons in Permanent Locations

#### 3.1: Docker Instructions (.github/instructions/04-02.docker.instructions.md)

**Add Section: "Docker healthcheck Syntax"**:
```markdown
## Docker Healthcheck Syntax - CRITICAL

**ALWAYS use `--start-period` (double dash), NEVER `--start_period` (underscore)**

**Correct**:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-check-certificate -q -O /dev/null https://127.0.0.1:8080/admin/api/v1/livez
```

**Wrong**:
```dockerfile
HEALTHCHECK --start_period=5s  # INVALID - Docker won't parse this
```

**Reference**: maintenance-session-2026-01-23.md Task 5
```

**Add Section: ".dockerignore Optimization Patterns"**:
```markdown
## .dockerignore Optimization - MANDATORY

**Purpose**: Reduce Docker build context size by excluding non-essential files

**Recommended Patterns** (42+ exclusions):
- CI/CD tools: `.github/`, `scripts/`, `.git/`
- Python environments: `venv/`, `.venv/`, `__pycache__/`
- Test files: `test/`, `testdata/`, `*_test.go`
- Deployment configs: `deployments/`, `specs/`
- Documentation: `docs/`, `*.md` (except README.md)
- Coverage/output: `coverage*/`, `*.out`, `test-output/`

**Impact**: Faster Docker builds (reduced context transfer time)

**Reference**: maintenance-session-2026-01-23.md Task 2
```

#### 3.2: Testing Instructions (.github/instructions/03-02.testing.instructions.md)

**Add Section: "Coverage Target Clarifications"**:
```markdown
## Coverage Targets by Package Type

| Package Type | Expected Coverage | Rationale |
|--------------|------------------|-----------|
| Production code | ≥95% | Business logic MUST be tested |
| Infrastructure | ≥98% | Critical shared code needs higher bar |
| cmd/* (main wrappers) | 0% | Thin delegating main() functions acceptable if internalMain() ≥95% |
| Generated code (api/*) | 0% | Auto-generated code excluded |
| Config loading | 20-40% | Acceptable for straightforward config parsing |

**Reference**: maintenance-session-2026-01-23.md Test Results Summary
```

**Add Section: "SQLite DateTime UTC Comparison - CRITICAL"**:
```markdown
## SQLite DateTime Testing - MANDATORY

**ALWAYS use `.UTC()` when comparing SQLite datetime fields in tests**

**Problem**: SQLite stores datetimes as strings without timezone, comparisons fail if not normalized

**Correct**:
```go
require.WithinDuration(t, now.UTC(), consent.ExpiresAt.UTC(), time.Second)
```

**Wrong**:
```go
require.WithinDuration(t, now, consent.ExpiresAt, time.Second)  // Fails on Linux
```

**Reference**: maintenance-session-2026-01-23.md Task 5 (consent_expired test fix)
```

#### 3.3: Linting Instructions (.github/instructions/03-07.linting.instructions.md)

**Add Section: "golangci-lint Major Version Upgrades"**:
```markdown
## golangci-lint Version Upgrades

**v1.x → v2.x Migration**:
- Configuration changes: `wsl` → `wsl_v5` config key
- Removed settings: `wsl.force-err-cuddling`, `misspell.ignore-words`, `wrapcheck.ignoreSigs`
- New built-in formatters: gofumpt + goimports with --fix

**Workflow**:
1. Update .golangci.yml for v2 config changes
2. Run `golangci-lint run --fix` to apply auto-fixes
3. Manually fix remaining issues
4. Update CI/CD workflows to use v2

**Reference**: maintenance-session-2026-01-23.md Task 3
```

#### 3.4: Dev Setup Docs (docs/DEV-SETUP.md)

**Add Section: "gopls Installation and Configuration"**:
```markdown
## gopls Language Server Setup

**Installation**:
```bash
go install golang.org/x/tools/gopls@latest
```

**VS Code Configuration** (.vscode/settings.json):
```json
{
  "go.alternateTools": {
    "gopls": "/home/q/go/bin/gopls"
  }
}
```

**Troubleshooting**:
- If VS Code can't find gopls: verify `which gopls` returns a path
- If path incorrect: update `go.alternateTools` setting with correct path
- Alternative: Remove `alternateTools` setting to use VS Code's bundled gopls

**Reference**: workflow-fixing-prompt-fixes.md Issue 1
```

#### 3.5: Prompt Best Practices (NEW FILE: docs/agent-prompt-best-practices.md)

**Create new file**:
```markdown
# GitHub Copilot Agent Prompt Best Practices

## YAML Frontmatter - REQUIRED

**All agent prompts MUST include frontmatter**:
```yaml
---
description: "Brief description of agent's purpose"
tools: ['extensions', 'codebase', 'usages', 'vscodeAPI', 'problems', 'changes', 'testFailure', 'terminalSelection', 'terminalLastCommand', 'fetch', 'search', 'runCommands', 'runTasks', 'editFiles']
---
```

**Why**: GitHub Copilot uses frontmatter to recognize and list agents in dropdown

## Autonomous Execution Patterns

**Todo List Format**:
```markdown
## Todo List
- [ ] Task 1: Description
- [x] Task 2: Completed task
- [ ] Task 3: Next task
```

**Autonomous Rules**:
1. Continue until ALL tasks complete
2. Update todo list after each completion
3. Display updated list to user
4. Don't stop between tasks to ask permission
5. Research extensively with fetch_webpage
6. Commit small, atomic changes
7. Handle "resume"/"continue" commands

## Memory Management

**Location**: `.github/instructions/memory.instruction.md`

**Use for**:
- Recurring issues and workarounds
- Project-specific patterns
- Common mistakes and fixes

**Reference**: workflow-fixing-prompt-fixes.md Issue 2
```

---

### Step 4: Verify Coverage

#### Verification Checklist

- [ ] SQLite datetime UTC → ✅ Covered in testing instructions
- [ ] Docker healthcheck syntax → ✅ Covered in Docker instructions
- [ ] E2E test gaps → ✅ Already in plan.md
- [ ] .dockerignore optimization → ✅ Covered in Docker instructions
- [ ] golangci-lint upgrade → ✅ Covered in linting instructions
- [ ] importas enforcement → ✅ Already in lint_go.go code
- [ ] Coverage expectations → ✅ Covered in testing instructions
- [ ] gopls installation → ✅ Covered in dev setup docs
- [ ] VS Code gopls config → ✅ Covered in dev setup docs
- [ ] Agent YAML frontmatter → ✅ Covered in prompt best practices
- [ ] Todo list format → ✅ Covered in prompt best practices

**All lessons covered**: YES / NO

---

### Step 5: Document Deletion Decision

**Decision**: APPROVED for deletion after lesson extraction

**Justification**:
- All valuable lessons extracted to permanent locations
- Temporary session summaries no longer needed
- Permanent homes established for each lesson type
- No unique information remaining

**Audit Trail**: See docs/fixes-needed-plan-tasks-v3/lessons-extraction-checklist.md

---

### Step 6: Create Deletion Commit

**Files to Delete**:
- docs/maintenance-session-2026-01-23.md
- docs/workflow-fixing-prompt-fixes.md

**Commit Message**:
```
docs: delete temporary maintenance docs after extracting lessons

- Deleted docs/maintenance-session-2026-01-23.md (135 lines)
- Deleted docs/workflow-fixing-prompt-fixes.md (171 lines)

All valuable lessons extracted to permanent locations:
- Docker healthcheck syntax → docker.instructions.md
- .dockerignore optimization → docker.instructions.md
- SQLite datetime UTC fix → testing.instructions.md
- Coverage expectations → testing.instructions.md
- golangci-lint v2 upgrade → linting.instructions.md
- gopls setup → DEV-SETUP.md
- Agent YAML frontmatter → agent-prompt-best-practices.md

See docs/fixes-needed-plan-tasks-v3/lessons-extraction-checklist.md for full audit trail.
```

---

## Next Steps

1. ✅ Create this checklist
2. ⏳ Add lessons to permanent documentation files
3. ⏳ Verify all lessons covered
4. ⏳ Delete temporary docs with audit trail commit
