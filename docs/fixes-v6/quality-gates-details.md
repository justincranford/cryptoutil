# Quality Gates — Details (fixes-v6)

Comprehensive audit of every project file against ARCHITECTURE.md. All findings are **MANDATORY AND BLOCKING**.

Back to [quality-gates-summary.md](quality-gates-summary.md)

---

## CRITICAL Priority

### F-6.1: math/rand in production code (FIPS violation)

- **File**: `internal/apps/jose/ja/repository/audit_repository.go` line 12
- **Severity**: CRITICAL
- **Issue**: Uses `math/rand` for audit sampling decisions. Architecture MANDATES `crypto/rand` ALWAYS (FIPS 140-3 compliance). `math/rand` is deterministic and predictable.
- **Fix**: Replace with `crypto/rand` based float generation or use `cryptoutil/internal/shared/util/random` utilities.

### F-6.2: context.Background() in HTTP handlers

- **File**: `internal/apps/template/service/server/apis/sessions.go` lines 100, 138
- **Severity**: CRITICAL
- **Issue**: Uses `context.Background()` instead of Fiber request context in HTTP handlers. Breaks trace correlation (OpenTelemetry spans not linked to request) and cancellation propagation.
- **Fix**: Use `c.UserContext()` from Fiber context to propagate request-scoped context.

### F-6.3: InsecureSkipVerify conditional in production CLI

- **File**: `internal/apps/template/service/cli/http_client.go` lines 72, 114
- **Severity**: CRITICAL
- **Issue**: `InsecureSkipVerify: caCertPool == nil` falls back to insecure TLS when no CA cert provided. Architecture says NEVER `InsecureSkipVerify: true` in production.
- **Fix**: Fail with clear error when no CA cert is available instead of silently disabling verification.

---

## HIGH Priority

### F-6.4: ValidateUUIDs wraps wrong error (bug)

- **File**: `internal/shared/util/random/uuid.go` line 53
- **Severity**: HIGH (bug)
- **Issue**: When `ValidateUUID` returns an error, the wrapping uses hardcoded `ErrUUIDsCantBeNil` instead of the actual `err`. This masks real validation failures (zero-UUID, max-UUID).
- **Fix**: Change `cryptoutilSharedApperr.ErrUUIDsCantBeNil` to `err` in the `%w` verb.

### F-6.5: Copy-paste bug — "sqlite" in PostgreSQL function

- **File**: `internal/shared/container/postgres.go` line 39
- **Severity**: HIGH (bug)
- **Issue**: `fmt.Errorf("failed to start sqlite container: %w", err)` inside `StartPostgres` function.
- **Fix**: Change to `"failed to start postgres container: %w"`.

### F-6.6: Generic error messages leak JWK context

- **File**: `internal/shared/apperr/app_errors.go` lines 13-15
- **Severity**: HIGH
- **Issue**: `ErrCantBeNil` says `"jwks can't be nil"` and `ErrCantBeEmpty` says `"jwks can't be empty"` but are used as GENERIC sentinel errors across UUID validation, JWE, JWS — all different domains.
- **Fix**: Change to `"value can't be nil"` and `"value can't be empty"`.

### F-6.7: Shared packages import from apps/template (dependency inversion)

- **Files**: `shared/pool/pool.go`, `shared/crypto/jose/jwkgen_service.go`, `shared/container/containers_util.go`, `shared/container/postgres.go`
- **Severity**: HIGH
- **Issue**: 4 production files in `internal/shared/` import `cryptoutil/internal/apps/template/service/telemetry`. Shared code should NOT depend on application-layer code.
- **Fix**: Extract telemetry interface/types to `internal/shared/telemetry/` and have `apps/template` implement it.

### F-6.8: Error sentinels typed as string not error

- **Files**: `shared/crypto/jose/jws_jwk_util.go` line 28, `jwe_jwk_util.go` line 30
- **Severity**: HIGH
- **Issue**: `ErrInvalidJWSJWKKidUUID` and `ErrInvalidJWEJWKKidUUID` are declared as `var ... = "string"` instead of `errors.New(...)`. Bypasses Go's `error` interface and `errors.Is()`/`errors.As()`.
- **Fix**: Change to `var ErrInvalidJWSJWKKidUUID = errors.New("...")` and update `ValidateUUID` to accept `error` instead of `*string`.

### F-6.9: Test helper file invisible — space in filename

- **File**: `internal/shared/util/random/usernames_passwords_test util.go`
- **Severity**: HIGH
- **Issue**: Filename contains a literal space. Go ignores it (not `_test.go` suffix pattern match). Test helpers in this file are invisible to the compiler.
- **Fix**: Rename to `usernames_passwords_test_util.go` (underscore instead of space).

### F-6.10: Identity magic package not in shared/magic

- **Files**: `internal/apps/identity/magic/` (12 files, 754 lines)
- **Severity**: HIGH
- **Issue**: Architecture MANDATES "ALL magic constants MUST be consolidated in `internal/shared/magic/`". Identity has a parallel magic package.
- **Fix**: Move generic constants to `internal/shared/magic/magic_identity_*.go`. Keep identity-specific OIDC/MFA/adaptive constants in shared magic with appropriate naming.

### F-6.11: PKI CA magic package not in shared/magic

- **File**: `internal/apps/pki/ca/magic/magic.go` (28 lines)
- **Severity**: HIGH
- **Fix**: Move to `internal/shared/magic/magic_pki_ca.go`.

### F-6.12: Identity config magic file not in shared/magic

- **File**: `internal/apps/identity/config/magic.go` (62 lines)
- **Severity**: HIGH
- **Issue**: Hardcoded ports (8200, 8300, 8400, 9090), timeouts (30, 60), pool sizes (25, 5) outside shared magic.
- **Fix**: Move to appropriate `internal/shared/magic/` files.

### F-6.13: Test files exceed 500-line hard limit

- **Files**:
  - `internal/apps/sm/kms/server/businesslogic/businesslogic_crud_test.go` (514 lines)
  - `internal/apps/sm/kms/server/businesslogic/oam_orm_mapper_test.go` (506 lines)
  - `internal/shared/crypto/tls/tls_error_paths_test.go` (504 lines)
- **Severity**: HIGH
- **Fix**: Split each into focused files by functionality.

### F-6.14: Identity server & cmd packages have zero tests

- **Files**:
  - `internal/apps/identity/server/` (4 files: authz_server.go, idp_server.go, rs_server.go, server_manager.go)
  - `internal/apps/identity/cmd/` (7 files, 432-line identity_cli.go)
- **Severity**: HIGH
- **Fix**: Add comprehensive tests. CLI should use thin main → testable `internalMain` pattern.

### F-6.15: //nolint:wsl violations (5 instances)

- **Files**: `template/service/telemetry/telemetry_service_helpers.go` (lines 134, 158), `identity/idp/unified/idp.go` (3 instances)
- **Severity**: HIGH
- **Issue**: Architecture says NEVER use `//nolint:wsl` — restructure code to group related logic.
- **Fix**: Restructure code to satisfy wsl linter naturally.

---

## MEDIUM Priority

### F-6.16: 35 test files missing t.Parallel()

- **Affected packages**: `cipher/im/server/apis`, `template/service/config`, `template/service/server/application`, `template/service/server/listener`, `template/service/telemetry`, `jose/ja/server/apis`, `pki/ca/observability`, `pki/ca/server/config`, `template/service/config/tls_generator`, `shared/crypto/jose`, `shared/crypto/hash`, `identity/rs`, and 23 more.
- **Severity**: MEDIUM
- **Fix**: Add `t.Parallel()` to all test functions and subtests.

### F-6.17: pool.go long if/else chain (12 branches)

- **File**: `internal/shared/pool/pool.go` lines 425-447
- **Severity**: MEDIUM
- **Fix**: Refactor `validateConfig` to switch statement.

### F-6.18–F-6.22: Files approaching 500-line hard limit

| File | Lines | Priority |
|------|-------|----------|
| `shared/pool/pool.go` | 451 | MEDIUM |
| `shared/crypto/certificate/certificates.go` | 474 | MEDIUM |
| `identity/issuer/jws.go` | 494 | MEDIUM |
| `pki/ca/cli/cli.go` | 492 | MEDIUM |
| `workflow/workflow_executor.go` | 491 | MEDIUM |

- **Fix**: Proactively split before they exceed 500.

### F-6.23: Container package has zero tests

- **File**: `internal/shared/container/` (3 files)
- **Severity**: MEDIUM
- **Issue**: Infrastructure code requiring ≥98% coverage. Zero tests exist.
- **Fix**: Add basic unit tests. For Docker-dependent functions, use build tags.

### F-6.24: Empty shared/barrier/ directory

- **File**: `internal/shared/barrier/`
- **Severity**: MEDIUM
- **Issue**: Referenced in architecture but actual implementation is in `apps/template/service/server/barrier/`.
- **Fix**: Remove empty directory or add `doc.go` explaining the redirected location.

### F-6.25: TLS chain.go constants outside magic

- **File**: `internal/shared/crypto/tls/chain.go` lines 126-132
- **Severity**: MEDIUM
- **Issue**: `DefaultCAChainLength = 3`, `DefaultCADuration = 10 * 365 * 24 * time.Hour`, `DefaultEndEntityDuration = 365 * 24 * time.Hour` should be in magic.
- **Fix**: Move to `internal/shared/magic/magic_pkix.go` (where similar constants already exist).

### F-6.26: Algorithm string constants in jose files not in magic

- **File**: `internal/shared/crypto/jose/jws_jwk_util.go` lines 30-42
- **Severity**: MEDIUM
- **Issue**: 13 algorithm string constants (`algRS512`, `algPS512`, etc.) defined locally instead of magic.
- **Fix**: Move to `internal/shared/magic/magic_jose.go`.

### F-6.27: pwdgen.go policy constants outside magic

- **File**: `internal/shared/pwdgen/pwdgen.go` lines 46-52
- **Severity**: MEDIUM
- **Fix**: Move to `magic_pwdgen.go` or keep local with documented exemption.

### F-6.28: unseal_keys_service else-if chain

- **File**: `internal/apps/template/service/server/barrier/unsealkeysservice/unseal_keys_service_sharedsecrets.go` lines 54-70
- **Severity**: MEDIUM
- **Fix**: Refactor 6-branch else-if to switch statement.

### F-6.29: context.Background() in barrier services

- **Files**: `root_keys_service.go` lines 56, 97; `intermediate_keys_service.go` lines 60, 94
- **Severity**: MEDIUM
- **Fix**: Accept and propagate `ctx context.Context` parameter.

### F-6.30: 91 httptest.NewServer usages in tests

- **Files**: Various test files
- **Severity**: MEDIUM
- **Issue**: Architecture says ALWAYS use Fiber `app.Test()` for HTTP handler tests. 91 usages of `httptest.NewServer`.
- **Fix**: Where testing handlers, prefer `app.Test()`. For external service simulation, `httptest.NewServer` is acceptable.

### F-6.31: Observability tests — 30 standalone functions

- **File**: `internal/apps/pki/ca/observability/observability_test.go`
- **Severity**: MEDIUM
- **Fix**: Consolidate related tests into table-driven tests.

### F-6.32: identity/rp/ and identity/spa/ have zero tests

- **Files**: `internal/apps/identity/rp/` (4 files), `identity/spa/` (4 files)
- **Severity**: MEDIUM
- **Fix**: Add tests for service entry points.

### F-6.33: pki/ca/domain/ has zero tests

- **File**: `internal/apps/pki/ca/domain/` (2 files)
- **Severity**: MEDIUM
- **Fix**: Add domain object validation tests.

### F-6.34: //nolint:wrapcheck,thelper blanket suppression

- **File**: `internal/shared/crypto/keygen/keygen_error_paths_test.go`
- **Severity**: MEDIUM
- **Fix**: Remove blanket nolint, fix underlying issues.

### F-6.35: jose package name mismatch

- **File**: `internal/shared/crypto/jose/` — directory `jose/` but `package crypto`
- **Severity**: MEDIUM
- **Issue**: Confusing for readers who expect `package jose`. Go allows it but naming should match.
- **Fix**: Consider renaming to `package jose` (requires updating all import aliases).

### F-6.36: Duplicate identity/demo constants with demo package

- **Files**: `internal/apps/identity/demo/demo.go`, `internal/apps/demo/identity.go`, `internal/apps/demo/integration.go`
- **Severity**: MEDIUM
- **Fix**: De-duplicate shared credentials and ports into magic package.

### F-6.37: Unused sentinel errors in database/sharding.go

- **File**: `internal/shared/database/sharding.go` line 96
- **Severity**: MEDIUM
- **Issue**: 6 of 8 sentinel errors declared but never used.
- **Fix**: Remove unused sentinels or mark with future-use comments.

### F-6.38: SQL interpolation in sharding (defense in depth)

- **File**: `internal/shared/database/sharding.go` line 101
- **Severity**: MEDIUM
- **Issue**: `SET search_path TO "%s"` and `CREATE SCHEMA IF NOT EXISTS "%s"` use `fmt.Sprintf`. Risk is low (tenant IDs = UUIDs) but defense-in-depth matters.
- **Fix**: Validate `schemaName` format before interpolation.

### F-6.39: Fmt.Errorf without %w audit needed

- **Files**: ~1,089 instances across codebase
- **Severity**: MEDIUM
- **Issue**: Many `fmt.Errorf` calls don't use `%w` wrapping. Some are validation errors (acceptable), but some wrap function return values that MUST use `%w`.
- **Fix**: Audit top offenders: `jwk_util_validate.go` (30), `certificates.go` (27), `jws_jwk_util.go` (25).

### F-6.39.1: semantic file names

- **Files**: *_cov*.go outside out cicd
- **Severity**: MEDIUM
- **Issue**: Many files have meaningless names or subnames, related to increasing code coverage to reach required coverage and mutations thresholds, for unit/integration/e2e tests
- **Fix**: List all les in internal/*, filter on filenames with names that don't match the semantic meaning of the contents of the file or the package; rename the file, or consolidate the tests contented in it to other meaningfully named test files within the package; take care to avoid duplicate tests, and take care to ensure tests following doc/ARCHITECTURE.md requirements like table-driven happy path tests, table-driven sad path tests, t.parallel(), use `requires` instead of `asserts`, and all of the other test constraints dictated by docs/ARCHITECTURE.md

---

## LOW Priority

### F-6.40: CICD import alias convention deviation

- **Files**: Various `internal/apps/cicd/` files
- **Severity**: LOW
- **Issue**: Uses `lintGo*`, `formatGo*` instead of `cryptoutil*` convention.
- **Fix**: Document as intentional exception or standardize.

### F-6.41: Hardcoded test password in testutil.go

- **File**: `internal/shared/testutil/testutil.go` line 127
- **Severity**: LOW
- **Fix**: Use dynamic generation or document why static is intentional.

### F-6.42: E2E test constants could be in magic

- **File**: `internal/apps/identity/test/e2e/constants.go`
- **Severity**: LOW
- **Fix**: Evaluate for consolidation.

### F-6.43: Template CLI constants not in magic

- **File**: `internal/apps/template/service/cli/constants.go`
- **Severity**: LOW
- **Fix**: Evaluate for consolidation.

### F-6.44: ValidateUUID takes *string pointer unnecessarily

- **File**: `internal/shared/util/random/uuid.go` line 34
- **Severity**: LOW
- **Fix**: Change to `msg string` value parameter.

### F-6.45: Hardcoded boundary UUIDs in tests (documented)

- **Files**: Various test files
- **Severity**: LOW
- **Issue**: Acceptable for boundary/edge-case testing.
- **Fix**: Document intent in comments.

### F-6.46: pool.go //nolint:errcheck on close()

- **File**: `internal/shared/pool/pool.go` lines 399-400
- **Severity**: LOW
- **Issue**: Channel close cannot fail. Acceptable.
- **Fix**: Add brief justification comments.
