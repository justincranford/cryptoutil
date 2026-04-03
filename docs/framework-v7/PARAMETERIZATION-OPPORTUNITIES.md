# Parameterization Opportunities

**Status**: Part A complete (moved to PARAMETERIZATION-DONE.md). This file contains only
deferred, incomplete, missed, skipped, or newly discovered opportunities.
**Updated**: 2026-04-03
**See also**: [PARAMETERIZATION-DONE.md](PARAMETERIZATION-DONE.md) — all 18 completed items.

---

## Executive Summary

Two items remain deferred from the original 20. Seven new opportunities discovered via deep
analysis of the full project state — covering agent/skill drift enforcement gaps, multi-language
testing standards, Claude Code autonomous configuration, and additional linter coverage expansion.

| ID | Title | Priority | Status |
|----|-------|----------|--------|
| #02 | Generative Deployment Scaffold Command | Low | ⏸️ Deferred — violates cicd-lint constraints |
| #14 | Instruction File Slot Reservation Table | Low | ⏸️ Deferred — too little value at current scale |
| #21 | Claude Command YAML Frontmatter + Drift Validation | **CRITICAL** | ❌ Not implemented |
| #22 | Multi-Language Parameterized Testing | High | ❌ Not implemented |
| #23 | Copilot↔Claude Skill Body Content Drift | High | ❌ Not implemented |
| #24 | Claude Code Continuous Execution Configuration | Medium | ❌ Not documented |
| #25 | Agent Self-Containment Linter | Medium | ❌ Not implemented |
| #26 | ARCHITECTURE.md Section Link Validity | Medium | ❌ Not implemented |
| #27 | lint-go-test Expansion Beyond require-over-assert | Medium | ❌ Not implemented |

---

## Deferred Items (from original 20)

### #02 — Generative Deployment Scaffold Command ⏸️ DEFERRED

**Status**: Deferred — violates cicd-lint architectural constraints.

**Reason**: `cicd-lint` is exclusively for linting, formatting, and operational cleanup. A
`scaffold --ps-id=NEW-SVC` command violates both constraints:
- `--ps-id=NEW-SVC` is a customization parameter (violates subcommands-only rule)
- Generating Dockerfiles, compose files, config overlays, and secrets is content generation
  (violates no-generation rule)

**Alternative**: New services are onboarded manually. The `entity-registry-completeness`,
`deployment-dir-completeness`, and `config-overlay-freshness` fitness linters detect gaps
and missing artifacts after manual creation. The `skeleton-template` service provides
a copy-and-modify reference.

---

### #14 — Instruction File Slot Reservation Table ⏸️ DEFERRED

**Status**: Deferred — not enough value at current scale.

**Reason**: The `NN-NN.name.instructions.md` numbering convention is a lightweight implicit
scheme. A formal reservation table adds governance overhead without a clear benefit given the
current 18 files. May reconsider if the number grows or if numbering collisions occur.

**Alternative**: The existing `copilot-instructions.md` table serves as an informal slot
registry. Manual coordination suffices.

---

## New Discoveries: Not Yet Implemented

---

### #21 — Claude Command YAML Frontmatter + Drift Validation ❌ CRITICAL

**Impact**: HIGH — closes a critical gap in the artifact drift enforcement chain.

**Current state**: ALL 14 `.claude/commands/*.md` files are missing YAML frontmatter entirely.
The `lint-skill-command-drift` linter only validates:
1. Claude command file exists for each Copilot skill
2. Claude command body contains a reference to the Copilot skill path string

It does NOT validate:
- Whether the Claude command has YAML frontmatter at all
- Whether the `description` field matches the Copilot skill's `description`
- Whether the `argument-hint` field (if present in the skill) matches

This means the description visible to Claude Code users when browsing `/` commands is absent
or incorrect, AND there is zero automatic enforcement of description alignment between Copilot
skills and Claude commands.

**Examples of gap** (all 14 commands missing frontmatter):
```
.claude/commands/test-table-driven.md   — NO YAML frontmatter
.claude/commands/coverage-analysis.md  — NO YAML frontmatter
.claude/commands/fips-audit.md         — NO YAML frontmatter
... (11 more)
```

**Expected Claude command format** (mirroring skill `description` and `argument-hint`):
```yaml
---
name: test-table-driven
description: "Generate table-driven tests conforming to cryptoutil project standards. Use when
  writing or reviewing tests to ensure correct parallelism, typed test data, assertions, proper
  subtest structure, and TestMain for heavyweight resources."
argument-hint: "[package or function name]"
---

**Full Copilot original**: [.github/skills/test-table-driven/SKILL.md](...)
```

**Drift between Copilot skill `description` and Claude command `description`** constitutes
AI drift — Claude Code users see a different command description than Copilot users see for
the same logical operation.

**Required work**:
1. Add YAML frontmatter to all 14 `.claude/commands/*.md` files
2. Extend `CheckSkillCommandDrift()` in `docs_validation/skill_command_drift.go` to:
   - Detect missing YAML frontmatter in Claude command (error)
   - Extract `description` from skill YAML frontmatter
   - Extract `description` from Claude command YAML frontmatter
   - Report error if they differ
   - Extract `argument-hint` from skill YAML frontmatter if present
   - Report error if Claude command `argument-hint` differs
3. Update ARCHITECTURE.md Section 2.1.5 (Skill Catalog) to document that Claude commands
   MUST have YAML frontmatter with `name`, `description`, `argument-hint`
4. Update `06-02.agent-format.instructions.md` Section `lint-skill-command-drift` to document
   the frontmatter validation requirement
5. Update `skill-scaffold` SKILL.md and Claude command to document that the corresponding
   Claude command must have matching YAML frontmatter
6. Fix `lint-docs` to detect the 14 missing-frontmatter violations and exit non-zero

**Claude command `name` field convention**: The Claude command name should match the skill
directory name exactly (e.g., `name: test-table-driven`, NOT `name: claude-test-table-driven`
— commands use bare names, unlike agents which use the `claude-` prefix).

**Quantified scope**: 14 Claude command files + `skill_command_drift.go` extension +
ARCHITECTURE.md §2.1.5 + instruction file §06-02.

**Pre-requisites**: None — purely additive to existing infrastructure.

**Fitness enforcement**: Extended `lint-skill-command-drift` in `lint-docs`; exits non-zero
on missing frontmatter, description mismatch, or argument-hint mismatch.

---

### #22 — Multi-Language Parameterized Testing ❌ HIGH

**Impact**: HIGH — enforces parameterized (table-driven) patterns across ALL languages
used in the repository, not just Go.

**Current state**: The repository uses 3 languages for testing:
- **Go**: All 10 PS-ID services, CI-CD tooling, shared packages — covered by `test-table-driven`
  skill and `lint-go-test` (require-over-assert only) and fitness linters (parallel-tests, etc.)
- **Java** (Gatling API): `test/load/src/test/java/cryptoutil/*.java` — 4 simulation files;
  use Gatling Java API (io.gatling.javaapi), NOT JUnit; no linting beyond Maven build
- **Python**: Utility scripts via `pyproject.toml` which declares `pytest` + `pytest-cov` +
  `pytest-benchmark` + `pytest-mock`; no Python test files exist yet beyond ad-hoc scripts

**User intent**: Single skill file per artifact (not separate per language) to minimize drift
risk. Parameterize the content within the skill for all languages.

**Gap 1 — test-table-driven skill is Go-ONLY**:
The `test-table-driven` SKILL.md description says "Generate table-driven **Go** tests".
The Claude command description also says "Go tests". Both must be updated to cover all 3 languages.

**Gap 2 — No lint-java-test cicd-lint subcommand**:
Gatling Java simulation files should use `io.gatling.javaapi.core.CoreDsl.scenario()` with
`feed()` for parameterized scenarios and structured data feeders. No automatic enforcement exists.

**Gap 3 — No lint-python-test cicd-lint subcommand**:
Python tests must use `@pytest.mark.parametrize` for all multi-case tests. No enforcement exists.

**Gap 4 — ARCHITECTURE.md Section 10 is Go-centric**:
Section 10 (Testing Architecture) covers Go testing patterns comprehensively but has no
subsections covering Java or Python testing requirements.

**Required work**:

**A. Update `test-table-driven` skill + Claude command (single file, multi-language)**:
- Keep single `SKILL.md` (user instruction: avoid drift from multiple files)
- Retitle description: "Generate parameterized tests for Go, Java (Gatling), and Python
  conforming to cryptoutil project standards"
- Add H2 section per language:
  - **## Go** (existing content, no change)
  - **## Java (Gatling)**: `scenario()` with `feed()` for data feeders, parameterized feeders,
    `exec()` chains — show happy-path scenario and error-status scenario
  - **## Python (pytest)**: `@pytest.mark.parametrize` with both happy-path and sad-path params,
    `pytest.fixture` with `scope="session"` for heavyweight resources, `pytest-mock` for
    error injection

**B. Update `.claude/commands/test-table-driven.md`**:
- Add YAML frontmatter (see #21)
- Update description to match expanded multi-language skill description
- Update summary section to reference all 3 languages

**C. Update ARCHITECTURE.md Section 10**:
- Add Section 10.9 (Java Load Test Patterns) — Gatling parameterized scenarios, feeders,
  scenario naming convention (`{feature}-happy-path`, `{feature}-error-cases`)
- Add Section 10.10 (Python Test Patterns) — `@pytest.mark.parametrize`, `pytest.fixture`,
  `pytest-benchmark` for performance-sensitive operations, `scope="session"` for shared resources
- Add note in Section 10.1 that all 3 languages MUST use parameterized tests for multi-case coverage

**D. Add `lint-java-test` cicd-lint subcommand** (new `lint_javatest/` package):
- Scans `test/load/src/test/java/**/*.java` for Gatling simulation files
- Validates: each simulation uses `feed()` for data variation (not hardcoded single user),
  uses `pause()` for realistic think time, scenario names follow `{feature}-{variant}` convention
- Registered in `cicd.go` as `lint-java-test`

**E. Add `lint-python-test` cicd-lint subcommand** (new `lint_pytest/` package):
- Scans for `*.py` files in project (excluding test-output/)
- Validates: test functions use `@pytest.mark.parametrize` for multi-case tests,
  fixture scopes declared, `pytest-cov` configuration present in `pyproject.toml`
- Registered in `cicd.go` as `lint-python-test`

**F. Update ARCHITECTURE.md cicd-lint command table**:
- Add `lint-java-test` row
- Add `lint-python-test` row
- Update linter count from 11 to 13

**Quantified scope**: 1 skill SKILL.md expanded; 1 Claude command updated; 2 new cicd
subcommands (`lint-java-test`, `lint-python-test`); ARCHITECTURE.md 2 new subsections.

**Pre-requisites**: #21 (Claude command frontmatter must be in place before updating commands).

---

### #23 — Copilot Skill → Claude Command Body Content Drift ❌ HIGH

**Impact**: HIGH — closes the semantic drift gap between Copilot skill bodies and Claude
command summaries.

**Current state**: The `lint-skill-command-drift` linter validates that:
1. A Claude command exists for each Copilot skill
2. The Claude command body contains a reference to the skill path string
3. (After #21) The Claude command `description` matches the skill `description`

What it does NOT validate: whether the **key rules** in the Copilot skill are reflected in the
Claude command summary. A Copilot skill could add 5 new critical rules (e.g., "NEVER use
SpecificAPI"), update the template, and the Claude command body would remain stale — but
`lint-skill-command-drift` would still pass.

**The architectural relationship** (Copilot skill vs. Claude command) is intentionally asymmetric:
- **Copilot skill** (`.github/skills/NAME/SKILL.md`): FULL authoritative content — complete rules,
  multiple templates, detailed `## References` section
- **Claude command** (`.claude/commands/NAME.md`): SHORT summary — key rules only, 1 template,
  reference to full Copilot skill

This asymmetry means 1:1 body comparison (like agent drift) is WRONG for skills/commands.

**Proposed approach**: Each Copilot skill MUST declare its "Key Rules" in a normalized `## Key Rules`
H2 section. The Claude command MUST contain a `## Key Rules` section that references the same
rule bullets (as a subset, not verbatim copy). The drift linter checks:
1. Skill has `## Key Rules` H2 section
2. Claude command has `## Key Rules` H2 section
3. Every rule bullet in the Claude command also appears in the skill (command is a SUBSET)
4. If the skill's rule count grows significantly (more than 2 new bullets), flag as potentially stale

**Current rule section alignment audit** (all 14 pairs):
- `test-table-driven`: skill has `## Key Rules`, command has `## Rules` — ~aligned but heading inconsistent
- `coverage-analysis`: skill has `## Gap Categories`, command has no equivalent — DRIFT
- `fips-audit`: skill has `## BANNED`, command has `## BANNED` — roughly aligned
- `test-benchmark-gen`, `test-fuzz-gen`: skill has `## Rules`, command has `## Rules` — consistent
- `contract-test-gen`, `migration-create`, `new-service`, `openapi-codegen`, `propagation-check`,
  `skill-scaffold`, `agent-scaffold`, `instruction-scaffold`, `fitness-function-gen`: need audit

**Required work**:
1. Audit all 14 skill/command pairs for rule section alignment
2. Define normalization: each skill MUST have `## Key Rules`; commands MUST mirror it with same heading
3. Update all skills and commands to use normalized `## Key Rules` heading
4. Extend `CheckSkillCommandDrift()` to validate rule section presence and non-empty content
5. Update ARCHITECTURE.md §2.1.5 to document the `## Key Rules` normalization requirement
6. Update `skill-scaffold` SKILL.md to include `## Key Rules` in the template

**Quantified scope**: 14 skill/command pairs audited and aligned; linter extended.

**Pre-requisites**: #21 (frontmatter in place first).

---

### #24 — Claude Code Continuous Execution Configuration ❌ MEDIUM

**Impact**: MEDIUM — removes the operational gap between Copilot beast-mode and equivalent
continuous execution in Claude Code.

**Current state**: The repository has `claude-beast-mode` agent at `.claude/agents/beast-mode.md`
that instructs Claude Code to execute autonomously. However, there is NO documentation in
ARCHITECTURE.md, CLAUDE.md, or `docs/DEV-SETUP.md` explaining:
1. How to invoke Claude Code in a mode that does not pause for confirmation
2. What settings enable non-interactive autonomous operation
3. How to configure `.claude/settings.local.json` to reduce per-tool friction
4. What environment variables control Claude Code behavior

**Claude Code autonomous execution options** (require verification against current Claude docs):
- Invoke `/claude-beast-mode` as a slash command in Claude Code's chat
- Use `claude` CLI with `--dangerously-skip-permissions` for fully non-interactive runs
- `.claude/settings.local.json` can specify `"permissions": {"allow": [...]}` to pre-authorize tool invocations
- Use `claude --print` for single-turn non-interactive execution

**Required work**:
1. Research current Claude Code configuration options for autonomous/non-interactive execution
2. Add Section 14.9 (Claude Code Autonomous Execution) to ARCHITECTURE.md covering:
   - How to invoke beast-mode agent in Claude Code
   - `.claude/settings.local.json` permission pre-authorization patterns
   - CLI flags for non-interactive use
   - Comparison with Copilot beast-mode agent invocation
3. Add Claude Code autonomous execution guidance to CLAUDE.md
4. Update `.claude/settings.local.json` with recommended permissions configuration
5. Update `claude-beast-mode` agent to reference the new ARCHITECTURE.md section

**Quantified scope**: 1 new ARCHITECTURE.md section + CLAUDE.md update + settings file update.

**Pre-requisites**: None — documentation only.

---

### #25 — Agent Self-Containment Linter ❌ MEDIUM

**Impact**: MEDIUM — mechanically enforces the self-containment principle that every agent
must reference ARCHITECTURE.md.

**Current state**: Instruction file `06-02.agent-format.instructions.md` defines the
"Agent Self-Containment Checklist":
```
- Agents generating implementation plans MUST reference ARCHITECTURE.md testing (Section 10)
- Agents modifying code MUST reference coding standards (Sections 11, 14)
- Agents with ZERO ARCHITECTURE.md references are NON-COMPLIANT
```

`lint-agent-drift` validates equality between Copilot and Claude agent files, but does NOT
validate that the agent bodies actually contain ARCHITECTURE.md section references.

**Opportunity**: New `lint-agent-self-containment` sub-linter (part of `lint-docs`):
- Scans all `.github/agents/*.agent.md` files
- Extracts body content (after frontmatter)
- Checks: does body contain at least one `ARCHITECTURE.md` reference? (error if not)
- Reports: which agents are non-compliant

**Quantified scope**: 4 existing agents + any future agents; simple pattern check.

**Pre-requisites**: None.

**Fitness enforcement**: New `lint_agent_self_containment/` sub-linter added to `lint_docs.go`.

---

### #26 — ARCHITECTURE.md Section Link Validity ❌ MEDIUM

**Impact**: MEDIUM — prevents dangling `See [ARCHITECTURE.md Section X.Y](...)` references
that accumulate every time ARCHITECTURE.md sections are renumbered or renamed.

**Current state**: Instruction files contain ~200+ `See [ARCHITECTURE.md Section X.Y](...)` cross-reference
links. When ARCHITECTURE.md sections are renumbered, the instruction file references silently fail to
navigate to the right anchor. No linter detects this drift.

**Opportunity**: New `lint-docs section-link-validity` sub-linter:
1. Extract all H1–H4 headings from `docs/ARCHITECTURE.md` and compute their anchor IDs
   (GitHub Markdown anchor format: lowercase, hyphens for spaces, alphanumeric only)
2. Scan all `.github/instructions/*.instructions.md`, `.github/agents/*.agent.md`,
   `.github/skills/**/*.md`, and `.claude/agents/*.md` files for
   `](../../docs/ARCHITECTURE.md#ANCHOR)` patterns
3. Report any anchor that does not match a known heading in ARCHITECTURE.md

**Quantified scope**: ~200+ cross-reference links across 18 instruction files, 4 agents, 14 skills.

**Pre-requisites**: None — standalone.

**Fitness enforcement**: New `lint_architecture_links/` sub-linter added to `lint_docs.go`.

---

### #27 — lint-go-test Expansion Beyond require-over-assert ❌ MEDIUM

**Impact**: MEDIUM — closes gaps between what `lint-go-test` enforces and what the
architecture mandates for Go test files.

**Current state**: `internal/apps/tools/cicd_lint/lint_gotest/lint_gotest.go` has exactly
ONE registered linter (`require-over-assert`). The following mandatory architecture rules have NO
machine enforcement in either `lint-go-test` or fitness linters:

**Gap A — Hardcoded UUID detection** (ARCH §10.2 Forbidden pattern #4):
- Detects: `uuid.MustParse("00000000-...")` or `uuid.Parse("12345678-...")` with literal strings
  in test files
- Required: `googleUuid.NewV7()` for all test data IDs; only `googleUuid.UUID{}` and
  full-byte max UUID literal constructions are permitted for edge-case tests
- Currently: `import_alias_formula` enforces alias name but not the `MustParse` literal pattern

**Gap B — Real HTTP server detection in tests** (ARCH §10.2 Forbidden pattern #2):
- Detects: `httptest.NewServer(...)` in `*_test.go` files
- Required: `app.Test(req, -1)` (Fiber in-memory transport) for all handler tests
- Currently: No enforcement

**Gap C — `time.Sleep` in test files** (ARCH §10.2 timing targets):
- Detects: `time.Sleep(...)` in `*_test.go` files (prefer testcontainers wait strategies
  or `WaitForReady` over manual polling sleeps)
- Currently: No enforcement

**Required work**:
1. Add `lint_gotest/hardcoded_uuid/` sub-linter — forbid `uuid.MustParse` with literal strings
   in test files
2. Add `lint_gotest/real_http_server/` sub-linter — forbid `httptest.NewServer` in test files
3. Add `lint_gotest/test_sleep/` sub-linter — forbid `time.Sleep` in test files
4. Register all 3 in `registeredLinters` in `lint_gotest.go`
5. Update ARCHITECTURE.md §9.10 cicd-lint command table to reflect expanded lint-go-test

**Quantified scope**: 3 new sub-linters in `lint_gotest/`; combined coverage of all Go test files.

**Pre-requisites**: None.

---

## Documentation Gaps (Require ARCHITECTURE.md + Instruction File Updates)

The following are not separate parameterization items but are documentation accuracy gaps
discovered during this analysis:

### D1 — Claude Command YAML Frontmatter Not Documented

**Files**: ARCHITECTURE.md §2.1.5; `06-02.agent-format.instructions.md`
**Gap**: The `lint-skill-command-drift` section says only "Claude command references the skill
path string". Does not document that Claude commands MUST have YAML frontmatter or that
`description` must match.
**Fix**: Update §2.1.5 to specify Claude command YAML frontmatter format and required fields;
update `06-02.agent-format.instructions.md` `lint-skill-command-drift` paragraph accordingly.

### D2 — Section 10 Testing Architecture Has No Java or Python Coverage

**Files**: ARCHITECTURE.md §10
**Gap**: All testing patterns (t.Parallel, TestMain, Fiber app.Test, require over assert,
UUIDv7 data) are documented as Go-specific only. Gatling (Java) and Python (pytest) tests
have no documented architecture requirements in §10.
**Fix**: Add §10.9 Java Load Test Patterns + §10.10 Python Test Patterns. Add language note to §10.1.

### D3 — CLAUDE.md Does Not Document Autonomous Mode Invocation

**Files**: CLAUDE.md
**Gap**: CLAUDE.md lists the `claude-beast-mode` agent in the agents table, but there is no
documentation on HOW to invoke it in non-interactive/continuous mode (CLI flags, settings.local.json).
**Fix**: Add "Autonomous Execution" section or sidebar note explaining invocation options.

---

## Impact Ranking Summary

| Rank | ID | Title | Priority | Blocker? |
|------|----|-------|----------|----------|
| 1 | #21 | Claude Command YAML Frontmatter | CRITICAL | Yes — lint-docs should fail |
| 2 | #22 | Multi-Language Parameterized Testing | High | No |
| 3 | #23 | Copilot Skill → Claude Command Body Drift | High | No |
| 4 | #24 | Claude Code Continuous Execution Config | Medium | No |
| 5 | #25 | Agent Self-Containment Linter | Medium | No |
| 6 | #26 | ARCH.md Section Link Validity | Medium | No |
| 7 | #27 | lint-go-test Expansion | Medium | No |

**Recommended implementation order**: #21 first (blocking), then #22 (highest user value),
then #23 (completes skill/command drift chain), then #24–#27 and D-series in parallel.

---

## Dependency Graph (New Items)

```
#21 Claude Command Frontmatter
  ├── → #22 Multi-Language Testing (commands updated after frontmatter added)
  └── → #23 Skill Body Drift (description alignment enables rule alignment check)

#24 Claude Code Config ── standalone
#25 Agent Self-Containment ── standalone
#26 ARCH.md Link Validity ── standalone
#27 lint-go-test Expansion ── standalone

D1, D2, D3: Documentation fixes, no code dependencies
```

---

## Relationship to Existing Infrastructure

| New Item | Extends | Uses |
|----------|---------|------|
| #21 | `docs_validation/skill_command_drift.go` | Existing `splitMarkdownFrontmatter()` + `parseAgentFrontmatter()` |
| #22 | New `lint_javatest/`, `lint_pytest/` packages | Existing `cicd.go` registration pattern |
| #23 | `docs_validation/skill_command_drift.go` | Extended from #21; `## Key Rules` section parser |
| #24 | ARCHITECTURE.md §14, CLAUDE.md | `.claude/settings.local.json` |
| #25 | `lint_docs/lint_docs.go` | New sub-linter using existing `readFileFn` pattern |
| #26 | `lint_docs/lint_docs.go` | New sub-linter; ARCHITECTURE.md heading anchor parser |
| #27 | `lint_gotest/lint_gotest.go` | New sub-linters following existing `LinterFunc` interface |
