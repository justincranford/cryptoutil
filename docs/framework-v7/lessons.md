# Lessons — Parameterization Opportunities

**Created**: 2026-03-29
**Last Updated**: 2026-03-30

## Phase 1: Foundation — Entity Registry YAML

**Completed**: 2026-03-30 (sessions 1-3)

### What Worked

- **Struct-based YAML loader with gopkg.in/yaml.v3**: Clean, idiomatic loader with validation at load time prevents runtime issues.
- **init() panic for missing registry**: Using `init()` to load the registry means any malformed YAML fails fast at program start, not silently at first use.
- **AllProducts/AllProductServices/AllSuites preserved unchanged**: Zero callers disrupted by replacing hardcoded structs with YAML loading.

### What Didn't Work

- **97.6% coverage ceiling**: The `init()` panic path and `os.Getwd()` error path in `findRegistryYAMLPath` are structurally unreachable in unit tests. Documented as ceiling.

### Patterns

- **Registry YAML location**: `api/cryptosuite-registry/registry.yaml` with JSON Schema at `api/cryptosuite-registry/registry-schema.json`.
- **Fitness linter validates YAML schema**: `entity-registry-schema` linter runs the JSON Schema validation at CI/CD time.

## Phase 2: Standalone Linters — No Registry Dependency

**Completed**: 2026-03-30 (sessions 4-8)

### What Worked

- **Test seam pattern**: Package-level `var seamFn = realFn` with `t.Cleanup(func() { seamFn = original })` enables comprehensive blackbox/whitebox testing of OS interactions (ReadFile, WalkDir, Getwd). Tests using seams MUST NOT use `t.Parallel()`.
- **YAML profiles use `any` for `default_curve_or_size`**: Using `any` instead of a specific type for fields that can be null or a mix of string/int is the correct approach for YAML deserialization with `gopkg.in/yaml.v3`.
- **Pre-commit hooks auto-format JSON**: `pretty-format-json` hook modifies JSON files. Always expect CRLF/format-related failures on first commit; commit twice.
- **Magic constants are mandatory**: The `literal-use [blocking]` linter catches all string/int literals that have corresponding magic constants. Always run `go run ./cmd/cicd-lint lint-go 2>&1 | Select-String "literal-use"` after introducing any string/int literals.
- **`min_days: 0` is valid** for short-lived certificates (kubernetes-workload, ssh-user). Don't assume minimum 1 day.
- **`default_curve_or_size: null` is valid** for Ed25519 (no curve/size parameter needed). Use `k.DefaultAlgorithm != magic.EdCurveEd25519` as guard.
- **AST-based alias validation**: The `import_alias_formula` linter uses `go/ast` and `go/parser` for accurate Go source analysis — avoids false positives from regex-based approaches.

### What Didn't Work

- **Using hardcoded literals in test struct fields**: Violated literal-use gate. Must always use magic constants even in test files.
- **Assuming profile count from tasks.md** (stated "25 profiles"): Actual count is 24. Always count directly from filesystem, not documentation.

### Root Causes

- Literal-use violations in test files: Not importing/using magic constants when constructing test data structs (e.g., `365` instead of `magic.DefaultCertificateMaxAgeDays`, `"RSA"` instead of `magic.KeyTypeRSA`).
- Gremlins timing out on Windows: System-level issue. Deferred to CI/CD as documented in tasks.md.

### Patterns

- **Fitness linter template**: Test file should include 3 seam tests (ReadFile error, WalkDir error, Getwd error) + happy path + ~10-15 violation tests + ~5 direct unit tests on validation functions = ~30 total tests for ≥95% coverage.
- **PKI-CA profile validation exceptions**: min_days=0 OK (short-lived), null default_curve_or_size OK for Ed25519.
- **Gremlins on Windows**: All mutants time out. Use CI/CD for mutation testing.

## Phase 3: Derivation Functions — Registry Consumers

**Completed**: 2026-03-30 (session 9, commit d76097707)

### What Worked

- **Derivation functions in a separate file**: Placing all derivation functions in `registry/derivations.go` (not `registry.go`) keeps the registry loader clean and makes derivations independently testable.
- **Fixture-based test pattern for derivations**: Table-driven tests iterating all 10 PS-IDs provide fast confirmation that formulas apply uniformly.
- **Injectable functions for test seams**: Using `var otlpReadDirFn = os.ReadDir` / `var otlpReadFileFn = os.ReadFile` (and equivalents) and restoring in `defer` allows testing error paths without OS setup.
- **wsl_v5 blank-line-before-defer rule**: When a seam test does `original := varFn` then `defer func() { varFn = original }()`, a blank line is REQUIRED between the assignment and the defer. Pattern: `original := x`, blank line, `defer func() { x = original }()`.
- **TestCheck_DelegatesToCheckInDir with os.Chdir**: When `Check()` uses `"."` as root and tests run from package dirs, use `os.Chdir(projectRoot)` + defer restore (non-parallel) to test the real entry point.

### What Didn't Work

- **Wrong ValidComposeServiceNames() variants**: Initial implementation used `sqlite`, `pg-1`, `pg-2` instead of the canonical `sqlite-1`, `postgres-1`, `postgres-2`. Always verify variant names against actual compose service names in deployments/.
- **Assuming compose_port_formula TestCheck_DelegatesToCheckInDir passes silently**: That package's linter skips missing files — so `Check(".")` returns nil even from package dir. But compose_db_naming and compose_service_names FAIL on missing files. Each linter's error behavior differs.
- **Magic literal `:8080` in test string**: Used literal `":8080"` as container port suffix. Must use `cryptoutilSharedMagic.TestPort`. The `lint-go literal-use [blocking]` rule catches this.

### Root Causes

- **Variant name mismatch**: Existing compose files used `sqlite-1`/`postgres-1`/`postgres-2` but the registry constants had been written as `sqlite`/`pg-1`/`pg-2`. Root cause: created constants from memory rather than checking actual compose files.
- **wsl_v5 violations in seam tests**: The `wsl_v5` linter requires blank lines around `defer` in certain cases. Sequential seam tests that save and restore package-level vars are a reliable trigger.
- **4 deployment config YAML files with wrong OTLP names**: `jose-ja` and `skeleton-template` postgresql config files had `otlp-service: "jose-e2e"` — a stale copy-paste from when the configs were scaffolded. The new `checkDeploymentConfigs()` function caught this correctly.

### Patterns

- **Deployment config YAML naming**: Config overlay files use `postgresql` in filename (e.g. `jose-ja-app-postgresql-1.yml`) but the OTLP service name value uses `postgres` not `postgresql` (e.g. `jose-ja-postgres-1`). These are intentionally different conventions.
- **Three-tier compose structure**: SERVICE tier = `deployments/{psid}/compose.yml`, PRODUCT tier = `deployments/{product}/compose.yml`, SUITE tier = `deployments/cryptoutil/compose.yml`. Port offsets: +0/+10000/+20000.
- **Compose port formula variant offsets**: `sqlite-1`=+0, `postgres-1`=+1, `postgres-2`=+2 (base_port + tier_offset + variant_offset).
- **Sequential seam test NON-PARALLEL requirement**: Tests modifying package-level `var` function pointers MUST NOT use `t.Parallel()`. Document this with comment `// Sequential seam test — must not use t.Parallel().`

## Phase 4: Secret & Config Schema Validation

**Completed**: 2026-03-30 (sessions 10-11, commits bd217babd, 2e57c132e + Phase 4 commit)

### What Worked

- **`go:embed` for YAML schemas**: Embedding YAML validation templates via `//go:embed schema.yaml` on a `var embeddedYAML []byte` seam makes the schema portable with the binary while remaining overridable in tests.
- **Schema-driven linter rewrite**: Replacing hardcoded regex in `secret_content.go` with a loaded YAML schema reduces brittleness and makes new secret types additive (add YAML entry, no Go changes needed).
- **Skip logic for missing config dirs**: When a linter iterates all 10 PS-IDs from registry but test temp dirs only contain a subset, returning `nil` (skip) when `deployments/{psid}/config/` doesn't exist enables clean unit test isolation per PS-ID.
- **variantSuffixes map keys use ComposeVariant constants**: Template YAML `variant:` names must match `ComposeVariant*` constants (`sqlite-1`, `postgres-1`, `postgres-2`)—not the file suffix keywords (`postgresql-1.yml`). The map handles the namespace translation.
- **8 missing `database-url` keys discovered by linter**: The linter immediately caught real drift (jose-ja, sm-im, skeleton-template, pki-ca sqlite overlays missing `database-url`). The linter correctly enforced the invariant before the config files existed.

### What Didn't Work

- **Initial template YAML used `postgresql-1` as variant name**: Template was written with `postgresql-1` matching the file name suffix, but `variantSuffixes` map keys use `ComposeVariant*` constants which are `postgres-1`. Root cause: didn't verify constant values before writing YAML.
- **wsl_v5 and nlreturn violations in linter code**: Required blank lines before `continue` (nlreturn) and before `if err` after multiple statements (wsl_v5). These are systematic violations that need consideration when writing any loop body.

### Root Causes

- **Variant name mismatch**: Template YAML must use values matching Go `ComposeVariant*` constants, not the file naming convention. Always verify constant values before writing YAML.
- **pki-ca dual database key**: pki-ca config overlays use both `database.dsn` (legacy) and `database-url` (standard). The linter correctly requires `database-url`; both keys co-exist with a comment explaining the dual-key design.
- **nlreturn/wsl_v5 in range loops**: Any `for range` loop body with multi-line logic needs careful blank line management. Pattern: blank line before `continue`, blank line before `if err :=` when preceded by multi-statement blocks.

### Patterns

- **go:embed + overlay seam**: `var overlayTemplatesYAML []byte` (embedded) + `var overlayReadFileFn = os.ReadFile` (injectable). Both can be overridden in non-parallel seam tests for error path coverage.
- **Config overlay template YAML location convention**: Template embedded in linter package (`lint_fitness/config_overlay_freshness/config-overlay-templates.yaml`). Documentation copy in lint_deployments (`lint_deployments/config-overlay-templates.yaml`).
- **Database-url standard for SQLite overlays**: ALL sqlite overlay files (`*-app-sqlite-1.yml`, `*-app-sqlite-2.yml`) MUST contain `database-url: "sqlite://file::memory:?cache=shared"`. PostgreSQL overlays MUST NOT contain `database-url`.
- **`OTLPServiceSMKMS` constant**: Use `cryptoutilSharedMagic.OTLPServiceSMKMS` instead of literal `"sm-kms"` in all test code.

## Phase 5: Deployment & Build Validation

**Completed**: 2026-03-31 (Tasks 5.1–5.3)

### What Worked

- **Explicit `entrypoint` field per PS-ID in registry.yaml**: When ENTRYPOINT patterns are
  inconsistent across PS-IDs (own binary vs. suite binary + subcommand + start), adding an
  explicit `entrypoint` field to each PS-ID in `registry.yaml` is cleaner than deriving the
  pattern algorithmically. Enables `DockerfileEntrypoint(psID)` derivation with zero ambiguity.
- **Exact title matching via `OTLPServiceName()` for PS-ID dirs**: Switching from substring
  match to exact case-insensitive comparison for PS-ID Dockerfiles caught a real bug —
  `identity-spa` had title `cryptoutil-identity-spa-rp` (wrong `-rp` suffix from copy-paste).
  Suite/product dirs kept substring match (legitimate multi-service images).
- **Set-membership via `buildValidServiceSet()` with PS-ID prefix filter**: For compose
  service name validation, check only services whose name starts with `psID+"-"` against the
  valid set built from `ValidComposeServiceNames()` + `DBServiceName()`. Infrastructure helper
  services (`pki-init`, `healthcheck-secrets`, `builder-*`) don't carry the PS-ID prefix and
  pass through naturally without an exemption list.
- **`buildValidServiceSet()` computed once in `CheckInDir`**: Build the full valid set once
  and pass it as a parameter to `checkServiceNames()` for all 10 PS-IDs, avoiding 10 redundant
  set builds per invocation.

### What Didn't Work

- **Initial set-membership checked ALL service names**: Rejected legitimate infrastructure
  services (`pki-init`, `healthcheck-secrets`, `builder-{psID}`) present in all 10 real compose
  files. Fix: only validate services with the PS-ID prefix via `strings.HasPrefix`.
- **Magic literal in `want []string{...}` test expected values**: Even in test assertion slices
  like `want: []string{..., "identity-authz", ...}`, the identity PS-ID string literal
  triggers `literal-use [blocking]`. Must use `cryptoutilSharedMagic.OTLPService*` constants
  in expected values too — not just production code.

### Root Causes

- **identity-spa title bug**: Dockerfile had `cryptoutil-identity-spa-rp` — the `-rp` suffix
  was an incorrect copy-paste from `identity-rp`. Registry-driven exact matching caught this;
  substring matching had silently accepted it for months.
- **magic literal-use in test `want` slices**: The `lint-go literal-use` scanner examines ALL
  Go string literals, including struct field initializers inside table-driven test `want` slices.
  Pattern: all PS-ID string literals in tests must use the `cryptoutilSharedMagic` constants.

### Patterns

- **ENTRYPOINT registry field**: For services with non-uniform ENTRYPOINT, declare canonical
  entrypoint array in `registry.yaml` per PS-ID under the `entrypoint` key. Schema enforces
  `minItems: 1`. Derivation: `DockerfileEntrypoint(psID) []string`.
- **Structural coverage ceilings on Windows**: `os.Stat` permission errors and
  `bufio.Scanner.Err()` I/O errors cannot be injected without OS-level mocks. Accept
  93–97% per function as structural ceiling; overall ≥95% is the practical quality gate.
  Document the ceiling in test comments.
- **Compose prefix filter**: When validating compose service names by set membership, scope
  to `strings.HasPrefix(svc, psID+"-")` to automatically exclude infrastructure services
  without needing an explicit exemption list.
- **Pre-commit double-commit pattern**: PowerShell `WriteAllText` writes system line endings.
  `end-of-file-fixer` and `mixed-line-ending` pre-commit hooks fix on first commit attempt;
  second commit succeeds. Always expect two `git commit` invocations when new files are
  written via PowerShell.

## Phase 6: API & Health Completeness

**Lessons Learned:**

1. **`SkeletonTemplateServiceID` vs `SkeletonTemplateServiceName`**: When filtering the skeleton-template PSID during service iteration, use `SkeletonTemplateServiceID = "skeleton-template"` (not `SkeletonTemplateServiceName = "template"`). PSID filtering must use the PSID constant, not the service name.

2. **`NewLogger` signature**: `NewLogger(operation string)` takes a string, not a bool. Use `NewLogger("test")` in test helpers.

3. **Health path count discrepancy**: Task description said "6 paths" but there are only 5 required health paths (`/service/api/v1/health`, `/browser/api/v1/health`, `/admin/api/v1/livez`, `/admin/api/v1/readyz`, `/admin/api/v1/shutdown`). Plan docs may contain minor count errors — verify by checking the actual magic constants.

4. **Test file scanning**: The health-path-completeness linter scans ALL top-level `.go` files in a service directory, including `_test.go` files. This was intentional — sm-im's service health path is documented in its `http_test.go`. When checking real workspace, test files count toward compliance.

5. **Pre-commit hook blocks commit for pre-existing issues**: The `cicd-lint-all` pre-commit hook runs lint-go, which fails if ANY magic-aliases violation exists anywhere in the codebase. Pre-existing violations (`domainMaxLength` in `usernames_passwords_test_util.go`, `numConcurrentInfoOps` in `sysinfo_all.go`) blocked all new commits until fixed. Always run `go run ./cmd/cicd-lint lint-go` before committing new code.

6. **`magic-usage` literal-use violations**: The `magic-usage` linter flags literal octal file permissions (e.g. `0o755`, `0o600`) as blocking violations. Use the corresponding magic constants: `FilePermOwnerReadWriteExecuteGroupOtherReadExecute` for `0o755` and `FilePermissionsDefault` for `0o600`.

7. **Fitness skill template**: The `.github/skills/fitness-function-gen/SKILL.md` prescribes `cryptoutilSharedMagic.CacheFilePermissions` for `0o600` and `cryptoutilSharedMagic.CICDTempDirPermissions` for `0o755`, but the magic-usage linter recommends `FilePermissionsDefault` and `FilePermOwnerReadWriteExecuteGroupOtherReadExecute`. Both are semantically equivalent, but the linter recommendation is authoritative. The SKILL.md should be updated.

8. **Usage file documentation gap**: All services need both `/service/api/v1/health` (for service-to-service clients) and `/browser/api/v1/health` (for browser clients) in their `*_usage.go` commentary. The health-path-completeness linter now enforces this as a fitness function.

**Phase 6 Quality Gate Results:**
- Build: ✅ clean
- golangci-lint: ✅ 0 issues on Phase 6 code
- lint-go: ✅ SUCCESS (after fixing pre-existing magic-alias violations)
- lint-fitness: ✅ SUCCESS (all 68 linters pass)
- Coverage: ✅ 96.3% across all lint_fitness packages (≥95% gate met)
- Tests: ✅ all pass


## Phase 7: Knowledge Propagation

*(To be filled during Phase 7 execution)*
