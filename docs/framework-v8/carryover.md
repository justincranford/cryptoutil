# Framework-v8 Carryover Items

**Created**: 2026-04-05
**Source**: Analysis of `docs/framework-v7/` (plan.md, tasks.md, lessons.md) + `docs/target-structure.md` Section N

Numbered list, prioritized from highest to lowest impact. Items marked ✅ are complete.

---

## 1. Move `internal/shared/apperr/` → `internal/apps/framework/apperr/` [HIGH]

**Current state**: `internal/shared/apperr/` contains `app_errors.go`, `http_errors.go`, and
`http_status_line_and_code.go` — application-level HTTP error abstraction used by all services.

**Why HIGH**: This is an architectural correctness issue. `internal/shared/` is for truly cross-cutting
infrastructure utilities (crypto, telemetry, magic constants). Application error types that
map to HTTP status codes are a **service framework concern** — they belong in
`internal/apps/framework/apperr/` alongside `ServerConfig`, `DatabaseConfig`, etc. Keeping them
in `internal/shared/` violates the layering rule established during framework-v7.

**Action**: Move the package, update all import paths, run `go build ./...` + `golangci-lint run`.

---

## 2. Create Product-Level Dockerfiles (5 Products) [HIGH]

**Current state**: Product-level deployment directories (`deployments/sm/`, `deployments/jose/`,
`deployments/pki/`, `deployments/identity/`, `deployments/skeleton/`) are missing Dockerfiles.

**Why HIGH**: The `dockerfile_labels` fitness linter and `deployment_dir_completeness` validator
already enforce per-service Dockerfiles. Product-level Dockerfiles are needed to support
product-tier Docker Compose deployments (compose.yml at the product level launches all services
within a product via a single product-built image). Without these, the product-tier deployments
are structurally incomplete.

**Template**: Use the existing `skeleton-template/Dockerfile` as the base. Each product Dockerfile
must include the standard OCI labels (`org.opencontainers.image.title`, etc.) matching the product
display name.

**Action**: Create `deployments/{sm,jose,pki,identity,skeleton}/Dockerfile` (5 files).

---

## 2.1. Migrate `.claude/commands/` → `.claude/skills/` + Update Linter [HIGH]

**Current state**: Claude Code's preferred format for custom slash commands is now a
directory-based skill at `.claude/skills/<name>/SKILL.md` (following the
[Agent Skills open standard](https://agentskills.io/)). The 14 existing command files at
`.claude/commands/*.md` are the legacy format — still supported but superseded.

**The `lint-skill-command-drift` linter** currently checks `.claude/commands/<name>.md` against
Copilot skills. After migration, it must check `.claude/skills/<name>/SKILL.md` instead.

**Why HIGH**: The dual canonical strategy (Copilot Skills ↔ Claude Skills) is a foundational
architectural principle. As long as the linter checks the legacy path, drift between Copilot
skills and Claude skills is undetected. New skills created using the correct format (`sync-copilot-claude`
attempted to do this) are not yet validated by the linter.

**Action** (3 steps):
1. For each of the 14 files in `.claude/commands/`: create `.claude/skills/<name>/SKILL.md`
   directory and file (copy + adapt frontmatter; body stays identical to Copilot skill body),
   test via `/<name>` in Claude Code, then delete the command file
2. Update `lint-skill-command-drift` Go implementation in `internal/apps/tools/cicd_lint/lint_docs/`
   to check `.claude/skills/<name>/SKILL.md` instead of `.claude/commands/<name>.md`
3. Verify `go run ./cmd/cicd-lint lint-docs` still passes with zero errors after migration

**Skills to migrate**: agent-scaffold, contract-test-gen, coverage-analysis, fips-audit,
fitness-function-gen, instruction-scaffold, migration-create, new-service, openapi-codegen,
propagation-check, skill-scaffold, test-benchmark-gen, test-fuzz-gen, test-table-driven (14 total).

---

## 3. Fitness Linter: `usage/service_browser_health_paths` [MEDIUM]

**Current state**: Each PS-ID has a `{ps-id}_usage.go` file with CLI usage strings. These usage
strings MUST mention both `/service/**` and `/browser/**` health paths because every PS-ID exposes
both endpoints. This is not currently enforced by any fitness linter.

**Why MEDIUM**: Discovered during framework-v7 Task 5.5 (deleted 5 product-level usage files).
The deletion of `identity/authz/authz_usage.go` etc. revealed that ensuring PS-ID usage files
contain both health path variants is a code quality invariant that should be automatically verified.

**Pattern**: The linter should scan all `*_usage.go` files under `internal/apps/{PS-ID}/` and
verify they contain both the string `/service/api/v1/health` and `/browser/api/v1/health`.
Files that omit either path are flagged as violations.

**Action**: Create `internal/apps/tools/cicd_lint/lint_fitness/usage_health_path_completeness/`
with `lint.go` + `lint_test.go`, register in the fitness registry.

---

## 4. ✅ Create `docs/framework-v8/claude.md` — Claude AI Best Practices [MEDIUM]

**Status**: COMPLETED in framework-v8 session (2026-04-05).

**What was done**: Created `docs/framework-v8/claude.md` covering Claude Code file structure
(`.claude/` directory layout), CLAUDE.md format guidelines, skill YAML frontmatter reference,
agent frontmatter reference, path-scoped rules (`.claude/rules/`), the Agent Skills open standard
(agentskills.io), corrected dual canonical strategy (Skills → Claude Skills, not Commands),
and migration checklist from legacy commands to skills.

---

## 5. ✅ Create Copilot Skill: `sync-copilot-claude` [MEDIUM]

**Status**: COMPLETED in framework-v8 session (2026-04-05).

**What was done**: Created `.github/skills/sync-copilot-claude/SKILL.md` (Copilot) and
`.claude/skills/sync-copilot-claude/SKILL.md` (Claude — using the new preferred directory format).
The skill covers audit, pair creation, migration workflow, and legacy status checks.

---

## 6. `const-redefine` Linter: Verify Blocking in CI/CD [MEDIUM]

**Current state**: The `magic-usage` sub-linter within `lint-go` enforces two categories:
`literal-use` (BLOCKING) and `const-redefine` (was informational, corrected to BLOCKING in
Task 3.4). This correction may not be propagated to all CI/CD pipeline stages.

**Why MEDIUM**: `const-redefine` detects values that are re-declared as local constants outside
the magic package — always a violation. If CI/CD still treats this as informational, the fix
will not prevent regressions.

**Action**: Verify `go run ./cmd/cicd-lint lint-go` reports `const-redefine` violations with
exit code 1. Run the lint-go fitness test suite to confirm blocking behavior. If not blocking,
trace through `magic_usage.go` and update the severity classification.

---

## 7. Load Test Refactoring: All Tiers [LOW]

**Current state**: `test/load/` (Gatling, Java 21, Maven) covers only some service-level
scenarios. Per `docs/target-structure.md` Section I, the target is:
- 10 service-level load test scenarios (one per PS-ID)
- 5 product-level load test scenarios (one per product)
- 1 suite-level load test scenario

**Why LOW**: Load tests do not block CI/CD and require Java/Gatling expertise to extend.
However, the gap means production throughput characteristics at product and suite levels are
unknown until these are created.

**Action**: Extend `test/load/src/` to add product-level and suite-level simulation classes.
Ensure `pom.xml` is updated with the new simulation entry points.

---

## 8. Debug Log Cleanup in Barrier Service [LOW]

**Current state**: `internal/apps/framework/service/server/barrier/intermediate_keys_service.go`
contains ~10 `log.Printf("DEBUG ...")` statements from development debugging.

**Why LOW**: Debug logging at INFO level pollutes production logs. These should be converted to
proper structured logging (zap) at DEBUG level or removed if they are no longer needed.

**Action**: Replace `log.Printf("DEBUG ...")` calls with `logger.Debug(...)` using zap structured
logging, or remove if the debug context is no longer needed after initial implementation.
