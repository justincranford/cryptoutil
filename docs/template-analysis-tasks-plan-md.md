# Template Analysis: tasks.md and plan.md

**Date**: 2026-02-05
**Focus**: Comparing EXPECTED structure (in agent template) vs ACTUAL structure (in real v10/v9/v8 files)

---

## EXECUTIVE SUMMARY

The EXPECTED templates in .github/agents/plan-tasks-quizme.agent.md were too simplistic for real-world complex work. Analysis of actual v10, v9, and v8 tasks.md/plan.md files revealed patterns and features needed for:

1. **Tracking complex findings** (Evidence sections)
2. **Documenting context** (Purpose, Executive Summary)
3. **Quality enforcement** (Quality Mandate section)
4. **Phase clarity** (Phase Objectives)
5. **Artifact traceability** (Evidence Archive)

**IMPROVEMENTS APPLIED**: Updated both templates in agent file to support these patterns.

---

## KEY DIFFERENCES FOUND

### plan.md Template

| Aspect | EXPECTED | ACTUAL v10/v9/v8 | Fix Applied |
|--------|----------|------------------|------------|
| Purpose field |  None |  Present | Added Purpose |
| Executive Summary |  None |  Rich findings | Added Executive Summary |
| Phase Objectives |  Implicit |  Explicit | Added Objectives + Success |
| Risk scenarios |  Generic |  Concrete (E2E timeouts) | Added examples |
| Background section |  None |  Context | Added context to Overview |

### tasks.md Template

| Aspect | EXPECTED | ACTUAL v10/v9/v8 | Fix Applied |
|--------|----------|------------------|------------|
| Quality Mandate |  None |  Top section | Added Quality Mandate section |
| Evidence sections |  Not mentioned |  Per-task paths | Added Evidence subsections |
| Test commands |  Generic |  Specific commands | Enhanced Acceptance Criteria |
| Phase Objectives |  None |  Before tasks | Added Phase Objectives |
| Deferred Work |  Not mentioned |  Section | Added Notes / Deferred Work |
| Evidence Archive |  Not mentioned |  Artifact links | Added Evidence Archive |

---

## WHY THESE DIFFERENCES EXIST

### Root Cause Analysis

1. **Simple templates inadequate for complex work**
   - Real work: 20-50 tasks, multiple phases, unknowns, analysis needed
   - Simple templates: 2-3 tasks, clear path, linear progression

2. **Quality enforcement requires upfront mandate**
   - Real world issue: work gets deferred when blockers found late
   - Quality Mandate section enforces "all issues are blockers" thinking

3. **Complex analysis needs evidence tracking**
   - Real world issue: decisions made during root cause analysis need documentation
   - Evidence sections provide artifact links for future reference

4. **Phase boundaries need explicit objectives**
   - Real world issue: tasks become disconnected without phase context
   - Phase Objectives clarify what each phase builds and validates

5. **Artifact traceability essential for learning**
   - Real world issue: E2E timeout analysis, mutation testing, coverage gaps need artifacts
   - Evidence Archive enables future iteration planning

---

## LESSONS LEARNED - ANALYSIS PATTERN

When comparing EXPECTED vs ACTUAL:

1. **Read the ACTUAL files first** - understand real patterns, not assumptions
2. **Identify gaps systematically** - missing sections, added features, format differences
3. **Root cause each difference** - why was feature added? what problem solved?
4. **Know the context** - complex work vs simple tasks require different templates
5. **Apply selectively** - enhancements for real work, simplicity for straightforward tasks

---

## IMPROVEMENTS COMMITTED TO AGENT

 plan-tasks-quizme.agent.md updated with:
- Enhanced plan.md template (Purpose, Executive Summary, Objectives, Success)
- Enhanced tasks.md template (Quality Mandate, Evidence, Phase Objectives, Deferred Work, Evidence Archive)
- Detailed Acceptance Criteria with test commands
- Structured phase context
- Rationale for Quality Mandate section

See agent file lines 280-520 for complete enhanced templates.

---

# Plan.md Analysis - Repeat for plan.md (vs tasks.md)

**Date**: 2026-02-05
**Exercise**: Repeat EXPECTED vs ACTUAL analysis, substituting plan.md instead of tasks.md

---

## 1. EXPECTED PLAN.MD STRUCTURE (from agent template, post-enhancement)

### Core Elements (11 sections)

1. **Metadata**: Status, Created, Last Updated, Purpose
2. **Overview**: Brief description of work, goals, scope  
3. **Executive Summary** (optional): For complex work - Critical Context, Assumptions & Risks
4. **Technical Context**: Language, Framework, Database, Dependencies, Related Files
5. **Phases**: Phase 0-4 structure with Objective, Tasks, Success, Time (Xh)
6. **Technical Decisions**: Topic, Chosen, Rationale, Alternatives, Impact, Evidence
7. **Risk Assessment**: Risk | Probability | Impact | Mitigation table format
8. **Quality Gates**: 7-item MANDATORY checklist
9. **Success Criteria**: [ ] Checklist format (6-7 items)
10. **Related Documents**: Links to tasks.md, research.md, other plans

### Key Pattern: EXPECTED is **template-driven** (structured, predictable format)

---

## 2. ACTUAL PLAN.MD STRUCTURES

### v9 Plan.md Additions Beyond EXPECTED

| Feature | EXPECTED | ACTUAL v9 | Pattern |
|---------|----------|-----------|---------|
| Background section | ??? None | ??? YES | Context from prior phases |
| Executive Decisions format | Generic "Technical Decisions" | Options A/B/C/D format | Structured choice documentation |
| Decision Evidence | Basic Evidence field | Specific "Rationale" + examples | Detailed reasoning with context |
| Success Criteria | Checkbox format | Checkbox + emoji mixes (??? vs [x]) | Visual status indicators |
| Time Tracking | Only Estimated (Xh) | Actual vs Estimated both shown | Real data for future planning |
| Phase-level Status | Implied | Explicit (??? COMPLETE, ??? IN PROGRESS) | Immediate phase readiness assessment |
| Port Standards Reference | Not in template | Embedded table | Context-specific reference data |

### v10 Plan.md Additions Beyond EXPECTED

| Feature | EXPECTED | ACTUAL v10 | Pattern |
|---------|----------|-----------|---------|
| Quality Mandate section | ??? None (not in plan.md) | ??? YES | Explicit quality enforcement upfront |
| Root Cause Hypothesis | Not present | "Suspected: cipher-im's 180s timeout (increased from 90s due to cascade dependencies...)" | Analysis with reasoning |
| Comparative Analysis Tables | Implied but not shown | Dockerfile Location Inconsistency, cmd/ Structure Inconsistency, E2E Test Pattern, Health Endpoint tables | Evidence-based comparative analysis |
| Finding Field | Not explicit | "**Finding**: cipher-im maintains Dockerfile in BOTH cmd/ AND deployments/ - creates drift risk" | Explicit structured observations |
| Root Cause Analysis Section | Not present | Dedicated sections for each major issue discovery | Systematic issue documentation |
| Action Required Field | Not present | "**Action Required**: Verify jose-ja/sm-kms E2E actually exist and pass..." | Next-step clarity in analysis sections |

### v8 Plan.md (if analyzed) - Would show similar patterns

---

## 3. SEMANTIC ASPECTS ANALYSIS

### What Makes ACTUAL Plans Work Better Than EXPECTED

**EXPECTED Plan Template is Linear**:
- Phase 0  Phase 1  Phase 2  Phase 3  Phase 4
- Clear progression, minimal backtracking
- Works for straightforward implementation

**ACTUAL v9/v10 Plans are Non-Linear**:
- Discovery phase with unknowns
- Hypotheses about root causes
- Comparative analysis across services
- Deferred decisions (Options B/C expansion in v9)
- Strategic decisions affecting multiple phases

**Key Insight**: plan.md template assumes discovery is done (Technical Context known). Real work **IS the discovery process**. v9/v10 plans document the discovery itself.

---

## 4. DIFFERENCES: ROOT CAUSE ANALYSIS

### Why ACTUAL Plans Diverge from EXPECTED

**Root Cause 1: Prior Knowledge Unknown**
- EXPECTED assumes all unknowns resolved before plan.md written
- ACTUAL (v10): E2E timeout root cause "Suspected" (hypothesis, not proven)
- ACTUAL (v9): Prior work incomplete, so must reference V8 context
- **Fix**: Add Background section to plan.md template (DONE in enhancement)

**Root Cause 2: Complex Analysis Needs Documentation**
- EXPECTED: Generic section for "Technical Decisions"
- ACTUAL (v9): Structured "Executive Decisions" with Options A/B/C/D format
- ACTUAL (v10): Root cause analysis tables (3x comparative analysis across services)
- **Fix**: Enhance Technical Decisions section with decision alternatives format

**Root Cause 3: Quality Mandate Belongs in Plan Too**
- EXPECTED: plan.md has no Quality Mandate (only in tasks.md)
- ACTUAL (v10): Explicit "Quality Mandate - MANDATORY" section at top
- **Fix**: Add Quality Mandate section to plan.md template

**Root Cause 4: Risk Assessment Needs Concrete Examples**
- EXPECTED: Generic table format (Risk | Probability | Impact | Mitigation)
- ACTUAL (v10): Example populated table with actual risk - "E2E timeouts"
- **Fix**: Already enhanced with concrete examples in template

**Root Cause 5: Success Criteria Don't Show Phase Progress**
- EXPECTED: Only final checkboxes
- ACTUAL (v9): Checkboxes WITHIN each phase (??? COMPLETE vs [ ] TODO)
- **Fix**: Need to show status at phase level, not just final level

---

## 5. IMPROVEMENTS DISCOVERED FOR PLAN.MD TEMPLATE

### Enhancement Candidates (8 findings)

| # | Enhancement | EXPECTED | ACTUAL Pattern | Priority |
|---|-------------|----------|-----------------|----------|
| 1 | Quality Mandate section in plan.md | ??? Absent | ??? Present (v10) | HIGH |
| 2 | Background section (prior work context) | ??? Absent | ??? Present (v9) | HIGH |
| 3 | Executive Decisions Options A/B/C/D format | ??? Generic | ??? Structured (v9) | HIGH |
| 4 | Root Cause Analysis section | ??? Implicit | ??? Explicit tables (v10) | MEDIUM |
| 5 | Phase-level Status indicators | ??? Implicit | ??? Explicit (??? COMPLETE, [ ] TODO) | MEDIUM |
| 6 | Time Tracking (Actual vs Estimated) | ??? Estimated only | ??? Both shown (v9) | MEDIUM |
| 7 | Hypothesis/Suspected findings | ??? Not mentioned | ??? "Suspected: ..." (v10) | LOW |
| 8 | Action Required field | ??? Implicit | ??? Explicit (v10) | LOW |

---

## 6. COMPARISON: SEMANTIC ASPECTS v9 vs v10

### v9 Plan.md Semantic Focus

- **Structural Focus**: How to organize lint-ports enhancements (Options A/B/C/D decisions)
- **Context Focus**: Background from V8, deferred work
- **Efficiency Focus**: Actual vs Estimated time tracking for future planning
- **Status Focus**: Phase-level status (??? COMPLETE vs ??? DEFERRED)

### v10 Plan.md Semantic Focus

- **Root Cause Focus**: Comparative analysis across services to identify divergence
- **Evidence Focus**: Detailed structural analysis tables with concrete findings
- **Hypothesis Focus**: "Suspected: ..." reasoning about why cipher-im unique fails
- **Quality Focus**: Explicit Quality Mandate at top (NEW pattern v10 introduced)

---

## 7. RECOMMENDATIONS FOR AGENT TEMPLATE

### Current Enhancement Status (from prior session)

 **Already Added**:
- Purpose field
- Executive Summary section (Critical Context + Assumptions & Risks)
- Phase Objectives (Objective + Success fields)
- Risk Assessment with concrete examples
- Quality Gates MANDATORY checklist

### Still Needed (identified in this analysis)

??? **Candidate 1: Quality Mandate Section** (HIGH priority)
- V10 added this to plan.md explicitly
- Should be at top, similar to tasks.md mandate
- Template should show emoji format: "??? Fix issues immediately", etc.

??? **Candidate 2: Background Section** (HIGH priority)
- V9 used explicitly: "V8 successfully completed all core objectives..."
- Provides: Context from prior phases, what was deferred, what v9 carries forward
- Should be AFTER Overview, BEFORE Executive Summary

??? **Candidate 3: Executive Decisions Format** (HIGH priority)
- V9 structured as: Options (A/B/C/D) + Selected + Rationale + Evidence
- Current template has "Technical Decisions" which is generic
- Should rename and structure explicitly with Options format

??? **Candidate 4: Root Cause Analysis Section** (MEDIUM priority - only for complex plans)
- V10 added: Comparative tables analyzing across services
- Should be optional, but when needed, structure as:
  - Issue description
  - Comparative analysis table
  - Finding (explicit structured observation)
  - Action Required (next steps)

??? **Candidate 5: Phase-level Status** (MEDIUM priority)
- V9/V10 show: ??? COMPLETE, [ ] TODO, ??? DEFERRED, etc.
- Should allow phase-level status tracking within phases section

---

## 8. CURRENT PLAN.MD TEMPLATE ASSESSMENT

### Completeness Scoring

| Aspect | Score | Status |
|--------|-------|--------|
| Metadata | 100% | ??? Complete (Status, Created, Updated, Purpose) |
| Overview | 100% | ??? Complete |
| Executive Summary | 100% | ??? Complete (Critical Context, Assumptions & Risks) |
| Technical Context | 100% | ??? Complete |
| Phases structure | 85% | ??? Good but missing phase-level status tracking |
| Technical Decisions | 70% | ??? Weak - generic vs structured Options format |
| Risk Assessment | 100% | ??? Complete with concrete examples |
| Quality Gates | 100% | ??? Complete checklist |
| Success Criteria | 80% | ??? Good but missing phase-level indicators |
| **Missing Sections** | ??? | Quality Mandate, Background, Root Cause Analysis |

### Overall Completeness: **75-80%** (good foundation, specific enhancements needed)

---

## 9. NEXT STEPS FOR AGENT TEMPLATE

### Phase A: High-Priority Enhancements (should be applied)

1. Add Quality Mandate section (top of template)
2. Add Background section (after Overview)
3. Rename "Technical Decisions" to "Executive Decisions" with Options format
4. Add phase-level Status field to Phases section

### Phase B: Medium-Priority Enhancements (conditional, complex plans only)

1. Add Root Cause Analysis section (optional, with sub-sections for Issue table + Finding + Action Required)
2. Add Time Tracking to Phase sections (Estimated | Actual)

### Phase C: Low-Priority Enhancements (documentation)

1. Add guidance: "When to use: quality-mandate section (complex work with unknowns only)"
2. Add examples: Options A/B/C format, comparative analysis tables
3. Document Phase-level status indicators

---

## Summary: plan.md Analysis Complete

**Key Finding**: Plan.md template is 75-80% complete. Real-world v9/v10 plans added:
1. Quality Mandate (Quality-first thinking)
2. Background (Prior work context)
3. Executive Decisions format (Structured choice documentation)
4. Root Cause Analysis tables (Evidence-based findings)
5. Phase-level status (Immediate phase readiness)

**Recommendation**: Apply High-Priority enhancements (4 items) to agent template to better capture complex planning patterns observed in v9/v10.
