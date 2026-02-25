# Architecture Strategy Review: Markdown Link Propagation

**Purpose**: Evaluate the effectiveness of the ARCHITECTURE.md → Copilot Instructions → Agents
→ CICD linters → code → tests → configs → deployments → workflows propagation strategy.

## 1. The Strategy

ARCHITECTURE.md (3,819 lines, 133 sections) serves as the **single source of truth** for all
architectural decisions. Changes flow outward through a chain:

```
docs/ARCHITECTURE.md (source of truth)
    ├── .github/instructions/*.instructions.md (18 files, compressed summaries)
    │       └── via "See [ARCHITECTURE.md Section X.Y]" cross-references (148 total)
    ├── .github/agents/*.agent.md (5 files, self-contained execution)
    │       └── via direct ARCHITECTURE.md references (15 total)
    ├── internal/cmd/cicd/ (linter code)
    │       └── via comment references to ARCHITECTURE.md sections (10 files)
    ├── Go source code (implementations)
    │       └── via patterns described in ARCHITECTURE.md
    ├── deployments/ and configs/ (compose, YAML)
    │       └── via deployment architecture (Section 12)
    └── .github/workflows/*.yml (CI/CD pipelines)
            └── via CI/CD section (Section 9.7)
```

The propagation model from Section 12.7 describes this as "chunk-based verbatim copying of
semantic units" where instruction files contain compressed summaries with cross-references
back to ARCHITECTURE.md.

## 2. Quantitative Analysis

### 2.1 Coverage Metrics

| Metric | Value | Assessment |
|--------|-------|------------|
| Total ARCHITECTURE.md sections (## + ###) | 133 | - |
| Sections referenced by instructions/agents | 48 (36%) | LOW |
| Sections with NO cross-references | 85 (64%) | HIGH |
| Cross-references in instructions | 148 | Dense |
| Cross-references in agents | 15 | Sparse |
| Go/YAML files referencing ARCHITECTURE.md | 10 | Minimal |
| Agent with zero references | implementation-planning (0) | GAP |

### 2.2 Reference Distribution (Instruction Files)

| Instruction File | References | Coverage |
|-----------------|------------|----------|
| 04-01.deployment | 18 | Heaviest — covers Sections 9.7, 10.3, 12.* |
| 03-02.testing | 17 | Dense — covers Section 10.* |
| 02-05.security | 13 | Broad — covers Sections 6.*, 12.3, 12.6 |
| 02-01.architecture | 13 | Broad — covers Sections 2-5 |
| 03-04.data-infrastructure | 11 | Focused — covers Sections 5.2, 7.* |
| 03-03.golang | 9 | Focused — covers Sections 2.4, 4.4, 11.1 |
| 03-05.linting | 9 | Focused — covers Sections 2.5, 9.9, 11.3 |
| 02-04.openapi | 9 | Focused — covers Section 8.* |
| 01-02.beast-mode | 7 | Behavioral — covers Sections 1.3, 11.2, 13 |
| 02-03.observability | 7 | Focused — covers Section 9.4 |
| 02-02.versions | 6 | Focused — covers Sections 2.2, 9.7, 11.3 |
| 02-06.authn | 6 | Focused — covers Section 6.9 |
| 05-02.git | 5 | Focused — covers Sections 11.2, 13.2 |
| 06-01.evidence-based | 5 | Focused — covers Sections 2.4, 11.2 |
| 06-02.agent-format | 4 | Focused — covers Section 2.1 |
| 03-01.coding | 4 | Focused — covers Sections 11.*, 13.1 |
| 01-01.terminology | 3 | Light — covers Sections 1.3, 2.2 |
| 05-01.cross-platform | 2 | Light — covers Sections 9.7, 10.2 |

### 2.3 Unreferenced Sections (Notable Gaps)

**Entire sections with zero downstream references**:

| Section | Lines | Impact |
|---------|-------|--------|
| 1. Executive Summary | 69-180 | Low (context only) |
| 3. Product Suite Architecture | 367-648 | Medium (service catalog) |
| 4. System Architecture (partial) | 648-914 | Medium (container arch) |
| 6.6 JOSE Architecture | 1253-1297 | High (JOSE-specific patterns) |
| 6.8 MFA Strategy | 1297-1373 | Medium (MFA patterns) |
| 7.1-7.5 Data Architecture (partial) | 1373-1500 | High (DB schema, isolation, migration) |
| 8.3 API Versioning | 1572-1606 | Medium (N-1 compat) |
| 8.5 API Security | 1606-1628 | High (API security patterns) |
| 9.1-9.3, 9.5-9.6, 9.8 Infrastructure | 1630-1873 | High (CLI, config, orchestration) |
| 10.4 E2E Testing | 2097-2302 | High (E2E patterns) |
| 10.6-10.9 Load/Race/Property Testing | 2302-2400 | Medium (advanced testing) |
| 14. Operational Excellence | 3570-3600 | Medium (monitoring, DR) |
| Appendix A-C | 3603-3800 | Low (reference tables) |

## 3. Effectiveness Assessment

### 3.1 What the Strategy Does Well

**1. Single Source of Truth is Clear**
ARCHITECTURE.md is unambiguously THE reference. Every instruction file points back to it.
There's no confusion about where canonical information lives.

**2. Compressed Summaries are Practical**
Instruction files don't duplicate ARCHITECTURE.md content. They provide concise, actionable
directives with "See [Section X.Y]" for depth. This is the right compression model.

**3. Cross-References are Machine-Parseable**
The `See [ARCHITECTURE.md Section X.Y](...)` format is grep-friendly. A CI/CD linter
could validate all references resolve to existing sections.

**4. Instruction Files are Well-Organized**
The hierarchical numbering (01-terminology through 06-evidence) creates natural reading
order. Each file has a clear domain.

**5. Linters Encode Architecture**
10 Go/YAML files reference ARCHITECTURE.md in comments, linking code enforcement directly
to architectural decisions. This is a strong pattern.

### 3.2 What the Strategy Does Poorly

**1. No Propagation Trigger Mechanism**
When ARCHITECTURE.md changes, NOTHING automatically triggers updates to instruction files.
The mapping in Section 12.7 has only 5 entries for 148 actual references. There's no CI/CD
check that says "Section 10.2 changed — update 03-02.testing.instructions.md."

**2. 64% of Sections Have No Downstream Consumers**
85 of 133 sections exist only in ARCHITECTURE.md. They're not compressed into any
instruction file, referenced by any agent, or enforced by any linter. These are
"documentation-only" — valuable for humans but invisible to automated tools.

**3. Agents Are Mostly Disconnected**
Only 15 references across 5 agents vs 148 across instructions. The planning agent
(implementation-planning.agent.md) has ZERO references, meaning plans are created without
direct architectural guidance. This undermines the "source of truth" claim.

**4. No Validation of Reference Freshness**
Cross-references like `See [ARCHITECTURE.md Section 10.2]` point to section NUMBERS, not
content hashes. If Section 10.2 content changes substantially but the number stays the
same, downstream files silently become stale. There's no staleness detection.

**5. Instruction File → Code Gap is Wide**
Instruction files describe patterns. Code implements them. But there's no formal link
between "Section 10.2.1 Table-Driven Test Pattern" in the instruction file and actual
test code. Compliance is enforced by linters (tparallel, thelper) for SOME patterns,
but many patterns (seam injection, coverage analysis) have no enforcement.

**6. Bidirectional Propagation is Missing**
The model is strictly top-down: ARCHITECTURE.md → instructions → agents → code.
But lessons learned flow bottom-up: code experience → post-mortems → ARCHITECTURE.md.
There's no systematic "experience feedback" mechanism. The lessons.md in this archive
is an ad-hoc version of what should be a formal pattern.

### 3.3 Silent Propagation Failures

These are documented requirements in ARCHITECTURE.md that have NO enforcement downstream:

| ARCHITECTURE.md Requirement | Expected Enforcement | Actual Status |
|---------------------------|---------------------|---------------|
| Section 6.6: JOSE Architecture | jose instruction file | No instruction file covers JOSE arch |
| Section 7.1: DB Schema Patterns | data-infrastructure instructions | Instructions cover GORM but not schema patterns |
| Section 8.3: API Versioning (N-1 compat) | openapi instructions | Instructions mention REST but not versioning |
| Section 8.5: API Security | security instructions | Security instructions cover crypto, not API security |
| Section 9.1: CLI Patterns | golang instructions | Instructions mention CLI but not full patterns from 9.1 |
| Section 9.2: Config Architecture | data-infrastructure instructions | Config YAML mentioned but not full architecture |
| Section 10.4: E2E Testing Strategy | testing instructions | Testing instructions mention E2E timing but not strategy |
| Section 14: Operational Excellence | (none) | No instruction file covers ops at all |

## 4. Propagation Chain Analysis

### 4.1 Complete Propagation Chain

When a change is made to ARCHITECTURE.md, the full propagation chain is:

```
1. ARCHITECTURE.md change (source)
    │
    ├─► 2. Instruction files (.github/instructions/)
    │       └─ 18 files with 148 "See [ARCHITECTURE.md]" references
    │       └─ TRIGGER: Manual review of affected references
    │       └─ VERIFICATION: None (no CI/CD check)
    │
    ├─► 3. Agent files (.github/agents/)
    │       └─ 5 files with 15 references
    │       └─ TRIGGER: None (agents rarely updated)
    │       └─ VERIFICATION: None
    │
    ├─► 4. CICD linters (internal/cmd/cicd/)
    │       └─ 10 files with ARCHITECTURE.md comments
    │       └─ TRIGGER: When enforced pattern changes
    │       └─ VERIFICATION: Linter tests
    │
    ├─► 5. Source code
    │       └─ Implements patterns from ARCHITECTURE.md
    │       └─ TRIGGER: Developer awareness
    │       └─ VERIFICATION: Linters (partial), tests, code review
    │
    ├─► 6. Deployment configs
    │       └─ deployments/, configs/
    │       └─ TRIGGER: Deployment architecture changes
    │       └─ VERIFICATION: cicd lint-deployments validate-all (62 validators)
    │
    ├─► 7. CI/CD workflows
    │       └─ .github/workflows/
    │       └─ TRIGGER: Workflow architecture changes
    │       └─ VERIFICATION: Workflow-specific linters
    │
    └─► 8. Documentation companions
            └─ ARCHITECTURE-INDEX.md (278 lines)
            └─ TRIGGER: Section adds/removes/reorganizes
            └─ VERIFICATION: None (manual sync noted in header)
```

### 4.2 Propagation Gaps (What SHOULD Be Triggered But ISN'T)

| Change | Expected Propagation | Current Status |
|--------|---------------------|----------------|
| New section added to ARCHITECTURE.md | ARCHITECTURE-INDEX.md update | Manual only |
| Section content changed | Instruction file update | No trigger mechanism |
| Section renumbered | All cross-references update | Manual grep-and-fix |
| New pattern added | Instruction file + possibly linter | No trigger |
| Pattern removed/deprecated | Instruction file + linter + code | No trigger |
| Port assignment changed | deployment instructions + compose files | Validators check compose, NOT instructions |
| Quality target changed | testing + evidence + beast-mode instructions | Manual only |
| New agent created | copilot-instructions.md table update | Manual only |
| Agent lacks ARCHITECTURE.md refs | (nothing) | No validation exists |

## 5. Recommendations

### 5.1 Short-Term (Low Effort)

**R1: Complete the Section 12.7 Mapping Table**
Replace the 5-entry table with the 14-row section-level mapping from ARCH-SUGGESTIONS.md
Suggestion 4. This is pure documentation — no code changes.

**R2: Add ARCHITECTURE.md References to implementation-planning.agent.md**
The planning agent needs architectural context. Add 3-5 references to key sections
(testing, quality, coding, deployment). One file change.

**R3: Create a Reference Validation Script**
A simple script that extracts all `See [ARCHITECTURE.md Section X.Y]` references,
resolves them against actual ARCHITECTURE.md section headers, and reports broken links.
Could be a pre-commit hook or cicd subcommand.

### 5.2 Medium-Term (Moderate Effort)

**R4: Add `cicd lint-propagation` Linter**
A CICD subcommand that:
1. Extracts all ARCHITECTURE.md section headers and their line ranges.
2. Cross-references against instruction files and agents.
3. Reports sections with zero downstream references (currently 85/133).
4. Reports instruction files referencing non-existent sections.
5. Optionally warns when ARCHITECTURE.md was modified more recently than its downstream files.

**R5: Add Instruction File Coverage Targets**
Track what percentage of ARCHITECTURE.md sections have downstream instruction file coverage.
Current: 36%. Target: 60% (cover all "High impact" sections from table in Section 3.3).
Accept that Appendices and Executive Summary don't need instruction file coverage.

### 5.3 Long-Term (Higher Effort)

**R6: Content Hash-Based Staleness Detection**
For each instruction file cross-reference, store the SHA-256 hash of the referenced
ARCHITECTURE.md section content at the time the instruction file was last updated.
A CI/CD check compares current hashes against stored hashes and flags stale references.

**R7: Bidirectional Feedback Loop**
Formalize the bottom-up flow: code experience → lessons.md → ARCHITECTURE.md updates.
After each major plan completion, require a lessons.md that maps experience to
ARCHITECTURE.md gaps. This review document is itself an instance of this pattern.

**R8: Agent Architecture Awareness Scoring**
For each agent, compute an "architecture awareness" score based on the number of
ARCHITECTURE.md sections it references relative to its function scope. Planning agents
should reference testing/quality/coding. Execution agents should reference all of those
plus deployment. Scores below threshold trigger CI/CD warnings.

## 6. Propagation Impact of ARCH-SUGGESTIONS.md

If all 8 suggestions from ARCH-SUGGESTIONS.md were implemented, the following propagations
would be triggered:

### 6.1 Direct Propagations (ARCHITECTURE.md → Downstream)

| Suggestion | ARCH Section | Downstream Target | Type |
|-----------|-------------|-------------------|------|
| 1. Test Seam | 10.2 (new 10.2.4) | 03-02.testing.instructions.md | Add subsection |
| 1. Test Seam | 10.2 (new 10.2.4) | implementation-execution.agent.md | Add reference |
| 2. Coverage Ceiling | 10.2.3 (append) | 03-02.testing.instructions.md | Update section |
| 2. Coverage Ceiling | 10.2.3 (append) | 06-01.evidence-based.instructions.md | Add reference |
| 3. OTel Constraints | 9.4 (append) | 02-03.observability.instructions.md | Add table |
| 3. OTel Constraints | 9.4 (append) | 04-01.deployment.instructions.md | Add reference |
| 3. OTel Constraints | 9.4 (config) | otel-collector-config.yaml | Fix config |
| 4. Propagation Map | 12.7 (replace) | docs/ARCHITECTURE.md (self) | Replace table |
| 4. Propagation Map | 12.7 (replace) | .github/copilot-instructions.md | Update ref |
| 5. Plan Lifecycle | 13 (new 13.6) | 06-01.evidence-based.instructions.md | Add reference |
| 5. Plan Lifecycle | 13 (new 13.6) | implementation-planning.agent.md | Add pattern |
| 5. Plan Lifecycle | 13 (new 13.6) | implementation-execution.agent.md | Add reference |
| 6. Per-Pkg Coverage | 2.5 + 10.2.3 | 03-02.testing.instructions.md | Add exception table |
| 6. Per-Pkg Coverage | 2.5 + 10.2.3 | 06-01.evidence-based.instructions.md | Add reference |
| 7. Agent Requirements | 2.1.1 (append) | 06-02.agent-format.instructions.md | Add checklist |
| 7. Agent Requirements | 2.1.1 (append) | implementation-planning.agent.md | Add refs |
| 7. Agent Requirements | 2.1.1 (append) | All agents (audit) | Verify refs |
| 8. Blocker Escalation | 13 (new 13.7) | 01-02.beast-mode.instructions.md | Add reference |
| 8. Blocker Escalation | 13 (new 13.7) | 06-01.evidence-based.instructions.md | Add reference |

### 6.2 Indirect Propagations (Downstream → Further Downstream)

| Source Change | Triggers | Further Target |
|--------------|----------|----------------|
| 03-02.testing.instructions.md updated | Beast-mode reads it | Beast-mode quality checks |
| implementation-planning.agent.md updated | Future plans | Plan quality |
| otel-collector-config.yaml fixed | E2E tests | E2E pass/fail |
| copilot-instructions.md updated | All chat sessions | Session awareness |

### 6.3 Missing/Failing Propagations (Known Debt)

These propagations SHOULD exist but currently DON'T:

| Source | Should Propagate To | Current Status | Impact |
|--------|-------------------|----------------|--------|
| ARCHITECTURE.md Section 6.6 (JOSE) | jose-specific instruction file | NO instruction file | JOSE patterns undocumented for agents |
| ARCHITECTURE.md Section 8.3 (Versioning) | openapi instructions | NOT covered | API versioning not enforced |
| ARCHITECTURE.md Section 8.5 (API Security) | security instructions | NOT covered | API security patterns not in agent context |
| ARCHITECTURE.md Section 9.1 (CLI Patterns) | golang instructions | Partial | CLI patterns not fully propagated |
| ARCHITECTURE.md Section 9.2 (Config) | data-infrastructure instructions | Partial | Config architecture gaps |
| ARCHITECTURE.md Section 10.4 (E2E Strategy) | testing instructions | NOT covered | E2E strategy not in agent context |
| ARCHITECTURE.md Section 14 (Ops) | (no instruction file) | MISSING | Ops patterns have zero agent awareness |
| ARCHITECTURE-INDEX.md | ARCHITECTURE.md sync | Manual only | Index may be stale |
| All instruction files | Agent awareness | Agents don't load instructions | 148 refs invisible to agents |

### 6.4 The Fundamental Tension

The strategy has a structural tension:

1. **ARCHITECTURE.md is comprehensive** (3,819 lines, 133 sections) — good for humans.
2. **Instruction files compress it** (18 files, ~500 lines each) — good for Copilot chat.
3. **Agents are isolated** (don't inherit instructions) — clean execution boundary.
4. **But agents need the same knowledge** — 15 refs vs 148 in instructions.

This means the agent system operates with ~10% of the architectural context that Copilot
chat sessions have. The planning agent operates with 0%. This is the single largest
gap in the propagation strategy.

**Potential Resolution**:
- Option A: Agents load relevant instruction files explicitly (breaks isolation principle)
- Option B: Agents embed their own compressed architectural context (duplication)
- Option C: Agents use `fetch_webpage` to read ARCHITECTURE.md at runtime (dynamic, fragile)
- Option D: A "shared context" file that both instructions and agents reference (middle ground)

None of these are ideal. The current strategy implicitly assumes that Copilot chat
(with instructions) handles most work, and agents handle specific tasks with enough
embedded context. This assumption holds for execution agents (which follow plan.md)
but fails for planning agents (which need broad architectural awareness to create
good plans).

## 7. Conclusion

The markdown link propagation strategy is **sound in principle** but **incomplete in practice**.

**Strengths**: Clear source of truth, machine-parseable cross-references, compressed
instruction files, linter-enforced patterns.

**Weaknesses**: No propagation triggers, 64% unreferenced sections, agents mostly
disconnected, no staleness detection, no bidirectional feedback.

**Priority Fixes**:
1. Complete the propagation mapping (Section 12.7) — immediate, documentation-only
2. Add ARCHITECTURE.md refs to planning agent — immediate, one file
3. Create reference validation script — short-term, one cicd subcommand
4. Formalize coverage ceiling analysis — short-term, testing instructions update
5. Add content-hash staleness detection — medium-term, CI/CD infrastructure
