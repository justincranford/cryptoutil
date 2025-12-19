# Speckit Copilot Instructions - Proposed 06-01.speckit.instructions.md

**Purpose**: Guidelines for LLM agents working with Speckit methodology during iterative development
**Status**: Draft for review - consider adding to .github/instructions/ as 06-01.speckit.instructions.md

---

## Core Principles

### Iterative Development Reality

**CRITICAL**: Speckit presents as sequential workflow (constitution → specify → clarify (optional) → plan → tasks → analyze (optional) → implement), but **real development is iterative**. Constitution, spec, and plan are **living documents** that evolve through implementation feedback, not static prerequisites.

**Balance**: Follow Speckit structure for organization, but embrace continuous refinement of earlier documents based on implementation discoveries.

---

## Continuous Document Evolution

### During Implementation: Update Specs Immediately

**NEVER wait until phase/iteration complete to update specifications**
**ALWAYS update constitution/spec/plan when discovering constraints, clarifications, or architectural insights**

**Pattern: Mini-Cycle Feedback (Every 3-5 Tasks)**:

```
1. Complete task group (3-5 related tasks)
2. Document in DETAILED.md Section 2 timeline
3. Identify learnings affecting requirements/constraints
4. Update constitution.md/spec.md/plan.md immediately
5. Commit changes with traceability
6. Continue to next task group
```

### Implementation-Driven Constraints

**Add living sections to constitution.md**:

```markdown
## Implementation-Driven Constraints (LIVING SECTION)
*Updated continuously during implementation*

### [Category] (e.g., Database Layer, Testing, Security)
- **Constraint**: [Brief description]
  - **Rationale**: [Why this constraint exists]
  - **Source**: [DETAILED.md entry, commit hash, EXECUTIVE.md section]
  - **Impact**: [What code/config must follow this]
  - **Date Discovered**: YYYY-MM-DD
```

**Examples**:

- SQLite GORM transaction configuration (MaxOpenConns=5)
- Windows test server binding (127.0.0.1 only)
- Mutation testing thresholds (≥98% efficacy)
- OAuth token caching requirements

---

## Feedback Integration Workflow

### Track Specification Updates

**Create and maintain `specs/NNN-cryptoutil/FEEDBACK-MAP.md`**:

```markdown
| Date | Source | Insight | Target Doc | Status |
|------|--------|---------|------------|--------|
| YYYY-MM-DD | DETAILED.md Phase X.Y | [Brief insight] | constitution.md §Section | ✅ Applied |
| YYYY-MM-DD | EXECUTIVE.md Risk #N | [Brief insight] | spec.md §Section | ⏳ Pending |
```

**Update frequency**: After each mini-cycle (3-5 tasks), scan DETAILED.md Section 2 timeline for insights requiring spec updates.

### Upward Feedback in EXECUTIVE.md

**Add section to EXECUTIVE.md**:

```markdown
## Upward Feedback: Specification Changes

### Constitution Updates
- [x] Applied YYYY-MM-DD: [Description] → [Constitution section]
- [ ] Pending: [Description] → [Constitution section]

### Spec Updates
- [x] Applied YYYY-MM-DD: [Description] → [Spec section]
- [ ] Pending: [Description] → [Spec section]

### Plan Updates
- [ ] Pending: [Description] → [Plan section]
```

---

## Traceability Requirements

**ALWAYS link specification updates back to implementation source**:

- ✅ "Source: DETAILED.md 2025-12-15 Phase 1.2.3, commit abc1234"
- ✅ "Source: EXECUTIVE.md Risk #3, encountered in Phase 2 Task 1.2"
- ❌ "Updated requirement based on implementation" (no source citation)

**Why**: Enables understanding of why constraints exist, and validates their continued necessity.

---

## Hard Restart Pattern (If Absolutely Necessary)

**Prefer continuous evolution over hard restarts**, but if restart required:

### Pre-Restart Checklist

1. **Extract learnings from EXECUTIVE.md**: Copy all Lessons Learned, Risks, Workarounds
2. **Extract insights from DETAILED.md Section 2**: Scan timeline for architecture decisions, requirement clarifications
3. **Update constitution.md**: Add all Implementation-Driven Constraints
4. **Update spec.md**: Add all requirement refinements and clarifications
5. **Update plan.md**: Incorporate process improvements and discovered risks
6. **Create bridge document**: `CONTEXT-FROM-ITERATION-NNN.md` with key insights
7. **Place bridge in `.specify/memory/`**: Ensures LLM agent picks up context on restart

### Restart Execution

```bash
# Copy current speckit with updated docs
cp -r specs/002-cryptoutil specs/003-cryptoutil-iteration-002
cd specs/003-cryptoutil-iteration-002

# Keep: constitution.md, spec.md, clarify.md (NOW WITH FEEDBACK INTEGRATED)
# Remove: plan.md, tasks.md, implement/ (regenerate from enriched specs)
rm plan.md tasks.md
rm -rf implement/

# Restart from planning with enriched context
# /speckit.plan (using updated constitution + spec with implementation feedback)
```

---

## Anti-Patterns - NEVER DO THESE

1. **Batch Spec Updates**: ❌ Waiting until iteration complete to update specs
   → ✅ Update immediately when insights discovered (mini-cycle pattern)

2. **Frozen Prerequisites**: ❌ Treating constitution/spec as immutable after initial creation
   → ✅ Treat as living documents that evolve with implementation

3. **Context Isolation**: ❌ Keeping learnings only in EXECUTIVE.md/DETAILED.md
   → ✅ Explicitly integrate insights back into constitution/spec/plan

4. **Missing Traceability**: ❌ Updating specs without source citations
   → ✅ Always cite DETAILED.md dates, commit hashes, EXECUTIVE.md sections

5. **Hard Restarts Without Bridge**: ❌ Copying/truncating without preserving context
   → ✅ Create explicit bridge documents before restarting

---

## Integration with Existing Instructions

**Placement Consideration**: This file could become `06-01.speckit.instructions.md` in `.github/instructions/`, positioned after evidence-based completion (05-01) and before any product-specific instructions.

**Relationship to Other Instructions**:

- Complements evidence-based completion (05-01): Feedback loops provide evidence for evolving requirements
- Works with git workflow (03-03): Commit spec updates incrementally with implementation changes
- Supports documentation standards (constitution.md): Add living sections for implementation feedback

**Activation**: Apply these patterns when working on specs/NNN-cryptoutil/ directories with Speckit workflow files (constitution.md, spec.md, plan.md, tasks.md, DETAILED.md, EXECUTIVE.md).

---

## Summary: Key Behavioral Changes

1. **Continuous Evolution**: Update specs during implementation, not after
2. **Mini-Cycles**: Feedback loop every 3-5 tasks, not every phase
3. **Living Sections**: Add "Implementation-Driven Constraints" to constitution.md
4. **Feedback Tracking**: Use FEEDBACK-MAP.md and EXECUTIVE.md upward feedback sections
5. **Traceability**: Always cite sources (DETAILED.md entries, commit hashes)
6. **Bridge Documents**: Create explicit context handoffs for restarts

**Core Principle**: Speckit is a spiral, not a waterfall. Each implementation pass enriches the specifications for the next iteration.
