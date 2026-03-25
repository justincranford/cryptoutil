# Implementation Plan - Framework Brainstorm Execution

**Status**: Phase 1 In Progress
**Created**: 2026-03-07
**Based On**: docs/framework-brainstorm/00-overview.md through 08-recommendations.md

## Companion Documents

1. **plan.md** (this file) — phases, objectives, decisions
2. **tasks.md** — task checklist per phase
3. **lessons.md** — persistent memory: what worked, what did not, root causes, patterns

> **Note**: Framework v1 implementation is tracked in `docs/framework-v1/plan.md` and
> `docs/framework-v1/tasks.md`. This brainstorm plan covers (1) closing Framework v1,
> (2) copilot artifact synchronization, and (3) any future brainstorm items deferred
> from Framework v1.

---

## Executive Summary

The brainstorm research (docs/framework-brainstorm/00-overview.md through 08-recommendations.md)
identified improvement priorities for cryptoutil's internal framework. Framework v1
(`docs/framework-v1/`) implemented the highest-ROI items. This plan does three things:

1. **Closes Framework v1** — completes unclosed quality gates and knowledge propagation
2. **Synchronizes Copilot Artifacts** — ensures every agent, skill, instruction, and
   ARCHITECTURE.md section reflects the same artifact scope (code, tests, config,
   deployments, workflows, documents)
3. **Documents Deferred Work** — captures unimplemented brainstorm items for future plans

---

## Brainstorm Research Summary

### Implemented in Framework v1

| Item | Status | Where |
|------|--------|-------|
| P0-1: ServiceContract interface | ✅ Done | `server/contract.go` + compile-time assertions |
| P0-2: air live reload | ✅ Done | `.air.toml` per-service targets |
| P2-2: Architecture fitness functions (`lint-fitness`) | ✅ Done | `cmd/cicd-lint/lint_fitness/` |
| P1-4: Shared test infrastructure | ✅ Done | `template/service/testing/` packages |
| P1-2: Cross-service contract test suite | ✅ Done | `template/service/testing/contract/` |
| Simplified builder pattern | ✅ Done | Template builder Phase 2/3 |

### Excluded (User Decision, Framework v1)

| Item | Decision | Rationale |
|------|----------|-----------|
| P0-3: Skeleton full CRUD reference | Excluded | Contract + fitness = stronger enforcement |
| P1-1: cicd new-service scaffolding | Excluded | 9 services exist; unlikely to add more |
| P1-3: cicd diff-skeleton conformance | Excluded | Superseded by fitness functions |
| P2-1: Service manifest declaration | Excluded | Replaced by simplified builder defaults |
| P2-3: OpenAPI-to-Repository codegen | Excluded | Not wanted |
| P3-1: Module system (fx/Wire) | Excluded | Overkill |
| P3-2: Extract framework module | Excluded | Premature — all 10 services internal |

---

## Phases

### Phase 1: Framework v1 Closure (~2 days) [Status: 🔄 In Progress]

**Objective**: Complete all unclosed Framework v1 work — stale status fields, Phase 7
quality gates, Phase 8 knowledge propagation.

**Sub-phases**:
1. Update stale ❌ task status fields in framework-v1/tasks.md (packages exist, just not marked)
2. Run Phase 7 quality gates: `go test ./...`, coverage, lint-fitness  
3. Complete Phase 8: ARCHITECTURE.md updates (lint-fitness, shared test infra)
4. Verify Phase 8: skills (contract-test-gen ✅, fitness-function-gen ✅), agents (updated this session ✅)
5. Final framework-v1 commit and mark plan complete

**Success**: All 48/48 framework-v1 tasks show ✅. Docs and code consistent.

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 2: Copilot Skills Completeness Audit (~1 day) [Status: ☐ TODO]

**Objective**: Audit each of the 14 skills to ensure they cover ALL artifact types:
code, tests, config, deployments, workflows, and documentation — not just code.

**Skills under audit** (14 total):
- test-table-driven, test-fuzz-gen, test-benchmark-gen, coverage-analysis
- migration-create, fips-audit, propagation-check
- openapi-codegen, agent-scaffold, instruction-scaffold, skill-scaffold
- new-service, contract-test-gen, fitness-function-gen

**For each skill, verify**:
- Does the skill guidance apply only to code, or does it also address docs/config/workflows?
- Are there mentions of when the skill applies to deployments, CI/CD, or documentation?
- Are cross-artifact implications explicit (e.g., "adding a migration also means updating docs/CONFIG-SCHEMA.md")?

**Success**: Each skill either has explicit cross-artifact scope OR is correctly scoped as code-only with documented rationale.

**Post-Mortem**: Record observations in lessons.md; identify skills needing follow-up tasks.

---

### Phase 3: Copilot Agents Completeness Audit (~1 day) [Status: ☐ TODO]

**Objective**: Audit each of the 5 agents to ensure their self-evaluation scope
covers ALL artifact types: code, tests, config, deployments, workflows, documentation,
agents, skills, and instructions.

**Agents under audit** (5 total):
- beast-mode, doc-sync, fix-workflows
- implementation-planning, implementation-execution

**For each agent, verify**:
1. Does post-mortem self-evaluation explicitly mention code, tests, config, deployments, workflows, documentation?
2. Does the commit strategy reflect semantic grouping per artifact type?
3. Does the quality gate checklist cover all artifact types (not just code)?
4. Are there cross-artifact consistency checks (e.g., if code adds a pattern, does ARCHITECTURE.md need updating)?

**Note**: implementation-execution and implementation-planning were updated this session to include code/tests/workflows/docs in post-mortem scope.

**Success**: All 5 agents have explicit cross-artifact self-evaluation scope with no gaps.

**Post-Mortem**: Record gaps found in lessons.md.

---

### Phase 4: Copilot Instructions Completeness Audit (~1 day) [Status: ☐ TODO]

**Objective**: Audit each instruction file to ensure rules apply consistently
across ALL artifact types where relevant.

**Instructions under audit** (18 files in .github/instructions/):
- Architecture, security, authn, observability, openapi, versions
- Coding, testing, golang, data-infrastructure, linting
- Deployment, cross-platform, git
- Evidence-based, beast-mode, agent-format, terminology

**For each instruction, verify**:
1. Do the rules apply to code, tests, config, AND documentation consistently?
2. Do deployment instructions cover config file patterns, not just Docker/k8s?
3. Do testing instructions cover test fixtures, test data, test helpers, not just test code?
4. Are there implicit assumptions that only apply to code (and not documented constraints for docs/config)?

**Success**: Each instruction file has consistent cross-artifact guidance with no implicit code-only assumptions.

**Post-Mortem**: Record gaps in lessons.md; create follow-up tasks for contradictions.

---

### Phase 5: ARCHITECTURE.md Cross-Artifact Completeness Review (~1 day) [Status: ☐ TODO]

**Objective**: Ensure ARCHITECTURE.md documents patterns for ALL artifact types,
not just code. Every major pattern must have documentation explaining how it applies
to code, tests, config, deployments, workflows, and documentation.

**Sections under review**:
- Section 10 (Testing): Does it cover test helpers, fixtures, shared infrastructure?
- Section 11 (Quality): Does it cover docs quality, config quality, deployment quality?
- Section 12 (Deployment): Does it cover config schema validation, secret management?
- Section 13 (Development): Does it cover periodic commits for docs/config changes?
- Section 5 (Service Template): Does it fully document builder pattern, fitness functions?

**Specific gaps identified so far**:
1. lint-fitness command NOT documented in ARCHITECTURE.md (Phase 8.2 from framework-v1)
2. Shared test infrastructure packages NOT documented (testdb, testserver, fixtures, assertions, healthclient)
3. Air live reload NOT documented in ARCHITECTURE.md

**Success**: Architecture patterns consistently document all artifact types. lint-docs passes.

**Post-Mortem**: Record what was added and why in lessons.md.

---

### Phase 6: Knowledge Propagation [Status: ☐ TODO]

**Objective**: Apply insights from Phases 1-5 to permanent project artifacts.

**Propagation targets**:
1. ARCHITECTURE.md sections updated (Phases 1+5)
2. @source/@propagate blocks in instruction files updated to match ARCHITECTURE.md
3. Agent files updated where Phase 2-3 audit found gaps
4. Skill files updated where Phase 2 audit found gaps
5. lint-docs validate-propagation passes

**Success**: `go run ./cmd/cicd-lint lint-docs validate-propagation` passes. All artifacts consistent.

---

### Phase 7: Future Brainstorm Items (Deferred) [Status: ☐ TODO]

**Objective**: Document deferred brainstorm items as first-class tasks for future plans,
with enough context that a future agent can execute without re-reading all brainstorm docs.

**Deferred items** (from framework-v1 explicit exclusions, re-evaluated):
1. **cicd new-service scaffolding** (P1-1): Low priority unless 10th+ service added.
   When revisited: use `text/template` + skeleton-template as source, generate into
   `internal/apps/PRODUCT/SERVICE/`.
2. **OpenAPI-to-Repository codegen** (P2-3): User said "not wanted" in framework-v1.
   Reconsider only if CRUD boilerplate becomes painful.
3. **Extract framework module** (P3-2): Revisit when >15 services OR when external
   consumers of the framework are identified.

**Note**: P0-3 (Skeleton CRUD), P1-3 (diff-skeleton), P2-1 (manifest), P3-1 (module system)
remain excluded by user decision. These are NOT re-evaluated in this plan.

**Success**: Each deferred item has a standalone task in tasks.md with enough context
for future execution.

---

## Success Criteria

- [ ] Framework v1 marked 100% complete (48/48 tasks)
- [ ] All 14 skills audited for cross-artifact completeness
- [ ] All 5 agents audited for cross-artifact post-mortem scope
- [ ] All 18 instruction files audited for cross-artifact guidance
- [ ] ARCHITECTURE.md documents: lint-fitness, shared test infra, air live reload
- [ ] lint-docs validate-propagation passes
- [ ] All deferred brainstorm items documented in tasks.md for future plans
- [ ] Quality gates: build clean, lint clean, tests pass

---

## Decisions

### Decision 1: Scope of "Cross-Artifact Completeness"

**Question**: Does every skill/agent/instruction need to cover EVERY artifact type?

**Decision**: No — skills are task-specific and should be correctly scoped. A skill that
is code-only (e.g., `fips-audit`) is correct as-is. The audit checks for IMPLICIT code-only
assumptions that should be explicit, not that every skill must cover all artifact types.

**Rationale**: Force-fitting all artifact types into every skill would reduce clarity.
The goal is: "if a pattern applies to multiple artifact types, it should be documented.
If it only applies to code, that should be explicitly noted."

### Decision 2: Framework-brainstorm vs framework-v2

**Question**: Should this be called "framework-v2" instead of framework-brainstorm execution?

**Decision**: Keep in framework-brainstorm/ until a concrete v2 emerges. This plan adds
artifact sync + documentation closure. If concrete new framework features are needed
(new skills, new fitness functions), create docs/framework-v2/ at that time.
