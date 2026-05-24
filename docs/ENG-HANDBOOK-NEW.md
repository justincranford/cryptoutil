---
title: cryptoutil Architecture - Proposed Structured Handbook
version: 0.1
date: 2026-05-23
status: Phase 1 Working Draft
purpose: New handbook structure that separates semantic sections from downstream appendix mirrors.
---

# cryptoutil Architecture - Proposed Structured Handbook

## 1. Scope

This document is a first-phase redesign of [docs/ENG-HANDBOOK.md](docs/ENG-HANDBOOK.md).
It does not replace the current handbook yet.
Its purpose is to establish a cleaner structure for both humans and agents and to define a new referential-integrity model for downstream Copilot and Claude artifacts.

### 1.1 Phase Goal

The target model is:

1. Semantic sections live before the appendixes.
2. Those sections contribute programmatic chunks to appendix blocks.
3. Appendix blocks are the only downstream-facing mirror surfaces.
4. Appendix blocks then propagate to Copilot and Claude files.

The intended chain is:

`Sections -> Appendixes -> Copilot/Claude files`

### 1.2 Phase 1 Boundary

This phase focuses on structure, flow, and reverse-engineering.
It does not yet attempt a full semantic rewrite of every propagated downstream artifact.
The appendixes in this draft are organized as the future downstream control plane and document how the current flat propagation system can be lifted into a two-layer model.

## 2. Problems In The Current Handbook

The current handbook is strong as a single source of truth, but it has three maintenance costs:

1. Propagation blocks are distributed across the full document, so downstream relationships are locally correct but globally hard to inspect.
2. The semantic narrative and the downstream synchronization surfaces are interleaved, which makes both human reading and agent ingestion noisier than necessary.
3. Multi-target propagation is currently expressed directly from section-local blocks to many instruction and agent files, which obscures composition when several handbook concepts combine into one downstream artifact.

## 3. Design Objectives

The new structure optimizes for four things.

1. Human readability: core architecture content should read top-down without repeated downstream file details.
2. Agent readability: rules, tables, and canonical contracts should be grouped into compact sections with predictable anchors.
3. Referential integrity: every downstream artifact should have one obvious appendix home.
4. Future linter enforcement: validation should be able to enforce both layers of propagation, not only the final downstream layer.

## 4. Proposed Top-Level Structure

The recommended future handbook shape is:

1. Executive Summary
2. Audience And Consumption Modes
3. Core Platform Model
4. Service And Product Architecture
5. Security And Cryptography
6. Data And Storage
7. API And Contract Surfaces
8. Observability And Operations
9. Tooling, CI, And Deployment
10. Testing And Quality
11. Development Workflow And Local Operations
12. Customization Surfaces
13. Propagation Architecture
14. Migration Plan
15. Appendixes

This ordering front-loads architecture and policy while moving downstream file mirrors to the end.

## 5. Reverse-Engineered Semantic Buckets

A compact handbook should prefer stable semantic buckets over hundreds of locally placed propagation snippets.
The current handbook content can be re-grouped into these upstream buckets.

### 5.1 Terminology And Normative Language

Current chunk families:

1. `rfc-2119-keywords`
2. `emphasis-keywords`
3. `abbreviations`

Primary downstream consumers:

1. Copilot terminology instruction
2. CLAUDE loader surface

### 5.2 Execution And Agent Behavior

Current chunk families:

1. `quality-attributes`
2. `end-of-turn-commit-protocol`
3. `mandatory-review-passes`
4. `agent-self-containment`
5. `per-task-status-updates`
6. `lessons-md-structure`

Primary downstream consumers:

1. Beast mode instruction
2. Evidence-based instruction
3. Agent-format instruction
4. Agent pairs

### 5.3 Architecture And Service Template

Current chunk families:

1. `service-framework-components`
2. `tls-provision-mode`
3. `three-tier-database-strategy`
4. `sqlite-barrier-outside-tx`

Primary downstream consumers:

1. Architecture instruction
2. Data-infrastructure instruction

### 5.4 Security, Authn, And Authz

Current chunk families:

1. `key-principles`
2. `session-token-formats`
3. `headless-authn`
4. `browser-authn`
5. `mfa-combinations`
6. `authz-methods`
7. `secrets-detection-strategy`
8. `tls-client-policy`

Primary downstream consumers:

1. Security instruction
2. Authentication instruction

### 5.5 API, OpenAPI, And Status Contracts

Current chunk families:

1. `base-initialisms`
2. `http-status-codes`

Primary downstream consumers:

1. OpenAPI instruction
2. OpenAPI-oriented skills

### 5.6 Observability And Deployment Tooling

Current chunk families:

1. `otel-collector-constraints`
2. `docker-compose-rules`
3. `cicd-command-naming`
4. `cicd-lint-constraints`
5. `cicd-bulk-hook-architecture`
6. `docker-desktop-startup`
7. `docker-desktop-upgrade`
8. `infrastructure-blocker-escalation`

Primary downstream consumers:

1. Observability instruction
2. Deployment instruction
3. Cross-platform instruction
4. Linting instruction
5. Workflow-fix agent

### 5.7 Testing, Quality, And Go Development

Current chunk families:

1. `utf8-without-bom`
2. `test-file-suffixes`
3. `production-closure-body-coverage`
4. `sequential-test-exemption`
5. `disable-keep-alives-test-transport`
6. `timeout-double-multiplication-antipattern`
7. `postgres-mtls-client-identity`
8. `mutation-common-survivors`
9. `format-go-protection`
10. `validator-error-aggregation`
11. `crypto-acronyms-caps`
12. `conventional-commits`
13. `incremental-commits`
14. `restore-from-baseline`
15. `scripting-language-policy`
16. `platform-line-ending-operations`

Primary downstream consumers:

1. Testing instruction
2. Coding instruction
3. Golang instruction
4. Git instruction
5. Cross-platform instruction
6. Skills and agents that embed handbook-derived rules

## 6. Proposed Propagation Architecture

### 6.1 New Layering Contract

The future linter should enforce two different relations.

1. `section-source -> appendix-block`
2. `appendix-block -> downstream-file`

The first layer expresses semantic composition.
The second layer expresses distribution.

### 6.2 Why This Is Better

This model creates three benefits.

1. Upstream semantic editing becomes local to the narrative sections.
2. Downstream audits become local to the appendixes.
3. The linter can prove both semantic completeness and downstream completeness separately.

### 6.3 Suggested Marker Taxonomy

The current markers are sufficient for one layer but not expressive enough for the new chain.
The future validator can evolve with a second marker family.

Suggested conceptual model:

1. `@contribute` from semantic sections into an appendix block id.
2. `@appendix-propagate` from an appendix block to downstream targets.

Illustrative shape:

```html
<!-- @contribute to-appendix="copilot-instruction-01-01" as="terminology-rfc-2119" -->
... semantic content ...
<!-- @/contribute -->

<!-- @appendix-propagate to=".github/instructions/01-01.terminology.instructions.md" as="copilot-instruction-01-01" -->
... assembled appendix block ...
<!-- @/appendix-propagate -->
```

Phase 1 does not implement this validator change yet.
It defines the target shape so the linter can be upgraded intentionally instead of implicitly.

## 7. Reverse-Engineered Downstream Appendix Plan

### 7.1 Appendix Families

This handbook should end with exactly three downstream-facing appendix families plus the CLAUDE loader subsection.

1. Copilot instruction appendix
2. Copilot and Claude agent-pair appendix
3. Copilot and Claude skill-pair appendix
4. CLAUDE loader subsection within the instruction appendix family

### 7.2 Why CLAUDE.md Lives With Instructions

`CLAUDE.md` is not just another documentation file.
It is the loader that enumerates and imports the Copilot instruction set for Claude-side use.
For maintenance purposes it belongs beside the instruction appendix, not beside agents or skills.

## 8. Appendix A - Copilot Instruction Targets

Each subsection below is the future home of one downstream instruction-file mirror.
In the final model, each appendix block is assembled from one or more semantic sections above, then propagated to the target file.

### A.1 01-01.terminology.instructions.md

Target: `.github/instructions/01-01.terminology.instructions.md`

Expected source contributions:

1. Terminology and normative language
2. Abbreviation policy

Current chunk ids:

1. `rfc-2119-keywords`
2. `emphasis-keywords`
3. `abbreviations`

### A.2 01-02.beast-mode.instructions.md

Target: `.github/instructions/01-02.beast-mode.instructions.md`

Expected source contributions:

1. Execution and agent behavior
2. Quality strategy

Current chunk ids:

1. `quality-attributes`
2. `end-of-turn-commit-protocol`

### A.3 02-01.architecture.instructions.md

Target: `.github/instructions/02-01.architecture.instructions.md`

Current chunk ids:

1. `service-framework-components`
2. `tls-provision-mode`

### A.4 02-02.versions.instructions.md

Target: `.github/instructions/02-02.versions.instructions.md`

Current chunk ids:

1. `minimum-versions`

### A.5 02-03.observability.instructions.md

Target: `.github/instructions/02-03.observability.instructions.md`

Current chunk ids:

1. `otel-collector-constraints`

### A.6 02-04.openapi.instructions.md

Target: `.github/instructions/02-04.openapi.instructions.md`

Current chunk ids:

1. `base-initialisms`
2. `http-status-codes`

### A.7 02-05.security.instructions.md

Target: `.github/instructions/02-05.security.instructions.md`

Current chunk ids:

1. `secrets-detection-strategy`
2. `tls-client-policy`

### A.8 02-06.authn.instructions.md

Target: `.github/instructions/02-06.authn.instructions.md`

Current chunk ids:

1. `key-principles`
2. `session-token-formats`
3. `headless-authn`
4. `browser-authn`
5. `mfa-combinations`
6. `authz-methods`

### A.9 03-01.coding.instructions.md

Target: `.github/instructions/03-01.coding.instructions.md`

Current chunk ids:

1. `validator-error-aggregation`
2. `format-go-protection`

### A.10 03-02.testing.instructions.md

Target: `.github/instructions/03-02.testing.instructions.md`

Current chunk ids:

1. `test-file-suffixes`
2. `three-tier-database-strategy`
3. `production-closure-body-coverage`
4. `sequential-test-exemption`
5. `disable-keep-alives-test-transport`
6. `timeout-double-multiplication-antipattern`
7. `postgres-mtls-client-identity`
8. `mutation-common-survivors`

### A.11 03-03.golang.instructions.md

Target: `.github/instructions/03-03.golang.instructions.md`

Current chunk ids:

1. `crypto-acronyms-caps`

### A.12 03-04.data-infrastructure.instructions.md

Target: `.github/instructions/03-04.data-infrastructure.instructions.md`

Current chunk ids:

1. `three-tier-database-strategy`
2. `sqlite-barrier-outside-tx`

### A.13 03-05.linting.instructions.md

Target: `.github/instructions/03-05.linting.instructions.md`

Current chunk ids:

1. `utf8-without-bom`
2. `cicd-bulk-hook-architecture`

### A.14 04-01.deployment.instructions.md

Target: `.github/instructions/04-01.deployment.instructions.md`

Current chunk ids:

1. `docker-compose-rules`
2. `cicd-command-naming`
3. `cicd-lint-constraints`
4. `docker-compose-verification-in-scope`

### A.15 05-01.cross-platform.instructions.md

Target: `.github/instructions/05-01.cross-platform.instructions.md`

Current chunk ids:

1. `scripting-language-policy`
2. `docker-desktop-startup`
3. `docker-desktop-upgrade`

### A.16 05-02.git.instructions.md

Target: `.github/instructions/05-02.git.instructions.md`

Current chunk ids:

1. `conventional-commits`
2. `incremental-commits`
3. `restore-from-baseline`
4. `platform-line-ending-operations`

### A.17 06-01.evidence-based.instructions.md

Target: `.github/instructions/06-01.evidence-based.instructions.md`

Current chunk ids:

1. `mandatory-review-passes`
2. `infrastructure-blocker-escalation`
3. `per-task-status-updates`

### A.18 06-02.agent-format.instructions.md

Target: `.github/instructions/06-02.agent-format.instructions.md`

Current chunk ids:

1. `agent-self-containment`

### A.19 06-03.tool-efficiency.instructions.md

Target: `.github/instructions/06-03.tool-efficiency.instructions.md`

Current state:

1. No direct propagated chunk ids today
2. Primarily local instruction content
3. Good candidate for future semantic extraction from tooling sections

### A.20 CLAUDE.md Loader Surface

Target: `CLAUDE.md`

This subsection should eventually mirror only the handbook-derived parts of `CLAUDE.md` that enumerate and explain instruction loading, agent loading, and skill loading.
It should not necessarily absorb all Claude-specific local glue.

Expected source contributions:

1. Instruction inventory
2. Agent inventory
3. Skill inventory
4. Handbook positioning and architecture references

## 9. Appendix B - Copilot And Claude Agent Pairs

Each agent-pair subsection represents one shared body surface plus the pair-specific frontmatter differences that stay outside the propagated body.

### B.1 beast-mode

Targets:

1. `.github/agents/beast-mode.agent.md`
2. `.claude/agents/beast-mode.md`

Current shared chunk ids already embedded from the handbook:

1. `mandatory-review-passes`
2. `platform-line-ending-operations`
3. `cicd-bulk-hook-architecture`

Future appendix composition should also absorb the common autonomous-execution contract as a single appendix-owned body.

### B.2 fix-workflows

Targets:

1. `.github/agents/fix-workflows.agent.md`
2. `.claude/agents/fix-workflows.md`

Current handbook-derived shared chunk ids:

1. `mandatory-review-passes`
2. `platform-line-ending-operations`

### B.3 implementation-execution

Targets:

1. `.github/agents/implementation-execution.agent.md`
2. `.claude/agents/implementation-execution.md`

Current handbook-derived shared chunk ids:

1. `mandatory-review-passes`
2. `platform-line-ending-operations`
3. `lessons-md-structure`
4. `per-task-status-updates`
5. `docker-compose-verification-in-scope`

### B.4 implementation-planning

Targets:

1. `.github/agents/implementation-planning.agent.md`
2. `.claude/agents/implementation-planning.md`

Current handbook-derived shared chunk ids:

1. `mandatory-review-passes`
2. `platform-line-ending-operations`
3. `lessons-md-structure`
4. `per-task-status-updates`
5. `docker-compose-verification-in-scope`

## 10. Appendix C - Copilot And Claude Skill Pairs

Skill pairs already have identical bodies today.
That makes them the cleanest downstream family for the future appendix model.

### C.1 Skills Inventory

Pairs:

1. `copilot-customization`
2. `coverage-analysis`
3. `fips-audit`
4. `fitness-function-gen`
5. `migration-create`
6. `new-service`
7. `openapi-codegen`
8. `propagation-check`
9. `psid-template-sync`
10. `sync-copilot-claude`
11. `test-benchmark-gen`
12. `test-fuzz-gen`
13. `test-table-driven`

### C.2 Pair Structure Rule

For each skill pair, the future appendix should own one shared body block and should treat frontmatter differences as pair-local metadata outside the propagated block.

That gives a stable pattern:

1. semantic sections contribute handbook-derived rules
2. appendix skill block assembles the canonical shared body
3. Copilot and Claude skill files consume that exact block

### C.3 High-Value Early Candidates

The best initial appendix-backed skill migrations are:

1. `propagation-check`
2. `sync-copilot-claude`
3. `copilot-customization`
4. `test-table-driven`
5. `openapi-codegen`

These are the most handbook-coupled skill bodies and therefore benefit most from an explicit section-to-appendix model.

## 11. Migration Checklist

### 11.1 Phase 1 Exit Criteria

Phase 1 is complete when:

1. the new structure is clear
2. every current downstream target has one appendix home
3. every current chunk id is assigned to at least one semantic bucket
4. the old flat propagation model has been reverse-engineered into appendix ownership

### 11.2 Phase 2 Entry Criteria

Phase 2 should begin once the linter strategy is chosen for enforcing:

1. section-to-appendix completeness
2. appendix-to-downstream completeness
3. appendix ownership uniqueness
4. pair-body identity for agents and skills

## 12. Open Questions

1. Should the future linter assemble appendix blocks physically, or validate declarative composition only.
2. Should `CLAUDE.md` be fully appendix-backed or only partially handbook-backed.
3. Should instruction appendixes mirror full files or only handbook-derived bodies plus local glue.
4. Should agent and skill frontmatter stay hand-maintained or become generated metadata.

## 13. Recommendation

Do not rewrite the downstream files yet.
First upgrade the propagation model conceptually and structurally, then teach the linter about the two-layer contract, then begin moving the most handbook-coupled downstream artifacts into appendix-owned blocks.

That sequence keeps the first migration stable and reviewable.
