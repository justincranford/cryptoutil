# Speckit Feedback Loop Analysis and Solutions

**Date**: December 19, 2025
**Context**: Analysis of feedback loop challenges in Speckit methodology
**Purpose**: Provide concrete patterns for integrating implementation learnings back into specifications

---

## Core Problem Analysis

Speckit presents as sequential workflow, but real software development is iterative. The fundamental issues:

1. **One-way flow assumption**: Speckit steps assume prerequisites are complete and frozen
2. **Context isolation**: Implementation insights (EXECUTIVE.md, DETAILED.md) exist separately from earlier artifacts
3. **LLM agent amnesia**: When restarting, the agent doesn't have implementation context—only sees original documents
4. **Missing feedback mechanism**: No explicit pattern for bubbling implementation learnings back to constitution/spec/plan

**What causes restarts to fail:**

- Treating earlier documents as "done" instead of living documents
- Attempting "hard restarts" (copy, truncate, restart) instead of continuous refinement
- Storing learnings only in implementation docs without integrating them into constitution/spec/plan
- Waiting until implementation is "complete" before trying to update earlier steps

---

## Recommended Solutions

### 1. Continuous Document Evolution

**DON'T**: Wait until implementation complete to update specifications
**DO**: Update constitution/spec/plan immediately when discovering insights

**Pattern**: Append implementation-driven constraints to constitution.md

```markdown
# Constitution.md Enhancement

## Implementation-Driven Constraints (LIVING SECTION)
*Updated continuously during implementation*

### Database Layer
- **Constraint**: SQLite with GORM requires MaxOpenConns=5 for transaction support
  - **Rationale**: GORM transactions need separate connections; MaxOpenConns=1 causes deadlock
  - **Source**: Phase 1.2.3, DETAILED.md 2025-12-15, commit abc123
  - **Impact**: All SQLite configurations must set this

### Testing Environment
- **Constraint**: Test servers MUST bind 127.0.0.1, never 0.0.0.0
  - **Rationale**: Windows Firewall exception prompts block CI/CD
  - **Source**: EXECUTIVE.md Risk #3, Phase 2.1.1
  - **Impact**: All test configurations, integration tests
```

---

### 2. Explicit Feedback Integration Workflow

Create `specs/002-cryptoutil/FEEDBACK-MAP.md`:

```markdown
# Implementation → Specification Feedback Map

## Purpose
Track which implementation insights update which specification sections.

## Active Feedback Items

| Date | DETAILED.md Entry | Insight | Target Document | Status |
|------|-------------------|---------|-----------------|--------|
| 2025-12-15 | Phase 1.2.3 | SQLite MaxOpenConns | constitution.md §Database | ✅ Updated |
| 2025-12-16 | Risk #3 | Windows Firewall | spec.md §Testing | ✅ Updated |
| 2025-12-17 | Phase 2.1.1 | OAuth token validation | plan.md §Phase 2 | ⏳ Pending |

## Integration Process
1. After each task completion, check DETAILED.md Section 2 timeline
2. Identify insights that change requirements/constraints/approach
3. Update target document immediately (don't batch)
4. Mark in this table for tracking
```

---

### 3. Mini-Cycle Pattern (Tighter Feedback Loops)

Don't wait for full phase completion—use 3-5 task groups:

```
Implement Task Group (3-5 tasks)
  ↓
Update DETAILED.md Section 2 (timeline entry)
  ↓
Extract learnings → Update constitution/spec/plan immediately
  ↓
Commit updated docs
  ↓
Continue to next task group
```

**Example**:

```bash
# Complete Phase 1, Tasks 1.1-1.3 (Database setup)
# Discover: SQLite requires specific GORM config

# IMMEDIATELY update specs (don't wait)
# Edit constitution.md, add SQLite constraint
git add specs/002-cryptoutil/constitution.md
git commit -m "docs(constitution): add SQLite GORM MaxOpenConns constraint from Phase 1.2"

# Continue Phase 1, Tasks 1.4-1.6
```

---

### 4. Bridge Document for Hard Restarts

**If you must restart from earlier step**, create explicit bridge:

```markdown
# CONTEXT-FROM-ITERATION-001.md

## Key Architectural Decisions
- SQLite GORM configuration pattern (MaxOpenConns=5)
- Test server binding strategy (127.0.0.1 only)
- Mutation testing targets (≥98% efficacy)

## Discovered Requirements
- OAuth token caching needed (performance bottleneck in Phase 2)
- WebAuthn requires browser-specific implementations
- Rate limiting needs separate per-IP and per-tenant limits

## Process Improvements
- Continuous document evolution works better than batch updates
- Mini-cycles (3-5 tasks) provide good feedback frequency
- FEEDBACK-MAP.md essential for tracking updates
```

Place in `.specify/memory/` so LLM agent picks it up on restart.

---

### 5. Spec.md Enhancement Pattern

Add living sections to spec.md:

```markdown
# spec.md Enhancement

## Implementation Feedback Loop (LIVING SECTION)

### Architectural Discoveries
- **Database Layer**: Originally specified "use GORM", but implementation revealed SQLite requires specific connection pool settings for GORM transactions. Updated architecture to mandate MaxOpenConns=5 for SQLite with GORM.

### Requirements Refinements
- **Testing Requirements**: Added mandatory requirement: "All test servers MUST bind to 127.0.0.1 to prevent Windows Firewall prompts" (discovered during Phase 1 testing)

### Clarifications from Implementation
- **Mutation Testing**: Originally vague "use mutation testing", now clarified to ≥98% efficacy, ≥90% mutant coverage, gremlins tool specific
```

---

### 6. EXECUTIVE.md Feedback Section

Add upward feedback tracking to EXECUTIVE.md:

```markdown
# EXECUTIVE.md Enhancement

## Upward Feedback: Specification Changes Required

### Constitution Updates Needed
1. ✅ **APPLIED** (2025-12-16): Add SQLite GORM constraint
2. ⏳ **PENDING**: Add OAuth token caching requirement (discovered Phase 2.3)
3. ⏳ **PENDING**: Add WebAuthn browser compatibility constraint (discovered Phase 3.1)

### Spec Updates Needed
1. ✅ **APPLIED** (2025-12-16): Clarify mutation testing targets
2. ⏳ **PENDING**: Add explicit JWT validation requirements (Phase 2.4 revealed ambiguity)

### Plan Updates Needed
1. ⏳ **PENDING**: Revise Phase 3 approach based on WebAuthn complexity
2. ⏳ **PENDING**: Add Phase 2.5 for OAuth token caching (not in original plan)
```

---

## Anti-Patterns to Avoid

1. **Batch Updates**: ❌ Don't wait until phase complete
   → ✅ Update specs immediately when insights discovered

2. **Frozen Artifacts**: ❌ Don't treat constitution/spec as "done"
   → ✅ Treat as living documents that evolve with implementation

3. **Context Islands**: ❌ Don't keep learnings isolated in EXECUTIVE.md
   → ✅ Explicitly integrate insights back into specification documents

4. **Hard Restart Without Bridge**: ❌ Don't copy/delete expecting context preservation
   → ✅ Create explicit bridge documents (FEEDBACK-MAP.md, CONTEXT-FROM-PREVIOUS.md)

5. **Ignoring Mini-Cycles**: ❌ Don't wait for full implementation
   → ✅ Use 3-5 task groups as mini-cycle boundaries

6. **Missing Traceability**: ❌ Don't update specs without linking sources
   → ✅ Always cite DETAILED.md dates, commit hashes, EXECUTIVE.md sections

---

## Recommended Immediate Actions

1. **Right now**: Add "Implementation-Driven Constraints" section to constitution.md
2. **Today**: Create FEEDBACK-MAP.md to start tracking spec updates
3. **This week**: Establish mini-cycle pattern (update specs after every 3-5 tasks)
4. **Next iteration**: Use enhanced documents as base—they'll have feedback baked in

---

## Key Insight

**Don't restart the process; evolve the documents continuously.**

Speckit's sequential structure is a scaffold, not a straitjacket. Treat it as a spiral where each pass through implementation enriches the earlier artifacts for the next cycle. The goal is continuous refinement, not periodic restarts.
