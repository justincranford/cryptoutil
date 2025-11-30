# Feature Template Documentation

**Purpose**: Reusable templates and methodologies for planning and executing complex feature implementations with LLM Agent autonomy

**Created**: November 23, 2025
**Version**: 1.0
**Target Audience**: LLM Agents (primary), Human Developers (reference)

---

## 📁 Directory Contents

| File | Purpose | Primary Audience | Usage Frequency |
|------|---------|------------------|-----------------|
| `feature-template.md` | Complete feature planning template with all sections | Human + LLM | Once per new feature |
| `usage-example.md` | Filled-out example showing how to use template | Human + LLM | Reference during planning |
| `agent-quick-reference.md` | Condensed checklist for LLM autonomous execution | LLM Agent | Every task during implementation |
| `README.md` | This file - directory overview and usage guide | Human | Initial setup |

---

## 🎯 When to Use This Template

### ✅ Use Template For

- **New feature development**: Multi-day/week implementation efforts
- **Major refactoring initiatives**: Restructuring existing code
- **Cross-cutting infrastructure changes**: Changes affecting multiple services
- **Service extraction/migration**: Moving code between packages/services
- **Complex bug fixes**: Bugs requiring architectural changes

### ❌ Don't Use Template For

- **Simple bug fixes**: Use issue tracker, not full feature template
- **Minor enhancements**: Single file changes, small tweaks
- **Documentation updates**: Use README directly
- **Dependency updates**: Use automated tools (go-update-direct-dependencies)
- **Linting fixes**: Use golangci-lint --fix

---

## 🚀 Quick Start Guide

### For Human Developers (Planning Phase)

**Step 1: Copy Template**

```bash
# Create feature directory
mkdir -p docs/<FEATURE_ID>

# Copy template to feature directory
cp docs/feature-template/feature-template.md docs/<FEATURE_ID>/MASTER-PLAN.md
```

**Step 2: Fill Out Template**

- Read `usage-example.md` for guidance
- Fill in Executive Summary (current reality, goals, constraints)
- Define Implementation Tasks table (task breakdown with dependencies)
- Create individual task documents:
  - Naming: `01-<TASK>.md`, `02-<TASK>.md` (zero-padded sequential)
  - Structure: Copy from `task-template.md` or follow MASTER-PLAN task sections
  - Dependencies: Reference parent/blocker tasks in Prerequisites section
  - Acceptance: Define measurable success criteria per task
- Customize acceptance criteria per task
- Scale sections based on feature complexity (remove unused sections)

**Step 3: Review and Refine**

- Validate task dependencies (no circular dependencies)
- Verify acceptance criteria are measurable
- Check risk assessment is realistic
- Ensure constraints are enforceable

**Step 4: Prepare for LLM Agent**

- Ensure all task docs are created and complete
- Verify acceptance criteria are clear and testable
- Add task-specific notes for complex tasks
- Reference `agent-quick-reference.md` in master plan

**Step 5: Handoff to LLM Agent**

```
USER: "Implement feature using master plan at docs/<FEATURE_ID>/MASTER-PLAN.md
Follow continuous work directives in agent-quick-reference.md
Work until all tasks complete or 950k tokens used"
```

### For LLM Agents (Execution Phase)

**Pre-Session (5 min)**

1. Read `agent-quick-reference.md` (THIS IS YOUR BIBLE)
2. Read master plan document (understand goals, constraints, task sequence)
3. Read ALL task documents (understand full scope)
4. Check `manage_todo_list` (identify current status)

**During Session (until 950k tokens OR all tasks complete)**

1. Follow per-task loop in `agent-quick-reference.md`
2. NEVER stop between tasks (zero-text rule)
3. Create post-mortem after EVERY task
4. Mark complete → IMMEDIATELY start next task
5. Check token usage after each tool call

**End Session (ONLY when stopping)**

1. Commit pending changes
2. Update `manage_todo_list`
3. Push commits
4. Provide brief summary to user

---

## 📋 Template Structure Overview

### feature-template.md Sections

**Executive Summary** (Page 1-3)

- Feature overview and status
- Current reality and production blockers
- Completion metrics
- Remediation approach

**Goals and Objectives** (Page 3-6)

- Primary goals with success criteria
- Secondary goals (nice-to-have)
- Non-goals (explicitly out of scope)
- Constraints and boundaries

**Context and Baseline** (Page 6-9)

- Historical context (previous attempts, related work)
- Baseline assessment (current implementation status)
- Code analysis (LOC, coverage, TODO count)
- Stakeholder analysis

**Architecture and Design** (Page 9-14)

- System architecture diagrams
- Component breakdown
- Design patterns used
- Directory structure
- API design
- Database schema
- Security design

**Implementation Tasks** (Page 14-18)

- Task organization and numbering
- Implementation tasks table (ID, status, priority, effort, dependencies, risk)
- Task dependency graph
- Implementation phases

**Task Execution Instructions** (Page 18-24)

- LLM Agent continuous work directives (CRITICAL)
- Task execution checklist
- Testing guidelines
- Tool usage rules

**Post-Mortem and Corrective Actions** (Page 24-32)

- Post-mortem structure and template
- Corrective action patterns
- Real-world SDLC practices
- Lessons learned framework

**Quality Gates and Acceptance Criteria** (Page 32-35)

- Universal acceptance criteria (all tasks)
- Task-specific criteria templates
- Quality gate enforcement (pre-commit, pre-push, CI/CD, production)

**Risk Management** (Page 35-37)

- Risk categories and assessment matrix
- Risk mitigation strategies
- Risk monitoring and escalation

**Success Metrics** (Page 37-39)

- Completion metrics
- Performance metrics
- Business metrics
- Quality metrics

**Appendix** (Page 39-40)

- Terminology
- References (internal docs, external standards)
- Version history
- Template usage guidelines

---

## 🎓 Key Concepts

### Continuous Work Directive

**The PRIMARY RULE for LLM Agents:**

```
NEVER STOP until:
1. All tasks complete, OR
2. Token usage ≥ 950k (95% of 1M budget), OR
3. Explicit user command to stop
```

**NOT stopping conditions:**

- Time elapsed
- Commits made
- Tasks complete (check for more work)
- "Feeling" like it's time to stop

**Pattern**: commit → IMMEDIATE tool call (zero text between)

### Post-Mortem Discipline

**EVERY task gets a post-mortem:**

- File: `##-<TASK>-POSTMORTEM.md`
- Created: Immediately after task completion
- Purpose: Document bugs, omissions, learnings, corrective actions
- Not optional: Critical for continuous improvement

**Corrective actions include:**

- Append new task docs
- Insert hierarchical subtasks
- Update upcoming task acceptance criteria
- Document pattern improvements
- Update risk register

### Quality Gate Enforcement

**Four checkpoints:**

1. **Pre-commit**: Local hooks (linting, formatting, test patterns)
2. **Pre-push**: Local validation (all tests, coverage, dependencies)
3. **PR Merge**: CI/CD automation (security scans, benchmarks)
4. **Production**: Deployment gates (load tests, smoke tests, rollback verified)

Each task must pass quality gates before marking complete.

### Task Granularity

**Guideline**: Each task should be 4-12 hours of work

**Too granular** (< 2 hours):

- Creates excessive overhead (commits, post-mortems)
- Fragments related work
- Slows progress

**Too coarse** (> 16 hours):

- Difficult to track progress
- Risky (large changes harder to review)
- Violates atomic commit principle

**Sweet spot** (4-12 hours):

- Meaningful units of work
- Clear acceptance criteria
- Atomic commits
- Manageable risk

---

## 📊 Template Customization Guidelines

### Sections to Always Keep

**NEVER remove these sections** (critical for LLM Agent execution):

- Executive Summary
- Goals and Objectives
- Implementation Tasks (table with dependencies)
- Task Execution Instructions
- Post-Mortem and Corrective Actions
- Quality Gates and Acceptance Criteria

### Sections You Can Remove

**Remove if not applicable**:

- Historical Context (new features without previous attempts)
- Stakeholder Analysis (internal tools with obvious stakeholders)
- Database Schema (non-database features)
- API Design (non-API features)
- Security Design (non-security-critical features)

### Sections You Should Customize

**Always customize these**:

- Constraints and Boundaries (project-specific rules)
- Design Patterns (domain-specific patterns)
- Directory Structure (project layout)
- Task-Specific Acceptance Criteria (feature requirements)
- Risk Assessment (feature-specific risks)

### Scaling by Complexity

| Feature Complexity | Tasks | Template Pages | Customization |
|-------------------|-------|----------------|---------------|
| Simple | 3-5 | 10-15 | Remove 40% of sections |
| Medium | 6-12 | 15-25 | Remove 20% of sections |
| Complex | 13-20 | 25-40 | Use full template |
| Epic | 20+ | 40+ | Add custom sections |

---

## 🔍 Example Features by Complexity

### Simple Feature (5 tasks, 10 pages)

**Example**: Add new client authentication method

- Tasks: 01-domain-model, 02-authenticator, 03-handler-integration, 04-tests, 05-docs
- Timeline: 1 week
- Template customization: Removed historical context, stakeholder analysis, reduced risk matrix

### Medium Feature (8 tasks, 20 pages)

**Example**: OAuth 2.1 Authorization Server (see `usage-example.md`)

- Tasks: 01-domain, 02-repos, 03-authz-flow, 04-tokens, 05-introspect, 06-client-auth-basic, 07-client-auth-jwt, 08-e2e
- Timeline: 2-3 weeks
- Template customization: Condensed historical context, focused risk assessment, domain-specific sections

### Complex Feature (15 tasks, 35 pages)

**Example**: Identity V2 Implementation (see `docs/02-identityV2/MASTER-PLAN.md`)

- Tasks: 20 tasks across foundation, core, advanced, integration
- Timeline: 2-3 months
- Template customization: Full template used, added domain-specific sections for OAuth/OIDC

### Epic Feature (25+ tasks, 50+ pages)

**Example**: Multi-service refactor (see `docs/01-refactor/PLANS-INDEX.md`)

- Tasks: 20+ tasks across planning, code migration, CLI, infrastructure, testing
- Timeline: 6+ months
- Template customization: Added service group taxonomy, dependency analysis, migration phases

---

## 🎯 Success Patterns

### What Works Well

**Clear acceptance criteria**:

- Specific: "Authorization code flow matches RFC 6749 Section 4.1"
- Measurable: "P95 latency < 100ms"
- Testable: "All tests pass with ≥85% coverage"

**Granular tasks**:

- 4-12 hour effort per task
- Clear dependencies
- Atomic commits
- Distinct deliverables

**Comprehensive post-mortems**:

- Document ALL issues (bugs, omissions, pattern problems)
- List corrective actions (immediate + deferred)
- Create new task docs for gaps
- Update upcoming tasks

**Continuous work discipline**:

- Zero-text between tool calls
- Work until 950k tokens used
- Create post-mortem for EVERY task
- NEVER stop mid-stream

### What Doesn't Work

**Vague acceptance criteria**:

- ❌ "Feature works well"
- ❌ "Good test coverage"
- ❌ "Fast enough"

**Monolithic tasks**:

- ❌ Single task: "Implement entire OAuth 2.1 server" (100+ hours)
- ❌ No clear deliverables
- ❌ Difficult to track progress

**Skipping post-mortems**:

- ❌ "No issues, skip post-mortem"
- ❌ "Too much work to document"
- ❌ "I'll remember the lessons"

**Stopping mid-stream**:

- ❌ Stopping after commits "to check in"
- ❌ Asking permission to continue
- ❌ Providing status updates between tasks

---

## 📚 Related Documentation

**Instruction Files** (MUST READ before implementation):

- `.github/copilot-instructions.md` - Primary instructions
- `.github/instructions/01-01.coding.instructions.md` - Coding patterns and standards
- `.github/instructions/01-02.testing.instructions.md` - Testing patterns and best practices
- `.github/instructions/01-03.golang.instructions.md` - Go project structure and standards
- `.github/instructions/01-04.database.instructions.md` - Database operations and ORM patterns
- `.github/instructions/01-05.security.instructions.md` - Security implementation patterns
- `.github/instructions/01-06.linting.instructions.md` - Code quality and linting standards

**Example Master Plans** (reference implementations):

- `docs/02-identityV2/MASTER-PLAN.md` - Complex feature (20 tasks)
- `docs/01-refactor/PLANS-INDEX.md` - Epic refactor (20+ tasks)
- `docs/04-identity/identity_master.md` - Original identity implementation

**Supporting Documents**:

- `docs/PLANS-INDEX.md` - Index of all documentation plans
- `docs/pre-commit-hooks.md` - Quality gate automation
- `docs/DEV-SETUP.md` - Development environment setup

---

## 🔄 Template Evolution

### Version History

| Version | Date | Changes | Impact |
|---------|------|---------|--------|
| 1.0 | 2025-11-23 | Initial template based on 50+ planning docs analysis | Baseline |

### Feedback Loop

**After each feature using this template:**

1. Create feature post-mortem: `docs/<FEATURE_ID>/FEATURE-POSTMORTEM.md`
2. Document template gaps: What sections were missing?
3. Document template bloat: What sections were unnecessary?
4. Update template: Incorporate improvements
5. Version bump: Increment version, document changes

**Template improvement sources:**

- Feature post-mortems (what worked/didn't work)
- LLM Agent feedback (blockers encountered)
- Human developer feedback (usability issues)
- Industry best practices (new SDLC patterns)

### Contribution Guidelines

**To propose template improvements:**

1. Create detailed RFC: `docs/feature-template/rfcs/<RFC_ID>-<TITLE>.md`
2. Document rationale: Why is this improvement needed?
3. Provide examples: Show before/after with real features
4. Assess impact: What features would benefit? What's the effort to adopt?
5. Review and discuss: Get team consensus
6. Merge improvements: Update template, version bump, document in history

---

## 🎓 Learning Resources

### Spec-Driven Development

**Concept**: Define complete specifications before implementation

- **Specs**: OpenAPI (APIs), Database Schema (data models), Test Cases (behavior)
- **Benefits**: Clear contracts, parallel development, automated validation
- **Tools**: OpenAPI Specification, JSON Schema, Test Data Builders

**Application to Template**:

- Architecture and Design section defines specs
- Task acceptance criteria validate against specs
- Quality gates enforce spec compliance

### LLM Agent Planning Patterns

**Emerging Best Practices**:

1. **Continuous Execution**: Work until budget exhausted, not time-based
2. **Post-Mortem Discipline**: Document learnings after every unit of work
3. **Progressive Enhancement**: Build foundation first, layer features
4. **Quality Gates**: Automated checkpoints prevent regression
5. **Corrective Action Loops**: Feed learnings into future tasks

**Resources**:

- [Anthropic's Claude Guide](https://docs.anthropic.com/claude/docs) - LLM best practices
- [GitHub Copilot Documentation](https://docs.github.com/en/copilot) - AI-assisted development
- [OpenAI GPT Best Practices](https://platform.openai.com/docs/guides/gpt-best-practices) - Prompt engineering

### Real-World SDLC Patterns

**Agile/Scrum**:

- Sprint planning → Feature planning (goals, tasks, timeline)
- Daily standups → Post-mortems (progress, blockers, learnings)
- Retrospectives → Corrective actions (process improvements)

**DevOps**:

- CI/CD pipelines → Quality gates (automated validation)
- Monitoring → Success metrics (performance, quality, business)
- Incident response → Risk management (identification, mitigation)

**Lean**:

- Value stream mapping → Task dependency graph (critical path)
- Kaizen (continuous improvement) → Post-mortem corrective actions
- Waste elimination → Template customization (remove unused sections)

---

## 🚀 Getting Started Checklist

**For your first feature using this template:**

- [ ] Read `feature-template.md` completely (understand all sections)
- [ ] Read `usage-example.md` (see realistic application)
- [ ] Read `agent-quick-reference.md` (understand execution pattern)
- [ ] Copy template to new feature directory
- [ ] Fill out Executive Summary (current reality, goals, metrics)
- [ ] Define Implementation Tasks table (breakdown with dependencies)
- [ ] Create individual task documents (##-<TASK>.md)
- [ ] Customize acceptance criteria (make specific and measurable)
- [ ] Remove inapplicable sections (scale to feature complexity)
- [ ] Review with team (validate approach, get buy-in)
- [ ] Handoff to LLM Agent (provide master plan path, reference quick ref)

**During implementation:**

- [ ] LLM Agent follows `agent-quick-reference.md`
- [ ] Post-mortem created after EVERY task
- [ ] Quality gates enforced before task completion
- [ ] Corrective actions applied (new tasks, pattern improvements)
- [ ] Continuous work until all tasks complete OR 950k tokens

**After feature completion:**

- [ ] Create feature post-mortem
- [ ] Document template gaps/improvements
- [ ] Update template if needed
- [ ] Share learnings with team
- [ ] Celebrate success! 🎉

---

**You now have a complete feature planning and execution framework. Use it to build amazing features with LLM Agent autonomy!**
