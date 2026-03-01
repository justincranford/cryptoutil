# Fixes v9 - Tasks

## Phase 1: Quality Review Passes Rework

### Task 1.1: Update ARCHITECTURE.md Section 2.5
- [ ] Rewrite review passes to check ALL 8 quality attributes per pass
- [ ] Add continuation rule (3 minimum, 5 maximum)
- [ ] Add criteria for when to continue beyond pass 3
- [ ] Ensure @propagate blocks updated for both targets

### Task 1.2: Update beast-mode.instructions.md
- [ ] Verify @source block matches new ARCHITECTURE.md content
- [ ] Add cross-reference to Section 11.1 for quality attributes

### Task 1.3: Update evidence-based.instructions.md
- [ ] Verify @source block matches new ARCHITECTURE.md content
- [ ] Ensure review passes are generic (not docs-specific)

### Task 1.4: Update beast-mode.agent.md
- [ ] Update Mandatory Review Passes section with new format
- [ ] Ensure examples are generic or clearly labeled

### Task 1.5: Update Other Agents
- [ ] doc-sync.agent.md - Update review passes section
- [ ] fix-workflows.agent.md - Update review passes section
- [ ] implementation-execution.agent.md - Update review passes section
- [ ] implementation-planning.agent.md - Update review passes section

---

## Phase 2: Agent Semantic Analysis

### Task 2.1: Review beast-mode.agent.md for Generic vs Specific
- [ ] Identify all domain-specific examples
- [ ] Option A: Make examples generic (preferred)
- [ ] Option B: Label examples as "e.g., for Go projects"
- [ ] Ensure continuous execution rules apply to ANY work type

### Task 2.2: Confirm Other Agents Are Correctly Scoped
- [ ] doc-sync.agent.md - Confirm documentation-specific (NO CHANGES)
- [ ] fix-workflows.agent.md - Confirm workflows-specific (NO CHANGES)
- [ ] implementation-execution.agent.md - Confirm plan-execution-specific (NO CHANGES)
- [ ] implementation-planning.agent.md - Confirm planning-specific (NO CHANGES)

---

## Phase 3: ARCHITECTURE.md Optimization

### Task 3.1: Identify Duplication
- [ ] Map all occurrences of quality attributes
- [ ] Map all occurrences of infrastructure blocker escalation
- [ ] Map all occurrences of CLI patterns
- [ ] Map all occurrences of port assignments

### Task 3.2: Consolidate Duplications
- [ ] Quality attributes → Single source in Section 11.1, propagate to others
- [ ] Infrastructure blocker → Keep in 13.7, cross-reference from 2.5
- [ ] CLI patterns → Consolidate to Section 9.1
- [ ] Port assignments → Summary in 3.4, details in Appendix B

### Task 3.3: Resolve Contradictions
- [ ] Review all coverage target mentions for consistency
- [ ] Update review pass count from "3" to "3-5"

### Task 3.4: Address Omissions
- [ ] Add skills documentation section (or reference to VS Code docs)
- [ ] Add agent/skill/instruction decision guidance
- [ ] Document review pass continuation criteria

---

## Phase 4: doc-sync Agent Propagation

### Task 4.1: Add Missing Cross-References
- [ ] Add reference to Section 12.7 Documentation Propagation Strategy
- [ ] Add reference to Section 11.4 Documentation Standards

### Task 4.2: Verify Existing References
- [ ] Confirm Section 2.5 reference is current

---

## Phase 5: Copilot Skills Migration

### Task 5.1: Document Decision Framework
- [ ] Add decision tree to ARCHITECTURE.md or separate doc
- [ ] Document when to use instructions vs agents vs skills

### Task 5.2: Identify Priority Skill Candidates
- [ ] High: test-table-driven skill (most reusable)
- [ ] Medium: compose-validation skill
- [ ] Medium: propagation-check skill
- [ ] Low: coverage-analysis skill

### Task 5.3: Create Initial Skill Structure (Optional)
- [ ] Create .github/skills/ directory
- [ ] Create SKILL.md template
- [ ] Document skill creation process

---

## Quality Review Passes (Meta)

### Pass 1: Full Quality Check
- [ ] Correctness: All changes functionally correct
- [ ] Completeness: All tasks addressed
- [ ] Thoroughness: Evidence-based validation
- [ ] Reliability: Quality gates enforced
- [ ] Efficiency: Optimized for maintainability
- [ ] Accuracy: Root causes addressed
- [ ] NO Time Pressure: Not rushed
- [ ] NO Premature Completion: Evidence exists

### Pass 2: Full Quality Check
- [ ] (Same 8 attributes verified again)

### Pass 3: Full Quality Check
- [ ] (Same 8 attributes verified again)
- [ ] Decision: Continue to pass 4? (if significant issues found)

### Pass 4: (If Needed)
- [ ] (Same 8 attributes verified again)

### Pass 5: (If Needed)
- [ ] (Same 8 attributes verified again)
- [ ] Diminishing returns reached - done
