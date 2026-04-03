# Lessons — Parameterization Opportunities

**Created**: 2026-03-29
**Last Updated**: 2026-03-31
**Status**: COMPLETE — All 7 phases finished. Reorganized into actionable suggestions.

---

## Suggestions for Permanent Artifacts

The following numbered suggestions distill all lessons from Phases 1–7 into actionable items
organized by target artifact. Each item describes a specific change or addition.

---

### A. docs/ARCHITECTURE.md

1. **Document the `api/cryptosuite-registry/` directory** in Section 4.4.1 (Go Project Structure) as the canonical YAML-based entity registry location, alongside the Go registry at `internal/apps/tools/cicd_lint/lint_fitness/registry/`.

2. **Update the fitness linter count** from 59 to 68 in Section 9.11.1 (Fitness Sub-Linter Catalog). Add the 10 new linters: `api_path_registry`, `compose_port_formula`, `config_overlay_freshness`, `entity_registry_schema`, `fitness_registry_completeness`, `health_path_completeness`, `import_alias_formula`, `pki_ca_profile_schema`, `subcommand_completeness`, `test_file_suffix_structure`.

3. **Add Section 9.11.3 (Fitness Linter Registry YAML)** documenting that `lint-fitness-registry.yaml` is the machine-readable catalog of all fitness linters, validated by the `fitness_registry_completeness` linter.

4. **Document the `go:embed` + seam pattern** in Section 10.2.4 (Test Seam Injection Pattern): `var embeddedYAML []byte` with `//go:embed schema.yaml` as a seam overridable in non-parallel tests for error path coverage.

5. **Add PKI-CA profile validation exceptions** to Section 6.5 (PKI Architecture): `min_days: 0` is valid for short-lived certificates (kubernetes-workload, ssh-user), and `default_curve_or_size: null` is valid for Ed25519 (no curve/size parameter needed).

6. **Document the deployment config YAML naming convention** in Section 12.2 or 13.2: Config overlay files use `postgresql` in filename (e.g., `jose-ja-app-postgresql-1.yml`) but the OTLP service name value inside uses `postgres` not `postgresql` (e.g., `jose-ja-postgres-1`). These are intentionally different conventions.

7. **Document the compose port formula** in Section 3.4.1 (Port Design Principles): `base_port + tier_offset + variant_offset` where variant offsets are `sqlite-1`=+0, `postgres-1`=+1, `postgres-2`=+2.

8. **Document the 5 required health paths** in Section 5.5 (Health Check Patterns): `/service/api/v1/health`, `/browser/api/v1/health`, `/admin/api/v1/livez`, `/admin/api/v1/readyz`, `/admin/api/v1/shutdown`. The current documentation says "6 paths" but actually 5 are required.

9. **Document the ENTRYPOINT registry field** in Section 12.2.1 (Dockerfile Parameterization): For services with non-uniform ENTRYPOINT patterns, the canonical entrypoint array is declared per PS-ID in `registry.yaml` under the `entrypoint` key.

10. **Document coverage ceiling pattern** in Section 10.2.3 (Coverage Targets): `filepath.Abs` errors, `os.Stat` permission errors, and `bufio.Scanner.Err()` I/O errors cannot be injected without OS-level mocks. Accept 93–99% per function as structural ceiling; document in test comments.

---

### B. Copilot Instructions / Skills / Agents

1. **Update `.github/skills/fitness-function-gen/SKILL.md`**: Replace `cryptoutilSharedMagic.CacheFilePermissions` with `cryptoutilSharedMagic.FilePermissionsDefault` for `0o600`, and `cryptoutilSharedMagic.CICDTempDirPermissions` with `cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute` for `0o755`. The `magic-usage` linter recommendation is authoritative.

2. **Update `.github/skills/fitness-function-gen/SKILL.md`**: Add the fitness linter test template pattern — test file should include 3 seam tests (ReadFile error, WalkDir error, Getwd error) + happy path + ~10-15 violation tests + ~5 direct unit tests on validation functions = ~30 total tests for ≥95% coverage.

3. **Update `03-02.testing.instructions.md`**: Add the `wsl_v5` blank-line-before-defer rule for seam tests. When a seam test does `original := varFn` then `defer func() { varFn = original }()`, a blank line is REQUIRED between the assignment and the defer statement.

4. **Update `03-01.coding.instructions.md`**: Add the `nlreturn` rule for `for range` loops — blank line before `continue`, blank line before `if err :=` when preceded by multi-statement blocks.

5. **Update `03-02.testing.instructions.md`**: Document that the `literal-use [blocking]` linter applies to ALL Go string/int/octal literals including test `want` slices, struct initializers, and table-driven test expected values. Magic constants are required everywhere.

---

### C. GitHub Workflows / CI/CD

1. **Add gremlins per-package strategy to CI workflow**: Gremlins times out when run on all packages at once on Windows. CI/CD should run mutation testing per-package or per-linter to avoid timeouts. Document this in `ci-mutation.yml`.

2. **Pre-commit double-commit pattern**: Document in CI/CD notes that PowerShell `WriteAllText` writes system line endings; `end-of-file-fixer` and `mixed-line-ending` pre-commit hooks fix on first commit attempt; second commit succeeds. Always expect two `git commit` invocations when new files are written via PowerShell.

---

### D. Coding Patterns

1. **Always verify magic constant VALUES** before writing YAML registry entries or test expectations. Drift between constant names and values is the most common source of test failures across all phases. Read the magic file, don't rely on memory.

2. **Compose variant name mismatch**: Template YAML `variant:` names must match `ComposeVariant*` constants (`sqlite-1`, `postgres-1`, `postgres-2`) — not the file suffix keywords (`postgresql-1.yml`). The `variantSuffixes` map handles the namespace translation.

3. **Struct-based YAML loader with `gopkg.in/yaml.v3`**: Clean, idiomatic loader with validation at load time prevents runtime issues. Use `init()` panic for missing registry to fail fast at program start.

4. **Derivation functions in a separate `derivations.go` file**: Keeps the registry loader clean and makes derivations independently testable with fixture-based table-driven tests.

5. **Set-membership with PS-ID prefix filter**: When validating compose service names, scope to `strings.HasPrefix(svc, psID+"-")` to automatically exclude infrastructure services without needing an explicit exemption list.

6. **`SkeletonTemplateServiceID` vs `SkeletonTemplateServiceName`**: When filtering the skeleton-template PS-ID during service iteration, use `SkeletonTemplateServiceID = "skeleton-template"` (PS-ID constant), not `SkeletonTemplateServiceName = "template"` (service name constant).

7. **YAML `any` type for mixed null/string/int fields**: Use `any` instead of a specific type for fields like `default_curve_or_size` that can be null or a mix of string/int. This is the correct approach for YAML deserialization with `gopkg.in/yaml.v3`.

---

### E. Testing Patterns

1. **Sequential seam test NON-PARALLEL requirement**: Tests modifying package-level `var` function pointers MUST NOT use `t.Parallel()`. Document with comment `// Sequential: modifies package-level seam (not parallel-safe).`

2. **TestCheck_DelegatesToCheckInDir with `os.Chdir`**: When `Check()` uses `"."` as root and tests run from package dirs, use `os.Chdir(projectRoot)` + defer restore (non-parallel) to test the real entry point.

3. **Pre-commit hook blocks commit for pre-existing issues**: The `cicd-lint-all` pre-commit hook runs `lint-go`, which fails if ANY magic-aliases violation exists anywhere in the codebase. Always run `go run ./cmd/cicd-lint lint-go` before committing new code.

4. **Linter error behavior differs per package**: `compose_port_formula` Test `Check_DelegatesToCheckInDir` returns nil when files are missing (skip behavior), but `compose_db_naming` and `compose_service_names` FAIL on missing files. Each linter's error behavior differs — don't assume uniform skip/fail behavior.

5. **Test file scanning includes `_test.go`**: The `health-path-completeness` linter scans ALL top-level `.go` files in a service directory, including test files. sm-im's service health path is documented in its `http_test.go` — test files count toward compliance when checking the real workspace.

---

### F. Deployment / Config Patterns

1. **Database-url standard for SQLite overlays**: ALL sqlite overlay files MUST contain `database-url: "sqlite://file::memory:?cache=shared"`. PostgreSQL overlays MUST NOT contain `database-url`.

2. **Config overlay template YAML**: Template is embedded in the linter package via `go:embed` (`lint_fitness/config_overlay_freshness/config-overlay-templates.yaml`). Separate documentation copy exists in `lint_deployments/`.

3. **pki-ca dual database key**: pki-ca config overlays use both `database.dsn` (legacy) and `database-url` (standard). The config-overlay-freshness linter correctly requires `database-url`; both keys co-exist.

4. **8 missing `database-url` keys were discovered** by the config-overlay-freshness linter in jose-ja, sm-im, skeleton-template, and pki-ca sqlite overlays. The linter immediately caught real drift — demonstrating the parameterization approach works.

5. **identity-spa Dockerfile title bug caught**: Dockerfile had `cryptoutil-identity-spa-rp` (wrong `-rp` suffix from copy-paste from identity-rp). Registry-driven exact matching caught this; substring matching had silently accepted it for months.

---

### G. Linting / Fitness Patterns

1. **AST-based import alias validation**: The `import_alias_formula` linter uses `go/ast` and `go/parser` for accurate Go source analysis — avoids false positives from regex-based approaches.

2. **`buildValidServiceSet()` computed once**: Build the full valid set once in `CheckInDir` and pass as parameter to `checkServiceNames()` for all 10 PS-IDs — avoids 10 redundant set builds.

3. **Skip logic for missing config dirs**: When a linter iterates all 10 PS-IDs from registry but test temp dirs only contain a subset, return `nil` (skip) when `deployments/{psid}/config/` doesn't exist — enables clean unit test isolation.

4. **Schema-driven linter rewrite**: Replacing hardcoded regex with a loaded YAML schema reduces brittleness and makes new validation types additive (add YAML entry, no Go changes needed).

5. **All services need both `/service/api/v1/health` and `/browser/api/v1/health`**: The `health-path-completeness` linter enforces this. Both paths must appear in usage documentation or handler registration.

---

## Quality Gate Results (Final)

| Gate | Status |
|------|--------|
| Build | ✅ clean |
| golangci-lint | ✅ 0 issues |
| lint-go | ✅ SUCCESS |
| lint-fitness | ✅ SUCCESS (all 68 linters pass) |
| Coverage | ✅ ≥95% across all lint_fitness packages |
| Tests | ✅ all pass |
| Mutation testing | ⏳ Deferred to CI/CD (gremlins panics/times out on Windows) |
| Race detector | ⏳ Deferred to CI/CD (requires CGO_ENABLED=1) |

---

## Phase 1 (Continuation): Parameterization Items #21–#27

*(To be filled during Phase 1 execution)*

---

## Phase 2 (Continuation): TLS Init Refactoring

*(To be filled during Phase 2 execution)*

---

## Phase 3 (Continuation): Framework CLI & Magic Cleanup

*(To be filled during Phase 3 execution)*

---

## Phase 4 (Continuation): Config Test File Reorganization

*(To be filled during Phase 4 execution)*

---

## Phase 5 (Continuation): Identity Product Refactoring

*(To be filled during Phase 5 execution)*

---

## Phase 6 (Continuation): Knowledge Propagation

*(To be filled during Phase 6 execution)*
