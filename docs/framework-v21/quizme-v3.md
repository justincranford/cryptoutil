# Quizme - Framework v21 Phase 2 API Design Decisions (Round 3)

Purpose: Resolve only the remaining unresolved decisions (Q1 and Q2).

Resolved and no longer pending:
1. Q3 readiness default: admin readyz (`/admin/api/v1/readyz`) is mandatory for integration orchestration.
2. Q4 port policy: tests bind both listeners to port 0 and orchestration returns resolved URLs.
3. Q5 migration strategy: one-pass direct migration with no compatibility wrappers.

Canonical basis already available:
1. [docs/ENG-HANDBOOK.md](docs/ENG-HANDBOOK.md): integration tests + TestMain + dynamic test ports + admin readiness endpoints.
2. [.github/instructions/02-01.architecture.instructions.md](.github/instructions/02-01.architecture.instructions.md): tests MUST use port 0 and readiness endpoint definitions.
3. Propagation/sync validation: `go run ./cmd/cicd-lint lint-docs` passed in this session (instructions/agents/skills drift checks passed).

## Terminology Clarification (to remove ambiguity)

1. "Per-suite" in this quiz means per Go test package binary lifecycle (one `TestMain(m *testing.M)` setup shared by all tests in that package).
2. "Per-suite" does NOT mean cryptoutil product suite (`cmd/cryptoutil`) and does NOT mean product-level or PS-ID-level runtime deployment scope.
3. "Per-test" means each test function (or each subtest if chosen) gets its own fixture lifecycle.

## Question 1: Default Fixture Scope Model for test_orch_integration

What is being decided:
1. The default fixture lifecycle in integration tests started by `test_orch_integration`.
2. Whether setup/teardown happens once per package (shared) or per test function (isolated).

Concrete example for each option:

**A) Per-suite shared fixture default**
```go
var env *IntegrationEnv

func TestMain(m *testing.M) {
    env = NewIntegrationEnv() // one app + one DB per package
    code := m.Run()
    env.Close()
    os.Exit(code)
}

func TestA(t *testing.T) { use(env) }
func TestB(t *testing.T) { use(env) }
```

**B) Per-test isolated fixture default**
```go
func TestA(t *testing.T) {
    env := NewIntegrationEnv() // fresh app + DB
    t.Cleanup(env.Close)
    use(env)
}
```

**C) Hybrid shared DB + per-test app instance**
```go
var sharedDB *gorm.DB

func TestMain(m *testing.M) {
    sharedDB = NewSharedDB()
    os.Exit(m.Run())
}

func TestA(t *testing.T) {
    app := NewApp(sharedDB) // fresh app each test
    t.Cleanup(app.Close)
}
```

**D) Hybrid shared app + per-test DB namespace**
```go
var app *App

func TestMain(m *testing.M) {
    app = NewApp(nil)
    os.Exit(m.Run())
}

func TestA(t *testing.T) {
    ns := NewDBNamespace(t.Name())
    t.Cleanup(ns.Drop)
    app.UseNamespace(ns)
}
```

**Question**: Which default should `test_orch_integration` enforce?

**A)** Per-package (per `TestMain`) shared fixture by default; opt-in per-test isolation.
**B)** Per-test isolated fixture by default; opt-in shared fixture.
**C)** Shared DB + per-test app instance by default.
**D)** Shared app + per-test DB namespace by default.
**E)** Per-package (per `TestMain`) shared fixture by default; opt-in per-test isolation. THIS IS ALREADY COVERED IN DOCS/ENG-HANDBOOK.md, AND PROPAGATED TO COPILOT+CLAUDE INSTRUCTIONS/AGENTS/SKILLS!!!!!!!!!!!!!!!!!!!!

**Answer**: E

**Rationale**: This determines startup cost, parallel safety, and state-leak risk across 37 remaining migrations.

## Question 2: Error-Path Fixture Creation Contract

What is being decided:
1. How integration tests create deterministic failure scenarios.
2. How to avoid ad hoc, inconsistent failure setup in each package.

Concrete example for each option:

**A) Explicit factory APIs (pre-broken fixture constructors)**
```go
env := testorch.NewClosedDBFixture(t)      // DB already closed
env := testorch.NewInvalidTLSFixture(t)    // bad client cert / trust
env := testorch.NewInvalidDSNFixture(t)    // startup DSN failure path
```

**B) Hook injection callbacks (mutate valid fixture before startup)**
```go
env := testorch.NewFixture(t, testorch.WithMutator(func(cfg *Config) {
    cfg.DatabaseURL = "postgres://bad:dsn"
}))
```
Meaning: "hook injection callback" = a caller-supplied function that edits fixture config/state at a specific hook point.

**C) Failure profile enum + framework builder**
```go
env := testorch.NewFixtureWithProfile(t, testorch.FailureClosedDB)
env := testorch.NewFixtureWithProfile(t, testorch.FailureInvalidTLS)
```
Meaning: "failure profile enum" = named constants representing supported failure modes.

**D) Direct suite-managed failure setup (no framework contract)**
```go
// Each package writes custom code to break DB/TLS/startup behavior.
// No shared framework API.
```

**Question**: Which mechanism should be standard?

**A)** Explicit factory APIs returning pre-broken fixtures.
**B)** Hook injection callbacks that mutate a valid fixture before startup.
**C)** Failure-profile enum plus framework builders.
**D)** Direct suite-managed setup (no framework-level error fixture contract).
**E)**

**Answer**: A

**Rationale**: This determines readability, consistency, and maintenance overhead for all error-path integration tests.
