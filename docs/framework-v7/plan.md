# Implementation Plan — Parameterization Opportunities

**Status**: Planning
**Created**: 2026-03-29
**Last Updated**: 2026-03-29
**Purpose**: Implement the 18 ranked parameterization opportunities from
[PARAMETERIZATION-OPPORTUNITIES.md](PARAMETERIZATION-OPPORTUNITIES.md) to replace prose-described
and manually-maintained invariants with machine-enforced validation.

## Quality Mandate — MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:

- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers — NO exceptions:**

- ✅ **Fix issues immediately** — STOP and address
- ✅ **Treat as BLOCKING** — ALL issues block progress to next phase or task
- ✅ **Document root causes** — part of planning AND implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues
- ✅ **NEVER de-prioritize quality** — evidence-based verification is ALWAYS highest priority

## Overview

This plan implements 18 parameterization opportunities that transform the cryptoutil repository
from prose-described conventions to machine-enforced invariants. Each opportunity introduces a
machine-readable schema or derivation function, paired with a fitness linter that validates the
actual codebase matches the declared model.

**Scope**: 5 products, 10 PS-IDs, 1 suite, 57+ fitness linters, 50 config overlays, 140 secret
files, 10 Dockerfiles, 10 compose files, 18 instruction files, 25 PKI profiles.

## Background

The current entity registry (`registry.go`) defines products, product-services, and suites as Go
structs. This approach works but has critical limitations:

- Only Go code can consume the registry at compile time.
- Derived values (ports, SQL identifiers, OTLP names) are re-computed independently in consumers.
- No single source of truth connects the registry to deployment artifacts.

Prior work completed: scripting language policy documentation, cicd-lint constraints codification,
PARAMETERIZATION-OPPORTUNITIES.md reorganization into Part A (18 items) / Part B (2 deferred).

## Technical Context

- **Language**: Go 1.26.1
- **Registry**: `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`
- **Magic constants**: `internal/shared/magic/magic_*.go` (46 files)
- **Fitness linters**: `internal/apps/tools/cicd_lint/lint_fitness/` (57 sub-linters)
- **OpenAPI specs**: `api/{PS-ID}/openapi_spec*.yaml` (8 of 10 services have specs)
- **Config overlays**: `deployments/{PS-ID}/config/{PS-ID}-app-{variant}.yml` (50 files)
- **Secret files**: `deployments/{PS-ID}/secrets/*.secret` (140 files across 3 tiers)
- **Dependencies**: gopkg.in/yaml.v3 (YAML parsing), oapi-codegen (OpenAPI)

## Executive Decision (Pending — See quizme-v1.md Q1)

### Decision 1: Registry YAML Location

**Options**:

- A: `api/cryptosuite-registry/registry.yml` + `generate.go` invoking `./cmd/tools/cicd-registry`
  → `api/cryptosuite-registry/registry.go` (generated). Standard Go project layout per
  [golang-standards/project-layout/api](https://github.com/golang-standards/project-layout/tree/master/api).
- B: `api/cryptosuite-registry/registry.yml` read by `cicd-lint lint-fitness` to build
  parameterized model in-memory and diff against actual code, emitting errors for all deviations.
  No code generation — pure validation.
- C: `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.yaml` alongside existing
  `registry.go`. Consumed at compile time. Registry stays internal.
- D: Repository root `registry.yaml`. Simplest location. Consumed by all consumers.
- E:

**Decision**: TBD — awaiting user answer in quizme-v1.md.

### Decision 2: Code Generation vs Pure Validation

**Options**:

- A: **Generate-and-validate** — YAML → `go generate` → `registry.go`; fitness linters also read
  YAML at runtime. Advantage: generated Go code eliminates manual struct maintenance. Risk:
  adds a build step dependency.
- B: **Validate-only** — YAML is the schema; fitness linters load it at runtime and build the
  parameterized model in-memory; existing `registry.go` is maintained manually but linters
  verify it matches YAML. Advantage: no new `cmd/` binary, no code generation. Risk: YAML
  and Go registry can silently drift if lint is not run.
- C: **Hybrid** — YAML is source of truth; `registry.go` is generated; fitness linters import
  the generated Go code (not YAML). Advantage: strong compile-time guarantees. Risk: more
  complex build pipeline.
- D: **YAML replaces Go entirely** — no `registry.go` at all; all consumers parse YAML at
  startup. Advantage: single source. Risk: YAML parsing at every fitness linter invocation;
  breaks existing import patterns.
- E:

**Decision**: TBD — awaiting user answer in quizme-v1.md.

## Phases

### Phase 1: Foundation — Entity Registry YAML (16h) [Status: ☐ TODO]

**Objective**: Create the canonical `registry.yaml` schema and integrate it with the existing
Go registry.

**Items**: #01 Machine-Readable Entity Registry Schema.

**Work**:

- Define YAML schema covering suite, products (with configurable display names), and all 10
  product-services (with base_port, pg_host_port, migration_range, api_resources using actual
  OpenAPI paths).
- Create JSON Schema for validation.
- Implement registry YAML loader in Go (gopkg.in/yaml.v3).
- Based on Decision 1 & 2: either generate `registry.go` or validate YAML↔Go consistency.
- Add `entity-registry-schema` fitness linter.
- Tests: ≥95% coverage on loader/validator; mutation testing ≥95%.

**Success**: `registry.yaml` exists with all 1 suite + 5 products + 10 PS-IDs; fitness linter
validates schema on every commit; existing fitness linters continue working.

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

### Phase 2: Standalone Linters — No Registry Dependency (20h) [Status: ☐ TODO]

**Objective**: Implement the 5 standalone opportunities that have no dependency on #01.

**Items**: #03 @propagate Coverage, #15 Fitness Sub-Linter Registry, #18 Test File Suffix Rules, #19 Import Alias Formula, #20 X.509 Certificate Profile Schema.

**Work**:

- #03: Create `required-propagations.yaml` manifest; extend `lint-docs` with coverage check.
- #15: Create `lint-fitness-registry.yaml`; add registry-completeness check in `lint_fitness.go`.
- #18: Create `test-file-suffix-rules.yaml`; new `test-file-suffix-structure` fitness linter.
- #19: Create alias map YAML; new `import-alias-formula` fitness linter.
- #20: Create `pki-ca-profile-schema.json`; validate 25 profile files.

**Success**: All 5 linters pass with zero violations; all have ≥95% test coverage.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: Derivation Functions — Registry Consumers (16h) [Status: ☐ TODO]

**Objective**: Implement derivation functions that consume registry data to compute derived values.

**Items**: #04 Port Formula, #11 SQL Identifier Derivation, #10 OTLP/Compose Naming Formula.

**Work**:

- #04: Add `PublicPort()`, `AdminPort()`, `PostgresPort()` functions; enhance `ValidatePorts`
  fitness linter with formula-level verification.
- #11: Add `PSIDToSQLID()`, `DatabaseName()`, `DatabaseUser()`, `PostgresServiceName()` functions;
  update `compose-db-naming` and `secret-content` linters.
- #10: Add `OTLPServiceName()`, `ComposeServiceName()`, `ValidOTLPServiceNames()` functions;
  enhance `otlp-service-name-pattern` and `compose-service-names` linters with entity registry
  cross-reference.

**Success**: All derivation functions have ≥98% test coverage; enhanced linters cross-reference
registry entities and reject typos.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 4: Secret & Config Schema Validation (16h) [Status: ☐ TODO]

**Objective**: Define machine-readable schemas for secrets and config overlays;
validate all existing files against them.

**Items**: #12 Secret File Content Schema, #05 Parameterized Secret Validation, #06 Config Overlay Template Validation.

**Work**:

- #12: Create `secret-schemas.yaml` with 14 format entries; rewrite `secret-content` linter to
  read schema instead of hardcoded regex.
- #05: Integrate `secret-schemas.yaml` with entity registry for prefix derivation; validate all
  420 secret instances.
- #06: Create config overlay templates; add `config-overlay-freshness` fitness linter to detect
  drift between templates and actual overlay files.

**Success**: All 420 secrets validated against schemas; 50 config overlays validated against
templates; zero false positives.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 5: Deployment & Build Validation (12h) [Status: ☐ TODO]

**Objective**: Validate Dockerfiles, migration ranges, and compose naming against registry.

**Items**: #07 Migration Ranges, #08 Dockerfile Labels, #16 Compose Instance Naming.

**Work**:

- #07: Add migration_range_start/end per PS-ID; enhance `migration-range-compliance` with
  cross-service collision detection.
- #08: Enhance `dockerfile-labels` with registry-derived expected values; new
  `dockerfile-entrypoint-formula` check.
- #16: Change `compose-service-names` from regex-match to set-membership check using
  `ValidComposeServiceNames()`.

**Success**: All migration ranges are non-overlapping; Dockerfile labels match registry;
compose names validated by set membership.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 6: API & Health Completeness (10h) [Status: ☐ TODO]

**Objective**: Validate API paths and health endpoints against registry declarations.

**Items**: #09 CLI Subcommand Completeness, #13 API Path Parameter Registry, #17 Health Path Completeness.

**Work**:

- #09: Define subcommand requirements per role; new `subcommand-completeness` fitness linter
  traversing Go AST.
- #13: Validate OpenAPI spec paths against `api_resources` declared in registry using actual
  paths from existing specs (e.g., `/elastickey`, `/elastic-jwks`, `/messages/tx`).
- #17: New `health-path-completeness` linter checking all 6 health paths per service.

**Success**: All 50 required subcommand handlers tracked; API paths match registry declarations;
60 health path appearances verified.

**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 7: Knowledge Propagation (4h) [Status: ☐ TODO]

**Objective**: Apply lessons learned to permanent artifacts — NEVER skip this phase.

- Review lessons.md from all prior phases.
- Update ARCHITECTURE.md with new patterns and decisions.
- Update agents, skills, instructions, code, tests, workflows, and docs where warranted.
- Verify propagation integrity (`go run ./cmd/cicd-lint lint-docs validate-propagation`).

**Success**: All artifact updates committed; propagation check passes.

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| YAML schema design proves insufficient | Medium | High | Start with minimal schema, iterate |
| Existing fitness linters break during migration | Medium | Medium | Incremental changes, run full suite after each |
| Registry location causes import cycle | Low | High | `api/` directory avoids internal/ import restrictions |
| Config overlay drift too large to auto-detect | Low | Medium | Start with structural validation, not byte-exact |
| OpenAPI spec inconsistencies across services | High | Medium | Document known deviations, fix incrementally |

## Quality Gates — MANDATORY

**Per-Action Quality Gates**:

- ✅ All tests pass (`go test ./...`) — 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) — zero errors
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) — zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets** (from copilot instructions):

- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code: ≥98% line coverage
- ✅ Generated code: Excluded from coverage

**Mutation Testing Targets**:

- ✅ Production code: ≥95%
- ✅ Infrastructure/utility code: ≥98%

**Per-Phase Quality Gates**:

- ✅ Unit + integration tests complete before moving to next phase
- ✅ Deployment validators pass (`go run ./cmd/cicd-lint lint-deployments` — when relevant)
- ✅ Race detector clean (`go test -race -count=2 ./...`)

## Success Criteria

- [ ] All 18 opportunities implemented with fitness linters
- [ ] All quality gates passing
- [ ] Registry YAML covers suite + 5 products + 10 PS-IDs with actual OpenAPI paths
- [ ] Documentation updated (ARCHITECTURE.md, instructions)
- [ ] CI/CD workflows green
- [ ] Evidence archived (test output, logs, analysis)

## ARCHITECTURE.md Cross-References — MANDATORY

| Topic | ARCHITECTURE.md Section | When to Reference |
|-------|------------------------|-------------------|
| Testing Strategy | [Section 10](../../docs/ARCHITECTURE.md#10-testing-architecture) | ALL phases |
| Unit Testing | [Section 10.2](../../docs/ARCHITECTURE.md#102-unit-testing-strategy) | ALL phases |
| Quality Gates | [Section 11.2](../../docs/ARCHITECTURE.md#112-quality-gates) | ALL phases |
| Code Quality | [Section 11.3](../../docs/ARCHITECTURE.md#113-code-quality-standards) | ALL phases |
| Coding Standards | [Section 14.1](../../docs/ARCHITECTURE.md#141-coding-standards) | ALL phases |
| Version Control | [Section 14.2](../../docs/ARCHITECTURE.md#142-version-control) | ALL phases |
| Entity Registry | [Section 9.11.2](../../docs/ARCHITECTURE.md#9112-entity-registry) | Phase 1 |
| Fitness Linters | [Section 9.11.1](../../docs/ARCHITECTURE.md#9111-fitness-sub-linter-catalog) | ALL phases |
| Deployment Architecture | [Section 12](../../docs/ARCHITECTURE.md#12-deployment-architecture) | Phases 4, 5 |
| Secrets Management | [Section 13.3](../../docs/ARCHITECTURE.md#133-secrets-management-in-deployments) | Phase 4 |
| API Architecture | [Section 8](../../docs/ARCHITECTURE.md#8-api-architecture) | Phase 6 |
| Plan Lifecycle | [Section 14.6](../../docs/ARCHITECTURE.md#146-plan-lifecycle-management) | ALL phases |
| Post-Mortem | [Section 14.8](../../docs/ARCHITECTURE.md#148-phase-post-mortem--knowledge-propagation) | ALL phases |
