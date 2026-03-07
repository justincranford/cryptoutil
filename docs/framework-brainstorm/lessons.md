# Lessons Learned - Framework Brainstorm Execution

**Purpose**: Persistent memory for the framework brainstorm execution plan. Records
what worked, what did not, root causes, and patterns for use throughout the plan and
for knowledge propagation to ARCHITECTURE.md, agents, skills, and instructions.

---

## Phase 1: Framework v1 Closure

*(To be filled after Phase 1 completes)*

---

## Phase 2: Copilot Skills Audit

*(To be filled after Phase 2 completes)*

---

## Phase 3: Copilot Agents Audit

*(To be filled after Phase 3 completes)*

---

## Phase 4: Instructions Audit

*(To be filled after Phase 4 completes)*

---

## Phase 5: ARCHITECTURE.md Completeness

*(To be filled after Phase 5 completes)*

---

## Cross-Phase Lessons

### Session 1 (2026-03-07): Context Recovery + Agent Updates

**What was found**:
1. Framework v1 implementation is 58% complete (28/48 tasks). Phase 5 shared test
   packages actually exist in the codebase but task status fields were never updated.
2. framework-brainstorm/ had 9 research docs but no plan.md or tasks.md — the
   brainstorm research was disconnected from an executable plan.
3. SKILL.md files had invalid `metadata: domain:` frontmatter blocks not in VS Code spec.
   These caused YAML validation warnings.

**What was done**:
1. Fixed: removed `metadata:` blocks from all 14 SKILL.md files (commit 2acabc715)
2. Added: lessons.md as 5th companion document to implementation-planning and implementation-execution agents
3. Added: post-mortem step to every phase template in implementation-planning agent
4. Expanded: implementation-execution Phase-Based Post-Mortem to include code/tests/workflows/docs
5. Expanded: ARCHITECTURE.md 13.8.1 and 13.8.2 to include all artifact types
6. Added: Semantic Grouping & Periodic Commits directives to beast-mode, doc-sync, fix-workflows agents
7. Created: framework-brainstorm/plan.md and tasks.md (this session)

**Root Causes**:
- SKILL.md `metadata:` blocks: VS Code agent skills spec was not consulted during initial skill creation; a project-specific convention was invented without verifying against the actual spec
- Disconnected brainstorm docs: The brainstorm research was done but never translated into an executable plan with task tracking
- Stale task status: Phase 5 packages were implemented as part of tasks 5.7/5.8 work but individual task statuses 5.1-5.6 were not back-filled

**Patterns**:
- Always verify toolkit/platform specs (VS Code, GitHub, etc.) before adding custom conventions
- Research docs MUST be followed by plan.md + tasks.md before ending a session
- When a "higher-level" task (5.7: migrate) is marked done, verify that "lower-level" prerequisite tasks (5.1-5.6: create) are also marked done

**Candidates for Knowledge Propagation**:
- SKILL.md frontmatter: ARCHITECTURE.md 2.1.5 already updated ✅
- lessons.md as 5th companion doc: ARCHITECTURE.md 13.8 and agents already updated ✅
- Post-mortem across all artifact types: ARCHITECTURE.md 13.8.1/13.8.2 and agents already updated ✅
- Periodic commits: ARCHITECTURE.md 13.2.2 already has it; propagated to agents ✅
- lint-fitness not in ARCHITECTURE.md: Task 1.3 / Phase 5 Task 5.1 to address
- Shared test infra packages not in ARCHITECTURE.md: Task 1.3 / Phase 5 Task 5.2 to address

