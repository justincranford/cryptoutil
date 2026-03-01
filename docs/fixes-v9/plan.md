# Fixes v9 - Quality Review Passes, Agent Semantics, ARCHITECTURE.md Optimization, Skills Migration

## Executive Summary

This plan addresses 5 major areas:
1. **Quality Review Passes Rework** - Make review passes generic and comprehensive
2. **Agent Semantic Analysis** - Ensure beast-mode is generic, others are specific
3. **ARCHITECTURE.md Optimization** - Compact, deduplicate, fix contradictions, fill omissions
4. **doc-sync Agent Propagation** - Ensure all necessary ARCHITECTURE.md content is propagated
5. **Copilot Skills Migration** - Identify candidates for migration from instructions/agents to skills

---

## Phase 1: Quality Review Passes Rework

### Current State (PROBLEM)

The current review passes in ARCHITECTURE.md Section 2.5 and beast-mode.agent.md are:
- **Pass 1 — Completeness**: Only checks completeness
- **Pass 2 — Correctness**: Only checks correctness  
- **Pass 3 — Quality**: Only checks coverage/mutation

### Target State (SOLUTION)

Each review pass MUST check ALL 8 quality attributes:
1. ✅ **Correctness**: Code is functionally correct
2. ✅ **Completeness**: No tasks/steps skipped
3. ✅ **Thoroughness**: Evidence-based validation
4. ✅ **Reliability**: Quality gates enforced (≥95%/98%)
5. ✅ **Efficiency**: Optimized for maintainability
6. ✅ **Accuracy**: Root cause addressed, not symptoms
7. ❌ **NO Time Pressure**: Never rushed
8. ❌ **NO Premature Completion**: Evidence-based completion only

### Review Pass Rules

- **Minimum**: 3 review passes
- **Maximum**: 5 review passes
- **Continuation Rule**: If significant issues found in pass 3, continue passes 4-5 until diminishing returns
- **Scope**: Generic for ALL work types (code, docs, config, tests, infrastructure)

### Files to Update

1. `docs/ARCHITECTURE.md` Section 2.5 (source of truth)
2. `.github/instructions/01-02.beast-mode.instructions.md` (@source)
3. `.github/instructions/06-01.evidence-based.instructions.md` (@source)
4. `.github/agents/beast-mode.agent.md`
5. `.github/agents/doc-sync.agent.md`
6. `.github/agents/fix-workflows.agent.md`
7. `.github/agents/implementation-execution.agent.md`
8. `.github/agents/implementation-planning.agent.md`

---

## Phase 2: Agent Semantic Analysis

### Agent Purpose Matrix

| Agent | Scope | Purpose | Should Be Generic? |
|-------|-------|---------|-------------------|
| beast-mode | ANY work | Continuous autonomous execution | ✅ YES - MUST be generic |
| doc-sync | Documentation | Sync docs, prevent sprawl | ❌ NO - documentation-specific |
| fix-workflows | GitHub Actions | Fix/optimize workflows | ❌ NO - workflows-specific |
| implementation-execution | Plan execution | Execute plan.md/tasks.md | ❌ NO - plan-execution-specific |
| implementation-planning | Plan creation | Create plan.md/tasks.md | ❌ NO - planning-specific |

### Analysis Findings

1. **beast-mode.agent.md** - Currently has some Go-specific examples (go build, golangci-lint). These should be made generic OR moved to domain-specific sections with clear "example for Go" labels.

2. **doc-sync.agent.md** - Correctly scoped to documentation. Has specific tables for doc types, update patterns, propagation sources. KEEP AS-IS.

3. **fix-workflows.agent.md** - Correctly scoped to GitHub Actions. Has workflow-specific patterns, YAML syntax, CI/CD. KEEP AS-IS.

4. **implementation-execution.agent.md** - Correctly scoped to plan execution. References plan.md/tasks.md. KEEP AS-IS.

5. **implementation-planning.agent.md** - Correctly scoped to plan creation. Has plan/tasks templates. KEEP AS-IS.

### Recommendations

- **beast-mode**: Reword examples to be generic, or clearly label as "example commands"
- **Others**: No changes needed - they are correctly domain-specific

---

## Phase 3: ARCHITECTURE.md Optimization

### Document Statistics
- **Current Size**: 4,445 lines
- **Sections**: 14 major sections + 3 appendices
- **Propagate Markers**: 34 @propagate, 32 @/propagate

### Identified Issues

#### 3.1 Duplication Candidates

1. **Quality Attributes** - Repeated in:
   - Section 1.3 Core Principles
   - Section 2.5 Quality Strategy
   - Section 11.1 Maximum Quality Strategy
   - Multiple agent files
   
   **Recommendation**: Consolidate to single canonical source in Section 11.1, propagate to others

2. **Infrastructure Blocker Escalation** - Duplicated in:
   - Section 13.7 Infrastructure Blocker Escalation
   - Section 2.5 Quality Strategy (partial)
   
   **Recommendation**: Keep in 13.7, remove from 2.5 (just cross-reference)

3. **CLI Patterns** - Appear in:
   - Section 4.4.7 CLI Patterns
   - Section 9.1 CLI Patterns & Strategy
   
   **Recommendation**: Consolidate to Section 9.1, update 4.4.7 to cross-reference

4. **Port Assignments** - Detailed in:
   - Section 3.4 Port Assignments & Networking
   - Appendix B.1 Service Port Assignments
   - Appendix B.2 Database Port Assignments
   
   **Recommendation**: Keep tables in Appendix B, Section 3.4 should summarize and reference

#### 3.2 Potential Contradictions

1. **Coverage Targets** - Different values mentioned:
   - "≥95%/98%" (general)
   - "≥95% production, ≥98% infrastructure/utility" (specific)
   
   **Status**: Not a contradiction - context-dependent. Clarify context.

2. **Review Pass Count** - Need to update from "exactly 3" to "3-5 with continuation rule"

#### 3.3 Potential Omissions

1. **Skills Documentation** - No section about Copilot Skills (new VS Code feature)
2. **Agent vs Skill vs Instruction Decision Tree** - Missing guidance on when to use each
3. **Review Pass Continuation Criteria** - Not defined when to continue to passes 4-5

---

## Phase 4: doc-sync Agent Propagation Analysis

### Current State

doc-sync.agent.md has only ONE cross-reference to ARCHITECTURE.md:
- Section 2.5 Quality Strategy (review passes)

### Missing Propagations

The doc-sync agent should reference these ARCHITECTURE.md sections:

1. **Section 12.7 Documentation Propagation Strategy** - Core to doc-sync purpose
2. **Section 11.4 Documentation Standards** - Doc quality requirements
3. **Section 2.5 Mandatory Review Passes** - Already present ✅
4. **Section B.6 Instruction File Reference** - File organization

### Recommendations

Add cross-references to doc-sync.agent.md for sections 12.7 and 11.4.

---

## Phase 5: Copilot Skills Migration Candidates

### Skills vs Instructions vs Agents Decision Tree

| Feature | Instructions | Agents | Skills |
|---------|-------------|--------|--------|
| Scope | Always loaded or file-pattern | On-demand invocation | On-demand, auto-detected |
| Content | Text only | Text + tools + handoffs | Text + scripts + examples |
| Portability | VS Code only | VS Code only | Cross-platform (CLI, coding agent) |
| Best For | Standards/guidelines | Workflows/automation | Specialized tasks with resources |

### Migration Candidates

#### 5.1 From Instructions to Skills

| Instruction File | Candidate Content | Rationale |
|------------------|-------------------|-----------|
| 03-02.testing.instructions.md | Table-driven test patterns | Could include test templates, example files |
| 04-01.deployment.instructions.md | Docker/compose patterns | Could include compose templates, Dockerfile examples |
| 03-05.linting.instructions.md | Lint fix workflows | Could include lint scripts, config examples |

#### 5.2 From Agents to Skills

| Agent | Candidate Capability | Rationale |
|-------|---------------------|-----------|
| fix-workflows | GitHub Actions debugging | Self-contained task with specific patterns |
| doc-sync | Propagation checking | Could be automated script + instructions |

#### 5.3 New Skills to Create

| Skill Name | Purpose | Contents |
|------------|---------|----------|
| test-table-driven | Create table-driven Go tests | SKILL.md + test-template.go + examples/ |
| compose-validation | Validate Docker Compose files | SKILL.md + validation scripts |
| propagation-check | Verify @propagate/@source sync | SKILL.md + lint-docs integration |
| coverage-analysis | Analyze and improve coverage | SKILL.md + coverage scripts |

---

## Success Criteria

### Phase 1 Success
- [ ] Review passes updated to check ALL 8 quality attributes per pass
- [ ] Continuation rule (3-5 passes) documented
- [ ] All @propagate/@source chains updated
- [ ] All agents updated with new review pass format

### Phase 2 Success
- [ ] beast-mode.agent.md examples generic or clearly labeled
- [ ] Other agents confirmed as correctly domain-specific

### Phase 3 Success
- [ ] Duplication reduced by 10-15% without losing semantic meaning
- [ ] All contradictions resolved
- [ ] All omissions addressed

### Phase 4 Success
- [ ] doc-sync.agent.md has cross-references to sections 12.7, 11.4

### Phase 5 Success
- [ ] Skills decision documented
- [ ] Migration candidates identified and prioritized
