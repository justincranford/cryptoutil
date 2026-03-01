# Fixes v9 - Quality Review Passes, Agent Semantics, ARCHITECTURE.md Optimization, Skills Migration

**Status**: Planning → Ready for Execution
**Created**: 2026-03-01
**Last Updated**: 2026-03-01 (quizme-v1 answers merged)

## Quality Mandate - MANDATORY

- ✅ **Correctness**: ALL changes must be accurate and semantically correct
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: lint-docs, build, lint, tests must all pass
- ✅ **Efficiency**: Optimized for clarity and maintainability, NOT speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark tasks complete without objective evidence

**ALL issues are blockers. Resources (time, tokens) are NEVER a constraint.**

---

## Executive Summary

Six phases of improvement, all to be implemented:

1. **Phase 1: Quality Review Passes** — Rewrite review passes so EACH pass checks ALL 8 quality attributes; min 3 passes, max 5; continue to pass 4 whenever pass 3 finds ANY issue
2. **Phase 2: Agent Semantics** — beast-mode stays generic for execution but KEEPS Go-specific examples (labeled); other agents confirmed domain-specific
3. **Phase 3: ARCHITECTURE.md Optimization** — Consolidate duplications, resolve contradictions, fill omissions; target <4,000 lines
4. **Phase 4: doc-sync Propagation** — Add missing cross-references and propagate content from 3 ARCHITECTURE.md sections
5. **Phase 5: Copilot Skills Planning** — Inventory all skill candidates; each candidate gated by quizme-v2 decision
6. **Phase 6: Validation** — lint-docs, build, lint, tests; 3–5 review passes

---

## Phase 1: Quality Review Passes Rework

### Current State (PROBLEM)

Section 2.5 of ARCHITECTURE.md and all @propagate/@source targets define review passes as:
- **Pass 1 — Completeness**: Only checks completeness
- **Pass 2 — Correctness**: Only checks correctness
- **Pass 3 — Quality**: Only checks coverage/mutation

### Target State (SOLUTION)

Each review pass MUST check ALL 8 quality attributes:

1. ✅ **Correctness** — Code is functionally correct, no regressions
2. ✅ **Completeness** — All requested items addressed, no steps skipped
3. ✅ **Thoroughness** — Evidence-based validation, all edge cases
4. ✅ **Reliability** — Quality gates enforced (build, lint, test, coverage)
5. ✅ **Efficiency** — Optimized for maintainability, not speed
6. ✅ **Accuracy** — Root cause addressed, not just symptoms
7. ❌ **NO Time Pressure** — Work not rushed, all checks performed
8. ❌ **NO Premature Completion** — Evidence required before marking complete

### Review Pass Rules (Decided)

- **Minimum**: 3 review passes
- **Maximum**: 5 review passes
- **Continuation rule**: If pass 3 finds **ANY issue** (not just "significant"), continue to pass 4
- If pass 4 still finds issues, continue to pass 5
- **Scope**: ALL work types — code, docs, config, tests, infrastructure, deployments

### Files to Update

1. `docs/ARCHITECTURE.md` Section 2.5 — canonical source of truth
2. `.github/instructions/01-02.beast-mode.instructions.md` — @source block
3. `.github/instructions/06-01.evidence-based.instructions.md` — @source block
4. `.github/agents/beast-mode.agent.md` — Mandatory Review Passes section
5. `.github/agents/doc-sync.agent.md` — Mandatory Review Passes section
6. `.github/agents/fix-workflows.agent.md` — Mandatory Review Passes section
7. `.github/agents/implementation-execution.agent.md` — Mandatory Review Passes section
8. `.github/agents/implementation-planning.agent.md` — Mandatory Review Passes section

---

## Phase 2: Agent Semantic Analysis

### Agent Purpose Matrix

| Agent | Scope | Generic? | Decision |
|-------|-------|----------|----------|
| beast-mode | ANY work type | ✅ Generic execution | KEEP structure; KEEP Go examples labeled |
| doc-sync | Documentation | ❌ Domain-specific | NO CHANGES to scope |
| fix-workflows | GitHub Actions | ❌ Domain-specific | NO CHANGES to scope |
| implementation-execution | Plan execution | ❌ Domain-specific | NO CHANGES to scope |
| implementation-planning | Plan creation | ❌ Domain-specific | NO CHANGES to scope |

### beast-mode Analysis (Q5 + Q6 Answered)

**Decision**: Dual structure — generic continuous execution principles + Go-specific Quality Gate examples (clearly labeled as "Go project examples").

**Rationale**: This is primarily a Go project. Generic execution principles serve ALL work types. Go-specific examples are the practical Quality Gates for this project's implementation work. The two must coexist — removing Go examples would lose actionable guidance.

**Changes needed**:
- Review pass section → update to new all-8-attributes format (Phase 1 handles this)
- Go-specific Quality Gate examples → KEEP, add label "Quality Gate Commands (Go Projects)"
- All execution/continuation principles → KEEP generic phrasing ("build", "lint", "test")
- Verify no Go-isms leak into generic execution principles section

### Other Agents

All confirmed as correctly domain-specific — NO CHANGES to scope or purpose.

---

## Phase 3: ARCHITECTURE.md Optimization

### Constraints

- **Target size**: <4,000 lines (from 4,445 lines — reduce by ~450 lines minimum)
- **Rule**: Do NOT over-condense; do NOT sacrifice clarity, completeness, correctness, or thoroughness
- **Propagate markers**: Embrace automation — lint-docs validates sync. MORE propagation is better, not worse.
- **Cross-references**: Continue "See Section X.Y for..." pattern — helps navigation, reduces duplication

### 3.1 Quality Attributes Duplication (Q8 → B: Consolidate to 11.1)

**Problem**: Quality attributes table repeated verbatim in 1.3, 2.5, 11.1.

**Solution**:
- Section 11.1 → canonical source, full list with @propagate markers
- Section 1.3 → @source block (or brief mention + cross-reference to 11.1)
- Section 2.5 → @source block (or brief mention + cross-reference to 11.1)

**Expected savings**: ~40–60 lines

### 3.2 CLI Patterns Duplication (Q9 → C: Consolidate to 4.4.7)

**Problem**: CLI patterns appear in both Section 4.4.7 (code structure) and Section 9.1 (CLI patterns & strategy).

**Solution**:
- Section **4.4.7** → canonical full CLI patterns source
- Section 9.1 → cross-reference only ("See Section 4.4.7 for CLI patterns")

**Rationale**: 4.4.7 is the Go project structure section — CLI patterns naturally belong under code structure. Section 9.1 is architectural strategy; it should describe WHY not HOW.

**Expected savings**: ~30–50 lines

### 3.3 Port Assignments (Q10 → C: Consolidate to 3.4, DELETE Appendix B.1 + B.2)

**Problem**: Port assignment tables duplicated in Section 3.4 AND Appendix B.1/B.2.

**Solution**:
- Section 3.4 → canonical full port assignment tables (both service and database ports)
- Appendix B.1 → **DELETE** (Service Port Assignments)
- Appendix B.2 → **DELETE** (Database Port Assignments)
- Remaining Appendix B sections → **RENUMBER** (e.g., old B.3 becomes B.1, etc.)
- Index/TOC → update references

**Expected savings**: ~80–120 lines (two full port tables removed)

### 3.4 Infrastructure Blocker Escalation (Q11 → A: Keep in BOTH)

**Decision**: Keep content in BOTH Section 13.7 (Infrastructure Blocker Escalation) and Section 2.5 (Quality Strategy).

**Rationale**: 13.7 is the authoritative/detailed source; 2.5 is where quality gates are enforced — both readers need to see it. No change needed.

### 3.5 Copilot Skills Section (Q12 → A: Add new section)

**Problem**: ARCHITECTURE.md has no section about Copilot Skills — a new VS Code feature used in this project.

**Solution**: Add **Section 2.X Copilot Skills Architecture** (or renumber as fits):
- Brief description of VS Code Copilot Skills and their role
- How skills are organized in this project (`.github/skills/`)
- Naming conventions, structure conventions
- Reference link to VS Code docs (do NOT duplicate VS Code docs content)
- Relationship to agents and instructions

**Note**: Content specific to THIS PROJECT's skills patterns — not a duplication of VS Code official docs.

### 3.6 Agent/Skill/Instruction Guidance (Q13 → B: Add to 2.1)

**Problem**: No guidance on when to use agents vs skills vs instructions.

**Solution**: Add **concise high-level decision matrix** to Section 2.1 Agent Orchestration:
- 4-row table: Instructions / Agents / Skills — scope, trigger, best for
- NOT a detailed decision tree — keep concise
- Cross-reference to new skills section (3.5 above) for details

### 3.7 Review Pass Count Consistency

Update all occurrences of "exactly 3" or "minimum 3" review passes to reflect new "3–5 with continuation on any issue in pass 3" rule. (Phase 1 handles the canonical source; Phase 3 sweeps for all other mentions.)

### Expected Total Savings

| Change | Est. Lines Saved |
|--------|-----------------|
| Quality attributes consolidation | 40–60 |
| CLI patterns consolidation | 30–50 |
| Appendix B.1 + B.2 deletion | 80–120 |
| New skills section added | −30 to −50 (adds content) |
| Agent guidance in 2.1 | −10 to −15 (adds content) |
| **Total net** | **140–175 lines** |

**Result**: 4,445 − 150 ≈ 4,295 lines. Further reduction from prose deduplication should reach <4,000 target.

---

## Phase 4: doc-sync Agent Propagation

### Current State

doc-sync.agent.md has only ONE cross-reference to ARCHITECTURE.md:
- ✅ Section 2.5 Mandatory Review Passes

### Missing (Q14 → All three: A, B, C)

| Section | Title | Why Needed |
|---------|-------|-----------|
| Section 12.7 | Documentation Propagation Strategy | Core to doc-sync's purpose — defines @propagate/@source system |
| Section 11.4 | Documentation Standards | Documentation quality requirements doc-sync enforces |
| Appendix B.6 | Instruction File Reference | Lists all instruction files doc-sync is responsible for |

### Action (Q15 → A: Propagate like other instruction files)

- Add @source blocks for all three sections (not just cross-references)
- Relevant content should be propagated INTO doc-sync just like 01-02.beast-mode.instructions.md propagates content

---

## Phase 5: Copilot Skills Planning

### Status: Candidates Listed — Individual Decisions in quizme-v2.md

The user confirmed: YES create skills, ADD all candidates to plan.md/tasks.md/quizme-v2.md for per-candidate review. Q17–Q21 answered "PROBABLY / WAIT for more skill examples" — quizme-v2.md presents each candidate concretely for final decision.

### Skill Candidates Inventory

#### Group A: Test Generation Skills

| Skill | Wraps/Extends | Source Content |
|-------|--------------|----------------|
| `test-table-driven` | New | `03-02.testing.instructions.md` — table-driven test pattern |
| `test-fuzz-gen` | New | `03-02.testing.instructions.md` — fuzz test pattern |
| `test-benchmark-gen` | New | `03-02.testing.instructions.md` — benchmark test pattern |

#### Group B: Infrastructure/Deployment Skills

| Skill | Wraps/Extends | Source Content |
|-------|--------------|----------------|
| `compose-validator` | Wraps `cicd lint-deployments` | `04-01.deployment.instructions.md` |
| `migration-create` | New | `03-04.data-infrastructure.instructions.md` |
| `service-scaffold` | New | `02-01.architecture.instructions.md` + builder pattern |

#### Group C: Code Quality Skills

| Skill | Wraps/Extends | Source Content |
|-------|--------------|----------------|
| `coverage-analysis` | New | `03-02.testing.instructions.md` — coverage targets |
| `fips-audit` | New | `02-05.security.instructions.md` — FIPS 140-3 check |

#### Group D: Documentation Skills

| Skill | Wraps/Extends | Source Content |
|-------|--------------|----------------|
| `propagation-check` | Wraps `cicd lint-docs` | `doc-sync.agent.md` partial |
| `openapi-codegen` | New | `02-04.openapi.instructions.md` — codegen config |

#### Group E: Agent/Instruction Scaffolding Skills

| Skill | Wraps/Extends | Source Content |
|-------|--------------|----------------|
| `agent-scaffold` | New | `06-02.agent-format.instructions.md` |
| `instruction-scaffold` | New | `06-02.agent-format.instructions.md` |

### Per-Candidate Final Decisions

All candidates are in **quizme-v2.md** Q1–Q12 for per-candidate YES/NO/DEFER decisions.

---

## Phase 6: Validation

### Quality Gates

1. `go build ./...` — clean
2. `go build -tags e2e,integration ./...` — clean
3. `golangci-lint run --fix && golangci-lint run` — 0 issues
4. `go test ./... -shuffle=on -count=1` — all pass
5. `go run ./cmd/cicd lint-docs` — propagation verified
6. `go run ./cmd/cicd lint-text` — UTF-8 clean

### Review Passes (New Format)

All 8 attributes checked per pass, 3–5 passes, continue to pass 4 if pass 3 finds ANY issue.

---

## Decisions Log (from quizme-v1.md)

| Q | Decision | Impact |
|---|----------|--------|
| Q1 | Each pass checks ALL 8 attributes | Phase 1 scope |
| Q2 | Min 3, max 5 passes | Phase 1 rules |
| Q3 | Continue on ANY issue in pass 3 | Phase 1 continuation rule |
| Q4 | ALL work types | Phase 1 scope |
| Q5 | KEEP Go-specific examples | Phase 2 beast-mode |
| Q6 | Dual: generic execution + Go QG examples | Phase 2 beast-mode |
| Q7 | Keep other agents domain-specific | Phase 2 no-change |
| Q8 | Consolidate quality attributes to 11.1, propagate | Phase 3.1 |
| Q9 | Consolidate CLI patterns to **4.4.7** (not 9.1) | Phase 3.2 |
| Q10 | Consolidate ports to **3.4**, DELETE Appendix B.1+B.2, resequence | Phase 3.3 |
| Q11 | Keep infra blocker in BOTH 13.7 and 2.5 | Phase 3.4 no-change |
| Q12 | Add new ARCHITECTURE.md section for skills | Phase 3.5 |
| Q13 | Add concise matrix to Section 2.1 | Phase 3.6 |
| Q14 | doc-sync add: 12.7, 11.4, B.6 | Phase 4 |
| Q15 | Propagate content like other instruction files | Phase 4 |
| Q16 | YES skills, individual decisions via quizme-v2 | Phase 5 |
| Q17–Q21 | PROBABLY/WAIT — in quizme-v2 | Phase 5 |
| Q22 | Target <4,000 lines | Phase 3 constraint |
| Q23 | Embrace propagation, automation scales it | All phases |
| Q24 | Continue cross-references | All phases |
| Q25 | Implement ALL — time/tokens NEVER a constraint | All phases |
