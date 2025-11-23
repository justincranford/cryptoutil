# LLM Agent Quick Reference for Feature Template

**Purpose**: Condensed checklist for LLM agents during autonomous feature implementation sessions

**Read This First**: Before starting any multi-task feature implementation

**Full Template**: See `feature-template.md` for complete details

---

## 🎯 PRIMARY DIRECTIVE: NEVER STOP UNTIL ALL TASKS COMPLETE

**Token Budget**: Work until 950k/1M tokens used (95% utilization)

**Stop Conditions**: ONLY when tokens ≥950k OR explicit user command

**Not Stop Conditions**: Time elapsed, tasks complete, commits made, summaries provided

---

## ⚡ Continuous Work Pattern

```text
START → Read task doc → Implement → Test → Commit → Mark complete → IMMEDIATELY next task → ...
```

**ZERO TEXT between tool calls:**

- ❌ WRONG: commit → "Working on Task 2..." → create_file
- ✅ RIGHT: commit → create_file (zero characters between)

**Tool call chaining:**

```text
read_file → create_file → runTests → run_in_terminal (commit) → manage_todo_list → read_file (next)
```

---

## 📋 Pre-Implementation Checklist

Before starting feature:

- [ ] Read master plan document thoroughly
- [ ] Read ALL task documents (01-##.md)
- [ ] Identify critical path dependencies
- [ ] Check for parallel execution opportunities
- [ ] Review instruction files (.github/instructions/*.md)

---

## 🔄 Per-Task Loop

**For EACH task until ALL complete:**

### 1. Pre-Implementation (2 min)

- [ ] Read task doc: `read_file` on `##-<TASK>.md`
- [ ] Understand acceptance criteria
- [ ] Check dependencies complete
- [ ] Review related code patterns

### 2. Implementation (varies)

- [ ] Create/modify files per spec
- [ ] Follow coding standards (see instructions)
- [ ] Add inline documentation (godoc)
- [ ] Handle errors explicitly

### 3. Testing (5-10 min)

- [ ] Write table-driven tests
- [ ] Test happy + sad paths
- [ ] Use `t.Parallel()` always
- [ ] Run: `runTests` tool (NEVER `go test`)
- [ ] Achieve coverage target (≥85% infra, ≥80% features)

### 4. Quality (5 min)

- [ ] Auto-fix: `golangci-lint run --fix`
- [ ] Fix remaining issues manually
- [ ] Verify no TODOs introduced
- [ ] Check import aliases correct

### 5. Commit (1 min)

- [ ] Stage: `git add <files>`
- [ ] Commit: `git commit -m "type(scope): description"`
- [ ] Use `--no-verify` for speed during iteration

### 6. Post-Mortem (5-10 min)

- [ ] Create: `##-<TASK>-POSTMORTEM.md`
- [ ] Document bugs/fixes
- [ ] Document omissions
- [ ] List corrective actions
- [ ] Identify instruction violations
- [ ] **Create new task docs** for corrective actions requiring significant work
- [ ] **Add subtasks to manage_todo_list** for quick fixes

### 7. Handoff (0 min - IMMEDIATE)

- [ ] Mark complete: `manage_todo_list`
- [ ] **IMMEDIATELY** read next task doc
- [ ] **NO STOPPING, NO SUMMARY**

---

## 🚫 Anti-Patterns to Avoid

**NEVER do these:**

- ❌ Stop after commits
- ❌ Provide status updates between tasks
- ❌ Ask "Should I continue?"
- ❌ Summarize completed work mid-session
- ❌ Use `go test` in terminal (use `runTests` tool)
- ❌ Remove `t.Parallel()` to fix test failures
- ❌ Skip post-mortem ("too much work")
- ❌ Mark task complete with failing tests
- ❌ Ignore linting errors ("will fix later")

**ALWAYS do these:**

- ✅ Tool calls only (zero text between)
- ✅ Work continuously until 950k tokens
- ✅ Create post-mortem for EVERY task
- ✅ Fix failing tests before moving on
- ✅ Use `runTests` tool exclusively
- ✅ Enable `t.Parallel()` always
- ✅ Commit frequently (atomic units)

---

## 🎯 Quality Gates (Every Task)

**Before marking task complete:**

### Code Quality

- [ ] Zero compilation errors
- [ ] Zero linting errors: `golangci-lint run`
- [ ] No hardcoded values (use magic*.go)
- [ ] Errors wrapped with context
- [ ] No TODO comments

### Testing

- [ ] All tests pass: `runTests`
- [ ] Coverage meets threshold
- [ ] Table-driven test pattern
- [ ] `t.Parallel()` enabled
- [ ] Happy + sad paths covered

### Documentation

- [ ] Godoc comments on exports
- [ ] README updated if needed
- [ ] OpenAPI updated if API changes
- [ ] Migration guide if breaking
- [ ] Post-mortem corrective actions → new task docs OR subtasks

### Architecture

- [ ] Follows directory structure
- [ ] Respects domain boundaries
- [ ] Import aliases correct
- [ ] Design patterns used correctly

---

## 🔧 Tool Usage Rules

**File Operations:**

- ✅ `create_file`, `replace_string_in_file`, `multi_replace_string_in_file`
- ❌ NEVER use shell redirection (`>`, `>>`) for file creation

**Testing:**

- ✅ `runTests` tool exclusively
- ❌ NEVER `go test` in terminal (can hang)

**Git:**

- ✅ `git add`, `git commit`, `git status`, `git log`
- ❌ NEVER GitKraken MCP tools (mcp_gitkraken_*)

**Python (if needed):**

- ✅ `install_python_packages`, `configure_python_environment`
- ❌ NEVER manual `pip install` commands

**Directory:**

- ✅ `list_dir` tool
- ❌ NEVER `ls`, `dir`, `Get-ChildItem` commands

---

## 📊 Progress Tracking

**After each task:**

```text
manage_todo_list → mark task X complete
manage_todo_list → mark task X+1 in-progress
read_file → ##-<NEXT_TASK>.md
```

**Check token usage** (from `<system_warning>`):

```text
Token usage: X/1000000; Y remaining
```

**Continue working if**: Y remaining > 50,000 (5% buffer)

**Stop working if**: Y remaining ≤ 50,000 OR explicit user command

---

## 🐛 Error Handling

**Test failures:**

1. Analyze output
2. Fix implementation
3. Re-run tests
4. Document in post-mortem
5. NEVER skip or disable

**Linting failures:**

1. Run `golangci-lint run --fix`
2. Fix remaining manually
3. Re-run linter
4. Commit fixes

**Blocked task:**

1. Document blocker in post-mortem
2. Create new task for blocker
3. Continue with parallel tasks
4. Return after blocker resolved

---

## 📝 Commit Message Format

**Pattern**: `type(scope): description`

**Types**:

- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring
- `test`: Test additions/changes
- `docs`: Documentation updates
- `chore`: Maintenance tasks

**Examples**:

```bash
git commit -m "feat(authz): implement authorization code flow"
git commit -m "fix(authz): correct PKCE S256 hash algorithm"
git commit -m "test(authz): add RFC 7636 PKCE test vectors"
git commit -m "docs(authz): update OAuth 2.1 compliance checklist"
```

---

## 🎓 Post-Mortem Essentials

**EVERY task MUST have**: `##-<TASK>-POSTMORTEM.md`

**Minimum sections**:

1. **Implementation Summary**: What was done
2. **Issues Encountered**: Bugs, omissions, suboptimal patterns, test failures, instruction violations
3. **Corrective Actions**: Immediate (current task), deferred (future tasks), new task docs, pattern improvements
4. **Lessons Learned**: What went well, what needs improvement
5. **Metrics**: Time, coverage, quality, complexity

**Template location**: `feature-template.md` section "Post-Mortem and Corrective Actions"

---

## 🎯 Success Criteria Per Task

**Universal criteria** (ALL tasks):

- ✅ Code compiles
- ✅ Tests pass (≥85% coverage infra, ≥80% features)
- ✅ Linting clean
- ✅ No TODOs
- ✅ Documentation updated
- ✅ Post-mortem created
- ✅ Post-mortem corrective actions → new task docs OR subtasks

**Task-specific criteria**:

- See individual task docs (`##-<TASK>.md` "Acceptance Criteria" section)

---

## 🚀 Speed Optimization

**During iteration:**

- Use `git commit --no-verify` (skip pre-commit hooks)
- Use `runTests` with specific package: `runTests -files ./internal/specific/package`
- Batch related file operations: `multi_replace_string_in_file`

**Before task complete:**

- Run `golangci-lint run --fix`
- Run `runTests` on full package
- Verify coverage threshold

**Before push:**

- Run pre-commit hooks: `pre-commit run --all-files`
- Run full test suite: `go test ./... -cover`
- Verify CI will pass

---

## 📚 Reference Documents

**Must read before implementation:**

- `.github/copilot-instructions.md` - Primary instructions
- `.github/instructions/01-01.coding.instructions.md` - Coding patterns
- `.github/instructions/01-02.testing.instructions.md` - Testing patterns
- `.github/instructions/01-03.golang.instructions.md` - Go standards

**Reference during implementation:**

- `docs/feature-template/feature-template.md` - Full template
- `docs/feature-template/usage-example.md` - Example feature
- Master plan doc (e.g., `docs/02-identityV2/MASTER-PLAN.md`)
- Task docs (`##-<TASK>.md`)

---

## ⚡ Quick Start Workflow

**First 5 minutes of session:**

1. Read master plan: `read_file` → master plan doc
2. Read all task docs: `read_file` → all `##-<TASK>.md` files (parallel if possible)
3. Check current status: `manage_todo_list` → read
4. Identify next task: Look for first `IN_PROGRESS` or `PLANNED` status
5. Start immediately: `read_file` → next task doc → implement

**Critical Directive**: NO ASKING, NO PLANNING, JUST START

---

## 🎯 Remember

**You are an AUTONOMOUS agent**:

- Work continuously until all tasks complete OR 950k tokens used
- Don't stop for user approval (only when blocked)
- Don't ask questions with obvious answers (use best judgment)
- Don't provide status updates (just work)
- Don't celebrate task completion (just continue)

**Your job**: Execute the plan, fix issues, document learnings, keep going.

**Your output**: Completed tasks with post-mortems, working code, passing tests.

**NOT your job**: Stopping, summarizing, asking permission, waiting.

---

## 📞 When to Stop and Ask

**ONLY stop and ask when:**

1. Genuine blocker with no clear path forward (external dependency, spec ambiguity)
2. Token budget at 95% (950k/1M used)
3. Explicit user command to stop
4. Discovery of critical security issue requiring immediate attention

**NOT when:**

- Task is "hard" (figure it out, that's the job)
- Test is failing (fix it, don't ask)
- Not sure about implementation detail (make best judgment, document in post-mortem)
- Completed all planned tasks (check docs for more work)

---

## ✅ Session End Checklist

**When approaching 950k tokens OR all tasks complete:**

- [ ] Commit any pending changes
- [ ] Create final post-mortem if mid-task
- [ ] Update `manage_todo_list` with current status
- [ ] Push commits: `git push`
- [ ] **ONLY THEN** provide summary to user
