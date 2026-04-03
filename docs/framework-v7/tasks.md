# Tasks — Framework v7 (Continuation)

**Status**: 1 of 37 tasks complete (3%)
**Last Updated**: 2026-04-02
**Created**: 2026-04-02

## Quality Mandate — MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:

- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability, NOT implementation speed
- ✅ **Accuracy**: Root cause addressed, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer
- ❌ **Premature Completion**: Objective evidence required before marking complete

**ALL issues are blockers — NO exceptions.**

---

## Task Checklist

### Phase 1: Parameterization Items #21–#27

**Phase Objective**: Implement 7 new parameterization items. Also: migrate #02 and #14 to DONE,
delete both PARAMETERIZATION files.

**Prerequisite (before starting 1.1)**

#### Task 1.0: PARAMETERIZATION File Cleanup

- **Status**: ✅ (Done during planning)
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add #02 and #14 to PARAMETERIZATION-DONE.md as permanently deferred
  (NEVER). Then delete `PARAMETERIZATION-OPPORTUNITIES.md` and `PARAMETERIZATION-DONE.md`.
- **Acceptance Criteria**:
  - [x] #02 entry added to PARAMETERIZATION-DONE.md (status: NEVER)
  - [x] #14 entry added to PARAMETERIZATION-DONE.md (status: NEVER)
  - [x] `PARAMETERIZATION-OPPORTUNITIES.md` deleted
  - [x] `PARAMETERIZATION-DONE.md` deleted
- **Files**:
  - `docs/framework-v7/PARAMETERIZATION-DONE.md` (update then delete)
  - `docs/framework-v7/PARAMETERIZATION-OPPORTUNITIES.md` (delete)

#### Task 1.1: #21 — Claude Command YAML Frontmatter + Drift Validation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Dependencies**: None
- **Description**: Add YAML frontmatter to all 14 `.claude/commands/*.md` files. Extend
  `CheckSkillCommandDrift()` to validate frontmatter presence, `description` match, and
  `argument-hint` match. Claude command `name` field uses bare skill name (NOT `claude-` prefix).
  Update ARCHITECTURE.md §2.1.5 and instruction file §06-02.
- **Acceptance Criteria**:
  - [ ] All 14 `.claude/commands/*.md` files have `---` YAML frontmatter with `name`, `description`
  - [ ] `CheckSkillCommandDrift()` validates frontmatter presence (fails if missing)
  - [ ] `CheckSkillCommandDrift()` validates `description` matches between Copilot skill and Claude command
  - [ ] `lint-docs` exits non-zero on missing or mismatched frontmatter
  - [ ] ARCHITECTURE.md §2.1.5 documents Claude command frontmatter rules
  - [ ] §06-02 instruction rules updated
  - [ ] Tests ≥95% coverage on new validation logic
- **Files**:
  - `.claude/commands/*.md` (all 14 files)
  - `internal/apps/tools/cicd_lint/lint_docs/` (CheckSkillCommandDrift)
  - `docs/ARCHITECTURE.md`
  - `.github/instructions/06-02.agent-format.instructions.md`

#### Task 1.2: #22 — Multi-Language Parameterized Testing Standards

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Dependencies**: None
- **Description**: Expand `test-table-driven` skill and Claude command to cover Go, Java
  (Gatling), and Python (pytest). Add `lint-java-test` and `lint-python-test` cicd-lint
  subcommands. Update ARCHITECTURE.md §10 with §10.9 (Java/Gatling) and §10.10 (Python/pytest).
  Update cicd-lint command table to show 13 linter commands.
- **Acceptance Criteria**:
  - [ ] `test-table-driven` skill updated with Java and Python sections
  - [ ] Claude command updated to match
  - [ ] `lint-java-test` sub-linter implemented and registered
  - [ ] `lint-python-test` sub-linter implemented and registered
  - [ ] ARCHITECTURE.md §10.9 and §10.10 added
  - [ ] cicd-lint command table shows 13 linter commands
  - [ ] Tests ≥95%
- **Files**:
  - `.github/skills/test-table-driven/SKILL.md`
  - `.claude/commands/test-table-driven.md`
  - `internal/apps/tools/cicd_lint/` (new lint-java-test, lint-python-test)
  - `docs/ARCHITECTURE.md`

#### Task 1.3: #23 — Copilot↔Claude Skill Body Content Drift

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Dependencies**: 1.1 (drift checker extended)
- **Description**: Normalize `## Key Rules` heading in all 14 skill/command pairs. Extend
  `CheckSkillCommandDrift()` to validate rule section presence. Every skill MUST have
  `## Key Rules`; every Claude command MUST mirror it.
- **Acceptance Criteria**:
  - [ ] All 14 skills have `## Key Rules` section
  - [ ] All 14 Claude commands mirror the `## Key Rules` section from skill
  - [ ] `CheckSkillCommandDrift()` errors if section is missing
  - [ ] `lint-docs` rejects mismatches
  - [ ] Tests ≥95%
- **Files**:
  - `.github/skills/*/SKILL.md` (all 14)
  - `.claude/commands/*.md` (all 14)
  - `internal/apps/tools/cicd_lint/lint_docs/` (CheckSkillCommandDrift)

#### Task 1.4: #24 — Claude Code Continuous Execution Config

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: None
- **Description**: Add ARCHITECTURE.md §14.9 documenting Claude Code autonomous execution
  options (beast-mode agent invocation, settings.local.json, CLI flags). Update CLAUDE.md.
  Create/update `.claude/settings.local.json` with appropriate defaults.
- **Acceptance Criteria**:
  - [ ] ARCHITECTURE.md §14.9 added documenting all three execution options
  - [ ] CLAUDE.md updated with reference to §14.9
  - [ ] `.claude/settings.local.json` exists with reasonable defaults
  - [ ] No `lint-docs` failures from new section
- **Files**:
  - `docs/ARCHITECTURE.md`
  - `CLAUDE.md`
  - `.claude/settings.local.json`

#### Task 1.5: #25 — Agent Self-Containment Linter

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Dependencies**: None
- **Description**: New `lint_agent_self_containment/` sub-linter in lint-docs. Scans
  `.github/agents/*.agent.md` bodies; errors if no `ARCHITECTURE.md` reference found.
- **Acceptance Criteria**:
  - [ ] `lint_agent_self_containment/lint_agent_self_containment.go` implemented
  - [ ] Registered in `lint_docs.go`
  - [ ] Fails for agents with zero ARCHITECTURE.md references
  - [ ] Passes for all current compliant agents
  - [ ] Tests ≥95%
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_docs/lint_agent_self_containment/`
  - `internal/apps/tools/cicd_lint/lint_docs/lint_docs.go`

#### Task 1.6: #26 — ARCHITECTURE.md Section Link Validity

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Dependencies**: None
- **Description**: New `lint_architecture_links/` sub-linter in lint-docs. Extracts H1–H4
  headings from ARCHITECTURE.md; validates all `](../../docs/ARCHITECTURE.md#ANCHOR)` references
  in instruction/agent/skill files resolve to real headings.
- **Acceptance Criteria**:
  - [ ] `lint_architecture_links/lint_architecture_links.go` implemented
  - [ ] Extracts all real anchors from ARCHITECTURE.md using heading → anchor conversion
  - [ ] Scans all `.github/instructions/`, `.github/agents/`, `.github/skills/` files
  - [ ] Errors on any `#ANCHOR` that doesn't correspond to a real heading
  - [ ] All existing references are valid (fix any broken ones found during implementation)
  - [ ] Registered in `lint_docs.go`
  - [ ] Tests ≥95%
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_docs/lint_architecture_links/`
  - `internal/apps/tools/cicd_lint/lint_docs/lint_docs.go`

#### Task 1.7: #27 — lint-go-test Expansion (3 New Sub-Linters)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Dependencies**: None
- **Description**: Add three new sub-linters to `lint_gotest/`:
  - `hardcoded_uuid/`: forbid `uuid.MustParse("literal-uuid-string")` in test files
  - `real_http_server/`: forbid `httptest.NewServer(` in test files
  - `test_sleep/`: forbid `time.Sleep(` in test files
  All registered in `lint_gotest.go`. Fix any existing violations found.
- **Acceptance Criteria**:
  - [ ] `hardcoded_uuid/` sub-linter implemented and registered
  - [ ] `real_http_server/` sub-linter implemented and registered
  - [ ] `test_sleep/` sub-linter implemented and registered
  - [ ] All 3 registered in `lint_gotest.go`
  - [ ] ARCHITECTURE.md §9.10 cicd-lint table updated
  - [ ] Existing violations fixed (or exempted with documented reason)
  - [ ] Tests ≥95% for each sub-linter
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_go_test/lint_gotest_hardcoded_uuid/`
  - `internal/apps/tools/cicd_lint/lint_go_test/lint_gotest_real_http_server/`
  - `internal/apps/tools/cicd_lint/lint_go_test/lint_gotest_test_sleep/`
  - `internal/apps/tools/cicd_lint/lint_go_test/lint_gotest.go`
  - `docs/ARCHITECTURE.md`

#### Task 1.8: Phase 1 Quality Gates

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: 1.0, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7
- **Description**: Verify all Phase 1 quality gates. Update lessons.md Phase 1.
- **Acceptance Criteria**:
  - [ ] `go test ./... -shuffle=on` passes 100%
  - [ ] `golangci-lint run` clean
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
  - [ ] PARAMETERIZATION files deleted
  - [ ] lessons.md Phase 1 section updated

---

### Phase 2: TLS Init Refactoring

**Phase Objective**: Refactor tls/init.go: remove legacy manual parsing, add 3-tier Init
functions, add configurable FIPS signing algorithm, move pkiInitName to magic.

#### Task 2.1: Move pkiInitName to Magic Package

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: `pkiInitName = "pki-init"` is a local var in `internal/apps/cryptoutil/cryptoutil.go`.
  Add `PSIDPKIInit = "pki-init"` to an appropriate `internal/shared/magic/` file and replace the
  local variable. Verify `golangci-lint run` passes.
- **Acceptance Criteria**:
  - [ ] Magic constant added to `internal/shared/magic/`
  - [ ] Local variable in `cryptoutil.go` replaced with magic constant
  - [ ] `golangci-lint run` clean
  - [ ] `go build ./...` clean

#### Task 2.2: Remove Backward-Compat from Init()

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: None
- **Description**: Replace manual `strings.HasPrefix` arg parsing in `Init()` with pflag
  parsing identical to `InitForService()`. Remove backward-compat comment.
- **Acceptance Criteria**:
  - [ ] `Init()` uses pflag for all flag parsing
  - [ ] `strings.HasPrefix` manual arg parsing removed
  - [ ] Backward-compat comment removed
  - [ ] Tests pass

#### Task 2.3: Add InitForProduct() and InitForSuite()

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Dependencies**: 2.2
- **Description**: Add `InitForProduct(productID string, args []string, ...)` and
  `InitForSuite(suiteID string, args []string, ...)` following the same pflag pattern as
  `InitForService()`. Add productID/suiteID as SAN DNS entries. Wire into product/suite CLI
  entry points in `cmd/`.
- **Acceptance Criteria**:
  - [ ] `InitForProduct()` implemented
  - [ ] `InitForSuite()` implemented
  - [ ] Both wired in product and suite entry points
  - [ ] Tests for both new functions ≥95%

#### Task 2.4: Add Configurable Signing Algorithm

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Dependencies**: 2.2
- **Description**: Add `--signing-algorithm` pflag to all 4 Init functions. Default:
  `ECDSA-P384-SHA384` (FIPS-approved). Support: P256+SHA256, P384+SHA384, P521+SHA512,
  RSA-2048+SHA256, RSA-3072+SHA256, RSA-4096+SHA256. Add magic constants for each name string.
  Validate flag value; pass to TLS generator.
- **Acceptance Criteria**:
  - [ ] `--signing-algorithm` flag present in all 4 Init functions
  - [ ] Default is `ECDSA-P384-SHA384`
  - [ ] All 6 algorithms accepted; invalid values rejected with clear error
  - [ ] Magic constants defined for each algorithm name
  - [ ] Tests verify all algorithms configurable

#### Task 2.5: Phase 2 Quality Gates

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: 2.1, 2.2, 2.3, 2.4
- **Description**: Verify Phase 2 quality gates. Update lessons.md Phase 2.
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/framework/tls/...` passes 100%
  - [ ] Coverage ≥95%
  - [ ] `golangci-lint run` clean
  - [ ] lessons.md Phase 2 updated

---

### Phase 3: Framework CLI & Magic Cleanup

**Phase Objective**: Move CLI constants to magic, add cicd-lint to pre-commit, remove
function-var redeclarations, consolidate CLI code, fix formatting.

#### Task 3.1: Move constants.go to Magic Package

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: None
- **Description**: Move 8 constants from `service/cli/constants.go` to `internal/shared/magic/`
  (e.g., `magic_cli.go`). Delete `constants.go`. Update all callers. Verify `lint-go literal-use`
  passes after change.
- **Acceptance Criteria**:
  - [ ] All 8 constants in `internal/shared/magic/`
  - [ ] `service/cli/constants.go` deleted
  - [ ] All callers updated
  - [ ] `go run ./cmd/cicd-lint lint-go` passes
  - [ ] `golangci-lint run` clean

#### Task 3.2: Add cicd-lint Hooks to Pre-Commit

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Add `lint-go`, `lint-go-test`, and `lint-fitness` as local pre-commit hooks
  in `.pre-commit-config.yaml`.
- **Acceptance Criteria**:
  - [ ] `lint-go` hook added with `stages: [pre-commit]`
  - [ ] `lint-go-test` hook added with `stages: [pre-commit]`
  - [ ] `lint-fitness` hook added with `stages: [pre-commit]`
  - [ ] UTF-8 without BOM (checked by existing lint-text hook)
  - [ ] `pre-commit run --all-files` passes

#### Task 3.3: Remove Function-Variable Redeclarations in user_auth.go

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Remove `var` block in `user_auth.go` that redeclares package-level functions
  as variables. Call functions directly at all usage sites.
- **Acceptance Criteria**:
  - [ ] `templateClientGenerateUsernameSimpleFn` var removed; call-sites use function directly
  - [ ] `templateClientGeneratePasswordSimpleFn` var removed; call-sites use function directly
  - [ ] `templateClientJSONMarshalFn` var removed; call-sites use `json.Marshal` directly
  - [ ] `go test ./...` passes

#### Task 3.4: New lint-go Sub-Linter for Function-Variable Redeclaration

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Dependencies**: 3.3
- **Description**: Add `lint_go/function_var_redeclaration/` sub-linter. Detects
  `var xxx = pkg.FunctionName` in non-test production `.go` files. Excludes `*_test.go`
  and `export_test.go` (seam injection valid). Register in `lint_go.go`. Tests ≥95%.
- **Acceptance Criteria**:
  - [ ] Sub-linter implemented and registered
  - [ ] Detects the pattern in production code; misses valid seam injection in test files
  - [ ] Passes on current codebase (after 3.3 cleans the only violation)
  - [ ] Tests ≥95%

#### Task 3.5: Consolidate Duplicate CLI Help/Version Code

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: 3.1
- **Description**: After moving constants, verify `cli.IsHelpRequest()` is the sole help
  checker. Remove any inline `args[0] == "help" || args[0] == "--help"` duplications. Same
  for version.
- **Acceptance Criteria**:
  - [ ] No inline help/version string comparisons outside `cli.IsHelpRequest()` / `cli.IsVersionRequest()`
  - [ ] `golangci-lint run` clean

#### Task 3.6: Parameterize health_commands.go

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Dependencies**: None
- **Description**: Extract shared logic from `HealthCommand`, `LivezCommand`, `ReadyzCommand`
  into a private `httpGetCommand(...)` helper. Keep 3 public functions as thin wrappers.
- **Acceptance Criteria**:
  - [ ] Private `httpGetCommand` helper exists
  - [ ] All 3 public functions delegate to it
  - [ ] All 3 command tests still pass
  - [ ] Coverage ≥95%

#### Task 3.7: Fix product_router_test.go Formatting

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Run `golangci-lint run --fix ./internal/apps/framework/product/cli/...` to
  apply gofumpt to `product_router_test.go`. Verify file passes gofumpt check.
- **Acceptance Criteria**:
  - [ ] `product_router_test.go` passes `golangci-lint run`
  - [ ] File uses proper Go tab indentation

#### Task 3.8: Phase 3 Quality Gates

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7
- **Description**: Verify Phase 3 quality gates. Update lessons.md Phase 3.
- **Acceptance Criteria**:
  - [ ] `go test ./...` passes 100%
  - [ ] `golangci-lint run` clean
  - [ ] `go run ./cmd/cicd-lint lint-go` passes
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
  - [ ] lessons.md Phase 3 updated

---

### Phase 4: Config Test File Reorganization

**Phase Objective**: Rename non-semantic test files in `internal/apps/framework/service/config/`
to semantically-named alternatives.

#### Task 4.1: Scan and Plan Renames

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Read `config_coverage_test.go`, `config_gaps_test.go`, and
  `config_test_util_coverage_test.go`. Determine the semantic domain of each test.
  Document the rename mapping before executing it.
- **Acceptance Criteria**:
  - [ ] Each file's test domain documented (e.g., "tests TLS PEM parsing", "tests factory defaults")
  - [ ] Rename mapping agreed before files moved

#### Task 4.2: Execute Renames and Verify

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: 4.1
- **Description**: Execute the rename mapping from 4.1. If test content spans multiple domains,
  split the file. After rename, verify all tests pass.
- **Acceptance Criteria**:
  - [ ] No test file contains `coverage` or `gaps` in its name
  - [ ] `config_test_util_coverage_test.go` renamed to `config_test_util_test.go`
  - [ ] All tests pass after rename
  - [ ] `golangci-lint run` clean

#### Task 4.3: Phase 4 Quality Gates

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: 4.1, 4.2
- **Description**: Verify Phase 4 quality gates. Update lessons.md Phase 4.
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/framework/service/config/...` 100%
  - [ ] Coverage unchanged or improved
  - [ ] lessons.md Phase 4 updated

---

### Phase 5: Identity Product Refactoring

**Phase Objective**: Enforce architectural layering in `internal/apps/identity/`:
framework-generic code to `framework/`, PS-ID-specific code to PS-IDs, identity-domain
code stays.

#### Task 5.1: Migrate Common Config Types to Framework

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Dependencies**: None
- **Description**: `identity/config/config.go` defines `ServerConfig`, `DatabaseConfig`,
  `SessionConfig`, `ObservabilityConfig` — non-identity types. Move to
  `internal/apps/framework/` config package. Update all callers.
- **Acceptance Criteria**:
  - [ ] `ServerConfig`, `DatabaseConfig`, `SessionConfig`, `ObservabilityConfig` in framework
  - [ ] `identity/config/config.go` either deleted or only contains identity-domain types
  - [ ] All callers updated
  - [ ] `go build ./...` clean

#### Task 5.2: Migrate Loader to Framework

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: 5.1
- **Description**: Move `identity/config/loader.go` (`LoadFromFile`, `SaveToFile`) to framework
  config package. Update all callers.
- **Acceptance Criteria**:
  - [ ] `LoadFromFile` and `SaveToFile` live in framework
  - [ ] All callers updated
  - [ ] Tests pass

#### Task 5.3: Migrate Framework-Generic Defaults to Framework

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: 5.1
- **Description**: Move `defaultDatabaseConfig()`, `defaultSessionConfig()`, and
  `defaultObservabilityConfig()` from `identity/config/defaults.go` to framework defaults.
- **Acceptance Criteria**:
  - [ ] Three framework-generic defaults in framework package
  - [ ] `identity/config/defaults.go` only contains identity-domain defaults
  - [ ] `go build ./...` clean

#### Task 5.4: Migrate PS-ID-Specific Defaults to PS-ID Directories

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Dependencies**: 5.3
- **Description**: Move `defaultAuthZConfig()` → `identity-authz/`, `defaultIDPConfig()` →
  `identity-idp/`, `defaultRSConfig()` → `identity-rs/`. Add missing `defaultRPConfig()` →
  `identity-rp/` and `defaultSPAConfig()` → `identity-spa/`. If `identity/config/defaults.go`
  becomes empty, delete it.
- **Acceptance Criteria**:
  - [ ] All 5 PS-ID defaults in their respective PS-ID directories
  - [ ] `identity/config/defaults.go` deleted or only contains product-level identity defaults
  - [ ] `go build ./...` clean

#### Task 5.5: Remove Duplicate PS-ID Usage Constants from Product-Level Subdirs

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: None
- **Description**: `identity/authz/authz_usage.go` duplicates `identity-authz/authz_usage.go`.
  Confirm there is no unique content in product-level `identity/authz/`, then delete the directory.
  Same for `identity/idp/`. Canonical location is the PS-ID directory.
- **Acceptance Criteria**:
  - [ ] `internal/apps/identity/authz/` deleted (after confirming content is duplicate or merged)
  - [ ] `internal/apps/identity/idp/` deleted (same)
  - [ ] PS-ID directories (`identity-authz/`, `identity-idp/`) retain canonical usage constants
  - [ ] `go build ./...` clean

#### Task 5.6: Classify and Verify Remaining Product-Level Files

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: 5.1, 5.2, 5.3, 5.4, 5.5
- **Description**: Walk every file in `internal/apps/identity/` NOT inside a PS-ID subdirectory.
  Classify each by domain. Verify `identity.go`, `apperr/`, `domain/`, `email/`, `issuer/`,
  `jobs/`, `mfa/`, `ratelimit/`, `repository/` are all correctly owned by identity (none need to
  move). Fix any misclassified files.
- **Acceptance Criteria**:
  - [ ] All product-level files explicitly reviewed
  - [ ] Any misclassified files moved to correct location
  - [ ] `go build ./...` clean

#### Task 5.7: Phase 5 Quality Gates

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6
- **Description**: Verify Phase 5 quality gates. Update lessons.md Phase 5.
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/identity/...` 100%
  - [ ] `go test ./internal/apps/identity-authz/...` 100%
  - [ ] Coverage ≥95%
  - [ ] `golangci-lint run` clean
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
  - [ ] lessons.md Phase 5 updated

---

### Phase 6: Knowledge Propagation

**Phase Objective**: Apply lessons learned from all phases to permanent artifacts.

#### Task 6.1: Review lessons.md and Identify Updates

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: All preceding phases
- **Description**: Read all lessons.md phase sections. Identify ARCHITECTURE.md contradictions,
  omissions, outdated patterns. List agent/skill/instruction files needing updates.
- **Acceptance Criteria**:
  - [ ] All lessons.md sections read
  - [ ] Update list created

#### Task 6.2: Update ARCHITECTURE.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: 6.1
- **Description**: Apply identified ARCHITECTURE.md updates for new patterns.
- **Acceptance Criteria**:
  - [ ] All identified sections updated
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 6.3: Update Agents, Skills, Instructions

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: 6.1
- **Description**: Apply lessons to agents, skills, instructions, code, tests, workflows, docs.
  Separate commit per artifact type.
- **Acceptance Criteria**:
  - [ ] All identified artifact files updated
  - [ ] Separate commits per artifact type

#### Task 6.4: Verify Propagation Integrity

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: 6.2, 6.3
- **Description**: Run propagation check. Fix any drift found.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] All commits pushed

---

## Cross-Cutting Tasks

### Testing

- [ ] Unit tests ≥95% coverage (production), ≥98% (new linters)
- [ ] Mutation testing ≥98% for new linters, ≥95% production
- [ ] Race detector clean: `go test -race ./...`
- [ ] No skipped tests

### Code Quality

- [ ] Linting passes: `golangci-lint run ./...`
- [ ] Build-tagged: `golangci-lint run --build-tags e2e,integration ./...`
- [ ] `go run ./cmd/cicd-lint lint-fitness` passes

### Documentation

- [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes
- [ ] ARCHITECTURE.md updated per-phase

---

## Notes / Deferred Work

None at this time. Items #02 and #14 formally deferred to NEVER (see plan.md Background).
