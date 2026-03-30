# Lessons — Parameterization Opportunities

**Created**: 2026-03-29
**Last Updated**: 2026-03-29

## Phase 1: Foundation — Entity Registry YAML

*(To be filled during Phase 1 execution)*

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

*(To be filled during Phase 3 execution)*

## Phase 4: Secret & Config Schema Validation

*(To be filled during Phase 4 execution)*

## Phase 5: Deployment & Build Validation

*(To be filled during Phase 5 execution)*

## Phase 6: API & Health Completeness

*(To be filled during Phase 6 execution)*

## Phase 7: Knowledge Propagation

*(To be filled during Phase 7 execution)*
