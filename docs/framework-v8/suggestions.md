# ENG-HANDBOOK Suggestions

**Created**: 2026-04-08
**Source**: Deep analysis of docs/ENG-HANDBOOK.md

---

## Inconsistencies

1. **Rate limiting units conflict: req/sec vs req/min** (§1.2 vs §8.5.2) —
   §1.2 says "100 req/sec" for browser APIs; §8.5.2 says "100 req/min" for public APIs.
   This is a 100x difference. One is wrong.

2. **Port ranges in §5.3.1 are stale** (§5.3.1 vs §3.4) —
   §5.3.1 shows "8080-8089 KMS, 8100-8149 Identity" but §3.4 canonical port table shows
   sm-kms at 8000-8099 and identity at 8400-8899. §5.3.1 values are from an older layout.

3. **E2E app instance count conflicts** (§2.2 vs §10.1 vs §13.1.5) —
   §2.2 says "2x PostgreSQL + 2x SQLite" (4 total), §10.1 says "3 app instances
   (2 PostgreSQL + 1 SQLite)", §13.1.5 defines 4 app-variant configs. Reconcile to one number.

4. **Docker Compose minimum version: "v2+" vs "v2.24+"** (§B.1 vs §12.3.5) —
   §B.1 minimum versions table says "v2+" but §12.3.5 requires "v2.24+" for `include:`
   deduplication and `!override` tag. The propagated version is dangerously lax.

## Duplicate / Misnumbered Sections

1. **Duplicate section number 12.3.5** —
   Two sections share "12.3.5": "Recursive Include Architecture" (~line 4504) and "Canonical
   Docker Compose Service Command Pattern" (~line 5156). The second is misnumbered and misplaced.

## Stale Cross-References

1. **§6.10 references "Section 12.6" but links to §13.3** —
   The display text says "Section 12.6" but the anchor points to §13.3. Update display text.

2. **§13.2 references "Section 12.4.5" which was renumbered to §13.1.5** —
   The anchor `#1245-config-file-naming-strategy` is dead. Update to point to §13.1.5.

3. **Cross-reference index: §13.4.7 anchor is wrong** —
   `#1347-propagation-coverage-accounting` points to "Migration Strategy" not
   "Propagation Coverage Accounting."

## Outdated Content

1. **Fitness sub-linter count: "18+" should be ~68** (§1.2) —
   §1.2 says "18+ fitness sub-linters" but §9.11.1 catalogs 68. Update the executive summary.

2. **Stale "v6 CREATE" framework version reference** (§12.2.1) —
    "0 product-level Dockerfiles exist (v6 CREATE)" should reference current framework version
    or use a status description instead of a version number.

## Redundancies

1. **Secret value format table duplicated** (§4.4.6 vs §12.3.3) —
    Both locations have nearly identical postgres-url.secret format tables. Declare one canonical
    and cross-reference from the other.

2. **Port offset strategy documented twice** (§3.4.1 vs §12.3.4) —
    SERVICE/PRODUCT/SUITE offset rules (+0/+10000/+20000) appear in both. Consolidate.

## Missing Cross-References

1. **§12.3.5 Recursive Include Architecture has no inbound references** —
    No cross-references from §9.10 (CICD Command Architecture) or §12.2 (Dockerfile patterns)
    despite being a major architectural decision.

2. **§14.11 Claude Code doesn't cross-reference §2.1 or §B.5** —
    §14.11 describes execution modes but doesn't link to §2.1 (Agent Orchestration Strategy)
    or §B.5 (Agent Catalog).

3. **§13.1.5 Config File Naming not referenced from §5.2** —
    Developers start at §5.2 (Service Builder) when adding services but there's no link to
    §13.1.5 for the required config file naming convention.

## Missing Content

1. **OTel Collector version not specified** (§9.4) —
    No minimum version for `otel-collector-contrib` Docker image despite being critical infra.

2. **gremlins version missing from §B.1 minimum versions table** —
    gremlins v0.6.0+ mentioned in §10.5 but absent from the propagated versions table.

3. **Alpine base image version not tracked** (§B.1) —
    Dockerfile examples use `alpine:3.19` but this isn't in the minimum versions table.

4. **No minimum versions for Fiber, GORM, oapi-codegen, testcontainers-go** —
    Listed as technology choices in §A.2/§B.2 but no minimum versions specified. All are Go
    module dependencies that could introduce breaking API changes.

5. **§12.4 Environment Strategy and §12.5 Release Management are stubs** —
    Each has ~3 lines with no architecture decisions or actionable patterns. Every surrounding
    section has 20-100+ lines of detail.

6. **§15 Operational Excellence subsections are minimal stubs** (§15.1-15.5) —
    Monitoring, Incident Management, Performance, Capacity Planning, Disaster Recovery each
    have ~4 bullet points. Referenced from instruction files but contain no substantive content.

## Version Drift

1. **Docker Compose version propagation pushes wrong minimum** —
    Same as #4: the `@propagate` chunk in §B.1 pushes "v2+" to instruction files while the
    actual requirement is v2.24+. The propagation system broadcasts the wrong minimum.

## Structural

1. **§13 heading structure anomaly** —
    Three sections (12.3.5-duplicate, 12.4, 12.5) are positioned inside §13's line range but
    carry §12 numbering. Appears to be a restructuring artifact from when §12 was split.

2. **§13.4.7 propagation tracking table may be stale** —
    The table lists 38 chunks but doesn't track completeness or last validation date. May have
    drifted from `required-propagations.yaml`.

---

## golangci-lint Configuration (.golangci.yml)

1. **Stale `importas` aliases reference old package paths** —
    ~15 alias entries point to `cryptoutil/internal/cmd/cicd_lint/*` (old structure).
    Current structure is `cryptoutil/internal/apps/tools/cicd_lint/*`. These dead aliases
    are harmless but add maintenance noise. Remove or update them.

2. **`testpackage.skip-regexp: '.*_test\.go$'` disables testpackage entirely** —
    The skip regex matches ALL test files, making the `testpackage` linter a no-op.
    Either configure it properly or remove from the enabled list.

3. **`goheader` disabled due to corruption bug in golangci-lint v2** —
    Comment says "monitor v2.8+ for fix." After upgrading past v2.7.2, re-evaluate
    enabling `goheader` for copyright header enforcement.

4. **Growing `nilerr` exclusion list for lint_deployments validators** —
    8 validators each get a dedicated nilerr exclusion. Consider refactoring the
    validation result pattern to avoid returning nil error alongside captured errors,
    or add a single directory-level exclusion for `lint_deployments/`.

---

## cicd-lint Linters/Formatters

1. **Fitness sub-linter count: actual 70 vs ENG-HANDBOOK "18+"** —
    The "18+" claim in §1.2 (lines 182, 214) is severely outdated. Update to "70+"
    or use a dynamic reference ("see §9.11.1 for current catalog").

2. **lint-fitness-registry.yaml may be stale** —
    If the registry YAML tracks sub-linter names, verify it matches the current 70
    directories under `internal/apps/tools/cicd_lint/lint_fitness/`.

---

## Copilot/Claude Instructions

1. **Instructions are comprehensive** — All 18 instruction files cover the documented
    domains with appropriate cross-references. No missing topic areas identified.

2. **03-05.linting.instructions.md mentions "wsl -> wsl_v5 config key"** — This is
    accurate for golangci-lint v2. No update needed.

3. **Potential gap: no instruction file for documentation standards** — The propagation
    system is documented in copilot-instructions.md and 06-02.agent-format.instructions.md,
    but there's no dedicated instruction for documentation writing standards (Markdown style,
    section structuring, cross-reference conventions). Low priority.

---

## Copilot/Claude Agents

1. **4 agents with dual canonical pairs** — All synchronized, no drift detected.
    Agent catalog in CLAUDE.md matches `.claude/agents/` directory contents.

2. **No agent for documentation-focused work** — beast-mode and implementation-*
    agents target code changes. A documentation-focused agent could enforce
    propagation rules and cross-reference consistency. Low priority.

---

## Copilot/Claude Skills

1. **15 skills with dual canonical pairs** — All synchronized, no drift detected.
    Skill catalog in CLAUDE.md matches `.claude/skills/` directory contents.

2. **No skill for deployment validation** — cicd-lint lint-deployments has 8
    validators, but there's no `/deployment-validate` skill to guide users through
    deployment changes. The existing `/new-service` skill covers creation but not
    modification of existing deployments. Low priority.
