# Implementation Plan — Framework v7 (Continuation)

**Status**: In Progress
**Created**: 2026-03-29
**Last Updated**: 2026-04-02
**Purpose**: Continuation of framework-v7 parameterization work. Phases 1–7 (all 18 original
parameterization items) are fully complete. This plan tracks the next wave of work: 7 new
parameterization items (#21–#27), TLS init refactoring, framework CLI and magic cleanup,
config test reorganization, and identity product refactoring.

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

**ALL issues are blockers — NO exceptions.**

## Background (Phases 1–7 Complete)

All 18 original parameterization opportunities are fully implemented, tested, and committed.

| Phase | Items | Status |
|-------|-------|--------|
| Phase 1 — Entity Registry YAML | #01 Machine-Readable Entity Registry Schema | ✅ Complete |
| Phase 2 — Standalone Linters | #03, #15, #18, #19, #20 | ✅ Complete |
| Phase 3 — Derivation Functions | #04, #11, #10 | ✅ Complete |
| Phase 4 — Secret & Config Schema | #12, #05, #06 | ✅ Complete |
| Phase 5 — Deployment & Build | #07, #08, #16 | ✅ Complete |
| Phase 6 — API & Health | #09, #13, #17 | ✅ Complete |
| Phase 7 — Knowledge Propagation | ARCHITECTURE.md, instructions, skills | ✅ Complete |

**Artifacts delivered**: `api/cryptosuite-registry/registry.yaml`, 68 fitness linters, derivation
functions, secret/config schemas, migration range enforcement, Dockerfile label validation, API
path registry, health path completeness, @propagate coverage manifest.

**Deferred to NEVER**: Items #02 (Generative Deployment Scaffold Command) and #14 (Instruction
File Slot Reservation Table) violate cicd-lint constraints or provide insufficient value.
Recorded in PARAMETERIZATION-DONE.md, then both PARAMETERIZATION files deleted.

## Overview

Six new work areas:

1. **Parameterization #21–#27**: Enforce Claude command frontmatter, multi-language testing
   standards, skill/command drift rules, Claude autonomous config, agent self-containment, ARCH
   link validity, and lint-go-test expansion.
2. **TLS Init Refactoring**: Replace manual arg parsing with pflag, add InitForProduct/InitForSuite,
   add configurable signing algorithm (default ECDSA P-384 + SHA-384), move pkiInitName to magic.
3. **Framework CLI & Magic Cleanup**: Move CLI constants to magic package, add cicd-lint to
   pre-commit, remove function variable redeclarations, consolidate duplicate CLI help code,
   parameterize health_commands.go, fix product_router_test.go formatting.
4. **Framework Config Test Reorganization**: Rename non-semantic test files
   (`_coverage`, `_gaps`) to semantically-named test files.
5. **Identity Product Refactoring**: Move common config types to framework, move PS-ID-specific
   defaults to their services, remove usage constants from product-level authz/idp subdirs.
6. **Knowledge Propagation**: Apply all lessons to permanent artifacts.

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: `internal/apps/framework/` (service, product, tls, builder)
- **Magic constants**: `internal/shared/magic/magic_*.go`
- **Fitness linters**: `internal/apps/tools/cicd_lint/lint_fitness/` (68 sub-linters)
- **Identity product**: `internal/apps/identity/` and 5 PS-IDs (`identity-{authz,idp,rs,rp,spa}`)
- **TLS init**: `internal/apps/framework/tls/init.go`
- **Pre-commit config**: `.pre-commit-config.yaml`

## Phases

### Phase 1: Parameterization Items #21–#27 (24h) [Status: ☐ TODO]

**Objective**: Implement all 7 new parameterization opportunities. These items enforce
Claude/Copilot artifact alignment, multi-language testing standards, and lint-go-test expansion.

**Prerequisite cleanup**: Before starting Phase 1 implementation, add #02 and #14 to
PARAMETERIZATION-DONE.md as permanently deferred, then delete both PARAMETERIZATION files.

**Items**:

- **#21 Claude Command YAML Frontmatter + Drift Validation** (CRITICAL): All 14
  `.claude/commands/*.md` files are missing YAML frontmatter. Extend `CheckSkillCommandDrift()`
  to validate presence, `description` match, and `argument-hint` match. Claude command `name`
  field uses bare skill name (e.g., `test-table-driven`), NOT `claude-` prefix. Update
  ARCHITECTURE.md §2.1.5 and instruction file §06-02.

- **#22 Multi-Language Parameterized Testing** (High): Expand `test-table-driven` skill and
  Claude command to cover Go, Java (Gatling), and Python (pytest). Add `lint-java-test` and
  `lint-python-test` cicd-lint subcommands. Update ARCHITECTURE.md §10 with §10.9 and §10.10.
  Update cicd-lint command table to show 13 linter commands.

- **#23 Copilot↔Claude Skill Body Content Drift** (High): Normalize `## Key Rules` heading
  in all 14 skill/command pairs. Extend `CheckSkillCommandDrift()` to validate rule section
  presence. Every skill MUST have `## Key Rules`; every Claude command MUST mirror it.

- **#24 Claude Code Continuous Execution Config** (Medium): Add ARCHITECTURE.md §14.9 covering
  Claude Code autonomous execution options (beast-mode agent invocation, settings.local.json,
  CLI flags). Update CLAUDE.md. Update `.claude/settings.local.json`.

- **#25 Agent Self-Containment Linter** (Medium): New `lint_agent_self_containment/` sub-linter
  in `lint-docs`. Scans `.github/agents/*.agent.md` bodies; errors if no `ARCHITECTURE.md`
  reference found. Tests ≥95%.

- **#26 ARCHITECTURE.md Section Link Validity** (Medium): New `lint_architecture_links/`
  sub-linter in `lint-docs`. Extracts H1–H4 headings from ARCHITECTURE.md; validates all
  `](../../docs/ARCHITECTURE.md#ANCHOR)` references in instruction/agent/skill files.
  Tests ≥95%.

- **#27 lint-go-test Expansion** (Medium): Three new sub-linters in `lint_gotest/`:
  - `hardcoded_uuid/`: forbid `uuid.MustParse("literal")` in test files
  - `real_http_server/`: forbid `httptest.NewServer(` in test files
  - `test_sleep/`: forbid `time.Sleep(` in test files
  All registered in `lint_gotest.go`. ARCHITECTURE.md §9.10 cicd-lint table updated.

**Success**: `lint-docs` exits non-zero on missing frontmatter; all 7 linters registered;
≥95% coverage on all new code; PARAMETERIZATION files deleted.

**Post-Mortem**: After quality gates pass, update lessons.md Phase 1 section.

---

### Phase 2: TLS Init Refactoring (12h) [Status: ☐ TODO]

**Objective**: Refactor `internal/apps/framework/tls/init.go` — remove legacy manual arg
parsing, add InitForProduct/Suite, add configurable signing algorithm, move pkiInitName to magic.

**Work**:

- **Move `pkiInitName` to magic**: In `internal/apps/cryptoutil/cryptoutil.go` the local
  variable `pkiInitName = "pki-init"` is a magic literal. Add a named constant to
  `internal/shared/magic/` (e.g., `PSIDPKIInit = "pki-init"` in `magic_psids.go`). Replace
  the local variable. Verify `golangci-lint run` passes after change.

- **Remove backward-compat from `Init()`**: The `Init()` function uses manual
  `strings.HasPrefix` arg parsing "for backward compatibility with Docker Compose entrypoint
  scripts." Products have NOT been released — backward compatibility is unnecessary. Replace
  with pflag parsing identical to `InitForService()`. Remove the backward-compat comment.

- **Add `InitForProduct()` and `InitForSuite()`**: `InitForService(serviceID, args, stdout, stderr)`
  exists; add matching:
  - `InitForProduct(productID string, args []string, stdout, stderr io.Writer) int`
  - `InitForSuite(suiteID string, args []string, stdout, stderr io.Writer) int`
  Same pflag approach as `InitForService()`. Add productID/suiteID as additional DNS SAN
  entries. Wire up in product and suite CLI entry points.

- **Configurable signing algorithm**: Add `--signing-algorithm` pflag (default:
  `ECDSA-P384-SHA384`, FIPS-approved). Supported: `ECDSA-P256-SHA256`, `ECDSA-P384-SHA384`,
  `ECDSA-P521-SHA512`, `RSA-2048-SHA256`, `RSA-3072-SHA256`, `RSA-4096-SHA256`. Validate the
  flag value; pass to TLS generator. Add magic constants for each algorithm name string.

**Success**: `tls/init.go` uses pflag uniformly in all public functions; all 3 tiers have Init
functions; signing algorithm is configurable with FIPS defaults; pkiInitName is a magic constant;
≥95% test coverage.

**Post-Mortem**: After quality gates pass, update lessons.md Phase 2 section.

---

### Phase 3: Framework CLI & Magic Cleanup (14h) [Status: ☐ TODO]

**Objective**: Eliminate magic value duplication in `service/cli/constants.go`, remove
function-variable redeclaration anti-pattern, consolidate duplicate help/version code,
parameterize health commands, add cicd-lint to pre-commit, and fix formatting.

**Work**:

- **Move `constants.go` to magic package**: `service/cli/constants.go` declares
  `helpCommand = "help"`, `helpFlag = "--help"`, `helpShortFlag = "-h"`, `urlFlag = "--url"`,
  `cacertFlag = "--cacert"`, `versionCommand = "version"`, `versionFlag = "--version"`,
  `versionShortFlag = "-v"`. These MUST be in `internal/shared/magic/`. Move all 8 to
  `magic_cli.go` (or appropriate file). Delete `constants.go`. Update all callers.
  Verify `lint-go literal-use` passes.

- **Add cicd-lint to pre-commit**: Currently cicd-lint only runs in CI/CD. Add as pre-commit
  stage hooks in `.pre-commit-config.yaml` so commit rejections show lint reasons to Copilot:
  ```yaml
  - id: lint-go
    entry: go
    args: [run, ./cmd/cicd-lint/main.go, lint-go]
    stages: [pre-commit]
  - id: lint-go-test
    entry: go
    args: [run, ./cmd/cicd-lint/main.go, lint-go-test]
    stages: [pre-commit]
  - id: lint-fitness
    entry: go
    args: [run, ./cmd/cicd-lint/main.go, lint-fitness]
    stages: [pre-commit]
  ```

- **Remove function-variable redeclarations in `user_auth.go`**: Remove the `var` block
  declaring `templateClientGenerateUsernameSimpleFn`, `templateClientGeneratePasswordSimpleFn`,
  `templateClientJSONMarshalFn`. Call `cryptoutilSharedUtilRandom.GenerateUsernameSimple`,
  `cryptoutilSharedUtilRandom.GeneratePasswordSimple`, and `json.Marshal` directly at call sites.

- **New lint-go sub-linter for function-variable redeclaration**: Add
  `lint_go/function_var_redeclaration/` sub-linter. Detects `var xxx = pkg.FunctionName`
  patterns in non-test production `.go` files (excludes `*_test.go` and `export_test.go`
  where seam injection is valid). Register in `lint_go.go`. Tests ≥95%.

- **Consolidate duplicate CLI help/version code**: After moving constants to magic, verify
  that `cli.IsHelpRequest()` is the sole canonical help checker. Remove any inline
  `args[0] == "help" || args[0] == "--help"` logic that duplicates it. Same for version.

- **Parameterize `health_commands.go`**: Extract shared logic from `HealthCommand`,
  `LivezCommand`, `ReadyzCommand` into a private `httpGetCommand(...)` helper. Keep the 3
  public functions as thin parameterized wrappers. Tests verify all 3 paths still work.

- **Fix `product_router_test.go` formatting**: File has incorrect indentation. Run
  `golangci-lint run --fix ./internal/apps/framework/product/cli/...` to apply gofumpt.
  Verify file is properly formatted before commit.

**Success**: No magic literals in CLI; cicd-lint in pre-commit; `user_auth.go` calls functions
directly; function-var-redeclaration linter passes; health commands share implementation;
`product_router_test.go` passes gofumpt.

**Post-Mortem**: After quality gates pass, update lessons.md Phase 3 section.

---

### Phase 4: Config Test File Reorganization (6h) [Status: ☐ TODO]

**Objective**: Rename/reorganize test files with non-semantic names in
`internal/apps/framework/service/config/`.

**Affected files (scan contents, then reorganize)**:

- `config_coverage_test.go` — contains tests for real semantic code (TLS PEM bytes, factory
  functions). Rename to match the production code being tested (e.g., `config_tls_test.go`
  or merge into existing `config_parse_test.go`).
- `config_gaps_test.go` — scan, then split into files by tested feature
  (`config_loading_test.go`, `config_validation_test.go`, etc.).
- `config_test_util_coverage_test.go` — tests for `config_test_util.go`. Rename to
  `config_test_util_test.go`.

**Rule**: Test file names MUST reflect the production code or feature being tested,
NOT why they exist (coverage, gaps). After rename, verify all tests pass.

**Post-Mortem**: After quality gates pass, update lessons.md Phase 4 section.

---

### Phase 5: Identity Product Refactoring (20h) [Status: ☐ TODO]

**Objective**: Enforce architectural layering in `internal/apps/identity/`:
framework-generic code moves to `framework/`, PS-ID-specific code moves to PS-IDs,
identity-domain-specific code stays in `identity/`.

**Work**:

**5A — Migrate common config types to framework**:
`identity/config/config.go` defines `ServerConfig`, `DatabaseConfig`, `SessionConfig`,
`ObservabilityConfig` — non-identity types. Move to `internal/apps/framework/` (likely
joining existing framework config types). Identity's `Config` struct then references framework
types. Update all callers.

**5B — Migrate loader.go to framework**:
`identity/config/loader.go` implements `LoadFromFile()` / `SaveToFile()` using `os.ReadFile` +
`yaml.Unmarshal` + `os.ExpandEnv`. This is a general config loading pattern. Move to framework.
Update all callers in identity and any other consumers.

**5C — Migrate common default configs to framework**:
`identity/config/defaults.go` `defaultDatabaseConfig()`, `defaultSessionConfig()`, and
`defaultObservabilityConfig()` belong in framework defaults (not identity). Move them.
`defaultAuthZConfig()`, `defaultIDPConfig()`, `defaultRSConfig()` belong in their PS-ID
service directories. Move them to `identity-authz/`, `identity-idp/`, `identity-rs/`.
Add missing `defaultRPConfig()` and `defaultSPAConfig()` to `identity-rp/` and `identity-spa/`.

**5D — Remove duplicate PS-ID usage constants from product subdirs**:
`identity/authz/authz_usage.go` provides PS-ID-specific CLI usage strings. These are ALREADY
in `identity-authz/authz_usage.go`. Confirm there is no unique content in the product-level
file, then delete `internal/apps/identity/authz/` subdirectory. Same for `identity/idp/`.

**5E — Classify remaining non-service-subdir files**:
Walk every file in `identity/` that is NOT inside a PS-ID service directory. Classify each:
- `identity.go`, `identity_test.go` — product entry point, stays
- `apperr/` — identity-domain errors, stays
- `domain/` — OAuth2/OIDC domain types, stays
- `email/` — email service (identity-shared), stays
- `issuer/` — JWT/JWE issuer (identity-shared), stays
- `jobs/` — background jobs (identity-shared), stays
- `mfa/` — MFA services (identity-shared), stays
- `ratelimit/` — in-memory rate limiter for identity services, stays (comment says
  "for identity services"; it is identity-domain rate limiting, not framework)
- `repository/` — data repositories (identity-shared), stays

**Success**: `identity/config/` has only identity-domain-specific types; common types in
framework; PS-ID-specific defaults in PS-IDs; duplicate usage files removed; all tests pass.

**Post-Mortem**: After quality gates pass, update lessons.md Phase 5 section.

---

### Phase 6: Knowledge Propagation (4h) [Status: ☐ TODO]

**Objective**: Apply lessons learned from Phases 1–5 to permanent artifacts — NEVER skip.

- Review `lessons.md` from all prior phases.
- Update ARCHITECTURE.md with new patterns and decisions discovered in Phases 1–5.
- Update agents, skills, instructions, code, tests, workflows, and docs where warranted.
- Verify propagation integrity (`go run ./cmd/cicd-lint lint-docs validate-propagation`).

**Success**: All artifact updates committed; propagation check passes.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Identity config migration breaks callers | High | High | Scan all callers before moving; update one at a time |
| Magic constant move breaks literal-use lint | Medium | Medium | Run lint-go after each constant move |
| InitForProduct/Suite not wired in CLI entry points | Medium | Medium | Audit all cmd/ entry points |
| Health command refactor breaks fitness linter | Low | Medium | Run lint-fitness after refactor |
| Claude command frontmatter breaks lint-docs | Low | Low | Additive only; existing body unchanged |

## Quality Gates — MANDATORY

- ✅ All tests pass (`go test ./...`) — 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`)
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`)
- ✅ `go run ./cmd/cicd-lint lint-fitness` passes after each phase
- ✅ `go run ./cmd/cicd-lint lint-docs` passes after each phase
- ✅ Coverage: ≥95% production, ≥98% infrastructure/utility linters
- ✅ Mutation: ≥98% new linters, ≥95% production code

## Success Criteria

- [ ] All 7 parameterization items (#21–#27) implemented with linters
- [ ] PARAMETERIZATION-OPPORTUNITIES.md and PARAMETERIZATION-DONE.md deleted
- [ ] `tls/init.go` refactored: pflag, 3 tiers, configurable FIPS signing, pkiInitName magic
- [ ] `constants.go` deleted; CLI constants in magic package
- [ ] cicd-lint hooks in pre-commit
- [ ] Function-var-redeclaration linter implemented
- [ ] Config test files renamed to semantic names
- [ ] Identity product refactoring complete (common → framework, PS-ID-specific → PS-IDs)
- [ ] All quality gates passing

## ARCHITECTURE.md Cross-References — MANDATORY

| Topic | Section | Phases |
|-------|---------|--------|
| Testing Strategy | [§10](../../docs/ARCHITECTURE.md#10-testing-architecture) | 1, 6 |
| Quality Gates | [§11.2](../../docs/ARCHITECTURE.md#112-quality-gates) | ALL |
| Magic Values | [§11.1.4](../../docs/ARCHITECTURE.md#1114-magic-values-organization) | 2, 3 |
| Agent Architecture | [§2.1.1](../../docs/ARCHITECTURE.md#211-agent-architecture) | 1 |
| Skill Catalog | [§2.1](../../docs/ARCHITECTURE.md#21-agent-orchestration-strategy) | 1 |
| CICD Command Arch | [§9.10](../../docs/ARCHITECTURE.md#910-cicd-command-architecture) | 1, 3 |
| Pre-commit Hooks | [§9.9](../../docs/ARCHITECTURE.md#99-pre-commit-hook-architecture) | 3 |
| Service Framework | [§5.1](../../docs/ARCHITECTURE.md#51-service-framework-pattern) | 2, 5 |
| Security / FIPS | [§6.4](../../docs/ARCHITECTURE.md#64-cryptographic-architecture) | 2 |
| Import Aliases | [§11.1.3](../../docs/ARCHITECTURE.md#1113-import-alias-conventions) | 3 |
| Version Control | [§14.2](../../docs/ARCHITECTURE.md#142-version-control) | ALL |
| Post-Mortem | [§14.8](../../docs/ARCHITECTURE.md#148-phase-post-mortem--knowledge-propagation) | 6 |
