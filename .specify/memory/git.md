# Git and Documentation Standards - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/05-02.git.instructions.md`

## Git Workflow for Copilot Chat Sessions

### Terminal Commands

**ALWAYS use terminal git commands** (NOT Copilot chat commands):
- `git status` - Check working tree state
- `git add -A` - Stage all changes
- `git commit -m "message"` - Commit with message
- `git push origin main` - Push to remote
- `git log --oneline` - View commit history
- `git diff` - View uncommitted changes

### Commit vs Push Strategy

**Commit Frequently, Push Strategically**:
- **Commit**: For logical units of work (atomic changes)
- **Push**: When ready for CI/CD validation and peer review
- **Local iteration**: Acceptable to use `git commit --no-verify` for quick iteration while actively changing code
- **Push readiness**: ALWAYS run `go test`, `golangci-lint`, and required pre-commit checks before pushing or opening PR

**Pre-push Hooks**:
- Run automatically before push to enforce quality gates
- Check: Dependency updates, linting, formatting
- Failure: Fix issues before force-pushing

**Dependency Management**:
- Update dependencies incrementally
- Validate with tests between updates
- Document breaking changes in commit messages

### Branch Strategy

**Feature Branches**:
- Use feature branches for development work
- Merge to main only after CI passes
- Keep branches short-lived (days, not weeks)
- Rebase on main before merging

**Commit Hygiene**:
- Keep commits atomic and focused (one logical change per commit)
- Use conventional commit messages (see below)
- Avoid "WIP" commits in main branch

---

## Conventional Commits - MANDATORY

**Format**: `<type>[optional scope]: <description>`

### Types

- **feat**: New feature for users
- **fix**: Bug fix for users
- **docs**: Documentation changes
- **style**: Code formatting, whitespace, missing semicolons (NOT CSS)
- **refactor**: Code change that neither fixes bug nor adds feature
- **perf**: Performance improvement
- **test**: Adding or correcting tests
- **build**: Changes to build system or dependencies
- **ci**: CI/CD configuration changes
- **chore**: Maintenance tasks (dependency updates, etc.)
- **revert**: Reverts a previous commit

### Rules

- **Imperative mood**: "add feature" NOT "added feature"
- **Lowercase**: All lowercase for type and description
- **No period**: Don't end subject with period
- **Subject length**: Keep under 72 characters
- **Breaking changes**: Add `!` after type OR use `BREAKING CHANGE:` in body

### Examples

```bash
# Feature
git commit -m "feat(auth): add OAuth2 client credentials flow"

# Bug fix
git commit -m "fix(database): prevent connection pool exhaustion"

# Breaking change
git commit -m "feat(api)!: remove deprecated v1 endpoints"

# Scoped documentation
git commit -m "docs(readme): add Docker Compose setup instructions"

# Refactoring
git commit -m "refactor(crypto): extract key derivation to separate package"
```

### Body and Footer (Optional)

```bash
git commit -m "feat(api): add pagination to list endpoints

Add page and size query parameters to all list endpoints.
Default page size is 50, maximum is 1000.

BREAKING CHANGE: List endpoints now return paginated results.
Clients must handle pagination or specify large page size.

Fixes #123"
```

---

## Commit Strategy - Incremental vs Amend - CRITICAL

### ALWAYS Commit Incrementally (NOT Amend)

**Why Incremental Commits Matter**:
- **Preserves full timeline** of changes and decisions
- **Enables git bisect** to identify when bugs were introduced
- **Allows selective revert** of specific fixes
- **Shows thought process** and iterative improvement
- **Easier code review** - each logical change independently reviewable

### NEVER Use Amend Repeatedly

**WRONG Pattern**:

```bash
# ❌ BAD: Amend repeatedly (loses history, masks mistakes, hard to bisect)
git commit -m "fix"
# Run tests, find another issue
git add more_fixes
git commit --amend
# Run tests, find yet another issue
git add even_more_fixes
git commit --amend  # Original fix context completely lost!
```

**Result**: Single squashed commit with no context, impossible to bisect, unclear what each fix addressed

### ALWAYS Commit Each Logical Unit

**CORRECT Pattern**:

```bash
# ✅ GOOD: Commit each logical unit independently
git commit -m "fix(format_go): restore clean baseline from 07192eac"
# Run tests, verify baseline works

git commit -m "fix(format_go): add defensive check with filepath.Abs()"
# Run tests, verify defensive check works

git commit -m "test(format_go): verify self_modification_test catches regressions"
# Clear progression, easy to bisect, reviewable history
```

**Result**: 3 focused commits with clear context, easy to bisect if regression occurs, shows iterative problem-solving

### When to Use Amend (Rare Cases)

**ONLY acceptable uses**:
- Fixing typos in commit message IMMEDIATELY after commit (within 1 minute, before push)
- Adding forgotten files to most recent commit (within 1 minute, before push)

**NEVER amend**:
- After pushing (breaks shared history)
- Repeatedly during debugging session
- To hide incremental fixes
- As default workflow pattern

---

## Restore from Clean Baseline Pattern - CRITICAL

### When to Use This Pattern

**ALWAYS restore from clean baseline FIRST when**:
- Fixing regressions or corrupted code
- Multiple failed fix attempts have occurred
- Current HEAD state is uncertain
- Code worked previously but broken now

### Why This Matters

**Problem**: HEAD may be corrupted by previous failed attempts
- Incremental fixes on corrupted base compound the problem
- "Fixing" code that's already broken creates confusion
- Lost track of what changed between working and broken state

**Solution**: Start from known-good state, apply targeted fix

### Step-by-Step Pattern

**1. Find Last Known-Good Commit**:

```bash
# Search for baseline commits
git log --oneline --grep="baseline" | head -5

# Or find specific working commit
git log --oneline --all | grep "test"
git log --oneline --since="2 weeks ago"

# Use git bisect if needed
git bisect start
git bisect bad HEAD
git bisect good <known-good-commit>
```

**2. Restore Entire Package from Clean Commit**:

```bash
# Restore all files in package from clean commit
git checkout <clean-commit-hash> -- path/to/package/

# Example: Restore format_go package
git checkout 07192eac -- internal/cmd/cicd/format_go/
```

**3. Verify Baseline Works**:

```bash
# Run tests for restored package
go test ./path/to/package/

# Check git status (should show only restored files)
git status
```

**4. Apply ONLY the New Fix**:

```bash
# Edit specific file with targeted change
# Example: Add defensive check for absolute paths
vim internal/cmd/cicd/format_go/enforce_any.go
# Add filepath.Abs() check
```

**5. Verify Fix Works Independently**:

```bash
# Run tests again
go test ./path/to/package/

# Verify fix addresses specific issue
```

**6. Commit as NEW Commit** (NOT amend):

```bash
git commit -m "fix(package): add defensive check for X"
```

### Common Mistakes to Avoid

**❌ Assuming HEAD is Correct**:
- HEAD may be corrupted from previous attempts
- Always verify baseline works before applying new fixes

**❌ Applying "One More Fix" on Corrupted Code**:
- Each fix compounds the problem
- Restore clean baseline first

**❌ Mixing Baseline Restoration with New Fixes**:
- Separate commits for restoration vs new fixes
- Makes history clear and reviewable

**❌ Using Amend Instead of New Commits**:
- Loses evidence of restoration process
- Harder to track what was fixed vs restored

---

## Terminal Command Auto-Approval

### Pattern Checking Workflow

**When executing terminal commands through Copilot**:

1. **Check Pattern Match**: Verify if command matches `chat.tools.terminal.autoApprove` patterns in `.vscode/settings.json`
2. **Track Unmatched**: Maintain list of unmatched commands during session
3. **End-of-Session Review**: Ask user if they'd like to add new auto-approve patterns
4. **Pattern Recommendations**:
   - **Auto-Enable (true)**: Safe, informational, build commands
   - **Auto-Disable (false)**: Destructive, dangerous, system-altering commands

### Auto-Enable Candidates

**Safe read-only and build operations**:
- Read-only operations: `status`, `list`, `inspect`, `logs`, `history`
- Build and test: `build`, `test`, `format`, `lint`
- Informational: `version`, `info`, `df`
- Development workflow: `fetch`, `status`, `diff`

### Auto-Disable Candidates

**Potentially dangerous operations**:
- Destructive: `rm`, `delete`, `prune`, `reset`, `kill`
- Network: `push`, `pull` from remotes
- System modifications: `install`, `update`, edit configurations
- File system changes: create, update, delete files/directories
- Container execution: `exec`, `run` interactive containers

### Pattern Format

**Regex patterns in .vscode/settings.json**:

```json
{
  "chat.tools.terminal.autoApprove": {
    "/^git (status|log|diff)/": true,
    "/^go (test|build|fmt)/": true,
    "/^docker (ps|images|inspect)/": true,
    "/^git (push|reset --hard)/": false
  }
}
```

**Rules**:
- Use established regex with `^` anchor
- Group related subcommands with alternation `(cmd1|cmd2|cmd3)`
- Include comments explaining security rationale

---

## PowerShell Notes

### Command Chaining

**Use semicolon (`;`) to chain commands** (NOT `&&`):

```powershell
# ✅ CORRECT
git add -A; git commit -m "message"; git status

# ❌ WRONG (bash syntax doesn't work in PowerShell)
git add -A && git commit -m "message" && git status
```

### Unix Utilities Not Available

**PowerShell does not include Unix utilities by default**:
- ❌ `sed` - NOT available
- ❌ `awk` - NOT available
- ❌ `grep` - Use `Select-String` instead

**Prefer Git built-in capabilities**:
- ✅ `git diff -- path/to/file` - View file diffs
- ✅ `git show <commit> -- path/to/file` - View file at specific commit
- ✅ `git grep 'pattern'` - Search repository content
- ✅ `Get-Content file | Select-String 'pattern'` - Simple grep-like search

**Cross-platform diffs**:
- Use `git diff -- <file>` over `sed` pipelines
- Works consistently across environments
- Yields consistent results for reviewers and CI

---

## Pull Request Descriptions

### Title

**Format**: `type(scope): description`
- Use conventional commit format
- Keep under 72 characters
- Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore

### Sections

**What**:
- Clear description of what PR does (present tense)
- Focus on user/system impact

**Why**:
- Business/technical rationale
- Impact on users or system
- Problem being solved

**How**:
- High-level implementation approach
- Key technical decisions
- Architecture changes

**Testing**:
- How tested (unit, integration, E2E)
- Coverage impact (before/after percentages)
- Manual testing steps (if applicable)

**Breaking Changes** (if applicable):
- Migration guidance
- API changes with examples
- Deprecation timeline

**Documentation**:
- README changes
- Migration guides
- API documentation updates

### Code Review Checklist

**Security**:
- No sensitive data exposure (credentials, keys, tokens)
- Proper input validation
- Secure defaults (TLS 1.3, FIPS algorithms)

**Quality**:
- Tests added/updated (coverage ≥95% production, ≥98% infrastructure)
- Linting passes (`golangci-lint run --fix`)
- Documentation updated

**Performance**:
- No regressions (benchmark comparison)
- Memory leaks addressed
- Database query optimization (N+1 queries)

**Operations**:
- Logging appropriate (structured logging, no PII)
- Monitoring/metrics added
- Deployment impact documented

### PR Size Guidelines

| Size | Lines Changed | Risk | Strategy |
|------|--------------|------|----------|
| Small | <200 | Low | Single focused change, fast review |
| Medium | 200-500 | Moderate | Multiple related changes, thorough review |
| Large | 500+ | High | Complex feature, consider splitting |
| Epic | 1000+ | Very High | Break down into smaller, independently deployable PRs |

**Large PR Mitigation**:
- Split into smaller PRs with feature flags
- Use stacked PRs (PR 1 merged before PR 2 opened)
- Provide detailed testing plan and migration guide

---

## Session Documentation Strategy - CRITICAL

### NEVER Create Standalone Session Documentation

**MANDATORY: ALWAYS append session work to `specs/001-cryptoutil/implement/DETAILED.md` Section 2 timeline**

### Append-Only Timeline Pattern (Required)

**Format**:

```markdown
### YYYY-MM-DD: Brief Session Title
- Work completed: Summary of tasks (commit hashes)
- Key findings: Important discoveries or blockers
- Coverage/quality metrics: Before/after numbers
- Violations found: Any issues discovered
- Next steps: Outstanding work or follow-up needed
- Related commits: [abc1234] description
```

### Example Correct Append

```markdown
### 2025-12-14: Jose Coverage Improvement Attempt
- Added 60+ comprehensive tests (commit 81e3260d) - all passing
- Coverage remained 84.2% (no improvement) - tests duplicated existing paths
- Identified real gaps: unused functions (23%), Is*/Extract* defaults (83-86%)
- Lessons: MUST analyze baseline coverage HTML BEFORE writing tests
- Violations found: individual test functions, 1371-line file, standalone doc
- Updated copilot instructions with mandatory patterns (commit abc1234)
```

### Violations to Avoid

**NEVER create files like**:
- ❌ `docs/SESSION-2025-12-14-JOSE-COVERAGE.md` (standalone session doc)
- ❌ `docs/session-*.md` (any dated session documentation)
- ❌ `docs/analysis-*.md` (standalone analysis documents)
- ❌ `docs/work-log-*.md` (separate work logs)

**Why This Matters**:
- Prevents documentation bloat (dozens of orphaned session files)
- Single source of truth for implementation timeline
- Easier to search and review work history
- Maintains chronological narrative flow
- Reduces maintenance burden

### When to Create New Documentation

**ONLY create new docs for**:
- Permanent feature specifications (`specs/*/README.md`, `TASKS.md`)
- Reference guides that users need (`docs/DEMO-GUIDE.md`, `docs/DEV-SETUP.md`)
- Post-mortem analysis requiring deep dive (`docs/P0.X-*.md`)
- Architecture Decision Records (ADRs)

**Rule of Thumb**: If it's session-specific work → append to DETAILED.md. If it's permanent reference material → create dedicated doc.

---

## Core Documentation Organization

**Keep docs in 2 main files**:
- `README.md` - Main project documentation (setup, usage, architecture overview)
- `docs/README.md` - Deep dive documentation (detailed architecture, advanced topics)

### Adding to Existing README

**ALWAYS add to existing README.md** instead of creating new markdown files:
- New feature? Add section to README.md
- New script? Document in README.md Scripts section
- New tool? Add to README.md Tools section

**NEVER create separate documentation files** for scripts or tools:
- ❌ `docs/SCRIPT-NAME.md` - Document in README.md instead
- ❌ `docs/TOOL-NAME.md` - Document in README.md instead

---

## TODO Documentation Organization

### Critical Requirements

**Keep 1-6 `./docs/todos-*.md` files**:
- Track missing or incomplete items before ending sessions
- Review before ending sessions, remove completed items
- **Delete completed tasks immediately** - don't mark as done, remove from file
- Always ensure files contain ONLY active, actionable tasks

### Implementation Guidelines

**Large cleanups**:
- Use `create_file` to rewrite entire file with only active tasks
- Don't use `replace_string_in_file` for large deletions

**Failed replacements**:
- Create clean version in new file
- Replace original with `create_file`

**Avoid complex replace_string_in_file**:
- Large text blocks often fail due to whitespace mismatches
- Simpler to rewrite entire section with `create_file`
