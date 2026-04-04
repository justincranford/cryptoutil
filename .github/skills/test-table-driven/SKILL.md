---
name: test-table-driven
description: "Generate table-driven Go tests conforming to cryptoutil project standards. Use when writing or reviewing Go tests to ensure correct t.Parallel() usage, UUIDv7 test data, require over assert, proper subtest structure, and TestMain for heavyweight resources."
argument-hint: "[package or function name]"
---

Generate table-driven Go tests conforming to cryptoutil project standards.

## Purpose

Use this skill when writing or reviewing Go tests. Ensures correct patterns:
`t.Parallel()`, `UUIDv7` test data, `require` over `assert`, proper subtests.

## Key Rules

- `t.Parallel()` MANDATORY on parent and ALL subtests
- Use `googleUuid.NewV7()` for test data IDs (thread-safe, unique, no conflicts)
- `require` package (fail-fast) over `assert` (continue-on-failure)
- Table-driven for ALL multi-case tests (happy path AND sad path)
- TestMain for heavyweight resources (DB, servers, containers) — one per package
- Fiber `app.Test()` for ALL HTTP handler tests (no real network listeners)
- SQLite DateTime: ALWAYS use `time.Now().UTC()` when comparing timestamps
- Timing: unit tests MUST complete in <15s per package; full suite <180s
- Probability-based execution: use `TestProbAlways=100`, `TestProbQuarter=25`, `TestProbTenth=10` for expensive algorithm variant tests (RSA sizes, ECDSA curves)

## Template

```go
func TestXxx_Description(t *testing.T) {
t.Parallel()
tests := []struct {
name    string
input   TypeA
want    TypeB
wantErr string
}{
{name: "happy path basic", input: ..., want: ...},
{name: "error case missing field", input: ..., wantErr: "missing X"},
}
for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()
got, err := FunctionUnderTest(tc.input)
if tc.wantErr != "" {
require.ErrorContains(t, err, tc.wantErr)
return
}
require.NoError(t, err)
require.Equal(t, tc.want, got)
})
}
}
```

## Fiber Handler Testing Pattern

**ALWAYS** use Fiber's in-memory testing for HTTP handler tests — never start real listeners:

```go
func TestListMessages_Handler(t *testing.T) {
 t.Parallel()

 app := fiber.New(fiber.Config{DisableStartupMessage: true})
 msgRepo := repository.NewMessageRepository(testDB)
 handler := NewPublicServer(nil, msgRepo, nil, nil, nil)
 app.Get("/browser/api/v1/messages", handler.ListMessages)

 req := httptest.NewRequest("GET", "/browser/api/v1/messages", nil)
 req.Header.Set("X-Tenant-ID", testTenantID.String())

 resp, err := app.Test(req, -1) // in-memory, <1ms, no network binding
 require.NoError(t, err)
 defer resp.Body.Close()

 require.Equal(t, 200, resp.StatusCode)
}
```

Benefits: no port conflicts, no Windows Firewall popups, tests run in <1ms.

## TestMain Pattern (heavyweight resources)

```go
var (
testDB     *gorm.DB
testServer *Server
)

func TestMain(m *testing.M) {
ctx := context.Background()
container, _ := postgres.RunContainer(ctx,
postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewV7())),
postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewV7())),
)
defer container.Terminate(ctx)
// initialize testDB, testServer...
os.Exit(m.Run())
}
```

## Error Path Testing via Function-Param Injection

**MANDATORY**: Use function-parameter injection (struct fields or fn params), NOT package-level `var xxxFn`. Tests that use struct fields are parallel-safe.

```go
// Struct method error path test
func TestDoSomething_EncryptError(t *testing.T) {
 t.Parallel()
 sm := setupSessionManager(t)
 sm.encryptBytesFn = func(_ []joseJwk.Key, _ []byte) (*joseJwe.Message, []byte, error) {
  return nil, nil, fmt.Errorf("injected encrypt error")
 }
 _, err := sm.DoSomething(ctx, input)
 require.ErrorContains(t, err, "injected encrypt error")
}
```

See [ARCHITECTURE.md §10.2.4](../../../docs/ARCHITECTURE.md#1024-test-seam-injection-pattern) for full decision matrix.

## Java / Gatling Load Test Pattern

Java Gatling simulations in `test/load/src/test/java/cryptoutil/` MUST follow these standards:

- **Secure RNG**: ALWAYS use `java.security.SecureRandom`, NEVER `new Random()` or `Math.random()`
- **Parameterization**: Use `System.getProperty("key", "default")` for all configurable values (base URLs, user counts, durations)
- **Simulation extension**: All simulation classes MUST extend `Simulation` — do not extend other test frameworks
- **Validated by**: `cicd-lint lint-java-test` — checks for insecure random number generation

**Correct pattern:**

```java
import java.security.SecureRandom;
import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;
import static io.gatling.javaapi.core.CoreDsl.*;

public class MyApiSimulation extends Simulation {
    private static final SecureRandom SECURE_RANDOM = new SecureRandom();
    private static final String BASE_URL = System.getProperty("baseUrl", "https://localhost:8080");
    private static final int USERS   = Integer.parseInt(System.getProperty("users", "10"));

    HttpProtocolBuilder protocol = http.baseUrl(BASE_URL);

    ScenarioBuilder scn = scenario("MyScenario")
        .exec(http("request").get("/service/api/v1/health").check(status().is(200)));

    { setUp(scn.injectOpen(atOnceUsers(USERS))).protocols(protocol); }
}
```

**Violations detected by `lint-java-test`:**
- `new Random()` — replace with `new SecureRandom()`
- `Math.random()` — replace with `secureRandom.nextDouble()`

## Python / pytest Pattern

Python test files (when present) MUST use pytest style:

- **pytest functions**: Use standalone `def test_*()` functions, NOT `class MyTest(unittest.TestCase)`
- **Parameterization**: Use `@pytest.mark.parametrize` decorator, NOT `self.assertEqual` loops
- **Assertions**: Use bare `assert` statements, NOT `self.assert*()` methods
- **File naming**: Test files MUST be named `test_*.py` or `*_test.py`
- **Validated by**: `cicd-lint lint-python-test` — checks for unittest.TestCase antipatterns

**Correct pattern:**

```python
import pytest

@pytest.mark.parametrize("value,expected", [
    ("valid",   True),
    ("invalid", False),
])
def test_validate_input(value, expected):
    result = validate_input(value)
    assert result == expected


@pytest.fixture
def client(base_url):
    return ApiClient(base_url)


def test_health_check(client):
    resp = client.get("/service/api/v1/health")
    assert resp.status_code == 200
```

**Violations detected by `lint-python-test` (in `test_*.py` and `*_test.py` files only):**
- `class MyTest(unittest.TestCase)` — replace with standalone functions
- `from unittest import TestCase` — use pytest instead
- `self.assert*(...)` calls — use bare `assert` or `pytest.raises()`

## References

Read [ARCHITECTURE.md Section 10.2 Unit Testing Strategy](../../../docs/ARCHITECTURE.md#102-unit-testing-strategy) for full testing requirements — apply all forbidden patterns, `t.Parallel()` rules, `TestMain` requirements, and coverage targets from this section.

Read [ARCHITECTURE.md Section 10.2.2 Fiber Handler Testing](../../../docs/ARCHITECTURE.md#1022-fiber-handler-testing-apptest) for handler test patterns — apply `app.Test()` for ALL HTTP handler tests.

Read [ARCHITECTURE.md Section 10.3.2 Test Isolation](../../../docs/ARCHITECTURE.md#1032-test-isolation-with-tparallel) for parallelism requirements — ensure `t.Parallel()` is applied correctly at all levels.

Read [ARCHITECTURE.md Section 10.3.6 Shared Test Infrastructure](../../../docs/ARCHITECTURE.md#1036-shared-test-infrastructure) for shared test helpers — use `testdb.NewInMemorySQLiteDB(t)`, `testserver.StartAndWait`, `fixtures.CreateTestTenant/Realm/User`, `assertions.AssertHealthy`, and `healthclient.NewHealthClient` when these test patterns apply to test infrastructure packages.
