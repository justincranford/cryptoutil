# Implementation Plan - Framework V18: ENG-HANDBOOK.md Knowledge Propagation

**Status**: Planning
**Created**: 2026-04-26
**Last Updated**: 2026-04-26
**Purpose**: Propagate implementation-specific details from four reference docs
(`target-structure.md`, `tls-structure.md`, `deployment-templates.md`, `claude-structure.md`)
into the canonical `ENG-HANDBOOK.md`, plus one targeted fix to `tls-structure.md` for the
Admin CA Bundle section. After completion, the four suggestion docs are deleted.

---

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL additions accurate; no copy-paste errors |
| Completeness | ALL suggestion items addressed; NO items skipped |
| Thoroughness | Evidence-based validation at every step |
| Reliability | `lint-docs` and `lint-fitness` clean after every phase |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Overview

V18 is a documentation-only plan. No Go code is added. The work is:

1. **ENG-HANDBOOK.md additions** from four reference docs — content that exists in the
   reference docs but is missing from the handbook's canonical sections.
2. **tls-structure.md fix** — add `--cert`/`--key`/`--ca-cert` CLI flag documentation to the
   Admin CA Bundle section (Item 1 in `tls-structure-suggestions.md` is PARTIAL).

The plan is organized into five phases:
- **Phase 1**: ENG-HANDBOOK.md additions from `target-structure.md` (11 items → §2.1, §4.4, §B.7, §11.1.4, §12.2.1)
- **Phase 2**: `tls-structure.md` fix + ENG-HANDBOOK.md additions from `tls-structure.md` (5 items → §6.5, §6.11)
- **Phase 3**: ENG-HANDBOOK.md additions from `deployment-templates.md` (11 items → §12.2.1, §12.3.1, §12.3.3, §12.3.5, §13.2, §13.6)
- **Phase 4**: ENG-HANDBOOK.md additions from `claude-structure.md` (11 items → §2.1.1, §2.1.5, §14.11)
- **Phase 5**: Knowledge Propagation verification (`lint-docs`)

---

## Background

### Source Documents (Reference Docs)

| Doc | Purpose | ENG-HANDBOOK.md Target |
|-----|---------|----------------------|
| `docs/target-structure.md` | Concrete file/dir inventory, permissions, catalogs | §2.1, §4.4, §B.7, §11.1.4, §12.2.1 |
| `docs/tls-structure.md` | PKI cert categories, naming, count formulas, TLS modes | §6.5, §6.11 |
| `docs/deployment-templates.md` | Template rules (DF/CO/CF/SC/PC/SU), parameterization | §12.2.1, §12.3.1, §12.3.3, §12.3.5, §13.2, §13.6 |
| `docs/claude-structure.md` | Claude Code `.claude/` structure, skill/agent fields | §2.1.1, §2.1.5, §14.11 |

### Suggestion Docs (Deleted After V18 Complete)

Four suggestion docs were created to track documentation gaps. After V18 completes all items,
the suggestion docs are deleted since all remaining work is captured in ENG-HANDBOOK.md:

- `docs/tls-structure-suggestions.md` — Items 1–5 actionable; Items 6–7 obsolete (V12/V13 phases complete)
- `docs/target-structure-suggestions.md` — All 11 items actionable
- `docs/deployment-templates-suggestions.md` — All 11 items actionable
- `docs/claude-structure-suggestions.md` — All 11 items actionable

---

## Phase 1: ENG-HANDBOOK.md Additions from target-structure.md

**Phase Objective**: Add 11 missing catalog entries, tables, and inventory sections from
`target-structure.md` into the appropriate ENG-HANDBOOK.md sections.

| Task | Item | ENG-HANDBOOK.md Section | Content |
|------|------|------------------------|---------|
| 1.0 | Pre-flight | — | Build health + lint-fitness baseline |
| 1.1 | File Permission Convention Table | §4.4.1 | octal permission table (640/440/750/etc.) |
| 1.2 | Root-Level File Inventory | §4.4.1 | 23+ root config files that MUST exist |
| 1.3 | Root Hidden Directory Inventory | §4.4 | `.cicd-lint/`, `.semgrep/`, `.vscode/`, `.well-known/`, `.zap/` |
| 1.4 | .github/ Top-Level File Catalog | §2.1 | copilot-instructions.md, dependabot.yml, SECURITY.md, etc. |
| 1.5 | GitHub Actions Catalog | §B.7 | 15 reusable actions with purpose descriptions |
| 1.6 | Concrete Service Subdirectory Inventory | §4.4.4 | Per-PS-ID actual subdirectories (e.g., pki-ca 15+ dirs) |
| 1.7 | Identity Shared Package Catalog | §4.4.4 | `internal/apps/identity/` shared packages (9 packages) |
| 1.8 | Complete magic/ File Listing | §11.1.4 | 42 `magic_*.go` files organized by domain |
| 1.9 | Other Top-Level Directories | §4.4.1 | `scripts/`, `workflow-reports/`, `test-output/`, `pkg/` |
| 1.10 | Dockerfile Parameterization Table | §12.2.1 | Per-tier image.title, binary, EXPOSE, HEALTHCHECK, ENTRYPOINT |
| 1.11 | Pending Work Inventory | §4.4.6 | Product Dockerfiles missing (all 5 products) |
| 1.12 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

---

## Phase 2: tls-structure.md Fix + ENG-HANDBOOK.md Additions from tls-structure.md

**Phase Objective**: Fix the partial Admin CA Bundle documentation in `tls-structure.md`, then add
5 items from `tls-structure.md` into ENG-HANDBOOK.md §6.5 and §6.11.

Items 6–7 from `tls-structure-suggestions.md` are **OBSOLETE** (V12/V13 phases complete;
`pki-init-order.md` is now a standalone doc at `docs/pki-init-order.md`).

| Task | Item | Target Doc | Content |
|------|------|-----------|---------|
| 2.0 | Admin CA Bundle fix | `tls-structure.md` | Add `--cert`/`--key`/`--ca-cert` flags to Policy Alignment section |
| 2.1 | Admin CA Bundle → ENG-HANDBOOK.md | §6.5 (PKI Architecture) | Admin mTLS client trust requirement |
| 2.2 | tls-config.yml Dynamic Cert Pattern | §6.11 (TLS Config) | TLSModeAutoGenerate/PreGenerated/Mixed table; SAME-AS-DIR-NAME convention |
| 2.3 | Realm Dynamic Binding (Decision 8) | §6.5 | pki-init realm list from registry.yaml; per-PS-ID realm defaults |
| 2.4 | postgres vs postgres-1/2 Naming | §6.11 | Shared domain vs per-instance naming rationale |
| 2.5 | Directory Count Formula Derivation | §6.11 or tls-structure.md | 26 global + 64 per-PS-ID × 10 = 630 formula with Cat 9 correction |
| 2.6 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

---

## Phase 3: ENG-HANDBOOK.md Additions from deployment-templates.md

**Phase Objective**: Add 11 items covering parameterization tables, Dockerfile/Compose/config
enforceable rule catalogs, PostgreSQL mTLS cert reference, inconsistency inventory, and template
syntax spec into ENG-HANDBOOK.md §12–§13.

| Task | Item | ENG-HANDBOOK.md Section | Content |
|------|------|------------------------|---------|
| 3.0 | Pre-phase lint baseline | — | Verify lint-docs clean before changes |
| 3.1 | Complete Parameterization Table | §13.6 or §13.7 | Entity, port, build/container params; PS-ID matrix |
| 3.2 | Container UID/GID Security Rationale | §12.2.1 | UID 65532 rationale, ARG parameterization, debug override |
| 3.3 | Dockerfile Rules DF-01–DF-24 | §12.2.1 | 24 machine-checkable Dockerfile rules |
| 3.4 | Compose Rules CO-01–CO-22 | §12.3.1 | 22 compose rules including named-volume mandate |
| 3.5 | Deployment Config Rules CF-01–CF-17 | §13.2 | 17 config overlay rules (incl. PostgreSQL mTLS cert paths) |
| 3.6 | Standalone Config Rules SC-01–SC-06 | §13.2 | 6 standalone config rules (127.0.0.1 bind, port consistency) |
| 3.7 | Product/Suite Compose Rules PC/SU | §12.3.5 | PC-01–PC-06 (product) and SU-01–SU-04 (suite) rules |
| 3.8 | Secret File Value Patterns | §12.3.3 | Complete 14-secret table with exact value format patterns |
| 3.9 | PostgreSQL mTLS Cert Reference Table | §6.11.4 | Per-node cert ownership; PKI Cat 10–14 reference |
| 3.10 | Current Inconsistency Inventory | §13.6 or Appendix M | Dockerfile Pattern A/B/C bugs; config snake_case PS-IDs |
| 3.11 | Template Syntax Specification | §13.6 | `__KEY__` format; path-level vs content-level; template catalog |
| 3.12 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

---

## Phase 4: ENG-HANDBOOK.md Additions from claude-structure.md

**Phase Objective**: Add 11 items covering `.claude/` directory structure, CLAUDE.md format,
skill/agent frontmatter fields, dynamic context injection, path-scoped rules, and agentskills.io
open standard into ENG-HANDBOOK.md §2.1 and §14.11.

| Task | Item | ENG-HANDBOOK.md Section | Content |
|------|------|------------------------|---------|
| 4.0 | Pre-phase lint baseline | — | Verify lint-docs clean before changes |
| 4.1 | .claude/ Directory Structure Reference | §14.11 or §2.1 | Full `.claude/` tree (agents/, skills/, rules/, settings.json, etc.) |
| 4.2 | User-Level ~/.claude/ Structure | §14.11 | `~/.claude/` layout; loading order vs project CLAUDE.md |
| 4.3 | CLAUDE.md Format and Loading Behavior | §14.11 | User message delivery, /compact survival, @import syntax |
| 4.4 | Required CLAUDE.md Sections for cryptoutil | §14.11 | Canonical section structure for this project's CLAUDE.md |
| 4.5 | Complete Skill Frontmatter Fields | §2.1.5 | `allowed-tools`, `model`, `effort`, `context`, `agent`, `paths`, `shell` |
| 4.6 | Dynamic Context Injection Syntax | §2.1.5 | Backtick-bang blocks; `$ARGUMENTS`, `$0/$N`, `${CLAUDE_SESSION_ID}` |
| 4.7 | Skill Body Structure Template | §2.1.5 | Recommended SKILL.md body structure for cryptoutil |
| 4.8 | Sub-Agent Frontmatter Fields | §2.1.1 | `disallowedTools`, `permissionMode`, `maxTurns`, `skills`, `memory`, `color` |
| 4.9 | Path-Scoped Rules (.claude/rules/) | §14.11 | Auto-load behavior, `paths` frontmatter, recommended cryptoutil rule files |
| 4.10 | agentskills.io Open Standard Context | §2.1.5 | Cross-agent shared frontmatter; multi-tool adoption |
| 4.11 | CLAUDE.md Length and Scoping Strategy | §14.11 | Per-directory CLAUDE.md, path-scoped rules for monorepos |
| 4.12 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

---

## Phase 5: Knowledge Propagation Verification

**Phase Objective**: Run `lint-docs` and verify zero propagation drift across all modified
ENG-HANDBOOK.md sections and their corresponding instruction files.

| Task | Description |
|------|-------------|
| 5.1 | Run `go run ./cmd/cicd-lint lint-docs` — verify ALL checks pass |
| 5.2 | Fix any `validate-chunks`, `validate-coverage`, or `lint-agent-drift` violations found |

---

## ENG-HANDBOOK.md Section Map

| Phase | Suggestion Doc | ENG-HANDBOOK.md Sections Modified |
|-------|---------------|----------------------------------|
| Phase 1 | target-structure-suggestions.md | §2.1, §4.4, §4.4.1, §4.4.4, §4.4.6, §B.7, §11.1.4, §12.2.1 |
| Phase 2 | tls-structure-suggestions.md | §6.5, §6.11 + tls-structure.md fix |
| Phase 3 | deployment-templates-suggestions.md | §6.11.4, §12.2.1, §12.3.1, §12.3.3, §12.3.5, §13.2, §13.6 |
| Phase 4 | claude-structure-suggestions.md | §2.1.1, §2.1.5, §14.11 |
| Phase 5 | — | Verification only |

---

## Out of Scope

- No Go code changes
- No new fitness linters (that is V17 scope)
- No changes to `docs/pki-init-order.md` (Items 6–7 from tls-structure-suggestions.md are obsolete)
- No changes to deployment artifacts (`deployments/`, `configs/`) — only documentation
