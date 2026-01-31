---
name: doc-sync
description: Synchronize documentation across project - prevent sprawl and ensure consistency
tools:
  - edit/editFiles
  - execute/runInTerminal
  - execute/getTerminalOutput
  - read/problems
  - search/codebase
  - search/usages
  - search/changes
model: claude-sonnet-4
argument-hint: "[source-of-truth-file]"
---

# Documentation Synchronization Prompt

## Documentation Reference Table

Quick reference for all synchronization-eligible documentation across the cryptoutil project.

| Document Path | Type | Purpose | Update Triggers | Propagates To |
|---------------|------|---------|----------------|---------------|
| **Source of Truth (NEVER delete, ALWAYS update)** |
| `.github/copilot-instructions.md` | Entrypoint | Main agent instruction loader | New instruction files, instruction reorganization | All agents, instruction files |
| `.github/constitution.md` | Constitutional | Overarching project principles | Fundamental principle changes, new constitutional rules | All docs, all code |
| `.github/instructions/*.instructions.md` | Constitutional | Domain-specific detailed rules | Pattern discovery, anti-pattern identification, new best practices | READMEs, agent prompts, templates |
| `docs/arch/ARCHITECTURE.md` | Architectural | System design, service patterns, quality gates | Architectural changes, pattern discovery, quality gate adjustments | Service templates, agent prompts, copilot instructions |
| **Spec Kit (Update frequently, specialized content)** |
| `docs/speckit/constitution.md` | Constitutional | Spec Kit principles and rules | Constitution changes, Spec Kit workflow updates | Spec Kit templates, workflows |
| `docs/speckit/templates/*.md` | Templates | Reusable Spec Kit document templates | Template improvements, new document types | New spec documents |
| `specs/*/spec.md` | Specifications | Product/feature specifications | Requirement changes, scope changes | Implementation plans, tasks |
| `specs/*/clarify.md` | Clarifications | Q&A for ambiguities | Unknowns discovered during planning, user answers | Plans, tasks |
| `specs/*/plan.md` | Plans | High-level implementation plans | Scope changes, phase adjustments, new blockers discovered | Tasks, DETAILED |
| `specs/*/tasks.md` | Tasks | Detailed actionable checklists | Task completion, blocker discovery, mid-execution phase creation | Plan (checkbox sync), DETAILED |
| `specs/*/analyze/*.md` | Analysis | Deep-dive investigations | Complex technical analysis needed | Plans, tasks, DETAILED |
| `specs/*/DETAILED.md` | Timeline | Chronological work log (Section 2 append-only) | Every work session, every task completion, every blocker | EXECUTIVE (summaries) |
| `specs/*/EXECUTIVE.md` | Executive | High-level summaries for stakeholders | Major milestones, phase completions | External reporting |
| **Custom Plans (Ephemeral, delete after completion)** |
| `docs/fixes-needed-plan-tasks-v*/plan.md` | Plans | Custom fix campaign plans | Mid-execution phase discovery, blocker resolution | tasks.md (checkbox sync), completed.md |
| `docs/fixes-needed-plan-tasks-v*/tasks.md` | Tasks | Custom fix campaign task lists | Task completion, checkbox updates | plan.md (checkbox sync), completed.md |
| `docs/fixes-needed-plan-tasks-v*/completed.md` | Archive | Completed tasks with evidence | Task completion with evidence | Lessons learned extraction |
| **Reference (Update as needed, user-facing)** |
| `README.md` | Reference | Project overview, quick start, core concepts | Major feature additions, architecture changes, setup process changes | DEV-SETUP, other READMEs |
| `docs/README.md` | Reference | Developer deep-dive, architecture details | Architecture changes, workflow changes | Project README, DEV-SETUP |
| `docs/DEV-SETUP.md` | Reference | Development environment setup | Tool version updates, new dependencies, setup process changes | README, workflow docs |
| `docs/arch/SERVICE-TEMPLATE-*.md` | Architectural | Service template patterns | Template pattern changes, new template features | Service implementations, copilot instructions |
| `docs/pre-commit-hooks.md` | Reference | Pre-commit hook documentation | Hook configuration changes, new hooks, formatter updates | .pre-commit-config.yaml, copilot instructions |
| `.github/agents/*.agent.md` | Agent Prompts | Custom agent workflows | Workflow improvements, pattern discovery, anti-pattern identification | Copilot instructions, constitution |
| **Evidence (Temporary, test-output/ only, NEVER commit)** |
| `test-output/<analysis-type>/*.md` | Analysis | Session-specific analysis artifacts | Every analysis session | Lessons learned, copilot instructions (as needed) |
| `test-output/<analysis-type>/*.{cov,html,log,txt}` | Artifacts | Coverage, mutation, benchmark results | Every test/analysis run | Analysis documents, completion evidence |

### Document Type Definitions

- **Entrypoint**: First file read by agents, loads all other instructions
- **Constitutional**: Fundamental rules that apply to all work
- **Architectural**: System design and service patterns
- **Templates**: Reusable document structures
- **Specifications**: Detailed product/feature requirements
- **Clarifications**: Q&A for ambiguities (unknowns/risks/inefficiencies)
- **Plans**: High-level implementation strategies
- **Tasks**: Detailed actionable checklists with acceptance criteria
- **Analysis**: Deep-dive technical investigations
- **Timeline**: Chronological work log (DETAILED.md Section 2, append-only)
- **Executive**: High-level summaries for stakeholders
- **Archive**: Completed work with evidence
- **Reference**: User-facing guides and documentation
- **Agent Prompts**: Custom workflow automation
- **Evidence**: Temporary analysis artifacts (NEVER commit)

### Quick Reference: Common Update Scenarios

| Change Type | Start Here | Then Update |
|-------------|-----------|-------------|
| New quality gate | `docs/arch/ARCHITECTURE.md` | `.github/instructions/06-01.quality-gates.md`, `docs/speckit/constitution.md`, `.github/agents/plan-tasks-implement.agent.md` |
| New execution rule | `.github/constitution.md` | `.github/instructions/01-02.beast-mode.md`, `.github/agents/plan-tasks-implement.agent.md`, `docs/speckit/constitution.md` |
| New architectural pattern | `docs/arch/ARCHITECTURE.md` | `.github/instructions/02-*.md` (relevant domain), `docs/arch/SERVICE-TEMPLATE-*.md`, service implementations |
| New testing pattern | `.github/instructions/03-02.testing.md` | `docs/arch/SERVICE-TEMPLATE-TESTING.md`, `README.md`, `.github/agents/*.agent.md` (if workflow change) |
| New pre-commit hook | `.pre-commit-config.yaml` | `docs/pre-commit-hooks.md`, `.github/instructions/03-07.linting.md` |
| Lessons learned discovered | `test-output/<session>/lessons-learned.md` | Copilot instructions (anti-patterns), README (best practices), agent prompts (validation steps) |

---

## Purpose

Systematically identify and synchronize related documentation across the cryptoutil project to prevent documentation sprawl and ensure consistency.

**When to Use**:

- When updating any source of truth document (copilot instructions, constitution, architecture)
- Before creating new documentation (check if existing docs need updates first)
- After discovering new patterns, anti-patterns, or lessons learned

**Anti-Pattern**: Creating new documentation without checking if existing docs need updates

---

## Documentation Hierarchy and Ownership

### Source of Truth Copilot Instructions Documents (Update These First)

**C1. Main Copilot Instructions** (`.github/copilot-instructions.md`)

- **Purpose**: Core principles, file references table, prompt usage guidance
- **Update Triggers**: New workflow patterns, prompt additions/removals, new instruction file categories
- **Propagate To**: README.md (high-level mention of agent capabilities)

**C2. Copilot Supplementary Instructions** (`.github/instructions/*.instructions.md`)

- **26 instruction files** covering: terminology, continuous work, SpecKit, architecture, security, testing, coding standards
- **Purpose**: Agent behavior rules, quality gates, standards enforcement, technical patterns
- **Update Triggers**: New patterns discovered, regression prevention, anti-pattern documentation, quality gate changes
- **Propagate To**: Specs (architectural constraints), plans (implementation patterns), architecture docs (patterns), prompts (execution rules)

### Source of Truth Architecture Documents (Update These First)

A1. **Architecture** (`docs/arch/ARCHITECTURE.md`)

- **Purpose**: Repository Source of Truth, System design, patterns, component relationships
- **Update Triggers**: Architectural changes, new patterns, component additions
- **Propagate To**: Service template docs, copilot instructions (02-01.architecture)

A2. **Service Template Docs** (`docs/arch/*.md`)

- **Purpose**: Reusable patterns, examples, guidance
- **Sources**: Architecture (patterns), copilot instructions (standards), implementation (validated patterns)
- **Update Triggers**: Pattern validation, anti-pattern discovery, template refactoring

### Spec Kit Documents (Update After Source of Truth)

SK1. **Constitution** (`.specify/memory/constitution.md`)

- **Purpose**: Spec Kit specific Source of Truth, Project vision, quality gates, non-functional requirements, architectural constraints
- **Update Triggers**: Quality gate changes, new NFRs, architectural decisions
- **Propagate To**: Specs, plans, copilot instructions

SK2. **Templates** (`.specify/templates`)
    - **Purpose**: Standardized document structures for specs, plans, tasks
    - **Sources**: SpecKit methodology, copilot instructions (SpecKit)
    - **Update Triggers**: Methodology changes, new document types

SK3. **Specs** Rerun /speckit.spec to update `specs/*/spec.md`

- **Purpose**: Feature requirements, constraints, acceptance criteria
- **Sources**: Constitution (quality gates), architecture (patterns)
- **Update Triggers**: Requirement changes, clarifications, constitution updates

SK4. **Clarify Docs** Rerun /speckit.clarify to update `specs/*/clarify.md` and create next `specs/*/CLARIFY-QUIZME-##.md`
    - **Purpose**: Clarify contains Knowns and Unknowns, Quizme contains only unknowns and open questions
    - **Sources**: Specs (requirements), copilot instructions (clarification patterns)
    - **Update Triggers**: New questions, requirement clarifications

SK5. **Plans** Rerun /speckit.plan to update `specs/*/plan.md`

- **Purpose**: Implementation phases, task breakdown, dependencies
- **Sources**: Specs (requirements), constitution (quality gates), copilot instructions (standards)
- **Update Triggers**: Spec changes, quality gate updates, new phase discoveries

SK6. **Tasks** Rerun /speckit.tasks to update `specs/*/tasks.md`

- **Purpose**: Execution checklists, completion criteria, evidence tracking
- **Sources**: Plans (phases/tasks), constitution (quality gates), copilot instructions (evidence requirements)
- **Update Triggers**: Plan updates, quality gate changes, new task discoveries

SK7. **Analyze** Rerun /speckit.analyze to update `specs/*/analyze.md`

- **Purpose**: Gap analysis, risk assessment, test coverage analysis
- **Sources**: Plans (tasks), copilot instructions (testing standards)
- **Update Triggers**: Task completions, test results, risk discoveries

SK8. **DETAILED.md** Apply updates from `specs/*/tasks.md` to `specs/*/implement/DETAILED.md`

- **Purpose**: Task checkboxes for progress tracking, Append-only timeline of session notes
- **Sources**: Tasks (checkboxes), implementation (discoveries)
- **Update Triggers**: Task completion, lessons learned, constraint discoveries

SK9. **EXECUTIVE.md** Apply updates from `specs/*/implement/DETAILED.md` to `specs/*/EXECUTIVE.md`

- **Purpose**: High-level stakeholder overview, progress metrics, blockers, risks
- **Sources**: DETAILED.md (progress), tasks (completion %), quality metrics
- **Update Triggers**: Phase completion, blocker discovery, risk identification

### Custom Plan Documents (Update After Source of Truth)

P1. **Fixes Needed Plan** `docs/fixes-needed-*/plan.md`

- **Purpose**: High-level issues and fixes
- **Sources**: Specs (specs/clarify/plan/aanlyze/tasks), copilot instructions (execution rules)
- **Update Triggers**: New issues discovered, fix implementations, task completions

P2. **Fixes Needed Tasks** `docs/fixes-needed-*/tasks.md`
    - **Purpose**: Detailed task checkbox lists for fixes needed
    - **Sources**: Fixes Needed Plan (tasks), copilot instructions (evidence requirements)
    - **Update Triggers**: Task completions, new fix discoveries

### Reference Documentation (Update When Patterns Stabilize)

R1. **READMEs** (`README.md`, `docs/README.md`, `specs/*/README.md`)
    - **Purpose**: Getting started, quick reference, navigation
    - **Sources**: All above (high-level summaries)
    - **Update Triggers**: Major project changes, new workflows, documentation reorganization

---

## Synchronization Workflow

### Step 1: Identify What Changed

**Questions**:

- Which source of truth document was updated? (copilot instructions, constitution, architecture, templates)
- What type of change? (quality gates, execution rules, architectural patterns, security standards, testing standards, SpecKit methodology)
- What are the new/changed values or rules?

**Action**: Note specific changes to propagate

---

### Step 2: Find All Affected Documents

**Use grep to find references**:

```bash
# Example: Find all references to a value or pattern
grep -r "<search-term>" docs/ specs/ .github/
```

**Action**: Create list of files requiring updates with line numbers

---

### Step 3: Update Documents in Order

**Update Order** (source of truth → derived → implementation → reference):

1. **Source of Truth**: Copilot instructions, constitution, architecture, templates (ALREADY UPDATED - this is the trigger)
2. **Derived Documents**: Specs, plans, tasks, analyze docs
3. **Implementation Docs**: DETAILED.md, EXECUTIVE.md
4. **Reference Docs**: READMEs, guides

**For Each Document**:

- Read current content around change location
- Update values/rules to match source of truth
- Preserve document-specific context (adapt, don't copy verbatim)
- Use `replace_string_in_file` or `multi_replace_string_in_file`

---

### Step 4: Validate Consistency

**Cross-Document Validation**:

```bash
# Verify all documents use same values
grep -r "<value>" docs/ specs/ .github/ | sort | uniq -c

# Check for no orphaned old values
grep -r "<old-value>" docs/ specs/ .github/ | wc -l  # Should be 0
```

**Checklist**:

- [ ] All documents reference same values
- [ ] No conflicting guidance across documents
- [ ] No orphaned references to old values

---

### Step 5: Commit with Audit Trail

**Conventional Commit Format**:

```
docs(sync): synchronize [change type] across documentation

Updated documents:
- .github/instructions/XX.instructions.md: [change]
- specs/002-cryptoutil/spec.md: [change]
- docs/arch/ARCHITECTURE.md: [change]

Changes:
- [Specific values/rules changed]

Verification:
- grep search shows consistent values across all docs
- No conflicting guidance found

Related: [task reference if applicable]
```

---

## Documentation Sprawl Prevention

### Anti-Patterns to Avoid

❌ **Creating Summary Docs** (COMPLETION-SUMMARY.md, ANALYSIS.md)

- Violation: Duplicates data from authoritative sources
- Prevention: Append to DETAILED.md Section 2 timeline instead

❌ **Creating Backup Docs** (plan_backup.md, tasks_backup.md)

- Violation: Git history already provides backup
- Prevention: Use `git checkout <hash> -- file` to restore

❌ **Creating Verbose Analysis Docs** (SESSION-*.md, COMPLETION-ANALYSIS.md)

- Violation: User never reviews these, they accumulate as debt
- Prevention: Append to DETAILED.md Section 2 with concise findings

❌ **Creating Specialized Docs** (PHASE-X-COVERAGE-GAPS.md)

- Violation: Specialized content should be in main plan/spec sections
- Prevention: Update existing plan/spec sections with details

### Lean Documentation Rules

✅ **ALWAYS prefer updating existing docs over creating new ones**
✅ **ALWAYS check if information exists elsewhere before creating new doc**
✅ **ALWAYS append to DETAILED.md Section 2 for session-specific work**
✅ **ALWAYS update plan.md/tasks.md in-place (no backups)**
✅ **ALWAYS use git history for rollback (not backup files)**

---

## Verification Checklist

Before ending documentation sync:

- [ ] All source of truth documents updated (copilot instructions, constitution, architecture, templates)
- [ ] All derived documents synchronized (specs, plans, tasks, analyze)
- [ ] All implementation docs updated (DETAILED.md, EXECUTIVE.md)
- [ ] All reference docs current (READMEs, guides)
- [ ] grep searches show consistent values across all docs
- [ ] No conflicting guidance found
- [ ] Git commit created with comprehensive audit trail
- [ ] No new documentation sprawl (summaries, backups, verbose analyses, specialized docs)

---

### Scenario 2: New Execution Rule Discovered

**Trigger**: Discovered regression where agent asked "Should I proceed?" despite instructions

**Documents Requiring Sync**:

1. Copilot Instructions (01-02.beast-mode) → Add "NEVER ask 'Should I proceed?'" to prohibited patterns
2. Prompts (autonomous-execution.prompt.md) → Add to continuous execution rules
3. Prompts (beast-mode-3.1.prompt.md) → Add to autonomous behavior section
4. Plans (plan.md) → Update execution workflow section

**Sync Workflow**:

```bash
# Step 1: Add rule to copilot instructions
# Edit .github/instructions/01-02.beast-mode.instructions.md

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
