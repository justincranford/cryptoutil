---
name: test-table-driven
description: "Generate table-driven Go tests conforming to cryptoutil project standards. Use when writing or reviewing Go tests to ensure correct t.Parallel() usage, UUIDv7 test data, require over assert, proper subtest structure, and TestMain for heavyweight resources."
argument-hint: "[package or function name]"
---

Generate table-driven Go tests for the specified function or package.

**Full Copilot original**: [.github/skills/test-table-driven/SKILL.md](.github/skills/test-table-driven/SKILL.md)

## Key Rules

- `t.Parallel()` MANDATORY in every test function and every subtest
- Use `googleUuid.NewV7()` for all test data IDs (not hardcoded strings or UUIDs)
- Use `require` (not `assert`) — fail-fast on first failure
- Table-driven with named subtests via `t.Run(tc.name, ...)`
- HTTP handler tests: use `app.Test()` (Fiber in-memory, NO real servers)
- TestMain pattern for heavyweight resources (DB, servers) — init once per package
- `<15s` per package; parallel execution required

## Template

```go
func TestFunctionName(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    InputType{ID: googleUuid.NewV7()},
            expected: ExpectedType{...},
        },
        {
            name:    "invalid input",
            input:   InputType{},
            wantErr: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            result, err := FunctionUnderTest(tc.input)

            if tc.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            require.Equal(t, tc.expected, result)
        })
    }
}
```

## TestMain Pattern (heavyweight resources)

```go
var (
    sharedDB     *gorm.DB
    sharedServer *application.Application
)

func TestMain(m *testing.M) {
    ctx := context.Background()
    var err error

    sharedDB, err = repository.InitSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, migrationsFS)
    if err != nil {
        log.Fatalf("failed to init db: %v", err)
    }

    os.Exit(m.Run())
}
```

## Error Path Testing via Function-Param Injection

Use struct fields (for methods) or fn parameters (for functions) — NOT package-level `var xxxFn`. Struct field injection is parallel-safe.

```go
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

See [ARCHITECTURE.md §10.2.4](../../docs/ARCHITECTURE.md#1024-test-seam-injection-pattern).
