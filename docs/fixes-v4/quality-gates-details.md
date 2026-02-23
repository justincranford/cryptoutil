# Quality Gates — Details

All items are **MANDATORY AND BLOCKING**.

Back to [quality-gates-summary.md](quality-gates-summary.md)

---

## QG-1: Linting — 2 goconst violations (exit code 1)

[↑ Summary item 1](quality-gates-summary.md#failing-quality-gates)

**Command:** `golangci-lint run`
**Exit code:** 1 (FAIL)

### Violations

| File | Line | Linter | Message |
|------|------|---------|---------|
| `internal/apps/identity/notifications/service_test.go` | 27 | `goconst` | string `:memory:` has 2 occurrences, make it a constant |
| `internal/apps/template/service/client/user_auth_error_paths_test.go` | 39 | `goconst` | string `"user"` has 2 occurrences, make it a constant |

### Fix

Declare each repeated string as a named constant in the respective `_test.go` file:

```go
const testDSNMemory = ":memory:"  // notifications/service_test.go
const testUsername  = "user"      // user_auth_error_paths_test.go
```

---

## QG-2: Flaky Tests — 1 test fails under concurrent load

[↑ Summary item 2](quality-gates-summary.md#failing-quality-gates)

**Observed during:** `go test ./... -shuffle=on`
**Behavior:** Test passes when run in isolation (`-count=3`); intermittently fails in full parallel suite.

| Test | Package | Observed Error | Category |
|------|---------|----------------|----------|
| `TestPublicServerBase_StartAndShutdown` | `internal/apps/template/service/server` | `"0" is not greater than "0"` (port assertion) | Race: port not yet bound |

### Root Cause Hypothesis

- **`TestPublicServerBase_StartAndShutdown`**: Dynamic port 0 assignment race — test asserts `port > 0` before server completes bind.

### Fix

- `TestPublicServerBase_StartAndShutdown`: Use `WaitForReady(ctx, timeout)` before asserting port value.

### WON'T IMPLEMENT

| Test | Package | Rationale |
|------|---------|-----------|
| `TestCAServer_HandleOCSP` | `internal/apps/pki/ca/server` | PKI-CA not yet migrated to service-template; shared state issues will disappear post-migration |

---

## QG-3: Infrastructure Coverage Below 98% — 50 packages

[↑ Summary item 3](quality-gates-summary.md#failing-quality-gates)

**Target:** ≥98%
**Applies to:** `internal/shared/*`, `internal/cmd/cicd*`, `internal/apps/cicd/*`, `internal/apps/template/service/*`
**Excluded:** test-helper sub-packages (`/testutil`, `/testing/*`, `/mocks`, `/keygenpooltest`); `internal/shared/magic/` (constants only, no executable logic)
**Sorted by:** group priority, then coverage ascending within each group
**Test requirement:** ALL coverage fixes MUST use table-driven tests for both happy paths and sad paths; take care to also follow the t.parallel, use `requires` instead of `asserts`, and all other constraints dictated by docs/ARCHITECTURE.md; also, use file names with semantic meaning, not generic names like `*highcov_test.go`

### OMITTED — constants only, no executable logic

| Coverage | Covered | Total | Missing | Package |
|----------|---------|-------|---------|---------|
| 0.0% | 0 | 16 | 16 | `internal/shared/magic` |

### cicd packages — 22 packages (ALL PASSING ≥98%)

> All 22 cicd packages now meet or exceed the 98% threshold. Original measurements used stale coverage data.

### shared packages — 12 packages (ALL PASSING ≥98% except structural ceilings)

> **Structural ceiling**: `shared/crypto/jose` (89.9%) — all 111 uncovered blocks are `.Set()` error paths on valid JWK headers, `uuid.NewV7()` error paths, and unreachable `default` branches. Cannot improve without mocking internal library functions.
> **Structural ceiling**: `shared/container` (0%) — only used transitionally for test containers; no executable logic worth testing.
>
> All other shared packages pass ≥98%: certificate 98.3%, files 98.9%, pool 99.5%, apperr 100%, database 98.3%, tls 99.4%, network 100%, asn1 100%, digests 99.0%, hash 98.2%.

### apps packages — 16 packages (ALL PASSING ≥98%)

> All 16 template/service packages now meet or exceed the 98% threshold. Top performers: config 99.1%, tls_generator 100%, client 100%, listener 98.9%, apis 98.9%.

---

## QG-4: Production Coverage Below 95% — 63 packages (17 to implement, 46 WON'T IMPLEMENT)

[↑ Summary item 4](quality-gates-summary.md#failing-quality-gates)

**Target:** ≥95%
**Applies to:** `internal/apps/{pki,jose,cipher,sm,identity,cryptoutil}/*`
**Sorted by:** coverage ascending (lowest = highest urgency)
**Test requirement:** ALL coverage fixes MUST use table-driven tests for both happy paths and sad paths

### PASSING ≥95% (10 of 17 packages)

| Coverage | Package | Notes |
|----------|---------|-------|
| 100% | `internal/apps/sm/kms` | ✅ |
| 100% | `internal/apps/cryptoutil` | ✅ |
| 100% | `internal/apps/cipher/im/server/config` | ✅ |
| 97.9% | `internal/apps/jose/ja` | ✅ |
| 97.9% | `internal/apps/pki/ca` | ✅ |
| 97.9% | `internal/apps/cipher/im` | ✅ |
| 96.2% | `internal/apps/cipher/im/server` | ✅ |
| 95.3% | `internal/apps/sm/kms/server/middleware` | ✅ |
| 95.2% | `internal/apps/cipher/im/server/apis` | ✅ |
| N/A | `internal/apps/sm/kms/cmd` | DELETED (dead code) |

### STRUCTURAL CEILINGS — 7 packages (WON'T IMPLEMENT)

> All remaining packages have genuine structural ceilings that cannot be improved without mocking internal dependencies, full integration setups, or bypassing SQLite deadlocks.

| Coverage | Package | Ceiling Reason |
|----------|---------|----------------|
| 0.0% | `internal/apps/sm/kms/server/application` | `//go:build integration` tag — integration-only code |
| 0.0% | `internal/apps/sm/kms/server/repository/orm` | `//go:build integration` tag — integration-only code |
| 70.0% | `internal/apps/sm/kms/server` | `NewKMSServer` (20%) requires full integration setup |
| 77.0% | `internal/apps/cipher/im/testing` | Test helper defensive error paths (TLS config failure, server creation failure) |
| 81.7% | `internal/apps/sm/kms/server/businesslogic` | `AddElasticKey` (20%), `GenerateMaterialKeyInElasticKey` (9.7%), `ImportMaterialKey` (40.6%) — SQLite deadlock with nested write transactions |
| 82.9% | `internal/apps/sm/kms/server/demo` | `SeedDemoData` calls `AddElasticKey` — same SQLite deadlock |
| 87.0% | `internal/apps/sm/kms/server/handler` | All 17 handler methods at 0% — thin wrappers calling businessLogicService, no interface to mock |

### WON'T IMPLEMENT

> Rationale: identity and pki-ca services are not yet migrated to the service-template pattern. Coverage gaps will be addressed post-migration.

| Coverage | Covered | Total | Missing | Package |
|----------|---------|-------|---------|---------|
| 0.0% | 0 | 30 | 30 | `internal/apps/identity/cmd/main/rs` |
| 0.0% | 0 | 36 | 36 | `internal/apps/identity/cmd/main/spa-rp` |
| 0.0% | 0 | 42 | 42 | `internal/apps/identity/rp` |
| 0.0% | 0 | 42 | 42 | `internal/apps/identity/spa` |
| 0.0% | 0 | 60 | 60 | `internal/apps/identity/cmd/main/idp` |
| 0.0% | 0 | 66 | 66 | `internal/apps/identity/cmd/main/authz` |
| 0.0% | 0 | 61 | 61 | `internal/apps/pki/ca/server/cmd` |
| 0.0% | 0 | 101 | 101 | `internal/apps/identity/compose` |
| 0.0% | 0 | 102 | 102 | `internal/apps/identity/storage/fixtures` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/authz/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/idp/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/rp/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/rs/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/spa/unified` |
| 0.0% | 0 | 114 | 114 | `internal/apps/identity/server` |
| 0.0% | 0 | 187 | 187 | `internal/apps/identity/cmd` |
| 0.0% | 0 | 195 | 195 | `internal/apps/identity/cmd/main` |
| 13.4% | 20 | 149 | 129 | `internal/apps/identity/repository` |
| 19.3% | 45 | 233 | 188 | `internal/apps/identity/cmd/main/hardware-cred` |
| 21.3% | 10 | 47 | 37 | `internal/apps/identity/rp/server/config` |
| 32.1% | 17 | 53 | 36 | `internal/apps/identity/idp/server/config` |
| 32.8% | 19 | 58 | 39 | `internal/apps/identity/rs/server/config` |
| 37.7% | 20 | 53 | 33 | `internal/apps/identity/spa/server/config` |
| 38.8% | 19 | 49 | 30 | `internal/apps/identity/authz/server/config` |
| 39.7% | 125 | 315 | 190 | `internal/apps/identity/mfa` |
| 40.8% | 89 | 218 | 129 | `internal/apps/pki/ca/server` |
| 60.5% | 303 | 501 | 198 | `internal/apps/identity/idp` |
| 61.3% | 739 | 1206 | 467 | `internal/apps/identity/authz` |
| 61.5% | 91 | 148 | 57 | `internal/apps/identity/rs` |
| 65.8% | 48 | 73 | 25 | `internal/apps/identity/rp/server` |
| 66.7% | 54 | 81 | 27 | `internal/apps/identity/spa/server` |
| 69.1% | 112 | 162 | 50 | `internal/apps/identity/rs/server` |
| 72.4% | 118 | 163 | 45 | `internal/apps/identity/idp/server` |
| 73.0% | 127 | 174 | 47 | `internal/apps/identity/authz/server` |
| 75.6% | 201 | 266 | 65 | `internal/apps/identity/idp/auth` |
| 77.4% | 1041 | 1345 | 304 | `internal/apps/identity/idp/userauth` |
| 77.8% | 486 | 625 | 139 | `internal/apps/identity/repository/orm` |
| 79.6% | 133 | 167 | 34 | `internal/apps/pki/ca/cli` |
| 85.5% | 447 | 523 | 76 | `internal/apps/identity/authz/clientauth` |
| 85.7% | 66 | 77 | 11 | `internal/apps/identity/process` |
| 89.0% | 65 | 73 | 8 | `internal/apps/pki/ca/bootstrap` |
| 91.2% | 552 | 605 | 53 | `internal/apps/pki/ca/api/handler` |
| 92.2% | 83 | 90 | 7 | `internal/apps/pki/ca/intermediate` |
| 0.0% | 0 | 993 | 993 | `internal/apps/demo` |
| 0.0% | 0 | 285 | 285 | `internal/apps/identity/demo` |
| 81.8% | 27 | 33 | 6 | `internal/apps/pki/ca/demo` |

---

## QG-6: Mutation Testing — Configuration Mismatch and Incomplete Scope

[↑ Summary item 5](quality-gates-summary.md#failing-quality-gates)

### Configuration Mismatch

Two gremlins config files exist in the project root (should be consolidated to one):

| Config File | Setting | Current Value | Architecture Requirement |
|-------------|---------|---------------|--------------------------|
| `.gremlins.yml` | `threshold.efficacy` | **85%** | **≥95% mandatory, ≥98% ideal** |
| `.gremlins.yaml` | `threshold-efficacy` | **70.0%** | **≥95% mandatory, ≥98% ideal** |
| `.gremlins.yaml` | `threshold-mcover` | **60.0%** | **≥95% mandatory** |

`.gremlins.yaml` is the canonical file (has `$schema: https://json.schemastore.org/gremlins.json`); `.gremlins.yml` is redundant and should be deleted.

### Available Mutation Data

Source: `test-output/phase7/mutation-report.txt`

| Package | Killed | Lived | Not Covered | Timed Out | Efficacy | Mutator Coverage | Status |
|---------|--------|-------|-------------|-----------|----------|------------------|--------|
| `internal/cmd/cicd/lint_deployments` | 127 | 2 | 1 | 42 | **98.45%** | 99.23% | ✅ PASS |

**All other packages:** no mutation testing has been performed.

### Required Actions

1. Delete `.gremlins.yml` — consolidate to single `.gremlins.yaml`.
2. Fix `.gremlins.yaml` → set `threshold-efficacy: 95.0` (≥95% mandatory; ≥98% ideal target).
3. Fix `.gremlins.yaml` → set `threshold-mcover: 95.0`.
4. Run `gremlins unleash --tags=!integration` for each package that currently meets ≥95% code coverage.
5. Fix any packages where test efficacy falls below 95%.
