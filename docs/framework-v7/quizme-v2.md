# Quiz — Framework v7 Phase 0 Seam & Linter Decisions

**Version**: quizme-v2  
**Created**: 2026-04-03  
**Answers required before**: Task 0.10 (Q6) and Task 0.11 (Q1–Q5)

Answer each question with A, B, C, D, or E (write letter in **Answer:** field).

---

## Background: Seam Pattern Categories

60+ package-level function-variable seams (`var xxxFn = pkg.Func`) exist across 5 categories.
Each has distinct trade-offs for parallel-safety, API impact, and maintenance cost.

The fundamental problem: `t.Parallel()` cannot be used safely when two goroutines mutate the
same package-level var. `saveRestoreSeams(t)` + `t.Cleanup` restores state AFTER the test
but does NOT prevent concurrent access DURING the test (data race on `-race`).

Options available for each category:

| Option | Description | Parallel-safe | API impact | Effort |
|--------|-------------|--------------|------------|--------|
| A | `// Sequential:` exemption — no refactor | ❌ no | None | Zero |
| B | Function-param injection — pass `fn` at call site | ✅ yes | Medium (callers change) | Medium |
| C | Interface/struct injection — inject via constructor | ✅ yes | Larger | Large |
| D | Remove seam — accept structural ceiling | varies | None | Small |

---

## Question 1: Category A — Fitness Linter OS I/O Seams (~20 seams)

**Examples** (linters in `internal/apps/tools/cicd_lint/lint_fitness/`):

```go
// cmd_main_pattern/cmd_main_pattern.go
var cmdMainWalkFn = filepath.Walk

// cross_service_import_isolation/cross_service_import_isolation.go
var crossServiceWalkFn = filepath.Walk

// fitness_registry_completeness/fitness_registry_completeness.go
var fitnessRegistryReadFileFn = os.ReadFile
var fitnessRegistryReadDirFn  = os.ReadDir
var fitnessGetwdFn             = os.Getwd

// test_file_suffix_structure/test_file_suffix_structure.go
var testFileSuffixReadFileFn  = os.ReadFile
var testFileSuffixWalkDirFn   = filepath.WalkDir
var testFileSuffixGetwdFn     = os.Getwd
```

These seams exist to cover error paths like "Walk returns error" or "Getwd fails" in CI
tool tests. They live entirely in tool (non-production) code.

**Which approach should be used for these ~20 fitness-linter OS I/O seams?**

**A)** Keep `// Sequential: mutates package-level seam vars` exemption on all seam tests. No
refactoring. Accept that seam tests cannot run in parallel. Appropriate for tool code where
parallel test performance is not a priority.

**B)** Refactor each linter function to accept OS I/O functions as parameters, e.g.:
`func Lint(logger, walkFn, readFileFn)`. Tests inject stubs as call-site arguments. All
seam tests can then use `t.Parallel()`. Significant but mechanical refactor (~20 function
signatures change).

**C)** Group related OS I/O functions into an injectable struct per linter:
`type ioFuncs struct { walkFn, readFileFn, getwdFn }`. Pass via constructor or `Lint()`
param. Cleaner per-linter API; similar effort to B.

**D)** Remove all fitness-linter OS I/O seams. Test only happy paths in unit tests; cover
error paths via integration tests with real OS behaviour. Accept reduced coverage on error
paths as structural ceiling.

**E)**

**Answer:**

**Rationale**: Category A seams affect ~10 linter packages. B or C would require updating
every linter function signature AND registered test. D would reduce coverage scores.

---

## Question 2: Category B — Crypto/Random Seams (~9 seams)

**Examples** (shared production library in `internal/shared/crypto/`):

```go
// internal/shared/crypto/digests/hkdf.go
var digestsHKDFReadFn = func(reader interface{ Read([]byte) (int, error) }, buf []byte) (int, error) { ... }

// internal/shared/crypto/digests/pbkdf2.go
var digestsRandReadFn = crand.Read

// internal/shared/crypto/hash/hash_high_fixed_provider.go
var hashHighFixedHKDFFn = cryptoutilSharedCryptoDigests.HKDF

// internal/shared/crypto/pbkdf2/pbkdf2.go
var pbkdf2CrandReadFn = func(b []byte) (int, error) { return crand.Read(b) }

// internal/shared/barrier/unsealkeysservice/unseal_keys_service.go
var hkdfWithSHA256Fn = cryptoutilSharedCryptoDigests.HKDFwithSHA256
```

These seams exist in production crypto primitives to inject deterministic random or error
conditions in unit tests. Crypto callers call these functions billions of times in
production; any API change ripples to all call sites.

**Which approach should be used for these ~9 crypto/random seams?**

**A)** Keep `// Sequential:` exemption on all seam tests. Accept non-parallel seam tests in
crypto packages. Crypto functions keep their simple signatures; callers unchanged.

**B)** Add `io.Reader` parameter to relevant crypto functions, e.g.:
`func HKDF(rand io.Reader, ...)`. Tests inject `bytes.NewReader(...)`. Callers pass
`crand.Reader` in production. API break for all crypto callers.

**C)** Group crypto dependencies into an injectable struct passed via constructor or
functional option per crypto service. Seam tests inject the struct with a test double.
Targets only the service layer, not the primitive functions.

**D)** Remove crypto seams entirely. Test crypto correctness with real `crand.Read`. Error
paths (e.g., "read returns 0 bytes") are accepted as structural ceiling not coverable
without OS-level mocking. Document in coverage ceiling analysis.

**E)**

**Answer:**

**Rationale**: Crypto functions are production primitives called by all services. API
changes cascade widely. D eliminates coverage for rare-but-possible entropy depletion paths.

---

## Question 3: Category C — Network/Server Seams (~5 seams)

**Examples** (framework listener in `internal/apps/framework/service/server/listener/`):

```go
// listener/admin.go
var generateTLSMaterialFn = cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial
var adminListenFn = func(ctx context.Context, network, address string) (net.Listener, error) { ... }
var adminAppListenerFn = func(app *fiber.App, ln net.Listener) error { ... }

// listener/public.go
var publicListenFn = func(ctx context.Context, network, address string) (net.Listener, error) { ... }
var publicAppListenerFn = func(app *fiber.App, ln net.Listener) error { ... }

// server/public_server_base.go
var appListenerFn = func(app *fiber.App, ln net.Listener) error { ... }
```

These seams let tests inject failures for "listen returns error" and "app.Listen returns
error". Server tests already use TestMain (heavyweight) and often real ports.

**Which approach should be used for the ~5 network/server seams?**

**A)** Keep `// Sequential:` exemption. Server listener tests are already heavyweight
(TestMain, real ports). Making them parallel would require dynamic ports anyway; Sequential
is acceptable given the existing test structure.

**B)** Add `WithListenFn(fn)` and `WithAppListenerFn(fn)` functional options to the server
builder. Tests inject no-op or error-producing functions via builder options. Package-level
vars removed; truly parallel-safe.

**C)** Inject a `net.Listener` directly into the server constructor. Tests pass a pre-bound
listener (e.g., `net.Listen("tcp", "127.0.0.1:0")`); no seams needed. Error paths tested
by passing a closed listener.

**D)** Remove all listener seams. Accept that "listen fails" error paths have no unit-test
coverage. Coverage from integration/E2E tests absorbs the gap.

**E)**

**Answer:**

**Rationale**: Server startup tests are already TestMain-based and take 1–2s. Seam behaviour
under B or C requires new builder plumbing. D narrows error-path coverage.

---

## Question 4: Category D — Internal Framework Dependency Seams (~6 seams)

**Examples** (framework initialization in `internal/apps/framework/service/server/`):

```go
// service_framework.go
var newTelemetryServiceFn = func(ctx, config) (*TelemetryService, error) { ... }
var newJWKGenServiceFn    = func(ctx, telemetry, devMode) (*JWKGenService, error) { ... }

// barrier/intermediate_keys_service.go
var intermediateGenerateJWEJWKFn = func(svc *JWKGenService) (...) { ... }

// application/application_basic.go
var newJWKGenServiceFn = cryptoutilSharedCryptoJose.NewJWKGenService

// service/server/realms/service.go
var realmsServiceHashSecretPBKDF2Fn = cryptoutilSharedCryptoHash.HashSecretPBKDF2

// service/server/repository/migrations.go
var migrateNewWithInstanceFn = migrate.NewWithInstance
```

These seams let unit tests inject failures at framework initialization (telemetry init,
JWK service init, migration engine init). Callers are framework internals, not user code.

**Which approach should be used for the ~6 framework dependency seams?**

**A)** Keep `// Sequential:` exemption. Framework init tests are rarely run in parallel;
Sequential cost is negligible given the already-heavyweight server test setup.

**B)** Define `TelemetryFactory`, `JWKFactory`, `MigrationFactory` interfaces. Inject via
constructor (`NewServiceFramework(ctx, config, factories...)`). Package-level vars removed;
parallel-safe. Medium refactor confined to framework internals.

**C)** Add functional options `WithTelemetryFactory(fn)`, `WithJWKFactory(fn)` to the server
builder. Elegant; backward-compatible; parallel-safe after option injection.

**D)** Remove factory seams. Test framework init only at integration level (TestMain with
real services). Accept that "telemetry init fails" and "JWK init fails" paths have no unit
coverage.

**E)**

**Answer:**

**Rationale**: Framework init happens once per server. Options B or C cleanly eliminate
global state. D removes coverage for init-failure resilience paths.

---

## Question 5: Category E — Single-Use Utility Seams (~5 seams)

**Examples** (scattered across multiple packages):

```go
// internal/apps/sm-im/client/message.go
var jsonMarshalFn = json.Marshal

// internal/apps/tools/cicd_lint/format_gotest/thelper/thelper.go
var printerFprintFn = func(output io.Writer, fset *token.FileSet, node any) error { ... }

// internal/apps/framework/service/server/middleware/session.go
var sessionMiddlewareStringsSplitNFn = strings.SplitN

// internal/apps/tools/cicd_lint/lint_workflow/github_actions/github_actions.go
var checkActionVersionsConcurrentlyFn = checkActionVersionsConcurrently

// internal/apps/tools/cicd_workflow/workflow.go
var cleanupFn = defaultCleanup
```

Each is a standalone seam in a single file, used for one specific error path.

**Which approach should be used for these ~5 single-use utility seams?**

**A)** Keep `// Sequential:` exemption on each seam test. Minimal maintenance cost; these
are isolated low-risk seams with no shared state across packages.

**B)** Convert each to a function parameter: e.g., `sm-im/client/message.go` receives
`marshalFn func(any) ([]byte, error)` instead of using the package var. Each injected
independently; no global state; parallel-safe. Small per-case refactor.

**C)** Remove all single-use seams and replace with interface injection at the type level.
Over-engineered for single-use cases.

**D)** Remove all single-use seams. Test the function with real `json.Marshal`, real
`strings.SplitN` etc. Error paths for standard library functions are structural ceilings;
document in coverage analysis. Accept reduced coverage.

**E)**

**Answer:**

**Rationale**: Each seam is one small package-level var in one file. B costs 1 function
signature change per case. D reduces coverage on rarely-failing stdlib paths.

---

## Question 6: Linter Absent-Directory Handling

**Context**: During planning, two linters were found to handle absent per-PS-ID service
directories inconsistently vs the majority of the 68 fitness linters.

**Majority pattern (correct, 60+ linters)**:
```go
if _, err := os.Stat(dir); os.IsNotExist(err) {
    return nil   // skip gracefully; absent dir is not a violation
}
```

**Minority pattern (found in 2 linters)**:
```go
// health_path_completeness: checkServiceHealthPaths()
entries, err := healthPathReadDirFn(svcDir)
if err != nil {
    return nil, fmt.Errorf("read service dir %s: %w", psID, err)
    // Returns HARD ERROR even if err is os.IsNotExist
}

// api_path_registry: collectSpecPaths()
entries, err := apiPathReadDirFn(apiDir)
if err != nil {
    return nil, fmt.Errorf("cannot read api directory %s: %w", apiDir, err)
    // Same: hard error on absent dir
}
```

**How should `health_path_completeness` and `api_path_registry` handle an absent per-PS-ID
service/API directory?**

**A)** Consistent with majority — skip gracefully: add `os.IsNotExist` check before the
`ReadDir` call; return `nil` (no violations) when directory is absent. Lenient; allows
PS-IDs being developed to pass the linter during scaffolding.

**B)** Treat absence as a VIOLATION — add "missing handler directory for PS-ID X" to the
violations list (not a hard error, not nil). Strict; all PS-IDs must have the directory;
CI fails with a clear violation message rather than a hard error.

**C)** Keep returning HARD ERROR for absent directories — inconsistent with majority but
intent may be intentional strictness. Caller framework receives error and aborts entire
lint run (not just that PS-ID).

**D)** Split: `health_path_completeness` → violation (health paths required for all PS-IDs);
`api_path_registry` → skip nil (API dir is optional during development).

**E)**

**Answer:**

**Rationale**: These are the only two linters where absence fails the entire lint run rather
than adding a violation or skipping. All other linters treat absence as "nothing to check."

---

*Answer all 6 questions above, then signal implementation to begin.*
