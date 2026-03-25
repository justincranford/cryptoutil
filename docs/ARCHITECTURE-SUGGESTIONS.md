# ARCHITECTURE.md Improvement Suggestions

**Date**: 2026-02-17
**Status**: Review-first (do NOT apply without review)
**Scope**: Efficiency, quality, correctness, completeness, thoroughness, and general quality gates for `docs/ARCHITECTURE.md`

---

## Document Profile (Current State)

| Metric | Value |
|--------|-------|
| Total lines | 5,362 |
| Total characters | ~263,755 |
| Estimated tokens (Claude Sonnet 4.6) | ~75,000 |
| Instruction files (18) token cost | ~36,500 |
| copilot-instructions.md token cost | ~1,700 |
| **Combined context budget** | **~113,500 tokens (56.8% of 200K window)** |
| Top-level sections (##) | 14 |
| Total headings | ~430 |
| @propagate markers | 43 |
| Code blocks | ~83 |
| Table rows | ~480 |

**Largest sections by line count**:

| # | Section | Lines | % of doc |
|---|---------|-------|----------|
| 12 | Deployment Architecture | 1,016 | 18.9% |
| 10 | Testing Architecture | 756 | 14.1% |
| 9 | Infrastructure Architecture | 608 | 11.3% |
| 4 | System Architecture | 385 | 7.2% |
| 6 | Security Architecture | 374 | 7.0% |
| 2 | Strategic Vision | 350 | 6.5% |
| 3 | Product Suite | 330 | 6.2% |

---

## Suggestions

### Category A: Token Budget Reduction (Efficiency)

These suggestions reduce token consumption without losing information.

#### 1. Extract Deployment Secret Format Examples to a Satellite Reference

**Section**: 12.3.3 Secrets Coordination Strategy (lines ~4000-4300)

**Problem**: The secret file format examples (unseal keys, hash pepper, PostgreSQL credentials, TLS secrets, healthcheck-secrets service YAML, cross-reference docs) occupy ~200 lines of highly structured, rarely-changing reference content. LLM agents rarely need the exact byte format of a secret file during normal development — they need the rules (use Docker secrets, chmod 440, never inline).

**Suggestion**: Extract the _format examples and sample values_ into a new `docs/SECRETS-REFERENCE.md`. Keep the rules, naming conventions, and tier strategy in ARCHITECTURE.md. Add a one-line cross-reference: `See [SECRETS-REFERENCE.md](SECRETS-REFERENCE.md) for secret file format examples and sample values.`

**Token savings**: ~2,500-3,000 tokens.

**Trade-off**: Agents creating new deployments would need to open a second file. Mitigated by instruction files already containing the essential rules.

---

#### 2. Consolidate the Three-Tier Deployment Hierarchy Compose Examples

**Section**: 12.3.4 Multi-Level Deployment Hierarchy (lines ~4300-4500)

**Problem**: Three near-identical YAML blocks (SUITE, PRODUCT, SERVICE compose patterns) show the same structure with trivially different values (paths, port offsets, pepper prefixes). The accompanying tables (Port Offset Strategy, Layered Pepper Strategy, Docker Compose `include` Semantics) repeat information already stated in Section 3.4 (Port Assignments) and 12.3.3 (Secrets Coordination).

**Suggestion**: Replace the three full compose examples with a single parameterized template showing `{TIER}` placeholders, plus a compact 3-row table showing the per-tier differences. Remove the Port Offset Strategy table (it duplicates Section 3.4.1 Port Design Principles, which already has the +0/+10000/+20000 rule).

**Token savings**: ~1,500-2,000 tokens.

---

#### 3. Compress the Fitness Sub-Linter Catalog (Section 9.11)

**Section**: 9.11 Architecture Fitness Functions (57 sub-linters)

**Problem**: The full 57-entry fitness sub-linter catalog is valuable for developers authoring new linters but unlikely to be needed during normal agent interactions. Each entry includes name, package, description, and error/warning classification.

**Suggestion**: Keep a summary table with linter _categories_ (8-10 rows: file-size, parallel-tests, banned-names, entity-registry, etc.) with counts per category. Move the full 57-entry catalog to a generated comment block in the linter registry code or to `docs/FITNESS-LINTERS.md`. Add a one-line cross-reference.

**Token savings**: ~2,000-2,500 tokens.

---

#### 4. Reduce Code Examples in Section 10 (Testing Architecture)

**Section**: 10.1-10.8 (lines 2808-3563)

**Problem**: Section 10 contains 13+ full Go code blocks (benchmark templates, TestMain patterns, table-driven test templates, contract test boilerplate, Fiber app.Test patterns, etc.). Many of these are also reproduced in `03-02.testing.instructions.md`. LLM agents receive both the ARCHITECTURE.md section and the instruction file, creating double exposure.

**Suggestion**: In ARCHITECTURE.md, keep one canonical example per testing pattern (e.g., one table-driven test, one TestMain, one benchmark) and replace the others with brief descriptions + cross-references to the instruction file or to actual test files in the codebase. The instruction file is the "tactical quick reference" and should hold the examples; ARCHITECTURE.md should hold the _strategy and rationale_.

**Token savings**: ~2,000-3,000 tokens.

**Trade-off**: Agents using ARCHITECTURE.md without instruction files (e.g., custom agents in isolation mode) would lose some examples. Mitigated by agent self-containment checklist requiring ARCHITECTURE.md references.

---

#### 5. De-duplicate @propagate Blocks with Multi-Target Syntax

**Section**: Throughout (43 markers)

**Problem**: The `@propagate` system requires one block per target file, causing verbatim duplication. The `infrastructure-blocker-escalation` chunk appears in two `@propagate` blocks (lines 5055-5069), and `mandatory-review-passes` appears in two blocks (lines 560-600). Each duplicated block adds ~300-600 tokens of identical content.

**Suggestion**: Extend the propagation marker syntax to support multi-target in a single block:
```html
<!-- @propagate to="01-02.beast-mode.instructions.md, 06-01.evidence-based.instructions.md" as="mandatory-review-passes" -->
```
This would require updating the `lint-docs validate-propagation` parser, but would eliminate all intra-document content duplication from multi-target propagation. Currently there are ~4 such duplicated blocks.

**Token savings**: ~1,200-1,500 tokens.

**Trade-off**: Requires code change to `lint-docs`. Low complexity — the parser already extracts `to` and `as` attributes; adding comma-separated `to` values is straightforward.

---

#### 6. Replace Verbose Directory Trees with Compact Tables

**Section**: 4.4.6 Deployments (lines ~1050-1300)

**Problem**: The three-tier deployment directory tree (`deployments/{PS-ID}/`, `deployments/{PRODUCT}/`, `deployments/{SUITE}/`) uses ASCII art tree structures (`├──`, `└──`) expanded across ~150 lines. Much of this is repetitive (14 secret files listed 3 times with trivially different filenames).

**Suggestion**: Replace with a single compact table showing directory structure per tier, plus one unexpanded example:

| Tier | compose.yml | Dockerfile | secrets/ (count) | config/ | otel-config |
|------|-------------|------------|-------------------|---------|-------------|
| SERVICE | required | required | 14 `.secret` | required | optional |
| PRODUCT | required | required | 14 `.secret` + 4 `.never` | — | — |
| SUITE | required | — | 14 `.secret` + 4 `.never` | — | — |

Keep one fully expanded SERVICE-level tree as the canonical example. Remove the PRODUCT and SUITE trees (they differ only by the addition of `.never` marker files, which the table captures).

**Token savings**: ~1,500-2,000 tokens.

---

#### 7. Compress the Authentication Realm Type Table (Section 7.2.1)

**Section**: 7.2.1 Authentication Realms

**Problem**: The 23+ row realm type table repeats similar patterns (each realm type has: name, factor type, storage, federation). Many are minor variants (e.g., JWE/JWS/Opaque session cookie, JWE/JWS/Opaque session token — 6 near-identical rows differing only in token format).

**Suggestion**: Group by pattern with a compact format:
```
Session cookies: JWE | JWS | Opaque (browser only, SQL storage)
Session tokens: JWE | JWS | Opaque (headless only, SQL storage)
MFA factors: TOTP | HOTP | WebAuthn | Push | Recovery (federated only, SQL storage)
```

**Token savings**: ~800-1,200 tokens.

---

### Category B: Correctness & Consistency

These suggestions fix errors, inconsistencies, or misleading content.

#### 8. Fix Section 13.5 Numbering (Out of Order)

**Section**: 13.5 Development Workflow

**Problem**: Sub-sections are numbered out of order:
- 13.5.1 Spec Structure Patterns (line 4911)
- 13.5.2 Terminal Command Auto-Approval (line 4918)
- 13.5.3 Session Documentation Strategy (line 4925)
- **13.5.5** Air Live Reload (line 4932) ← should be 13.5.4
- **13.5.4** Docker Desktop Startup (line 4957) ← should be 13.5.5

**Suggestion**: Renumber so they are sequential. Since Docker Desktop Startup has `@propagate` blocks with stable anchors (`#1354-docker-desktop-startup---critical`), the safest fix is to swap the numbers: rename current 13.5.5 to 13.5.4 (Air) and current 13.5.4 to 13.5.5 (Docker). This changes the Docker Desktop anchor from `#1354` to `#1355`.

**Impact**: Need to update all cross-references to `#1354` and `#1355` across ARCHITECTURE.md and instruction files. Grep for `1354` and `1355` to find them.

---

#### 9. Reconcile Duplicate Port Assignment Documentation

**Sections**: 3.4 Port Assignments & Networking, 3.4.1 Port Design Principles, 12.3.4 Port Offset Strategy

**Problem**: Port assignment rules are stated in three places:
1. Section 3.4: Full service catalog with host ports
2. Section 3.4.1: Three deployment type offsets (+0, +10000, +20000)
3. Section 12.3.4 Port Offset Strategy table: Same three offsets with sm-kms example

**Suggestion**: Define port rules once in Section 3.4.1 and cross-reference from 12.3.4. Remove the Port Offset Strategy table from 12.3.4 (it adds no information beyond 3.4.1).

---

#### 10. Clarify the `end-of-turn-commit-protocol` Propagation Target

**Section**: 2.4 Implementation Strategy (line ~460)

**Problem**: The `end-of-turn-commit-protocol` block has `@propagate to=".github/instructions/01-02.beast-mode.instructions.md"` but the beast-mode instruction file is loaded via the mode instructions (injected at runtime), not via the standard instructions directory scan. This means the propagation target is architecturally correct but the block is consumed differently than other instruction file blocks — it goes into the mode prompt, not the instructions context.

**Suggestion**: Add a comment noting this distinction for maintainers:
```html
<!-- NOTE: Target is beast-mode, which is injected as modeInstructions at runtime, not via standard instructions scan -->
```
This prevents future maintainers from assuming all `@propagate` targets are consumed identically.

---

#### 11. Remove Stale "TBD" Entries in Appendix B

**Sections**: B.5 Agent Catalog, B.6 CI/CD Workflow Catalog, B.7 Reusable Action Catalog, B.8 Linter Rule Reference

**Problem**: Multiple table cells contain "TBD" placeholders:
- B.5: fix-workflows and beast-mode have `Tools: TBD`, `Handoffs: TBD`
- B.6: ci-race, ci-benchmark, ci-sast, ci-dast, ci-e2e, ci-load all have `Duration: TBD`
- B.7: "Additional actions | TBD | TBD | TBD"
- B.8: "... | (30+ total linters) | TBD | TBD | TBD"

**Suggestion**: Either fill in the values (agents have defined tools in their `.agent.md` files; workflow durations can be estimated from `timeout` values) or remove the TBD rows entirely. TBD rows consume tokens without providing value and risk agents treating them as actionable items.

---

#### 12. Fix Healthcheck Pattern Numbering (Section 12.3.1)

**Section**: 12.3.1 Docker Compose Deployment, Health Checks subsection

**Problem**: Three healthcheck patterns are listed as numbered items, but all three use `1.` instead of `1.`, `2.`, `3.`:
```markdown
1. **Service-only** (native HEALTHCHECK):
...
1. **Job-only** (validation job, ExitCode=0 required):
...
1. **Service with healthcheck job** (external sidecar):
```

**Suggestion**: Fix numbering to `1.`, `2.`, `3.`.

---

### Category C: Completeness

These suggestions identify missing information or gaps.

#### 13. Add `@propagate` Markers for the 3 Non-Propagated Instruction Files

**Sections**: Various

**Problem**: Section 12.7.7 notes that 3 instruction files (03-04.data-infrastructure, 03-05.linting, copilot-instructions) have zero `@propagate` chunks because "their content is condensed quick-reference summaries." However:
- `03-04.data-infrastructure` contains the critical "3-Tier Database Strategy" which is also stated verbatim in Section 10.1 — this IS propagatable.
- `03-05.linting` contains the "UTF-8 without BOM" mandate which exists verbatim in ARCHITECTURE.md Section 9.9.3.

**Suggestion**: Add `@propagate` markers for:
- `three-tier-database-strategy` → `03-02.testing.instructions.md` (already has an `@source` block for `test-file-suffixes`; this would add the DB strategy)
- `utf8-without-bom` → `03-05.linting.instructions.md`

This brings propagation coverage from 15/18 to 17/18 instruction files.

---

#### 14. Document the `@propagate` Marker Syntax in a Machine-Readable Schema

**Section**: 12.7.2 Propagation Marker System

**Problem**: The marker syntax is documented in prose, but the validation algorithm (12.7.5) relies on regex extraction. There is no formal grammar or schema. If a future maintainer introduces a variant (e.g., `@propagate from=...` in a target file), the validator would silently miss it.

**Suggestion**: Add a formal BNF-like or regex specification:
```
@propagate-open  ::= '<!-- @propagate to="' PATH '" as="' CHUNK_ID '" -->'
@propagate-close ::= '<!-- @/propagate -->'
@source-open     ::= '<!-- @source from="' PATH '" as="' CHUNK_ID '" -->'
@source-close    ::= '<!-- @/source -->'
PATH             ::= [a-zA-Z0-9_./-]+
CHUNK_ID         ::= [a-z0-9-]+
```
This is 5-6 lines and makes the contract explicit.

---

#### 15. Add Anchor Stability Policy

**Section**: Document Metadata or Section 2

**Problem**: Many instruction files and agents reference ARCHITECTURE.md anchors (e.g., `#1354-docker-desktop-startup---critical`). Section renumbering (like suggestion #8) would break all cross-references. There is no documented policy for anchor stability.

**Suggestion**: Add a brief anchor stability policy:
- Section numbers in anchors are stable unless explicitly renumbered in a documented change.
- When renumbering, grep for old anchor patterns and update all references atomically.
- Consider named anchors (e.g., `#docker-desktop-startup`) instead of numbered anchors for frequently-referenced sections.

---

### Category D: Thoroughness & Quality Gates

These suggestions improve validation, testing, or enforcement coverage.

#### 16. Add `@propagate` Drift Detection to Pre-Commit Hooks

**Section**: 9.9 Pre-Commit Hook Architecture

**Problem**: `@propagate`/`@source` drift is currently caught only by `cicd lint-docs validate-propagation` (CI/CD). It is NOT in the pre-commit hook sequence. An agent or developer can commit drifted content, and the error only surfaces when CI/CD runs.

**Suggestion**: Add `lint-docs validate-propagation` to the `cicd-lint-all` pre-commit hook group. This catches drift before commit, consistent with the "zero escaped violations" philosophy.

**Trade-off**: Adds a few seconds to pre-commit time. Justified by the high cost of propagation drift (stale instruction files mislead agents).

---

#### 17. Add Config File Count Validation to Deployment Linter

**Section**: 12.4.5 Config File Naming Strategy

**Problem**: The naming strategy requires 5 config files per service (`{PS-ID}-app-{common,sqlite-1,sqlite-2,postgresql-1,postgresql-2}.yml`). The linter validates naming patterns, but the document does not specify whether the linter validates the _exact count_ (i.e., that all 5 variants exist). A service missing `sqlite-2.yml` might pass naming validation but fail the completeness check.

**Suggestion**: Clarify whether the validator enforces the presence of all 5 variants. If it does, document this. If it doesn't, consider adding it — it's a meaningful completeness check that prevents incomplete deployments.

---

#### 18. Add Token Budget Warning to Documentation Standards

**Section**: 11.4 Documentation Standards

**Problem**: There is no documented maximum size target for ARCHITECTURE.md. At 75K tokens it consumes 37.5% of a 200K context window before any code is loaded. As the document grows with new services and features, it could exceed 50% alone.

**Suggestion**: Add a token budget guideline:
- ARCHITECTURE.md SHOULD stay below 60K tokens (~4,800 lines).
- When approaching the limit, apply extraction strategies (satellite reference docs for format examples, generated tables, code-as-source-of-truth).
- Instruction files in aggregate SHOULD stay below 30K tokens.
- Combined budget (ARCHITECTURE.md + instruction files + copilot-instructions.md) SHOULD stay below 100K tokens (50% of minimum expected context window).

This creates a measurable quality gate for documentation size.

---

### Category E: Structural Improvements

These suggestions improve readability, navigation, or organization.

#### 19. Split Section 12 (Deployment Architecture) into Two Sections

**Section**: 12. Deployment Architecture (1,016 lines, 18.9% of document)

**Problem**: Section 12 is disproportionately large. It contains two distinct concerns:
1. **Deployment Patterns** (12.1-12.3): CI/CD automation, build pipeline, deployment patterns, secrets coordination — these are _operational_ concerns.
2. **Deployment Validation** (12.4-12.8): Directory structure validation, config file architecture, secrets management linting, documentation propagation — these are _tooling/quality_ concerns.

**Suggestion**: Split into:
- **Section 12: Deployment Architecture** (12.1-12.3, ~500 lines): Patterns, pipelines, secrets.
- **Section 15: Deployment Tooling & Validation** (current 12.4-12.8, ~500 lines): Linters, validators, config schema, propagation.

Alternatively, move 12.7 (Documentation Propagation) to a new Section 15 or an Appendix, since documentation propagation is conceptually distinct from deployment.

**Trade-off**: Changes all Section 13/14 numbering and cross-references. A major renumbering event, but it would bring all sections below 600 lines.

---

#### 20. Move Section 14 (Operational Excellence) Content to Appendix

**Section**: 14. Operational Excellence (52 lines)

**Problem**: Section 14 is extremely thin (52 lines) compared to other sections. Its content (monitoring, incident management, performance, capacity planning, disaster recovery) consists of brief bullet-point summaries with no detail — essentially placeholder text. It provides minimal value to LLM agents.

**Suggestion**: Either:
- (a) Move to Appendix D as a compact reference table, saving a top-level section number.
- (b) Expand with actionable detail (runbook templates, alert thresholds, Grafana dashboard configs) — but only when the content exists in the codebase.
- (c) Leave as-is but add a note: "This section will be expanded when operational monitoring is implemented."

Option (c) is lowest risk. Option (a) saves ~1% token overhead from reduced heading hierarchy.

---

#### 21. Add a Document Map / Reading Guide

**Section**: Top of document (after Document Organization, before Section 1)

**Problem**: At 5,362 lines with 14 sections, navigating ARCHITECTURE.md requires knowing the section numbering scheme. The current "Architecture at a Glance" (Section 1.5) provides a text summary, but no visual map or reading guide that matches sections to use cases.

**Suggestion**: Add a concise reading guide table:

| If you need to... | Read section(s) |
|--------------------|----------------|
| Understand the product suite | 1, 3 |
| Add a new service | 4.4, 5.1, 5.2 |
| Configure deployment | 12.1-12.3 |
| Write/fix tests | 10 |
| Understand security | 6 |
| Add/modify API endpoints | 8 |
| Fix CI/CD or linting | 9, 11.3 |
| Understand secrets | 6.10, 12.3.3, 12.6 |

This adds ~15 lines but significantly improves navigation for both humans and agents.

---

#### 22. Add Cross-Reference Index for Frequently-Referenced Topics

**Section**: Appendix or new Section

**Problem**: Some topics (secrets, TLS, port assignments, health checks) are referenced from 4+ different sections. An agent searching for "secrets" would need to know about Sections 4.4.6, 6.10, 12.3.3, 12.6, and the instruction files. Currently there is no central index.

**Suggestion**: Add a compact cross-reference index in an Appendix:

| Topic | Primary Section | Also Referenced In |
|-------|----------------|-------------------|
| Secrets | 12.3.3 | 4.4.6, 6.10, 12.6, 02-05.security |
| TLS | 6.11 | 5.3, 12.3.3, 02-05.security |
| Ports | 3.4 | 12.3.4, 04-01.deployment |
| Health checks | 5.5 | 12.3.1, 02-01.architecture |

This adds ~20 lines and prevents agents from missing relevant context scattered across sections.

---

### Category F: @propagate System Improvements

#### 23. Formalize the "Glue Content" Maintenance Strategy

**Section**: 12.7.6 Feasibility Constraints

**Problem**: The propagation system validates that `@propagate` blocks match their `@source` counterparts byte-for-byte. However, the ~20% "glue content" (headings, `See` cross-references, transitional paragraphs) in instruction files is NOT validated. If an ARCHITECTURE.md section is reorganized, the glue content in instruction files becomes stale. There is no detection mechanism for stale glue.

**Suggestion**: Document a maintenance checklist:
- When renumbering ARCHITECTURE.md sections, grep all instruction files for `See [ARCHITECTURE.md Section X.Y]` patterns and update.
- Consider a lightweight linter that validates anchor targets in `See` cross-references actually exist in ARCHITECTURE.md.

---

#### 24. Track Propagation Coverage Metrics

**Section**: 12.7.7 Migration Strategy

**Problem**: The document states "15 files have 1+ propagation chunks (26 total chunk pairs)" and "3 files are structural glue only." This is a manual accounting that will drift as chunks are added. There is no automated way to track propagation coverage percentage.

**Suggestion**: Add a `lint-docs propagation-coverage` sub-command that reports:
- Total instruction files, files with ≥1 chunk, coverage percentage
- Total lines in instruction files, lines inside `@source` blocks, coverage percentage
- This gives an objective metric for how much of the instruction content is auto-validated vs manually maintained.

---

### Category G: Miscellaneous

#### 25. Remove or Archive `docs/ARCHITECTURE-INDEX.md` Reference

**Section**: Document Metadata, last line

**Problem**: The Document Metadata section references `docs/ARCHITECTURE-INDEX.md` as "Agent lookup reference." Verify this file still exists and is current. If it was a temporary artifact or has been superseded, remove the reference.

---

#### 26. Review `docs/CONFIG-SCHEMA.md` vs Hardcoded Schema

**Section**: 12.5 Config File Architecture

**Problem**: Section 12.5 states "Config file schema is HARDCODED in Go (`validate_schema.go`). No external schema files." However, Section 12.4.8 references `[CONFIG-SCHEMA.md](/docs/CONFIG-SCHEMA.md)` for "complete config file schema with all supported keys." These two statements are contradictory — either CONFIG-SCHEMA.md exists as a reference doc, or the schema is code-only.

**Suggestion**: Clarify the relationship: if CONFIG-SCHEMA.md exists as a human-readable reference (not a machine-consumed schema), state that. If it was deprecated in favor of code-only schema, remove the reference.

---

## Summary Statistics

| Category | Count | Estimated Token Savings |
|----------|-------|------------------------|
| A: Token Budget Reduction | 7 | ~11,500-15,700 tokens |
| B: Correctness & Consistency | 5 | — (no size change) |
| C: Completeness | 3 | +200 tokens (small additions) |
| D: Quality Gates | 3 | — (no size change) |
| E: Structural Improvements | 4 | ~-200 tokens (minor additions) |
| F: @propagate System | 2 | — (tooling changes) |
| G: Miscellaneous | 2 | — (minor edits) |
| **Total** | **26** | **~11,000-15,000 tokens net savings** |

Applying all Category A suggestions would reduce ARCHITECTURE.md from ~75K to ~60-64K tokens, bringing the combined context budget from 56.8% to ~49-51% of a 200K window — below the recommended 50% threshold.

---

## Prioritized Implementation Order

**Phase 1 — Quick wins (correctness fixes, no token impact)**:
- #8 Fix Section 13.5 numbering
- #11 Remove stale TBD entries
- #12 Fix healthcheck pattern numbering

**Phase 2 — Token reduction (high impact, low risk)**:
- #1 Extract secret format examples
- #6 Replace verbose directory trees
- #4 Reduce testing code examples
- #7 Compress realm type table

**Phase 3 — Structural improvements (moderate risk)**:
- #2 Consolidate deployment compose examples
- #3 Compress fitness linter catalog
- #9 Reconcile duplicate port docs
- #21 Add reading guide

**Phase 4 — Infrastructure improvements (requires code changes)**:
- #5 Multi-target @propagate syntax
- #16 Pre-commit propagation check
- #24 Propagation coverage metrics

**Phase 5 — Major restructuring (high risk, high reward)**:
- #19 Split Section 12
- #18 Token budget quality gate
- #15 Anchor stability policy
