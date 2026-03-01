# Fixes v9 - Tasks

**Status**: 0 of 48 tasks complete (0%)
**Created**: 2026-03-01
**Last Updated**: 2026-03-01 (quizme-v1 answers merged)

## Quality Mandate - MANDATORY

Every task MUST be completed with ALL 8 quality attributes verified each review pass. Resources (time, tokens) are NEVER a constraint. All issues are blockers — fix immediately, never defer.

---

## Phase 1: Quality Review Passes Rework

**Objective**: Rewrite review pass format so EACH pass checks ALL 8 quality attributes. Min 3 passes, max 5. Continue to pass 4 whenever pass 3 finds ANY issue.

### Task 1.1: Update ARCHITECTURE.md Section 2.5 (canonical source)
- [ ] Locate Section 2.5 "Mandatory Review Passes"
- [ ] Rewrite review passes: EACH pass checks all 8 attributes (not one per pass)
- [ ] List all 8 attributes in each pass description: Correctness, Completeness, Thoroughness, Reliability, Efficiency, Accuracy, NO Time Pressure, NO Premature Completion
- [ ] Add continuation rule: "If pass 3 finds ANY issue, continue to pass 4"
- [ ] Add pass 4 rule: "If pass 4 finds issues, continue to pass 5 (maximum)"
- [ ] Change "exactly 3 passes" phrasing to "minimum 3, maximum 5"
- [ ] Confirm @propagate markers are correctly placed for both targets
- **Files**: `docs/ARCHITECTURE.md` Section 2.5

### Task 1.2: Update beast-mode.instructions.md @source block
- [ ] Locate @source block referencing Section 2.5 mandatory-review-passes
- [ ] Update @source content to match new ARCHITECTURE.md Section 2.5 verbatim
- [ ] Verify @source/@@propagate chain is valid (lint-docs passes)
- **Files**: `.github/instructions/01-02.beast-mode.instructions.md`

### Task 1.3: Update evidence-based.instructions.md @source block
- [ ] Locate @source block referencing mandatory-review-passes in evidence-based instructions
- [ ] Update @source content to match new ARCHITECTURE.md Section 2.5 verbatim
- [ ] Verify generic phrasing (not docs-specific)
- **Files**: `.github/instructions/06-01.evidence-based.instructions.md`

### Task 1.4: Update beast-mode.agent.md review passes
- [ ] Locate "Mandatory Review Passes" section in beast-mode.agent.md
- [ ] Rewrite to match new format: each pass = all 8 attributes
- [ ] Add continuation rule (ANY issue in pass 3 → continue)
- [ ] Ensure generic phrasing for execution principles
- [ ] Keep "Quality Gate Commands (Go Projects)" label on Go-specific examples
- **Files**: `.github/agents/beast-mode.agent.md`

### Task 1.5: Update doc-sync.agent.md review passes
- [ ] Locate "Mandatory Review Passes" section
- [ ] Rewrite to match new format
- **Files**: `.github/agents/doc-sync.agent.md`

### Task 1.6: Update fix-workflows.agent.md review passes
- [ ] Locate "Mandatory Review Passes" section
- [ ] Rewrite to match new format
- **Files**: `.github/agents/fix-workflows.agent.md`

### Task 1.7: Update implementation-execution.agent.md review passes
- [ ] Locate "Mandatory Review Passes" section
- [ ] Rewrite to match new format
- **Files**: `.github/agents/implementation-execution.agent.md`

### Task 1.8: Update implementation-planning.agent.md review passes
- [ ] Locate "Mandatory Review Passes" section
- [ ] Rewrite to match new format
- **Files**: `.github/agents/implementation-planning.agent.md`

### Task 1.9: Sweep for stale "3 passes" mentions
- [ ] `grep -rn "3 review passes\|3 sequential\|exactly 3\|three review" .github/ docs/ARCHITECTURE.md`
- [ ] Update all stale mentions to "minimum 3, maximum 5" (or equivalent)

---

## Phase 2: Agent Semantic Analysis

**Objective**: beast-mode keeps Go-specific examples (labeled); confirm other agents are correctly scoped.

### Task 2.1: beast-mode.agent.md — label Go-specific Quality Gates
- [ ] Locate Quality Gate Commands section (go build, golangci-lint, go test lines)
- [ ] Add label/heading: "Quality Gate Commands (Go Projects)"
- [ ] Review all generic execution principles — ensure no Go-isms in generic sections
- [ ] Confirm continuous execution rules use generic language ("build command", "lint command") in principle text, Go examples only in labeled example blocks
- **Files**: `.github/agents/beast-mode.agent.md`

### Task 2.2: Confirm other agents correctly scoped (no changes needed)
- [ ] doc-sync.agent.md — read and confirm documentation scope is correct
- [ ] fix-workflows.agent.md — read and confirm GitHub Actions scope is correct
- [ ] implementation-execution.agent.md — read and confirm plan-execution scope is correct
- [ ] implementation-planning.agent.md — read and confirm planning scope is correct
- [ ] Document: "Confirmed — no scope changes required"

---

## Phase 3: ARCHITECTURE.md Optimization

**Objective**: Target <4,000 lines. Consolidate duplications, resolve contradictions, add skills and agent/skill/instruction guidance.

### Task 3.1: Consolidate quality attributes to Section 11.1 (Q8)
- [ ] Read Section 11.1 — confirm it has the canonical full quality attributes list
- [ ] Add @propagate marker in Section 11.1 around quality attributes block
- [ ] Read Section 1.3 Core Principles — replace full attribute list with @source block
- [ ] Read Section 2.5 Quality Strategy — replace full attribute list with @source block (or clear cross-reference if already handled by Phase 1)
- [ ] Run lint-docs to verify propagation chain
- **Expected**: ~40–60 lines saved

### Task 3.2: Consolidate CLI patterns to Section 4.4.7 (Q9 → 4.4.7 is canonical)
- [ ] Read Section 4.4.7 and Section 9.1 — identify full overlap
- [ ] Section 4.4.7 → keep all CLI pattern content (canonical)
- [ ] Section 9.1 CLI Patterns & Strategy → remove duplicated HOW content, keep WHY/strategy text, add cross-reference: "See Section 4.4.7 for CLI pattern implementation"
- [ ] Verify all "See Section 9.1" cross-references in instruction files → check if any need updating to 4.4.7
- **Expected**: ~30–50 lines saved

### Task 3.3: Port assignments — consolidate to 3.4, DELETE Appendix B.1 + B.2 (Q10)
- [ ] Read Section 3.4 — confirm it has service port table and database port table
- [ ] Read Appendix B.1 (Service Port Assignments) — compare to Section 3.4
- [ ] Read Appendix B.2 (Database Port Assignments) — compare to Section 3.4
- [ ] If Section 3.4 is missing any rows from Appendix B — ADD them to 3.4 first
- [ ] **DELETE** Appendix B.1 entirely
- [ ] **DELETE** Appendix B.2 entirely
- [ ] Resequence remaining Appendix B sections (old B.3 → B.1, old B.4 → B.2, etc.)
- [ ] Update TOC/index to reflect new B.# numbering
- [ ] Search for all cross-references to "Appendix B.1" and "Appendix B.2" — update to Section 3.4
- [ ] Search for "Appendix B.3", "B.4" etc. — update to new numbers
- **Expected**: ~80–120 lines saved

### Task 3.4: Verify infrastructure blocker escalation in both 13.7 and 2.5 (Q11)
- [ ] Read Section 13.7 — confirm full content present
- [ ] Read Section 2.5 — confirm partial content present
- [ ] Verify the two are consistent (not contradictory)
- [ ] NO deletion required — keep in both locations
- [ ] If inconsistent, update 2.5 to match 13.7

### Task 3.5: Add Copilot Skills section to ARCHITECTURE.md (Q12)
- [ ] Determine best section number (likely new subsection within Section 2.1 or new Section 2.X)
- [ ] Write section content:
  - What VS Code Copilot Skills are (brief, not duplicating VS Code docs)
  - Reference link to official VS Code Copilot Customization docs
  - How skills are organized in THIS project: `.github/skills/` directory
  - File naming conventions: `SKILL.md` files
  - Relationship to agents (on-demand, specialized) and instructions (always loaded)
  - Where skills are catalogued/inventoried in this doc
- [ ] Add to TOC
- **Adds**: ~30–50 lines (net cost, offset by Appendix B savings)

### Task 3.6: Add agent/skill/instruction matrix to Section 2.1 (Q13)
- [ ] Read Section 2.1 Agent Orchestration Strategy
- [ ] Add concise 4-row decision matrix table: Instructions / Agents / Skills / When to Use
- [ ] Keep it high-level — NOT a detailed decision tree
- [ ] Cross-reference to new skills section (Task 3.5)
- **Adds**: ~10–15 lines

### Task 3.7: Review pass count sweep (Q2/Q3 — any missed occurrences)
- [ ] `grep -n "3 passes\|3 review\|three pass\|minimum 3\|exactly 3" docs/ARCHITECTURE.md`
- [ ] Update any occurrences not already handled by Task 1.1

### Task 3.8: Line count verification
- [ ] `wc -l docs/ARCHITECTURE.md` — record before and after
- [ ] Target: <4,000 lines
- [ ] If not reached: identify additional prose deduplication opportunities (no semantic loss)

---

## Phase 4: doc-sync Agent Propagation

**Objective**: Add cross-references and @source propagation for sections 12.7, 11.4, and B.6.

### Task 4.1: Add Section 12.7 documentation propagation strategy to doc-sync
- [ ] Read `docs/ARCHITECTURE.md` Section 12.7 Documentation Propagation Strategy
- [ ] Identify @propagate markers in Section 12.7
- [ ] Add @source block for Section 12.7 content into `doc-sync.agent.md`
- [ ] Add cross-reference: "See ARCHITECTURE.md Section 12.7..."
- **Files**: `.github/agents/doc-sync.agent.md`

### Task 4.2: Add Section 11.4 documentation standards to doc-sync
- [ ] Read `docs/ARCHITECTURE.md` Section 11.4 Documentation Standards
- [ ] Identify @propagate markers in Section 11.4
- [ ] Add @source block for Section 11.4 content into `doc-sync.agent.md`
- [ ] Add cross-reference: "See ARCHITECTURE.md Section 11.4..."
- **Files**: `.github/agents/doc-sync.agent.md`

### Task 4.3: Add Appendix B.6 instruction file reference to doc-sync
- [ ] Read `docs/ARCHITECTURE.md` Appendix B.6 (Instruction File Reference table)
- [ ] Note: B.6 numbering may change after Phase 3.3 (Appendix B resequencing) — resolve before this task
- [ ] Add @source block or cross-reference for B.6 into `doc-sync.agent.md`
- **Files**: `.github/agents/doc-sync.agent.md`
- **Dependency**: Complete Task 3.3 first (Appendix B renumbering)

### Task 4.4: Verify existing Section 2.5 reference in doc-sync
- [ ] Confirm Section 2.5 cross-reference is current after Phase 1 changes
- [ ] Update if stale

---

## Phase 5: Copilot Skills Planning

**Objective**: Each skill candidate has YES/NO/DEFER decision from quizme-v2.md. Only implement skills with YES decision.

### Task 5.0: Review quizme-v2.md answers (blocking all 5.x tasks below)
- [ ] User answers quizme-v2.md
- [ ] Merge answers into this tasks.md (mark 5.x tasks as applicable)

### Task 5.1: Group A — Test Generation Skills (per quizme-v2 decision)
- [ ] `test-table-driven` skill — implement if YES in quizme-v2 Q1
- [ ] `test-fuzz-gen` skill — implement if YES in quizme-v2 Q2
- [ ] `test-benchmark-gen` skill — implement if YES in quizme-v2 Q3

### Task 5.2: Group B — Infrastructure/Deployment Skills (per quizme-v2 decision)
- [ ] `compose-validator` skill — implement if YES in quizme-v2 Q4
- [ ] `migration-create` skill — implement if YES in quizme-v2 Q5
- [ ] `service-scaffold` skill — implement if YES in quizme-v2 Q6

### Task 5.3: Group C — Code Quality Skills (per quizme-v2 decision)
- [ ] `coverage-analysis` skill — implement if YES in quizme-v2 Q7
- [ ] `fips-audit` skill — implement if YES in quizme-v2 Q8

### Task 5.4: Group D — Documentation Skills (per quizme-v2 decision)
- [ ] `propagation-check` skill — implement if YES in quizme-v2 Q9
- [ ] `openapi-codegen` skill — implement if YES in quizme-v2 Q10

### Task 5.5: Group E — Scaffolding Skills (per quizme-v2 decision)
- [ ] `agent-scaffold` skill — implement if YES in quizme-v2 Q11
- [ ] `instruction-scaffold` skill — implement if YES in quizme-v2 Q12

### Task 5.6: Create .github/skills/ infrastructure (if any Group A–E task = YES)
- [ ] Create `.github/skills/` directory
- [ ] Create `SKILL.md` template
- [ ] Add `.github/skills/` documentation to ARCHITECTURE.md skills section (Task 3.5)

---

## Phase 6: Validation

**Objective**: All quality gates pass after all phases complete.

### Task 6.1: Build validation
- [ ] `go build ./...` — clean
- [ ] `go build -tags e2e,integration ./...` — clean

### Task 6.2: Lint validation
- [ ] `golangci-lint run --fix` — 0 issues
- [ ] `golangci-lint run --build-tags e2e,integration --fix` — 0 issues

### Task 6.3: Test validation
- [ ] `go test ./... -shuffle=on -count=1` — all pass, exit 0

### Task 6.4: Documentation validation
- [ ] `go run ./cmd/cicd lint-docs` — all propagation verified
- [ ] `go run ./cmd/cicd lint-text` — UTF-8 clean

### Task 6.5: Review passes (1 of minimum 3)

**Pass 1** — Check ALL 8 quality attributes:
- [ ] Correctness: All changes in phases 1–5 are accurate
- [ ] Completeness: All 48 tasks addressed
- [ ] Thoroughness: Evidence from lint-docs, build, lint, test
- [ ] Reliability: All quality gates pass
- [ ] Efficiency: No unnecessary changes made
- [ ] Accuracy: Root causes addressed (not just symptoms)
- [ ] NO Time Pressure: Not rushed
- [ ] NO Premature Completion: Evidence exists for all tasks

**Pass 2** — Check ALL 8 quality attributes:
- [ ] (Same 8 attributes; fresh eyes on all changes)

**Pass 3** — Check ALL 8 quality attributes:
- [ ] (Same 8 attributes)
- [ ] Decision: ANY issues found? → Continue to Pass 4

**Pass 4** (if ANY issue found in Pass 3):
- [ ] (Same 8 attributes)
- [ ] Decision: Issues resolved? If not → Pass 5

**Pass 5** (maximum — if Pass 4 still finds issues):
- [ ] (Same 8 attributes)
- [ ] Diminishing returns reached → complete

### Task 6.6: Git commit
- [ ] `git add -A`
- [ ] `git commit -m "feat: quality review passes rework, architecture optimization, skills planning"`
- [ ] `git push`
