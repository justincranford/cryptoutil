# Fixes v9 - Quiz for User Review

**Instructions**: For each question, select the option you want to apply. Mark your choice by changing `[ ]` to `[x]`.

---

## Section 1: Quality Review Passes

### Q1: Review Pass Structure

The current review passes check ONE quality attribute per pass (Pass 1=Completeness, Pass 2=Correctness, Pass 3=Quality). How should review passes work?

- [ ] **A)** Keep current structure - one attribute per pass is clearer
- [x] **B)** Each pass checks ALL 8 quality attributes (Correctness, Completeness, Thoroughness, Reliability, Efficiency, Accuracy, NO Time Pressure, NO Premature Completion)
- [ ] **C)** First 2 passes check all attributes, Pass 3 focuses on documentation

---

### Q2: Review Pass Count

How many review passes should be required?

- [ ] **A)** Exactly 3 passes (current)
- [x] **B)** Minimum 3, maximum 5 (continue if significant issues found in pass 3)
- [ ] **C)** Minimum 2, maximum 5 (reduce baseline for simple tasks)
- [ ] **D)** At least 3, no maximum (continue until zero issues)

---

### Q3: Continuation Criteria

When should a 4th or 5th pass be performed?

- [x] **A)** Whenever Pass 3 finds ANY issue
- [ ] **B)** When Pass 3 finds SIGNIFICANT issues (>2 gaps, regressions, or quality gate failures)
- [ ] **C)** Always do 5 passes regardless
- [ ] **D)** User decides case-by-case

---

### Q4: Review Pass Scope

What types of work should review passes apply to?

- [ ] **A)** Code changes only
- [ ] **B)** Code and documentation only
- [x] **C)** ALL work types (code, docs, config, tests, infrastructure, deployments)
- [ ] **D)** Different pass counts for different work types

---

## Section 2: Agent Semantics

### Q5: beast-mode.agent.md Generic vs Specific

The beast-mode agent has Go-specific examples (go build, golangci-lint). How should this be handled?

- [x] **A)** Keep Go-specific examples - this is primarily a Go project
- [ ] **B)** Make all examples generic (e.g., "build command", "lint command")
- [ ] **C)** Keep as-is but add labels like "Example for Go projects"
- [ ] **D)** Add examples for multiple languages (Go, Python, Java)

---

### Q6: beast-mode Purpose

What should beast-mode's scope be?

- [x] **A)** Generic continuous execution for ANY work type
- [ ] **B)** Primarily code tasks with some generic guidance
- [x] **C)** Keep dual: generic principles + language-specific Quality Gates

---

### Q7: Other Agent Scopes

The other agents (doc-sync, fix-workflows, implementation-*) are domain-specific. Should they be?

- [x] **A)** YES - keep them domain-specific (current design)
- [ ] **B)** NO - make them more generic
- [ ] **C)** Merge some agents (e.g., merge doc-sync into implementation-execution)

---

## Section 3: ARCHITECTURE.md Optimization

### Q8: Quality Attributes Duplication

Quality attributes appear in multiple sections (1.3, 2.5, 11.1). How should this be handled?

- [ ] **A)** Keep all duplications for context-specific emphasis
- [x] **B)** Consolidate to Section 11.1, use @propagate to others, cross-reference from 1.3/2.5
- [ ] **C)** Keep in 11.1 only, remove from other sections entirely

---

### Q9: CLI Patterns Duplication

CLI patterns appear in both Section 4.4.7 and Section 9.1. How should this be handled?

- [ ] **A)** Keep both - 4.4.7 for code structure, 9.1 for strategy
- [ ] **B)** Consolidate to 9.1, make 4.4.7 a cross-reference
- [x] **C)** Consolidate to 4.4.7, make 9.1 a cross-reference

---

### Q10: Port Assignments

Port assignments appear in Section 3.4 and Appendix B.1/B.2. How should this be handled?

- [ ] **A)** Keep detailed tables in both locations
- [ ] **B)** Summary in 3.4, detailed tables only in Appendix B
- [x] **C)** Consolidate in 3.4, remove Appendix B.1 and B.2; fix appendix B.# numbering

---

### Q11: Infrastructure Blocker Escalation

Infrastructure blocker escalation appears in Section 13.7 and partially in Section 2.5. How should this be handled?

- [x] **A)** Keep in both locations
- [ ] **B)** Single source in 13.7, cross-reference from 2.5
- [ ] **C)** Merge into Section 2.5 (quality strategy context)

---

### Q12: Missing Skills Documentation

ARCHITECTURE.md has no section about Copilot Skills. Should it?

- [x] **A)** YES - add a new section about the skills it supports and how they fit into the architecture/strategy; include reference link to VS Code docs but don't duplicate content from them unnecessarily, only in context of how the specific skills are organized, structured, and used in this project
- [ ] **B)** YES - add brief mention in Section 2.1 (Agent Orchestration)
- [ ] **C)** NO - skills are VS Code feature, not project architecture
- [ ] **D)** Reference VS Code docs, don't duplicate

---

### Q13: Agent/Skill/Instruction Decision Tree

Should ARCHITECTURE.md include guidance on when to use agents vs skills vs instructions?

- [ ] **A)** YES - add decision tree as new section
- [x] **B)** YES - add to Section 2.1 Agent Orchestration, but keep it concise and focused on high-level guidance rather than detailed decision tree
- [ ] **C)** NO - this is VS Code/Copilot specific, not architecture
- [ ] **D)** Add to separate doc (e.g., docs/COPILOT-CUSTOMIZATION.md)

---

## Section 4: doc-sync Agent Propagation

### Q14: Missing Cross-References

doc-sync.agent.md only references Section 2.5. Should it also reference:

- [x] **A)** Section 12.7 (Documentation Propagation Strategy) - YES
- [x] **B)** Section 11.4 (Documentation Standards) - YES
- [x] **C)** Section B.6 (Instruction File Reference) - YES

(Mark YES or NO after each)

---

### Q15: doc-sync ARCHITECTURE.md Propagation

Should doc-sync.agent.md have @source blocks from ARCHITECTURE.md?

- [x] **A)** YES - propagate relevant content like other instruction files
- [ ] **B)** NO - agents should only cross-reference, not duplicate content
- [ ] **C)** PARTIAL - propagate critical rules only (review passes)

---

## Section 5: Copilot Skills Migration

### Q16: Should We Create Skills?

Should this project create Copilot Skills (in .github/skills/)?

- [x] **A)** YES - migrate suitable content from instructions/agents; ADD THEM TO THE plan.md and tasks.md, and quizme-v2.md, for me to review and decide which ones i want to implement
- [ ] **B)** YES - create new skills for specialized tasks
- [ ] **C)** NO - current instructions/agents structure is sufficient
- [ ] **D)** LATER - defer until skills feature is more mature

---

### Q17: test-table-driven Skill

Should we create a skill for generating table-driven Go tests?

- [ ] **A)** YES - high reuse value, include test templates
- [ ] **B)** NO - keep in instructions (03-02.testing)
- [x] **C)** PROBABLY - wait for more skill examples

---

### Q18: compose-validation Skill

Should we create a skill for Docker Compose validation?

- [ ] **A)** YES - include validation scripts and compose templates
- [ ] **B)** NO - cicd lint-deployments already handles this
- [x] **C)** PROBABLY - skill wraps cicd lint-deployments - wait for more skill examples

---

### Q19: propagation-check Skill

Should we create a skill for checking @propagate/@source sync?

- [ ] **A)** YES - useful for doc maintenance
- [ ] **B)** NO - cicd lint-docs already handles this
- [x] **C)** PROBABLY - skill wraps cicd lint-docs - wait for more skill examples

---

### Q20: Instruction → Skill Migration Priority

Which instruction file content would benefit most from becoming a skill?

- [ ] **A)** 03-02.testing.instructions.md (test patterns + templates)
- [ ] **B)** 04-01.deployment.instructions.md (compose patterns + templates)
- [ ] **C)** 03-05.linting.instructions.md (lint workflows + scripts)
- [ ] **D)** None - keep all as instructions
- [x] **E)** DON'T KNOW YET - wait for more skill examples to evaluate which content is best suited for skills vs instructions

---

### Q21: Agent → Skill Migration

Should any agent functionality move to skills?

- [ ] **A)** fix-workflows → GitHub Actions debugging skill
- [ ] **B)** doc-sync → propagation checking skill
- [ ] **C)** Keep all agents as-is
- [ ] **D)** Create complementary skills, don't migrate agents
- [x] **E)** DON'T KNOW YET - wait for more skill examples to evaluate which content is best suited for skills vs instructions

---

## Section 6: Additional Considerations

### Q22: Document Line Count

ARCHITECTURE.md is 4,445 lines. Is this acceptable?

- [ ] **A)** YES - comprehensive is good, keep growing as needed
- [x] **B)** REDUCE - try to target <4,000 lines through deduplication, but don't over-condensing, don't sacrifice clarity or completeness or correctness or thoroughness or reliability or efficiency
- [ ] **C)** SPLIT - create separate docs for some sections
- [ ] **D)** RESTRUCTURE - appendices could be separate files

---

### Q23: Propagation Count

There are 34 @propagate markers. Is this manageable?

- [x] **A)** YES - automated validation (lint-docs) handles sync; automation is the only way to scale up propagation in a large doc like this, so we should embrace it rather than shy away from it; the benefits of propagation (consistency, single source of truth) outweigh the maintenance overhead
- [ ] **B)** TOO MANY - reduce to critical content only
- [ ] **C)** EXPAND - add more propagation for consistency

---

### Q24: Cross-Reference Density

Many sections end with "See Section X.Y for..." Should this pattern:

- [x] **A)** CONTINUE - helps navigation and avoids duplication
- [ ] **B)** REDUCE - too many cross-refs, increases reading friction
- [ ] **C)** STANDARDIZE - every section should have cross-refs in same format

---

### Q25: Implementation Priority

If resources are limited, what should be prioritized?

- [ ] **A)** Quality review passes rework (Phase 1) - affects all work
- [ ] **B)** Agent semantics (Phase 2) - affects agent behavior
- [ ] **C)** ARCHITECTURE.md optimization (Phase 3) - affects maintainability
- [ ] **D)** Skills migration (Phase 5) - new capability
- [x] **E)** Resources like time and tokens are NEVER a constraint - implement all changes as designed

---

## Summary

After completing your selections:
1. Save this file
2. Run implementation with marked choices
3. Unchosen options will be skipped

**Decision Legend:**
- `[x]` = Apply this change
- `[ ]` = Skip this option
