# ARCHITECTURE.md Improvement Suggestions

**Date**: 2026-03-28
**Status**: Review-first (do NOT apply without review)
**Scope**: Deep analysis of `docs/ARCHITECTURE.md`, Copilot instructions/agents/skills,
lint-fitness sub-linters, and INFRA-TOOL structure

---

## Document Profile (Current State — 2026-03-28)

| Metric | Value | vs. 2026-03-26 |
|--------|-------|----------------|
| Total lines | ~5,350 | +1,400 (§12 split added new §13/§14/§15 headings) |
| @propagate blocks | 39 | +1 (utf8-without-bom, three-tier-db added) |
| Fitness sub-linters | 55 | unchanged |
| Registered services | 10 | unchanged |
| Skills in §2.1.5 catalog | 12 | unchanged (3 missing vs actual 15) |
| Agents in §2.1.2 catalog | 4 | missing Explore (5 actual) |

**Largest sections (current)**:

| # | Section | Approx Lines | Notes |
|---|---------|-------------|-------|
| 12 | Deployment Architecture | ~550 | Split partner of new §13 |
| 13 | Deployment Tooling & Validation | ~530 | New—carved from old §12 |
| 10 | Testing Architecture | ~480 | Still largest single domain section |
| 9 | Infrastructure Architecture | ~420 | Contains 55-linter catalog |
| 6 | Security Architecture | ~350 | |
| 14 | Development Practices | ~300 | Was §13, now properly sized |
| 15 | Operational Excellence | ~65 | THIN — placeholder content only |

---

## Open Suggestions

### Category A: Token Budget Reduction

These suggestions reduce token consumption without losing information. Combined potential savings: ~10,000–14,000 tokens.

#### 1. Extract Deployment Secret Format Examples to a Satellite Reference

**Section**: 12.3.3 Secrets Coordination Strategy

**Problem**: ~200 lines of secret file format examples (exact hex format for unseal keys, base64 pepper values, sample PostgreSQL credential content, unseal-Nof5 naming) occupy high-cost space. Agents rarely need byte-level format during normal development — they need the rules (Docker secrets, chmod 440, never inline).

**Suggestion**: Extract format examples and sample values to `docs/SECRETS-REFERENCE.md`. Keep rules, naming conventions, tier strategy in ARCHITECTURE.md. Replace block with: `See [SECRETS-REFERENCE.md](SECRETS-REFERENCE.md) for secret file format examples and sample values.`

**Token savings**: ~2,500–3,000 tokens.

---

#### 2. Consolidate Three-Tier Deployment Hierarchy Compose Examples

**Section**: 12.3.4 Multi-Level Deployment Hierarchy

**Problem**: Three near-identical YAML blocks (SUITE, PRODUCT, SERVICE compose patterns) show the same structure with trivially different values. Port Offset Strategy table duplicates §3.4.1 (the +0/+10000/+20000 rule).

**Suggestion**: Replace with one parameterized template using `{TIER}` placeholders + a compact 3-row difference table. Delete the Port Offset Strategy table (fully covered by §3.4.1 Port Design Principles).

**Token savings**: ~1,500–2,000 tokens.

---

#### 3. Compress the Fitness Sub-Linter Catalog (Section 9.11.1)

**Section**: 9.11.1 Fitness Sub-Linter Catalog (55 total)

**Problem**: The 55-entry catalog (name, package, description, error/warning classification per entry) is ~200+ lines. Detailed per-linter information is valuable when authoring new linters but unlikely needed during normal agent sessions.

**Suggestion**: Keep a category-level summary table (8–10 rows: file-size, parallel-tests, banned-names, entity-registry, compose-validation, etc.) with counts per category. Move full 55-entry detail to a generated comment block in the registry or to `docs/FITNESS-LINTERS.md`.

**Token savings**: ~2,000–2,500 tokens.

---

#### 4. Reduce Code Examples in Section 10

**Section**: 10.1–10.8

**Problem**: Section 10 contains 13+ full Go code blocks. Many are also in `03-02.testing.instructions.md`, creating double exposure in agent context. ARCHITECTURE.md should hold strategy and rationale; instruction files should hold tactical examples.

**Suggestion**: Keep one canonical example per pattern (one table-driven test, one TestMain, one benchmark). Replace remaining blocks with brief descriptions + cross-references to the instruction file or actual test files in the codebase.

**Token savings**: ~2,000–3,000 tokens.

**Trade-off**: Agents in isolation mode (no instruction files) lose examples. Mitigated by agent self-containment checklist.

---

#### 5. De-duplicate @propagate Blocks with Multi-Target Syntax

**Section**: Throughout (39 markers)

**Problem**: The `mandatory-review-passes` chunk propagates to two targets (01-02 and 06-01) in separate `@propagate` blocks (lines 560 and 584). Same for `infrastructure-blocker-escalation` (lines 5053 and 5059). Each duplicated block repeats 300–600 tokens of identical content inside ARCHITECTURE.md.

**Suggestion**: Extend marker syntax to support comma-separated targets:
```html
<!-- @propagate to="01-02.beast-mode.instructions.md, 06-01.evidence-based.instructions.md" as="mandatory-review-passes" -->
```
Requires updating `lint-docs validate-propagation` parser to split the `to=` attribute on commas. Low complexity.

**Token savings**: ~1,200–1,500 tokens.

---

#### 6. Replace Verbose Deployment Directory Trees with Compact Tables

**Section**: 4.4.6 Deployments (ASCII art tree structures)

**Problem**: Three-tier deployment directory trees use ASCII art across ~150 lines. 14 secret files are listed 3× with trivially different filenames.

**Suggestion**: Replace with a single compact table (tier × component with counts), plus one fully expanded SERVICE-level tree as canonical example. Remove PRODUCT and SUITE trees.

**Token savings**: ~1,500–2,000 tokens.

---

#### 7. Compress the Authentication Realm Type Table (Section 7.2.1)

**Section**: 7.2.1 Authentication Realms

**Problem**: 23+ row table with near-identical token format variants (JWE/JWS/Opaque session cookie, JWE/JWS/Opaque session token = 6 nearly-identical rows) consuming ~100 lines.

**Suggestion**: Group by pattern:
```
Browser Session cookies: JWE | JWS | Opaque (browser only, SQL storage)
Non-browser Session tokens: JWE | JWS | Opaque (headless only, SQL storage)
MFA factors: TOTP | HOTP | WebAuthn | Push | Recovery (federated only, SQL storage)
```
Saves ~800–1,200 tokens with no information loss.

---

### Category B: Correctness & Stale References

#### 8. Fix Stale Section Numbers in `@propagate agent-self-containment` Block

**Location**: `docs/ARCHITECTURE.md` line 290–299 (propagated to `06-02.agent-format.instructions.md`)

**Problem**: The `agent-self-containment` @propagate block contains stale section references introduced by the §12→§13/§14/§15 split:

| Current (Wrong) | Correct |
|-----------------|---------|
| "coding standards (Section 13)" | §13 = Deployment Tooling; should say "Section 14" |
| "coding standards (Sections 11, 13)" | Should say "Sections 11, 14" |
| "deployment architecture (Section 12)" | Now spans §12 and §13; should say "Sections 12, 13" |
| "Section 12.7 (Documentation Propagation)" | §12.7 no longer exists; should say "Section 13.4" |

**Impact**: Agents following this instruction look at §13 (Deployment Tooling) for coding standards — completely wrong section. This bug survives `lint-docs` validation because BOTH the source block and the @source copy are identically wrong (in sync). The drift detector passes despite the stale content. **HIGH PRIORITY.**

**Fix**: Update the @propagate block in ARCHITECTURE.md, then re-propagate to `06-02.agent-format.instructions.md` (the @source block must match byte-for-byte).

---

#### 9. Fix Stale "Section 12.X" Display Text in 8+ Instruction Files

**Problem**: After the §12 split, many links in instruction files have correct anchors (resolve to §13.X targets) but wrong display text showing old "12.X" numbers. The `lint-docs` anchor validator reports these as valid (anchors exist), so the bug escapes automated detection. **HIGH PRIORITY.**

**Affected files**:

| File | Display Text | Correct Display |
|------|-------------|-----------------|
| `03-01.coding.instructions.md` | "Section 12.8 Validator Error Aggregation Pattern" | "Section 13.5" |
| `03-04.data-infrastructure.instructions.md` | "Section 12.5 Config File Architecture" | "Section 13.2" |
| `02-01.architecture.instructions.md` | "Section 12.5 Config File Architecture" | "Section 13.2" |
| `02-05.security.instructions.md` | "Section 12.6 Secrets Management in Deployments" | "Section 13.3" |
| `04-01.deployment.instructions.md` (line 55) | "Section 12.6 Secrets Management" | "Section 13.3" |
| `04-01.deployment.instructions.md` (line 211) | "Section 12.4.11" | "Section 13.1.11" |
| `04-01.deployment.instructions.md` (line 213) | "Section 12.6" | "Section 13.3" |
| `06-02.agent-format.instructions.md` (glue, line 41) | "Section 12.7" (no anchor) | "Section 13.4" |
| `.github/skills/propagation-check/SKILL.md` | "Section 12.7 Documentation Propagation Strategy" | "Section 13.4" |
| `.github/skills/instruction-scaffold/SKILL.md` | "Section 12.7 Documentation Propagation Strategy" | "Section 13.4" |

**Grep pattern** to find all: `Section 12\.` across `.github/**/*.md`

---

#### 10. Add Display-Text Accuracy Check to `lint-docs validate-propagation`

**Problem**: The propagation validator verifies anchor targets exist in ARCHITECTURE.md. It does NOT verify that the display text ("Section 12.7") matches the actual heading at that anchor ("Section 13.4 Documentation Propagation Strategy"). After any renumbering, links become "valid but misleading."

**Suggestion**: Add a sub-check that:
1. For each `[display text](docs/ARCHITECTURE.md#anchor)` link, extracts the anchor.
2. Looks up the first heading at that anchor in ARCHITECTURE.md.
3. Emits a WARNING if a section number in the display text does not match the anchor's section number.

**Complexity**: Medium. The heading-to-anchor mapping already exists in the propagation validator; needs a display-text extractor and fuzzy number comparison.

---

#### 11. Anchor Stability Policy

**Section**: Document Metadata or §2

**Problem**: Many instruction files and agents reference numbered anchors (e.g., `#134-documentation-propagation-strategy`). Section splits broke all cross-references to moved sections. There is no documented policy for anchor stability or renumbering procedures.

**Suggestion**: Add a brief policy:
- MUST grep `.github/**/*.md` and `docs/**/*.md` for old anchor patterns after any renumbering.
- SHOULD prefer stable named anchors (`#documentation-propagation-strategy`) over numbered anchors for sections likely to be referenced from agent/skill/instruction files.
- The `lint-docs validate-propagation` broken-reference check is the safety net — treat any broken references as blocking.

---

#### 12. Config File Count Validation in Deployment Linter

**Section**: 13.1.5 Config File Naming Strategy

**Problem**: The naming strategy requires exactly 5 config overlay files per service (`{PS-ID}-app-{common,sqlite-1,sqlite-2,postgresql-1,postgresql-2}.yml`). The linter validates naming patterns but does not enforce the count. A service missing `sqlite-2.yml` passes naming validation while being incomplete for deployment.

**Suggestion**: Add a completeness check to `lint-deployments validate-naming`: verify that each `deployments/{PS-ID}/config/` directory contains all 5 required overlay files. Emit an error if any are missing.

---

#### 13. Regression Test for `cicd-workflow` Magic Constant

**Location**: `internal/shared/magic/magic_cicd.go`

**Problem**: After the `workflow` → `cicd-workflow` rename, `CICDCmdDirWorkflow = "cicd-workflow"`. There is no test asserting the expected string value. A future refactor could silently change it.

**Suggestion**: Add a test in `magic_cicd_test.go` (or `cmd_anti_pattern_test.go`):
```go
require.Equal(t, "cicd-workflow", cryptoutilSharedMagic.CICDCmdDirWorkflow)
require.Equal(t, "cicd-lint", cryptoutilSharedMagic.CICDCmdDirCicdLint)
```
Trivial test, prevents naming regressions.

---

### Category C: Completeness Gaps

#### 14. Update §2.1.2 Agent Catalog and §2.1.5 Skill Catalogue

**Location**: `docs/ARCHITECTURE.md` §2.1.2 (line 301) and §2.1.5

**Problem — Agents**: The `Explore` subagent exists in `.github/copilot-instructions.md` available-agents but is absent from ARCHITECTURE.md §2.1.2 Agent Catalog. Confirmed absent via grep.

**Problem — Skills**: The catalog lists 12 skills. Three skills exist in `.github/skills/` and appear in `.github/copilot-instructions.md` but are absent from §2.1.5:

| Missing Skill | File | Purpose |
|--------------|------|---------|
| `contract-test-gen` | `.github/skills/contract-test-gen/SKILL.md` | Generate cross-service contract compliance tests |
| `fitness-function-gen` | `.github/skills/fitness-function-gen/SKILL.md` | Create new architecture fitness function linters |
| `agent-customization` | `copilot-skill:/agent-customization/SKILL.md` | Create/update VS Code customization files (built-in skill) |

**Action**: Add `Explore` to §2.1.2. Add 3 missing skills to §2.1.5. Note: `agent-customization` uses the built-in `copilot-skill://` scheme (not a `.github/skills/` file) — document this distinction.

---

#### 15. Add Missing Agents

**Current agents** (5): `beast-mode`, `fix-workflows`, `implementation-execution`, `implementation-planning`, `Explore`.

**Missing agents**:

| Suggested Agent | Purpose | Priority |
|----------------|---------|----------|
| `security-audit` | Orchestrates FIPS audit → gosec → govulncheck → SAST → DAST → consolidated report | High |
| `coverage-boost` | Analyzes coverage gaps and generates targeted tests to reach ≥95% thresholds | Medium |
| `dependency-update` | Updates Go dependencies, checks CVEs, runs full tests | Medium |

**Highest priority**: `security-audit` — security scanning is a complex multi-step workflow that benefits most from agent orchestration.

---

#### 16. Add Missing Skills

**Current skills** (15): agent-scaffold, agent-customization, contract-test-gen, coverage-analysis, fips-audit, fitness-function-gen, instruction-scaffold, migration-create, new-service, openapi-codegen, propagation-check, skill-scaffold, test-benchmark-gen, test-fuzz-gen, test-table-driven.

**Missing skills**:

| Suggested Skill | Purpose | Priority |
|----------------|---------|----------|
| `deployment-gen` | Generate complete deployment structure for a new service (compose.yml, Dockerfile, secrets/, config/) | Medium |
| `secret-gen` | Generate Docker secrets with correct format, naming, hex/base64 values, and tier prefix | **High** |
| `api-handler` | Map OpenAPI operation to strict server handler implementation boilerplate | Medium |

**Highest priority**: `secret-gen` — wrong hex values in unseal secrets break HKDF derivation silently. A skill with format validation prevents the #1 deployment mistake.

---

#### 17. Expand §15 Operational Excellence

**Section**: §15 (65 lines across 5 subsections)

**Problem**: §15 is 65 lines total for Monitoring, Incident Management, Performance, Capacity Planning, and DR. Each subsection is 3–6 bullet points — placeholder text with no actionable detail for agents.

**Suggestion** (choose one):
- (a) Expand with detail: alert thresholds, Grafana dashboard IDs, incident escalation workflow, runbook template, DR drill checklist. Right answer when operational work exists.
- (b) Add a note: "This section will be expanded when operational monitoring is implemented. See §9.4 Telemetry Strategy for current observability patterns." Minimal, honest, stops agents from treating placeholders as specifications.
- (c) Move to Appendix D as a compact reference, freeing the section number.

**Recommendation**: Option (b) now, evolve to (a) as the product matures.

---

#### 18. Add a Document Map / Reading Guide

**Section**: Top of document (after Document Organization)

**Problem**: At 5,350+ lines with 15 sections, navigating ARCHITECTURE.md requires knowing the numbering scheme. No reading guide maps user intent to sections.

**Suggestion**: Add a concise reading guide table (~15 lines) after the Document Organization section:

| If you need to... | Read section(s) |
|---------------------|----------------|
| Understand the product suite | 1, 3 |
| Add a new service | 4.4, 5.1, 5.2 |
| Write or fix tests | 10 |
| Configure deployment | 12, 13.1 |
| Understand secrets management | 6.10, 12.3.3, 13.3 |
| Add or modify API endpoints | 8 |
| Fix CI/CD or linting | 9, 11.3 |
| Debug auth or sessions | 6.9, 7.2 |
| Understand the @propagate system | 13.4 |

---

#### 19. Add Cross-Reference Index for Frequently-Referenced Topics

**Section**: Appendix (new Appendix D, or inline in Document Organization)

**Problem**: Topics like secrets, TLS, port assignments, and health checks are referenced from 4+ separate sections. No central navigation aid.

**Suggestion**: Add a compact cross-reference index (~20 lines):

| Topic | Primary | Also In |
|-------|---------|---------|
| Secrets | 12.3.3 | 6.10, 13.3, 02-05.security |
| TLS / mTLS | 6.11, 6.5 | 5.3, 12.3.3 |
| Ports | 3.4, 3.4.1 | 02-01.architecture |
| Health checks | 5.5 | 02-01.architecture |
| Testing database tiers | 10.1 | 03-02.testing, 03-04.data-infra |
| @propagate system | 13.4 | 06-02.agent-format |

---

#### 20. Propagation Coverage Metrics Sub-Command

**Section**: §13.4 Propagation Strategy

**Problem**: §13.4.7 references propagation coverage percentages as manual accounting that drifts as files are added.

**Suggestion**: Add `lint-docs propagation-coverage` sub-command that reports:
- Total instruction files vs. files with ≥1 `@source` block (percentage)
- Total lines in instruction files vs. lines inside `@source` blocks (percentage)
- Identifies zero-coverage files by name

This makes coverage a measurable, continuously-tracked metric rather than a manually updated claim.

---

#### 21. Token Budget Guideline in §11.4 Documentation Standards

**Section**: 11.4 Documentation Standards

**Problem**: No maximum size target for ARCHITECTURE.md. After the §12 split added new §13/§14/§15 content, the document grew back toward ~5,350 lines. Without a documented budget, the document can grow unbounded.

**Suggestion**: Add a token budget guideline:
- ARCHITECTURE.md SHOULD stay below 60K tokens (~4,800 lines).
- Combined context budget (ARCHITECTURE.md + 18 instruction files + copilot-instructions.md) SHOULD stay below 100K tokens (50% of 200K window).
- When approaching limits, apply extraction strategies: satellite docs for format examples, generated catalogs, code-as-source-of-truth.
- Add `lint-docs token-budget` sub-command that measures and warns when thresholds are breached.

---

### Category D: Fitness & Quality Gates

#### 22. Define INFRA-TOOL CLI Pattern in §4.4.7

**Section**: §4.4.7 CLI Patterns

**Problem**: §4.4.7 defines PRODUCT, PRODUCT-SERVICE, and SUITE patterns. `cicd-lint` and `cicd-workflow` are referenced in examples but are never formally defined as an INFRA-TOOL CLI pattern type. No rules exist about naming, location, or entry function signature.

**Suggestion**: Add an "INFRA-TOOL Pattern" subsection:
```
INFRA-TOOL Pattern: cmd/cicd-{tool}/main.go → internal/apps/tools/cicd_{tool}/cicd_{tool}.go
- ALL INFRA-TOOL cmd dirs MUST be prefixed cicd-
- Internal package dirs use underscore: cicd_{tool}/
- NOT registered in entity registry (not a product-service)
- Whitelisted in cmd-entry-whitelist fitness linter
```

---

#### 23. Add `infra-tool-naming` Fitness Linter

**Problem**: `cmd-entry-whitelist` whitelists known INFRA-TOOLs but does NOT enforce that new INFRA-TOOLs follow the `cicd-*` naming convention. A developer could add `cmd/release-tool/` and bypass all naming checks.

**Suggestion**: Add a fitness linter that:
1. Identifies all `cmd/` entries not matching product/service/suite patterns.
2. Validates those entries are prefixed with `cicd-`.
3. Validates matching `internal/apps/tools/` entries use `cicd_` prefix.
4. Emits an error otherwise.

**Prerequisite**: Suggestion #22 (INFRA-TOOL pattern definition) should be documented first.

---

#### 24. Investigate `admin-port-exposure` and `validate-admin` Linter Overlap

**Problem**: Two fitness linters may check overlapping concerns:
- `admin-port-exposure`: checks admin port binding in compose files.
- `validate-admin`: validates admin port configuration.

If both verify that the admin port binds to `127.0.0.1:9090`, one is redundant.

**Action**: Read both implementations (`lint_fitness/admin_port_exposure/` and `lint_fitness/validate_admin/`), determine if there is genuine duplication, and either consolidate or clearly differentiate their responsibilities in §9.11.1.

---

#### 25. Add `magic-constant-location` Fitness Linter

**Problem**: `internal/shared/magic/` is the mandatory location for all magic constants. The `mnd` golangci-lint linter catches inline literals but does NOT catch package-local `const x = ...` declarations (which bypass `mnd` entirely). A developer can create `internal/myservice/constants.go` and add package-local string/int constants without triggering any lint warning.

**Suggestion**: Add a fitness linter that scans all non-`magic/` packages for `const` declarations containing suspicious values (port-like integers, algorithm name strings like `"AES"`, `"RSA"`, timeout durations). Initial focus: `const` with integer values between 1000–65535 (port range) in non-test, non-generated files.

---

#### 26. Remove Hard Count from §9.11.1 Heading

**Location**: `docs/ARCHITECTURE.md` line 2686

**Problem**: The heading `#### 9.11.1 Fitness Sub-Linter Catalog (55 total)` hardcodes the linter count. When a linter is added, the doc must be manually updated. When it drifts, the document is misleading.

**Suggestion**: Change heading to `#### 9.11.1 Fitness Sub-Linter Catalog` (drop the count). Add a CI assertion that `len(registeredLinters) == count of lint_fitness subdirs excluding registry/`. The code becomes the source of truth.

---

#### 27. Categorize the Fitness Sub-Linter Catalog

**Section**: §9.11.1

**Problem**: 55 sub-linters are listed in registration order. Finding a specific linter or identifying category coverage gaps requires scanning the entire list.

**Suggestion**: Organize with category sub-headers:

| Category | Example Linters | Count |
|----------|----------------|-------|
| Code Quality | `file-size-limit`, `magic-constants`, `cgo-ban` | ~10 |
| Naming Conventions | `file-naming-conventions`, `package-naming` | ~8 |
| Testing | `test-patterns`, `parallel-tests`, `literal-use` | ~8 |
| Architecture | `entity-registry`, `entity-registry-completeness`, `cmd-entry-whitelist` | ~6 |
| Deployment | `compose-service-naming`, `admin-port-exposure`, `validate-ports` | ~6 |
| Documentation | `docs-*`, `banned-product-names` | ~5 |
| Infrastructure | `cmd-anti-pattern`, `infra-tool-naming` | ~4 |

---

### Category E: Propagation Gaps

These rules exist in the codebase and instruction files but lack `@propagate` protection. An ARCHITECTURE.md change can silently break the instruction without the drift detector firing.

#### 28. Add @propagate for SQLite+Barrier Rule (§5.2.4)

**Location**: `docs/ARCHITECTURE.md` line 1373 (§5.2.4 Database Compatibility Rules)

**Problem**: The SQLite+Barrier outside-transactions rule is documented in §5.2.4 with full root-cause explanation and the correct pattern (`ORM.Create → commit → barrier.Encrypt → ORM.Update`). The rule is repeated in `03-04.data-infrastructure.instructions.md` but there is no `@propagate` block protecting it from drift.

**Action**: Add `@propagate to=".github/instructions/03-04.data-infrastructure.instructions.md" as="sqlite-barrier-outside-tx"` around the SQLite+Barrier subsection in §5.2.4. Update the corresponding `@source` block in `03-04.data-infrastructure.instructions.md`.

---

#### 29. Add @propagate for §10.3.4 Critical Test Patterns

**Location**: `docs/ARCHITECTURE.md` line 3075 (§10.3.4 Test HTTP Client Patterns) and §10.2.5

**Problem**: §10.3.4 and §10.2.5 document three critically important rules in `03-02.testing.instructions.md`:
1. **`DisableKeepAlives: true`** requirement (prevents 90-second Fiber shutdown hang)
2. **Timeout double-multiplication anti-pattern** (`time.Duration * time.Second` = ~158-year timeout)
3. **Sequential test exemption** (`// Sequential: <reason>` comment pattern at §10.2.5)

None have `@propagate` blocks. All are mentioned in the instruction file without drift protection.

**Action**: Add `@propagate` blocks to §10.3.4 and §10.2.5:
- `as="disable-keep-alives-test-transport"` → `03-02.testing.instructions.md`
- `as="timeout-double-multiplication-antipattern"` → `03-02.testing.instructions.md`
- `as="sequential-test-exemption"` → `03-02.testing.instructions.md`

---

#### 30. Add README to `docs/framework-v3/` through `docs/framework-v6/`

**Problem**: Four directories (`framework-v3/`, `framework-v4/`, `framework-v5/`, `framework-v6/`) contain completed planning documents. None have a `README.md`. Developers exploring `docs/` encounter these without context.

**Suggestion**: Add a `README.md` to each directory explaining "Historical planning document for completed framework phase X — do not apply; retained for reference only." Alternatively, move all four to `docs/ARCHIVE/` with a single `README.md` covering all phases.

---

#### 31. Resolve Remaining Appendix B.7 TBD

**Location**: `docs/ARCHITECTURE.md` B.7 Reusable Action Catalog (~line 5294)

**Problem**: One TBD row remains: `Additional actions | TBD | TBD | TBD`. TBD rows consume tokens without providing value.

**Action**: Either fill in actual reusable actions from `.github/actions/` (beyond `docker-images-pull`) or remove the TBD row entirely.

---

## Prioritized Implementation Order

### Phase 1 — Blocking / Immediate

1. **#8** Fix @propagate `agent-self-containment` block stale section references (§13→§14, §12.7→§13.4) — actively misdirects agents to wrong section for coding standards.
2. **#9** Fix stale "Section 12.X" display text in 8+ instruction and skill files.

### Phase 2 — Quality Gates (medium effort, high value)

1. **#28** Add @propagate for SQLite+Barrier rule (§5.2.4).
2. **#29** Add @propagate for §10.3.4 critical test patterns.
3. **#14** Update §2.1.2 Agent Catalog (add Explore) and §2.1.5 Skill Catalogue (add 3 missing skills).
4. **#12** Add config file count validation to deployment linter.
5. **#13** Add regression test for cicd-workflow magic constant.

### Phase 3 — Completeness (fill documented gaps)

1. **#18** Add reading guide table to top of ARCHITECTURE.md.
2. **#17** Add expansion note to §15 Operational Excellence.
3. **#26** Remove hard count from §9.11.1 heading.
4. **#15** Create `security-audit` agent.
5. **#16** Create `secret-gen` skill.

### Phase 4 — Structural Improvements

1. **#22** Define INFRA-TOOL CLI pattern in §4.4.7.
2. **#23** Add `infra-tool-naming` fitness linter.
3. **#27** Categorize fitness sub-linter catalog in §9.11.1.
4. **#10** Add display-text accuracy check to lint-docs.
5. **#11** Add anchor stability policy.

### Phase 5 — Token Budget Reduction

1. **#1** Extract secret format examples to satellite reference.
2. **#4** Reduce testing code examples in §10.
3. **#6** Replace verbose directory trees with compact tables.
4. **#7** Compress authentication realm type table.
5. **#3** Compress fitness linter catalog section.
6. **#5** Multi-target @propagate syntax (code change required).

### Phase 6 — Low-Priority Cleanup

1. **#19** Add cross-reference index for frequently-referenced topics.
2. **#20** Add propagation coverage metrics sub-command (code change required).
3. **#21** Add token budget guideline to §11.4 Documentation Standards.
4. **#24** Investigate admin port linter overlap.
5. **#25** Add `magic-constant-location` fitness linter.
6. **#30** Add README files to all four framework-vX directories.
7. **#31** Resolve remaining Appendix B.7 TBD row.
