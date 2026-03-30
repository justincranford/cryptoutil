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

*(To be filled during Phase 4 execution)*

## Phase 5: Deployment & Build Validation

*(To be filled during Phase 5 execution)*

## Phase 6: API & Health Completeness

*(To be filled during Phase 6 execution)*

## Phase 7: Knowledge Propagation

*(To be filled during Phase 7 execution)*
