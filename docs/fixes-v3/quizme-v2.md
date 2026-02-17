# Quizme v2 - ARCHITECTURE.md and Deep Implementation Patterns

**Purpose**: Clarify deep architectural decisions for ARCHITECTURE.md updates, propagation strategy, and implementation rigor

**Context**: Plan.md and tasks.md created with executor decisions from quizme-v1. Now need architectural rigor decisions for implementation.

---

## Question 1: ARCHITECTURE.md Section Depth

**Question**: How detailed should ARCHITECTURE.md sections 12.4-12.6 be?

**Context**: Phase 4 adds 3 new sections (Deployment Validation, Config File Architecture, Secrets Management). Depth affects maintenance burden and implementation guidance.

**Impact**: Affects Phase 4 Task 4.1-4.3 LOE and documentation completeness

**A)** Minimal: Brief overview, defer to code comments and validator implementations (low maintenance, minimal guidance)

**B)** Moderate: Core principles, key patterns, examples for each validator (balanced guidance)

**C)** Comprehensive: Detailed rules, extensive examples, decision rationale, edge cases, all 8 validators fully documented (maximum guidance, higher maintenance)

**D)** Reference-heavy: Link to external docs (Docker Compose spec, CONFIG-SCHEMA.md), minimal duplication (low doc drift risk)

**E)**

**Answer**:

**Rationale**: Affects documentation depth, maintenance burden, implementation guidance clarity

---

## Question 2: CONFIG-SCHEMA.md Integration Strategy

**Question**: How should ValidateSchema reference CONFIG-SCHEMA.md?

**Context**: Phase 3 Task 3.3 implements ValidateSchema. CONFIG-SCHEMA.md doc exists but needs programmatic access for validation.

**Impact**: Affects Task 3.3 implementation approach

**A)** Parse CONFIG-SCHEMA.md markdown at runtime (flexible, slower, parsing complexity)

**B)** Generate Go types from CONFIG-SCHEMA.md (type-safe, fast, build step required)

**C)** Hardcode schema in Go (fastest, but CONFIG-SCHEMA.md and code drift risk)

**D)** Embed CONFIG-SCHEMA.md as string, parse once at init (balanced, compiled-in doc)

**E)**

**Answer**:

**Rationale**: Trade-off between flexibility, performance, and maintenance

---

## Question 3: Pre-Commit Hook Performance Strategy

**Question**: How should pre-commit optimize validation performance?

**Context**: 8 validators + 50+ config files = potential 30-60s pre-commit time. Developers expect <5s pre-commit.

**Impact**: Affects Phase 3 Task 3.9 pre-commit integration

**A)** Run all validators sequentially (simple, slow, 30-60s)

**B)** Run validators in parallel (faster, complex, 10-15s)

**C)** Cache validation results (file hash-based), skip unchanged (fastest, complexity, 1-5s for incremental)

**D)** Validate only staged files (minimal, misses cross-file consistency)

**E)**

**Answer**:

**Rationale**: Developer experience vs implementation complexity trade-off

---

## Question 4: Template Pattern Validation Depth

**Question**: What defines "follows template pattern" for PRODUCT/SUITE configs?

**Context**: Phase 2 creates PRODUCT/SUITE configs "following template pattern" (Decision 4, Decision 7). Need concrete validation rules.

**Impact**: Affects Phase 2 Task 2.1-2.6 acceptance criteria and Phase 3 ValidateSchema

**A)** Naming convention only: {PRODUCT}-app-{variant}.yml (surface-level)

**B)** Naming + key structure: Required keys match template (structural)

**C)** Naming + key structure + value patterns: Port offsets, delegation patterns, secret paths (deep validation)

**D)** Template-driven generation: Generate from template, NO manual editing (strictest, limits flexibility)

**E)**

**Answer**:

**Rationale**: Defines rigor level for template compliance

---

## Question 5: Propagation Verification Strategy

**Question**: How to verify ARCHITECTURE.md patterns fully propagated to instruction files?

**Context**: Phase 5 updates instruction files with ARCHITECTURE.md deployment patterns. Need systematic verification.

**Impact**: Affects Phase 5 Task 5.3 verification approach and completeness confidence

**A)** Manual review: Read ARCHITECTURE.md, check each instruction file (error-prone)

**B)** Keyword search: Grep for deployment/config/validation terms (finds references, not completeness)

**C)** Semantic diff: Extract concepts from ARCHITECTURE.md, verify in instructions (complex, thorough)

**D)** Checklist-based: Pre-defined list of patterns to verify in each file (systematic, maintainable)

**E)**

**Answer**:

**Rationale**: Balance between thoroughness and practicality

---

## Question 6: CICD Validator Error Message Strategy

**Question**: How verbose should validator error messages be?

**Context**: 8 validators will report errors. Error message quality affects debugging speed.

**Impact**: Affects all Phase 3 validator implementations (Tasks 3.1-3.8)

**A)** Minimal: "Validation failed" + error type (terse)

**B)** Standard: Error type + file path + line number (Go-style)

**C)** Verbose: Error type + context + suggestion + ARCHITECTURE.md reference (helpful)

**D)** Interactive: Prompt for fix with options (too complex for CI/CD)

**E)**

**Answer**:

**Rationale**: Developer debugging experience vs implementation complexity

---

## Question 7: Secrets Validation Scope

**Question**: How aggressively should ValidateSecrets detect credential patterns?

**Context**: Phase 3 Task 3.8 implements ValidateSecrets. Balance between false positives and security.

**Impact**: Affects Task 3.8 pattern detection rules

**A)** Conservative: Only common keywords (password, secret, token, key) in values (baseline)

**B)** Moderate: Common keywords + pattern detection (base64, hex strings >16 chars) (balanced)

**C)** Aggressive: Common keywords + patterns + entropy analysis (high-entropy strings) (fewer leaks, more false positives)

**D)** Allowlist-based: Flag ALL non-secret-path values, require explicit allowlist (strictest, initial setup burden)

**E)**

**Answer**:

**Rationale**: Security rigor vs false positive tolerance

---

## Question 8: ARCHITECTURE.md Diagram Strategy

**Question**: Should ARCHITECTURE.md sections include diagrams?

**Context**: Phase 4 adds sections 12.4-12.6. Diagrams improve comprehension but add maintenance.

**Impact**: Affects Phase 4 Task 4.1-4.3 LOE and documentation clarity

**A)** No diagrams: Text-only (low maintenance, harder to visualize)

**B)** ASCII diagrams: Simple text diagrams (Git-friendly, limited expressiveness)

**C)** Mermaid diagrams: Code-based diagrams (Git-friendly, expressive, requires renderer)

**D)** External diagrams: Link to draw.io/Excalidraw files (flexible, separate maintenance)

**E)**

**Answer**:

**Rationale**: Documentation clarity vs maintenance burden

---

## Question 9: Mutation Testing Exemptions

**Question**: Should any validators be exempt from ≥98% mutation testing?

**Context**: Infrastructure code requires ≥98% mutation score. Some validators have trivial logic (e.g., kebab-case regex check).

**Impact**: Affects Phase 3 Task 3.10 mutation testing LOE

**A)** No exemptions: ALL validators ≥98% (strictest rigor)

**B)** Exempt trivial validators: ValidateKebabCase, ValidateNaming if simple regex (practical)

**C)** Exempt by complexity: Only complex validators (ValidateSchema, ValidateConsistency) require ≥98% (risk-based)

**D)** User decision per validator: Document exemption rationale in code (flexible)

**E)**

**Answer**:

**Rationale**: Quality rigor vs practical implementation trade-offs

---

## Question 10: Phase 0 Research Results Documentation

**Question**: Should Phase 0 research findings be documented in plan.md?

**Context**: Agent instructions say Phase 0 is internal research (NOT output documentation). Quizme-v2 may surface unknowns.

**Impact**: Clarifies whether Phase 0 findings go in plan.md or stay in test-output/

**A)** Internal only: Findings stay in test-output/phase0-research/, NOT in plan.md (agent pattern adherence)

**B)** Summary in plan.md: Brief summary of Phase 0 findings in "Background" or "Executive Summary" (hybrid)

**C)** Full documentation: Phase 0 findings fully documented in plan.md (violates agent pattern)

**D)** Reference-based: Plan.md references test-output/phase0-research/ for details (clearest separation)

**E)**

**Answer**:

**Rationale**: Clarifies agent Phase 0 pattern interpretation

---

## Instructions for User

**Fill in Answer field with: A, B, C, D, or E (with optional explanation)**

**After answering**:
1. Agent will merge answers into plan.md Executive Decisions
2. Agent will update tasks.md acceptance criteria
3. Agent will delete quizme-v2.md
4. Agent will commit changes

**Answer Format Example**:
```
**Answer**: C - I want comprehensive docs even if higher maintenance
```
