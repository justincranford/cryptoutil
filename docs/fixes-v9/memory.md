---
applyTo: '**'
---
# Fixes v9 Execution Memory

## Current Phase: Phase 1 - Quality Review Passes Rework
## Status: In Progress

## Key Facts
- ARCHITECTURE.md is 4445 lines (target: <4000)
- skills/ dir does not exist yet
- All agents have old 3-pass format with Pass1=Completeness, Pass2=Correctness, Pass3=Quality
- B.1 and B.2 are just summary cross-references (safe to delete - full tables in 3.2 and 3.4)
- Section 9.1 is brief, minimal content to remove
- All 5 agents + 2 instruction files need review pass updates

## New Review Passes Text (canonical)
Each pass checks ALL 8 quality attributes. Min 3, max 5.
Continuation: pass3 finds ANY issue → pass4. pass4 still has issues → pass5.

## Files Tracking
- docs/ARCHITECTURE.md (primary, 4445 lines)
- .github/agents/beast-mode.agent.md (482 lines)
- .github/agents/doc-sync.agent.md
- .github/agents/fix-workflows.agent.md
- .github/agents/implementation-execution.agent.md
- .github/agents/implementation-planning.agent.md
- .github/instructions/01-02.beast-mode.instructions.md
- .github/instructions/06-01.evidence-based.instructions.md
