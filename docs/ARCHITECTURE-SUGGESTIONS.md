# ARCHITECTURE.md Improvement Suggestions

**Date**: 2026-03-26
**Status**: Review-first (do NOT apply without review)
**Scope**: Deep analysis of `docs/ARCHITECTURE.md`, Copilot instructions/agents/skills, lint-fitness sub-linters, and INFRA-TOOL structure

---

## Document Profile (Current State — 2026-03-26)

| Metric | Value | vs. 2026-02-17 |
|--------|-------|----------------|
| Total lines | 3,936 | -1,426 (-26.6%) |
| Total characters | 260,334 | ~-3,400 (-1.3%) |
| Estimated tokens (Claude Sonnet 4.6) | ~55,000 | -20,000 |
| Instruction files (18) token cost | ~36,500 | unchanged |
| copilot-instructions.md token cost | ~1,700 | unchanged |
| **Combined context budget** | **~93,200 tokens (46.6% of 200K window)** | was 56.8% |
| Top-level sections (##) | 14 | unchanged |
| Total headings | 383 | -47 (-10.9%) |
| @propagate markers | 38 | -5 (-11.6%) |
| Code blocks | ~60 | -23 |

**Largest sections (current)**:

| # | Section | Lines | % of doc |
|---|---------|-------|----------|
| 12 | Deployment Architecture | 1,015 | 25.8% |
| 10 | Testing Architecture | ~480 | 12.2% |
| 9 | Infrastructure Architecture | ~420 | 10.7% |
| 6 | Security Architecture | ~350 | 8.9% |
| 2 | Strategic Vision | ~280 | 7.1% |
| 4 | System Architecture | ~250 | 6.4% |
| 14 | Operational Excellence | 35 | 0.9% |

**Notable changes since Feb 2026**:
- ARCHITECTURE.md shrank 26.6% (1,426 lines) due to framework-v6 restructuring.
- Section 12 grew proportionally: was 18.9%, now **25.8%** — now dominates the document by a wider margin.
- Section 14 remains extremely thin at 35 lines.
- `@propagate` markers decreased 5 (from 43 to 38) — some propagation blocks removed during restructuring.
- Combined context budget improved from 56.8% to 46.6% — now at a healthy target.

---

## Previous Suggestions Status (2026-02-17 Backlog)

Of the 26 Feb 2026 suggestions, approximately 6 were partially applied during framework-v6 restructuring (evidenced by the 1,426-line reduction, heading count decrease, and @propagate marker reduction). The remaining 20 are still open and valid.

### Applied (estimated ~6, partially)

Evidence of application: document shrank 1,426 lines; headings -47; @propagate -5. Most likely applied as side-effects of framework-v6 work (not as direct suggestion implementations):
- **#2 partial**: Deployment compose examples may have been compressed.
- **#3 partial**: Fitness linter catalog may have been condensed.
- **#4 partial**: Test code examples reduced (Section 10 shrank significantly).
- **#6 partial**: Directory trees may have been simplified.
- **#1 partial**: Some secret format content extracted.

### Still Open (confirmed NOT applied)

| # | Suggestion | Priority | Status |
|---|-----------|----------|--------|
| 5 | Multi-target `@propagate` syntax | Low | Open (code change) |
| 8 | Fix Section 13.5 numbering (13.5.5 before 13.5.4) | **High** | ✅ Applied (renamed → 14.5.4/14.5.5) |
| 9 | Reconcile duplicate port docs | Low | Open |
| 10 | Clarify `end-of-turn-commit-protocol` propagation target | Low | ✅ Applied (NOTE comment added) |
| 11 | Remove stale TBD entries in Appendix B | **High** | ✅ Applied |
| 12 | Fix healthcheck numbered list (all `1.`) | Medium | ✅ Applied |
| 13 | Add `@propagate` for non-propagated instruction files | Medium | ✅ Applied (three-tier→03-04, utf8-bom→03-05 already were done) |
| 14 | Machine-readable `@propagate` grammar | Low | ✅ Applied (BNF added to §13.4.2) |
| 15 | Anchor stability policy | Medium | Open |
| 16 | Pre-commit propagation drift detection | Medium | Open |
| 17 | Config file count validation in deployment linter | Medium | Open |
| 18 | Token budget quality gate documentation | Medium | Open |
| 19 | Split Section 12 into two sections | **High** | ✅ Applied (§12.4-12.8 → new §13) |
| 20 | Expand / move Section 14 | Medium | Open |
| 21 | Add reading guide table | Medium | Open |
| 22 | Cross-reference index for frequently-referenced topics | Low | Open |
| 23 | Glue content maintenance strategy | Low | Open |
| 24 | Propagation coverage metrics command | Low | Open (code change) |
| 25 | Remove ARCHITECTURE-INDEX.md reference | Low | ✅ Applied (file deleted) |
| 26 | Reconcile CONFIG-SCHEMA.md vs hardcoded schema | Medium | ✅ Applied (file deleted, refs fixed) |

---

## Suggestions (Feb 2026 Backlog — Items #1–#26)

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

---

## New Suggestions (2026-03-26 — Items #27–#50)

### Category H: INFRA-TOOL Rename & Naming Convention

#### 27. Complete `workflow` → `cicd-workflow` Rename

**Context**: The project has 2 INFRA-TOOLs: `cicd-lint` and `workflow`. The `cicd-*` prefix applies to `cicd-lint` but NOT `workflow`. This inconsistency was identified and a rename to `cicd-workflow` is planned.

**Full rename scope**:

| File / Location | Required Change |
|----------------|-----------------|
| `internal/shared/magic/magic_cicd.go` | `CICDCmdDirWorkflow = "workflow"` → `"cicd-workflow"` |
| `internal/shared/magic/magic_workflows.go` | `UsageWorkflow` string: `"workflow"` → `"cicd-workflow"` |
| `cmd/workflow/` | `git mv cmd/workflow cmd/cicd-workflow` |
| `cmd/cicd-workflow/main.go` | Update import path and package comment |
| `internal/apps/tools/workflow/` | `git mv` → `internal/apps/tools/cicd_workflow/` |
| `internal/apps/tools/cicd_workflow/*.go` | `package workflow` → `package cicd_workflow`; update all hardcoded strings |
| `.github/agents/fix-workflows.agent.md` | 40+ occurrences of `./cmd/workflow` → `./cmd/cicd-workflow` |
| `.github/agents/beast-mode.agent.md` | 2 occurrences |
| `docs/ARCHITECTURE.md` | ~8 occurrences on lines ~1033, ~1309, ~3520, ~3627–3640 |

**Fitness linters** (`cmd_entry_whitelist`, `cmd_anti_pattern`): reference `CICDCmdDirWorkflow` via magic constant — auto-update when magic constant changes. No additional code changes needed.

**Status**: Planned — execute before implementing suggestion #29 (infra-tool-naming linter).

---

#### 28. Define INFRA-TOOL CLI Pattern in ARCHITECTURE.md Section 4.4.7

**Problem**: Section 4.4.7 CLI Patterns defines PRODUCT, PRODUCT-SERVICE, and SUITE patterns. INFRA-TOOLs (`cicd-lint`, `cicd-workflow`) are referenced in examples but are never formally defined as a CLI pattern type — there are no rules about naming, location, or entry function signature.

**Suggestion**: Add an "INFRA-TOOL Pattern" subsection to 4.4.7:

```
INFRA-TOOL Pattern: cmd/cicd-{tool}/main.go → internal/apps/tools/cicd_{tool}/cicd_{tool}.go
- ALL INFRA-TOOL cmd dirs MUST be prefixed with cicd-
- Internal package dirs use underscore: cicd_{tool}
- NOT registered in entity registry (not a product-service)
- Whitelisted in cmd-entry-whitelist fitness linter
```

**Impact**: Provides clear rules for future INFRA-TOOL additions (cicd-release, cicd-migrate, etc.). Update the INFRA-TOOL table in Section 9.10 to `cicd-lint, cicd-workflow` after rename #27.

---

#### 29. Add `infra-tool-naming` Fitness Linter

**Problem**: Neither `cmd-entry-whitelist` nor `cmd-anti-pattern` specifically enforces the `cicd-*` naming convention for INFRA-TOOLs. They whitelist known entries but don't ensure new INFRA-TOOLs follow the convention.

**Suggestion**: Add a new fitness linter `infra-tool-naming` that:
1. Identifies all `cmd/` entries not matching product/service/suite patterns.
2. Validates those entries (INFRA-TOOLs) are prefixed with `cicd-`.
3. Validates matching `internal/apps/tools/` entries use `cicd_` prefix.
4. Emits error otherwise.

**Prerequisite**: Complete rename #27 first so the new linter validates `cicd-workflow`, not the old `workflow`.

---

### Category I: Lint-Fitness Sub-Linter Analysis

**Current state**: 55 registered sub-linters confirmed in `lint_fitness.go`. Directory count matches: 55 dirs + 1 `registry/` helper = 56 total entries.

#### 30. Investigate Admin Port Linter Overlap

**Problem**: Two linters may check overlapping concerns:
- `admin-port-exposure` — checks admin port binding in compose files.
- `validate-admin` — validates admin port configuration.

If both validate that the admin port binds to `127.0.0.1:9090`, one is redundant. If they check different aspects (exposure vs. YAML key naming), they can coexist.

**Action**: Read both implementations and determine if there is genuine duplication. If yes, consolidate into one linter.

---

#### 31. Add `magic-constant-location` Fitness Linter

**Problem**: `internal/shared/magic/` is the MANDATORY location for ALL magic constants. The `mnd` golangci-lint linter catches inline literals but does NOT catch package-local `const x = ...` declarations that should be in `magic/`. A developer can create `internal/myservice/constants.go` with package-local constants and bypass the `mnd` linter entirely.

**Suggestion**: Add a fitness linter that scans all non-`magic/` packages for constant declarations and emits a warning if a `const` declaration contains a numeric value or a string that looks like a magic constant (e.g., port numbers, timeout values, algorithm names).

**Note**: False-positive tuning will be needed — some package-local constants are legitimately package-private. Start with string constants containing known magic patterns (e.g., `"AES"`, `"RSA"`, port-like integers).

---

#### 32. Remove Hard Count from Section 9.11.1 Heading

**Problem**: Section 9.11.1 heading says "(55 total)". Every time a linter is added, this number must be manually updated. When it falls out of sync, the document is misleading.

**Suggestion**: Remove the hard count from the heading. Change `Fitness Sub-Linter Catalog (55 total)` to `Fitness Sub-Linter Catalog` and state the count in a table or note that can auto-generate. Alternatively, add a CI step asserting `len(registeredLinters) == number of lint_fitness subdirs (excluding registry/)`.

---

#### 33. Add Regression Test for `workflow` Rename in `cmd_anti_pattern_test.go`

**Problem**: After the `workflow` → `cicd-workflow` rename (#27), the magic constant changes from `"workflow"` to `"cicd-workflow"`. Without a regression test, a future refactor could silently change it back.

**Suggestion**: Add a test in `cmd_anti_pattern_test.go` (or a new `magic_cicd_test.go`) asserting:
```go
require.Equal(t, "cicd-workflow", cryptoutilSharedMagic.CICDCmdDirWorkflow)
require.Equal(t, "cicd-lint", cryptoutilSharedMagic.CICDCmdDirCicdLint)
```
This is a trivial test that prevents naming regressions. No logic to test — just value assertions.

---

#### 34. Categorize the Fitness Sub-Linter Catalog in Section 9.11.1

**Problem**: The 55 sub-linters in Section 9.11.1 are listed in registration order (neither alphabetical nor by category). As the list grows, it becomes harder to quickly find a specific linter or identify category gaps.

**Suggestion**: Organize the catalog with category sub-headers:

| Category | Example Linters |
|----------|----------------|
| Code Quality | `file-size-limit`, `magic-constants`, `cgo-ban` |
| Naming Conventions | `file-naming-conventions`, `package-naming`, `directory-naming` |
| Testing | `test-patterns`, `parallel-tests`, `table-driven-tests` |
| Architecture | `entity-registry`, `entity-registry-completeness`, `cmd-entry-whitelist` |
| Deployment | `compose-service-naming`, `admin-port-exposure`, `validate-ports` |
| Documentation | `docs-*`, `banned-product-names` |
| Infrastructure | `cmd-anti-pattern`, linters specific to INFRA-TOOL dirs |

This restructuring adds ~15 lines but significantly improves catalog navigability.

---

### Category J: Agent & Skill Gaps

#### 35. Missing Agents: `security-audit` and `coverage-boost`

**Current agents**: `beast-mode`, `fix-workflows`, `implementation-execution`, `implementation-planning`, `Explore`.

**Missing agents**:

| Suggested Agent | Purpose | Frequency |
|----------------|---------|-----------|
| `security-audit` | Orchestrates SAST + FIPS audit + gosec + govulncheck + DAST | Quarterly |
| `coverage-boost` | Analyzes coverage gaps and generates targeted tests | Per-release |
| `dependency-update` | Updates Go dependencies, checks CVEs, runs full tests | Monthly |

**Highest priority**: `security-audit` — security scanning is a complex multi-step workflow (FIPS audit → gosec → govulncheck → SAST → DAST → report) that benefits most from agent orchestration.

---

#### 36. Missing Skills: `deployment-gen` and `secret-gen`

**Current skills (14)**: agent-scaffold, agent-customization, contract-test-gen, coverage-analysis, fips-audit, fitness-function-gen, instruction-scaffold, migration-create, new-service, openapi-codegen, propagation-check, skill-scaffold, test-benchmark-gen, test-fuzz-gen, test-table-driven.

**Missing skills**:

| Suggested Skill | Purpose | Gap Being Filled |
|----------------|---------|-----------------|
| `deployment-gen` | Generate complete deployment structure for a new service | Currently error-prone manual process |
| `secret-gen` | Generate Docker secrets with correct format, naming, hex values | Wrong format is the #1 deployment mistake |
| `api-handler` | Map OpenAPI operation to strict server handler implementation | Reduces boilerplate copy errors |

**Highest priority**: `secret-gen` — wrong hex values in unseal secrets break HKDF derivation silently; a skill with format validation would prevent this.

---

#### 37. `fix-workflows.agent.md` Has 40+ References to Old `cmd/workflow`

**Problem**: After the `workflow` → `cicd-workflow` rename (#27), `.github/agents/fix-workflows.agent.md` will have 40+ stale references to `./cmd/workflow`. These are BLOCKING — the agent would generate incorrect commands after the rename.

**Status**: Tracked in rename scope (#27). This item is a reminder that the agent file is part of the rename PR, NOT a separate task.

---

### Category K: Section-Level Correctness (New Findings)

#### 38. Update INFRA-TOOL Table in Section 9.10 After Rename

**Location**: ARCHITECTURE.md Section 9.10 CICD Command Architecture, INFRA-TOOL table.

**Problem**: After the `workflow` → `cicd-workflow` rename (#27), the table listing `cicd-lint, workflow` must become `cicd-lint, cicd-workflow`.

**Action**: After completing #27, grep ARCHITECTURE.md for occurrences of `workflow` in INFRA-TOOL-related contexts (not general workflow references) and update them. ~8 occurrences.

---

#### 39. Section 2.1.2 Agent Catalog — Verify `Explore` Agent Entry

**Problem**: The `Explore` subagent was added after Feb 2026 and may not appear in ARCHITECTURE.md Section 2.1.2 Agent Catalog.

**Action**: Read Section 2.1.2, verify `Explore` is listed with its purpose and argument hint. Add if missing. Also verify `agent-customization` skill is in Section 2.1.1 Skills Catalog.

---

#### 40. Appendix B TBD Entries — Fill or Remove

**Problem**: 12 TBD cells remain in Appendix B across 4 sections:
1. **B.5 Agent Catalog**: `fix-workflows` and `beast-mode` have `Tools: TBD`, `Handoffs: TBD` — these can be filled from actual `.agent.md` files.
2. **B.6 CI/CD Workflow Catalog**: `ci-race`, `ci-benchmark`, `ci-sast`, `ci-dast`, `ci-e2e`, `ci-load` all have `Duration: TBD` — fill from workflow timeout values.
3. **B.7 Reusable Action Catalog**: "Additional actions | TBD | TBD | TBD" — fill or remove.
4. **B.8 Linter Rule Reference**: "(30+ total linters) | TBD | TBD | TBD" — fill from `lint_fitness.go`.

**Token impact**: Filling adds ~200–400 tokens. If brevity preferred, remove TBD rows entirely (empty rows add zero value).

---

### Category L: Propagation & Documentation Issues

#### 41. Verify SQLite+Barrier Rule Has `@propagate` Coverage

**Problem**: `03-04.data-infrastructure.instructions.md` documents "SQLite + Barrier Outside Transactions (CRITICAL)" with the specific deadlock explanation and correct pattern. This is also referenced as coming from ARCHITECTURE.md Section 5.2.4. If Section 5.2.4 does NOT have a `@propagate` block for this rule, then the instruction file contains un-validated content that can drift.

**Action**: Read ARCHITECTURE.md Section 5.2.4 and verify the SQLite+Barrier constraint has a `@propagate` marker sending it to `03-04.data-infrastructure.instructions.md`. Add if missing.

---

#### 42. Add `@source` Blocks to 2 Non-Propagated Instruction Files

**Problem**: `03-04.data-infrastructure.instructions.md` and `03-05.linting.instructions.md` still have zero `@source` blocks (confirmed in Feb 2026 suggestion #13 as NOT applied). Two clear candidates for propagation:
- `three-tier-database-strategy` → `03-04.data-infrastructure`
- `utf8-without-bom` → `03-05.linting`

**Action**: Apply Feb suggestion #13.

---

#### 43. Archive `docs/framework-v5/` and `docs/framework-v6/`

**Problem**: These directories contain implementation planning documents for completed framework development phases. Keeping them in `docs/` root adds cognitive overhead for humans and agents exploring the docs directory.

**Suggestion**: Move to `docs/ARCHIVE/framework-v5/` and `docs/ARCHIVE/framework-v6/`, or add a `README.md` in each stating "Historical planning for completed framework phase — do not apply."

---

#### 44. Link `ARCHITECTURE-TODO.md` from ARCHITECTURE.md

**Problem**: `docs/ARCHITECTURE-TODO.md` exists as a peer of ARCHITECTURE.md but is not referenced from ARCHITECTURE.md. Agents or developers may not know it exists.

**Suggestion**: Add a one-line reference near the Document Organization section: `For pending architecture improvements, see [ARCHITECTURE-TODO.md](ARCHITECTURE-TODO.md).`

---

#### 45. Anti-Pattern Documentation: Timeout Double-Multiplication

**Problem**: The anti-pattern `context.WithTimeout(ctx, magic.DefaultDataServerShutdownTimeout * time.Second)` producing a ~158-year timeout is documented in `03-02.testing.instructions.md` (under "Section 10.3.4 Test HTTP Client Patterns"). This should also be in ARCHITECTURE.md Section 10.3.4 with a `@propagate` marker so it's machine-verified.

**Suggestion**: Add the anti-pattern note with `@propagate` to ARCHITECTURE.md Section 10.3.4.

---

#### 46. Verify `DisableKeepAlives` Requirement Has `@propagate` Coverage

**Problem**: The `DisableKeepAlives: true` requirement for HTTP test transports (preventing 90-second Fiber shutdown hang) is documented in `03-02.testing.instructions.md` Section 10.3.4. If the corresponding `@propagate` block in ARCHITECTURE.md is missing, this critical rule has no drift protection.

**Action**: Verify ARCHITECTURE.md Section 10.3.4 has a `@propagate` block for `disable-keep-alives-test-transport`.

---

#### 47. Verify Sequential Test Exemption Rule Has `@propagate` Coverage

**Problem**: The `// Sequential: <reason>` comment exemption pattern (must appear within 10 lines before function declaration) is documented in `03-02.testing.instructions.md` Section 10.2.5. If this is not in a `@propagate` block, it has no drift protection.

**Action**: Verify ARCHITECTURE.md Section 10.2.5 has a `@propagate` block. Add if missing.

---

#### 48. Section 14 Operational Excellence — Add Expansion Note

**Problem**: Section 14 is 35 lines of placeholder content (monitoring, incident management, performance, capacity, disaster recovery) with no actionable detail for agents. At 0.9% of the document, it contributes almost nothing.

**Suggestion**: Add a brief note: "This section will be expanded when the operational monitoring stack (Prometheus alerts, Grafana dashboards, runbooks) is implemented in the codebase. Until then, see Section 9.4 Telemetry Strategy for current observability patterns."

This sets appropriate expectations without creating false completeness.

---

#### 49. Add Token Budget Quality Gate to ARCHITECTURE.md

**Problem**: Feb suggestion #18 (token budget warning) is still open. With the document now at 46.6% (improved from 56.8%), there is now headroom. But the lack of a documented target means it could creep back up as sections grow.

**Suggestion**: Add to ARCHITECTURE.md Section 11.4 Documentation Standards:
- ARCHITECTURE.md SHOULD stay below 4,500 lines (~55K tokens).
- Combined context budget (ARCHITECTURE.md + instruction files + copilot-instructions.md) SHOULD stay below 100K tokens (50% of 200K window).
- Document size is a tracked quality gate, measured during each `lint-docs` run.

---

#### 50. Section 12 Split Is Now More Urgent (25.8% of Document)

**Problem**: Section 12 was 18.9% of the document in Feb 2026. After framework-v6 restructuring, it is now **25.8%** — the largest section by a wider margin, and proportionally even more dominant. The split suggested in Feb (#19) is now a higher priority.

**Revised suggestion**: Split at 12.4 boundary:
- **Section 12: Deployment Architecture** (12.1–12.3): ~400 lines — patterns, pipelines, secrets.
- **Section 15: Deployment Tooling & Validation** (12.4–12.8): ~600 lines — linters, validators, config schema, documentation propagation.

Section 12.7 (Documentation Propagation) could also be extracted to a standalone Section 15 or Appendix, since it has no operational deployment dependency.

**Trade-off**: Requires renumbering Sections 13 and 14 → 14 and 16. All `#13xx` and `#14xx` anchors in cross-references must update. Significant but feasible in one focused pass.

---

## Summary Statistics (Updated)

### Feb 2026 Suggestions (#1–#26)

| Category | Count | Estimated Token Savings |
|----------|-------|------------------------|
| A: Token Budget Reduction | 7 | ~11,500–15,700 tokens |
| B: Correctness & Consistency | 5 | — |
| C: Completeness | 3 | +200 tokens |
| D: Quality Gates | 3 | — |
| E: Structural Improvements | 4 | ~-200 tokens |
| F: @propagate System | 2 | — |
| G: Miscellaneous | 2 | — |
| **Feb subtotal** | **26** | **~11,000–15,000 tokens net savings** |

### Mar 2026 New Suggestions (#27–#50)

| Category | Count | Primary Impact |
|----------|-------|----------------|
| H: INFRA-TOOL Rename & Convention | 3 | Naming consistency, fitness linter |
| I: Lint-Fitness Analysis | 5 | Overlap removal, new linters |
| J: Agent & Skill Gaps | 3 | New agents, new skills |
| K: Section-Level Correctness | 4 | TBD removal, catalog accuracy |
| L: Propagation & Documentation | 10 | Drift protection, doc debt |
| **Mar subtotal** | **24** | — |

### Combined

| Metric | Value |
|--------|-------|
| Total suggestions | 50 |
| Open (never applied) | ~44 |
| Applied/partially applied | ~6 (estimated) |
| High priority | 6 |
| Medium priority | 20 |
| Low priority | 24 |

---

## Prioritized Implementation Order (Updated March 2026)

### Phase 1 — Blocking / Immediate (must do now)

1. ✅ **#8** Fix Section 13.5 numbering (13.5.5 before 13.5.4) — renamed to 14.5.4/14.5.5.
2. ✅ **#11** Remove stale TBD entries from Appendix B — done.
3. ✅ **#12** Fix healthcheck numbered list (all `1.`) — done.
4. ✅ **#19** Split Section 12 (§12.4-12.8 → new §13 "Deployment Tooling & Validation") — done.
5. **#27** Execute `workflow` → `cicd-workflow` rename (blocks #29, #33, #38).
6. **#38** Update INFRA-TOOL table in Section 9.10 after rename (part of #27 commit).

### Phase 2 — Quality Gates (medium effort, high value)

1. **#28** Define INFRA-TOOL CLI pattern in Section 4.4.7.
2. **#29** Add `infra-tool-naming` fitness linter (post-rename).
3. **#16** Add propagation drift detection to pre-commit hooks.
4. **#41** Verify SQLite+Barrier rule has `@propagate` coverage.
5. ✅ **#13** Add `@source` blocks to non-propagated instruction files — done (03-04 three-tier-database-strategy added; 03-05 utf8-without-bom was already done).

### Phase 3 — Completeness (fill documented gaps)

1. ✅ **#11** Appendix B TBD entries filled from actual agent/workflow files — done.
2. **#39** Verify `Explore` agent and `agent-customization` skill in Section 2.1.
3. **#21** Add reading guide table to top of ARCHITECTURE.md.
4. **#35** Create `security-audit` agent.
5. **#36** Create `secret-gen` and `deployment-gen` skills.

### Phase 4 — Structural Improvements (moderate risk)

1. **#34** Categorize fitness linter catalog in Section 9.11.1.
2. **#20** Expand or move Section 14.
3. **#32** Remove hard count from Section 9.11.1 heading.
4. **#49** Add token budget quality gate to Section 11.4.

### Phase 5 — Token Budget Reduction (efficiency)

1. **#1** Extract secret format examples to satellite reference.
2. **#4** Reduce testing code examples.
3. **#6** Replace verbose directory trees with compact tables.
4. **#7** Compress authentication realm type table.
5. **#3** Compress fitness linter catalog section.

### Phase 6 — Tooling & Infrastructure (requires code changes)

1. **#5** Multi-target `@propagate` syntax support.
2. **#24** Propagation coverage metrics command.
3. **#31** Add `magic-constant-location` fitness linter.
4. **#33** Add regression test for INFRA-TOOL name constants.

### Phase 7 — Low-Priority Cleanup

1. **#9** Reconcile duplicate port assignment documentation.
2. ✅ **#10** Clarify `end-of-turn-commit-protocol` propagation target — NOTE comment added to §2.4.
3. ✅ **#14** Machine-readable `@propagate` grammar — BNF added to §13.4.2.
4. **#15** Anchor stability policy.
5. **#22** Cross-reference index for frequently-referenced topics.
6. **#23** Glue content maintenance strategy.
7. **#43** Archive framework-v5/v6 planning docs.
8. ~~**#44** Link ARCHITECTURE-TODO.md from ARCHITECTURE.md~~ — obsolete: ARCHITECTURE-TODO.md deleted.
9. **#45–#47** Add `@propagate` coverage for testing anti-patterns.
10. **#48** Add expansion note to Section 14.
11. ~~**#25** Remove or archive ARCHITECTURE-INDEX.md reference~~ — obsolete: ARCHITECTURE-INDEX.md deleted.
12. ~~**#26** Reconcile CONFIG-SCHEMA.md vs hardcoded schema~~ — obsolete: CONFIG-SCHEMA.md deleted, refs fixed.
