---
description: Documentation Synchronization Checklist - ensures consistency across all project documentation
tools:
  - read_file
  - grep_search
  - file_search
  - replace_string_in_file
  - multi_replace_string_in_file
---

# Documentation Synchronization Prompt

## Purpose

Systematically identify and synchronize related documentation across the cryptoutil project to prevent documentation sprawl and ensure consistency.

**When to Use**:
- After making changes to quality gates, standards, or patterns
- Before creating new documentation (check if existing docs need updates)
- When updating copilot instructions (propagate to related specs/plans/architecture)
- After completing major work phases (sync lessons to permanent homes)

**Anti-Pattern**: Creating new documentation without checking if existing docs need updates

---

## Documentation Hierarchy and Ownership

### Source of Truth Documents (Update These First)

1. **Copilot Instructions** (`.github/instructions/*.instructions.md`)
   - 26 instruction files covering: terminology, continuous work, SpecKit, architecture, security, testing, coding standards
   - **Purpose**: Agent behavior rules, quality gates, standards enforcement
   - **Update Triggers**: New patterns discovered, regression prevention, anti-pattern documentation
   - **Propagate To**: Specs, plans, prompts, READMEs, architecture docs

2. **Constitution** (`.specify/memory/constitution.md`)
   - **Purpose**: Project vision, quality gates, non-functional requirements, architectural constraints
   - **Update Triggers**: Quality gate changes, new NFRs, architectural decisions
   - **Propagate To**: Specs, plans, copilot instructions

3. **Architecture** (`docs/arch/ARCHITECTURE.md`)
   - **Purpose**: System design, patterns, component relationships
   - **Update Triggers**: Architectural changes, new patterns, component additions
   - **Propagate To**: Service template docs, copilot instructions (02-01.architecture)

### Derived Documents (Update After Source of Truth)

4. **Specs** (`specs/*/spec.md`, `specs/*/clarify.md`)
   - **Purpose**: Feature requirements, constraints, acceptance criteria
   - **Sources**: Constitution (quality gates), architecture (patterns)
   - **Update Triggers**: Requirement changes, clarifications, constitution updates

5. **Plans** (`specs/*/plan.md`, `docs/fixes-needed-*/plan.md`)
   - **Purpose**: Implementation phases, task breakdown, dependencies
   - **Sources**: Specs (requirements), constitution (quality gates), copilot instructions (standards)
   - **Update Triggers**: Spec changes, quality gate updates, new phase discoveries

6. **Tasks** (`specs/*/tasks.md`, `docs/fixes-needed-*/tasks.md`)
   - **Purpose**: Execution checklists, completion criteria, evidence tracking
   - **Sources**: Plans (phases/tasks), constitution (quality gates), copilot instructions (evidence requirements)
   - **Update Triggers**: Plan updates, quality gate changes, new task discoveries

### Implementation Documentation (Update During/After Implementation)

7. **DETAILED.md** (`specs/*/implement/DETAILED.md`)
   - **Purpose**: Append-only timeline, task checklist, session notes
   - **Sources**: Tasks (checkboxes), implementation (discoveries)
   - **Update Triggers**: Task completion, lessons learned, constraint discoveries

8. **EXECUTIVE.md** (`specs/*/implement/EXECUTIVE.md`)
   - **Purpose**: Stakeholder overview, progress metrics, blockers, risks
   - **Sources**: DETAILED.md (progress), tasks (completion %), quality metrics
   - **Update Triggers**: Phase completion, blocker discovery, risk identification

### Reference Documentation (Update When Patterns Stabilize)

9. **Service Template Docs** (`docs/service-template/*.md`)
   - **Purpose**: Reusable patterns, examples, guidance
   - **Sources**: Architecture (patterns), copilot instructions (standards), implementation (validated patterns)
   - **Update Triggers**: Pattern validation, anti-pattern discovery, template refactoring

10. **READMEs** (`README.md`, `docs/README.md`, `specs/*/README.md`)
    - **Purpose**: Getting started, quick reference, navigation
    - **Sources**: All above (high-level summaries)
    - **Update Triggers**: Major project changes, new workflows, documentation reorganization

---

## Cross-Cutting Concerns Requiring Sync

### Quality Gates
**Source**: Constitution → Copilot Instructions (01-02, 03-02, 06-01)
**Propagate To**:
- [ ] Specs (acceptance criteria sections)
- [ ] Plans (completion criteria for phases)
- [ ] Tasks (evidence requirements per task)
- [ ] DETAILED.md (validation checklists)

**Current Values**:
- Coverage: ≥95% production, ≥98% infrastructure/utility
- Mutation: ≥85% production (early), ≥98% production (later), ≥98% infrastructure
- Build: `go build ./...` clean (0 errors)
- Linting: `golangci-lint run` clean (0 warnings)
- Tests: `go test ./...` (100% pass, 0 skips without tracking)
- Timing: <15s per package unit tests, <180s full unit suite, <240s E2E tests

### Execution Rules
**Source**: Copilot Instructions (01-02.continuous-work) → Prompts (autonomous-execution, beast-mode-3.1)
**Propagate To**:
- [ ] Plans (execution workflow sections)
- [ ] Tasks (completion criteria patterns)
- [ ] Other prompts (workflow-fixing, plan-tasks-quizme)

**Current Rules**:
- NEVER stop until user clicks STOP button
- NEVER ask "Should I proceed?" or "Shall I continue?"
- ALWAYS commit after each task completion
- ALWAYS use incremental commits (NOT amend)
- ALWAYS provide evidence before marking tasks complete

### Architectural Patterns
**Source**: Architecture → Copilot Instructions (02-01, 02-02, 02-03) → Service Template Docs
**Propagate To**:
- [ ] Specs (architectural constraints sections)
- [ ] Plans (implementation approach sections)
- [ ] Service Template docs (examples, guides)

**Current Patterns**:
- Dual HTTPS endpoints (public + admin)
- Two request paths (/service/** vs /browser/**)
- Multi-tenancy (tenant_id for data, realm_id for auth only)
- ServerBuilder pattern (eliminates 260+ lines boilerplate)
- Merged migrations (template 1001-1004, domain 2001+)

### Security Standards
**Source**: Copilot Instructions (02-07, 02-08, 02-09, 03-06) → Constitution
**Propagate To**:
- [ ] Specs (security requirements sections)
- [ ] Plans (security validation tasks)
- [ ] Architecture (security patterns)

**Current Standards**:
- FIPS 140-3 ALWAYS enabled (NEVER disable)
- PBKDF2-HMAC-SHA256 for password hashing (600k iterations)
- crypto/rand ALWAYS (NEVER math/rand)
- TLS 1.3+ with full cert chain validation
- Docker secrets (NEVER environment variables)

### Testing Standards
**Source**: Copilot Instructions (03-02.testing) → Constitution
**Propagate To**:
- [ ] Plans (test coverage tasks)
- [ ] Tasks (test evidence requirements)
- [ ] Specs (test acceptance criteria)

**Current Standards**:
- TestMain for heavyweight dependencies (PostgreSQL, services)
- Table-driven tests with t.Parallel()
- Port 0 for dynamic allocation (tests)
- SQLite in-memory for fast tests
- Race detector with count=2+ for probabilistic execution

---

## Synchronization Workflow

### Step 1: Identify Change Scope

**Question**: What type of change occurred?
- [ ] Quality gate update (coverage/mutation threshold change)
- [ ] Execution rule change (continuous work, autonomous execution)
- [ ] Architectural pattern change (new service template requirement)
- [ ] Security standard change (cryptography, authentication)
- [ ] Testing standard change (coverage, mutation, test patterns)
- [ ] Documentation pattern change (SpecKit methodology update)

**Action**: Read relevant source of truth docs listed above

---

### Step 2: Find Impacted Documents

**For Quality Gate Changes**:
```bash
# Find all references to coverage thresholds
grep -r "≥95\|≥98\|≥85" docs/ specs/ .github/instructions/

# Find all references to mutation scores
grep -r "mutation.*85\|mutation.*98\|gremlins" docs/ specs/ .github/instructions/

# Find all references to timing targets
grep -r "<15s\|<180s\|<240s" docs/ specs/ .github/instructions/
```

**For Execution Rule Changes**:
```bash
# Find all references to continuous execution
grep -r "NEVER stop\|NEVER ask\|Should I proceed\|autonomous" .github/prompts/ docs/fixes-needed-*/

# Find all references to commit patterns
grep -r "commit.*task\|incremental commit\|NEVER amend" docs/fixes-needed-*/ specs/*/
```

**For Architectural Pattern Changes**:
```bash
# Find all references to dual HTTPS
grep -r "dual.*HTTPS\|public.*admin.*server\|9090" docs/ .github/instructions/

# Find all references to multi-tenancy
grep -r "tenant_id\|realm_id\|multi-tenancy" docs/ .github/instructions/

# Find all references to ServerBuilder
grep -r "ServerBuilder\|merged.*migration\|260\+ lines" docs/ .github/instructions/
```

---

### Step 3: Verify Document Relationships

**Checklist**:
- [ ] Read source of truth document (copilot instructions, constitution, architecture)
- [ ] Identify specific changed section (quality gates, execution rules, patterns)
- [ ] Search for dependent docs referencing changed content (use grep patterns above)
- [ ] Read dependent docs to understand current state and required updates
- [ ] Create list of updates required with file paths and line numbers

---

### Step 4: Propagate Changes

**Pattern**: Update from source outward (constitution → specs → plans → tasks → implementation docs)

**For Each Impacted Document**:
1. Read current content around change location
2. Determine if update required (value changed, new rule, deprecated pattern)
3. Use `replace_string_in_file` for single change
4. Use `multi_replace_string_in_file` for batch updates (≤10 similar changes)
5. Verify update preserves document context and formatting

**Quality Gates**:
- [ ] Changed values match source of truth exactly
- [ ] Document-specific context preserved (don't copy verbatim, adapt to doc purpose)
- [ ] Cross-references updated (if referencing specific instruction file sections)
- [ ] No orphaned references (old values, deprecated patterns)

---

### Step 5: Validate Consistency

**Cross-Document Validation**:
```bash
# Verify all quality gates use same values
grep -r "≥95.*production\|≥98.*infrastructure" docs/ specs/ .github/instructions/ | sort | uniq -c

# Verify all mutation scores consistent
grep -r "≥85.*mutation\|≥98.*mutation" docs/ specs/ .github/instructions/ | sort | uniq -c

# Verify execution rules consistent
grep -r "NEVER stop until.*STOP\|NEVER ask.*proceed" .github/prompts/ .github/instructions/ | sort | uniq -c
```

**Checklist**:
- [ ] All documents use same quality gate values
- [ ] All documents reference same execution rules
- [ ] All documents describe same architectural patterns
- [ ] No conflicting guidance across documents

---

### Step 6: Commit with Audit Trail

**Conventional Commit Format**:
```
docs(sync): synchronize [change type] across documentation

Updated documents:
- .github/instructions/XX-YY.*.instructions.md: [specific change]
- specs/002-cryptoutil/spec.md: [specific change]
- docs/fixes-needed-plan-tasks/plan.md: [specific change]

Changes:
- Quality gates: Coverage ≥95%/98%, mutation ≥85%/98%
- Execution rules: NEVER stop without STOP button
- Architectural patterns: Dual HTTPS, multi-tenancy

Verification:
- grep search shows consistent values across all docs
- No conflicting guidance found
- All dependent docs updated

Related: docs/fixes-needed-plan-tasks/tasks.md (P#.#)
```

---

## Common Sync Scenarios

### Scenario 1: Updated Quality Gates

**Trigger**: Changed coverage target from ≥90% to ≥95%

**Documents Requiring Sync**:
1. Constitution → Update quality gates section
2. Copilot Instructions (06-01.evidence-based) → Update coverage targets
3. Specs (002-cryptoutil/spec.md) → Update acceptance criteria
4. Plans (fixes-needed-PLAN.md) → Update phase completion criteria
5. Tasks (fixes-needed-TASKS.md) → Update evidence requirements

**Sync Workflow**:
```bash
# Step 1: Update constitution
# Edit .specify/memory/constitution.md quality gates section

# Step 2: Find all references to old value
grep -r "≥90.*coverage" docs/ specs/ .github/instructions/

# Step 3: Use multi_replace_string_in_file for batch update
# Replace all instances of "≥90% coverage" with "≥95% coverage"

# Step 4: Verify consistency
grep -r "≥95.*coverage" docs/ specs/ .github/instructions/ | wc -l
grep -r "≥90.*coverage" docs/ specs/ .github/instructions/ | wc -l  # Should be 0

# Step 5: Commit with audit trail
```

---

### Scenario 2: New Execution Rule Discovered

**Trigger**: Discovered regression where agent asked "Should I proceed?" despite instructions

**Documents Requiring Sync**:
1. Copilot Instructions (01-02.continuous-work) → Add "NEVER ask 'Should I proceed?'" to prohibited patterns
2. Prompts (autonomous-execution.prompt.md) → Add to continuous execution rules
3. Prompts (beast-mode-3.1.prompt.md) → Add to autonomous behavior section
4. Plans (fixes-needed-PLAN.md) → Update execution workflow section

**Sync Workflow**:
```bash
# Step 1: Add rule to copilot instructions
# Edit .github/instructions/01-02.continuous-work.instructions.md

# Step 2: Find all execution rule sections
grep -rn "Execution.*Rule\|Continuous.*Execution\|NEVER ask" .github/prompts/

# Step 3: Update each prompt file
# Use replace_string_in_file for each prompt

# Step 4: Verify new rule present everywhere
grep -r "NEVER ask.*Should I proceed" .github/ | wc -l  # Should match number of prompt files + copilot instructions

# Step 5: Commit with audit trail
```

---

### Scenario 3: Architectural Pattern Change

**Trigger**: ServerBuilder pattern now requires merged migrations (template 1001-1004 + domain 2001+)

**Documents Requiring Sync**:
1. Architecture (docs/arch/ARCHITECTURE.md) → Update ServerBuilder section
2. Copilot Instructions (02-02.service-template) → Update ServerBuilder requirements
3. Copilot Instructions (03-08.server-builder) → Add merged migrations pattern
4. Service Template Docs (docs/service-template/*.md) → Update examples
5. Specs → Update architectural constraints

**Sync Workflow**:
```bash
# Step 1: Update architecture doc
# Edit docs/arch/ARCHITECTURE.md ServerBuilder section

# Step 2: Find all ServerBuilder references
grep -r "ServerBuilder\|merged.*migration" docs/ .github/instructions/ specs/

# Step 3: Update copilot instructions
# Use multi_replace_string_in_file for batch updates

# Step 4: Update service template docs with examples
# Use replace_string_in_file for each example

# Step 5: Verify consistency
grep -r "merged.*migration.*1001.*2001" docs/ .github/instructions/ | wc -l

# Step 6: Commit with audit trail
```

---

## Documentation Sprawl Prevention

### Anti-Patterns to Avoid

❌ **Creating Summary Docs** (COMPLETION-SUMMARY.md, ANALYSIS.md)
- Violation: Duplicates data from authoritative sources (tasks.md, DETAILED.md)
- Prevention: Append to DETAILED.md Section 2 timeline instead

❌ **Creating Backup Docs** (plan_backup.md, tasks_backup.md)
- Violation: Git history already provides backup and rollback
- Prevention: Use `git checkout <hash> -- file` to restore from history

❌ **Creating Verbose Analysis Docs** (SESSION-*.md, COMPLETION-ANALYSIS.md)
- Violation: User never reviews these, they accumulate as documentation debt
- Prevention: Append to DETAILED.md Section 2 with concise findings

❌ **Creating Specialized Gap Analysis Docs** (PHASE-2.4-COVERAGE-GAPS.md)
- Violation: Gaps should be documented in main plan Phase X section
- Prevention: Update plan.md Phase X with specific gap details

### Lean Documentation Rules

✅ **ALWAYS prefer updating existing docs over creating new ones**
✅ **ALWAYS check if information exists elsewhere before creating new doc**
✅ **ALWAYS append to DETAILED.md Section 2 for session-specific work**
✅ **ALWAYS update plan.md/tasks.md in-place (no backups)**
✅ **ALWAYS use git history for rollback (not backup files)**

---

## Verification Checklist

Before ending documentation sync session:

- [ ] All source of truth documents updated (copilot instructions, constitution, architecture)
- [ ] All derived documents synchronized (specs, plans, tasks)
- [ ] All implementation docs updated (DETAILED.md, EXECUTIVE.md)
- [ ] All reference docs current (service template, READMEs)
- [ ] grep searches show consistent values across all docs
- [ ] No conflicting guidance found
- [ ] Git commit created with comprehensive audit trail
- [ ] No new documentation sprawl (summaries, backups, verbose analyses)

---

## Quick Reference: Document Update Order

**When Changing Quality Gates**:
1. Constitution → Copilot Instructions (06-01) → Specs → Plans → Tasks → DETAILED.md

**When Changing Execution Rules**:
1. Copilot Instructions (01-02) → Prompts (all) → Plans → Tasks

**When Changing Architectural Patterns**:
1. Architecture → Copilot Instructions (02-*) → Service Template → Specs → Plans

**When Changing Security Standards**:
1. Copilot Instructions (02-07, 02-08, 02-09, 03-06) → Constitution → Specs → Plans

**When Discovering Lessons Learned**:
1. DETAILED.md Section 2 → Copilot Instructions (anti-patterns) → READMEs (if applicable)
